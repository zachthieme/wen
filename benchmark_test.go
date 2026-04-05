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
			b.ReportAllocs()
			for b.Loop() {
				_, _ = ParseRelative(bm.input, benchRef)
			}
		})
	}
}

func BenchmarkParseMulti(b *testing.B) {
	b.Run("every_friday_in_april", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = ParseMulti("every friday in april", benchRef)
		}
	})
	b.Run("fallback_to_single", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = ParseMulti("tomorrow", benchRef)
		}
	})
}

func BenchmarkLexer(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{"simple", "tomorrow"},
		{"moderate", "next friday at 3pm"},
		{"complex", "first monday of next month at 3pm"},
		{"long_with_noise", "the beginning of the next quarter"},
		{"numeric_heavy", "12:30pm"},
		{"ordinal_suffix", "march 3rd 2027"},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				l := newLexer(bm.input)
				l.tokenize()
			}
		})
	}
}

func BenchmarkCountWorkdays(b *testing.B) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	b.Run("full_year", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			CountWorkdays(start, end)
		}
	})
	b.Run("single_week", func(b *testing.B) {
		b.ReportAllocs()
		weekEnd := start.AddDate(0, 0, 7)
		for b.Loop() {
			CountWorkdays(start, weekEnd)
		}
	})
}

func BenchmarkFiscalQuarter(b *testing.B) {
	b.Run("calendar_year", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			FiscalQuarter(3, 2026, 1)
		}
	})
	b.Run("oct_fiscal_year", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			FiscalQuarter(3, 2026, 10)
		}
	})
}
