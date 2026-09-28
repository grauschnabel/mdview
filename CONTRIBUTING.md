# Contributing

Bug reports and small, focused pull requests are welcome.

## Setup

```sh
# Go 1.26 or newer is required; the distribution package may be too old,
# in which case install Go from https://go.dev/dl/
sudo apt install golang build-essential pkg-config libgtk-4-dev libwebkitgtk-6.0-dev
make test
make install   # local install to ~/.local
```

## Before opening a pull request

- `gofmt -l .` must print nothing and `go vet ./... && go test ./...` must pass (CI runs the same).
- Write code comments and documentation in English.
- Keep the security model intact: no relaxing of the CSP, no enabling navigation.
- Add a test in `render_test.go` for rendering changes and an entry in `CHANGELOG.md`.

The project is intentionally small. Features that need a settings UI, plugins or
editing support are out of scope.
