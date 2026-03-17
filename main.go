package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/zachthieme/wen/calendar"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
	"golang.org/x/term"
)

var (
	version        = "dev"
	errNoSelection = errors.New("no date selected")
)

// Package-level parser — initialized once, reused on every call.
var dateParser *when.Parser

func init() {
	dateParser = when.New(nil)
	dateParser.Add(en.All...)
	dateParser.Add(common.All...)
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, errNoSelection) {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
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
				return fmt.Errorf("failed to read from stdin")
			}
			input = ""
		} else {
			input = strings.TrimSpace(scanner.Text())
		}
	}

	if input == "" {
		fmt.Println(time.Now().Format(calendar.DateLayout))
		return nil
	}

	result, err := parseDate(input, time.Now())
	if err != nil {
		return err
	}
	fmt.Println(result.Format(calendar.DateLayout))
	return nil
}

var weekdays = map[string]time.Weekday{
	"sunday": time.Sunday, "sun": time.Sunday,
	"monday": time.Monday, "mon": time.Monday,
	"tuesday": time.Tuesday, "tue": time.Tuesday,
	"wednesday": time.Wednesday, "wed": time.Wednesday,
	"thursday": time.Thursday, "thu": time.Thursday,
	"friday": time.Friday, "fri": time.Friday,
	"saturday": time.Saturday, "sat": time.Saturday,
}

// parseRelativeWeekday handles "this/next/last <weekday>" patterns.
// Returns the resolved date and true if matched, or zero time and false.
func parseRelativeWeekday(input string, ref time.Time) (time.Time, bool) {
	lower := strings.ToLower(strings.TrimSpace(input))
	parts := strings.Fields(lower)
	if len(parts) != 2 {
		return time.Time{}, false
	}

	prefix := parts[0]
	if prefix != "this" && prefix != "next" && prefix != "last" {
		return time.Time{}, false
	}

	target, ok := weekdays[parts[1]]
	if !ok {
		return time.Time{}, false
	}

	refDay := ref.Weekday()
	switch prefix {
	case "this":
		diff := int(target) - int(refDay)
		return ref.AddDate(0, 0, diff), true
	case "next":
		// Advance to start of next week (Sunday), then offset by target weekday.
		daysToNextSunday := (7 - int(refDay)) % 7
		if daysToNextSunday == 0 {
			daysToNextSunday = 7 // if ref is Sunday, "next" means the following week
		}
		diff := daysToNextSunday + int(target)
		return ref.AddDate(0, 0, diff), true
	case "last":
		diff := int(refDay) - int(target)
		if diff <= 0 {
			diff += 7
		}
		return ref.AddDate(0, 0, -diff), true
	}

	return time.Time{}, false
}

func parseDate(input string, ref time.Time) (time.Time, error) {
	if t, ok := parseRelativeWeekday(input, ref); ok {
		return t, nil
	}

	result, err := dateParser.Parse(input, ref)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse date %q: %w", input, err)
	}
	if result == nil {
		return time.Time{}, fmt.Errorf("could not parse date %q", input)
	}
	return result.Time, nil
}

func printHelp() {
	fmt.Print(`wen - a natural language date tool

Usage:
  wen                        Print today's date
  wen <natural language>     Parse a date (e.g., "next friday", "march 25 2026")
  echo "tomorrow" | wen      Parse date from stdin
  wen cal                    Interactive calendar at current month
  wen cal <month/date>       Interactive calendar at a specific month

Flags:
  -h, --help       Show this help
  -v, --version    Show version

Calendar keybindings:
  h/l, arrows      Previous / next day
  j/k, arrows      Next / previous week
  H/L              Previous / next month
  J/K              Next / previous year
  t                Jump to today
  w                Toggle week numbers
  y                Yank date to clipboard
  ?                Toggle help bar
  Enter            Select date and exit
  q, Esc           Quit without selecting

Config: ~/.config/wen/config.yaml
`)
}

func runCalendar(args []string) error {
	today := time.Now()
	cursor := today

	if len(args) > 0 {
		input := strings.Join(args, " ")
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
	m := calendar.New(cursor, today, cfg)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("calendar: %w", err)
	}

	result, ok := finalModel.(calendar.Model)
	if !ok {
		return fmt.Errorf("unexpected internal state")
	}
	if result.IsSelected() {
		fmt.Println(result.Cursor().Format(calendar.DateLayout))
		return nil
	}
	return errNoSelection
}
