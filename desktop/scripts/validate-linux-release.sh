#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

APP="bgit-desktop"
VERSION="${1:-0.12.5}"
BINARY="build/bin/${APP}-linux-amd64"
PACKAGE="build/bin/${APP}-linux-amd64.deb"
CHECK_DIR="$(mktemp -d)"

cleanup() {
    rm -rf "$CHECK_DIR"
}
trap cleanup EXIT

fail() {
    echo "ERROR: $1"
    exit 1
}

echo "======================================"
echo " BGIT Linux Release Validation"
echo " Version: $VERSION"
echo "======================================"

echo
echo "[1/7] Checking build environment"

go version
gcc --version | head -n 1
ld --version | head -n 1

export GOOS=linux
export GOARCH=amd64
export GOAMD64=v1
export CGO_ENABLED=1
export CGO_CFLAGS="-march=x86-64 -mtune=generic"
export CGO_CXXFLAGS="-march=x86-64 -mtune=generic"
export CGO_LDFLAGS="-march=x86-64 -mtune=generic"

echo "GOAMD64=$GOAMD64"
echo "CGO_CFLAGS=$CGO_CFLAGS"
echo "CGO_CXXFLAGS=$CGO_CXXFLAGS"
echo "CGO_LDFLAGS=$CGO_LDFLAGS"

echo
echo "[2/7] Building Linux desktop application"

wails build \
    -platform linux/amd64 \
    -tags webkit2_41 \
    -clean \
    -nopackage \
    -o "${APP}-linux-amd64"

[[ -f "$BINARY" ]] || fail "Build executable not found"

echo
echo "[3/7] Validating executable architecture"

file "$BINARY"
readelf -h "$BINARY" | grep -E 'Class:|Machine:'

ISA_OUTPUT="$(readelf -n "$BINARY")"
echo "$ISA_OUTPUT" | grep -i "ISA needed" || true

if echo "$ISA_OUTPUT" | grep -Eq 'x86-64-v[23]'; then
    fail "Executable requires x86-64-v2/v3"
fi

echo "PASS: Executable meets x86-64 baseline"

echo
echo "[4/7] Creating Debian package"

rm -f "$PACKAGE"

bash build/linux/package-deb.sh \
    "$VERSION" \
    "$BINARY" \
    "build/bin"

[[ -f "$PACKAGE" ]] || fail "Debian package was not created"

echo
echo "[5/7] Inspecting Debian package"

dpkg-deb --info "$PACKAGE"
dpkg-deb --contents "$PACKAGE"

echo
echo "[6/7] Extracting and validating packaged executable"

dpkg-deb -x "$PACKAGE" "$CHECK_DIR"

PACKAGED_BINARY="$CHECK_DIR/usr/bin/$APP"

[[ -f "$PACKAGED_BINARY" ]] || fail "Executable missing from package"

PACKAGED_ISA="$(readelf -n "$PACKAGED_BINARY")"
echo "$PACKAGED_ISA" | grep -i "ISA needed" || true

if echo "$PACKAGED_ISA" | grep -Eq 'x86-64-v[23]'; then
    fail "Debian package contains incompatible executable"
fi

echo "PASS: Debian package meets x86-64 baseline"

echo
echo "[7/7] Final summary"

sha256sum "$BINARY" "$PACKAGE"

echo
echo "======================================"
echo " ALL LINUX RELEASE CHECKS PASSED"
echo "======================================"