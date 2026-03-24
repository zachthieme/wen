package wen

import (
	"testing"
	"time"
)

var benchRef = time.Date(2026, 3, 18, 14, 30, 0, 0, time.UTC)

func BenchmarkParse(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{"relative_day", "tomorrow"},
		{"mod_weekday", "next friday"},
		{"last_weekday", "last monday"},
		{"offset_days", "in 3 days"},
		{"offset_ago", "2 weeks ago"},
		{"absolute_date", "march 25 2026"},
		{"ordinal_weekday", "first monday of april"},
		{"last_weekday_in_month", "last friday in november"},
		{"boundary", "end of next quarter"},
		{"time_expression", "tomorrow at 3pm"},
		{"time_colon", "tomorrow at 12:30"},
		{"named_time", "tomorrow at noon"},
		{"period_next_week", "next week"},
		{"period_last_month", "last month"},
		{"cardinal_words", "in five days"},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for b.Loop() {
				_, _ = ParseRelative(bm.input, benchRef)
			}
		})
	}
}

func BenchmarkParseMulti(b *testing.B) {
	b.Run("every_friday_in_april", func(b *testing.B) {
		for b.Loop() {
			_, _ = ParseMulti("every friday in april", benchRef)
		}
	})
}

func BenchmarkLexer(b *testing.B) {
	b.Run("complex_input", func(b *testing.B) {
		for b.Loop() {
			l := newLexer("first monday of next month at 3pm")
			l.tokenize()
		}
	})
}
