package calendar

import (
	"strings"
	"testing"
	"time"

	"github.com/zachthieme/wen"

	tea "github.com/charmbracelet/bubbletea"
)

// Test helpers (date, runeMsg, specialMsg, press) are in testhelpers_test.go.

func TestNewRow(t *testing.T) {
	t.Parallel()
	cursor := date(2026, time.March, 15)
	today := date(2026, time.March, 30)
	m := NewRow(cursor, today, DefaultConfig())

	if m.Cursor() != cursor {
		t.Errorf("Cursor() = %s, want %s", m.Cursor().Format(wen.DateLayout), cursor.Format(wen.DateLayout))
	}
	if m.IsQuit() {
		t.Error("expected IsQuit() to be false")
	}
	if m.Selected() {
		t.Error("expected Selected() to be false")
	}
	if m.HasRange() {
		t.Error("expected HasRange() to be false")
	}
}

func TestNewRowWithHighlightedDates(t *testing.T) {
	t.Parallel()
	cursor := date(2026, time.March, 15)
	today := date(2026, time.March, 30)
	highlights := map[time.Time]bool{
		dateKey(date(2026, time.March, 20)): true,
		dateKey(date(2026, time.March, 25)): true,
	}
	m := NewRow(cursor, today, DefaultConfig(), WithHighlightedDates(highlights))

	if len(m.highlightedDates) != 2 {
		t.Errorf("expected 2 highlighted dates, got %d", len(m.highlightedDates))
	}
	if !m.highlightedDates[dateKey(date(2026, time.March, 20))] {
		t.Error("expected March 20 to be highlighted")
	}
}

func TestRowInitReturnsMidnightTick(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := NewRow(today, today, DefaultConfig())
	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init() to return a non-nil cmd for midnight tick")
	}
}

func TestRowMidnightTickUpdatesToday(t *testing.T) {
	t.Parallel()
	oldToday := date(2026, time.March, 17)
	m := NewRow(oldToday, oldToday, DefaultConfig())

	updated, cmd := m.Update(midnightTickMsg{})
	m = updated.(RowModel)

	now := time.Now()
	expected := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if m.today != expected {
		t.Errorf("today = %s, want %s", m.today.Format(wen.DateLayout), expected.Format(wen.DateLayout))
	}
	if cmd == nil {
		t.Error("expected non-nil cmd for next midnight tick")
	}
}

func TestRowUpdateHighlightChanged(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := NewRow(today, today, DefaultConfig())

	newDates := map[time.Time]bool{
		time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC): true,
	}
	updated, cmd := m.Update(highlightChangedMsg{
		dates:   newDates,
		watcher: nil,
		path:    "/test/path",
	})
	m = updated.(RowModel)

	if len(m.highlightedDates) != 1 {
		t.Errorf("expected 1 highlighted date, got %d", len(m.highlightedDates))
	}
	key := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	if !m.highlightedDates[key] {
		t.Error("expected 2026-03-25 to be highlighted")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd for next file watch")
	}
}

func TestRowView(t *testing.T) {
	t.Parallel()
	cursor := date(2026, time.March, 15)
	m := NewRow(cursor, cursor, DefaultConfig())
	view := m.View()

	if !strings.Contains(view, "Mr") {
		t.Errorf("View() should contain month abbreviation 'Mr', got: %q", view)
	}
	// Should contain day header abbreviations
	if !strings.Contains(view, "Su") {
		t.Errorf("View() should contain day header 'Su', got: %q", view)
	}
	if !strings.Contains(view, "Mo") {
		t.Errorf("View() should contain day header 'Mo', got: %q", view)
	}
	// Should contain the cursor day
	if !strings.Contains(view, "15") {
		t.Errorf("View() should contain cursor day '15', got: %q", view)
	}
}

func TestRowViewHelpBar(t *testing.T) {
	t.Parallel()
	cursor := date(2026, time.March, 15)
	m := NewRow(cursor, cursor, DefaultConfig())
	m.showHelp = true
	view := m.View()

	// Help bar should contain key binding descriptions
	if !strings.Contains(view, "quit") {
		t.Errorf("View() with help should contain 'quit', got: %q", view)
	}
}

func TestRowVisibleWindowFitsFullStrip(t *testing.T) {
	t.Parallel()
	m := NewRow(date(2026, time.March, 15), date(2026, time.March, 15), DefaultConfig())
	m.termWidth = 200 // wider than any strip
	output := m.View()
	// Full March strip should have all 31 days
	if !strings.Contains(output, "31") {
		t.Error("expected full strip with day 31")
	}
}

func TestRowVisibleWindowTrimmed(t *testing.T) {
	t.Parallel()
	m := NewRow(date(2026, time.March, 15), date(2026, time.March, 15), DefaultConfig())
	m.termWidth = 50 // narrow: fits (50-2)/3 = 16 days
	output := m.View()
	// Cursor (15) should be visible
	if !strings.Contains(output, "15") {
		t.Error("expected cursor day 15 in narrow view")
	}
	// Day 1 might not be visible (cursor centered)
	// Strip should be narrower than 50
	lines := strings.Split(output, "\n")
	// Should have fewer columns than full strip (107 chars)
	if len(lines[0]) > 0 && len(lines[0]) > 60 {
		t.Errorf("trimmed strip should be narrower, got %d chars", len(lines[0]))
	}
}

func TestRowVisibleWindowResize(t *testing.T) {
	t.Parallel()
	m := NewRow(date(2026, time.March, 15), date(2026, time.March, 15), DefaultConfig())
	// Simulate resize
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m = updated.(RowModel)
	if m.termWidth != 60 {
		t.Errorf("termWidth = %d, want 60", m.termWidth)
	}
}

func TestRowVisibleWindowJulianNarrower(t *testing.T) {
	t.Parallel()
	// Julian cells are 4 chars wide vs 3 for normal, so fewer days should fit.
	normal := NewRow(date(2026, time.March, 15), date(2026, time.March, 15), DefaultConfig())
	normal.termWidth = 80

	julian := NewRow(date(2026, time.March, 15), date(2026, time.March, 15), DefaultConfig(), WithJulian(true))
	julian.termWidth = 80

	// Normal: maxDays = (80-2)/3 = 26
	// Julian: maxDays = (80-3)/4 = 19
	normalStart, normalEnd := stripWindow(2026, time.March, 0, time.Local)
	julianStart, julianEnd := stripWindow(2026, time.March, 0, time.Local)

	ns, ne := normal.visibleWindow(normalStart, normalEnd)
	js, je := julian.visibleWindow(julianStart, julianEnd)

	normalDays := dayCount(ns, ne)
	julianDays := dayCount(js, je)

	if julianDays >= normalDays {
		t.Errorf("julian should show fewer days than normal at same width: julian=%d, normal=%d", julianDays, normalDays)
	}
}

func TestRowViewNoHelpByDefault(t *testing.T) {
	t.Parallel()
	cursor := date(2026, time.March, 15)
	m := NewRow(cursor, cursor, DefaultConfig())
	view := m.View()

	if strings.Contains(view, "quit") {
		t.Errorf("View() without help should not contain 'quit', got: %q", view)
	}
}

func TestRowNavigation(t *testing.T) {
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
		{"next day (right)", specialMsg(tea.KeyRight), today, date(2026, time.March, 18)},
		{"prev day (left)", specialMsg(tea.KeyLeft), today, date(2026, time.March, 16)},

		// Month navigation (j/k/arrows)
		{"next month (j)", runeMsg("j"), today, date(2026, time.April, 17)},
		{"prev month (k)", runeMsg("k"), today, date(2026, time.February, 17)},
		{"next month (down)", specialMsg(tea.KeyDown), today, date(2026, time.April, 17)},
		{"prev month (up)", specialMsg(tea.KeyUp), today, date(2026, time.February, 17)},

		// Jump to today
		{"jump to today (t)", runeMsg("t"), date(2026, time.June, 15), today},

		// Month clamping
		{"next month clamps day", runeMsg("j"), date(2026, time.January, 31), date(2026, time.February, 28)},
		{"prev month clamps day", runeMsg("k"), date(2026, time.March, 31), date(2026, time.February, 28)},

		// Day wrapping across months
		{"day wrap forward", runeMsg("l"), date(2026, time.March, 31), date(2026, time.April, 1)},
		{"day wrap backward", runeMsg("h"), date(2026, time.March, 1), date(2026, time.February, 28)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := NewRow(tt.start, today, DefaultConfig())
			updated, _ := m.Update(tt.msg)
			got := updated.(RowModel)
			if got.cursor != tt.expected {
				t.Errorf("got %s, want %s", got.cursor.Format(wen.DateLayout), tt.expected.Format(wen.DateLayout))
			}
		})
	}
}

func TestRowVimMotions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		start    time.Time
		keys     []string
		expected time.Time
		config   Config
	}{
		{
			name:     "week start from mid-week",
			start:    date(2026, time.March, 18), // Wednesday
			keys:     []string{"b"},
			expected: date(2026, time.March, 15), // Sunday (week start)
		},
		{
			name:     "week start repeat",
			start:    date(2026, time.March, 18), // Wednesday
			keys:     []string{"b", "b"},
			expected: date(2026, time.March, 8), // Previous Sunday
		},
		{
			name:     "week end from mid-week",
			start:    date(2026, time.March, 18), // Wednesday
			keys:     []string{"e"},
			expected: date(2026, time.March, 21), // Saturday (week end)
		},
		{
			name:     "week end repeat",
			start:    date(2026, time.March, 18), // Wednesday
			keys:     []string{"e", "e"},
			expected: date(2026, time.March, 28), // Next Saturday
		},
		{
			name:     "month start",
			start:    date(2026, time.March, 18),
			keys:     []string{"0"},
			expected: date(2026, time.March, 1),
		},
		{
			name:     "month end",
			start:    date(2026, time.March, 18),
			keys:     []string{"$"},
			expected: date(2026, time.March, 31),
		},
		{
			name:     "week start with Monday config",
			start:    date(2026, time.March, 18), // Wednesday
			keys:     []string{"b"},
			expected: date(2026, time.March, 16), // Monday
			config: Config{
				WeekStartDay:  1,
				WeekNumbering: WeekNumberingUS,
				Theme:         "default",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := tt.config
			if cfg.Theme == "" {
				cfg = DefaultConfig()
			}
			m := NewRow(tt.start, tt.start, cfg)
			for _, k := range tt.keys {
				m = press(m, k)
			}
			if m.cursor != tt.expected {
				t.Errorf("got %s, want %s", m.cursor.Format(wen.DateLayout), tt.expected.Format(wen.DateLayout))
			}
		})
	}
}

func TestWeekStartDate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		t            time.Time
		weekStartDay int
		expected     time.Time
	}{
		{
			name:         "Wednesday to Sunday start",
			t:            date(2026, time.March, 18),
			weekStartDay: 0,
			expected:     date(2026, time.March, 15),
		},
		{
			name:         "Sunday to previous Sunday",
			t:            date(2026, time.March, 15),
			weekStartDay: 0,
			expected:     date(2026, time.March, 8),
		},
		{
			name:         "Wednesday to Monday start",
			t:            date(2026, time.March, 18),
			weekStartDay: 1,
			expected:     date(2026, time.March, 16),
		},
		{
			name:         "Monday to previous Monday",
			t:            date(2026, time.March, 16),
			weekStartDay: 1,
			expected:     date(2026, time.March, 9),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := weekStartDate(tt.t, tt.weekStartDay)
			if got != tt.expected {
				t.Errorf("weekStartDate(%s, %d) = %s, want %s",
					tt.t.Format("2006-01-02"), tt.weekStartDay,
					got.Format("2006-01-02"), tt.expected.Format("2006-01-02"))
			}
		})
	}
}

func TestWeekEndDate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		t            time.Time
		weekStartDay int
		expected     time.Time
	}{
		{
			name:         "Wednesday to Saturday end (Sunday start)",
			t:            date(2026, time.March, 18),
			weekStartDay: 0,
			expected:     date(2026, time.March, 21),
		},
		{
			name:         "Saturday to next Saturday",
			t:            date(2026, time.March, 21),
			weekStartDay: 0,
			expected:     date(2026, time.March, 28),
		},
		{
			name:         "Wednesday to Sunday end (Monday start)",
			t:            date(2026, time.March, 18),
			weekStartDay: 1,
			expected:     date(2026, time.March, 22),
		},
		{
			name:         "Sunday to next Sunday (Monday start)",
			t:            date(2026, time.March, 22),
			weekStartDay: 1,
			expected:     date(2026, time.March, 29),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := weekEndDate(tt.t, tt.weekStartDay)
			if got != tt.expected {
				t.Errorf("weekEndDate(%s, %d) = %s, want %s",
					tt.t.Format("2006-01-02"), tt.weekStartDay,
					got.Format("2006-01-02"), tt.expected.Format("2006-01-02"))
			}
		})
	}
}

func TestRowQuitExits(t *testing.T) {
	t.Parallel()
	m := NewRow(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	updated, cmd := m.Update(runeMsg("q"))
	m = updated.(RowModel)
	if !m.IsQuit() {
		t.Error("expected IsQuit to be true")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestRowEscQuits(t *testing.T) {
	t.Parallel()
	m := NewRow(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	updated, cmd := m.Update(specialMsg(tea.KeyEscape))
	m = updated.(RowModel)
	if !m.IsQuit() {
		t.Error("expected IsQuit to be true")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestRowCtrlCForceQuits(t *testing.T) {
	t.Parallel()
	m := NewRow(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	// Enter range mode first to test force quit escapes everything
	m = press(m, "v")
	m = press(m, "l")
	updated, cmd := m.Update(specialMsg(tea.KeyCtrlC))
	m = updated.(RowModel)
	if !m.IsQuit() {
		t.Error("expected IsQuit to be true after ctrl+c")
	}
	if cmd == nil {
		t.Error("expected quit command after ctrl+c")
	}
}

func TestRowEnterSelects(t *testing.T) {
	t.Parallel()
	cursor := date(2026, time.April, 15)
	m := NewRow(cursor, date(2026, time.March, 17), DefaultConfig())
	updated, cmd := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(RowModel)
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

func TestRowToggleHelp(t *testing.T) {
	t.Parallel()
	m := NewRow(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	if m.showHelp {
		t.Error("expected showHelp false initially")
	}
	m = press(m, "?")
	if !m.showHelp {
		t.Error("expected showHelp true after toggle")
	}
	m = press(m, "?")
	if m.showHelp {
		t.Error("expected showHelp false after second toggle")
	}
}

func TestRowVisualSelectEnter(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := NewRow(today, today, DefaultConfig())
	// Press v to anchor
	m = press(m, "v")
	// Move 5 days right
	for range 5 {
		m = press(m, "l")
	}
	// Press Enter to confirm
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(RowModel)
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

func TestRowVisualSelectCancel(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := NewRow(today, today, DefaultConfig())
	// Press v to anchor
	m = press(m, "v")
	// Move right
	m = press(m, "l")
	// Press Esc to cancel range (not quit)
	updated, cmd := m.Update(specialMsg(tea.KeyEscape))
	m = updated.(RowModel)
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
	m = updated.(RowModel)
	if !m.IsQuit() {
		t.Error("expected IsQuit to be true after second Esc")
	}
	if cmd == nil {
		t.Error("expected quit command after second Esc")
	}
}

func TestRowSameDayRange(t *testing.T) {
	t.Parallel()
	today := date(2026, time.March, 17)
	m := NewRow(today, today, DefaultConfig())
	// Press v to anchor
	m = press(m, "v")
	// Immediately press Enter (same day)
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(RowModel)
	if !m.Selected() {
		t.Error("expected Selected to be true")
	}
	if m.HasRange() {
		t.Error("expected HasRange to be false for same day")
	}
}

func TestRowToggleJulian(t *testing.T) {
	t.Parallel()
	m := NewRow(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	if m.julian {
		t.Error("expected julian false initially")
	}
	if m.dayFmt.cellWidth != 2 {
		t.Errorf("expected cellWidth 2 initially, got %d", m.dayFmt.cellWidth)
	}
	updated, _ := m.Update(runeMsg("J"))
	m = updated.(RowModel)
	if !m.julian {
		t.Error("expected julian true after toggle")
	}
	if m.dayFmt.cellWidth != 3 {
		t.Errorf("expected cellWidth 3 after julian toggle, got %d", m.dayFmt.cellWidth)
	}
	updated, _ = m.Update(runeMsg("J"))
	m = updated.(RowModel)
	if m.julian {
		t.Error("expected julian false after second toggle")
	}
	if m.dayFmt.cellWidth != 2 {
		t.Errorf("expected cellWidth 2 after second toggle, got %d", m.dayFmt.cellWidth)
	}
}

func TestRowWithJulian(t *testing.T) {
	t.Parallel()
	m := NewRow(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig(), WithJulian(true))
	if !m.julian {
		t.Error("expected julian to be true")
	}
}

func TestRowWithPrintMode(t *testing.T) {
	t.Parallel()
	m := NewRow(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig(), WithPrintMode(true))
	if !m.printMode {
		t.Error("expected printMode to be true")
	}
}

func TestRowRangeReverseOrder(t *testing.T) {
	t.Parallel()
	start := date(2026, time.March, 20)
	m := NewRow(start, start, DefaultConfig())
	// Press v to anchor at March 20
	m = press(m, "v")
	// Move 5 left
	for range 5 {
		m = press(m, "h")
	}
	// Press Enter to confirm
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(RowModel)
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
