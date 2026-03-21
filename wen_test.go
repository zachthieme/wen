package wen

import (
	"testing"
	"time"
)

// ref is Wednesday March 18, 2026 at 14:30 UTC
var ref = time.Date(2026, 3, 18, 14, 30, 0, 0, time.UTC)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestRelativeDay(t *testing.T) {
	tests := []struct {
		input string
		want  time.Time
	}{
		{"today", date(2026, 3, 18)},
		{"tomorrow", date(2026, 3, 19)},
		{"yesterday", date(2026, 3, 17)},
		{"Today", date(2026, 3, 18)},
		{"TOMORROW", date(2026, 3, 19)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
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

func TestModWeekday(t *testing.T) {
	// ref = Wednesday March 18, 2026
	tests := []struct {
		input string
		want  time.Time
	}{
		// Bare weekday = "this" = current week (Sun-Sat)
		{"thursday", date(2026, 3, 19)},
		{"monday", date(2026, 3, 16)},   // already passed this week
		{"wednesday", date(2026, 3, 18)}, // today
		{"sunday", date(2026, 3, 15)},    // start of this week

		// "this" modifier
		{"this thursday", date(2026, 3, 19)},
		{"this monday", date(2026, 3, 16)},

		// "next" modifier = next week
		{"next thursday", date(2026, 3, 26)},
		{"next monday", date(2026, 3, 23)},
		{"next wednesday", date(2026, 3, 25)},

		// "last" modifier = last week
		{"last thursday", date(2026, 3, 12)},
		{"last monday", date(2026, 3, 9)},
		{"last sunday", date(2026, 3, 8)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
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

func TestRelativeOffset(t *testing.T) {
	// ref = Wednesday March 18, 2026 14:30 UTC
	tests := []struct {
		input string
		want  time.Time
	}{
		{"in 3 days", date(2026, 3, 21)},
		{"in 1 week", date(2026, 3, 25)},
		{"in 2 months", date(2026, 5, 18)},
		{"in 1 year", date(2027, 3, 18)},
		{"2 weeks ago", date(2026, 3, 4)},
		{"3 days ago", date(2026, 3, 15)},
		{"1 month ago", date(2026, 2, 18)},
		{"3 months from now", date(2026, 6, 18)},
		{"2 weeks from now", date(2026, 4, 1)},
		{"in 2 hours", ref.Add(2 * time.Hour)},
		{"in 30 minutes", ref.Add(30 * time.Minute)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
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

func TestCountedWeekday(t *testing.T) {
	// ref = Wednesday March 18, 2026
	tests := []struct {
		input string
		want  time.Time
	}{
		{"in 1 monday", date(2026, 3, 23)},    // next Monday
		{"in 4 mondays", date(2026, 4, 13)},   // 4th Monday from now
		{"in 1 friday", date(2026, 3, 20)},    // this Friday
		{"in 2 fridays", date(2026, 3, 27)},   // next Friday after that
		{"in 1 wednesday", date(2026, 3, 25)}, // next Wednesday (not today)
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
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

func TestLexer(t *testing.T) {
	tests := []struct {
		input  string
		tokens []token
	}{
		{
			input: "today",
			tokens: []token{
				{Kind: tokenRelativeDay, Value: "today", Position: 0},
			},
		},
		{
			input: "next thursday",
			tokens: []token{
				{Kind: tokenModifier, Value: "next", Position: 0},
				{Kind: tokenWeekday, Value: "thursday", Weekday: time.Thursday, Position: 5},
			},
		},
		{
			input: "in 3 days",
			tokens: []token{
				{Kind: tokenPreposition, Value: "in", Position: 0},
				{Kind: tokenNumber, Value: "3", IntVal: 3, Position: 3},
				{Kind: tokenUnit, Value: "day", Position: 5},
			},
		},
		{
			input: "in 4 mondays",
			tokens: []token{
				{Kind: tokenPreposition, Value: "in", Position: 0},
				{Kind: tokenNumber, Value: "4", IntVal: 4, Position: 3},
				{Kind: tokenWeekday, Value: "monday", Weekday: time.Monday, Position: 5},
			},
		},
		{
			input: "the third thursday in april",
			tokens: []token{
				{Kind: tokenNoise, Value: "the", Position: 0},
				{Kind: tokenOrdinal, Value: "third", IntVal: 3, Position: 4},
				{Kind: tokenWeekday, Value: "thursday", Weekday: time.Thursday, Position: 10},
				{Kind: tokenPreposition, Value: "in", Position: 19},
				{Kind: tokenMonth, Value: "april", Month: time.April, Position: 22},
			},
		},
		{
			input: "march 15",
			tokens: []token{
				{Kind: tokenMonth, Value: "march", Month: time.March, Position: 0},
				{Kind: tokenNumber, Value: "15", IntVal: 15, Position: 6},
			},
		},
		{
			input: "april 3rd 2025",
			tokens: []token{
				{Kind: tokenMonth, Value: "april", Month: time.April, Position: 0},
				{Kind: tokenOrdinal, Value: "3rd", IntVal: 3, Position: 6},
				{Kind: tokenNumber, Value: "2025", IntVal: 2025, Position: 10},
			},
		},
		{
			input: "at 3pm",
			tokens: []token{
				{Kind: tokenPreposition, Value: "at", Position: 0},
				{Kind: tokenNumber, Value: "3", IntVal: 3, Position: 3},
				{Kind: tokenMeridiem, Value: "pm", Position: 4},
			},
		},
		{
			input: "at 15:00",
			tokens: []token{
				{Kind: tokenPreposition, Value: "at", Position: 0},
				{Kind: tokenNumber, Value: "15", IntVal: 15, Position: 3},
				{Kind: tokenColon, Value: ":", Position: 5},
				{Kind: tokenNumber, Value: "00", IntVal: 0, Position: 6},
			},
		},
		{
			input: "end of month",
			tokens: []token{
				{Kind: tokenBoundary, Value: "end", Position: 0},
				{Kind: tokenPreposition, Value: "of", Position: 4},
				{Kind: tokenUnit, Value: "month", Position: 7},
			},
		},
		{
			input: "2 weeks ago",
			tokens: []token{
				{Kind: tokenNumber, Value: "2", IntVal: 2, Position: 0},
				{Kind: tokenUnit, Value: "week", Position: 2},
				{Kind: tokenPreposition, Value: "ago", Position: 8},
			},
		},
		{
			input: "3 months from now",
			tokens: []token{
				{Kind: tokenNumber, Value: "3", IntVal: 3, Position: 0},
				{Kind: tokenUnit, Value: "month", Position: 2},
				{Kind: tokenPreposition, Value: "from", Position: 9},
				{Kind: tokenPreposition, Value: "now", Position: 14},
			},
		},
		{
			input: "TOMORROW",
			tokens: []token{
				{Kind: tokenRelativeDay, Value: "tomorrow", Position: 0},
			},
		},
		{
			input: "Next Thursday",
			tokens: []token{
				{Kind: tokenModifier, Value: "next", Position: 0},
				{Kind: tokenWeekday, Value: "thursday", Weekday: time.Thursday, Position: 5},
			},
		},
		{
			input: "at noon",
			tokens: []token{
				{Kind: tokenPreposition, Value: "at", Position: 0},
				{Kind: tokenNamedTime, Value: "noon", Position: 3},
			},
		},
		{
			input: "beginning of next week",
			tokens: []token{
				{Kind: tokenBoundary, Value: "beginning", Position: 0},
				{Kind: tokenPreposition, Value: "of", Position: 10},
				{Kind: tokenModifier, Value: "next", Position: 13},
				{Kind: tokenUnit, Value: "week", Position: 18},
			},
		},
		{
			input: "first monday of march",
			tokens: []token{
				{Kind: tokenOrdinal, Value: "first", IntVal: 1, Position: 0},
				{Kind: tokenWeekday, Value: "monday", Weekday: time.Monday, Position: 6},
				{Kind: tokenPreposition, Value: "of", Position: 13},
				{Kind: tokenMonth, Value: "march", Month: time.March, Position: 16},
			},
		},
		{
			input: "last friday in november",
			tokens: []token{
				{Kind: tokenModifier, Value: "last", Position: 0},
				{Kind: tokenWeekday, Value: "friday", Weekday: time.Friday, Position: 5},
				{Kind: tokenPreposition, Value: "in", Position: 12},
				{Kind: tokenMonth, Value: "november", Month: time.November, Position: 15},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := newLexer(tt.input)
			tokens := l.tokenize()

			// Remove EOF for comparison
			got := tokens[:len(tokens)-1]
			if len(got) != len(tt.tokens) {
				t.Fatalf("token count: got %d, want %d\ngot:  %+v", len(got), len(tt.tokens), got)
			}
			for i, want := range tt.tokens {
				g := got[i]
				if g.Kind != want.Kind {
					t.Errorf("token[%d].Kind = %v, want %v", i, g.Kind, want.Kind)
				}
				if g.Value != want.Value {
					t.Errorf("token[%d].Value = %q, want %q", i, g.Value, want.Value)
				}
				if g.IntVal != want.IntVal {
					t.Errorf("token[%d].IntVal = %d, want %d", i, g.IntVal, want.IntVal)
				}
				if want.Kind == tokenWeekday && g.Weekday != want.Weekday {
					t.Errorf("token[%d].Weekday = %v, want %v", i, g.Weekday, want.Weekday)
				}
				if want.Kind == tokenMonth && g.Month != want.Month {
					t.Errorf("token[%d].Month = %v, want %v", i, g.Month, want.Month)
				}
				if g.Position != want.Position {
					t.Errorf("token[%d].Position = %d, want %d", i, g.Position, want.Position)
				}
			}
		})
	}
}

func TestAbsoluteDate(t *testing.T) {
	tests := []struct {
		input string
		want  time.Time
	}{
		{"march 15", date(2026, 3, 15)},
		{"april 3rd 2025", date(2025, 4, 3)},
		{"december 25", date(2026, 12, 25)},
		{"january 1 2027", date(2027, 1, 1)},
		{"april 3rd", date(2026, 4, 3)},
		// February already passed — use next year
		{"february 14", date(2027, 2, 14)},
		// Year-only (no day) — defaults to 1st of month
		{"march 2027", date(2027, 3, 1)},
		{"december 2026", date(2026, 12, 1)},
		{"january 2028", date(2028, 1, 1)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
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

func TestPeriodRef(t *testing.T) {
	// ref = Wednesday March 18, 2026
	// Sunday of this week = March 15
	tests := []struct {
		input string
		opts  []Option
		want  time.Time
	}{
		// PeriodStart (default)
		{"this week", nil, date(2026, 3, 15)},
		{"next week", nil, date(2026, 3, 22)},
		{"last week", nil, date(2026, 3, 8)},
		{"this month", nil, date(2026, 3, 1)},
		{"next month", nil, date(2026, 4, 1)},
		{"last month", nil, date(2026, 2, 1)},

		// PeriodSame
		{"next week", []Option{WithPeriodSame()}, date(2026, 3, 25)},
		{"last week", []Option{WithPeriodSame()}, date(2026, 3, 11)},
		{"next month", []Option{WithPeriodSame()}, date(2026, 4, 18)},
		{"last month", []Option{WithPeriodSame()}, date(2026, 2, 18)},

		// Boundaries
		{"end of month", nil, time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)},
		{"beginning of month", nil, date(2026, 3, 1)},
		{"beginning of next week", nil, date(2026, 3, 22)},
		{"end of next month", nil, time.Date(2026, 4, 30, 23, 59, 59, 0, time.UTC)},
		{"beginning of last month", nil, date(2026, 2, 1)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseRelative(tt.input, ref, tt.opts...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeExpr(t *testing.T) {
	tests := []struct {
		input string
		want  time.Time
	}{
		// Standalone time (applied to ref date)
		{"at 3pm", time.Date(2026, 3, 18, 15, 0, 0, 0, time.UTC)},
		{"at 3 pm", time.Date(2026, 3, 18, 15, 0, 0, 0, time.UTC)},
		{"at 11am", time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC)},
		{"at 15:00", time.Date(2026, 3, 18, 15, 0, 0, 0, time.UTC)},
		{"at 3:30pm", time.Date(2026, 3, 18, 15, 30, 0, 0, time.UTC)},
		{"at noon", time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)},
		{"at midnight", time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)},
		{"noon", time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)},
		{"midnight", time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)},
		{"3pm", time.Date(2026, 3, 18, 15, 0, 0, 0, time.UTC)},
		// 12am = midnight, 12pm = noon
		{"at 12am", time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)},
		{"at 12pm", time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
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

func TestCombinedExpr(t *testing.T) {
	tests := []struct {
		input string
		want  time.Time
	}{
		{"tomorrow at 3pm", time.Date(2026, 3, 19, 15, 0, 0, 0, time.UTC)},
		{"next thursday at noon", time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)},
		{"march 15 at 9am", time.Date(2026, 3, 15, 9, 0, 0, 0, time.UTC)},
		{"yesterday at midnight", time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC)},
		{"next thursday at 15:00", time.Date(2026, 3, 26, 15, 0, 0, 0, time.UTC)},
		{"today at 8:30am", time.Date(2026, 3, 18, 8, 30, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
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

func TestErrors(t *testing.T) {
	tests := []struct {
		input    string
		wantPos  int
		wantFind string
	}{
		{"in 4 flurbs", 5, "flurbs"},
		{"blah", 0, "blah"},
		{"", 0, ""},
		{"next", 4, ""},  // unexpected EOF
		{"fifth monday in february", 0, ""},              // Feb 2026 has only 4 Mondays
		{"pizza", 0, "pizza"},                             // gibberish
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseRelative(tt.input, ref)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			pe, ok := err.(*ParseError)
			if !ok {
				t.Fatalf("expected *ParseError, got %T", err)
			}
			if pe.Input != tt.input {
				t.Errorf("Input = %q, want %q", pe.Input, tt.input)
			}
			if tt.wantPos > 0 && pe.Position != tt.wantPos {
				t.Errorf("Position = %d, want %d", pe.Position, tt.wantPos)
			}
			if tt.wantFind != "" && pe.Found != tt.wantFind {
				t.Errorf("Found = %q, want %q", pe.Found, tt.wantFind)
			}
			if len(pe.Expected) == 0 {
				t.Error("Expected should not be empty")
			}
			// Verify Error() produces readable output
			errStr := pe.Error()
			if errStr == "" {
				t.Error("Error() should not return empty string")
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		ref   time.Time
		opts  []Option
		want  time.Time
	}{
		{
			name:  "year boundary: next month from december",
			input: "next month",
			ref:   time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC),
			want:  date(2027, 1, 1),
		},
		{
			name:  "last month from january",
			input: "last month",
			ref:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			want:  date(2025, 12, 1),
		},
		{
			name:  "leap year: feb 29",
			input: "february 29",
			ref:   time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC),
			want:  date(2028, 2, 29),
		},
		{
			name:  "next month on jan 31 with PeriodSame",
			input: "next month",
			ref:   time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
			opts:  []Option{WithPeriodSame()},
			want:  date(2026, 3, 3), // Go's AddDate rolls over
		},
		{
			name:  "ref on Sunday: this monday is in current week",
			input: "this monday",
			ref:   time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC), // Sunday
			want:  date(2026, 3, 23), // Monday of this Sun-Sat week
		},
		{
			name:  "timezone preservation",
			input: "tomorrow",
			ref:   time.Date(2026, 3, 18, 14, 0, 0, 0, time.FixedZone("EST", -5*3600)),
			want:  time.Date(2026, 3, 19, 0, 0, 0, 0, time.FixedZone("EST", -5*3600)),
		},
		{
			name:  "end of february",
			input: "end of month",
			ref:   time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC),
			want:  time.Date(2026, 2, 28, 23, 59, 59, 0, time.UTC),
		},
		{
			name:  "end of february leap year",
			input: "end of month",
			ref:   time.Date(2028, 2, 15, 0, 0, 0, 0, time.UTC),
			want:  time.Date(2028, 2, 29, 23, 59, 59, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRelative(tt.input, tt.ref, tt.opts...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrdinalWeekdayInMonth(t *testing.T) {
	tests := []struct {
		input string
		ref   time.Time
		want  time.Time
	}{
		// April 2026: starts on Wednesday
		// Thursdays: 2, 9, 16, 23, 30
		{"the third thursday in april", ref, date(2026, 4, 16)},
		{"third thursday in april", ref, date(2026, 4, 16)},
		{"first thursday in april", ref, date(2026, 4, 2)},

		// March 2026: starts on Sunday
		// Mondays: 2, 9, 16, 23, 30
		{"first monday of march", ref, date(2026, 3, 2)},
		{"second monday of march", ref, date(2026, 3, 9)},

		// November 2026: starts on Sunday
		// Fridays: 6, 13, 20, 27
		{"last friday in november", ref, date(2026, 11, 27)},

		// If month already passed, use next year
		// ref is March 2026, so "february" means Feb 2027
		// Feb 2027: starts on Monday
		// Mondays: 1, 8, 15, 22
		{"first monday of february", ref, date(2027, 2, 1)},

		// Optional preposition — "in"/"of" not required
		{"first monday march", ref, date(2026, 3, 2)},
		{"third thursday april", ref, date(2026, 4, 16)},

		// Explicit year
		{"third monday in march 2027", ref, date(2027, 3, 15)},
		{"first thursday april 2028", ref, date(2028, 4, 6)},
		{"last friday in november 2027", ref, date(2027, 11, 26)},

		// Unsupported patterns should error
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseRelative(tt.input, tt.ref)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseConvenience(t *testing.T) {
	// Smoke test for Parse() - just verify it doesn't error on valid input
	_, err := Parse("tomorrow")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func FuzzParse(f *testing.F) {
	seeds := []string{
		"today", "tomorrow", "yesterday",
		"next friday", "last monday", "this wednesday",
		"in 3 days", "2 weeks ago", "5 months from now",
		"march 15 2026", "march 2027", "december 25",
		"first monday of march", "last friday in november",
		"beginning of next week", "end of this month",
		"at noon", "at 3pm", "march 15 at 15:30",
		"in 3 fridays",
		"", "   ", "not a date", "999999999 days ago",
		"the the the", "at at at", "next next next",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(_ *testing.T, input string) {
		// Must not panic regardless of input.
		_, _ = ParseRelative(input, ref)
	})
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid day: february 30", "february 30"},
		{"invalid day: april 31", "april 31"},
		{"invalid day: february 29 non-leap year", "february 29"},
		{"invalid hour 24h: at 25:00", "at 25:00"},
		{"invalid minute: at 3:99", "at 3:99"},
		{"invalid meridiem hour too high: at 13pm", "at 13pm"},
		{"invalid meridiem hour zero: at 0pm", "at 0pm"},
		{"invalid meridiem hour colon: at 13:00pm", "at 13:00pm"},
		{"invalid meridiem hour zero colon: at 0:30am", "at 0:30am"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRelative(tt.input, ref)
			if err == nil {
				t.Errorf("expected error for %q, got nil", tt.input)
			}
		})
	}
}

func TestBoundaryQuarterYear(t *testing.T) {
	tests := []struct {
		input string
		want  time.Time
	}{
		// ref = March 18, 2026 (Q1)
		{"end of quarter", time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)},
		{"beginning of quarter", date(2026, 1, 1)},
		{"end of next quarter", time.Date(2026, 6, 30, 23, 59, 59, 0, time.UTC)},
		{"beginning of next quarter", date(2026, 4, 1)},
		{"end of last quarter", time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)},
		{"beginning of last quarter", date(2025, 10, 1)},
		{"end of year", time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)},
		{"beginning of year", date(2026, 1, 1)},
		{"end of next year", time.Date(2027, 12, 31, 23, 59, 59, 0, time.UTC)},
		{"beginning of last year", date(2025, 1, 1)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
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

func TestOrdinalWeekdayNextMonth(t *testing.T) {
	// ref = March 18, 2026
	// Next month = April 2026, starts on Wednesday
	// Mondays in April: 6, 13, 20, 27
	tests := []struct {
		input string
		want  time.Time
	}{
		{"first monday of next month", date(2026, 4, 6)},
		{"second friday of next month", date(2026, 4, 10)},
		{"first monday of last month", date(2026, 2, 2)},
		{"third wednesday of this month", date(2026, 3, 18)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
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

func TestEveryWeekdayInMonth(t *testing.T) {
	// ref = March 18, 2026
	// Fridays in April 2026: 3, 10, 17, 24
	got, err := ParseMulti("every friday in april", ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []time.Time{
		date(2026, 4, 3),
		date(2026, 4, 10),
		date(2026, 4, 17),
		date(2026, 4, 24),
	}
	if len(got) != len(want) {
		t.Fatalf("got %d dates, want %d", len(got), len(want))
	}
	for i, w := range want {
		if !got[i].Equal(w) {
			t.Errorf("date[%d] = %v, want %v", i, got[i], w)
		}
	}
}

func TestEveryWeekdayInMonthWithYear(t *testing.T) {
	got, err := ParseMulti("every monday in march 2027", ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// March 2027 starts on Monday
	// Mondays: 1, 8, 15, 22, 29
	want := []time.Time{
		date(2027, 3, 1),
		date(2027, 3, 8),
		date(2027, 3, 15),
		date(2027, 3, 22),
		date(2027, 3, 29),
	}
	if len(got) != len(want) {
		t.Fatalf("got %d dates, want %d", len(got), len(want))
	}
	for i, w := range want {
		if !got[i].Equal(w) {
			t.Errorf("date[%d] = %v, want %v", i, got[i], w)
		}
	}
}

func TestParseMultiFallsBack(t *testing.T) {
	// Single-date expressions should still work via ParseMulti
	got, err := ParseMulti("next friday", ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 date, got %d", len(got))
	}
}

func TestWithPeriodStartExplicit(t *testing.T) {
	// Verify explicit WithPeriodStart matches default behavior
	r := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)
	got1, _ := ParseRelative("next week", r)
	got2, _ := ParseRelative("next week", r, WithPeriodStart())
	if !got1.Equal(got2) {
		t.Errorf("WithPeriodStart should match default: got %v vs %v", got1, got2)
	}
}
