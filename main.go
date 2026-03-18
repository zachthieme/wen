// Command wen is a natural language date tool that parses human-readable date
// expressions and provides an interactive terminal calendar.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

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
				return fmt.Errorf("failed to read from stdin: %w", err)
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

Calendar flags:
  --padding-top N      Top padding in lines (default: from config or 0)
  --padding-right N    Right padding in characters (default: from config or 0)
  --padding-bottom N   Bottom padding in lines (default: from config or 0)
  --padding-left N     Left padding in characters (default: from config or 0)

Calendar keybindings:
  h/l, ←/→         Previous / next day
  j/k, ↑/↓         Next / previous week
  H/L              Previous / next month
  J/K              Next / previous year
  t                Jump to today
  w                Toggle week numbers
  ?                Toggle help bar
  q, Esc           Quit

Exit codes:
  0    Success (date printed)
  2    Error (parse failure, invalid input, etc.)

Config: ~/.config/wen/config.yaml
`)
}

func runCalendar(args []string) error {
	today := time.Now()

	fs := flag.NewFlagSet("cal", flag.ContinueOnError)
	paddingTop := fs.Int("padding-top", 0, "top padding (lines)")
	paddingRight := fs.Int("padding-right", 0, "right padding (characters)")
	paddingBottom := fs.Int("padding-bottom", 0, "bottom padding (lines)")
	paddingLeft := fs.Int("padding-left", 0, "left padding (characters)")
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

	m := calendar.New(cursor, today, cfg)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("calendar: %w", err)
	}

	if _, ok := finalModel.(calendar.Model); !ok {
		return fmt.Errorf("unexpected internal state")
	}
	return nil
}
