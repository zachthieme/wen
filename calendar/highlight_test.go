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
