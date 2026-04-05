package calendar

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadHighlightedDates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	t.Run("valid file", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(dir, "dates.json")
		if err := os.WriteFile(path, []byte(`["2026-03-25", "2026-04-01"]`), 0644); err != nil {
			t.Fatal(err)
		}

		dates, warnings := LoadHighlightedDates(path)
		if len(warnings) != 0 {
			t.Errorf("expected no warnings, got %v", warnings)
		}
		if len(dates) != 2 {
			t.Fatalf("expected 2 dates, got %d", len(dates))
		}
		if !dates[time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)] {
			t.Error("expected 2026-03-25 to be highlighted")
		}
		if !dates[time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)] {
			t.Error("expected 2026-04-01 to be highlighted")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		missingPath := filepath.Join(dir, "nonexistent.json")
		dates, warnings := LoadHighlightedDates(missingPath)
		if dates != nil {
			t.Error("expected nil for missing file")
		}
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
		}
		if !strings.Contains(warnings[0].Message, "not found") {
			t.Errorf("expected 'not found' warning, got %q", warnings[0].Message)
		}
	})

	t.Run("malformed JSON", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(dir, "bad.json")
		if err := os.WriteFile(path, []byte(`not json`), 0644); err != nil {
			t.Fatal(err)
		}
		dates, warnings := LoadHighlightedDates(path)
		if dates != nil {
			t.Error("expected nil for malformed JSON")
		}
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
		}
		if !strings.Contains(warnings[0].Message, "not valid JSON") {
			t.Errorf("expected 'not valid JSON' warning, got %q", warnings[0].Message)
		}
	})

	t.Run("empty path", func(t *testing.T) {
		t.Parallel()
		dates, warnings := LoadHighlightedDates("")
		if dates != nil {
			t.Error("expected nil for empty path")
		}
		if len(warnings) != 0 {
			t.Errorf("expected no warnings for empty path, got %v", warnings)
		}
	})

	t.Run("invalid dates produce per-entry warnings", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(dir, "mixed.json")
		if err := os.WriteFile(path, []byte(`["2026-03-25", "not-a-date", "2026-04-01"]`), 0644); err != nil {
			t.Fatal(err)
		}
		dates, warnings := LoadHighlightedDates(path)
		if len(dates) != 2 {
			t.Fatalf("expected 2 valid dates, got %d", len(dates))
		}
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning for invalid date, got %d: %v", len(warnings), warnings)
		}
		if !strings.Contains(warnings[0].Message, "not-a-date") {
			t.Errorf("expected warning to mention 'not-a-date', got %q", warnings[0].Message)
		}
	})
}

func TestExpandTilde(t *testing.T) {
	t.Parallel()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	t.Run("expands tilde prefix", func(t *testing.T) {
		t.Parallel()
		got := expandTilde("~/foo/bar")
		want := filepath.Join(home, "foo", "bar")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("leaves absolute path unchanged", func(t *testing.T) {
		t.Parallel()
		got := expandTilde("/absolute/path")
		if got != "/absolute/path" {
			t.Errorf("got %q, want %q", got, "/absolute/path")
		}
	})

	t.Run("leaves empty string unchanged", func(t *testing.T) {
		t.Parallel()
		got := expandTilde("")
		if got != "" {
			t.Errorf("got %q, want %q", got, "")
		}
	})
}

func TestResolveHighlightSource(t *testing.T) {
	t.Parallel()
	t.Run("flag takes priority", func(t *testing.T) {
		t.Parallel()
		got := ResolveHighlightSource("/flag/path", "/config/path")
		if got != "/flag/path" {
			t.Errorf("expected flag path, got %q", got)
		}
	})

	t.Run("config fallback", func(t *testing.T) {
		t.Parallel()
		got := ResolveHighlightSource("", "/config/path")
		if got != "/config/path" {
			t.Errorf("expected config path, got %q", got)
		}
	})
}

func TestWithHighlightSource(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "dates.json")
	if err := os.WriteFile(path, []byte(`["2026-03-25"]`), 0644); err != nil {
		t.Fatal(err)
	}

	today := time.Date(2026, time.March, 17, 0, 0, 0, 0, time.Local)
	m := New(today, today, DefaultConfig(), WithHighlightSource(path))

	if m.highlightPath != path {
		t.Errorf("highlightPath = %q, want %q", m.highlightPath, path)
	}
	key := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	if !m.highlightedDates[key] {
		t.Error("expected 2026-03-25 to be highlighted")
	}
	if len(m.Warnings()) != 0 {
		t.Errorf("expected no warnings, got %v", m.Warnings())
	}
}

func TestWithHighlightSourceMissing(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, time.March, 17, 0, 0, 0, 0, time.Local)
	m := New(today, today, DefaultConfig(), WithHighlightSource("/nonexistent/file.json"))

	if m.highlightedDates != nil {
		t.Error("expected nil highlightedDates for missing file")
	}
	if len(m.Warnings()) == 0 {
		t.Error("expected warning for missing highlight file")
	}
}

func TestWithHighlightedDatesClearsHighlightPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "dates.json")
	if err := os.WriteFile(path, []byte(`["2026-03-25"]`), 0644); err != nil {
		t.Fatal(err)
	}

	today := time.Date(2026, time.March, 17, 0, 0, 0, 0, time.Local)
	manualDates := map[time.Time]bool{
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC): true,
	}
	// WithHighlightSource first, then WithHighlightedDates — last option wins.
	m := New(today, today, DefaultConfig(),
		WithHighlightSource(path),
		WithHighlightedDates(manualDates),
	)

	if m.highlightPath != "" {
		t.Errorf("expected highlightPath cleared, got %q", m.highlightPath)
	}
	if !m.highlightedDates[time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)] {
		t.Error("expected manual date 2026-04-01 to be highlighted")
	}
	if m.highlightedDates[time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)] {
		t.Error("expected file date 2026-03-25 to NOT be highlighted")
	}
}
