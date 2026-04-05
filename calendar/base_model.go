package calendar

import (
	"fmt"
	"time"

	"github.com/zachthieme/wen"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
)

// baseModel holds the shared state between Model and RowModel.
// It is embedded in both and provides common accessors, lifecycle helpers,
// and message handling. Pointer-receiver methods mutate the embedded copy
// that Bubble Tea's Update returns; value-receiver methods are read-only.
type baseModel struct {
	cursor           time.Time
	today            time.Time
	quit             bool
	selected         bool
	rangeAnchor      *time.Time
	highlightedDates map[time.Time]bool
	highlightPath    string
	activeWatcher    *fsnotify.Watcher
	config           Config
	help             help.Model
	styles           resolvedStyles
	showHelp         bool
	julian           bool
	printMode        bool
	dayFmt           dayFormat
	termWidth        int
	termHeight       int
	months           int
	warnings         []Warning
}

// Option configures calendar model properties.
type Option func(*baseModel)

// WithHighlightedDates sets dates to be visually highlighted.
// Clears any highlight source path, disabling file watching.
func WithHighlightedDates(dates map[time.Time]bool) Option {
	return func(b *baseModel) {
		b.highlightedDates = dates
		b.highlightPath = ""
	}
}

// WithJulian enables Julian day-of-year numbering.
func WithJulian(on bool) Option {
	return func(b *baseModel) {
		b.julian = on
	}
}

// WithPrintMode enables non-interactive print mode (suppresses cursor styling).
func WithPrintMode(on bool) Option {
	return func(b *baseModel) {
		b.printMode = on
	}
}

// WithMonths sets the number of months to display side by side (grid calendar).
// This option only affects [Model]; [RowModel] always displays a single strip.
func WithMonths(n int) Option {
	return func(b *baseModel) {
		if n < 1 {
			n = 1
		}
		b.months = n
	}
}

// WithTermWidth sets the terminal width for print mode rendering.
func WithTermWidth(w int) Option {
	return func(b *baseModel) {
		b.termWidth = w
	}
}

// resolvedStyles holds pre-computed lipgloss styles for calendar rendering.
type resolvedStyles struct {
	cursor      lipgloss.Style
	cursorToday lipgloss.Style
	today       lipgloss.Style
	highlight   lipgloss.Style
	rangeDay    lipgloss.Style
	title       lipgloss.Style
	weekNum     lipgloss.Style
	dayHeader   lipgloss.Style
	helpBar     lipgloss.Style
}

// Warnings returns any warnings collected during initialization (e.g. highlight parse issues).
func (b baseModel) Warnings() []Warning { return b.warnings }

// IsQuit reports whether the user quit without selecting.
func (b baseModel) IsQuit() bool { return b.quit }

// Selected reports whether the user selected a date with Enter.
func (b baseModel) Selected() bool { return b.selected }

// Cursor returns the currently selected date.
func (b baseModel) Cursor() time.Time { return b.cursor }

// HasRange reports whether the user confirmed a multi-day range selection.
func (b baseModel) HasRange() bool {
	return b.selected && b.rangeAnchor != nil && !b.rangeAnchor.Equal(b.cursor)
}

// RangeStart returns the earlier date of the confirmed range, or zero if no range.
func (b baseModel) RangeStart() time.Time {
	if !b.HasRange() {
		return time.Time{}
	}
	if b.rangeAnchor.Before(b.cursor) {
		return *b.rangeAnchor
	}
	return b.cursor
}

// RangeEnd returns the later date of the confirmed range, or zero if no range.
func (b baseModel) RangeEnd() time.Time {
	if !b.HasRange() {
		return time.Time{}
	}
	if b.rangeAnchor.After(b.cursor) {
		return *b.rangeAnchor
	}
	return b.cursor
}

// initCmds returns the tea.Cmds that both Model and RowModel schedule from Init().
func (b baseModel) initCmds() []tea.Cmd {
	cmds := []tea.Cmd{scheduleMidnightTick(b.today)}
	if b.highlightPath != "" {
		cmds = append(cmds, startFileWatcher(b.highlightPath))
	}
	return cmds
}

// handleMsg processes messages shared between Model and RowModel.
// It returns the command to send and whether the message was handled.
// If handled is true, the caller should return immediately.
func (b *baseModel) handleMsg(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.help.Width = msg.Width
		b.termWidth = msg.Width
		b.termHeight = msg.Height
		return nil, true
	case watcherErrMsg:
		b.warnings = append(b.warnings, Warning{
			Key:     "watcher",
			Message: fmt.Sprintf("file watcher error: %v", msg.err),
		})
		return nil, true
	case midnightTickMsg:
		now := time.Now()
		b.today = wen.TruncateDay(now)
		return scheduleMidnightTick(now), true
	case highlightChangedMsg:
		b.highlightedDates = msg.dates
		b.activeWatcher = msg.watcher
		return waitForNextChange(msg.watcher, msg.path), true
	}
	return nil, false
}

// closeWatcher nil-checks and closes the activeWatcher.
func (b *baseModel) closeWatcher() {
	if b.activeWatcher != nil {
		_ = b.activeWatcher.Close()
		b.activeWatcher = nil
	}
}

// doQuit sets quit, closes the watcher, and returns tea.Quit.
func (b *baseModel) doQuit() tea.Cmd {
	b.quit = true
	b.closeWatcher()
	return tea.Quit
}

// doSelect sets selected, closes the watcher, and returns tea.Quit.
func (b *baseModel) doSelect() tea.Cmd {
	b.selected = true
	b.closeWatcher()
	return tea.Quit
}

// doVisualSelect sets the range anchor to the current cursor position.
func (b *baseModel) doVisualSelect() {
	anchor := b.cursor
	b.rangeAnchor = &anchor
}

// cancelRange clears the range anchor and reports whether one was active.
func (b *baseModel) cancelRange() bool {
	if b.rangeAnchor != nil {
		b.rangeAnchor = nil
		return true
	}
	return false
}
