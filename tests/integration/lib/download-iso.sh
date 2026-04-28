#!/bin/bash
# Downloads ISO images for test VMs
# Usage: ./download-iso.sh <distro> <version> <desktop>

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ISOS_DIR="$(cd "$SCRIPT_DIR/../isos" && pwd)"
DOWNLOAD_DIR="$ISOS_DIR"

check_and_download() {
	local iso_url="$1"
	local iso_file="$2"

	if [ -f "$iso_file" ]; then
		echo "ISO already exists: $iso_file"
		return 0
	fi

	if curl -sIL -o /dev/null -w "%{http_code}" "$iso_url" 2>/dev/null | grep -q "200"; then
		echo "Downloading $iso_url ..."
		curl -L -o "$iso_file" "$iso_url"
	else
		echo "ERROR: ISO not found at $iso_url"
		return 1
	fi
}

download_debian() {
	local version="$1"
	local desktop="$2"
	local iso_file="$DOWNLOAD_DIR/debian-${version}-${desktop}.iso"

	local iso_url
	case "${version}-${desktop}" in
		12-gnome|12-kde)
			iso_url="https://cdimage.debian.org/cdimage/archive/12.13.0/amd64/iso-cd/debian-12.13.0-amd64-netinst.iso"
			;;
		13-gnome|13-kde)
			iso_url="https://cdimage.debian.org/mirror/cdimage/archive/13.3.0/amd64/iso-cd/debian-13.4.0-amd64-netinst.iso"
			;;
		*)
			echo "Unsupported Debian version/desktop: $version $desktop"
			return 1
			;;
	esac

	check_and_download "$iso_url" "$iso_file"
}

download_ubuntu() {
	local version="$1"
	local desktop="$2"
	local iso_file="$DOWNLOAD_DIR/ubuntu-${version}-${desktop}.iso"

	local iso_url
	case "${version}" in
		24.04)
			iso_url="https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso"
			;;
		25.10)
			iso_url="https://releases.ubuntu.com/25.10/ubuntu-25.10-live-server-amd64.iso"
			;;
		*)
			echo "Unsupported Ubuntu version: $version"
			return 1
			;;
	esac

	check_and_download "$iso_url" "$iso_file"
}

download_fedora() {
	local version="$1"
	local desktop="$2"
	local iso_file="$DOWNLOAD_DIR/fedora-${version}-${desktop}.iso"

	local iso_url
	case "${version}-${desktop}" in
		43-gnome|43-kde)
    	iso_url="https://download.fedoraproject.org/pub/fedora/linux/releases/43/Everything/x86_64/iso/Fedora-Everything-netinst-x86_64-43-1.6.iso"
    	;;
		*)
			echo "Unsupported Fedora version/desktop: $version $desktop"
			return 1
			;;
	esac

	check_and_download "$iso_url" "$iso_file"
}

main() {
	local distro="$1"
	local version="$2"
	local desktop="$3"

	if [ -z "$distro" ] || [ -z "$version" ] || [ -z "$desktop" ]; then
		echo "Usage: $0 <distro> <version> <desktop>"
		echo "  distro: debian, ubuntu, fedora"
		echo "  version: 12, 13, 24.04, 25.10, 43 (depending on the <distro> name)"
		echo "  desktop: gnome, kde"
		exit 1
	fi

	case "$distro" in
		debian)
			download_debian "$version" "$desktop"
			;;
		ubuntu)
			download_ubuntu "$version" "$desktop"
			;;
		fedora)
			download_fedora "$version" "$desktop"
			;;
		*)
			echo "Unsupported distro: $distro"
			exit 1
			;;
	esac
}

main "$@"
