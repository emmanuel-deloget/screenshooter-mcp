#!/bin/bash
# Build Debian .deb package
# Usage: ./build-debian.sh [version] [arch]
# Example: ./build-debian.sh 1.0.0 amd64

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
VERSION="${1:-dev}"
ARCH="${2:-amd64}"

case "$ARCH" in
    amd64|arm64) ;;
    *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

mkdir -p "$SCRIPT_DIR/output"

echo "Building Debian package: version=$VERSION arch=$ARCH"

# Run build script inside debian container
podman run --rm \
    -v "$PROJECT_DIR:/project:ro" \
    -v "$SCRIPT_DIR/scripts:/scripts:ro" \
    -v "$SCRIPT_DIR/output:/output" \
    -w /project \
    debian:bookworm-slim \
    bash /scripts/build-debian.sh "$ARCH" "$VERSION"

echo "Package built successfully"
ls -la "$SCRIPT_DIR/output/"