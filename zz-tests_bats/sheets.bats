#! /usr/bin/env bats

setup() {
  load "$(dirname "$BATS_TEST_FILE")/common.bash"
  export output
}

teardown() {
  chflags_and_rm
}

function read_spreadsheet_returns_values { # @test
  run run_mcp_tool_call "readSpreadsheet" '{"spreadsheetId":"mock-sheet-id-456","range":"Sheet1!A1:B3"}'
  assert_success
  local row_count
  row_count=$(echo "$output" | jq '.values | length')
  assert_equal "$row_count" "3"
}

function read_spreadsheet_returns_header_row { # @test
  run run_mcp_tool_call "readSpreadsheet" '{"spreadsheetId":"mock-sheet-id-456","range":"Sheet1!A1:B3"}'
  assert_success
  local header
  header=$(echo "$output" | jq -r '.values[0][0]')
  assert_equal "$header" "Name"
}

function read_spreadsheet_returns_data_rows { # @test
  run run_mcp_tool_call "readSpreadsheet" '{"spreadsheetId":"mock-sheet-id-456","range":"Sheet1!A1:B3"}'
  assert_success
  local name
  name=$(echo "$output" | jq -r '.values[1][0]')
  assert_equal "$name" "Alice"
  local score
  score=$(echo "$output" | jq -r '.values[1][1]')
  assert_equal "$score" "95"
}
