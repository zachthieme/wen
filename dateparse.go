package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
)

// dateParser is the package-level parser — initialized once, reused on every call.
var dateParser = newDateParser()

func newDateParser() *when.Parser {
	p := when.New(nil)
	p.Add(en.All...)
	p.Add(common.All...)
	return p
}

var weekdays = map[string]time.Weekday{
	"sunday": time.Sunday, "sun": time.Sunday,
	"monday": time.Monday, "mon": time.Monday,
	"tuesday": time.Tuesday, "tue": time.Tuesday,
	"wednesday": time.Wednesday, "wed": time.Wednesday,
	"thursday": time.Thursday, "thu": time.Thursday,
	"friday": time.Friday, "fri": time.Friday,
	"saturday": time.Saturday, "sat": time.Saturday,
}

// parseRelativeWeekday handles "this/next/last <weekday>" patterns.
// Returns the resolved date and true if matched, or zero time and false.
func parseRelativeWeekday(input string, ref time.Time) (time.Time, bool) {
	lower := strings.ToLower(strings.TrimSpace(input))
	parts := strings.Fields(lower)
	if len(parts) != 2 {
		return time.Time{}, false
	}

	prefix := parts[0]
	if prefix != "this" && prefix != "next" && prefix != "last" {
		return time.Time{}, false
	}

	target, ok := weekdays[parts[1]]
	if !ok {
		return time.Time{}, false
	}

	refDay := ref.Weekday()
	switch prefix {
	case "this":
		// "this <weekday>" returns the given weekday within the current week,
		// which may be in the past or future relative to ref. For example, on a
		// Wednesday, "this monday" returns the preceding Monday (2 days ago) and
		// "this friday" returns the upcoming Friday (2 days ahead). When ref falls
		// on the target day (e.g., "this sunday" on a Sunday), it returns ref itself.
		diff := int(target) - int(refDay)
		return ref.AddDate(0, 0, diff), true
	case "next":
		// Advance to the start of next week (Sunday), then offset by target weekday.
		// When ref is already Sunday, (7 - 0) % 7 == 0, so the guard below ensures
		// we skip ahead a full week rather than treating ref as "start of next week."
		daysToNextSunday := (7 - int(refDay)) % 7
		if daysToNextSunday == 0 {
			daysToNextSunday = 7
		}
		diff := daysToNextSunday + int(target)
		return ref.AddDate(0, 0, diff), true
	case "last":
		// "last <weekday>" returns the most recent past occurrence of that day.
		// When ref falls on the target day (e.g., "last tuesday" on a Tuesday),
		// diff == 0, so we add 7 to go back a full week — unlike "this," which
		// would return today.
		diff := int(refDay) - int(target)
		if diff <= 0 {
			diff += 7
		}
		return ref.AddDate(0, 0, -diff), true
	}

	return time.Time{}, false
}

func parseDate(input string, ref time.Time) (time.Time, error) {
	if t, ok := parseRelativeWeekday(input, ref); ok {
		return t, nil
	}

	result, err := dateParser.Parse(input, ref)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse date %q: %w", input, err)
	}
	if result == nil {
		return time.Time{}, fmt.Errorf("could not parse date %q", input)
	}
	return result.Time, nil
}
