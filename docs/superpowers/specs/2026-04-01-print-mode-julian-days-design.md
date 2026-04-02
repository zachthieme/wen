# Non-Interactive Print Mode & Julian Day-of-Year Numbering

## Overview

Two features that let `wen` fully replace the Unix `cal` utility:

1. **Non-interactive print mode** -- render the calendar to stdout and exit, no TUI session
2. **Julian day-of-year numbering** -- display days as their position in the year (1--365/366) instead of day-of-month

## Feature 1: Non-Interactive Print Mode

### Invocation

- `--print` / `-p` flag on `wen cal` and `wen row` forces print mode on a TTY
- Auto-detect: when stdout is not a TTY (piped or redirected), print mode activates automatically
- Logic: `printMode = flagPrint || !term.IsTerminal(stdout)`

### Behavior

- Build the calendar Model or RowModel as normal, applying all config and flags
- No cursor highlighting (the concept of a "selected day" doesn't apply in print mode)
- Today is highlighted with styling when output is a TTY; ANSI is stripped automatically by lipgloss when piped
- Highlighted dates from `--highlight-file` / config are styled when output is a TTY
- Call `View()` on the model, write the result to stdout, exit
- No `tea.NewProgram`, no alt screen, no event loop
- Help bar is not shown

### Flags and config respected

All existing flags and config options apply to print mode:

- `--months N` / `-N` (cal only) -- multiple months side by side
- `--padding-*` -- padding around the output
- `--highlight-file` -- highlighted dates
- Config: `week_start_day`, `show_week_numbers`, `week_numbering`, `fiscal_year_start`, `show_fiscal_quarter`, `show_quarter_bar`, `theme`, `colors`, `highlight_source`

### Examples

```
wen cal --print              # print current month
wen cal --print march        # print March of current year
wen cal --print march 2027   # print March 2027
wen cal --print -3           # print 3 months centered on current
wen row --print              # print strip calendar
wen cal | cat                # auto-detect: prints without TUI
```

### Implementation approach

In `cal.go` and `row.go`, after flag parsing and config setup:

1. Check if print mode is active (`--print` flag or stdout is not a TTY)
2. Build the model with the same constructor (`calendar.New` / `calendar.NewRow`)
3. Set `printMode` on the model via a new `WithPrintMode(bool)` option -- this suppresses cursor styling in `renderGrid` / `renderStripDays` (the cursor date is still used to determine which month to display, so we can't zero it out)
4. Call `View()` to get the rendered string
5. Write to stdout and return (skip the `tea.NewProgram` / `p.Run()` path)

Terminal detection uses `golang.org/x/term.IsTerminal(int(os.Stdout.Fd()))`. This package is already an indirect dependency via charmbracelet.

## Feature 2: Julian Day-of-Year Numbering

### What it does

Replaces day-of-month numbers (1--31) with day-of-year numbers (1--366). January 1 = 1, February 1 = 32, December 31 = 365 (or 366 in leap years).

### Grid calendar changes

- **Cell width**: 2 chars -> 3 chars (Julian days are up to 3 digits)
- **Day headers**: 2 chars -> 3 chars (`Su` -> `Sun`, `Mo` -> `Mon`, etc.)
- **`dayGridWidth`**: 20 -> 27 (7 columns x 3 chars + 6 spaces)
- **Day formatting**: `strconv.Itoa(day)` with 2-char padding -> `strconv.Itoa(t.YearDay())` with 3-char padding
- Leading blanks in the first week use 4-char spacing (`"    "`) instead of 3 (`"   "`)

### Strip calendar changes

- **Cell width**: `%2d` -> `%3d` using `YearDay()`
- **Day headers**: 2-char abbreviations -> 3-char abbreviations
- **Leading prefix**: 3-char month abbreviation column stays the same width

### TUI toggle

`J` (shift-J) toggles Julian mode on/off in both views. Added to the keymap and help bar.

**Grid calendar key conflict**: `J` is currently bound to "next year". Reassign it:
- `J` becomes Julian toggle
- "Next year" moves to `N` (mnemonic: next year). "Prev year" `K` moves to `P` (prev year). `N`/`P` are currently unmapped.
- **Row calendar**: `J` is unmapped (row has no year navigation), so no conflict.

### CLI flag

- `--julian` / `-j` on both `wen cal` and `wen row`
- Starts the TUI in Julian mode (still toggleable with `J`)
- In print mode, prints Julian days

### Config file

```yaml
julian: false  # default: show day-of-month
```

- CLI `--julian` overrides config
- TUI `J` key overrides both at runtime

### Model changes

- Add `julian bool` to `Model` and `RowModel`
- Add `printMode bool` to `Model` and `RowModel` (suppresses cursor styling)
- Add `WithJulian(bool) ModelOption` and `WithRowJulian(bool) RowModelOption`
- Add `WithPrintMode(bool) ModelOption` and `WithRowPrintMode(bool) RowModelOption`
- `renderGrid`: check `m.julian` to decide cell width and day number source
- `renderStripDays`: same check
- `renderDayHeaders`: check `m.julian` to decide header width
- `renderStripDayHeaders`: same check

### Width cascading

When Julian mode is on:
- `dayGridWidth` changes from 20 to 27
- `renderTitle` width changes to match
- `wrapWithWeekNums` column width adjusts
- `renderMultiMonth` column width adjusts
- Quarter bar width adjusts

These are all derived from `dayGridWidth`, so a method like `m.gridWidth() int` that returns 20 or 27 based on `m.julian` keeps it DRY.

## Testing

### Non-interactive print mode

- **In-process tests** (`cmd/wen/main_test.go`): call `run(&buf, args)` with `--print` and verify output
  - `wen cal --print` -- produces a month grid
  - `wen cal --print march 2026` -- deterministic output for a known month
  - `wen cal --print -3 march 2026` -- three months side by side
  - `wen row --print` -- produces strip output
  - `wen row --print march 2026` -- deterministic strip
- **CLI integration tests**: build binary, pipe stdout, verify no TUI launched and output matches
- **TTY auto-detect**: test that piping stdout triggers print mode (integration test with `exec.Command` piping to a buffer)

### Julian mode

- **Unit tests** (`calendar/render_test.go`): render grid with `julian=true`, verify:
  - Day numbers are YearDay values (e.g., Jan 1 = 1, Feb 1 = 32)
  - Cells are 3 chars wide
  - Headers are 3-char abbreviations
  - Grid width is 27
- **Unit tests** (`calendar/row_render_test.go`): render strip with julian, verify 3-char cells and YearDay values
- **In-process CLI tests**: `wen cal --print --julian march 2026` produces julian output
- **Combined**: `wen cal --print --julian -3`
- **Config**: test `julian: true` in config is respected, `--julian=false` overrides

### Edge cases

- Julian day 366 on Dec 31 of a leap year
- Julian days across month boundaries in multi-month view
- Julian mode in strip calendar with terminal width constraints (3-char cells reduce visible days)
- Print mode with `--highlight-file` on a non-TTY (no ANSI, but highlighted dates should still appear in the data -- they just won't be visually distinct)

## Scope boundaries

**In scope:**
- `--print` / `-p` flag and TTY auto-detect
- `--julian` / `-j` flag and `J` key toggle
- `julian` config option
- Julian cell width changes in grid and strip
- Tests for both features

**Out of scope:**
- Year view (`cal -y` equivalent) -- can be done with `-12` already
- Replacing `cal` as a shell alias (user's choice)
- Any changes to the date parser
- Any changes to `wen diff` or `wen rel`
