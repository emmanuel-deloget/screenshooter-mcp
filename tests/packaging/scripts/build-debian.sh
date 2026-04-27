#!/bin/bash
# Build script that runs inside the container

set -e

ARCH="${1}"
VERSION="${2}"

echo "Building Debian package: version=$VERSION arch=$ARCH"

# Install dependencies
apt-get update
apt-get install -y ruby ruby-dev rubygems binutils ca-certificates curl || true

# Install Go
curl -sL "https://go.dev/dl/go1.26.2.linux-${ARCH}.tar.gz" -o /tmp/go.tar.gz || {
	echo "cannot download go compiler"
	exit 1
}
tar -C /usr/local -xzf /tmp/go.tar.gz
rm /tmp/go.tar.gz
export PATH=/usr/local/go/bin:$PATH

# Install fpm
gem install fpm --no-document

rm -rf /output/control /output/pkg /output/screenshooter-mcp /output/*.deb

# Build the Go binary
cd /project
GOARCH="${ARCH}" go build -buildvcs=false -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /output/screenshooter-mcp ./cmd/screenshooter-mcp-server

cd /output

# Create package structure
mkdir -p pkg/usr/bin
mkdir -p pkg/etc/screenshooter-mcp
mkdir -p pkg/etc/xdg/autostart
mkdir -p pkg/usr/lib/systemd/user
mkdir -p pkg/usr/lib/screenshooter-mcp
mkdir -p pkg/usr/share/screenshooter-mcp/extensions/
mkdir -p control

cp -fpR /project/gnome-extension/* pkg/usr/share/screenshooter-mcp/extensions/

echo '{"log_level":"info","color":"auto","listen":"127.0.0.1:11777"}' > pkg/etc/screenshooter-mcp/config.json

cp /project/scripts/packaging/com.deloget.ScreenshooterMCP-server.desktop pkg/etc/xdg/autostart/com.deloget.ScreenshooterMCP.desktop
cp /project/scripts/packaging/authorize-portal.sh pkg/usr/lib/screenshooter-mcp/authorize-portal.sh
cp /project/scripts/packaging/screenshooter-mcp-server.service pkg/usr/lib/systemd/user/screenshooter-mcp.service
cp screenshooter-mcp pkg/usr/bin/screenshooter-mcp-server

cp /project/scripts/packaging/debian-postinst-server.sh control/postinst
cp /project/scripts/packaging/debian-prerm-server.sh control/prerm

chmod 755 pkg/usr/lib/screenshooter-mcp/authorize-portal.sh
chmod 755 control/postinst
chmod 755 control/prerm

# Determine debian architecture name
case "$ARCH" in
    amd64) DEB_ARCH=x86_64 ;;
    arm64) DEB_ARCH=aarch64 ;;
esac

# Build package
fpm -s dir -t deb \
    -C pkg \
    -n screenshooter-mcp-server \
    -v "$VERSION" \
    --architecture "$ARCH" \
    --description "Screenshooter MCP Server (HTTP mode)" \
    --maintainer "Emmanuel Deloget <emmanuel@deloget.com>" \
    --url "https://github.com/emmanuel-deloget/screenshooter-mcp" \
    --license MIT \
    --vendor "Emmanuel Deloget" \
    --depends systemd \
    --after-install control/postinst \
    --before-remove control/prerm

echo "Package built: $(ls /output/*.deb 2>/dev/null | head -1)"
