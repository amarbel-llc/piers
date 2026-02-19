# BATS Test Infrastructure Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Bootstrap a portable BATS test suite that validates the MCP server's JSON-RPC protocol, tool registration, and tool execution using mocked Google API responses.

**Architecture:** Add `MOCK_AUTH=1` env var support to `src/clients.ts` that bypasses Google OAuth and returns mock API clients with hardcoded responses. BATS tests send raw JSON-RPC messages over stdin to the built `dist/index.js` binary and assert on stdout responses using jq. The test infrastructure follows the batman/robin conventions (sandcastle isolation, bats-assert, TAP output).

**Tech Stack:** Nix flake (nixpkgs stable + batman), bats + bats-assert + bats-assert-additions, just, jq, Node.js

---

### Task 1: Create flake.nix

**Files:**
- Create: `flake.nix`

**Step 1: Write flake.nix**

```nix
{
  description = "Google Docs MCP Server";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/23d72dabcb3b12469f57b37170fcbc1789bd7457";
    nixpkgs-master.url = "github:NixOS/nixpkgs/b28c4999ed71543e71552ccfd0d7e68c581ba7e9";
    utils.url = "https://flakehub.com/f/numtide/flake-utils/0.1.102";
    batman.url = "github:amarbel-llc/batman";
  };

  outputs =
    {
      self,
      nixpkgs,
      nixpkgs-master,
      utils,
      batman,
    }:
    utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = (with pkgs; [
            just
            jq
            nodejs_latest
          ]) ++ [
            batman.packages.${system}.bats
            batman.packages.${system}.bats-libs
          ];
        };
      }
    );
}
```

**Step 2: Lock the flake inputs**

Run: `nix flake lock` (in repo root)
Expected: `flake.lock` created successfully

**Step 3: Verify devShell works**

Run: `nix develop --command bash -c "node --version && bats --version && just --version && jq --version"`
Expected: All four commands print version strings

**Step 4: Add flake files to git tracking**

Run: `git add flake.nix flake.lock`

**Step 5: Commit**

```bash
git commit -m "feat: add nix flake with batman for bats testing"
```

---

### Task 2: Create justfiles

**Files:**
- Create: `justfile`
- Create: `zz-tests_bats/justfile`

**Step 1: Create root justfile**

```makefile
default:
    @just --list

build:
    npm run build

test-bats: build
    just zz-tests_bats/test

test: test-bats
```

**Step 2: Create zz-tests_bats directory**

Run: `mkdir -p zz-tests_bats`

**Step 3: Create test justfile**

`zz-tests_bats/justfile`:

```makefile
bats_timeout := "10"

test-targets *targets="*.bats":
  BATS_TEST_TIMEOUT="{{bats_timeout}}" \
    bats --tap --jobs {{num_cpus()}} {{targets}}

test-tags *tags:
  BATS_TEST_TIMEOUT="{{bats_timeout}}" \
    bats --tap --jobs {{num_cpus()}} --filter-tags {{tags}} *.bats

test: (test-targets "*.bats")
```

**Step 4: Verify justfile lists recipes**

Run: `just`
Expected: Lists `build`, `test-bats`, `test` recipes

**Step 5: Commit**

```bash
git add justfile zz-tests_bats/justfile
git commit -m "feat: add justfiles for build and bats test orchestration"
```

---

### Task 3: Add mock auth support to the app

**Files:**
- Create: `src/mockClients.ts`
- Modify: `src/clients.ts:14-53`

**Step 1: Create src/mockClients.ts**

This module provides mock Google API clients that return hardcoded data. Each mock object mimics the googleapis client interface just enough for the tool handlers to work.

```typescript
// src/mockClients.ts
//
// Mock Google API clients for testing. Activated by MOCK_AUTH=1.

const MOCK_DOC = {
  documentId: 'mock-doc-id-123',
  title: 'Mock Document',
  body: {
    content: [
      {
        paragraph: {
          elements: [
            {
              textRun: {
                content: 'Hello from the mock document.\n',
              },
            },
          ],
        },
      },
    ],
  },
  tabs: [],
};

const MOCK_SPREADSHEET_VALUES = {
  range: 'Sheet1!A1:B3',
  majorDimension: 'ROWS',
  values: [
    ['Name', 'Score'],
    ['Alice', '95'],
    ['Bob', '87'],
  ],
};

const MOCK_FILE_LIST = {
  files: [
    {
      id: 'mock-doc-id-123',
      name: 'Mock Document',
      mimeType: 'application/vnd.google-apps.document',
      modifiedTime: '2025-01-15T10:30:00.000Z',
      createdTime: '2025-01-01T08:00:00.000Z',
      webViewLink: 'https://docs.google.com/document/d/mock-doc-id-123/edit',
      owners: [{ displayName: 'Test User', emailAddress: 'test@example.com' }],
    },
    {
      id: 'mock-sheet-id-456',
      name: 'Mock Spreadsheet',
      mimeType: 'application/vnd.google-apps.spreadsheet',
      modifiedTime: '2025-01-14T09:00:00.000Z',
      createdTime: '2025-01-02T08:00:00.000Z',
      webViewLink: 'https://docs.google.com/spreadsheets/d/mock-sheet-id-456/edit',
      owners: [{ displayName: 'Test User', emailAddress: 'test@example.com' }],
    },
  ],
};

function ok(data: any) {
  return { data, status: 200, statusText: 'OK', headers: {}, config: {} };
}

export function createMockDocsClient(): any {
  return {
    documents: {
      get: async () => ok(MOCK_DOC),
      batchUpdate: async () => ok({ replies: [] }),
      create: async () => ok(MOCK_DOC),
    },
  };
}

export function createMockDriveClient(): any {
  return {
    files: {
      list: async () => ok(MOCK_FILE_LIST),
      get: async () => ok(MOCK_FILE_LIST.files[0]),
      create: async () => ok(MOCK_FILE_LIST.files[0]),
      update: async () => ok(MOCK_FILE_LIST.files[0]),
      copy: async () => ok({ ...MOCK_FILE_LIST.files[0], id: 'mock-copy-id' }),
      delete: async () => ok({}),
    },
    permissions: {
      create: async () => ok({ id: 'mock-permission-id' }),
    },
    comments: {
      list: async () => ok({ comments: [] }),
      get: async () => ok({ id: 'mock-comment-id', content: 'Mock comment' }),
      create: async () => ok({ id: 'mock-comment-id', content: 'Mock comment' }),
      delete: async () => ok({}),
    },
    replies: {
      create: async () => ok({ id: 'mock-reply-id', content: 'Mock reply' }),
    },
  };
}

export function createMockSheetsClient(): any {
  return {
    spreadsheets: {
      get: async () =>
        ok({
          spreadsheetId: 'mock-sheet-id-456',
          properties: { title: 'Mock Spreadsheet' },
          sheets: [{ properties: { sheetId: 0, title: 'Sheet1', index: 0 } }],
        }),
      values: {
        get: async () => ok(MOCK_SPREADSHEET_VALUES),
        update: async () => ok({ updatedCells: 6, updatedRows: 3 }),
        append: async () => ok({ updates: { updatedCells: 2, updatedRows: 1 } }),
        clear: async () => ok({ clearedRange: 'Sheet1!A1:B3' }),
      },
      batchUpdate: async () => ok({ replies: [] }),
      create: async () =>
        ok({
          spreadsheetId: 'mock-new-sheet-id',
          properties: { title: 'New Spreadsheet' },
        }),
    },
  };
}
```

**Step 2: Modify src/clients.ts to use mocks when MOCK_AUTH=1**

Replace the `initializeGoogleClient` function body. Add import at top and early-return check:

At the top of `src/clients.ts`, add:
```typescript
import { createMockDocsClient, createMockDriveClient, createMockSheetsClient } from './mockClients.js';
```

Replace the `initializeGoogleClient` function to add a mock check at the beginning:

```typescript
export async function initializeGoogleClient() {
  if (googleDocs && googleDrive && googleSheets)
    return { authClient, googleDocs, googleDrive, googleSheets };

  // Mock mode: skip real auth entirely
  if (process.env.MOCK_AUTH === '1') {
    googleDocs = createMockDocsClient();
    googleDrive = createMockDriveClient();
    googleSheets = createMockSheetsClient();
    logger.info('MOCK_AUTH enabled: using mock Google API clients.');
    return { authClient: null, googleDocs, googleDrive, googleSheets };
  }

  if (!authClient) {
    // ... rest of existing code unchanged
```

**Step 3: Build the project**

Run: `npm run build`
Expected: Compiles without errors

**Step 4: Verify mock mode starts the server**

Run: `echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.0.1"}}}' | MOCK_AUTH=1 timeout 3s node dist/index.js 2>/dev/null | head -1`
Expected: JSON response containing `"protocolVersion"` (server starts and responds)

**Step 5: Commit**

```bash
git add src/mockClients.ts src/clients.ts
git commit -m "feat: add MOCK_AUTH=1 mode for testing without Google credentials"
```

---

### Task 4: Create bats common.bash and first protocol test

**Files:**
- Create: `zz-tests_bats/common.bash`
- Create: `zz-tests_bats/protocol.bats`

**Step 1: Create common.bash**

```bash
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
MCP_BIN="${MCP_BIN:-$(dirname "$BATS_TEST_FILE")/../dist/index.js}"

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
    | MOCK_AUTH=1 timeout --preserve-status 5s node "$MCP_BIN" 2>/dev/null \
    | grep -F "\"id\":$response_id" \
    | head -1)

  if [ -z "$response" ]; then
    echo "no response for id $response_id"
    return 1
  fi

  echo "$response"
}

# Standard MCP initialization handshake messages
MCP_INIT='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"bats-test","version":"0.0.1"}}}'
MCP_INITIALIZED='{"jsonrpc":"2.0","method":"notifications/initialized"}'

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

  # Extract the text content from the result
  echo "$response" | jq -r '.result.content[0].text'
}
```

**Step 2: Create protocol.bats**

```bash
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
  assert_output --partial "serverInfo"
}

function tools_list_returns_tools { # @test
  run run_mcp_tools_list
  assert_success
  local tool_count
  tool_count=$(echo "$output" | jq '.result.tools | length')
  # Should have a substantial number of tools (44+ per CLAUDE.md)
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
```

**Step 3: Build and run the protocol tests**

Run: `npm run build && just zz-tests_bats/test-targets protocol.bats`
Expected: All tests pass with TAP output

**Step 4: Commit**

```bash
git add zz-tests_bats/common.bash zz-tests_bats/protocol.bats
git commit -m "feat: add bats protocol tests for MCP lifecycle and tool registration"
```

---

### Task 5: Add docs tool tests

**Files:**
- Create: `zz-tests_bats/docs.bats`

**Step 1: Create docs.bats**

```bash
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
  # JSON format should contain the document structure
  assert_output --partial "textRun"
}

function read_document_returns_markdown_format { # @test
  run run_mcp_tool_call "readDocument" '{"documentId":"mock-doc-id-123","format":"markdown"}'
  assert_success
  assert_output --partial "Hello from the mock document."
}
```

**Step 2: Run docs tests**

Run: `just zz-tests_bats/test-targets docs.bats`
Expected: All tests pass

**Step 3: Commit**

```bash
git add zz-tests_bats/docs.bats
git commit -m "feat: add bats tests for document tool execution"
```

---

### Task 6: Add drive tool tests

**Files:**
- Create: `zz-tests_bats/drive.bats`

**Step 1: Create drive.bats**

```bash
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
```

**Step 2: Run drive tests**

Run: `just zz-tests_bats/test-targets drive.bats`
Expected: All tests pass

**Step 3: Commit**

```bash
git add zz-tests_bats/drive.bats
git commit -m "feat: add bats tests for drive tool execution"
```

---

### Task 7: Add sheets tool tests

**Files:**
- Create: `zz-tests_bats/sheets.bats`

**Step 1: Create sheets.bats**

```bash
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
```

**Step 2: Run sheets tests**

Run: `just zz-tests_bats/test-targets sheets.bats`
Expected: All tests pass

**Step 3: Commit**

```bash
git add zz-tests_bats/sheets.bats
git commit -m "feat: add bats tests for sheets tool execution"
```

---

### Task 8: Run full test suite and verify

**Step 1: Run full suite**

Run: `just test`
Expected: All tests pass (build succeeds, all .bats files pass with TAP output)

**Step 2: Verify parallel execution works**

Run: `just zz-tests_bats/test`
Expected: Tests run in parallel across all .bats files, all pass

---

### Summary of files created/modified

| File | Action |
|------|--------|
| `flake.nix` | Create |
| `flake.lock` | Generated |
| `justfile` | Create |
| `src/mockClients.ts` | Create |
| `src/clients.ts` | Modify (add mock import + early return) |
| `zz-tests_bats/justfile` | Create |
| `zz-tests_bats/common.bash` | Create |
| `zz-tests_bats/protocol.bats` | Create |
| `zz-tests_bats/docs.bats` | Create |
| `zz-tests_bats/drive.bats` | Create |
| `zz-tests_bats/sheets.bats` | Create |
