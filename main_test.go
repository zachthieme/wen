package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zachthieme/wen/calendar"
)

var testBinary string

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	dir, err := os.MkdirTemp("", "wen-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %s\n", err)
		return 1
	}
	defer func() { _ = os.RemoveAll(dir) }()

	testBinary = filepath.Join(dir, "wen")
	cmd := exec.Command("go", "build", "-o", testBinary, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %s\n%s", err, out)
		return 1
	}

	return m.Run()
}

func TestNoArgs_PrintsToday(t *testing.T) {
	today := time.Now().Format(calendar.DateLayout)
	cmd := exec.Command(testBinary)
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	if got != today {
		t.Errorf("got %q, want %q", got, today)
	}
}

func TestPositionalArg(t *testing.T) {
	cmd := exec.Command(testBinary, "next friday")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	if _, err := time.Parse(calendar.DateLayout, got); err != nil {
		t.Errorf("output %q is not a valid yyyy-mm-dd date", got)
	}
}

func TestStdinMode(t *testing.T) {
	cmd := exec.Command(testBinary)
	cmd.Stdin = strings.NewReader("next friday\n")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	if _, err := time.Parse(calendar.DateLayout, got); err != nil {
		t.Errorf("output %q is not a valid yyyy-mm-dd date", got)
	}
}

func TestParseFailure(t *testing.T) {
	cmd := exec.Command(testBinary, "pizza")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit code")
	}
	got := string(out)
	if !strings.Contains(got, `could not parse date "pizza"`) {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestThisVsNextWeekday(t *testing.T) {
	// Reference: Tuesday March 17, 2026
	ref := time.Date(2026, 3, 17, 12, 0, 0, 0, time.Local)

	tests := []struct {
		input string
		want  string
	}{
		{"this thursday", "2026-03-19"},
		{"next thursday", "2026-03-26"},
		{"this tuesday", "2026-03-17"},
		{"next tuesday", "2026-03-24"},
		{"this monday", "2026-03-16"},
		{"next monday", "2026-03-23"},
		{"this friday", "2026-03-20"},
		{"next friday", "2026-03-27"},
		{"this sunday", "2026-03-15"},
		{"next sunday", "2026-03-22"},
		{"this saturday", "2026-03-21"},
		{"next saturday", "2026-03-28"},
		{"this thu", "2026-03-19"},
		{"next fri", "2026-03-27"},
		{"last thursday", "2026-03-12"},
		{"last tuesday", "2026-03-10"},
		{"last sunday", "2026-03-15"},
		{"last sat", "2026-03-14"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := parseRelativeWeekday(tt.input, ref)
			if !ok {
				t.Fatalf("expected match for %q", tt.input)
			}
			if got.Format(calendar.DateLayout) != tt.want {
				t.Errorf("parseRelativeWeekday(%q) = %s, want %s", tt.input, got.Format(calendar.DateLayout), tt.want)
			}
		})
	}
}

func TestRelativeWeekdayDoesNotMatchOtherInputs(t *testing.T) {
	ref := time.Date(2026, 3, 17, 12, 0, 0, 0, time.Local)
	inputs := []string{"tomorrow", "2 weeks ago", "march 20th", "pizza", "next", "this"}
	for _, input := range inputs {
		if _, ok := parseRelativeWeekday(input, ref); ok {
			t.Errorf("expected no match for %q", input)
		}
	}
}

func TestPositionalArgTakesPrecedenceOverStdin(t *testing.T) {
	cmd := exec.Command(testBinary, "march 25 2026")
	cmd.Stdin = strings.NewReader("next friday\n")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	want := "2026-03-25"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
