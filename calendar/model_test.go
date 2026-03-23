package calendar

import (
	"os"
	"path/filepath"
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
	if m.weekNumPos != WeekNumOff {
		t.Errorf("expected weekNumPos off initially, got %d", m.weekNumPos)
	}
	m = pressKey(m, "w")
	if m.weekNumPos != WeekNumLeft {
		t.Errorf("expected weekNumPos left after first toggle, got %d", m.weekNumPos)
	}
	m = pressKey(m, "w")
	if m.weekNumPos != WeekNumRight {
		t.Errorf("expected weekNumPos right after second toggle, got %d", m.weekNumPos)
	}
	m = pressKey(m, "w")
	if m.weekNumPos != WeekNumOff {
		t.Errorf("expected weekNumPos off after third toggle, got %d", m.weekNumPos)
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

func TestEnterSelects(t *testing.T) {
	cursor := date(2026, time.April, 15)
	m := New(cursor, date(2026, time.March, 17), DefaultConfig())
	updated, cmd := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.Selected() {
		t.Error("expected Selected to be true")
	}
	if m.IsQuit() {
		t.Error("expected IsQuit to be false")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
	if m.Cursor() != cursor {
		t.Errorf("got cursor %v, want %v", m.Cursor(), cursor)
	}
}

func TestVisualSelectEnter(t *testing.T) {
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	// Press v to anchor
	m = pressKey(m, "v")
	// Move 5 days right
	for i := 0; i < 5; i++ {
		m = pressKey(m, "l")
	}
	// Press Enter to confirm
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.InRange() {
		t.Error("expected InRange to be true")
	}
	if m.RangeStart() != date(2026, time.March, 17) {
		t.Errorf("RangeStart got %s, want 2026-03-17", m.RangeStart().Format(DateLayout))
	}
	if m.RangeEnd() != date(2026, time.March, 22) {
		t.Errorf("RangeEnd got %s, want 2026-03-22", m.RangeEnd().Format(DateLayout))
	}
}

func TestVisualSelectCancel(t *testing.T) {
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	// Press v to anchor
	m = pressKey(m, "v")
	// Move right
	m = pressKey(m, "l")
	// Press Esc to cancel range (not quit)
	updated, cmd := m.Update(specialMsg(tea.KeyEscape))
	m = updated.(Model)
	if m.IsQuit() {
		t.Error("expected IsQuit to be false after first Esc (cancel range)")
	}
	if m.InRange() {
		t.Error("expected InRange to be false after Esc cancel")
	}
	if cmd != nil {
		t.Error("expected no quit command after first Esc")
	}
	// Press Esc again to quit
	updated, cmd = m.Update(specialMsg(tea.KeyEscape))
	m = updated.(Model)
	if !m.IsQuit() {
		t.Error("expected IsQuit to be true after second Esc")
	}
	if cmd == nil {
		t.Error("expected quit command after second Esc")
	}
}

func TestVisualSelectReanchor(t *testing.T) {
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	// Press v to anchor
	m = pressKey(m, "v")
	// Move 2 right
	m = pressKey(m, "l")
	m = pressKey(m, "l")
	// Press v again to re-anchor at March 19
	m = pressKey(m, "v")
	// Move 1 right
	m = pressKey(m, "l")
	// Press Enter to confirm
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.InRange() {
		t.Error("expected InRange to be true")
	}
	if m.RangeStart() != date(2026, time.March, 19) {
		t.Errorf("RangeStart got %s, want 2026-03-19", m.RangeStart().Format(DateLayout))
	}
	if m.RangeEnd() != date(2026, time.March, 20) {
		t.Errorf("RangeEnd got %s, want 2026-03-20", m.RangeEnd().Format(DateLayout))
	}
}

func TestRangeReverseOrder(t *testing.T) {
	start := date(2026, time.March, 20)
	m := New(start, start, DefaultConfig())
	// Press v to anchor at March 20
	m = pressKey(m, "v")
	// Move 5 left
	for i := 0; i < 5; i++ {
		m = pressKey(m, "h")
	}
	// Press Enter to confirm
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.InRange() {
		t.Error("expected InRange to be true")
	}
	if m.RangeStart() != date(2026, time.March, 15) {
		t.Errorf("RangeStart got %s, want 2026-03-15", m.RangeStart().Format(DateLayout))
	}
	if m.RangeEnd() != date(2026, time.March, 20) {
		t.Errorf("RangeEnd got %s, want 2026-03-20", m.RangeEnd().Format(DateLayout))
	}
}

func TestEnterWithoutRange(t *testing.T) {
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	// Press Enter without v
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.Selected() {
		t.Error("expected Selected to be true")
	}
	if m.InRange() {
		t.Error("expected InRange to be false")
	}
}

func TestSameDayRange(t *testing.T) {
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	// Press v to anchor
	m = pressKey(m, "v")
	// Immediately press Enter (same day)
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.Selected() {
		t.Error("expected Selected to be true")
	}
	if m.InRange() {
		t.Error("expected InRange to be false for same day")
	}
}

func TestCtrlCInRangeMode(t *testing.T) {
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	// Press v to anchor
	m = pressKey(m, "v")
	// Move right
	m = pressKey(m, "l")
	// Press ctrl+c to force quit
	updated, cmd := m.Update(specialMsg(tea.KeyCtrlC))
	m = updated.(Model)
	if !m.IsQuit() {
		t.Error("expected IsQuit to be true after ctrl+c")
	}
	if cmd == nil {
		t.Error("expected quit command after ctrl+c")
	}
}

func TestMidnightTickUpdatesToday(t *testing.T) {
	// Start with today = March 17
	oldToday := date(2026, time.March, 17)
	m := New(oldToday, oldToday, DefaultConfig())

	// Simulate midnight tick
	updated, cmd := m.Update(midnightTickMsg{})
	m = updated.(Model)

	// today should be updated to the real current time
	now := time.Now()
	expected := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if m.today != expected {
		t.Errorf("today = %s, want %s", m.today.Format(DateLayout), expected.Format(DateLayout))
	}

	// Should return a non-nil cmd to schedule the next tick
	if cmd == nil {
		t.Error("expected non-nil cmd for next midnight tick")
	}
}

func TestInitReturnsMidnightTick(t *testing.T) {
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init() to return a non-nil cmd for midnight tick")
	}
}

func TestUpdateHighlightChangedMsg(t *testing.T) {
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())

	// Simulate receiving a highlightChangedMsg
	newDates := map[time.Time]bool{
		time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC): true,
	}
	updated, cmd := m.Update(highlightChangedMsg{
		dates:   newDates,
		watcher: nil, // cmd is checked but never executed, so nil watcher is safe here
		path:    "/test/path",
	})
	m = updated.(Model)

	if len(m.highlightedDates) != 1 {
		t.Errorf("expected 1 highlighted date, got %d", len(m.highlightedDates))
	}
	key := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	if !m.highlightedDates[key] {
		t.Error("expected 2026-03-25 to be highlighted")
	}

	// Should return a cmd to wait for next change
	if cmd == nil {
		t.Error("expected non-nil cmd for next file watch")
	}
}

func TestUpdateHighlightChangedMsgNilDates(t *testing.T) {
	today := date(2026, time.March, 17)
	initialDates := map[time.Time]bool{
		time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC): true,
	}
	m := New(today, today, DefaultConfig(), WithHighlightedDates(initialDates))

	// Simulate file deletion (nil dates)
	updated, _ := m.Update(highlightChangedMsg{
		dates:   nil,
		watcher: nil,
		path:    "/test/path",
	})
	m = updated.(Model)

	if m.highlightedDates != nil {
		t.Errorf("expected nil highlightedDates, got %d dates", len(m.highlightedDates))
	}
}

func TestInitWithHighlightSource(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dates.json")
	if err := os.WriteFile(path, []byte(`["2026-03-25"]`), 0644); err != nil {
		t.Fatal(err)
	}

	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig(), WithHighlightSource(path))

	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init() to return a non-nil cmd")
	}
}
