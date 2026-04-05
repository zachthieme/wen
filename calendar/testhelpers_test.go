package calendar

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// date creates a time.Time at midnight in the local timezone.
func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

// runeMsg builds a tea.KeyMsg for a printable key string.
func runeMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

// specialMsg builds a tea.KeyMsg for a special key type (arrows, enter, etc.).
func specialMsg(key tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: key}
}

// press sends a rune key to a Bubble Tea model and returns the updated model.
// Works with any concrete model type (Model, RowModel) via generics.
func press[M tea.Model](m M, key string) M {
	updated, _ := m.Update(runeMsg(key))
	return any(updated).(M)
}
