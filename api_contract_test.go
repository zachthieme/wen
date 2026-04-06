package wen_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zachthieme/wen"
)

// contractRef is Wednesday March 18, 2026 at 14:30 UTC.
// Separate from api_test.go's ref to avoid collisions.
var contractRef = time.Date(2026, 3, 18, 14, 30, 0, 0, time.UTC)

// ---------------------------------------------------------------------------
// 1. Parse function signatures and basic behavior
// ---------------------------------------------------------------------------

func TestContractParseReturnsValidDate(t *testing.T) {
	t.Parallel()
	got, err := wen.Parse("tomorrow")
	if err != nil {
		t.Fatalf("Parse(\"tomorrow\") returned error: %v", err)
	}
	if got.IsZero() {
		t.Fatal("Parse(\"tomorrow\") returned zero time")
	}
}

func TestContractParseRelativeWithRef(t *testing.T) {
	t.Parallel()
	got, err := wen.ParseRelative("today", contractRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("ParseRelative(\"today\", ref) = %v, want %v", got, want)
	}
}

func TestContractParseMultiReturnsSlice(t *testing.T) {
	t.Parallel()
	dates, err := wen.ParseMulti("every friday in march", contractRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dates) < 4 {
		t.Fatalf("expected at least 4 dates, got %d", len(dates))
	}
	for i, d := range dates {
		if d.Weekday() != time.Friday {
			t.Errorf("dates[%d] weekday = %v, want Friday", i, d.Weekday())
		}
		if d.Month() != time.March {
			t.Errorf("dates[%d] month = %v, want March", i, d.Month())
		}
	}
}

func TestContractParseToExprReturnsExpr(t *testing.T) {
	t.Parallel()
	expr, err := wen.ParseToExpr("tomorrow")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expr == nil {
		t.Fatal("ParseToExpr(\"tomorrow\") returned nil Expr")
	}
}

// ---------------------------------------------------------------------------
// 2. Error contract
// ---------------------------------------------------------------------------

func TestContractInvalidInputReturnsParseError(t *testing.T) {
	t.Parallel()
	_, err := wen.ParseRelative("not a date", contractRef)
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
	var pe *wen.ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *wen.ParseError, got %T", err)
	}
}

func TestContractParseErrorHasInput(t *testing.T) {
	t.Parallel()
	const input = "zzz gibberish qqq"
	_, err := wen.ParseRelative(input, contractRef)
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
	var pe *wen.ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *wen.ParseError, got %T", err)
	}
	if pe.Input != input {
		t.Errorf("ParseError.Input = %q, want %q", pe.Input, input)
	}
}

func TestContractParseErrorUnwrapsContextCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := wen.ParseRelativeContext(ctx, "tomorrow", contractRef)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("errors.Is(err, context.Canceled) = false, want true")
	}
}

// ---------------------------------------------------------------------------
// 3. Option contracts
// ---------------------------------------------------------------------------

func TestContractWithFiscalYearStartValidRange(t *testing.T) {
	t.Parallel()
	// Months 1-12 must be accepted.
	for month := 1; month <= 12; month++ {
		_, err := wen.ParseRelative("tomorrow", contractRef, wen.WithFiscalYearStart(month))
		if err != nil {
			t.Errorf("WithFiscalYearStart(%d) should not error: %v", month, err)
		}
	}
	// 0 and 13 must be rejected.
	for _, bad := range []int{0, 13} {
		_, err := wen.ParseRelative("tomorrow", contractRef, wen.WithFiscalYearStart(bad))
		if err == nil {
			t.Errorf("WithFiscalYearStart(%d) should return an error", bad)
		}
	}
}

func TestContractWithPeriodStartIsDefault(t *testing.T) {
	t.Parallel()
	got, err := wen.ParseRelative("next week", contractRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Default PeriodStart: "next week" resolves to start of next week (Sunday).
	want := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("default \"next week\" = %v, want %v (Sunday)", got, want)
	}
}

func TestContractWithPeriodSamePreservesDay(t *testing.T) {
	t.Parallel()
	got, err := wen.ParseRelative("next week", contractRef, wen.WithPeriodSame())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// PeriodSame: "next week" = ref + 7 days.
	want := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("WithPeriodSame \"next week\" = %v, want %v (ref+7)", got, want)
	}
}

// ---------------------------------------------------------------------------
// 4. Utility function contracts
// ---------------------------------------------------------------------------

func TestContractTruncateDayPreservesLocation(t *testing.T) {
	t.Parallel()
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("failed to load location: %v", err)
	}
	input := time.Date(2026, 6, 15, 14, 30, 45, 123, loc)
	got := wen.TruncateDay(input)
	if got.Location() != loc {
		t.Errorf("Location = %v, want %v", got.Location(), loc)
	}
}

func TestContractTruncateDayZerosTime(t *testing.T) {
	t.Parallel()
	input := time.Date(2026, 6, 15, 14, 30, 45, 999, time.UTC)
	got := wen.TruncateDay(input)
	if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
		t.Errorf("time components = %02d:%02d:%02d.%d, want 00:00:00.0",
			got.Hour(), got.Minute(), got.Second(), got.Nanosecond())
	}
}

func TestContractDaysInFebruary(t *testing.T) {
	t.Parallel()
	t.Run("non-leap 2026", func(t *testing.T) {
		t.Parallel()
		if got := wen.DaysIn(2026, time.February, time.UTC); got != 28 {
			t.Errorf("DaysIn(2026, Feb) = %d, want 28", got)
		}
	})
	t.Run("leap 2024", func(t *testing.T) {
		t.Parallel()
		if got := wen.DaysIn(2024, time.February, time.UTC); got != 29 {
			t.Errorf("DaysIn(2024, Feb) = %d, want 29", got)
		}
	})
}

func TestContractCountWorkdaysMonToMon(t *testing.T) {
	t.Parallel()
	mon1 := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC) // Monday
	mon2 := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC) // Next Monday
	got := wen.CountWorkdays(mon1, mon2)
	if got != 5 {
		t.Errorf("CountWorkdays(Mon, next Mon) = %d, want 5", got)
	}
}

func TestContractCountWorkdaysSymmetric(t *testing.T) {
	t.Parallel()
	a := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)
	b := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
	ab := wen.CountWorkdays(a, b)
	ba := wen.CountWorkdays(b, a)
	if ab != ba {
		t.Errorf("CountWorkdays(a,b) = %d, CountWorkdays(b,a) = %d; want equal", ab, ba)
	}
}

func TestContractLookupMonthCaseInsensitive(t *testing.T) {
	t.Parallel()
	inputs := []string{"january", "January", "JANUARY", "jan"}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			got, ok := wen.LookupMonth(input)
			if !ok {
				t.Fatalf("LookupMonth(%q) returned false", input)
			}
			if got != time.January {
				t.Errorf("LookupMonth(%q) = %v, want January", input, got)
			}
		})
	}
}

func TestContractFiscalQuarterCalendarYear(t *testing.T) {
	t.Parallel()
	cases := []struct {
		month int
		wantQ int
	}{
		{1, 1}, {4, 2}, {7, 3}, {10, 4},
	}
	for _, tc := range cases {
		t.Run(time.Month(tc.month).String(), func(t *testing.T) {
			t.Parallel()
			q, fy := wen.FiscalQuarter(tc.month, 2026, 1)
			if q != tc.wantQ {
				t.Errorf("FiscalQuarter(%d, 2026, 1) quarter = %d, want %d", tc.month, q, tc.wantQ)
			}
			if fy != 2026 {
				t.Errorf("FiscalQuarter(%d, 2026, 1) fiscal year = %d, want 2026", tc.month, fy)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 5. DateLayout constant contract
// ---------------------------------------------------------------------------

func TestContractDateLayout(t *testing.T) {
	t.Parallel()
	if wen.DateLayout != "2006-01-02" {
		t.Errorf("DateLayout = %q, want %q", wen.DateLayout, "2006-01-02")
	}
}
