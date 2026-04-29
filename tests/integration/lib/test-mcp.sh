#!/bin/bash
# Main test orchestrator
# Usage: ./test-mcp.sh <vm-ip> <distro> <version> <desktop> <mode>
#
# This script runs inside the VM via SSH and coordinates all tests.

set -Ee

VM_IP="$1"
DISTRO="$2"
VERSION="$3"
DESKTOP="$4"
MODE="$5"

if [ -z "$VM_IP" ] || [ -z "$DISTRO" ] || [ -z "$VERSION" ] || [ -z "$DESKTOP" ] || [ -z "$MODE" ]; then
	echo "Usage: $0 <vm-ip> <distro> <version> <desktop> <mode>"
	exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KEYS_DIR="$(cd "$SCRIPT_DIR/../keys" && pwd)"
PKG_DIR="$(cd "$SCRIPT_DIR/../pkg" && pwd)"
OUTPUT_DIR="/tmp/screenshooter-mcp-images"
SSH_OPTS="-o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o ConnectTimeout=10 -i ${KEYS_DIR}/test-key"
SSH="ssh $SSH_OPTS tester@${VM_IP}"
SCP="scp $SSH_OPTS"

echo "=== ScreenshooterMCP Integration Test ==="
echo "  VM: $VM_IP"
echo "  Distro: $DISTRO $VERSION ($DESKTOP)"
echo "  Mode: $MODE"
echo ""

# Function to run command on VM
run_on_vm() {
	$SSH "$@"
}

# Function to copy file to VM
copy_to_vm() {
	local src="$1"
	local dest="$2"
	$SCP "$src" "tester@${VM_IP}:${dest}"
}

# Function to copy file from VM
copy_from_vm() {
	local src="$1"
	local dest="$2"
	$SCP "tester@${VM_IP}:${src}" "$dest"
}

wait_vm_state() {
	local cmd="$1"
	local expected="$2"
	local cnt="$3"

	[ -n "$cnt" ] || cnt=15

	for i in $(seq 1 ${cnt}); do
		r=$(run_on_vm "$cmd" || true)
		[ "$r" = "$expected" ] && {
			echo "  condition <$cmd> = <$expected> is true after $i seconds"
			return 0
		}
		sleep 1
	done
	echo "  condition <$cmd> != <$expected> is true after $cnt seconds"
	return 1
}

echo "[1/8] Waiting for VM to be ready..."
if ! $SSH "echo ok" 2>/dev/null | grep -q ok; then
	echo "ERROR: Cannot SSH to VM"
	exit 1
fi
echo "  OK"

echo "[2/8] Creating output directory..."
run_on_vm "mkdir -p $OUTPUT_DIR"

echo "[3/8] Uploading test tool..."
TEST_MCP="${SCRIPT_DIR}/../shared/test-mcp/test-mcp"
if [ ! -f "$TEST_MCP" ]; then
	echo "ERROR: test-mcp binary not found at $TEST_MCP"
	echo "Run: cd shared/test-mcp && go build -o test-mcp ."
	exit 1
fi
copy_to_vm "$TEST_MCP" "/tmp/test-mcp"
run_on_vm "chmod +x /tmp/test-mcp"

echo "[4/8] Uploading package..."
case "$DISTRO" in
	debian|ubuntu)
		PKG_FILE=$(ls "$PKG_DIR"/screenshooter-mcp-server_*.deb 2>/dev/null | head -1)
		;;
	fedora)
		PKG_FILE=$(ls "$PKG_DIR"/screenshooter-mcp-server-*.rpm 2>/dev/null | head -1)
		;;
esac

if [ -z "$PKG_FILE" ]; then
	echo "ERROR: Package not found in $PKG_DIR"
	echo "Run: ./lib/download-package.sh $DISTRO"
	exit 1
fi

copy_to_vm "$PKG_FILE" "/tmp/package.deb"
echo "  Package: $(basename "$PKG_FILE")"

echo "[5/8] Installing package..."

case "$DISTRO" in
	debian|ubuntu)
		run_on_vm "sudo dpkg -i /tmp/package.deb"
		run_on_vm "dpkg -l screenshooter-mcp-server | grep screenshooter"
		;;
	fedora)
		run_on_vm "sudo rpm -ivh /tmp/package.deb"
		run_on_vm "rpm -q screenshooter-mcp-server"
		;;
esac
echo "  OK"

if [ "$DESKTOP" = "gnome" ] && [ "$MODE" == "wayland" ]; then
	echo "[6/8] [1] Wait for a seat0 session..."
	for i in $(seq 1 20); do
		session=$(run_on_vm "loginctl list-sessions --no-legend" | grep 'tester' | grep seat | awk '{ print $1 }')
		[ -n "${session}" ] && {
			echo "  ... session found: ${session}"
			break
		}
		echo "  ... waited $i second(s)"
		sleep 1
	done
else
	echo "[6/8] [1] nothing to do..."
fi
echo "  OK"

echo "[6/8] [2] Starting MCP server..."
run_on_vm "systemctl --user daemon-reload"
run_on_vm "systemctl --user enable --now screenshooter-mcp.service"
sleep 2
echo "  OK"

if [ "$DESKTOP" = "gnome" ] && [ "$MODE" == "wayland" ]; then
	echo "[6/8] [3] restarting the user session"
	run_on_vm "sudo loginctl terminate-session ${session}" || true
	echo "  ... wait 3 seconds after terminate-session ${session}"
	sleep 3
	run_on_vm "sudo systemctl restart gdm" || true
	echo "  OK"

	echo "[6/8] [4] Wait for a seat0 session..."
	wait_vm_state "loginctl list-sessions --no-legend | grep tester | grep seat | grep -q '.' && echo yes" "yes"
	wait_vm_state "systemctl --user is-active screenshooter-mcp.service" "active"
	sleep 1
	echo "  OK"
fi

echo "[7/8] Running MCP tools test..."
TEST_FAILED=0
run_on_vm "OUTPUT_DIR=$OUTPUT_DIR /tmp/test-mcp http://localhost:11777" || {
	echo "ERROR: test-mcp failed"
	TEST_FAILED=1
}
if [ "$TEST_FAILED" -eq 0 ]; then
	echo "  OK"
fi

echo "[8/8] Downloading results..."
IMAGES_DIR="${SCRIPT_DIR}/../images/${DISTRO}-${VERSION}-${DESKTOP}-${MODE}"
mkdir -p "$IMAGES_DIR"
run_on_vm "ls -la $OUTPUT_DIR/" || true
run_on_vm "cd $OUTPUT_DIR && for f in *; do echo \"\$f\"; done" | while read -r f; do
	if [ -n "$f" ] && [ "$f" != "*" ]; then
		copy_from_vm "$OUTPUT_DIR/$f" "$IMAGES_DIR/$f" || true
	fi
done
echo "  Images saved to: $IMAGES_DIR"

echo ""
if [ "$TEST_FAILED" -ne 0 ]; then
	echo "=== Test Completed with Errors ==="
	exit 1
fi
echo "=== Test Completed Successfully ==="
echo "Results: $IMAGES_DIR"