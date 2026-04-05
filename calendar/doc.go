// Package calendar provides an interactive terminal calendar built with
// [Bubble Tea]. It supports grid and strip (row) views, date highlighting
// from JSON files with live-reload, range selection, multiple themes,
// week number display (US and ISO 8601), Julian day-of-year mode,
// fiscal quarter tracking, and print mode for non-interactive output.
//
// The calendar uses the [wen] parser for natural language date arguments
// but does not depend on it at runtime — it operates on [time.Time] values.
//
// [Bubble Tea]: https://github.com/charmbracelet/bubbletea
// [wen]: https://github.com/zachthieme/wen
package calendar
