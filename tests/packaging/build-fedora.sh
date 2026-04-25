#!/bin/bash
# Build Fedora .rpm package
# Usage: ./build-fedora.sh [version] [arch]
# Example: ./build-fedora.sh 1.0.0 x86_64

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

echo "Building Fedora package: version=$VERSION arch=$ARCH"

# Run build script inside fedora container
podman run --rm \
    -v "$PROJECT_DIR:/project:ro" \
    -v "$SCRIPT_DIR/scripts:/scripts:ro" \
    -v "$SCRIPT_DIR/output:/output" \
    -w /project \
    fedora:43 \
    bash /scripts/build-fedora.sh "$ARCH" "$VERSION"

echo "Package built successfully"
ls -la "$SCRIPT_DIR/output/"