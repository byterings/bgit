#!/bin/sh
# Builds the Debian/Ubuntu package for the bgit desktop app.
#
#   desktop/build/linux/package-deb.sh <version> <binary> <output-dir>
#
# Set DEB_MAINTAINER ("Name <email>") to record a contact address in the
# package; without it only the name is recorded.
#
# Example, from the repository root after `wails build`:
#   desktop/build/linux/package-deb.sh 1.2.3 desktop/build/bin/bgit-desktop-linux-amd64 desktop/build/bin
set -eu

if [ $# -ne 3 ]; then
  echo "usage: $0 <version> <binary> <output-dir>" >&2
  exit 2
fi

VERSION="$1"
BINARY="$2"
OUTPUT_DIR="$3"
HERE="$(cd "$(dirname "$0")" && pwd)"

if [ ! -f "$BINARY" ]; then
  echo "binary not found: $BINARY" >&2
  exit 1
fi

# Debian versions use "~" before a pre-release suffix so that 1.0.0~beta sorts
# before 1.0.0.
DEB_VERSION="$(printf '%s' "$VERSION" | sed 's/-/~/')"
ARCH="amd64"
MAINTAINER="${DEB_MAINTAINER:-ByteRings}"
STAGE="$(mktemp -d)"
trap 'rm -rf "$STAGE"' EXIT
chmod 0755 "$STAGE"

install -D -m 0755 "$BINARY" "$STAGE/usr/bin/bgit-desktop"
install -D -m 0644 "$HERE/bgit-desktop.desktop" "$STAGE/usr/share/applications/bgit-desktop.desktop"
install -D -m 0644 "$HERE/bgit-desktop.png" "$STAGE/usr/share/icons/hicolor/256x256/apps/bgit-desktop.png"

INSTALLED_KB="$(du -sk "$STAGE" | cut -f1)"
mkdir -p "$STAGE/DEBIAN"
cat > "$STAGE/DEBIAN/control" <<CONTROL
Package: bgit-desktop
Version: $DEB_VERSION
Section: devel
Priority: optional
Architecture: $ARCH
Maintainer: $MAINTAINER
Installed-Size: $INSTALLED_KB
Depends: libgtk-3-0, libwebkit2gtk-4.1-0, git, openssh-client
Homepage: https://bgitcli.com
Description: Manage multiple Git identities and repositories
 bgit keeps each repository on the right Git identity and SSH key, and
 checks every push before it leaves the machine.
CONTROL

mkdir -p "$OUTPUT_DIR"
# The file name carries no version so the "latest release" download link stays
# the same from release to release.
OUTPUT="$OUTPUT_DIR/bgit-desktop-linux-${ARCH}.deb"
dpkg-deb --build --root-owner-group "$STAGE" "$OUTPUT" >/dev/null
echo "$OUTPUT"
