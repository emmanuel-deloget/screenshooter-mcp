#!/bin/bash
# Creates Fedora kickstart configuration
# Usage: ./create-kickstart.sh <version> <ssh_pubkey>

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
# Kickstart for Fedora
# https://docs.fedoraproject.org/en-US/fedora/latest/html/install-guide/appe-Kickstart_Syntax_Reference.html

# System language
lang en_US.UTF-8

# Keyboard layouts
keyboard us

# System timezone
timezone UTC --utc

# Root password (empty, SSH key only)
rootpw --iscrypted --allow-empty
sshkey --username=root "${SSH_KEY}"

# Use text mode install
text

# Run the Setup Agent on first boot
firstboot --enable

# System services
services --enabled=sshd,cloud-init

# SELinux
selinux --enforcing

# Network information
network --bootproto=dhcp --device=link --activate --ipv6=auto --type

# Halt after installation
shutdown

# System bootloader
bootloader --location=mbr

# Clear disk
clearpart --all --initlabel --drives=sda

# Partition layout
part /boot --fstype=xfs --size=512 --ondisk=sda
part pv.01 --size=6144 --fstype=xfs --grow --ondisk=sda
volgroup vg_root pv.01
logvol / --fstype=xfs --size=4096 --name=lv_root --volgroup=vg_root --grow
logvol swap --fstype=swap --size=2048 --name=lv_swap --volgroup=vg_root

# Package selection (desktop environment added via live iso, but ensure packages)
%packages
@^workstation-product-environment
openssh-server
cloud-init
vim
%end

# Post installation
%post
#!/bin/bash
useradd -m -s /bin/bash tester
echo 'tester ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/tester
chmod 0440 /etc/sudoers.d/tester
mkdir -p /home/tester/.ssh
echo '${SSH_KEY}' > /home/tester/.ssh/authorized_keys
chown -R tester:tester /home/tester/.ssh
chmod 0700 /home/tester/.ssh
chmod 0600 /home/tester/.ssh/authorized_keys
%end
EOF