package calendar

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	WeekNumberingUS  = "us"
	WeekNumberingISO = "iso"
)

// ThemeColors defines the color scheme for calendar UI elements.
type ThemeColors struct {
	Cursor     string `yaml:"cursor"`
	Today      string `yaml:"today"`
	Title      string `yaml:"title"`
	WeekNumber string `yaml:"week_number"`
	DayHeader  string `yaml:"day_header"`
	HelpBar    string `yaml:"help_bar"`
}

// Config holds user preferences for the calendar display.
type Config struct {
	ShowWeekNumbers bool        `yaml:"show_week_numbers"`
	WeekNumbering   string      `yaml:"week_numbering"`
	WeekStartDay    int         `yaml:"week_start_day"`
	Theme           string      `yaml:"theme"`
	Colors          ThemeColors `yaml:"colors"`
}

// DefaultConfig returns a Config with sensible defaults (US week numbering, Sunday start, default theme).
func DefaultConfig() Config {
	return Config{
		ShowWeekNumbers: false,
		WeekNumbering:   WeekNumberingUS,
		WeekStartDay:    0,
		Theme:           "default",
	}
}

func (c *Config) Normalize() []string {
	var warnings []string
	if c.WeekNumbering != WeekNumberingUS && c.WeekNumbering != WeekNumberingISO {
		warnings = append(warnings, "invalid config value for \"week_numbering\", using default")
		c.WeekNumbering = WeekNumberingUS
	}
	if c.WeekStartDay != 0 && c.WeekStartDay != 1 {
		warnings = append(warnings, "invalid config value for \"week_start_day\", using default")
		c.WeekStartDay = 0
	}
	if c.WeekNumbering == WeekNumberingISO {
		c.WeekStartDay = 1
	}
	if _, ok := themePresets[c.Theme]; !ok {
		warnings = append(warnings, "invalid config value for \"theme\", using default")
		c.Theme = "default"
	}
	return warnings
}

func (c Config) ResolvedColors() ThemeColors {
	base := themePresets[c.Theme]
	if c.Colors.Cursor != "" {
		base.Cursor = c.Colors.Cursor
	}
	if c.Colors.Today != "" {
		base.Today = c.Colors.Today
	}
	if c.Colors.Title != "" {
		base.Title = c.Colors.Title
	}
	if c.Colors.WeekNumber != "" {
		base.WeekNumber = c.Colors.WeekNumber
	}
	if c.Colors.DayHeader != "" {
		base.DayHeader = c.Colors.DayHeader
	}
	if c.Colors.HelpBar != "" {
		base.HelpBar = c.Colors.HelpBar
	}
	return base
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
	},
	"dracula": {
		Cursor:     "#ff79c6",
		Today:      "#50fa7b",
		Title:      "#bd93f9",
		WeekNumber: "#6272a4",
		DayHeader:  "#8be9fd",
		HelpBar:    "#6272a4",
	},
	"nord": {
		Cursor:     "#88c0d0",
		Today:      "#a3be8c",
		Title:      "#81a1c1",
		WeekNumber: "#4c566a",
		DayHeader:  "#8fbcbb",
		HelpBar:    "#4c566a",
	},
}

// LoadConfig reads the user's config from the XDG config path, creating a default file if none exists.
func LoadConfig() (Config, []string) {
	path := configPath()
	return loadConfigFromPath(path)
}

func loadConfigFromPath(path string) (Config, []string) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			writeDefaultConfig(path)
		}
		return cfg, nil
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), []string{"invalid config file, using defaults"}
	}
	warnings := cfg.Normalize()
	return cfg, warnings
}

func configPath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "wen", "config.yaml")
}

func writeDefaultConfig(path string) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not create config directory: %v\n", err)
		return
	}
	content := `# Week numbers
show_week_numbers: false
week_numbering: us    # "us" or "iso"
week_start_day: 0     # 0=Sunday, 1=Monday

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
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not write default config: %v\n", err)
	}
}
