package calendar

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

func applyColor(s lipgloss.Style, color string) lipgloss.Style {
	if color != "" {
		return s.Foreground(lipgloss.Color(color))
	}
	return s
}

func buildStyles(colors ThemeColors) resolvedStyles {
	cursorStyle := applyColor(lipgloss.NewStyle().Reverse(true), colors.Cursor)
	todayStyle := applyColor(lipgloss.NewStyle().Bold(true).Underline(true), colors.Today)
	// Pre-compose cursor+today so View() avoids nested Render calls and
	// the double-reset ANSI sequences they produce.
	cursorTodayStyle := lipgloss.NewStyle().Reverse(true).Bold(true).Underline(true)
	if colors.Cursor != "" {
		cursorTodayStyle = cursorTodayStyle.Foreground(lipgloss.Color(colors.Cursor))
	}
	highlightStyle := lipgloss.NewStyle().Bold(true)
	if colors.Highlight != "" {
		highlightStyle = highlightStyle.Foreground(lipgloss.Color(colors.Highlight))
	} else {
		highlightStyle = highlightStyle.Underline(true)
	}
	// rangeDay uses Reverse as a default (visible without color).
	// When a Range color is set, Foreground replaces Reverse entirely
	// for a cleaner look — intentionally different from other styles
	// which layer color on top of their base treatment.
	rangeDayStyle := lipgloss.NewStyle().Reverse(true)
	if colors.Range != "" {
		rangeDayStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Range))
	}
	return resolvedStyles{
		cursor:      cursorStyle,
		cursorToday: cursorTodayStyle,
		today:       todayStyle,
		highlight:   highlightStyle,
		rangeDay:    rangeDayStyle,
		title:       applyColor(lipgloss.NewStyle().Bold(true), colors.Title),
		weekNum:     applyColor(lipgloss.NewStyle().Faint(true), colors.WeekNumber),
		dayHeader:   applyColor(lipgloss.NewStyle().Faint(true), colors.DayHeader),
		helpBar:     applyColor(lipgloss.NewStyle().Faint(true), colors.HelpBar),
	}
}

func newHelpModel(colors ThemeColors) help.Model {
	h := help.New()
	h.ShowAll = true

	helpStyle := applyColor(lipgloss.NewStyle().Faint(true), colors.HelpBar)

	h.Styles.ShortKey = helpStyle
	h.Styles.ShortDesc = helpStyle
	h.Styles.ShortSeparator = helpStyle
	h.Styles.FullKey = helpStyle
	h.Styles.FullDesc = helpStyle
	h.Styles.FullSeparator = helpStyle

	return h
}
