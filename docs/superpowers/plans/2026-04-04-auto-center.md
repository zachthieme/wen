# Auto-Center Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace manual padding configuration with automatic `lipgloss.Place()` centering so both calendar views render centered in the terminal like a dialog box.

**Architecture:** Remove the four `Padding*` config fields, their CLI flags, validation, and the `styles.padding` lipgloss style. Add `termWidth`/`termHeight` fields to both models, store dimensions from `WindowSizeMsg`, and wrap `View()` output with `lipgloss.Place()`. Print mode skips centering.

**Tech Stack:** Go, Bubble Tea, lipgloss

---

### Task 1: Remove padding from Config and styles

**Files:**
- Modify: `calendar/config.go:46-47` (remove `MaxPadding` const)
- Modify: `calendar/config.go:62-77` (remove `Padding*` fields from `Config`)
- Modify: `calendar/config.go:127-135` (remove padding validation from `Normalize()`)
- Modify: `calendar/config.go:240-282` (remove padding section from default config template)
- Modify: `calendar/model.go:38-49` (remove `padding` from `resolvedStyles`)
- Modify: `calendar/styles.go:15-49` (remove `padding` from `buildStyles` return — it's not there, it's set in `New()`, so this file needs no change)

- [ ] **Step 1: Remove `MaxPadding` const and `Padding*` fields from Config**

In `calendar/config.go`, remove the `MaxPadding` constant (line 47) and the four padding fields from the `Config` struct (lines 73-76):

```go
// Remove this line:
const MaxPadding = 20

// Remove these four lines from the Config struct:
PaddingTop        int         `yaml:"padding_top"`
PaddingRight      int         `yaml:"padding_right"`
PaddingBottom     int         `yaml:"padding_bottom"`
PaddingLeft       int         `yaml:"padding_left"`
```

- [ ] **Step 2: Remove padding validation from `Normalize()`**

In `calendar/config.go`, remove the padding clamping loop from `Normalize()` (lines 127-135):

```go
// Remove this entire block:
for _, p := range []*int{&c.PaddingTop, &c.PaddingRight, &c.PaddingBottom, &c.PaddingLeft} {
    if *p < 0 {
        warnings = append(warnings, "negative padding value clamped to 0")
        *p = 0
    } else if *p > MaxPadding {
        warnings = append(warnings, fmt.Sprintf("padding value %d exceeds maximum, clamped to %d", *p, MaxPadding))
        *p = MaxPadding
    }
}
```

Also remove the `"fmt"` import if it becomes unused (check — `fmt` is also used by other warnings in `Normalize()` and by `writeDefaultConfig`, so it will still be needed).

- [ ] **Step 3: Remove padding section from default config template**

In `calendar/config.go`, remove the padding comment block from `writeDefaultConfig()` (lines 277-281):

```go
// Remove these lines from the content string:
# Padding (0-20, can also be set via --padding-* CLI flags):
# padding_top: 0
# padding_right: 0
# padding_bottom: 0
# padding_left: 0
```

- [ ] **Step 4: Remove `padding` field from `resolvedStyles`**

In `calendar/model.go`, remove the `padding` field from the `resolvedStyles` struct (line 48):

```go
// Remove this line:
padding     lipgloss.Style
```

- [ ] **Step 5: Remove padding style assignment from `New()`**

In `calendar/model.go`, remove lines 136-138 that set `m.styles.padding`:

```go
// Remove these lines:
m.styles.padding = lipgloss.NewStyle().Padding(
    cfg.PaddingTop, cfg.PaddingRight, cfg.PaddingBottom, cfg.PaddingLeft,
)
```

- [ ] **Step 6: Remove padding style assignment from `NewRow()`**

In `calendar/row_model.go`, remove lines 91-93 that set `m.styles.padding`:

```go
// Remove these lines:
m.styles.padding = lipgloss.NewStyle().Padding(
    cfg.PaddingTop, cfg.PaddingRight, cfg.PaddingBottom, cfg.PaddingLeft,
)
```

- [ ] **Step 7: Run tests to see what breaks**

Run: `make check`

Expected: Compilation errors in tests that reference `PaddingTop`, `PaddingLeft`, etc. and in CLI code that references padding flags. That's fine — we'll fix those in the next tasks.

- [ ] **Step 8: Commit**

```bash
git add calendar/config.go calendar/model.go calendar/row_model.go
git commit -m "refactor: remove padding config, styles, and validation"
```

---

### Task 2: Remove padding CLI flags

**Files:**
- Modify: `cmd/wen/cal.go:60-64` (remove padding flag declarations)
- Modify: `cmd/wen/cal.go:85-111` (remove padding override logic and re-normalize)
- Modify: `cmd/wen/row.go:18-21` (remove padding flag declarations)
- Modify: `cmd/wen/row.go:39-66` (remove padding override logic and re-normalize)

- [ ] **Step 1: Remove padding flags and override logic from `runCalendar`**

In `cmd/wen/cal.go`, remove the four padding flag declarations (lines 61-64):

```go
// Remove these four lines:
paddingTop := fs.Int("padding-top", 0, "top padding (lines)")
paddingRight := fs.Int("padding-right", 0, "right padding (characters)")
paddingBottom := fs.Int("padding-bottom", 0, "bottom padding (lines)")
paddingLeft := fs.Int("padding-left", 0, "left padding (characters)")
```

Remove the padding override block (lines 94-106) and the re-normalize block (lines 108-111). The first `cfg.Normalize()` call (line 90) and its comment can stay since it handles non-padding config warnings. But the comment on lines 86-89 references padding — update it:

```go
// Before (lines 85-111):
cfg := ctx.cfg

// Print config warnings to stderr.
// Note: cfg was already normalized during newAppContext, but we re-load
// here because runCalendar needs to apply CLI padding overrides and
// re-normalize. We use the already-loaded cfg to avoid a second disk read.
for _, w := range cfg.Normalize() {
    fmt.Fprintf(os.Stderr, "warning: %s\n", w)
}

// Override config padding with explicitly-set CLI flags.
fs.Visit(func(f *flag.Flag) {
    switch f.Name {
    case "padding-top":
        cfg.PaddingTop = *paddingTop
    case "padding-right":
        cfg.PaddingRight = *paddingRight
    case "padding-bottom":
        cfg.PaddingBottom = *paddingBottom
    case "padding-left":
        cfg.PaddingLeft = *paddingLeft
    }
})

// Re-normalize after CLI overrides to clamp padding values.
for _, w := range cfg.Normalize() {
    fmt.Fprintf(os.Stderr, "warning: %s\n", w)
}

// After:
cfg := ctx.cfg

// Print config warnings to stderr.
for _, w := range cfg.Normalize() {
    fmt.Fprintf(os.Stderr, "warning: %s\n", w)
}
```

- [ ] **Step 2: Remove padding flags and override logic from `runRow`**

In `cmd/wen/row.go`, apply the same removals:

Remove four padding flag declarations (lines 18-21):

```go
// Remove these four lines:
paddingTop := fs.Int("padding-top", 0, "top padding (lines)")
paddingRight := fs.Int("padding-right", 0, "right padding (characters)")
paddingBottom := fs.Int("padding-bottom", 0, "bottom padding (lines)")
paddingLeft := fs.Int("padding-left", 0, "left padding (characters)")
```

Remove the padding override block (lines 49-61) and re-normalize block (lines 63-66). Update the comment:

```go
// Before (lines 39-66):
cfg := ctx.cfg

// Print config warnings to stderr.
// Note: cfg was already normalized during newAppContext, but we re-load
// here because runRow needs to apply CLI padding overrides and
// re-normalize. We use the already-loaded cfg to avoid a second disk read.
for _, w := range cfg.Normalize() {
    fmt.Fprintf(os.Stderr, "warning: %s\n", w)
}

// Override config padding with explicitly-set CLI flags.
fs.Visit(func(f *flag.Flag) {
    switch f.Name {
    case "padding-top":
        cfg.PaddingTop = *paddingTop
    case "padding-right":
        cfg.PaddingRight = *paddingRight
    case "padding-bottom":
        cfg.PaddingBottom = *paddingBottom
    case "padding-left":
        cfg.PaddingLeft = *paddingLeft
    }
})

// Re-normalize after CLI overrides to clamp padding values.
for _, w := range cfg.Normalize() {
    fmt.Fprintf(os.Stderr, "warning: %s\n", w)
}

// After:
cfg := ctx.cfg

// Print config warnings to stderr.
for _, w := range cfg.Normalize() {
    fmt.Fprintf(os.Stderr, "warning: %s\n", w)
}
```

- [ ] **Step 3: Remove unused imports from `cal.go`**

After removing the `fs.Visit` block, the `flag` import reference via `f *flag.Flag` in the Visit closure is gone. However `flag` is still used for `flag.NewFlagSet` and `flag.ContinueOnError` and `flag.ErrHelp`, so it stays. Check if any other imports became unused.

- [ ] **Step 4: Commit**

```bash
git add cmd/wen/cal.go cmd/wen/row.go
git commit -m "refactor: remove padding CLI flags from cal and row commands"
```

---

### Task 3: Add terminal dimensions and centering to Model

**Files:**
- Modify: `calendar/model.go:17-36` (add `termWidth`, `termHeight` fields)
- Modify: `calendar/model.go:172-175` (store dimensions in `WindowSizeMsg` handler)
- Modify: `calendar/view.go:70-97` (replace `padding.Render` with `lipgloss.Place` in `renderSingleMonth`)
- Modify: `calendar/view.go:99-173` (replace `padding.Render` with `lipgloss.Place` in `renderMultiMonth`)

- [ ] **Step 1: Add terminal dimension fields to Model**

In `calendar/model.go`, add two fields to the `Model` struct after the `styles` field (line 35):

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
	activeWatcher    *fsnotify.Watcher // closed on quit to unblock watcher goroutine
	config           Config
	keys             keyMap
	help             help.Model
	styles           resolvedStyles
	termWidth        int
	termHeight       int
}
```

- [ ] **Step 2: Store terminal dimensions from WindowSizeMsg**

In `calendar/model.go`, update the `WindowSizeMsg` case in `Update()`:

```go
// Before:
case tea.WindowSizeMsg:
    m.help.Width = msg.Width
    return m, nil

// After:
case tea.WindowSizeMsg:
    m.help.Width = msg.Width
    m.termWidth = msg.Width
    m.termHeight = msg.Height
    return m, nil
```

- [ ] **Step 3: Replace padding with centering in `renderSingleMonth`**

In `calendar/view.go`, replace the padding line at the end of `renderSingleMonth()`:

```go
// Before (lines 94-96):
output := b.String()
output = m.styles.padding.Render(output)
return output

// After:
output := b.String()
if m.termWidth > 0 && m.termHeight > 0 {
    return lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Center, lipgloss.Center, output)
}
return output
```

- [ ] **Step 4: Replace padding with centering in `renderMultiMonth`**

In `calendar/view.go`, replace the padding line at the end of `renderMultiMonth()`:

```go
// Before (lines 170-172):
output := result.String()
output = m.styles.padding.Render(output)
return output

// After:
output := result.String()
if m.termWidth > 0 && m.termHeight > 0 {
    return lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Center, lipgloss.Center, output)
}
return output
```

- [ ] **Step 5: Run tests**

Run: `make check`

Expected: Calendar tests pass. Config/CLI tests still broken (fixed in Task 5).

- [ ] **Step 6: Commit**

```bash
git add calendar/model.go calendar/view.go
git commit -m "feat: center calendar Model in terminal with lipgloss.Place"
```

---

### Task 4: Add terminal height and centering to RowModel

**Files:**
- Modify: `calendar/row_model.go:17-35` (add `termHeight` field)
- Modify: `calendar/row_model.go:157-160` (store height in `WindowSizeMsg` handler)
- Modify: `calendar/row_model.go:236` (remove padding subtraction from `visibleWindow`)
- Modify: `calendar/row_model.go:265-284` (replace `padding.Render` with `lipgloss.Place` in `View()`)

- [ ] **Step 1: Add `termHeight` field to RowModel**

In `calendar/row_model.go`, add `termHeight` after `termWidth` (line 35):

```go
type RowModel struct {
	cursor           time.Time
	today            time.Time
	quit             bool
	selected         bool
	rangeAnchor      *time.Time
	highlightedDates map[time.Time]bool
	highlightPath    string
	activeWatcher    *fsnotify.Watcher // closed on quit to unblock watcher goroutine
	config           Config
	keys             rowKeyMap
	help             help.Model
	styles           resolvedStyles
	showHelp         bool
	julian           bool
	printMode        bool
	dayFmt           dayFormat
	termWidth        int
	termHeight       int
}
```

- [ ] **Step 2: Store terminal height from WindowSizeMsg**

In `calendar/row_model.go`, update the `WindowSizeMsg` case:

```go
// Before:
case tea.WindowSizeMsg:
    m.help.Width = msg.Width
    m.termWidth = msg.Width
    return m, nil

// After:
case tea.WindowSizeMsg:
    m.help.Width = msg.Width
    m.termWidth = msg.Width
    m.termHeight = msg.Height
    return m, nil
```

- [ ] **Step 3: Remove padding subtraction from `visibleWindow`**

In `calendar/row_model.go`, the `visibleWindow` method subtracts config padding from available width (line 236). Since padding is gone, use `m.termWidth` directly:

```go
// Before:
availWidth := m.termWidth - m.config.PaddingLeft - m.config.PaddingRight

// After:
availWidth := m.termWidth
```

- [ ] **Step 4: Replace padding with centering in `View()`**

In `calendar/row_model.go`, replace the padding line at the end of `View()`:

```go
// Before (line 283):
return m.styles.padding.Render(b.String())

// After:
output := b.String()
if m.termWidth > 0 && m.termHeight > 0 {
    return lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Center, lipgloss.Center, output)
}
return output
```

- [ ] **Step 5: Run tests**

Run: `make check`

Expected: Row model tests pass (the `visibleWindow` test uses `termWidth = 80` with no padding, so the math becomes `availWidth = 80` — same result since config padding defaults to 0). Config/CLI tests still broken.

- [ ] **Step 6: Commit**

```bash
git add calendar/row_model.go
git commit -m "feat: center RowModel in terminal with lipgloss.Place"
```

---

### Task 5: Remove padding tests and fix remaining test failures

**Files:**
- Modify: `calendar/view_test.go:106-147` (remove `TestRenderWithLeftPadding` and `TestRenderWithTopPadding`)
- Modify: `calendar/config_test.go:138-208` (remove all padding config tests)
- Modify: `cmd/wen/main_test.go:369` (remove `--padding-top` test case from `TestExpandMonthShorthand`)
- Modify: `cmd/wen/main_test.go:534-567` (remove `TestCalFlagParsing`)

- [ ] **Step 1: Remove padding view tests**

In `calendar/view_test.go`, remove these two functions entirely:
- `TestRenderWithLeftPadding` (lines 106-130)
- `TestRenderWithTopPadding` (lines 132-147)

- [ ] **Step 2: Remove padding config tests**

In `calendar/config_test.go`, remove these five functions entirely:
- `TestLoadConfigWithPadding` (lines 138-153)
- `TestDefaultConfigPaddingZero` (lines 155-160)
- `TestNormalizeClampsNegativePadding` (lines 162-181)
- `TestNormalizeClampsExcessivePadding` (lines 183-197)
- `TestNormalizeValidPaddingNoWarnings` (lines 199-208)

Also remove the `MaxPadding` reference — check if `config_test.go` imports anything that's now unused. The test file imports `os`, `path/filepath`, `testing` — these are still used by remaining tests.

- [ ] **Step 3: Remove padding test case from `TestExpandMonthShorthand`**

In `cmd/wen/main_test.go`, remove the `--padding-top` test case from the table (line 369):

```go
// Remove this line:
{[]string{"--padding-top", "2"}, []string{"--padding-top", "2"}},
```

- [ ] **Step 4: Remove `TestCalFlagParsing`**

In `cmd/wen/main_test.go`, remove the entire `TestCalFlagParsing` function (lines 534-567). This test exclusively tests padding flag parsing, which no longer exists.

Also check if removing this test makes `flag` import unused in `main_test.go`. The `flag` package is imported on the line — check if other tests use it.

- [ ] **Step 5: Run full test suite**

Run: `make check`

Expected: All tests pass, no lint errors.

- [ ] **Step 6: Commit**

```bash
git add calendar/view_test.go calendar/config_test.go cmd/wen/main_test.go
git commit -m "test: remove padding-related tests"
```

---

### Task 6: Final verification

- [ ] **Step 1: Run `make check`**

Run: `make check`

Expected: All tests pass, lint clean.

- [ ] **Step 2: Build and smoke-test**

Run: `go build ./cmd/wen && ./wen cal`

Expected: Calendar renders centered in the terminal. Resize the terminal — calendar recenters.

Run: `./wen row`

Expected: Strip calendar renders centered.

Run: `./wen cal --print`

Expected: Raw output, no centering, no extra whitespace.

- [ ] **Step 3: Commit any fixups**

If any issues found, fix and commit.
