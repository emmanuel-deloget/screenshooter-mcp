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
