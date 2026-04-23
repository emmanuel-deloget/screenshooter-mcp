#!/bin/bash
# Creates Debian preseed configuration
# Usage: ./create-preseed.sh <version> <ssh_pubkey>

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
# Preseed for Debian netinst
# https://www.debian.org/releases/stable/amd64/apb.en.html

#### Contents of the preconfiguration file
### Localization
d-i debian-installer/locale string en_US.UTF-8
d-i debian-installer/language string en
d-i debian-installer/country string US

### Network configuration
d-i netcfg/choose_interface select auto
d-i netcfg/get_hostname string debian-${VERSION}-test
d-i netcfg/get_domain string localdomain

### Mirror settings
d-i mirror/countryselect select US
d-i mirror/http/hostname string deb.debian.org
d-i mirror/http/directory string /debian
d-i mirror/http/proxy string

### Partitioning
d-i partman-auto/method string regular
d-i partman-auto/choose_recipe select atomic
d-i partman-partitioning/confirm_write_new_label boolean true
d-i partman/choose_partition select finish
d-i partman/confirm boolean true
d-i partman/confirm_nooverwrite boolean true

### Account setup
d-i passwd/make-user boolean true
d-i passwd/user-uid string 1000
d-i passwd/user-fullname string Tester User
d-i passwd/username string tester
d-i passwd/user-password-crypted password ''
d-i user-setup/allow-password-empty boolean true
d-i user-setup/encrypt-home boolean false

### Package selection
tasksel tasksel/first multiselect standard, desktop, gnome-desktop, kde-desktop
d-i pkgsel/include string openssh-server cloud-init vim

### Bootloader
d-i grub-installer/bootdev string default

### Post installation
d-i preseed/late_command string \\
    echo 'tester ALL=(ALL) NOPASSWD: ALL' > /target/etc/sudoers.d/tester; \\
    chmod 0440 /target/etc/sudoers.d/tester; \\
    mkdir -p /target/home/tester/.ssh; \\
    echo '${SSH_KEY}' > /target/home/tester/.ssh/authorized_keys; \\
    chown -R tester:tester /target/home/tester/.ssh; \\
    chmod 0700 /target/home/tester/.ssh; \\
    chmod 0600 /target/home/tester/.ssh/authorized_keys
EOF