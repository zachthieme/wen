# Date Range Selection in Calendar Mode

## Summary

Add vim-style visual range selection to the interactive calendar TUI. Press `v` to anchor a start date, navigate to extend the range, press `Enter` to confirm. Output is two lines (start and end date), composable with tools like `git log --since=... --until=...`.

## Motivation

Single-date selection covers most use cases, but date ranges are common in scripting: filtering logs, generating reports, scheduling. Currently users must run `wen cal` twice or type dates manually. Range selection makes this a single interactive action.

## Design

### Model State

Add one field to `calendar.Model`:

```go
rangeAnchor *time.Time // nil = normal mode, non-nil = range mode
```

The existing `cursor`, `selected`, and `quit` fields are unchanged.

When a range is confirmed (Enter in range mode), `selected` is set to `true` AND `rangeAnchor` remains non-nil. This distinguishes three exit states:

| State | `selected` | `rangeAnchor` | `quit` |
|-------|-----------|--------------|--------|
| Quit without selecting | false | nil | true |
| Single date selected | true | nil | false |
| Range selected | true | non-nil | false |

### Keybindings

| Key | Normal Mode | Range Mode |
|-----|------------|------------|
| `v` | Set anchor at cursor, enter range mode | Move anchor to cursor (restart range) |
| `Enter` | Set `selected = true`, quit | Set `selected = true`, quit (anchor stays) |
| `Esc`/`q` | Set `quit = true`, quit | Clear anchor, return to normal mode (don't quit) |
| `ctrl+c` | Set `quit = true`, quit | Set `quit = true`, quit (hard interrupt, always exits) |
| Navigation (hjkl etc.) | Move cursor | Move cursor (anchor stays fixed) |

`ctrl+c` must be split into its own key binding separate from `Esc`/`q` so it always exits immediately regardless of mode. The existing `Quit` binding is split into `Quit` (`q`, `esc`) and `ForceQuit` (`ctrl+c`).

When `Esc`/`q` is pressed in range mode, it clears the anchor and returns to normal mode rather than quitting. A second `Esc`/`q` in normal mode quits as usual.

### Same-Day Range

If the user presses `v` then immediately `Enter` (anchor == cursor), this is treated as a **single-date selection**, not a range. `InRange()` returns false because anchor equals cursor — there is no range. This avoids outputting two identical lines and matches the user's intent (they didn't move to define a range).

### Rendering

Days between anchor and cursor (inclusive) receive the `rangeDay` style. Style priority from highest to lowest:

1. `cursor` / `cursorToday` — the cell under the cursor
2. `today` — today's date (when not under cursor)
3. `rangeDay` — dates in the selected range (excluding cursor)
4. `highlighted` — externally highlighted dates
5. Plain — default rendering

Range highlighting works across month boundaries. In the `renderGrid` function, the range check is performed against the full date (`time.Date(year, month, day, ...)`) — not just the day number — so it works correctly for any visible month in both single and multi-month views.

#### Multi-Month View

`renderMultiMonth` calls `renderGrid` for each visible month. `renderGrid` already receives `year`, `month`, and has access to `m.rangeAnchor` and `m.cursor` via the model receiver. The range membership check (`isInRange(date, anchor, cursor)`) operates on absolute dates, so it automatically highlights the correct days in every visible month without any special multi-month logic.

### Styles

Add `rangeDay` field to `resolvedStyles`:

```go
type resolvedStyles struct {
    // ... existing fields ...
    rangeDay lipgloss.Style
}
```

In `buildStyles`, construct the range style following the same pattern as `highlight`:

```go
rangeDayStyle := lipgloss.NewStyle().Reverse(true)
if colors.Range != "" {
    rangeDayStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Range))
} else {
    rangeDayStyle = lipgloss.NewStyle().Reverse(true)
}
```

When `Range` color is empty (the `default` theme), fall back to `Reverse(true)` for a visible but color-agnostic highlight. This matches the pattern where `highlight` falls back to `Bold+Underline` when no color is set.

### Theme Integration

Add `Range` field to `ThemeColors`:

```go
Range string `yaml:"range"`
```

Add to `ResolvedColors()`:

```go
Range: mergeColor(base.Range, c.Colors.Range),
```

Default colors per theme:

| Theme | Range Color |
|-------|------------|
| default | (empty — fallback to Reverse) |
| catppuccin-mocha | `#a6e3a1` |
| dracula | `#50fa7b` |
| nord | `#a3be8c` |

Config override: `colors.range: "#hexcolor"`.

### Public API

New methods on `Model`:

```go
// InRange reports whether the user confirmed a date range selection.
// Returns false if: selected is false (user quit), anchor is nil (no range started),
// or anchor equals cursor (same-day "range" is a single selection).
// Implementation: m.selected && m.rangeAnchor != nil && !m.rangeAnchor.Equal(m.cursor)
func (m Model) InRange() bool

// RangeStart returns the earlier date of the range.
// Returns the zero time if InRange() is false.
func (m Model) RangeStart() time.Time

// RangeEnd returns the later date of the range.
// Returns the zero time if InRange() is false.
func (m Model) RangeEnd() time.Time
```

`RangeStart` and `RangeEnd` return chronologically ordered dates regardless of selection direction. They return `time.Time{}` (zero value) when `InRange()` is false — callers should check `InRange()` first, matching the `Selected()`/`Cursor()` pattern.

### CLI Output

In `runCalendar`:

```go
if cal.InRange() {
    fmt.Fprintln(ctx.w, cal.RangeStart().Format(wen.DateLayout))
    fmt.Fprintln(ctx.w, cal.RangeEnd().Format(wen.DateLayout))
} else if cal.Selected() {
    fmt.Fprintln(ctx.w, cal.Cursor().Format(wen.DateLayout))
}
```

Output format:
```
2026-03-21
2026-04-02
```

Composable: `wen cal | head -1` for start, `tail -1` for end.

### Help Bar

Add `VisualSelect` and `ForceQuit` to `keyMap`. Split the existing `Quit` binding:

- `Quit`: `q`, `esc` — help text: `q/esc`
- `ForceQuit`: `ctrl+c` — no help text (implicit)
- `VisualSelect`: `v` — help text: `range`

Update `ShortHelp` to include `VisualSelect`. Update `FullHelp` to include `VisualSelect` alongside `Select`.

## Files Modified

- `calendar/model.go` — rangeAnchor field, split quit/force-quit bindings, range API methods, Update() logic
- `calendar/view.go` — rangeDay in resolvedStyles, buildStyles, isInRange helper, renderGrid range check
- `calendar/config.go` — Range color in ThemeColors, presets, ResolvedColors, default config template
- `cmd/wen/main.go` — range output branch in runCalendar
- `calendar/model_test.go` — test range selection flow
- `calendar/view_test.go` — test range rendering
- `README.md` — document v keybinding and range output

## Testing

- `TestVisualSelectEnter`: press v, navigate, press Enter — verify InRange, RangeStart, RangeEnd, selected
- `TestVisualSelectCancel`: press v, press Esc — verify anchor cleared, not quit
- `TestVisualSelectReanchor`: press v, navigate, press v again — verify anchor moved
- `TestRangeReverseOrder`: anchor after cursor — verify RangeStart < RangeEnd
- `TestEnterWithoutRange`: Enter without v — verify single-date behavior unchanged
- `TestSameDayRange`: press v, immediately Enter — verify InRange is false, Selected is true (single date)
- `TestCtrlCInRangeMode`: press v, press ctrl+c — verify quit is true (hard exit)
- `TestRangeRendering`: verify range days get range style in View() output
- `TestRangeRenderingMultiMonth`: verify range spanning months highlights correctly in multi-month view
