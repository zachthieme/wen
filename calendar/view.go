package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

func applyColor(s lipgloss.Style, color string) lipgloss.Style {
	if color != "" {
		return s.Foreground(lipgloss.Color(color))
	}
	return s
}

func buildStyles(colors ThemeColors) resolvedStyles {
	cursorStyle := applyColor(lipgloss.NewStyle().Reverse(true), colors.Cursor)
	todayStyle := applyColor(lipgloss.NewStyle().Bold(true).Underline(true), colors.Today)
	// Pre-compose cursor+today so View() avoids nested Render calls and
	// the double-reset ANSI sequences they produce.
	cursorTodayStyle := lipgloss.NewStyle().Reverse(true).Bold(true).Underline(true)
	if colors.Cursor != "" {
		cursorTodayStyle = cursorTodayStyle.Foreground(lipgloss.Color(colors.Cursor))
	}
	return resolvedStyles{
		cursor:      cursorStyle,
		cursorToday: cursorTodayStyle,
		today:       todayStyle,
		title:       applyColor(lipgloss.NewStyle().Bold(true), colors.Title),
		weekNum:     applyColor(lipgloss.NewStyle().Faint(true), colors.WeekNumber),
		dayHeader:   applyColor(lipgloss.NewStyle().Faint(true), colors.DayHeader),
		helpBar:     applyColor(lipgloss.NewStyle().Faint(true), colors.HelpBar),
	}
}

func newHelpModel(colors ThemeColors) help.Model {
	h := help.New()
	h.ShowAll = true

	helpStyle := applyColor(lipgloss.NewStyle().Faint(true), colors.HelpBar)

	h.Styles.ShortKey = helpStyle
	h.Styles.ShortDesc = helpStyle
	h.Styles.ShortSeparator = helpStyle
	h.Styles.FullKey = helpStyle
	h.Styles.FullDesc = helpStyle
	h.Styles.FullSeparator = helpStyle

	return h
}

var dayNames = [7]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}

// View produces the calendar view string for the model state.
func (m Model) View() string {
	st := m.styles
	var b strings.Builder
	year, month, _ := m.cursor.Date()
	loc := m.cursor.Location()
	startDay := m.config.WeekStartDay

	// Title
	gridWidth := 20
	if m.showWeekNumbers {
		gridWidth = 23
	}
	title := fmt.Sprintf("%s %d", month, year)
	padding := max((gridWidth-len(title))/2, 0)
	b.WriteString(st.title.Render(strings.Repeat(" ", padding) + title))
	b.WriteString("\n")

	// Day headers
	if m.showWeekNumbers {
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
	if m.showWeekNumbers {
		wn := weekNumber(first, m.config.WeekNumbering)
		b.WriteString(st.weekNum.Render(fmt.Sprintf("%2d", wn)) + " ")
	}

	// Leading spaces
	b.WriteString(strings.Repeat("   ", weekday))

	_, _, cursorDay := m.cursor.Date()
	todayYear, todayMonth, todayDay := m.today.Date()

	for day := 1; day <= days; day++ {
		dayStr := fmt.Sprintf("%2d", day)

		isCursor := day == cursorDay
		isToday := year == todayYear && month == todayMonth && day == todayDay

		switch {
		case isCursor && isToday:
			dayStr = st.cursorToday.Render(dayStr)
		case isCursor:
			dayStr = st.cursor.Render(dayStr)
		case isToday:
			dayStr = st.today.Render(dayStr)
		}

		b.WriteString(dayStr)

		col := (weekday + day) % 7
		if col == 0 && day < days {
			b.WriteString("\n")
			if m.showWeekNumbers {
				nextDay := time.Date(year, month, day+1, 0, 0, 0, 0, loc)
				wn := weekNumber(nextDay, m.config.WeekNumbering)
				b.WriteString(st.weekNum.Render(fmt.Sprintf("%2d", wn)) + " ")
			}
		} else if day < days {
			b.WriteString(" ")
		}
	}

	b.WriteString("\n")

	if m.statusMsg != "" {
		b.WriteString(st.helpBar.Render(m.statusMsg))
		b.WriteString("\n")
	}

	if m.showHelp {
		b.WriteString("\n")
		b.WriteString(m.help.View(m.keys))
		b.WriteString("\n")
	}

	output := b.String()
	if m.config.PaddingTop > 0 || m.config.PaddingRight > 0 || m.config.PaddingBottom > 0 || m.config.PaddingLeft > 0 {
		padStyle := lipgloss.NewStyle().Padding(
			m.config.PaddingTop,
			m.config.PaddingRight,
			m.config.PaddingBottom,
			m.config.PaddingLeft,
		)
		output = padStyle.Render(output)
	}
	return output
}

func weekNumber(t time.Time, numbering string) int {
	if numbering == WeekNumberingISO {
		_, wn := t.ISOWeek()
		return wn
	}
	jan1 := time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location())
	dayOfYear := t.YearDay()
	jan1Weekday := int(jan1.Weekday())
	return (dayOfYear+jan1Weekday-1)/7 + 1
}
