# Code Review Fixes Design

Close the four gaps identified during principal-engineer code review:
Model/RowModel duplication, LoadConfig test coverage, highlight error
swallowing, and AST interface coverage.

## 1. Embedded baseModel (Model/RowModel Deduplication)

### New file: `calendar/base_model.go`

Extract shared state into an embedded struct:

```go
type baseModel struct {
    cursor, today    time.Time
    quit, selected   bool
    rangeAnchor      *time.Time
    highlightedDates map[time.Time]bool
    highlightPath    string
    activeWatcher    *fsnotify.Watcher
    config           Config
    help             help.Model
    styles           resolvedStyles
    showHelp         bool
    julian, printMode bool
    dayFmt           dayFormat
    termWidth, termHeight int
    warnings         []string
}
```

### Shared methods on `*baseModel`

- **Accessors** (promoted via embedding): `IsQuit`, `Selected`, `Cursor`,
  `InRange`, `RangeStart`, `RangeEnd`
- **`initCmds() []tea.Cmd`** — midnight tick + optional watcher start
- **`handleMsg(msg tea.Msg) (tea.Cmd, bool)`** — handles WindowSizeMsg,
  watcherErrMsg, midnightTickMsg, highlightChangedMsg. Returns `(cmd, true)`
  if handled so the caller can short-circuit.
- **`closeWatcher()`** — the 4-line cleanup block (nil-check, Close, nil-assign),
  called from quit/select paths.
- **`doQuit() tea.Cmd`** — sets quit, closes watcher, returns `tea.Quit`.
- **`doSelect() tea.Cmd`** — sets selected, closes watcher, returns `tea.Quit`.
- **`doVisualSelect()`** — sets rangeAnchor to cursor.
- **`cancelRange() bool`** — clears rangeAnchor if set, returns whether it was
  set (for Esc/q fallthrough to quit).

### Model and RowModel after extraction

```go
type Model struct {
    baseModel
    weekNumPos WeekNumPos
    months     int
    keys       keyMap
}

type RowModel struct {
    baseModel
    keys rowKeyMap
}
```

Each model's `Update()` starts with:

```go
if cmd, handled := m.handleMsg(msg); handled {
    return m, cmd
}
// model-specific key dispatch using m.doQuit(), m.doSelect(), etc.
```

### What stays model-specific

- **keyMap / rowKeyMap** — different structs, different bindings, stay separate.
- **View()** — completely different rendering. Stays in view.go / row_model.go.
- **RowModel.visibleWindow()** — viewport trimming, stays in row_model.go.
- **RowModel style stripping** — mosh underline workaround in NewRow().
- **ModelOption / RowModelOption** — separate types closing over `*Model` /
  `*RowModel`, but base fields accessed directly via embedding.

### Tests

Existing tests remain unchanged (they test the public interface). Add
`base_model_test.go` with focused tests for `handleMsg`, `closeWatcher`,
`doQuit`, `doSelect`, `cancelRange`.

## 2. LoadConfig Test Coverage

### Problem

`LoadConfig()` and `configPath()` are at 0% coverage. `loadConfigFromPath()`
is already well-tested. The gap is the thin glue that resolves XDG paths.

### New tests in `config_test.go`

No production code changes. Four new tests using `t.Setenv()` (cannot use
`t.Parallel()` since they mutate env):

1. **TestLoadConfig_XDGOverride** — Set `XDG_CONFIG_HOME` to a temp dir
   containing a valid config.yaml, call `LoadConfig()`, verify values loaded.
2. **TestLoadConfig_MissingCreatesDefault** — Set `XDG_CONFIG_HOME` to an empty
   temp dir, call `LoadConfig()`, verify it returns `DefaultConfig()` and the
   default file was created on disk.
3. **TestConfigPath_XDGSet** — Set `XDG_CONFIG_HOME`, verify `configPath()`
   returns the expected path.
4. **TestConfigPath_XDGUnset** — Unset `XDG_CONFIG_HOME`, verify `configPath()`
   falls back to `~/.config/wen/config.yaml`.

## 3. LoadHighlightedDates Warnings

### Signature change

```go
// Before
func LoadHighlightedDates(path string) map[time.Time]bool

// After
func LoadHighlightedDates(path string) (map[time.Time]bool, []string)
```

### Warning cases

| Condition | Warning message |
|-----------|----------------|
| File doesn't exist | `"highlight file not found: <path>"` |
| JSON unmarshal fails | `"highlight file is not valid JSON: <path>"` |
| Individual date invalid | `"skipping invalid date <value> in <path>"` |
| Empty path | No warning — `(nil, nil)`. Not configured, not an error. |

### Caller updates (4 sites)

1. **`WithHighlightSource()`** in highlight.go — collect warnings into
   `baseModel.warnings`.
2. **`WithRowHighlightSource()`** in row_model.go — same via baseModel embedding.
3. **`highlightChangedMsg` handler** in baseModel.handleMsg — watcher reload.
   Warnings dropped silently (mid-session TUI warnings are disruptive).
4. **`startFileWatcher()`** in watcher.go — initial load inside goroutine.
   Warnings dropped (startup warnings already surfaced by option).

### How warnings reach the user

`baseModel` gains a `warnings []string` field and a `Warnings() []string`
accessor (exported, promoted via embedding). Highlight-loading options append
to it. The CLI layer (`cal.go`, `row.go`) reads them after construction, same
pattern as config warnings:

```go
for _, w := range m.Warnings() {
    fmt.Fprintf(os.Stderr, "warning: %s\n", w)
}
```

Watcher reloads stay silent — the user already saw the file load successfully
at startup, and mid-session stderr writes in a TUI are disruptive.

### Tests

Update existing highlight_test.go cases to assert the second return value. Add
cases for each warning message.

## 4. AST Interface Coverage

### Compile-time assertions in `ast.go`

```go
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

### New test file: `ast_test.go`

Single table-driven test calling `.dateExpr()` on each concrete type. Trivial
test whose purpose is closing the coverage gap and documenting the full AST
node set.

### Effect

ast.go coverage: 0% to 100%. Total project coverage: ~88.8% to ~89.2%.

## Scope and constraints

- **Public API changes:** `LoadHighlightedDates` signature adds a `[]string`
  return. `Model` and `RowModel` field layout changes but exported methods and
  constructors are unchanged.
- **No new dependencies.**
- **All existing tests must continue to pass.** New tests added for each change.
- **`make check` must pass** (tests + lint) before completion.
