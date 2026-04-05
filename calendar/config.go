package calendar

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Week numbering schemes.
const (
	WeekNumberingUS  = "us"
	WeekNumberingISO = "iso"
)

// WeekNumPos controls where week numbers appear relative to the grid.
type WeekNumPos int

// Week number display positions.
const (
	WeekNumOff   WeekNumPos = iota // hidden
	WeekNumLeft                    // left of grid (standard cal -w style)
	WeekNumRight                   // right of grid
)

// String constants for YAML/config normalization.
const (
	weekNumStrOff   = ""
	weekNumStrLeft  = "left"
	weekNumStrRight = "right"
)

// parseWeekNumPos converts a config string to the typed WeekNumPos enum.
func parseWeekNumPos(s string) WeekNumPos {
	switch s {
	case "left":
		return WeekNumLeft
	case "right":
		return WeekNumRight
	default:
		return WeekNumOff
	}
}

// ThemeColors defines the color scheme for calendar UI elements.
type ThemeColors struct {
	Cursor     string `yaml:"cursor"`
	Today      string `yaml:"today"`
	Title      string `yaml:"title"`
	WeekNumber string `yaml:"week_number"`
	DayHeader  string `yaml:"day_header"`
	HelpBar    string `yaml:"help_bar"`
	Highlight  string `yaml:"highlight"`
	Range      string `yaml:"range"`
}

// Config holds user preferences for the calendar display.
type Config struct {
	ShowWeekNumbers   string      `yaml:"show_week_numbers"`
	WeekNumbering     string      `yaml:"week_numbering"`
	WeekStartDay      int         `yaml:"week_start_day"`
	FiscalYearStart   int         `yaml:"fiscal_year_start"`
	ShowFiscalQuarter bool        `yaml:"show_fiscal_quarter"`
	ShowQuarterBar    bool        `yaml:"show_quarter_bar"`
	Julian            bool        `yaml:"julian"`
	Theme             string      `yaml:"theme"`
	Colors            ThemeColors `yaml:"colors"`
	HighlightSource   string      `yaml:"highlight_source"`
}

// DefaultConfig returns a Config with sensible defaults (US week numbering, Sunday start, default theme).
func DefaultConfig() Config {
	return Config{
		ShowWeekNumbers: "",
		WeekNumbering:   WeekNumberingUS,
		WeekStartDay:    0,
		FiscalYearStart: 1,
		Theme:           "default",
	}
}

// Normalize validates config values, resetting invalid ones to defaults, and returns any warnings.
// ISO week numbering silently forces WeekStartDay to Monday as a constraint (not a validation error).
func (c *Config) Normalize() []string {
	var warnings []string
	if c.WeekNumbering != WeekNumberingUS && c.WeekNumbering != WeekNumberingISO {
		warnings = append(warnings, "invalid config value for \"week_numbering\", using default")
		c.WeekNumbering = WeekNumberingUS
	}
	// Normalize show_week_numbers: true → "left" (standard cal -w style), false → ""
	switch c.ShowWeekNumbers {
	case "true":
		c.ShowWeekNumbers = weekNumStrLeft
	case "false", "":
		c.ShowWeekNumbers = weekNumStrOff
	case weekNumStrLeft, weekNumStrRight:
		// valid
	default:
		warnings = append(warnings, "invalid config value for \"show_week_numbers\", using default")
		c.ShowWeekNumbers = weekNumStrOff
	}
	if c.WeekStartDay != 0 && c.WeekStartDay != 1 {
		warnings = append(warnings, "invalid config value for \"week_start_day\", using default")
		c.WeekStartDay = 0
	}
	if c.WeekNumbering == WeekNumberingISO {
		c.WeekStartDay = 1
	}
	if c.FiscalYearStart < 1 || c.FiscalYearStart > 12 {
		if c.FiscalYearStart != 0 { // 0 = unset, treat as default
			warnings = append(warnings, "invalid config value for \"fiscal_year_start\", using default (1=January)")
		}
		c.FiscalYearStart = 1
	}
	if _, ok := themePresets[c.Theme]; !ok {
		warnings = append(warnings, "invalid config value for \"theme\", using default")
		c.Theme = "default"
	}
	return warnings
}

func mergeColor(base, override string) string {
	if override != "" {
		return override
	}
	return base
}

// ResolvedColors merges theme preset colors with any user-specified overrides.
func (c Config) ResolvedColors() ThemeColors {
	base := themePresets[c.Theme]
	return ThemeColors{
		Cursor:     mergeColor(base.Cursor, c.Colors.Cursor),
		Today:      mergeColor(base.Today, c.Colors.Today),
		Title:      mergeColor(base.Title, c.Colors.Title),
		WeekNumber: mergeColor(base.WeekNumber, c.Colors.WeekNumber),
		DayHeader:  mergeColor(base.DayHeader, c.Colors.DayHeader),
		HelpBar:    mergeColor(base.HelpBar, c.Colors.HelpBar),
		Highlight:  mergeColor(base.Highlight, c.Colors.Highlight),
		Range:      mergeColor(base.Range, c.Colors.Range),
	}
}

var themePresets = map[string]ThemeColors{
	"default": {},
	"catppuccin-mocha": {
		Cursor:     "#f5c2e7",
		Today:      "#a6e3a1",
		Title:      "#89b4fa",
		WeekNumber: "#6c7086",
		DayHeader:  "#94e2d5",
		HelpBar:    "#6c7086",
		Highlight:  "#f9e2af",
		Range:      "#a6e3a1",
	},
	"dracula": {
		Cursor:     "#ff79c6",
		Today:      "#50fa7b",
		Title:      "#bd93f9",
		WeekNumber: "#6272a4",
		DayHeader:  "#8be9fd",
		HelpBar:    "#6272a4",
		Highlight:  "#f1fa8c",
		Range:      "#50fa7b",
	},
	"nord": {
		Cursor:     "#88c0d0",
		Today:      "#a3be8c",
		Title:      "#81a1c1",
		WeekNumber: "#4c566a",
		DayHeader:  "#8fbcbb",
		HelpBar:    "#4c566a",
		Highlight:  "#ebcb8b",
		Range:      "#a3be8c",
	},
}

// knownConfigKeys lists all valid top-level config keys. Used to detect typos.
var knownConfigKeys = map[string]bool{
	"show_week_numbers":   true,
	"week_numbering":      true,
	"week_start_day":      true,
	"fiscal_year_start":   true,
	"show_fiscal_quarter": true,
	"show_quarter_bar":    true,
	"julian":              true,
	"theme":               true,
	"colors":              true,
	"highlight_source":    true,
}

// LoadConfig reads the user's config from the XDG config path, creating a default file if none exists.
func LoadConfig() (Config, []string) {
	path, err := configPath()
	if err != nil {
		return DefaultConfig(), []string{err.Error()}
	}
	return loadConfigFromPath(path)
}

func loadConfigFromPath(path string) (Config, []string) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return cfg, []string{fmt.Sprintf("could not read config: %v", err)}
		}
		if wErr := writeDefaultConfig(path); wErr != nil {
			return cfg, []string{wErr.Error()}
		}
		return cfg, nil
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), []string{"invalid config file, using defaults"}
	}

	// Check for unknown top-level keys (typo detection).
	var warnings []string
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err == nil {
		for key := range raw {
			if !knownConfigKeys[key] {
				warnings = append(warnings, fmt.Sprintf("unknown config key %q", key))
			}
		}
	}

	warnings = append(warnings, cfg.Normalize()...)
	return cfg, warnings
}

func configPath() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not determine home directory: %w", err)
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "wen", "config.yaml"), nil
}

func writeDefaultConfig(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}
	content := `# Week numbers
show_week_numbers: false  # false, true/"left" (standard), or "right"
week_numbering: us    # "us" or "iso"
week_start_day: 0     # 0=Sunday, 1=Monday

# Fiscal year start month (1-12, default: 1=January)
# Affects "end of quarter", "beginning of quarter", etc.
# Example: fiscal_year_start: 10  # October (common US federal/corporate)
# fiscal_year_start: 1

# Show fiscal quarter in calendar title (e.g., "March 2026 · Q2 FY26")
# Requires fiscal_year_start > 1 to take effect.
# show_fiscal_quarter: false

# Show quarter progress bar below the calendar grid
# show_quarter_bar: false

# Julian day-of-year numbering (shows day 1-366 instead of day of month)
# julian: false

# Theme (built-in: "default", "catppuccin-mocha", "dracula", "nord")
theme: default

# Override individual colors (hex values, override theme):
# colors:
#   cursor: "#f5c2e7"
#   today: "#a6e3a1"
#   title: "#89b4fa"
#   week_number: "#6c7086"
#   day_header: "#94e2d5"
#   help_bar: "#6c7086"
#   highlight: "#f9e2af"
#   range: "#a6e3a1"

# Highlighted dates source (JSON array of yyyy-mm-dd strings):
# highlight_source: ~/.local/share/pike/due.json
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("could not write default config: %w", err)
	}
	return nil
}
