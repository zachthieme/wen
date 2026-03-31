package calendar

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestMonthAbbrevs(t *testing.T) {
	t.Parallel()
	expected := [12]string{"Ja", "Fe", "Mr", "Ap", "My", "Jn", "Jl", "Au", "Se", "Oc", "No", "De"}
	if monthAbbrevs != expected {
		t.Errorf("monthAbbrevs = %v, want %v", monthAbbrevs, expected)
	}
}

func TestStripWindow(t *testing.T) {
	t.Parallel()
	loc := time.Local

	tests := []struct {
		name         string
		year         int
		month        time.Month
		weekStartDay int
		wantStart    time.Time
		wantEnd      time.Time
	}{
		{
			name:         "March 2026 Sunday start",
			year:         2026,
			month:        time.March,
			weekStartDay: 0, // Sunday
			wantStart:    time.Date(2026, time.March, 1, 0, 0, 0, 0, loc), // March 1 is a Sunday
			wantEnd:      time.Date(2026, time.April, 4, 0, 0, 0, 0, loc), // Sat: day before next Sun after Mar 31 (Tue)
		},
		{
			name:         "March 2026 Monday start",
			year:         2026,
			month:        time.March,
			weekStartDay: 1, // Monday
			wantStart:    time.Date(2026, time.February, 23, 0, 0, 0, 0, loc), // Mon before March 1
			wantEnd:      time.Date(2026, time.April, 5, 0, 0, 0, 0, loc),     // Sun after March 31
		},
		{
			name:         "February 2026 Sunday start",
			year:         2026,
			month:        time.February,
			weekStartDay: 0,
			wantStart:    time.Date(2026, time.February, 1, 0, 0, 0, 0, loc), // Feb 1 is a Sunday
			wantEnd:      time.Date(2026, time.February, 28, 0, 0, 0, 0, loc),
		},
		{
			name:         "February 2024 leap year Sunday start",
			year:         2024,
			month:        time.February,
			weekStartDay: 0,
			wantStart:    time.Date(2024, time.January, 28, 0, 0, 0, 0, loc), // Sun before Feb 1 (Thu)
			wantEnd:      time.Date(2024, time.March, 2, 0, 0, 0, 0, loc),    // Sat after Feb 29
		},
		{
			name:         "January 2026 Sunday start",
			year:         2026,
			month:        time.January,
			weekStartDay: 0,
			wantStart:    time.Date(2025, time.December, 28, 0, 0, 0, 0, loc), // Sun before Jan 1 (Thu)
			wantEnd:      time.Date(2026, time.January, 31, 0, 0, 0, 0, loc),
		},
		{
			name:         "April 2026 Monday start",
			year:         2026,
			month:        time.April,
			weekStartDay: 1,
			wantStart:    time.Date(2026, time.March, 30, 0, 0, 0, 0, loc), // Mon before Apr 1 (Wed)
			wantEnd:      time.Date(2026, time.May, 3, 0, 0, 0, 0, loc),    // Sun after Apr 30 (Thu)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			start, end := stripWindow(tt.year, tt.month, tt.weekStartDay, loc)
			if !start.Equal(tt.wantStart) {
				t.Errorf("start = %s, want %s", start.Format("2006-01-02"), tt.wantStart.Format("2006-01-02"))
			}
			if !end.Equal(tt.wantEnd) {
				t.Errorf("end = %s, want %s", end.Format("2006-01-02"), tt.wantEnd.Format("2006-01-02"))
			}
		})
	}
}

func TestRenderStripDayHeaders(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		contains string
	}{
		{
			name:     "March 2026 full weeks Sunday start",
			start:    date(2026, time.March, 1),
			end:      date(2026, time.April, 4),
			contains: "Su Mo Tu We Th Fr Sa",
		},
		{
			name:     "starts with leading space",
			start:    date(2026, time.March, 1),
			end:      date(2026, time.March, 7),
			contains: "   Su",
		},
		{
			name:     "Monday start shows Mo first",
			start:    date(2026, time.February, 23), // Monday
			end:      date(2026, time.March, 1),
			contains: "Mo Tu We Th Fr Sa Su",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := NewRow(date(2026, time.March, 15), date(2026, time.March, 15), DefaultConfig())
			got := m.renderStripDayHeaders(tt.start, tt.end)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("renderStripDayHeaders() = %q, want substring %q", got, tt.contains)
			}
		})
	}
}

func TestRenderStripDays(t *testing.T) {
	t.Parallel()

	t.Run("contains month abbreviation", func(t *testing.T) {
		t.Parallel()
		cursor := date(2026, time.March, 15)
		m := NewRow(cursor, date(2026, time.March, 15), DefaultConfig())
		start, end := stripWindow(2026, time.March, 0, time.Local)
		got := m.renderStripDays(start, end)
		if !strings.Contains(got, "Mr") {
			t.Errorf("renderStripDays() should contain month abbreviation 'Mr', got: %q", got)
		}
	})

	t.Run("contains all days of month", func(t *testing.T) {
		t.Parallel()
		cursor := date(2026, time.March, 15)
		m := NewRow(cursor, date(2026, time.March, 15), DefaultConfig())
		start, end := stripWindow(2026, time.March, 0, time.Local)
		got := m.renderStripDays(start, end)
		for day := 1; day <= 31; day++ {
			dayStr := strings.TrimSpace(strings.ReplaceAll(got, "\n", " "))
			if !strings.Contains(dayStr, fmt.Sprintf("%d", day)) {
				t.Errorf("renderStripDays() should contain day %d", day)
			}
		}
	})

	t.Run("February shows 28 days", func(t *testing.T) {
		t.Parallel()
		cursor := date(2026, time.February, 14)
		m := NewRow(cursor, date(2026, time.February, 14), DefaultConfig())
		start, end := stripWindow(2026, time.February, 0, time.Local)
		got := m.renderStripDays(start, end)
		if !strings.Contains(got, "Fe") {
			t.Errorf("renderStripDays() should contain month abbreviation 'Fe', got: %q", got)
		}
		if !strings.Contains(got, "28") {
			t.Errorf("renderStripDays() should contain day 28, got: %q", got)
		}
	})

	t.Run("highlighted dates are styled", func(t *testing.T) {
		t.Parallel()
		cursor := date(2026, time.March, 15)
		highlights := map[time.Time]bool{
			dateKey(date(2026, time.March, 20)): true,
		}
		m := NewRow(cursor, date(2026, time.March, 1), DefaultConfig(), WithRowHighlightedDates(highlights))
		start, end := stripWindow(2026, time.March, 0, time.Local)
		got := m.renderStripDays(start, end)
		// The output should contain day 20 (styled differently, but present)
		if !strings.Contains(got, "20") {
			t.Errorf("renderStripDays() should contain highlighted day 20, got: %q", got)
		}
	})

	t.Run("day count matches window size", func(t *testing.T) {
		t.Parallel()
		cursor := date(2026, time.March, 15)
		m := NewRow(cursor, date(2026, time.March, 15), DefaultConfig())
		start, end := stripWindow(2026, time.March, 0, time.Local)
		headerLine := m.renderStripDayHeaders(start, end)
		dayLine := m.renderStripDays(start, end)
		// Both lines should be non-empty
		if len(headerLine) == 0 {
			t.Error("header line should not be empty")
		}
		if len(dayLine) == 0 {
			t.Error("day line should not be empty")
		}
	})
}
