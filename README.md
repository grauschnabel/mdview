# mdview

[![CI](https://github.com/grauschnabel/mdview/actions/workflows/ci.yml/badge.svg)](https://github.com/grauschnabel/mdview/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/grauschnabel/mdview)](https://github.com/grauschnabel/mdview/releases)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Minimal Markdown viewer for Linux. Double-click a `.md` file, read it, and it
reloads automatically whenever the file changes.

Written in Go, rendered with [goldmark](https://github.com/yuin/goldmark) in a
GTK 3 window using WebKitGTK.

## Features

- Opens a single Markdown file in a small window (`mdview file.md`)
- **Live reload** when the file changes, keeping the scroll position
  (works with editors that save atomically)
- GitHub-flavoured Markdown: tables, task lists, strikethrough, autolinks, raw HTML
- Relative images resolve against the file's directory
- **Hardened by default:** scripts in the document never run (CSP
  `default-src 'none'; script-src 'none'`), all navigation and external
  resources are blocked
- Quit with `q`, `Ctrl+Q` or `Ctrl+W`

## Install

### Debian / Ubuntu / Pop!\_OS (.deb)

Download the `.deb` from the releases page, then:

```sh
sudo apt install ./mdview_*_amd64.deb
```

Requires Ubuntu 24.04 / Debian 13 or newer (`libwebkit2gtk-4.1`).

### From source

Build dependencies:

```sh
sudo apt install golang build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev
```

Then:

```sh
make install          # installs to ~/.local (binary + desktop entry)
make uninstall        # removes it again
make deb              # builds dist/mdview_<version>_<arch>.deb
```

`make install` also registers mdview for `text/markdown`, so it shows up under
"Open With" and can be set as the default for `.md` files.

## Usage

```sh
mdview README.md
```

## Not (yet) supported

Footnotes, definition lists, math, Mermaid and syntax highlighting are not
enabled. They appear as plain text; see `testdata/test.md` for a showcase of
what does and does not render.

## Development

```sh
go vet ./... && go test ./...
```

The rendering (`render.go`) is platform independent; only the window and
WebView layer (`main.go`, `webview.go`) is tied to GTK/WebKitGTK.

## License

MIT, see [LICENSE](LICENSE).
