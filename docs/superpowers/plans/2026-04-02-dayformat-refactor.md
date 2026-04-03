# dayFormat Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace 9 scattered `if m.julian` conditionals with a single pre-computed `dayFormat` struct, eliminating rendering branching.

**Architecture:** A `dayFormat` struct captures cell width, grid width, prefix width, day name array, and day formatting function. It is computed once from the `julian` bool in model constructors and on toggle. All rendering functions read fields from the struct instead of checking `m.julian`.

**Tech Stack:** Go, Bubble Tea (calendar package)

---

### Task 1: Add dayFormat struct and constructors

**Files:**
- Modify: `calendar/render.go`
- Test: `calendar/render_test.go`

- [ ] **Step 1: Write failing tests for dayFormat constructors**

In `calendar/render_test.go`, replace `TestGridWidth` with:

```go
func TestDayFormat(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		df := normalDayFormat()
		if df.cellWidth != 2 {
			t.Errorf("cellWidth = %d, want 2", df.cellWidth)
		}
		if df.gridWidth != 20 {
			t.Errorf("gridWidth = %d, want 20", df.gridWidth)
		}
		if df.prefixWidth != 3 {
			t.Errorf("prefixWidth = %d, want 3", df.prefixWidth)
		}
		if df.names != dayNames {
			t.Errorf("names = %v, want dayNames", df.names)
		}
		got := df.formatDay(2026, time.March, 5, time.Local)
		if got != " 5" {
			t.Errorf("formatDay = %q, want %q", got, " 5")
		}
	})
	t.Run("julian", func(t *testing.T) {
		df := julianDayFormat()
		if df.cellWidth != 3 {
			t.Errorf("cellWidth = %d, want 3", df.cellWidth)
		}
		if df.gridWidth != 27 {
			t.Errorf("gridWidth = %d, want 27", df.gridWidth)
		}
		if df.prefixWidth != 4 {
			t.Errorf("prefixWidth = %d, want 4", df.prefixWidth)
		}
		if df.names != dayNamesLong {
			t.Errorf("names = %v, want dayNamesLong", df.names)
		}
		// March 5 2026 = yearday 64
		got := df.formatDay(2026, time.March, 5, time.Local)
		if got != " 64" {
			t.Errorf("formatDay = %q, want %q", got, " 64")
		}
	})
	t.Run("dayFormatFor false", func(t *testing.T) {
		df := dayFormatFor(false)
		if df.cellWidth != 2 {
			t.Errorf("dayFormatFor(false) cellWidth = %d, want 2", df.cellWidth)
		}
	})
	t.Run("dayFormatFor true", func(t *testing.T) {
		df := dayFormatFor(true)
		if df.cellWidth != 3 {
			t.Errorf("dayFormatFor(true) cellWidth = %d, want 3", df.cellWidth)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run TestDayFormat -v`
Expected: FAIL — `normalDayFormat`, `julianDayFormat`, `dayFormatFor` undefined

- [ ] **Step 3: Add dayFormat struct and constructors to render.go**

In `calendar/render.go`, replace the `dayGridWidth`, `julianGridWidth` constants and `gridWidth()` method with:

```go
// dayFormat captures rendering dimensions that vary between normal and julian mode.
type dayFormat struct {
	cellWidth   int                                                        // character width of a day number (2 or 3)
	gridWidth   int                                                        // character width of the 7-column day grid
	prefixWidth int                                                        // total width of the strip's leading prefix column
	names       [7]string                                                  // day-of-week abbreviations
	formatDay   func(year int, month time.Month, day int, loc *time.Location) string // formats a day number
}

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

Remove these lines:
- `const dayGridWidth = 20` (line 17-18)
- `const julianGridWidth = 27` (line 20-21)
- The `gridWidth()` method (lines 45-51)

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run TestDayFormat -v`
Expected: PASS

Note: Other tests will fail at this point because `gridWidth()` was removed and `dayGridWidth`/`julianGridWidth` were removed. That's expected — Task 2 fixes them.

- [ ] **Step 5: Commit**

```bash
git add calendar/render.go calendar/render_test.go
git commit -m "refactor(calendar): add dayFormat struct and constructors"
```

---

### Task 2: Wire dayFormat into Model and RowModel, update rendering

**Files:**
- Modify: `calendar/model.go`
- Modify: `calendar/row_model.go`
- Modify: `calendar/render.go`
- Modify: `calendar/row_render.go`
- Modify: `calendar/view.go`

This task adds the `dayFmt` field, sets it in constructors and toggle handlers, and replaces all `if m.julian` rendering conditionals with `dayFmt` field access.

- [ ] **Step 1: Add dayFmt field to Model and set in New()**

In `calendar/model.go`, add `dayFmt dayFormat` to the Model struct after `printMode`:

```go
type Model struct {
	cursor           time.Time
	today            time.Time
	quit             bool
	selected         bool
	rangeAnchor      *time.Time
	weekNumPos       WeekNumPos
	showHelp         bool
	months           int
	julian           bool
	printMode        bool
	dayFmt           dayFormat
	highlightedDates map[time.Time]bool
	highlightPath    string
	activeWatcher    *fsnotify.Watcher
	config           Config
	keys             keyMap
	help             help.Model
	styles           resolvedStyles
}
```

In `New()`, after the `for _, opt := range opts` loop (line 140), add:

```go
	m.dayFmt = dayFormatFor(m.julian)
	return m
```

(Replace the existing bare `return m` on line 141.)

- [ ] **Step 2: Add dayFmt field to RowModel and set in NewRow()**

In `calendar/row_model.go`, add `dayFmt dayFormat` to the RowModel struct after `printMode`:

```go
type RowModel struct {
	cursor           time.Time
	today            time.Time
	quit             bool
	selected         bool
	rangeAnchor      *time.Time
	highlightedDates map[time.Time]bool
	highlightPath    string
	activeWatcher    *fsnotify.Watcher
	config           Config
	keys             rowKeyMap
	help             help.Model
	styles           resolvedStyles
	showHelp         bool
	julian           bool
	printMode        bool
	dayFmt           dayFormat
	termWidth        int
}
```

In `NewRow()`, after the `for _, opt := range opts` loop, add `m.dayFmt = dayFormatFor(m.julian)` before the `return m`.

- [ ] **Step 3: Update ToggleJulian in both Update() handlers**

In `calendar/model.go` Update(), change:
```go
case key.Matches(msg, m.keys.ToggleJulian):
	m.julian = !m.julian
```
to:
```go
case key.Matches(msg, m.keys.ToggleJulian):
	m.julian = !m.julian
	m.dayFmt = dayFormatFor(m.julian)
```

In `calendar/row_model.go` Update(), make the same change:
```go
case key.Matches(msg, m.keys.ToggleJulian):
	m.julian = !m.julian
	m.dayFmt = dayFormatFor(m.julian)
```

- [ ] **Step 4: Replace julian conditionals in render.go**

In `renderTitle`, replace `m.gridWidth()` with `m.dayFmt.gridWidth`:
```go
titleStyle := m.styles.title.Width(m.dayFmt.gridWidth).Align(lipgloss.Center)
```

In `renderDayHeaders`, replace the julian branching:
```go
func (m Model) renderDayHeaders(b *strings.Builder) {
	startDay := m.config.WeekStartDay
	headers := make([]string, 7)
	for i := range 7 {
		headers[i] = m.dayFmt.names[(startDay+i)%7]
	}
	b.WriteString(m.styles.dayHeader.Render(strings.Join(headers, " ")))
	b.WriteString("\n")
}
```

In `renderGrid`, replace the cellWidth computation, blankCell, and day formatting:

Replace lines 92-96 (`cellWidth := 2; if m.julian { cellWidth = 3 }`):
```go
	cellWidth := m.dayFmt.cellWidth
```

Replace lines 105-111 (`var dayStr string; if m.julian { ... } else { ... }`):
```go
		dayStr := m.dayFmt.formatDay(year, month, day, loc)
```

- [ ] **Step 5: Replace julian conditionals in row_render.go**

Replace `renderStripDayHeaders`:
```go
func (m RowModel) renderStripDayHeaders(start, end time.Time) string {
	var b strings.Builder
	b.WriteString(strings.Repeat(" ", m.dayFmt.prefixWidth))
	first := true
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if !first {
			b.WriteString(" ")
		}
		b.WriteString(m.dayFmt.names[d.Weekday()])
		first = false
	}
	return m.styles.dayHeader.Render(b.String())
}
```

In `renderStripDays`, replace the prefix spacing (lines 93-98):
```go
	b.WriteString(m.styles.title.Render(abbrev))
	b.WriteString(strings.Repeat(" ", m.dayFmt.prefixWidth-2))
```

Replace the day formatting (lines 114-119):
```go
		dayStr := m.dayFmt.formatDay(d.Year(), d.Month(), d.Day(), loc)
```

Note: `loc` is the cursor's location — add it before the loop:
```go
	loc := m.cursor.Location()
```

Wait — `loc` is already used implicitly via `d` which is derived from `start`/`end`. But `formatDay` needs a `*time.Location`. The strip days are iterated from `start` to `end`, and `d.Location()` gives it. However, `renderStripDays` already has `loc` defined on line 86: `loc := m.cursor.Location()`. But `d.YearDay()` (used by the julian formatter) doesn't depend on location — `YearDay()` is computed from the date's own location. So pass `d.Location()`:

```go
		dayStr := m.dayFmt.formatDay(d.Year(), d.Month(), d.Day(), d.Location())
```

Actually, `d` is iterated starting from `start` which comes from `stripWindow` using `loc`. So `d.Location()` will be `loc`. Using `d.Location()` is cleaner and correct.

- [ ] **Step 6: Replace gridWidth() calls in view.go**

In `renderSingleMonth` (line 78):
```go
m.renderQuarterBar(&core, m.dayFmt.gridWidth)
```

In `renderMultiMonth` (line 134):
```go
colWidth := m.dayFmt.gridWidth
```

- [ ] **Step 7: Replace julian conditionals in visibleWindow (row_model.go)**

Replace lines 239-246:
```go
	// Each day cell is cellWidth+1 chars (number + space separator).
	// The prefix occupies prefixWidth chars before the first cell.
	// Total for N days: prefixWidth + N*(cellWidth+1) - 1 (no trailing space).
	// Solving for N: (availWidth - prefixWidth + 1) / (cellWidth + 1)
	cellW := m.dayFmt.cellWidth + 1
	maxDays := (availWidth - m.dayFmt.prefixWidth + 1) / cellW
```

- [ ] **Step 8: Run all tests**

Run: `cd /home/zach/code/wen && go test ./calendar/ -v`
Expected: All PASS

Run: `cd /home/zach/code/wen && make check`
Expected: All PASS, 0 lint issues

- [ ] **Step 9: Commit**

```bash
git add calendar/model.go calendar/row_model.go calendar/render.go calendar/row_render.go calendar/view.go
git commit -m "refactor(calendar): replace julian conditionals with dayFormat fields"
```

---

### Task 3: Update toggle tests to verify dayFmt updates

**Files:**
- Modify: `calendar/model_test.go`
- Modify: `calendar/row_model_test.go`

- [ ] **Step 1: Update TestToggleJulian in model_test.go**

Replace the existing test:

```go
func TestToggleJulian(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	if m.julian {
		t.Error("expected julian false initially")
	}
	if m.dayFmt.gridWidth != 20 {
		t.Errorf("expected gridWidth 20 initially, got %d", m.dayFmt.gridWidth)
	}
	m = pressKey(m, "J")
	if !m.julian {
		t.Error("expected julian true after toggle")
	}
	if m.dayFmt.gridWidth != 27 {
		t.Errorf("expected gridWidth 27 after julian toggle, got %d", m.dayFmt.gridWidth)
	}
	m = pressKey(m, "J")
	if m.julian {
		t.Error("expected julian false after second toggle")
	}
	if m.dayFmt.gridWidth != 20 {
		t.Errorf("expected gridWidth 20 after second toggle, got %d", m.dayFmt.gridWidth)
	}
}
```

- [ ] **Step 2: Update TestRowToggleJulian in row_model_test.go**

Replace the existing test:

```go
func TestRowToggleJulian(t *testing.T) {
	t.Parallel()
	m := NewRow(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	if m.julian {
		t.Error("expected julian false initially")
	}
	if m.dayFmt.cellWidth != 2 {
		t.Errorf("expected cellWidth 2 initially, got %d", m.dayFmt.cellWidth)
	}
	updated, _ := m.Update(runeMsg("J"))
	m = updated.(RowModel)
	if !m.julian {
		t.Error("expected julian true after toggle")
	}
	if m.dayFmt.cellWidth != 3 {
		t.Errorf("expected cellWidth 3 after julian toggle, got %d", m.dayFmt.cellWidth)
	}
	updated, _ = m.Update(runeMsg("J"))
	m = updated.(RowModel)
	if m.julian {
		t.Error("expected julian false after second toggle")
	}
	if m.dayFmt.cellWidth != 2 {
		t.Errorf("expected cellWidth 2 after second toggle, got %d", m.dayFmt.cellWidth)
	}
}
```

- [ ] **Step 3: Run tests**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run 'TestToggleJulian|TestRowToggleJulian' -v`
Expected: PASS

- [ ] **Step 4: Run full suite**

Run: `cd /home/zach/code/wen && make check`
Expected: All PASS, 0 lint issues

- [ ] **Step 5: Commit**

```bash
git add calendar/model_test.go calendar/row_model_test.go
git commit -m "test(calendar): verify dayFmt updates on julian toggle"
```

---

### Task 4: Clean up removed constants from test references

**Files:**
- Modify: `calendar/render_test.go`

Check if any remaining tests reference `dayGridWidth` or `julianGridWidth` directly.

- [ ] **Step 1: Search for stale references**

Run: `cd /home/zach/code/wen && grep -rn 'dayGridWidth\|julianGridWidth' calendar/`

If any are found in test files (e.g., `TestRenderQuarterBar` uses `dayGridWidth`), replace with the literal value `20` or use `normalDayFormat().gridWidth`.

- [ ] **Step 2: Fix any references found**

In `calendar/render_test.go`, if `dayGridWidth` is referenced in `TestRenderQuarterBar` (line ~267):
```go
m.renderQuarterBar(&b, normalDayFormat().gridWidth)
```

In `calendar/view_test.go`, `dayGridWidth` is referenced in the `wrapWithWeekNums` comment only — that was already fixed in the earlier stale-comments commit.

- [ ] **Step 3: Run full suite**

Run: `cd /home/zach/code/wen && make check`
Expected: All PASS, 0 lint issues

- [ ] **Step 4: Commit (if changes were needed)**

```bash
git add calendar/render_test.go
git commit -m "refactor(calendar): remove stale dayGridWidth references from tests"
```
