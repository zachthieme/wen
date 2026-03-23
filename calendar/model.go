// Package calendar provides an interactive terminal calendar UI built on Bubble Tea.
package calendar

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
)

// DateLayout is the standard date format used for output (yyyy-mm-dd).
//
// Deprecated: Use wen.DateLayout instead.
const DateLayout = "2006-01-02"

// Model holds the state for the interactive calendar TUI.
type Model struct {
	cursor           time.Time
	today            time.Time
	quit             bool
	selected         bool
	rangeAnchor      *time.Time
	weekNumPos       WeekNumPos
	showHelp         bool
	months           int
	highlightedDates map[time.Time]bool
	highlightPath    string
	activeWatcher    *fsnotify.Watcher // closed on quit to unblock watcher goroutine
	config           Config
	keys             keyMap
	help             help.Model
	styles           resolvedStyles
}

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
	padding     lipgloss.Style
}

// IsQuit reports whether the user quit without selecting.
func (m Model) IsQuit() bool { return m.quit }

// Selected reports whether the user selected a date with Enter.
func (m Model) Selected() bool { return m.selected }

// Cursor returns the currently selected date.
func (m Model) Cursor() time.Time { return m.cursor }

// InRange reports whether the user confirmed a multi-day range selection.
func (m Model) InRange() bool {
	return m.selected && m.rangeAnchor != nil && !m.rangeAnchor.Equal(m.cursor)
}

// RangeStart returns the earlier date of the confirmed range, or zero if no range.
func (m Model) RangeStart() time.Time {
	if !m.InRange() {
		return time.Time{}
	}
	if m.rangeAnchor.Before(m.cursor) {
		return *m.rangeAnchor
	}
	return m.cursor
}

// RangeEnd returns the later date of the confirmed range, or zero if no range.
func (m Model) RangeEnd() time.Time {
	if !m.InRange() {
		return time.Time{}
	}
	if m.rangeAnchor.After(m.cursor) {
		return *m.rangeAnchor
	}
	return m.cursor
}

// ModelOption configures optional Model properties.
type ModelOption func(*Model)

// WithHighlightedDates sets dates to be visually highlighted in the calendar.
// Clears any highlight source path, disabling file watching.
func WithHighlightedDates(dates map[time.Time]bool) ModelOption {
	return func(m *Model) {
		m.highlightedDates = dates
		m.highlightPath = ""
	}
}

// WithMonths sets the number of months to display side by side.
func WithMonths(n int) ModelOption {
	return func(m *Model) {
		if n < 1 {
			n = 1
		}
		m.months = n
	}
}

// New creates a calendar Model with the given cursor position, today's date, and configuration.
func New(cursor, today time.Time, cfg Config, opts ...ModelOption) Model {
	colors := cfg.ResolvedColors()
	m := Model{
		cursor:     stripTime(cursor),
		today:      stripTime(today),
		weekNumPos: parseWeekNumPos(cfg.ShowWeekNumbers),
		months:     1,
		config:     cfg,
		keys:       defaultKeyMap(),
		help:       newHelpModel(colors),
	}
	m.styles = buildStyles(colors)
	m.styles.padding = lipgloss.NewStyle().Padding(
		cfg.PaddingTop, cfg.PaddingRight, cfg.PaddingBottom, cfg.PaddingLeft,
	)
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

func stripTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// midnightTickMsg is sent when the clock crosses midnight, triggering a
// refresh of the "today" highlight.
type midnightTickMsg struct{}

// scheduleMidnightTick returns a tea.Cmd that fires at the next midnight.
func scheduleMidnightTick(now time.Time) tea.Cmd {
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return tea.Tick(time.Until(next), func(_ time.Time) tea.Msg {
		return midnightTickMsg{}
	})
}

// Init schedules the midnight tick (to refresh the "today" highlight at midnight)
// and, if a highlight source path is configured, starts an fsnotify file watcher.
func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, scheduleMidnightTick(m.today))
	if m.highlightPath != "" {
		cmds = append(cmds, startFileWatcher(m.highlightPath))
	}
	return tea.Batch(cmds...)
}

// closeWatcher closes the active fsnotify watcher if one exists, unblocking
// the watcher goroutine so it can exit cleanly.
func (m *Model) closeWatcher() {
	if m.activeWatcher != nil {
		_ = m.activeWatcher.Close()
		m.activeWatcher = nil
	}
}

// Update handles input messages and updates model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
	case midnightTickMsg:
		now := time.Now()
		m.today = stripTime(now)
		return m, scheduleMidnightTick(now)
	case highlightChangedMsg:
		m.highlightedDates = msg.dates
		m.activeWatcher = msg.watcher
		return m, waitForNextChange(msg.watcher, msg.path)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.ForceQuit):
			m.quit = true
			m.closeWatcher()
			return m, tea.Quit
		case key.Matches(msg, m.keys.VisualSelect):
			anchor := m.cursor
			m.rangeAnchor = &anchor
			return m, nil
		case key.Matches(msg, m.keys.Select):
			m.selected = true
			m.closeWatcher()
			return m, tea.Quit
		case key.Matches(msg, m.keys.Quit):
			if m.rangeAnchor != nil {
				m.rangeAnchor = nil
				return m, nil
			}
			m.quit = true
			m.closeWatcher()
			return m, tea.Quit
		case key.Matches(msg, m.keys.Left):
			m.cursor = m.cursor.AddDate(0, 0, -1)
		case key.Matches(msg, m.keys.Right):
			m.cursor = m.cursor.AddDate(0, 0, 1)
		case key.Matches(msg, m.keys.Up):
			m.cursor = m.cursor.AddDate(0, 0, -7)
		case key.Matches(msg, m.keys.Down):
			m.cursor = m.cursor.AddDate(0, 0, 7)
		case key.Matches(msg, m.keys.PrevMonth):
			m.cursor = shiftDate(m.cursor, 0, -1)
		case key.Matches(msg, m.keys.NextMonth):
			m.cursor = shiftDate(m.cursor, 0, 1)
		case key.Matches(msg, m.keys.PrevYear):
			m.cursor = shiftDate(m.cursor, -1, 0)
		case key.Matches(msg, m.keys.NextYear):
			m.cursor = shiftDate(m.cursor, 1, 0)
		case key.Matches(msg, m.keys.Today):
			m.cursor = m.today
		case key.Matches(msg, m.keys.ToggleWeeks):
			switch m.weekNumPos {
			case WeekNumOff:
				m.weekNumPos = WeekNumLeft
			case WeekNumLeft:
				m.weekNumPos = WeekNumRight
			case WeekNumRight:
				m.weekNumPos = WeekNumOff
			}
		case key.Matches(msg, m.keys.ToggleHelp):
			m.showHelp = !m.showHelp
		}
	}
	return m, nil
}

func shiftDate(t time.Time, years, months int) time.Time {
	y, m, d := t.Date()
	target := time.Date(y+years, m+time.Month(months), 1, 0, 0, 0, 0, t.Location())
	maxDay := daysInMonth(target.Year(), target.Month(), t.Location())
	if d > maxDay {
		d = maxDay
	}
	return time.Date(target.Year(), target.Month(), d, 0, 0, 0, 0, t.Location())
}

func daysInMonth(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}

type keyMap struct {
	Left         key.Binding
	Right        key.Binding
	Up           key.Binding
	Down         key.Binding
	PrevMonth    key.Binding
	NextMonth    key.Binding
	PrevYear     key.Binding
	NextYear     key.Binding
	Today        key.Binding
	ToggleWeeks  key.Binding
	ToggleHelp   key.Binding
	VisualSelect key.Binding
	Select       key.Binding
	Quit         key.Binding
	ForceQuit    key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/←", "prev day"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/→", "next day"),
		),
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "prev week"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "next week"),
		),
		PrevMonth: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "prev month"),
		),
		NextMonth: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "next month"),
		),
		PrevYear: key.NewBinding(
			key.WithKeys("K"),
			key.WithHelp("K", "prev year"),
		),
		NextYear: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("J", "next year"),
		),
		Today: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "today"),
		),
		ToggleWeeks: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "weeks"),
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

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.VisualSelect, k.Select, k.Quit, k.ToggleHelp}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.Up, k.Down},
		{k.PrevMonth, k.NextMonth, k.PrevYear, k.NextYear},
		{k.Today, k.ToggleWeeks},
		{k.VisualSelect, k.Select, k.Quit},
	}
}
