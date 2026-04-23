#!/bin/bash
# VM lifecycle management helpers
# Usage: ./vm-lifecycle.sh <command> [vm-name]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VMS_DIR="$(cd "$SCRIPT_DIR/../vms" && pwd)"
KEYS_DIR="$(cd "$SCRIPT_DIR/../keys" && pwd)"

# this is required if we want to enable networking on our VMs ; and it shall be used
# through all operations, including virt-install and so on.
export LIBVIRT_DEFAULT_URI=qemu:///system

COMMAND="$1"
VM_NAME="$2"

get_vm_ip() {
	local vm="$1"
	local ip_file="${VMS_DIR}/${vm}.ip"
	local ip

	if [ -f "$ip_file" ]; then
		ip=$(cat "$ip_file")
	else
		ip=$(virsh domifaddr "$vm" 2>/dev/null | grep ipv4 | head -n 1 | awk '{ print $4 }' | sed 's,/.*$,,')
	fi

	echo "$ip"
}

ssh_vm() {
	local vm="$1"
	shift
	local ip
	ip=$(get_vm_ip "$vm")

	if [ -z "$ip" ]; then
		echo "Cannot determine VM IP for $vm"
		return 1
	fi

	ssh -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o ConnectTimeout=10 -i "${KEYS_DIR}/test-key" "tester@${ip}" "$@"
}

scp_to_vm() {
	local vm="$1"
	local src="$2"
	shift 2
	local ip
	ip=$(get_vm_ip "$vm")

	if [ -z "$ip" ]; then
		echo "Cannot determine VM IP for $vm"
		return 1
	fi

	scp -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o ConnectTimeout=10 -i "${KEYS_DIR}/test-key" "$src" "tester@${ip}:$@"
}

scp_from_vm() {
	local vm="$1"
	shift
	local ip
	ip=$(get_vm_ip "$vm")

	if [ -z "$ip" ]; then
		echo "Cannot determine VM IP for $vm"
		return 1
	fi

	scp -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o ConnectTimeout=10 -i "${KEYS_DIR}/test-key" "tester@${ip}:$@"
}

destroy_vm() {
	local vm="$1"

	echo "Destroying VM: $vm"
	virsh destroy "$vm" 2>/dev/null || true

	local vm_image="${VMS_DIR}/${vm}.qcow2"
	if [ -f "$vm_image" ]; then
		echo "Removing VM image: $vm_image"
		rm -f "$vm_image"
	fi

	rm -f "${VMS_DIR}/${vm}.ip" "${VMS_DIR}/${vm}.env"

	virsh undefine "$vm" 2>/dev/null || true
}

start_vm() {
	local vm="$1"

	echo "Starting VM: $vm"
	virsh start "$vm"

	local ip
	ip=$(get_vm_ip "$vm")

	if [ -z "$ip" ]; then
		echo "Waiting for IP..."
		sleep 10
		ip=$(get_vm_ip "$vm")
	fi

	echo "VM IP: $ip"
}

stop_vm() {
	local vm="$1"

	echo "Stopping VM: $vm"
	virsh shutdown "$vm" 2>/dev/null || virsh destroy "$vm" 2>/dev/null || true
}

wait_for_ssh() {
	local vm="$1"
	local timeout="${2:-300}"
	local count=0

	while [ $count -lt $timeout ]; do
		local ip
		ip=$(get_vm_ip "$vm")

		if [ -n "$ip" ]; then
			if ssh -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o ConnectTimeout=5 -i "${KEYS_DIR}/test-key" "tester@${ip}" "echo ok" 2>/dev/null | grep -q ok; then
				echo "SSH ready at $ip"
				return 0
			fi
		fi

		sleep 5
		count=$((count + 5))
	done

	echo "SSH did not become ready within ${timeout}s"
	return 1
}

case "$COMMAND" in
	destroy)
		if [ -z "$VM_NAME" ]; then
			echo "Usage: $0 destroy <vm-name>"
			exit 1
		fi
		destroy_vm "$VM_NAME"
		;;
	start)
		if [ -z "$VM_NAME" ]; then
			echo "Usage: $0 start <vm-name>"
			exit 1
		fi
		start_vm "$VM_NAME"
		;;
	stop)
		if [ -z "$VM_NAME" ]; then
			echo "Usage: $0 stop <vm-name>"
			exit 1
		fi
		stop_vm "$VM_NAME"
		;;
	ssh)
		if [ -z "$VM_NAME" ]; then
			echo "Usage: $0 ssh <vm-name> <command>"
			exit 1
		fi
		shift 2
		ssh_vm "$VM_NAME" "$@"
		;;
	scp-to)
		if [ -z "$VM_NAME" ]; then
			echo "Usage: $0 scp-to <vm-name> <source> <dest>"
			exit 1
		fi
		shift 2
		scp_to_vm "$VM_NAME" "$@"
		;;
	scp-from)
		if [ -z "$VM_NAME" ]; then
			echo "Usage: $0 scp-from <vm-name> <source> <dest>"
			exit 1
		fi
		shift 2
		scp_from_vm "$VM_NAME" "$@"
		;;
	wait-ssh)
		if [ -z "$VM_NAME" ]; then
			echo "Usage: $0 wait-ssh <vm-name> [timeout]"
			exit 1
		fi
		wait_for_ssh "$VM_NAME" "${3:-300}"
		;;
	ip)
		if [ -z "$VM_NAME" ]; then
			echo "Usage: $0 ip <vm-name>"
			exit 1
		fi
		get_vm_ip "$VM_NAME"
		;;
	list)
		echo "Available VMs:"
		virsh list --all --name 2>/dev/null | grep test- || echo "  (none)"
		;;
	*)
		echo "Usage: $0 <command> [args]"
		echo ""
		echo "Commands:"
		echo "  destroy <vm-name>	 Destroy VM and remove image"
		echo "  start <vm-name>	   Start a stopped VM"
		echo "  stop <vm-name>		Stop a running VM"
		echo "  ssh <vm-name> <cmd>   Run command via SSH"
		echo "  scp-to <vm-name> <src> <dest>  Copy file to VM"
		echo "  scp-from <vm-name> <src> <dest>  Copy file from VM"
		echo "  wait-ssh <vm-name> [timeout]  Wait for SSH to be ready"
		echo "  ip <vm-name>		  Get VM IP address"
		echo "  list				  List all test VMs"
		;;
esac