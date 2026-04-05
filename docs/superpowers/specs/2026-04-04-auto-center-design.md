# Auto-Center Calendar in Terminal

## Summary

Replace the manual padding configuration with automatic centering. Both the calendar view (Model) and strip view (RowModel) render centered horizontally and vertically in the terminal, like a dialog box. Uses `lipgloss.Place()`.

## Changes

### Config removal

- Remove `PaddingTop`, `PaddingRight`, `PaddingBottom`, `PaddingLeft` from the `Config` struct
- Remove their YAML parsing, validation (min/max), and default values
- Remove `--padding-top`, `--padding-right`, `--padding-bottom`, `--padding-left` CLI flags from `cal.go` and `row.go`
- Remove `styles.padding` from `resolvedStyles`

### Model changes

- **Model**: Add `termWidth` and `termHeight` int fields
- **RowModel**: Add `termHeight` int field (already has `termWidth`)
- Both models store both dimensions from `tea.WindowSizeMsg` in their `Update()` methods

### View changes

- Replace `m.styles.padding.Render(output)` with `lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Center, lipgloss.Center, output)` in both `Model.View()` and `RowModel.View()`
- Guard against zero dimensions (print mode): return raw output without `Place()`

### Print mode

Print mode does not receive `WindowSizeMsg`. Output is returned raw with no centering and no padding. This is correct — print mode output is for piping/capturing.

### Testing

- Remove tests asserting padding behavior
- Update tests constructing `Config` structs to drop padding fields
- No new centering unit tests — `lipgloss.Place()` is a framework function
- `make check` must pass clean
