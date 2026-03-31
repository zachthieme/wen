package calendar

import (
	"time"

	"github.com/zachthieme/wen"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// RowModel holds the state for the interactive strip calendar TUI.
type RowModel struct {
	cursor           time.Time
	today            time.Time
	quit             bool
	selected         bool
	rangeAnchor      *time.Time
	highlightedDates map[time.Time]bool
	highlightPath    string
	activeWatcher    *fsnotify.Watcher
	config           Config
	keys             rowKeyMap
	help             help.Model
	styles           resolvedStyles
	showHelp         bool
}

// RowModelOption configures optional RowModel properties.
type RowModelOption func(*RowModel)

// WithRowHighlightedDates sets dates to be visually highlighted in the row calendar.
// Clears any highlight source path, disabling file watching.
func WithRowHighlightedDates(dates map[time.Time]bool) RowModelOption {
	return func(m *RowModel) {
		m.highlightedDates = dates
		m.highlightPath = ""
	}
}

// WithRowHighlightSource sets the path to a JSON file of dates to highlight.
// It expands ~ to the user's home directory, performs the initial load, and
// enables file watching when Init() runs.
func WithRowHighlightSource(path string) RowModelOption {
	return func(m *RowModel) {
		m.highlightPath = expandTilde(path)
		m.highlightedDates = LoadHighlightedDates(m.highlightPath)
	}
}

// NewRow creates a RowModel with the given cursor position, today's date, and configuration.
func NewRow(cursor, today time.Time, cfg Config, opts ...RowModelOption) RowModel {
	colors := cfg.ResolvedColors()
	m := RowModel{
		cursor: wen.TruncateDay(cursor),
		today:  wen.TruncateDay(today),
		config: cfg,
		keys:   defaultRowKeyMap(),
		help:   newHelpModel(colors),
		styles: buildStyles(colors),
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// IsQuit reports whether the user quit without selecting.
func (m RowModel) IsQuit() bool { return m.quit }

// Selected reports whether the user selected a date with Enter.
func (m RowModel) Selected() bool { return m.selected }

// Cursor returns the currently selected date.
func (m RowModel) Cursor() time.Time { return m.cursor }

// Init is a stub that satisfies the tea.Model interface.
// Will be fully implemented in a later task.
func (m RowModel) Init() tea.Cmd {
	if m.activeWatcher != nil {
		// Placeholder: watcher integration comes in a later task.
		_ = m.activeWatcher
	}
	return nil
}

// Update is a stub that satisfies the tea.Model interface.
// Will be fully implemented in a later task.
func (m RowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	_ = msg
	return m, nil
}

// View is a stub that satisfies the tea.Model interface.
// Will be fully implemented in a later task.
func (m RowModel) View() string {
	if m.showHelp {
		// Placeholder: help rendering comes in a later task.
		_ = m.help.View(m.keys)
	}
	return ""
}

type rowKeyMap struct {
	Left         key.Binding
	Right        key.Binding
	PrevMonth    key.Binding
	NextMonth    key.Binding
	WeekStart    key.Binding
	WeekEnd      key.Binding
	MonthStart   key.Binding
	MonthEnd     key.Binding
	Today        key.Binding
	ToggleHelp   key.Binding
	VisualSelect key.Binding
	Select       key.Binding
	Quit         key.Binding
	ForceQuit    key.Binding
}

func defaultRowKeyMap() rowKeyMap {
	return rowKeyMap{
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/\u2190", "prev day"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/\u2192", "next day"),
		),
		PrevMonth: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/\u2191", "prev month"),
		),
		NextMonth: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/\u2193", "next month"),
		),
		WeekStart: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "week start"),
		),
		WeekEnd: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "week end"),
		),
		MonthStart: key.NewBinding(
			key.WithKeys("0"),
			key.WithHelp("0", "month start"),
		),
		MonthEnd: key.NewBinding(
			key.WithKeys("$"),
			key.WithHelp("$", "month end"),
		),
		Today: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "today"),
		),
		ToggleHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		VisualSelect: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "range"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "force quit"),
		),
	}
}

// ShortHelp returns bindings for the short help view.
func (k rowKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.VisualSelect, k.Select, k.Quit, k.ToggleHelp}
}

// FullHelp returns bindings for the full help view.
func (k rowKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.PrevMonth, k.NextMonth},
		{k.WeekStart, k.WeekEnd, k.MonthStart, k.MonthEnd},
		{k.Today, k.VisualSelect, k.Select, k.Quit},
	}
}
