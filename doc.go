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
//
// # Grammar
//
// The parser accepts the following grammar (informal BNF). Noise words
// ("the", "a") are silently skipped between tokens.
//
//	input        = dateExpr [timeExpr] | timeExpr
//	dateExpr     = relativeDay
//	             | [modifier] weekday
//	             | modifier ("week" | "month")
//	             | "in" number (weekday | unit)
//	             | number unit ("ago" | "from" "now")
//	             | ordinal weekday [prep] (month [year] | modifier "month")
//	             | "last" weekday prep month [year]
//	             | month (day [year] | year)
//	             | boundary "of" [modifier] ("week" | "month" | "quarter" | "year")
//	multiExpr    = "every" weekday [prep] month [year]
//	timeExpr     = ["at"] (namedTime | number ":" number [meridiem] | number meridiem | "at" number)
//	relativeDay  = "today" | "tomorrow" | "yesterday"
//	modifier     = "this" | "next" | "last"
//	boundary     = "beginning" | "end"
//	prep         = "in" | "of"
//	namedTime    = "noon" | "midnight"
//	meridiem     = "am" | "pm"
//	ordinal      = "first" | "1st" | ... | "twelfth" | "12th"
//	number       = digit+ | "one" | "two" | ... | "thirty"
//	unit         = "day" | "week" | "month" | "quarter" | "year" | "hour" | "minute"
//	day          = number | ordinal
//	year         = number (> 31)
//	weekday      = "monday" | ... | "sunday"
//	month        = "january" | ... | "december"
//
// # Grammar Stability
//
// The grammar may be extended in minor versions (new keywords, new
// productions). Existing expressions will not change their resolved meaning
// in minor or patch releases. If a future grammar extension would alter
// the meaning of a previously valid expression, it will be gated behind a
// major version bump or an opt-in [Option].
//
// # Quick Start
//
// Parse a single date relative to now:
//
//	t, err := wen.Parse("next friday")
//
// Parse relative to a specific reference time:
//
//	ref := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)
//	t, err := wen.ParseRelative("in 3 days", ref)
//
// Parse with time-of-day:
//
//	t, err := wen.ParseRelative("march 25 at 3pm", ref)
//
// Enumerate multiple dates:
//
//	dates, err := wen.ParseMulti("every friday in april", ref)
//
// Use fiscal year quarters:
//
//	t, err := wen.ParseRelative("beginning of quarter", ref,
//	    wen.WithFiscalYearStart(10)) // Q1 starts in October
//
// All Context-accepting variants ([ParseContext], [ParseRelativeContext],
// [ParseMultiContext]) honor cancellation and deadlines.
package wen
