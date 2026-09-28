#!/bin/sh
# Builds dist/mdview_<version>_<arch>.deb from the current source tree.
#
# Usage: packaging/build-deb.sh [version]
#
# Without an argument the version is derived from `git describe`. A pre-release
# suffix such as "-rc1" becomes "~rc1" so that it sorts before the final release.
set -eu

# Make file modes independent of the caller's umask.
umask 022

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

if [ $# -ge 1 ]; then
    VERSION="$1"
else
    VERSION="$(git -C "$ROOT" describe --tags --always 2>/dev/null | sed 's/^v//' || true)"
    VERSION="${VERSION:-0.0.0}"
fi
# Debian versions must start with a digit; "~" marks a pre-release.
VERSION="$(printf '%s' "$VERSION" | sed 's/-\(rc\|alpha\|beta\)/~\1/')"

ARCH="$(dpkg --print-architecture)"
PKG="$ROOT/dist/mdview_${VERSION}_${ARCH}"

rm -rf "$PKG"
mkdir -p "$PKG/DEBIAN" "$PKG/usr/bin" "$PKG/usr/share/applications" \
         "$PKG/usr/share/icons/hicolor/scalable/apps" "$PKG/usr/share/doc/mdview"

# PIE is the Debian default for executables; -trimpath keeps the build
# reproducible; the version is injected into main.version.
(cd "$ROOT" && go build -trimpath -buildmode=pie \
    -ldflags "-s -w -X main.version=$VERSION" -o "$PKG/usr/bin/mdview" .)

install -m644 "$ROOT/packaging/mdview.desktop" "$PKG/usr/share/applications/mdview.desktop"
install -m644 "$ROOT/packaging/mdview.svg" "$PKG/usr/share/icons/hicolor/scalable/apps/mdview.svg"
# DEP-5 copyright file, including the licences of the bundled Go modules.
install -m644 "$ROOT/packaging/copyright" "$PKG/usr/share/doc/mdview/copyright"
install -m644 "$ROOT/CHANGELOG.md" "$PKG/usr/share/doc/mdview/changelog"
gzip -9n "$PKG/usr/share/doc/mdview/changelog"

SIZE="$(du -sk "$PKG/usr" | cut -f1)"

cat > "$PKG/DEBIAN/control" <<CONTROL
Package: mdview
Version: $VERSION
Section: utils
Priority: optional
Architecture: $ARCH
Installed-Size: $SIZE
Depends: libc6, libglib2.0-0t64 | libglib2.0-0, libgtk-4-1, libjavascriptcoregtk-6.0-1, libwebkitgtk-6.0-4
Maintainer: Martin Kaffanke <martin@kaffanke.info>
Homepage: https://github.com/grauschnabel/mdview
Description: Minimal Markdown viewer with live reload
 Opens a Markdown file in a small GTK window, renders it with goldmark and
 reloads automatically when the file changes. JavaScript inside the document
 is blocked and navigation is disabled.
CONTROL

fakeroot dpkg-deb --root-owner-group --build "$PKG" "$ROOT/dist/" >/dev/null
echo "$ROOT/dist/mdview_${VERSION}_${ARCH}.deb"
