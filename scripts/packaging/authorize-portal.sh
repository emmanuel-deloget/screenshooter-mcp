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
portal_available() {
	gdbus call --session \
		--dest org.freedesktop.DBus \
		--object-path /org/freedesktop/DBus \
		--method org.freedesktop.DBus.NameHasOwner \
		"org.freedesktop.portal.Desktop" 2>/dev/null | grep -q "true"
}

if is_gnome_session; then
	# On GNOME, the portal must be available for screenshot authorization
	if ! portal_available; then
		echo "ERROR: xdg-desktop-portal not available, cannot authorize screenshots"
		exit 1
	fi

	# check is screenshooting is enabled
	has_auth_screenshot || auth_screenshot || true

	# check if out own extension is enabled
	has_extension || enable_extension || true
else
	# check is screenshooting is enabled
	has_auth_screenshot || auth_screenshot || true
fi
