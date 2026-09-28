# Version shown by `mdview --version` and used for the .deb file name.
VERSION ?= 0.2.1
# Install location; the default needs no root privileges.
PREFIX ?= $(HOME)/.local
BINDIR  = $(PREFIX)/bin
APPDIR  = $(PREFIX)/share/applications
ICONDIR = $(PREFIX)/share/icons/hicolor/scalable/apps

.PHONY: build install uninstall test deb

# Build the binary into the project directory.
build:
	go build -ldflags "-X main.version=$(VERSION)" -o mdview .

test:
	go test ./...

# Build dist/mdview_<version>_<arch>.deb (needs fakeroot and dpkg-deb).
deb:
	packaging/build-deb.sh $(VERSION)

# Install the binary, desktop entry and icon, and make mdview the default
# application for Markdown files. The Exec line is rewritten to an absolute
# path because ~/.local/bin is not always on the desktop's PATH.
install: build
	install -Dm755 mdview $(BINDIR)/mdview
	install -d $(APPDIR)
	sed 's|^Exec=mdview|Exec=$(BINDIR)/mdview|' packaging/mdview.desktop > $(APPDIR)/mdview.desktop
	chmod 644 $(APPDIR)/mdview.desktop
	install -Dm644 packaging/mdview.svg $(ICONDIR)/mdview.svg
	-update-desktop-database $(APPDIR)
	-xdg-mime default mdview.desktop text/markdown text/x-markdown

uninstall:
	rm -f $(BINDIR)/mdview $(APPDIR)/mdview.desktop $(ICONDIR)/mdview.svg
	-update-desktop-database $(APPDIR)
