package calendar

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/zachthieme/wen"

	tea "github.com/charmbracelet/bubbletea"
)

var propToday = time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)
var propCursor = propToday

func TestPropertyNavigateRightLeftReturnsToStart(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	m := New(propCursor, propToday, cfg)

	for n := 1; n <= 30; n++ {
		t.Run(fmt.Sprintf("%d_steps", n), func(t *testing.T) {
			current := m
			for i := 0; i < n; i++ {
				updated, _ := current.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
				current = updated.(Model)
			}
			for i := 0; i < n; i++ {
				updated, _ := current.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
				current = updated.(Model)
			}
			if !current.Cursor().Equal(m.Cursor()) {
				t.Errorf("after %d right + %d left: got %v, want %v", n, n, current.Cursor(), m.Cursor())
			}
		})
	}
}

func TestPropertyNavigateUpDownReturnsToStart(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	m := New(propCursor, propToday, cfg)

	for n := 1; n <= 10; n++ {
		t.Run(fmt.Sprintf("%d_weeks", n), func(t *testing.T) {
			current := m
			for i := 0; i < n; i++ {
				updated, _ := current.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
				current = updated.(Model)
			}
			for i := 0; i < n; i++ {
				updated, _ := current.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
				current = updated.(Model)
			}
			if !current.Cursor().Equal(m.Cursor()) {
				t.Errorf("after %d down + %d up: got %v, want %v", n, n, current.Cursor(), m.Cursor())
			}
		})
	}
}

func TestPropertyEveryDayAppearsInGrid(t *testing.T) {
	t.Parallel()
	months := []time.Month{
		time.January, time.February, time.March, time.April,
		time.May, time.June, time.July, time.August,
		time.September, time.October, time.November, time.December,
	}
	for _, mo := range months {
		t.Run(mo.String(), func(t *testing.T) {
			cursor := time.Date(2026, mo, 15, 0, 0, 0, 0, time.UTC)
			cfg := DefaultConfig()
			m := New(cursor, propToday, cfg, WithPrintMode(true))
			output := m.View()
			daysInMonth := wen.DaysIn(2026, mo, time.UTC)
			for d := 1; d <= daysInMonth; d++ {
				dayStr := fmt.Sprintf("%2d", d)
				if !strings.Contains(output, dayStr) {
					t.Errorf("month %s missing day %d in output", mo, d)
				}
			}
		})
	}
}
