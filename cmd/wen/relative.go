package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/zachthieme/wen"
)

func runRelative(ctx appContext, args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(ctx.w, "today")
		return nil
	}

	input := strings.Join(args, " ")
	t, err := wen.ParseRelative(input, ctx.now, ctx.parseOpts...)
	if err != nil {
		return fmt.Errorf("could not parse date %q: %w", input, err)
	}
	today := time.Date(ctx.now.Year(), ctx.now.Month(), ctx.now.Day(), 0, 0, 0, 0, time.UTC)
	target := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	days := int(target.Sub(today).Hours() / 24)

	switch {
	case days == 0:
		fmt.Fprintln(ctx.w, "today")
	case days == 1:
		fmt.Fprintln(ctx.w, "tomorrow")
	case days == -1:
		fmt.Fprintln(ctx.w, "yesterday")
	case days > 1:
		fmt.Fprintf(ctx.w, "in %d days\n", days)
	default:
		fmt.Fprintf(ctx.w, "%d days ago\n", -days)
	}
	return nil
}
