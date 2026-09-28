// Command mdview is a minimal Markdown viewer with live reload.
//
// It renders a single Markdown file with goldmark, shows it in a GTK 4 window
// backed by WebKitGTK and re-renders whenever the file changes on disk.
// Documents are treated as untrusted; see SECURITY.md for the security model.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

const usage = `usage: mdview <file.md>

Opens a Markdown file in a window and reloads it when the file changes.
Quit with q, Ctrl+Q or Ctrl+W.

  -h, --help       show this help
  -v, --version    show the version
`

// fail prints err to stderr in the conventional "program: message" form and
// exits with status 1.
func fail(err error) {
	fmt.Fprintf(os.Stderr, "mdview: %v\n", err)
	os.Exit(1)
}

// main parses the command line, renders the file once, starts the file watcher
// and then blocks in the GTK main loop until the window is closed.
//
// Exit status: 0 on success, 1 on runtime errors, 2 on usage errors.
func main() {
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "-h", "--help":
			fmt.Print(usage)
			return
		case "-v", "--version":
			fmt.Println("mdview", version)
			return
		}
	}
	if len(os.Args) != 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	filePath := os.Args[1]
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		fail(err)
	}
	// Follow symlinks so that the watcher sees writes to the real file and
	// relative images resolve next to it.
	if resolved, err := filepath.EvalSymlinks(absPath); err == nil {
		absPath = resolved
	}
	baseDir := filepath.Dir(absPath)

	// render reads and converts the file. It is re-run on every change so the
	// content is always fresh from disk.
	render := func() (string, error) {
		mdBytes, err := os.ReadFile(absPath)
		if err != nil {
			return "", err
		}
		return RenderMarkdown(mdBytes)
	}

	htmlContent, err := render()
	if err != nil {
		fail(err)
	}
	baseURI, err := BaseURI(baseDir)
	if err != nil {
		fail(err)
	}

	// Live reload is optional: if the watcher cannot start (e.g. inotify limit
	// reached) the viewer still works, just without reloading.
	stop, err := watchFile(absPath, func() {
		htmlContent, err := render()
		if err != nil {
			// e.g. file briefly missing during an atomic save; the next event retries
			fmt.Fprintf(os.Stderr, "mdview: reload: %v\n", err)
			return
		}
		ReloadViewer(htmlContent, baseURI)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "mdview: live reload disabled: %v\n", err)
	} else {
		defer stop()
	}

	title := filepath.Base(filePath) + " — mdview"
	if err := RunViewer(title, htmlContent, baseURI); err != nil {
		fail(err)
	}
}
