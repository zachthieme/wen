package calendar

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

func runeMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

func specialMsg(key tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: key}
}

func pressKey(m Model, key string) Model {
	updated, _ := m.Update(runeMsg(key))
	return updated.(Model)
}

func TestNavigation(t *testing.T) {
	today := date(2026, time.March, 17)

	tests := []struct {
		name     string
		msg      tea.KeyMsg
		start    time.Time
		expected time.Time
	}{
		// Day navigation
		{"next day (l)", runeMsg("l"), today, date(2026, time.March, 18)},
		{"prev day (h)", runeMsg("h"), today, date(2026, time.March, 16)},
		{"next day (→)", specialMsg(tea.KeyRight), today, date(2026, time.March, 18)},
		{"prev day (←)", specialMsg(tea.KeyLeft), today, date(2026, time.March, 16)},

		// Week navigation
		{"next week (j)", runeMsg("j"), today, date(2026, time.March, 24)},
		{"prev week (k)", runeMsg("k"), today, date(2026, time.March, 10)},
		{"next week (↓)", specialMsg(tea.KeyDown), today, date(2026, time.March, 24)},
		{"prev week (↑)", specialMsg(tea.KeyUp), today, date(2026, time.March, 10)},

		// Month navigation
		{"next month (L)", runeMsg("L"), today, date(2026, time.April, 17)},
		{"prev month (H)", runeMsg("H"), today, date(2026, time.February, 17)},

		// Month clamping
		{"next month clamps day", runeMsg("L"), date(2026, time.January, 31), date(2026, time.February, 28)},
		{"prev month clamps day", runeMsg("H"), date(2026, time.March, 31), date(2026, time.February, 28)},

		// Year navigation
		{"next year (J)", runeMsg("J"), today, date(2027, time.March, 17)},
		{"prev year (K)", runeMsg("K"), today, date(2025, time.March, 17)},
		{"next year leap day clamp", runeMsg("J"), date(2024, time.February, 29), date(2025, time.February, 28)},

		// Boundary wrapping
		{"day wrap forward", runeMsg("l"), date(2026, time.March, 31), date(2026, time.April, 1)},
		{"day wrap backward", runeMsg("h"), date(2026, time.March, 1), date(2026, time.February, 28)},
		{"week wrap forward", runeMsg("j"), date(2026, time.March, 28), date(2026, time.April, 4)},

		// Jump to today
		{"jump to today (t)", runeMsg("t"), date(2026, time.June, 15), today},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.start, today, DefaultConfig())
			updated, _ := m.Update(tt.msg)
			got := updated.(Model)
			if got.cursor != tt.expected {
				t.Errorf("got %s, want %s", got.cursor.Format(DateLayout), tt.expected.Format(DateLayout))
			}
		})
	}
}

func TestToggleWeekNumbers(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	if m.showWeekNumbers {
		t.Error("expected showWeekNumbers false initially")
	}
	m = pressKey(m, "w")
	if !m.showWeekNumbers {
		t.Error("expected showWeekNumbers true after toggle")
	}
	m = pressKey(m, "w")
	if m.showWeekNumbers {
		t.Error("expected showWeekNumbers false after second toggle")
	}
}

func TestToggleHelp(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	if m.showHelp {
		t.Error("expected showHelp false initially")
	}
	m = pressKey(m, "?")
	if !m.showHelp {
		t.Error("expected showHelp true after toggle")
	}
}

func TestQuitExits(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	updated, cmd := m.Update(runeMsg("q"))
	m = updated.(Model)
	if !m.IsQuit() {
		t.Error("expected IsQuit to be true")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestEscQuits(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	updated, cmd := m.Update(specialMsg(tea.KeyEscape))
	m = updated.(Model)
	if !m.IsQuit() {
		t.Error("expected IsQuit to be true")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}
