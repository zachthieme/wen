# zdate — Natural Language Date CLI

## Overview

A Go CLI tool that converts natural language date expressions to `yyyy-mm-dd` format. Uses `olebedev/when` for natural language parsing.

## CLI Interface

```
zdate                          # print today's date
zdate "next friday"            # parse and print
zdate --now "2026-01-01" "next friday"  # override reference date
echo "next friday" | zdate     # read from stdin
```

### Flags

- `--now yyyy-mm-dd` — Override the reference date (defaults to today). Applies to all modes.

### Input Modes

1. **No args, TTY stdin** — Print today's date (or `--now` date if provided).
2. **Positional argument** — Parse the argument as a natural language date.
3. **Piped stdin, no positional arg** — Read one line from stdin, parse it.

TTY detection distinguishes mode 1 from mode 3: if stdin is a terminal, print today; if stdin is a pipe, read from it.

### Output

- **Success** — Print `yyyy-mm-dd` to stdout (newline-terminated), exit 0.
- **Failure** — Print error message to stderr, exit 1.

## Error Handling

| Condition | stderr message | Exit code |
|-----------|---------------|-----------|
| Parse failure | `error: could not parse date "<input>"` | 1 |
| Invalid `--now` value | `error: invalid --now date "<value>", expected yyyy-mm-dd` | 1 |
| Stdin read failure | `error: failed to read from stdin` | 1 |

## Dependencies

- `olebedev/when` — Natural language date parsing. Pluggable rules system, supports relative dates ("tomorrow", "next friday", "2 weeks ago"), informal absolute dates ("march 20th"), and conversational dates ("last tuesday", "end of month"). 1.5k GitHub stars, actively maintained.

## Project Structure

```
zdate/
├── main.go          # CLI entry point, flag parsing, stdin/tty detection
├── go.mod
└── go.sum
```

Single file — thin CLI wrapper around `olebedev/when`. No reason to split until complexity warrants it.

## Supported Input Examples

- "tomorrow", "yesterday"
- "next friday", "last tuesday"
- "in 3 days", "2 weeks ago"
- "march 20th", "jan 1 2027", "12/25/2026"
- "end of month"
