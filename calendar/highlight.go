package calendar

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const defaultHighlightPath = ".local/share/pike/due.json"

// expandTilde replaces a leading ~ with the user's home directory.
// Returns the path unchanged if it doesn't start with ~ or if the home
// directory cannot be determined.
func expandTilde(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}

// LoadHighlightedDates loads highlighted dates from a JSON file.
// The file should contain a JSON array of yyyy-mm-dd strings.
// Returns nil with no warnings for an empty path (not configured).
// Returns nil with a warning if the file is missing or malformed.
// Individual unparseable date entries produce per-entry warnings.
func LoadHighlightedDates(path string) (map[time.Time]bool, []string) {
	if path == "" {
		return nil, nil
	}

	path = expandTilde(path)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, []string{fmt.Sprintf("highlight file not found: %s", path)}
	}

	var dates []string
	if err := json.Unmarshal(data, &dates); err != nil {
		return nil, []string{fmt.Sprintf("highlight file is not valid JSON: %s", path)}
	}

	var warnings []string
	result := make(map[time.Time]bool, len(dates))
	for _, s := range dates {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("skipping invalid date %q in %s", s, path))
			continue
		}
		result[t] = true
	}
	if len(result) == 0 {
		return nil, warnings
	}
	return result, warnings
}

// WithHighlightSource sets the path to a JSON file of dates to highlight.
// It expands ~ to the user's home directory, performs the initial load, and
// enables file watching when Init() runs. If both WithHighlightSource and
// WithHighlightedDates are used, the last one applied wins (WithHighlightedDates
// clears the highlight path, disabling file watching).
func WithHighlightSource(path string) ModelOption {
	return func(m *Model) {
		m.highlightPath = expandTilde(path)
		dates, warnings := LoadHighlightedDates(m.highlightPath)
		m.highlightedDates = dates
		m.warnings = append(m.warnings, warnings...)
	}
}

// ResolveHighlightSource determines the highlight file path based on priority:
// 1. CLI flag (if non-empty)
// 2. Config setting (if non-empty)
// 3. Default path (~/.local/share/pike/due.json) if it exists
func ResolveHighlightSource(flagValue, configValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if configValue != "" {
		return configValue
	}

	// Check default path
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	defaultPath := filepath.Join(home, defaultHighlightPath)
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}
	return ""
}
