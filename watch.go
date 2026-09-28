// This file implements file watching with debounced change notifications.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// debounce merges the several events one save often produces into one reload.
const debounce = 100 * time.Millisecond

// watchFile calls onChange whenever absPath changes, always from the same
// background goroutine, so calls never overlap. absPath must not be a symlink.
// It watches the parent directory because many editors save atomically via
// rename, which would detach a watch placed on the file itself.
// The returned function stops watching.
func watchFile(absPath string, onChange func()) (stop func(), err error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := w.Add(filepath.Dir(absPath)); err != nil {
		w.Close()
		return nil, err
	}

	go func() {
		var fire <-chan time.Time
		for {
			select {
			case ev, ok := <-w.Events:
				if !ok {
					return
				}
				if filepath.Clean(ev.Name) != absPath || !ev.Has(fsnotify.Write|fsnotify.Create|fsnotify.Rename) {
					continue
				}
				fire = time.After(debounce)
			case <-fire:
				fire = nil
				onChange()
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				fmt.Fprintf(os.Stderr, "mdview: watch error: %v\n", err)
			}
		}
	}()

	return func() { w.Close() }, nil
}
