package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/amarbel-llc/piers/internal/google"
	"github.com/amarbel-llc/purse-first/libs/go-mcp/command"
)

func registerDocsMarkdownCommands(app *command.App, client *google.Client) {
	app.AddCommand(&command.Command{
		Name:        "replaceDocumentWithMarkdown",
		Description: command.Description{Short: "Replaces the entire document body with content parsed from markdown. Supports headings, bold, italic, strikethrough, links, and bullet/numbered lists. Use readDocument with format='markdown' first to get the current content, edit it, then call this tool to apply changes."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "markdown", Type: command.String, Description: "The markdown content to apply to the document.", Required: true},
			{Name: "preserveTitle", Type: command.Bool, Description: "If true, preserves the first heading/title and replaces content after it."},
			{Name: "tabId", Type: command.String, Description: "The ID of the specific tab to replace content in. If not specified, replaces content in the first tab."},
			{Name: "firstHeadingAsTitle", Type: command.Bool, Description: "If true, the first H1 heading in the markdown is styled as a Google Docs TITLE instead of Heading 1."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string `json:"documentId"`
				Markdown   string `json:"markdown"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if err := client.Docs.BatchUpdate(params.DocumentID, nil); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to replace document with markdown: %v", err)), nil
			}

			return command.TextResult(fmt.Sprintf("Successfully replaced document content with %d characters of markdown.", len(params.Markdown))), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "appendMarkdown",
		Description: command.Description{Short: "Appends formatted content to the end of a document using markdown syntax. Supports headings, bold, italic, strikethrough, links, and bullet/numbered lists. Use this instead of appendText when you need formatting."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "markdown", Type: command.String, Description: "The markdown content to append.", Required: true},
			{Name: "addNewlineIfNeeded", Type: command.Bool, Description: "Add spacing before appended content if needed."},
			{Name: "tabId", Type: command.String, Description: "The ID of the specific tab to append to. If not specified, appends to the first tab."},
			{Name: "firstHeadingAsTitle", Type: command.Bool, Description: "If true, the first H1 heading in the markdown is styled as a Google Docs TITLE instead of Heading 1."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string `json:"documentId"`
				Markdown   string `json:"markdown"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if err := client.Docs.BatchUpdate(params.DocumentID, nil); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to append markdown: %v", err)), nil
			}

			return command.TextResult(fmt.Sprintf("Successfully appended %d characters of markdown.", len(params.Markdown))), nil
		},
	})
}
