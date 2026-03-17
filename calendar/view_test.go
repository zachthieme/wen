package calendar

import (
	"strings"
	"testing"
	"time"
)

func TestRenderMarch2026(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17))
	output := Render(m)

	if !strings.Contains(output, "March 2026") {
		t.Error("expected 'March 2026' in output")
	}

	if !strings.Contains(output, "Su Mo Tu We Th Fr Sa") {
		t.Error("expected day headers in output")
	}

	lines := strings.Split(output, "\n")
	found := false
	for _, line := range lines {
		if strings.Contains(line, " 1 ") || strings.HasPrefix(strings.TrimSpace(line), "1 ") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected day 1 in output")
	}
}

func TestRenderFebruary2026(t *testing.T) {
	m := New(date(2026, time.February, 14), date(2026, time.March, 17))
	output := Render(m)

	if !strings.Contains(output, "February 2026") {
		t.Error("expected 'February 2026' in output")
	}
}
