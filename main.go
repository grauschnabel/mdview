package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: mdview <file.md>")
		os.Exit(2)
	}

	filePath := os.Args[1]
	mdBytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mdview: %v\n", err)
		os.Exit(1)
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mdview: %v\n", err)
		os.Exit(1)
	}
	baseDir := filepath.Dir(absPath)

	htmlContent, err := RenderMarkdown(mdBytes, baseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mdview: %v\n", err)
		os.Exit(1)
	}

	baseURI, err := BaseURI(baseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mdview: %v\n", err)
		os.Exit(1)
	}

	gtk.Init(nil)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mdview: %v\n", err)
		os.Exit(1)
	}

	title := filepath.Base(absPath) + " — mdview"
	win.SetTitle(title)
	win.SetDefaultSize(900, 700)
	win.SetPosition(gtk.WIN_POS_CENTER)

	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	// Key bindings: q (no modifier), Ctrl+Q, Ctrl+W → quit
	win.Connect("key-press-event", func(_ *gtk.Window, ev *gdk.Event) bool {
		keyEvent := gdk.EventKeyNewFromEvent(ev)
		key := keyEvent.KeyVal()
		state := gdk.ModifierType(keyEvent.State()) & gtk.AcceleratorGetDefaultModMask()

		switch {
		case key == gdk.KEY_q && state == 0:
			gtk.MainQuit()
			return true
		case key == gdk.KEY_q && state == gdk.CONTROL_MASK:
			gtk.MainQuit()
			return true
		case key == gdk.KEY_w && state == gdk.CONTROL_MASK:
			gtk.MainQuit()
			return true
		}
		return false
	})

	webview := NewWebView()
	webview.AddToWindow(uintptr(win.Native()))
	webview.LoadHTML(htmlContent, baseURI)

	watchFile(absPath, func() {
		mdBytes, err := os.ReadFile(absPath)
		if err != nil {
			return // file briefly missing during an atomic save; next event retries
		}
		htmlContent, err := RenderMarkdown(mdBytes, baseDir)
		if err != nil {
			return
		}
		glib.IdleAdd(func() { webview.Reload(htmlContent, baseURI) })
	})

	win.ShowAll()
	gtk.Main()
}

// watchFile calls onChange (from a background goroutine) whenever absPath changes.
// It watches the parent directory because many editors save atomically via
// rename, which would detach a watch placed on the file itself.
func watchFile(absPath string, onChange func()) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "mdview: live reload disabled: %v\n", err)
		return
	}
	if err := w.Add(filepath.Dir(absPath)); err != nil {
		fmt.Fprintf(os.Stderr, "mdview: live reload disabled: %v\n", err)
		return
	}

	go func() {
		var timer *time.Timer
		for {
			select {
			case ev, ok := <-w.Events:
				if !ok {
					return
				}
				if filepath.Clean(ev.Name) != absPath || !ev.Has(fsnotify.Write|fsnotify.Create|fsnotify.Rename) {
					continue
				}
				// Debounce: one save often produces several events.
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(100*time.Millisecond, onChange)
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				fmt.Fprintf(os.Stderr, "mdview: watch error: %v\n", err)
			}
		}
	}()
}
