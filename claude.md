# Google Docs MCP Server

Go MCP server (purse-first/go-mcp) with 41 tools for Google Docs, Sheets, and Drive.

## Build & Test

```sh
just build          # nix build → result/bin/piers
just build-go       # go build (with gomod2nix)
just test           # go test + bats tests
just test-bats      # bats integration tests only
just test-go        # go unit tests only
just fmt            # go fmt
just deps           # go mod tidy + gomod2nix
```

## Tool Categories

| Category   | Count | Examples                                                                              |
| ---------- | ----- | ------------------------------------------------------------------------------------- |
| Docs       | 5     | `readDocument`, `appendText`, `insertText`, `deleteRange`, `listTabs`                 |
| Markdown   | 2     | `replaceDocumentWithMarkdown`, `appendMarkdown`                                       |
| Formatting | 2     | `applyTextStyle`, `applyParagraphStyle`                                               |
| Structure  | 3     | `insertTable`, `insertPageBreak`, `insertImage`                                       |
| Comments   | 6     | `listComments`, `getComment`, `addComment`, `replyToComment`, `resolveComment`, `deleteComment` |
| Sheets     | 11    | `readSpreadsheet`, `writeSpreadsheet`, `appendRows`, `clearRange`, `createSpreadsheet`, `listSpreadsheets` |
| Drive      | 12    | `listDocuments`, `searchDocuments`, `getDocumentInfo`, `createFolder`, `moveFile`, `copyFile`, `createDocument` |

## Known Limitations

- **Comment anchoring:** Programmatically created comments appear in "All Comments" but aren't visibly anchored to text in the UI
- **Resolved status:** May not persist in Google Docs UI (Drive API limitation)
- **Mock mode:** Set `MOCK_AUTH=1` for testing without real Google credentials

## Parameter Patterns

- **Document ID:** Extract from URL: `docs.google.com/document/d/DOCUMENT_ID/edit`
- **Text targeting:** Use `textToFind` + `matchInstance` OR `startIndex`/`endIndex`
- **Colors:** Hex format `#RRGGBB` or `#RGB`
- **Alignment:** `START`, `END`, `CENTER`, `JUSTIFIED` (not LEFT/RIGHT)
- **Indices:** 1-based, ranges are [start, end)
- **Tabs:** Optional `tabId` parameter (defaults to first tab)

## Source Files

| File                              | Contains                                              |
| --------------------------------- | ----------------------------------------------------- |
| `cmd/piers/main.go`              | Entry point, stdio transport, server setup            |
| `internal/google/client.go`      | Client struct, MOCK_AUTH support                      |
| `internal/google/docs.go`        | Document types, DocsService interface                 |
| `internal/google/drive.go`       | DriveFile/Comment types, DriveService interface       |
| `internal/google/sheets.go`      | Spreadsheet types, SheetsService interface            |
| `internal/google/mock.go`        | Mock implementations for testing                      |
| `internal/tools/registry.go`     | RegisterAll, wires all tool groups                    |
| `internal/tools/docs.go`         | readDocument, appendText, insertText, deleteRange, listTabs |
| `internal/tools/drive.go`        | 12 Drive tools (list, search, create, move, copy, etc.) |
| `internal/tools/sheets.go`       | 11 Sheets tools (read, write, append, format, etc.)   |
| `internal/tools/comments.go`     | 6 comment tools                                       |
| `internal/tools/docs_structure.go` | insertTable, insertPageBreak, insertImage           |
| `internal/tools/docs_formatting.go` | applyTextStyle, applyParagraphStyle               |
| `internal/tools/docs_markdown.go` | replaceDocumentWithMarkdown, appendMarkdown          |
