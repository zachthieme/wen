package calendar

import (
	"fmt"
	"strings"
	"time"
)

// monthAbbrevs provides 2-character month abbreviations for the strip row view.
var monthAbbrevs = [12]string{"Ja", "Fe", "Mr", "Ap", "My", "Jn", "Jl", "Au", "Se", "Oc", "No", "De"}

// stripWindow computes the week-aligned start and end dates for a strip view
// of the given month. The window starts on weekStartDay on or before the 1st,
// and ends the day before the next weekStartDay on or after the last day.
func stripWindow(year int, month time.Month, weekStartDay int, loc *time.Location) (start, end time.Time) {
	first := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	offset := (int(first.Weekday()) - weekStartDay + 7) % 7
	start = first.AddDate(0, 0, -offset)

	last := time.Date(year, month+1, 0, 0, 0, 0, 0, loc) // last day of month
	daysUntilEnd := (weekStartDay - int(last.Weekday()) - 1 + 7) % 7
	end = last.AddDate(0, 0, daysUntilEnd)
	return start, end
}

// renderStripDayHeaders produces the first row of the strip: a 3-character
// leading space followed by repeating day-of-week abbreviations for each day
// from start to end (inclusive).
func (m RowModel) renderStripDayHeaders(start, end time.Time) string {
	var b strings.Builder
	b.WriteString("   ") // 3-char leading space for month abbreviation column
	first := true
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if !first {
			b.WriteString(" ")
		}
		b.WriteString(dayNames[d.Weekday()])
		first = false
	}
	return m.styles.dayHeader.Render(b.String())
}

// renderStripDays produces the second row of the strip: a 2-character month
// abbreviation followed by a space, then day numbers with cursor/today/
// highlight/range/padding styling.
func (m RowModel) renderStripDays(start, end time.Time) string {
	year, month, _ := m.cursor.Date()
	loc := m.cursor.Location()
	first := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	last := time.Date(year, month+1, 0, 0, 0, 0, 0, loc)

	abbrev := monthAbbrevs[month-1]

	var b strings.Builder
	b.WriteString(m.styles.title.Render(abbrev))
	b.WriteString(" ")

	todayKey := dateKey(m.today)
	cursorKey := dateKey(m.cursor)

	var anchorKey time.Time
	if m.rangeAnchor != nil {
		anchorKey = dateKey(*m.rangeAnchor)
	}

	firstDay := true
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if !firstDay {
			b.WriteString(" ")
		}

		dayStr := fmt.Sprintf("%2d", d.Day())
		dk := dateKey(d)
		inMonth := !d.Before(first) && !d.After(last)

		isCursor := dk.Equal(cursorKey)
		isToday := dk.Equal(todayKey)
		isHighlighted := m.highlightedDates[dk]
		isRangeDay := false
		if m.rangeAnchor != nil {
			isRangeDay = isInRange(dk, anchorKey, cursorKey)
		}

		switch {
		case !inMonth:
			dayStr = m.styles.weekNum.Render(dayStr)
		case isCursor && isToday:
			dayStr = m.styles.cursorToday.Render(dayStr)
		case isCursor:
			dayStr = m.styles.cursor.Render(dayStr)
		case isToday:
			dayStr = m.styles.today.Render(dayStr)
		case isRangeDay:
			dayStr = m.styles.rangeDay.Render(dayStr)
		case isHighlighted:
			dayStr = m.styles.highlight.Render(dayStr)
		}

		b.WriteString(dayStr)
		firstDay = false
	}
	return b.String()
}
