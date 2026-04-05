package calendar

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	cfg := DefaultConfig()
	colors := cfg.ResolvedColors()
	if colors.Title != "" {
		t.Error("default theme should have empty title color (terminal default)")
	}
}

func TestResolveThemeCatppuccin(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.Theme = "catppuccin-mocha"
	colors := cfg.ResolvedColors()
	if colors.Title != "#89b4fa" {
		t.Errorf("expected catppuccin title color, got %q", colors.Title)
	}
}

func TestCustomColorsOverrideTheme(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.Theme = "catppuccin-mocha"
	cfg.Colors.Title = "#ff0000"
	colors := cfg.ResolvedColors()
	if colors.Title != "#ff0000" {
		t.Errorf("expected custom title color, got %q", colors.Title)
	}
}

func TestISOForcesMonday(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

func TestNormalizeFiscalYearStart(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

func TestJulianConfigField(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	if cfg.Julian {
		t.Error("expected Julian to default to false")
	}

	yamlData := []byte("julian: true\n")
	var loaded Config
	if err := yaml.Unmarshal(yamlData, &loaded); err != nil {
		t.Fatal(err)
	}
	if !loaded.Julian {
		t.Error("expected Julian to be true after loading from YAML")
	}
}

func TestLoadConfigUnwritableDir(t *testing.T) {
	t.Parallel()
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

func TestLoadConfigUnknownKeyWarns(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("theme: default\nshwo_week_numbers: true\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, warnings := loadConfigFromPath(path)
	if cfg.Theme != "default" {
		t.Errorf("expected theme 'default', got %q", cfg.Theme)
	}
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "shwo_week_numbers") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning about unknown key 'shwo_week_numbers', got %v", warnings)
	}
}

func TestLoadConfigNoWarningForValidKeys(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("theme: dracula\nshow_week_numbers: left\nweek_numbering: us\nweek_start_day: 0\nfiscal_year_start: 1\njulian: false\nshow_fiscal_quarter: false\nshow_quarter_bar: false\nhighlight_source: /tmp/test.json\ncolors:\n  cursor: \"#ff0000\"\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	_, warnings := loadConfigFromPath(path)
	for _, w := range warnings {
		if strings.Contains(w, "unknown") {
			t.Errorf("unexpected unknown-key warning: %s", w)
		}
	}
}

func TestConfigPath_XDGSet(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-xdg")
	got, err := configPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/tmp/test-xdg/wen/config.yaml"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestConfigPath_XDGUnset(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	got, err := configPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".config", "wen", "config.yaml")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLoadConfig_XDGOverride(t *testing.T) {
	dir := t.TempDir()
	wenDir := filepath.Join(dir, "wen")
	if err := os.MkdirAll(wenDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := []byte("theme: dracula\nfiscal_year_start: 10\n")
	if err := os.WriteFile(filepath.Join(wenDir, "config.yaml"), content, 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", dir)
	cfg, warnings := LoadConfig()
	if cfg.Theme != "dracula" {
		t.Errorf("expected theme 'dracula', got %q", cfg.Theme)
	}
	if cfg.FiscalYearStart != 10 {
		t.Errorf("expected FiscalYearStart 10, got %d", cfg.FiscalYearStart)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
}

func TestLoadConfig_MissingCreatesDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, warnings := LoadConfig()
	if cfg.Theme != "default" {
		t.Errorf("expected default theme, got %q", cfg.Theme)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}

	// Verify default config file was created on disk.
	created := filepath.Join(dir, "wen", "config.yaml")
	if _, err := os.Stat(created); err != nil {
		t.Errorf("expected default config file at %s, got error: %v", created, err)
	}
}
