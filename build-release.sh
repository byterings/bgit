#!/bin/bash

# Build script for creating release binaries for all platforms
# Users do NOT need Go installed to use these binaries

VERSION=${1:-v0.1.0}
OUTPUT_DIR="release"

echo "Building bgit $VERSION for all platforms..."
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build for Linux (AMD64)
echo "Building for Linux (AMD64)..."
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION" -o "$OUTPUT_DIR/bgit-linux-amd64" .
if [ $? -eq 0 ]; then
    echo "✓ Linux (AMD64) build successful"
else
    echo "✗ Linux (AMD64) build failed"
    exit 1
fi

# Build for Linux (ARM64) - for Raspberry Pi, etc.
echo "Building for Linux (ARM64)..."
GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=$VERSION" -o "$OUTPUT_DIR/bgit-linux-arm64" .
if [ $? -eq 0 ]; then
    echo "✓ Linux (ARM64) build successful"
else
    echo "✗ Linux (ARM64) build failed"
    exit 1
fi

# Build for macOS (Intel)
echo "Building for macOS (Intel)..."
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION" -o "$OUTPUT_DIR/bgit-darwin-amd64" .
if [ $? -eq 0 ]; then
    echo "✓ macOS (Intel) build successful"
else
    echo "✗ macOS (Intel) build failed"
    exit 1
fi

# Build for macOS (Apple Silicon)
echo "Building for macOS (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=$VERSION" -o "$OUTPUT_DIR/bgit-darwin-arm64" .
if [ $? -eq 0 ]; then
    echo "✓ macOS (Apple Silicon) build successful"
else
    echo "✗ macOS (Apple Silicon) build failed"
    exit 1
fi

# Build for Windows (AMD64)
echo "Building for Windows (AMD64)..."
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION" -o "$OUTPUT_DIR/bgit-windows-amd64.exe" .
if [ $? -eq 0 ]; then
    echo "✓ Windows (AMD64) build successful"
else
    echo "✗ Windows (AMD64) build failed"
    exit 1
fi

# Build for Windows (ARM64) - for Windows on ARM
echo "Building for Windows (ARM64)..."
GOOS=windows GOARCH=arm64 go build -ldflags "-X main.Version=$VERSION" -o "$OUTPUT_DIR/bgit-windows-arm64.exe" .
if [ $? -eq 0 ]; then
    echo "✓ Windows (ARM64) build successful"
else
    echo "✗ Windows (ARM64) build failed"
    exit 1
fi

echo ""
echo "All builds completed successfully!"
echo "Binaries are in the '$OUTPUT_DIR' directory:"
echo ""
ls -lh "$OUTPUT_DIR/"

echo ""
echo "Creating checksums..."
cd "$OUTPUT_DIR"
sha256sum * > SHA256SUMS
cd ..

echo ""
echo "✓ Release build complete for version $VERSION"
echo ""
echo "Next steps:"
echo "  1. Test the binaries on different platforms"
echo "  2. Create a GitHub release: gh release create $VERSION"
echo "  3. Upload binaries from $OUTPUT_DIR/"
echo "  4. Users can download and use WITHOUT installing Go!"
