# Integrate Natural Language Date Parser into Wen

## Overview

Replace the `olebedev/when` dependency with a built-in natural language date parser. Restructure the repository so the parser is importable as `github.com/zachthieme/wen` and the CLI lives at `cmd/wen/`. This follows the standard Go library+CLI single-repo pattern.

## Current State

```
wen/
  main.go           # package main — CLI entry point
  main_test.go      # CLI tests (190 lines)
  dateparse.go      # two-tier date parsing: custom weekday handler + olebedev/when
  calendar/         # interactive TUI calendar (bubbletea)
    config.go / model.go / view.go + tests
  Makefile / README.md / flake.nix / go.mod
```

Date parsing currently uses:
1. A custom `parseRelativeWeekday()` for "this/next/last weekday" patterns
2. `olebedev/when` as fallback for everything else
3. Output format: `yyyy-mm-dd` via `DateLayout = "2006-01-02"`

## Target State

```
wen/                              # module: github.com/zachthieme/wen
  wen.go                          # package wen — public API (Parse, ParseRelative, Options)
  token.go                        # token types
  errors.go                       # ParseError struct
  lexer.go                        # single-pass lexer
  parser.go                       # recursive descent parser + date resolution
  wen_test.go                     # parser library tests (~80 test cases)
  cmd/
    wen/
      main.go                     # CLI entry point (moved from root)
      main_test.go                # CLI tests (moved from root)
  calendar/                       # untouched
    config.go / model.go / view.go + tests
  Makefile / README.md / flake.nix / go.mod
```

## Changes Required

### 1. Move CLI to `cmd/wen/`

Move `main.go` and `main_test.go` into `cmd/wen/`. Update the calendar import path in both files from `"github.com/zachthieme/wen/calendar"` (already correct if the module path stays the same).

### 2. Copy parser library to root

Copy the 6 library files from the natural-date project into the wen repo root:
- `naturaldate.go` → `wen.go` (rename package to `wen`)
- `token.go` (rename package)
- `errors.go` (rename package)
- `lexer.go` (rename package)
- `parser.go` (rename package)
- `naturaldate_test.go` → `wen_test.go` (rename package)

Public API becomes:
```go
import "github.com/zachthieme/wen"

t, err := wen.Parse("next thursday")
t, err := wen.ParseRelative("in 4 mondays", ref)
t, err := wen.Parse("next week", wen.WithPeriodSame())
```

### 3. Replace `dateparse.go` with simplified version

The current `parseDate` signature is `parseDate(input string, ref time.Time) (time.Time, error)`. The new version preserves this signature and the error wrapping format (required by `main_test.go`):

```go
package main

import (
    "fmt"
    "time"

    "github.com/zachthieme/wen"
)

func parseDate(input string, ref time.Time) (time.Time, error) {
    t, err := wen.ParseRelative(input, ref)
    if err != nil {
        return time.Time{}, fmt.Errorf("could not parse date %q: %w", input, err)
    }
    return t, nil
}
```

The custom `parseRelativeWeekday()`, `newDateParser()`, the `weekdays` map, and the `olebedev/when` import are all removed — the wen library handles this/next/last natively.

### 4. Update Makefile

Change build target from `go build .` to `go build ./cmd/wen`. Update install target similarly.

### 5. Update test build path

`main_test.go` builds the binary with `exec.Command("go", "build", "-o", testBinary, ".")`. This must change to `"./cmd/wen"` since the CLI moved.

### 6. Remove olebedev/when

Run `go mod tidy` to remove the unused dependency.

Update the Nix build expression if it references the build path.

### 7. Change library week start to Sunday

The natural-date library currently uses Monday-start weeks (Mon-Sun). Change `weekdayInWeek` to use Sunday-start weeks (Sun-Sat) to match the current wen behavior and Go's `time.Weekday` convention (Sunday=0).

The change is in `weekdayInWeek` in `parser.go`:

```go
func weekdayInWeek(ref time.Time, target time.Weekday, weekOffset int) time.Time {
    refDow := int(ref.Weekday())    // Sunday=0 ... Saturday=6
    targetDow := int(target)
    sunday := truncateDay(ref).AddDate(0, 0, -refDow)
    sunday = sunday.AddDate(0, 0, 7*weekOffset)
    return sunday.AddDate(0, 0, targetDow)
}
```

Update `wen_test.go` expectations for tests affected by the week boundary change (any test involving Sunday).

## Semantic Alignment

With Sunday-start weeks, the library matches the current wen behavior exactly:

| Expression (on Wednesday) | Current wen | New wen library |
|---|---|---|
| "this thursday" | +1 day | +1 day |
| "this monday" | -2 days | -2 days |
| "this sunday" | -3 days (past) | -3 days (past) |
| "next thursday" | +8 days | +8 days |
| "last thursday" | -6 days | -6 days |

No semantic differences — existing `main_test.go` tests should pass without changes to expectations.

## New capabilities (not in olebedev/when)

- "in 4 mondays" (counted weekdays)
- "the third thursday in april" (ordinal weekday in month)
- "end of month", "beginning of next week" (period boundaries)
- Structured `ParseError` with position and expected tokens
- Configurable period resolution (WithPeriodStart / WithPeriodSame)

## Testing Strategy

1. **Library tests** (`wen_test.go`): ~80 test cases covering all expression types
2. **CLI tests** (`cmd/wen/main_test.go`): Existing tests verify end-to-end behavior
3. **Calendar tests**: Unchanged, independent of date parsing
4. **Verification**: Run `make test` — all existing tests must pass

## What Doesn't Change

- CLI interface (flags, stdin, output format)
- Calendar package (completely untouched)
- CLI test assertions (same inputs → same outputs)
- Module path (`github.com/zachthieme/wen`)
- README usage examples (CLI commands stay the same)
