package calendar

import (
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const DateLayout = "2006-01-02"

type Model struct {
	Cursor          time.Time
	Today           time.Time
	selected        bool
	quit            bool
	ShowWeekNumbers bool
	ShowHelp        bool
	Config          Config
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
	m.clipboardCmd = resolveClipboardCmd()
	return m
}

func stripTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func (m Model) Init() tea.Cmd {
	return nil
}

type yankMsg struct{}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case yankMsg:
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
				if m.clipboardCmd != nil {
					text := m.Cursor.Format(DateLayout)
					cmdArgs := m.clipboardCmd
					return m, func() tea.Msg {
						cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
						cmd.Stdin = strings.NewReader(text)
						_ = cmd.Run()
						return yankMsg{}
					}
				}
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	return Render(m)
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
