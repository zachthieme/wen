package wen

import "time"

// dateExpr is the internal AST node interface. Each concrete type represents
// a distinct production in the date grammar. The parser produces these nodes;
// the resolver converts them to time.Time values.
type dateExpr interface {
	dateExpr() // marker method
}

// relativeDayExpr represents "today", "tomorrow", or "yesterday".
type relativeDayExpr struct {
	Day string // "today", "tomorrow", "yesterday"
}

// modWeekdayExpr represents "[modifier] weekday" (e.g., "next friday").
type modWeekdayExpr struct {
	Modifier string       // "this", "next", "last", ""
	Weekday  time.Weekday
}

// relativeOffsetExpr represents "N unit ago/from now" or "in N unit".
type relativeOffsetExpr struct {
	N         int
	Unit      string // "day", "week", "month", "year", "hour", "minute"
	Direction int    // +1 forward, -1 backward
}

// countedWeekdayExpr represents "in N weekdays" (e.g., "in 2 fridays").
type countedWeekdayExpr struct {
	Count   int
	Weekday time.Weekday
}

// ordinalWeekdayExpr represents "Nth weekday in/of month [year]"
// or "Nth weekday of next/last month".
type ordinalWeekdayExpr struct {
	N             int
	Weekday       time.Weekday
	Month         time.Month // 0 when MonthModifier is set
	Year          int        // 0 = infer from ref
	MonthModifier string     // "this", "next", "last" — set when Month is 0
}

// lastWeekdayInMonthExpr represents "last weekday in month [year]".
type lastWeekdayInMonthExpr struct {
	Weekday time.Weekday
	Month   time.Month
	Year    int // 0 = infer from ref
}

// absoluteDateExpr represents "month day [year]" or "month year".
type absoluteDateExpr struct {
	Month time.Month
	Day   int
	Year  int // 0 = infer from ref
}

// periodRefExpr represents "this/next/last week/month".
type periodRefExpr struct {
	Modifier string // "this", "next", "last"
	Unit     string // "week", "month"
}

// boundaryExpr represents "beginning/end of [modifier] unit".
type boundaryExpr struct {
	Boundary string // "beginning", "end"
	Modifier string // "this", "next", "last"
	Unit     string // "week", "month", "quarter", "year"
}

// multiDateExpr represents "every weekday in month [year]".
type multiDateExpr struct {
	Weekday time.Weekday
	Month   time.Month
	Year    int // 0 = infer from ref
}

// withTimeExpr wraps a date expression with a time-of-day.
// When Date is nil, the time applies to "today" (standalone time expression).
type withTimeExpr struct {
	Date   dateExpr // nil for standalone time expressions
	Hour   int
	Minute int
}

func (*relativeDayExpr) dateExpr()        {}
func (*modWeekdayExpr) dateExpr()         {}
func (*relativeOffsetExpr) dateExpr()     {}
func (*countedWeekdayExpr) dateExpr()     {}
func (*ordinalWeekdayExpr) dateExpr()     {}
func (*lastWeekdayInMonthExpr) dateExpr() {}
func (*absoluteDateExpr) dateExpr()       {}
func (*periodRefExpr) dateExpr()          {}
func (*boundaryExpr) dateExpr()           {}
func (*multiDateExpr) dateExpr()          {}
func (*withTimeExpr) dateExpr()           {}
