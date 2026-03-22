# Smart Title + Quarter Progress Bar

## Summary

Two independent calendar view improvements: (1) omit the year from the title when viewing the current year, (2) add an optional quarter progress bar below the calendar grid.

## Feature 1: Smart Title (Current Year Omission)

### Behavior

- When the displayed month is in the current year: render `March` instead of `March 2026`
- When the displayed month is in a different year: render `March 2027` (unchanged)
- The fiscal quarter suffix is unaffected. It only appears when both `show_fiscal_quarter: true` AND `fiscal_year_start > 1` (existing behavior). Example: `March · Q2 FY26` with fiscal start in October.
- No config changes — this is a rendering tweak

### Implementation

In `renderTitle`, compare `year` to `m.today.Year()`. If equal, omit the year from the title string.

```go
var title string
if year == m.today.Year() {
    title = month.String()
} else {
    title = fmt.Sprintf("%s %d", month, year)
}
```

The fiscal quarter suffix block remains unchanged — it appends to `title` regardless of whether the year was included.

### Multi-Month View

Each month column renders its own title via `renderTitle`. Months in the current year show month-only; months in other years (e.g., December 2025 in a `-3` view from January 2026) show month + year. This is correct and requires no special handling.

## Feature 2: Quarter Progress Bar

### Behavior

A single-line progress bar rendered below the calendar grid showing how far through the current quarter the cursor date is.

Format: `Q1 ████████░░░░ 67%`

- `Q1` — quarter number based on cursor date
- Visual bar — filled and empty segments proportional to progress
- `67%` — percentage through the quarter

The bar works with any `fiscal_year_start` value, including the default of 1 (calendar quarters: Q1=Jan-Mar, Q2=Apr-Jun, etc.). It does not depend on `show_fiscal_quarter`.

### Config

New boolean config field, independent of all existing fiscal settings:

```yaml
# Show quarter progress bar below calendar
# show_quarter_bar: false
```

Add `ShowQuarterBar bool \`yaml:"show_quarter_bar"\`` to the Config struct. Default is `false`. No normalization needed (boolean field). Add the commented-out line to the `writeDefaultConfig` template after the `show_fiscal_quarter` line.

### Rendering

- Rendered after the grid and before the help bar
- In `renderSingleMonth`: call `m.renderQuarterBar(&b)` after `renderGrid`, before the help bar check
- In `renderMultiMonth`: call `m.renderQuarterBar(&result)` after the multi-month join loop, before the help bar check. This produces a single bar below all month columns.
- Bar width: 12 characters for the bar itself (filled + empty = 12). Full line is approximately `Q1 ████████████ 100%` = 20 chars, fitting `dayGridWidth`. When week numbers are enabled, the bar stays at this fixed width (no extension needed — it's a summary, not grid-aligned).

Bar characters:
- Filled: `█` (U+2588)
- Empty: `░` (U+2591)

### Styling

- Quarter label (`Q1`) and percentage: title style
- Filled bar segments: title style
- Empty bar segments: faint style (matches week numbers)

### Quarter Start/End Date Calculation

To compute progress, we need the quarter's start and end dates. Given the cursor date and `fiscal_year_start`:

1. Call `wen.FiscalQuarter(cursorMonth, cursorYear, fiscalYearStart)` → returns `(quarter, fiscalYear)`
2. Determine the fiscal year's calendar start year:
   - If `cursorMonth >= fiscalYearStart`: fiscal year started this calendar year (`fyCalStart = cursorYear`)
   - If `cursorMonth < fiscalYearStart`: fiscal year started last calendar year (`fyCalStart = cursorYear - 1`)
3. Quarter start month (1-indexed): `startMonth = fiscalYearStart + (quarter-1)*3`
4. Handle month overflow with modular arithmetic:
   - `startMonth = ((startMonth - 1) % 12) + 1`
   - If `startMonth` overflowed past 12, increment the year accordingly
5. Quarter start date: `time.Date(startYear, startMonth, 1, ...)`
6. Quarter end date: first day of next quarter minus one second, or more simply: `time.Date(startYear, startMonth+3, 1, ...).AddDate(0,0,-1)` for the last day (inclusive)
7. Days elapsed: cursor day - quarter start day (inclusive of start, inclusive of cursor) = `cursor.Sub(quarterStart).Hours()/24 + 1` (using UTC-normalized dates)
8. Total days in quarter: `quarterEnd.Sub(quarterStart).Hours()/24 + 1`
9. Progress: `daysElapsed / totalDays`
10. Filled segments: `int(progress * 12)`, empty: `12 - filled`

Example: cursor = March 18 2026, fiscal_year_start = 1 (calendar quarters)
- FiscalQuarter(3, 2026, 1) → Q1, FY2026
- Q1 starts Jan 1, ends Mar 31 → 90 days total
- Days elapsed: Jan 1 to Mar 18 inclusive = 77 days
- Progress: 77/90 = 85.6% → 10 filled, 2 empty
- Output: `Q1 ██████████░░ 86%`

### State

No new model state needed. The bar is computed from `m.cursor`, `m.today`, and `m.config` on every render.

## Files Modified

- `calendar/model.go` — no changes (no new state)
- `calendar/view.go` — modify `renderTitle` for smart year, add `renderQuarterBar` method, call it from `renderSingleMonth` and `renderMultiMonth`
- `calendar/config.go` — add `ShowQuarterBar bool` field to Config, add to default config template
- `calendar/view_test.go` — test smart title (current year vs other year), test quarter bar rendering
- `README.md` — document `show_quarter_bar` config option

## Testing

- `TestSmartTitleCurrentYear`: title omits year when year matches today
- `TestSmartTitleOtherYear`: title includes year when year differs from today
- `TestSmartTitleWithFiscalQuarter`: fiscal suffix still appears with smart title (requires fiscal_year_start > 1 and show_fiscal_quarter)
- `TestQuarterBarRendering`: bar appears when config enabled, contains Q label and percentage
- `TestQuarterBarHiddenByDefault`: bar not in output with default config
- `TestQuarterBarFiscalQuarter`: bar uses fiscal quarters when fiscal_year_start configured
- `TestQuarterBarProgress`: verify progress calculation at known dates (start of quarter = 0%, end = 100%, mid = ~50%)
