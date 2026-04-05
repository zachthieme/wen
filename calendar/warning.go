package calendar

// Warning represents a non-fatal issue encountered during configuration
// loading or file parsing. The operation completed with default or partial
// values.
type Warning struct {
	// Key identifies what triggered the warning: a config field name
	// (e.g., "week_numbering"), a file path, or an unknown config key.
	Key string

	// Message is the human-readable description.
	Message string
}

// Error implements the error interface so warnings can be used where errors
// are expected.
func (w Warning) Error() string { return w.Message }
