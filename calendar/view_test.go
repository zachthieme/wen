package calendar

import (
	"strings"
	"testing"
	"time"

	"github.com/zachthieme/wen"
)

func TestRenderMarch2026(t *testing.T) {
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
	// Use a different year for today so the year appears in the title
	m := New(date(2026, time.February, 14), date(2025, time.March, 17), DefaultConfig())
	output := m.View()

	if !strings.Contains(output, "February 2026") {
		t.Error("expected 'February 2026' in output")
	}
}

func TestRenderWithWeekNumbers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ShowWeekNumbers = "left"
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	output := m.View()

	if !strings.Contains(output, "Wk") {
		t.Error("expected 'Wk' header when week numbers enabled")
	}
}

func TestRenderWithoutWeekNumbers(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	output := m.View()

	if strings.Contains(output, "Wk") {
		t.Error("should not have 'Wk' header when week numbers disabled")
	}
}

func TestRenderMondayStart(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WeekStartDay = 1
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	output := m.View()

	if !strings.Contains(output, "Mo Tu We Th Fr Sa Su") {
		t.Error("expected Monday-start day headers")
	}
}

func TestRenderHelpBar(t *testing.T) {
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
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	output := m.View()

	if strings.Contains(output, "prev day") {
		t.Error("help bar should not appear by default")
	}
}

func TestWeekNumberUS(t *testing.T) {
	d := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.Local)
	wn := weekNumber(d, "us")
	// March 1 is day 60 of 2026. Jan 1 is Thursday (weekday 4).
	// (60 + 4 - 1) / 7 + 1 = 63/7 + 1 = 9 + 1 = 10
	if wn != 10 {
		t.Errorf("expected US week 10 for March 1 2026, got %d", wn)
	}
}

func TestWeekNumberISO(t *testing.T) {
	d := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.Local)
	wn := weekNumber(d, "iso")
	_, expected := d.ISOWeek()
	if wn != expected {
		t.Errorf("expected ISO week %d for March 1 2026, got %d", expected, wn)
	}
}

func TestRenderWithLeftPadding(t *testing.T) {
	cfg := DefaultConfig()
	noPad := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	noPadOutput := noPad.View()

	cfg.PaddingLeft = 3
	withPad := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	padOutput := withPad.View()

	// Padded output should be different from unpadded.
	if padOutput == noPadOutput {
		t.Error("expected padded output to differ from unpadded output")
	}
	// Each line with content should be at least 3 chars wider due to left padding.
	noPadLines := strings.Split(noPadOutput, "\n")
	padLines := strings.Split(padOutput, "\n")
	for i, line := range padLines {
		if strings.Contains(line, "March") {
			if i < len(noPadLines) && len(line) <= len(noPadLines[i]) {
				t.Errorf("expected padded title line to be wider, got len %d vs %d", len(line), len(noPadLines[i]))
			}
			break
		}
	}
}

func TestRenderWithTopPadding(t *testing.T) {
	cfg := DefaultConfig()
	noPad := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	noPadOutput := noPad.View()

	cfg.PaddingTop = 2
	withPad := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	padOutput := withPad.View()

	// Lip Gloss adds top padding as extra lines (with spaces, not bare newlines).
	noPadLines := len(strings.Split(noPadOutput, "\n"))
	padLines := len(strings.Split(padOutput, "\n"))
	if padLines < noPadLines+2 {
		t.Errorf("expected at least 2 more lines with top padding, got %d vs %d", padLines, noPadLines)
	}
}

func TestFiscalQuarter(t *testing.T) {
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
			q, fy := wen.FiscalQuarter(tt.month, tt.year, tt.fyStart)
			if q != tt.wantQ || fy != tt.wantFY {
				t.Errorf("FiscalQuarter(%d, %d, %d) = Q%d FY%d, want Q%d FY%d",
					tt.month, tt.year, tt.fyStart, q, fy, tt.wantQ, tt.wantFY)
			}
		})
	}
}

func TestRenderFiscalQuarterTitle(t *testing.T) {
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
	cfg := DefaultConfig()
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	output := m.View()
	if strings.Contains(output, "FY") {
		t.Error("expected no fiscal quarter in default config")
	}
}

func TestRenderMultiMonth(t *testing.T) {
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
	cursor := date(2027, time.March, 17)
	today := date(2026, time.March, 17)
	m := New(cursor, today, DefaultConfig())
	output := m.View()
	if !strings.Contains(output, "March 2027") {
		t.Error("expected 'March 2027' for non-current year")
	}
}

func TestSmartTitleWithFiscalQuarter(t *testing.T) {
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
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	output := m.View()
	if strings.Contains(output, "█") || strings.Contains(output, "░") {
		t.Error("quarter bar should not appear with default config")
	}
}

func TestQuarterBarRendering(t *testing.T) {
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

func TestQuarterBarProgress(t *testing.T) {
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
