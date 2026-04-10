// Package calendar provides an interactive terminal calendar UI built on Bubble Tea.
package calendar

import (
	"time"

	"github.com/zachthieme/wen"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Model holds the state for the interactive calendar TUI.
type Model struct {
	baseModel
	weekNumPos WeekNumPos
	keys       keyMap
}

// New creates a calendar Model with the given cursor position, today's date, and configuration.
func New(cursor, today time.Time, cfg Config, opts ...Option) Model {
	colors := cfg.ResolvedColors()
	m := Model{
		baseModel: baseModel{
			cursor: wen.TruncateDay(cursor),
			today:  wen.TruncateDay(today),
			config: cfg,
			help:   newHelpModel(colors),
			styles: buildStyles(colors),
			months: 1,
		},
		weekNumPos: parseWeekNumPos(cfg.ShowWeekNumbers),
		keys:       defaultKeyMap(),
	}
	for _, opt := range opts {
		opt(&m.baseModel)
	}
	m.dayFmt = dayFormatFor(m.julian)
	return m
}

// dateCheckInterval is the period between date-change checks. Go timers use
// the monotonic clock, which pauses during system suspend. A short interval
// ensures the "today" highlight recovers quickly after resume instead of
// waiting for a single long-duration midnight timer to expire.
const dateCheckInterval = time.Minute

// dateCheckMsg is sent periodically so the model can refresh its "today" value.
type dateCheckMsg struct{}

// scheduleDateCheck returns a tea.Cmd that fires after dateCheckInterval.
func scheduleDateCheck() tea.Cmd {
	return tea.Tick(dateCheckInterval, func(_ time.Time) tea.Msg {
		return dateCheckMsg{}
	})
}

// Init schedules the periodic date check (to refresh the "today" highlight)
// and, if a highlight source path is configured, starts an fsnotify file watcher.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.initCmds()...)
}

// Update handles input messages and updates model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, handled := m.handleMsg(msg); handled {
		return m, cmd
	}
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, m.keys.ForceQuit):
			cmd := m.doQuit()
			return m, cmd
		case key.Matches(msg, m.keys.VisualSelect):
			m.doVisualSelect()
			return m, nil
		case key.Matches(msg, m.keys.Select):
			cmd := m.doSelect()
			return m, cmd
		case key.Matches(msg, m.keys.Quit):
			if m.cancelRange() {
				return m, nil
			}
			cmd := m.doQuit()
			return m, cmd
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
		case key.Matches(msg, m.keys.ToggleJulian):
			m.julian = !m.julian
			m.dayFmt = dayFormatFor(m.julian)
		case key.Matches(msg, m.keys.ToggleHelp):
			m.showHelp = !m.showHelp
		}
	}
	return m, nil
}

func shiftDate(t time.Time, years, months int) time.Time {
	y, m, d := t.Date()
	target := time.Date(y+years, m+time.Month(months), 1, 0, 0, 0, 0, t.Location())
	maxDay := wen.DaysIn(target.Year(), target.Month(), t.Location())
	if d > maxDay {
		d = maxDay
	}
	return time.Date(target.Year(), target.Month(), d, 0, 0, 0, 0, t.Location())
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
	ToggleJulian key.Binding
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
			key.WithKeys("P"),
			key.WithHelp("P", "prev year"),
		),
		NextYear: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "next year"),
		),
		Today: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "today"),
		),
		ToggleWeeks: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "weeks"),
		),
		ToggleJulian: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("J", "julian"),
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
		{k.Today, k.ToggleWeeks, k.ToggleJulian},
		{k.VisualSelect, k.Select, k.Quit},
	}
}
