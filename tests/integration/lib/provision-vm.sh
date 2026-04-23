#!/bin/bash
# Provisions a test VM from base image
# Usage: ./provision-vm.sh <distro> <version> <desktop> <mode>
#
# Creates a VM clone, configures X11/Wayland mode, and starts it.
# VM is stored in ../vms/

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASES_DIR="$(cd "$SCRIPT_DIR/../bases" && pwd)"
VMS_DIR="${SCRIPT_DIR}/../vms"
mkdir -p "$VMS_DIR"
VMS_DIR="$(cd "$SCRIPT_DIR/../vms" && pwd)"
KEYS_DIR="$(cd "$SCRIPT_DIR/../keys" && pwd)"

DISTRO="$1"
VERSION="$2"
DESKTOP="$3"
MODE="$4"

if [ -z "$DISTRO" ] || [ -z "$VERSION" ] || [ -z "$DESKTOP" ] || [ -z "$MODE" ]; then
	echo "Usage: $0 <distro> <version> <desktop> <mode>"
	echo "  distro: debian, ubuntu, fedora"
	echo "  version: 12, 13, 24.04, 25.10, 26.04, 42, 43"
	echo "  desktop: gnome, kde"
	echo "  mode: x11, wayland"
	exit 1
fi

VM_NAME="test-${DISTRO}-${VERSION}-${DESKTOP}-${MODE}"
BASE_IMAGE="${BASES_DIR}/${DISTRO}-${VERSION}-${DESKTOP}.qcow2"
VM_IMAGE="${VMS_DIR}/${VM_NAME}.qcow2"
BASE_NVRAM="/var/lib/libvirt/qemu/nvram/base-${DISTRO}-${VERSION}-${DESKTOP}_VARS.fd"
VM_NVRAM="${VMS_DIR}/${VM_NAME}_VARS.fd"

if [ ! -f "$BASE_IMAGE" ]; then
	echo "Base image not found: $BASE_IMAGE"
	echo "Run create-base-image.sh first."
	exit 1
fi

mkdir -p "$VMS_DIR"

echo "Provisioning VM: $VM_NAME"
echo "  Base: $BASE_IMAGE"
echo "  Target: $VM_IMAGE"
echo "  Mode: $MODE"

# Clone base image (copy to ensure standalone boot)
if [ -f "$VM_IMAGE" ]; then
	echo "VM image already exists, removing..."
	rm -f "$VM_IMAGE"
fi

cp "$BASE_IMAGE" "$VM_IMAGE"
echo "Change the owner of the vm image [requires root]"
sudo chown "$USER:$USER" "$VM_IMAGE"
chmod 0664 "$VM_IMAGE"

# Ensure disk is bootable - re-read the disk to update permissions
chmod 0664 "$VM_IMAGE"

# Configure X11/Wayland mode
configure_display_mode() {
	local disk="$1"
	local mode="$2"

	case "$mode" in
		x11)
			echo "Configuring X11 mode..."
			virt-customize -a "$disk" \
				--firstboot-command "sed -i 's/^#*WaylandEnable=.*/WaylandEnable=false/' /etc/gdm3/custom.conf 2>/dev/null || sed -i 's/^#*WaylandEnable=.*/WaylandEnable=false/' /etc/gdm/custom.conf 2>/dev/null || true"
			;;
		wayland)
			echo "Configuring Wayland mode..."
			virt-customize -a "$disk" \
				--firstboot-command "sed -i 's/^#*WaylandEnable=.*/#WaylandEnable=true/' /etc/gdm3/custom.conf 2>/dev/null || sed -i 's/^#*WaylandEnable=.*/#WaylandEnable=true/' /etc/gdm/custom.conf 2>/dev/null || true"
			;;
	esac
}

configure_display_mode "$VM_IMAGE" "$MODE"

echo "Setting up EFI fallback boot entry..."
case "$DISTRO" in
    debian)
        EFI_SRC="/boot/efi/EFI/debian/grubx64.efi"
        ;;
    ubuntu)
        EFI_SRC="/boot/efi/EFI/ubuntu/grubx64.efi"
        ;;
    fedora)
        EFI_SRC="/boot/efi/EFI/fedora/grubx64.efi"
        ;;
esac

virt-customize -a "$VM_IMAGE" \
    --run-command "mkdir -p /boot/efi/EFI/BOOT" \
    --run-command "cp ${EFI_SRC} /boot/efi/EFI/BOOT/bootx64.efi" \
    --install sudo \
		--run-command "echo 'tester ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/tester" \
		--run-command "chmod 440 /etc/sudoers.d/tester"

if ! virsh net-info default &>/dev/null; then
	echo "Defining the 'default' network"
	virsh net-define /dev/stdin <<-EOF
	<network>
		<name>default</name>
		<forward mode='nat'/>
		<bridge name='virbr0' stp='on' delay='0'/>
		<ip address='192.168.122.1' netmask='255.255.255.0'>
			<dhcp>
				<range start='192.168.122.2' end='192.168.122.254'/>
			</dhcp>
		</ip>
	</network>
	EOF
fi

if ! virsh net-info default | grep -q "Active:.*yes"; then
	virsh net-start default
fi

virsh net-autostart default

virsh net-list --all

# Define and start VM
echo "Starting VM..."

virsh destroy "$VM_NAME" 2>/dev/null || true
virsh undefine "$VM_NAME" --nvram 2>/dev/null || true

echo "Copying NVRAM [requires root]"
sudo cp "${BASE_NVRAM}" "${VM_NVRAM}"
sudo chown "$USER:$USER" "$VM_NVRAM"
chmod 0644 "$VM_NVRAM"

virsh define /dev/stdin << EOF
<domain type='kvm'>
  <name>$VM_NAME</name>
  <memory unit='MiB'>8192</memory>
  <vcpu>2</vcpu>
	<os>
		<type arch='x86_64' machine='q35'>hvm</type>
		<loader readonly='yes' type='pflash'>/usr/share/OVMF/OVMF_CODE.fd</loader>
		<nvram template='/usr/share/OVMF/OVMF_VARS.fd'/>
		<boot dev='hd'/>
	</os>
  <features>
    <acpi/>
    <apic/>
  </features>
  <cpu mode='host-model'/>
  <devices>
		<disk type='file'>
			<driver name='qemu' type='qcow2'/>
			<source file='$VM_IMAGE'/>
			<target dev='vda' bus='virtio'/>
		</disk>
		<interface type='network'>
			<source network='default'/>
			<model type='virtio'/>
		</interface>
		<graphics type='spice' autoport='yes'/>
    <video>
      <model type='virtio'/>
    </video>
    <channel type='spicevmc'>
      <target type='virtio' name='com.redhat.spice.0'/>
    </channel>
    <console type='pty'>
      <target type='serial'/>
    </console>
  </devices>
</domain>
EOF

virsh start "$VM_NAME"

echo "VM $VM_NAME started"
echo "Waiting for boot and SSH..."

wait_for_ssh() {
	local vm="$1"
	local timeout=300
	local count=0

	while [ $count -lt $timeout ]; do
		local ip
		ip=$(virsh domifaddr "$vm" 2>/dev/null | grep ipv4 | head -n 1 | awk '{ print $4 }' | sed 's,/.*$,,')

		if [ -n "$ip" ]; then
			echo "VM IP: $ip" >&2
			echo "Checking SSH..." >&2
			if ssh -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o ConnectTimeout=5 -i "${KEYS_DIR}/test-key" "tester@${ip}" "echo ok" 2>/dev/null | grep -q ok; then
				echo "SSH is ready" >&2
				echo "$ip"
				return 0
			fi
		fi

		sleep 5
		count=$((count + 5))
		echo "Waiting... ${count}s" >&2
	done

	echo "SSH did not become ready within ${timeout}s" >&2
	return 1
}

IP=$(wait_for_ssh "$VM_NAME")

echo "VM provisioned successfully:"
echo "  Name: $VM_NAME"
echo "  IP: $IP"
echo "$IP" > "${VMS_DIR}/${VM_NAME}.ip"

echo "VM_IP=$IP" > "${VMS_DIR}/${VM_NAME}.env"