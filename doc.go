// Package wen parses natural language date and time expressions into [time.Time] values.
//
// It supports relative expressions ("tomorrow", "next friday", "in 3 days"),
// absolute dates ("march 15 2027"), ordinal patterns ("first monday of april"),
// time-of-day ("at 3pm"), boundaries ("end of next quarter"), and multi-date
// expressions ("every friday in april"). All parsing is relative to a reference
// time, defaulting to [time.Now].
//
// The parser is zero-dependency beyond the Go standard library. Configure
// behavior with functional [Option] values such as [WithFiscalYearStart] and
// [WithPeriodSame].
package wen
