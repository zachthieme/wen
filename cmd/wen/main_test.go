package main

import (
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

	// Isolate tests from the user's real config so results are deterministic.
	if err := os.Setenv("XDG_CONFIG_HOME", dir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set XDG_CONFIG_HOME: %s\n", err)
		return 1
	}

	testBinary = filepath.Join(dir, "wen")
	cmd := exec.Command("go", "build", "-o", testBinary, "github.com/zachthieme/wen/cmd/wen")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %s\n%s", err, out)
		return 1
	}

	return m.Run()
}

func TestNoArgs_PrintsToday(t *testing.T) {
	// Capture time before and after to handle midnight boundary.
	before := time.Now().Format(wen.DateLayout)
	cmd := exec.Command(testBinary)
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	after := time.Now().Format(wen.DateLayout)
	got := strings.TrimSpace(string(out))
	if got != before && got != after {
		t.Errorf("got %q, want %q or %q", got, before, after)
	}
}

func TestPositionalArg(t *testing.T) {
	cmd := exec.Command(testBinary, "next", "friday")
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

func TestHelpFlag(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "--help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := string(out)
	if !strings.Contains(got, "wen - a natural language date tool") {
		t.Errorf("help output missing header: %s", got)
	}
	if !strings.Contains(got, "cal, calendar") {
		t.Errorf("help output missing cal subcommand: %s", got)
	}
	if !strings.Contains(got, "rel, relative") {
		t.Errorf("help output missing rel subcommand: %s", got)
	}
}

func TestVersionFlag(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "--version")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	if !strings.HasPrefix(got, "wen ") {
		t.Errorf("version output should start with 'wen ', got %q", got)
	}
}

func TestMultiWordArgs(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "march", "25", "2026")
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

func TestMultipleParseErrors(t *testing.T) {
	t.Parallel()
	inputs := []string{"xyzzy", "42 blobs ago", "next flurb"}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			cmd := exec.Command(testBinary, input)
			out, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("expected non-zero exit for %q", input)
			}
			if !strings.Contains(string(out), "could not parse date") {
				t.Errorf("unexpected error message for %q: %s", input, out)
			}
		})
	}
}

func TestDiff(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"basic days", []string{"diff", "march 1 2026", "march 11 2026"}, "10 days"},
		{"reversed dates", []string{"diff", "march 11 2026", "march 1 2026"}, "10 days"},
		{"same date", []string{"diff", "march 1 2026", "march 1 2026"}, "0 days"},
		{"weeks even", []string{"diff", "--weeks", "march 1 2026", "march 15 2026"}, "2 weeks"},
		{"weeks remainder", []string{"diff", "--weeks", "march 1 2026", "march 11 2026"}, "1 week, 3 days"},
		{"weeks reversed", []string{"diff", "--weeks", "march 11 2026", "march 1 2026"}, "1 week, 3 days"},
		{"workdays", []string{"diff", "--workdays", "march 2 2026", "march 6 2026"}, "4 workdays"},
		{"workdays reversed", []string{"diff", "--workdays", "march 6 2026", "march 2 2026"}, "4 workdays"},
		{"weeks trailing flag", []string{"diff", "march 1 2026", "march 15 2026", "--weeks"}, "2 weeks"},
		{"workdays trailing flag", []string{"diff", "march 2 2026", "march 6 2026", "--workdays"}, "4 workdays"},
		{"weeks between flags and dates", []string{"diff", "march 1 2026", "march 8 2026", "--weeks"}, "1 week"},
		{"workdays between flags and dates", []string{"diff", "march 2 2026", "--workdays", "march 4 2026"}, "2 workdays"},
		{"to separator", []string{"diff", "march 1 2026", "to", "march 11 2026"}, "10 days"},
		{"until separator", []string{"diff", "march 1 2026", "until", "march 11 2026"}, "10 days"},
		{"to with multiword exprs", []string{"diff", "march 1 2026", "to", "march 15 2026"}, "14 days"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(testBinary, tt.args...)
			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			got := strings.TrimSpace(string(out))
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDiffErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
	}{
		{"no args", []string{"diff"}},
		{"one arg", []string{"diff", "2026-03-01"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(testBinary, tt.args...)
			_, err := cmd.Output()
			if err == nil {
				t.Fatal("expected non-zero exit code")
			}
		})
	}
}

func TestFormatFlag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"custom format", []string{"--format", "01/02/2006", "march 25 2026"}, "03/25/2026"},
		{"long format", []string{"--format", "January 2, 2006", "march 25 2026"}, "March 25, 2026"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(testBinary, tt.args...)
			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			got := strings.TrimSpace(string(out))
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatNoArgs(t *testing.T) {
	t.Parallel()
	before := time.Now().Format("01/02/2006")
	cmd := exec.Command(testBinary, "--format", "01/02/2006")
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	after := time.Now().Format("01/02/2006")
	got := strings.TrimSpace(string(out))
	if got != before && got != after {
		t.Errorf("got %q, want %q or %q", got, before, after)
	}
}

func TestRelativeSubcommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"rel today", []string{"rel", "today"}, "today"},
		{"relative today", []string{"relative", "today"}, "today"},
		{"rel tomorrow", []string{"rel", "tomorrow"}, "tomorrow"},
		{"rel yesterday", []string{"rel", "yesterday"}, "yesterday"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(testBinary, tt.args...)
			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			got := strings.TrimSpace(string(out))
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRelativeNoArgs(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "rel")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	if got != "today" {
		t.Errorf("got %q, want %q", got, "today")
	}
}

func TestExpandMonthShorthand(t *testing.T) {
	tests := []struct {
		input []string
		want  []string
	}{
		{[]string{"-3"}, []string{"--months", "3"}},
		{[]string{"-12"}, []string{"--months", "12"}},
		{[]string{"-3", "march"}, []string{"--months", "3", "march"}},
		{[]string{"-h"}, []string{"-h"}}, // not a number, leave alone
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.input, " "), func(t *testing.T) {
			got := expandMonthShorthand(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFormatDoesNotEatSubcommand(t *testing.T) {
	t.Parallel()
	// `wen --format diff ...` should error, not silently consume "diff" as the format value.
	cmd := exec.Command(testBinary, "--format", "diff", "march 1 2026", "march 15 2026")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error when --format value is a subcommand name")
	}
	got := string(out)
	if !strings.Contains(got, "subcommand") {
		t.Errorf("expected error mentioning subcommand, got: %s", got)
	}
}

func TestSubcommandAliases(t *testing.T) {
	t.Parallel()
	// "calendar" should work the same as "cal" — we can't test the TUI,
	// but we can test that "relative" works the same as "rel".
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"relative alias", []string{"relative", "today"}, "today"},
		{"rel alias", []string{"rel", "today"}, "today"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(testBinary, tt.args...)
			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			got := strings.TrimSpace(string(out))
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseCalArgs(t *testing.T) {
	today := time.Date(2026, 3, 21, 12, 0, 0, 0, time.Local)
	tests := []struct {
		name      string
		args      []string
		wantMonth time.Month
		wantYear  int
	}{
		{"empty", nil, time.March, 2026},
		{"month only", []string{"march"}, time.March, 2026},
		{"month abbrev", []string{"dec"}, time.December, 2026},
		{"month and year", []string{"march", "2027"}, time.March, 2027},
		{"month mixed case", []string{"APRIL"}, time.April, 2026},
		{"december 2025", []string{"december", "2025"}, time.December, 2025},
		{"ignores small numbers as year", []string{"march", "32"}, time.March, 2026},
		{"ignores two-digit year", []string{"march", "99"}, time.March, 2026},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCalArgs(tt.args, today)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Month() != tt.wantMonth {
				t.Errorf("month: got %v, want %v", got.Month(), tt.wantMonth)
			}
			if got.Year() != tt.wantYear {
				t.Errorf("year: got %d, want %d", got.Year(), tt.wantYear)
			}
		})
	}
}

func TestRunWithWriter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"parse date", []string{"march 25 2026"}, "2026-03-25"},
		{"format flag", []string{"--format", "01/02/2006", "march 25 2026"}, "03/25/2026"},
		{"relative today", []string{"rel", "today"}, "today"},
		{"relative tomorrow", []string{"rel", "tomorrow"}, "tomorrow"},
		{"relative yesterday", []string{"rel", "yesterday"}, "yesterday"},
		{"relative future", []string{"rel", "march 30 2030"}, "in"},
		{"relative past", []string{"rel", "march 1 2020"}, "ago"},
		{"diff days", []string{"diff", "march 1 2026", "march 11 2026"}, "10 days"},
		{"diff weeks", []string{"diff", "--weeks", "march 1 2026", "march 15 2026"}, "2 weeks"},
		{"diff weeks remainder", []string{"diff", "--weeks", "march 1 2026", "march 10 2026"}, "1 week, 2 days"},
		{"diff workdays", []string{"diff", "--workdays", "march 2 2026", "march 6 2026"}, "4 workdays"},
		{"diff to separator", []string{"diff", "march 1 2026", "to", "march 11 2026"}, "10 days"},
		{"diff until separator", []string{"diff", "march 1 2026", "until", "march 11 2026"}, "10 days"},
		{"help flag", []string{"--help"}, "wen - a natural language date tool"},
		{"short help flag", []string{"-h"}, "wen - a natural language date tool"},
		{"version flag", []string{"--version"}, "wen dev"},
		{"short version flag", []string{"-v"}, "wen dev"},
		{"relative no args", []string{"rel"}, "today"},
		{"cal print", []string{"cal", "--print"}, "Su Mo Tu"},
		{"cal print named month", []string{"cal", "--print", "march"}, "March"},
		{"cal print month year", []string{"cal", "--print", "december", "2027"}, "December 2027"},
		{"cal print julian", []string{"cal", "--print", "--julian"}, "Sun Mon Tue"},
		{"cal print multi month", []string{"cal", "--print", "-3"}, "Su Mo Tu"},
		{"cal auto print", []string{"cal"}, "Su Mo Tu"},
		{"row print", []string{"row", "--print"}, "Mo Tu"},
		{"row auto print", []string{"row"}, "Mo Tu"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf strings.Builder
			err := run(&buf, tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := strings.TrimSpace(buf.String())
			if !strings.Contains(got, tt.want) {
				t.Errorf("got %q, want substring %q", got, tt.want)
			}
		})
	}
}

func TestCalHighlightWarningOnStderr(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "cal", "--print", "--highlight-file", "/nonexistent/wen-test-file.json")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	// stdout has the calendar output; we only care about stderr here.
	if err := cmd.Run(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := stderr.String()
	if !strings.Contains(got, "not found") {
		t.Errorf("expected stderr warning about missing highlight file, got %q", got)
	}
}

func TestRunErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"invalid date", []string{"pizza"}, "could not parse date"},
		{"format missing value", []string{"--format"}, "--format requires a value"},
		{"format eats subcommand", []string{"--format", "diff"}, "subcommand"},
		{"format eats rel", []string{"--format", "rel"}, "subcommand"},
		{"format eats cal", []string{"--format", "cal"}, "subcommand"},
		{"diff missing args", []string{"diff"}, "diff requires two"},
		{"diff one arg", []string{"diff", "today"}, "diff requires two"},
		{"diff unparseable", []string{"diff", "pizza", "cake"}, "could not parse date"},
		{"relative invalid", []string{"rel", "pizza"}, "could not parse date"},
		{"cal invalid month", []string{"cal", "--print", "pizza"}, "could not parse date"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf strings.Builder
			err := run(&buf, tt.args)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error %q should contain %q", err.Error(), tt.want)
			}
		})
	}
}

// TestDiffAcrossDST is a regression test for DST-related off-by-one errors.
// The bug: date diff used d2.Sub(d1).Hours()/24 with local times, which could
// produce 23 or 25 hours across a DST boundary, yielding wrong day counts.
// The fix normalizes to UTC before computing. These tests verify the fix holds.
func TestDiffAcrossDST(t *testing.T) {
	t.Parallel()

	// --- Part 1: test countWorkdays directly with UTC dates ---
	// March 8 2026 is the US spring-forward date. Using UTC ensures
	// the function is immune to local timezone DST transitions.
	t.Run("countWorkdays across DST boundary", func(t *testing.T) {
		t.Parallel()
		// March 1 (Sun) to March 15 (Sun) 2026 — crosses spring-forward on March 8.
		// Workdays: Mar 2-6 (5) + Mar 9-13 (5) = 10.
		start := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC)
		got := wen.CountWorkdays(start, end)
		if got != 10 {
			t.Errorf("CountWorkdays(March 1 -> March 15) = %d, want 10", got)
		}
	})

	t.Run("countWorkdays to DST day itself", func(t *testing.T) {
		t.Parallel()
		// March 1 (Sun) to March 8 (Sun) 2026 — the DST transition day.
		// Workdays: Mar 2-6 (Mon-Fri) = 5.
		start := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, time.March, 8, 0, 0, 0, 0, time.UTC)
		got := wen.CountWorkdays(start, end)
		if got != 5 {
			t.Errorf("CountWorkdays(March 1 -> March 8) = %d, want 5", got)
		}
	})

	// --- Part 2: test through the binary with absolute dates crossing DST ---
	t.Run("binary diff days across DST", func(t *testing.T) {
		t.Parallel()
		// March 1 to March 15 = 14 calendar days (not 13).
		cmd := exec.Command(testBinary, "diff", "march 1 2026", "march 15 2026")
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		got := strings.TrimSpace(string(out))
		if got != "14 days" {
			t.Errorf("got %q, want %q", got, "14 days")
		}
	})

	t.Run("binary diff weeks across DST", func(t *testing.T) {
		t.Parallel()
		// March 1 to March 15 = exactly 2 weeks.
		cmd := exec.Command(testBinary, "diff", "march 1 2026", "march 15 2026", "--weeks")
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		got := strings.TrimSpace(string(out))
		if got != "2 weeks" {
			t.Errorf("got %q, want %q", got, "2 weeks")
		}
	})
}

func TestParseCalArgsErrors(t *testing.T) {
	today := time.Date(2026, 3, 21, 12, 0, 0, 0, time.Local)

	t.Run("completely invalid input", func(t *testing.T) {
		_, err := parseCalArgs([]string{"xyzzy"}, today)
		if err == nil {
			t.Error("expected error for invalid input \"xyzzy\", got nil")
		}
	})

	t.Run("empty slice returns today", func(t *testing.T) {
		got, err := parseCalArgs([]string{}, today)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Year() != today.Year() || got.Month() != today.Month() || got.Day() != today.Day() {
			t.Errorf("got %v, want %v", got, today)
		}
	})
}

func TestCountWorkdays(t *testing.T) {
	tests := []struct {
		name  string
		start time.Time
		end   time.Time
		want  int
	}{
		{
			name:  "same day yields 0 workdays",
			start: time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC), // Wednesday
			end:   time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC),
			want:  0,
		},
		{
			name:  "Saturday to Monday yields 0 workdays (weekend only)",
			start: time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC), // Saturday
			end:   time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC), // Monday
			want:  0,
		},
		{
			name:  "Monday to Saturday yields 5 workdays",
			start: time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC), // Monday
			end:   time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC), // Saturday
			want:  5,
		},
		{
			name:  "reversed args still works (swaps internally)",
			start: time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC), // Saturday
			end:   time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC), // Monday
			want:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wen.CountWorkdays(tt.start, tt.end)
			if got != tt.want {
				t.Errorf("CountWorkdays(%s, %s) = %d, want %d",
					tt.start.Format("2006-01-02"), tt.end.Format("2006-01-02"),
					got, tt.want)
			}
		})
	}
}

func TestDiffSameDateFlags(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"same date weeks", []string{"diff", "--weeks", "march 1 2026", "march 1 2026"}, "0 weeks"},
		{"same date workdays", []string{"diff", "--workdays", "march 1 2026", "march 1 2026"}, "0 workdays"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf strings.Builder
			err := run(&buf, tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := strings.TrimSpace(buf.String())
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatFlagGuardsAllSubcommands(t *testing.T) {
	t.Parallel()
	// Verify all subcommand names are rejected as --format values.
	subcmds := []string{"cal", "calendar", "diff", "rel", "relative", "row", "--help", "-h", "--version", "-v"}
	for _, sub := range subcmds {
		t.Run(sub, func(t *testing.T) {
			t.Parallel()
			var buf strings.Builder
			err := run(&buf, []string{"--format", sub})
			if err == nil {
				t.Errorf("expected error for --format %s", sub)
			}
		})
	}
}

func TestCalMonthsFlagShorthand(t *testing.T) {
	t.Parallel()
	// -m N should work the same as --months N and -N.
	// Verify -m is accepted as a flag by the cal subcommand.
	cmd := exec.Command(testBinary, "cal", "-m", "3", "--help")
	out, _ := cmd.CombinedOutput()
	if strings.Contains(string(out), "flag provided but not defined") {
		t.Errorf("-m flag not recognized: %s", out)
	}
}

func TestPlural(t *testing.T) {
	tests := []struct {
		n    int
		word string
		want string
	}{
		{0, "day", "days"},
		{1, "day", "day"},
		{2, "day", "days"},
		{1, "week", "week"},
		{5, "week", "weeks"},
	}
	for _, tt := range tests {
		got := plural(tt.n, tt.word)
		if got != tt.want {
			t.Errorf("plural(%d, %q) = %q, want %q", tt.n, tt.word, got, tt.want)
		}
	}
}

func TestRowSubcommandHelp(t *testing.T) {
	t.Parallel()
	var buf strings.Builder
	err := run(&buf, []string{"--help"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "wen row") {
		t.Error("help should mention 'wen row'")
	}
}

func TestCalPrint(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"basic month", []string{"cal", "--print", "march", "2027"}, "March 2027"},
		{"contains day headers", []string{"cal", "--print", "march", "2026"}, "Su Mo Tu We Th Fr Sa"},
		{"contains days", []string{"cal", "--print", "march", "2026"}, "31"},
		{"multi month", []string{"cal", "--print", "-3", "march", "2027"}, "February 2027"},
		{"multi month has april", []string{"cal", "--print", "-3", "march", "2027"}, "April 2027"},
		{"julian mode", []string{"cal", "--print", "--julian", "march", "2026"}, " 60"},
		{"julian 3-char headers", []string{"cal", "--print", "--julian", "march", "2026"}, "Sun Mon Tue"},
		{"short flags", []string{"cal", "-p", "-j", "march", "2026"}, " 60"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf strings.Builder
			err := run(&buf, tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := buf.String()
			if !strings.Contains(got, tt.want) {
				t.Errorf("got:\n%s\nwant substring %q", got, tt.want)
			}
		})
	}
}

func TestRowPrint(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"basic strip", []string{"row", "--print", "march", "2026"}, "Mr"},
		{"contains day headers", []string{"row", "--print", "march", "2026"}, "Su"},
		{"julian mode", []string{"row", "--print", "--julian", "march", "2026"}, "Sun"},
		{"julian yearday", []string{"row", "--print", "--julian", "march", "2026"}, " 60"},
		{"short flags", []string{"row", "-p", "-j", "march", "2026"}, "Sun"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf strings.Builder
			err := run(&buf, tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := buf.String()
			if !strings.Contains(got, tt.want) {
				t.Errorf("got:\n%s\nwant substring %q", got, tt.want)
			}
		})
	}
}

func TestFormatGuardsRow(t *testing.T) {
	t.Parallel()
	var buf strings.Builder
	err := run(&buf, []string{"--format", "row"})
	if err == nil {
		t.Error("expected error when --format value is 'row'")
	}
}

func TestHelpContainsNewFlags(t *testing.T) {
	t.Parallel()
	var buf strings.Builder
	err := run(&buf, []string{"--help"})
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{"--print", "--julian", "J                Toggle Julian", "N/P"} {
		if !strings.Contains(got, want) {
			t.Errorf("help should contain %q", want)
		}
	}
}

func TestCalPrintBinary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"print flag", []string{"cal", "--print", "march", "2027"}, "March 2027"},
		{"piped stdout", []string{"cal", "march", "2027"}, "March 2027"},
		{"julian flag", []string{"cal", "--print", "--julian", "march", "2027"}, " 60"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(testBinary, tt.args...)
			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("unexpected error: %s\n%s", err, out)
			}
			got := string(out)
			if !strings.Contains(got, tt.want) {
				t.Errorf("got:\n%s\nwant substring %q", got, tt.want)
			}
		})
	}
}

func TestRowPrintBinary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"print flag", []string{"row", "--print", "march", "2026"}, "Mr"},
		{"piped stdout", []string{"row", "march", "2026"}, "Mr"},
		{"julian flag", []string{"row", "--print", "--julian", "march", "2026"}, "Sun"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(testBinary, tt.args...)
			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("unexpected error: %s\n%s", err, out)
			}
			got := string(out)
			if !strings.Contains(got, tt.want) {
				t.Errorf("got:\n%s\nwant substring %q", got, tt.want)
			}
		})
	}
}

func TestFormatFlagMissingValue(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "--format")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit code")
	}
	if !strings.Contains(string(out), "--format requires a value") {
		t.Errorf("unexpected error: %s", out)
	}
}

func TestDiffMissingArgs(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "diff")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit code for diff with no args")
	}
	got := string(out)
	if !strings.Contains(got, "usage") && !strings.Contains(got, "requires") {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestDiffSingleArg(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "diff", "tomorrow")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit code for diff with single arg")
	}
	got := string(out)
	if !strings.Contains(got, "two dates") && !strings.Contains(got, "usage") && !strings.Contains(got, "could not") {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestHighlightFileFlag(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	hlFile := filepath.Join(dir, "dates.json")
	if err := os.WriteFile(hlFile, []byte(`["2026-04-01","2026-04-15"]`), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(testBinary, "cal", "--print", "--highlight-file", hlFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error: %s\n%s", err, out)
	}
	if len(out) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestHelpOutputContent(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "--help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := string(out)
	if !strings.Contains(got, "Usage:") {
		t.Errorf("help output missing 'Usage:': %s", got)
	}
	if !strings.Contains(got, "Subcommands:") {
		t.Errorf("help output missing 'Subcommands:': %s", got)
	}
}

func TestRelativeSubcommandAlias(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "rel", "tomorrow")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	if got == "" {
		t.Error("expected non-empty output from rel alias")
	}
}

func TestCalPrintCurrentMonth(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "cal", "--print")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	if got == "" {
		t.Errorf("expected non-empty output")
	}
	if !strings.Contains(got, "Su Mo Tu") {
		t.Errorf("expected day headers in output, got:\n%s", got)
	}
}

func TestCalPrintNamedMonth(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "cal", "--print", "march")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	if !strings.Contains(got, "March") {
		t.Errorf("expected output to contain 'March', got:\n%s", got)
	}
}

func TestCalPrintMultiMonth(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "cal", "--print", "-3")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	// With -3 centered on the current month, we expect at least two month names.
	now := time.Now()
	prevMonth := now.AddDate(0, -1, 0).Month().String()
	nextMonth := now.AddDate(0, 1, 0).Month().String()
	if !strings.Contains(got, prevMonth) {
		t.Errorf("expected output to contain %q, got:\n%s", prevMonth, got)
	}
	if !strings.Contains(got, nextMonth) {
		t.Errorf("expected output to contain %q, got:\n%s", nextMonth, got)
	}
}

func TestCalPrintJulian(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "cal", "--print", "--julian")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	if !strings.Contains(got, "Sun Mon Tue") {
		t.Errorf("expected 3-char day headers in julian mode, got:\n%s", got)
	}
}

func TestRowPrintCurrentMonth(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "row", "--print")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	if got == "" {
		t.Errorf("expected non-empty output")
	}
}

func TestCalPrintMonthYear(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "cal", "--print", "december", "2027")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	if !strings.Contains(got, "December 2027") {
		t.Errorf("expected output to contain 'December 2027', got:\n%s", got)
	}
}

func TestCalPrintInvalidMonth(t *testing.T) {
	t.Parallel()
	cmd := exec.Command(testBinary, "cal", "--print", "pizza")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit code for invalid month")
	}
	got := string(out)
	if !strings.Contains(got, "could not parse date") {
		t.Errorf("expected parse error message, got: %s", got)
	}
}
