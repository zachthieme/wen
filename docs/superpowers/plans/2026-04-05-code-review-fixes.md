# Code Review Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close four code-review gaps: AST interface coverage, highlight warning returns, Model/RowModel deduplication via embedded baseModel, and LoadConfig test coverage.

**Architecture:** Four independent-to-semi-dependent changes. Tasks 1, 2, and 5 are independent. Task 3 (baseModel) depends on Task 2 (highlight warnings). Task 4 (CLI warnings) depends on Task 3.

**Tech Stack:** Go 1.25, Bubble Tea, fsnotify, golangci-lint

---

### Task 1: AST compile-time assertions and coverage

**Files:**
- Modify: `ast.go:87-98` (add compile-time assertions above marker methods)
- Create: `ast_test.go`

- [ ] **Step 1: Write `ast_test.go`**

```go
package wen

import "testing"

func TestDateExprInterface(t *testing.T) {
	t.Parallel()
	nodes := []dateExpr{
		&relativeDayExpr{},
		&modWeekdayExpr{},
		&relativeOffsetExpr{},
		&countedWeekdayExpr{},
		&ordinalWeekdayExpr{},
		&lastWeekdayInMonthExpr{},
		&absoluteDateExpr{},
		&periodRefExpr{},
		&boundaryExpr{},
		&multiDateExpr{},
		&withTimeExpr{},
	}
	for _, n := range nodes {
		n.dateExpr()
	}
}
```

- [ ] **Step 2: Run test to verify it passes**

Run: `go test -run TestDateExprInterface -v .`
Expected: PASS

- [ ] **Step 3: Add compile-time assertions to `ast.go`**

Insert between the `withTimeExpr` struct definition (line 86) and the marker methods (line 88):

```go
// Compile-time interface satisfaction checks.
var (
	_ dateExpr = (*relativeDayExpr)(nil)
	_ dateExpr = (*modWeekdayExpr)(nil)
	_ dateExpr = (*relativeOffsetExpr)(nil)
	_ dateExpr = (*countedWeekdayExpr)(nil)
	_ dateExpr = (*ordinalWeekdayExpr)(nil)
	_ dateExpr = (*lastWeekdayInMonthExpr)(nil)
	_ dateExpr = (*absoluteDateExpr)(nil)
	_ dateExpr = (*periodRefExpr)(nil)
	_ dateExpr = (*boundaryExpr)(nil)
	_ dateExpr = (*multiDateExpr)(nil)
	_ dateExpr = (*withTimeExpr)(nil)
)
```

- [ ] **Step 4: Run full test suite**

Run: `make check`
Expected: All tests pass, no lint issues.

- [ ] **Step 5: Commit**

```bash
git add ast.go ast_test.go
git commit -m "test: add AST interface coverage and compile-time assertions"
```

---

### Task 2: LoadHighlightedDates warning returns

**Files:**
- Modify: `calendar/highlight.go:29-57` (change signature, add warnings)
- Modify: `calendar/highlight.go:65-70` (WithHighlightSource caller)
- Modify: `calendar/row_model.go:53-57` (WithRowHighlightSource caller)
- Modify: `calendar/watcher.go:91` (watchLoop caller)
- Modify: `calendar/highlight_test.go` (update all LoadHighlightedDates assertions)

- [ ] **Step 1: Update `LoadHighlightedDates` tests to expect `(map, []string)` return**

Replace the entire `TestLoadHighlightedDates` function in `calendar/highlight_test.go`:

```go
func TestLoadHighlightedDates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	t.Run("valid file", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(dir, "dates.json")
		if err := os.WriteFile(path, []byte(`["2026-03-25", "2026-04-01"]`), 0644); err != nil {
			t.Fatal(err)
		}

		dates, warnings := LoadHighlightedDates(path)
		if len(warnings) != 0 {
			t.Errorf("expected no warnings, got %v", warnings)
		}
		if len(dates) != 2 {
			t.Fatalf("expected 2 dates, got %d", len(dates))
		}
		if !dates[time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)] {
			t.Error("expected 2026-03-25 to be highlighted")
		}
		if !dates[time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)] {
			t.Error("expected 2026-04-01 to be highlighted")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		missingPath := filepath.Join(dir, "nonexistent.json")
		dates, warnings := LoadHighlightedDates(missingPath)
		if dates != nil {
			t.Error("expected nil for missing file")
		}
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
		}
		if !strings.Contains(warnings[0], "not found") {
			t.Errorf("expected 'not found' warning, got %q", warnings[0])
		}
	})

	t.Run("malformed JSON", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(dir, "bad.json")
		if err := os.WriteFile(path, []byte(`not json`), 0644); err != nil {
			t.Fatal(err)
		}
		dates, warnings := LoadHighlightedDates(path)
		if dates != nil {
			t.Error("expected nil for malformed JSON")
		}
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
		}
		if !strings.Contains(warnings[0], "not valid JSON") {
			t.Errorf("expected 'not valid JSON' warning, got %q", warnings[0])
		}
	})

	t.Run("empty path", func(t *testing.T) {
		t.Parallel()
		dates, warnings := LoadHighlightedDates("")
		if dates != nil {
			t.Error("expected nil for empty path")
		}
		if len(warnings) != 0 {
			t.Errorf("expected no warnings for empty path, got %v", warnings)
		}
	})

	t.Run("invalid dates produce per-entry warnings", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(dir, "mixed.json")
		if err := os.WriteFile(path, []byte(`["2026-03-25", "not-a-date", "2026-04-01"]`), 0644); err != nil {
			t.Fatal(err)
		}
		dates, warnings := LoadHighlightedDates(path)
		if len(dates) != 2 {
			t.Fatalf("expected 2 valid dates, got %d", len(dates))
		}
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning for invalid date, got %d: %v", len(warnings), warnings)
		}
		if !strings.Contains(warnings[0], "not-a-date") {
			t.Errorf("expected warning to mention 'not-a-date', got %q", warnings[0])
		}
	})
}
```

Also add `"strings"` to the import block if not already present.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -run TestLoadHighlightedDates -v ./calendar/`
Expected: FAIL — `LoadHighlightedDates` returns 1 value, tests expect 2.

- [ ] **Step 3: Update `LoadHighlightedDates` in `calendar/highlight.go`**

Replace the function (lines 29–57) with:

```go
// LoadHighlightedDates loads highlighted dates from a JSON file.
// The file should contain a JSON array of yyyy-mm-dd strings.
// Returns nil with no warnings for an empty path (not configured).
// Returns nil with a warning if the file is missing or malformed.
// Individual unparseable date entries produce per-entry warnings.
func LoadHighlightedDates(path string) (map[time.Time]bool, []string) {
	if path == "" {
		return nil, nil
	}

	path = expandTilde(path)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, []string{fmt.Sprintf("highlight file not found: %s", path)}
	}

	var dates []string
	if err := json.Unmarshal(data, &dates); err != nil {
		return nil, []string{fmt.Sprintf("highlight file is not valid JSON: %s", path)}
	}

	var warnings []string
	result := make(map[time.Time]bool, len(dates))
	for _, s := range dates {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("skipping invalid date %q in %s", s, path))
			continue
		}
		result[t] = true
	}
	if len(result) == 0 {
		return nil, warnings
	}
	return result, warnings
}
```

Add `"fmt"` to the import block in `highlight.go`.

- [ ] **Step 4: Update callers to accept second return value**

In `calendar/highlight.go` line 68 (`WithHighlightSource`), change:

```go
m.highlightedDates = LoadHighlightedDates(m.highlightPath)
```

to:

```go
m.highlightedDates, _ = LoadHighlightedDates(m.highlightPath)
```

In `calendar/row_model.go` line 56 (`WithRowHighlightSource`), change:

```go
m.highlightedDates = LoadHighlightedDates(m.highlightPath)
```

to:

```go
m.highlightedDates, _ = LoadHighlightedDates(m.highlightPath)
```

In `calendar/watcher.go` line 91 (`watchLoop`), change:

```go
dates := LoadHighlightedDates(path)
```

to:

```go
dates, _ := LoadHighlightedDates(path)
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `make check`
Expected: All tests pass, no lint issues.

- [ ] **Step 6: Commit**

```bash
git add calendar/highlight.go calendar/highlight_test.go calendar/row_model.go calendar/watcher.go
git commit -m "feat: return warnings from LoadHighlightedDates instead of failing silently"
```

---

### Task 3: Extract baseModel from Model and RowModel

**Files:**
- Create: `calendar/base_model.go`
- Create: `calendar/base_model_test.go`
- Modify: `calendar/model.go` (slim down to embed baseModel)
- Modify: `calendar/row_model.go` (slim down to embed baseModel)
- Modify: `calendar/highlight.go:65-70` (wire up warnings)

- [ ] **Step 1: Create `calendar/base_model.go` with struct and all methods**

```go
package calendar

import (
	"time"

	"github.com/zachthieme/wen"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
)

// baseModel holds state shared between Model and RowModel.
type baseModel struct {
	cursor           time.Time
	today            time.Time
	quit             bool
	selected         bool
	rangeAnchor      *time.Time
	highlightedDates map[time.Time]bool
	highlightPath    string
	activeWatcher    *fsnotify.Watcher
	config           Config
	help             help.Model
	styles           resolvedStyles
	showHelp         bool
	julian           bool
	printMode        bool
	dayFmt           dayFormat
	termWidth        int
	termHeight       int
	warnings         []string
}

// resolvedStyles holds pre-computed lipgloss styles for all calendar elements.
type resolvedStyles struct {
	cursor      lipgloss.Style
	cursorToday lipgloss.Style
	today       lipgloss.Style
	highlight   lipgloss.Style
	rangeDay    lipgloss.Style
	title       lipgloss.Style
	weekNum     lipgloss.Style
	dayHeader   lipgloss.Style
	helpBar     lipgloss.Style
}

// Warnings returns any warnings collected during model construction.
func (b baseModel) Warnings() []string { return b.warnings }

// IsQuit reports whether the user quit without selecting.
func (b baseModel) IsQuit() bool { return b.quit }

// Selected reports whether the user selected a date with Enter.
func (b baseModel) Selected() bool { return b.selected }

// Cursor returns the currently selected date.
func (b baseModel) Cursor() time.Time { return b.cursor }

// InRange reports whether the user confirmed a multi-day range selection.
func (b baseModel) InRange() bool {
	return b.selected && b.rangeAnchor != nil && !b.rangeAnchor.Equal(b.cursor)
}

// RangeStart returns the earlier date of the confirmed range, or zero if no range.
func (b baseModel) RangeStart() time.Time {
	if !b.InRange() {
		return time.Time{}
	}
	if b.rangeAnchor.Before(b.cursor) {
		return *b.rangeAnchor
	}
	return b.cursor
}

// RangeEnd returns the later date of the confirmed range, or zero if no range.
func (b baseModel) RangeEnd() time.Time {
	if !b.InRange() {
		return time.Time{}
	}
	if b.rangeAnchor.After(b.cursor) {
		return *b.rangeAnchor
	}
	return b.cursor
}

// initCmds returns the initial tea.Cmds shared by both models:
// a midnight tick and, if configured, a file watcher.
func (b baseModel) initCmds() []tea.Cmd {
	cmds := []tea.Cmd{scheduleMidnightTick(b.today)}
	if b.highlightPath != "" {
		cmds = append(cmds, startFileWatcher(b.highlightPath))
	}
	return cmds
}

// handleMsg handles messages common to both Model and RowModel.
// Returns (cmd, true) if the message was handled.
func (b *baseModel) handleMsg(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.help.Width = msg.Width
		b.termWidth = msg.Width
		b.termHeight = msg.Height
		return nil, true
	case watcherErrMsg:
		return nil, true
	case midnightTickMsg:
		now := time.Now()
		b.today = wen.TruncateDay(now)
		return scheduleMidnightTick(now), true
	case highlightChangedMsg:
		b.highlightedDates = msg.dates
		b.activeWatcher = msg.watcher
		return waitForNextChange(msg.watcher, msg.path), true
	}
	return nil, false
}

// closeWatcher shuts down the active file watcher if one is running.
func (b *baseModel) closeWatcher() {
	if b.activeWatcher != nil {
		_ = b.activeWatcher.Close()
		b.activeWatcher = nil
	}
}

// doQuit marks the model as quit, cleans up resources, and returns tea.Quit.
func (b *baseModel) doQuit() tea.Cmd {
	b.quit = true
	b.closeWatcher()
	return tea.Quit
}

// doSelect marks the model as having a selection, cleans up, and returns tea.Quit.
func (b *baseModel) doSelect() tea.Cmd {
	b.selected = true
	b.closeWatcher()
	return tea.Quit
}

// doVisualSelect sets the range anchor to the current cursor position.
func (b *baseModel) doVisualSelect() {
	anchor := b.cursor
	b.rangeAnchor = &anchor
}

// cancelRange clears the range anchor if set. Returns true if a range was active.
func (b *baseModel) cancelRange() bool {
	if b.rangeAnchor != nil {
		b.rangeAnchor = nil
		return true
	}
	return false
}
```

- [ ] **Step 2: Create `calendar/base_model_test.go`**

```go
package calendar

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleMsg_WindowSizeMsg(t *testing.T) {
	t.Parallel()
	b := baseModel{}
	cmd, handled := b.handleMsg(tea.WindowSizeMsg{Width: 80, Height: 24})
	if !handled {
		t.Fatal("expected WindowSizeMsg to be handled")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	if b.termWidth != 80 || b.termHeight != 24 {
		t.Errorf("got %dx%d, want 80x24", b.termWidth, b.termHeight)
	}
}

func TestHandleMsg_WatcherErrMsg(t *testing.T) {
	t.Parallel()
	b := baseModel{}
	cmd, handled := b.handleMsg(watcherErrMsg{err: nil})
	if !handled {
		t.Fatal("expected watcherErrMsg to be handled")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestHandleMsg_Unhandled(t *testing.T) {
	t.Parallel()
	b := baseModel{}
	_, handled := b.handleMsg(tea.KeyMsg{})
	if handled {
		t.Error("expected KeyMsg to not be handled by baseModel")
	}
}

func TestCloseWatcher_Nil(t *testing.T) {
	t.Parallel()
	b := baseModel{}
	b.closeWatcher() // must not panic
	if b.activeWatcher != nil {
		t.Error("expected nil watcher after close")
	}
}

func TestDoQuit(t *testing.T) {
	t.Parallel()
	b := baseModel{}
	cmd := b.doQuit()
	if !b.quit {
		t.Error("expected quit to be true")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestDoSelect(t *testing.T) {
	t.Parallel()
	b := baseModel{}
	cmd := b.doSelect()
	if !b.selected {
		t.Error("expected selected to be true")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestDoVisualSelect(t *testing.T) {
	t.Parallel()
	cursor := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	b := baseModel{cursor: cursor}
	b.doVisualSelect()
	if b.rangeAnchor == nil {
		t.Fatal("expected rangeAnchor to be set")
	}
	if !b.rangeAnchor.Equal(cursor) {
		t.Error("expected rangeAnchor to equal cursor")
	}
}

func TestCancelRange(t *testing.T) {
	t.Parallel()

	t.Run("with active range", func(t *testing.T) {
		t.Parallel()
		anchor := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		b := baseModel{rangeAnchor: &anchor}
		if !b.cancelRange() {
			t.Error("expected true when range was active")
		}
		if b.rangeAnchor != nil {
			t.Error("expected rangeAnchor to be nil")
		}
	})

	t.Run("without active range", func(t *testing.T) {
		t.Parallel()
		b := baseModel{}
		if b.cancelRange() {
			t.Error("expected false when no range active")
		}
	})
}

func TestInitCmds(t *testing.T) {
	t.Parallel()

	t.Run("without highlight path", func(t *testing.T) {
		t.Parallel()
		b := baseModel{today: time.Date(2026, 4, 5, 0, 0, 0, 0, time.Local)}
		cmds := b.initCmds()
		if len(cmds) != 1 {
			t.Errorf("expected 1 cmd, got %d", len(cmds))
		}
	})

	t.Run("with highlight path", func(t *testing.T) {
		t.Parallel()
		b := baseModel{
			today:         time.Date(2026, 4, 5, 0, 0, 0, 0, time.Local),
			highlightPath: "/some/path.json",
		}
		cmds := b.initCmds()
		if len(cmds) != 2 {
			t.Errorf("expected 2 cmds, got %d", len(cmds))
		}
	})
}

func TestWarnings(t *testing.T) {
	t.Parallel()
	b := baseModel{warnings: []string{"w1", "w2"}}
	got := b.Warnings()
	if len(got) != 2 || got[0] != "w1" || got[1] != "w2" {
		t.Errorf("got %v, want [w1 w2]", got)
	}
}
```

- [ ] **Step 3: Verify base_model_test.go compiles and passes**

Run: `go test -run 'TestHandleMsg|TestCloseWatcher|TestDoQuit|TestDoSelect|TestDoVisualSelect|TestCancelRange|TestInitCmds|TestWarnings' -v ./calendar/`
Expected: PASS — all new tests pass against the new code.

- [ ] **Step 4: Refactor `calendar/model.go` to embed baseModel**

Replace the entire `model.go` file with the slimmed-down version. Key changes:
- Remove `resolvedStyles` struct (moved to base_model.go)
- `Model` embeds `baseModel`, keeps only `weekNumPos`, `months`, `keys`
- Remove duplicate accessors (IsQuit, Selected, Cursor, InRange, RangeStart, RangeEnd)
- `New()` initializes `baseModel` via named field
- `Init()` delegates to `initCmds()`
- `Update()` delegates to `handleMsg()` then uses `doQuit()`, `doSelect()`, `doVisualSelect()`, `cancelRange()`

New `calendar/model.go`:

```go
// Package calendar provides an interactive terminal calendar UI built on Bubble Tea.
package calendar

import (
	"time"

	"github.com/zachthieme/wen"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Model holds the state for the interactive calendar TUI.
type Model struct {
	baseModel
	weekNumPos WeekNumPos
	months     int
	keys       keyMap
}

// ModelOption configures optional Model properties.
type ModelOption func(*Model)

// WithHighlightedDates sets dates to be visually highlighted in the calendar.
// Clears any highlight source path, disabling file watching.
func WithHighlightedDates(dates map[time.Time]bool) ModelOption {
	return func(m *Model) {
		m.highlightedDates = dates
		m.highlightPath = ""
	}
}

// WithMonths sets the number of months to display side by side.
func WithMonths(n int) ModelOption {
	return func(m *Model) {
		if n < 1 {
			n = 1
		}
		m.months = n
	}
}

// WithJulian enables Julian day-of-year numbering.
func WithJulian(on bool) ModelOption {
	return func(m *Model) {
		m.julian = on
	}
}

// WithPrintMode enables non-interactive print mode (suppresses cursor styling).
func WithPrintMode(on bool) ModelOption {
	return func(m *Model) {
		m.printMode = on
	}
}

// New creates a calendar Model with the given cursor position, today's date, and configuration.
func New(cursor, today time.Time, cfg Config, opts ...ModelOption) Model {
	colors := cfg.ResolvedColors()
	m := Model{
		baseModel: baseModel{
			cursor: wen.TruncateDay(cursor),
			today:  wen.TruncateDay(today),
			config: cfg,
			help:   newHelpModel(colors),
			styles: buildStyles(colors),
		},
		weekNumPos: parseWeekNumPos(cfg.ShowWeekNumbers),
		months:     1,
		keys:       defaultKeyMap(),
	}
	for _, opt := range opts {
		opt(&m)
	}
	m.dayFmt = dayFormatFor(m.julian)
	return m
}

// midnightTickMsg is sent when the clock crosses midnight, triggering a
// refresh of the "today" highlight.
type midnightTickMsg struct{}

// scheduleMidnightTick returns a tea.Cmd that fires at the next midnight.
func scheduleMidnightTick(now time.Time) tea.Cmd {
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return tea.Tick(time.Until(next), func(_ time.Time) tea.Msg {
		return midnightTickMsg{}
	})
}

// Init schedules the midnight tick (to refresh the "today" highlight at midnight)
// and, if a highlight source path is configured, starts an fsnotify file watcher.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.initCmds()...)
}

// Update handles input messages and updates model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, handled := m.handleMsg(msg); handled {
		return m, cmd
	}
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, m.keys.ForceQuit):
			return m, m.doQuit()
		case key.Matches(msg, m.keys.VisualSelect):
			m.doVisualSelect()
			return m, nil
		case key.Matches(msg, m.keys.Select):
			return m, m.doSelect()
		case key.Matches(msg, m.keys.Quit):
			if m.cancelRange() {
				return m, nil
			}
			return m, m.doQuit()
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
			switch m.weekNumPos {
			case WeekNumOff:
				m.weekNumPos = WeekNumLeft
			case WeekNumLeft:
				m.weekNumPos = WeekNumRight
			case WeekNumRight:
				m.weekNumPos = WeekNumOff
			}
		case key.Matches(msg, m.keys.ToggleJulian):
			m.julian = !m.julian
			m.dayFmt = dayFormatFor(m.julian)
		case key.Matches(msg, m.keys.ToggleHelp):
			m.showHelp = !m.showHelp
		}
	}
	return m, nil
}

func shiftDate(t time.Time, years, months int) time.Time {
	y, m, d := t.Date()
	target := time.Date(y+years, m+time.Month(months), 1, 0, 0, 0, 0, t.Location())
	maxDay := wen.DaysIn(target.Year(), target.Month(), t.Location())
	if d > maxDay {
		d = maxDay
	}
	return time.Date(target.Year(), target.Month(), d, 0, 0, 0, 0, t.Location())
}

type keyMap struct {
	Left         key.Binding
	Right        key.Binding
	Up           key.Binding
	Down         key.Binding
	PrevMonth    key.Binding
	NextMonth    key.Binding
	PrevYear     key.Binding
	NextYear     key.Binding
	Today        key.Binding
	ToggleWeeks  key.Binding
	ToggleJulian key.Binding
	ToggleHelp   key.Binding
	VisualSelect key.Binding
	Select       key.Binding
	Quit         key.Binding
	ForceQuit    key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/←", "prev day"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/→", "next day"),
		),
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "prev week"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "next week"),
		),
		PrevMonth: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "prev month"),
		),
		NextMonth: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "next month"),
		),
		PrevYear: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "prev year"),
		),
		NextYear: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "next year"),
		),
		Today: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "today"),
		),
		ToggleWeeks: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "weeks"),
		),
		ToggleJulian: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("J", "julian"),
		),
		ToggleHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		VisualSelect: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "range"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "force quit"),
		),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.VisualSelect, k.Select, k.Quit, k.ToggleHelp}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.Up, k.Down},
		{k.PrevMonth, k.NextMonth, k.PrevYear, k.NextYear},
		{k.Today, k.ToggleWeeks, k.ToggleJulian},
		{k.VisualSelect, k.Select, k.Quit},
	}
}
```

- [ ] **Step 5: Run tests to verify Model refactor preserves behavior**

Run: `go test -race -count=1 ./calendar/`
Expected: PASS — all existing calendar tests still pass.

- [ ] **Step 6: Refactor `calendar/row_model.go` to embed baseModel**

Replace the entire `row_model.go` file. Key changes:
- `RowModel` embeds `baseModel`, keeps only `keys`
- Remove duplicate accessors
- `Init()` delegates to `initCmds()`
- `Update()` delegates to `handleMsg()` then uses shared action helpers

New `calendar/row_model.go`:

```go
package calendar

import (
	"strings"
	"time"

	"github.com/zachthieme/wen"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RowModel holds the state for the interactive strip calendar TUI.
type RowModel struct {
	baseModel
	keys rowKeyMap
}

// RowModelOption configures optional RowModel properties.
type RowModelOption func(*RowModel)

// WithRowHighlightedDates sets dates to be visually highlighted in the row calendar.
// Clears any highlight source path, disabling file watching.
func WithRowHighlightedDates(dates map[time.Time]bool) RowModelOption {
	return func(m *RowModel) {
		m.highlightedDates = dates
		m.highlightPath = ""
	}
}

// WithRowHighlightSource sets the path to a JSON file of dates to highlight.
// It expands ~ to the user's home directory, performs the initial load, and
// enables file watching when Init() runs.
func WithRowHighlightSource(path string) RowModelOption {
	return func(m *RowModel) {
		m.highlightPath = expandTilde(path)
		dates, warnings := LoadHighlightedDates(m.highlightPath)
		m.highlightedDates = dates
		m.warnings = append(m.warnings, warnings...)
	}
}

// WithRowJulian enables Julian day-of-year numbering in the row calendar.
func WithRowJulian(on bool) RowModelOption {
	return func(m *RowModel) {
		m.julian = on
	}
}

// WithRowPrintMode enables non-interactive print mode (suppresses cursor styling).
func WithRowPrintMode(on bool) RowModelOption {
	return func(m *RowModel) {
		m.printMode = on
	}
}

// WithTermWidth returns a copy of the model with the terminal width set.
// Used in print mode where no WindowSizeMsg is received.
func (m RowModel) WithTermWidth(w int) RowModel {
	m.termWidth = w
	return m
}

// NewRow creates a RowModel with the given cursor position, today's date, and configuration.
func NewRow(cursor, today time.Time, cfg Config, opts ...RowModelOption) RowModel {
	colors := cfg.ResolvedColors()
	m := RowModel{
		baseModel: baseModel{
			cursor: wen.TruncateDay(cursor),
			today:  wen.TruncateDay(today),
			config: cfg,
			help:   newHelpModel(colors),
			styles: buildStyles(colors),
		},
		keys: defaultRowKeyMap(),
	}
	// Strip Underline from row styles. lipgloss renders Underline per-character
	// (each char gets its own ANSI open/close), which causes terminals like mosh
	// to miscalculate cursor positions and misalign the strip columns.
	m.styles.today = m.styles.today.Underline(false)
	m.styles.cursorToday = m.styles.cursorToday.Underline(false)
	m.styles.highlight = m.styles.highlight.Underline(false)
	for _, opt := range opts {
		opt(&m)
	}
	m.dayFmt = dayFormatFor(m.julian)
	return m
}

// Init schedules the midnight tick and, if a highlight source path is configured,
// starts an fsnotify file watcher.
func (m RowModel) Init() tea.Cmd {
	return tea.Batch(m.initCmds()...)
}

// Update handles input messages and updates model state.
func (m RowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, handled := m.handleMsg(msg); handled {
		return m, cmd
	}
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, m.keys.ForceQuit):
			return m, m.doQuit()
		case key.Matches(msg, m.keys.VisualSelect):
			m.doVisualSelect()
			return m, nil
		case key.Matches(msg, m.keys.Select):
			return m, m.doSelect()
		case key.Matches(msg, m.keys.Quit):
			if m.cancelRange() {
				return m, nil
			}
			return m, m.doQuit()
		case key.Matches(msg, m.keys.Left):
			m.cursor = m.cursor.AddDate(0, 0, -1)
		case key.Matches(msg, m.keys.Right):
			m.cursor = m.cursor.AddDate(0, 0, 1)
		case key.Matches(msg, m.keys.PrevMonth):
			m.cursor = shiftDate(m.cursor, 0, -1)
		case key.Matches(msg, m.keys.NextMonth):
			m.cursor = shiftDate(m.cursor, 0, 1)
		case key.Matches(msg, m.keys.Today):
			m.cursor = m.today
		case key.Matches(msg, m.keys.ToggleJulian):
			m.julian = !m.julian
			m.dayFmt = dayFormatFor(m.julian)
		case key.Matches(msg, m.keys.WeekStart):
			m.cursor = weekStartDate(m.cursor, m.config.WeekStartDay)
		case key.Matches(msg, m.keys.WeekEnd):
			m.cursor = weekEndDate(m.cursor, m.config.WeekStartDay)
		case key.Matches(msg, m.keys.MonthStart):
			y, mo, _ := m.cursor.Date()
			m.cursor = time.Date(y, mo, 1, 0, 0, 0, 0, m.cursor.Location())
		case key.Matches(msg, m.keys.MonthEnd):
			y, mo, _ := m.cursor.Date()
			m.cursor = time.Date(y, mo+1, 0, 0, 0, 0, 0, m.cursor.Location())
		case key.Matches(msg, m.keys.ToggleHelp):
			m.showHelp = !m.showHelp
		}
	}
	return m, nil
}

// visibleWindow trims the full strip window to fit within the terminal width,
// centering on the cursor. If the full window fits, it is returned unchanged.
func (m RowModel) visibleWindow(fullStart, fullEnd time.Time) (time.Time, time.Time) {
	availWidth := m.termWidth
	if availWidth <= 0 {
		return fullStart, fullEnd
	}

	totalDays := dayCount(fullStart, fullEnd)
	// Each day cell is cellWidth+1 chars (number + space separator).
	// The prefix occupies prefixWidth chars before the first cell.
	// Total for N days: prefixWidth + N*(cellWidth+1) - 1 (no trailing space).
	// Solving for N: (availWidth - prefixWidth + 1) / (cellWidth + 1)
	cellW := m.dayFmt.cellWidth + 1
	maxDays := (availWidth - m.dayFmt.prefixWidth + 1) / cellW
	if maxDays <= 0 {
		maxDays = 1
	}
	if totalDays <= maxDays {
		return fullStart, fullEnd
	}

	cursorOffset := dayCount(fullStart, m.cursor) - 1 // 0-indexed
	startOffset := max(cursorOffset-maxDays/2, 0)
	if startOffset+maxDays > totalDays {
		startOffset = totalDays - maxDays
	}

	return fullStart.AddDate(0, 0, startOffset), fullStart.AddDate(0, 0, startOffset+maxDays-1)
}

// View produces the strip calendar view string for the model state.
func (m RowModel) View() string {
	year, month, _ := m.cursor.Date()
	loc := m.cursor.Location()
	fullStart, fullEnd := stripWindow(year, month, m.config.WeekStartDay, loc)
	start, end := m.visibleWindow(fullStart, fullEnd)

	var b strings.Builder
	b.WriteString(m.renderStripDayHeaders(start, end))
	b.WriteString("\n")
	b.WriteString(m.renderStripDays(start, end))
	b.WriteString("\n")

	if m.showHelp {
		b.WriteString("\n")
		b.WriteString(m.help.View(m.keys))
		b.WriteString("\n")
	}

	output := b.String()
	if m.termWidth > 0 && m.termHeight > 0 {
		return lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Left, lipgloss.Center, output)
	}
	return output
}

type rowKeyMap struct {
	Left         key.Binding
	Right        key.Binding
	PrevMonth    key.Binding
	NextMonth    key.Binding
	WeekStart    key.Binding
	WeekEnd      key.Binding
	MonthStart   key.Binding
	MonthEnd     key.Binding
	Today        key.Binding
	ToggleJulian key.Binding
	ToggleHelp   key.Binding
	VisualSelect key.Binding
	Select       key.Binding
	Quit         key.Binding
	ForceQuit    key.Binding
}

func defaultRowKeyMap() rowKeyMap {
	return rowKeyMap{
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/\u2190", "prev day"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/\u2192", "next day"),
		),
		PrevMonth: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/\u2191", "prev month"),
		),
		NextMonth: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/\u2193", "next month"),
		),
		WeekStart: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "week start"),
		),
		WeekEnd: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "week end"),
		),
		MonthStart: key.NewBinding(
			key.WithKeys("0"),
			key.WithHelp("0", "month start"),
		),
		MonthEnd: key.NewBinding(
			key.WithKeys("$"),
			key.WithHelp("$", "month end"),
		),
		Today: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "today"),
		),
		ToggleJulian: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("J", "julian"),
		),
		ToggleHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		VisualSelect: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "range"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "force quit"),
		),
	}
}

// ShortHelp returns bindings for the short help view.
func (k rowKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.VisualSelect, k.Select, k.Quit, k.ToggleHelp}
}

// FullHelp returns bindings for the full help view.
func (k rowKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.PrevMonth, k.NextMonth},
		{k.WeekStart, k.WeekEnd, k.MonthStart, k.MonthEnd},
		{k.Today, k.ToggleJulian, k.VisualSelect, k.Select, k.Quit},
	}
}
```

- [ ] **Step 7: Wire up highlight warnings in `calendar/highlight.go`**

Replace the `WithHighlightSource` function (change `_, _` to collect warnings):

```go
// WithHighlightSource sets the path to a JSON file of dates to highlight.
// It expands ~ to the user's home directory, performs the initial load, and
// enables file watching when Init() runs. If both WithHighlightSource and
// WithHighlightedDates are used, the last one applied wins (WithHighlightedDates
// clears the highlight path, disabling file watching).
func WithHighlightSource(path string) ModelOption {
	return func(m *Model) {
		m.highlightPath = expandTilde(path)
		dates, warnings := LoadHighlightedDates(m.highlightPath)
		m.highlightedDates = dates
		m.warnings = append(m.warnings, warnings...)
	}
}
```

- [ ] **Step 8: Update highlight option tests for warnings**

Add a warnings assertion to `TestWithHighlightSourceMissing` in `calendar/highlight_test.go`:

After the existing `m.highlightedDates != nil` check, add:

```go
	if len(m.Warnings()) == 0 {
		t.Error("expected warning for missing highlight file")
	}
```

Add a no-warnings assertion to `TestWithHighlightSource`:

After the existing `m.highlightedDates` check, add:

```go
	if len(m.Warnings()) != 0 {
		t.Errorf("expected no warnings, got %v", m.Warnings())
	}
```

- [ ] **Step 9: Run full test suite**

Run: `make check`
Expected: All tests pass, no lint issues.

- [ ] **Step 10: Commit**

```bash
git add calendar/base_model.go calendar/base_model_test.go calendar/model.go calendar/row_model.go calendar/highlight.go calendar/highlight_test.go
git commit -m "refactor: extract baseModel to deduplicate Model and RowModel"
```

---

### Task 4: Surface highlight warnings in CLI

**Files:**
- Modify: `cmd/wen/cal.go:91` (add warning printing after model creation)
- Modify: `cmd/wen/row.go:45` (add warning printing after model creation)
- Modify: `cmd/wen/main_test.go` (add test for warning output)

- [ ] **Step 1: Write CLI test for highlight warnings on stderr**

Add to `cmd/wen/main_test.go`:

```go
func TestCalHighlightWarningOnStderr(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "cal", "--print", "--highlight-file", "/nonexistent/wen-test-file.json")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	// stdout has the calendar output; we only care about stderr here.
	if err := cmd.Run(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := stderr.String()
	if !strings.Contains(got, "not found") {
		t.Errorf("expected stderr warning about missing highlight file, got %q", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run TestCalHighlightWarningOnStderr -v ./cmd/wen/`
Expected: FAIL — no warning printed to stderr yet.

- [ ] **Step 3: Add warning printing to `cmd/wen/cal.go`**

In `runCalendar`, after `m := calendar.New(...)` (line 91), add:

```go
	for _, w := range m.Warnings() {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}
```

- [ ] **Step 4: Add warning printing to `cmd/wen/row.go`**

In `runRow`, after `m := calendar.NewRow(...)` (line 45), add:

```go
	for _, w := range m.Warnings() {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `make check`
Expected: All tests pass, no lint issues.

- [ ] **Step 6: Commit**

```bash
git add cmd/wen/cal.go cmd/wen/row.go cmd/wen/main_test.go
git commit -m "feat: surface highlight file warnings on stderr in cal and row subcommands"
```

---

### Task 5: LoadConfig and configPath test coverage

**Files:**
- Modify: `calendar/config_test.go` (add tests for LoadConfig and configPath)

- [ ] **Step 1: Write tests for `configPath` and `LoadConfig`**

Add to `calendar/config_test.go`:

```go
func TestConfigPath_XDGSet(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-xdg")
	got, err := configPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/tmp/test-xdg/wen/config.yaml"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestConfigPath_XDGUnset(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	got, err := configPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".config", "wen", "config.yaml")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLoadConfig_XDGOverride(t *testing.T) {
	dir := t.TempDir()
	wenDir := filepath.Join(dir, "wen")
	if err := os.MkdirAll(wenDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := []byte("theme: dracula\nfiscal_year_start: 10\n")
	if err := os.WriteFile(filepath.Join(wenDir, "config.yaml"), content, 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", dir)
	cfg, warnings := LoadConfig()
	if cfg.Theme != "dracula" {
		t.Errorf("expected theme 'dracula', got %q", cfg.Theme)
	}
	if cfg.FiscalYearStart != 10 {
		t.Errorf("expected FiscalYearStart 10, got %d", cfg.FiscalYearStart)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
}

func TestLoadConfig_MissingCreatesDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, warnings := LoadConfig()
	if cfg.Theme != "default" {
		t.Errorf("expected default theme, got %q", cfg.Theme)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}

	// Verify default config file was created on disk.
	created := filepath.Join(dir, "wen", "config.yaml")
	if _, err := os.Stat(created); err != nil {
		t.Errorf("expected default config file at %s, got error: %v", created, err)
	}
}
```

Note: These tests use `t.Setenv()` which automatically restores the env var and prevents `t.Parallel()`. This is correct — env mutation tests must not run in parallel.

- [ ] **Step 2: Run tests to verify they pass**

Run: `go test -run 'TestConfigPath|TestLoadConfig_XDG|TestLoadConfig_Missing' -v ./calendar/`
Expected: PASS

- [ ] **Step 3: Run full test suite**

Run: `make check`
Expected: All tests pass, no lint issues.

- [ ] **Step 4: Commit**

```bash
git add calendar/config_test.go
git commit -m "test: add coverage for LoadConfig and configPath"
```
