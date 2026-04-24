#!/bin/bash
# Downloads package from GitHub releases
# Usage: ./download-package.sh <distro> [version]
#
# Checks local directory first, downloads from GitHub if not found.
# Package is stored in ../pkg/

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
PKG_DIR="${SCRIPT_DIR}/../pkg"
mkdir -p "${PKG_DIR}"
PKG_DIR="$(cd "$SCRIPT_DIR/../pkg" && pwd)"
REPO="emmanuel-deloget/screenshooter-mcp"

DISTRO="$1"
VERSION="${2:-latest}"

mkdir -p "$PKG_DIR"

find_local_package() {
	local distro="$1"
	local pattern

	case "$distro" in
		debian|ubuntu)
			pattern="screenshooter-mcp-server_*.deb"
			;;
		fedora)
			pattern="screenshooter-mcp-server-*.rpm"
			;;
		*)
			echo "Unsupported distro: $distro"
			return 1
			;;
	esac

	ls "$PKG_DIR"/${pattern} 2>/dev/null | head -1
}

download_from_github() {
	local distro="$1"
	local version="$2"

	echo "Fetching latest tag info from GitHub..."
	local tag
	if [ "$version" = "latest" ]; then
		tag=$(curl -sSL "https://api.github.com/repos/${REPO}/tags" | jq -r '.[0].name // empty')
	else
		tag="$version"
	fi

	if [ -z "$tag" ]; then
		echo "Failed to fetch tag info"
		return 1
	fi

	echo "Latest tag: $tag"

	local assets
	assets=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/tags/${tag}" | jq -r '.assets[] | .name + " " + .browser_download_url')

	case "$distro" in
		debian|ubuntu)
			local deb_url
			deb_url=$(echo "$assets" | grep -E 'screenshooter-mcp-server_[0-9].*amd64\.deb$' | awk '{print $2}' | head -1)
			if [ -n "$deb_url" ]; then
				echo "Downloading $deb_url ..."
				curl -L -o "${PKG_DIR}/$(basename "$deb_url")" "$deb_url"
			else
				echo "No .deb package found for $tag"
				return 1
			fi
			;;
		fedora)
			local rpm_url
			rpm_url=$(echo "$assets" | grep -E 'screenshooter-mcp-server-[0-9].*x86_64\.rpm$' | awk '{print $2}' | head -1)
			if [ -n "$rpm_url" ]; then
				echo "Downloading $rpm_url ..."
				curl -L -o "${PKG_DIR}/$(basename "$rpm_url")" "$rpm_url"
			else
				echo "No .rpm package found for $tag"
				return 1
			fi
			;;
	esac
}

main() {
	if [ -z "$DISTRO" ]; then
		echo "Usage: $0 <distro> [version]"
		echo "  distro: debian, ubuntu, fedora"
		echo "  version: specific version or 'latest' (default)"
		exit 1
	fi

	local pkg
	pkg=$(find_local_package "$DISTRO")

	if [ -n "$pkg" ]; then
		echo "Found local package: $pkg"
	else
		download_from_github "$DISTRO" "$VERSION"
		pkg=$(find_local_package "$DISTRO")
	fi

	if [ -z "$pkg" ]; then
		echo "Package not found"
		exit 1
	fi

	echo "$pkg"
}

main "$@"