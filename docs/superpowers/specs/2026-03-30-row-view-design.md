# `wen row` — Strip Calendar View

## Overview

A new `wen row` subcommand that renders an interactive, two-row calendar strip in the terminal using Bubble Tea.

```
Su Mo Tu We Th Fr Sa Su Mo Tu We Th Fr Sa Su Mo Tu We Th Fr Sa Su Mo Tu We Th Fr Sa Su Mo Tu
Mr  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31
```

- **Row 1:** Repeating 2-char day-of-week headers (Su, Mo, Tu, We, Th, Fr, Sa)
- **Row 2:** 2-char month abbreviation + day numbers

## Layout

### Week-Aligned Month

The strip always starts on the configured `WeekStartDay` on or before the 1st of the month, and ends on the corresponding week-end-day on or after the last day. This means:

- Width varies from 28–37 columns depending on month length and start-day alignment
- Padding days from the previous/next month are shown with dimmed styling
- Clean week boundaries are always visible

### Month Abbreviations

2-char abbreviations using first two letters of each month, ensuring uniqueness:

| Month | Abbr |
|-------|------|
| January | Ja |
| February | Fe |
| March | Mr |
| April | Ap |
| May | My |
| June | Jn |
| July | Jl |
| August | Au |
| September | Se |
| October | Oc |
| November | No |
| December | De |

The abbreviation is prepended to row 2 before the first day number, occupying 3 characters (2 letters + 1 space). It replaces a traditional title row. Row 1 (day headers) has 3 characters of leading space to align with the day numbers.

## Navigation

| Key | Action |
|-----|--------|
| `h` / `←` | Previous day |
| `l` / `→` | Next day |
| `b` | Beginning of current week; if already there, previous week |
| `e` | End of current week; if already there, next week |
| `0` | First day of month |
| `$` | Last day of month |
| `k` / `↑` | Previous month (same day number, clamped) |
| `j` / `↓` | Next month (same day number, clamped) |
| `t` | Jump to today |
| `v` | Start/anchor range selection |
| `enter` | Select date or confirm range |
| `q` / `esc` | Quit (or cancel active range) |
| `ctrl+c` | Force quit |
| `?` | Toggle help bar |

### Cursor on Padding Days

The cursor can move onto padding days from adjacent months. Moving onto a padding day transitions the view to that month. For example:

- Cursor on March 31, press `l` → cursor lands on April 1, strip re-renders for April
- Cursor on a leading padding day (e.g., Feb 28 before March 1), press `h` → strip re-renders for February with cursor on Feb 28

This provides seamless day-by-day month navigation via `h`/`l`, while `j`/`k` jump by whole months.

## Features

Full parity with the grid calendar view:

- **Cursor styling** — highlighted cell for current cursor position
- **Today highlight** — distinct style for today's date
- **Highlighted dates** — from JSON file (`--highlight-file` or config `highlightSource`)
- **Range selection** — `v` to anchor, move cursor, `enter` to confirm; styled range days
- **Themes** — all existing themes (default, catppuccin-mocha, dracula, nord) and custom color overrides apply
- **Midnight tick** — today highlight updates at midnight
- **File watcher** — highlight file changes are picked up live

### Not Included

- No fiscal quarter display
- No week numbers
- No multi-month view (single month only; navigate with `j`/`k`)

## Architecture

### New Files

- `calendar/row_model.go` — `RowModel` struct implementing `tea.Model` with `Init()`, `Update()`, `View()`
- `calendar/row_render.go` — strip rendering: `renderStripDayHeaders()`, `renderStripDays()`
- `calendar/row_model_test.go` — model tests (key handling, cursor movement, month transitions)
- `calendar/row_render_test.go` — render tests (output correctness, padding, styling)
- `cmd/wen/row.go` — CLI subcommand wiring, flag parsing, `runRow()` function

### Shared Logic (no refactoring needed)

These existing free functions and types are reused directly:

- `shiftDate(t, years, months)` — cursor month/year clamping
- `scheduleMidnightTick()` / `midnightTickMsg` — midnight refresh
- `startFileWatcher()` / `highlightChangedMsg` / `watcherErrMsg` — file watching
- `resolvedStyles` / `buildStyles()` — theme styling
- `Config` / `LoadConfig()` / `Normalize()` — configuration
- `dateKey()` — UTC date normalization for map lookups
- `isInRange()` — range membership check
- `dayNames` — day-of-week abbreviation array

### RowModel Struct

```go
type RowModel struct {
    cursor           time.Time
    today            time.Time
    quit             bool
    selected         bool
    rangeAnchor      *time.Time
    highlightedDates map[time.Time]bool
    highlightPath    string
    activeWatcher    *fsnotify.Watcher
    config           Config
    keys             rowKeyMap
    help             help.Model
    styles           resolvedStyles
    showHelp         bool
}
```

Fields mirror `Model` minus grid-specific state (`weekNumPos`, `months`). The `rowKeyMap` differs from `keyMap` — no week toggle, no year navigation, adds `b`/`e`/`0`/`$`.

### RowModel.Update()

Handles the same message types as `Model.Update()`:

- `tea.WindowSizeMsg` — update help width
- `watcherErrMsg` — silent degradation
- `midnightTickMsg` — refresh today
- `highlightChangedMsg` — update highlighted dates
- `tea.KeyMsg` — dispatch to row-specific key bindings

Key handling for `h`/`l` crossing month boundaries: when cursor moves before day 1 or after last day of the current month, the view transitions to the adjacent month automatically (the cursor date just changes; `View()` renders whatever month the cursor is in).

### RowModel.View()

Calls `renderStripDayHeaders()` and `renderStripDays()`, appends optional help bar, applies padding.

### Rendering

`renderStripDayHeaders(startDate, endDate, config)` — produces row 1 by repeating day-of-week abbreviations for each day in the window.

`renderStripDays(startDate, endDate, cursor, today, month, highlights, rangeAnchor, styles)` — produces row 2: month abbreviation in first cell, then each day number with appropriate styling (cursor, today, highlight, range, dimmed for padding days).

The rendering functions compute the week-aligned window from the cursor's month:
1. Find the 1st of the cursor's month
2. Walk back to the nearest `WeekStartDay` on or before the 1st
3. Find the last day of the cursor's month
4. Walk forward to the nearest week-end-day on or after the last

## CLI Integration

### Subcommand

`wen row` — registered alongside `cal`, `diff`, `relative` in `cmd/wen/main.go`.

### Flags

- `--highlight-file` — path to JSON highlighted dates file
- `--padding-top`, `--padding-right`, `--padding-bottom`, `--padding-left` — display padding

No `--months` flag. Config file settings for `theme`, `colors`, `weekStartDay`, and `highlightSource` all apply.

### Output

On selection (`enter`), prints the selected date (or date range) to stdout in the configured `--format`, matching `wen cal` behavior.
