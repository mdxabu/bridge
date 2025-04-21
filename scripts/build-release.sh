#!/bin/bash
set -e

# Get current version from version.go
VERSION=$(grep -oP 'Version = "\K[^"]+' internal/version/version.go)
COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date +%FT%T%z)

echo "Building Bridge $VERSION (commit: $COMMIT, built: $BUILD_DATE)"

# Create dist directory
mkdir -p dist

# Platforms to build for
PLATFORMS=("windows/amd64" "linux/amd64" "darwin/amd64" "darwin/arm64")

# Build for each platform
for PLATFORM in "${PLATFORMS[@]}"; do
    OS="${PLATFORM%/*}"
    ARCH="${PLATFORM#*/}"
    
    echo "Building for $OS/$ARCH..."
    
    # Set output binary name
    if [ "$OS" = "windows" ]; then
        OUTPUT="dist/bridge_${VERSION}_${OS}_${ARCH}.exe"
    else
        OUTPUT="dist/bridge_${VERSION}_${OS}_${ARCH}"
    fi
    
    # Build
    GOOS=$OS GOARCH=$ARCH go build -ldflags "-X github.com/mdxabu/bridge/internal/version.BuildDate=$BUILD_DATE -X github.com/mdxabu/bridge/internal/version.GitCommit=$COMMIT" -o "$OUTPUT" ./cmd/bridge
    
    # Create checksum
    if [ "$OS" = "windows" ]; then
        sha256sum "$OUTPUT" > "${OUTPUT}.sha256"
    else
        shasum -a 256 "$OUTPUT" > "${OUTPUT}.sha256"
    fi
done

echo "Release build complete. Binaries available in dist/"
