package calendar

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ShowWeekNumbers {
		t.Error("expected ShowWeekNumbers false by default")
	}
	if cfg.WeekNumbering != "us" {
		t.Errorf("expected week_numbering 'us', got %q", cfg.WeekNumbering)
	}
	if cfg.WeekStartDay != 0 {
		t.Errorf("expected week_start_day 0, got %d", cfg.WeekStartDay)
	}
	if cfg.Theme != "default" {
		t.Errorf("expected theme 'default', got %q", cfg.Theme)
	}
}

func TestResolveThemeDefault(t *testing.T) {
	cfg := DefaultConfig()
	colors := cfg.ResolvedColors()
	if colors.Title != "" {
		t.Error("default theme should have empty title color (terminal default)")
	}
}

func TestResolveThemeCatppuccin(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Theme = "catppuccin-mocha"
	colors := cfg.ResolvedColors()
	if colors.Title != "#89b4fa" {
		t.Errorf("expected catppuccin title color, got %q", colors.Title)
	}
}

func TestCustomColorsOverrideTheme(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Theme = "catppuccin-mocha"
	cfg.Colors.Title = "#ff0000"
	colors := cfg.ResolvedColors()
	if colors.Title != "#ff0000" {
		t.Errorf("expected custom title color, got %q", colors.Title)
	}
}

func TestISOForcesMonday(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WeekNumbering = "iso"
	cfg.WeekStartDay = 0
	warnings := cfg.Normalize()
	if cfg.WeekStartDay != 1 {
		t.Errorf("ISO should force week_start_day to 1, got %d", cfg.WeekStartDay)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for valid ISO config, got %v", warnings)
	}
}

func TestInvalidWeekStartDayDefaults(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WeekStartDay = 5
	warnings := cfg.Normalize()
	if cfg.WeekStartDay != 0 {
		t.Errorf("invalid week_start_day should default to 0, got %d", cfg.WeekStartDay)
	}
	if len(warnings) == 0 {
		t.Error("expected a warning for invalid week_start_day, got none")
	}
}

func TestNormalizeWarnsInvalidWeekNumbering(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WeekNumbering = "bogus"
	warnings := cfg.Normalize()
	if cfg.WeekNumbering != WeekNumberingUS {
		t.Errorf("expected week_numbering reset to %q, got %q", WeekNumberingUS, cfg.WeekNumbering)
	}
	if len(warnings) == 0 {
		t.Error("expected a warning for invalid week_numbering, got none")
	}
}

func TestNormalizeWarnsInvalidTheme(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Theme = "nonexistent-theme"
	warnings := cfg.Normalize()
	if cfg.Theme != "default" {
		t.Errorf("expected theme reset to 'default', got %q", cfg.Theme)
	}
	if len(warnings) == 0 {
		t.Error("expected a warning for invalid theme, got none")
	}
}

func TestLoadConfigFromYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("theme: dracula\nshow_week_numbers: true\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, warnings := loadConfigFromPath(path)
	if cfg.Theme != "dracula" {
		t.Errorf("expected theme 'dracula', got %q", cfg.Theme)
	}
	if !cfg.ShowWeekNumbers {
		t.Error("expected ShowWeekNumbers true")
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for valid config, got %v", warnings)
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("{{invalid yaml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, warnings := loadConfigFromPath(path)
	if cfg.Theme != "default" {
		t.Errorf("expected default theme on invalid YAML, got %q", cfg.Theme)
	}
	if len(warnings) == 0 {
		t.Error("expected a warning for invalid YAML, got none")
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	cfg, warnings := loadConfigFromPath("/nonexistent/config.yaml")
	if cfg.Theme != "default" {
		t.Errorf("expected default theme on missing file, got %q", cfg.Theme)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for missing file, got %v", warnings)
	}
}
