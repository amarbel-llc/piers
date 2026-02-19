# Go Rewrite Design: piers MCP Server

## Summary

Rewrite the piers Google Docs/Sheets/Drive MCP server from TypeScript (FastMCP) to Go using the purse-first framework (go-mcp). The Go binary replaces `dist/index.js` as the MCP server, maintaining identical tool names so the existing bats test suite validates the new implementation without changes.

## Decision Record

- **Framework:** purse-first (go-mcp) — same framework as grit
- **Auth library:** `google.golang.org/api` (official Go client libraries)
- **Architecture:** Approach C — single `internal/tools` package with grouped files, separate `internal/google` package for API wrappers
- **Repo:** Same repo (piers), Go code alongside existing TypeScript
- **Tool scope:** All 41 registered tools (see catalog below)

## Architecture

```
cmd/piers/main.go                  # Entry point: command.NewApp + server + transport
internal/google/client.go          # OAuth2 + MOCK_AUTH=1 support
internal/google/docs.go            # docs_v1.Service wrapper
internal/google/drive.go           # drive_v3.Service wrapper
internal/google/sheets.go          # sheets_v4.Service wrapper
internal/tools/registry.go         # RegisterAll(app, client)
internal/tools/docs.go             # 5 tools: readDocument, appendText, insertText, deleteRange, listTabs
internal/tools/docs_markdown.go    # 2 tools: replaceDocumentWithMarkdown, appendMarkdown
internal/tools/docs_structure.go   # 3 tools: insertTable, insertPageBreak, insertImage
internal/tools/docs_formatting.go  # 2 tools: applyTextStyle, applyParagraphStyle
internal/tools/comments.go         # 6 tools: listComments, getComment, addComment, replyToComment, resolveComment, deleteComment
internal/tools/drive.go            # 12 tools: listDocuments, searchDocuments, getDocumentInfo, createFolder, listFolderContents, getFolderInfo, moveFile, copyFile, renameFile, deleteFile, createDocument, createDocumentFromTemplate
internal/tools/sheets.go           # 11 tools: readSpreadsheet, writeSpreadsheet, appendRows, clearRange, getSpreadsheetInfo, addSheet, createSpreadsheet, listSpreadsheets, formatCells, freezeRowsAndColumns, setDropdownValidation
go.mod
go.sum
```

### Entry Point Pattern (from grit)

```go
// cmd/piers/main.go
func main() {
    app := tools.RegisterAll(googleClient)
    registry := server.NewToolRegistry()
    app.RegisterMCPTools(registry)

    t := transport.NewStdio()
    srv, err := server.New(t, server.Options{
        ServerName:    app.Name,
        ServerVersion: app.Version,
        Tools:         registry,
    })
    srv.Serve()
}
```

### Google API Client Layer

`internal/google/client.go` creates a `Client` struct holding three service wrappers:

```go
type Client struct {
    Docs   *DocsService
    Drive  *DriveService
    Sheets *SheetsService
}

func NewClient(ctx context.Context) (*Client, error) {
    if os.Getenv("MOCK_AUTH") == "1" {
        return newMockClient(), nil
    }
    // OAuth2 flow using google.golang.org/api/option
}
```

Each wrapper (e.g., `DocsService`) encapsulates the `google.golang.org/api/docs/v1` service and exposes typed methods that tools call. Mock implementations return hardcoded data matching the existing TypeScript mocks.

### Tool Registration Pattern

```go
// internal/tools/registry.go
func RegisterAll(client *google.Client) *command.App {
    app := command.NewApp("piers", "MCP server for Google Docs, Sheets, and Drive")
    app.Version = "1.0.0"

    registerDocsCommands(app, client)
    registerDocsMarkdownCommands(app, client)
    registerDocsStructureCommands(app, client)
    registerDocsFormattingCommands(app, client)
    registerCommentCommands(app, client)
    registerDriveCommands(app, client)
    registerSheetsCommands(app, client)

    return app
}
```

Each `register*Commands` function adds commands via `app.AddCommand()`:

```go
// internal/tools/docs.go
func registerDocsCommands(app *command.App, client *google.Client) {
    app.AddCommand(&command.Command{
        Name:        "readDocument",
        Description: "Read the content of a Google Document",
        Params: []command.Param{
            {Name: "documentId", Description: "The document ID", Required: true},
            {Name: "format", Description: "Output format: text, json, or markdown", Required: false},
            {Name: "maxLength", Description: "Maximum character length", Required: false},
            {Name: "tabId", Description: "Target tab ID", Required: false},
        },
        Run: func(ctx context.Context, args json.RawMessage, prompter command.Prompter) (*command.Result, error) {
            // Parse args, call client.Docs.Get(), format response
        },
    })
    // ... more tools
}
```

## Tool Catalog (41 tools, exact MCP names preserved)

### Docs (5)
| Tool Name | Source File |
|-----------|------------|
| `readDocument` | `docs.go` |
| `appendText` | `docs.go` |
| `insertText` | `docs.go` |
| `deleteRange` | `docs.go` |
| `listTabs` | `docs.go` |

### Docs — Structure (3)
| Tool Name | Source File |
|-----------|------------|
| `insertTable` | `docs_structure.go` |
| `insertPageBreak` | `docs_structure.go` |
| `insertImage` | `docs_structure.go` |

### Docs — Formatting (2)
| Tool Name | Source File |
|-----------|------------|
| `applyTextStyle` | `docs_formatting.go` |
| `applyParagraphStyle` | `docs_formatting.go` |

### Docs — Comments (6)
| Tool Name | Source File |
|-----------|------------|
| `listComments` | `comments.go` |
| `getComment` | `comments.go` |
| `addComment` | `comments.go` |
| `replyToComment` | `comments.go` |
| `resolveComment` | `comments.go` |
| `deleteComment` | `comments.go` |

### Docs — Markdown (2)
| Tool Name | Source File |
|-----------|------------|
| `replaceDocumentWithMarkdown` | `docs_markdown.go` |
| `appendMarkdown` | `docs_markdown.go` |

### Drive (12)
| Tool Name | Source File |
|-----------|------------|
| `listDocuments` | `drive.go` |
| `searchDocuments` | `drive.go` |
| `getDocumentInfo` | `drive.go` |
| `createFolder` | `drive.go` |
| `listFolderContents` | `drive.go` |
| `getFolderInfo` | `drive.go` |
| `moveFile` | `drive.go` |
| `copyFile` | `drive.go` |
| `renameFile` | `drive.go` |
| `deleteFile` | `drive.go` |
| `createDocument` | `drive.go` |
| `createDocumentFromTemplate` | `drive.go` |

### Sheets (11)
| Tool Name | Source File |
|-----------|------------|
| `readSpreadsheet` | `sheets.go` |
| `writeSpreadsheet` | `sheets.go` |
| `appendRows` | `sheets.go` |
| `clearRange` | `sheets.go` |
| `getSpreadsheetInfo` | `sheets.go` |
| `addSheet` | `sheets.go` |
| `createSpreadsheet` | `sheets.go` |
| `listSpreadsheets` | `sheets.go` |
| `formatCells` | `sheets.go` |
| `freezeRowsAndColumns` | `sheets.go` |
| `setDropdownValidation` | `sheets.go` |

## Nix Build

Update `flake.nix` to build the Go binary alongside the existing Node.js devShell:

- Add `go` devenv input or use `pkgs.buildGoModule`
- Binary output: `piers` (replaces `node dist/index.js`)
- Keep batman/bats inputs for testing
- Update `justfile` to build Go binary and run bats against it
- Update `common.bash` `run_mcp()` to invoke the Go binary instead of `node dist/index.js`

## Mock Auth

`MOCK_AUTH=1` is handled in `internal/google/client.go`. Mock wrappers return the same hardcoded data as the TypeScript mocks:

- **Docs:** Document with body text "Hello from the mock document."
- **Drive:** File list with 2 files (mock-doc-id-1 "Mock Document", mock-doc-id-2 "Another Mock Doc")
- **Sheets:** Spreadsheet with headers ["Name", "Score"] and rows [["Alice", "95"], ["Bob", "87"]]

This ensures the existing bats tests pass unchanged against the Go binary.

## Testing Strategy

1. **Bats tests (existing):** Validate MCP protocol, tool registration, and tool behavior over stdio. These are the primary acceptance criteria — if all bats tests pass against the Go binary, the rewrite is functionally correct.
2. **Go unit tests:** Test `internal/google` wrappers and individual tool logic in isolation.
3. **`nix flake check`:** Build verification.

## Markdown Conversion

The TypeScript implementation uses `markdown-it` for parsing. The Go implementation needs equivalent markdown parsing for `replaceDocumentWithMarkdown` and `appendMarkdown`. Options:

- `github.com/yuin/goldmark` — most popular, extensible, CommonMark-compliant
- `github.com/gomarkdown/markdown` — also mature, supports many extensions

Recommend `goldmark` for CommonMark compliance matching markdown-it's default behavior.

## Shared Drives

All Drive operations must set `SupportsAllDrives(true)` and (for list operations) `IncludeItemsFromAllDrives(true)`, matching the TypeScript implementation.
