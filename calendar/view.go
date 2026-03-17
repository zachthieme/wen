package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	cursorStyle = lipgloss.NewStyle().Reverse(true)
	todayStyle  = lipgloss.NewStyle().Bold(true).Underline(true)
	titleStyle  = lipgloss.NewStyle().Bold(true)
)

func Render(m Model) string {
	var b strings.Builder

	year, month, _ := m.Cursor.Date()

	// Title
	title := fmt.Sprintf("%s %d", month, year)
	padding := (20 - len(title)) / 2
	if padding < 0 {
		padding = 0
	}
	b.WriteString(titleStyle.Render(strings.Repeat(" ", padding) + title))
	b.WriteString("\n")

	// Day headers
	b.WriteString("Su Mo Tu We Th Fr Sa\n")

	// First day of month
	first := time.Date(year, month, 1, 0, 0, 0, 0, m.Cursor.Location())
	weekday := int(first.Weekday())
	days := daysInMonth(year, month)

	// Leading spaces
	b.WriteString(strings.Repeat("   ", weekday))

	for day := 1; day <= days; day++ {
		current := time.Date(year, month, day, 0, 0, 0, 0, m.Cursor.Location())
		dayStr := fmt.Sprintf("%2d", day)

		isCursor := current.Equal(m.Cursor)
		isToday := current.Equal(m.Today)

		switch {
		case isCursor && isToday:
			dayStr = cursorStyle.Render(todayStyle.Render(dayStr))
		case isCursor:
			dayStr = cursorStyle.Render(dayStr)
		case isToday:
			dayStr = todayStyle.Render(dayStr)
		}

		b.WriteString(dayStr)

		col := (weekday + day) % 7
		if col == 0 && day < days {
			b.WriteString("\n")
		} else if day < days {
			b.WriteString(" ")
		}
	}

	b.WriteString("\n")
	return b.String()
}
