package calendar

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func FuzzConfigNormalize(f *testing.F) {
	seeds := []string{
		"",
		"theme: default",
		"theme: catppuccin-mocha\nshow_week_numbers: left",
		"week_numbering: iso\nweek_start_day: 1",
		"{invalid yaml",
		"fiscal_year_start: 999",
		"show_week_numbers: maybe",
		"\x00\xff\xfe",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(_ *testing.T, input string) {
		var cfg Config
		_ = yaml.Unmarshal([]byte(input), &cfg)
		cfg.Normalize()
	})
}

func FuzzLoadHighlightedDates(f *testing.F) {
	seeds := []string{
		`["2026-01-01"]`,
		`["2026-01-01","2026-12-31"]`,
		`[]`,
		`["invalid-date"]`,
		`not json`,
		`{"key":"value"}`,
		"",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.json")
		if err := os.WriteFile(path, []byte(input), 0644); err != nil {
			return
		}
		_, _ = LoadHighlightedDates(path)
	})
}
