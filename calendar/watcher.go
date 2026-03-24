package calendar

import (
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

const debounceDuration = 150 * time.Millisecond

// highlightChangedMsg is sent when the highlight source file changes.
// It carries the reloaded dates, the watcher handle (for reuse), and the path.
type highlightChangedMsg struct {
	dates   map[time.Time]bool
	watcher *fsnotify.Watcher
	path    string
}

// watcherErrMsg is sent when the file watcher encounters an error.
type watcherErrMsg struct{ err error }

// startFileWatcher returns a tea.Cmd that creates an fsnotify watcher on the
// parent directory of path, waits for a change to the target file, and returns
// a highlightChangedMsg with the reloaded dates.
func startFileWatcher(path string) tea.Cmd {
	return func() tea.Msg {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return watcherErrMsg{err: err}
		}
		dir := filepath.Dir(path)
		if err := watcher.Add(dir); err != nil {
			_ = watcher.Close()
			return watcherErrMsg{err: err}
		}
		return watchLoop(watcher, path)
	}
}

// waitForNextChange returns a tea.Cmd that reuses an existing watcher to wait
// for the next file change. Uses the same debounced watchLoop.
func waitForNextChange(watcher *fsnotify.Watcher, path string) tea.Cmd {
	return func() tea.Msg {
		return watchLoop(watcher, path)
	}
}

// watchLoop blocks on the watcher's event channel, debounces rapid events,
// and returns a highlightChangedMsg when the target file changes.
func watchLoop(watcher *fsnotify.Watcher, path string) tea.Msg {
	target := filepath.Base(path)
	debounce := time.NewTimer(time.Hour)
	debounce.Stop()
	triggered := false

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if filepath.Base(event.Name) != target {
				continue
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}
			// Reset debounce timer
			if !debounce.Stop() && triggered {
				select {
				case <-debounce.C:
				default:
				}
			}
			debounce.Reset(debounceDuration)
			triggered = true

		case <-debounce.C:
			dates := LoadHighlightedDates(path)
			return highlightChangedMsg{
				dates:   dates,
				watcher: watcher,
				path:    path,
			}

		case _, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			// Continue listening on errors
		}
	}
}
