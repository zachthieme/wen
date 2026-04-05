package wen

import (
	"context"
	"fmt"
	"time"
)

// resolver.go contains date resolution logic — the "semantic" side of parsing
// that converts AST nodes (produced by the parser) into concrete time.Time values.
// The grammar recognition (parse* methods) lives in parser.go.

// resolver converts Expr AST nodes into time.Time values.
type resolver struct {
	ref   time.Time
	opts  options
	input string
	ctx   context.Context
}

func newResolver(ref time.Time, opts options, input string) *resolver {
	return &resolver{ref: ref, opts: opts, input: input}
}

// resolve dispatches an Expr to the appropriate resolution method and
// returns the resulting time.Time.
func (r *resolver) resolve(expr Expr) (time.Time, error) {
	switch e := expr.(type) {
	case *RelativeDayExpr:
		return r.resolveRelativeDay(e)
	case *ModWeekdayExpr:
		return r.resolveModWeekday(e)
	case *RelativeOffsetExpr:
		return r.resolveRelativeOffset(e)
	case *CountedWeekdayExpr:
		return r.resolveCountedWeekday(e)
	case *OrdinalWeekdayExpr:
		return r.resolveOrdinalWeekdayInMonth(e)
	case *LastWeekdayInMonthExpr:
		return r.resolveLastWeekdayInMonth(e)
	case *AbsoluteDateExpr:
		return r.resolveAbsoluteDate(e)
	case *PeriodRefExpr:
		return r.resolvePeriodRef(e)
	case *BoundaryExpr:
		return r.resolveBoundary(e)
	case *WithTimeExpr:
		return r.resolveWithTime(e)
	default:
		return time.Time{}, &ParseError{
			Input:    r.input,
			Position: -1,
			Expected: []string{"date expression"},
			Found:    fmt.Sprintf("%T", expr),
		}
	}
}

// resolveMulti handles expressions that may produce multiple dates.
// For [MultiDateExpr] it returns all occurrences; otherwise it wraps a single result.
func (r *resolver) resolveMulti(expr Expr) ([]time.Time, error) {
	if r.ctx != nil {
		if err := r.ctx.Err(); err != nil {
			return nil, &ParseError{
				Input:    r.input,
				Position: -1,
				Expected: []string{"date expression"},
				Cause:    err,
			}
		}
	}
	if e, ok := expr.(*MultiDateExpr); ok {
		return r.resolveMultiDate(e)
	}
	t, err := r.resolve(expr)
	if err != nil {
		return nil, err
	}
	return []time.Time{t}, nil
}

func (r *resolver) resolveRelativeDay(e *RelativeDayExpr) (time.Time, error) {
	base := TruncateDay(r.ref)
	switch e.Day {
	case "today":
		return base, nil
	case "tomorrow":
		return base.AddDate(0, 0, 1), nil
	case "yesterday":
		return base.AddDate(0, 0, -1), nil
	}
	return time.Time{}, &ParseError{
		Input:    r.input,
		Position: -1,
		Expected: []string{"today", "tomorrow", "yesterday"},
		Found:    e.Day,
	}
}

// resolveModWeekday resolves a weekday with an optional modifier
// ("this", "next", "last") to the corresponding date relative to ref.
func (r *resolver) resolveModWeekday(e *ModWeekdayExpr) (time.Time, error) {
	var weekOffset int
	switch e.Modifier {
	case "this", "":
		weekOffset = 0
	case "next":
		weekOffset = 1
	case "last":
		weekOffset = -1
	}
	return weekdayInWeek(r.ref, e.Weekday, weekOffset), nil
}

// weekdayInWeek returns the given weekday in the week offset from ref's week.
// Weeks run Sunday through Saturday.
func weekdayInWeek(ref time.Time, target time.Weekday, weekOffset int) time.Time {
	refDay := int(ref.Weekday()) // Sunday=0 ... Saturday=6
	targetDay := int(target)
	// Find Sunday of ref's week
	sunday := TruncateDay(ref).AddDate(0, 0, -refDay)
	// Apply week offset
	sunday = sunday.AddDate(0, 0, 7*weekOffset)
	// Find target day in that week
	return sunday.AddDate(0, 0, targetDay)
}

// resolveRelativeOffset applies an offset of n units (day, week, month, year, hour, minute)
// in the given direction (+1 forward, -1 backward) relative to ref.
func (r *resolver) resolveRelativeOffset(e *RelativeOffsetExpr) (time.Time, error) {
	amount := e.N * e.Direction
	switch e.Unit {
	case "day":
		return TruncateDay(r.ref).AddDate(0, 0, amount), nil
	case "week":
		return TruncateDay(r.ref).AddDate(0, 0, amount*7), nil
	case "month":
		return TruncateDay(r.ref).AddDate(0, amount, 0), nil
	case "year":
		return TruncateDay(r.ref).AddDate(amount, 0, 0), nil
	case "hour":
		return r.ref.Add(time.Duration(amount) * time.Hour), nil
	case "minute":
		return r.ref.Add(time.Duration(amount) * time.Minute), nil
	}
	return time.Time{}, &ParseError{
		Input:    r.input,
		Position: -1,
		Expected: []string{"day", "week", "month", "year", "hour", "minute"},
		Found:    e.Unit,
	}
}

// resolveCountedWeekday finds the Nth occurrence of a weekday after ref
// (e.g., "in 2 fridays" = the 2nd Friday after today).
func (r *resolver) resolveCountedWeekday(e *CountedWeekdayExpr) (time.Time, error) {
	if e.Count <= 0 {
		return time.Time{}, &ParseError{
			Input:    r.input,
			Position: -1,
			Expected: []string{"positive number"},
			Found:    fmt.Sprintf("%d", e.Count),
		}
	}
	// Start from the day after ref
	d := TruncateDay(r.ref).AddDate(0, 0, 1)
	// Find the first occurrence of the target weekday
	for d.Weekday() != e.Weekday {
		d = d.AddDate(0, 0, 1)
	}
	// Advance by (count-1) weeks to get the Nth occurrence
	d = d.AddDate(0, 0, 7*(e.Count-1))
	return d, nil
}

// resolveOrdinalWeekdayInMonth finds the Nth occurrence of a weekday in the given
// month and year. Returns an error if the month has fewer than N occurrences.
func (r *resolver) resolveOrdinalWeekdayInMonth(e *OrdinalWeekdayExpr) (time.Time, error) {
	month := e.Month
	year := e.Year

	if e.MonthModifier != "" {
		// "first monday of next month" — compute month/year from modifier
		targetMonth, targetYear := shiftMonth(r.ref.Month(), r.ref.Year(), modifierDelta(e.MonthModifier))
		month = targetMonth
		year = targetYear
	} else if year == 0 {
		year = r.ref.Year()
		if month < r.ref.Month() {
			year++
		}
	}

	// Find the first occurrence of target weekday in the month
	first := time.Date(year, month, 1, 0, 0, 0, 0, r.ref.Location())
	d := first
	for d.Weekday() != e.Weekday {
		d = d.AddDate(0, 0, 1)
	}
	// Advance to the Nth occurrence
	d = d.AddDate(0, 0, 7*(e.N-1))

	// Verify still in the same month
	if d.Month() != month {
		// Count max occurrences of target weekday in this month
		maxOccurrences := 0
		c := first
		for c.Weekday() != e.Weekday {
			c = c.AddDate(0, 0, 1)
		}
		for c.Month() == month {
			maxOccurrences++
			c = c.AddDate(0, 0, 7)
		}
		return time.Time{}, &ParseError{
			Input:    r.input,
			Position: -1,
			Expected: []string{fmt.Sprintf("%d or fewer", maxOccurrences)},
			Found:    fmt.Sprintf("%d", e.N),
		}
	}
	return d, nil
}

// resolveLastWeekdayInMonth finds the last occurrence of a weekday in the given
// month and year by scanning backward from the month's last day.
func (r *resolver) resolveLastWeekdayInMonth(e *LastWeekdayInMonthExpr) (time.Time, error) {
	year := e.Year
	if year == 0 {
		year = r.ref.Year()
		if e.Month < r.ref.Month() {
			year++
		}
	}

	// Start from last day of the month
	firstOfNext := time.Date(year, e.Month+1, 1, 0, 0, 0, 0, r.ref.Location())
	lastDay := firstOfNext.AddDate(0, 0, -1)
	d := lastDay
	for d.Weekday() != e.Weekday {
		d = d.AddDate(0, 0, -1)
	}
	return d, nil
}

// resolveAbsoluteDate resolves "month day [year]" or "month year" patterns,
// inferring year and validating the day.
func (r *resolver) resolveAbsoluteDate(e *AbsoluteDateExpr) (time.Time, error) {
	year := e.Year
	if year == 0 {
		year = r.ref.Year()
		if e.Month < r.ref.Month() {
			year++
		}
	}

	maxDay := DaysIn(year, e.Month, r.ref.Location())
	if e.Day < 1 || e.Day > maxDay {
		return time.Time{}, &ParseError{
			Input:    r.input,
			Position: -1,
			Expected: []string{fmt.Sprintf("day between 1 and %d", maxDay)},
			Found:    fmt.Sprintf("%d", e.Day),
		}
	}

	return time.Date(year, e.Month, e.Day, 0, 0, 0, 0, r.ref.Location()), nil
}

// resolvePeriodRef resolves "this/next/last week/month" using the configured
// PeriodMode (start-of-period vs same-day offset).
func (r *resolver) resolvePeriodRef(e *PeriodRefExpr) (time.Time, error) {
	ref := TruncateDay(r.ref)

	switch e.Unit {
	case "week":
		dayOfWeek := int(ref.Weekday())
		sunday := ref.AddDate(0, 0, -dayOfWeek)

		switch e.Modifier {
		case "this":
			return sunday, nil
		case "next":
			if r.opts.periodMode == PeriodSame {
				return ref.AddDate(0, 0, 7), nil
			}
			return sunday.AddDate(0, 0, 7), nil
		case "last":
			if r.opts.periodMode == PeriodSame {
				return ref.AddDate(0, 0, -7), nil
			}
			return sunday.AddDate(0, 0, -7), nil
		}

	case "month":
		delta := modifierDelta(e.Modifier)
		// PeriodSame uses AddDate to preserve day-of-month rollover semantics
		// (e.g., Jan 31 + 1 month = Mar 3).
		if r.opts.periodMode == PeriodSame && delta != 0 {
			return ref.AddDate(0, delta, 0), nil
		}
		targetMonth, targetYear := shiftMonth(ref.Month(), ref.Year(), delta)
		return time.Date(targetYear, targetMonth, 1, 0, 0, 0, 0, ref.Location()), nil
	}
	return time.Time{}, &ParseError{
		Input:    r.input,
		Position: -1,
		Expected: []string{"week", "month"},
		Found:    e.Unit,
	}
}

// resolveBoundary computes the beginning or end of a week, month, quarter, or
// year, adjusted by the modifier. Quarter boundaries respect fiscal year settings.
func (r *resolver) resolveBoundary(e *BoundaryExpr) (time.Time, error) {
	ref := TruncateDay(r.ref)
	loc := ref.Location()

	switch e.Unit {
	case "week":
		dayOfWeek := int(ref.Weekday())
		sunday := ref.AddDate(0, 0, -dayOfWeek)
		switch e.Modifier {
		case "next":
			sunday = sunday.AddDate(0, 0, 7)
		case "last":
			sunday = sunday.AddDate(0, 0, -7)
		}
		if e.Boundary == "beginning" {
			return sunday, nil
		}
		// end = Saturday 23:59:59
		return sunday.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second), nil

	case "month":
		targetMonth, targetYear := shiftMonth(ref.Month(), ref.Year(), modifierDelta(e.Modifier))
		if e.Boundary == "beginning" {
			return time.Date(targetYear, targetMonth, 1, 0, 0, 0, 0, loc), nil
		}
		// end = last day 23:59:59
		firstOfNext := time.Date(targetYear, targetMonth+1, 1, 0, 0, 0, 0, loc)
		return firstOfNext.Add(-time.Second), nil

	case "quarter":
		fyStart := r.opts.fiscalYearStart
		if fyStart < 1 || fyStart > 12 {
			fyStart = 1
		}
		// Determine which fiscal year the ref falls in.
		// The fiscal year "starts" in calendar year fyYear.
		fyYear := ref.Year()
		if int(ref.Month()) < fyStart {
			fyYear--
		}
		// Fiscal quarter index (0-3) within this fiscal year.
		fiscalMonth := (int(ref.Month()) - fyStart + monthsPerYear) % monthsPerYear
		quarterIdx := fiscalMonth / monthsPerQuarter
		switch e.Modifier {
		case "next":
			quarterIdx++
		case "last":
			quarterIdx--
		}
		// Convert back to calendar month/year.
		// Each fiscal quarter starts at fyStart + quarterIdx*monthsPerQuarter months from the FY start year.
		totalMonths := (fyStart - 1) + quarterIdx*monthsPerQuarter // 0-indexed calendar month
		year := fyYear
		for totalMonths < 0 {
			totalMonths += monthsPerYear
			year--
		}
		for totalMonths >= monthsPerYear {
			totalMonths -= monthsPerYear
			year++
		}
		startMonth := time.Month(totalMonths + 1)
		if e.Boundary == "beginning" {
			return time.Date(year, startMonth, 1, 0, 0, 0, 0, loc), nil
		}
		// end = last day of quarter 23:59:59
		endMonth := startMonth + monthsPerQuarter
		firstOfNext := time.Date(year, endMonth, 1, 0, 0, 0, 0, loc)
		return firstOfNext.Add(-time.Second), nil

	case "year":
		year := ref.Year()
		switch e.Modifier {
		case "next":
			year++
		case "last":
			year--
		}
		if e.Boundary == "beginning" {
			return time.Date(year, time.January, 1, 0, 0, 0, 0, loc), nil
		}
		// end = Dec 31 23:59:59
		return time.Date(year, time.December, 31, 23, 59, 59, 0, loc), nil
	}
	return time.Time{}, &ParseError{
		Input:    r.input,
		Position: -1,
		Expected: []string{"week", "month", "quarter", "year"},
		Found:    e.Unit,
	}
}

// resolveMultiDate enumerates all occurrences of a weekday in the given month.
func (r *resolver) resolveMultiDate(e *MultiDateExpr) ([]time.Time, error) {
	year := e.Year
	if year == 0 {
		year = r.ref.Year()
		if e.Month < r.ref.Month() {
			year++
		}
	}

	loc := r.ref.Location()
	first := time.Date(year, e.Month, 1, 0, 0, 0, 0, loc)
	d := first
	for d.Weekday() != e.Weekday {
		d = d.AddDate(0, 0, 1)
	}
	var results []time.Time
	for d.Month() == e.Month {
		results = append(results, d)
		d = d.AddDate(0, 0, 7)
	}
	return results, nil
}

// resolveWithTime resolves the inner date expression (or uses today if nil)
// and applies the time-of-day.
func (r *resolver) resolveWithTime(e *WithTimeExpr) (time.Time, error) {
	var base time.Time
	if e.Date == nil {
		base = TruncateDay(r.ref)
	} else {
		var err error
		base, err = r.resolve(e.Date)
		if err != nil {
			return time.Time{}, err
		}
	}
	return setTime(base, e.Hour, e.Minute), nil
}

// setTime returns base with the time-of-day set to hour:min.
func setTime(base time.Time, hour, min int) time.Time {
	return time.Date(base.Year(), base.Month(), base.Day(), hour, min, 0, 0, base.Location())
}

// applyMeridiem converts a 12-hour clock hour with an "am"/"pm" suffix to 24-hour format.
func applyMeridiem(hour int, meridiem string) int {
	if meridiem == "am" {
		if hour == 12 {
			return 0
		}
		return hour
	}
	// pm
	if hour == 12 {
		return 12
	}
	return hour + 12
}
