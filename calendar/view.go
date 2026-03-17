package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func buildStyles(colors ThemeColors) (cursor, today, title, weekNum, dayHeader, helpBar lipgloss.Style) {
	cursor = lipgloss.NewStyle().Reverse(true)
	today = lipgloss.NewStyle().Bold(true).Underline(true)
	title = lipgloss.NewStyle().Bold(true)
	weekNum = lipgloss.NewStyle().Faint(true)
	dayHeader = lipgloss.NewStyle().Faint(true)
	helpBar = lipgloss.NewStyle().Faint(true)

	if colors.Cursor != "" {
		cursor = cursor.Foreground(lipgloss.Color(colors.Cursor))
	}
	if colors.Today != "" {
		today = today.Foreground(lipgloss.Color(colors.Today))
	}
	if colors.Title != "" {
		title = title.Foreground(lipgloss.Color(colors.Title))
	}
	if colors.WeekNumber != "" {
		weekNum = lipgloss.NewStyle().Foreground(lipgloss.Color(colors.WeekNumber))
	}
	if colors.DayHeader != "" {
		dayHeader = lipgloss.NewStyle().Foreground(lipgloss.Color(colors.DayHeader))
	}
	if colors.HelpBar != "" {
		helpBar = lipgloss.NewStyle().Foreground(lipgloss.Color(colors.HelpBar))
	}

	return
}

var dayNames = [7]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}

func Render(m Model) string {
	colors := m.Config.ResolvedColors()
	cursorSt, todaySt, titleSt, weekNumSt, dayHeaderSt, helpBarSt := buildStyles(colors)

	var b strings.Builder
	year, month, _ := m.Cursor.Date()
	startDay := m.Config.WeekStartDay

	// Title
	gridWidth := 20
	if m.ShowWeekNumbers {
		gridWidth = 23
	}
	title := fmt.Sprintf("%s %d", month, year)
	padding := (gridWidth - len(title)) / 2
	if padding < 0 {
		padding = 0
	}
	b.WriteString(titleSt.Render(strings.Repeat(" ", padding) + title))
	b.WriteString("\n")

	// Day headers
	if m.ShowWeekNumbers {
		b.WriteString(weekNumSt.Render("Wk") + " ")
	}
	headers := make([]string, 7)
	for i := 0; i < 7; i++ {
		headers[i] = dayNames[(startDay+i)%7]
	}
	b.WriteString(dayHeaderSt.Render(strings.Join(headers, " ")))
	b.WriteString("\n")

	// First day of month
	first := time.Date(year, month, 1, 0, 0, 0, 0, m.Cursor.Location())
	weekday := (int(first.Weekday()) - startDay + 7) % 7
	days := daysInMonth(year, month)

	// Week number for first row
	if m.ShowWeekNumbers {
		wn := weekNumber(first, m.Config.WeekNumbering)
		b.WriteString(weekNumSt.Render(fmt.Sprintf("%2d", wn)) + " ")
	}

	// Leading spaces
	b.WriteString(strings.Repeat("   ", weekday))

	for day := 1; day <= days; day++ {
		current := time.Date(year, month, day, 0, 0, 0, 0, m.Cursor.Location())
		dayStr := fmt.Sprintf("%2d", day)

		isCursor := current.Equal(m.Cursor)
		isToday := current.Equal(m.Today)

		switch {
		case isCursor && isToday:
			dayStr = cursorSt.Render(todaySt.Render(dayStr))
		case isCursor:
			dayStr = cursorSt.Render(dayStr)
		case isToday:
			dayStr = todaySt.Render(dayStr)
		}

		b.WriteString(dayStr)

		col := (weekday + day) % 7
		if col == 0 && day < days {
			b.WriteString("\n")
			// Week number for next row
			if m.ShowWeekNumbers {
				nextDay := time.Date(year, month, day+1, 0, 0, 0, 0, m.Cursor.Location())
				wn := weekNumber(nextDay, m.Config.WeekNumbering)
				b.WriteString(weekNumSt.Render(fmt.Sprintf("%2d", wn)) + " ")
			}
		} else if day < days {
			b.WriteString(" ")
		}
	}

	b.WriteString("\n")

	// Help bar
	if m.ShowHelp {
		b.WriteString("\n")
		help := "h/l:day  j/k:week  H/L:month  J/K:year  t:today  y:yank  w:weeks  enter:select  q:quit"
		b.WriteString(helpBarSt.Render(help))
		b.WriteString("\n")
	}

	return b.String()
}

func weekNumber(t time.Time, numbering string) int {
	if numbering == "iso" {
		_, wn := t.ISOWeek()
		return wn
	}
	// US: week 1 contains Jan 1, weeks start Sunday
	jan1 := time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location())
	dayOfYear := t.YearDay()
	jan1Weekday := int(jan1.Weekday())
	return (dayOfYear+jan1Weekday-1)/7 + 1
}
