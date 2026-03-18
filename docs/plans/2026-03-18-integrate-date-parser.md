# Integrate Date Parser Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace olebedev/when with a built-in natural language date parser, restructure repo as library+CLI.

**Architecture:** Copy parser library files from /home/zach/code/natural-date to wen repo root as `package wen`. Move CLI to cmd/wen/. Replace dateparse.go with a thin wrapper calling `wen.ParseRelative`. Change week start from Monday to Sunday.

**Tech Stack:** Go standard library only (no new dependencies).

**Spec:** `docs/specs/2026-03-18-integrate-date-parser-design.md`

---

## File Structure (after)

```
wen/
  wen.go              # package wen — public API (was naturaldate.go)
  token.go            # token types
  errors.go           # ParseError
  lexer.go            # lexer
  parser.go           # parser + date resolution
  wen_test.go         # library tests (was naturaldate_test.go)
  cmd/
    wen/
      main.go         # CLI (moved from root)
      main_test.go    # CLI tests (moved from root)
      dateparse.go    # thin wrapper calling wen.ParseRelative (replaces old dateparse.go)
  calendar/           # untouched
  Makefile            # updated build paths
  go.mod              # olebedev/when removed
```

---

### Task 1: Copy Library Files + Rename Package

**Files:**
- Create: `wen.go` (copy from `/home/zach/code/natural-date/naturaldate.go`)
- Create: `token.go` (copy from `/home/zach/code/natural-date/token.go`)
- Create: `errors.go` (copy from `/home/zach/code/natural-date/errors.go`)
- Create: `lexer.go` (copy from `/home/zach/code/natural-date/lexer.go`)
- Create: `parser.go` (copy from `/home/zach/code/natural-date/parser.go`)
- Create: `wen_test.go` (copy from `/home/zach/code/natural-date/naturaldate_test.go`)

- [ ] **Step 1: Copy all 6 files**

```bash
cd /home/zach/code/wen
cp /home/zach/code/natural-date/naturaldate.go wen.go
cp /home/zach/code/natural-date/token.go token.go
cp /home/zach/code/natural-date/errors.go errors.go
cp /home/zach/code/natural-date/lexer.go lexer.go
cp /home/zach/code/natural-date/parser.go parser.go
cp /home/zach/code/natural-date/naturaldate_test.go wen_test.go
```

- [ ] **Step 2: Rename package from `naturaldate` to `wen` in all files**

In every copied file, replace `package naturaldate` with `package wen`. The files are:
- `wen.go`
- `token.go`
- `errors.go`
- `lexer.go`
- `parser.go`
- `wen_test.go`

- [ ] **Step 3: Verify library tests pass**

Run: `go test -run "TestLexer|TestRelativeDay|TestModWeekday|TestRelativeOffset|TestCountedWeekday|TestOrdinalWeekdayInMonth|TestAbsoluteDate|TestTimeExpr|TestCombinedExpr|TestParseConvenience|TestWithPeriodStartExplicit" -v`

Expected: PASS (the library tests should work as-is with the package rename, since they don't depend on the old package name)

- [ ] **Step 4: Commit**

```bash
git add wen.go token.go errors.go lexer.go parser.go wen_test.go
git commit -m "feat: add date parser library as package wen"
```

---

### Task 2: Change Week Start from Monday to Sunday

**Files:**
- Modify: `parser.go` — change `weekdayInWeek`, `resolvePeriodRef`, `resolveBoundary`
- Modify: `wen_test.go` — update test expectations for Sunday-start weeks

- [ ] **Step 1: Update `weekdayInWeek` in `parser.go`**

Replace the current `weekdayInWeek` function:

```go
// weekdayInWeek returns the given weekday in the week offset from ref's week.
// Weeks run Sunday through Saturday.
func weekdayInWeek(ref time.Time, target time.Weekday, weekOffset int) time.Time {
	refDow := int(ref.Weekday())    // Sunday=0 ... Saturday=6
	targetDow := int(target)
	// Find Sunday of ref's week
	sunday := truncateDay(ref).AddDate(0, 0, -refDow)
	// Apply week offset
	sunday = sunday.AddDate(0, 0, 7*weekOffset)
	// Find target day in that week
	return sunday.AddDate(0, 0, targetDow)
}
```

- [ ] **Step 2: Update `resolvePeriodRef` week branch in `parser.go`**

Replace the `case "week":` branch in `resolvePeriodRef`:

```go
	case "week":
		dow := int(ref.Weekday())
		sunday := ref.AddDate(0, 0, -dow)

		switch modifier {
		case "this":
			return sunday
		case "next":
			if p.opts.periodMode == PeriodSame {
				return ref.AddDate(0, 0, 7)
			}
			return sunday.AddDate(0, 0, 7)
		case "last":
			if p.opts.periodMode == PeriodSame {
				return ref.AddDate(0, 0, -7)
			}
			return sunday.AddDate(0, 0, -7)
		}
```

- [ ] **Step 3: Update `resolveBoundary` week branch in `parser.go`**

Replace the `case "week":` branch in `resolveBoundary`:

```go
	case "week":
		dow := int(ref.Weekday())
		sunday := ref.AddDate(0, 0, -dow)
		switch modifier {
		case "next":
			sunday = sunday.AddDate(0, 0, 7)
		case "last":
			sunday = sunday.AddDate(0, 0, -7)
		}
		if boundary == "beginning" {
			return sunday
		}
		// end = Saturday 23:59:59
		return sunday.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
```

- [ ] **Step 4: Update `wen_test.go` expectations**

The following test values change from Monday-start to Sunday-start (ref = Wednesday March 18, 2026):

In `TestModWeekday`:
- `{"sunday", date(2026, 3, 22)}` → `{"sunday", date(2026, 3, 15)}`
- `{"last sunday", date(2026, 3, 15)}` → `{"last sunday", date(2026, 3, 8)}`

In `TestPeriodRef`:
- `{"this week", nil, date(2026, 3, 16)}` → `{"this week", nil, date(2026, 3, 15)}`
- `{"next week", nil, date(2026, 3, 23)}` → `{"next week", nil, date(2026, 3, 22)}`
- `{"last week", nil, date(2026, 3, 9)}` → `{"last week", nil, date(2026, 3, 8)}`
- `{"beginning of next week", nil, date(2026, 3, 23)}` → `{"beginning of next week", nil, date(2026, 3, 22)}`

In `TestEdgeCases`:
- The "ref on Sunday: this week Monday" test: Sunday March 22 is now the start of a new Sun-Sat week, so "this monday" = March 23 (not March 16):
  - `want: date(2026, 3, 16)` → `want: date(2026, 3, 23)`
  - Update the test name to reflect: `"ref on Sunday: this monday is tomorrow"`

- [ ] **Step 5: Run library tests to verify**

Run: `go test -run "TestModWeekday|TestPeriodRef|TestEdgeCases" -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add parser.go wen_test.go
git commit -m "feat: change week start from Monday to Sunday"
```

---

### Task 3: Move CLI to cmd/wen/ + Replace dateparse.go

**Files:**
- Move: `main.go` → `cmd/wen/main.go`
- Move: `main_test.go` → `cmd/wen/main_test.go`
- Delete: `dateparse.go` (old root-level file)
- Create: `cmd/wen/dateparse.go` (simplified replacement)

- [ ] **Step 1: Create cmd/wen/ directory and move files**

```bash
mkdir -p cmd/wen
git mv main.go cmd/wen/main.go
git mv main_test.go cmd/wen/main_test.go
```

- [ ] **Step 2: Delete old dateparse.go**

```bash
git rm dateparse.go
```

- [ ] **Step 3: Create new `cmd/wen/dateparse.go`**

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

- [ ] **Step 4: Verify build**

Run: `go build ./cmd/wen`
Expected: clean build

- [ ] **Step 5: Commit**

```bash
git add cmd/wen/dateparse.go
git commit -m "refactor: move CLI to cmd/wen/, replace dateparse with wen library"
```

---

### Task 4: Update Build Paths, Tests, and Cleanup

**Files:**
- Modify: `cmd/wen/main_test.go` — fix build path, rewrite weekday tests
- Modify: `Makefile` — update build/install paths
- Modify: `flake.nix` — add subPackages if needed
- Modify: `go.mod` / `go.sum` — remove olebedev/when

- [ ] **Step 1: Update `cmd/wen/main_test.go` build path**

In `runTests`, change line 31:

```go
// Old:
cmd := exec.Command("go", "build", "-o", testBinary, ".")
// New:
cmd := exec.Command("go", "build", "-o", testBinary, "./cmd/wen")
```

- [ ] **Step 2: Rewrite `TestThisVsNextWeekday` in `cmd/wen/main_test.go`**

The current test calls `parseRelativeWeekday()` which no longer exists. Rewrite to call `parseDate()`:

```go
func TestThisVsNextWeekday(t *testing.T) {
	// Reference: Tuesday March 17, 2026
	ref := time.Date(2026, 3, 17, 12, 0, 0, 0, time.Local)

	tests := []struct {
		input string
		want  string
	}{
		{"this thursday", "2026-03-19"},
		{"next thursday", "2026-03-26"},
		{"this tuesday", "2026-03-17"},
		{"next tuesday", "2026-03-24"},
		{"this monday", "2026-03-16"},
		{"next monday", "2026-03-23"},
		{"this friday", "2026-03-20"},
		{"next friday", "2026-03-27"},
		{"this sunday", "2026-03-15"},
		{"next sunday", "2026-03-22"},
		{"this saturday", "2026-03-21"},
		{"next saturday", "2026-03-28"},
		{"this thu", "2026-03-19"},
		{"next fri", "2026-03-27"},
		{"last thursday", "2026-03-12"},
		{"last tuesday", "2026-03-10"},
		{"last sunday", "2026-03-08"},
		{"last sat", "2026-03-14"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseDate(tt.input, ref)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if got.Format("2006-01-02") != tt.want {
				t.Errorf("parseDate(%q) = %s, want %s", tt.input, got.Format("2006-01-02"), tt.want)
			}
		})
	}
}
```

Note: `"last sunday"` changed from `"2026-03-15"` to `"2026-03-08"` — with Sunday-start weeks, "last sunday" is the Sunday of the previous week, not the current week's Sunday (which is "this sunday"). The old behavior had "this sunday" and "last sunday" returning the same date, which was a quirk of the simple arithmetic.

- [ ] **Step 3: Rewrite `TestRelativeWeekdayDoesNotMatchOtherInputs`**

The old test called `parseRelativeWeekday()` which no longer exists. Some inputs that didn't match the old function DO match the new library ("tomorrow", "2 weeks ago"). Replace with a test that verifies `parseDate` rejects truly invalid input:

```go
func TestParseDateRejectsInvalidInput(t *testing.T) {
	ref := time.Date(2026, 3, 17, 12, 0, 0, 0, time.Local)
	inputs := []string{"pizza", "next", "this"}
	for _, input := range inputs {
		if _, err := parseDate(input, ref); err == nil {
			t.Errorf("expected error for %q", input)
		}
	}
}
```

- [ ] **Step 4: Remove the `calendar` import from `cmd/wen/main_test.go`**

The test file imports `"github.com/zachthieme/wen/calendar"` for `calendar.DateLayout`. Since the test no longer calls `parseRelativeWeekday` (which returned `time.Time` formatted with `calendar.DateLayout`), the new `TestThisVsNextWeekday` uses `"2006-01-02"` directly. However, `TestNoArgs_PrintsToday` still uses `calendar.DateLayout`, so the import stays. No change needed.

- [ ] **Step 5: Update `Makefile`**

```makefile
.PHONY: build test lint check install

build:
	go build -o wen ./cmd/wen

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run

check: test lint

install:
	go install ./cmd/wen
```

- [ ] **Step 6: Remove olebedev/when**

```bash
go mod tidy
```

This removes `github.com/olebedev/when` and its transitive dependencies from `go.mod` and `go.sum`.

- [ ] **Step 7: Run full test suite**

Run: `go test -race -count=1 -v ./...`
Expected: ALL tests pass — library tests (`wen_test.go`), CLI tests (`cmd/wen/main_test.go`), calendar tests (`calendar/*_test.go`)

- [ ] **Step 8: Verify build and vet**

```bash
go build ./cmd/wen
go vet ./...
```

- [ ] **Step 9: Commit**

```bash
git add -A
git commit -m "chore: update build paths, tests, remove olebedev/when"
```

---

## Verification

After all tasks, run the full check:

```bash
make check   # runs test + lint
make build   # produces ./wen binary
./wen next friday
./wen "the third thursday in april"
./wen "in 4 mondays"
```
