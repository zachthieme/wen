package wen

import (
	"fmt"
	"time"
)

// resolver.go contains date resolution logic — the "semantic" side of parsing
// that converts recognized grammar patterns into concrete time.Time values.
// The grammar recognition (parse* methods) lives in parser.go.

// resolveModWeekday resolves a weekday token with an optional modifier
// ("this", "next", "last") to the corresponding date relative to ref.
func (p *parser) resolveModWeekday(modifier string, tok token) (time.Time, bool) {
	p.advance() // consume weekday token
	target := tok.Weekday
	var weekOffset int
	switch modifier {
	case "this", "":
		weekOffset = 0
	case "next":
		weekOffset = 1
	case "last":
		weekOffset = -1
	}
	return weekdayInWeek(p.ref, target, weekOffset), true
}

// weekdayInWeek returns the given weekday in the week offset from ref's week.
// Weeks run Sunday through Saturday.
func weekdayInWeek(ref time.Time, target time.Weekday, weekOffset int) time.Time {
	refDay := int(ref.Weekday())    // Sunday=0 ... Saturday=6
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
func (p *parser) resolveRelativeOffset(n int, unit string, direction int) (time.Time, bool) {
	amount := n * direction
	switch unit {
	case "day":
		return TruncateDay(p.ref).AddDate(0, 0, amount), true
	case "week":
		return TruncateDay(p.ref).AddDate(0, 0, amount*7), true
	case "month":
		return TruncateDay(p.ref).AddDate(0, amount, 0), true
	case "year":
		return TruncateDay(p.ref).AddDate(amount, 0, 0), true
	case "hour":
		return p.ref.Add(time.Duration(amount) * time.Hour), true
	case "minute":
		return p.ref.Add(time.Duration(amount) * time.Minute), true
	}
	p.recordError(p.makeError("day", "week", "month", "year", "hour", "minute"))
	return time.Time{}, false
}

// resolveCountedWeekday finds the Nth occurrence of a weekday after ref
// (e.g., "in 2 fridays" = the 2nd Friday after today).
func (p *parser) resolveCountedWeekday(count int, target time.Weekday) (time.Time, bool) {
	if count <= 0 {
		p.recordError(p.makeError("positive number"))
		return time.Time{}, false
	}
	// Start from the day after ref
	d := TruncateDay(p.ref).AddDate(0, 0, 1)
	// Find the first occurrence of the target weekday
	for d.Weekday() != target {
		d = d.AddDate(0, 0, 1)
	}
	// Advance by (count-1) weeks to get the Nth occurrence
	d = d.AddDate(0, 0, 7*(count-1))
	return d, true
}

// resolveOrdinalWeekdayInMonth finds the Nth occurrence of a weekday in the given
// month and year. Returns false if the month has fewer than N occurrences.
func (p *parser) resolveOrdinalWeekdayInMonth(n int, target time.Weekday, month time.Month, explicitYear int) (time.Time, bool) {
	year := explicitYear
	if year == 0 {
		year = p.ref.Year()
		if month < p.ref.Month() {
			year++
		}
	}

	// Find the first occurrence of target weekday in the month
	first := time.Date(year, month, 1, 0, 0, 0, 0, p.ref.Location())
	d := first
	for d.Weekday() != target {
		d = d.AddDate(0, 0, 1)
	}
	// Advance to the Nth occurrence
	d = d.AddDate(0, 0, 7*(n-1))

	// Verify still in the same month
	if d.Month() != month {
		// Count max occurrences of target weekday in this month
		maxOccurrences := 0
		c := first
		for c.Weekday() != target {
			c = c.AddDate(0, 0, 1)
		}
		for c.Month() == month {
			maxOccurrences++
			c = c.AddDate(0, 0, 7)
		}
		p.recordError(&ParseError{
			Input:    p.input,
			Position: 0,
			Expected: []string{fmt.Sprintf("%d or fewer", maxOccurrences)},
			Found:    fmt.Sprintf("%d", n),
		})
		return time.Time{}, false
	}
	return d, true
}

// resolveLastWeekdayInMonth finds the last occurrence of a weekday in the given
// month and year by scanning backward from the month's last day.
func (p *parser) resolveLastWeekdayInMonth(target time.Weekday, month time.Month, explicitYear int) (time.Time, bool) {
	year := explicitYear
	if year == 0 {
		year = p.ref.Year()
		if month < p.ref.Month() {
			year++
		}
	}

	// Start from last day of the month
	firstOfNext := time.Date(year, month+1, 1, 0, 0, 0, 0, p.ref.Location())
	lastDay := firstOfNext.AddDate(0, 0, -1)
	d := lastDay
	for d.Weekday() != target {
		d = d.AddDate(0, 0, -1)
	}
	return d, true
}

// resolvePeriodRef resolves "this/next/last week/month" using the configured
// PeriodMode (start-of-period vs same-day offset).
func (p *parser) resolvePeriodRef(modifier, unit string) (time.Time, bool) {
	ref := TruncateDay(p.ref)

	switch unit {
	case "week":
		dayOfWeek := int(ref.Weekday())
		sunday := ref.AddDate(0, 0, -dayOfWeek)

		switch modifier {
		case "this":
			return sunday, true
		case "next":
			if p.opts.periodMode == PeriodSame {
				return ref.AddDate(0, 0, 7), true
			}
			return sunday.AddDate(0, 0, 7), true
		case "last":
			if p.opts.periodMode == PeriodSame {
				return ref.AddDate(0, 0, -7), true
			}
			return sunday.AddDate(0, 0, -7), true
		}

	case "month":
		delta := modifierDelta(modifier)
		// PeriodSame uses AddDate to preserve day-of-month rollover semantics
		// (e.g., Jan 31 + 1 month = Mar 3).
		if p.opts.periodMode == PeriodSame && delta != 0 {
			return ref.AddDate(0, delta, 0), true
		}
		targetMonth, targetYear := shiftMonth(ref.Month(), ref.Year(), delta)
		return time.Date(targetYear, targetMonth, 1, 0, 0, 0, 0, ref.Location()), true
	}
	p.recordError(p.makeError("week", "month"))
	return time.Time{}, false
}

// resolveBoundary computes the beginning or end of a week, month, quarter, or
// year, adjusted by the modifier. Quarter boundaries respect fiscal year settings.
func (p *parser) resolveBoundary(boundary, modifier, unit string) (time.Time, bool) {
	ref := TruncateDay(p.ref)
	loc := ref.Location()

	switch unit {
	case "week":
		dayOfWeek := int(ref.Weekday())
		sunday := ref.AddDate(0, 0, -dayOfWeek)
		switch modifier {
		case "next":
			sunday = sunday.AddDate(0, 0, 7)
		case "last":
			sunday = sunday.AddDate(0, 0, -7)
		}
		if boundary == "beginning" {
			return sunday, true
		}
		// end = Saturday 23:59:59
		return sunday.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second), true

	case "month":
		targetMonth, targetYear := shiftMonth(ref.Month(), ref.Year(), modifierDelta(modifier))
		if boundary == "beginning" {
			return time.Date(targetYear, targetMonth, 1, 0, 0, 0, 0, loc), true
		}
		// end = last day 23:59:59
		firstOfNext := time.Date(targetYear, targetMonth+1, 1, 0, 0, 0, 0, loc)
		return firstOfNext.Add(-time.Second), true

	case "quarter":
		fyStart := p.opts.fiscalYearStart
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
		switch modifier {
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
		if boundary == "beginning" {
			return time.Date(year, startMonth, 1, 0, 0, 0, 0, loc), true
		}
		// end = last day of quarter 23:59:59
		endMonth := startMonth + monthsPerQuarter
		firstOfNext := time.Date(year, endMonth, 1, 0, 0, 0, 0, loc)
		return firstOfNext.Add(-time.Second), true

	case "year":
		year := ref.Year()
		switch modifier {
		case "next":
			year++
		case "last":
			year--
		}
		if boundary == "beginning" {
			return time.Date(year, time.January, 1, 0, 0, 0, 0, loc), true
		}
		// end = Dec 31 23:59:59
		return time.Date(year, time.December, 31, 23, 59, 59, 0, loc), true
	}
	p.recordError(p.makeError("week", "month", "quarter", "year"))
	return time.Time{}, false
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
