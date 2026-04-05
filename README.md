# wen

> **wen** — a natural language date tool for your terminal.
>
> Parse dates the way you think about them. Pick dates from an interactive calendar. Get answers in `yyyy-mm-dd`.

---

You don't want to do date math in your head. You want to type `wen next friday` and get `2026-03-20`. You want to browse a calendar in your terminal. wen does both.

## Installation

### Nix

```bash
# Run directly (uses prebuilt binary)
nix run github:zachthieme/wen

# Install into profile
nix profile install github:zachthieme/wen

# Build from source instead
nix build github:zachthieme/wen#wen-src
```

### Go

```bash
go install github.com/zachthieme/wen/cmd/wen@latest
```

Or build locally:

```bash
go build -o wen ./cmd/wen
```

### Flags

| Flag | Description |
|------|-------------|
| `-h`, `--help` | Show help |
| `-v`, `--version` | Show version |
| `--format <layout>` | Output format (Go time layout string, default: `2006-01-02`) |

## Usage

### Date Parsing

```bash
# Today's date
wen

# Natural language dates
wen next friday
wen this thursday
wen last tuesday
wen tomorrow
wen "march 25 2026"
wen "2 weeks ago"
wen "two weeks ago"
wen "in five days"

# Boundaries
wen "end of month"
wen "end of quarter"
wen "end of year"
wen "beginning of next quarter"

# Ordinal weekday in month
wen "first monday of next month"
wen "third thursday in april"

# Multi-date output
wen "every friday in april"

# Pipe from stdin
echo "next friday" | wen
```

Output is always `yyyy-mm-dd` unless `--format` is specified.

### Custom Output Format

```bash
wen --format "January 2, 2006" next friday
# → March 27, 2026
```

### Relative Output

```bash
wen rel "april 1 2026"
# → in 11 days

wen relative "march 25 2026"
# → in 4 days
```

### Date Diff

```bash
wen diff today "april 10"
# → 20 days

wen diff today "april 10" --weeks
# → 2 weeks, 6 days

wen diff today "april 10" --workdays
# → 14 workdays
```

### Interactive Calendar

```bash
# Open calendar at current month
wen cal

# Open at a specific month (assumes current year)
wen cal december
wen calendar march

# Open at a specific month and year
wen cal december 2027
```

Navigate with vim keys (or arrow keys). Press `Enter` to select a date and print it to stdout (useful for scripting: `git log --since=$(wen cal)`). Press `q`, `Esc`, or `ctrl+c` to quit without selecting.

For date ranges, press `v` to anchor a start date, navigate to the end date, then press `Enter`. Both dates are printed (one per line), composable with tools like `git log --since=$(wen cal | head -1) --until=$(wen cal | tail -1)`. Press `Esc` or `q` to cancel the range and return to normal selection. Press `v` again to move the anchor.

#### Keybindings

| Key | Action |
|-----|--------|
| `h` / `l` / `←` / `→` | Previous / next day |
| `j` / `k` / `↓` / `↑` | Next / previous week |
| `H` / `L` | Previous / next month |
| `J` / `K` | Next / previous year |
| `t` | Jump to today |
| `w` | Cycle week numbers: off → left → right → off |
| `?` | Toggle help bar |
| `Enter` | Select date (print to stdout and exit) |
| `v` | Start range selection (move cursor, then Enter to confirm) |
| `q` / `Esc` | Quit (cancels range first if active) |
| `ctrl+c` | Force quit |

The calendar highlights today and your cursor position. Navigation wraps across boundaries (e.g., `l` on March 31 moves to April 1). Month and year jumps clamp the day (e.g., Jan 31 + `L` = Feb 28, Feb 29 + `J` = Feb 28).

#### Calendar Flags

| Flag | Description |
|------|-------------|
| `--months N` | Number of months to display side by side |
| `-N` | Shorthand for `--months N` (e.g., `-3` for three months) |
| `--highlight-file P` | Path to JSON file with dates to highlight |

#### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success (date printed) |
| `2` | Error (parse failure, invalid input, etc.) |

## Configuration

Config lives at `~/.config/wen/config.yaml` (created automatically on first `wen cal`). Config only affects the calendar — date parsing stays zero-config.

```yaml
# Week numbers
show_week_numbers: false  # false, true/"left" (standard), or "right"
week_numbering: us    # "us" or "iso"
week_start_day: 0     # 0=Sunday, 1=Monday

# Fiscal year start month (1-12, default: 1=January)
# Affects "end of quarter", "beginning of quarter", etc.
# Example: fiscal_year_start: 10  # October (common US federal/corporate)
# fiscal_year_start: 1

# Show fiscal quarter in calendar title (e.g., "March 2026 · Q2 FY26")
# Requires fiscal_year_start > 1 to take effect.
# show_fiscal_quarter: false

# Show quarter progress bar below calendar (e.g., "Q1 ████████░░░░ 23wd")
# show_quarter_bar: false

# Theme (built-in: "default", "catppuccin-mocha", "dracula", "nord")
theme: default

# Override individual colors (hex values, override theme):
# colors:
#   cursor: "#f5c2e7"
#   today: "#a6e3a1"
#   title: "#89b4fa"
#   week_number: "#6c7086"
#   day_header: "#94e2d5"
#   help_bar: "#6c7086"
#   highlight: "#f9e2af"
#   range: "#a6e3a1"

# Highlighted dates source (JSON array of yyyy-mm-dd strings):
# highlight_source: ~/.local/share/pike/due.json
```

ISO week numbering forces Monday as the week start day.

## Library Usage

`wen` is also a Go library. Import it to parse natural language dates in your own programs:

```go
import "github.com/zachthieme/wen"

// Parse relative to now
t, err := wen.Parse("next friday")

// Parse relative to a specific reference time
t, err := wen.ParseRelative("march 25 at 3pm", refTime)

// Control how period references resolve
t, err := wen.ParseRelative("next week", refTime, wen.WithPeriodSame())

// Parse multi-date expressions
dates, err := wen.ParseMulti("every friday in april", refTime)

// Fiscal quarter calculation
q, fy := wen.FiscalQuarter(3, 2026, 10) // → Q2, FY2026

// Month name lookup
month, ok := wen.LookupMonth("sep") // → time.September, true
```

See the [package documentation](https://pkg.go.dev/github.com/zachthieme/wen) for full API details and examples.

## Examples

```bash
# Quick date lookup
wen "last tuesday"

# Pipe into other commands
wen next friday | xargs -I{} echo "Meeting on {}"
```

## Project Structure

```
cmd/wen/
  main.go            CLI entry: subcommand routing, flag parsing
wen.go               Library API: Parse, ParseRelative, ParseMulti, FiscalQuarter
lexer.go             Tokenizer: keywords, numbers, ordinals, cardinals, meridiems
parser.go            Recursive descent parser: dates, times, boundaries, ranges
token.go             Token type definitions
errors.go            ParseError with position and input context
calendar/
  config.go          Config loading: YAML, themes, XDG path, validation
  model.go           Bubble Tea model: cursor, range selection, key bindings
  styles.go          Style construction: theme resolution, color application
  render.go          Content rendering: grid, title, quarter bar, date math
  view.go            Layout composition: single/multi-month, week number wrapping
  highlight.go       Highlighted dates: JSON loading, path resolution
```
