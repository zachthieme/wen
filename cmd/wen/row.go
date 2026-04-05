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
	var cf calendarFlags
	cf.register(fs)
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

	highlightPath, julian, printMode := cf.resolve(ctx.cfg)

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

	m := calendar.NewRow(cursor, ctx.now, ctx.cfg, modelOpts...)
	for _, w := range m.Warnings() {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

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
