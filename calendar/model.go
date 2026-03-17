package calendar

import (
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Cursor          time.Time
	Today           time.Time
	selected        bool
	quit            bool
	ShowWeekNumbers bool
	ShowHelp        bool
	Config          Config
	styles          resolvedStyles
}

type resolvedStyles struct {
	cursor    lipgloss.Style
	today     lipgloss.Style
	title     lipgloss.Style
	weekNum   lipgloss.Style
	dayHeader lipgloss.Style
	helpBar   lipgloss.Style
}

func (m Model) IsSelected() bool { return m.selected }
func (m Model) IsQuit() bool     { return m.quit }

func New(cursor, today time.Time, cfg Config) Model {
	m := Model{
		Cursor:          stripTime(cursor),
		Today:           stripTime(today),
		ShowWeekNumbers: cfg.ShowWeekNumbers,
		Config:          cfg,
	}
	m.styles = buildStyles(cfg.ResolvedColors())
	return m
}

func stripTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func (m Model) Init() tea.Cmd {
	return nil
}

// yankMsg signals that a clipboard write completed (or failed silently).
type yankMsg struct{}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case yankMsg:
		// clipboard write finished, nothing to do
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.selected = true
			return m, tea.Quit
		case tea.KeyEscape:
			m.quit = true
			return m, tea.Quit
		case tea.KeyLeft:
			m.Cursor = m.Cursor.AddDate(0, 0, -1)
		case tea.KeyRight:
			m.Cursor = m.Cursor.AddDate(0, 0, 1)
		case tea.KeyUp:
			m.Cursor = m.Cursor.AddDate(0, 0, -7)
		case tea.KeyDown:
			m.Cursor = m.Cursor.AddDate(0, 0, 7)
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "q":
				m.quit = true
				return m, tea.Quit
			case "h":
				m.Cursor = m.Cursor.AddDate(0, 0, -1)
			case "l":
				m.Cursor = m.Cursor.AddDate(0, 0, 1)
			case "k":
				m.Cursor = m.Cursor.AddDate(0, 0, -7)
			case "j":
				m.Cursor = m.Cursor.AddDate(0, 0, 7)
			case "H":
				m.Cursor = shiftDate(m.Cursor, 0, -1)
			case "L":
				m.Cursor = shiftDate(m.Cursor, 0, 1)
			case "K":
				m.Cursor = shiftDate(m.Cursor, -1, 0)
			case "J":
				m.Cursor = shiftDate(m.Cursor, 1, 0)
			case "t":
				m.Cursor = m.Today
			case "w":
				m.ShowWeekNumbers = !m.ShowWeekNumbers
			case "?":
				m.ShowHelp = !m.ShowHelp
			case "y":
				text := m.Cursor.Format("2006-01-02")
				return m, yankToClipboardCmd(text)
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	return Render(m)
}

// shiftDate moves a date by the given years and months, clamping the day
// to the last day of the target month. Consolidates nextMonth/prevMonth/
// nextYear/prevYear into a single function.
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

// yankToClipboardCmd returns a tea.Cmd that writes text to the system
// clipboard asynchronously, so the TUI doesn't block.
func yankToClipboardCmd(text string) tea.Cmd {
	return func() tea.Msg {
		yankToClipboard(text)
		return yankMsg{}
	}
}

func yankToClipboard(text string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		if path, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command(path)
		} else if path, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command(path, "-selection", "clipboard")
		} else if path, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command(path, "--clipboard", "--input")
		} else {
			return
		}
	default:
		return
	}
	cmd.Stdin = strings.NewReader(text)
	_ = cmd.Run()
}
