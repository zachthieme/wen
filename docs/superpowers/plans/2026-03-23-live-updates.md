# Live Updates Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add live updates to the calendar TUI — midnight tick to refresh the "today" highlight and fsnotify-based file watching to reload highlighted dates when the source file changes.

**Architecture:** Two independent features wired into the existing Bubble Tea model. The midnight tick uses `tea.Tick` to fire at midnight and update `m.today`. File watching creates an fsnotify watcher in a cmd goroutine (not stored on the model due to value receivers), communicates via `tea.Msg`, and includes 150ms debounce. Both features are activated in `Init()`.

**Tech Stack:** Go, Bubble Tea (bubbletea v1.3.10), fsnotify

**Spec:** `docs/superpowers/specs/2026-03-23-live-updates-design.md`

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `calendar/highlight.go` | Modify | Add `expandTilde` helper, refactor `LoadHighlightedDates` to use it, add `WithHighlightSource` option |
| `calendar/highlight_test.go` | Modify | Add tests for `expandTilde`, `WithHighlightSource` |
| `calendar/watcher.go` | Create | `startFileWatcher`, `watchLoop`, `waitForNextChange`, message types |
| `calendar/watcher_test.go` | Create | Integration tests for file watcher + debounce |
| `calendar/model.go` | Modify | Add `highlightPath` field, `midnightTickMsg`, update `Init()` and `Update()` |
| `calendar/model_test.go` | Modify | Add midnight tick tests |
| `cmd/wen/main.go` | Modify | Switch from `WithHighlightedDates` to `WithHighlightSource` |
| `go.mod` / `go.sum` | Modify | Add `github.com/fsnotify/fsnotify` dependency |

---

### Task 1: Add `expandTilde` helper and refactor `LoadHighlightedDates`

**Files:**
- Modify: `calendar/highlight.go:20-27` (inline tilde expansion)
- Test: `calendar/highlight_test.go`

- [ ] **Step 1: Write the failing test for `expandTilde`**

In `calendar/highlight_test.go`, add:

```go
func TestExpandTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	t.Run("expands tilde prefix", func(t *testing.T) {
		got := expandTilde("~/foo/bar")
		want := filepath.Join(home, "foo/bar")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("leaves absolute path unchanged", func(t *testing.T) {
		got := expandTilde("/absolute/path")
		if got != "/absolute/path" {
			t.Errorf("got %q, want %q", got, "/absolute/path")
		}
	})

	t.Run("leaves empty string unchanged", func(t *testing.T) {
		got := expandTilde("")
		if got != "" {
			t.Errorf("got %q, want %q", got, "")
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./calendar/ -run TestExpandTilde -v`
Expected: FAIL — `expandTilde` undefined

- [ ] **Step 3: Implement `expandTilde` and refactor `LoadHighlightedDates`**

In `calendar/highlight.go`, add the helper and replace the inline expansion:

```go
// expandTilde replaces a leading ~ with the user's home directory.
// Returns the path unchanged if it doesn't start with ~ or if the home
// directory cannot be determined.
func expandTilde(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}
```

Then replace lines 20-27 of `LoadHighlightedDates` (the inline `~` expansion block) with:

```go
	path = expandTilde(path)
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./calendar/ -run "TestExpandTilde|TestLoadHighlightedDates" -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add calendar/highlight.go calendar/highlight_test.go
git commit -m "refactor: extract expandTilde helper from LoadHighlightedDates"
```

---

### Task 2: Add `WithHighlightSource` model option

**Files:**
- Modify: `calendar/highlight.go` (add option)
- Modify: `calendar/model.go:19-33` (add `highlightPath` field)
- Test: `calendar/highlight_test.go`

- [ ] **Step 1: Write the failing test**

In `calendar/highlight_test.go`, add:

```go
func TestWithHighlightSource(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dates.json")
	if err := os.WriteFile(path, []byte(`["2026-03-25"]`), 0644); err != nil {
		t.Fatal(err)
	}

	today := time.Date(2026, time.March, 17, 0, 0, 0, 0, time.Local)
	m := New(today, today, DefaultConfig(), WithHighlightSource(path))

	if m.highlightPath != path {
		t.Errorf("highlightPath = %q, want %q", m.highlightPath, path)
	}
	key := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	if !m.highlightedDates[key] {
		t.Error("expected 2026-03-25 to be highlighted")
	}
}

func TestWithHighlightSourceMissing(t *testing.T) {
	today := time.Date(2026, time.March, 17, 0, 0, 0, 0, time.Local)
	m := New(today, today, DefaultConfig(), WithHighlightSource("/nonexistent/file.json"))

	if m.highlightedDates != nil {
		t.Error("expected nil highlightedDates for missing file")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./calendar/ -run "TestWithHighlightSource" -v`
Expected: FAIL — `WithHighlightSource` undefined, `highlightPath` unknown field

- [ ] **Step 3: Add `highlightPath` field and `WithHighlightSource` option**

In `calendar/model.go`, add to the `Model` struct (after `highlightedDates`):

```go
	highlightPath    string
```

In `calendar/highlight.go`, add:

```go
// WithHighlightSource sets the path to a JSON file of dates to highlight.
// It expands ~ to the user's home directory, performs the initial load, and
// enables file watching when Init() runs. Last option wins if both
// WithHighlightSource and WithHighlightedDates are provided.
func WithHighlightSource(path string) ModelOption {
	return func(m *Model) {
		m.highlightPath = expandTilde(path)
		m.highlightedDates = LoadHighlightedDates(m.highlightPath)
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./calendar/ -run "TestWithHighlightSource" -v`
Expected: all PASS

- [ ] **Step 5: Run full test suite to check for regressions**

Run: `go test ./...`
Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add calendar/model.go calendar/highlight.go calendar/highlight_test.go
git commit -m "feat: add WithHighlightSource option for live file-based highlights"
```

---

### Task 3: Implement midnight tick

**Files:**
- Modify: `calendar/model.go:128-131` (Init), `calendar/model.go:134-189` (Update)
- Test: `calendar/model_test.go`

- [ ] **Step 1: Write the failing test for midnight tick message handling**

In `calendar/model_test.go`, add:

```go
func TestMidnightTickUpdatesToday(t *testing.T) {
	// Start with today = March 17
	oldToday := date(2026, time.March, 17)
	m := New(oldToday, oldToday, DefaultConfig())

	// Simulate midnight tick
	updated, cmd := m.Update(midnightTickMsg{})
	m = updated.(Model)

	// today should be updated to the real current time (which is not March 17)
	// We can't assert exact value, but we can verify it changed if we're not
	// on March 17 2026, or verify it's a valid date
	now := time.Now()
	expected := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if m.today != expected {
		t.Errorf("today = %s, want %s", m.today.Format(DateLayout), expected.Format(DateLayout))
	}

	// Should return a non-nil cmd to schedule the next tick
	if cmd == nil {
		t.Error("expected non-nil cmd for next midnight tick")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./calendar/ -run TestMidnightTickUpdatesToday -v`
Expected: FAIL — `midnightTickMsg` undefined

- [ ] **Step 3: Implement midnight tick**

In `calendar/model.go`, add the message type and scheduling function (before or after the existing `Init` function):

```go
// midnightTickMsg is sent when the clock crosses midnight, triggering a
// refresh of the "today" highlight.
type midnightTickMsg struct{}

// scheduleMidnightTick returns a tea.Cmd that fires at the next midnight.
func scheduleMidnightTick(now time.Time) tea.Cmd {
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return tea.Tick(time.Until(next), func(t time.Time) tea.Msg {
		return midnightTickMsg{}
	})
}
```

Update `Init()` to return the midnight tick:

```go
func (m Model) Init() tea.Cmd {
	return scheduleMidnightTick(m.today)
}
```

Add the case to `Update()`, inside the `switch msg := msg.(type)` block, before the `tea.KeyMsg` case:

```go
	case midnightTickMsg:
		now := time.Now()
		m.today = stripTime(now)
		return m, scheduleMidnightTick(now)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./calendar/ -run TestMidnightTickUpdatesToday -v`
Expected: PASS

- [ ] **Step 5: Write test that Init returns a non-nil cmd**

In `calendar/model_test.go`, add:

```go
func TestInitReturnsMidnightTick(t *testing.T) {
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init() to return a non-nil cmd for midnight tick")
	}
}
```

- [ ] **Step 6: Run test to verify it passes**

Run: `go test ./calendar/ -run TestInitReturnsMidnightTick -v`
Expected: PASS

- [ ] **Step 7: Run full test suite**

Run: `go test ./...`
Expected: all PASS

- [ ] **Step 8: Commit**

```bash
git add calendar/model.go calendar/model_test.go
git commit -m "feat: add midnight tick to refresh today highlight"
```

---

### Task 4: Add fsnotify dependency

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Add fsnotify dependency**

```bash
go get github.com/fsnotify/fsnotify
```

- [ ] **Step 2: Verify it was added**

Run: `go mod tidy && grep fsnotify go.mod`
Expected: line containing `github.com/fsnotify/fsnotify`

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add fsnotify dependency for file watching"
```

---

### Task 5: Implement file watcher

**Files:**
- Create: `calendar/watcher.go`
- Test: `calendar/watcher_test.go`

- [ ] **Step 1: Write the integration test**

Create `calendar/watcher_test.go`:

```go
package calendar

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchLoopDetectsFileChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dates.json")

	// Write initial file
	initial := []string{"2026-03-25"}
	data, _ := json.Marshal(initial)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Start the watcher cmd
	cmd := startFileWatcher(path)
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}

	// Modify the file in a goroutine after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		updated := []string{"2026-03-25", "2026-04-01"}
		data, _ := json.Marshal(updated)
		_ = os.WriteFile(path, data, 0644)
	}()

	// The cmd blocks until a change is detected — run it with a timeout
	done := make(chan highlightChangedMsg, 1)
	go func() {
		msg := cmd()
		if msg != nil {
			done <- msg.(highlightChangedMsg)
		}
	}()

	select {
	case msg := <-done:
		if len(msg.dates) != 2 {
			t.Errorf("expected 2 dates, got %d", len(msg.dates))
		}
		if msg.watcher == nil {
			t.Error("expected non-nil watcher in message")
		}
		if msg.path != path {
			t.Errorf("path = %q, want %q", msg.path, path)
		}
		// Clean up watcher
		msg.watcher.Close()
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for file change detection")
	}
}

func TestWatchLoopFileDeleted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dates.json")

	// Write initial file
	data, _ := json.Marshal([]string{"2026-03-25"})
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	cmd := startFileWatcher(path)
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}

	// Delete the file after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = os.Remove(path)
	}()

	done := make(chan highlightChangedMsg, 1)
	go func() {
		msg := cmd()
		if msg != nil {
			done <- msg.(highlightChangedMsg)
		}
	}()

	select {
	case msg := <-done:
		if msg.dates != nil {
			t.Errorf("expected nil dates after deletion, got %d dates", len(msg.dates))
		}
		msg.watcher.Close()
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for delete detection")
	}
}

func TestStartFileWatcherMissingParentDir(t *testing.T) {
	cmd := startFileWatcher("/nonexistent/parent/dir/dates.json")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	// The cmd should return nil when it can't watch
	msg := cmd()
	if msg != nil {
		t.Errorf("expected nil msg for missing parent dir, got %T", msg)
	}
}

func TestWaitForNextChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dates.json")

	data, _ := json.Marshal([]string{"2026-03-25"})
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Create a watcher manually to test waitForNextChange
	cmd := startFileWatcher(path)

	// Trigger first change
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = os.WriteFile(path, []byte(`["2026-04-01"]`), 0644)
	}()

	done := make(chan highlightChangedMsg, 1)
	go func() {
		msg := cmd()
		if msg != nil {
			done <- msg.(highlightChangedMsg)
		}
	}()

	select {
	case msg := <-done:
		// Now test waitForNextChange with the returned watcher
		cmd2 := waitForNextChange(msg.watcher, msg.path)
		if cmd2 == nil {
			t.Fatal("expected non-nil cmd from waitForNextChange")
		}

		// Trigger second change
		go func() {
			time.Sleep(50 * time.Millisecond)
			_ = os.WriteFile(path, []byte(`["2026-05-01", "2026-06-01"]`), 0644)
		}()

		done2 := make(chan highlightChangedMsg, 1)
		go func() {
			msg2 := cmd2()
			if msg2 != nil {
				done2 <- msg2.(highlightChangedMsg)
			}
		}()

		select {
		case msg2 := <-done2:
			if len(msg2.dates) != 2 {
				t.Errorf("expected 2 dates on second change, got %d", len(msg2.dates))
			}
			msg2.watcher.Close()
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for second file change")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for first file change")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./calendar/ -run "TestWatchLoop|TestStartFileWatcher|TestWaitForNextChange" -v`
Expected: FAIL — `startFileWatcher`, `highlightChangedMsg`, `waitForNextChange` undefined

- [ ] **Step 3: Implement the watcher**

Create `calendar/watcher.go`:

```go
package calendar

import (
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

const debounceDuration = 150 * time.Millisecond

// highlightChangedMsg is sent when the highlight source file changes.
// It carries the reloaded dates, the watcher handle (for reuse), and the path.
type highlightChangedMsg struct {
	dates   map[time.Time]bool
	watcher *fsnotify.Watcher
	path    string
}

// startFileWatcher returns a tea.Cmd that creates an fsnotify watcher on the
// parent directory of path, waits for a change to the target file, and returns
// a highlightChangedMsg with the reloaded dates.
func startFileWatcher(path string) tea.Cmd {
	return func() tea.Msg {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return nil
		}
		dir := filepath.Dir(path)
		if err := watcher.Add(dir); err != nil {
			watcher.Close()
			return nil
		}
		return watchLoop(watcher, path)
	}
}

// waitForNextChange returns a tea.Cmd that reuses an existing watcher to wait
// for the next file change. Uses the same debounced watchLoop.
func waitForNextChange(watcher *fsnotify.Watcher, path string) tea.Cmd {
	return func() tea.Msg {
		return watchLoop(watcher, path)
	}
}

// watchLoop blocks on the watcher's event channel, debounces rapid events,
// and returns a highlightChangedMsg when the target file changes.
func watchLoop(watcher *fsnotify.Watcher, path string) tea.Msg {
	target := filepath.Base(path)
	debounce := time.NewTimer(0)
	if !debounce.Stop() {
		<-debounce.C
	}
	triggered := false

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if filepath.Base(event.Name) != target {
				continue
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}
			// Reset debounce timer
			if !debounce.Stop() && triggered {
				select {
				case <-debounce.C:
				default:
				}
			}
			debounce.Reset(debounceDuration)
			triggered = true

		case <-debounce.C:
			dates := LoadHighlightedDates(path)
			return highlightChangedMsg{
				dates:   dates,
				watcher: watcher,
				path:    path,
			}

		case _, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			// Continue listening on errors
		}
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./calendar/ -run "TestWatchLoop|TestStartFileWatcher|TestWaitForNextChange" -v -timeout 30s`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add calendar/watcher.go calendar/watcher_test.go
git commit -m "feat: add fsnotify-based file watcher for highlight source"
```

---

### Task 6: Wire file watcher into Init/Update

**Files:**
- Modify: `calendar/model.go:128-131` (Init), `calendar/model.go:134-189` (Update)
- Test: `calendar/model_test.go`

- [ ] **Step 1: Write the failing test**

In `calendar/model_test.go`, add `"os"` and `"path/filepath"` to the import block, then add:

```go
func TestUpdateHighlightChangedMsg(t *testing.T) {
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())

	// Simulate receiving a highlightChangedMsg
	newDates := map[time.Time]bool{
		time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC): true,
	}
	updated, cmd := m.Update(highlightChangedMsg{
		dates:   newDates,
		watcher: nil, // cmd is checked but never executed, so nil watcher is safe here
		path:    "/test/path",
	})
	m = updated.(Model)

	if len(m.highlightedDates) != 1 {
		t.Errorf("expected 1 highlighted date, got %d", len(m.highlightedDates))
	}
	key := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	if !m.highlightedDates[key] {
		t.Error("expected 2026-03-25 to be highlighted")
	}

	// Should return a cmd to wait for next change
	if cmd == nil {
		t.Error("expected non-nil cmd for next file watch")
	}
}

func TestUpdateHighlightChangedMsgNilDates(t *testing.T) {
	today := date(2026, time.March, 17)
	initialDates := map[time.Time]bool{
		time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC): true,
	}
	m := New(today, today, DefaultConfig(), WithHighlightedDates(initialDates))

	// Simulate file deletion (nil dates)
	updated, _ := m.Update(highlightChangedMsg{
		dates:   nil,
		watcher: nil,
		path:    "/test/path",
	})
	m = updated.(Model)

	if m.highlightedDates != nil {
		t.Errorf("expected nil highlightedDates, got %d dates", len(m.highlightedDates))
	}
}

func TestInitWithHighlightSource(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dates.json")
	if err := os.WriteFile(path, []byte(`["2026-03-25"]`), 0644); err != nil {
		t.Fatal(err)
	}

	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig(), WithHighlightSource(path))

	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init() to return a non-nil cmd")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./calendar/ -run "TestUpdateHighlightChangedMsg|TestInitWithHighlightSource" -v`
Expected: FAIL — Update doesn't handle `highlightChangedMsg`

- [ ] **Step 3: Wire into Init and Update**

Update `Init()` in `calendar/model.go`:

```go
func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, scheduleMidnightTick(m.today))
	if m.highlightPath != "" {
		cmds = append(cmds, startFileWatcher(m.highlightPath))
	}
	return tea.Batch(cmds...)
}
```

Add the case to `Update()`, inside the `switch msg := msg.(type)` block (after `midnightTickMsg`, before `tea.KeyMsg`):

```go
	case highlightChangedMsg:
		m.highlightedDates = msg.dates
		return m, waitForNextChange(msg.watcher, msg.path)
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./calendar/ -run "TestUpdateHighlightChangedMsg|TestInitWithHighlightSource|TestInitReturnsMidnightTick" -v`
Expected: all PASS

- [ ] **Step 5: Run full test suite**

Run: `go test ./...`
Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add calendar/model.go calendar/model_test.go
git commit -m "feat: wire file watcher and midnight tick into Init/Update"
```

---

### Task 7: Update main.go to use `WithHighlightSource`

**Files:**
- Modify: `cmd/wen/main.go:277-284`

- [ ] **Step 1: Update `runCalendar` to use `WithHighlightSource`**

In `cmd/wen/main.go`, replace the highlight loading block (lines 277-284):

```go
	// Load highlighted dates from file (priority: --highlight-file > config > default path).
	highlightPath := calendar.ResolveHighlightSource(*highlightFile, cfg.HighlightSource)
	highlightedDates := calendar.LoadHighlightedDates(highlightPath)

	var modelOpts []calendar.ModelOption
	if highlightedDates != nil {
		modelOpts = append(modelOpts, calendar.WithHighlightedDates(highlightedDates))
	}
```

With:

```go
	// Resolve highlight source (priority: --highlight-file > config > default path).
	highlightPath := calendar.ResolveHighlightSource(*highlightFile, cfg.HighlightSource)

	var modelOpts []calendar.ModelOption
	if highlightPath != "" {
		modelOpts = append(modelOpts, calendar.WithHighlightSource(highlightPath))
	}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./cmd/wen/`
Expected: success, no errors

- [ ] **Step 3: Run full test suite**

Run: `go test ./...`
Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add cmd/wen/main.go
git commit -m "refactor: use WithHighlightSource for live highlight updates in calendar"
```

---

### Task 8: Final validation

- [ ] **Step 1: Run full test suite with race detector**

Run: `go test -race ./...`
Expected: all PASS, no data races

- [ ] **Step 2: Run linter**

Run: `golangci-lint run` (if configured; check `.golangci.yml`)
Expected: no new issues

- [ ] **Step 3: Manual smoke test**

Create a test highlight file and run the calendar:

```bash
mkdir -p /tmp/wen-test
echo '["2026-03-25", "2026-03-28"]' > /tmp/wen-test/dates.json
go run ./cmd/wen cal --highlight-file /tmp/wen-test/dates.json
```

While the calendar is open in another terminal:

```bash
echo '["2026-03-25", "2026-03-28", "2026-03-30"]' > /tmp/wen-test/dates.json
```

Verify: the new date appears highlighted without restarting the calendar.

- [ ] **Step 4: Clean up test files**

```bash
rm -rf /tmp/wen-test
```
