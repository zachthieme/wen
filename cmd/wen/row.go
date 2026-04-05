package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/zachthieme/wen"
	"github.com/zachthieme/wen/calendar"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

func runRow(ctx appContext, args []string) error {
	fs := flag.NewFlagSet("row", flag.ContinueOnError)
	highlightFile := fs.String("highlight-file", "", "path to JSON file with dates to highlight")
	printFlag := fs.Bool("print", false, "print strip calendar and exit (non-interactive)")
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

	highlightPath := calendar.ResolveHighlightSource(*highlightFile, cfg.HighlightSource)

	// Resolve julian: CLI flag overrides config
	julian := cfg.Julian || *julianFlag

	// Determine print mode: explicit flag or non-TTY stdout
	printMode := *printFlag || !term.IsTerminal(int(os.Stdout.Fd()))

	var modelOpts []calendar.RowModelOption
	if highlightPath != "" {
		modelOpts = append(modelOpts, calendar.WithRowHighlightSource(highlightPath))
	}
	if julian {
		modelOpts = append(modelOpts, calendar.WithRowJulian(true))
	}
	if printMode {
		modelOpts = append(modelOpts, calendar.WithRowPrintMode(true))
	}

	m := calendar.NewRow(cursor, ctx.now, cfg, modelOpts...)

	if printMode {
		width := 80
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
			width = w
		}
		m = m.WithTermWidth(width)
		fmt.Fprint(ctx.w, m.View())
		return nil
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("row: %w", err)
	}

	row, ok := finalModel.(calendar.RowModel)
	if !ok {
		return fmt.Errorf("unexpected internal state")
	}
	if row.InRange() {
		fmt.Fprintln(ctx.w, row.RangeStart().Format(wen.DateLayout))
		fmt.Fprintln(ctx.w, row.RangeEnd().Format(wen.DateLayout))
	} else if row.Selected() {
		fmt.Fprintln(ctx.w, row.Cursor().Format(wen.DateLayout))
	}
	return nil
}
