#!/bin/bash
# Creates cloud-init user-data for Ubuntu autoinstall
# Usage: ./create-cloud-init.sh <version> <ssh_pubkey>

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSION="$1"
SSH_PUBKEY="$2"

if [ -z "$VERSION" ] || [ -z "$SSH_PUBKEY" ]; then
	echo "Usage: $0 <version> <ssh_pubkey_file>"
	exit 1
fi

if [ ! -f "$SSH_PUBKEY" ]; then
	echo "SSH public key file not found: $SSH_PUBKEY"
	exit 1
fi

SSH_KEY="$(cat "$SSH_PUBKEY")"

cat << EOF
#cloud-config
autoinstall:
  version: 3
  identity:
    hostname: ubuntu-${VERSION}-test
    username: tester
    password: ''  # no password, key-only auth
  ssh:
    install-server: true
    authorized-keys:
      - "${SSH_KEY}"
  packages:
    - openssh-server
    - cloud-init
    - vim
  late-commands:
    - echo 'tester ALL=(ALL) NOPASSWD: ALL' > /target/etc/sudoers.d/tester
    - chmod 0440 /target/etc/sudoers.d/tester
EOF