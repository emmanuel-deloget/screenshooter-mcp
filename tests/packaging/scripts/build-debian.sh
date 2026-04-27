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

cat > pkg/etc/xdg/autostart/com.deloget.ScreenshooterMCP.desktop << 'EOF'
[Desktop Entry]
Type=Application
Name=Screenshooter MCP Server
Exec=systemctl --user start screenshooter-mcp.service
NoDisplay=true
EOF

# Create authorize-portal script
cat > pkg/usr/lib/screenshooter-mcp/authorize-portal.sh << 'EOF'
#!/bin/sh

set -e

is_gnome_session() {
	gdbus call --session \
		--dest org.freedesktop.DBus \
		--object-path /org/freedesktop/DBus \
		--method org.freedesktop.DBus.NameHasOwner 'org.gnome.Shell' 2>/dev/null | grep -q "true"
}

has_auth_screenshot() {
	gdbus call --session \
		--dest org.freedesktop.impl.portal.PermissionStore \
		--object-path /org/freedesktop/impl/portal/PermissionStore \
		--method org.freedesktop.impl.portal.PermissionStore.Lookup \
		"screenshot" "screenshot" 2>/dev/null | grep -q "'yes'"
}

auth_screenshot() {
	echo "allowing screenshot"
	gdbus call --session \
		--dest org.freedesktop.impl.portal.PermissionStore \
		--object-path /org/freedesktop/impl/portal/PermissionStore \
		--method org.freedesktop.impl.portal.PermissionStore.Set \
		"screenshot" true "screenshot" "{'': ['yes']}" "<byte 0x00>" > /dev/null 2>&1
}

has_extension() {
	gnome-extensions info screenshooter-mcp@deloget.com 2>/dev/null | grep -qi 'Enabled: Yes' && {
		gnome-extensions info screenshooter-mcp@deloget.com 2>/dev/null | grep -qi 'State: ACTIVE' && return 0
	}
	gnome-extensions info screenshooter-mcp@deloget.com 2>/dev/null | grep -qi 'State: ENABLED'
}

get_gnome_shell_version() {
	gnome-shell --version | sed 's/^.*\([0-9][0-9]\).*/\1/'
}

has_valid_extension() {
	local v=$(get_gnome_shell_version)
	local metadata="${HOME}/.local/share/gnome-shell/extensions/screenshooter-mcp@deloget.com/metadata.json"

	if ! [ -f "${metadata}" ]; then
		return 1
	fi
	grep '"shell_version"' ${metadata} | grep -q "${v}"
}

copy_extension() {
	local v=$(get_gnome_shell_version)
	local src
	local dst

	dst="${HOME}/.local/share/gnome-shell/extensions/screenshooter-mcp@deloget.com"

	case ${v} in
	4[34])
		src=/usr/share/screenshooter-mcp/extensions/screenshooter-mcp@deloget.com_legacy
		;;
	4[56789]|50)
		src=/usr/share/screenshooter-mcp/extensions/screenshooter-mcp@deloget.com_modern
		;;
	esac

	mkdir -p ${dst}
	cp -fpR ${src}/* ${dst}/
}

enable_extension() {
	has_valid_extension || {
		echo "installing extension screenshooter-mcp@deloget.com"
		copy_extension
	}
	echo "enabling gnome extension screenshooter-mcp@deloget.com"
	gsettings set org.gnome.shell disable-user-extensions false 2> /dev/null || true
	gnome-extensions disable screenshooter-mcp@deloget.com > /dev/null 2>&1 || true
	gnome-extensions enable screenshooter-mcp@deloget.com > /dev/null 2>&1
}

# check if the portal is here
gdbus call --session \
	--dest org.freedesktop.DBus \
	--object-path /org/freedesktop/DBus \
	--method org.freedesktop.DBus.NameHasOwner \
	"org.freedesktop.portal.Desktop" 2>/dev/null | grep -q "true"

if is_gnome_session; then
	# check is screenshooting is enabled
	has_auth_screenshot || auth_screenshot || true

	# check if out own extension is enabled
	has_extension || enable_extension || true
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
Requires=graphical-session.target

[Service]
Type=simple
ExecStartPre=/usr/lib/screenshooter-mcp/authorize-portal.sh
ExecStart=/usr/bin/screenshooter-mcp-server --listen 127.0.0.1:11777
Restart=on-failure
RestartSec=5
TimeoutStartSec=30

[Install]
WantedBy=default.target
EOF

# Create postinst script
cat > control/postinst << 'EOF'
#!/bin/sh
set -e

is_gnome_session() {
	local uid="${1}"
	su - "$(id -nu "$uid")" -c \
		"DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/$uid/bus \
			gdbus call --session \
			--dest org.freedesktop.DBus \
			--object-path /org/freedesktop/DBus \
			--method org.freedesktop.DBus.NameHasOwner 'org.gnome.Shell'" 2>/dev/null | grep -q "true"
}

get_gnome_shell_version() {
	gnome-shell --version | sed 's/^.*\([0-9][0-9]\).*/\1/'
}

copy_extension_for_user() {
	local v=$(get_gnome_shell_version)
	local src
	local dst

	dst="/home/${1}/.local/share/gnome-shell/extensions/screenshooter-mcp@deloget.com"

	case ${v} in
	4[34])
		src=/usr/share/screenshooter-mcp/extensions/screenshooter-mcp@deloget.com_legacy
		;;
	4[56789]|50)
		src=/usr/share/screenshooter-mcp/extensions/screenshooter-mcp@deloget.com_modern
		;;
	esac

	mkdir -p ${dst}
	cp -fpR ${src}/* ${dst}/
}

show_action=0
for uid in $(loginctl list-users --no-legend 2>/dev/null | awk '{print $1}'); do
  if [ -d "/run/user/$uid" ]; then
  	if is_gnome_session "$uid"; then
  		copy_extension_for_user "$(id -nu "$uid")"
			show_action=1
  	fi
    su - "$(id -nu "$uid")" -c "systemctl --user daemon-reload" 2>/dev/null || true
    su - "$(id -nu "$uid")" -c "systemctl --user enable screenshooter-mcp.service" 2>/dev/null || true
  fi
done
if [ ${show_action} -eq 1 ]; then
	echo ""
	echo "┌─────────────────────────────────────────────────────────────┐"
	echo "│  Screenshooter MCP: action required                         │"
	echo "│                                                             │"
	echo "│  Please log out and log back in to your desktop session     │"
	echo "│  to activate the Screenshooter MCP GNOME extension.         │"
	echo "└─────────────────────────────────────────────────────────────┘"
	echo ""
fi
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
    su - "$(id -nu "$uid")" -c \
      "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/$uid/bus gnome-extensions disable screenshooter-mcp@deloget.com" 2>/dev/null || true
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

echo "Package built: $(ls /output/*.deb 2>/dev/null | head -1)"
