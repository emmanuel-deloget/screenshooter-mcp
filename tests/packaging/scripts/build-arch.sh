#!/bin/bash
# Build script that runs inside the Arch container

set -e

ARCH="${1}"
VERSION="${2}"

echo "Building Arch Linux server package: version=$VERSION arch=$ARCH"

# Install dependencies
pacman -Syu --noconfirm base-devel go ca-certificates

_GOARCH=amd64
if [ ${ARCH} == aarch64 ] ; then
	_GOARCH=arm64
fi

# Convert version for Arch pkgver (no hyphens allowed)
PKGVER="${VERSION#v}"
PKGVER="${PKGVER//-/.}"

# Build the Go binary
cd /project
GOARCH="${_GOARCH}" go build -buildvcs=false -trimpath \
	-ldflags="-s -w -X main.version=${VERSION}" \
	-o /output/screenshooter-mcp ./cmd/screenshooter-mcp-server

# Create build directory for makepkg
rm -rf /output/build
mkdir -p /output/build
cd /output/build

# Prepare PKGBUILD
sed "s/PKGVER_PLACEHOLDER/${PKGVER}/g;s/ARCH_PLACEHOLDER/${ARCH}/g" \
	/project/scripts/packaging/PKGBUILD-server.in > PKGBUILD

cat PKGBUILD | sed 's/^/PKGBUILD > /'

# Copy only needed files
cp /project/scripts/packaging/arch-server.install .install
cp /project/scripts/packaging/authorize-portal.sh .
cp /project/scripts/packaging/screenshooter-mcp-server.service .
cp /project/scripts/packaging/com.deloget.ScreenshooterMCP-server.desktop .
cp -r /project/gnome-extension .
cp /output/screenshooter-mcp .

# Build package as unprivileged user
useradd -m builder || true
chown -R builder:builder .
if [ "${ARCH}" = aarch64 ]; then
	su builder -c "CARCH=aarch64 makepkg --noconfirm --ignorearch"
else
	su builder -c "makepkg --noconfirm"
fi

# Copy result to output
cp *.pkg.tar.zst /output/ 2>/dev/null || true

echo "Package built: $(ls /output/*.pkg.tar.zst 2>/dev/null | head -1)"

chown -R root:root .
