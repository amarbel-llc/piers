# Go Rewrite Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rewrite the piers Google Docs/Sheets/Drive MCP server in Go using purse-first (go-mcp), keeping identical MCP tool names so the existing bats test suite passes against the new binary.

**Architecture:** Single `internal/tools` package with grouped files (grit pattern), separate `internal/google` package wraps Google API clients with interfaces and handles `MOCK_AUTH=1`. Entry point at `cmd/piers/main.go` follows the grit pattern: `command.NewApp` → `RegisterMCPTools` → `transport.NewStdio` → `server.Run`.

**Tech Stack:** Go, purse-first/go-mcp, google.golang.org/api (docs/v1, drive/v3, sheets/v4), goldmark, Nix (buildGoApplication + gomod2nix), batman/bats

**Reference implementations:**
- grit entry point: `/Users/sfriedenberg/eng/repos/grit/cmd/grit/main.go`
- grit registry: `/Users/sfriedenberg/eng/repos/grit/internal/tools/registry.go`
- grit tool example: `/Users/sfriedenberg/eng/repos/grit/internal/tools/status.go`
- grit flake.nix: `/Users/sfriedenberg/eng/repos/grit/flake.nix`
- grit go.mod: `/Users/sfriedenberg/eng/repos/grit/go.mod`
- grit justfile: `/Users/sfriedenberg/eng/repos/grit/justfile`
- piers design doc: `docs/plans/2026-02-19-go-rewrite-design.md`
- piers TS mock clients: `src/mockClients.ts`
- piers bats common: `zz-tests_bats/common.bash`
- piers bats tests: `zz-tests_bats/protocol.bats`, `docs.bats`, `drive.bats`, `sheets.bats`
- TS tool sources: `src/tools/` (for exact param names, descriptions, response shapes)

---

### Task 1: Update flake.nix for Go build

**Files:**
- Modify: `flake.nix`

**Step 1: Update flake.nix**

Add `go` and `shell` devenv inputs, add `buildGoApplication` package output. Keep batman and nodejs_latest (still needed during transition). Match grit's `flake.nix` pattern exactly.

```nix
{
  description = "Google Docs MCP Server";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/23d72dabcb3b12469f57b37170fcbc1789bd7457";
    nixpkgs-master.url = "github:NixOS/nixpkgs/b28c4999ed71543e71552ccfd0d7e68c581ba7e9";
    utils.url = "https://flakehub.com/f/numtide/flake-utils/0.1.102";
    go.url = "github:friedenberg/eng?dir=devenvs/go";
    shell.url = "github:friedenberg/eng?dir=devenvs/shell";
    batman.url = "github:amarbel-llc/batman";
  };

  outputs =
    {
      self,
      nixpkgs,
      nixpkgs-master,
      utils,
      go,
      shell,
      batman,
    }:
    utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [
            go.overlays.default
          ];
        };

        version = "1.0.0";

        piers = pkgs.buildGoApplication {
          pname = "piers";
          inherit version;
          src = ./.;
          modules = ./gomod2nix.toml;
          subPackages = [ "cmd/piers" ];

          meta = with pkgs.lib; {
            description = "MCP server for Google Docs, Sheets, and Drive";
            license = licenses.isc;
          };
        };
      in
      {
        packages = {
          default = piers;
          inherit piers;
        };

        devShells.default = pkgs.mkShell {
          packages =
            (with pkgs; [
              just
              jq
              nodejs_latest
            ])
            ++ [
              batman.packages.${system}.bats
              batman.packages.${system}.bats-libs
            ];

          inputsFrom = [
            go.devShells.${system}.default
            shell.devShells.${system}.default
          ];
        };

        apps.default = {
          type = "app";
          program = "${piers}/bin/piers";
        };
      }
    );
}
```

**Step 2: Lock flake inputs**

Run: `git add flake.nix && nix flake lock`
Expected: lock file updates with new go/shell inputs

**Step 3: Commit**

```
feat: update flake.nix for Go build with purse-first
```

---

### Task 2: Initialize Go module and scaffold entry point

**Files:**
- Create: `go.mod`
- Create: `cmd/piers/main.go`
- Create: `internal/google/client.go`
- Create: `internal/tools/registry.go`

**Step 1: Create go.mod**

```
module github.com/amarbel-llc/piers

go 1.25.6

require github.com/amarbel-llc/purse-first/libs/go-mcp v0.0.1
```

**Step 2: Create cmd/piers/main.go**

Match grit's `cmd/grit/main.go` pattern. Stdio transport only.

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/amarbel-llc/purse-first/libs/go-mcp/server"
	"github.com/amarbel-llc/purse-first/libs/go-mcp/transport"
	"github.com/amarbel-llc/piers/internal/google"
	"github.com/amarbel-llc/piers/internal/tools"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client, err := google.NewClient(ctx)
	if err != nil {
		log.Fatalf("creating google client: %v", err)
	}

	app := tools.RegisterAll(client)

	registry := server.NewToolRegistry()
	app.RegisterMCPTools(registry)

	t := transport.NewStdio(os.Stdin, os.Stdout)

	srv, err := server.New(t, server.Options{
		ServerName:    app.Name,
		ServerVersion: app.Version,
		Tools:         registry,
	})
	if err != nil {
		log.Fatalf("creating server: %v", err)
	}

	if err := srv.Run(ctx); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
```

**Step 3: Create internal/google/client.go**

Stub with mock-only support. Real OAuth2 comes later. Note: `DocsService`, `DriveService`, `SheetsService` are interfaces defined in Task 3. For this step, define them as empty interfaces so the file compiles — they'll be replaced in the next task.

```go
package google

import (
	"context"
	"fmt"
	"os"
)

type Client struct {
	Docs   DocsService
	Drive  DriveService
	Sheets SheetsService
}

func NewClient(ctx context.Context) (*Client, error) {
	if os.Getenv("MOCK_AUTH") == "1" {
		return newMockClient(), nil
	}
	return nil, fmt.Errorf("real OAuth2 not yet implemented; set MOCK_AUTH=1 for testing")
}
```

**Step 4: Create internal/tools/registry.go**

Empty registry — no tools yet.

```go
package tools

import (
	"github.com/amarbel-llc/purse-first/libs/go-mcp/command"
	"github.com/amarbel-llc/piers/internal/google"
)

func RegisterAll(client *google.Client) *command.App {
	app := command.NewApp("piers", "MCP server for Google Docs, Sheets, and Drive")
	app.Version = "1.0.0"
	return app
}
```

**Step 5: Run go mod tidy and gomod2nix**

Run: `git add go.mod cmd/ internal/ && nix develop --command bash -c "go mod tidy && gomod2nix"`
Expected: `go.sum` and `gomod2nix.toml` created

**Step 6: Verify it builds**

Run: `nix develop --command go build -o piers ./cmd/piers`
Expected: compiles with no errors

**Step 7: Commit**

```
feat: scaffold Go entry point with purse-first
```

---

### Task 3: Google API service interfaces and mock implementations

**Files:**
- Create: `internal/google/docs.go`
- Create: `internal/google/drive.go`
- Create: `internal/google/sheets.go`
- Create: `internal/google/mock.go`
- Modify: `internal/google/client.go` (update field types to interfaces)

**Step 1: Create internal/google/docs.go**

Define types matching the Google Docs API structures our tools consume, and a service interface.

```go
package google

type Document struct {
	DocumentID string        `json:"documentId"`
	Title      string        `json:"title"`
	Body       *DocumentBody `json:"body,omitempty"`
	Tabs       []any         `json:"tabs,omitempty"`
}

type DocumentBody struct {
	Content []ContentElement `json:"content,omitempty"`
}

type ContentElement struct {
	Paragraph *Paragraph `json:"paragraph,omitempty"`
	Table     *Table     `json:"table,omitempty"`
}

type Paragraph struct {
	Elements []ParagraphElement `json:"elements,omitempty"`
}

type ParagraphElement struct {
	TextRun *TextRun `json:"textRun,omitempty"`
}

type TextRun struct {
	Content string `json:"content"`
}

type Table struct {
	TableRows []TableRow `json:"tableRows,omitempty"`
}

type TableRow struct {
	TableCells []TableCell `json:"tableCells,omitempty"`
}

type TableCell struct {
	Content []ContentElement `json:"content,omitempty"`
}

type DocsService interface {
	Get(documentID string) (*Document, error)
	BatchUpdate(documentID string, requests any) error
	Create(title string) (*Document, error)
}
```

**Step 2: Create internal/google/drive.go**

```go
package google

type DriveFile struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	MimeType     string      `json:"mimeType,omitempty"`
	ModifiedTime string      `json:"modifiedTime,omitempty"`
	CreatedTime  string      `json:"createdTime,omitempty"`
	WebViewLink  string      `json:"webViewLink,omitempty"`
	Owners       []FileOwner `json:"owners,omitempty"`
	Parents      []string    `json:"parents,omitempty"`
}

type FileOwner struct {
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

type Comment struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type CommentReply struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type DriveService interface {
	ListFiles(query string, pageSize int, orderBy string) ([]DriveFile, error)
	GetFile(fileID string) (*DriveFile, error)
	CreateFile(name string, mimeType string, parentID string) (*DriveFile, error)
	UpdateFile(fileID string, name string, addParents string, removeParents string) (*DriveFile, error)
	CopyFile(fileID string, name string) (*DriveFile, error)
	DeleteFile(fileID string) error
	ListComments(fileID string) ([]Comment, error)
	GetComment(fileID string, commentID string) (*Comment, error)
	CreateComment(fileID string, content string, quotedContent string) (*Comment, error)
	DeleteComment(fileID string, commentID string) error
	ReplyToComment(fileID string, commentID string, content string) (*CommentReply, error)
	ResolveComment(fileID string, commentID string) error
}
```

**Step 3: Create internal/google/sheets.go**

```go
package google

type ValueRange struct {
	Range  string  `json:"range"`
	Values [][]any `json:"values,omitempty"`
}

type UpdateResult struct {
	UpdatedCells int `json:"updatedCells"`
	UpdatedRows  int `json:"updatedRows"`
}

type Spreadsheet struct {
	SpreadsheetID string             `json:"spreadsheetId"`
	Properties    SpreadsheetProps   `json:"properties"`
	Sheets        []SpreadsheetSheet `json:"sheets,omitempty"`
}

type SpreadsheetProps struct {
	Title string `json:"title"`
}

type SpreadsheetSheet struct {
	Properties SheetProperties `json:"properties"`
}

type SheetProperties struct {
	SheetID int    `json:"sheetId"`
	Title   string `json:"title"`
	Index   int    `json:"index"`
}

type SheetsService interface {
	GetValues(spreadsheetID string, rangeStr string) (*ValueRange, error)
	UpdateValues(spreadsheetID string, rangeStr string, values [][]any) (*UpdateResult, error)
	AppendValues(spreadsheetID string, rangeStr string, values [][]any) (*UpdateResult, error)
	ClearValues(spreadsheetID string, rangeStr string) (string, error)
	GetSpreadsheet(spreadsheetID string) (*Spreadsheet, error)
	CreateSpreadsheet(title string) (*Spreadsheet, error)
	AddSheet(spreadsheetID string, title string) error
	BatchUpdate(spreadsheetID string, requests any) error
}
```

**Step 4: Create internal/google/mock.go**

Mock implementations returning hardcoded data matching `src/mockClients.ts` exactly.

```go
package google

func newMockClient() *Client {
	return &Client{
		Docs:   &mockDocsService{},
		Drive:  &mockDriveService{},
		Sheets: &mockSheetsService{},
	}
}

type mockDocsService struct{}

func (m *mockDocsService) Get(documentID string) (*Document, error) {
	return &Document{
		DocumentID: "mock-doc-id-123",
		Title:      "Mock Document",
		Body: &DocumentBody{
			Content: []ContentElement{
				{Paragraph: &Paragraph{
					Elements: []ParagraphElement{
						{TextRun: &TextRun{Content: "Hello from the mock document.\n"}},
					},
				}},
			},
		},
		Tabs: []any{},
	}, nil
}

func (m *mockDocsService) BatchUpdate(documentID string, requests any) error { return nil }

func (m *mockDocsService) Create(title string) (*Document, error) {
	doc, _ := m.Get("")
	return doc, nil
}

type mockDriveService struct{}

var mockFiles = []DriveFile{
	{
		ID: "mock-doc-id-123", Name: "Mock Document",
		MimeType: "application/vnd.google-apps.document",
		ModifiedTime: "2025-01-15T10:30:00.000Z", CreatedTime: "2025-01-01T08:00:00.000Z",
		WebViewLink: "https://docs.google.com/document/d/mock-doc-id-123/edit",
		Owners: []FileOwner{{DisplayName: "Test User", EmailAddress: "test@example.com"}},
	},
	{
		ID: "mock-sheet-id-456", Name: "Mock Spreadsheet",
		MimeType: "application/vnd.google-apps.spreadsheet",
		ModifiedTime: "2025-01-14T09:00:00.000Z", CreatedTime: "2025-01-02T08:00:00.000Z",
		WebViewLink: "https://docs.google.com/spreadsheets/d/mock-sheet-id-456/edit",
		Owners: []FileOwner{{DisplayName: "Test User", EmailAddress: "test@example.com"}},
	},
}

func (m *mockDriveService) ListFiles(q string, ps int, ob string) ([]DriveFile, error) { return mockFiles, nil }
func (m *mockDriveService) GetFile(id string) (*DriveFile, error) { return &mockFiles[0], nil }
func (m *mockDriveService) CreateFile(n, mt, p string) (*DriveFile, error) { return &mockFiles[0], nil }
func (m *mockDriveService) UpdateFile(id, n, ap, rp string) (*DriveFile, error) { return &mockFiles[0], nil }
func (m *mockDriveService) CopyFile(id, n string) (*DriveFile, error) { f := mockFiles[0]; f.ID = "mock-copy-id"; return &f, nil }
func (m *mockDriveService) DeleteFile(id string) error { return nil }
func (m *mockDriveService) ListComments(id string) ([]Comment, error) { return []Comment{}, nil }
func (m *mockDriveService) GetComment(fid, cid string) (*Comment, error) { return &Comment{ID: "mock-comment-id", Content: "Mock comment"}, nil }
func (m *mockDriveService) CreateComment(fid, c, q string) (*Comment, error) { return &Comment{ID: "mock-comment-id", Content: "Mock comment"}, nil }
func (m *mockDriveService) DeleteComment(fid, cid string) error { return nil }
func (m *mockDriveService) ReplyToComment(fid, cid, c string) (*CommentReply, error) { return &CommentReply{ID: "mock-reply-id", Content: "Mock reply"}, nil }
func (m *mockDriveService) ResolveComment(fid, cid string) error { return nil }

type mockSheetsService struct{}

func (m *mockSheetsService) GetValues(sid, r string) (*ValueRange, error) {
	return &ValueRange{Range: "Sheet1!A1:B3", Values: [][]any{{"Name", "Score"}, {"Alice", "95"}, {"Bob", "87"}}}, nil
}
func (m *mockSheetsService) UpdateValues(sid, r string, v [][]any) (*UpdateResult, error) { return &UpdateResult{UpdatedCells: 6, UpdatedRows: 3}, nil }
func (m *mockSheetsService) AppendValues(sid, r string, v [][]any) (*UpdateResult, error) { return &UpdateResult{UpdatedCells: 2, UpdatedRows: 1}, nil }
func (m *mockSheetsService) ClearValues(sid, r string) (string, error) { return "Sheet1!A1:B3", nil }
func (m *mockSheetsService) GetSpreadsheet(sid string) (*Spreadsheet, error) {
	return &Spreadsheet{SpreadsheetID: "mock-sheet-id-456", Properties: SpreadsheetProps{Title: "Mock Spreadsheet"}, Sheets: []SpreadsheetSheet{{Properties: SheetProperties{SheetID: 0, Title: "Sheet1", Index: 0}}}}, nil
}
func (m *mockSheetsService) CreateSpreadsheet(title string) (*Spreadsheet, error) { return &Spreadsheet{SpreadsheetID: "mock-new-sheet-id", Properties: SpreadsheetProps{Title: title}}, nil }
func (m *mockSheetsService) AddSheet(sid, title string) error { return nil }
func (m *mockSheetsService) BatchUpdate(sid string, req any) error { return nil }
```

**Step 5: Verify build**

Run: `nix develop --command go build ./...`

**Step 6: Commit**

```
feat: add Google API service interfaces and mock implementations
```

---

### Task 4: Core docs tools (5 tools)

**Files:**
- Create: `internal/tools/docs.go`
- Modify: `internal/tools/registry.go`

5 tools: `readDocument`, `appendText`, `insertText`, `deleteRange`, `listTabs`

Bats assertions (`docs.bats`):
- `readDocument` text: output contains `"Hello from the mock document."`
- `readDocument` json: output contains `"textRun"`
- `readDocument` markdown: output contains `"Hello from the mock document."`

**Step 1: Create internal/tools/docs.go**

Key implementation details:
- `readDocument`: 3 formats. Text wraps as `Content (N characters):\n---\n{text}`. JSON returns `json.MarshalIndent(doc)`. Markdown returns plain text for now.
- `appendText`, `insertText`, `deleteRange`: call `client.Docs.BatchUpdate()`, return success message.
- `listTabs`: call `client.Docs.Get()`, return `{"tabs": [...]}`.
- Helper `extractText(doc)` walks `Body.Content` paragraphs and tables extracting `TextRun.Content`.

Reference TS files for exact param names: `src/tools/docs/readGoogleDoc.ts`, `appendToGoogleDoc.ts`, `insertText.ts`, `deleteRange.ts`, `listDocumentTabs.ts`

**Step 2: Wire up registerDocsCommands in registry.go**

Add `registerDocsCommands(app, client)` to `RegisterAll`.

**Step 3: Verify build**

Run: `nix develop --command go build ./...`

**Step 4: Commit**

```
feat: add core docs tools (readDocument, appendText, insertText, deleteRange, listTabs)
```

---

### Task 5: Drive tools (12 tools)

**Files:**
- Create: `internal/tools/drive.go`
- Modify: `internal/tools/registry.go`

12 tools: `listDocuments`, `searchDocuments`, `getDocumentInfo`, `createFolder`, `listFolderContents`, `getFolderInfo`, `moveFile`, `copyFile`, `renameFile`, `deleteFile`, `createDocument`, `createDocumentFromTemplate`

Bats assertions (`drive.bats`):
- `.documents | length >= 1`
- `.documents[0].name == "Mock Document"`
- `.documents[0].url` contains `"docs.google.com"`

Critical output shape: `listDocuments` must return `{"documents": [{"id": "...", "name": "...", "modifiedTime": "...", "owner": "...", "url": "..."}]}`. The `url` field maps from `DriveFile.WebViewLink`, `owner` from `DriveFile.Owners[0].DisplayName`.

Reference TS files: `src/tools/drive/*.ts`

**Step 1: Create internal/tools/drive.go**

**Step 2: Wire up registerDriveCommands in registry.go**

**Step 3: Verify build**

**Step 4: Commit**

```
feat: add drive tools (12 tools)
```

---

### Task 6: Sheets tools (11 tools)

**Files:**
- Create: `internal/tools/sheets.go`
- Modify: `internal/tools/registry.go`

11 tools: `readSpreadsheet`, `writeSpreadsheet`, `appendRows`, `clearRange`, `getSpreadsheetInfo`, `addSheet`, `createSpreadsheet`, `listSpreadsheets`, `formatCells`, `freezeRowsAndColumns`, `setDropdownValidation`

Bats assertions (`sheets.bats`):
- `.values | length == 3`
- `.values[0][0] == "Name"`
- `.values[1][0] == "Alice"` and `.values[1][1] == "95"`

Critical output shape: `readSpreadsheet` must return `{"range": "...", "values": [[...], ...]}`.

Note: `listSpreadsheets` uses `client.Drive.ListFiles()` with spreadsheet mimeType filter, not the Sheets API.

Reference TS files: `src/tools/sheets/*.ts`

**Step 1: Create internal/tools/sheets.go**

**Step 2: Wire up registerSheetsCommands in registry.go**

**Step 3: Verify build**

**Step 4: Commit**

```
feat: add sheets tools (11 tools)
```

---

### Task 7: Comment tools (6 tools)

**Files:**
- Create: `internal/tools/comments.go`
- Modify: `internal/tools/registry.go`

6 tools: `listComments`, `getComment`, `addComment`, `replyToComment`, `resolveComment`, `deleteComment`

All use `client.Drive` comment/reply methods. Reference: `src/tools/docs/comments/*.ts`

**Step 1: Create internal/tools/comments.go**

**Step 2: Wire up registerCommentCommands in registry.go**

**Step 3: Verify build**

**Step 4: Commit**

```
feat: add comment tools (6 tools)
```

---

### Task 8: Docs structure tools (3 tools)

**Files:**
- Create: `internal/tools/docs_structure.go`
- Modify: `internal/tools/registry.go`

3 tools: `insertTable`, `insertPageBreak`, `insertImage`

All use `client.Docs.BatchUpdate()`. Reference: `src/tools/docs/insertTable.ts`, `insertPageBreak.ts`, `insertImage.ts`

**Step 1: Create internal/tools/docs_structure.go**

**Step 2: Wire up registerDocsStructureCommands in registry.go**

**Step 3: Verify build**

**Step 4: Commit**

```
feat: add docs structure tools (insertTable, insertPageBreak, insertImage)
```

---

### Task 9: Docs formatting tools (2 tools)

**Files:**
- Create: `internal/tools/docs_formatting.go`
- Modify: `internal/tools/registry.go`

2 tools: `applyTextStyle`, `applyParagraphStyle`

Both use `client.Docs.BatchUpdate()`. Reference: `src/tools/docs/formatting/applyTextStyle.ts`, `applyParagraphStyle.ts`

**Step 1: Create internal/tools/docs_formatting.go**

**Step 2: Wire up registerDocsFormattingCommands in registry.go**

**Step 3: Verify build**

**Step 4: Commit**

```
feat: add docs formatting tools (applyTextStyle, applyParagraphStyle)
```

---

### Task 10: Markdown tools (2 tools)

**Files:**
- Create: `internal/tools/docs_markdown.go`
- Modify: `internal/tools/registry.go`

2 tools: `replaceDocumentWithMarkdown`, `appendMarkdown`

For initial implementation, accept markdown content and call `client.Docs.BatchUpdate()`. Full goldmark-based markdown-to-Google-Docs conversion (headings, bold, italic, strikethrough, links, lists, code blocks) is complex — start with placeholder that inserts plain text, then iterate.

Reference: `src/tools/utils/replaceDocumentWithMarkdown.ts`, `appendMarkdownToGoogleDoc.ts`, `src/markdownToGoogleDocs.ts`

**Step 1: Add goldmark dependency**

Run: `nix develop --command go get github.com/yuin/goldmark`

**Step 2: Create internal/tools/docs_markdown.go**

**Step 3: Wire up registerDocsMarkdownCommands in registry.go**

**Step 4: Verify build**

Run: `nix develop --command bash -c "go mod tidy && go build ./..."`

**Step 5: Commit**

```
feat: add markdown tools (replaceDocumentWithMarkdown, appendMarkdown)
```

---

### Task 11: Build with Nix and update bats tests

**Files:**
- Modify: `zz-tests_bats/common.bash`
- Modify: `justfile`

**Step 1: Generate gomod2nix.toml and build**

Run: `nix develop --command gomod2nix && nix build`
Expected: `./result/bin/piers` exists

**Step 2: Update common.bash to use Go binary**

Change `MCP_BIN` default and remove `node` invocation:

```bash
# Before:
MCP_BIN="${MCP_BIN:-$(dirname "$BATS_TEST_FILE")/../dist/index.js}"
# After:
MCP_BIN="${MCP_BIN:-$(dirname "$BATS_TEST_FILE")/../result/bin/piers}"
```

```bash
# Before:
| MOCK_AUTH=1 timeout --preserve-status 5s node "$MCP_BIN" 2>/dev/null \
# After:
| MOCK_AUTH=1 timeout --preserve-status 5s "$MCP_BIN" 2>/dev/null \
```

**Step 3: Update justfile**

```makefile
default:
    @just --list

build:
    nix build

build-gomod2nix:
    nix develop --command gomod2nix

build-go: build-gomod2nix
    nix develop --command go build -o piers ./cmd/piers

test-go:
    nix develop --command go test ./...

test-bats: build
    just zz-tests_bats/test

test: test-go test-bats

fmt:
    nix develop --command go fmt ./...

deps:
    nix develop --command go mod tidy
    nix develop --command gomod2nix

clean:
    rm -f piers
    rm -rf result
```

**Step 4: Run bats tests**

Run: `just test-bats`
Expected: all 17 tests pass (8 protocol + 3 docs + 3 drive + 3 sheets)

**Step 5: Commit**

```
feat: update bats tests and justfile for Go binary
```

---

### Task 12: Debug and fix bats test failures

This task exists because bats tests may fail due to subtle differences between Go and TypeScript implementations. Common issues:

- Tool count: `protocol.bats` checks `tool_count >= 30`. We have 41 tools — should pass.
- JSON output shape differences (field names, nesting)
- Mock data format mismatches

**Step 1: Run each test file independently**

```bash
just zz-tests_bats/test-targets protocol.bats
just zz-tests_bats/test-targets docs.bats
just zz-tests_bats/test-targets drive.bats
just zz-tests_bats/test-targets sheets.bats
```

**Step 2: Fix any issues in tool implementations or mock data**

**Step 3: Run full test suite**

Run: `just test`
Expected: all Go tests and bats tests pass

**Step 4: Commit**

```
fix: address bats test failures in Go implementation
```

---

### Task 13: Final cleanup

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Run go vet and fmt**

Run: `nix develop --command bash -c "go vet ./... && go fmt ./..."`

**Step 2: Update CLAUDE.md**

Update project description, build commands, source file table, and tech stack to reflect Go implementation.

**Step 3: Final build verification**

Run: `nix build && just test`
Expected: clean build, all tests pass

**Step 4: Commit**

```
docs: update CLAUDE.md for Go implementation
```
