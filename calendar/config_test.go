package calendar

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ShowWeekNumbers != "" {
		t.Errorf("expected ShowWeekNumbers empty by default, got %q", cfg.ShowWeekNumbers)
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
	if cfg.ShowWeekNumbers != "left" {
		t.Errorf("expected ShowWeekNumbers 'left' (from true), got %q", cfg.ShowWeekNumbers)
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

func TestLoadConfigWithPadding(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("padding_top: 1\npadding_right: 2\npadding_bottom: 3\npadding_left: 4\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, warnings := loadConfigFromPath(path)
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
	if cfg.PaddingTop != 1 || cfg.PaddingRight != 2 || cfg.PaddingBottom != 3 || cfg.PaddingLeft != 4 {
		t.Errorf("padding mismatch: got top=%d right=%d bottom=%d left=%d",
			cfg.PaddingTop, cfg.PaddingRight, cfg.PaddingBottom, cfg.PaddingLeft)
	}
}

func TestDefaultConfigPaddingZero(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.PaddingTop != 0 || cfg.PaddingRight != 0 || cfg.PaddingBottom != 0 || cfg.PaddingLeft != 0 {
		t.Error("expected all padding values to be 0 by default")
	}
}

func TestNormalizeClampsNegativePadding(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PaddingTop = -5
	cfg.PaddingLeft = -1
	warnings := cfg.Normalize()
	if cfg.PaddingTop != 0 {
		t.Errorf("expected PaddingTop clamped to 0, got %d", cfg.PaddingTop)
	}
	if cfg.PaddingLeft != 0 {
		t.Errorf("expected PaddingLeft clamped to 0, got %d", cfg.PaddingLeft)
	}
	// Untouched fields should remain at zero default.
	if cfg.PaddingRight != 0 || cfg.PaddingBottom != 0 {
		t.Errorf("expected untouched padding fields to remain 0, got right=%d bottom=%d",
			cfg.PaddingRight, cfg.PaddingBottom)
	}
	if len(warnings) != 2 {
		t.Errorf("expected 2 warnings for 2 negative padding values, got %d", len(warnings))
	}
}

func TestNormalizeClampsExcessivePadding(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PaddingRight = 50
	cfg.PaddingBottom = 100
	warnings := cfg.Normalize()
	if cfg.PaddingRight != MaxPadding {
		t.Errorf("expected PaddingRight clamped to %d, got %d", MaxPadding, cfg.PaddingRight)
	}
	if cfg.PaddingBottom != MaxPadding {
		t.Errorf("expected PaddingBottom clamped to %d, got %d", MaxPadding, cfg.PaddingBottom)
	}
	if len(warnings) != 2 {
		t.Errorf("expected 2 warnings for 2 excessive padding values, got %d", len(warnings))
	}
}

func TestNormalizeValidPaddingNoWarnings(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PaddingTop = 2
	cfg.PaddingLeft = 5
	warnings := cfg.Normalize()
	if cfg.PaddingTop != 2 || cfg.PaddingLeft != 5 {
		t.Error("valid padding values should not be changed")
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for valid padding, got %v", warnings)
	}
}

func TestNormalizeFiscalYearStart(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FiscalYearStart = 10
	warnings := cfg.Normalize()
	if cfg.FiscalYearStart != 10 {
		t.Errorf("expected FiscalYearStart 10, got %d", cfg.FiscalYearStart)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for valid fiscal_year_start, got %v", warnings)
	}

	// Invalid value
	cfg.FiscalYearStart = 13
	warnings = cfg.Normalize()
	if cfg.FiscalYearStart != 1 {
		t.Errorf("expected FiscalYearStart reset to 1, got %d", cfg.FiscalYearStart)
	}
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning for invalid fiscal_year_start, got %d", len(warnings))
	}
}

func TestLoadConfigFiscalYearStart(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("fiscal_year_start: 10\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, warnings := loadConfigFromPath(path)
	if cfg.FiscalYearStart != 10 {
		t.Errorf("expected FiscalYearStart 10, got %d", cfg.FiscalYearStart)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/subdir/config.yaml"
	cfg, warnings := loadConfigFromPath(path)
	if cfg.Theme != "default" {
		t.Errorf("expected default theme on missing file, got %q", cfg.Theme)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for missing file in writable dir, got %v", warnings)
	}
}

func TestLoadConfigUnwritableDir(t *testing.T) {
	// Use /proc as a directory we definitely cannot create subdirs in,
	// and a path that will not already have a config file.
	cfg, warnings := loadConfigFromPath("/proc/nonexistent-wen-test/config.yaml")
	if cfg.Theme != "default" {
		t.Errorf("expected default theme, got %q", cfg.Theme)
	}
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning for unwritable path, got %d: %v", len(warnings), warnings)
	}
}
