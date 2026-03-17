# wen — Natural Language Date CLI + Interactive Calendar

## Overview

Rename `zdate` to `wen` and add an interactive terminal calendar with vim-style navigation and date picking. The tool becomes a two-mode utility: natural language date parsing (default) and an interactive calendar picker (subcommand).

## Rename

All references to `zdate` become `wen`:
- Go module name
- Binary name
- GitHub repo (`zachthieme/zdate` → `zachthieme/wen`)
- goreleaser config
- Nix flake
- CI/CD workflows
- `.gitignore`
- Makefile
- Tests

The `--now` flag is removed entirely.

## CLI Interface

```
wen                          # print today's date as yyyy-mm-dd
wen next friday              # parse natural language date, print yyyy-mm-dd
echo "tomorrow" | wen        # read from stdin, parse, print yyyy-mm-dd
wen cal                      # interactive calendar starting at current month
wen cal december 2026        # interactive calendar starting at Dec 2026
```

No flags. `cal` is a subcommand; everything else is date parsing mode.

### Input Mode Detection (no subcommand)

1. **Positional args** (not starting with `cal`) — parse as natural language date
2. **Piped stdin, no args** — read one line, parse
3. **No args, TTY stdin** — print today's date

If both positional args and piped stdin, positional args take precedence.

### Subcommand: `cal`

- `wen cal` — launch interactive calendar at current month, cursor on today
- `wen cal <natural language>` — parse remaining args as a month/date, start calendar there with cursor on the 1st of that month
- If the argument can't be parsed, print error to stderr and exit 1

## Calendar Display

Compact `cal`-style layout rendered with lipgloss:

```
     March 2026
Su Mo Tu We Th Fr Sa
 1  2  3  4  5  6  7
 8  9 10 11 12 13 14
15 16 17 18 19 20 21
22 23 24 25 26 27 28
29 30 31
```

- Month title centered above the day grid
- Cursor date: distinct highlight (e.g., reverse video via lipgloss)
- Today: subtle highlight (e.g., bold or underline) even when cursor is elsewhere
- When cursor is on today, both highlights combine

## Keybindings

| Key | Action |
|-----|--------|
| `h` | Previous day (wraps across month boundaries, view follows) |
| `l` | Next day (wraps across month boundaries, view follows) |
| `k` | Previous week (wraps across month boundaries) |
| `j` | Next week (wraps across month boundaries) |
| `H` | Previous month (same day number, clamped — e.g., Mar 31 → Feb 28) |
| `L` | Next month (same day number, clamped — e.g., Jan 31 → Feb 28) |
| `Enter` | Print cursor date as `yyyy-mm-dd` to stdout, exit 0 |
| `q` / `Esc` | Exit without output, exit 1 |

## Architecture

```
wen/
├── main.go              # CLI entry: subcommand routing, date parsing
├── calendar/
│   ├── model.go         # bubbletea model: state, navigation, key handling
│   └── view.go          # lipgloss rendering: month grid, highlights, cursor
├── go.mod
└── go.sum
```

- `main.go` — thin entry point. Detects `cal` subcommand vs. date parsing mode. Date parsing logic stays inline (~15 lines).
- `calendar/model.go` — bubbletea `Model` implementing `Init`, `Update`, `View`. Holds cursor date (year, month, day) and today's date. Handles all key events and navigation logic.
- `calendar/view.go` — lipgloss-based rendering. Builds the month grid string, applies cursor and today highlights. Pure function of the model state.

## Error Handling

| Condition | stderr message | Exit code |
|-----------|---------------|-----------|
| Parse failure (date mode) | `error: could not parse date "<input>"` | 1 |
| Parse failure (cal mode) | `error: could not parse date "<input>"` | 1 |
| Stdin read failure | `error: failed to read from stdin` | 1 |
| Calendar cancelled (q/Esc) | (no output) | 1 |

## Dependencies

- `olebedev/when` — natural language date parsing (existing)
- `golang.org/x/term` — TTY detection (existing)
- `charmbracelet/bubbletea` — TUI framework (new)
- `charmbracelet/lipgloss` — styling (new)

## Testing

- Existing date parsing tests carry over (updated for module rename). With `--now` removed, tests that need determinism should use environment-based clock injection (e.g., `WEN_NOW=2026-03-17`) or accept that the no-arg test compares against `time.Now()` at test time (current approach).
- Calendar model tests: unit test the `Update` function with synthetic key messages to verify navigation logic (day/week/month wrapping, clamping). The model accepts a starting date, so tests are fully deterministic.
- Calendar view tests: snapshot test the `View` output for a known date to verify grid layout
- Integration test: build binary, run `wen cal`, send keystrokes, verify stdout on Enter
