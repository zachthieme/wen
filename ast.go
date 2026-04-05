package wen

import "time"

// Expr is a parsed date expression node. Concrete implementations are
// [RelativeDayExpr], [ModWeekdayExpr], [RelativeOffsetExpr],
// [CountedWeekdayExpr], [OrdinalWeekdayExpr], [LastWeekdayInMonthExpr],
// [AbsoluteDateExpr], [PeriodRefExpr], [BoundaryExpr], [MultiDateExpr],
// and [WithTimeExpr].
//
// The interface is sealed: only types defined in this package satisfy it.
// Use a type switch to inspect the concrete expression type.
type Expr interface {
	expr() // unexported marker — prevents external implementations
}

// RelativeDayExpr represents "today", "tomorrow", or "yesterday".
type RelativeDayExpr struct {
	Day string // "today", "tomorrow", "yesterday"
}

// ModWeekdayExpr represents "[modifier] weekday" (e.g., "next friday").
type ModWeekdayExpr struct {
	Modifier string       // "this", "next", "last", ""
	Weekday  time.Weekday
}

// RelativeOffsetExpr represents "N unit ago/from now" or "in N unit".
type RelativeOffsetExpr struct {
	N         int
	Unit      string // "day", "week", "month", "year", "hour", "minute"
	Direction int    // +1 forward, -1 backward
}

// CountedWeekdayExpr represents "in N weekdays" (e.g., "in 2 fridays").
type CountedWeekdayExpr struct {
	Count   int
	Weekday time.Weekday
}

// OrdinalWeekdayExpr represents "Nth weekday in/of month [year]"
// or "Nth weekday of next/last month".
type OrdinalWeekdayExpr struct {
	N             int
	Weekday       time.Weekday
	Month         time.Month // 0 when MonthModifier is set
	Year          int        // 0 = infer from ref
	MonthModifier string     // "this", "next", "last" — set when Month is 0
}

// LastWeekdayInMonthExpr represents "last weekday in month [year]".
type LastWeekdayInMonthExpr struct {
	Weekday time.Weekday
	Month   time.Month
	Year    int // 0 = infer from ref
}

// AbsoluteDateExpr represents "month day [year]" or "month year".
type AbsoluteDateExpr struct {
	Month time.Month
	Day   int
	Year  int // 0 = infer from ref
}

// PeriodRefExpr represents "this/next/last week/month".
type PeriodRefExpr struct {
	Modifier string // "this", "next", "last"
	Unit     string // "week", "month"
}

// BoundaryExpr represents "beginning/end of [modifier] unit".
type BoundaryExpr struct {
	Boundary string // "beginning", "end"
	Modifier string // "this", "next", "last"
	Unit     string // "week", "month", "quarter", "year"
}

// MultiDateExpr represents "every weekday in month [year]".
type MultiDateExpr struct {
	Weekday time.Weekday
	Month   time.Month
	Year    int // 0 = infer from ref
}

// WithTimeExpr wraps a date expression with a time-of-day.
// When Date is nil, the time applies to "today" (standalone time expression).
type WithTimeExpr struct {
	Date   Expr // nil for standalone time expressions
	Hour   int
	Minute int
}

// Compile-time interface satisfaction checks.
var (
	_ Expr = (*RelativeDayExpr)(nil)
	_ Expr = (*ModWeekdayExpr)(nil)
	_ Expr = (*RelativeOffsetExpr)(nil)
	_ Expr = (*CountedWeekdayExpr)(nil)
	_ Expr = (*OrdinalWeekdayExpr)(nil)
	_ Expr = (*LastWeekdayInMonthExpr)(nil)
	_ Expr = (*AbsoluteDateExpr)(nil)
	_ Expr = (*PeriodRefExpr)(nil)
	_ Expr = (*BoundaryExpr)(nil)
	_ Expr = (*MultiDateExpr)(nil)
	_ Expr = (*WithTimeExpr)(nil)
)

func (*RelativeDayExpr) expr()        {}
func (*ModWeekdayExpr) expr()         {}
func (*RelativeOffsetExpr) expr()     {}
func (*CountedWeekdayExpr) expr()     {}
func (*OrdinalWeekdayExpr) expr()     {}
func (*LastWeekdayInMonthExpr) expr() {}
func (*AbsoluteDateExpr) expr()       {}
func (*PeriodRefExpr) expr()          {}
func (*BoundaryExpr) expr()           {}
func (*MultiDateExpr) expr()          {}
func (*WithTimeExpr) expr()           {}
