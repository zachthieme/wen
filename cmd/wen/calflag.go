package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zachthieme/wen/calendar"
	"golang.org/x/term"
)

// calendarFlags holds flags shared by the cal and row subcommands.
type calendarFlags struct {
	highlightFile string
	printMode     bool
	julian        bool
}

// register adds the shared calendar flags to a FlagSet.
func (f *calendarFlags) register(fs *flag.FlagSet) {
	fs.StringVar(&f.highlightFile, "highlight-file", "", "path to JSON file with dates to highlight")
	fs.BoolVar(&f.printMode, "print", false, "print and exit (non-interactive)")
	fs.BoolVar(&f.printMode, "p", false, "shorthand for --print")
	fs.BoolVar(&f.julian, "julian", false, "show Julian day-of-year numbers")
	fs.BoolVar(&f.julian, "j", false, "shorthand for --julian")
}

// resolve computes derived values from flags and config. It prints config
// normalization warnings to stderr and returns the resolved values.
func (f *calendarFlags) resolve(cfg calendar.Config) (highlightPath string, julian, printMode bool) {
	for _, w := range cfg.Normalize() {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}
	highlightPath = calendar.ResolveHighlightSource(f.highlightFile, cfg.HighlightSource)
	julian = cfg.Julian || f.julian
	printMode = f.printMode || !term.IsTerminal(int(os.Stdout.Fd()))
	return
}
