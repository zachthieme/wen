package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/zachthieme/wen"

	"github.com/charmbracelet/lipgloss"
)

var dayNames = [7]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}

// dayGridWidth is the character width of the 7-column day grid.
const dayGridWidth = 20

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

// dateKey normalises a time.Time to a UTC midnight key suitable for map lookups.
func dateKey(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
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

// countQuarterWorkdaysLeft counts workdays remaining from the day after cursor
// through qEnd (inclusive) using a closed-form calculation.
func countQuarterWorkdaysLeft(cursor, qEnd time.Time) int {
	start := cursor.AddDate(0, 0, 1) // day after cursor
	if start.After(qEnd) {
		return 0
	}
	totalDays := int(qEnd.Sub(start).Hours()/24) + 1
	fullWeeks := totalDays / 7
	remaining := totalDays % 7
	workdays := fullWeeks * 5
	startDow := int(start.Weekday()) // Sunday=0
	for i := range remaining {
		dow := (startDow + i) % 7
		if dow != int(time.Saturday) && dow != int(time.Sunday) {
			workdays++
		}
	}
	return workdays
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
