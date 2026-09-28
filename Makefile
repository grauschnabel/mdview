VERSION ?= 0.2.0
PREFIX ?= $(HOME)/.local
BINDIR  = $(PREFIX)/bin
APPDIR  = $(PREFIX)/share/applications

.PHONY: build install uninstall test deb

build:
	go build -o mdview .

test:
	go test ./...

deb:
	packaging/build-deb.sh $(VERSION)

install: build
	install -Dm755 mdview $(BINDIR)/mdview
	install -Dm644 packaging/mdview.desktop $(APPDIR)/mdview.desktop
	-update-desktop-database $(APPDIR)
	-xdg-mime default mdview.desktop text/markdown text/x-markdown

uninstall:
	rm -f $(BINDIR)/mdview $(APPDIR)/mdview.desktop
	-update-desktop-database $(APPDIR)
