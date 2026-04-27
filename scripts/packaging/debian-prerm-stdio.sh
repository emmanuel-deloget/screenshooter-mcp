#!/bin/sh
set -e
for uid in $(loginctl list-users --no-legend 2>/dev/null | awk '{print $1}'); do
  if [ -d "/run/user/$uid" ]; then
    su - "$(id -nu "$uid")" -c \
      "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/$uid/bus gnome-extensions disable screenshooter-mcp@deloget.com" 2>/dev/null || true
  fi
done
