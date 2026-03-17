package main

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	binary := t.TempDir() + "/zdate"
	cmd := exec.Command("go", "build", "-o", binary, ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %s\n%s", err, out)
	}
	return binary
}

func TestNoArgs_PrintsToday(t *testing.T) {
	bin := buildBinary(t)
	today := time.Now().Format("2006-01-02")
	cmd := exec.Command(bin, "--now", today)
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
	bin := buildBinary(t)
	cmd := exec.Command(bin, "--now", "2026-03-17", "next friday")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	want := "2026-03-20"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNowFlagNoArgs(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "--now", "2026-06-15")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	want := "2026-06-15"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStdinMode(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "--now", "2026-03-17")
	cmd.Stdin = strings.NewReader("next friday\n")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	want := "2026-03-20"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFailure(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "pizza")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit code")
	}
	got := string(out)
	if !strings.Contains(got, `error: could not parse date "pizza"`) {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestInvalidNowFlag(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "--now", "not-a-date")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit code")
	}
	got := string(out)
	if !strings.Contains(got, `error: invalid --now date "not-a-date", expected yyyy-mm-dd`) {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestPositionalArgTakesPrecedenceOverStdin(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "--now", "2026-03-17", "next friday")
	cmd.Stdin = strings.NewReader("yesterday\n")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	want := "2026-03-20" // next friday, not yesterday
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
