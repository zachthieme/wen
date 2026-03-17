package calendar

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const DateLayout = "2006-01-02"

// Model holds the state for the interactive calendar TUI.
type Model struct {
	Cursor          time.Time
	Today           time.Time
	selected        bool
	quit            bool
	ShowWeekNumbers bool
	ShowHelp        bool
	StatusMsg       string // transient status message (e.g., yank confirmation)
	Config          Config
	keys            keyMap
	help            help.Model
	styles          resolvedStyles
	clipboardCmd    []string // resolved clipboard command, nil if unavailable
}

type resolvedStyles struct {
	cursor    lipgloss.Style
	today     lipgloss.Style
	title     lipgloss.Style
	weekNum   lipgloss.Style
	dayHeader lipgloss.Style
	helpBar   lipgloss.Style
}

// IsSelected reports whether the user confirmed a date selection.
func (m Model) IsSelected() bool { return m.selected }

// IsQuit reports whether the user quit without selecting.
func (m Model) IsQuit() bool { return m.quit }

// New creates a calendar Model with the given cursor position, today's date, and configuration.
func New(cursor, today time.Time, cfg Config) Model {
	colors := cfg.ResolvedColors()
	m := Model{
		Cursor:          stripTime(cursor),
		Today:           stripTime(today),
		ShowWeekNumbers: cfg.ShowWeekNumbers,
		Config:          cfg,
		keys:            defaultKeyMap(),
		help:            newHelpModel(colors),
	}
	m.styles = buildStyles(colors)
	m.clipboardCmd = resolveClipboardCmd()
	return m
}

func stripTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func (m Model) Init() tea.Cmd {
	return nil
}

type yankMsg struct{ err error }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case yankMsg:
		if msg.err != nil {
			m.StatusMsg = fmt.Sprintf("yank failed: %v", msg.err)
		} else {
			m.StatusMsg = "yanked"
		}
	case tea.KeyMsg:
		m.StatusMsg = ""
		switch {
		case key.Matches(msg, m.keys.Select):
			m.selected = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Quit):
			m.quit = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Left):
			m.Cursor = m.Cursor.AddDate(0, 0, -1)
		case key.Matches(msg, m.keys.Right):
			m.Cursor = m.Cursor.AddDate(0, 0, 1)
		case key.Matches(msg, m.keys.Up):
			m.Cursor = m.Cursor.AddDate(0, 0, -7)
		case key.Matches(msg, m.keys.Down):
			m.Cursor = m.Cursor.AddDate(0, 0, 7)
		case key.Matches(msg, m.keys.PrevMonth):
			m.Cursor = shiftDate(m.Cursor, 0, -1)
		case key.Matches(msg, m.keys.NextMonth):
			m.Cursor = shiftDate(m.Cursor, 0, 1)
		case key.Matches(msg, m.keys.PrevYear):
			m.Cursor = shiftDate(m.Cursor, -1, 0)
		case key.Matches(msg, m.keys.NextYear):
			m.Cursor = shiftDate(m.Cursor, 1, 0)
		case key.Matches(msg, m.keys.Today):
			m.Cursor = m.Today
		case key.Matches(msg, m.keys.ToggleWeeks):
			m.ShowWeekNumbers = !m.ShowWeekNumbers
		case key.Matches(msg, m.keys.ToggleHelp):
			m.ShowHelp = !m.ShowHelp
		case key.Matches(msg, m.keys.Yank):
			if m.clipboardCmd != nil {
				text := m.Cursor.Format(DateLayout)
				cmdArgs := m.clipboardCmd
				return m, func() tea.Msg {
					cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
					cmd.Stdin = strings.NewReader(text)
					return yankMsg{err: cmd.Run()}
				}
			}
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

// resolveClipboardCmd finds the clipboard command once at startup.
func resolveClipboardCmd() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"pbcopy"}
	case "linux":
		if path, err := exec.LookPath("wl-copy"); err == nil {
			return []string{path}
		}
		if path, err := exec.LookPath("xclip"); err == nil {
			return []string{path, "-selection", "clipboard"}
		}
		if path, err := exec.LookPath("xsel"); err == nil {
			return []string{path, "--clipboard", "--input"}
		}
	}
	return nil
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
	Yank        key.Binding
	Select      key.Binding
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
		Yank: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "yank"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "quit"),
		),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.Select, k.Quit, k.ToggleHelp}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.Up, k.Down},
		{k.PrevMonth, k.NextMonth, k.PrevYear, k.NextYear},
		{k.Today, k.Yank, k.ToggleWeeks},
		{k.Select, k.Quit},
	}
}
