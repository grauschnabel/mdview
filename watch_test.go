package main

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// startWatch watches file and returns a counter of onChange calls.
func startWatch(t *testing.T, file string) *atomic.Int32 {
	t.Helper()
	var calls atomic.Int32
	stop, err := watchFile(file, func() { calls.Add(1) })
	if err != nil {
		t.Fatalf("watchFile: %v", err)
	}
	t.Cleanup(stop)
	return &calls
}

// waitFor polls until the counter reaches want or the timeout expires.
func waitFor(calls *atomic.Int32, want int32) bool {
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if calls.Load() >= want {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func TestWatchWrite(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.md")
	os.WriteFile(file, []byte("one"), 0o644)
	calls := startWatch(t, file)

	os.WriteFile(file, []byte("two"), 0o644)
	if !waitFor(calls, 1) {
		t.Fatal("no callback after write")
	}
}

func TestWatchDebounce(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.md")
	os.WriteFile(file, []byte("one"), 0o644)
	calls := startWatch(t, file)

	for i := 0; i < 5; i++ {
		os.WriteFile(file, []byte("x"), 0o644)
		time.Sleep(10 * time.Millisecond)
	}
	if !waitFor(calls, 1) {
		t.Fatal("no callback")
	}
	time.Sleep(3 * debounce)
	if n := calls.Load(); n != 1 {
		t.Errorf("expected 1 debounced callback, got %d", n)
	}
}

func TestWatchAtomicSave(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.md")
	os.WriteFile(file, []byte("one"), 0o644)
	calls := startWatch(t, file)

	tmp := filepath.Join(dir, ".a.md.tmp")
	os.WriteFile(tmp, []byte("two"), 0o644)
	if err := os.Rename(tmp, file); err != nil {
		t.Fatal(err)
	}
	if !waitFor(calls, 1) {
		t.Fatal("no callback after atomic save")
	}
	// A second atomic save must still be seen (the watch must not be detached).
	calls.Store(0)
	os.WriteFile(tmp, []byte("three"), 0o644)
	os.Rename(tmp, file)
	if !waitFor(calls, 1) {
		t.Fatal("no callback after second atomic save")
	}
}

func TestWatchIgnoresSiblings(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.md")
	os.WriteFile(file, []byte("one"), 0o644)
	calls := startWatch(t, file)

	os.WriteFile(filepath.Join(dir, "other.md"), []byte("x"), 0o644)
	time.Sleep(3 * debounce)
	if n := calls.Load(); n != 0 {
		t.Errorf("sibling file triggered %d callbacks", n)
	}
}
