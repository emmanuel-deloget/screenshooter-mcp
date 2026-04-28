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
	echo "  version: 12, 13, 24.04, 25.10, 43 (depending on the <distro> name)"
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
# Fix slow boot on Ubuntu: cloud-init searches multiple datasources with timeouts.
# Restrict to NoCloud only (our seed ISO) to eliminate 60-120s boot delay.
configure_ubuntu_fast_boot() {
	local disk="$1"

	virt-customize -a "$disk" \
		--run-command "mkdir -p /etc/cloud/cloud.cfg.d" \
		--run-command "printf 'datasource_list: [ NoCloud ]\n' > /etc/cloud/cloud.cfg.d/90_dpkg.cfg"
}

configure_display_mode_gnome_debian() {
	local disk="$1"
	local mode="$2"

	case "$mode" in
		x11)
			echo "Configuring X11 mode..."
			virt-customize -a "$disk" \
				--install xserver-xorg \
				--firstboot-command "sed -i 's/^#*WaylandEnable=.*/WaylandEnable=false/' /etc/gdm3/daemon.conf 2>/dev/null || true" \
				--firstboot-command "sed -i 's/^#WaylandEnable=false/WaylandEnable=false/' /etc/gdm3/custom.conf 2> /dev/null || true"
			;;
		wayland)
			echo "Wayland is the default mode, nothing to do"
			;;
	esac

	virt-customize -a "$VM_IMAGE" \
		--run-command "sed -i 's/^#\s*AutomaticLoginEnable\s*=.*/AutomaticLoginEnable=true/' /etc/gdm3/daemon.conf" \
		--run-command "sed -i 's/^#\s*AutomaticLogin\s*=.*/AutomaticLogin=tester/' /etc/gdm3/daemon.conf"
}

configure_display_mode_gnome_ubuntu() {
	local disk="$1"
	local mode="$2"

	case "$mode" in
		x11)
			echo "Configuring X11 mode..."
			virt-customize -a "$disk" \
				--install xserver-xorg \
				--firstboot-command "sed -i 's/^#*WaylandEnable=.*/WaylandEnable=false/' /etc/gdm3/daemon.conf 2>/dev/null || true" \
				--firstboot-command "sed -i 's/^#WaylandEnable=false/WaylandEnable=false/' /etc/gdm3/custom.conf 2> /dev/null || true"
			;;
		wayland)
			echo "Wayland is the default mode, nothing to do"
			;;
	esac

	virt-customize -a "$VM_IMAGE" \
		--run-command "sed -i 's/^#\s*AutomaticLoginEnable\s*=.*/AutomaticLoginEnable=true/' /etc/gdm3/custom.conf || true" \
		--run-command "sed -i 's/^#\s*AutomaticLogin\s*=.*/AutomaticLogin=tester/' /etc/gdm3/custom.conf || true"
}

configure_display_gnome_mode_fedora() {
	local disk="$1"
	local mode="$2"

	# there is no x11 mode in fedora, but we still need to do some adjustment

	virt-customize -a "$VM_IMAGE" \
		--run-command "grubby --update-kernel=ALL --args='console=ttyS0,115200n8'" \
		--run-command "sed -i 's/^#\s*AutomaticLoginEnable\s*=.*/AutomaticLoginEnable=true/' /etc/gdm/custom.conf" \
		--run-command "sed -i 's/^#\s*AutomaticLogin\s*=.*/AutomaticLogin=tester/' /etc/gdm/custom.conf"
}

configure_display_mode_kde_debian() {
	local disk="$1"
	local mode="$2"

	# Install KDE portal backend for Wayland screenshot support
	virt-customize -a "$VM_IMAGE" \
		--install xdg-desktop-portal-kde

	case "$mode" in
		x11)
			echo "Configuring for X11"
			virt-customize -a "$VM_IMAGE" \
				--run-command "mkdir -p /etc/sddm.conf.d" \
				--run-command "printf '[Autologin]\nUser=tester\nSession=plasma\n' > /etc/sddm.conf.d/autologin.conf"
			;;
		wayland)
			echo "Configuring for Wayland..."
			# On Debian, the Wayland session is 'plasma' (plasma.desktop in wayland-sessions)
			# DisplayServer=wayland tells SDDM to use wayland-sessions, not xsessions
			virt-customize -a "$VM_IMAGE" \
				--run-command "mkdir -p /etc/sddm.conf.d" \
				--run-command "printf '[Autologin]\nUser=tester\nSession=plasma\n' > /etc/sddm.conf.d/autologin.conf" \
				--run-command "printf '[General]\nDisplayServer=wayland\nDefaultSession=plasma.desktop\n' > /etc/sddm.conf.d/wayland.conf"
			;;
	esac
}

configure_display_mode_kde_ubuntu() {
	local disk="$1"
	local mode="$2"

	# Fix network: kde-plasma-desktop does not include network-manager
	virt-customize -a "$VM_IMAGE" \
		--install network-manager

	case "$mode" in
		x11)
			echo "Configuring for X11"
			virt-customize -a "$VM_IMAGE" \
				--install sddm-theme-breeze \
				--run-command "mkdir -p /etc/sddm.conf.d" \
				--run-command "printf '[Autologin]\nUser=tester\nSession=plasma\n' > /etc/sddm.conf.d/autologin.conf"
			;;
		wayland)
			echo "Configuring for Wayland..."
			# plasmawayland session requires plasma-workspace+plasma-session-wayland
			virt-customize -a "$VM_IMAGE" \
				--install sddm-theme-breeze,plasma-workspace,plasma-session-wayland \
				--run-command "mkdir -p /etc/sddm.conf.d" \
				--run-command "printf '[Autologin]\nUser=tester\nSession=plasmawayland\n' > /etc/sddm.conf.d/autologin.conf" \
				--run-command "printf '[General]\nDefaultSession=plasmawayland.desktop\n' > /etc/sddm.conf.d/wayland.conf"
			;;
	esac
}

configure_display_kde_mode_fedora() {
	local disk="$1"
	local mode="$2"

	# there is no x11 mode in fedora, but we still need to do some adjustment

	virt-customize -a "$VM_IMAGE" \
		--run-command "grubby --update-kernel=ALL --args='console=ttyS0,115200n8'" \
		--run-command "mkdir -p /etc/sddm.conf.d" \
		--run-command "printf '[Autologin]\nUser=tester\nSession=plasma\n' > /etc/sddm.conf.d/autologin.conf" \
		--run-command "printf '[General]\nDefaultSession=plasma-wayland.desktop\n' > /etc/sddm.conf.d/wayland.conf"
}

configure_display_mode() {
	local disk="$1"
	local mode="$2"
	local distro="$3"
	local desktop="$4"

	case "${distro}-${desktop}" in
		debian-gnome)
			configure_display_mode_gnome_debian "${disk}" "${mode}"
			;;
		ubuntu-gnome)
			configure_ubuntu_fast_boot "${disk}"
			configure_display_mode_gnome_ubuntu "${disk}" "${mode}"
			;;
		debian-kde)
			configure_display_mode_kde_debian "${disk}" "${mode}"
			;;
		ubuntu-kde)
			configure_ubuntu_fast_boot "${disk}"
			configure_display_mode_kde_ubuntu "${disk}" "${mode}"
			;;
		fedora-gnome)
			configure_display_gnome_mode_fedora "${disk}" "${mode}"
			;;
		fedora-kde)
			configure_display_kde_mode_fedora "${disk}" "${mode}"
			;;
	esac
}

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
		--run-command "sed -i 's/^SELINUX=.*/SELINUX=disabled/' /etc/selinux/config || true" \
    --install sudo \
		--run-command "echo 'tester ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/tester" \
		--run-command "chmod 440 /etc/sudoers.d/tester" \
		--run-command "mkdir -p /var/lib/systemd/linger" \
		--run-command "touch /var/lib/systemd/linger/tester"

configure_display_mode "$VM_IMAGE" "$MODE" "$DISTRO" "$DESKTOP"

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

# Define and start VM
echo "Starting VM..."

virsh destroy "$VM_NAME" 2>/dev/null || true
virsh undefine "$VM_NAME" --nvram 2>/dev/null || true

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