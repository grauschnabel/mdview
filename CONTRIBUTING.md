# Contributing

Bug reports and small, focused pull requests are welcome.

## Setup

```sh
sudo apt install golang build-essential pkg-config libgtk-4-dev libwebkitgtk-6.0-dev
make test
make install   # local install to ~/.local
```

## Before opening a pull request

- `go vet ./... && go test ./...` must pass (CI runs the same).
- Keep the security model intact: no relaxing of the CSP, no enabling navigation.
- Add a test in `render_test.go` for rendering changes and an entry in `CHANGELOG.md`.

The project is intentionally small. Features that need a settings UI, plugins or
editing support are out of scope.
