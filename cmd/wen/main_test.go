package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zachthieme/wen"
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
	cmd := exec.Command("go", "build", "-o", testBinary, "github.com/zachthieme/wen/cmd/wen")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %s\n%s", err, out)
		return 1
	}

	return m.Run()
}

func TestNoArgs_PrintsToday(t *testing.T) {
	today := time.Now().Format(wen.DateLayout)
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
	if _, err := time.Parse(wen.DateLayout, got); err != nil {
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
	if _, err := time.Parse(wen.DateLayout, got); err != nil {
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
		{"last sunday", "2026-03-08"},
		{"last sat", "2026-03-14"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseDate(tt.input, ref)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if got.Format("2006-01-02") != tt.want {
				t.Errorf("parseDate(%q) = %s, want %s", tt.input, got.Format("2006-01-02"), tt.want)
			}
		})
	}
}

func TestParseDateRejectsInvalidInput(t *testing.T) {
	ref := time.Date(2026, 3, 17, 12, 0, 0, 0, time.Local)
	inputs := []string{"pizza", "next", "this"}
	for _, input := range inputs {
		if _, err := parseDate(input, ref); err == nil {
			t.Errorf("expected error for %q", input)
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

func TestCalFlagParsing(t *testing.T) {
	fs := flag.NewFlagSet("cal", flag.ContinueOnError)
	paddingTop := fs.Int("padding-top", 0, "")
	paddingRight := fs.Int("padding-right", 0, "")
	paddingBottom := fs.Int("padding-bottom", 0, "")
	paddingLeft := fs.Int("padding-left", 0, "")

	err := fs.Parse([]string{"--padding-top", "3", "--padding-left", "2", "march", "2026"})
	if err != nil {
		t.Fatal(err)
	}
	if *paddingTop != 3 {
		t.Errorf("expected padding-top 3, got %d", *paddingTop)
	}
	if *paddingRight != 0 {
		t.Errorf("expected padding-right 0 (default), got %d", *paddingRight)
	}
	if *paddingBottom != 0 {
		t.Errorf("expected padding-bottom 0 (default), got %d", *paddingBottom)
	}
	if *paddingLeft != 2 {
		t.Errorf("expected padding-left 2, got %d", *paddingLeft)
	}
	remaining := strings.Join(fs.Args(), " ")
	if remaining != "march 2026" {
		t.Errorf("expected remaining args 'march 2026', got %q", remaining)
	}

	// Verify Visit() only sees explicitly-set flags.
	var visited []string
	fs.Visit(func(f *flag.Flag) { visited = append(visited, f.Name) })
	if len(visited) != 2 {
		t.Errorf("expected 2 visited flags, got %d: %v", len(visited), visited)
	}
}
