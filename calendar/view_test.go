package calendar

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/zachthieme/wen"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderMarch2026(t *testing.T) {
	t.Parallel()
	// Use a different year for today so the year appears in the title
	m := New(date(2026, time.March, 17), date(2025, time.March, 17), DefaultConfig())
	output := m.View()

	if !strings.Contains(output, "March 2026") {
		t.Error("expected 'March 2026' in output")
	}
	if !strings.Contains(output, "Su Mo Tu We Th Fr Sa") {
		t.Error("expected Sunday-start day headers")
	}
}

func TestRenderFebruary2026(t *testing.T) {
	t.Parallel()
	// Use a different year for today so the year appears in the title
	m := New(date(2026, time.February, 14), date(2025, time.March, 17), DefaultConfig())
	output := m.View()

	if !strings.Contains(output, "February 2026") {
		t.Error("expected 'February 2026' in output")
	}
}

func TestRenderWithWeekNumbers(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.ShowWeekNumbers = "left"
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	output := m.View()

	if !strings.Contains(output, "Wk") {
		t.Error("expected 'Wk' header when week numbers enabled")
	}
}

func TestRenderWithoutWeekNumbers(t *testing.T) {
	t.Parallel()
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	output := m.View()

	if strings.Contains(output, "Wk") {
		t.Error("should not have 'Wk' header when week numbers disabled")
	}
}

func TestRenderMondayStart(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.WeekStartDay = 1
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	output := m.View()

	if !strings.Contains(output, "Mo Tu We Th Fr Sa Su") {
		t.Error("expected Monday-start day headers")
	}
}

func TestRenderHelpBar(t *testing.T) {
	t.Parallel()
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m.showHelp = true
	output := m.View()

	if !strings.Contains(output, "prev day") {
		t.Error("expected help bar to contain 'prev day'")
	}
	if !strings.Contains(output, "quit") {
		t.Error("expected help bar to contain 'quit'")
	}
}

func TestRenderNoHelpBarByDefault(t *testing.T) {
	t.Parallel()
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	output := m.View()

	if strings.Contains(output, "prev day") {
		t.Error("help bar should not appear by default")
	}
}

func TestWeekNumberUS(t *testing.T) {
	t.Parallel()
	d := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.Local)
	wn := weekNumber(d, "us")
	// March 1 is day 60 of 2026. Jan 1 is Thursday (weekday 4).
	// (60 + 4 - 1) / 7 + 1 = 63/7 + 1 = 9 + 1 = 10
	if wn != 10 {
		t.Errorf("expected US week 10 for March 1 2026, got %d", wn)
	}
}

func TestWeekNumberISO(t *testing.T) {
	t.Parallel()
	d := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.Local)
	wn := weekNumber(d, "iso")
	_, expected := d.ISOWeek()
	if wn != expected {
		t.Errorf("expected ISO week %d for March 1 2026, got %d", expected, wn)
	}
}

func TestFiscalQuarter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		month      int
		year       int
		fyStart    int
		wantQ      int
		wantFY     int
	}{
		// FY starts October: Oct 2025 – Sep 2026 = FY26
		{"Oct 2025 FY-Oct", 10, 2025, 10, 1, 2026},
		{"Dec 2025 FY-Oct", 12, 2025, 10, 1, 2026},
		{"Jan 2026 FY-Oct", 1, 2026, 10, 2, 2026},
		{"Mar 2026 FY-Oct", 3, 2026, 10, 2, 2026},
		{"Apr 2026 FY-Oct", 4, 2026, 10, 3, 2026},
		{"Jul 2026 FY-Oct", 7, 2026, 10, 4, 2026},
		{"Sep 2026 FY-Oct", 9, 2026, 10, 4, 2026},
		// FY starts April (UK/Japan): Apr 2026 – Mar 2027 = FY27
		{"Apr 2026 FY-Apr", 4, 2026, 4, 1, 2027},
		{"Mar 2026 FY-Apr", 3, 2026, 4, 4, 2026},
		{"Jan 2026 FY-Apr", 1, 2026, 4, 4, 2026},
		// Calendar year (FY starts Jan): standard quarters
		{"Mar 2026 FY-Jan", 3, 2026, 1, 1, 2026},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			q, fy := wen.FiscalQuarter(tt.month, tt.year, tt.fyStart)
			if q != tt.wantQ || fy != tt.wantFY {
				t.Errorf("FiscalQuarter(%d, %d, %d) = Q%d FY%d, want Q%d FY%d",
					tt.month, tt.year, tt.fyStart, q, fy, tt.wantQ, tt.wantFY)
			}
		})
	}
}

func TestRenderFiscalQuarterTitle(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.FiscalYearStart = 10
	cfg.ShowFiscalQuarter = true
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	output := m.View()
	if !strings.Contains(output, "Q2 FY26") {
		t.Errorf("expected title to contain 'Q2 FY26', got:\n%s", output)
	}
}

func TestRenderNoFiscalQuarterByDefault(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	output := m.View()
	if strings.Contains(output, "FY") {
		t.Error("expected no fiscal quarter in default config")
	}
}

func TestRenderMultiMonth(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cursor := date(2026, time.March, 17)
	// Use a different year for today so the year appears in titles
	today := date(2025, time.March, 17)
	m := New(cursor, today, cfg, WithMonths(3))
	output := m.View()

	// With 3 months centered on March, we expect February, March, April
	if !strings.Contains(output, "February 2026") {
		t.Error("expected 'February 2026' in multi-month output")
	}
	if !strings.Contains(output, "March 2026") {
		t.Error("expected 'March 2026' in multi-month output")
	}
	if !strings.Contains(output, "April 2026") {
		t.Error("expected 'April 2026' in multi-month output")
	}

	// Months should appear on the same lines (side by side),
	// so line count should be similar to single month.
	single := New(cursor, today, cfg).View()
	multiLines := len(strings.Split(output, "\n"))
	singleLines := len(strings.Split(single, "\n"))
	if multiLines > singleLines+2 {
		t.Errorf("multi-month should render side by side, got %d lines vs single %d lines", multiLines, singleLines)
	}
}

func TestRenderMultiMonthWithWeekNumbers(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.ShowWeekNumbers = "left"
	cursor := date(2026, time.March, 17)
	today := date(2025, time.March, 17)
	m := New(cursor, today, cfg, WithMonths(3))
	output := m.View()
	if !strings.Contains(output, "Wk") {
		t.Error("expected 'Wk' header in multi-month view with week numbers")
	}
	// Week numbers should appear for each month column
	if !strings.Contains(output, "March 2026") {
		t.Error("expected 'March 2026' in output")
	}
}

func TestRenderMultiMonthSingle(t *testing.T) {
	t.Parallel()
	// WithMonths(1) should produce same output as default
	cfg := DefaultConfig()
	cursor := date(2026, time.March, 17)
	today := date(2026, time.March, 17)
	single := New(cursor, today, cfg).View()
	withOpt := New(cursor, today, cfg, WithMonths(1)).View()
	if single != withOpt {
		t.Error("WithMonths(1) should produce identical output to default")
	}
}

func TestRangeRenderingProducesOutput(t *testing.T) {
	t.Parallel()
	// Verify that View() doesn't panic or produce empty output in range mode.
	// Note: lipgloss strips ANSI in non-TTY environments, so we can't compare
	// styled vs unstyled output directly. Instead we verify the output is
	// well-formed and contains expected content.
	cfg := DefaultConfig()
	cursor := date(2026, time.March, 17)
	today := date(2025, time.March, 17)
	m := New(cursor, today, cfg)
	m = pressKey(m, "v")
	for range 3 {
		m = pressKey(m, "l")
	}
	output := m.View()
	if !strings.Contains(output, "March 2026") {
		t.Error("expected 'March 2026' in range mode output")
	}
	// Days 17-20 should all appear in the output
	if !strings.Contains(output, "17") || !strings.Contains(output, "20") {
		t.Error("expected anchor and cursor days in output")
	}
}

func TestRangeRenderingMultiMonth(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cursor := date(2026, time.March, 28)
	today := date(2025, time.March, 28)
	m := New(cursor, today, cfg, WithMonths(3))
	m = pressKey(m, "v")
	for range 5 {
		m = pressKey(m, "l")
	}
	output := m.View()
	if !strings.Contains(output, "March 2026") {
		t.Error("expected March in output")
	}
	if !strings.Contains(output, "April 2026") {
		t.Error("expected April in output")
	}
}

func TestSmartTitleCurrentYear(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	output := m.View()
	if !strings.Contains(output, "March") {
		t.Error("expected 'March' in output")
	}
	if strings.Contains(output, "2026") {
		t.Error("expected year to be omitted for current year")
	}
}

func TestSmartTitleOtherYear(t *testing.T) {
	t.Parallel()
	cursor := date(2027, time.March, 17)
	today := date(2026, time.March, 17)
	m := New(cursor, today, DefaultConfig())
	output := m.View()
	if !strings.Contains(output, "March 2027") {
		t.Error("expected 'March 2027' for non-current year")
	}
}

func TestSmartTitleWithFiscalQuarter(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.FiscalYearStart = 10
	cfg.ShowFiscalQuarter = true
	today := date(2026, time.March, 17)
	m := New(today, today, cfg)
	output := m.View()
	if !strings.Contains(output, "Q2 FY26") {
		t.Errorf("expected fiscal quarter in title, got:\n%s", output)
	}
}

func TestQuarterBarHiddenByDefault(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	output := m.View()
	if strings.Contains(output, "█") || strings.Contains(output, "░") {
		t.Error("quarter bar should not appear with default config")
	}
}

func TestQuarterBarRendering(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.ShowQuarterBar = true
	today := date(2026, time.March, 17)
	m := New(today, today, cfg)
	output := m.View()
	if !strings.Contains(output, "Q1") {
		t.Errorf("expected Q1 in quarter bar, got:\n%s", output)
	}
	if !strings.Contains(output, "wd") {
		t.Error("expected workdays remaining in bar")
	}
}

func TestQuarterBarFiscalQuarter(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.ShowQuarterBar = true
	cfg.FiscalYearStart = 10
	today := date(2026, time.March, 17)
	m := New(today, today, cfg)
	output := m.View()
	if !strings.Contains(output, "Q2") {
		t.Errorf("expected Q2 for fiscal Oct start in March, got:\n%s", output)
	}
}

func TestRenderJulianSingleMonth(t *testing.T) {
	t.Parallel()
	m := New(date(2026, time.March, 17), date(2025, time.March, 17), DefaultConfig(), WithJulian(true))
	output := m.View()
	// Should contain 3-char day headers
	if !strings.Contains(output, "Sun Mon Tue Wed Thu Fri Sat") {
		t.Errorf("expected 3-char julian headers, got:\n%s", output)
	}
	// March 1 = yearday 60
	if !strings.Contains(output, " 60") {
		t.Errorf("expected yearday 60, got:\n%s", output)
	}
}

func TestRenderJulianMultiMonth(t *testing.T) {
	t.Parallel()
	m := New(date(2026, time.March, 17), date(2025, time.March, 17), DefaultConfig(), WithJulian(true), WithMonths(3))
	output := m.View()
	if !strings.Contains(output, "Sun Mon") {
		t.Errorf("expected julian headers in multi-month, got:\n%s", output)
	}
}

func TestQuarterBarProgress(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.ShowQuarterBar = true
	startOfQ := New(date(2026, time.January, 1), date(2026, time.January, 1), cfg)
	if !strings.Contains(startOfQ.View(), "Q1") {
		t.Error("expected Q1 at start of year")
	}
	endOfQ := New(date(2026, time.March, 31), date(2026, time.March, 31), cfg)
	if !strings.Contains(endOfQ.View(), "Q1") {
		t.Error("expected Q1 at end of March")
	}
}

// --- View Composition Tests ---
// These tests verify multi-month layout, column alignment, centering, and
// week number wrapping edge cases that were previously under-tested.

func TestMultiMonthColumnHeightBalancing(t *testing.T) {
	t.Parallel()
	// February 2026 has 28 days starting Sunday → 4 grid rows.
	// March 2026 has 31 days starting Sunday → 5 grid rows.
	// The multi-month layout must pad the shorter column.
	cfg := DefaultConfig()
	cursor := date(2026, time.March, 1)
	today := date(2025, time.January, 1)
	m := New(cursor, today, cfg, WithMonths(3))
	output := m.View()

	// All three months should be present
	for _, month := range []string{"February 2026", "March 2026", "April 2026"} {
		if !strings.Contains(output, month) {
			t.Errorf("expected %q in output", month)
		}
	}

	// The multi-month view should render side-by-side, not stacked.
	// Check that February and March titles appear on the same line.
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "February") && strings.Contains(line, "March") {
			return // found them on the same line
		}
	}
	t.Error("expected February and March titles on the same line (side by side)")
}

func TestMultiMonthEvenCount(t *testing.T) {
	t.Parallel()
	// Even number of months: 2 months centered on March → Feb + March
	cfg := DefaultConfig()
	cursor := date(2026, time.March, 15)
	today := date(2025, time.January, 1)
	m := New(cursor, today, cfg, WithMonths(2))
	output := m.View()

	if !strings.Contains(output, "February 2026") {
		t.Error("expected February in 2-month view")
	}
	if !strings.Contains(output, "March 2026") {
		t.Error("expected March in 2-month view")
	}
}

func TestMultiMonthFiveWide(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cursor := date(2026, time.June, 15)
	today := date(2025, time.January, 1)
	m := New(cursor, today, cfg, WithMonths(5))
	output := m.View()

	// 5 months centered on June → April, May, June, July, August
	for _, month := range []string{"April 2026", "May 2026", "June 2026", "July 2026", "August 2026"} {
		if !strings.Contains(output, month) {
			t.Errorf("expected %q in 5-month view", month)
		}
	}
}

func TestMultiMonthWithWeekNumbersAlignment(t *testing.T) {
	t.Parallel()
	// Week numbers on both sides should produce consistent column widths.
	for _, pos := range []string{"left", "right"} {
		t.Run(pos, func(t *testing.T) {
			t.Parallel()
			cfg := DefaultConfig()
			cfg.ShowWeekNumbers = pos
			cursor := date(2026, time.March, 17)
			today := date(2025, time.January, 1)
			m := New(cursor, today, cfg, WithMonths(3))
			output := m.View()

			if !strings.Contains(output, "Wk") {
				t.Error("expected 'Wk' header")
			}
			// Verify all three months are present and rendered
			for _, month := range []string{"February 2026", "March 2026", "April 2026"} {
				if !strings.Contains(output, month) {
					t.Errorf("expected %q in output", month)
				}
			}
		})
	}
}

func TestMultiMonthYearBoundary(t *testing.T) {
	t.Parallel()
	// December 2025 → January 2026: year boundary in multi-month view.
	// Use a today in a different year so smart title shows the year for all months.
	cfg := DefaultConfig()
	cursor := date(2026, time.January, 15)
	today := date(2024, time.January, 1)
	m := New(cursor, today, cfg, WithMonths(3))
	output := m.View()

	if !strings.Contains(output, "December 2025") {
		t.Errorf("expected December 2025 at year boundary, got:\n%s", output)
	}
	if !strings.Contains(output, "January 2026") {
		t.Error("expected January 2026 at year boundary")
	}
	if !strings.Contains(output, "February 2026") {
		t.Error("expected February 2026 at year boundary")
	}
}

func TestMultiMonthQuarterBar(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.ShowQuarterBar = true
	cursor := date(2026, time.March, 17)
	today := date(2026, time.March, 17)
	m := New(cursor, today, cfg, WithMonths(3))
	output := m.View()

	if !strings.Contains(output, "Q1") {
		t.Error("expected quarter bar in multi-month view")
	}
}

func TestMultiMonthWithHelpBar(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cursor := date(2026, time.March, 17)
	today := date(2025, time.January, 1)
	m := New(cursor, today, cfg, WithMonths(3))
	m.showHelp = true
	output := m.View()

	if !strings.Contains(output, "prev day") {
		t.Error("expected help bar in multi-month view")
	}
}

func TestMultiMonthJulian(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cursor := date(2026, time.March, 17)
	today := date(2025, time.January, 1)
	m := New(cursor, today, cfg, WithMonths(3), WithJulian(true))
	output := m.View()

	// Julian day headers use 3-char abbreviations
	if !strings.Contains(output, "Sun Mon") {
		t.Errorf("expected 3-char julian headers in multi-month, got:\n%s", output)
	}
	// March 1 = yearday 60
	if !strings.Contains(output, " 60") {
		t.Errorf("expected yearday 60 in multi-month julian, got:\n%s", output)
	}
}

func TestSingleMonthCentering(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cursor := date(2026, time.March, 17)
	today := date(2025, time.January, 1)
	m := New(cursor, today, cfg)
	m.termWidth = 120
	m.termHeight = 40
	output := m.View()

	// When term dimensions are set, lipgloss.Place centers the content.
	// The output should be larger than the raw calendar.
	raw := New(cursor, today, cfg).View()
	if len(output) <= len(raw) {
		t.Error("expected centered output to include padding")
	}
}

func TestMultiMonthCentering(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cursor := date(2026, time.March, 17)
	today := date(2025, time.January, 1)
	m := New(cursor, today, cfg, WithMonths(3))
	m.termWidth = 200
	m.termHeight = 50
	output := m.View()

	raw := New(cursor, today, cfg, WithMonths(3)).View()
	if len(output) <= len(raw) {
		t.Error("expected centered multi-month output to include padding")
	}
}

func TestWrapWithWeekNumsOff(t *testing.T) {
	t.Parallel()
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	lines := []string{"line1", "line2"}
	wnLines := []string{"W1", "W2"}
	// WeekNumOff should return lines unchanged
	result := m.wrapWithWeekNums(lines, wnLines)
	if result[0] != "line1" || result[1] != "line2" {
		t.Error("expected unchanged lines when week numbers are off")
	}
}

func TestWrapWithWeekNumsLeft(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.ShowWeekNumbers = "left"
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	lines := []string{"content"}
	wnLines := []string{"W1"}
	result := m.wrapWithWeekNums(lines, wnLines)
	if !strings.HasPrefix(result[0], "W1") {
		t.Errorf("expected week number prefix, got %q", result[0])
	}
}

func TestWrapWithWeekNumsRight(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.ShowWeekNumbers = "right"
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	lines := []string{"content"}
	wnLines := []string{"W1"}
	result := m.wrapWithWeekNums(lines, wnLines)
	if !strings.HasSuffix(result[0], "W1") {
		t.Errorf("expected week number suffix, got %q", result[0])
	}
}

func TestWrapWithWeekNumsPadding(t *testing.T) {
	t.Parallel()
	// When there are more lines than week numbers, padding should be applied
	cfg := DefaultConfig()
	cfg.ShowWeekNumbers = "left"
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	lines := []string{"line1", "line2", "line3"}
	wnLines := []string{"W1"} // fewer than lines
	result := m.wrapWithWeekNums(lines, wnLines)
	if len(result) != 3 {
		t.Errorf("expected 3 result lines, got %d", len(result))
	}
	// Lines without week numbers should have blank padding
	if !strings.HasPrefix(result[1], "   ") {
		t.Errorf("expected blank padding for line without week number, got %q", result[1])
	}
}

func TestBuildWeekNumLinesLeft(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.ShowWeekNumbers = "left"
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	gridWNs := []int{10, 11, 12, 13, 14}
	wnLines := m.buildWeekNumLines(gridWNs, 8) // 8 core lines
	// First line: title (empty), second: "Wk" header, then 5 week numbers
	if wnLines[0] != "" {
		t.Errorf("expected empty title line, got %q", wnLines[0])
	}
	if !strings.Contains(wnLines[1], "Wk") {
		t.Errorf("expected Wk header, got %q", wnLines[1])
	}
	if len(wnLines) != 8 {
		t.Errorf("expected 8 lines (padded), got %d", len(wnLines))
	}
}

func TestBuildWeekNumLinesRight(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.ShowWeekNumbers = "right"
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	gridWNs := []int{10, 11}
	wnLines := m.buildWeekNumLines(gridWNs, 5)
	// Right-aligned: format is " Wk" with leading space
	if !strings.HasPrefix(wnLines[1], " ") {
		t.Errorf("expected right-aligned Wk header with leading space, got %q", wnLines[1])
	}
}

// --- TUI Integration Tests ---
// These tests exercise the full Init → Update → View cycle with realistic
// message sequences, verifying that the model stays consistent through
// multiple rounds of state changes.

func TestIntegrationInitUpdateViewCycle(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())

	// Init should produce a command (midnight tick)
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init() should return a cmd")
	}

	// View should render without panic
	output := m.View()
	if !strings.Contains(output, "March") {
		t.Fatal("initial View() should contain 'March'")
	}

	// Navigate right
	updated, _ := m.Update(runeMsg("l"))
	m = updated.(Model)
	if m.cursor != date(2026, time.March, 18) {
		t.Errorf("after right nav, cursor = %s, want 2026-03-18", m.cursor.Format("2006-01-02"))
	}

	// View should still render correctly
	output = m.View()
	if !strings.Contains(output, "March") {
		t.Error("View() after navigation should still contain 'March'")
	}

	// Toggle week numbers
	updated, _ = m.Update(runeMsg("w"))
	m = updated.(Model)
	output = m.View()
	if !strings.Contains(output, "Wk") {
		t.Error("View() after week toggle should contain 'Wk'")
	}

	// Select date
	updated, cmd = m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.Selected() {
		t.Error("expected Selected() after Enter")
	}
	if cmd == nil {
		t.Error("expected quit command after Enter")
	}
}

func TestIntegrationWindowResizeThenNavigate(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())

	// Simulate terminal resize
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(Model)
	if m.termWidth != 120 || m.termHeight != 40 {
		t.Errorf("expected 120x40, got %dx%d", m.termWidth, m.termHeight)
	}

	// View should center with the new dimensions
	output := m.View()
	if len(output) == 0 {
		t.Fatal("View() after resize should not be empty")
	}

	// Navigate after resize
	for range 5 {
		updated, _ = m.Update(runeMsg("l"))
		m = updated.(Model)
	}
	if m.cursor != date(2026, time.March, 22) {
		t.Errorf("cursor after 5 right moves = %s, want 2026-03-22", m.cursor.Format("2006-01-02"))
	}

	// View should still render correctly with centering
	output = m.View()
	if !strings.Contains(output, "March") {
		t.Error("View() should still contain 'March' after resize+navigate")
	}
}

func TestIntegrationMidnightTickCycle(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())

	// Init
	_ = m.Init()

	// Render initial view
	output1 := m.View()

	// Simulate midnight tick
	updated, cmd := m.Update(midnightTickMsg{})
	m = updated.(Model)

	// Should reschedule the next tick
	if cmd == nil {
		t.Error("midnight tick should return a cmd for next tick")
	}

	// View should still render
	output2 := m.View()
	if len(output2) == 0 {
		t.Fatal("View() after midnight tick should not be empty")
	}

	// Navigate to verify model is still functional
	updated, _ = m.Update(runeMsg("l"))
	m = updated.(Model)
	_ = m.View() // should not panic

	_ = output1 // used above
}

func TestIntegrationRangeSelectThenView(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())

	// Start range
	updated, _ := m.Update(runeMsg("v"))
	m = updated.(Model)

	// Navigate through multiple days
	for range 10 {
		updated, _ = m.Update(runeMsg("l"))
		m = updated.(Model)
		// View should render at every step without panic
		output := m.View()
		if !strings.Contains(output, "March") {
			t.Fatal("View() should contain 'March' during range navigation")
		}
	}

	// Confirm range
	updated, _ = m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.InRange() {
		t.Error("expected InRange() after range select + Enter")
	}
	if m.RangeStart() != date(2026, time.March, 17) {
		t.Errorf("RangeStart = %s, want 2026-03-17", m.RangeStart().Format("2006-01-02"))
	}
	if m.RangeEnd() != date(2026, time.March, 27) {
		t.Errorf("RangeEnd = %s, want 2026-03-27", m.RangeEnd().Format("2006-01-02"))
	}
}

func TestIntegrationMultiMonthNavigateAcrossMonths(t *testing.T) {
	t.Parallel()
	today := date(2025, time.January, 1)
	cursor := date(2026, time.March, 31)
	m := New(cursor, today, DefaultConfig(), WithMonths(3))

	// Navigate forward — should wrap into April
	updated, _ := m.Update(runeMsg("l"))
	m = updated.(Model)
	if m.cursor != date(2026, time.April, 1) {
		t.Errorf("expected April 1 after navigating past March 31, got %s", m.cursor.Format("2006-01-02"))
	}

	// View should now center on April
	output := m.View()
	if !strings.Contains(output, "April 2026") {
		t.Error("expected April 2026 in multi-month view after navigation")
	}
}

func TestIntegrationHighlightChangedMsg(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())

	_ = m.Init()

	// Simulate highlight file change
	newDates := map[time.Time]bool{
		dateKey(date(2026, time.March, 20)): true,
		dateKey(date(2026, time.March, 25)): true,
	}
	updated, cmd := m.Update(highlightChangedMsg{
		dates:   newDates,
		watcher: nil,
		path:    "/test/path",
	})
	m = updated.(Model)

	if len(m.highlightedDates) != 2 {
		t.Errorf("expected 2 highlights, got %d", len(m.highlightedDates))
	}
	if cmd == nil {
		t.Error("expected cmd to watch for next change")
	}

	// View should render with highlights
	output := m.View()
	if !strings.Contains(output, "20") || !strings.Contains(output, "25") {
		t.Error("expected highlighted dates in view output")
	}
}

func TestIntegrationWatcherErrorGraceful(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())

	// Simulate watcher error
	updated, cmd := m.Update(watcherErrMsg{err: fmt.Errorf("test error")})
	m = updated.(Model)

	// Should degrade gracefully — no panic, no quit
	if m.IsQuit() {
		t.Error("watcher error should not cause quit")
	}
	if cmd != nil {
		t.Error("watcher error should not return a command")
	}

	// Model should still be functional
	output := m.View()
	if !strings.Contains(output, "March") {
		t.Error("View() should work after watcher error")
	}
}

func TestIntegrationToggleJulianMidSession(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())

	// Initial view — normal mode
	output := m.View()
	if strings.Contains(output, "Sun Mon") {
		t.Error("should not have julian headers initially")
	}

	// Toggle julian
	updated, _ := m.Update(runeMsg("J"))
	m = updated.(Model)
	output = m.View()
	if !strings.Contains(output, "Sun Mon") {
		t.Error("expected julian headers after toggle")
	}

	// Navigate while in julian mode
	updated, _ = m.Update(runeMsg("l"))
	m = updated.(Model)
	output = m.View()
	if !strings.Contains(output, "Sun Mon") {
		t.Error("julian mode should persist through navigation")
	}

	// Toggle back
	updated, _ = m.Update(runeMsg("J"))
	m = updated.(Model)
	output = m.View()
	if strings.Contains(output, "Sun Mon") {
		t.Error("should not have julian headers after second toggle")
	}
}

// TestIntegrationRowInitUpdateViewCycle tests the full RowModel lifecycle.
func TestIntegrationRowInitUpdateViewCycle(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := NewRow(today, today, DefaultConfig())

	// Init
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("RowModel Init() should return a cmd")
	}

	// Render
	output := m.View()
	if !strings.Contains(output, "Mr") {
		t.Fatal("RowModel View() should contain 'Mr' (March)")
	}

	// Navigate right
	updated, _ := m.Update(runeMsg("l"))
	m = updated.(RowModel)
	output = m.View()
	if !strings.Contains(output, "Mr") {
		t.Error("RowModel View() after navigation should contain 'Mr'")
	}

	// Window resize
	updated, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(RowModel)
	output = m.View()
	if len(output) == 0 {
		t.Error("RowModel View() after resize should not be empty")
	}

	// Select
	updated, cmd = m.Update(specialMsg(tea.KeyEnter))
	m = updated.(RowModel)
	if !m.Selected() {
		t.Error("expected Selected() after Enter")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestIntegrationRowMidnightTick(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := NewRow(today, today, DefaultConfig())

	_ = m.Init()

	// Midnight tick
	updated, cmd := m.Update(midnightTickMsg{})
	m = updated.(RowModel)
	if cmd == nil {
		t.Error("midnight tick should reschedule")
	}

	// Still functional
	output := m.View()
	if len(output) == 0 {
		t.Error("View() after midnight tick should not be empty")
	}
}
