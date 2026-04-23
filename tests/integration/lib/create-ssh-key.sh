#!/bin/bash
# Generates SSH key for test VM access

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KEYS_DIR="$(cd "$SCRIPT_DIR/../keys" && pwd)"
KEY_FILE="$KEYS_DIR/test-key"

if [ -f "$KEY_FILE" ] && [ -f "${KEY_FILE}.pub" ]; then
    echo "SSH key already exists at $KEY_FILE"
    exit 0
fi

echo "Generating SSH key for test VM access..."
mkdir -p "$KEYS_DIR"
ssh-keygen -t ed25519 -f "$KEY_FILE" -N "" -C "screenshooter-mcp test key"
chmod 600 "$KEY_FILE"
chmod 644 "${KEY_FILE}.pub"

echo "SSH key generated successfully:"
echo "  Private: $KEY_FILE"
echo "  Public:  ${KEY_FILE}.pub"