#!/bin/bash
# Creates base VM image from ISO using virt-install with autoinstall
# Usage: ./create-base-image.sh <distro> <version> <desktop>
#
# Creates a base image with:
# - tester user with SSH key-based access
# - sudo privileges
# - cloud-init and SSH server enabled
#
# Base images are stored in ../bases/

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KEYS_DIR="$(cd "$SCRIPT_DIR/../keys" && pwd)"
ISOS_DIR="$(cd "$SCRIPT_DIR/../isos" && pwd)"
BASES_DIR="$(cd "$SCRIPT_DIR/../bases" && pwd)"
SSH_KEY="${KEYS_DIR}/test-key.pub"

DISTRO="$1"
VERSION="$2"
DESKTOP="$3"

if [ -z "$DISTRO" ] || [ -z "$VERSION" ] || [ -z "$DESKTOP" ]; then
	echo "Usage: $0 <distro> <version> <desktop>"
	echo "  distro: debian, ubuntu, fedora"
	echo "  version: 12, 13, 24.04, 25.10, 26.04, 42, 43"
	echo "  desktop: gnome, kde"
	exit 1
fi

VM_NAME="base-${DISTRO}-${VERSION}-${DESKTOP}"
BASE_IMAGE="${BASES_DIR}/${DISTRO}-${VERSION}-${DESKTOP}.qcow2"
ISO_FILE="${ISOS_DIR}/${DISTRO}-${VERSION}-${DESKTOP}.iso"

if [ -f "$BASE_IMAGE" ]; then
	echo "Base image already exists: $BASE_IMAGE"
	exit 0
fi

if [ ! -f "$SSH_KEY" ]; then
	echo "SSH key not found. Run create-ssh-key.sh first."
	exit 1
fi

if [ ! -f "$ISO_FILE" ]; then
	echo "ISO not found. Run download-iso.sh first."
	exit 1
fi

mkdir -p "$BASES_DIR"

wait_for_shutdown() {
	local vm="$1"
	local timeout=60
	local count=0

	while virsh domstate "$vm" 2>/dev/null | grep -q running; do
		sleep 1
		count=$((count + 1))
		if [ $count -ge $timeout ]; then
			echo "Warning: VM did not shutdown within ${timeout}s, forcing..."
			virsh destroy "$vm" 2>/dev/null || true
			break
		fi
	done
}

get_install_with_task() {
	case "${1}" in
	[gG][nN][oO][mM][eE])
		echo gnome-desktop
		;;
	[kK][dD][eE])
		echo kde-desktop
		;;
	*)
		echo gnome-desktop
		;;
	esac
}

create_debian_image() {
	local sz=30

	[ "${2}" = "kde" ] && sz=45

	echo "Starting virt-install for Debian..."
	virt-install \
		--name "$VM_NAME" \
		--memory 8192 \
		--vcpus 2 \
		--disk path="$BASE_IMAGE",format=qcow2,size=${sz} \
		--location "$ISO_FILE" \
		--graphics spice \
		--video virtio \
		--osinfo debian${1} \
		--unattended "user-login=tester,user-password-file=${KEYS_DIR}/user-password-file,admin-password-file=${KEYS_DIR}/admin-password-file" \
		--extra-args "tasksel:tasksel/first=$(get_install_with_task ${2})" \
		--boot uefi \
		--transient \
		--wait 60

	echo "Change the owner of the generated base image"
	sudo chown "$USER:$USER" "$BASE_IMAGE"
	chmod 0644 "$BASE_IMAGE"

	echo "Post-install customization..."
	virt-customize \
		-a "$BASE_IMAGE" \
		--install spice-vdagent,spice-webdavd,qemu-guest-agent,openssh-client,openssh-server,cloud-init,sudo \
		--ssh-inject "tester:file:${SSH_KEY}" \
		--run-command "echo 'tester ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/tester" \
		--run-command "chmod 440 /etc/sudoers.d/tester" \
		--run-command "systemctl enable cloud-init" \
		--run-command "systemctl enable cloud-init-local" \
		--run-command "systemctl enable cloud-config" \
		--run-command "systemctl enable cloud-final" \
		--run-command "systemctl enable ssh || systemctl enable sshd" \
		--run-command "systemctl enable spice-vdagentd" \
		--run-command "systemctl enable qemu-guest-agent" \
		--selinux-relabel || true
}

create_ubuntu_image() {
	echo "Starting virt-install for Ubuntu..."
	virt-install \
		--name "$VM_NAME" \
		--memory 8192 \
		--vcpus 2 \
		--disk path="$BASE_IMAGE",format=qcow2,size=30 \
		--location "$ISO_FILE" \
		--graphics spice \
		--video virtio \
		--osinfo ubuntu${1} \
		--unattended "user-login=tester,user-password-file=${KEYS_DIR}/user-password-file,admin-password-file=${KEYS_DIR}/admin-password-file" \
		--boot uefi \
		--transient \
		--wait 60

	echo "Change the owner of the generated base image"
	sudo chown "$USER:$USER" "$BASE_IMAGE"
	chmod 0644 "$BASE_IMAGE"

	echo "Post-install customization..."
	virt-customize \
		-a "$BASE_IMAGE" \
		--install spice-vdagent,spice-webdavd,qemu-guest-agent,openssh-client,openssh-server,cloud-init,sudo \
		--ssh-inject "tester:file:${SSH_KEY}" \
		--run-command "echo 'tester ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/tester" \
		--run-command "chmod 440 /etc/sudoers.d/tester" \
		--run-command "systemctl enable cloud-init" \
		--run-command "systemctl enable cloud-init-local" \
		--run-command "systemctl enable cloud-config" \
		--run-command "systemctl enable cloud-final" \
		--run-command "systemctl enable ssh || systemctl enable sshd" \
		--run-command "systemctl enable spice-vdagentd" \
		--run-command "systemctl enable qemu-guest-agent" \
		--selinux-relabel || true
}

get_install_with_dnf() {
	case "${1}" in
	[gG][nN][oO][mM][eE])
		echo "@gnome-desktop"
		;;
	[kK][dD][eE])
		echo "@kde-desktop-environment"
		;;
	*)
		echo "@gnome-desktop"
		;;
	esac
}

create_fedora_image() {
	echo "Starting virt-install for Fedora..."

	virt-install \
		--name "$VM_NAME" \
		--memory 8192 \
		--vcpus 2 \
		--disk path="$BASE_IMAGE",format=qcow2,size=30 \
		--location "$ISO_FILE" \
		--graphics spice \
		--video virtio \
		--osinfo fedora${1} \
		--unattended "user-login=tester,user-password-file=${KEYS_DIR}/user-password-file,admin-password-file=${KEYS_DIR}/admin-password-file" \
		--boot uefi \
		--transient \
		--wait 60

	echo "Change the owner of the generated base image"
	sudo chown "$USER:$USER" "$BASE_IMAGE"
	chmod 0644 "$BASE_IMAGE"

	echo "Post-install customization..."
	virt-customize \
		-a "$BASE_IMAGE" \
		--install spice-vdagent,spice-webdavd,qemu-guest-agent,openssh,openssh-server,cloud-init,sudo \
		--run-command "dnf install -y $(get_install_with_dnf "${2}")" \
		--run-command "dnf install -y openssh" \
		--ssh-inject "tester:file:${SSH_KEY}" \
		--run-command "echo 'tester ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/tester" \
		--run-command "chmod 440 /etc/sudoers.d/tester" \
		--run-command "systemctl enable cloud-init" \
		--run-command "systemctl enable cloud-init-local" \
		--run-command "systemctl enable cloud-config" \
		--run-command "systemctl enable cloud-final" \
		--run-command "systemctl enable ssh || systemctl enable sshd" \
		--run-command "systemctl enable spice-vdagentd" \
		--run-command "systemctl enable qemu-guest-agent" \
		--selinux-relabel || true
}

echo "Creating base image for $DISTRO $VERSION ($DESKTOP)..."
echo "  VM name: $VM_NAME"
echo "  Base image: $BASE_IMAGE"
echo "  ISO: $ISO_FILE"

case "$DISTRO" in
	debian)
		create_debian_image $VERSION $DESKTOP
		;;
	ubuntu)
		create_ubuntu_image $VERSION $DESKTOP
		;;
	fedora)
		create_fedora_image $VERSION $DESKTOP
		;;
	*)
		echo "Unsupported distro: $DISTRO"
		exit 1
		;;
esac

echo "Base image created: $BASE_IMAGE"