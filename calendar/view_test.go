package calendar

import (
	"strings"
	"testing"
	"time"
)

func TestRenderMarch2026(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	output := Render(m)

	if !strings.Contains(output, "March 2026") {
		t.Error("expected 'March 2026' in output")
	}
	if !strings.Contains(output, "Su Mo Tu We Th Fr Sa") {
		t.Error("expected Sunday-start day headers")
	}
}

func TestRenderFebruary2026(t *testing.T) {
	m := New(date(2026, time.February, 14), date(2026, time.March, 17), DefaultConfig())
	output := Render(m)

	if !strings.Contains(output, "February 2026") {
		t.Error("expected 'February 2026' in output")
	}
}

func TestRenderWithWeekNumbers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ShowWeekNumbers = true
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	output := Render(m)

	if !strings.Contains(output, "Wk") {
		t.Error("expected 'Wk' header when week numbers enabled")
	}
}

func TestRenderWithoutWeekNumbers(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	output := Render(m)

	if strings.Contains(output, "Wk") {
		t.Error("should not have 'Wk' header when week numbers disabled")
	}
}

func TestRenderMondayStart(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WeekStartDay = 1
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg)
	output := Render(m)

	if !strings.Contains(output, "Mo Tu We Th Fr Sa Su") {
		t.Error("expected Monday-start day headers")
	}
}

func TestRenderHelpBar(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m.ShowHelp = true
	output := Render(m)

	if !strings.Contains(output, "h/l:day") {
		t.Error("expected help bar content")
	}
	if !strings.Contains(output, "J/K:year") {
		t.Error("expected year jump in help bar")
	}
}

func TestRenderNoHelpBarByDefault(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	output := Render(m)

	if strings.Contains(output, "h/l:day") {
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
