package calendar

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "update golden files")

func TestGoldenView(t *testing.T) {
	t.Parallel()
	// Use a fixed "today" different from cursor year so year always appears in title.
	today := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name   string
		cursor time.Time
		cfg    Config
		opts   []Option
	}{
		{
			name:   "single_month",
			cursor: time.Date(2026, time.March, 17, 0, 0, 0, 0, time.Local),
			cfg:    DefaultConfig(),
		},
		{
			name:   "monday_start",
			cursor: time.Date(2026, time.March, 17, 0, 0, 0, 0, time.Local),
			cfg: func() Config {
				c := DefaultConfig()
				c.WeekStartDay = 1
				return c
			}(),
		},
		{
			name:   "week_numbers_left",
			cursor: time.Date(2026, time.March, 17, 0, 0, 0, 0, time.Local),
			cfg: func() Config {
				c := DefaultConfig()
				c.ShowWeekNumbers = "left"
				return c
			}(),
		},
		{
			name:   "julian",
			cursor: time.Date(2026, time.January, 15, 0, 0, 0, 0, time.Local),
			cfg:    DefaultConfig(),
			opts:   []Option{WithJulian(true)},
		},
		{
			name:   "multi_month_3",
			cursor: time.Date(2026, time.March, 17, 0, 0, 0, 0, time.Local),
			cfg:    DefaultConfig(),
			opts:   []Option{WithMonths(3)},
		},
		{
			name:   "print_mode",
			cursor: time.Date(2026, time.March, 17, 0, 0, 0, 0, time.Local),
			cfg:    DefaultConfig(),
			opts:   []Option{WithPrintMode(true)},
		},
		{
			name:   "fiscal_quarter",
			cursor: time.Date(2026, time.March, 17, 0, 0, 0, 0, time.Local),
			cfg: func() Config {
				c := DefaultConfig()
				c.FiscalYearStart = 10
				c.ShowFiscalQuarter = true
				return c
			}(),
		},
		{
			name:   "february_leap_year",
			cursor: time.Date(2024, time.February, 15, 0, 0, 0, 0, time.Local),
			cfg:    DefaultConfig(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			opts := append([]Option{WithPrintMode(true)}, tt.opts...)
			m := New(tt.cursor, today, tt.cfg, opts...)
			got := m.View()

			// Strip ANSI escape sequences for deterministic comparison.
			got = stripANSI(got)

			golden := filepath.Join("testdata", tt.name+".golden")
			if *update {
				if err := os.WriteFile(golden, []byte(got), 0644); err != nil {
					t.Fatalf("failed to update golden file: %v", err)
				}
				return
			}

			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("failed to read golden file (run with -update to create): %v", err)
			}
			if got != string(want) {
				t.Errorf("output does not match golden file %s\n--- got ---\n%s\n--- want ---\n%s", golden, got, string(want))
			}
		})
	}
}

// stripANSI removes ANSI escape sequences from s for deterministic golden comparisons.
func stripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Skip until we hit a letter (the terminator)
			j := i + 2
			for j < len(s) && (s[j] < 'A' || s[j] > 'Z') && (s[j] < 'a' || s[j] > 'z') {
				j++
			}
			if j < len(s) {
				j++ // skip the terminator
			}
			i = j
		} else {
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String()
}
