package calendar

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadHighlightedDates(t *testing.T) {
	dir := t.TempDir()

	t.Run("valid file", func(t *testing.T) {
		path := filepath.Join(dir, "dates.json")
		if err := os.WriteFile(path, []byte(`["2026-03-25", "2026-04-01"]`), 0644); err != nil {
			t.Fatal(err)
		}

		dates := LoadHighlightedDates(path)
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
		dates := LoadHighlightedDates(filepath.Join(dir, "nonexistent.json"))
		if dates != nil {
			t.Error("expected nil for missing file")
		}
	})

	t.Run("malformed JSON", func(t *testing.T) {
		path := filepath.Join(dir, "bad.json")
		if err := os.WriteFile(path, []byte(`not json`), 0644); err != nil {
			t.Fatal(err)
		}
		dates := LoadHighlightedDates(path)
		if dates != nil {
			t.Error("expected nil for malformed JSON")
		}
	})

	t.Run("empty path", func(t *testing.T) {
		dates := LoadHighlightedDates("")
		if dates != nil {
			t.Error("expected nil for empty path")
		}
	})

	t.Run("invalid dates skipped", func(t *testing.T) {
		path := filepath.Join(dir, "mixed.json")
		if err := os.WriteFile(path, []byte(`["2026-03-25", "not-a-date", "2026-04-01"]`), 0644); err != nil {
			t.Fatal(err)
		}
		dates := LoadHighlightedDates(path)
		if len(dates) != 2 {
			t.Fatalf("expected 2 valid dates, got %d", len(dates))
		}
	})
}

func TestExpandTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	t.Run("expands tilde prefix", func(t *testing.T) {
		got := expandTilde("~/foo/bar")
		want := filepath.Join(home, "foo", "bar")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("leaves absolute path unchanged", func(t *testing.T) {
		got := expandTilde("/absolute/path")
		if got != "/absolute/path" {
			t.Errorf("got %q, want %q", got, "/absolute/path")
		}
	})

	t.Run("leaves empty string unchanged", func(t *testing.T) {
		got := expandTilde("")
		if got != "" {
			t.Errorf("got %q, want %q", got, "")
		}
	})
}

func TestResolveHighlightSource(t *testing.T) {
	t.Run("flag takes priority", func(t *testing.T) {
		got := ResolveHighlightSource("/flag/path", "/config/path")
		if got != "/flag/path" {
			t.Errorf("expected flag path, got %q", got)
		}
	})

	t.Run("config fallback", func(t *testing.T) {
		got := ResolveHighlightSource("", "/config/path")
		if got != "/config/path" {
			t.Errorf("expected config path, got %q", got)
		}
	})
}

func TestWithHighlightSource(t *testing.T) {
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
}

func TestWithHighlightSourceMissing(t *testing.T) {
	today := time.Date(2026, time.March, 17, 0, 0, 0, 0, time.Local)
	m := New(today, today, DefaultConfig(), WithHighlightSource("/nonexistent/file.json"))

	if m.highlightedDates != nil {
		t.Error("expected nil highlightedDates for missing file")
	}
}

func TestWithHighlightedDatesClearsHighlightPath(t *testing.T) {
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
