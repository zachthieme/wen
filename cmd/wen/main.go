// Command wen is a natural language date tool that parses human-readable date
// expressions and provides an interactive terminal calendar.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/zachthieme/wen"
	"github.com/zachthieme/wen/calendar"

	"golang.org/x/term"
)

var version = "dev"

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
	cfg, warnings := calendar.LoadConfig()
	for _, warn := range warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", warn)
	}
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

func parseDate(input string, ref time.Time, opts ...wen.Option) (time.Time, error) {
	t, err := wen.ParseRelative(input, ref, opts...)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse date %q: %w", input, err)
	}
	return t, nil
}
