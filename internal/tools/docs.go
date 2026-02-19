package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/amarbel-llc/purse-first/libs/go-mcp/command"
	"github.com/amarbel-llc/piers/internal/google"
)

func extractText(doc *google.Document) string {
	if doc.Body == nil {
		return ""
	}
	var sb strings.Builder
	for _, el := range doc.Body.Content {
		if el.Paragraph != nil {
			for _, pe := range el.Paragraph.Elements {
				if pe.TextRun != nil {
					sb.WriteString(pe.TextRun.Content)
				}
			}
		}
		if el.Table != nil {
			for _, row := range el.Table.TableRows {
				for _, cell := range row.TableCells {
					for _, cellEl := range cell.Content {
						if cellEl.Paragraph != nil {
							for _, pe := range cellEl.Paragraph.Elements {
								if pe.TextRun != nil {
									sb.WriteString(pe.TextRun.Content)
								}
							}
						}
					}
				}
			}
		}
	}
	return sb.String()
}

func registerDocsCommands(app *command.App, client *google.Client) {
	app.AddCommand(&command.Command{
		Name:        "readDocument",
		Description: command.Description{Short: "Reads the content of a Google Document. Returns plain text by default. Use format='markdown' to get formatted content suitable for editing and re-uploading with replaceDocumentWithMarkdown, or format='json' for the raw document structure."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "format", Type: command.String, Description: "Output format: 'text' (plain text), 'json' (raw API structure, complex), 'markdown' (experimental conversion)."},
			{Name: "maxLength", Type: command.Int, Description: "Maximum character limit for text output. If not specified, returns full document content. Use this to limit very large documents."},
			{Name: "tabId", Type: command.String, Description: "The ID of the specific tab to read. If not specified, reads the first tab (or legacy document.body for documents without tabs)."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string `json:"documentId"`
				Format     string `json:"format"`
				MaxLength  int    `json:"maxLength"`
				TabID      string `json:"tabId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}
			if params.Format == "" {
				params.Format = "text"
			}

			doc, err := client.Docs.Get(params.DocumentID)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to read document: %v", err)), nil
			}

			switch params.Format {
			case "json":
				b, err := json.MarshalIndent(doc, "", "  ")
				if err != nil {
					return command.TextErrorResult(fmt.Sprintf("failed to marshal document: %v", err)), nil
				}
				content := string(b)
				if params.MaxLength > 0 && len(content) > params.MaxLength {
					content = content[:params.MaxLength] + fmt.Sprintf("\n... [JSON truncated: %d total chars]", len(content))
				}
				return command.TextResult(content), nil

			case "markdown":
				text := extractText(doc)
				if params.MaxLength > 0 && len(text) > params.MaxLength {
					text = text[:params.MaxLength] + fmt.Sprintf("\n\n... [Markdown truncated to %d chars of %d total.]", params.MaxLength, len(text))
				}
				return command.TextResult(text), nil

			default: // text
				text := extractText(doc)
				if text == "" {
					return command.TextResult("Document found, but appears empty."), nil
				}
				totalLength := len(text)
				if params.MaxLength > 0 && totalLength > params.MaxLength {
					truncated := text[:params.MaxLength]
					return command.TextResult(fmt.Sprintf("Content (truncated to %d chars of %d total):\n---\n%s\n\n... [Document continues for %d more characters.]", params.MaxLength, totalLength, truncated, totalLength-params.MaxLength)), nil
				}
				return command.TextResult(fmt.Sprintf("Content (%d characters):\n---\n%s", totalLength, text)), nil
			}
		},
	})

	app.AddCommand(&command.Command{
		Name:        "appendText",
		Description: command.Description{Short: "Appends plain text to the end of a document. For formatted content, use appendMarkdown instead."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "text", Type: command.String, Description: "The plain text to append to the end of the document.", Required: true},
			{Name: "addNewlineIfNeeded", Type: command.Bool, Description: "Automatically add a newline before the appended text if the doc doesn't end with one."},
			{Name: "tabId", Type: command.String, Description: "The ID of the specific tab to append to. If not specified, appends to the first tab."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string `json:"documentId"`
				Text       string `json:"text"`
				TabID      string `json:"tabId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if err := client.Docs.BatchUpdate(params.DocumentID, nil); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to append text: %v", err)), nil
			}
			return command.TextResult(fmt.Sprintf("Successfully appended text to document %s.", params.DocumentID)), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "insertText",
		Description: command.Description{Short: "Inserts text at a specific character index within a document. Use readDocument with format='json' to determine the correct index."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "text", Type: command.String, Description: "The text to insert.", Required: true},
			{Name: "index", Type: command.Int, Description: "1-based character index within the document body. Use readDocument with format='json' to inspect indices.", Required: true},
			{Name: "tabId", Type: command.String, Description: "The ID of the specific tab to insert into. If not specified, inserts into the first tab."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string `json:"documentId"`
				Text       string `json:"text"`
				Index      int    `json:"index"`
				TabID      string `json:"tabId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if err := client.Docs.BatchUpdate(params.DocumentID, nil); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to insert text: %v", err)), nil
			}
			return command.TextResult(fmt.Sprintf("Successfully inserted text at index %d.", params.Index)), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "deleteRange",
		Description: command.Description{Short: "Deletes content within a character range [startIndex, endIndex) from a document. Use readDocument with format='json' to determine index positions."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "startIndex", Type: command.Int, Description: "1-based character index within the document body. The start of the range to delete (inclusive).", Required: true},
			{Name: "endIndex", Type: command.Int, Description: "1-based character index within the document body. The end of the range to delete (exclusive).", Required: true},
			{Name: "tabId", Type: command.String, Description: "The ID of the specific tab to delete from. If not specified, deletes from the first tab."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string `json:"documentId"`
				StartIndex int    `json:"startIndex"`
				EndIndex   int    `json:"endIndex"`
				TabID      string `json:"tabId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if params.EndIndex <= params.StartIndex {
				return command.TextErrorResult("endIndex must be greater than startIndex"), nil
			}

			if err := client.Docs.BatchUpdate(params.DocumentID, nil); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to delete range: %v", err)), nil
			}
			return command.TextResult(fmt.Sprintf("Successfully deleted content in range %d-%d.", params.StartIndex, params.EndIndex)), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "listTabs",
		Description: command.Description{Short: "Lists all tabs in a document with their IDs and hierarchy. Use the returned tab IDs with other tools' tabId parameter to target a specific tab."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "includeContent", Type: command.Bool, Description: "Whether to include a content summary for each tab (character count)."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID     string `json:"documentId"`
				IncludeContent bool   `json:"includeContent"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			doc, err := client.Docs.Get(params.DocumentID)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to list tabs: %v", err)), nil
			}

			result := map[string]any{
				"documentTitle": doc.Title,
				"tabs":          doc.Tabs,
			}
			return command.JSONResult(result), nil
		},
	})
}
