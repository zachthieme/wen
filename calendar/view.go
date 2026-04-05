package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// monthGap is the spacing between side-by-side months.
const monthGap = "   "

// View produces the calendar view string for the model state.
func (m Model) View() string {
	if m.months <= 1 {
		return m.renderSingleMonth()
	}
	return m.renderMultiMonth()
}

// wrapWithWeekNums takes lines rendered at the current grid width and prepends/appends
// week number columns. Lines without a corresponding week number get blank padding.
func (m Model) wrapWithWeekNums(lines []string, weekNums []string) []string {
	if m.weekNumPos == WeekNumOff {
		return lines
	}
	wnWidth := 3 // "Wk" or " N" padded to 2 chars + 1 space
	result := make([]string, len(lines))
	for i, line := range lines {
		wn := ""
		if i < len(weekNums) && weekNums[i] != "" {
			wn = weekNums[i]
		} else {
			wn = strings.Repeat(" ", wnWidth)
		}
		if m.weekNumPos == WeekNumLeft {
			result[i] = wn + line
		} else {
			result[i] = line + wn
		}
	}
	return result
}

// buildWeekNumLines formats week number annotations for a month column.
func (m Model) buildWeekNumLines(gridWNs []int, coreLineCount int) []string {
	fmtWN := func(s string) string {
		if m.weekNumPos == WeekNumLeft {
			return s + " "
		}
		return " " + s
	}
	var wnLines []string
	wnLines = append(wnLines, "") // title — no week number
	if m.weekNumPos != WeekNumOff {
		wnLines = append(wnLines, fmtWN(m.styles.weekNum.Render("Wk")))
	} else {
		wnLines = append(wnLines, "")
	}
	for _, wn := range gridWNs {
		wnLines = append(wnLines, fmtWN(m.styles.weekNum.Render(fmt.Sprintf("%2d", wn))))
	}
	for len(wnLines) < coreLineCount {
		wnLines = append(wnLines, "")
	}
	return wnLines
}

func (m Model) renderSingleMonth() string {
	var core strings.Builder
	year, month, cursorDay := m.cursor.Date()
	loc := m.cursor.Location()

	m.renderTitle(&core, month, year)
	m.renderDayHeaders(&core)
	gridWNs := m.renderGrid(&core, year, month, cursorDay, loc)
	m.renderQuarterBar(&core, m.dayFmt.gridWidth)

	coreLines := strings.Split(strings.TrimRight(core.String(), "\n"), "\n")
	wnLines := m.buildWeekNumLines(gridWNs, len(coreLines))
	wrapped := m.wrapWithWeekNums(coreLines, wnLines)

	var b strings.Builder
	b.WriteString(strings.Join(wrapped, "\n"))
	b.WriteString("\n")

	if m.showHelp {
		b.WriteString("\n")
		b.WriteString(m.help.View(m.keys))
		b.WriteString("\n")
	}

	output := b.String()
	if m.termWidth > 0 && m.termHeight > 0 {
		return lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Center, lipgloss.Center, output)
	}
	return output
}

func (m Model) renderMultiMonth() string {
	cursorYear, cursorMonth, cursorDay := m.cursor.Date()
	loc := m.cursor.Location()

	// Determine starting month offset: center the cursor month
	startOffset := -(m.months / 2)

	// Render each month into lines, applying week numbers per column.
	monthLines := make([][]string, m.months)
	for i := range m.months {
		var core strings.Builder
		t := time.Date(cursorYear, cursorMonth+time.Month(startOffset+i), 1, 0, 0, 0, 0, loc)
		y, mo, _ := t.Date()
		cd := 0
		if y == cursorYear && mo == cursorMonth {
			cd = cursorDay
		}
		m.renderTitle(&core, mo, y)
		m.renderDayHeaders(&core)
		gridWNs := m.renderGrid(&core, y, mo, cd, loc)

		coreLines := strings.Split(strings.TrimRight(core.String(), "\n"), "\n")
		wnLines := m.buildWeekNumLines(gridWNs, len(coreLines))
		monthLines[i] = m.wrapWithWeekNums(coreLines, wnLines)
	}

	// Find max lines across all months
	maxLines := 0
	for _, lines := range monthLines {
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	// Column width includes week numbers if enabled.
	colWidth := m.dayFmt.gridWidth
	if m.weekNumPos != WeekNumOff {
		colWidth += 3 // " Wk" or "Wk " = 2 chars + 1 space
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
				result.WriteString(line)
				if i < len(monthLines)-1 {
					visible := lipgloss.Width(line)
					if visible < colWidth {
						result.WriteString(strings.Repeat(" ", colWidth-visible))
					}
				}
			} else if i < len(monthLines)-1 {
				result.WriteString(strings.Repeat(" ", colWidth))
			}
		}
		result.WriteString("\n")
	}
	totalWidth := colWidth*m.months + len(monthGap)*(m.months-1)
	m.renderQuarterBar(&result, totalWidth)

	if m.showHelp {
		result.WriteString("\n")
		result.WriteString(m.help.View(m.keys))
		result.WriteString("\n")
	}

	output := result.String()
	if m.termWidth > 0 && m.termHeight > 0 {
		return lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Center, lipgloss.Center, output)
	}
	return output
}
