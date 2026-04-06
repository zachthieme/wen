package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	// monthGap is the spacing between side-by-side months.
	monthGap = "   "
	// weekNumColWidth is the character width of the week number column ("Wk" or " N" padded to 2 chars + 1 space).
	weekNumColWidth = 3
)

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
	wnWidth := weekNumColWidth
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

// joinColumnsHorizontal joins multiple columns of lines side by side.
// Each column is padded to colWidth using lipgloss.Width for ANSI-aware
// measurement. Columns are separated by gap.
func joinColumnsHorizontal(columns [][]string, colWidth int, gap string) string {
	maxLines := 0
	for _, col := range columns {
		if len(col) > maxLines {
			maxLines = len(col)
		}
	}

	var result strings.Builder
	for row := range maxLines {
		for i, col := range columns {
			if i > 0 {
				result.WriteString(gap)
			}
			if row < len(col) {
				line := col[row]
				result.WriteString(line)
				if i < len(columns)-1 {
					visible := lipgloss.Width(line)
					if visible < colWidth {
						result.WriteString(strings.Repeat(" ", colWidth-visible))
					}
				}
			} else {
				result.WriteString(strings.Repeat(" ", colWidth))
			}
		}
		result.WriteString("\n")
	}
	return result.String()
}

func (m Model) renderMultiMonth() string {
	cursorYear, cursorMonth, cursorDay := m.cursor.Date()
	loc := m.cursor.Location()

	startOffset := -(m.months / 2)

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

	colWidth := m.dayFmt.gridWidth
	if m.weekNumPos != WeekNumOff {
		colWidth += weekNumColWidth
	}

	var result strings.Builder
	result.WriteString(joinColumnsHorizontal(monthLines, colWidth, monthGap))

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
