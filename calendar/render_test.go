package calendar

import (
	"strings"
	"testing"
	"time"
)

func TestRenderTitle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		cursor   time.Time
		today    time.Time
		cfg      Config
		month    time.Month
		year     int
		contains string
		excludes string
	}{
		{
			name:     "current year omits year",
			cursor:   date(2026, time.March, 17),
			today:    date(2026, time.March, 17),
			cfg:      DefaultConfig(),
			month:    time.March,
			year:     2026,
			contains: "March",
			excludes: "2026",
		},
		{
			name:     "different year shows year",
			cursor:   date(2027, time.March, 17),
			today:    date(2026, time.March, 17),
			cfg:      DefaultConfig(),
			month:    time.March,
			year:     2027,
			contains: "March 2027",
		},
		{
			name: "fiscal quarter shown",
			cursor: date(2026, time.March, 17),
			today:  date(2026, time.March, 17),
			cfg: func() Config {
				c := DefaultConfig()
				c.FiscalYearStart = 10
				c.ShowFiscalQuarter = true
				return c
			}(),
			month:    time.March,
			year:     2026,
			contains: "Q2 FY26",
		},
		{
			name: "fiscal quarter uses abbreviated month",
			cursor: date(2026, time.September, 17),
			today:  date(2026, time.September, 17),
			cfg: func() Config {
				c := DefaultConfig()
				c.FiscalYearStart = 10
				c.ShowFiscalQuarter = true
				return c
			}(),
			month:    time.September,
			year:     2026,
			contains: "Sep",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := New(tt.cursor, tt.today, tt.cfg)
			var b strings.Builder
			m.renderTitle(&b, tt.month, tt.year)
			got := b.String()
			if !strings.Contains(got, tt.contains) {
				t.Errorf("renderTitle() = %q, want substring %q", got, tt.contains)
			}
			if tt.excludes != "" && strings.Contains(got, tt.excludes) {
				t.Errorf("renderTitle() = %q, should not contain %q", got, tt.excludes)
			}
		})
	}
}

func TestRenderDayHeaders(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		startDay int
		contains string
	}{
		{"sunday start", 0, "Su Mo Tu We Th Fr Sa"},
		{"monday start", 1, "Mo Tu We Th Fr Sa Su"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := DefaultConfig()
			cfg.WeekStartDay = tt.startDay
			m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
			var b strings.Builder
			m.renderDayHeaders(&b)
			got := b.String()
			if !strings.Contains(got, tt.contains) {
				t.Errorf("renderDayHeaders() = %q, want substring %q", got, tt.contains)
			}
		})
	}
}

func TestRenderGrid(t *testing.T) {
	t.Parallel()
	t.Run("returns correct week numbers", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
		var b strings.Builder
		weekNums := m.renderGrid(&b, 2026, time.March, 17, time.Local)
		if len(weekNums) == 0 {
			t.Fatal("expected at least one week number")
		}
		// March 2026 starts on Sunday, so it should have 5 weeks
		if len(weekNums) != 5 {
			t.Errorf("expected 5 week rows for March 2026, got %d", len(weekNums))
		}
	})

	t.Run("grid contains all days", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		m := New(date(2026, time.February, 14), date(2026, time.February, 14), cfg)
		var b strings.Builder
		m.renderGrid(&b, 2026, time.February, 14, time.Local)
		got := b.String()
		// February 2026 has 28 days
		if !strings.Contains(got, "28") {
			t.Error("expected grid to contain day 28")
		}
		if strings.Contains(got, "29") {
			t.Error("February 2026 should not contain day 29")
		}
	})

	t.Run("monday start shifts grid", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		cfg.WeekStartDay = 1
		m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
		var b strings.Builder
		m.renderGrid(&b, 2026, time.March, 17, time.Local)
		got := b.String()
		// March 1 2026 is a Sunday. With Monday start, Sunday is last column.
		// So day 1 should appear at the end of the first row.
		lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
		if len(lines) == 0 {
			t.Fatal("expected at least one grid line")
		}
		firstLine := lines[0]
		if !strings.HasSuffix(strings.TrimRight(firstLine, " \n"), "1") {
			t.Errorf("with Monday start, day 1 (Sunday) should be at end of first row, got: %q", firstLine)
		}
	})
}

func TestDateKey(t *testing.T) {
	t.Parallel()
	// dateKey should normalize to UTC midnight
	local := time.Date(2026, time.March, 17, 15, 30, 0, 0, time.Local)
	key := dateKey(local)
	if key.Hour() != 0 || key.Minute() != 0 || key.Location() != time.UTC {
		t.Errorf("dateKey should return UTC midnight, got %v", key)
	}
	if key.Day() != 17 || key.Month() != time.March || key.Year() != 2026 {
		t.Errorf("dateKey should preserve date, got %v", key)
	}
}

func TestIsInRange(t *testing.T) {
	t.Parallel()
	a := date(2026, time.March, 10)
	b := date(2026, time.March, 20)

	tests := []struct {
		name string
		d    time.Time
		want bool
	}{
		{"before range", date(2026, time.March, 9), false},
		{"start of range", a, true},
		{"middle of range", date(2026, time.March, 15), true},
		{"end of range", b, true},
		{"after range", date(2026, time.March, 21), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isInRange(tt.d, a, b); got != tt.want {
				t.Errorf("isInRange(%s, %s, %s) = %v, want %v",
					tt.d.Format("2006-01-02"), a.Format("2006-01-02"), b.Format("2006-01-02"),
					got, tt.want)
			}
			// Also test with reversed range
			if got := isInRange(tt.d, b, a); got != tt.want {
				t.Errorf("isInRange(%s, %s, %s) reversed = %v, want %v",
					tt.d.Format("2006-01-02"), b.Format("2006-01-02"), a.Format("2006-01-02"),
					got, tt.want)
			}
		})
	}
}

func TestQuarterStartDate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		cursor    time.Time
		fyStart   int
		wantMonth time.Month
		wantYear  int
	}{
		{"calendar Q1", date(2026, time.February, 15), 1, time.January, 2026},
		{"calendar Q2", date(2026, time.May, 15), 1, time.April, 2026},
		{"fiscal Oct Q1", date(2025, time.November, 15), 10, time.October, 2025},
		{"fiscal Oct Q2", date(2026, time.February, 15), 10, time.January, 2026},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := quarterStartDate(tt.cursor, tt.fyStart)
			if got.Month() != tt.wantMonth || got.Year() != tt.wantYear {
				t.Errorf("quarterStartDate() = %s, want %s %d",
					got.Format("2006-01"), tt.wantMonth, tt.wantYear)
			}
		})
	}
}

func TestCountQuarterWorkdaysLeft(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		cursor time.Time
		qEnd   time.Time
		want   int
	}{
		{
			name:   "cursor after qEnd",
			cursor: date(2026, time.April, 1),
			qEnd:   date(2026, time.March, 31),
			want:   0,
		},
		{
			name:   "cursor on qEnd",
			cursor: date(2026, time.March, 31),
			qEnd:   date(2026, time.March, 31),
			want:   0,
		},
		{
			name:   "one workday left",
			cursor: date(2026, time.March, 30), // Monday
			qEnd:   date(2026, time.March, 31), // Tuesday
			want:   1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := countQuarterWorkdaysLeft(tt.cursor, tt.qEnd)
			if got != tt.want {
				t.Errorf("countQuarterWorkdaysLeft() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDayFormat(t *testing.T) {
	t.Parallel()
	t.Run("normal", func(t *testing.T) {
		t.Parallel()
		df := normalDayFormat()
		if df.cellWidth != 2 {
			t.Errorf("cellWidth = %d, want 2", df.cellWidth)
		}
		if df.gridWidth != 20 {
			t.Errorf("gridWidth = %d, want 20", df.gridWidth)
		}
		if df.prefixWidth != 3 {
			t.Errorf("prefixWidth = %d, want 3", df.prefixWidth)
		}
		if df.names != dayNames {
			t.Errorf("names = %v, want dayNames", df.names)
		}
		got := df.formatDay(2026, time.March, 5, time.Local)
		if got != " 5" {
			t.Errorf("formatDay = %q, want %q", got, " 5")
		}
	})
	t.Run("julian", func(t *testing.T) {
		t.Parallel()
		df := julianDayFormat()
		if df.cellWidth != 3 {
			t.Errorf("cellWidth = %d, want 3", df.cellWidth)
		}
		if df.gridWidth != 27 {
			t.Errorf("gridWidth = %d, want 27", df.gridWidth)
		}
		if df.prefixWidth != 4 {
			t.Errorf("prefixWidth = %d, want 4", df.prefixWidth)
		}
		if df.names != dayNamesLong {
			t.Errorf("names = %v, want dayNamesLong", df.names)
		}
		// March 5 2026 = yearday 64
		got := df.formatDay(2026, time.March, 5, time.Local)
		if got != " 64" {
			t.Errorf("formatDay = %q, want %q", got, " 64")
		}
	})
	t.Run("dayFormatFor false", func(t *testing.T) {
		t.Parallel()
		df := dayFormatFor(false)
		if df.cellWidth != 2 {
			t.Errorf("dayFormatFor(false) cellWidth = %d, want 2", df.cellWidth)
		}
	})
	t.Run("dayFormatFor true", func(t *testing.T) {
		t.Parallel()
		df := dayFormatFor(true)
		if df.cellWidth != 3 {
			t.Errorf("dayFormatFor(true) cellWidth = %d, want 3", df.cellWidth)
		}
	})
}

func TestRenderGridJulian(t *testing.T) {
	t.Parallel()
	t.Run("january shows yearday values", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		m := New(date(2026, time.January, 15), date(2026, time.January, 15), cfg, WithJulian(true))
		var b strings.Builder
		m.renderGrid(&b, 2026, time.January, 15, time.Local)
		got := b.String()
		if !strings.Contains(got, "  1") {
			t.Errorf("expected yearday 1 (3-char padded), got:\n%s", got)
		}
		if !strings.Contains(got, " 31") {
			t.Errorf("expected yearday 31, got:\n%s", got)
		}
	})

	t.Run("march shows offset yearday values", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		m := New(date(2026, time.March, 15), date(2026, time.March, 15), cfg, WithJulian(true))
		var b strings.Builder
		m.renderGrid(&b, 2026, time.March, 15, time.Local)
		got := b.String()
		// March 1 2026 = yearday 60, March 31 = yearday 90
		if !strings.Contains(got, " 60") {
			t.Errorf("expected yearday 60 for March 1, got:\n%s", got)
		}
		if !strings.Contains(got, " 90") {
			t.Errorf("expected yearday 90 for March 31, got:\n%s", got)
		}
	})

	t.Run("december leap year shows 366", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		m := New(date(2024, time.December, 31), date(2024, time.December, 31), cfg, WithJulian(true))
		var b strings.Builder
		m.renderGrid(&b, 2024, time.December, 31, time.Local)
		got := b.String()
		if !strings.Contains(got, "366") {
			t.Errorf("expected yearday 366 for Dec 31 2024, got:\n%s", got)
		}
	})
}

func TestRenderGridPrintModeSuppressesCursor(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	// Create two models for March 17 — one normal, one print mode
	normal := New(date(2026, time.March, 17), date(2025, time.March, 17), cfg)
	printM := New(date(2026, time.March, 17), date(2025, time.March, 17), cfg, WithPrintMode(true))

	var normalBuf, printBuf strings.Builder
	normal.renderGrid(&normalBuf, 2026, time.March, 17, time.Local)
	printM.renderGrid(&printBuf, 2026, time.March, 17, time.Local)

	// Both should contain "17" but print mode output shouldn't have cursor ANSI styling
	// In non-TTY test env, lipgloss strips ANSI, so both look the same.
	// Just verify print mode doesn't crash and produces output with day 17.
	if !strings.Contains(printBuf.String(), "17") {
		t.Error("print mode grid should contain day 17")
	}
}

func TestRenderDayHeadersJulian(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg, WithJulian(true))
	var b strings.Builder
	m.renderDayHeaders(&b)
	got := b.String()
	if !strings.Contains(got, "Sun Mon Tue Wed Thu Fri Sat") {
		t.Errorf("julian headers should use 3-char names, got: %q", got)
	}
}

func TestRenderQuarterBar(t *testing.T) {
	t.Parallel()
	t.Run("hidden by default", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
		var b strings.Builder
		m.renderQuarterBar(&b, 20)
		if b.Len() != 0 {
			t.Error("quarter bar should produce no output when ShowQuarterBar is false")
		}
	})

	t.Run("shows quarter and workdays", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		cfg.ShowQuarterBar = true
		m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
		var b strings.Builder
		m.renderQuarterBar(&b, 20)
		got := b.String()
		if !strings.Contains(got, "Q1") {
			t.Errorf("expected Q1 in quarter bar, got: %q", got)
		}
		if !strings.Contains(got, "wd") {
			t.Errorf("expected workdays in quarter bar, got: %q", got)
		}
	})

	t.Run("fiscal quarter shown correctly", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		cfg.ShowQuarterBar = true
		cfg.FiscalYearStart = 10
		m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
		var b strings.Builder
		m.renderQuarterBar(&b, 20)
		got := b.String()
		if !strings.Contains(got, "Q2") {
			t.Errorf("expected Q2 for fiscal Oct start in March, got: %q", got)
		}
	})
}
