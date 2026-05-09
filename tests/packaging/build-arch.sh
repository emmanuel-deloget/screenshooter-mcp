#!/bin/bash
# Build Arch Linux .pkg.tar.zst package (server)
# Usage: ./build-arch.sh [version] [arch]
# Example: ./build-arch.sh 1.0.0 x86_64

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
VERSION="${1:-dev}"
ARCH="${2:-x86_64}"

case "$ARCH" in
    x86_64|aarch64) ;;
    *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

mkdir -p "$SCRIPT_DIR/output"

echo "Building Arch Linux package: version=$VERSION arch=$ARCH"

# Run build script inside archlinux container
podman run --rm \
    -v "$PROJECT_DIR:/project:ro" \
    -v "$SCRIPT_DIR/scripts:/scripts:ro" \
    -v "$SCRIPT_DIR/output:/output" \
    -w /project \
    archlinux:latest \
    bash /scripts/build-arch.sh "$ARCH" "$VERSION"

echo "Package built successfully"
ls -la "$SCRIPT_DIR/output/"
