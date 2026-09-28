# Changelog

All notable changes to this project are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and
the project uses [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.2.1] - 2026-09-28

### Security
- Fixed: a `<meta http-equiv="refresh">` in a document could navigate the window
  to another page (0.2.0 and earlier). Exactly one navigation per programmatic
  load is now allowed; everything else, including new windows, is ignored.
- JavaScript markup (`<script>`, event handlers, `javascript:` URLs) is disabled
  in WebKit itself, in addition to the CSP.
- The CSP now also sets `base-uri 'none'` and `form-action 'none'`.
- The context menu is disabled and the network session is ephemeral.

### Fixed
- Scroll position was lost or wrong when several reloads happened in quick succession.
- Reloads no longer overlap: file events are debounced in a single goroutine.
- Directory names containing `#`, `%` or `?` now resolve relative images correctly.
- Shortcuts `q` / `Ctrl+Q` / `Ctrl+W` also work with Caps Lock on.

### Added
- `--help` and `--version`; wrong usage exits with status 2.
- Reload errors are reported on stderr instead of being ignored.
- Debian package now ships the licences of bundled Go modules, an icon, and
  correct dependencies; `make install` installs the icon too.

### Changed
- Everything (docs, test document, code comments) is in English.
- Building requires Go 1.26 or newer.

## [0.2.0] - 2026-09-28

### Changed
- Ported to **GTK 4 and WebKitGTK 6.0**. The window layer is now a small C file
  (`viewer.c`); the gotk3 dependency is gone.
- Requires Ubuntu 24.04 / Debian 13 or newer. Use 0.1.0 (GTK 3) on older systems.

## [0.1.0] - 2026-09-28

### Added
- Markdown rendering with goldmark (tables, task lists, strikethrough, autolinks, raw HTML).
- GTK 3 window with WebKitGTK; quit with `q`, `Ctrl+Q` or `Ctrl+W`.
- Live reload when the file changes, keeping the scroll position.
- Desktop entry and icon for "Open With" / default `.md` handler.
- `make install`, `make deb` and a tag-triggered release workflow that publishes a `.deb`.

### Security
- Scripts in documents are blocked by CSP; navigation and external resources are disabled.

[Unreleased]: https://github.com/grauschnabel/mdview/compare/v0.2.1...HEAD
[0.2.1]: https://github.com/grauschnabel/mdview/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/grauschnabel/mdview/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/grauschnabel/mdview/releases/tag/v0.1.0
