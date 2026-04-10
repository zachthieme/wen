package calendar

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/zachthieme/wen"

	tea "github.com/charmbracelet/bubbletea"
)

// Test helpers (date, runeMsg, specialMsg, press) are in testhelpers_test.go.

func TestNavigation(t *testing.T) {
	t.Parallel()
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
		{"next year (N)", runeMsg("N"), today, date(2027, time.March, 17)},
		{"prev year (P)", runeMsg("P"), today, date(2025, time.March, 17)},
		{"next year leap day clamp", runeMsg("N"), date(2024, time.February, 29), date(2025, time.February, 28)},

		// Boundary wrapping
		{"day wrap forward", runeMsg("l"), date(2026, time.March, 31), date(2026, time.April, 1)},
		{"day wrap backward", runeMsg("h"), date(2026, time.March, 1), date(2026, time.February, 28)},
		{"week wrap forward", runeMsg("j"), date(2026, time.March, 28), date(2026, time.April, 4)},

		// Jump to today
		{"jump to today (t)", runeMsg("t"), date(2026, time.June, 15), today},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := New(tt.start, today, DefaultConfig())
			updated, _ := m.Update(tt.msg)
			got := updated.(Model)
			if got.cursor != tt.expected {
				t.Errorf("got %s, want %s", got.cursor.Format(wen.DateLayout), tt.expected.Format(wen.DateLayout))
			}
		})
	}
}

func TestToggleWeekNumbers(t *testing.T) {
	t.Parallel()
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	if m.weekNumPos != WeekNumOff {
		t.Errorf("expected weekNumPos off initially, got %d", m.weekNumPos)
	}
	m = press(m, "w")
	if m.weekNumPos != WeekNumLeft {
		t.Errorf("expected weekNumPos left after first toggle, got %d", m.weekNumPos)
	}
	m = press(m, "w")
	if m.weekNumPos != WeekNumRight {
		t.Errorf("expected weekNumPos right after second toggle, got %d", m.weekNumPos)
	}
	m = press(m, "w")
	if m.weekNumPos != WeekNumOff {
		t.Errorf("expected weekNumPos off after third toggle, got %d", m.weekNumPos)
	}
}

func TestToggleHelp(t *testing.T) {
	t.Parallel()
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	if m.showHelp {
		t.Error("expected showHelp false initially")
	}
	m = press(m, "?")
	if !m.showHelp {
		t.Error("expected showHelp true after toggle")
	}
}

func TestQuitExits(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	// Press v to anchor
	m = press(m, "v")
	// Move 5 days right
	for i := 0; i < 5; i++ {
		m = press(m, "l")
	}
	// Press Enter to confirm
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.HasRange() {
		t.Error("expected HasRange to be true")
	}
	if m.RangeStart() != date(2026, time.March, 17) {
		t.Errorf("RangeStart got %s, want 2026-03-17", m.RangeStart().Format(wen.DateLayout))
	}
	if m.RangeEnd() != date(2026, time.March, 22) {
		t.Errorf("RangeEnd got %s, want 2026-03-22", m.RangeEnd().Format(wen.DateLayout))
	}
}

func TestVisualSelectCancel(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	// Press v to anchor
	m = press(m, "v")
	// Move right
	m = press(m, "l")
	// Press Esc to cancel range (not quit)
	updated, cmd := m.Update(specialMsg(tea.KeyEscape))
	m = updated.(Model)
	if m.IsQuit() {
		t.Error("expected IsQuit to be false after first Esc (cancel range)")
	}
	if m.HasRange() {
		t.Error("expected HasRange to be false after Esc cancel")
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
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	// Press v to anchor
	m = press(m, "v")
	// Move 2 right
	m = press(m, "l")
	m = press(m, "l")
	// Press v again to re-anchor at March 19
	m = press(m, "v")
	// Move 1 right
	m = press(m, "l")
	// Press Enter to confirm
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.HasRange() {
		t.Error("expected HasRange to be true")
	}
	if m.RangeStart() != date(2026, time.March, 19) {
		t.Errorf("RangeStart got %s, want 2026-03-19", m.RangeStart().Format(wen.DateLayout))
	}
	if m.RangeEnd() != date(2026, time.March, 20) {
		t.Errorf("RangeEnd got %s, want 2026-03-20", m.RangeEnd().Format(wen.DateLayout))
	}
}

func TestRangeReverseOrder(t *testing.T) {
	t.Parallel()
	start := date(2026, time.March, 20)
	m := New(start, start, DefaultConfig())
	// Press v to anchor at March 20
	m = press(m, "v")
	// Move 5 left
	for i := 0; i < 5; i++ {
		m = press(m, "h")
	}
	// Press Enter to confirm
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.HasRange() {
		t.Error("expected HasRange to be true")
	}
	if m.RangeStart() != date(2026, time.March, 15) {
		t.Errorf("RangeStart got %s, want 2026-03-15", m.RangeStart().Format(wen.DateLayout))
	}
	if m.RangeEnd() != date(2026, time.March, 20) {
		t.Errorf("RangeEnd got %s, want 2026-03-20", m.RangeEnd().Format(wen.DateLayout))
	}
}

func TestEnterWithoutRange(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	// Press Enter without v
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.Selected() {
		t.Error("expected Selected to be true")
	}
	if m.HasRange() {
		t.Error("expected HasRange to be false")
	}
}

func TestSameDayRange(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	// Press v to anchor
	m = press(m, "v")
	// Immediately press Enter (same day)
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.Selected() {
		t.Error("expected Selected to be true")
	}
	if m.HasRange() {
		t.Error("expected HasRange to be false for same day")
	}
}

func TestCtrlCHasRangeMode(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	// Press v to anchor
	m = press(m, "v")
	// Move right
	m = press(m, "l")
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

func TestDateCheckUpdatesToday(t *testing.T) {
	t.Parallel()
	// Start with today = March 17
	oldToday := date(2026, time.March, 17)
	m := New(oldToday, oldToday, DefaultConfig())

	// Simulate date check tick
	updated, cmd := m.Update(dateCheckMsg{})
	m = updated.(Model)

	// today should be updated to the real current time
	now := time.Now()
	expected := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if m.today != expected {
		t.Errorf("today = %s, want %s", m.today.Format(wen.DateLayout), expected.Format(wen.DateLayout))
	}

	// Should return a non-nil cmd to schedule the next check
	if cmd == nil {
		t.Error("expected non-nil cmd for next date check")
	}
}

func TestInitReturnsDateCheck(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init() to return a non-nil cmd for date check")
	}
}

func TestUpdateHighlightChangedMsg(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

func TestWithJulian(t *testing.T) {
	t.Parallel()
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig(), WithJulian(true))
	if !m.julian {
		t.Error("expected julian to be true")
	}
}

func TestWithPrintMode(t *testing.T) {
	t.Parallel()
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig(), WithPrintMode(true))
	if !m.printMode {
		t.Error("expected printMode to be true")
	}
}

func TestInitWithHighlightSource(t *testing.T) {
	t.Parallel()
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

func TestToggleJulian(t *testing.T) {
	t.Parallel()
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	if m.julian {
		t.Error("expected julian false initially")
	}
	if m.dayFmt.gridWidth != 20 {
		t.Errorf("expected gridWidth 20 initially, got %d", m.dayFmt.gridWidth)
	}
	m = press(m, "J")
	if !m.julian {
		t.Error("expected julian true after toggle")
	}
	if m.dayFmt.gridWidth != 27 {
		t.Errorf("expected gridWidth 27 after julian toggle, got %d", m.dayFmt.gridWidth)
	}
	m = press(m, "J")
	if m.julian {
		t.Error("expected julian false after second toggle")
	}
	if m.dayFmt.gridWidth != 20 {
		t.Errorf("expected gridWidth 20 after second toggle, got %d", m.dayFmt.gridWidth)
	}
}

func TestYearNavigationRebound(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	// N = next year
	m = press(m, "N")
	if m.cursor != date(2027, time.March, 17) {
		t.Errorf("N should navigate to next year, got %s", m.cursor.Format("2006-01-02"))
	}
	// P = prev year
	m = press(m, "P")
	if m.cursor != date(2026, time.March, 17) {
		t.Errorf("P should navigate to prev year, got %s", m.cursor.Format("2006-01-02"))
	}
}
