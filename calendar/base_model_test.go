package calendar

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestBase() baseModel {
	cfg := DefaultConfig()
	colors := cfg.ResolvedColors()
	return baseModel{
		cursor: date(2026, time.March, 17),
		today:  date(2026, time.March, 17),
		config: cfg,
		help:   newHelpModel(colors),
		styles: buildStyles(colors),
	}
}

func TestHandleMsgWindowSize(t *testing.T) {
	t.Parallel()
	b := newTestBase()
	cmd, handled := b.handleMsg(tea.WindowSizeMsg{Width: 120, Height: 40})
	if !handled {
		t.Error("expected WindowSizeMsg to be handled")
	}
	if cmd != nil {
		t.Error("expected nil cmd for WindowSizeMsg")
	}
	if b.termWidth != 120 {
		t.Errorf("termWidth = %d, want 120", b.termWidth)
	}
	if b.termHeight != 40 {
		t.Errorf("termHeight = %d, want 40", b.termHeight)
	}
}

func TestHandleMsgWatcherErr(t *testing.T) {
	t.Parallel()
	b := newTestBase()
	cmd, handled := b.handleMsg(watcherErrMsg{err: nil})
	if !handled {
		t.Error("expected watcherErrMsg to be handled")
	}
	if cmd != nil {
		t.Error("expected nil cmd for watcherErrMsg")
	}
}

func TestHandleMsgUnhandledKeyMsg(t *testing.T) {
	t.Parallel()
	b := newTestBase()
	_, handled := b.handleMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if handled {
		t.Error("expected KeyMsg to not be handled by baseModel")
	}
}

func TestCloseWatcherNil(t *testing.T) {
	t.Parallel()
	b := newTestBase()
	b.activeWatcher = nil
	// Should not panic.
	b.closeWatcher()
	if b.activeWatcher != nil {
		t.Error("expected activeWatcher to remain nil")
	}
}

func TestDoQuit(t *testing.T) {
	t.Parallel()
	b := newTestBase()
	cmd := b.doQuit()
	if !b.quit {
		t.Error("expected quit to be true")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd from doQuit")
	}
}

func TestDoSelect(t *testing.T) {
	t.Parallel()
	b := newTestBase()
	cmd := b.doSelect()
	if !b.selected {
		t.Error("expected selected to be true")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd from doSelect")
	}
}

func TestDoVisualSelect(t *testing.T) {
	t.Parallel()
	b := newTestBase()
	b.doVisualSelect()
	if b.rangeAnchor == nil {
		t.Fatal("expected rangeAnchor to be set")
	}
	if !b.rangeAnchor.Equal(b.cursor) {
		t.Errorf("rangeAnchor = %s, want %s", b.rangeAnchor.Format("2006-01-02"), b.cursor.Format("2006-01-02"))
	}
}

func TestCancelRangeWithRange(t *testing.T) {
	t.Parallel()
	b := newTestBase()
	anchor := b.cursor
	b.rangeAnchor = &anchor
	wasSet := b.cancelRange()
	if !wasSet {
		t.Error("expected cancelRange to return true when range was set")
	}
	if b.rangeAnchor != nil {
		t.Error("expected rangeAnchor to be nil after cancel")
	}
}

func TestCancelRangeWithoutRange(t *testing.T) {
	t.Parallel()
	b := newTestBase()
	wasSet := b.cancelRange()
	if wasSet {
		t.Error("expected cancelRange to return false when no range was set")
	}
}

func TestInitCmdsWithoutHighlightPath(t *testing.T) {
	t.Parallel()
	b := newTestBase()
	cmds := b.initCmds()
	if len(cmds) != 1 {
		t.Errorf("expected 1 cmd (midnight tick), got %d", len(cmds))
	}
}

func TestInitCmdsWithHighlightPath(t *testing.T) {
	t.Parallel()
	b := newTestBase()
	b.highlightPath = "/some/path.json"
	cmds := b.initCmds()
	if len(cmds) != 2 {
		t.Errorf("expected 2 cmds (midnight tick + watcher), got %d", len(cmds))
	}
}

func TestWarningsAccessor(t *testing.T) {
	t.Parallel()
	b := newTestBase()
	if len(b.Warnings()) != 0 {
		t.Errorf("expected no warnings, got %v", b.Warnings())
	}
	b.warnings = append(b.warnings, "test warning")
	if len(b.Warnings()) != 1 {
		t.Errorf("expected 1 warning, got %d", len(b.Warnings()))
	}
	if b.Warnings()[0] != "test warning" {
		t.Errorf("warning = %q, want %q", b.Warnings()[0], "test warning")
	}
}
