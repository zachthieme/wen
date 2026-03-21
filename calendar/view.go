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
	highlightStyle := lipgloss.NewStyle().Bold(true)
	if colors.Highlight != "" {
		highlightStyle = highlightStyle.Foreground(lipgloss.Color(colors.Highlight))
	} else {
		highlightStyle = highlightStyle.Underline(true)
	}
	return resolvedStyles{
		cursor:      cursorStyle,
		cursorToday: cursorTodayStyle,
		today:       todayStyle,
		highlight:   highlightStyle,
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

// dayGridWidth is the character width of the 7-column day grid.
const dayGridWidth = 20

// monthGap is the spacing between side-by-side months.
const monthGap = "   "

// View produces the calendar view string for the model state.
func (m Model) View() string {
	if m.months <= 1 {
		return m.renderSingleMonth()
	}
	return m.renderMultiMonth()
}

func (m Model) renderSingleMonth() string {
	var b strings.Builder
	year, month, cursorDay := m.cursor.Date()
	loc := m.cursor.Location()

	m.renderTitle(&b, month, year)
	m.renderDayHeaders(&b)
	m.renderGrid(&b, year, month, cursorDay, loc)

	if m.showHelp {
		b.WriteString("\n")
		b.WriteString(m.help.View(m.keys))
		b.WriteString("\n")
	}

	output := b.String()
	if m.styles.hasPadding {
		output = m.styles.padding.Render(output)
	}
	return output
}

func (m Model) renderMultiMonth() string {
	cursorYear, cursorMonth, cursorDay := m.cursor.Date()
	loc := m.cursor.Location()

	// Determine starting month offset: center the cursor month
	startOffset := -(m.months / 2)

	// Render each month into lines
	monthLines := make([][]string, m.months)
	for i := range m.months {
		var b strings.Builder
		t := time.Date(cursorYear, cursorMonth+time.Month(startOffset+i), 1, 0, 0, 0, 0, loc)
		y, mo, _ := t.Date()
		cd := 0
		if y == cursorYear && mo == cursorMonth {
			cd = cursorDay
		}
		m.renderTitle(&b, mo, y)
		m.renderDayHeaders(&b)
		m.renderGrid(&b, y, mo, cd, loc)
		monthLines[i] = strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
	}

	// Find max lines across all months
	maxLines := 0
	for _, lines := range monthLines {
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	// Join side by side
	var result strings.Builder
	for row := range maxLines {
		for i, lines := range monthLines {
			if i > 0 {
				result.WriteString(monthGap)
			}
			if row < len(lines) {
				line := lines[row]
				// Pad to consistent width for alignment
				result.WriteString(line)
				// Pad with spaces to dayGridWidth for non-last months
				if i < len(monthLines)-1 {
					visible := visibleLen(line)
					if visible < dayGridWidth {
						result.WriteString(strings.Repeat(" ", dayGridWidth-visible))
					}
				}
			} else if i < len(monthLines)-1 {
				result.WriteString(strings.Repeat(" ", dayGridWidth))
			}
		}
		result.WriteString("\n")
	}

	if m.showHelp {
		result.WriteString("\n")
		result.WriteString(m.help.View(m.keys))
		result.WriteString("\n")
	}

	output := result.String()
	if m.styles.hasPadding {
		output = m.styles.padding.Render(output)
	}
	return output
}

// visibleLen returns the visible character length, stripping ANSI escape sequences.
func visibleLen(s string) int {
	n := 0
	inEsc := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z') {
				inEsc = false
			}
			continue
		}
		n++
	}
	return n
}

func (m Model) renderTitle(b *strings.Builder, month time.Month, year int) {
	title := fmt.Sprintf("%s %d", month, year)
	padding := max((dayGridWidth-len(title))/2, 0)
	b.WriteString(m.styles.title.Render(strings.Repeat(" ", padding) + title))
	b.WriteString("\n")
}

func (m Model) renderDayHeaders(b *strings.Builder) {
	startDay := m.config.WeekStartDay
	headers := make([]string, 7)
	for i := range 7 {
		headers[i] = dayNames[(startDay+i)%7]
	}
	b.WriteString(m.styles.dayHeader.Render(strings.Join(headers, " ")))
	if m.showWeekNumbers {
		b.WriteString(" " + m.styles.weekNum.Render("Wk"))
	}
	b.WriteString("\n")
}

func (m Model) renderGrid(b *strings.Builder, year int, month time.Month, cursorDay int, loc *time.Location) {
	st := m.styles
	startDay := m.config.WeekStartDay
	first := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	weekday := (int(first.Weekday()) - startDay + 7) % 7
	days := daysInMonth(year, month, loc)

	wn := 0
	if m.showWeekNumbers {
		wn = weekNumber(first, m.config.WeekNumbering)
	}

	// Leading spaces for first partial week
	b.WriteString(strings.Repeat("   ", weekday))

	todayYear, todayMonth, todayDay := m.today.Date()

	for day := 1; day <= days; day++ {
		dayStr := fmt.Sprintf("%2d", day)

		isCursor := day == cursorDay
		isToday := year == todayYear && month == todayMonth && day == todayDay
		isHighlighted := m.highlightedDates[time.Date(year, month, day, 0, 0, 0, 0, time.UTC)]

		switch {
		case isCursor && isToday:
			dayStr = st.cursorToday.Render(dayStr)
		case isCursor:
			dayStr = st.cursor.Render(dayStr)
		case isToday:
			dayStr = st.today.Render(dayStr)
		case isHighlighted:
			dayStr = st.highlight.Render(dayStr)
		}

		b.WriteString(dayStr)

		col := (weekday + day) % 7
		if col == 0 && day < days {
			// End of row — append week number, then newline
			if m.showWeekNumbers {
				b.WriteString(" " + st.weekNum.Render(fmt.Sprintf("%2d", wn)))
				nextDay := time.Date(year, month, day+1, 0, 0, 0, 0, loc)
				wn = weekNumber(nextDay, m.config.WeekNumbering)
			}
			b.WriteString("\n")
		} else if day < days {
			b.WriteString(" ")
		}
	}

	// Append week number to last row
	if m.showWeekNumbers {
		lastCol := (weekday + days) % 7
		if lastCol != 0 {
			b.WriteString(strings.Repeat("   ", 7-lastCol))
		}
		b.WriteString(" " + st.weekNum.Render(fmt.Sprintf("%2d", wn)))
	}
	b.WriteString("\n")
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
