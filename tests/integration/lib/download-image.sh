#!/bin/bash
# Downloads pre-built qcow2 images for test VMs
# Usage: ./download-image.sh <distro> <version> <desktop>
#
# Currently supports:
#   - arch: latest (from pkgbuild.com mirror)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ISOS_DIR="$(cd "$SCRIPT_DIR/../isos" && pwd)"
DOWNLOAD_DIR="$ISOS_DIR"

download_arch() {
	local version="$1"
	local desktop="$2"
	local image_file="$DOWNLOAD_DIR/arch-${version}.qcow2"

	if [ -f "$image_file" ]; then
		echo "Image already exists: $image_file"
		return 0
	fi

	local image_url="https://fastly.mirror.pkgbuild.com/images/latest/Arch-Linux-x86_64-cloudimg.qcow2"

	echo "Downloading $image_url ..."
	curl -L -o "$image_file" "$image_url"
}

main() {
	local distro="$1"
	local version="$2"
	local desktop="$3"

	if [ -z "$distro" ] || [ -z "$version" ] || [ -z "$desktop" ]; then
		echo "Usage: $0 <distro> <version> <desktop>"
		echo "  distro: arch"
		echo "  version: latest"
		echo "  desktop: gnome, kde"
		exit 1
	fi

	case "$distro" in
		arch)
			download_arch "$version" "$desktop"
			;;
		*)
			echo "Unsupported distro for image download: $distro"
			exit 1
			;;
	esac
}

main "$@"
