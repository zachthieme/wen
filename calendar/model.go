package calendar

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Cursor   time.Time
	Today    time.Time
	Selected bool
	Quit     bool
}

func New(cursor, today time.Time) Model {
	return Model{
		Cursor: stripTime(cursor),
		Today:  stripTime(today),
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
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	return ""
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

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
}
