# Live Updates: Midnight Tick & Highlight File Watching

## Problem

The calendar TUI has two staleness issues:

1. **Today highlight** — `m.today` is set once at construction. If the calendar stays open past midnight, the "today" highlight is wrong.
2. **Highlight source** — dates from `highlight_source` are loaded once at startup. If the file changes (e.g. a task manager updates due dates), the calendar doesn't reflect it.

## Design

### Feature 1: Midnight Tick

**Mechanism**: `Init()` returns a `tea.Cmd` that sleeps until the next midnight (via `time.Until`), then sends a `midnightTickMsg`. On receipt, `Update` sets `m.today = stripTime(time.Now())` and re-schedules the next tick.

**Why not a fixed interval?** A periodic tick (e.g. every minute) wastes cycles 99.93% of the time. Sleeping until midnight fires exactly once per day.

**Sleep/suspend edge case**: If the machine sleeps through midnight, Go's `time.After` fires immediately on wake since the duration has elapsed. `m.today` updates correctly.

**Message type**:

```go
type midnightTickMsg struct{}
```

**Scheduling function**:

```go
func scheduleMidnightTick(now time.Time) tea.Cmd {
    next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
    return tea.Tick(time.Until(next), func(t time.Time) tea.Msg {
        return midnightTickMsg{}
    })
}
```

**Update handler**:

```go
case midnightTickMsg:
    now := time.Now()
    m.today = stripTime(now)
    return m, scheduleMidnightTick(now)
```

### Feature 2: Highlight File Watching

**New dependency**: `github.com/fsnotify/fsnotify`

**New ModelOption**:

```go
func WithHighlightSource(path string) ModelOption {
    return func(m *Model) { m.highlightPath = path }
}
```

This stores the resolved path. On `Init()`, the model creates an fsnotify watcher and loads the initial dates. `WithHighlightedDates` remains available for programmatic use (no file watching).

**New model fields**:

```go
highlightPath string              // resolved path to highlight JSON file
watcher       *fsnotify.Watcher   // nil if no path configured
```

**Message types**:

```go
type highlightChangedMsg struct {
    dates map[time.Time]bool
}
type watcherReadyMsg struct{}
```

**Watcher lifecycle**:

1. `Init()` — if `highlightPath` is set, load initial dates and return `tea.Batch(startWatcher(), scheduleMidnightTick())`.
2. `startWatcher()` — creates `fsnotify.Watcher`, adds the file (and parent directory for create events), returns a cmd that blocks on the event channel.
3. On `Write`/`Create` event — debounce 150ms, then reload via `LoadHighlightedDates` and send `highlightChangedMsg`.
4. `Update` handles `highlightChangedMsg`: swap `m.highlightedDates`, return cmd to wait for next event.

**Debounce**: Editors often write files via tmp+rename or multiple write syscalls. After receiving an event, wait 150ms for further events before reloading. This avoids reading half-written files.

**Watching strategy**: Watch the parent directory (not just the file) so that delete+recreate patterns (common with atomic writes) are caught. Filter events to only react to the target filename.

**Failure modes** (all silent, consistent with existing `LoadHighlightedDates`):
- File doesn't exist at startup: no highlights, watcher watches parent dir for creation.
- File deleted while open: highlights clear, watcher stays active.
- File unreadable or malformed: highlights clear.
- Watcher error: log nothing, return cmd to keep listening.

### Combined Init()

```go
func (m Model) Init() tea.Cmd {
    var cmds []tea.Cmd
    cmds = append(cmds, scheduleMidnightTick(m.today))
    if m.highlightPath != "" {
        m.highlightedDates = LoadHighlightedDates(m.highlightPath)
        cmds = append(cmds, m.startFileWatcher())
    }
    return tea.Batch(cmds...)
}
```

### main.go Changes

`runCalendar` currently loads dates itself and passes `WithHighlightedDates(dates)`. After this change:
- It passes `WithHighlightSource(path)` instead when a highlight path is resolved.
- The model handles both initial load and live updates.
- `WithHighlightedDates` is kept but not used by main.go.

## What Doesn't Change

- `LoadHighlightedDates` — reused internally by the watcher.
- `ResolveHighlightSource` — still called in main.go to determine the path.
- View/render code — already reads `m.highlightedDates`, no changes needed.
- Config, themes, styles — untouched.
- `WithHighlightedDates` — stays for programmatic/library use.

## Testing

- **Midnight tick**: Unit test that sends `midnightTickMsg` and verifies `m.today` updates. Test scheduling by checking that `Init()` returns a non-nil cmd.
- **File watcher**: Integration test that writes a temp JSON file, creates a model with `WithHighlightSource`, modifies the file, and verifies highlights update. Use a short debounce for tests.
- **Debounce**: Test that rapid successive writes produce only one reload.
