#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

[ -f appimage-builder.yml ] || {
	echo "We're missing some crucial file to continue. Are we in the correct directory?"
	pwd
	exit 1
}

echo "=== Cleaning old artefacts ==="
{
	rm -rf AppDir appimage-builder-cache
	rm -f screenshooter-mcp-*-x86_64.AppImage
	rm -f screenshooter-mcp-server
} > /dev/null 2>&1

OLLAMA_VERSION="0.21.0"
VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "dev")
VERSION="${VERSION#v}"

CONTAINER_IMAGE="docker.io/appimagecrafters/appimage-builder:latest"

echo "=== Building screenshooter-mcp ${VERSION} ==="
echo "Ollama version: ${OLLAMA_VERSION}"

echo ""
echo "=== Step 1: Building Go server in ${PWD} ==="
eval "$(direnv export bash)" && go build -ldflags="-s -w" -o screenshooter-mcp-server ./cmd/screenshooter-mcp-server
echo "Server binary built: screenshooter-mcp-server"

echo ""
echo "=== Step 2: Downloading Ollama ${OLLAMA_VERSION} ==="
mkdir -p bin/ollama

OLLAMA_URL="https://github.com/ollama/ollama/releases/download/v${OLLAMA_VERSION}"
TGZ_URL="${OLLAMA_URL}/ollama-linux-amd64.tar.zst"

echo "Downloading Ollama (full bundle, zstd)..."
curl -fsSL "$TGZ_URL" | tar -I zstd -xf - -C bin/ollama

echo "Ollama downloaded and extracted"

echo ""
echo "=== Step 3: Building AppImage ==="
echo "Running appimage-builder in container..."

podman run --rm \
    --security-opt seccomp=unconfined \
    --security-opt apparmor=unconfined \
    -v "$(pwd):/app" \
    -w /app \
    "$CONTAINER_IMAGE" \
    /usr/local/bin/appimage-builder --recipe appimage-builder.yml --skip-test

echo ""
echo "=== Build complete ==="
echo "Output: screenshooter-mcp-${VERSION}-x86_64.AppImage"