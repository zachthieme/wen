# wen

> **wen** — a natural language date tool for your terminal.
>
> Parse dates the way you think about them. Pick dates from an interactive calendar. Get answers in `yyyy-mm-dd`.

---

You don't want to do date math in your head. You want to type `wen next friday` and get `2026-03-20`. You want to scroll through a calendar and hit Enter on the date you need. wen does both.

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
go install github.com/zachthieme/wen@latest
```

Or build locally:

```bash
go build -o wen .
```

## Usage

### Date Parsing

```bash
# Today's date
wen

# Natural language dates
wen next friday
wen tomorrow
wen "march 25 2026"
wen "2 weeks ago"

# Pipe from stdin
echo "next friday" | wen
```

Output is always `yyyy-mm-dd`.

### Interactive Calendar

```bash
# Open calendar at current month
wen cal

# Open calendar at a specific month
wen cal december 2026
```

Navigate with vim keys, press Enter to select a date (printed to stdout), or `q`/`Esc` to cancel.

#### Keybindings

| Key | Action |
|-----|--------|
| `h` / `l` | Previous / next day |
| `j` / `k` | Next / previous week |
| `H` / `L` | Previous / next month |
| `Enter` | Print selected date and exit |
| `q` / `Esc` | Exit without output |

The calendar highlights today and your cursor position. Day and month navigation wraps across boundaries (e.g., pressing `l` on March 31 moves to April 1). Month jumps clamp the day (e.g., Jan 31 + `L` = Feb 28).

## Examples

```bash
# Use in scripts
deadline=$(wen cal)
echo "Deadline set to $deadline"

# Quick date lookup
wen "last tuesday"

# Pipe into other commands
wen next friday | xargs -I{} echo "Meeting on {}"
```

## Project Structure

```
main.go              CLI entry: subcommand routing, date parsing
calendar/
  model.go           Bubbletea model: cursor state, navigation, key handling
  view.go            Lipgloss rendering: month grid, highlights
```
