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

func TestNextDay(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17))
	m = pressKey(m, "l")
	if m.Cursor != date(2026, time.March, 18) {
		t.Errorf("expected March 18, got %s", m.Cursor.Format("2006-01-02"))
	}
}

func TestPrevDay(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17))
	m = pressKey(m, "h")
	if m.Cursor != date(2026, time.March, 16) {
		t.Errorf("expected March 16, got %s", m.Cursor.Format("2006-01-02"))
	}
}

func TestNextWeek(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17))
	m = pressKey(m, "j")
	if m.Cursor != date(2026, time.March, 24) {
		t.Errorf("expected March 24, got %s", m.Cursor.Format("2006-01-02"))
	}
}

func TestPrevWeek(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17))
	m = pressKey(m, "k")
	if m.Cursor != date(2026, time.March, 10) {
		t.Errorf("expected March 10, got %s", m.Cursor.Format("2006-01-02"))
	}
}

func TestNextMonth(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17))
	m = pressKey(m, "L")
	if m.Cursor != date(2026, time.April, 17) {
		t.Errorf("expected April 17, got %s", m.Cursor.Format("2006-01-02"))
	}
}

func TestPrevMonth(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17))
	m = pressKey(m, "H")
	if m.Cursor != date(2026, time.February, 17) {
		t.Errorf("expected Feb 17, got %s", m.Cursor.Format("2006-01-02"))
	}
}

func TestNextMonthClampsDay(t *testing.T) {
	m := New(date(2026, time.January, 31), date(2026, time.January, 31))
	m = pressKey(m, "L")
	if m.Cursor != date(2026, time.February, 28) {
		t.Errorf("expected Feb 28, got %s", m.Cursor.Format("2006-01-02"))
	}
}

func TestPrevMonthClampsDay(t *testing.T) {
	m := New(date(2026, time.March, 31), date(2026, time.March, 31))
	m = pressKey(m, "H")
	if m.Cursor != date(2026, time.February, 28) {
		t.Errorf("expected Feb 28, got %s", m.Cursor.Format("2006-01-02"))
	}
}

func TestDayWrapForward(t *testing.T) {
	m := New(date(2026, time.March, 31), date(2026, time.March, 31))
	m = pressKey(m, "l")
	if m.Cursor != date(2026, time.April, 1) {
		t.Errorf("expected April 1, got %s", m.Cursor.Format("2006-01-02"))
	}
}

func TestDayWrapBackward(t *testing.T) {
	m := New(date(2026, time.March, 1), date(2026, time.March, 1))
	m = pressKey(m, "h")
	if m.Cursor != date(2026, time.February, 28) {
		t.Errorf("expected Feb 28, got %s", m.Cursor.Format("2006-01-02"))
	}
}

func TestWeekWrapForward(t *testing.T) {
	m := New(date(2026, time.March, 28), date(2026, time.March, 28))
	m = pressKey(m, "j")
	if m.Cursor != date(2026, time.April, 4) {
		t.Errorf("expected April 4, got %s", m.Cursor.Format("2006-01-02"))
	}
}

func TestEnterSelectsDate(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17))
	m = pressKey(m, "l")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if !m.Selected {
		t.Error("expected Selected to be true")
	}
	if m.Cursor != date(2026, time.March, 18) {
		t.Errorf("expected March 18, got %s", m.Cursor.Format("2006-01-02"))
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestQuitDoesNotSelect(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17))
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = updated.(Model)
	if m.Selected {
		t.Error("expected Selected to be false")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestJumpToToday(t *testing.T) {
	m := New(date(2026, time.June, 15), date(2026, time.March, 17))
	m = pressKey(m, "t")
	if m.Cursor != date(2026, time.March, 17) {
		t.Errorf("expected March 17, got %s", m.Cursor.Format("2006-01-02"))
	}
}

func TestEscDoesNotSelect(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17))
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updated.(Model)
	if m.Selected {
		t.Error("expected Selected to be false")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}
