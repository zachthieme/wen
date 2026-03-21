// Command wen is a natural language date tool that parses human-readable date
// expressions and provides an interactive terminal calendar.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
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

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func run(args []string) error {
	// Extract global flags (--format, --relative) before subcommand routing.
	format := wen.DateLayout
	relative := false
	var remaining []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			} else {
				return fmt.Errorf("--format requires a value")
			}
		case "--relative":
			relative = true
		default:
			remaining = append(remaining, args[i])
		}
	}
	args = remaining

	if len(args) > 0 {
		switch args[0] {
		case "-h", "--help":
			printHelp()
			return nil
		case "-v", "--version":
			fmt.Println("wen " + version)
			return nil
		case "cal":
			return runCalendar(args[1:])
		case "diff":
			return runDiff(args[1:], format)
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

	now := time.Now()

	// Load config for fiscal year start (affects quarter calculations).
	cfg, _ := calendar.LoadConfig()
	var parseOpts []wen.Option
	if cfg.FiscalYearStart > 1 {
		parseOpts = append(parseOpts, wen.WithFiscalYearStart(cfg.FiscalYearStart))
	}

	if input == "" {
		if relative {
			fmt.Println("today")
			return nil
		}
		fmt.Println(now.Format(format))
		return nil
	}

	if relative {
		return runRelative(input, now, parseOpts...)
	}

	// Try multi-date parse first (e.g., "every friday in april")
	results, err := wen.ParseMulti(input, now, parseOpts...)
	if err != nil {
		return fmt.Errorf("could not parse date %q: %w", input, err)
	}
	for _, r := range results {
		fmt.Println(r.Format(format))
	}
	return nil
}

func printHelp() {
	fmt.Print(`wen - a natural language date tool

Usage:
  wen                        Print today's date
  wen <natural language>     Parse a date (e.g., "next friday", "march 25 2026")
  echo "tomorrow" | wen      Parse date from stdin
  wen cal                    Interactive calendar at current month
  wen cal <month/date>       Interactive calendar at a specific month
  wen diff <date1> <date2>   Show days between two dates

Flags:
  -h, --help                 Show this help
  -v, --version              Show version
  --format <layout>          Output format (Go time layout, default: 2006-01-02)
  --relative                 Output human-readable relative string for a date

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

func runCalendar(args []string) error {
	today := time.Now()

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

	cursor := today
	remaining := fs.Args()
	if len(remaining) > 0 {
		input := strings.Join(remaining, " ")
		parsed, err := parseDate(input, today)
		if err != nil {
			return err
		}
		cursor = time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.Local)
	}

	cfg, warnings := calendar.LoadConfig()
	for _, w := range warnings {
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

	m := calendar.New(cursor, today, cfg, modelOpts...)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("calendar: %w", err)
	}

	cal, ok := finalModel.(calendar.Model)
	if !ok {
		return fmt.Errorf("unexpected internal state")
	}
	if cal.Selected() {
		fmt.Println(cal.Cursor().Format(wen.DateLayout))
	}
	return nil
}

func runDiff(args []string, format string) error {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	weeks := fs.Bool("weeks", false, "output in weeks")
	workdays := fs.Bool("workdays", false, "output in workdays")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	remaining := fs.Args()
	if len(remaining) < 2 {
		return fmt.Errorf("diff requires two date arguments\nusage: wen diff <date1> <date2>")
	}

	// Split remaining into two date expressions.
	// Try parsing progressively longer first-arg combinations.
	now := time.Now()
	var date1, date2 time.Time
	var found bool
	for split := 1; split < len(remaining); split++ {
		d1, err1 := parseDate(strings.Join(remaining[:split], " "), now)
		d2, err2 := parseDate(strings.Join(remaining[split:], " "), now)
		if err1 == nil && err2 == nil {
			date1, date2 = d1, d2
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("could not parse date arguments: %s", strings.Join(remaining, " "))
	}

	d1 := time.Date(date1.Year(), date1.Month(), date1.Day(), 0, 0, 0, 0, date1.Location())
	d2 := time.Date(date2.Year(), date2.Month(), date2.Day(), 0, 0, 0, 0, date2.Location())

	switch {
	case *weeks:
		totalDays := int(d2.Sub(d1).Hours() / 24)
		w := totalDays / 7
		rem := totalDays % 7
		if rem != 0 {
			fmt.Printf("%d weeks, %d days\n", w, rem)
		} else {
			fmt.Printf("%d weeks\n", w)
		}
	case *workdays:
		wd := countWorkdays(d1, d2)
		fmt.Printf("%d workdays\n", wd)
	default:
		totalDays := int(d2.Sub(d1).Hours() / 24)
		if totalDays < 0 {
			totalDays = -totalDays
		}
		fmt.Printf("%d days\n", totalDays)
	}
	return nil
}

func countWorkdays(start, end time.Time) int {
	if start.After(end) {
		start, end = end, start
	}
	count := 0
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		wd := d.Weekday()
		if wd != time.Saturday && wd != time.Sunday {
			count++
		}
	}
	return count
}

func runRelative(input string, ref time.Time, opts ...wen.Option) error {
	t, err := wen.ParseRelative(input, ref, opts...)
	if err != nil {
		return fmt.Errorf("could not parse date %q: %w", input, err)
	}
	today := time.Date(ref.Year(), ref.Month(), ref.Day(), 0, 0, 0, 0, ref.Location())
	target := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	days := int(target.Sub(today).Hours() / 24)

	switch {
	case days == 0:
		fmt.Println("today")
	case days == 1:
		fmt.Println("tomorrow")
	case days == -1:
		fmt.Println("yesterday")
	case days > 1:
		fmt.Printf("in %d days\n", days)
	default:
		fmt.Printf("%d days ago\n", -days)
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

func parseDate(input string, ref time.Time) (time.Time, error) {
	t, err := wen.ParseRelative(input, ref)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse date %q: %w", input, err)
	}
	return t, nil
}
