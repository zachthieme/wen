package calendar

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

func pressKey(m Model, key string) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated.(Model)
}

func pressArrow(m Model, key tea.KeyType) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: key})
	return updated.(Model)
}

func TestNextDay(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressKey(m, "l")
	if m.cursor != date(2026, time.March, 18) {
		t.Errorf("expected March 18, got %s", m.cursor.Format(DateLayout))
	}
}

func TestPrevDay(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressKey(m, "h")
	if m.cursor != date(2026, time.March, 16) {
		t.Errorf("expected March 16, got %s", m.cursor.Format(DateLayout))
	}
}

func TestNextDayArrow(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressArrow(m, tea.KeyRight)
	if m.cursor != date(2026, time.March, 18) {
		t.Errorf("expected March 18, got %s", m.cursor.Format(DateLayout))
	}
}

func TestPrevDayArrow(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressArrow(m, tea.KeyLeft)
	if m.cursor != date(2026, time.March, 16) {
		t.Errorf("expected March 16, got %s", m.cursor.Format(DateLayout))
	}
}

func TestNextWeek(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressKey(m, "j")
	if m.cursor != date(2026, time.March, 24) {
		t.Errorf("expected March 24, got %s", m.cursor.Format(DateLayout))
	}
}

func TestPrevWeek(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressKey(m, "k")
	if m.cursor != date(2026, time.March, 10) {
		t.Errorf("expected March 10, got %s", m.cursor.Format(DateLayout))
	}
}

func TestNextWeekArrow(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressArrow(m, tea.KeyDown)
	if m.cursor != date(2026, time.March, 24) {
		t.Errorf("expected March 24, got %s", m.cursor.Format(DateLayout))
	}
}

func TestPrevWeekArrow(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressArrow(m, tea.KeyUp)
	if m.cursor != date(2026, time.March, 10) {
		t.Errorf("expected March 10, got %s", m.cursor.Format(DateLayout))
	}
}

func TestNextMonth(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressKey(m, "L")
	if m.cursor != date(2026, time.April, 17) {
		t.Errorf("expected April 17, got %s", m.cursor.Format(DateLayout))
	}
}

func TestPrevMonth(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressKey(m, "H")
	if m.cursor != date(2026, time.February, 17) {
		t.Errorf("expected Feb 17, got %s", m.cursor.Format(DateLayout))
	}
}

func TestNextMonthClampsDay(t *testing.T) {
	m := New(date(2026, time.January, 31), date(2026, time.January, 31), DefaultConfig())
	m = pressKey(m, "L")
	if m.cursor != date(2026, time.February, 28) {
		t.Errorf("expected Feb 28, got %s", m.cursor.Format(DateLayout))
	}
}

func TestPrevMonthClampsDay(t *testing.T) {
	m := New(date(2026, time.March, 31), date(2026, time.March, 31), DefaultConfig())
	m = pressKey(m, "H")
	if m.cursor != date(2026, time.February, 28) {
		t.Errorf("expected Feb 28, got %s", m.cursor.Format(DateLayout))
	}
}

func TestNextYear(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressKey(m, "J")
	if m.cursor != date(2027, time.March, 17) {
		t.Errorf("expected 2027-03-17, got %s", m.cursor.Format(DateLayout))
	}
}

func TestPrevYear(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressKey(m, "K")
	if m.cursor != date(2025, time.March, 17) {
		t.Errorf("expected 2025-03-17, got %s", m.cursor.Format(DateLayout))
	}
}

func TestNextYearLeapDayClamp(t *testing.T) {
	m := New(date(2024, time.February, 29), date(2024, time.February, 29), DefaultConfig())
	m = pressKey(m, "J")
	if m.cursor != date(2025, time.February, 28) {
		t.Errorf("expected 2025-02-28, got %s", m.cursor.Format(DateLayout))
	}
}

func TestDayWrapForward(t *testing.T) {
	m := New(date(2026, time.March, 31), date(2026, time.March, 31), DefaultConfig())
	m = pressKey(m, "l")
	if m.cursor != date(2026, time.April, 1) {
		t.Errorf("expected April 1, got %s", m.cursor.Format(DateLayout))
	}
}

func TestDayWrapBackward(t *testing.T) {
	m := New(date(2026, time.March, 1), date(2026, time.March, 1), DefaultConfig())
	m = pressKey(m, "h")
	if m.cursor != date(2026, time.February, 28) {
		t.Errorf("expected Feb 28, got %s", m.cursor.Format(DateLayout))
	}
}

func TestWeekWrapForward(t *testing.T) {
	m := New(date(2026, time.March, 28), date(2026, time.March, 28), DefaultConfig())
	m = pressKey(m, "j")
	if m.cursor != date(2026, time.April, 4) {
		t.Errorf("expected April 4, got %s", m.cursor.Format(DateLayout))
	}
}

func TestToggleWeekNumbers(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	if m.showWeekNumbers {
		t.Error("expected ShowWeekNumbers false initially")
	}
	m = pressKey(m, "w")
	if !m.showWeekNumbers {
		t.Error("expected ShowWeekNumbers true after toggle")
	}
	m = pressKey(m, "w")
	if m.showWeekNumbers {
		t.Error("expected ShowWeekNumbers false after second toggle")
	}
}

func TestToggleHelp(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	if m.showHelp {
		t.Error("expected ShowHelp false initially")
	}
	m = pressKey(m, "?")
	if !m.showHelp {
		t.Error("expected ShowHelp true after toggle")
	}
}

func TestJumpToToday(t *testing.T) {
	m := New(date(2026, time.June, 15), date(2026, time.March, 17), DefaultConfig())
	m = pressKey(m, "t")
	if m.cursor != date(2026, time.March, 17) {
		t.Errorf("expected March 17, got %s", m.cursor.Format(DateLayout))
	}
}

func TestEnterSelectsDate(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressKey(m, "l")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if !m.IsSelected() {
		t.Error("expected Selected to be true")
	}
	if m.cursor != date(2026, time.March, 18) {
		t.Errorf("expected March 18, got %s", m.cursor.Format(DateLayout))
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestQuitDoesNotSelect(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = updated.(Model)
	if m.IsSelected() {
		t.Error("expected Selected to be false")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestEscDoesNotSelect(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updated.(Model)
	if m.IsSelected() {
		t.Error("expected Selected to be false")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestYankDoesNotPanic(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressKey(m, "y")
	if m.cursor != date(2026, time.March, 17) {
		t.Error("yank should not change cursor")
	}
}
