package wen

import "time"

// tokenKind classifies a lexed token.
type tokenKind int

const (
	tokenNumber      tokenKind = iota // "3", "15", "2025"
	tokenWeekday                      // "monday" through "sunday"
	tokenMonth                        // "january" through "december"
	tokenModifier                     // "this", "next", "last"
	tokenPreposition                  // "in", "from", "of", "at", "ago", "now"
	tokenUnit                         // "day", "week", "month", "year", "hour", "minute"
	tokenRelativeDay                  // "today", "tomorrow", "yesterday"
	tokenNamedTime                    // "noon", "midnight"
	tokenMeridiem                     // "am", "pm"
	tokenOrdinal                      // "1st", "first", etc. — IntVal holds numeric value
	tokenBoundary                     // "beginning", "end"
	tokenEvery                        // "every"
	tokenColon                        // ":"
	tokenNoise                        // "the", "a" — skipped by parser
	tokenUnknown                      // unrecognized word
	tokenEOF                          // end of input
)

func (k tokenKind) String() string {
	switch k {
	case tokenNumber:
		return "number"
	case tokenWeekday:
		return "weekday"
	case tokenMonth:
		return "month"
	case tokenModifier:
		return "modifier"
	case tokenPreposition:
		return "preposition"
	case tokenUnit:
		return "unit"
	case tokenRelativeDay:
		return "relative day"
	case tokenNamedTime:
		return "named time"
	case tokenMeridiem:
		return "meridiem"
	case tokenOrdinal:
		return "ordinal"
	case tokenBoundary:
		return "boundary"
	case tokenEvery:
		return "every"
	case tokenColon:
		return "colon"
	case tokenNoise:
		return "noise"
	case tokenUnknown:
		return "unknown"
	case tokenEOF:
		return "end of input"
	default:
		return "unknown"
	}
}

// token is a single lexed token with its classification, value, and position.
//
// Field validity by kind:
//   - tokenNumber:   IntVal (numeric value)
//   - tokenOrdinal:  IntVal (numeric value, e.g. 1 for "1st")
//   - tokenWeekday:  Weekday
//   - tokenMonth:    Month
//   - All kinds:     Value (lowercased text), Position (byte offset in input)
//
// Fields not listed for a given kind are zero-valued and should not be read.
type token struct {
	Kind     tokenKind
	Value    string       // original text (lowercased)
	IntVal   int          // numeric value for tokenNumber, tokenOrdinal
	Weekday  time.Weekday // set for tokenWeekday
	Month    time.Month   // set for tokenMonth
	Position int          // byte offset in original input
}
