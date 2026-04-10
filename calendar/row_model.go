package calendar

import (
	"strings"
	"time"

	"github.com/zachthieme/wen"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RowModel holds the state for the interactive strip calendar TUI.
type RowModel struct {
	baseModel
	keys rowKeyMap
}

// NewRow creates a RowModel with the given cursor position, today's date, and configuration.
func NewRow(cursor, today time.Time, cfg Config, opts ...Option) RowModel {
	colors := cfg.ResolvedColors()
	m := RowModel{
		baseModel: baseModel{
			cursor: wen.TruncateDay(cursor),
			today:  wen.TruncateDay(today),
			config: cfg,
			help:   newHelpModel(colors),
			styles: buildStyles(colors),
			months: 1,
		},
		keys: defaultRowKeyMap(),
	}
	// Strip Underline from row styles. lipgloss renders Underline per-character
	// (each char gets its own ANSI open/close), which causes terminals like mosh
	// to miscalculate cursor positions and misalign the strip columns.
	m.styles.today = m.styles.today.Underline(false)
	m.styles.cursorToday = m.styles.cursorToday.Underline(false)
	m.styles.highlight = m.styles.highlight.Underline(false)
	for _, opt := range opts {
		opt(&m.baseModel)
	}
	m.dayFmt = dayFormatFor(m.julian)
	return m
}

// Init schedules the periodic date check and, if a highlight source path is configured,
// starts an fsnotify file watcher.
func (m RowModel) Init() tea.Cmd {
	return tea.Batch(m.initCmds()...)
}

// Update handles input messages and updates model state.
func (m RowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case key.Matches(msg, m.keys.PrevMonth):
			m.cursor = shiftDate(m.cursor, 0, -1)
		case key.Matches(msg, m.keys.NextMonth):
			m.cursor = shiftDate(m.cursor, 0, 1)
		case key.Matches(msg, m.keys.Today):
			m.cursor = m.today
		case key.Matches(msg, m.keys.ToggleJulian):
			m.julian = !m.julian
			m.dayFmt = dayFormatFor(m.julian)
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

// visibleWindow trims the full strip window to fit within the terminal width,
// centering on the cursor. If the full window fits, it is returned unchanged.
func (m RowModel) visibleWindow(fullStart, fullEnd time.Time) (time.Time, time.Time) {
	availWidth := m.termWidth
	if availWidth <= 0 {
		return fullStart, fullEnd
	}

	totalDays := dayCount(fullStart, fullEnd)
	// Each day cell is cellWidth+1 chars (number + space separator).
	// The prefix occupies prefixWidth chars before the first cell.
	// Total for N days: prefixWidth + N*(cellWidth+1) - 1 (no trailing space).
	// Solving for N: (availWidth - prefixWidth + 1) / (cellWidth + 1)
	cellW := m.dayFmt.cellWidth + 1
	maxDays := (availWidth - m.dayFmt.prefixWidth + 1) / cellW
	if maxDays <= 0 {
		maxDays = 1
	}
	if totalDays <= maxDays {
		return fullStart, fullEnd
	}

	cursorOffset := dayCount(fullStart, m.cursor) - 1 // 0-indexed
	startOffset := max(cursorOffset-maxDays/2, 0)
	if startOffset+maxDays > totalDays {
		startOffset = totalDays - maxDays
	}

	return fullStart.AddDate(0, 0, startOffset), fullStart.AddDate(0, 0, startOffset+maxDays-1)
}

// View produces the strip calendar view string for the model state.
func (m RowModel) View() string {
	year, month, _ := m.cursor.Date()
	loc := m.cursor.Location()
	fullStart, fullEnd := stripWindow(year, month, m.config.WeekStartDay, loc)
	start, end := m.visibleWindow(fullStart, fullEnd)

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

	output := b.String()
	if m.termWidth > 0 && m.termHeight > 0 {
		return lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Left, lipgloss.Center, output)
	}
	return output
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
	ToggleJulian key.Binding
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

// ShortHelp returns bindings for the short help view.
func (k rowKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.VisualSelect, k.Select, k.Quit, k.ToggleHelp}
}

// FullHelp returns bindings for the full help view.
func (k rowKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.PrevMonth, k.NextMonth},
		{k.WeekStart, k.WeekEnd, k.MonthStart, k.MonthEnd},
		{k.Today, k.ToggleJulian, k.VisualSelect, k.Select, k.Quit},
	}
}
