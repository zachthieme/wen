package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"wen/calendar"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
	"golang.org/x/term"
)

const dateLayout = "2006-01-02"

func main() {
	args := os.Args[1:]

	if len(args) > 0 && args[0] == "cal" {
		runCalendar(args[1:])
		return
	}

	var input string
	switch {
	case len(args) > 0:
		input = strings.Join(args, " ")
	case !term.IsTerminal(int(os.Stdin.Fd())):
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "error: failed to read from stdin")
				os.Exit(1)
			}
			input = ""
		} else {
			input = strings.TrimSpace(scanner.Text())
		}
	}

	if input == "" {
		fmt.Println(time.Now().Format(dateLayout))
		return
	}

	result := parseDate(input, time.Now())
	fmt.Println(result.Format(dateLayout))
}

func parseDate(input string, ref time.Time) time.Time {
	w := when.New(nil)
	w.Add(en.All...)
	w.Add(common.All...)

	result, err := w.Parse(input, ref)
	if err != nil || result == nil {
		fmt.Fprintf(os.Stderr, "error: could not parse date %q\n", input)
		os.Exit(1)
	}
	return result.Time
}

func runCalendar(args []string) {
	today := time.Now()
	cursor := today

	if len(args) > 0 {
		input := strings.Join(args, " ")
		parsed := parseDate(input, today)
		cursor = time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.Local)
	}

	cfg := calendar.LoadConfig()
	m := calendar.New(cursor, today, cfg)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	result := finalModel.(calendar.Model)
	if result.Selected {
		fmt.Println(result.Cursor.Format(dateLayout))
	} else {
		os.Exit(1)
	}
}
