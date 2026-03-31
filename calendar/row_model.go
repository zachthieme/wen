package calendar

import (
	"strings"
	"time"

	"github.com/zachthieme/wen"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	m.styles.padding = lipgloss.NewStyle().Padding(
		cfg.PaddingTop, cfg.PaddingRight, cfg.PaddingBottom, cfg.PaddingLeft,
	)
	// Strip Underline from row styles. lipgloss renders Underline per-character
	// (each char gets its own ANSI open/close), which causes terminals like mosh
	// to miscalculate cursor positions and misalign the strip columns.
	m.styles.today = m.styles.today.Underline(false)
	m.styles.cursorToday = m.styles.cursorToday.Underline(false)
	m.styles.highlight = m.styles.highlight.Underline(false)
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

// InRange reports whether the user confirmed a multi-day range selection.
func (m RowModel) InRange() bool {
	return m.selected && m.rangeAnchor != nil && !m.rangeAnchor.Equal(m.cursor)
}

// RangeStart returns the earlier date of the confirmed range, or zero if no range.
func (m RowModel) RangeStart() time.Time {
	if !m.InRange() {
		return time.Time{}
	}
	if m.rangeAnchor.Before(m.cursor) {
		return *m.rangeAnchor
	}
	return m.cursor
}

// RangeEnd returns the later date of the confirmed range, or zero if no range.
func (m RowModel) RangeEnd() time.Time {
	if !m.InRange() {
		return time.Time{}
	}
	if m.rangeAnchor.After(m.cursor) {
		return *m.rangeAnchor
	}
	return m.cursor
}

// Init schedules the midnight tick and, if a highlight source path is configured,
// starts an fsnotify file watcher.
func (m RowModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, scheduleMidnightTick(m.today))
	if m.highlightPath != "" {
		cmds = append(cmds, startFileWatcher(m.highlightPath))
	}
	return tea.Batch(cmds...)
}

// Update handles input messages and updates model state.
func (m RowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		return m, nil
	case watcherErrMsg:
		return m, nil
	case midnightTickMsg:
		now := time.Now()
		m.today = wen.TruncateDay(now)
		return m, scheduleMidnightTick(now)
	case highlightChangedMsg:
		m.highlightedDates = msg.dates
		m.activeWatcher = msg.watcher
		return m, waitForNextChange(msg.watcher, msg.path)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.ForceQuit):
			m.quit = true
			if m.activeWatcher != nil {
				_ = m.activeWatcher.Close()
				m.activeWatcher = nil
			}
			return m, tea.Quit
		case key.Matches(msg, m.keys.VisualSelect):
			anchor := m.cursor
			m.rangeAnchor = &anchor
			return m, nil
		case key.Matches(msg, m.keys.Select):
			m.selected = true
			if m.activeWatcher != nil {
				_ = m.activeWatcher.Close()
				m.activeWatcher = nil
			}
			return m, tea.Quit
		case key.Matches(msg, m.keys.Quit):
			if m.rangeAnchor != nil {
				m.rangeAnchor = nil
				return m, nil
			}
			m.quit = true
			if m.activeWatcher != nil {
				_ = m.activeWatcher.Close()
				m.activeWatcher = nil
			}
			return m, tea.Quit
		case key.Matches(msg, m.keys.Left):
			m.cursor = m.cursor.AddDate(0, 0, -1)
		case key.Matches(msg, m.keys.Right):
			m.cursor = m.cursor.AddDate(0, 0, 1)
		case key.Matches(msg, m.keys.PrevMonth):
			m.cursor = shiftDate(m.cursor, 0, -1)
		case key.Matches(msg, m.keys.NextMonth):
			m.cursor = shiftDate(m.cursor, 0, 1)
		case key.Matches(msg, m.keys.Today):
			m.cursor = m.today
		case key.Matches(msg, m.keys.WeekStart):
			m.cursor = weekStartDate(m.cursor, m.config.WeekStartDay)
		case key.Matches(msg, m.keys.WeekEnd):
			m.cursor = weekEndDate(m.cursor, m.config.WeekStartDay)
		case key.Matches(msg, m.keys.MonthStart):
			y, mo, _ := m.cursor.Date()
			m.cursor = time.Date(y, mo, 1, 0, 0, 0, 0, m.cursor.Location())
		case key.Matches(msg, m.keys.MonthEnd):
			y, mo, _ := m.cursor.Date()
			m.cursor = time.Date(y, mo+1, 0, 0, 0, 0, 0, m.cursor.Location())
		case key.Matches(msg, m.keys.ToggleHelp):
			m.showHelp = !m.showHelp
		}
	}
	return m, nil
}

// View produces the strip calendar view string for the model state.
func (m RowModel) View() string {
	year, month, _ := m.cursor.Date()
	loc := m.cursor.Location()
	start, end := stripWindow(year, month, m.config.WeekStartDay, loc)

	var b strings.Builder
	b.WriteString(m.renderStripDayHeaders(start, end))
	b.WriteString("\n")
	b.WriteString(m.renderStripDays(start, end))
	b.WriteString("\n")

	if m.showHelp {
		b.WriteString("\n")
		b.WriteString(m.help.View(m.keys))
		b.WriteString("\n")
	}

	return m.styles.padding.Render(b.String())
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
