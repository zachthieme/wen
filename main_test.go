package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var testBinary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "wen-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %s\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	testBinary = filepath.Join(dir, "wen")
	cmd := exec.Command("go", "build", "-o", testBinary, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %s\n%s", err, out)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestNoArgs_PrintsToday(t *testing.T) {
	today := time.Now().Format("2006-01-02")
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
	if _, err := time.Parse("2006-01-02", got); err != nil {
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
	if _, err := time.Parse("2006-01-02", got); err != nil {
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
	if !strings.Contains(got, `error: could not parse date "pizza"`) {
		t.Errorf("unexpected error message: %s", got)
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
