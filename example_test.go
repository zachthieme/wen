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
