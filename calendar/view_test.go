package calendar

import (
	"strings"
	"testing"
	"time"
)

func TestRenderMarch2026(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	output := m.View()

	if !strings.Contains(output, "March 2026") {
		t.Error("expected 'March 2026' in output")
	}
	if !strings.Contains(output, "Su Mo Tu We Th Fr Sa") {
		t.Error("expected Sunday-start day headers")
	}
}

func TestRenderFebruary2026(t *testing.T) {
	m := New(date(2026, time.February, 14), date(2026, time.March, 17), DefaultConfig())
	output := m.View()

	if !strings.Contains(output, "February 2026") {
		t.Error("expected 'February 2026' in output")
	}
}

func TestRenderWithWeekNumbers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ShowWeekNumbers = true
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
		if strings.Contains(line, "March 2026") {
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
