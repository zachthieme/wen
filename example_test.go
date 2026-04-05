package wen_test

import (
	"fmt"
	"time"

	"github.com/zachthieme/wen"
)

func ExampleParse() {
	// Parse returns a date relative to time.Now().
	t, err := wen.Parse("next friday")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(t.Format(wen.DateLayout))
}

func ExampleParseRelative() {
	ref := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)

	t, err := wen.ParseRelative("tomorrow", ref)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(t.Format(wen.DateLayout))
	// Output: 2026-03-19
}

func ExampleParseRelative_withTime() {
	ref := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)

	t, err := wen.ParseRelative("march 25 at 3pm", ref)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(t.Format("2006-01-02 15:04"))
	// Output: 2026-03-25 15:00
}

func ExampleParseRelative_periodMode() {
	ref := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC) // Wednesday

	// PeriodStart (default): "next week" = Sunday of next week
	start, _ := wen.ParseRelative("next week", ref, wen.WithPeriodStart())

	// PeriodSame: "next week" = same weekday + 7 days
	same, _ := wen.ParseRelative("next week", ref, wen.WithPeriodSame())

	fmt.Println(start.Format(wen.DateLayout))
	fmt.Println(same.Format(wen.DateLayout))
	// Output:
	// 2026-03-22
	// 2026-03-25
}

func ExampleParseMulti() {
	ref := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)

	dates, err := wen.ParseMulti("every friday in april", ref)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, d := range dates {
		fmt.Println(d.Format(wen.DateLayout))
	}
	// Output:
	// 2026-04-03
	// 2026-04-10
	// 2026-04-17
	// 2026-04-24
}

func ExampleParseMulti_singleDate() {
	ref := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)

	// ParseMulti falls back to single-date parsing for non-multi expressions.
	dates, err := wen.ParseMulti("next friday", ref)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(len(dates), "date(s):", dates[0].Format(wen.DateLayout))
	// Output: 1 date(s): 2026-03-27
}

func ExampleCountWorkdays() {
	start := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC) // Monday
	end := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)   // next Monday

	fmt.Println(wen.CountWorkdays(start, end))
	// Output: 5
}

func ExampleFiscalQuarter() {
	// Standard calendar year (start = January)
	q, fy := wen.FiscalQuarter(3, 2026, 1)
	fmt.Printf("March 2026 (cal year): Q%d FY%d\n", q, fy)

	// Federal fiscal year (start = October)
	q, fy = wen.FiscalQuarter(3, 2026, 10)
	fmt.Printf("March 2026 (Oct FY):   Q%d FY%d\n", q, fy)
	// Output:
	// March 2026 (cal year): Q1 FY2026
	// March 2026 (Oct FY):   Q2 FY2026
}

func ExampleTruncateDay() {
	t := time.Date(2026, 3, 18, 14, 30, 45, 0, time.UTC)
	fmt.Println(wen.TruncateDay(t).Format("2006-01-02 15:04:05"))
	// Output: 2026-03-18 00:00:00
}

func ExampleDaysIn() {
	fmt.Println(wen.DaysIn(2026, time.February, time.UTC))
	fmt.Println(wen.DaysIn(2024, time.February, time.UTC)) // leap year
	// Output:
	// 28
	// 29
}

func ExampleLookupMonth() {
	m, ok := wen.LookupMonth("mar")
	fmt.Println(m, ok)

	_, ok = wen.LookupMonth("pizza")
	fmt.Println(ok)
	// Output:
	// March true
	// false
}

func ExampleWithFiscalYearStart() {
	ref := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)

	// Q1 starts in October → March is Q2
	t, err := wen.ParseRelative("beginning of quarter", ref, wen.WithFiscalYearStart(10))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(t.Format(wen.DateLayout))
	// Output: 2026-01-01
}
