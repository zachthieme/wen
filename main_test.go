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
	cmd := exec.Command(bin)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got := strings.TrimSpace(string(out))
	want := time.Now().Format("2006-01-02")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
