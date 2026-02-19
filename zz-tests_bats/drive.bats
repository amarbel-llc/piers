#! /usr/bin/env bats

setup() {
  load "$(dirname "$BATS_TEST_FILE")/common.bash"
  export output
}

teardown() {
  chflags_and_rm
}

function list_documents_returns_files { # @test
  run run_mcp_tool_call "listDocuments" '{}'
  assert_success
  local doc_count
  doc_count=$(echo "$output" | jq '.documents | length')
  [ "$doc_count" -ge 1 ]
}

function list_documents_contains_mock_doc { # @test
  run run_mcp_tool_call "listDocuments" '{}'
  assert_success
  local name
  name=$(echo "$output" | jq -r '.documents[0].name')
  assert_equal "$name" "Mock Document"
}

function list_documents_includes_urls { # @test
  run run_mcp_tool_call "listDocuments" '{}'
  assert_success
  local url
  url=$(echo "$output" | jq -r '.documents[0].url')
  assert_output --partial "docs.google.com"
}
