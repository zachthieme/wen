// Package calendar provides an interactive terminal calendar UI built on Bubble Tea.
package calendar

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DateLayout is the standard date format used for output (yyyy-mm-dd).
//
// Deprecated: Use wen.DateLayout instead.
const DateLayout = "2006-01-02"

// Model holds the state for the interactive calendar TUI.
type Model struct {
	cursor          time.Time
	today           time.Time
	quit            bool
	showWeekNumbers bool
	showHelp        bool
	config          Config
	keys            keyMap
	help            help.Model
	styles          resolvedStyles
}

type resolvedStyles struct {
	cursor      lipgloss.Style
	cursorToday lipgloss.Style
	today       lipgloss.Style
	title       lipgloss.Style
	weekNum     lipgloss.Style
	dayHeader   lipgloss.Style
	helpBar     lipgloss.Style
	padding     lipgloss.Style
	hasPadding  bool
}

// IsQuit reports whether the user quit without selecting.
func (m Model) IsQuit() bool { return m.quit }

// Cursor returns the currently selected date.
func (m Model) Cursor() time.Time { return m.cursor }

// New creates a calendar Model with the given cursor position, today's date, and configuration.
func New(cursor, today time.Time, cfg Config) Model {
	colors := cfg.ResolvedColors()
	m := Model{
		cursor:          stripTime(cursor),
		today:           stripTime(today),
		showWeekNumbers: cfg.ShowWeekNumbers,
		config:          cfg,
		keys:            defaultKeyMap(),
		help:            newHelpModel(colors),
	}
	m.styles = buildStyles(colors)
	if cfg.PaddingTop > 0 || cfg.PaddingRight > 0 || cfg.PaddingBottom > 0 || cfg.PaddingLeft > 0 {
		m.styles.hasPadding = true
		m.styles.padding = lipgloss.NewStyle().Padding(
			cfg.PaddingTop, cfg.PaddingRight, cfg.PaddingBottom, cfg.PaddingLeft,
		)
	}
	return m
}

func stripTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// Init satisfies the tea.Model interface.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles input messages and updates model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quit = true
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
			m.showWeekNumbers = !m.showWeekNumbers
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
	Left        key.Binding
	Right       key.Binding
	Up          key.Binding
	Down        key.Binding
	PrevMonth   key.Binding
	NextMonth   key.Binding
	PrevYear    key.Binding
	NextYear    key.Binding
	Today       key.Binding
	ToggleWeeks key.Binding
	ToggleHelp  key.Binding
	Quit        key.Binding
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
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q/esc", "quit"),
		),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.Quit, k.ToggleHelp}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.Up, k.Down},
		{k.PrevMonth, k.NextMonth, k.PrevYear, k.NextYear},
		{k.Today, k.ToggleWeeks},
		{k.Quit},
	}
}
