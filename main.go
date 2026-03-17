package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
	"golang.org/x/term"
)

const dateLayout = "2006-01-02"

func main() {
	args := os.Args[1:]

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

	w := when.New(nil)
	w.Add(en.All...)
	w.Add(common.All...)

	result, err := w.Parse(input, time.Now())
	if err != nil || result == nil {
		fmt.Fprintf(os.Stderr, "error: could not parse date %q\n", input)
		os.Exit(1)
	}

	fmt.Println(result.Time.Format(dateLayout))
}
