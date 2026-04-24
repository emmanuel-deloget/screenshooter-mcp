# ScreenshooterMCP Integration Tests

This directory contains integration tests that run the MCP server in VMs to verify end-to-end functionality.

## Overview

The tests create virtual machines using KVM/libvirt, install a distribution with aDesktop environment, and run the ScreenshooterMCP server to test screenshot capture functionality.

## Requirements

### Hardware Requirements
- **KVM-capable CPU** with hardware virtualization (Intel VT-x or AMD-V)
- **RAM**: 16GB+ recommended (8GB per VM, can run 2+ VMs simultaneously)
- **Disk Space**: 200GB+ recommended
  - Base images: ~30GB each (Debian/Ubuntu GNOME), ~45GB (KDE)
  - VM clones: Same size as base images
  - Plan for ~50GB per VM (base + clone)

### Software Requirements

The following packages must be installed on the host:

#### Essential Packages
```
# Debian/Ubuntu
apt install -y \
    qemu-system-x86 \
    qemu-utils \
    libvirt-daemon-system \
    libvirt-clients \
    virt-manager \
    virtinst \
    guestfs-tools \
    cloud-image-utils \
    genisoimage \
    jq \
    openssh-client
```

#### For osinfo-db (required for --osinfo flag with virt-install)
```
# Install a recent version
apt install -y libosinfo-bin

# Or update via
dpkg -l | grep osinfo
```

To check your osinfo-db version:
```
osinfo-query-os | grep -E "debian|ubuntu|fedora" | head
```

A recent osinfo-db (2024 or newer) is recommended for proper VM template detection.

### Network Requirements
- **libvirt default network** must be active (NAT mode)
- VMs require internet access to download packages during install

## Supported Configurations

| Distribution | Version | Desktop | Mode | Status |
|------------|---------|--------|------|--------|
| Debian     | 12, 13  | GNOME  | Wayland | ✓     |
| Debian     | 12, 13  | KDE    | X11   | ✓     |
| Debian     | 12, 13  | KDE    | Wayland | ✓     |
| Ubuntu    | 24.04   | GNOME  | Wayland | ✓     |
| Ubuntu    | 24.04   | KDE    | Wayland | ✓     |
| Ubuntu    | 26.04   | GNOME  | Wayland | ✓     |
| Ubuntu    | 26.04   | KDE    | Wayland | ✓     |
| Fedora    | 43      | GNOME  | Wayland | ✓     |
| Fedora    | 43      | KDE    | Wayland | ✓     |
| Fedora    | 43      | GNOME  | X11   | ✗     |
| Fedora    | 43      | KDE    | X11   | ✗     |

**Note**: Fedora 43 does not support X11 mode. The tests will reject these combinations.

## Wayland Test Status

**All Wayland tests are expected to fail** in the current project state. This is due to Wayland screen capture not being fully implemented.

The tests can create VMs and run the MCP server, but capture operations will fail when attempting to capture screens on Wayland.

Current workarounds:
1. Use X11 mode where supported (Debian KDE/X11, Ubuntu KDE/X11)
2. Wait for Wayland support to be implemented

## Usage

### Quick Start

```bash
cd tests/integration

# Run a single test
./run.sh debian 12 gnome wayland

# List available configurations
./run.sh

# Provision a VM without running tests
./lib/provision-vm.sh <distro> <version> <desktop> <mode>
```

### Creating Base Images

Base images must be created before running tests:

```bash
# Create base image for Debian 12 GNOME
./lib/create-base-image.sh debian 12 gnome

# Create base image for Ubuntu 24.04 KDE  
./lib/create-base-image.sh ubuntu 24.04 kde
```

### Downloading ISOs

ISOs are downloaded automatically when creating base images. To pre-download:

```bash
./lib/download-iso.sh debian 12 gnome
./lib/download-iso.sh ubuntu 24.04 kde
```

## Directory Structure

```
tests/integration/
├── bases/           # Base VM images (templates)
├── vms/           # VM clones for testing
├── isos/          # Downloaded ISO files
├── keys/          # SSH keys, passwords
├── pkg/           # Downloaded packages
├── lib/           # Helper scripts
│   ├── create-base-image.sh
│   ├── provision-vm.sh
│   ├── download-iso.sh
│   ├── download-package.sh
│   └── ...
├── shared/         # Test clients (test-mcp)
└── run.sh         # Main test runner
```

## Troubleshooting

### "No valid CPU mode"
Ensure CPU virtualization is enabled in BIOS/UEFI.

### "Network 'default' not found"
Run:
```bash
virsh net-define /dev/stdin <<EOF
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
virsh net-start default
virsh net-autostart default
```

### Disk space issues
Remove old VMs:
```bash
virsh destroy test-*
virsh undefine test-* --nvram
rm -f vms/*.qcow2
```

### Permission issues
Some operations require root:
```bash
sudo chown $USER:$USER /var/lib/libvirt/qemu/nvram/
```

## Notes

- Tests use **transient VMs** (destroyed after test)
- Each test run creates a fresh VM clone
- VM images use UEFI boot with OVMF
- Serial console is configured for debugging
- cloud-init is used for VM configuration (Ubuntu)