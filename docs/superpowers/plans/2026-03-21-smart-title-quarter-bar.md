# Smart Title + Quarter Progress Bar Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Omit the year from the calendar title when showing the current year, and add an optional quarter progress bar below the grid.

**Architecture:** Modify `renderTitle` for smart year display. Add `ShowQuarterBar` config field and a new `renderQuarterBar` method called from both single and multi-month render paths. Quarter math uses the existing `wen.FiscalQuarter` function plus a new `quarterStartDate` helper.

**Tech Stack:** Go, Bubble Tea, lipgloss

**Spec:** `docs/superpowers/specs/2026-03-21-smart-title-quarter-bar-design.md`

---

### Task 1: Config — Add ShowQuarterBar

**Files:**
- Modify: `calendar/config.go:32-46` (Config struct)
- Modify: `calendar/config.go:209` (writeDefaultConfig template)

- [ ] **Step 1: Add ShowQuarterBar to Config struct**

In `calendar/config.go`, add after `ShowFiscalQuarter bool` (line 38):

```go
ShowQuarterBar     bool        `yaml:"show_quarter_bar"`
```

- [ ] **Step 2: Add to writeDefaultConfig template**

In the `writeDefaultConfig` inline template, after the `# show_fiscal_quarter: false` line (line 209), add:

```
# show_quarter_bar: false
```

- [ ] **Step 3: Run tests**

Run: `go test ./calendar/ -count=1`
Expected: All pass (no behavioral change)

- [ ] **Step 4: Commit**

```bash
git add calendar/config.go
git commit -m "Add show_quarter_bar config field"
```

---

### Task 2: Smart Title — Omit current year

**Files:**
- Modify: `calendar/view.go:177-186` (renderTitle)
- Test: `calendar/view_test.go`

- [ ] **Step 1: Write failing tests**

Add to `calendar/view_test.go`:

```go
func TestSmartTitleCurrentYear(t *testing.T) {
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	output := m.View()
	// Should show "March" without "2026"
	if !strings.Contains(output, "March") {
		t.Error("expected 'March' in output")
	}
	if strings.Contains(output, "2026") {
		t.Error("expected year to be omitted for current year")
	}
}

func TestSmartTitleOtherYear(t *testing.T) {
	cursor := date(2027, time.March, 17)
	today := date(2026, time.March, 17)
	m := New(cursor, today, DefaultConfig())
	output := m.View()
	if !strings.Contains(output, "March 2027") {
		t.Error("expected 'March 2027' for non-current year")
	}
}

func TestSmartTitleWithFiscalQuarter(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FiscalYearStart = 10
	cfg.ShowFiscalQuarter = true
	today := date(2026, time.March, 17)
	m := New(today, today, cfg)
	output := m.View()
	// Title should be "March · Q2 FY26" (no year, with fiscal suffix)
	if !strings.Contains(output, "Q2 FY26") {
		t.Errorf("expected fiscal quarter in title, got:\n%s", output)
	}
	// Should NOT contain "March 2026" (year omitted)
	if strings.Contains(output, "2026") && !strings.Contains(output, "FY26") {
		t.Error("expected calendar year to be omitted")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./calendar/ -count=1 -run "TestSmartTitle"`
Expected: FAIL — `TestSmartTitleCurrentYear` fails because title still shows "March 2026"

- [ ] **Step 3: Implement smart title in renderTitle**

In `calendar/view.go`, replace the `renderTitle` method (lines 177-186):

```go
func (m Model) renderTitle(b *strings.Builder, month time.Month, year int) {
	var title string
	if year == m.today.Year() {
		title = month.String()
	} else {
		title = fmt.Sprintf("%s %d", month, year)
	}
	if m.config.ShowFiscalQuarter && m.config.FiscalYearStart > 1 {
		q, fy := wen.FiscalQuarter(int(month), year, m.config.FiscalYearStart)
		title += fmt.Sprintf(" · Q%d FY%02d", q, fy%100)
	}
	padding := max((dayGridWidth-len([]rune(title)))/2, 0)
	b.WriteString(m.styles.title.Render(strings.Repeat(" ", padding) + title))
	b.WriteString("\n")
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./calendar/ -count=1 -run "TestSmartTitle"`
Expected: All 3 pass

- [ ] **Step 5: Run full calendar tests**

Run: `go test ./calendar/ -count=1`
Expected: Some existing tests may fail because they check for "March 2026" when today is March 2026. Fix any that break by using a `today` that differs from the cursor year, or by checking for just "March".

- [ ] **Step 6: Commit**

```bash
git add calendar/view.go calendar/view_test.go
git commit -m "Smart title: omit year when viewing current year"
```

---

### Task 3: Quarter Progress Bar — Implementation

**Files:**
- Modify: `calendar/view.go` (add renderQuarterBar, call from renderSingleMonth and renderMultiMonth)
- Test: `calendar/view_test.go`

- [ ] **Step 1: Write failing tests**

Add to `calendar/view_test.go`:

```go
func TestQuarterBarHiddenByDefault(t *testing.T) {
	today := date(2026, time.March, 17)
	m := New(today, today, DefaultConfig())
	output := m.View()
	if strings.Contains(output, "Q1") && strings.Contains(output, "░") {
		t.Error("quarter bar should not appear with default config")
	}
}

func TestQuarterBarRendering(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ShowQuarterBar = true
	today := date(2026, time.March, 17)
	m := New(today, today, cfg)
	output := m.View()
	if !strings.Contains(output, "Q1") {
		t.Errorf("expected Q1 in quarter bar, got:\n%s", output)
	}
	if !strings.Contains(output, "█") {
		t.Error("expected filled bar character")
	}
	if !strings.Contains(output, "░") {
		t.Error("expected empty bar character")
	}
	if !strings.Contains(output, "%") {
		t.Error("expected percentage in bar")
	}
}

func TestQuarterBarFiscalQuarter(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ShowQuarterBar = true
	cfg.FiscalYearStart = 10 // Oct start: March is Q2
	today := date(2026, time.March, 17)
	m := New(today, today, cfg)
	output := m.View()
	if !strings.Contains(output, "Q2") {
		t.Errorf("expected Q2 for fiscal Oct start in March, got:\n%s", output)
	}
}

func TestQuarterBarProgress(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ShowQuarterBar = true

	// Jan 1 = start of Q1, ~1% progress
	startOfQ := New(date(2026, time.January, 1), date(2026, time.January, 1), cfg)
	startOutput := startOfQ.View()
	if !strings.Contains(startOutput, "Q1") {
		t.Error("expected Q1 at start of year")
	}

	// March 31 = end of Q1, ~100% progress
	endOfQ := New(date(2026, time.March, 31), date(2026, time.March, 31), cfg)
	endOutput := endOfQ.View()
	if !strings.Contains(endOutput, "Q1") {
		t.Error("expected Q1 at end of March")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./calendar/ -count=1 -run "TestQuarterBar"`
Expected: FAIL — `renderQuarterBar` doesn't exist yet

- [ ] **Step 3: Add quarterStartDate helper**

In `calendar/view.go`, add before `weekNumber`:

```go
// quarterStartDate returns the first day of the quarter containing the given date,
// using fiscalYearStart to determine quarter boundaries.
func quarterStartDate(cursor time.Time, fiscalYearStart int) time.Time {
	if fiscalYearStart < 1 || fiscalYearStart > 12 {
		fiscalYearStart = 1
	}
	month := int(cursor.Month())
	year := cursor.Year()

	// Determine fiscal year calendar start
	fyCalStart := year
	if month < fiscalYearStart {
		fyCalStart = year - 1
	}

	q, _ := wen.FiscalQuarter(month, year, fiscalYearStart)

	// Quarter start = fyCalStart + fiscalYearStart + (q-1)*3 months
	startMonth := fiscalYearStart + (q-1)*3
	startYear := fyCalStart
	for startMonth > 12 {
		startMonth -= 12
		startYear++
	}
	return time.Date(startYear, time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC)
}
```

- [ ] **Step 4: Add renderQuarterBar method**

In `calendar/view.go`, add after `quarterStartDate`:

```go
const barWidth = 12

func (m Model) renderQuarterBar(b *strings.Builder) {
	if !m.config.ShowQuarterBar {
		return
	}

	fyStart := m.config.FiscalYearStart
	if fyStart < 1 {
		fyStart = 1
	}

	cursorMonth := int(m.cursor.Month())
	cursorYear := m.cursor.Year()
	q, _ := wen.FiscalQuarter(cursorMonth, cursorYear, fyStart)

	qStart := quarterStartDate(m.cursor, fyStart)
	qEnd := time.Date(qStart.Year(), qStart.Month()+3, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)

	cursorUTC := time.Date(cursorYear, m.cursor.Month(), m.cursor.Day(), 0, 0, 0, 0, time.UTC)
	daysElapsed := int(cursorUTC.Sub(qStart).Hours()/24) + 1
	totalDays := int(qEnd.Sub(qStart).Hours()/24) + 1

	if totalDays <= 0 {
		totalDays = 1
	}
	progress := float64(daysElapsed) / float64(totalDays)
	if progress > 1 {
		progress = 1
	}
	if progress < 0 {
		progress = 0
	}

	filled := int(progress * barWidth)
	empty := barWidth - filled
	pct := int(progress * 100)

	label := fmt.Sprintf("Q%d ", q)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	suffix := fmt.Sprintf(" %d%%", pct)

	b.WriteString(m.styles.title.Render(label))
	b.WriteString(m.styles.title.Render(strings.Repeat("█", filled)))
	b.WriteString(m.styles.weekNum.Render(strings.Repeat("░", empty)))
	b.WriteString(m.styles.title.Render(suffix))
	_ = bar // suppress unused warning for the combined bar var
	b.WriteString("\n")
}
```

- [ ] **Step 5: Call renderQuarterBar from renderSingleMonth**

In `renderSingleMonth`, after `m.renderGrid(...)` (line 92) and before the help bar check, add:

```go
m.renderQuarterBar(&b)
```

- [ ] **Step 6: Call renderQuarterBar from renderMultiMonth**

In `renderMultiMonth`, after the multi-month join loop (after the `result.WriteString("\n")` closing the row loop, around line 161) and before the help bar check, add:

```go
m.renderQuarterBar(&result)
```

- [ ] **Step 7: Run tests**

Run: `go test ./calendar/ -count=1 -v`
Expected: All tests pass

- [ ] **Step 8: Commit**

```bash
git add calendar/view.go calendar/view_test.go
git commit -m "Add quarter progress bar below calendar grid"
```

---

### Task 4: README — Document new features

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Add show_quarter_bar to config section**

In the config YAML example in README.md, after the `# show_fiscal_quarter: false` line, add:

```yaml
# show_quarter_bar: false
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "Document show_quarter_bar config option"
```

---

### Task 5: Final verification

- [ ] **Step 1: Run full test suite**

Run: `go test ./... -count=1 -cover`
Expected: All pass

- [ ] **Step 2: Smoke test manually**

```bash
go build -o wen ./cmd/wen
# Test smart title (should show just "March" for current month)
./wen cal
# Test with year (should show "March 2027")
./wen cal march 2027
# Enable quarter bar in config and test
echo "show_quarter_bar: true" >> ~/.config/wen/config.yaml
./wen cal
# Test with fiscal year
echo "fiscal_year_start: 10" >> ~/.config/wen/config.yaml
./wen cal
# Test multi-month
./wen cal -3
# Clean up test config additions
```

- [ ] **Step 3: Commit any fixes**
