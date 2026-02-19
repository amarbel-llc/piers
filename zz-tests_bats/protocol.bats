#! /usr/bin/env bats

setup() {
  load "$(dirname "$BATS_TEST_FILE")/common.bash"
  export output
}

teardown() {
  chflags_and_rm
}

function initialize_returns_protocol_version { # @test
  run run_mcp 1 "$MCP_INIT"
  assert_success
  local version
  version=$(echo "$output" | jq -r '.result.protocolVersion')
  assert_equal "$version" "2024-11-05"
}

function initialize_returns_server_info { # @test
  run run_mcp 1 "$MCP_INIT"
  assert_success
  local name
  name=$(echo "$output" | jq -r '.result.serverInfo.name')
  [ -n "$name" ]
  [ "$name" != "null" ]
}

function initialize_returns_tool_capabilities { # @test
  run run_mcp 1 "$MCP_INIT"
  assert_success
  local has_tools
  has_tools=$(echo "$output" | jq '.result.capabilities.tools')
  [ "$has_tools" != "null" ]
}

function tools_list_returns_tools { # @test
  run run_mcp_tools_list
  assert_success
  local tool_count
  tool_count=$(echo "$output" | jq '.result.tools | length')
  [ "$tool_count" -ge 30 ]
}

function tools_list_contains_read_document { # @test
  run run_mcp_tools_list
  assert_success
  local has_tool
  has_tool=$(echo "$output" | jq '[.result.tools[].name] | index("readDocument")')
  [ "$has_tool" != "null" ]
}

function tools_list_contains_list_documents { # @test
  run run_mcp_tools_list
  assert_success
  local has_tool
  has_tool=$(echo "$output" | jq '[.result.tools[].name] | index("listDocuments")')
  [ "$has_tool" != "null" ]
}

function tools_list_contains_read_spreadsheet { # @test
  run run_mcp_tools_list
  assert_success
  local has_tool
  has_tool=$(echo "$output" | jq '[.result.tools[].name] | index("readSpreadsheet")')
  [ "$has_tool" != "null" ]
}

function read_document_schema_has_document_id { # @test
  run run_mcp_tools_list
  assert_success
  local schema
  schema=$(echo "$output" | jq '.result.tools[] | select(.name == "readDocument") | .inputSchema.properties.documentId')
  [ "$schema" != "null" ]
  [ -n "$schema" ]
}
