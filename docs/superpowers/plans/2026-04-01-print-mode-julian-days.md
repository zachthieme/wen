# Print Mode & Julian Day-of-Year Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add non-interactive `--print` mode and Julian day-of-year numbering to both `wen cal` and `wen row`, letting `wen` fully replace the Unix `cal` utility.

**Architecture:** Reuse existing `View()` rendering by adding `printMode` and `julian` booleans to both `Model` and `RowModel`. Print mode suppresses cursor styling and calls `View()` directly without the Bubble Tea event loop. Julian mode widens cells from 2 to 3 chars and displays `YearDay()` instead of day-of-month. A `gridWidth()` method centralizes the width calculation.

**Tech Stack:** Go, Bubble Tea, lipgloss, `golang.org/x/term` (already a dependency)

---

### Task 1: Add `julian` and `printMode` fields + options to grid Model

**Files:**
- Modify: `calendar/model.go:17-33` (Model struct)
- Modify: `calendar/model.go:85-126` (ModelOption, New)
- Test: `calendar/model_test.go`

- [ ] **Step 1: Write failing test for WithJulian option**

In `calendar/model_test.go`, add:

```go
func TestWithJulian(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig(), WithJulian(true))
	if !m.julian {
		t.Error("expected julian to be true")
	}
}

func TestWithPrintMode(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig(), WithPrintMode(true))
	if !m.printMode {
		t.Error("expected printMode to be true")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run 'TestWithJulian|TestWithPrintMode' -v`
Expected: FAIL — `m.julian` and `m.printMode` undefined, `WithJulian` and `WithPrintMode` undefined

- [ ] **Step 3: Add fields and options to Model**

In `calendar/model.go`, add `julian` and `printMode` fields to the `Model` struct (after the `months` field):

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
	highlightedDates map[time.Time]bool
	highlightPath    string
	activeWatcher    *fsnotify.Watcher
	config           Config
	keys             keyMap
	help             help.Model
	styles           resolvedStyles
}
```

Add the option functions after `WithMonths`:

```go
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run 'TestWithJulian|TestWithPrintMode' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add calendar/model.go calendar/model_test.go
git commit -m "feat(calendar): add julian and printMode fields to grid Model"
```

---

### Task 2: Add `julian` and `printMode` fields + options to RowModel

**Files:**
- Modify: `calendar/row_model.go:17-32` (RowModel struct)
- Test: `calendar/row_model_test.go`

- [ ] **Step 1: Write failing test for WithRowJulian option**

In `calendar/row_model_test.go`, add:

```go
func TestWithRowJulian(t *testing.T) {
	m := NewRow(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig(), WithRowJulian(true))
	if !m.julian {
		t.Error("expected julian to be true")
	}
}

func TestWithRowPrintMode(t *testing.T) {
	m := NewRow(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig(), WithRowPrintMode(true))
	if !m.printMode {
		t.Error("expected printMode to be true")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run 'TestWithRowJulian|TestWithRowPrintMode' -v`
Expected: FAIL

- [ ] **Step 3: Add fields and options to RowModel**

In `calendar/row_model.go`, add `julian` and `printMode` fields to `RowModel` (after `showHelp`):

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
	termWidth        int
}
```

Add option functions after `WithRowHighlightSource`:

```go
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run 'TestWithRowJulian|TestWithRowPrintMode' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add calendar/row_model.go calendar/row_model_test.go
git commit -m "feat(calendar): add julian and printMode fields to RowModel"
```

---

### Task 3: Add `julian` config option

**Files:**
- Modify: `calendar/config.go:62-76` (Config struct)
- Test: `calendar/config_test.go`

- [ ] **Step 1: Write failing test for julian config**

In `calendar/config_test.go`, add:

```go
func TestJulianConfigField(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Julian {
		t.Error("expected Julian to default to false")
	}

	yamlData := []byte("julian: true\n")
	var loaded Config
	if err := yaml.Unmarshal(yamlData, &loaded); err != nil {
		t.Fatal(err)
	}
	if !loaded.Julian {
		t.Error("expected Julian to be true after loading from YAML")
	}
}
```

(May need to add `"gopkg.in/yaml.v3"` to imports in config_test.go if not already present.)

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run TestJulianConfigField -v`
Expected: FAIL — `cfg.Julian` undefined

- [ ] **Step 3: Add Julian field to Config**

In `calendar/config.go`, add the `Julian` field to the `Config` struct:

```go
type Config struct {
	ShowWeekNumbers   string      `yaml:"show_week_numbers"`
	WeekNumbering     string      `yaml:"week_numbering"`
	WeekStartDay      int         `yaml:"week_start_day"`
	FiscalYearStart   int         `yaml:"fiscal_year_start"`
	ShowFiscalQuarter bool        `yaml:"show_fiscal_quarter"`
	ShowQuarterBar    bool        `yaml:"show_quarter_bar"`
	Julian            bool        `yaml:"julian"`
	Theme             string      `yaml:"theme"`
	Colors            ThemeColors `yaml:"colors"`
	HighlightSource   string      `yaml:"highlight_source"`
	PaddingTop        int         `yaml:"padding_top"`
	PaddingRight      int         `yaml:"padding_right"`
	PaddingBottom     int         `yaml:"padding_bottom"`
	PaddingLeft       int         `yaml:"padding_left"`
}
```

No changes needed to `DefaultConfig()` — `bool` zero value is `false`.

Add a comment to the default config YAML in `writeDefaultConfig`:

```
# Julian day-of-year numbering (shows day 1-366 instead of day of month)
# julian: false
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run TestJulianConfigField -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add calendar/config.go calendar/config_test.go
git commit -m "feat(config): add julian option to config"
```

---

### Task 4: Add `gridWidth()` method and Julian day headers to grid calendar

**Files:**
- Modify: `calendar/render.go:14-48` (dayNames, dayGridWidth, renderDayHeaders, renderTitle)
- Test: `calendar/render_test.go`

- [ ] **Step 1: Write failing tests**

In `calendar/render_test.go`, add:

```go
func TestGridWidth(t *testing.T) {
	t.Run("normal mode", func(t *testing.T) {
		m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
		if m.gridWidth() != 20 {
			t.Errorf("expected gridWidth 20, got %d", m.gridWidth())
		}
	})
	t.Run("julian mode", func(t *testing.T) {
		m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig(), WithJulian(true))
		if m.gridWidth() != 27 {
			t.Errorf("expected gridWidth 27, got %d", m.gridWidth())
		}
	})
}

func TestRenderDayHeadersJulian(t *testing.T) {
	cfg := DefaultConfig()
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), cfg, WithJulian(true))
	var b strings.Builder
	m.renderDayHeaders(&b)
	got := b.String()
	if !strings.Contains(got, "Sun Mon Tue Wed Thu Fri Sat") {
		t.Errorf("julian headers should use 3-char names, got: %q", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run 'TestGridWidth|TestRenderDayHeadersJulian' -v`
Expected: FAIL — `m.gridWidth` undefined

- [ ] **Step 3: Add gridWidth method, dayNamesLong, and update renderDayHeaders/renderTitle**

In `calendar/render.go`:

Add long day names and julian grid width constant:

```go
var dayNames = [7]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
var dayNamesLong = [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

// dayGridWidth is the character width of the 7-column day grid (2-char days).
const dayGridWidth = 20

// julianGridWidth is the character width of the 7-column day grid (3-char days).
const julianGridWidth = 27
```

Add the `gridWidth` method to `Model`:

```go
// gridWidth returns the character width of the day grid based on julian mode.
func (m Model) gridWidth() int {
	if m.julian {
		return julianGridWidth
	}
	return dayGridWidth
}
```

Update `renderTitle` to use `m.gridWidth()`:

```go
func (m Model) renderTitle(b *strings.Builder, month time.Month, year int) {
	// ... existing title logic unchanged ...
	titleStyle := m.styles.title.Width(m.gridWidth()).Align(lipgloss.Center)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")
}
```

Update `renderDayHeaders` to use 3-char names in julian mode:

```go
func (m Model) renderDayHeaders(b *strings.Builder) {
	startDay := m.config.WeekStartDay
	headers := make([]string, 7)
	names := dayNames
	if m.julian {
		names = dayNamesLong
	}
	for i := range 7 {
		headers[i] = names[(startDay+i)%7]
	}
	b.WriteString(m.styles.dayHeader.Render(strings.Join(headers, " ")))
	b.WriteString("\n")
}
```

Note: `names` needs to be declared as the array type. Since you can't assign `[7]string` to a var and switch, use a slice approach:

```go
func (m Model) renderDayHeaders(b *strings.Builder) {
	startDay := m.config.WeekStartDay
	headers := make([]string, 7)
	for i := range 7 {
		idx := (startDay + i) % 7
		if m.julian {
			headers[i] = dayNamesLong[idx]
		} else {
			headers[i] = dayNames[idx]
		}
	}
	b.WriteString(m.styles.dayHeader.Render(strings.Join(headers, " ")))
	b.WriteString("\n")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run 'TestGridWidth|TestRenderDayHeadersJulian' -v`
Expected: PASS

- [ ] **Step 5: Run all existing tests to check for regressions**

Run: `cd /home/zach/code/wen && go test ./calendar/ -v`
Expected: All tests PASS

- [ ] **Step 6: Commit**

```bash
git add calendar/render.go calendar/render_test.go calendar/model.go
git commit -m "feat(calendar): add gridWidth method and julian day headers"
```

---

### Task 5: Julian day rendering in grid calendar

**Files:**
- Modify: `calendar/render.go:64-130` (renderGrid)
- Test: `calendar/render_test.go`

- [ ] **Step 1: Write failing tests**

In `calendar/render_test.go`, add:

```go
func TestRenderGridJulian(t *testing.T) {
	t.Run("january shows yearday values", func(t *testing.T) {
		cfg := DefaultConfig()
		m := New(date(2026, time.January, 15), date(2026, time.January, 15), cfg, WithJulian(true))
		var b strings.Builder
		m.renderGrid(&b, 2026, time.January, 15, time.Local)
		got := b.String()
		// Jan 1 = yearday 1, Jan 31 = yearday 31
		if !strings.Contains(got, " 1") {
			t.Errorf("expected yearday 1, got:\n%s", got)
		}
		if !strings.Contains(got, " 31") {
			t.Errorf("expected yearday 31, got:\n%s", got)
		}
	})

	t.Run("march shows offset yearday values", func(t *testing.T) {
		cfg := DefaultConfig()
		m := New(date(2026, time.March, 15), date(2026, time.March, 15), cfg, WithJulian(true))
		var b strings.Builder
		m.renderGrid(&b, 2026, time.March, 15, time.Local)
		got := b.String()
		// March 1 2026 = yearday 60, March 31 = yearday 90
		if !strings.Contains(got, " 60") {
			t.Errorf("expected yearday 60 for March 1, got:\n%s", got)
		}
		if !strings.Contains(got, " 90") {
			t.Errorf("expected yearday 90 for March 31, got:\n%s", got)
		}
	})

	t.Run("julian cells are 3 chars wide", func(t *testing.T) {
		cfg := DefaultConfig()
		m := New(date(2026, time.March, 15), date(2026, time.March, 15), cfg, WithJulian(true))
		var b strings.Builder
		m.renderGrid(&b, 2026, time.March, 15, time.Local)
		got := b.String()
		lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
		// Each grid line with content should be julianGridWidth (27) chars
		for _, line := range lines {
			stripped := stripAnsi(line)
			trimmed := strings.TrimRight(stripped, " \n")
			if len(trimmed) > 0 && len(stripped) != julianGridWidth {
				// Last row may be shorter if not a full week, but padded rows should be 27
				// Just verify no line exceeds 27
				if len(stripped) > julianGridWidth {
					t.Errorf("julian grid line too wide (%d): %q", len(stripped), stripped)
				}
			}
		}
	})
}
```

Note: `stripAnsi` may need to be added as a test helper. If ANSI stripping is complex, just check for the presence of yearday numbers instead.

Actually, let's simplify the width test — checking for yearday values is sufficient. Remove the cell-width subtest and rely on the view-level test instead:

```go
func TestRenderGridJulian(t *testing.T) {
	t.Run("january shows yearday values", func(t *testing.T) {
		cfg := DefaultConfig()
		m := New(date(2026, time.January, 15), date(2026, time.January, 15), cfg, WithJulian(true))
		var b strings.Builder
		m.renderGrid(&b, 2026, time.January, 15, time.Local)
		got := b.String()
		if !strings.Contains(got, "  1") {
			t.Errorf("expected yearday 1 (3-char padded), got:\n%s", got)
		}
		if !strings.Contains(got, " 31") {
			t.Errorf("expected yearday 31, got:\n%s", got)
		}
	})

	t.Run("march shows offset yearday values", func(t *testing.T) {
		cfg := DefaultConfig()
		m := New(date(2026, time.March, 15), date(2026, time.March, 15), cfg, WithJulian(true))
		var b strings.Builder
		m.renderGrid(&b, 2026, time.March, 15, time.Local)
		got := b.String()
		// March 1 2026 = yearday 60, March 31 = yearday 90
		if !strings.Contains(got, " 60") {
			t.Errorf("expected yearday 60 for March 1, got:\n%s", got)
		}
		if !strings.Contains(got, " 90") {
			t.Errorf("expected yearday 90 for March 31, got:\n%s", got)
		}
	})

	t.Run("december leap year shows 366", func(t *testing.T) {
		cfg := DefaultConfig()
		m := New(date(2024, time.December, 31), date(2024, time.December, 31), cfg, WithJulian(true))
		var b strings.Builder
		m.renderGrid(&b, 2024, time.December, 31, time.Local)
		got := b.String()
		if !strings.Contains(got, "366") {
			t.Errorf("expected yearday 366 for Dec 31 2024, got:\n%s", got)
		}
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run TestRenderGridJulian -v`
Expected: FAIL — grid still shows day-of-month values

- [ ] **Step 3: Update renderGrid for julian mode**

In `calendar/render.go`, modify `renderGrid`:

```go
func (m Model) renderGrid(b *strings.Builder, year int, month time.Month, cursorDay int, loc *time.Location) []int {
	st := m.styles
	startDay := m.config.WeekStartDay
	first := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	weekday := (int(first.Weekday()) - startDay + 7) % 7
	days := wen.DaysIn(year, month, loc)

	wn := weekNumber(first, m.config.WeekNumbering)
	var weekNums []int
	weekNums = append(weekNums, wn)

	// Cell width: 3 for julian (e.g., " 60"), 2 for normal (e.g., " 5")
	cellWidth := 2
	if m.julian {
		cellWidth = 3
	}

	// Leading spaces for first partial week
	blankCell := strings.Repeat(" ", cellWidth+1) // cell + separator
	b.WriteString(strings.Repeat(blankCell, weekday))

	todayYear, todayMonth, todayDay := m.today.Date()

	for day := 1; day <= days; day++ {
		var dayStr string
		if m.julian {
			yd := time.Date(year, month, day, 0, 0, 0, 0, loc).YearDay()
			dayStr = fmt.Sprintf("%3d", yd)
		} else {
			dayStr = fmt.Sprintf("%2d", day)
		}

		isCursor := day == cursorDay && !m.printMode
		isToday := year == todayYear && month == todayMonth && day == todayDay
		isHighlighted := m.highlightedDates[dateKey(time.Date(year, month, day, 0, 0, 0, 0, loc))]

		isRangeDay := false
		if m.rangeAnchor != nil {
			dayDate := dateKey(time.Date(year, month, day, 0, 0, 0, 0, loc))
			anchorUTC := dateKey(*m.rangeAnchor)
			cursorUTC := dateKey(m.cursor)
			isRangeDay = isInRange(dayDate, anchorUTC, cursorUTC)
		}

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

		b.WriteString(dayStr)

		col := (weekday + day) % 7
		if col == 0 && day < days {
			nextDay := time.Date(year, month, day+1, 0, 0, 0, 0, loc)
			wn = weekNumber(nextDay, m.config.WeekNumbering)
			weekNums = append(weekNums, wn)
			b.WriteString("\n")
		} else if day < days {
			b.WriteString(" ")
		}
	}
	// Pad the last row to grid width so week numbers align.
	lastCol := (weekday + days) % 7
	if lastCol != 0 {
		b.WriteString(strings.Repeat(blankCell, 7-lastCol))
	}
	b.WriteString("\n")
	return weekNums
}
```

Also add `"fmt"` to imports if not already there (it is already imported).

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run TestRenderGridJulian -v`
Expected: PASS

- [ ] **Step 5: Run all calendar tests for regressions**

Run: `cd /home/zach/code/wen && go test ./calendar/ -v`
Expected: All PASS. The `strconv.Itoa` → `fmt.Sprintf("%2d")` change should be functionally equivalent for existing tests.

- [ ] **Step 6: Commit**

```bash
git add calendar/render.go calendar/render_test.go
git commit -m "feat(calendar): julian day-of-year rendering in grid"
```

---

### Task 6: Julian mode in strip calendar rendering

**Files:**
- Modify: `calendar/row_render.go:57-131` (renderStripDayHeaders, renderStripDays)
- Test: `calendar/row_render_test.go`

- [ ] **Step 1: Write failing tests**

In `calendar/row_render_test.go`, add:

```go
func TestRenderStripDayHeadersJulian(t *testing.T) {
	t.Parallel()
	m := NewRow(date(2026, time.March, 15), date(2026, time.March, 15), DefaultConfig(), WithRowJulian(true))
	start, end := stripWindow(2026, time.March, 0, time.Local)
	got := m.renderStripDayHeaders(start, end)
	if !strings.Contains(got, "Sun") {
		t.Errorf("julian strip headers should use 3-char names, got: %q", got)
	}
	if !strings.Contains(got, "Mon") {
		t.Errorf("julian strip headers should contain Mon, got: %q", got)
	}
}

func TestRenderStripDaysJulian(t *testing.T) {
	t.Parallel()
	cursor := date(2026, time.March, 15)
	m := NewRow(cursor, date(2026, time.March, 15), DefaultConfig(), WithRowJulian(true))
	start, end := stripWindow(2026, time.March, 0, time.Local)
	got := m.renderStripDays(start, end)
	// March 1 2026 = yearday 60
	if !strings.Contains(got, " 60") {
		t.Errorf("expected yearday 60 for March 1, got: %q", got)
	}
	// March 31 = yearday 90
	if !strings.Contains(got, " 90") {
		t.Errorf("expected yearday 90 for March 31, got: %q", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run 'TestRenderStripDayHeadersJulian|TestRenderStripDaysJulian' -v`
Expected: FAIL — still shows 2-char headers and day-of-month numbers

- [ ] **Step 3: Update strip rendering for julian mode**

In `calendar/row_render.go`, update `renderStripDayHeaders`:

```go
func (m RowModel) renderStripDayHeaders(start, end time.Time) string {
	var b strings.Builder
	// Leading space: 3 chars for month abbreviation column
	if m.julian {
		b.WriteString("    ") // 4 chars to align with 3-char day cells
	} else {
		b.WriteString("   ") // 3 chars for 2-char day cells
	}
	first := true
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if !first {
			b.WriteString(" ")
		}
		if m.julian {
			b.WriteString(dayNamesLong[d.Weekday()])
		} else {
			b.WriteString(dayNames[d.Weekday()])
		}
		first = false
	}
	return m.styles.dayHeader.Render(b.String())
}
```

Note: `dayNamesLong` is defined in `render.go` and accessible within the `calendar` package.

Update `renderStripDays`:

```go
func (m RowModel) renderStripDays(start, end time.Time) string {
	year, month, _ := m.cursor.Date()
	loc := m.cursor.Location()
	first := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	last := time.Date(year, month+1, 0, 0, 0, 0, 0, loc)

	abbrev := monthAbbrevs[month-1]

	var b strings.Builder
	if m.julian {
		b.WriteString(m.styles.title.Render(abbrev))
		b.WriteString("  ") // 2 spaces to align with 4-char header prefix
	} else {
		b.WriteString(m.styles.title.Render(abbrev))
		b.WriteString(" ")
	}

	todayKey := dateKey(m.today)
	cursorKey := dateKey(m.cursor)

	var anchorKey time.Time
	if m.rangeAnchor != nil {
		anchorKey = dateKey(*m.rangeAnchor)
	}

	firstDay := true
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if !firstDay {
			b.WriteString(" ")
		}

		var dayStr string
		if m.julian {
			dayStr = fmt.Sprintf("%3d", d.YearDay())
		} else {
			dayStr = fmt.Sprintf("%2d", d.Day())
		}
		dk := dateKey(d)
		inMonth := !d.Before(first) && !d.After(last)

		isCursor := dk.Equal(cursorKey) && !m.printMode
		isToday := dk.Equal(todayKey)
		isHighlighted := m.highlightedDates[dk]
		isRangeDay := false
		if m.rangeAnchor != nil {
			isRangeDay = isInRange(dk, anchorKey, cursorKey)
		}

		switch {
		case !inMonth:
			dayStr = m.styles.weekNum.Render(dayStr)
		case isCursor && isToday:
			dayStr = m.styles.cursorToday.Render(dayStr)
		case isCursor:
			dayStr = m.styles.cursor.Render(dayStr)
		case isToday:
			dayStr = m.styles.today.Render(dayStr)
		case isRangeDay:
			dayStr = m.styles.rangeDay.Render(dayStr)
		case isHighlighted:
			dayStr = m.styles.highlight.Render(dayStr)
		}

		b.WriteString(dayStr)
		firstDay = false
	}
	return b.String()
}
```

Add `"fmt"` to imports in `row_render.go` (currently only has `"strings"` and `"time"`).

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run 'TestRenderStripDayHeadersJulian|TestRenderStripDaysJulian' -v`
Expected: PASS

- [ ] **Step 5: Run all calendar tests for regressions**

Run: `cd /home/zach/code/wen && go test ./calendar/ -v`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add calendar/row_render.go calendar/row_render_test.go
git commit -m "feat(calendar): julian day-of-year rendering in strip calendar"
```

---

### Task 7: Update View() for julian grid width cascading

**Files:**
- Modify: `calendar/view.go:70-97` (renderSingleMonth)
- Modify: `calendar/view.go:99-173` (renderMultiMonth)
- Test: `calendar/view_test.go`

- [ ] **Step 1: Write failing tests**

In `calendar/view_test.go`, add:

```go
func TestRenderJulianSingleMonth(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2025, time.March, 17), DefaultConfig(), WithJulian(true))
	output := m.View()
	// Should contain 3-char day headers
	if !strings.Contains(output, "Sun Mon Tue Wed Thu Fri Sat") {
		t.Errorf("expected 3-char julian headers, got:\n%s", output)
	}
	// March 1 = yearday 60
	if !strings.Contains(output, " 60") {
		t.Errorf("expected yearday 60, got:\n%s", output)
	}
}

func TestRenderJulianMultiMonth(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2025, time.March, 17), DefaultConfig(), WithJulian(true), WithMonths(3))
	output := m.View()
	if !strings.Contains(output, "Sun Mon") {
		t.Errorf("expected julian headers in multi-month, got:\n%s", output)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run 'TestRenderJulianSingleMonth|TestRenderJulianMultiMonth' -v`
Expected: FAIL — title width and column width misaligned (may still pass if only checking content; the real issue is alignment)

- [ ] **Step 3: Update renderMultiMonth to use gridWidth()**

In `calendar/view.go`, update `renderMultiMonth` to use `m.gridWidth()` instead of `dayGridWidth`:

Replace the line:
```go
colWidth := dayGridWidth
```
with:
```go
colWidth := m.gridWidth()
```

And replace:
```go
m.renderQuarterBar(&result, totalWidth)
```
— this line already works since `totalWidth` is derived from `colWidth`.

Also, the `renderQuarterBar` in `renderSingleMonth` needs `m.gridWidth()`:

Replace:
```go
m.renderQuarterBar(&core, dayGridWidth)
```
with:
```go
m.renderQuarterBar(&core, m.gridWidth())
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run 'TestRenderJulianSingleMonth|TestRenderJulianMultiMonth' -v`
Expected: PASS

- [ ] **Step 5: Run all calendar tests for regressions**

Run: `cd /home/zach/code/wen && go test ./calendar/ -v`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add calendar/view.go calendar/view_test.go
git commit -m "feat(calendar): julian width cascading in view rendering"
```

---

### Task 8: `J` key toggle for Julian mode in grid TUI

**Files:**
- Modify: `calendar/model.go:244-325` (keyMap, defaultKeyMap, Update)
- Test: `calendar/model_test.go`

- [ ] **Step 1: Write failing tests**

In `calendar/model_test.go`, add:

```go
func TestToggleJulian(t *testing.T) {
	m := New(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	if m.julian {
		t.Error("expected julian false initially")
	}
	m = pressKey(m, "J")
	if !m.julian {
		t.Error("expected julian true after toggle")
	}
	m = pressKey(m, "J")
	if m.julian {
		t.Error("expected julian false after second toggle")
	}
}

func TestYearNavigationRebound(t *testing.T) {
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	// N = next year
	m = pressKey(m, "N")
	if m.cursor != date(2027, time.March, 17) {
		t.Errorf("N should navigate to next year, got %s", m.cursor.Format("2006-01-02"))
	}
	// P = prev year
	m = pressKey(m, "P")
	if m.cursor != date(2026, time.March, 17) {
		t.Errorf("P should navigate to prev year, got %s", m.cursor.Format("2006-01-02"))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run 'TestToggleJulian|TestYearNavigationRebound' -v`
Expected: FAIL — `J` still triggers next year, `N`/`P` not mapped

- [ ] **Step 3: Update keyMap and Update handler**

In `calendar/model.go`, update the `keyMap` struct — replace `PrevYear`/`NextYear` key bindings and add `ToggleJulian`:

```go
type keyMap struct {
	Left          key.Binding
	Right         key.Binding
	Up            key.Binding
	Down          key.Binding
	PrevMonth     key.Binding
	NextMonth     key.Binding
	PrevYear      key.Binding
	NextYear      key.Binding
	Today         key.Binding
	ToggleWeeks   key.Binding
	ToggleJulian  key.Binding
	ToggleHelp    key.Binding
	VisualSelect  key.Binding
	Select        key.Binding
	Quit          key.Binding
	ForceQuit     key.Binding
}
```

In `defaultKeyMap()`, change `NextYear` from `J` to `N`, `PrevYear` from `K` to `P`, and add `ToggleJulian`:

```go
NextYear: key.NewBinding(
	key.WithKeys("N"),
	key.WithHelp("N", "next year"),
),
PrevYear: key.NewBinding(
	key.WithKeys("P"),
	key.WithHelp("P", "prev year"),
),
```

Add after `ToggleWeeks`:

```go
ToggleJulian: key.NewBinding(
	key.WithKeys("J"),
	key.WithHelp("J", "julian"),
),
```

In the `Update` method, add a case for `ToggleJulian` (after the `ToggleWeeks` case):

```go
case key.Matches(msg, m.keys.ToggleJulian):
	m.julian = !m.julian
```

Update `FullHelp` to include `ToggleJulian`:

```go
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.Up, k.Down},
		{k.PrevMonth, k.NextMonth, k.PrevYear, k.NextYear},
		{k.Today, k.ToggleWeeks, k.ToggleJulian},
		{k.VisualSelect, k.Select, k.Quit},
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run 'TestToggleJulian|TestYearNavigationRebound' -v`
Expected: PASS

- [ ] **Step 5: Update existing navigation tests that use J/K**

In `calendar/model_test.go`, the `TestNavigation` table has entries for `"next year (J)"` and `"prev year (K)"`. Update them to use the new keys:

Change:
```go
{"next year (J)", runeMsg("J"), today, date(2027, time.March, 17)},
{"prev year (K)", runeMsg("K"), today, date(2025, time.March, 17)},
```
to:
```go
{"next year (N)", runeMsg("N"), today, date(2027, time.March, 17)},
{"prev year (P)", runeMsg("P"), today, date(2025, time.March, 17)},
```

Also update the leap day test:
```go
{"next year leap day clamp", runeMsg("N"), date(2024, time.February, 29), date(2025, time.February, 28)},
```

- [ ] **Step 6: Run all calendar tests**

Run: `cd /home/zach/code/wen && go test ./calendar/ -v`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add calendar/model.go calendar/model_test.go
git commit -m "feat(calendar): J key toggles julian mode, rebind year nav to N/P"
```

---

### Task 9: `J` key toggle for Julian mode in row TUI

**Files:**
- Modify: `calendar/row_model.go:255-331` (rowKeyMap, defaultRowKeyMap, Update)
- Test: `calendar/row_model_test.go`

- [ ] **Step 1: Write failing test**

In `calendar/row_model_test.go`, add:

```go
func TestRowToggleJulian(t *testing.T) {
	m := NewRow(date(2026, time.March, 17), date(2026, time.March, 17), DefaultConfig())
	if m.julian {
		t.Error("expected julian false initially")
	}
	updated, _ := m.Update(runeMsg("J"))
	m = updated.(RowModel)
	if !m.julian {
		t.Error("expected julian true after toggle")
	}
	updated, _ = m.Update(runeMsg("J"))
	m = updated.(RowModel)
	if m.julian {
		t.Error("expected julian false after second toggle")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run TestRowToggleJulian -v`
Expected: FAIL — `J` not mapped in row model

- [ ] **Step 3: Add ToggleJulian to rowKeyMap and Update**

In `calendar/row_model.go`, add `ToggleJulian` to `rowKeyMap`:

```go
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
```

In `defaultRowKeyMap()`, add after `Today`:

```go
ToggleJulian: key.NewBinding(
	key.WithKeys("J"),
	key.WithHelp("J", "julian"),
),
```

In the `Update` method, add a case before `ToggleHelp`:

```go
case key.Matches(msg, m.keys.ToggleJulian):
	m.julian = !m.julian
```

Update `FullHelp` to include `ToggleJulian`:

```go
func (k rowKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.PrevMonth, k.NextMonth},
		{k.WeekStart, k.WeekEnd, k.MonthStart, k.MonthEnd},
		{k.Today, k.ToggleJulian, k.VisualSelect, k.Select, k.Quit},
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/zach/code/wen && go test ./calendar/ -run TestRowToggleJulian -v`
Expected: PASS

- [ ] **Step 5: Run all calendar tests**

Run: `cd /home/zach/code/wen && go test ./calendar/ -v`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add calendar/row_model.go calendar/row_model_test.go
git commit -m "feat(calendar): J key toggles julian mode in strip calendar"
```

---

### Task 10: Non-interactive print mode in `wen cal`

**Files:**
- Modify: `cmd/wen/cal.go:55-138` (runCalendar)
- Test: `cmd/wen/main_test.go`

- [ ] **Step 1: Write failing tests**

In `cmd/wen/main_test.go`, add these to the `TestRunWithWriter` table (these use `run(&buf, args)` which writes to a `strings.Builder`, so stdout is not a TTY — auto-detect kicks in):

```go
{"cal print march 2026", []string{"cal", "march", "2026"}, "March 2026"},
{"cal print explicit flag", []string{"cal", "--print", "march", "2026"}, "March 2026"},
{"cal print multi month", []string{"cal", "--print", "-3", "march", "2026"}, "February 2026"},
{"cal print julian", []string{"cal", "--print", "--julian", "march", "2026"}, " 60"},
```

Wait — the auto-detect relies on `os.Stdout.Fd()`, but `run(&buf, ...)` writes to a `strings.Builder`, not stdout. The auto-detect needs to check whether the *actual* stdout is a TTY, not the writer. For `run()` tests where the writer is a `strings.Builder`, we need to pass `--print` explicitly or modify the auto-detect.

Actually, looking at the current code: `run(w io.Writer, args)` writes to `w`, but `runCalendar` creates `tea.NewProgram` which goes to the real terminal. For print mode, we just write to `ctx.w`. The auto-detect checks `os.Stdout` — in tests running via `go test`, stdout is typically not a TTY, so auto-detect would kick in. But this is fragile. Better to add `--print` explicitly in tests.

For the `run(&buf, args)` path: `runCalendar` currently always launches a TUI. With print mode, it should write to `ctx.w` and return. Let's test with `--print`:

```go
func TestCalPrint(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"basic month", []string{"cal", "--print", "march", "2026"}, "March 2026"},
		{"contains day headers", []string{"cal", "--print", "march", "2026"}, "Su Mo Tu We Th Fr Sa"},
		{"contains days", []string{"cal", "--print", "march", "2026"}, "31"},
		{"multi month", []string{"cal", "--print", "-3", "march", "2026"}, "February 2026"},
		{"multi month has april", []string{"cal", "--print", "-3", "march", "2026"}, "April 2026"},
		{"julian mode", []string{"cal", "--print", "--julian", "march", "2026"}, " 60"},
		{"julian 3-char headers", []string{"cal", "--print", "--julian", "march", "2026"}, "Sun Mon Tue"},
		{"short flags", []string{"cal", "-p", "-j", "march", "2026"}, " 60"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf strings.Builder
			err := run(&buf, tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := buf.String()
			if !strings.Contains(got, tt.want) {
				t.Errorf("got:\n%s\nwant substring %q", got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/zach/code/wen && go test ./cmd/wen/ -run TestCalPrint -v`
Expected: FAIL — `--print` and `--julian` flags not recognized

- [ ] **Step 3: Implement print mode in runCalendar**

In `cmd/wen/cal.go`, update `runCalendar`:

```go
func runCalendar(ctx appContext, args []string) error {
	args = expandMonthShorthand(args)

	fs := flag.NewFlagSet("cal", flag.ContinueOnError)
	paddingTop := fs.Int("padding-top", 0, "top padding (lines)")
	paddingRight := fs.Int("padding-right", 0, "right padding (characters)")
	paddingBottom := fs.Int("padding-bottom", 0, "bottom padding (lines)")
	paddingLeft := fs.Int("padding-left", 0, "left padding (characters)")
	highlightFile := fs.String("highlight-file", "", "path to JSON file with dates to highlight")
	monthCount := fs.Int("months", 1, "number of months to display side by side")
	fs.IntVar(monthCount, "m", 1, "shorthand for --months")
	printFlag := fs.Bool("print", false, "print calendar and exit (non-interactive)")
	fs.BoolVar(printFlag, "p", false, "shorthand for --print")
	julianFlag := fs.Bool("julian", false, "show Julian day-of-year numbers")
	fs.BoolVar(julianFlag, "j", false, "shorthand for --julian")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	cursor, err := parseCalArgs(fs.Args(), ctx.now, ctx.parseOpts...)
	if err != nil {
		return err
	}

	cfg := ctx.cfg

	for _, w := range cfg.Normalize() {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

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

	for _, w := range cfg.Normalize() {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

	highlightPath := calendar.ResolveHighlightSource(*highlightFile, cfg.HighlightSource)

	// Resolve julian: CLI flag overrides config
	julian := cfg.Julian || *julianFlag

	// Determine print mode: explicit flag or non-TTY stdout
	printMode := *printFlag || !term.IsTerminal(int(os.Stdout.Fd()))

	var modelOpts []calendar.ModelOption
	if highlightPath != "" {
		modelOpts = append(modelOpts, calendar.WithHighlightSource(highlightPath))
	}
	if *monthCount > 1 {
		modelOpts = append(modelOpts, calendar.WithMonths(*monthCount))
	}
	if julian {
		modelOpts = append(modelOpts, calendar.WithJulian(true))
	}
	if printMode {
		modelOpts = append(modelOpts, calendar.WithPrintMode(true))
	}

	m := calendar.New(cursor, ctx.now, cfg, modelOpts...)

	if printMode {
		fmt.Fprint(ctx.w, m.View())
		return nil
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("calendar: %w", err)
	}

	cal, ok := finalModel.(calendar.Model)
	if !ok {
		return fmt.Errorf("unexpected internal state")
	}
	if cal.InRange() {
		fmt.Fprintln(ctx.w, cal.RangeStart().Format(wen.DateLayout))
		fmt.Fprintln(ctx.w, cal.RangeEnd().Format(wen.DateLayout))
	} else if cal.Selected() {
		fmt.Fprintln(ctx.w, cal.Cursor().Format(wen.DateLayout))
	}
	return nil
}
```

Add `"golang.org/x/term"` to imports in `cal.go`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/zach/code/wen && go test ./cmd/wen/ -run TestCalPrint -v`
Expected: PASS

- [ ] **Step 5: Run all tests**

Run: `cd /home/zach/code/wen && make check`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add cmd/wen/cal.go cmd/wen/main_test.go
git commit -m "feat(cli): add --print and --julian flags to wen cal"
```

---

### Task 11: Non-interactive print mode in `wen row`

**Files:**
- Modify: `cmd/wen/row.go:15-89` (runRow)
- Test: `cmd/wen/main_test.go`

- [ ] **Step 1: Write failing tests**

In `cmd/wen/main_test.go`, add:

```go
func TestRowPrint(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"basic strip", []string{"row", "--print", "march", "2026"}, "Mr"},
		{"contains day headers", []string{"row", "--print", "march", "2026"}, "Su"},
		{"julian mode", []string{"row", "--print", "--julian", "march", "2026"}, "Sun"},
		{"julian yearday", []string{"row", "--print", "--julian", "march", "2026"}, " 60"},
		{"short flags", []string{"row", "-p", "-j", "march", "2026"}, "Sun"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf strings.Builder
			err := run(&buf, tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := buf.String()
			if !strings.Contains(got, tt.want) {
				t.Errorf("got:\n%s\nwant substring %q", got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/zach/code/wen && go test ./cmd/wen/ -run TestRowPrint -v`
Expected: FAIL — `--print` and `--julian` flags not recognized

- [ ] **Step 3: Implement print mode in runRow**

In `cmd/wen/row.go`, update `runRow`:

```go
func runRow(ctx appContext, args []string) error {
	fs := flag.NewFlagSet("row", flag.ContinueOnError)
	paddingTop := fs.Int("padding-top", 0, "top padding (lines)")
	paddingRight := fs.Int("padding-right", 0, "right padding (characters)")
	paddingBottom := fs.Int("padding-bottom", 0, "bottom padding (lines)")
	paddingLeft := fs.Int("padding-left", 0, "left padding (characters)")
	highlightFile := fs.String("highlight-file", "", "path to JSON file with dates to highlight")
	printFlag := fs.Bool("print", false, "print strip calendar and exit (non-interactive)")
	fs.BoolVar(printFlag, "p", false, "shorthand for --print")
	julianFlag := fs.Bool("julian", false, "show Julian day-of-year numbers")
	fs.BoolVar(julianFlag, "j", false, "shorthand for --julian")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	cursor, err := parseCalArgs(fs.Args(), ctx.now, ctx.parseOpts...)
	if err != nil {
		return err
	}

	cfg := ctx.cfg

	for _, w := range cfg.Normalize() {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

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

	for _, w := range cfg.Normalize() {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

	highlightPath := calendar.ResolveHighlightSource(*highlightFile, cfg.HighlightSource)

	julian := cfg.Julian || *julianFlag
	printMode := *printFlag || !term.IsTerminal(int(os.Stdout.Fd()))

	var modelOpts []calendar.RowModelOption
	if highlightPath != "" {
		modelOpts = append(modelOpts, calendar.WithRowHighlightSource(highlightPath))
	}
	if julian {
		modelOpts = append(modelOpts, calendar.WithRowJulian(true))
	}
	if printMode {
		modelOpts = append(modelOpts, calendar.WithRowPrintMode(true))
	}

	m := calendar.NewRow(cursor, ctx.now, cfg, modelOpts...)

	if printMode {
		// Set a reasonable default terminal width for print mode since
		// there's no WindowSizeMsg from Bubble Tea.
		m = m.WithTermWidth(80)
		fmt.Fprint(ctx.w, m.View())
		return nil
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("row: %w", err)
	}

	row, ok := finalModel.(calendar.RowModel)
	if !ok {
		return fmt.Errorf("unexpected internal state")
	}
	if row.InRange() {
		fmt.Fprintln(ctx.w, row.RangeStart().Format(wen.DateLayout))
		fmt.Fprintln(ctx.w, row.RangeEnd().Format(wen.DateLayout))
	} else if row.Selected() {
		fmt.Fprintln(ctx.w, row.Cursor().Format(wen.DateLayout))
	}
	return nil
}
```

Add `"golang.org/x/term"` to imports in `row.go`.

We need a `WithTermWidth` method on `RowModel` for print mode (since no `WindowSizeMsg` arrives). Add to `calendar/row_model.go`:

```go
// WithTermWidth returns a copy of the model with the terminal width set.
// Used in print mode where no WindowSizeMsg is received.
func (m RowModel) WithTermWidth(w int) RowModel {
	m.termWidth = w
	return m
}
```

Actually, a better approach: detect terminal width at print time. In `row.go`, before printing:

```go
if printMode {
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}
	m = m.WithTermWidth(width)
	fmt.Fprint(ctx.w, m.View())
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/zach/code/wen && go test ./cmd/wen/ -run TestRowPrint -v`
Expected: PASS

- [ ] **Step 5: Run all tests**

Run: `cd /home/zach/code/wen && make check`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add cmd/wen/row.go cmd/wen/main_test.go calendar/row_model.go
git commit -m "feat(cli): add --print and --julian flags to wen row"
```

---

### Task 12: Update help text and default config

**Files:**
- Modify: `cmd/wen/main.go:139-206` (printHelp)
- Modify: `calendar/config.go:234-278` (writeDefaultConfig)
- Test: `cmd/wen/main_test.go`

- [ ] **Step 1: Write failing tests**

In `cmd/wen/main_test.go`, add to the `TestRunWithWriter` table or add a new test:

```go
func TestHelpContainsPrintFlag(t *testing.T) {
	t.Parallel()
	var buf strings.Builder
	err := run(&buf, []string{"--help"})
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "--print") {
		t.Error("help should mention --print flag")
	}
	if !strings.Contains(got, "--julian") {
		t.Error("help should mention --julian flag")
	}
	if !strings.Contains(got, "J") {
		t.Error("help should mention J keybinding for julian")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/zach/code/wen && go test ./cmd/wen/ -run TestHelpContainsPrintFlag -v`
Expected: FAIL — help text doesn't mention `--print` or `--julian`

- [ ] **Step 3: Update help text**

In `cmd/wen/main.go`, update `printHelp` to add the new flags and update keybindings:

```go
func printHelp(w io.Writer) {
	fmt.Fprint(w, `wen - a natural language date tool

Usage:
  wen                            Print today's date
  wen <natural language>         Parse a date (e.g., "next friday", "march 25 2026")
  echo "tomorrow" | wen          Parse date from stdin

Subcommands:
  wen cal, calendar [month]      Interactive calendar (e.g., wen cal march)
  wen row [month]                Interactive strip calendar (e.g., wen row march)
  wen diff <date1> <date2>       Show days between two dates
  wen rel, relative <date>       Show human-readable relative distance

Flags:
  -h, --help                     Show this help
  -v, --version                  Show version
  --format <layout>              Output format (Go time layout, default: 2006-01-02)

Calendar flags:
  -p, --print            Print calendar and exit (non-interactive)
  -j, --julian           Show Julian day-of-year numbers
  --padding-top N        Top padding in lines (default: from config or 0)
  --padding-right N      Right padding in characters (default: from config or 0)
  --padding-bottom N     Bottom padding in lines (default: from config or 0)
  --padding-left N       Left padding in characters (default: from config or 0)
  -m, --months N         Number of months to display side by side (default: 1)
  -N                     Shorthand for --months N (e.g., -3 for three months)
  --highlight-file P     Path to JSON file with dates to highlight

Strip calendar flags (wen row):
  -p, --print            Print strip and exit (non-interactive)
  -j, --julian           Show Julian day-of-year numbers
  --padding-top N        Top padding in lines (default: from config or 0)
  --padding-right N      Right padding in characters (default: from config or 0)
  --padding-bottom N     Bottom padding in lines (default: from config or 0)
  --padding-left N       Left padding in characters (default: from config or 0)
  --highlight-file P     Path to JSON file with dates to highlight

Diff flags:
  --weeks              Output in weeks instead of days
  --workdays           Output in workdays instead of days

Calendar keybindings:
  h/l, ←/→         Previous / next day
  j/k, ↑/↓         Next / previous week
  H/L              Previous / next month
  N/P              Next / previous year
  t                Jump to today
  w                Toggle week numbers
  J                Toggle Julian day-of-year numbers
  ?                Toggle help bar
  v                Start range selection
  Enter            Select date and print to stdout
  q, Esc, ctrl+c   Quit

Strip calendar keybindings (wen row):
  h/l, ←/→         Previous / next day
  b/e               Start / end of week
  0/$               Start / end of month
  j/k, ↓/↑         Next / previous month
  t                Jump to today
  J                Toggle Julian day-of-year numbers
  v                Start range selection
  ?                Toggle help bar
  Enter            Select date and print to stdout
  q, Esc, ctrl+c   Quit

Non-interactive mode:
  wen cal --print           Print current month and exit
  wen cal --print march     Print March of current year
  wen cal --print -3        Print 3 months centered on current
  wen cal | cat             Auto-detect: prints without TUI when piped

Exit codes:
  0    Success (date printed)
  2    Error (parse failure, invalid input, etc.)

Config: ~/.config/wen/config.yaml
`)
}
```

Update `writeDefaultConfig` in `calendar/config.go` to add julian comment:

After the `show_quarter_bar` section, add:

```
# Julian day-of-year numbering (shows day 1-366 instead of day of month)
# julian: false
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/zach/code/wen && go test ./cmd/wen/ -run TestHelpContainsPrintFlag -v`
Expected: PASS

- [ ] **Step 5: Run all tests**

Run: `cd /home/zach/code/wen && make check`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add cmd/wen/main.go calendar/config.go cmd/wen/main_test.go
git commit -m "docs: update help text and default config for print mode and julian"
```

---

### Task 13: Integration tests via binary

**Files:**
- Test: `cmd/wen/main_test.go`

- [ ] **Step 1: Write integration tests**

In `cmd/wen/main_test.go`, add:

```go
func TestCalPrintBinary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"print flag", []string{"cal", "--print", "march", "2026"}, "March 2026"},
		{"piped stdout", []string{"cal", "march", "2026"}, "March 2026"},
		{"julian flag", []string{"cal", "--print", "--julian", "march", "2026"}, " 60"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(testBinary, tt.args...)
			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("unexpected error: %s\n%s", err, out)
			}
			got := string(out)
			if !strings.Contains(got, tt.want) {
				t.Errorf("got:\n%s\nwant substring %q", got, tt.want)
			}
		})
	}
}

func TestRowPrintBinary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"print flag", []string{"row", "--print", "march", "2026"}, "Mr"},
		{"piped stdout", []string{"row", "march", "2026"}, "Mr"},
		{"julian flag", []string{"row", "--print", "--julian", "march", "2026"}, "Sun"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(testBinary, tt.args...)
			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("unexpected error: %s\n%s", err, out)
			}
			got := string(out)
			if !strings.Contains(got, tt.want) {
				t.Errorf("got:\n%s\nwant substring %q", got, tt.want)
			}
		})
	}
}
```

Note: the "piped stdout" tests work because `exec.Command` captures stdout to a pipe (not a TTY), so auto-detect kicks in.

- [ ] **Step 2: Run integration tests**

Run: `cd /home/zach/code/wen && go test ./cmd/wen/ -run 'TestCalPrintBinary|TestRowPrintBinary' -v`
Expected: PASS

- [ ] **Step 3: Run full test suite**

Run: `cd /home/zach/code/wen && make check`
Expected: All PASS, all lint clean

- [ ] **Step 4: Commit**

```bash
git add cmd/wen/main_test.go
git commit -m "test: add integration tests for print mode and julian via binary"
```
