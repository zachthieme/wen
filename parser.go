package wen

import (
	"context"
	"fmt"
	"time"
)

// Grammar (informal BNF):
//
//   input        = dateExpr [timeExpr] | timeExpr
//   dateExpr     = relativeDay
//                | [modifier] weekday
//                | modifier ("week" | "month")
//                | "in" number (weekday | unit)
//                | number unit ("ago" | "from" "now")
//                | ordinal weekday [prep] (month [year] | modifier "month")
//                | "last" weekday prep month [year]
//                | month (day [year] | year)
//                | boundary "of" [modifier] ("week" | "month" | "quarter" | "year")
//   multiExpr    = "every" weekday [prep] month [year]
//   timeExpr     = ["at"] (namedTime | number ":" number [meridiem] | number meridiem | "at" number)
//   relativeDay  = "today" | "tomorrow" | "yesterday"
//   modifier     = "this" | "next" | "last"
//   boundary     = "beginning" | "end"
//   prep         = "in" | "of"
//   namedTime    = "noon" | "midnight"
//   meridiem     = "am" | "pm"
//   ordinal      = "first" | "1st" | ... | "twelfth" | "12th"
//   number       = digit+ | "one" | "two" | ... | "thirty"
//   unit         = "day" | "week" | "month" | "quarter" | "year" | "hour" | "minute"
//   day          = number | ordinal
//   year         = number (> 31)
//   weekday      = "monday" | ... | "sunday"
//   month        = "january" | ... | "december"
//
// Noise words ("the", "a") are silently skipped between tokens.

// maxDayOfMonth is the maximum valid day number; values above this are treated as years.
const maxDayOfMonth = 31

type parser struct {
	tokens  []token
	pos     int
	ref     time.Time
	opts    options
	input   string
	bestErr *ParseError
	ctx     context.Context
}

func newParser(tokens []token, ref time.Time, opts options, input string) *parser {
	return &parser{tokens: tokens, ref: ref, opts: opts, input: input}
}

func (p *parser) peek() token {
	if p.pos >= len(p.tokens) {
		return token{Kind: tokenEOF}
	}
	return p.tokens[p.pos]
}
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
		Input:    p.input,
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

// TruncateDay returns t with the time-of-day components zeroed out,
// preserving the location.
func TruncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// DaysIn returns the number of days in the given month and year.
func DaysIn(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}

// shiftMonth shifts a month by delta and adjusts the year on
// overflow (>12) or underflow (<1).
func shiftMonth(month time.Month, year, delta int) (time.Month, int) {
	m := int(month) + delta
	for m > 12 {
		m -= 12
		year++
	}
	for m < 1 {
		m += 12
		year--
	}
	return time.Month(m), year
}

// modifierDelta converts "next"/"last"/"this" to +1/-1/0.
func modifierDelta(modifier string) int {
	switch modifier {
	case "next":
		return 1
	case "last":
		return -1
	default:
		return 0
	}
}

func (p *parser) parse() (time.Time, error) {
	if p.ctx != nil {
		if err := p.ctx.Err(); err != nil {
			return time.Time{}, &ParseError{
				Input:    p.input,
				Position: p.pos,
				Expected: []string{"date or time expression"},
				Cause:    err,
			}
		}
	}
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
		result, matched = p.parseTimeExpr(TruncateDay(p.ref))
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

// parseDateExpr dispatches to the appropriate date production based on the
// current token kind (relative day, weekday, modifier, number, ordinal, month, or boundary).
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
	base := TruncateDay(p.ref)
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

// parseModifierExpr handles "this/next/last <weekday|week|month>" patterns.
func (p *parser) parseModifierExpr() (time.Time, bool) {
	saved := p.save()
	modifier := p.advance()
	p.skipNoise()

	tok := p.peek()
	if tok.Kind == tokenWeekday {
		// "last friday in november" is a distinct production from "last friday".
		if modifier.Value == "last" {
			if result, ok := p.tryLastWeekdayInMonth(tok); ok {
				return result, true
			}
		}
		return p.resolveModWeekday(modifier.Value, tok)
	}
	if tok.Kind == tokenUnit && (tok.Value == "week" || tok.Value == "month") {
		p.advance() // consume unit
		return p.resolvePeriodRef(modifier.Value, tok.Value)
	}

	p.recordError(p.makeError(
		fmt.Sprintf("weekday after %q", modifier.Value),
		fmt.Sprintf("week/month after %q", modifier.Value),
	))
	p.restore(saved)
	return time.Time{}, false
}

// tryLastWeekdayInMonth attempts to parse "last <weekday> in/of <month> [year]".
// The weekday token has been peeked but not consumed. Returns false and restores
// position if the pattern does not match.
func (p *parser) tryLastWeekdayInMonth(weekdayTok token) (time.Time, bool) {
	saved := p.save()
	p.advance() // consume weekday
	p.skipNoise()

	next := p.peek()
	if next.Kind != tokenPreposition || (next.Value != "in" && next.Value != "of") {
		p.restore(saved)
		return time.Time{}, false
	}
	p.advance() // consume "in"/"of"
	p.skipNoise()

	if p.peek().Kind != tokenMonth {
		p.restore(saved)
		return time.Time{}, false
	}
	monthTok := p.advance()

	// Optional year
	p.skipNoise()
	year := 0
	if p.peek().Kind == tokenNumber && p.peek().IntVal > maxDayOfMonth {
		year = p.advance().IntVal
	}
	return p.resolveLastWeekdayInMonth(weekdayTok.Weekday, monthTok.Month, year)
}

// parseInExpr handles "in <number> <weekday|unit>" patterns (e.g., "in 3 days", "in 2 fridays").
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

// parseNumberLeadExpr handles "<number> <unit> ago/from now" patterns (e.g., "3 days ago").
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

// parseOrdinalWeekdayInMonth handles "Nth weekday in/of month [year]" patterns
// (e.g., "first monday of april", "third wednesday of next month").
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
			ref := TruncateDay(p.ref)
			targetMonth, targetYear := shiftMonth(ref.Month(), ref.Year(), modifierDelta(mod.Value))
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

// parseAbsoluteDate handles "<month> <day> [year]" and "<month> <year>" patterns.
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

	maxDay := DaysIn(year, monthTok.Month, p.ref.Location())
	if day < 1 || day > maxDay {
		p.recordError(&ParseError{
			Input:    p.input,
			Position: saved,
			Expected: []string{fmt.Sprintf("day between 1 and %d", maxDay)},
			Found:    fmt.Sprintf("%d", day),
		})
		p.restore(saved)
		return time.Time{}, false
	}

	return time.Date(year, monthTok.Month, day, 0, 0, 0, 0, p.ref.Location()), true
}

// parseBoundaryExpr handles "beginning/end of [modifier] week/month/quarter/year".
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

// parseMultiDate handles "every <weekday> [in/of] <month> [year]" and returns
// all occurrences of that weekday in the given month.
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

// parseTimeExpr parses an optional time-of-day expression (e.g., "at 3pm",
// "15:00", "at noon") and applies it to the given base date.
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
