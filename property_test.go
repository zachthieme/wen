package wen

import (
	"fmt"
	"testing"
	"time"
)

// propRef is a distinct reference time for property tests, avoiding collision
// with the ref variable in wen_test.go.
var propRef = time.Date(2026, 4, 10, 9, 15, 33, 0, time.UTC) // Friday April 10, 2026

// propRefs is a set of reference times that exercise interesting boundaries.
var propRefs = []time.Time{
	propRef,
	time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC), // year end
	time.Date(2028, 2, 29, 12, 0, 0, 0, time.UTC),    // leap day
	time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),      // year start
	time.Date(2026, 3, 8, 2, 30, 0, 0, time.UTC),     // DST boundary (UTC is fine)
	time.Date(2026, 6, 15, 18, 45, 0, 0, time.UTC),   // mid-year
}

func TestPropertyTodayEqualsTruncateDay(t *testing.T) {
	t.Parallel()
	for _, r := range propRefs {
		t.Run(r.Format(time.RFC3339), func(t *testing.T) {
			got, err := ParseRelative("today", r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			want := TruncateDay(r)
			if !got.Equal(want) {
				t.Errorf("Parse(\"today\", %v) = %v, want TruncateDay = %v", r, got, want)
			}
		})
	}
}

func TestPropertyZeroOffsetIsToday(t *testing.T) {
	t.Parallel()
	expressions := []string{
		"in 0 days",
		"in 0 weeks",
		"in 0 months",
		"in 0 years",
	}
	for _, r := range propRefs {
		for _, expr := range expressions {
			t.Run(fmt.Sprintf("%s/%s", r.Format(time.DateOnly), expr), func(t *testing.T) {
				got, err := ParseRelative(expr, r)
				if err != nil {
					t.Fatalf("unexpected error for %q: %v", expr, err)
				}
				want := TruncateDay(r)
				if !got.Equal(want) {
					t.Errorf("Parse(%q, %v) = %v, want %v", expr, r, got, want)
				}
			})
		}
	}
}

func TestPropertyForwardMonotonicity(t *testing.T) {
	t.Parallel()
	units := []string{"days", "weeks", "months", "years"}
	for _, unit := range units {
		t.Run(unit, func(t *testing.T) {
			for m := 1; m <= 5; m++ {
				n := m + 1
				exprM := fmt.Sprintf("in %d %s", m, unit)
				exprN := fmt.Sprintf("in %d %s", n, unit)
				gotM, err := ParseRelative(exprM, propRef)
				if err != nil {
					t.Fatalf("unexpected error for %q: %v", exprM, err)
				}
				gotN, err := ParseRelative(exprN, propRef)
				if err != nil {
					t.Fatalf("unexpected error for %q: %v", exprN, err)
				}
				if !gotN.After(gotM) {
					t.Errorf("expected Parse(%q) > Parse(%q), got %v <= %v", exprN, exprM, gotN, gotM)
				}
			}
		})
	}
}

func TestPropertyBackwardMonotonicity(t *testing.T) {
	t.Parallel()
	units := []string{"days", "weeks", "months", "years"}
	for _, unit := range units {
		t.Run(unit, func(t *testing.T) {
			for m := 1; m <= 5; m++ {
				n := m + 1
				exprM := fmt.Sprintf("%d %s ago", m, unit)
				exprN := fmt.Sprintf("%d %s ago", n, unit)
				gotM, err := ParseRelative(exprM, propRef)
				if err != nil {
					t.Fatalf("unexpected error for %q: %v", exprM, err)
				}
				gotN, err := ParseRelative(exprN, propRef)
				if err != nil {
					t.Fatalf("unexpected error for %q: %v", exprN, err)
				}
				if !gotN.Before(gotM) {
					t.Errorf("expected Parse(%q) < Parse(%q), got %v >= %v", exprN, exprM, gotN, gotM)
				}
			}
		})
	}
}

func TestPropertyForwardBackwardSymmetry(t *testing.T) {
	t.Parallel()
	for n := 1; n <= 10; n++ {
		t.Run(fmt.Sprintf("%d_days", n), func(t *testing.T) {
			today := TruncateDay(propRef)
			fwd, err := ParseRelative(fmt.Sprintf("in %d days", n), propRef)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			bwd, err := ParseRelative(fmt.Sprintf("%d days ago", n), propRef)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			distFwd := fwd.Sub(today)
			distBwd := today.Sub(bwd)
			if distFwd != distBwd {
				t.Errorf("asymmetric: forward=%v, backward=%v (today=%v, fwd=%v, bwd=%v)",
					distFwd, distBwd, today, fwd, bwd)
			}
		})
	}
}

func TestPropertyMultiDateCountAndWeekday(t *testing.T) {
	t.Parallel()
	type testCase struct {
		weekday time.Weekday
		month   string
	}
	cases := []testCase{
		{time.Monday, "january"},
		{time.Tuesday, "february"},
		{time.Wednesday, "march"},
		{time.Thursday, "april"},
		{time.Friday, "may"},
		{time.Saturday, "june"},
		{time.Sunday, "july"},
	}
	for _, tc := range cases {
		expr := fmt.Sprintf("every %s in %s", tc.weekday, tc.month)
		t.Run(expr, func(t *testing.T) {
			dates, err := ParseMulti(expr, propRef)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", expr, err)
			}
			if len(dates) < 4 || len(dates) > 5 {
				t.Errorf("expected 4-5 dates for %q, got %d", expr, len(dates))
			}
			for _, d := range dates {
				if d.Weekday() != tc.weekday {
					t.Errorf("date %v is %s, want %s", d, d.Weekday(), tc.weekday)
				}
			}
		})
	}
}

func TestPropertyTomorrowEqualsTodayPlus24h(t *testing.T) {
	t.Parallel()
	for _, r := range propRefs {
		t.Run(r.Format(time.DateOnly), func(t *testing.T) {
			tomorrow, err := ParseRelative("tomorrow", r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			want := TruncateDay(r).AddDate(0, 0, 1)
			if !tomorrow.Equal(want) {
				t.Errorf("Parse(\"tomorrow\", %v) = %v, want %v", r, tomorrow, want)
			}
		})
	}
}

func TestPropertyTomorrowEqualsIn1Day(t *testing.T) {
	t.Parallel()
	for _, r := range propRefs {
		t.Run(r.Format(time.DateOnly), func(t *testing.T) {
			tomorrow, err := ParseRelative("tomorrow", r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			in1Day, err := ParseRelative("in 1 day", r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tomorrow.Equal(in1Day) {
				t.Errorf("tomorrow=%v != in 1 day=%v for ref %v", tomorrow, in1Day, r)
			}
		})
	}
}

func TestPropertyYesterdayEquals1DayAgo(t *testing.T) {
	t.Parallel()
	for _, r := range propRefs {
		t.Run(r.Format(time.DateOnly), func(t *testing.T) {
			yesterday, err := ParseRelative("yesterday", r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			oneDayAgo, err := ParseRelative("1 day ago", r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !yesterday.Equal(oneDayAgo) {
				t.Errorf("yesterday=%v != 1 day ago=%v for ref %v", yesterday, oneDayAgo, r)
			}
		})
	}
}

func TestPropertyOneWeekEquals7Days(t *testing.T) {
	t.Parallel()
	for _, r := range propRefs {
		t.Run(r.Format(time.DateOnly), func(t *testing.T) {
			in1Week, err := ParseRelative("in 1 week", r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			in7Days, err := ParseRelative("in 7 days", r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !in1Week.Equal(in7Days) {
				t.Errorf("in 1 week=%v != in 7 days=%v for ref %v", in1Week, in7Days, r)
			}
		})
	}
}
