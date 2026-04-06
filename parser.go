package wen

import (
	"context"
	"fmt"
	"time"
)

// Grammar (informal BNF):
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
// Noise words ("the", "a") are silently skipped between tokens.

// maxDayOfMonth is the maximum valid day number; values above this are treated as years.
const maxDayOfMonth = 31

type parser struct {
	tokens  []token
	pos     int
	input   string
	bestErr *ParseError
	ctx     context.Context
}

func newParser(tokens []token, input string) *parser {
	return &parser{tokens: tokens, input: input}
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

func (p *parser) fail(saved int, expected ...string) (Expr, bool) {
	p.recordError(p.makeError(expected...))
	p.restore(saved)
	return nil, false
}

// shiftMonth shifts a month by delta and adjusts the year on
// overflow (>12) or underflow (<1).
func shiftMonth(month time.Month, year, delta int) (time.Month, int) {
	m := int(month) - 1 + delta // 0-indexed
	year += m / 12
	m %= 12
	if m < 0 {
		m += 12
		year--
	}
	return time.Month(m + 1), year
}

// modifierDelta converts "next"/"last"/"this" to +1/-1/0.
// Panics on unrecognized modifiers — callers must only pass values
// produced by the lexer's tokenModifier classification.
func modifierDelta(modifier string) int {
	switch modifier {
	case "next":
		return 1
	case "last":
		return -1
	case "this", "":
		return 0
	default:
		panic(fmt.Sprintf("wen: unexpected modifier %q in modifierDelta", modifier))
	}
}

func (p *parser) parse() (Expr, error) {
	if p.ctx != nil {
		if err := p.ctx.Err(); err != nil {
			return nil, &ParseError{
				Input:    p.input,
				Position: p.pos,
				Expected: []string{"date or time expression"},
				Cause:    err,
			}
		}
	}
	p.skipNoise()

	expr, ok := p.parseDateExpr()
	if ok {
		// Try optional trailing time expression
		saved := p.save()
		p.skipNoise()
		if hour, minute, matched := p.parseTimeExpr(); matched {
			expr = &WithTimeExpr{Date: expr, Hour: hour, Minute: minute}
		} else {
			p.restore(saved)
		}
	} else {
		// Try standalone time expression
		p.pos = 0
		p.skipNoise()
		hour, minute, matched := p.parseTimeExpr()
		if !matched {
			return nil, p.finalError()
		}
		expr = &WithTimeExpr{Date: nil, Hour: hour, Minute: minute}
	}

	p.skipNoise()
	if p.peek().Kind != tokenEOF {
		return nil, p.makeError("end of input")
	}
	return expr, nil
}

// parseDateExpr dispatches to the appropriate date production based on the
// current token kind (relative day, weekday, modifier, number, ordinal, month, or boundary).
func (p *parser) parseDateExpr() (Expr, bool) {
	if p.ctx != nil {
		if err := p.ctx.Err(); err != nil {
			return nil, false
		}
	}
	p.skipNoise()
	tok := p.peek()

	switch tok.Kind {
	case tokenRelativeDay:
		return p.parseRelativeDay()
	case tokenWeekday:
		p.advance() // consume weekday token
		return &ModWeekdayExpr{Modifier: "this", Weekday: tok.Weekday}, true
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
	return nil, false
}

func (p *parser) parseRelativeDay() (Expr, bool) {
	tok := p.advance()
	return &RelativeDayExpr{Day: tok.Value}, true
}

// parseModifierExpr handles "this/next/last <weekday|week|month>" patterns.
func (p *parser) parseModifierExpr() (Expr, bool) {
	saved := p.save()
	modifier := p.advance()
	p.skipNoise()

	tok := p.peek()
	if tok.Kind == tokenWeekday {
		// "last friday in november" is a distinct production from "last friday".
		if modifier.Value == "last" {
			if expr, ok := p.tryLastWeekdayInMonth(tok); ok {
				return expr, true
			}
		}
		p.advance() // consume weekday token
		return &ModWeekdayExpr{Modifier: modifier.Value, Weekday: tok.Weekday}, true
	}
	if tok.Kind == tokenUnit && (tok.Value == "week" || tok.Value == "month") {
		p.advance() // consume unit
		return &PeriodRefExpr{Modifier: modifier.Value, Unit: tok.Value}, true
	}

	return p.fail(saved,
		fmt.Sprintf("weekday after %q", modifier.Value),
		fmt.Sprintf("week/month after %q", modifier.Value),
	)
}

// tryLastWeekdayInMonth attempts to parse "last <weekday> in/of <month> [year]".
// The weekday token has been peeked but not consumed. Returns false and restores
// position if the pattern does not match.
func (p *parser) tryLastWeekdayInMonth(weekdayTok token) (Expr, bool) {
	saved := p.save()
	p.advance() // consume weekday
	p.skipNoise()

	next := p.peek()
	if next.Kind != tokenPreposition || (next.Value != "in" && next.Value != "of") {
		p.restore(saved)
		return nil, false
	}
	p.advance() // consume "in"/"of"
	p.skipNoise()

	if p.peek().Kind != tokenMonth {
		p.restore(saved)
		return nil, false
	}
	monthTok := p.advance()

	// Optional year
	p.skipNoise()
	year := 0
	if p.peek().Kind == tokenNumber && p.peek().IntVal > maxDayOfMonth {
		year = p.advance().IntVal
	}
	return &LastWeekdayInMonthExpr{Weekday: weekdayTok.Weekday, Month: monthTok.Month, Year: year}, true
}

// parseInExpr handles "in <number> <weekday|unit>" patterns (e.g., "in 3 days", "in 2 fridays").
func (p *parser) parseInExpr() (Expr, bool) {
	saved := p.save()
	p.advance() // consume "in"
	p.skipNoise()

	if p.peek().Kind != tokenNumber {
		return p.fail(saved, "number")
	}
	num := p.advance()
	p.skipNoise()

	next := p.peek()
	switch next.Kind {
	case tokenWeekday:
		p.advance() // consume weekday
		return &CountedWeekdayExpr{Count: num.IntVal, Weekday: next.Weekday}, true
	case tokenUnit:
		p.advance()
		return &RelativeOffsetExpr{N: num.IntVal, Unit: next.Value, Direction: 1}, true
	}

	return p.fail(saved, "weekday", "unit")
}

// parseNumberLeadExpr handles "<number> <unit> ago/from now" patterns (e.g., "3 days ago").
func (p *parser) parseNumberLeadExpr() (Expr, bool) {
	saved := p.save()
	num := p.advance()
	p.skipNoise()

	if p.peek().Kind != tokenUnit {
		return p.fail(saved, "unit")
	}
	unit := p.advance()
	p.skipNoise()

	tok := p.peek()
	if tok.Kind == tokenPreposition && tok.Value == "ago" {
		p.advance()
		return &RelativeOffsetExpr{N: num.IntVal, Unit: unit.Value, Direction: -1}, true
	}
	if tok.Kind == tokenPreposition && tok.Value == "from" {
		p.advance()
		p.skipNoise()
		if p.peek().Kind == tokenPreposition && p.peek().Value == "now" {
			p.advance()
			return &RelativeOffsetExpr{N: num.IntVal, Unit: unit.Value, Direction: 1}, true
		}
		return p.fail(saved, "now")
	}

	return p.fail(saved, "ago", "from")
}

// parseOrdinalWeekdayInMonth handles "Nth weekday in/of month [year]" patterns
// (e.g., "first monday of april", "third wednesday of next month").
func (p *parser) parseOrdinalWeekdayInMonth() (Expr, bool) {
	saved := p.save()
	ordinal := p.advance() // consume ordinal
	p.skipNoise()

	if p.peek().Kind != tokenWeekday {
		return p.fail(saved, "weekday")
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
			return &OrdinalWeekdayExpr{
				N:             ordinal.IntVal,
				Weekday:       weekday.Weekday,
				MonthModifier: mod.Value,
			}, true
		}
		p.restore(modSaved)
	}

	if p.peek().Kind != tokenMonth {
		return p.fail(saved, "month")
	}
	month := p.advance()

	// Optional year
	p.skipNoise()
	year := 0
	if p.peek().Kind == tokenNumber && p.peek().IntVal > maxDayOfMonth {
		year = p.advance().IntVal
	}

	return &OrdinalWeekdayExpr{
		N:       ordinal.IntVal,
		Weekday: weekday.Weekday,
		Month:   month.Month,
		Year:    year,
	}, true
}

// parseAbsoluteDate handles "<month> <day> [year]" and "<month> <year>" patterns.
func (p *parser) parseAbsoluteDate() (Expr, bool) {
	saved := p.save()
	monthTok := p.advance() // consume month
	p.skipNoise()

	// Next token could be: day (number/ordinal), or year-only (number > 31)
	var day int
	year := 0
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
		return p.fail(saved, "day number")
	}

	// Optional year after day: "march 15 2027"
	if !hasYear {
		p.skipNoise()
		if p.peek().Kind == tokenNumber && p.peek().IntVal > maxDayOfMonth {
			year = p.advance().IntVal
		}
	}

	return &AbsoluteDateExpr{Month: monthTok.Month, Day: day, Year: year}, true
}

// parseBoundaryExpr handles "beginning/end of [modifier] week/month/quarter/year".
func (p *parser) parseBoundaryExpr() (Expr, bool) {
	saved := p.save()
	boundary := p.advance() // "beginning" or "end"
	p.skipNoise()

	if p.peek().Kind != tokenPreposition || p.peek().Value != "of" {
		return p.fail(saved, "of")
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
		return p.fail(saved, "week", "month", "quarter", "year")
	}
	unit := p.advance().Value

	return &BoundaryExpr{Boundary: boundary.Value, Modifier: modifier, Unit: unit}, true
}

// parseMultiDate handles "every <weekday> [in/of] <month> [year]" and returns
// a [MultiDateExpr] AST node.
func (p *parser) parseMultiDate() (Expr, bool) {
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
	year := 0
	if p.peek().Kind == tokenNumber && p.peek().IntVal > maxDayOfMonth {
		year = p.advance().IntVal
	}

	return &MultiDateExpr{Weekday: weekdayTok.Weekday, Month: monthTok.Month, Year: year}, true
}

// parseTimeExpr parses an optional time-of-day expression (e.g., "at 3pm",
// "15:00", "at noon") and returns the hour and minute components.
func (p *parser) parseTimeExpr() (hour int, minute int, ok bool) {
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
			return 12, 0, true
		}
		return 0, 0, true // midnight

	case tokenNumber:
		num := p.advance()
		switch {
		case p.peek().Kind == tokenColon:
			return p.parseColonTime(saved, num)
		case p.peek().Kind == tokenMeridiem:
			return p.parseMeridiemTime(saved, num)
		case hasAt && num.IntVal >= 0 && num.IntVal <= 23:
			return num.IntVal, 0, true
		}
	}

	p.restore(saved)
	return 0, 0, false
}

// parseColonTime handles "N:M [am/pm]" and "N:M" (24-hour) time formats.
// The number token has been consumed; the colon has been peeked.
func (p *parser) parseColonTime(saved int, num token) (int, int, bool) {
	p.advance() // consume ":"
	if p.peek().Kind != tokenNumber {
		p.restore(saved)
		return 0, 0, false
	}
	min := p.advance()
	h, m := num.IntVal, min.IntVal

	if m > 59 {
		p.recordError(&ParseError{
			Input:    p.input,
			Position: num.Position,
			Expected: []string{"valid time (minute 0-59)"},
			Found:    fmt.Sprintf("%d:%02d", h, m),
		})
		p.restore(saved)
		return 0, 0, false
	}

	if p.peek().Kind == tokenMeridiem {
		if h < 1 || h > 12 {
			p.recordError(&ParseError{
				Input:    p.input,
				Position: num.Position,
				Expected: []string{"valid time (hour 1-12 with am/pm)"},
				Found:    fmt.Sprintf("%d:%02d%s", h, m, p.peek().Value),
			})
			p.restore(saved)
			return 0, 0, false
		}
		h = applyMeridiem(h, p.advance().Value)
	} else if h > 23 {
		p.recordError(&ParseError{
			Input:    p.input,
			Position: num.Position,
			Expected: []string{"valid time (hour 0-23)"},
			Found:    fmt.Sprintf("%d:%02d", h, m),
		})
		p.restore(saved)
		return 0, 0, false
	}

	return h, m, true
}

// parseMeridiemTime handles "N am/pm" time format.
// The number token has been consumed; the meridiem has been peeked.
func (p *parser) parseMeridiemTime(saved int, num token) (int, int, bool) {
	if num.IntVal < 1 || num.IntVal > 12 {
		p.recordError(&ParseError{
			Input:    p.input,
			Position: num.Position,
			Expected: []string{"valid time (hour 1-12 with am/pm)"},
			Found:    fmt.Sprintf("%d%s", num.IntVal, p.peek().Value),
		})
		p.restore(saved)
		return 0, 0, false
	}
	h := applyMeridiem(num.IntVal, p.advance().Value)
	return h, 0, true
}
