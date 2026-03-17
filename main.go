package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
	"golang.org/x/term"
)

func main() {
	nowFlag := flag.String("now", "", "override reference date (yyyy-mm-dd)")
	flag.Parse()

	ref := time.Now()
	if *nowFlag != "" {
		parsed, err := time.Parse("2006-01-02", *nowFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: invalid --now date %q, expected yyyy-mm-dd\n", *nowFlag)
			os.Exit(1)
		}
		ref = parsed
	}

	args := flag.Args()

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
			// EOF with no data: behave as if no input was provided
			fmt.Println(ref.Format("2006-01-02"))
			return
		}
		input = strings.TrimSpace(scanner.Text())
	default:
		fmt.Println(ref.Format("2006-01-02"))
		return
	}

	w := when.New(nil)
	w.Add(en.All...)
	w.Add(common.All...)

	result, err := w.Parse(input, ref)
	if err != nil || result == nil {
		fmt.Fprintf(os.Stderr, "error: could not parse date %q\n", input)
		os.Exit(1)
	}

	fmt.Println(result.Time.Format("2006-01-02"))
}
