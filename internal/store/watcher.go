package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors the data directory for file changes and updates
// the SQLite index accordingly. It's used while the API server is running
// to keep the index in sync with external edits (ADR-002 §6).
//
// Usage:
//
//	w, err := store.NewWatcher(dataDir, index)
//	defer w.Close()
//	// Watcher runs in background goroutine
type Watcher struct {
	watcher *fsnotify.Watcher
	dataDir string
	index   *Index
	done    chan struct{}
	wg      sync.WaitGroup

	// OnError is called when the watcher encounters an error.
	// If nil, errors are printed to stderr.
	OnError func(error)

	// OnIndex is called after a file is successfully indexed.
	// Useful for logging or testing. May be nil.
	OnIndex func(path string)
}

// NewWatcher creates a file watcher that monitors dataDir for .md file
// changes and updates the index. It starts watching immediately in a
// background goroutine.
func NewWatcher(dataDir string, index *Index) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create watcher: %w", err)
	}

	w := &Watcher{
		watcher: fsw,
		dataDir: dataDir,
		index:   index,
		done:    make(chan struct{}),
	}

	// Add the data directory and all subdirectories
	if err := w.addRecursive(dataDir); err != nil {
		_ = fsw.Close()
		return nil, fmt.Errorf("watch %s: %w", dataDir, err)
	}

	// Start the event loop
	w.wg.Add(1)
	go w.loop()

	return w, nil
}

// Close stops the watcher and waits for the event loop to finish.
func (w *Watcher) Close() error {
	close(w.done)
	err := w.watcher.Close()
	w.wg.Wait()
	return err
}

// loop is the main event loop that processes filesystem events.
func (w *Watcher) loop() {
	defer w.wg.Done()

	for {
		select {
		case <-w.done:
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.reportError(fmt.Errorf("watcher: %w", err))
		}
	}
}

// handleEvent processes a single filesystem event.
func (w *Watcher) handleEvent(event fsnotify.Event) {
	path := event.Name

	// If a new directory is created, start watching it too
	if event.Has(fsnotify.Create) {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			if !strings.HasPrefix(filepath.Base(path), ".") {
				_ = w.watcher.Add(path)
			}
			return
		}
	}

	// Only process .md files, ignore everything else
	// Skip temp files from atomic writes, _box.md metadata, and non-markdown
	base := filepath.Base(path)
	if !strings.HasSuffix(path, ".md") || base == "_box.md" || strings.HasPrefix(base, ".tmp-") {
		return
	}

	// Ignore hidden directories
	relPath, err := filepath.Rel(w.dataDir, path)
	if err != nil {
		return
	}
	for _, part := range strings.Split(filepath.Dir(relPath), string(filepath.Separator)) {
		if strings.HasPrefix(part, ".") {
			return
		}
	}

	switch {
	case event.Has(fsnotify.Create) || event.Has(fsnotify.Write):
		w.indexFile(path, relPath)

	case event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename):
		w.removeFile(path, relPath)
	}
}

// indexFile parses a .md file and upserts it into the index.
func (w *Watcher) indexFile(absPath, relPath string) {
	task, err := ParseFile(absPath)
	if err != nil {
		w.reportError(fmt.Errorf("parse %s: %w", relPath, err))
		return
	}

	if err := w.index.UpsertTask(task, relPath); err != nil {
		w.reportError(fmt.Errorf("index %s: %w", relPath, err))
		return
	}

	if w.OnIndex != nil {
		w.OnIndex(relPath)
	}
}

// removeFile deletes a task from the index when its file is removed.
// We need to find the task ID from the filename since we can't parse a deleted file.
func (w *Watcher) removeFile(_, relPath string) {
	// The filename is the ULID (e.g., "01JQX00010.md")
	base := filepath.Base(relPath)
	id := strings.TrimSuffix(base, ".md")

	if err := w.index.DeleteTask(id); err != nil {
		w.reportError(fmt.Errorf("remove %s from index: %w", relPath, err))
	}
}

// addRecursive adds a directory and all non-hidden subdirectories to the watcher.
func (w *Watcher) addRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		// Skip hidden directories
		if path != root && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		return w.watcher.Add(path)
	})
}

func (w *Watcher) reportError(err error) {
	if w.OnError != nil {
		w.OnError(err)
	} else {
		fmt.Fprintf(os.Stderr, "watcher error: %v\n", err)
	}
}
