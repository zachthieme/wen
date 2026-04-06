package wen

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// ref is Wednesday March 18, 2026 at 14:30 UTC
var ref = time.Date(2026, 3, 18, 14, 30, 0, 0, time.UTC)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestRelativeDay(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
		// Cardinal number words
		{"two weeks ago", date(2026, 3, 4)},
		{"three days ago", date(2026, 3, 15)},
		{"in five days", date(2026, 3, 23)},
		{"in ten days", date(2026, 3, 28)},
		{"one month ago", date(2026, 2, 18)},
		// Large offsets — verifies shiftMonth handles big deltas correctly
		{"in 120 months", date(2036, 3, 18)},
		{"120 months ago", date(2016, 3, 18)},
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
	t.Parallel()
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
	t.Parallel()
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
		{
			input: "99999999999999999999",
			tokens: []token{
				{Kind: tokenUnknown, Value: "99999999999999999999", Position: 0},
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
		// Bare number after "at" — 24-hour time
		{"at 3", time.Date(2026, 3, 18, 3, 0, 0, 0, time.UTC)},
		{"at 15", time.Date(2026, 3, 18, 15, 0, 0, 0, time.UTC)},
		{"at 0", time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)},
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
	t.Parallel()
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
		// Bare number after "at"
		{"tomorrow at 3", time.Date(2026, 3, 19, 3, 0, 0, 0, time.UTC)},
		{"tomorrow at 15", time.Date(2026, 3, 19, 15, 0, 0, 0, time.UTC)},
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
	t.Parallel()
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
		{"99999999999999999999", 0, "99999999999999999999"}, // numeric overflow
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseRelative(tt.input, ref)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var pe *ParseError
			if !errors.As(err, &pe) {
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
	t.Parallel()
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
		{
			name:  "absolute date: march from december infers next year",
			input: "march 25",
			ref:   time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC),
			want:  date(2027, 3, 25),
		},
		{
			name:  "absolute date: december from january stays same year",
			input: "december 25",
			ref:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			want:  date(2026, 12, 25),
		},
		{
			name:  "absolute date: same month infers current year",
			input: "december 31",
			ref:   time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC),
			want:  date(2026, 12, 31),
		},
		{
			name:  "ordinal weekday: first monday of march from december",
			input: "first monday of march",
			ref:   time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC),
			want:  date(2027, 3, 1), // March 1, 2027 is a Monday
		},
		{
			name:  "last weekday in month: last friday in february from december",
			input: "last friday in february",
			ref:   time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC),
			want:  date(2027, 2, 26),
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

func TestMultiDateYearInference(t *testing.T) {
	t.Parallel()
	// Parsing "every tuesday in january" from a December ref should infer next year.
	decRef := time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC)
	dates, err := ParseMulti("every tuesday in january", decRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dates) == 0 {
		t.Fatal("expected dates, got none")
	}
	for _, d := range dates {
		if d.Year() != 2027 {
			t.Errorf("expected year 2027, got %d for %v", d.Year(), d)
		}
		if d.Month() != time.January {
			t.Errorf("expected January, got %v for %v", d.Month(), d)
		}
		if d.Weekday() != time.Tuesday {
			t.Errorf("expected Tuesday, got %v for %v", d.Weekday(), d)
		}
	}
}

func TestOrdinalWeekdayInMonth(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
		"", "   ", "not a date",
		"the the the", "at at at", "next next next",
		"999999999 days ago", // extreme offset — intentionally outside range check
	}
	for _, s := range seeds {
		f.Add(s)
	}

	// Reasonable bounds: a successful parse should not produce a date
	// more than 200 years from the reference time.
	const maxYears = 200
	lower := ref.AddDate(-maxYears, 0, 0)
	upper := ref.AddDate(maxYears, 0, 0)

	f.Fuzz(func(t *testing.T, input string) {
		got, err := ParseRelative(input, ref)
		if err != nil {
			return // parse failures are fine — we only assert on successes
		}

		// Property 1: result must not be the zero time.
		if got.IsZero() {
			t.Errorf("ParseRelative(%q) returned zero time without error", input)
		}

		// Property 2: result should be within a reasonable range of the reference.
		// Log (don't fail) extreme dates — intentional seeds like "999999999 days ago"
		// produce valid but extreme results.
		if got.Before(lower) || got.After(upper) {
			t.Logf("ParseRelative(%q) = %v, outside reasonable window of %v",
				input, got, ref)
		}

		// Property 3: successful ParseMulti must also not panic and must
		// return at least one result.
		results, err := ParseMulti(input, ref)
		if err == nil && len(results) == 0 {
			t.Errorf("ParseMulti(%q) returned 0 results without error", input)
		}
	})
}

func TestParseConcurrentSafety(t *testing.T) {
	t.Parallel()
	inputs := []string{
		"tomorrow", "next friday", "in 3 days", "march 15 2026",
		"last monday", "end of quarter", "every friday in april",
		"first monday of march", "2 weeks ago", "at noon",
	}
	var wg sync.WaitGroup
	for i := range 100 {
		input := inputs[i%len(inputs)]
		wg.Go(func() {
			// ParseRelative and ParseMulti must be safe for concurrent use.
			_, _ = ParseRelative(input, ref)
			_, _ = ParseMulti(input, ref)
		})
	}
	wg.Wait()
}

func TestValidation(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

func TestFiscalYearQuarters(t *testing.T) {
	t.Parallel()
	// ref = March 18, 2026
	// With fiscal year starting in October:
	// FY Q1 = Oct-Dec, Q2 = Jan-Mar, Q3 = Apr-Jun, Q4 = Jul-Sep
	// March is in FY Q2 (Jan-Mar)
	fyOct := WithFiscalYearStart(10)
	tests := []struct {
		name string
		input string
		opts  []Option
		want  time.Time
	}{
		// Calendar quarters (default, fiscal_year_start=1)
		{"cal end of quarter", "end of quarter", nil,
			time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)},
		{"cal beginning of quarter", "beginning of quarter", nil,
			date(2026, 1, 1)},

		// Fiscal year starting October: March is in Q2 (Jan-Mar)
		{"fy oct: end of quarter", "end of quarter", []Option{fyOct},
			time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)},
		{"fy oct: beginning of quarter", "beginning of quarter", []Option{fyOct},
			date(2026, 1, 1)},

		// Next quarter with fiscal Oct: Q2 is Jan-Mar, next is Q3 = Apr-Jun
		{"fy oct: end of next quarter", "end of next quarter", []Option{fyOct},
			time.Date(2026, 6, 30, 23, 59, 59, 0, time.UTC)},
		{"fy oct: beginning of next quarter", "beginning of next quarter", []Option{fyOct},
			date(2026, 4, 1)},

		// Last quarter with fiscal Oct: Q2 is Jan-Mar, last is Q1 = Oct-Dec
		{"fy oct: beginning of last quarter", "beginning of last quarter", []Option{fyOct},
			date(2025, 10, 1)},
		{"fy oct: end of last quarter", "end of last quarter", []Option{fyOct},
			time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)},

		// Fiscal year starting April (common in UK/Japan): March is Q4 (Jan-Mar)
		{"fy apr: beginning of quarter", "beginning of quarter", []Option{WithFiscalYearStart(4)},
			date(2026, 1, 1)},
		{"fy apr: end of quarter", "end of quarter", []Option{WithFiscalYearStart(4)},
			time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)},
		{"fy apr: beginning of next quarter", "beginning of next quarter", []Option{WithFiscalYearStart(4)},
			date(2026, 4, 1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestOrdinalWeekdayNextMonth(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	// Verify explicit WithPeriodStart matches default behavior
	r := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)
	got1, _ := ParseRelative("next week", r)
	got2, _ := ParseRelative("next week", r, WithPeriodStart())
	if !got1.Equal(got2) {
		t.Errorf("WithPeriodStart should match default: got %v vs %v", got1, got2)
	}
}

func TestFiscalQuarter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		month      int
		year       int
		startMonth int
		wantQ      int
		wantFY     int
	}{
		// 1. Standard calendar year (startMonth=1)
		{"cal: Jan→Q1", 1, 2026, 1, 1, 2026},
		{"cal: Apr→Q2", 4, 2026, 1, 2, 2026},
		{"cal: Jul→Q3", 7, 2026, 1, 3, 2026},
		{"cal: Oct→Q4", 10, 2026, 1, 4, 2026},

		// 2. US federal fiscal year (startMonth=10)
		{"us fed: Oct 2025→Q1 FY2026", 10, 2025, 10, 1, 2026},
		{"us fed: Jan 2026→Q2 FY2026", 1, 2026, 10, 2, 2026},
		{"us fed: Apr 2026→Q3 FY2026", 4, 2026, 10, 3, 2026},
		{"us fed: Jul 2026→Q4 FY2026", 7, 2026, 10, 4, 2026},
		{"us fed: Sep 2026→Q4 FY2026", 9, 2026, 10, 4, 2026},

		// 3. UK/Japan fiscal year (startMonth=4)
		{"uk: Apr 2026→Q1 FY2027", 4, 2026, 4, 1, 2027},
		{"uk: Jul 2026→Q2 FY2027", 7, 2026, 4, 2, 2027},
		{"uk: Jan 2026→Q4 FY2026", 1, 2026, 4, 4, 2026},

		// 4. December start (startMonth=12)
		{"dec start: Dec 2025→Q1 FY2026", 12, 2025, 12, 1, 2026},
		{"dec start: Mar 2026→Q2 FY2026", 3, 2026, 12, 2, 2026},
		{"dec start: Jun 2026→Q3 FY2026", 6, 2026, 12, 3, 2026},
		{"dec start: Nov 2026→Q4 FY2026", 11, 2026, 12, 4, 2026},

		// 5. Invalid startMonth values → treated as 1
		{"invalid startMonth=0 → Q1", 1, 2026, 0, 1, 2026},
		{"invalid startMonth=13 → Q1", 1, 2026, 13, 1, 2026},
		{"invalid startMonth=-1 → Q1", 1, 2026, -1, 1, 2026},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQ, gotFY := FiscalQuarter(tt.month, tt.year, tt.startMonth)
			if gotQ != tt.wantQ || gotFY != tt.wantFY {
				t.Errorf("FiscalQuarter(%d, %d, %d) = (Q%d, FY%d), want (Q%d, FY%d)",
					tt.month, tt.year, tt.startMonth, gotQ, gotFY, tt.wantQ, tt.wantFY)
			}
		})
	}
}

func TestBoundaryConditions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		ref   time.Time
		opts  []Option
		want  time.Time
	}{
		{
			name:  "year boundary: tomorrow from Dec 31",
			input: "tomorrow",
			ref:   time.Date(2025, 12, 31, 10, 0, 0, 0, time.UTC),
			want:  date(2026, 1, 1),
		},
		{
			name:  "year boundary: yesterday from Jan 1",
			input: "yesterday",
			ref:   time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
			want:  date(2025, 12, 31),
		},
		{
			name:  "month boundary: in 1 month from Jan 31 (Go AddDate rollover)",
			input: "in 1 month",
			ref:   time.Date(2026, 1, 31, 12, 0, 0, 0, time.UTC),
			want:  date(2026, 3, 3),
		},
		{
			name:  "month boundary: end of month from Feb in leap year 2028",
			input: "end of month",
			ref:   time.Date(2028, 2, 10, 0, 0, 0, 0, time.UTC),
			want:  time.Date(2028, 2, 29, 23, 59, 59, 0, time.UTC),
		},
		{
			name:  "month boundary: end of month from Feb in non-leap year",
			input: "end of month",
			ref:   time.Date(2027, 2, 10, 0, 0, 0, 0, time.UTC),
			want:  time.Date(2027, 2, 28, 23, 59, 59, 0, time.UTC),
		},
		{
			name:  "zero offset: in 0 days",
			input: "in 0 days",
			ref:   time.Date(2026, 3, 18, 14, 30, 0, 0, time.UTC),
			want:  date(2026, 3, 18),
		},
		{
			name:  "zero offset: in 0 weeks",
			input: "in 0 weeks",
			ref:   time.Date(2026, 3, 18, 14, 30, 0, 0, time.UTC),
			want:  date(2026, 3, 18),
		},
		{
			name:  "week at year boundary: beginning of week from Jan 1 2026 (Thursday)",
			input: "beginning of week",
			ref:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			want:  date(2025, 12, 28),
		},
		{
			name:  "end of next year from Dec 2026",
			input: "end of next year",
			ref:   time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC),
			want:  time.Date(2027, 12, 31, 23, 59, 59, 0, time.UTC),
		},
		{
			name:  "beginning of last year from Jan 2026",
			input: "beginning of last year",
			ref:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			want:  date(2025, 1, 1),
		},
		{
			name:  "end of quarter from boundary month March 31",
			input: "end of quarter",
			ref:   time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC),
			want:  time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC),
		},
		{
			name:  "fiscal year start=12: end of quarter from March",
			input: "end of quarter",
			ref:   time.Date(2026, 3, 18, 14, 30, 0, 0, time.UTC),
			opts:  []Option{WithFiscalYearStart(12)},
			want:  time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC),
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

func TestContextCancellation(t *testing.T) {
	t.Parallel()
	t.Run("cancelled context returns error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := ParseRelativeContext(ctx, "tomorrow", ref)
		if err == nil {
			t.Error("expected error from cancelled context")
		}
	})

	t.Run("valid context works normally", func(t *testing.T) {
		ctx := context.Background()
		got, err := ParseRelativeContext(ctx, "tomorrow", ref)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := date(2026, 3, 19)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("ParseContext works", func(t *testing.T) {
		_, err := ParseContext(context.Background(), "tomorrow")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("ParseMultiContext works", func(t *testing.T) {
		got, err := ParseMultiContext(context.Background(), "every friday in april", ref)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 4 {
			t.Errorf("expected 4 fridays, got %d", len(got))
		}
	})

	t.Run("ParseMultiContext cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := ParseMultiContext(ctx, "tomorrow", ref)
		if err == nil {
			t.Error("expected error from cancelled context")
		}
	})
}

func TestTruncateDay(t *testing.T) {
	t.Parallel()
	input := time.Date(2026, 3, 18, 14, 30, 45, 123, time.UTC)
	got := TruncateDay(input)
	want := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("TruncateDay() = %v, want %v", got, want)
	}
	// Preserves location
	loc := time.FixedZone("EST", -5*3600)
	input2 := time.Date(2026, 3, 18, 14, 30, 0, 0, loc)
	got2 := TruncateDay(input2)
	if got2.Location() != loc {
		t.Errorf("TruncateDay should preserve location, got %v", got2.Location())
	}
}

func TestDaysIn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		year  int
		month time.Month
		want  int
	}{
		{2026, time.February, 28},
		{2028, time.February, 29}, // leap year
		{2026, time.March, 31},
		{2026, time.April, 30},
	}
	for _, tt := range tests {
		got := DaysIn(tt.year, tt.month, time.UTC)
		if got != tt.want {
			t.Errorf("DaysIn(%d, %s) = %d, want %d", tt.year, tt.month, got, tt.want)
		}
	}
}

func TestNonASCIIInput(t *testing.T) {
	t.Parallel()
	// Non-ASCII input should not panic and should return an error (not a recognized expression).
	// The key behavior: "mañana" should be lexed as a single unknown token, not split on ñ.
	inputs := []string{"mañana", "demain", "übermorgen", "日曜日"}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			_, err := ParseRelative(input, ref)
			if err == nil {
				t.Errorf("expected error for non-English input %q", input)
			}
		})
	}
}

func TestErrorPaths(t *testing.T) {
	t.Parallel()
	ref := time.Date(2026, 3, 18, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    string
		wantErr  bool
		useMulti bool // if true, test via ParseMulti instead of ParseRelative
	}{
		{
			name:    "0th ordinal is invalid",
			input:   "0th monday in march",
			wantErr: true,
		},
		{
			name:    "tab characters in input parse as whitespace",
			input:   "\tnext\tfriday",
			wantErr: false,
		},
		{
			name:    "negative number in offset",
			input:   "in -5 days",
			wantErr: true,
		},
		{
			name:    "incomplete boundary: end of",
			input:   "end of",
			wantErr: true,
		},
		{
			name:     "incomplete multi-date: every",
			input:    "every",
			wantErr:  true,
			useMulti: true,
		},
		{
			name:    "invalid unit after boundary: beginning of foo",
			input:   "beginning of foo",
			wantErr: true,
		},
		{
			name:    "double spaces still parse",
			input:   "next  friday",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.useMulti {
				_, err := ParseMulti(tt.input, ref)
				if tt.wantErr && err == nil {
					t.Errorf("ParseMulti(%q): expected error, got nil", tt.input)
				}
				if !tt.wantErr && err != nil {
					t.Errorf("ParseMulti(%q): unexpected error: %v", tt.input, err)
				}
			} else {
				_, err := ParseRelative(tt.input, ref)
				if tt.wantErr && err == nil {
					t.Errorf("ParseRelative(%q): expected error, got nil", tt.input)
				}
				if !tt.wantErr && err != nil {
					t.Errorf("ParseRelative(%q): unexpected error: %v", tt.input, err)
				}
			}
		})
	}
}

func TestCountWorkdays(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		start time.Time
		end   time.Time
		want  int
	}{
		{"same_day", date(2026, 3, 18), date(2026, 3, 18), 0},
		{"one_weekday", date(2026, 3, 18), date(2026, 3, 19), 1},
		{"over_weekend", date(2026, 3, 20), date(2026, 3, 23), 1}, // Fri to Mon
		{"full_week", date(2026, 3, 16), date(2026, 3, 23), 5},
		{"two_weeks", date(2026, 3, 16), date(2026, 3, 30), 10},
		{"reversed", date(2026, 3, 23), date(2026, 3, 16), 5},
		{"start_on_weekend", date(2026, 3, 21), date(2026, 3, 23), 0}, // Sat to Mon
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CountWorkdays(tt.start, tt.end)
			if got != tt.want {
				t.Errorf("CountWorkdays(%s, %s) = %d, want %d",
					tt.start.Format(DateLayout), tt.end.Format(DateLayout), got, tt.want)
			}
		})
	}
}

func TestModifierErrorContext(t *testing.T) {
	t.Parallel()
	_, err := ParseRelative("last pizza", ref)
	if err == nil {
		t.Fatal("expected error")
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *ParseError, got %T", err)
	}
	// Error should mention the consumed modifier for context
	msg := pe.Error()
	if !strings.Contains(msg, "last") {
		t.Errorf("error message should mention consumed modifier 'last', got: %s", msg)
	}
}

func TestTokenKindStringExhaustive(t *testing.T) {
	t.Parallel()
	// Every tokenKind from tokenNumber (0) through tokenEOF must have
	// an explicit String() case. If a new kind is added without updating
	// String(), this test catches it.
	allKinds := []tokenKind{
		tokenNumber, tokenWeekday, tokenMonth, tokenModifier,
		tokenPreposition, tokenUnit, tokenRelativeDay, tokenNamedTime,
		tokenMeridiem, tokenOrdinal, tokenBoundary, tokenEvery,
		tokenColon, tokenNoise, tokenUnknown, tokenEOF,
	}
	for _, k := range allKinds {
		name := k.String()
		if name == "" {
			t.Errorf("tokenKind(%d).String() returned empty string", k)
		}
	}
	// Verify count matches: if someone adds a new token between tokenNumber
	// and tokenEOF without adding it to allKinds above, the total won't match.
	if int(tokenEOF) != len(allKinds)-1 {
		t.Errorf("tokenKind enum has %d values but test covers %d — update the allKinds list",
			int(tokenEOF)+1, len(allKinds))
	}
}

func TestSemanticErrorOmitsPosition(t *testing.T) {
	t.Parallel()
	// "february 30" is syntactically valid but semantically invalid
	_, err := ParseRelative("february 30", ref)
	if err == nil {
		t.Fatal("expected error for feb 30")
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *ParseError, got %T", err)
	}
	if pe.Position >= 0 {
		t.Errorf("semantic error should have Position < 0, got %d", pe.Position)
	}
	if strings.Contains(pe.Error(), "position") {
		t.Errorf("semantic error message should not mention position: %s", pe.Error())
	}
}

func TestParseMultiFallbackError(t *testing.T) {
	t.Parallel()
	// "every pizza" fails multi-date grammar then falls back to single-date.
	// This test guards the contract: after fallback, bestErr must reflect the
	// single-date parse, not a stale error from the multi-date attempt.
	// Currently parseMultiDate never calls recordError, so this is a
	// forward-looking guard — if multi-date parsing is extended to record
	// errors, the p.bestErr = nil reset in ParseMultiContext becomes load-bearing.
	_, err := ParseMulti("every pizza", ref)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *ParseError, got %T", err)
	}
	// The error should come from the single-date parse, not the multi-date attempt.
	// "every" is tokenEvery, which parseDateExpr doesn't handle — so the error should
	// mention "date expression", not "weekday" from the multi-date path.
	for _, exp := range pe.Expected {
		if exp == "weekday" {
			t.Errorf("error references multi-date grammar (Expected contains %q); should reflect single-date parse", exp)
		}
	}
}

func TestNoPositionConstant(t *testing.T) {
	t.Parallel()
	// Semantic errors (like invalid day count) use NoPosition
	_, err := ParseRelative("february 30", ref)
	if err == nil {
		t.Fatal("expected error")
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *ParseError, got %T", err)
	}
	if pe.Position != NoPosition {
		t.Errorf("Position = %d, want NoPosition (%d)", pe.Position, NoPosition)
	}
}

func TestTimeExprErrorMessages(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		wantSub string // substring that should appear in the error
	}{
		{"hour>23 in 24h", "at 25:00", "time"},
		{"minute>59", "at 3:99", "time"},
		{"meridiem hour>12", "at 13pm", "time"},
		{"meridiem hour=0", "at 0pm", "time"},
		{"meridiem colon hour>12", "at 13:00pm", "time"},
		{"meridiem colon hour=0", "at 0:30am", "time"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseRelative(tt.input, ref)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", tt.input)
			}
			var pe *ParseError
			if !errors.As(err, &pe) {
				t.Fatalf("expected *ParseError, got %T", err)
			}
			errStr := pe.Error()
			if !strings.Contains(errStr, tt.wantSub) {
				t.Errorf("error %q does not contain %q", errStr, tt.wantSub)
			}
		})
	}
}
