package wen

import (
	"fmt"
	"time"
)

// maxDayOfMonth is the maximum valid day number; values above this are treated as years.
const maxDayOfMonth = 31

type parser struct {
	tokens  []token
	pos     int
	ref     time.Time
	opts    options
	bestErr *ParseError
}

func newParser(tokens []token, ref time.Time, opts options) *parser {
	return &parser{tokens: tokens, ref: ref, opts: opts}
}

func (p *parser) peek() token     { return p.tokens[p.pos] }
func (p *parser) advance() token  { t := p.tokens[p.pos]; p.pos++; return t }
func (p *parser) save() int       { return p.pos }
func (p *parser) restore(pos int) { p.pos = pos }

func (p *parser) skipNoise() {
	for p.pos < len(p.tokens) && p.tokens[p.pos].Kind == tokenNoise {
		p.pos++
	}
}

func (p *parser) makeError(expected ...string) *ParseError {
	tok := p.peek()
	return &ParseError{
		Position: tok.Position,
		Expected: expected,
		Found:    tok.Value,
	}
}

func (p *parser) recordError(err *ParseError) {
	if p.bestErr == nil || err.Position >= p.bestErr.Position {
		p.bestErr = err
	}
}

func (p *parser) finalError() *ParseError {
	if p.bestErr != nil {
		return p.bestErr
	}
	return p.makeError("date or time expression")
}

func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func daysIn(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}

func (p *parser) parse() (time.Time, error) {
	p.skipNoise()

	result, ok := p.parseDateExpr()
	if ok {
		// Try optional trailing time expression
		saved := p.save()
		p.skipNoise()
		if t, matched := p.parseTimeExpr(result); matched {
			result = t
		} else {
			p.restore(saved)
		}
	} else {
		// Try standalone time expression
		p.pos = 0
		p.skipNoise()
		var matched bool
		result, matched = p.parseTimeExpr(truncateDay(p.ref))
		if !matched {
			return time.Time{}, p.finalError()
		}
	}

	p.skipNoise()
	if p.peek().Kind != tokenEOF {
		return time.Time{}, p.makeError("end of input")
	}
	return result, nil
}

func (p *parser) parseDateExpr() (time.Time, bool) {
	p.skipNoise()
	tok := p.peek()

	switch tok.Kind {
	case tokenRelativeDay:
		return p.parseRelativeDay()
	case tokenWeekday:
		return p.resolveModWeekday("this", tok)
	case tokenModifier:
		return p.parseModifierExpr()
	case tokenPreposition:
		if tok.Value == "in" {
			return p.parseInExpr()
		}
	case tokenNumber:
		return p.parseNumberLeadExpr()
	case tokenOrdinal:
		return p.parseOrdinalWeekdayInMonth()
	case tokenMonth:
		return p.parseAbsoluteDate()
	case tokenBoundary:
		return p.parseBoundaryExpr()
	}

	p.recordError(p.makeError("date expression"))
	return time.Time{}, false
}

func (p *parser) parseRelativeDay() (time.Time, bool) {
	tok := p.advance()
	base := truncateDay(p.ref)
	switch tok.Value {
	case "today":
		return base, true
	case "tomorrow":
		return base.AddDate(0, 0, 1), true
	case "yesterday":
		return base.AddDate(0, 0, -1), true
	}
	return time.Time{}, false
}

func (p *parser) parseModifierExpr() (time.Time, bool) {
	saved := p.save()
	modifier := p.advance()
	p.skipNoise()

	tok := p.peek()
	if tok.Kind == tokenWeekday {
		// Peek further: is this "last friday in november" or "last friday"?
		weekdaySaved := p.save()
		p.advance() // consume weekday
		p.skipNoise()
		next := p.peek()
		if next.Kind == tokenPreposition && (next.Value == "in" || next.Value == "of") {
			p.advance() // consume "in"/"of"
			p.skipNoise()
			if p.peek().Kind == tokenMonth {
				monthTok := p.advance()
				if modifier.Value == "last" {
					// Optional year
					p.skipNoise()
					year := 0
					if p.peek().Kind == tokenNumber && p.peek().IntVal > maxDayOfMonth {
						year = p.advance().IntVal
					}
					return p.resolveLastWeekdayInMonth(tok.Weekday, monthTok.Month, year)
				}
			}
		}
		p.restore(weekdaySaved)
		return p.resolveModWeekday(modifier.Value, tok)
	}
	if tok.Kind == tokenUnit && (tok.Value == "week" || tok.Value == "month") {
		p.advance() // consume unit
		return p.resolvePeriodRef(modifier.Value, tok.Value)
	}

	p.recordError(p.makeError("weekday", "week", "month"))
	p.restore(saved)
	return time.Time{}, false
}

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
	refDow := int(ref.Weekday())    // Sunday=0 ... Saturday=6
	targetDow := int(target)
	// Find Sunday of ref's week
	sunday := truncateDay(ref).AddDate(0, 0, -refDow)
	// Apply week offset
	sunday = sunday.AddDate(0, 0, 7*weekOffset)
	// Find target day in that week
	return sunday.AddDate(0, 0, targetDow)
}

func (p *parser) parseInExpr() (time.Time, bool) {
	saved := p.save()
	p.advance() // consume "in"
	p.skipNoise()

	if p.peek().Kind != tokenNumber {
		err := p.makeError("number")
		p.restore(saved)
		p.recordError(err)
		return time.Time{}, false
	}
	num := p.advance()
	p.skipNoise()

	next := p.peek()
	switch next.Kind {
	case tokenWeekday:
		p.advance() // consume weekday
		return p.resolveCountedWeekday(num.IntVal, next.Weekday)
	case tokenUnit:
		p.advance()
		return p.resolveRelativeOffset(num.IntVal, next.Value, 1)
	}

	p.recordError(p.makeError("weekday", "unit"))
	p.restore(saved)
	return time.Time{}, false
}

func (p *parser) parseNumberLeadExpr() (time.Time, bool) {
	saved := p.save()
	num := p.advance()
	p.skipNoise()

	if p.peek().Kind != tokenUnit {
		err := p.makeError("unit")
		p.restore(saved)
		p.recordError(err)
		return time.Time{}, false
	}
	unit := p.advance()
	p.skipNoise()

	tok := p.peek()
	if tok.Kind == tokenPreposition && tok.Value == "ago" {
		p.advance()
		return p.resolveRelativeOffset(num.IntVal, unit.Value, -1)
	}
	if tok.Kind == tokenPreposition && tok.Value == "from" {
		p.advance()
		p.skipNoise()
		if p.peek().Kind == tokenPreposition && p.peek().Value == "now" {
			p.advance()
			return p.resolveRelativeOffset(num.IntVal, unit.Value, 1)
		}
		err := p.makeError("now")
		p.restore(saved)
		p.recordError(err)
		return time.Time{}, false
	}

	err := p.makeError("ago", "from")
	p.restore(saved)
	p.recordError(err)
	return time.Time{}, false
}

func (p *parser) resolveRelativeOffset(n int, unit string, direction int) (time.Time, bool) {
	amount := n * direction
	switch unit {
	case "day":
		return truncateDay(p.ref).AddDate(0, 0, amount), true
	case "week":
		return truncateDay(p.ref).AddDate(0, 0, amount*7), true
	case "month":
		return truncateDay(p.ref).AddDate(0, amount, 0), true
	case "year":
		return truncateDay(p.ref).AddDate(amount, 0, 0), true
	case "hour":
		return p.ref.Add(time.Duration(amount) * time.Hour), true
	case "minute":
		return p.ref.Add(time.Duration(amount) * time.Minute), true
	}
	p.recordError(p.makeError("day", "week", "month", "year", "hour", "minute"))
	return time.Time{}, false
}

func (p *parser) resolveCountedWeekday(count int, target time.Weekday) (time.Time, bool) {
	if count <= 0 {
		p.recordError(p.makeError("positive number"))
		return time.Time{}, false
	}
	// Start from the day after ref
	d := truncateDay(p.ref).AddDate(0, 0, 1)
	// Find the first occurrence of the target weekday
	for d.Weekday() != target {
		d = d.AddDate(0, 0, 1)
	}
	// Advance by (count-1) weeks to get the Nth occurrence
	d = d.AddDate(0, 0, 7*(count-1))
	return d, true
}

func (p *parser) parseOrdinalWeekdayInMonth() (time.Time, bool) {
	saved := p.save()
	ordinal := p.advance() // consume ordinal
	p.skipNoise()

	if p.peek().Kind != tokenWeekday {
		err := p.makeError("weekday")
		p.restore(saved)
		p.recordError(err)
		return time.Time{}, false
	}
	weekday := p.advance()
	p.skipNoise()

	// Optional "in" or "of"
	if p.peek().Kind == tokenPreposition && (p.peek().Value == "in" || p.peek().Value == "of") {
		p.advance()
		p.skipNoise()
	}

	// Check for "next month" / "last month" / "this month" pattern
	if p.peek().Kind == tokenModifier {
		modSaved := p.save()
		mod := p.advance()
		p.skipNoise()
		if p.peek().Kind == tokenUnit && p.peek().Value == "month" {
			p.advance()
			ref := truncateDay(p.ref)
			var targetMonth time.Month
			targetYear := ref.Year()
			switch mod.Value {
			case "this":
				targetMonth = ref.Month()
			case "next":
				targetMonth = ref.Month() + 1
				if targetMonth > 12 {
					targetMonth = 1
					targetYear++
				}
			case "last":
				targetMonth = ref.Month() - 1
				if targetMonth < 1 {
					targetMonth = 12
					targetYear--
				}
			}
			return p.resolveOrdinalWeekdayInMonth(ordinal.IntVal, weekday.Weekday, targetMonth, targetYear)
		}
		p.restore(modSaved)
	}

	if p.peek().Kind != tokenMonth {
		err := p.makeError("month")
		p.restore(saved)
		p.recordError(err)
		return time.Time{}, false
	}
	month := p.advance()

	// Optional year
	p.skipNoise()
	year := 0
	if p.peek().Kind == tokenNumber && p.peek().IntVal > maxDayOfMonth {
		year = p.advance().IntVal
	}

	return p.resolveOrdinalWeekdayInMonth(ordinal.IntVal, weekday.Weekday, month.Month, year)
}

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
			Position: 0,
			Expected: []string{fmt.Sprintf("%d or fewer", maxOccurrences)},
			Found:    fmt.Sprintf("%d", n),
		})
		return time.Time{}, false
	}
	return d, true
}

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

func (p *parser) parseAbsoluteDate() (time.Time, bool) {
	saved := p.save()
	monthTok := p.advance() // consume month
	p.skipNoise()

	// Next token could be: day (number/ordinal), or year-only (number > 31)
	var day int
	year := p.ref.Year()
	hasYear := false

	switch p.peek().Kind {
	case tokenNumber:
		if p.peek().IntVal > maxDayOfMonth {
			// "march 2027" — year only, default to 1st of the month
			year = p.advance().IntVal
			hasYear = true
			day = 1
		} else {
			day = p.advance().IntVal
		}
	case tokenOrdinal:
		day = p.advance().IntVal
	default:
		err := p.makeError("day number")
		p.restore(saved)
		p.recordError(err)
		return time.Time{}, false
	}

	// Optional year after day: "march 15 2027"
	if !hasYear {
		p.skipNoise()
		if p.peek().Kind == tokenNumber && p.peek().IntVal > maxDayOfMonth {
			year = p.advance().IntVal
			hasYear = true
		}
	}

	// If no explicit year and the date has already passed, use next year
	if !hasYear && monthTok.Month < p.ref.Month() {
		year++
	}

	maxDay := daysIn(year, monthTok.Month, p.ref.Location())
	if day < 1 || day > maxDay {
		p.recordError(&ParseError{
			Position: saved,
			Expected: []string{fmt.Sprintf("day between 1 and %d", maxDay)},
			Found:    fmt.Sprintf("%d", day),
		})
		p.restore(saved)
		return time.Time{}, false
	}

	return time.Date(year, monthTok.Month, day, 0, 0, 0, 0, p.ref.Location()), true
}

func (p *parser) resolvePeriodRef(modifier, unit string) (time.Time, bool) {
	ref := truncateDay(p.ref)

	switch unit {
	case "week":
		dow := int(ref.Weekday())
		sunday := ref.AddDate(0, 0, -dow)

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
		firstOfMonth := time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, ref.Location())

		switch modifier {
		case "this":
			return firstOfMonth, true
		case "next":
			if p.opts.periodMode == PeriodSame {
				return ref.AddDate(0, 1, 0), true
			}
			return firstOfMonth.AddDate(0, 1, 0), true
		case "last":
			if p.opts.periodMode == PeriodSame {
				return ref.AddDate(0, -1, 0), true
			}
			return firstOfMonth.AddDate(0, -1, 0), true
		}
	}
	p.recordError(p.makeError("week", "month"))
	return time.Time{}, false
}

func (p *parser) parseBoundaryExpr() (time.Time, bool) {
	saved := p.save()
	boundary := p.advance() // "beginning" or "end"
	p.skipNoise()

	if p.peek().Kind != tokenPreposition || p.peek().Value != "of" {
		err := p.makeError("of")
		p.restore(saved)
		p.recordError(err)
		return time.Time{}, false
	}
	p.advance() // consume "of"
	p.skipNoise()

	// Optional modifier
	modifier := "this"
	if p.peek().Kind == tokenModifier {
		modifier = p.advance().Value
		p.skipNoise()
	}

	tok := p.peek()
	if tok.Kind != tokenUnit || (tok.Value != "week" && tok.Value != "month" && tok.Value != "quarter" && tok.Value != "year") {
		err := p.makeError("week", "month", "quarter", "year")
		p.restore(saved)
		p.recordError(err)
		return time.Time{}, false
	}
	unit := p.advance().Value

	return p.resolveBoundary(boundary.Value, modifier, unit)
}

func (p *parser) resolveBoundary(boundary, modifier, unit string) (time.Time, bool) {
	ref := truncateDay(p.ref)
	loc := ref.Location()

	switch unit {
	case "week":
		dow := int(ref.Weekday())
		sunday := ref.AddDate(0, 0, -dow)
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
		var targetMonth time.Month
		targetYear := ref.Year()
		switch modifier {
		case "this":
			targetMonth = ref.Month()
		case "next":
			targetMonth = ref.Month() + 1
			if targetMonth > 12 {
				targetMonth = 1
				targetYear++
			}
		case "last":
			targetMonth = ref.Month() - 1
			if targetMonth < 1 {
				targetMonth = 12
				targetYear--
			}
		}
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
		q := fiscalMonth / monthsPerQuarter
		switch modifier {
		case "next":
			q++
		case "last":
			q--
		}
		// Convert back to calendar month/year.
		// Each fiscal quarter starts at fyStart + q*monthsPerQuarter months from the FY start year.
		totalMonths := (fyStart - 1) + q*monthsPerQuarter // 0-indexed calendar month
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

func (p *parser) parseMultiDate() ([]time.Time, bool) {
	p.skipNoise()
	if p.peek().Kind != tokenEvery {
		return nil, false
	}
	saved := p.save()
	p.advance() // consume "every"
	p.skipNoise()

	if p.peek().Kind != tokenWeekday {
		p.restore(saved)
		return nil, false
	}
	weekdayTok := p.advance()
	p.skipNoise()

	// Optional "in" or "of"
	if p.peek().Kind == tokenPreposition && (p.peek().Value == "in" || p.peek().Value == "of") {
		p.advance()
		p.skipNoise()
	}

	if p.peek().Kind != tokenMonth {
		p.restore(saved)
		return nil, false
	}
	monthTok := p.advance()

	// Optional year
	p.skipNoise()
	year := p.ref.Year()
	if p.peek().Kind == tokenNumber && p.peek().IntVal > maxDayOfMonth {
		year = p.advance().IntVal
	} else if monthTok.Month < p.ref.Month() {
		year++
	}

	// Enumerate all occurrences of the weekday in the month
	loc := p.ref.Location()
	first := time.Date(year, monthTok.Month, 1, 0, 0, 0, 0, loc)
	d := first
	for d.Weekday() != weekdayTok.Weekday {
		d = d.AddDate(0, 0, 1)
	}
	var results []time.Time
	for d.Month() == monthTok.Month {
		results = append(results, d)
		d = d.AddDate(0, 0, 7)
	}
	return results, true
}

func (p *parser) parseTimeExpr(base time.Time) (time.Time, bool) {
	saved := p.save()
	p.skipNoise()

	// Optional "at" prefix
	hasAt := false
	if p.peek().Kind == tokenPreposition && p.peek().Value == "at" {
		p.advance()
		p.skipNoise()
		hasAt = true
	}

	tok := p.peek()

	switch tok.Kind {
	case tokenNamedTime:
		p.advance()
		if tok.Value == "noon" {
			return setTime(base, 12, 0), true
		}
		return setTime(base, 0, 0), true // midnight

	case tokenNumber:
		num := p.advance()

		// number:number [meridiem]
		if p.peek().Kind == tokenColon {
			p.advance() // consume ":"
			if p.peek().Kind == tokenNumber {
				min := p.advance()
				h, m := num.IntVal, min.IntVal
				if m > 59 {
					p.restore(saved)
					return time.Time{}, false
				}
				if p.peek().Kind == tokenMeridiem {
					if h < 1 || h > 12 {
						p.restore(saved)
						return time.Time{}, false
					}
					h = applyMeridiem(h, p.advance().Value)
				} else if h > 23 {
					p.restore(saved)
					return time.Time{}, false
				}
				return setTime(base, h, m), true
			}
			// "3:" with no minutes — fail
			p.restore(saved)
			return time.Time{}, false
		}

		// number meridiem
		if p.peek().Kind == tokenMeridiem {
			if num.IntVal < 1 || num.IntVal > 12 {
				p.restore(saved)
				return time.Time{}, false
			}
			h := applyMeridiem(num.IntVal, p.advance().Value)
			return setTime(base, h, 0), true
		}

		// Bare number after "at" — treat as 24-hour time (e.g., "at 3" = 03:00)
		if hasAt && num.IntVal >= 0 && num.IntVal <= 23 {
			return setTime(base, num.IntVal, 0), true
		}

		// Just a number — not a time expression
		p.restore(saved)
		return time.Time{}, false
	}

	p.restore(saved)
	return time.Time{}, false
}

func setTime(base time.Time, hour, min int) time.Time {
	return time.Date(base.Year(), base.Month(), base.Day(), hour, min, 0, 0, base.Location())
}

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
