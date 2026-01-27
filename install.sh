#!/bin/bash
# bgit installation script
# Usage: curl -sSL https://raw.githubusercontent.com/byterings/bgit/main/install.sh | bash

set -e

VERSION="0.1.0"
GITHUB_REPO="byterings/bgit"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case "$OS" in
    linux)
        BINARY="bgit-linux-$ARCH"
        ;;
    darwin)
        BINARY="bgit-darwin-$ARCH"
        ;;
    *)
        echo "Unsupported OS: $OS"
        echo "For Windows, please download from: https://github.com/$GITHUB_REPO/releases"
        exit 1
        ;;
esac

echo "Installing bgit v$VERSION for $OS/$ARCH..."
echo ""

# Download binary
DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/v$VERSION/$BINARY"
TMP_FILE=$(mktemp)

echo "Downloading from: $DOWNLOAD_URL"
if command -v curl &> /dev/null; then
    curl -sSL "$DOWNLOAD_URL" -o "$TMP_FILE"
elif command -v wget &> /dev/null; then
    wget -q "$DOWNLOAD_URL" -O "$TMP_FILE"
else
    echo "Error: curl or wget is required"
    exit 1
fi

# Make executable
chmod +x "$TMP_FILE"

# Install
echo "Installing to $INSTALL_DIR/bgit (requires sudo)..."
sudo mv "$TMP_FILE" "$INSTALL_DIR/bgit"

# Verify
echo ""
echo "Verifying installation..."
if command -v bgit &> /dev/null; then
    echo ""
    echo "bgit installed successfully!"
    bgit --version
    echo ""
    echo "Get started:"
    echo "  bgit add           # Add your first identity"
    echo "  bgit use <alias>   # Switch identity"
    echo "  bgit list          # List all identities"
    echo ""
else
    echo "Installation failed. Please check $INSTALL_DIR is in your PATH"
    exit 1
fi
