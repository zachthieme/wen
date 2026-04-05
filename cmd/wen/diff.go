package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/zachthieme/wen"
)

func runDiff(ctx appContext, args []string) error {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	weeks := fs.Bool("weeks", false, "output in weeks instead of days")
	workdays := fs.Bool("workdays", false, "output in workdays instead of days")

	// Partition args so flags can appear anywhere: before, between, or
	// after the date arguments (e.g., "wen diff today tomorrow --weeks").
	var flags, positional []string
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			flags = append(flags, a)
		} else {
			positional = append(positional, a)
		}
	}
	if err := fs.Parse(flags); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if len(positional) < 2 {
		return fmt.Errorf("diff requires two date arguments\nusage: wen diff <date1> <date2>")
	}

	// Split remaining into two date expressions using greedy-left search:
	// try parsing progressively longer first-arg combinations until both
	// halves parse successfully. This is correct because the parser requires
	// structure (e.g., a unit after a bare number), so the earliest valid
	// split point is always unambiguous — "3 days ago tomorrow" can only
	// split as ["3 days ago", "tomorrow"], never ["3", "days ago tomorrow"].
	var date1, date2 time.Time
	var found bool
	for split := 1; split < len(positional); split++ {
		d1, err1 := parseDate(strings.Join(positional[:split], " "), ctx.now, ctx.parseOpts...)
		d2, err2 := parseDate(strings.Join(positional[split:], " "), ctx.now, ctx.parseOpts...)
		if err1 == nil && err2 == nil {
			date1, date2 = d1, d2
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("could not parse date arguments: %s", strings.Join(positional, " "))
	}

	// Normalize to UTC to avoid DST-related hour differences when computing calendar days.
	d1 := time.Date(date1.Year(), date1.Month(), date1.Day(), 0, 0, 0, 0, time.UTC)
	d2 := time.Date(date2.Year(), date2.Month(), date2.Day(), 0, 0, 0, 0, time.UTC)

	switch {
	case *weeks:
		totalDays := int(d2.Sub(d1).Hours() / 24)
		if totalDays < 0 {
			totalDays = -totalDays
		}
		w := totalDays / 7
		rem := totalDays % 7
		if rem != 0 {
			fmt.Fprintf(ctx.w, "%d %s, %d %s\n", w, plural(w, "week"), rem, plural(rem, "day"))
		} else {
			fmt.Fprintf(ctx.w, "%d %s\n", w, plural(w, "week"))
		}
	case *workdays:
		wd := wen.CountWorkdays(d1, d2)
		fmt.Fprintf(ctx.w, "%d %s\n", wd, plural(wd, "workday"))
	default:
		totalDays := int(d2.Sub(d1).Hours() / 24)
		if totalDays < 0 {
			totalDays = -totalDays
		}
		fmt.Fprintf(ctx.w, "%d %s\n", totalDays, plural(totalDays, "day"))
	}
	return nil
}

func plural(n int, singular string) string {
	if n == 1 {
		return singular
	}
	return singular + "s"
}
