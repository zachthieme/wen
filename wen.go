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
	periodMode PeriodMode
}

// WithPeriodStart makes period references resolve to the start of the period (default).
func WithPeriodStart() Option { return func(o *options) { o.periodMode = PeriodStart } }

// WithPeriodSame makes period references resolve to the same relative day in the next period.
func WithPeriodSame() Option { return func(o *options) { o.periodMode = PeriodSame } }

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
