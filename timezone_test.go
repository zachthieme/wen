package wen

import (
	"testing"
	"time"
)

func TestTimezoneNonHourAligned(t *testing.T) {
	t.Parallel()

	// Load real IANA timezones with non-hour offsets.
	zones := []struct {
		name   string
		tzName string
	}{
		{"Kathmandu +05:45", "Asia/Kathmandu"},
		{"Chatham +12:45", "Pacific/Chatham"},
		{"Newfoundland -03:30", "America/St_Johns"},
		{"India +05:30", "Asia/Kolkata"},
		{"Iran +03:30", "Asia/Tehran"},
	}

	for _, z := range zones {
		loc, err := time.LoadLocation(z.tzName)
		if err != nil {
			t.Skipf("timezone %s not available: %v", z.tzName, err)
		}

		t.Run(z.name, func(t *testing.T) {
			t.Parallel()
			// Wednesday March 18, 2026 14:30 in the given timezone
			ref := time.Date(2026, 3, 18, 14, 30, 0, 0, loc)

			tests := []struct {
				input string
				want  time.Time
			}{
				{"today", time.Date(2026, 3, 18, 0, 0, 0, 0, loc)},
				{"tomorrow", time.Date(2026, 3, 19, 0, 0, 0, 0, loc)},
				{"yesterday", time.Date(2026, 3, 17, 0, 0, 0, 0, loc)},
				{"in 3 days", time.Date(2026, 3, 21, 0, 0, 0, 0, loc)},
				{"3 days ago", time.Date(2026, 3, 15, 0, 0, 0, 0, loc)},
				{"next friday", time.Date(2026, 3, 27, 0, 0, 0, 0, loc)},
				{"march 25 2026", time.Date(2026, 3, 25, 0, 0, 0, 0, loc)},
			}

			for _, tt := range tests {
				t.Run(tt.input, func(t *testing.T) {
					t.Parallel()
					got, err := ParseRelative(tt.input, ref)
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if !got.Equal(tt.want) {
						t.Errorf("got %v (loc=%s), want %v (loc=%s)",
							got, got.Location(), tt.want, tt.want.Location())
					}
					// Verify the result is in the same timezone as the reference
					if got.Location().String() != loc.String() {
						t.Errorf("result location = %s, want %s", got.Location(), loc)
					}
				})
			}
		})
	}
}

func TestTimezoneDSTTransition(t *testing.T) {
	t.Parallel()

	// US Eastern: DST spring forward March 8, 2026 at 2:00 AM
	eastern, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("America/New_York timezone not available")
	}

	// Reference time just before DST transition
	ref := time.Date(2026, 3, 7, 23, 0, 0, 0, eastern)

	tests := []struct {
		input string
		want  time.Time
	}{
		// "tomorrow" crosses the DST spring-forward boundary
		{"tomorrow", time.Date(2026, 3, 8, 0, 0, 0, 0, eastern)},
		{"in 2 days", time.Date(2026, 3, 9, 0, 0, 0, 0, eastern)},
		// Back across DST fall-back: Nov 1 2026 at 2:00 AM clocks fall back
		{"november 1 2026", time.Date(2026, 11, 1, 0, 0, 0, 0, eastern)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := ParseRelative(tt.input, ref)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimezoneMidnightCrossing(t *testing.T) {
	t.Parallel()

	// Test with a reference time very close to midnight in a non-UTC timezone.
	// This verifies that TruncateDay preserves the location correctly.
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Skip("Asia/Tokyo timezone not available")
	}

	// 23:59 in Tokyo on March 18 = 14:59 UTC on March 18
	ref := time.Date(2026, 3, 18, 23, 59, 59, 0, tokyo)

	got, err := ParseRelative("today", ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2026, 3, 18, 0, 0, 0, 0, tokyo)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}

	// "tomorrow" should be March 19 in Tokyo, not March 19 UTC
	got, err = ParseRelative("tomorrow", ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want = time.Date(2026, 3, 19, 0, 0, 0, 0, tokyo)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestTimezoneMonthBoundary(t *testing.T) {
	t.Parallel()

	// Verify month arithmetic works correctly across timezones.
	apia, err := time.LoadLocation("Pacific/Apia")
	if err != nil {
		t.Skip("Pacific/Apia timezone not available")
	}

	// Jan 31 in Apia — "in 1 month" should handle the 31→28 clamp
	ref := time.Date(2026, 1, 31, 12, 0, 0, 0, apia)

	got, err := ParseRelative("in 1 month", ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Go's AddDate(0,1,0) from Jan 31 = March 3 (28+3 days)
	// This is Go's documented behavior — verify we match it
	want := time.Date(2026, 1, 31, 0, 0, 0, 0, apia).AddDate(0, 1, 0)
	want = TruncateDay(want)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
