package wen

import (
	"fmt"
	"strings"
)

// ParseError is a structured error returned when parsing fails.
type ParseError struct {
	Input    string   // the original input string
	Position int      // byte offset where parsing failed
	Expected []string // what the parser expected at this position
	Found    string   // what was actually found
}

func (e *ParseError) Error() string {
	exp := strings.Join(e.Expected, " or ")
	var msg string
	if e.Found == "" {
		msg = fmt.Sprintf("unexpected end of input at position %d, expected %s", e.Position, exp)
	} else {
		msg = fmt.Sprintf("unexpected %q at position %d, expected %s", e.Found, e.Position, exp)
	}
	if e.Input != "" {
		msg = fmt.Sprintf("%s in %q", msg, e.Input)
	}
	return msg
}
