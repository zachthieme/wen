package calendar

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const defaultHighlightPath = ".local/share/pike/due.json"

// LoadHighlightedDates loads highlighted dates from a JSON file.
// The file should contain a JSON array of yyyy-mm-dd strings.
// Returns nil if the file doesn't exist or is malformed (fail silently).
func LoadHighlightedDates(path string) map[time.Time]bool {
	if path == "" {
		return nil
	}

	// Expand ~ prefix
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil
		}
		path = filepath.Join(home, path[1:])
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var dates []string
	if err := json.Unmarshal(data, &dates); err != nil {
		return nil
	}

	result := make(map[time.Time]bool, len(dates))
	for _, s := range dates {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			continue
		}
		result[t] = true
	}
	if len(result) == 0 {
		return nil
	}
	return result
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
