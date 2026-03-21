package wen

import (
	"errors"
	"time"
)

// DateLayout is the standard date format used for output (yyyy-mm-dd).
const DateLayout = "2006-01-02"

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
func WithFiscalYearStart(month int) Option {
	return func(o *options) {
		if month >= 1 && month <= 12 {
			o.fiscalYearStart = month
		}
	}
}

// Parse parses a natural language date/time expression relative to time.Now().
func Parse(input string, opts ...Option) (time.Time, error) {
	return ParseRelative(input, time.Now(), opts...)
}

// ParseRelative parses a natural language date/time expression relative to ref.
func ParseRelative(input string, ref time.Time, opts ...Option) (time.Time, error) {
	o := options{periodMode: PeriodStart}
	for _, opt := range opts {
		opt(&o)
	}
	l := newLexer(input)
	tokens := l.tokenize()
	p := newParser(tokens, ref, o)
	result, err := p.parse()
	if err != nil {
		var pe *ParseError
		if errors.As(err, &pe) {
			pe.Input = input
		}
		return time.Time{}, err
	}
	return result, nil
}

// ParseMulti parses expressions that may produce multiple dates (e.g., "every friday in april").
// Returns a slice of dates. Falls back to single-date parsing if not a multi-date expression.
func ParseMulti(input string, ref time.Time, opts ...Option) ([]time.Time, error) {
	o := options{periodMode: PeriodStart}
	for _, opt := range opts {
		opt(&o)
	}
	l := newLexer(input)
	tokens := l.tokenize()
	p := newParser(tokens, ref, o)

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
		var pe *ParseError
		if errors.As(err, &pe) {
			pe.Input = input
		}
		return nil, err
	}
	return []time.Time{result}, nil
}
