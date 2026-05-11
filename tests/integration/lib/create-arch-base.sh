#!/bin/bash
# Creates base Arch Linux image from pre-built qcow2 cloud image
# Usage: ./create-arch-base.sh <distro> <version> <desktop>
#
# Customizes the downloaded Arch cloud image with:
# - Desktop environment (GNOME or KDE)
# - tester user with SSH key-based access
# - sudo privileges
# - SSH server enabled
#
# Base images are stored in ../bases/

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KEYS_DIR="$(cd "$SCRIPT_DIR/../keys" && pwd)"
ISOS_DIR="$(cd "$SCRIPT_DIR/../isos" && pwd)"
BASES_DIR="$(cd "$SCRIPT_DIR/../bases" && pwd)"
SSH_KEY="${KEYS_DIR}/test-key.pub"
PASSWORD_FILE="${KEYS_DIR}/user-password-file"

if [ ! -f "$PASSWORD_FILE" ]; then
	echo "Password file not found: $PASSWORD_FILE"
	exit 1
fi

USER_PASSWORD="$(cat "$PASSWORD_FILE")"

DISTRO="$1"
VERSION="$2"
DESKTOP="$3"

if [ -z "$DISTRO" ] || [ -z "$VERSION" ] || [ -z "$DESKTOP" ]; then
	echo "Usage: $0 <distro> <version> <desktop>"
	echo "  distro: arch"
	echo "  version: latest"
	echo "  desktop: gnome, kde"
	exit 1
fi

VM_NAME="base-${DISTRO}-${VERSION}-${DESKTOP}"
SOURCE_IMAGE="${ISOS_DIR}/${DISTRO}-${VERSION}.qcow2"
BASE_IMAGE="${BASES_DIR}/${DISTRO}-${VERSION}-${DESKTOP}.qcow2"

if [ ! -f "$SSH_KEY" ]; then
	echo "SSH key not found. Run create-ssh-key.sh first."
	exit 1
fi

if [ ! -f "$SOURCE_IMAGE" ]; then
	echo "Source image not found. Run download-image.sh first."
	exit 1
fi

if [ -f "$BASE_IMAGE" ]; then
	echo "Base image already exists: $BASE_IMAGE"
	exit 0
fi

mkdir -p "$BASES_DIR"

# Determine disk size and packages based on desktop
case "${DESKTOP}" in
[gG][nN][oO][mM][eE])
	DISK_SIZE="30G"
	PACKAGES="gnome gnome-extra gdm spice-vdagent qemu-guest-agent openssh sudo cloud-init"
	;;
[kK][dD][Ee])
	DISK_SIZE="45G"
	PACKAGES="plasma sddm kde-applications spice-vdagent qemu-guest-agent openssh sudo cloud-init"
	;;
*)
	DISK_SIZE="30G"
	PACKAGES="gnome gnome-extra gdm spice-vdagent qemu-guest-agent openssh sudo cloud-init"
	;;
esac

echo "Creating base image for $DISTRO $VERSION ($DESKTOP)..."
echo "  VM name: $VM_NAME"
echo "  Source: $SOURCE_IMAGE"
echo "  Base image: $BASE_IMAGE"
echo "  Disk size: $DISK_SIZE"

# Use virt-resize to copy and expand in one step
echo "Copying and resizing disk to $DISK_SIZE..."
TEMP_IMAGE="${BASE_IMAGE}.tmp"
qemu-img create -f qcow2 "$TEMP_IMAGE" "$DISK_SIZE"
virt-resize --expand /dev/sda3 "$SOURCE_IMAGE" "$TEMP_IMAGE"
mv "$TEMP_IMAGE" "$BASE_IMAGE"

# Ensure btrfs filesystem is expanded
echo "Expanding btrfs filesystem..."
virt-customize \
	-a "$BASE_IMAGE" \
	--run-command "btrfs filesystem resize max / || true"

# Phase 1: Initialize pacman and install packages
echo "Installing packages (this may take a while)..."
virt-customize \
	-a "$BASE_IMAGE" \
	--run-command "pacman-key --init" \
	--run-command "pacman-key --populate archlinux" \
	--run-command "pacman -Syu --noconfirm" \
	--run-command "pacman -S --noconfirm ${PACKAGES}" \
	--run-command "fuser -km /dev 2>/dev/null || true" \
	--run-command "sync"

# Phase 2: Kernel, user, network, display manager config
echo "Configuring system..."
virt-customize \
	-a "$BASE_IMAGE" \
	--run-command "KVER=\$(ls /usr/lib/modules | head -1) && ln -sf /usr/lib/modules/\${KVER}/vmlinuz /boot/vmlinuz-linux" \
	--run-command "mkinitcpio -k /boot/vmlinuz-linux -g /boot/initramfs-linux.img" \
	--run-command "useradd -m -s /bin/bash tester" \
	--run-command "echo 'tester:${USER_PASSWORD}' | chpasswd" \
	--run-command "echo 'tester ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/tester" \
	--run-command "chmod 440 /etc/sudoers.d/tester" \
	--upload "${SSH_KEY}:/tmp/tester.pub" \
	--run-command "mkdir -p /home/tester/.ssh && cat /tmp/tester.pub >> /home/tester/.ssh/authorized_keys" \
	--run-command "chmod 700 /home/tester/.ssh && chmod 600 /home/tester/.ssh/authorized_keys" \
	--run-command "chown -R tester:tester /home/tester/.ssh" \
	--run-command "mkdir -p /etc/systemd/network && printf '[Match]\nName=eth0 enp* ens*\n\n[Network]\nDHCP=yes\n' > /etc/systemd/network/20-virtio.network" \
	--run-command "ln -sf /usr/lib/systemd/system/systemd-networkd.service /etc/systemd/system/multi-user.target.wants/systemd-networkd.service" \
	--run-command "ln -sf /usr/lib/systemd/system/systemd-resolved.service /etc/systemd/system/multi-user.target.wants/systemd-resolved.service" \
	--run-command "ln -sf /usr/lib/systemd/system/systemd-networkd-wait-online.service /etc/systemd/system/multi-user.target.wants/systemd-networkd-wait-online.service" \
	--run-command "ln -sf /usr/lib/systemd/system/sshd.service /etc/systemd/system/multi-user.target.wants/sshd.service" \
	--run-command "rm -f /tmp/tester.pub"

# Phase 3: Display manager configuration
echo "Configuring display manager..."
DESKTOP_LOWER="$(echo "${DESKTOP}" | tr '[:upper:]' '[:lower:]')"
case "${DESKTOP_LOWER}" in
	gnome)
		GDM_CONF=$(mktemp)
		cat > "$GDM_CONF" << 'GDM_EOF'
[daemon]
AutomaticLoginEnable=true
AutomaticLogin=tester

[security]
DisallowTCP=false
GDM_EOF
		virt-customize \
			-a "$BASE_IMAGE" \
			--upload "${GDM_CONF}:/etc/gdm/custom.conf" \
			--run-command "ln -sf /usr/lib/systemd/system/gdm.service /etc/systemd/system/display-manager.service"
		rm -f "$GDM_CONF"
		;;
	kde)
		SDDM_AUTOLOGIN=$(mktemp)
		cat > "$SDDM_AUTOLOGIN" << 'SDDM_EOF'
[Autologin]
User=tester
Session=plasma
SDDM_EOF
		SDDM_WAYLAND=$(mktemp)
		cat > "$SDDM_WAYLAND" << 'SDDM_EOF'
[General]
DefaultSession=plasma.desktop
SDDM_EOF
		virt-customize \
			-a "$BASE_IMAGE" \
			--run-command "mkdir -p /etc/sddm.conf.d" \
			--upload "${SDDM_AUTOLOGIN}:/etc/sddm.conf.d/autologin.conf" \
			--upload "${SDDM_WAYLAND}:/etc/sddm.conf.d/wayland.conf" \
			--run-command "ln -sf /usr/lib/systemd/system/sddm.service /etc/systemd/system/display-manager.service"
		rm -f "$SDDM_AUTOLOGIN" "$SDDM_WAYLAND"
		;;
esac

echo "Base image created: $BASE_IMAGE"
