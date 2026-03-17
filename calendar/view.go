package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func buildStyles(colors ThemeColors) resolvedStyles {
	s := resolvedStyles{
		cursor:    lipgloss.NewStyle().Reverse(true),
		today:     lipgloss.NewStyle().Bold(true).Underline(true),
		title:     lipgloss.NewStyle().Bold(true),
		weekNum:   lipgloss.NewStyle().Faint(true),
		dayHeader: lipgloss.NewStyle().Faint(true),
		helpBar:   lipgloss.NewStyle().Faint(true),
	}

	if colors.Cursor != "" {
		s.cursor = s.cursor.Foreground(lipgloss.Color(colors.Cursor))
	}
	if colors.Today != "" {
		s.today = s.today.Foreground(lipgloss.Color(colors.Today))
	}
	if colors.Title != "" {
		s.title = s.title.Foreground(lipgloss.Color(colors.Title))
	}
	if colors.WeekNumber != "" {
		s.weekNum = lipgloss.NewStyle().Foreground(lipgloss.Color(colors.WeekNumber))
	}
	if colors.DayHeader != "" {
		s.dayHeader = lipgloss.NewStyle().Foreground(lipgloss.Color(colors.DayHeader))
	}
	if colors.HelpBar != "" {
		s.helpBar = lipgloss.NewStyle().Foreground(lipgloss.Color(colors.HelpBar))
	}

	return s
}

var dayNames = [7]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}

func Render(m Model) string {
	st := m.styles
	var b strings.Builder
	year, month, _ := m.Cursor.Date()
	loc := m.Cursor.Location()
	startDay := m.Config.WeekStartDay

	// Title
	gridWidth := 20
	if m.ShowWeekNumbers {
		gridWidth = 23
	}
	title := fmt.Sprintf("%s %d", month, year)
	padding := max((gridWidth-len(title))/2, 0)
	b.WriteString(st.title.Render(strings.Repeat(" ", padding) + title))
	b.WriteString("\n")

	// Day headers
	if m.ShowWeekNumbers {
		b.WriteString(st.weekNum.Render("Wk") + " ")
	}
	headers := make([]string, 7)
	for i := range 7 {
		headers[i] = dayNames[(startDay+i)%7]
	}
	b.WriteString(st.dayHeader.Render(strings.Join(headers, " ")))
	b.WriteString("\n")

	// First day of month
	first := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	weekday := (int(first.Weekday()) - startDay + 7) % 7
	days := daysInMonth(year, month, loc)

	// Week number for first row
	if m.ShowWeekNumbers {
		wn := weekNumber(first, m.Config.WeekNumbering)
		b.WriteString(st.weekNum.Render(fmt.Sprintf("%2d", wn)) + " ")
	}

	// Leading spaces
	b.WriteString(strings.Repeat("   ", weekday))

	for day := 1; day <= days; day++ {
		current := time.Date(year, month, day, 0, 0, 0, 0, loc)
		dayStr := fmt.Sprintf("%2d", day)

		isCursor := current.Equal(m.Cursor)
		isToday := current.Equal(m.Today)

		switch {
		case isCursor && isToday:
			dayStr = st.cursor.Render(st.today.Render(dayStr))
		case isCursor:
			dayStr = st.cursor.Render(dayStr)
		case isToday:
			dayStr = st.today.Render(dayStr)
		}

		b.WriteString(dayStr)

		col := (weekday + day) % 7
		if col == 0 && day < days {
			b.WriteString("\n")
			if m.ShowWeekNumbers {
				nextDay := time.Date(year, month, day+1, 0, 0, 0, 0, loc)
				wn := weekNumber(nextDay, m.Config.WeekNumbering)
				b.WriteString(st.weekNum.Render(fmt.Sprintf("%2d", wn)) + " ")
			}
		} else if day < days {
			b.WriteString(" ")
		}
	}

	b.WriteString("\n")

	if m.ShowHelp {
		b.WriteString("\n")
		help := "h/l:day  j/k:week  H/L:month  J/K:year  t:today  y:yank  w:weeks  enter:select  q:quit"
		b.WriteString(st.helpBar.Render(help))
		b.WriteString("\n")
	}

	return b.String()
}

func weekNumber(t time.Time, numbering string) int {
	if numbering == "iso" {
		_, wn := t.ISOWeek()
		return wn
	}
	jan1 := time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location())
	dayOfYear := t.YearDay()
	jan1Weekday := int(jan1.Weekday())
	return (dayOfYear+jan1Weekday-1)/7 + 1
}
