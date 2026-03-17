package calendar

import (
	"os/exec"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Cursor          time.Time
	Today           time.Time
	Selected        bool
	Quit            bool
	ShowWeekNumbers bool
	ShowHelp        bool
	Config          Config
}

func New(cursor, today time.Time, cfg Config) Model {
	return Model{
		Cursor:          stripTime(cursor),
		Today:           stripTime(today),
		ShowWeekNumbers: cfg.ShowWeekNumbers,
		Config:          cfg,
	}
}

func stripTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.Selected = true
			return m, tea.Quit
		case tea.KeyEscape:
			m.Quit = true
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
				m.Quit = true
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
				m.Cursor = prevMonth(m.Cursor)
			case "L":
				m.Cursor = nextMonth(m.Cursor)
			case "K":
				m.Cursor = prevYear(m.Cursor)
			case "J":
				m.Cursor = nextYear(m.Cursor)
			case "t":
				m.Cursor = m.Today
			case "w":
				m.ShowWeekNumbers = !m.ShowWeekNumbers
			case "?":
				m.ShowHelp = !m.ShowHelp
			case "y":
				yankToClipboard(m.Cursor.Format("2006-01-02"))
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	return Render(m)
}

func nextMonth(t time.Time) time.Time {
	y, m, d := t.Date()
	next := time.Date(y, m+1, 1, 0, 0, 0, 0, t.Location())
	maxDay := daysInMonth(next.Year(), next.Month())
	if d > maxDay {
		d = maxDay
	}
	return time.Date(next.Year(), next.Month(), d, 0, 0, 0, 0, t.Location())
}

func prevMonth(t time.Time) time.Time {
	y, m, d := t.Date()
	prev := time.Date(y, m-1, 1, 0, 0, 0, 0, t.Location())
	maxDay := daysInMonth(prev.Year(), prev.Month())
	if d > maxDay {
		d = maxDay
	}
	return time.Date(prev.Year(), prev.Month(), d, 0, 0, 0, 0, t.Location())
}

func nextYear(t time.Time) time.Time {
	y, m, d := t.Date()
	maxDay := daysInMonth(y+1, m)
	if d > maxDay {
		d = maxDay
	}
	return time.Date(y+1, m, d, 0, 0, 0, 0, t.Location())
}

func prevYear(t time.Time) time.Time {
	y, m, d := t.Date()
	maxDay := daysInMonth(y-1, m)
	if d > maxDay {
		d = maxDay
	}
	return time.Date(y-1, m, d, 0, 0, 0, 0, t.Location())
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
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
	pipe, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	if err := cmd.Start(); err != nil {
		return
	}
	_, _ = pipe.Write([]byte(text))
	pipe.Close()
	_ = cmd.Wait()
}
