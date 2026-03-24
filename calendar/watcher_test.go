package calendar

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// watchTimeout is a generous deadline for filesystem-event tests so they
// don't flake on loaded CI runners.
const watchTimeout = 30 * time.Second

func TestWatchLoopDetectsFileChange(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem watcher test in short mode")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "dates.json")

	// Write initial file
	initial := []string{"2026-03-25"}
	data, _ := json.Marshal(initial)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Start the watcher cmd
	cmd := startFileWatcher(path)
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}

	// Modify the file in a goroutine after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		updated := []string{"2026-03-25", "2026-04-01"}
		data, _ := json.Marshal(updated)
		_ = os.WriteFile(path, data, 0644)
	}()

	// The cmd blocks until a change is detected — run it with a timeout
	done := make(chan highlightChangedMsg, 1)
	go func() {
		msg := cmd()
		if msg != nil {
			done <- msg.(highlightChangedMsg)
		}
	}()

	select {
	case msg := <-done:
		if len(msg.dates) != 2 {
			t.Errorf("expected 2 dates, got %d", len(msg.dates))
		}
		if msg.watcher == nil {
			t.Error("expected non-nil watcher in message")
		}
		if msg.path != path {
			t.Errorf("path = %q, want %q", msg.path, path)
		}
		// Clean up watcher
		_ = msg.watcher.Close()
	case <-time.After(watchTimeout):
		t.Fatal("timed out waiting for file change detection")
	}
}

func TestWatchLoopFileDeleted(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem watcher test in short mode")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "dates.json")

	// Write initial file
	data, _ := json.Marshal([]string{"2026-03-25"})
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	cmd := startFileWatcher(path)
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}

	// Delete the file after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = os.Remove(path)
	}()

	done := make(chan highlightChangedMsg, 1)
	go func() {
		msg := cmd()
		if msg != nil {
			done <- msg.(highlightChangedMsg)
		}
	}()

	select {
	case msg := <-done:
		if msg.dates != nil {
			t.Errorf("expected nil dates after deletion, got %d dates", len(msg.dates))
		}
		_ = msg.watcher.Close()
	case <-time.After(watchTimeout):
		t.Fatal("timed out waiting for delete detection")
	}
}

func TestStartFileWatcherMissingParentDir(t *testing.T) {
	cmd := startFileWatcher("/nonexistent/parent/dir/dates.json")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	// The cmd should return a watcherErrMsg when it can't watch
	msg := cmd()
	if _, ok := msg.(watcherErrMsg); !ok {
		t.Errorf("expected watcherErrMsg for missing parent dir, got %T", msg)
	}
}

func TestWaitForNextChange(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem watcher test in short mode")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "dates.json")

	data, _ := json.Marshal([]string{"2026-03-25"})
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Create a watcher manually to test waitForNextChange
	cmd := startFileWatcher(path)

	// Trigger first change
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = os.WriteFile(path, []byte(`["2026-04-01"]`), 0644)
	}()

	done := make(chan highlightChangedMsg, 1)
	go func() {
		msg := cmd()
		if msg != nil {
			done <- msg.(highlightChangedMsg)
		}
	}()

	select {
	case msg := <-done:
		// Now test waitForNextChange with the returned watcher
		cmd2 := waitForNextChange(msg.watcher, msg.path)
		if cmd2 == nil {
			t.Fatal("expected non-nil cmd from waitForNextChange")
		}

		// Trigger second change
		go func() {
			time.Sleep(50 * time.Millisecond)
			_ = os.WriteFile(path, []byte(`["2026-05-01", "2026-06-01"]`), 0644)
		}()

		done2 := make(chan highlightChangedMsg, 1)
		go func() {
			msg2 := cmd2()
			if msg2 != nil {
				done2 <- msg2.(highlightChangedMsg)
			}
		}()

		select {
		case msg2 := <-done2:
			if len(msg2.dates) != 2 {
				t.Errorf("expected 2 dates on second change, got %d", len(msg2.dates))
			}
			_ = msg2.watcher.Close()
		case <-time.After(watchTimeout):
			t.Fatal("timed out waiting for second file change")
		}
	case <-time.After(watchTimeout):
		t.Fatal("timed out waiting for first file change")
	}
}
