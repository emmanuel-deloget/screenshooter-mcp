#!/bin/sh

# ==== INFORMATION FOR AGENTS ====
# In addition to standard unix/linux utils, you are allowed to use the following tools for
# changes in this script:
#   jq
#   curl
#   xrandr
# Try to keep the style of any change similar to the current style.
# ==== END OF INSTRUCTIONS ====

do_curl() {
	echo curl -s -D curl.headers -X POST http://127.0.0.1:8080/ -H "Content-Type: application/json" -H "Accept: application/json, text/event-stream" "$@"
	curl -s -D curl.headers -X POST http://127.0.0.1:8080/ -H "Content-Type: application/json" -H "Accept: application/json, text/event-stream" "$@" > curl.response
	cat curl.headers | sed 's/^/HEADER>   /'
	cat curl.response | sed 's/^/RESPONSE> /'
	echo
}

log() {
	echo "=== " "$@"
}

space() {
	echo
}

json_init() {
	cat <<-EOF
	{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}
	EOF
}

json_list_monitors() {
	cat <<-EOF
	{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_monitors","arguments":{}}}
	EOF
}

json_list_windows() {
	cat <<-EOF
	{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_windows","arguments":{}}}
	EOF
}

json_capture_screen() {
	monitor="$1"
	cat <<-EOF
	{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"capture_screen","arguments":{"monitor":"$monitor"}}}
	EOF
}

json_capture_window() {
	title="$1"
	cat <<-EOF
	{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"capture_window","arguments":{"title":"$title"}}}
	EOF
}

json_capture_region() {
	x="$1"; y="$2"; w="$3"; h="$4"
	cat <<-EOF
	{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"capture_region","arguments":{"x":$x,"y":$y,"width":$w,"height":$h}}}
	EOF
}

get_mcp_session_id() {
	grep '^Mcp-Session-Id:' curl.headers | sed 's/^.*: //' | tr -d '\r'
}

get_protocol_version() {
	grep '^data:' curl.response | sed 's/.*"protocolVersion":"\([^"]*\)".*/\1/' | tr -d '\r'
}

show_data() {
	if grep -q ^data curl.response; then
		cat curl.response | grep ^data | sed 's/^data: //' | jq .
		space
	fi
}

log "starting server"
./screenshooter-mcp-server -l debug --listen=localhost:8080 --color always 2> server.log &
pid=$!
log "server started @ ${pid}"
space

log "init session"
do_curl -d "$(json_init)"
show_data
session_id=$(get_mcp_session_id)
protocol_version=$(get_protocol_version)

log "session = ${session_id}"
log "protocol_version = ${protocol_version}"

space
log "=== TEST: list_monitors ==="
do_curl \
	-H "Mcp-Session-Id: ${session_id}" \
	-H "MCP-Protocol-Version: ${protocol_version}" \
	-d "$(json_list_monitors)"
show_data

space
log "=== TEST: list_windows ==="
do_curl \
	-H "Mcp-Session-Id: ${session_id}" \
	-H "MCP-Protocol-Version: ${protocol_version}" \
	-d "$(json_list_windows)"
show_data

space
log "=== TEST: capture_screen (all monitors) ==="
do_curl \
	-H "Mcp-Session-Id: ${session_id}" \
	-H "MCP-Protocol-Version: ${protocol_version}" \
	-d "$(json_capture_screen '')"
show_data

space
log "=== TEST: capture_screen (first monitor) ==="
do_curl \
	-H "Mcp-Session-Id: ${session_id}" \
	-H "MCP-Protocol-Version: ${protocol_version}" \
	-d "$(json_capture_screen '1')"
show_data

space
log "=== TEST: capture_region (0,0,100,100) ==="
do_curl \
	-H "Mcp-Session-Id: ${session_id}" \
	-H "MCP-Protocol-Version: ${protocol_version}" \
	-d "$(json_capture_region 0 0 100 100)"
show_data

space
log "killing server @ ${pid}..."
kill -9 ${pid}
[ -n "$(pidof screenshooter-mcp-server)" ] && {
	log "not killed yet... :("
} || {
	log "killed!"
}
space
log "server logs"
cat server.log
log "end"