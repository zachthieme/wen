package main

import (
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

const (
	minYear = 1000 // lowest value accepted as a calendar year in parseCalArgs
	maxYear = 9999 // highest value accepted as a calendar year in parseCalArgs
)

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

	// Print config warnings to stderr.
	for _, w := range cfg.Normalize() {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

	// Resolve highlight source (priority: --highlight-file > config > default path).
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
