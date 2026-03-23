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
    return func(m *Model) {
        m.highlightPath = expandTilde(path)
        m.highlightedDates = LoadHighlightedDates(m.highlightPath)
    }
}
```

The option expands `~` to an absolute path (fsnotify requires absolute paths), then performs the initial date load during construction (since `Init()` has a value receiver and cannot mutate model fields). `WithHighlightedDates` remains available for programmatic use (no file watching).

**Precedence**: If both `WithHighlightSource` and `WithHighlightedDates` are provided, the last option wins (standard Go option pattern — options are applied in order).

**New model fields**:

```go
highlightPath string   // absolute path to highlight JSON file (tilde-expanded)
```

Note: the `fsnotify.Watcher` is **not** stored on the model. Since `Init()` and `Update()` use value receivers (required by `tea.Model`), storing a pointer to a watcher on the struct is fragile. Instead, the watcher is created inside the cmd goroutine returned by `Init()` and lives for the duration of that goroutine. The goroutine communicates solely via `tea.Msg` values.

**Message type**:

```go
type highlightChangedMsg struct {
    dates map[time.Time]bool
}
```

**Watcher lifecycle**:

1. `Init()` — if `highlightPath` is set, return `startFileWatcher(highlightPath)` as a cmd.
2. `startFileWatcher(path) tea.Cmd` — the returned cmd goroutine:
   a. Creates `fsnotify.NewWatcher()`, adds the parent directory of `path`.
   b. Enters an event loop: on `Write`/`Create`/`Remove`/`Rename` events where the event filename matches the target, debounce 150ms via `time.Timer` reset.
   c. When the debounce timer fires, reload via `LoadHighlightedDates(path)` and return `highlightChangedMsg{dates, watcher, path}`.
   d. If watcher creation or `Add()` fails (e.g. parent dir doesn't exist), return `nil` (silently give up — no watching).
3. `Update` handles `highlightChangedMsg`: assign `m.highlightedDates = msg.dates` (nil is valid — Go nil map reads return zero value `false`, which correctly clears all highlights). Return `waitForNextChange(msg.watcher, msg.path)`.
4. `waitForNextChange(watcher, path) tea.Cmd` — identical debounce loop to step 2b-c, but reuses the existing watcher rather than creating a new one. Same function can be extracted: both `startFileWatcher` and `waitForNextChange` call a shared `watchLoop(watcher, path) tea.Msg` that contains the event loop + debounce.
5. **Cleanup**: The watcher goroutine exits naturally when the Bubble Tea program shuts down (the process exits). For a CLI tool this is acceptable — the OS reclaims file descriptors. No explicit `Close()` is needed.

**highlightChangedMsg carries watcher handle**: The message includes the `*fsnotify.Watcher` and path so that `Update` can pass them to `waitForNextChange` without storing them on the model (which would be fragile with value receivers).

```go
type highlightChangedMsg struct {
    dates   map[time.Time]bool
    watcher *fsnotify.Watcher
    path    string
}
```

**Debounce implementation**: Both `startFileWatcher` and `waitForNextChange` share a `watchLoop` function. It uses a `time.Timer` internally: on each matching fsnotify event, reset the timer to 150ms. When the timer fires (no new events for 150ms), reload the file and return. This happens entirely within the blocking cmd goroutine.

**Watching strategy**: Watch the parent directory (not just the file) so that delete+recreate patterns (common with atomic writes) are caught. Filter events to only react to the target filename.

**Tilde expansion**: `LoadHighlightedDates` already handles `~` internally for loading, but `fsnotify.Watcher.Add()` requires absolute paths. The `WithHighlightSource` option expands `~` at construction time using `os.UserHomeDir()`. A shared `expandTilde` helper is added to `highlight.go`. `LoadHighlightedDates` is refactored to use this helper instead of its inline expansion.

**Failure modes** (all silent, consistent with existing `LoadHighlightedDates`):
- File doesn't exist at startup: no highlights, watcher watches parent dir for creation.
- File deleted while open: highlights clear (`LoadHighlightedDates` returns nil, assigned to `m.highlightedDates` — Go nil map reads return `false`, so all `isHighlighted` checks correctly fail). Watcher stays active on parent dir.
- File unreadable or malformed: highlights clear (same nil map behavior).
- Parent directory doesn't exist: `fsnotify.Watcher.Add()` fails, `startFileWatcher` returns nil (no watching, no crash). Highlights from initial load (if any) remain static.
- Watcher error channel event: goroutine continues listening.

### Combined Init()

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

Note: initial dates are loaded in `WithHighlightSource` during `New()`, not in `Init()`, because `Init()` has a value receiver.

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
- **File watcher**: Integration test that writes a temp JSON file, creates a model with `WithHighlightSource`, modifies the file, and verifies highlights update via the message type.
- **Tilde expansion**: Unit test for `expandTilde` helper.
