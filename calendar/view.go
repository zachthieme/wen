package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/zachthieme/wen"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

// dateKey normalises a time.Time to a UTC midnight key suitable for map lookups.
func dateKey(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

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
	// rangeDay uses Reverse as a default (visible without color).
	// When a Range color is set, Foreground replaces Reverse entirely
	// for a cleaner look — intentionally different from other styles
	// which layer color on top of their base treatment.
	rangeDayStyle := lipgloss.NewStyle().Reverse(true)
	if colors.Range != "" {
		rangeDayStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Range))
	}
	return resolvedStyles{
		cursor:      cursorStyle,
		cursorToday: cursorTodayStyle,
		today:       todayStyle,
		highlight:   highlightStyle,
		rangeDay:    rangeDayStyle,
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

// wrapWithWeekNums takes lines rendered at dayGridWidth and prepends/appends
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

func (m Model) renderSingleMonth() string {
	var core strings.Builder
	year, month, cursorDay := m.cursor.Date()
	loc := m.cursor.Location()

	m.renderTitle(&core, month, year)
	m.renderDayHeaders(&core)
	gridWNs := m.renderGrid(&core, year, month, cursorDay, loc)
	m.renderQuarterBar(&core, dayGridWidth)

	// Split into lines, apply week numbers, reassemble.
	coreLines := strings.Split(strings.TrimRight(core.String(), "\n"), "\n")

	// Week numbers align with: [0]=title (blank), [1]=headers ("Wk"), [2..]=grid rows
	var wnLines []string
	fmtWN := func(s string) string {
		if m.weekNumPos == WeekNumLeft {
			return s + " "
		}
		return " " + s
	}
	wnLines = append(wnLines, "") // title — no week number
	if m.weekNumPos != WeekNumOff {
		wnLines = append(wnLines, fmtWN(m.styles.weekNum.Render("Wk")))
	} else {
		wnLines = append(wnLines, "")
	}
	for _, wn := range gridWNs {
		wnLines = append(wnLines, fmtWN(m.styles.weekNum.Render(fmt.Sprintf("%2d", wn))))
	}
	// Pad remaining lines (quarter bar, etc.) with blanks
	for len(wnLines) < len(coreLines) {
		wnLines = append(wnLines, "")
	}

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
	output = m.styles.padding.Render(output)
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

		// Build week number annotations for this month column.
		var wnLines []string
		fmtWN := func(s string) string {
			if m.weekNumPos == WeekNumLeft {
				return s + " "
			}
			return " " + s
		}
		wnLines = append(wnLines, "") // title
		if m.weekNumPos != WeekNumOff {
			wnLines = append(wnLines, fmtWN(m.styles.weekNum.Render("Wk")))
		} else {
			wnLines = append(wnLines, "")
		}
		for _, wn := range gridWNs {
			wnLines = append(wnLines, fmtWN(m.styles.weekNum.Render(fmt.Sprintf("%2d", wn))))
		}
		for len(wnLines) < len(coreLines) {
			wnLines = append(wnLines, "")
		}

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
	colWidth := dayGridWidth
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
	output = m.styles.padding.Render(output)
	return output
}

func (m Model) renderTitle(b *strings.Builder, month time.Month, year int) {
	hasFQ := m.config.ShowFiscalQuarter && m.config.FiscalYearStart > 1
	// Use 3-letter month abbreviation when fiscal quarter is shown to fit within dayGridWidth.
	monthName := month.String()
	if hasFQ {
		monthName = monthName[:3]
	}
	var title string
	if year == m.today.Year() {
		title = monthName
	} else {
		title = fmt.Sprintf("%s %d", monthName, year)
	}
	if hasFQ {
		q, fy := wen.FiscalQuarter(int(month), year, m.config.FiscalYearStart)
		title += fmt.Sprintf(" · Q%d FY%02d", q, fy%100)
	}
	titleStyle := m.styles.title.Width(dayGridWidth).Align(lipgloss.Center)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")
}

func (m Model) renderDayHeaders(b *strings.Builder) {
	startDay := m.config.WeekStartDay
	headers := make([]string, 7)
	for i := range 7 {
		headers[i] = dayNames[(startDay+i)%7]
	}
	b.WriteString(m.styles.dayHeader.Render(strings.Join(headers, " ")))
	b.WriteString("\n")
}

func isInRange(d, a, b time.Time) bool {
	if a.After(b) {
		a, b = b, a
	}
	return !d.Before(a) && !d.After(b)
}

// renderGrid renders the day grid to b at dayGridWidth, returning per-row week numbers.
func (m Model) renderGrid(b *strings.Builder, year int, month time.Month, cursorDay int, loc *time.Location) []int {
	st := m.styles
	startDay := m.config.WeekStartDay
	first := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	weekday := (int(first.Weekday()) - startDay + 7) % 7
	days := daysInMonth(year, month, loc)

	wn := weekNumber(first, m.config.WeekNumbering)
	var weekNums []int
	weekNums = append(weekNums, wn)

	// Leading spaces for first partial week
	b.WriteString(strings.Repeat("   ", weekday))

	todayYear, todayMonth, todayDay := m.today.Date()

	for day := 1; day <= days; day++ {
		dayStr := fmt.Sprintf("%2d", day)

		isCursor := day == cursorDay
		isToday := year == todayYear && month == todayMonth && day == todayDay
		isHighlighted := m.highlightedDates[dateKey(time.Date(year, month, day, 0, 0, 0, 0, loc))]

		isRangeDay := false
		if m.rangeAnchor != nil {
			dayDate := dateKey(time.Date(year, month, day, 0, 0, 0, 0, loc))
			anchorUTC := dateKey(*m.rangeAnchor)
			cursorUTC := dateKey(m.cursor)
			isRangeDay = isInRange(dayDate, anchorUTC, cursorUTC)
		}

		switch {
		case isCursor && isToday:
			dayStr = st.cursorToday.Render(dayStr)
		case isCursor:
			dayStr = st.cursor.Render(dayStr)
		case isToday:
			dayStr = st.today.Render(dayStr)
		case isRangeDay:
			dayStr = st.rangeDay.Render(dayStr)
		case isHighlighted:
			dayStr = st.highlight.Render(dayStr)
		}

		b.WriteString(dayStr)

		col := (weekday + day) % 7
		if col == 0 && day < days {
			nextDay := time.Date(year, month, day+1, 0, 0, 0, 0, loc)
			wn = weekNumber(nextDay, m.config.WeekNumbering)
			weekNums = append(weekNums, wn)
			b.WriteString("\n")
		} else if day < days {
			b.WriteString(" ")
		}
	}
	// Pad the last row to dayGridWidth so week numbers align.
	lastCol := (weekday + days) % 7
	if lastCol != 0 {
		b.WriteString(strings.Repeat("   ", 7-lastCol))
	}
	b.WriteString("\n")
	return weekNums
}

func quarterStartDate(cursor time.Time, fiscalYearStart int) time.Time {
	if fiscalYearStart < 1 || fiscalYearStart > 12 {
		fiscalYearStart = 1
	}
	month := int(cursor.Month())
	year := cursor.Year()
	fyCalStart := year
	if month < fiscalYearStart {
		fyCalStart = year - 1
	}
	q, _ := wen.FiscalQuarter(month, year, fiscalYearStart)
	startMonth := fiscalYearStart + (q-1)*3
	startYear := fyCalStart
	for startMonth > 12 {
		startMonth -= 12
		startYear++
	}
	return time.Date(startYear, time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC)
}

// countQuarterWorkdaysLeft counts workdays remaining from the day after cursor through qEnd (inclusive).
func countQuarterWorkdaysLeft(cursor, qEnd time.Time) int {
	count := 0
	d := cursor.AddDate(0, 0, 1) // start from day after cursor
	for !d.After(qEnd) {
		wd := d.Weekday()
		if wd != time.Saturday && wd != time.Sunday {
			count++
		}
		d = d.AddDate(0, 0, 1)
	}
	return count
}

func (m Model) renderQuarterBar(b *strings.Builder, width int) {
	if !m.config.ShowQuarterBar {
		return
	}
	fyStart := m.config.FiscalYearStart
	if fyStart < 1 {
		fyStart = 1
	}
	q, _ := wen.FiscalQuarter(int(m.cursor.Month()), m.cursor.Year(), fyStart)
	qStart := quarterStartDate(m.cursor, fyStart)
	qEnd := time.Date(qStart.Year(), qStart.Month()+3, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
	cursorUTC := time.Date(m.cursor.Year(), m.cursor.Month(), m.cursor.Day(), 0, 0, 0, 0, time.UTC)
	daysElapsed := int(cursorUTC.Sub(qStart).Hours()/24) + 1
	totalDays := int(qEnd.Sub(qStart).Hours()/24) + 1
	if totalDays <= 0 {
		totalDays = 1
	}
	progress := float64(daysElapsed) / float64(totalDays)
	progress = max(0, min(1, progress))

	workdaysLeft := countQuarterWorkdaysLeft(cursorUTC, qEnd)

	// Format: "Q1 ████████░░░░ 23wd"
	// Label and suffix are variable-width; bar fills the remainder.
	label := fmt.Sprintf("Q%d ", q)
	suffix := fmt.Sprintf(" %dwd", workdaysLeft)
	bw := width - len(label) - len(suffix)
	if bw < 4 {
		bw = 4
	}

	filled := int(progress * float64(bw))
	empty := bw - filled

	// Center the bar within the given width.
	barContent := label + strings.Repeat("█", filled) + strings.Repeat("░", empty) + suffix
	barLen := lipgloss.Width(barContent)
	leftPad := 0
	if barLen < width {
		leftPad = (width - barLen) / 2
	}

	if leftPad > 0 {
		b.WriteString(strings.Repeat(" ", leftPad))
	}
	b.WriteString(m.styles.title.Render(label))
	b.WriteString(m.styles.title.Render(strings.Repeat("█", filled)))
	b.WriteString(m.styles.weekNum.Render(strings.Repeat("░", empty)))
	b.WriteString(m.styles.title.Render(suffix))
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
