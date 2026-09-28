#!/bin/sh
# Builds dist/mdview_<version>_<arch>.deb from the current source tree.
# Usage: packaging/build-deb.sh [version]
set -eu

VERSION="${1:-0.2.0}"
ARCH="$(dpkg --print-architecture)"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PKG="$ROOT/dist/mdview_${VERSION}_${ARCH}"

rm -rf "$PKG"
mkdir -p "$PKG/DEBIAN" "$PKG/usr/bin" "$PKG/usr/share/applications" \
         "$PKG/usr/share/icons/hicolor/scalable/apps" "$PKG/usr/share/doc/mdview"

(cd "$ROOT" && go build -trimpath -ldflags "-s -w" -o "$PKG/usr/bin/mdview" .)

install -m644 "$ROOT/packaging/mdview.desktop" "$PKG/usr/share/applications/mdview.desktop"
install -m644 "$ROOT/packaging/mdview.svg" "$PKG/usr/share/icons/hicolor/scalable/apps/mdview.svg"
install -m644 "$ROOT/LICENSE" "$PKG/usr/share/doc/mdview/copyright"

cat > "$PKG/DEBIAN/control" <<CONTROL
Package: mdview
Version: $VERSION
Section: utils
Priority: optional
Architecture: $ARCH
Depends: libwebkitgtk-6.0-4, libgtk-4-1
Maintainer: Martin Kaffanke <martin@kaffanke.info>
Description: Minimal Markdown viewer with live reload
 Opens a Markdown file in a small GTK window, renders it with goldmark and
 reloads automatically when the file changes. JavaScript inside the document
 is blocked and navigation is disabled.
CONTROL

fakeroot dpkg-deb --build "$PKG" "$ROOT/dist/" >/dev/null
echo "$ROOT/dist/mdview_${VERSION}_${ARCH}.deb"
