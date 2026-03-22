// Command wen is a natural language date tool that parses human-readable date
// expressions and provides an interactive terminal calendar.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/zachthieme/wen"
	"github.com/zachthieme/wen/calendar"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

var version = "dev"

const (
	minYear = 1000 // lowest value accepted as a calendar year in parseCalArgs
	maxYear = 9999 // highest value accepted as a calendar year in parseCalArgs
)

func main() {
	if err := run(os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

// appContext holds shared state loaded once at startup and threaded to subcommands.
type appContext struct {
	w         io.Writer
	now       time.Time
	cfg       calendar.Config
	parseOpts []wen.Option
	format    string
}

func newAppContext(w io.Writer) appContext {
	cfg, _ := calendar.LoadConfig()
	var parseOpts []wen.Option
	if cfg.FiscalYearStart > 1 {
		parseOpts = append(parseOpts, wen.WithFiscalYearStart(cfg.FiscalYearStart))
	}
	return appContext{
		w:         w,
		now:       time.Now(),
		cfg:       cfg,
		parseOpts: parseOpts,
		format:    wen.DateLayout,
	}
}

// isSubcommand reports whether s is a recognized subcommand or flag.
func isSubcommand(s string) bool {
	switch s {
	case "cal", "calendar", "diff", "rel", "relative",
		"-h", "--help", "-v", "--version":
		return true
	}
	return false
}

func run(w io.Writer, args []string) error {
	ctx := newAppContext(w)

	// Extract global --format flag before subcommand routing.
	// Guard: reject known subcommand names as --format values to prevent
	// `wen --format diff ...` from silently consuming "diff".
	var remaining []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--format" {
			if i+1 >= len(args) {
				return fmt.Errorf("--format requires a value")
			}
			next := args[i+1]
			if isSubcommand(next) {
				return fmt.Errorf("--format requires a value (got subcommand %q)", next)
			}
			ctx.format = next
			i++
		} else {
			remaining = append(remaining, args[i])
		}
	}
	args = remaining

	if len(args) > 0 {
		switch args[0] {
		case "-h", "--help":
			printHelp(ctx.w)
			return nil
		case "-v", "--version":
			fmt.Fprintln(ctx.w, "wen "+version)
			return nil
		case "cal", "calendar":
			return runCalendar(ctx, args[1:])
		case "diff":
			return runDiff(ctx, args[1:])
		case "rel", "relative":
			return runRelative(ctx, args[1:])
		}
	}

	var input string
	switch {
	case len(args) > 0:
		input = strings.Join(args, " ")
	case !term.IsTerminal(int(os.Stdin.Fd())):
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			input = ""
		} else {
			input = strings.TrimSpace(scanner.Text())
		}
	}

	if input == "" {
		fmt.Fprintln(ctx.w, ctx.now.Format(ctx.format))
		return nil
	}

	results, err := wen.ParseMulti(input, ctx.now, ctx.parseOpts...)
	if err != nil {
		return fmt.Errorf("could not parse date %q: %w", input, err)
	}
	for _, r := range results {
		fmt.Fprintln(ctx.w, r.Format(ctx.format))
	}
	return nil
}

func printHelp(w io.Writer) {
	fmt.Fprint(w, `wen - a natural language date tool

Usage:
  wen                            Print today's date
  wen <natural language>         Parse a date (e.g., "next friday", "march 25 2026")
  echo "tomorrow" | wen          Parse date from stdin

Subcommands:
  wen cal, calendar [month]      Interactive calendar (e.g., wen cal march)
  wen diff <date1> <date2>       Show days between two dates
  wen rel, relative <date>       Show human-readable relative distance

Flags:
  -h, --help                     Show this help
  -v, --version                  Show version
  --format <layout>              Output format (Go time layout, default: 2006-01-02)

Calendar flags:
  --padding-top N      Top padding in lines (default: from config or 0)
  --padding-right N    Right padding in characters (default: from config or 0)
  --padding-bottom N   Bottom padding in lines (default: from config or 0)
  --padding-left N     Left padding in characters (default: from config or 0)
  --months N           Number of months to display side by side (default: 1)
  -N                   Shorthand for --months N (e.g., -3 for three months)
  --highlight-file P   Path to JSON file with dates to highlight

Diff flags:
  --weeks              Output in weeks instead of days
  --workdays           Output in workdays instead of days

Calendar keybindings:
  h/l, ←/→         Previous / next day
  j/k, ↑/↓         Next / previous week
  H/L              Previous / next month
  J/K              Next / previous year
  t                Jump to today
  w                Toggle week numbers
  ?                Toggle help bar
  Enter            Select date and print to stdout
  q, Esc, ctrl+c   Quit

Exit codes:
  0    Success (date printed)
  2    Error (parse failure, invalid input, etc.)

Config: ~/.config/wen/config.yaml
`)
}

// parseCalArgs parses calendar positional args. Supports:
//   - (empty)              → current month
//   - "march"              → March of current year
//   - "march 2027"         → March 2027
//   - "december 2026"      → December 2026
//   - any parseable date   → that month
func parseCalArgs(args []string, today time.Time, opts ...wen.Option) (time.Time, error) {
	if len(args) == 0 {
		return today, nil
	}

	words := strings.Fields(strings.Join(args, " "))

	// Try "month" or "month year" first, using the library's month lookup.
	if m, ok := wen.LookupMonth(words[0]); ok {
		year := today.Year()
		if len(words) >= 2 {
			if y, err := strconv.Atoi(words[1]); err == nil && y >= minYear && y <= maxYear {
				year = y
			}
		}
		return time.Date(year, m, 1, 0, 0, 0, 0, time.Local), nil
	}

	// Fall back to full date parsing
	parsed, err := parseDate(strings.Join(words, " "), today, opts...)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.Local), nil
}

func runCalendar(ctx appContext, args []string) error {
	// Expand shorthand -N to --months N (e.g., -3 → --months 3)
	args = expandMonthShorthand(args)

	fs := flag.NewFlagSet("cal", flag.ContinueOnError)
	paddingTop := fs.Int("padding-top", 0, "top padding (lines)")
	paddingRight := fs.Int("padding-right", 0, "right padding (characters)")
	paddingBottom := fs.Int("padding-bottom", 0, "bottom padding (lines)")
	paddingLeft := fs.Int("padding-left", 0, "left padding (characters)")
	highlightFile := fs.String("highlight-file", "", "path to JSON file with dates to highlight")
	monthCount := fs.Int("months", 1, "number of months to display side by side")
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

	// Load highlighted dates from file (priority: --highlight-file > config > default path).
	highlightPath := calendar.ResolveHighlightSource(*highlightFile, cfg.HighlightSource)
	highlightedDates := calendar.LoadHighlightedDates(highlightPath)

	var modelOpts []calendar.ModelOption
	if highlightedDates != nil {
		modelOpts = append(modelOpts, calendar.WithHighlightedDates(highlightedDates))
	}
	if *monthCount > 1 {
		modelOpts = append(modelOpts, calendar.WithMonths(*monthCount))
	}

	m := calendar.New(cursor, ctx.now, cfg, modelOpts...)
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

func runDiff(ctx appContext, args []string) error {
	// Extract flags from anywhere in the arg list so that
	// `wen diff today tomorrow --weeks` works the same as
	// `wen diff --weeks today tomorrow`.
	weeks := false
	workdays := false
	var remaining []string
	for _, a := range args {
		switch a {
		case "--weeks":
			weeks = true
		case "--workdays":
			workdays = true
		default:
			remaining = append(remaining, a)
		}
	}

	if len(remaining) < 2 {
		return fmt.Errorf("diff requires two date arguments\nusage: wen diff <date1> <date2>")
	}

	// Split remaining into two date expressions.
	// Try parsing progressively longer first-arg combinations.
	var date1, date2 time.Time
	var found bool
	for split := 1; split < len(remaining); split++ {
		d1, err1 := parseDate(strings.Join(remaining[:split], " "), ctx.now, ctx.parseOpts...)
		d2, err2 := parseDate(strings.Join(remaining[split:], " "), ctx.now, ctx.parseOpts...)
		if err1 == nil && err2 == nil {
			date1, date2 = d1, d2
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("could not parse date arguments: %s", strings.Join(remaining, " "))
	}

	// Normalize to UTC to avoid DST-related hour differences when computing calendar days.
	d1 := time.Date(date1.Year(), date1.Month(), date1.Day(), 0, 0, 0, 0, time.UTC)
	d2 := time.Date(date2.Year(), date2.Month(), date2.Day(), 0, 0, 0, 0, time.UTC)

	switch {
	case weeks:
		totalDays := int(d2.Sub(d1).Hours() / 24)
		if totalDays < 0 {
			totalDays = -totalDays
		}
		w := totalDays / 7
		rem := totalDays % 7
		if rem != 0 {
			fmt.Fprintf(ctx.w, "%d weeks, %d days\n", w, rem)
		} else {
			fmt.Fprintf(ctx.w, "%d weeks\n", w)
		}
	case workdays:
		wd := countWorkdays(d1, d2)
		fmt.Fprintf(ctx.w, "%d workdays\n", wd)
	default:
		totalDays := int(d2.Sub(d1).Hours() / 24)
		if totalDays < 0 {
			totalDays = -totalDays
		}
		fmt.Fprintf(ctx.w, "%d days\n", totalDays)
	}
	return nil
}

func countWorkdays(start, end time.Time) int {
	if start.After(end) {
		start, end = end, start
	}
	totalDays := int(end.Sub(start).Hours() / 24)
	if totalDays == 0 {
		return 0
	}
	fullWeeks := totalDays / 7
	remaining := totalDays % 7
	workdays := fullWeeks * 5
	// Count workdays in the partial week starting from start's weekday.
	startDow := int(start.Weekday()) // Sunday=0
	for i := range remaining {
		dow := (startDow + i) % 7
		if dow != int(time.Saturday) && dow != int(time.Sunday) {
			workdays++
		}
	}
	return workdays
}

func runRelative(ctx appContext, args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(ctx.w, "today")
		return nil
	}

	input := strings.Join(args, " ")
	t, err := wen.ParseRelative(input, ctx.now, ctx.parseOpts...)
	if err != nil {
		return fmt.Errorf("could not parse date %q: %w", input, err)
	}
	today := time.Date(ctx.now.Year(), ctx.now.Month(), ctx.now.Day(), 0, 0, 0, 0, time.UTC)
	target := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	days := int(target.Sub(today).Hours() / 24)

	switch {
	case days == 0:
		fmt.Fprintln(ctx.w, "today")
	case days == 1:
		fmt.Fprintln(ctx.w, "tomorrow")
	case days == -1:
		fmt.Fprintln(ctx.w, "yesterday")
	case days > 1:
		fmt.Fprintf(ctx.w, "in %d days\n", days)
	default:
		fmt.Fprintf(ctx.w, "%d days ago\n", -days)
	}
	return nil
}

// expandMonthShorthand converts -N args to --months N (e.g., -3 → --months 3).
func expandMonthShorthand(args []string) []string {
	var result []string
	for _, arg := range args {
		if len(arg) >= 2 && arg[0] == '-' && arg[1] != '-' {
			if n, err := strconv.Atoi(arg[1:]); err == nil && n > 0 {
				result = append(result, "--months", strconv.Itoa(n))
				continue
			}
		}
		result = append(result, arg)
	}
	return result
}

func parseDate(input string, ref time.Time, opts ...wen.Option) (time.Time, error) {
	t, err := wen.ParseRelative(input, ref, opts...)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse date %q: %w", input, err)
	}
	return t, nil
}
