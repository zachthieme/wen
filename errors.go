package wen

import (
	"fmt"
	"strings"
)

// NoPosition indicates a ParseError that is not tied to a specific byte offset
// in the input (e.g., semantic errors like "february 30" or context cancellation).
const NoPosition = -1

// ParseError is a structured error returned when parsing fails.
type ParseError struct {
	Input    string   // the original input string
	Position int      // byte offset where parsing failed; -1 for semantic errors where position is not applicable
	Expected []string // what the parser expected at this position
	Found    string   // what was actually found
	Cause    error    // underlying error, if any (supports errors.Is / errors.As chains)
}

func (e *ParseError) Error() string {
	exp := strings.Join(e.Expected, " or ")
	var msg string
	switch {
	case e.Position < 0 && e.Found == "":
		msg = fmt.Sprintf("unexpected end of input, expected %s", exp)
	case e.Position < 0:
		msg = fmt.Sprintf("unexpected %q, expected %s", e.Found, exp)
	case e.Found == "":
		msg = fmt.Sprintf("unexpected end of input at position %d, expected %s", e.Position, exp)
	default:
		msg = fmt.Sprintf("unexpected %q at position %d, expected %s", e.Found, e.Position, exp)
	}
	if e.Input != "" {
		msg = fmt.Sprintf("%s in %q", msg, e.Input)
	}
	if e.Cause != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Cause)
	}
	return msg
}

// Unwrap returns the underlying cause, supporting [errors.Is] and [errors.As].
func (e *ParseError) Unwrap() error {
	return e.Cause
}
