# Changelog

### v1.10.1 — April 6, 2026

**Bug fixes:**
- Lower minimum terminal width floor from 40 to 33 for narrower terminal support.

**Testing:**
- Isolate CLI tests from user config to prevent test pollution from local settings.

**Infrastructure:**
- Add `flake.lock` for reproducible Nix builds.

---

### v1.10.0 — April 5, 2026

**Testing:**
- Timezone edge case tests: 5 non-hour-aligned IANA zones (Kathmandu, Chatham, St. Johns, Kolkata, Tehran), DST spring-forward crossing, midnight boundary, and month clamping across timezones.
- Year inference tests: absolute dates, ordinal weekdays, last-weekday-in-month, and multi-date expressions all verified to infer next year correctly when parsed from a December reference.
- ISO week year-boundary tests: Dec 29-31 belonging to week 1 of next year and Jan 1-3 belonging to last week of previous year.
- Golden file snapshot tests for calendar rendering (8 scenarios) with `-update` flag for regeneration.
- Public API contract tests (17 behavioral guarantees) in `api_contract_test.go`.
- CLI smoke tests for `cal --print` and `row --print` paths (both binary and in-process).

**Infrastructure:**
- Mutation testing via [gremlins](https://github.com/go-gremlins/gremlins): `make mutate` and CI job (informational, non-blocking).

**Improvements:**
- Reduced parser and resolver cyclomatic complexity via helper extraction.
- Extracted `joinColumnsHorizontal` from multi-month render path.
- Moved `TruncateDay` and `DaysIn` to `wen.go` as public API surface.
- Added `NoPosition` constant for `ParseError` sentinel value.
- Documented `Token` struct field validity per kind.

**Bug fixes:**
- Clamp terminal dimensions to minimum 40x10 to prevent rendering panics.
- Store watcher errors as warnings instead of silently discarding.
- Record diagnostic errors for invalid time expressions.
- Reset `bestErr` in `ParseMulti` single-date fallback to prevent stale errors.
- Replace `shiftMonth` loop with O(1) modular arithmetic.
- Semantic errors no longer report misleading position 0.

---

### v1.9.0 — April 5, 2026

**Architecture:**
- Internal AST layer between parser and resolver: grammar recognition (parser.go) is now fully decoupled from date math (resolver.go) via typed AST nodes (now exported as [Expr]). Public API unchanged.
- Removed dead `ref`/`opts` fields from parser struct — parser no longer holds any date-math state.

**New features:**
- Config typo detection: unknown YAML config keys now produce a warning (e.g., `unknown config key "shwo_week_numbers"`), catching silent typos that were previously ignored.

**Improvements:**
- Auto-center calendar and strip views in terminal using `lipgloss.Place`.
- Shared `calendarFlags` struct for `cal` and `row` subcommands, eliminating duplicated flag definitions and resolution logic.
- Package-level `doc.go` for both `wen` and `calendar` packages (Go convention for pkg.go.dev).
- `t.Parallel()` on all eligible tests across the codebase.

**Testing:**
- Algebraic property tests verifying parser invariants: today == TruncateDay(ref), zero offsets, forward/backward symmetry, monotonicity, tomorrow == in 1 day, week == 7 days, multi-date count/weekday correctness.

---

### v1.8.0 — April 2, 2026

**New features:**
- Non-interactive print mode: `wen cal --print` and `wen row --print` render to stdout and exit, no TUI session. Auto-detects piped stdout (e.g., `wen cal | cat`).
- Julian day-of-year numbering: `--julian` / `-j` flag displays days as their position in the year (1–366). Toggle with `J` in the TUI. Configurable via `julian: true` in config.
- Short flags: `-p` for `--print`, `-j` for `--julian` on both `cal` and `row` subcommands.

**Improvements:**
- Extracted `dayFormat` struct to consolidate julian/normal rendering dimensions, eliminating 9 scattered conditionals across rendering code.
- Year navigation rebound from `J`/`K` to `N`/`P` to free `J` for julian toggle.

---

### v1.7.0 — March 31, 2026

**New features:**
- `wen row` subcommand: interactive strip calendar — a compact, horizontal single-month view
- Vim-style navigation: `h`/`l` (day), `j`/`k` (month), `b`/`e` (week start/end), `0`/`$` (month start/end), `t` (today)
- Visual range selection (`v` to anchor, navigate, `Enter` to confirm) outputs one or two dates
- Responsive width: strip auto-trims and centers on cursor when terminal is narrow
- Highlighted dates and live file watching (`--highlight-file`) supported in strip view
- Midnight tick keeps "today" indicator accurate across day boundaries

**Bug fixes:**
- Strip Underline from row styles to fix mosh terminal cursor alignment caused by per-character ANSI wrapping

**Improvements:**
- Comprehensive test coverage for strip model (navigation, vim motions, range selection, quit/select behavior) and rendering (window calculation, day headers, month abbreviations)
- CLI integration tests for `wen row` subcommand

---

### v1.6.0 — March 26, 2026

**Improvements:**
- Informal BNF grammar comment in `parser.go` documenting all productions
- `shiftMonth()` and `modifierDelta()` helpers eliminate duplicated month overflow/underflow logic
- Documented `isLetter()` English-only design constraint in lexer
- `t.Parallel()` added to all eligible tests
- Coverage CI job, `make cover`/`make bench` targets
- Hardened watcher tests with longer timeout and `-short` skip
- Bumped CI actions to Node.js 24

---

### v1.5.1 — March 24, 2026

**New features:**
- `-m` shorthand flag for `--months` in `cal` subcommand (e.g., `wen cal -m 3`)
- `wen.CountWorkdays()` exported from core library

**Bug fixes:**
- Parser error messages now include consumed modifier for context (e.g., `expected weekday after "last"` instead of `expected weekday`)
- `atoi` panics on invariant violation instead of returning silent zero

**Improvements:**
- Deduplicated workday counting logic into `wen.CountWorkdays()` (was duplicated in `diff.go` and `render.go`)
- Documented watcher cancellation pattern (channel closure vs context.Context)
- Documented greedy-left heuristic in diff arg splitting
- Exhaustive `tokenKind` String() coverage test catches missing cases when new token types are added
- External `_test` package (`api_test.go`) verifies public API surface from a consumer's perspective

---

### v1.5.0 — March 21, 2026

**New features:**
- Date range selection in calendar: `v` to anchor, navigate, `Enter` to confirm. Outputs two dates.
- Subcommands with aliases: `cal`/`calendar`, `rel`/`relative`, `diff`
- Smart calendar title: omits year for current year, abbreviates month when fiscal quarter shown
- Quarter progress bar: `show_quarter_bar` config shows workdays remaining (e.g., `Q1 ████████░░░░ 23wd`)
- Week number positioning: `show_week_numbers` accepts `false`/`true`/`"left"`/`"right"`. `w` key cycles off → left → right → off
- Cardinal number words: `"two weeks ago"`, `"in five days"` now parse
- Bare hour after "at": `"tomorrow at 3"` resolves to 03:00
- Fiscal year config: `fiscal_year_start` and `show_fiscal_quarter` for custom quarter boundaries
- Date picker: `Enter` in calendar prints date to stdout for scripting
- Highlighted dates: `--highlight-file` flag and `highlight_source` config
- Multi-month view: `--months N` or `-3` shorthand
- Custom output format: `--format` flag with Go time layout strings
- Date diff: `wen diff <date1> <date2>` with `--weeks`/`--workdays`
- Month-only calendar: `wen cal march` opens March of current year
- Range color in theme system
- `wen.FiscalQuarter()`, `wen.LookupMonth()`, `wen.ParseMulti()` library exports

**Bug fixes:**
- DST off-by-one in date diff/relative (UTC normalization)
- Negative `--weeks` output for reversed dates
- Diff flags rejected after positional args
- `--format` consuming subcommand names
- Week numbers missing in multi-month view
- Quarter bar misaligned in multi-month view
- Timezone mismatch in highlight/range date lookups
- Year validation in calendar month parsing

**Improvements:**
- `io.Writer` threaded through CLI for testability (cmd/wen coverage 2% → 63%)
- `WeekNumPos` type-safe enum, `dateKey()` UTC helper
- Idiomatic lipgloss: `Align(Center)`, `Width()`, removed `hasPadding` bool
- `view.go` split into `styles.go`, `render.go`, `view.go`
- O(1) workday counting (closed-form formula)
- Config loaded once, threaded via `appContext`
- Comprehensive test hardening: boundary conditions, DST regression, error paths

---

### v1.4.1 — March 21, 2026

- Fix release workflow and test for unwritable config path.

---

### v1.4.0 — March 18, 2026

**New features:**
- Custom recursive descent date parser replaces olebedev/when dependency
- `wen` package is now a reusable Go library (`wen.Parse()`, `wen.ParseRelative()`)
- Period mode options: `WithPeriodStart()` / `WithPeriodSame()`
- Input validation rejects invalid dates (e.g., "february 30") and times (e.g., "at 25:00")

**Code quality:**
- Replaced panic("unreachable") calls with proper error returns
- Extracted magic number 31 to `maxDayOfMonth` constant
- Refactored calendar `View()` into `renderTitle()`, `renderDayHeaders()`, `renderGrid()`
- Extracted `dayGridWidth` constant for calendar rendering
- Makefile: added `clean` and `help` targets

**Testing:**
- Added fuzz testing (`FuzzParse`) in CI
- Added input validation tests (invalid days, hours, minutes, meridiem)
- Expanded CLI integration tests (help, version, multi-word args, error cases)
- CLI tests run in parallel with `t.Parallel()`

**Infrastructure:**
- CI test matrix: tests now run on both Linux and macOS
- Restructured CLI into `cmd/wen/` layout

---

### v1.3.0 — March 18, 2026

- Removed Enter/select feature (calendar no longer outputs a date on exit)
- Removed yank-to-clipboard feature (`y` key)
- Removed exit code 1 (no selection vs quit distinction)
- Cleaned up help text, keybindings, and README

---

### v1.2.1 — March 18, 2026

- Week number column moved to the right side of the calendar grid
- Title centering no longer shifts when toggling week numbers on/off

---

### v1.2.0 — March 18, 2026

**New features:**
- Calendar padding via `--padding-top`, `--padding-right`, `--padding-bottom`, `--padding-left` CLI flags
- Padding also configurable in `config.yaml` (clamped 0–20)
- Distinct exit codes: 0 (success), 1 (no selection), 2 (error)
- Yank (`y`) shows "no clipboard tool available" when no clipboard tool is found

**Code quality:**
- Extracted date parsing into `dateparse.go`
- Replaced `init()` with explicit initialization
- Pre-composed Lip Gloss styles (eliminated nested `Render` calls)
- Added `mergeColor`/`applyColor` helpers to reduce duplication
- Config errors surfaced as warnings instead of silent stderr writes
- Consolidated navigation tests into table-driven format
- Added package-level godoc comments

---

### v1.1.0 — March 17, 2026

Internal quality improvements. No user-facing behavior changes.

**Bubble Tea modernization:**
- Key bindings defined with `bubbles/key` — single source of truth for keys and help text
- Help bar rendered with `bubbles/help` instead of hardcoded string
- `Update()` uses `key.Matches()` instead of nested switch statements
- Model fields encapsulated behind accessors

**Go idiom fixes:**
- Config validation returns warnings instead of writing directly to stderr
- Error messages follow Go conventions (no `"error: "` prefix)
- Added `.golangci.yml` with stricter linters (gocritic, revive, misspell)
- Fixed `TestMain` defer-after-exit pattern

---

### v1.0.0 — March 17, 2026

First stable release. Natural language date parsing + interactive calendar picker.

**Date parsing:**
- Natural language dates via `olebedev/when`: "tomorrow", "2 weeks ago", "march 25 2026"
- Custom "this/next/last" weekday handling: "this thursday" vs "next thursday" give correct, different results
- Three input modes: positional args, piped stdin, no-args (prints today)
- Output always `yyyy-mm-dd`
- `--help` (`-h`) and `--version` (`-v`) flags

**Interactive calendar (`wen cal`):**
- `wen cal` opens at current month, `wen cal december 2026` at a specific month
- Vim-style navigation: `h`/`l` (day), `j`/`k` (week), `H`/`L` (month), `J`/`K` (year)
- Arrow keys mirror vim keys
- `t` to jump to today
- `w` toggles week numbers (US and ISO 8601)
- `?` toggles help bar
- `q`/`Esc` quits
- Day/week navigation wraps across boundaries, month/year jumps clamp day

**Themes and config:**
- Four built-in color themes: default, catppuccin-mocha, dracula, nord
- Custom color overrides per element
- `~/.config/wen/config.yaml` (auto-created on first run, respects `$XDG_CONFIG_HOME`)
- Configurable week start day (Sunday/Monday) and week numbering system

**Infrastructure:**
- CI pipeline with test and lint
- GoReleaser for cross-platform releases (linux/darwin x amd64/arm64)
- Nix flake with binary and source builds
- MIT license

---

For usage details, see [README](README.md).
