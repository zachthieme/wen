# dayFormat Refactor — Consolidate Julian Rendering Conditionals

## Overview

Extract a `dayFormat` struct that captures the rendering dimensions derived from `julian` mode. Replace 9 scattered `if m.julian` conditionals across `render.go`, `row_render.go`, and `row_model.go` with field accesses on a single pre-computed struct.

This is a pure refactor — no behavioral changes.

## Problem

The julian feature added `if m.julian` checks in 9 locations across 3 files. Each check independently derives the same underlying facts (cell width, day name array, grid width, prefix width, day formatting). This duplication:

- Makes it easy to miss a site when adding a new rendering dimension
- Causes bugs like the `visibleWindow` overflow (formula not updated for julian cell width)
- Requires every rendering function to know about the `julian` bool

## Solution: `dayFormat` struct

### Struct definition

In `calendar/render.go`:

```go
type dayFormat struct {
    cellWidth   int
    gridWidth   int
    prefixWidth int
    names       [7]string
    formatDay   func(year int, month time.Month, day int, loc *time.Location) string
}
```

Fields:
- `cellWidth` — character width of a single day number (2 for normal, 3 for julian)
- `gridWidth` — character width of the full 7-column day grid (20 for normal, 27 for julian)
- `prefixWidth` — total width of the strip's leading prefix column including separators (3 for normal, 4 for julian)
- `names` — day-of-week abbreviation array (2-char or 3-char)
- `formatDay` — formats a day number string given its date components

### Constructors

```go
func normalDayFormat() dayFormat {
    return dayFormat{
        cellWidth:   2,
        gridWidth:   20,
        prefixWidth: 3,
        names:       dayNames,
        formatDay: func(_ int, _ time.Month, day int, _ *time.Location) string {
            return fmt.Sprintf("%2d", day)
        },
    }
}

func julianDayFormat() dayFormat {
    return dayFormat{
        cellWidth:   3,
        gridWidth:   27,
        prefixWidth: 4,
        names:       dayNamesLong,
        formatDay: func(year int, month time.Month, day int, loc *time.Location) string {
            yd := time.Date(year, month, day, 0, 0, 0, 0, loc).YearDay()
            return fmt.Sprintf("%3d", yd)
        },
    }
}

func dayFormatFor(julian bool) dayFormat {
    if julian {
        return julianDayFormat()
    }
    return normalDayFormat()
}
```

### Model integration

Add `dayFmt dayFormat` field to both `Model` and `RowModel`.

Set in constructors:
- `New()` — after applying options: `m.dayFmt = dayFormatFor(m.julian)`
- `NewRow()` — same pattern

Update on `J` toggle in `Update()`:
- After `m.julian = !m.julian`, add `m.dayFmt = dayFormatFor(m.julian)`

### Replacements

| File | Before | After |
|---|---|---|
| `render.go` `gridWidth()` method | `if m.julian { return julianGridWidth }; return dayGridWidth` | Remove method. Use `m.dayFmt.gridWidth` |
| `render.go` `renderDayHeaders` | `if m.julian { dayNamesLong[idx] } else { dayNames[idx] }` | `m.dayFmt.names[idx]` |
| `render.go` `renderGrid` cellWidth | `cellWidth := 2; if m.julian { cellWidth = 3 }` | `m.dayFmt.cellWidth` |
| `render.go` `renderGrid` day format | `if m.julian { fmt.Sprintf("%3d", yd) } else { fmt.Sprintf("%2d", day) }` | `m.dayFmt.formatDay(year, month, day, loc)` |
| `render.go` `renderTitle` | `m.gridWidth()` | `m.dayFmt.gridWidth` |
| `row_render.go` `renderStripDayHeaders` prefix | `if m.julian { "    " } else { "   " }` | `strings.Repeat(" ", m.dayFmt.prefixWidth)` |
| `row_render.go` `renderStripDayHeaders` names | `if m.julian { dayNamesLong[...] } else { dayNames[...] }` | `m.dayFmt.names[d.Weekday()]` |
| `row_render.go` `renderStripDays` prefix | `if m.julian { "  " } else { " " }` | `strings.Repeat(" ", m.dayFmt.prefixWidth-2)` |
| `row_render.go` `renderStripDays` day format | `if m.julian { fmt.Sprintf("%3d", d.YearDay()) } else { fmt.Sprintf("%2d", d.Day()) }` | `m.dayFmt.formatDay(d.Year(), d.Month(), d.Day(), loc)` |
| `row_model.go` `visibleWindow` | `cellW := 3; prefixW := 2; if m.julian { ... }` | Derive `cellW` and overhead from `m.dayFmt.cellWidth` and `m.dayFmt.prefixWidth` |
| `view.go` `renderSingleMonth` | `m.gridWidth()` | `m.dayFmt.gridWidth` |
| `view.go` `renderMultiMonth` | `m.gridWidth()` | `m.dayFmt.gridWidth` |

### Constants removed

- `dayGridWidth` constant — value 20 moves into `normalDayFormat()`
- `julianGridWidth` constant — value 27 moves into `julianDayFormat()`
- `gridWidth()` method — replaced by `m.dayFmt.gridWidth`

### What stays unchanged

- `renderTitle` logic (fiscal quarter, year display) — only the width source changes
- `renderQuarterBar`, `weekNumber`, `quarterStartDate`, `countQuarterWorkdaysLeft` — no julian dependency
- `dateKey`, `isInRange` — pure utility functions
- `stripWindow`, `weekStartDate`, `weekEndDate`, `dayCount` — not julian-dependent
- Style-application switch (cursor/today/highlight/range) — unrelated
- `printMode` cursor suppression — unrelated
- `m.julian` bool — stays for config/CLI plumbing; `dayFmt` is the derived rendering state
- `WithJulian` / `WithRowJulian` options — still set `m.julian`; `dayFmt` is computed after options apply
- `dayNames` and `dayNamesLong` arrays — still defined, referenced by constructors

## Testing

### New tests

- `TestNormalDayFormat` — verify constructor returns cellWidth=2, gridWidth=20, prefixWidth=3, names=dayNames, formatDay produces `%2d`
- `TestJulianDayFormat` — verify constructor returns cellWidth=3, gridWidth=27, prefixWidth=4, names=dayNamesLong, formatDay produces YearDay
- `TestDayFormatFor` — verify `dayFormatFor(false)` and `dayFormatFor(true)` return the right variants

### Updated tests

- `TestToggleJulian` — verify `m.dayFmt.gridWidth` changes when J is toggled
- `TestRowToggleJulian` — same for RowModel
- `TestGridWidth` — remove (method no longer exists) or adapt to test `m.dayFmt.gridWidth` directly

### Unchanged tests

All existing rendering tests (`TestRenderGrid*`, `TestRenderStripDays*`, `TestRenderJulian*`, etc.) are behavior-preserving and should pass without modification.

## Scope boundaries

**In scope:**
- `dayFormat` struct + constructors in `calendar/render.go`
- `dayFmt` field on `Model` and `RowModel`
- Replace all julian conditionals in render/row_render/row_model with `dayFmt` access
- Remove `gridWidth()`, `dayGridWidth`, `julianGridWidth`
- Tests for the new struct and updated toggle tests

**Out of scope:**
- Any behavioral changes
- Config, CLI, help text changes
- Style system changes
- Any new features
