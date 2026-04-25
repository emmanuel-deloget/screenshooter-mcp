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
mkdir -p pkg/usr/lib/systemd/user
mkdir -p pkg/usr/lib/screenshooter-mcp
mkdir -p control

# Create authorize-portal script
cat > pkg/usr/lib/screenshooter-mcp/authorize-portal.sh << 'EOF'
#!/bin/sh
[ -f "${HOME}/.local/share/screenshooter-mcp/.portal-authorized" ] && exit 0
auth_screenshot() {
  gdbus call --session \
    --dest org.freedesktop.impl.portal.PermissionStore \
    --object-path /org/freedesktop/impl/portal/PermissionStore \
    --method org.freedesktop.impl.portal.PermissionStore.Set \
    "screenshot" true "screenshot" "{'': ['yes']}" "<byte 0x00>"
}
if auth_screenshot; then
  mkdir -p "${HOME}/.local/share/screenshooter-mcp"
  touch "${HOME}/.local/share/screenshooter-mcp/.portal-authorized"
fi
EOF
chmod 755 pkg/usr/lib/screenshooter-mcp/authorize-portal.sh

# Copy binary and config
cp screenshooter-mcp pkg/usr/bin/screenshooter-mcp-server
echo '{"log_level":"info","color":"auto","listen":"127.0.0.1:11777"}' > pkg/etc/screenshooter-mcp/config.json

# Create systemd service
cat > pkg/usr/lib/systemd/user/screenshooter-mcp.service << 'EOF'
[Unit]
Description=Screenshooter MCP Server
After=xdg-desktop-portal.service
Wants=xdg-desktop-portal.service

[Service]
Type=simple
ExecStartPre=/usr/lib/screenshooter-mcp/authorize-portal.sh
ExecStart=/usr/bin/screenshooter-mcp-server --listen 127.0.0.1:11777
Restart=on-failure

[Install]
WantedBy=default.target
EOF

# Create postinst script
cat > control/postinst << 'EOF'
#!/bin/sh
set -e
for uid in $(loginctl list-users --no-legend 2>/dev/null | awk '{print $1}'); do
  if [ -d "/run/user/$uid" ]; then
    su - "$(id -nu "$uid")" -c "systemctl --user daemon-reload" 2>/dev/null || true
    su - "$(id -nu "$uid")" -c "systemctl --user enable screenshooter-mcp.service" 2>/dev/null || true
  fi
done
EOF
chmod 755 control/postinst

# Create prerm script
cat > control/prerm << 'EOF'
#!/bin/sh
set -e
for uid in $(loginctl list-users --no-legend 2>/dev/null | awk '{print $1}'); do
  if [ -d "/run/user/$uid" ]; then
    su - "$(id -nu "$uid")" -c "systemctl --user stop screenshooter-mcp.service" 2>/dev/null || true
    su - "$(id -nu "$uid")" -c "systemctl --user disable screenshooter-mcp.service" 2>/dev/null || true
  fi
done
EOF
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

echo "Package built: screenshooter-mcp-server_${VERSION}_${DEB_ARCH}.deb"
