package wen

import (
	"context"
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
}

// WithPeriodStart makes period references resolve to the start of the period (default).
func WithPeriodStart() Option { return func(o *options) { o.periodMode = PeriodStart } }

// WithPeriodSame makes period references resolve to the same relative day in the next period.
func WithPeriodSame() Option { return func(o *options) { o.periodMode = PeriodSame } }

// WithFiscalYearStart sets the month (1-12) that begins the fiscal year.
// This affects quarter calculations: e.g., WithFiscalYearStart(10) makes
// Q1=Oct-Dec, Q2=Jan-Mar, Q3=Apr-Jun, Q4=Jul-Sep.
// Default is 1 (January), which gives standard calendar quarters.
// Values outside 1-12 are silently ignored (the default is retained).
func WithFiscalYearStart(month int) Option {
	return func(o *options) {
		if month >= 1 && month <= 12 {
			o.fiscalYearStart = month
		}
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
	fm := (month - startMonth + monthsPerYear) % monthsPerYear
	quarter = fm/monthsPerQuarter + 1
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

func buildParser(ctx context.Context, input string, ref time.Time, opts ...Option) *parser {
	o := options{periodMode: PeriodStart}
	for _, opt := range opts {
		opt(&o)
	}
	l := newLexer(input)
	tokens := l.tokenize()
	p := newParser(tokens, ref, o, input)
	p.ctx = ctx
	return p
}

// ParseRelative parses a natural language date/time expression relative to ref.
func ParseRelative(input string, ref time.Time, opts ...Option) (time.Time, error) {
	return ParseRelativeContext(context.Background(), input, ref, opts...)
}

// ParseRelativeContext is like ParseRelative but accepts a context for cancellation.
func ParseRelativeContext(ctx context.Context, input string, ref time.Time, opts ...Option) (time.Time, error) {
	p := buildParser(ctx, input, ref, opts...)
	result, err := p.parse()
	if err != nil {
		return time.Time{}, err
	}
	return result, nil
}

// ParseMulti parses expressions that may produce multiple dates (e.g., "every friday in april").
// Returns a slice of dates. Falls back to single-date parsing if not a multi-date expression.
func ParseMulti(input string, ref time.Time, opts ...Option) ([]time.Time, error) {
	return ParseMultiContext(context.Background(), input, ref, opts...)
}

// ParseMultiContext is like ParseMulti but accepts a context for cancellation.
func ParseMultiContext(ctx context.Context, input string, ref time.Time, opts ...Option) ([]time.Time, error) {
	p := buildParser(ctx, input, ref, opts...)

	// Try multi-date parse first
	if results, ok := p.parseMultiDate(); ok {
		p.skipNoise()
		if p.peek().Kind == tokenEOF {
			return results, nil
		}
	}

	// Fall back to single-date parse
	p.pos = 0
	result, err := p.parse()
	if err != nil {
		return nil, err
	}
	return []time.Time{result}, nil
}
