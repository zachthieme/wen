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

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// DateLayout is the standard date format used for output (yyyy-mm-dd).
const DateLayout = "2006-01-02"

const (
	monthsPerYear    = 12
	monthsPerQuarter = 3
)

// PeriodMode controls how bare period references like "next week" resolve.
type PeriodMode int

const (
	// PeriodStart resolves to the start of the period ("next week" = Monday).
	PeriodStart PeriodMode = iota
	// PeriodSame resolves to the same day in the next period ("next week" = same weekday + 7).
	PeriodSame
)

// Option configures parsing behavior.
type Option func(*options)

type options struct {
	periodMode      PeriodMode
	fiscalYearStart int // 1-12, month the fiscal year begins (default 1 = January)
	err             error
}

// WithPeriodStart makes period references resolve to the start of the period (default).
func WithPeriodStart() Option { return func(o *options) { o.periodMode = PeriodStart } }

// WithPeriodSame makes period references resolve to the same relative day in the next period.
func WithPeriodSame() Option { return func(o *options) { o.periodMode = PeriodSame } }

// WithFiscalYearStart sets the month (1-12) that begins the fiscal year.
// This affects quarter calculations: e.g., WithFiscalYearStart(10) makes
// Q1=Oct-Dec, Q2=Jan-Mar, Q3=Apr-Jun, Q4=Jul-Sep.
// Default is 1 (January), which gives standard calendar quarters.
// Values outside 1-12 cause a validation error at parse time.
func WithFiscalYearStart(month int) Option {
	return func(o *options) {
		if month < 1 || month > 12 {
			o.err = fmt.Errorf("invalid fiscal year start month %d: must be between 1 and 12", month)
			return
		}
		o.fiscalYearStart = month
	}
}

// FiscalQuarter returns the fiscal quarter (1-4) and fiscal year for a given
// calendar month and year, with the fiscal year starting in startMonth (1-12).
// The fiscal year number is the calendar year the FY ends in
// (e.g., startMonth=10: Oct 2025 → Q1 FY2026, Mar 2026 → Q2 FY2026).
// When startMonth is 1 (calendar year), Q1=Jan-Mar and fiscalYear equals year.
// Values of startMonth outside 1-12 are coerced to 1 (January).
func FiscalQuarter(month, year, startMonth int) (quarter, fiscalYear int) {
	if startMonth < 1 || startMonth > monthsPerYear {
		startMonth = 1
	}
	fiscalMonth := (month - startMonth + monthsPerYear) % monthsPerYear
	quarter = fiscalMonth/monthsPerQuarter + 1
	switch {
	case startMonth == 1:
		fiscalYear = year
	case month >= startMonth:
		fiscalYear = year + 1
	default:
		fiscalYear = year
	}
	return quarter, fiscalYear
}

// CountWorkdays returns the number of weekdays (Mon-Fri) in the half-open
// interval [start, end). That is, it counts start's weekday but not end's.
// If start equals end, returns 0.
// The order of start and end does not matter — the result is always non-negative.
func CountWorkdays(start, end time.Time) int {
	// Normalize to UTC midnight
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	end = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)
	if start.After(end) {
		start, end = end, start
	}
	totalDays := int(end.Sub(start).Hours() / 24)
	if totalDays == 0 {
		return 0
	}
	fullWeeks := totalDays / 7
	remaining := totalDays % 7
	workdays := fullWeeks * 5
	startDay := int(start.Weekday())
	for i := range remaining {
		dayOfWeek := (startDay + i) % 7
		if dayOfWeek != int(time.Saturday) && dayOfWeek != int(time.Sunday) {
			workdays++
		}
	}
	return workdays
}

// LookupMonth returns the time.Month for a month name or abbreviation (case-insensitive).
// Returns 0 and false if the name is not recognized.
func LookupMonth(name string) (time.Month, bool) {
	m, ok := months[strings.ToLower(name)]
	return m, ok
}

// Parse parses a natural language date/time expression relative to time.Now().
func Parse(input string, opts ...Option) (time.Time, error) {
	return ParseRelative(input, time.Now(), opts...)
}

// ParseContext is like Parse but accepts a context for cancellation.
func ParseContext(ctx context.Context, input string, opts ...Option) (time.Time, error) {
	return ParseRelativeContext(ctx, input, time.Now(), opts...)
}

func buildParser(ctx context.Context, input string, opts ...Option) (*parser, options, error) {
	o := options{periodMode: PeriodStart}
	for _, opt := range opts {
		opt(&o)
	}
	if o.err != nil {
		return nil, o, o.err
	}
	l := newLexer(input)
	tokens := l.tokenize()
	p := newParser(tokens, input)
	p.ctx = ctx
	return p, o, nil
}

// ParseRelative parses a natural language date/time expression relative to ref.
func ParseRelative(input string, ref time.Time, opts ...Option) (time.Time, error) {
	return ParseRelativeContext(context.Background(), input, ref, opts...)
}

// ParseRelativeContext is like ParseRelative but accepts a context for cancellation.
func ParseRelativeContext(ctx context.Context, input string, ref time.Time, opts ...Option) (time.Time, error) {
	p, o, err := buildParser(ctx, input, opts...)
	if err != nil {
		return time.Time{}, err
	}
	expr, err := p.parse()
	if err != nil {
		return time.Time{}, err
	}
	return newResolver(ref, o, input).resolve(expr)
}

// ParseMulti parses expressions that may produce multiple dates (e.g., "every friday in april").
// Returns a slice of dates. Falls back to single-date parsing if not a multi-date expression.
func ParseMulti(input string, ref time.Time, opts ...Option) ([]time.Time, error) {
	return ParseMultiContext(context.Background(), input, ref, opts...)
}

// ParseMultiContext is like ParseMulti but accepts a context for cancellation.
func ParseMultiContext(ctx context.Context, input string, ref time.Time, opts ...Option) ([]time.Time, error) {
	p, o, err := buildParser(ctx, input, opts...)
	if err != nil {
		return nil, err
	}

	r := newResolver(ref, o, input)

	// Try multi-date parse first
	if expr, ok := p.parseMultiDate(); ok {
		p.skipNoise()
		if p.peek().Kind == tokenEOF {
			return r.resolveMulti(expr)
		}
	}

	// Fall back to single-date parse
	p.pos = 0
	expr, err := p.parse()
	if err != nil {
		return nil, err
	}
	return r.resolveMulti(expr)
}
