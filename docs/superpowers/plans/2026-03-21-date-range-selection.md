# Date Range Selection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add vim-style visual range selection (`v` to anchor, `Enter` to confirm) to the wen calendar TUI, outputting two dates (start/end) to stdout.

**Architecture:** Add `rangeAnchor *time.Time` to the calendar Model. Split the existing Quit binding into Quit (`q`/`esc`) and ForceQuit (`ctrl+c`). Add `isInRange()` helper to view rendering. CLI prints two lines when a range is confirmed.

**Tech Stack:** Go, Bubble Tea, lipgloss

**Spec:** `docs/superpowers/specs/2026-03-21-date-range-selection-design.md`

---

### Task 1: Theme — Add Range color to config

**Files:**
- Modify: `calendar/config.go:20-29` (ThemeColors struct)
- Modify: `calendar/config.go:102-114` (ResolvedColors)
- Modify: `calendar/config.go:116-145` (themePresets)

- [ ] **Step 1: Add Range field to ThemeColors**

In `calendar/config.go`, add after `Highlight string`:

```go
Range     string `yaml:"range"`
```

- [ ] **Step 2: Add Range to ResolvedColors**

In `ResolvedColors()`, add after the Highlight merge line:

```go
Range:      mergeColor(base.Range, c.Colors.Range),
```

- [ ] **Step 3: Add Range colors to theme presets**

In `themePresets`, add to each theme:
- `"catppuccin-mocha"`: `Range: "#a6e3a1"`
- `"dracula"`: `Range: "#50fa7b"`
- `"nord"`: `Range: "#a3be8c"`
- `"default"`: no change (empty string triggers fallback)

- [ ] **Step 4: Update writeDefaultConfig template**

In the `writeDefaultConfig` function's inline config template string, after the `#   highlight:` line, add:

```
#   range: "#hexcolor"
```

- [ ] **Step 5: Run tests**

Run: `go test ./calendar/ -count=1`
Expected: All existing tests pass (no behavioral change yet)

- [ ] **Step 6: Commit**

```bash
git add calendar/config.go
git commit -m "Add Range color to theme system"
```

---

### Task 2: View — Add rangeDay style and isInRange helper

**Files:**
- Modify: `calendar/model.go:34-45` (resolvedStyles struct)
- Modify: `calendar/view.go:21-46` (buildStyles)
- Modify: `calendar/view.go:197-257` (renderGrid)

- [ ] **Step 1: Add rangeDay to resolvedStyles**

In `calendar/model.go`, add after `highlight lipgloss.Style`:

```go
rangeDay lipgloss.Style
```

- [ ] **Step 2: Build rangeDay style in buildStyles**

In `calendar/view.go`, after the `highlightStyle` block (lines 30-35), add:

```go
rangeDayStyle := lipgloss.NewStyle().Reverse(true)
if colors.Range != "" {
    rangeDayStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Range))
}
```

Add `rangeDay: rangeDayStyle,` to the returned `resolvedStyles` struct.

- [ ] **Step 3: Add isInRange helper**

In `calendar/view.go`, before `renderGrid`, add:

```go
// isInRange reports whether d falls between a and b (inclusive), regardless of order.
func isInRange(d, a, b time.Time) bool {
	if a.After(b) {
		a, b = b, a
	}
	return !d.Before(a) && !d.After(b)
}
```

- [ ] **Step 4: Add range check to renderGrid**

In `renderGrid`, after line 219 (`isHighlighted := ...`), add:

```go
isRangeDay := false
if m.rangeAnchor != nil {
    dayDate := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
    anchorUTC := time.Date(m.rangeAnchor.Year(), m.rangeAnchor.Month(), m.rangeAnchor.Day(), 0, 0, 0, 0, time.UTC)
    cursorUTC := time.Date(m.cursor.Year(), m.cursor.Month(), m.cursor.Day(), 0, 0, 0, 0, time.UTC)
    isRangeDay = isInRange(dayDate, anchorUTC, cursorUTC)
}
```

Note: All three dates are constructed in UTC using year/month/day components only. This matches the existing `isHighlighted` pattern (line 219) and avoids timezone-dependent comparison bugs where `time.Local` midnight differs from UTC midnight as absolute instants.

Update the style switch to insert `isRangeDay` between `isToday` and `isHighlighted`:

```go
switch {
case isCursor && isToday:
    dayStr = st.cursorToday.Render(dayStr)
case isCursor:
    dayStr = st.cursor.Render(dayStr)
case isToday:
    dayStr = st.today.Render(dayStr)
case isRangeDay:
    dayStr = st.rangeDay.Render(dayStr)
case isHighlighted:
    dayStr = st.highlight.Render(dayStr)
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./calendar/ -count=1`
Expected: All existing tests pass

- [ ] **Step 6: Commit**

```bash
git add calendar/model.go calendar/view.go
git commit -m "Add rangeDay style and isInRange helper to view rendering"
```

---

### Task 3: Model — Add rangeAnchor, split keybindings, range API

**Files:**
- Modify: `calendar/model.go:19-32` (Model struct)
- Modify: `calendar/model.go:47-54` (public API)
- Modify: `calendar/model.go:108-146` (Update)
- Modify: `calendar/model.go:162-233` (keyMap and defaultKeyMap)
- Modify: `calendar/model.go:235-246` (ShortHelp/FullHelp)
- Test: `calendar/model_test.go`

- [ ] **Step 1: Write failing tests**

Add to `calendar/model_test.go`:

```go
func TestVisualSelectEnter(t *testing.T) {
	cursor := date(2026, time.March, 17)
	m := New(cursor, cursor, DefaultConfig())
	// Press v to anchor
	m = pressKey(m, "v")
	// Navigate forward 5 days
	for range 5 {
		m = pressKey(m, "l")
	}
	// Press Enter to confirm range
	updated, cmd := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.Selected() {
		t.Error("expected Selected true")
	}
	if !m.InRange() {
		t.Error("expected InRange true")
	}
	if m.RangeStart() != cursor {
		t.Errorf("RangeStart = %v, want %v", m.RangeStart(), cursor)
	}
	want := date(2026, time.March, 22)
	if m.RangeEnd() != want {
		t.Errorf("RangeEnd = %v, want %v", m.RangeEnd(), want)
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestVisualSelectCancel(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressKey(m, "v")
	m = pressKey(m, "l")
	// Esc should cancel range, not quit
	updated, cmd := m.Update(specialMsg(tea.KeyEscape))
	m = updated.(Model)
	if m.IsQuit() {
		t.Error("expected IsQuit false after Esc in range mode")
	}
	if m.InRange() {
		t.Error("expected InRange false after cancel")
	}
	if cmd != nil {
		t.Error("expected no quit command")
	}
	// Second Esc should quit
	updated, cmd = m.Update(specialMsg(tea.KeyEscape))
	m = updated.(Model)
	if !m.IsQuit() {
		t.Error("expected IsQuit true after second Esc")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestVisualSelectReanchor(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressKey(m, "v")
	m = pressKey(m, "l")
	m = pressKey(m, "l")
	// Press v again to reanchor at current cursor
	m = pressKey(m, "v")
	m = pressKey(m, "l")
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	// Anchor should be March 19 (moved 2 then reanchored), end March 20
	if m.RangeStart() != date(2026, time.March, 19) {
		t.Errorf("RangeStart = %v, want March 19", m.RangeStart())
	}
	if m.RangeEnd() != date(2026, time.March, 20) {
		t.Errorf("RangeEnd = %v, want March 20", m.RangeEnd())
	}
}

func TestRangeReverseOrder(t *testing.T) {
	m := New(date(2026, time.March, 20), date(2026, time.March, 17), DefaultConfig())
	m = pressKey(m, "v")
	// Navigate backward
	for range 5 {
		m = pressKey(m, "h")
	}
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.InRange() {
		t.Error("expected InRange true")
	}
	// RangeStart should be the earlier date regardless of direction
	if m.RangeStart() != date(2026, time.March, 15) {
		t.Errorf("RangeStart = %v, want March 15", m.RangeStart())
	}
	if m.RangeEnd() != date(2026, time.March, 20) {
		t.Errorf("RangeEnd = %v, want March 20", m.RangeEnd())
	}
}

func TestEnterWithoutRange(t *testing.T) {
	cursor := date(2026, time.April, 15)
	m := New(cursor, date(2026, time.March, 17), DefaultConfig())
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.Selected() {
		t.Error("expected Selected true")
	}
	if m.InRange() {
		t.Error("expected InRange false when no v pressed")
	}
}

func TestSameDayRange(t *testing.T) {
	cursor := date(2026, time.March, 17)
	m := New(cursor, cursor, DefaultConfig())
	m = pressKey(m, "v")
	// Press Enter immediately without moving
	updated, _ := m.Update(specialMsg(tea.KeyEnter))
	m = updated.(Model)
	if !m.Selected() {
		t.Error("expected Selected true")
	}
	if m.InRange() {
		t.Error("expected InRange false for same-day")
	}
}

func TestCtrlCInRangeMode(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	m = pressKey(m, "v")
	m = pressKey(m, "l")
	// ctrl+c should force quit even in range mode
	updated, cmd := m.Update(specialMsg(tea.KeyCtrlC))
	m = updated.(Model)
	if !m.IsQuit() {
		t.Error("expected IsQuit true on ctrl+c")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./calendar/ -count=1 -run "TestVisualSelect|TestRange|TestSameDay|TestCtrlC|TestEnterWithout"`
Expected: FAIL — `InRange`, `RangeStart`, `RangeEnd` undefined; `pressKey(m, "v")` no effect

- [ ] **Step 3: Add rangeAnchor field to Model**

In `calendar/model.go`, add after `selected bool` (line 23):

```go
rangeAnchor *time.Time
```

- [ ] **Step 4: Add public API methods**

After `Cursor()` (line 54), add:

```go
// InRange reports whether the user confirmed a date range selection.
func (m Model) InRange() bool {
	return m.selected && m.rangeAnchor != nil && !m.rangeAnchor.Equal(m.cursor)
}

// RangeStart returns the earlier date of the range.
// Returns the zero time if InRange() is false.
func (m Model) RangeStart() time.Time {
	if !m.InRange() {
		return time.Time{}
	}
	if m.rangeAnchor.Before(m.cursor) {
		return *m.rangeAnchor
	}
	return m.cursor
}

// RangeEnd returns the later date of the range.
// Returns the zero time if InRange() is false.
func (m Model) RangeEnd() time.Time {
	if !m.InRange() {
		return time.Time{}
	}
	if m.rangeAnchor.After(m.cursor) {
		return *m.rangeAnchor
	}
	return m.cursor
}
```

- [ ] **Step 5: Split Quit binding, add VisualSelect and ForceQuit**

In the `keyMap` struct, replace `Quit key.Binding` with:

```go
VisualSelect key.Binding
Quit         key.Binding
ForceQuit    key.Binding
```

In `defaultKeyMap()`, replace the Quit binding with:

```go
VisualSelect: key.NewBinding(
    key.WithKeys("v"),
    key.WithHelp("v", "range"),
),
Quit: key.NewBinding(
    key.WithKeys("q", "esc"),
    key.WithHelp("q/esc", "quit"),
),
ForceQuit: key.NewBinding(
    key.WithKeys("ctrl+c"),
),
```

Update `ShortHelp`:
```go
func (k keyMap) ShortHelp() []key.Binding {
    return []key.Binding{k.Left, k.Right, k.VisualSelect, k.Select, k.Quit, k.ToggleHelp}
}
```

Update `FullHelp`:
```go
func (k keyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{
        {k.Left, k.Right, k.Up, k.Down},
        {k.PrevMonth, k.NextMonth, k.PrevYear, k.NextYear},
        {k.Today, k.ToggleWeeks},
        {k.VisualSelect, k.Select, k.Quit},
    }
}
```

- [ ] **Step 6: Update Update() for range mode logic**

Inside the existing `Update()` method, replace ONLY the inner `switch { case key.Matches... }` block within the `case tea.KeyMsg:` arm. Keep the outer `switch msg := msg.(type)` and the `case tea.WindowSizeMsg:` handler unchanged. The new inner switch:

```go
case tea.KeyMsg:
    switch {
    case key.Matches(msg, m.keys.ForceQuit):
        m.quit = true
        return m, tea.Quit
    case key.Matches(msg, m.keys.VisualSelect):
        anchor := m.cursor
        m.rangeAnchor = &anchor
    case key.Matches(msg, m.keys.Select):
        m.selected = true
        return m, tea.Quit
    case key.Matches(msg, m.keys.Quit):
        if m.rangeAnchor != nil {
            m.rangeAnchor = nil
        } else {
            m.quit = true
            return m, tea.Quit
        }
    case key.Matches(msg, m.keys.Left):
        m.cursor = m.cursor.AddDate(0, 0, -1)
    case key.Matches(msg, m.keys.Right):
        m.cursor = m.cursor.AddDate(0, 0, 1)
    case key.Matches(msg, m.keys.Up):
        m.cursor = m.cursor.AddDate(0, 0, -7)
    case key.Matches(msg, m.keys.Down):
        m.cursor = m.cursor.AddDate(0, 0, 7)
    case key.Matches(msg, m.keys.PrevMonth):
        m.cursor = shiftDate(m.cursor, 0, -1)
    case key.Matches(msg, m.keys.NextMonth):
        m.cursor = shiftDate(m.cursor, 0, 1)
    case key.Matches(msg, m.keys.PrevYear):
        m.cursor = shiftDate(m.cursor, -1, 0)
    case key.Matches(msg, m.keys.NextYear):
        m.cursor = shiftDate(m.cursor, 1, 0)
    case key.Matches(msg, m.keys.Today):
        m.cursor = m.today
    case key.Matches(msg, m.keys.ToggleWeeks):
        m.showWeekNumbers = !m.showWeekNumbers
    case key.Matches(msg, m.keys.ToggleHelp):
        m.showHelp = !m.showHelp
    }
```

- [ ] **Step 7: Run tests**

Run: `go test ./calendar/ -count=1 -v`
Expected: All tests pass including new range tests

- [ ] **Step 8: Commit**

```bash
git add calendar/model.go calendar/model_test.go
git commit -m "Add range selection: v to anchor, Enter to confirm, Esc to cancel"
```

---

### Task 4: View test — range rendering

**Files:**
- Test: `calendar/view_test.go`

- [ ] **Step 1: Write range rendering test**

Add to `calendar/view_test.go`:

```go
func TestRangeRendering(t *testing.T) {
	cfg := DefaultConfig()
	cursor := date(2026, time.March, 17)
	m := New(cursor, cursor, cfg)
	// Enter range mode: anchor at 17, move to 20
	m = pressKey(m, "v")
	for range 3 {
		m = pressKey(m, "l")
	}
	output := m.View()
	// The output should contain styled range days (18, 19 should have ANSI codes)
	// Day 17 was the anchor but cursor moved, so 17 is a range day too
	// Day 20 is the cursor (different style)
	// Verify the output is different from non-range rendering
	noRange := New(cursor, cursor, cfg)
	noRange = pressKey(noRange, "l")
	noRange = pressKey(noRange, "l")
	noRange = pressKey(noRange, "l")
	noRangeOutput := noRange.View()
	if output == noRangeOutput {
		t.Error("expected range rendering to differ from non-range rendering")
	}
}

func TestRangeRenderingMultiMonth(t *testing.T) {
	cfg := DefaultConfig()
	cursor := date(2026, time.March, 28)
	m := New(cursor, cursor, cfg, WithMonths(3))
	// Anchor at March 28, move into April
	m = pressKey(m, "v")
	for range 5 {
		m = pressKey(m, "l")
	}
	output := m.View()
	// Output should contain both March and April
	if !strings.Contains(output, "March 2026") {
		t.Error("expected March in output")
	}
	if !strings.Contains(output, "April 2026") {
		t.Error("expected April in output")
	}
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./calendar/ -count=1 -run "TestRangeRendering"`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add calendar/view_test.go
git commit -m "Add range rendering tests"
```

---

### Task 5: CLI — range output in runCalendar

**Files:**
- Modify: `cmd/wen/main.go:297-303` (runCalendar output)

- [ ] **Step 1: Update runCalendar output logic**

In `cmd/wen/main.go`, replace lines 301-303:

```go
if cal.Selected() {
    fmt.Fprintln(ctx.w, cal.Cursor().Format(wen.DateLayout))
}
```

with:

```go
if cal.InRange() {
    fmt.Fprintln(ctx.w, cal.RangeStart().Format(wen.DateLayout))
    fmt.Fprintln(ctx.w, cal.RangeEnd().Format(wen.DateLayout))
} else if cal.Selected() {
    fmt.Fprintln(ctx.w, cal.Cursor().Format(wen.DateLayout))
}
```

- [ ] **Step 2: Run all tests**

Run: `go test ./... -count=1`
Expected: All pass

- [ ] **Step 3: Commit**

```bash
git add cmd/wen/main.go
git commit -m "Output two-line date range when calendar range is selected"
```

---

### Task 6: README — document range selection

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Add v to keybindings table**

In the keybindings table (after the `Enter` row), the `v` row should already be missing. Add:

```markdown
| `v` | Start range selection (move cursor, then Enter to confirm) |
```

- [ ] **Step 2: Add range output documentation**

After the "Navigate with vim keys" paragraph, add:

```markdown
For date ranges, press `v` to anchor a start date, navigate to the end date, then press `Enter`. Both dates are printed (one per line):

```bash
# Select a date range interactively
wen cal
# Output:
# 2026-03-21
# 2026-04-02

# Use in scripts
git log --since=$(wen cal | head -1) --until=$(wen cal | tail -1)
```

Press `Esc` or `q` while in range mode to cancel and return to normal selection. Press `v` again to move the anchor.
```

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "Document range selection in README"
```

---

### Task 7: Final verification

- [ ] **Step 1: Run full test suite**

Run: `go test ./... -count=1 -cover`
Expected: All pass. Calendar coverage should increase.

- [ ] **Step 2: Smoke test manually**

```bash
go build -o wen ./cmd/wen
./wen cal              # press v, navigate, Enter — should print two dates
./wen cal              # press Enter without v — should print one date
./wen cal              # press v, Esc, q — should quit with no output
./wen cal -3           # press v, navigate across months, Enter — should work
```

- [ ] **Step 3: Commit any fixes**

If smoke testing reveals issues, fix and commit.
