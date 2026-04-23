#!/bin/bash
# Main test runner for ScreenshooterMCP integration tests
# Usage: ./run.sh <distro> <version> <desktop> <mode> [--keep-running]
#
# Supported distributions:
#   - debian: 12, 13
#   - ubuntu: 24.04, 25.10, 26.04
#   - fedora: 42, 43
#
# Supported desktops: gnome, kde
# Supported modes: x11, wayland

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB_DIR="$SCRIPT_DIR/lib"
KEYS_DIR="$SCRIPT_DIR/keys"
BASES_DIR="$SCRIPT_DIR/bases"
VMS_DIR="$SCRIPT_DIR/vms"
IMAGES_DIR="$SCRIPT_DIR/images"
SHARED_DIR="$SCRIPT_DIR/shared"
PKG_DIR="$SCRIPT_DIR/pkg"

# this is required if we want to enable networking on our VMs ; and it shall be used
# through all operations, including virt-install and so on.
export LIBVIRT_DEFAULT_URI=qemu:///system

DISTRO="$1"
VERSION="$2"
DESKTOP="$3"
MODE="$4"
KEEP_RUNNING=false

for arg in "$@"; do
	case "$arg" in
		--keep-running)
			KEEP_RUNNING=true
			;;
	esac
done

if [ -z "$DISTRO" ] || [ -z "$VERSION" ] || [ -z "$DESKTOP" ] || [ -z "$MODE" ]; then
	echo "Usage: $0 <distro> <version> <desktop> <mode> [--keep-running]"
	echo ""
	echo "Supported distributions:"
	echo "  - debian: 12, 13"
	echo "  - ubuntu: 24.04, 25.10, 26.04"
	echo "  - fedora: 42, 43"
	echo ""
	echo "Supported desktops: gnome, kde"
	echo "Supported modes: x11, wayland"
	echo ""
	echo "Example:"
	echo "  $0 debian 13 gnome wayland"
	echo "  $0 ubuntu 24.04 gnome x11 --keep-running"
	exit 1
fi

case "$DISTRO" in
	debian|ubuntu|fedora)
		;;
	*)
		echo "Unsupported distro: $DISTRO"
		exit 1
		;;
esac

case "$MODE" in
	x11|wayland)
		;;
	*)
		echo "Unsupported mode: $MODE (expected x11 or wayland)"
		exit 1
		;;
esac

case "$DESKTOP" in
	gnome|kde)
		;;
	*)
		echo "Unsupported desktop: $DESKTOP (expected gnome or kde)"
		exit 1
		;;
esac

VM_NAME="test-${DISTRO}-${VERSION}-${DESKTOP}-${MODE}"
BASE_IMAGE="${BASES_DIR}/${DISTRO}-${VERSION}-${DESKTOP}.qcow2"
VM_IMAGE="${VMS_DIR}/${VM_NAME}.qcow2"
IP_FILE="${VMS_DIR}/${VM_NAME}.ip"

cleanup() {
	if [ "$KEEP_RUNNING" = true ]; then
		echo "Keeping VM running (--keep-running specified)"
		echo "VM name: $VM_NAME"
		echo "To connect: ./lib/vm-lifecycle.sh ssh $VM_NAME"
		echo "To destroy: ./lib/vm-lifecycle.sh destroy $VM_NAME"
		return
	fi

	echo "Cleaning up VM..."
	"$LIB_DIR/vm-lifecycle.sh" destroy "$VM_NAME" 2>/dev/null || true
}

trap cleanup EXIT

echo "=== ScreenshooterMCP Integration Test ==="
echo "  Distribution: $DISTRO $VERSION ($DESKTOP)"
echo "  Display mode: $MODE"
echo ""

if [ ! -f "$KEYS_DIR/test-key" ]; then
	echo "Generating SSH key..."
	"$LIB_DIR/create-ssh-key.sh"
fi

echo "[1/8] Checking base image..."
if [ ! -f "$BASE_IMAGE" ]; then
	echo "Base image not found: $BASE_IMAGE"
	echo "Creating base image (this may take a while)..."

	"$LIB_DIR/download-iso.sh" "$DISTRO" "$VERSION" "$DESKTOP"
	"$LIB_DIR/create-base-image.sh" "$DISTRO" "$VERSION" "$DESKTOP"
fi
echo "  Base image ready: $BASE_IMAGE"

echo "[2/8] Provisioning VM..."
if [ -f "$VM_IMAGE" ]; then
	echo "Removing existing VM image..."
	rm -f "$VM_IMAGE"
fi

"$LIB_DIR/provision-vm.sh" "$DISTRO" "$VERSION" "$DESKTOP" "$MODE"
echo "  VM provisioned"

echo "[3/8] Waiting for VM to be ready..."
VM_IP=$(cat "$IP_FILE" 2>/dev/null || "$LIB_DIR/vm-lifecycle.sh" ip "$VM_NAME")

if [ -z "$VM_IP" ]; then
	echo "ERROR: Could not get VM IP"
	exit 1
fi
echo "  VM IP: $VM_IP"

"$LIB_DIR/vm-lifecycle.sh" wait-ssh "$VM_NAME" 120

echo "[4/8] Building test tool..."
TEST_MCP="${SHARED_DIR}/test-mcp/test-mcp"
if [ ! -f "$TEST_MCP" ]; then
	echo "Building test-mcp..."
	(cd "${SHARED_DIR}/test-mcp" && go build -o test-mcp .)
fi
if [ ! -f "$TEST_MCP" ]; then
	echo "ERROR: Failed to build test-mcp"
	exit 1
fi
echo "  Test tool ready"

echo "[5/8] Downloading package..."
"$LIB_DIR/download-package.sh" "$DISTRO"

echo "[6/8] Running test inside VM..."
"$LIB_DIR/test-mcp.sh" "$VM_IP" "$DISTRO" "$VERSION" "$DESKTOP" "$MODE"
echo "  Test completed"

echo "[7/8] Verifying results..."
RESULTS_DIR="${IMAGES_DIR}/${DISTRO}-${VERSION}-${DESKTOP}-${MODE}"
if [ -d "$RESULTS_DIR" ]; then
	echo "  Results saved to: $RESULTS_DIR"
	ls -la "$RESULTS_DIR/"
else
	echo "WARNING: No results directory found"
fi

echo ""
echo "=== Test Completed ==="
echo "Results: $RESULTS_DIR"