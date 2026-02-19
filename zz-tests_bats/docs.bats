#! /usr/bin/env bats

setup() {
  load "$(dirname "$BATS_TEST_FILE")/common.bash"
  export output
}

teardown() {
  chflags_and_rm
}

function read_document_returns_content { # @test
  run run_mcp_tool_call "readDocument" '{"documentId":"mock-doc-id-123"}'
  assert_success
  assert_output --partial "Hello from the mock document."
}

function read_document_returns_json_format { # @test
  run run_mcp_tool_call "readDocument" '{"documentId":"mock-doc-id-123","format":"json"}'
  assert_success
  assert_output --partial "textRun"
}

function read_document_returns_markdown_format { # @test
  run run_mcp_tool_call "readDocument" '{"documentId":"mock-doc-id-123","format":"markdown"}'
  assert_success
  assert_output --partial "Hello from the mock document."
}
