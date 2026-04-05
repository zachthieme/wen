package wen

import (
	"testing"
	"time"
)

var fuzzRef = time.Date(2026, 3, 18, 14, 30, 0, 0, time.UTC)

func FuzzParseRelative(f *testing.F) {
	// Seed with valid expressions covering each grammar branch.
	seeds := []string{
		"today",
		"tomorrow",
		"yesterday",
		"next friday",
		"last monday",
		"this wednesday",
		"in 3 days",
		"in 2 weeks",
		"5 months ago",
		"2 years from now",
		"first monday of april",
		"third wednesday in march",
		"last friday in november",
		"march 15 2027",
		"december 25",
		"january 2027",
		"beginning of next month",
		"end of this quarter",
		"end of last year",
		"next week",
		"last month",
		"tomorrow at 3pm",
		"at noon",
		"at 15:30",
		"at midnight",
		"every friday in april",
		// Edge cases and malformed input
		"",
		"pizza",
		"next",
		"in",
		"3",
		"the the the",
		"mañana",
		"in -5 days",
		"0th monday of march",
		"february 30",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(_ *testing.T, input string) {
		// Must not panic on any input.
		_, _ = ParseRelative(input, fuzzRef)
	})
}

func FuzzLexer(f *testing.F) {
	seeds := []string{
		"tomorrow at 3pm",
		"first monday of next month",
		"12:30am",
		"",
		"🎉 party time",
		"in\t5\ndays",
		"marchhhh",
		"1st 2nd 3rd 4th",
		"999999999999999999999999",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(_ *testing.T, input string) {
		// Must not panic on any input.
		l := newLexer(input)
		l.tokenize()
	})
}

func FuzzParseMulti(f *testing.F) {
	seeds := []string{
		"every friday in april",
		"every monday in january 2027",
		"tomorrow",
		"",
		"every",
		"every pizza",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(_ *testing.T, input string) {
		// Must not panic on any input.
		_, _ = ParseMulti(input, fuzzRef)
	})
}
