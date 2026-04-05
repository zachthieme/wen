package wen_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/zachthieme/wen"
)

// ref is Wednesday March 18, 2026 at 14:30 UTC — same as internal tests.
var ref = time.Date(2026, 3, 18, 14, 30, 0, 0, time.UTC)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestPublicParse(t *testing.T) {
	t.Parallel()
	// Parse uses time.Now() so we can only verify it doesn't error on valid input.
	_, err := wen.Parse("tomorrow")
	if err != nil {
		t.Fatalf("Parse(\"tomorrow\") returned error: %v", err)
	}
}

func TestPublicParseRelative(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  time.Time
	}{
		{"today", date(2026, 3, 18)},
		{"next friday", date(2026, 3, 27)},
		{"march 25 2026", date(2026, 3, 25)},
		{"3 days ago", date(2026, 3, 15)},
		{"in 2 weeks", date(2026, 4, 1)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := wen.ParseRelative(tt.input, ref)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublicParseMulti(t *testing.T) {
	t.Parallel()
	results, err := wen.ParseMulti("every friday in april", ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected multiple results, got none")
	}
	for _, r := range results {
		if r.Month() != time.April {
			t.Errorf("expected April, got %v", r.Month())
		}
		if r.Weekday() != time.Friday {
			t.Errorf("expected Friday, got %v", r.Weekday())
		}
	}
}

func TestPublicFiscalQuarter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		month      int
		year       int
		startMonth int
		wantQ      int
		wantFY     int
	}{
		{"calendar_q1", 2, 2026, 1, 1, 2026},
		{"oct_start_q1", 10, 2025, 10, 1, 2026},
		{"oct_start_q2", 1, 2026, 10, 2, 2026},
		{"invalid_start", 3, 2026, 0, 1, 2026}, // coerced to 1
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			q, fy := wen.FiscalQuarter(tt.month, tt.year, tt.startMonth)
			if q != tt.wantQ || fy != tt.wantFY {
				t.Errorf("FiscalQuarter(%d, %d, %d) = (%d, %d), want (%d, %d)",
					tt.month, tt.year, tt.startMonth, q, fy, tt.wantQ, tt.wantFY)
			}
		})
	}
}

func TestPublicLookupMonth(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  time.Month
		ok    bool
	}{
		{"january", time.January, true},
		{"Jan", time.January, true},
		{"MARCH", time.March, true},
		{"pizza", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, ok := wen.LookupMonth(tt.input)
			if ok != tt.ok || got != tt.want {
				t.Errorf("LookupMonth(%q) = (%v, %v), want (%v, %v)",
					tt.input, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestPublicCountWorkdays(t *testing.T) {
	t.Parallel()
	// Verify the exported function is accessible and returns expected values.
	start := date(2026, 3, 16) // Monday
	end := date(2026, 3, 23)   // Monday
	got := wen.CountWorkdays(start, end)
	if got != 5 {
		t.Errorf("CountWorkdays(Mon, next Mon) = %d, want 5", got)
	}
}

func TestPublicParseError(t *testing.T) {
	t.Parallel()
	_, err := wen.ParseRelative("not a date", ref)
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
	// Verify the error is a *wen.ParseError (public type)
	var pe *wen.ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *wen.ParseError, got %T", err)
	}
	if pe.Input != "not a date" {
		t.Errorf("ParseError.Input = %q, want %q", pe.Input, "not a date")
	}
}

func TestParseErrorUnwrap(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := wen.ParseRelativeContext(ctx, "tomorrow", ref)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	var pe *wen.ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *wen.ParseError, got %T", err)
	}
	if pe.Cause == nil {
		t.Fatal("expected Cause to be set on context cancellation error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("errors.Is(err, context.Canceled) = false, want true")
	}
}

func TestWithFiscalYearStartValidationError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		month int
	}{
		{"zero", 0},
		{"negative", -1},
		{"thirteen", 13},
		{"hundred", 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := wen.ParseRelative("tomorrow", ref, wen.WithFiscalYearStart(tt.month))
			if err == nil {
				t.Fatalf("WithFiscalYearStart(%d) should return a validation error", tt.month)
			}
			if !strings.Contains(err.Error(), "invalid fiscal year start month") {
				t.Errorf("error %q should mention invalid fiscal year start month", err.Error())
			}
		})
	}
}

func TestWithFiscalYearStartValidValues(t *testing.T) {
	t.Parallel()
	for month := 1; month <= 12; month++ {
		t.Run(fmt.Sprintf("month_%d", month), func(t *testing.T) {
			t.Parallel()
			_, err := wen.ParseRelative("tomorrow", ref, wen.WithFiscalYearStart(month))
			if err != nil {
				t.Fatalf("WithFiscalYearStart(%d) should not error: %v", month, err)
			}
		})
	}
}
