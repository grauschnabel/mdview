# Changelog

All notable changes to this project are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and
the project uses [Semantic Versioning](https://semver.org/).

## [0.1.0] - 2026-09-28

### Added
- Markdown rendering with goldmark (tables, task lists, strikethrough, autolinks, raw HTML).
- GTK 3 window with WebKitGTK; quit with `q`, `Ctrl+Q` or `Ctrl+W`.
- Live reload when the file changes, keeping the scroll position.
- Desktop entry and icon for "Open With" / default `.md` handler.
- `make install`, `make deb` and a tag-triggered release workflow that publishes a `.deb`.

### Security
- Scripts in documents are blocked by CSP; navigation and external resources are disabled.
