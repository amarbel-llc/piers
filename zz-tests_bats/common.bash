bats_load_library bats-support
bats_load_library bats-assert
bats_load_library bats-assert-additions

set_xdg() {
  loc="$(realpath "$1" 2>/dev/null)"
  export XDG_DATA_HOME="$loc/.xdg/data"
  export XDG_CONFIG_HOME="$loc/.xdg/config"
  export XDG_STATE_HOME="$loc/.xdg/state"
  export XDG_CACHE_HOME="$loc/.xdg/cache"
  export XDG_RUNTIME_HOME="$loc/.xdg/runtime"
  mkdir -p "$XDG_DATA_HOME" "$XDG_CONFIG_HOME" "$XDG_STATE_HOME" \
    "$XDG_CACHE_HOME" "$XDG_RUNTIME_HOME"
}

setup_test_home() {
  export REAL_HOME="$HOME"
  export HOME="$BATS_TEST_TMPDIR/home"
  mkdir -p "$HOME"
  set_xdg "$BATS_TEST_TMPDIR"
}

chflags_and_rm() {
  chflags -R nouchg "$BATS_TEST_TMPDIR" 2>/dev/null || true
  rm -rf "$BATS_TEST_TMPDIR"
}

# Path to the built MCP server binary
MCP_BIN="${MCP_BIN:-$(dirname "$BATS_TEST_FILE")/../result/bin/piers}"

# Standard MCP initialization handshake messages
MCP_INIT='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"bats-test","version":"0.0.1"}}}'
MCP_INITIALIZED='{"jsonrpc":"2.0","method":"notifications/initialized"}'

# Send JSON-RPC messages to the MCP server and capture the response for a
# specific request id.
# Usage: run_mcp <response_id> <json_line_1> [json_line_2] ...
# Sets $output to the full JSON-RPC response line matching the given id.
run_mcp() {
  local response_id="$1"
  shift

  local input=""
  for msg in "$@"; do
    input+="$msg"$'\n'
  done

  local response
  response=$(printf '%s' "$input" \
    | MOCK_AUTH=1 timeout --preserve-status 5s "$MCP_BIN" 2>/dev/null \
    | grep -F "\"id\":$response_id" \
    | head -1)

  if [ -z "$response" ]; then
    echo "no response for id $response_id"
    return 1
  fi

  echo "$response"
}

# Send an MCP tools/list request. Performs init handshake first.
# Sets $output to the tools/list response JSON.
run_mcp_tools_list() {
  local list_request='{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
  run_mcp 2 "$MCP_INIT" "$MCP_INITIALIZED" "$list_request"
}

# Send an MCP tools/call request. Performs init handshake first.
# Usage: run_mcp_tool_call <tool_name> <json_args>
# Sets $output to the result content text.
run_mcp_tool_call() {
  local tool_name="$1"
  local tool_args="$2"
  local call_request
  call_request=$(printf '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"%s","arguments":%s}}' "$tool_name" "$tool_args")

  local response
  response=$(run_mcp 3 "$MCP_INIT" "$MCP_INITIALIZED" "$call_request")
  if [ $? -ne 0 ]; then
    echo "$response"
    return 1
  fi

  echo "$response" | jq -r '.result.content[0].text'
}
