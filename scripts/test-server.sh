#!/bin/sh

set -x

do_curl() {
	curl -s -D curl.headers -X POST http://127.0.0.1:8080/ -H "Content-Type: application/json" -o curl.response "$@"
}

json_init() {
	cat <<-EOF | jq . -c
	{
		"jsonrpc":"2.0",
		"id":1,
		"method":"initialize",
		"params":{
			"protocolVersion":"2025-03-26",
			"capabilities":{},
			"clientInfo":{
				"name":"test",
				"version":"1.0"
			}
		}
	}
	EOF
}

json_list_monitors() {
	cat <<-EOF | jq . -c
	{
		"jsonrpc":"2.0",
		"id":2,
		"method":"tools/call",
		"params":{
			"name":"list_monitors",
			"arguments":{}
		}
	}
	EOF
}

get_data() {
	grep ^data: curl.response | sed 's/^data: //'
}

get_mcp_session_id() {
	grep '^Mcp-Session-Id:' curl.headers | sed 's/^.*: //'
}

do_curl -d "$(json_init)"
session_id=$(get_mcp_session_id)

do_curl  -H "Mcp-Session-Id: ${session_id}" -d "$(json_list_monitors)"
get_data | jq .