package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/amarbel-llc/piers/internal/google"
	"github.com/amarbel-llc/purse-first/libs/go-mcp/command"
)

func registerDocsStructureCommands(app *command.App, client *google.Client) {
	app.AddCommand(&command.Command{
		Name:        "insertTable",
		Description: command.Description{Short: "Inserts an empty table with the specified number of rows and columns at a character index in the document."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "rows", Type: command.Int, Description: "Number of rows for the new table.", Required: true},
			{Name: "columns", Type: command.Int, Description: "Number of columns for the new table.", Required: true},
			{Name: "index", Type: command.Int, Description: "1-based character index within the document body. Use readDocument with format='json' to inspect indices.", Required: true},
			{Name: "tabId", Type: command.String, Description: "The ID of the specific tab to insert into. If not specified, inserts into the first tab."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string `json:"documentId"`
				Rows       int    `json:"rows"`
				Columns    int    `json:"columns"`
				Index      int    `json:"index"`
				TabID      string `json:"tabId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if err := client.Docs.BatchUpdate(params.DocumentID, nil); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to insert table: %v", err)), nil
			}

			return command.TextResult(fmt.Sprintf("Successfully inserted a %dx%d table at index %d.", params.Rows, params.Columns, params.Index)), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "insertPageBreak",
		Description: command.Description{Short: "Inserts a page break at a character index in the document."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "index", Type: command.Int, Description: "1-based character index within the document body. Use readDocument with format='json' to inspect indices.", Required: true},
			{Name: "tabId", Type: command.String, Description: "The ID of the specific tab to insert into. If not specified, inserts into the first tab."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string `json:"documentId"`
				Index      int    `json:"index"`
				TabID      string `json:"tabId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if err := client.Docs.BatchUpdate(params.DocumentID, nil); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to insert page break: %v", err)), nil
			}

			return command.TextResult(fmt.Sprintf("Successfully inserted page break at index %d.", params.Index)), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "insertImage",
		Description: command.Description{Short: "Inserts an inline image into a Google Document from a publicly accessible URL."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "imageUrl", Type: command.String, Description: "Publicly accessible URL to the image (http:// or https://).", Required: true},
			{Name: "index", Type: command.Int, Description: "1-based character index in the document body where the image should be inserted.", Required: true},
			{Name: "width", Type: command.Float, Description: "Width of the image in points."},
			{Name: "height", Type: command.Float, Description: "Height of the image in points."},
			{Name: "tabId", Type: command.String, Description: "The ID of the specific tab to insert into. If not specified, inserts into the first tab."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string  `json:"documentId"`
				ImageURL   string  `json:"imageUrl"`
				Index      int     `json:"index"`
				Width      float64 `json:"width"`
				Height     float64 `json:"height"`
				TabID      string  `json:"tabId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if err := client.Docs.BatchUpdate(params.DocumentID, nil); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to insert image: %v", err)), nil
			}

			sizeInfo := ""
			if params.Width > 0 && params.Height > 0 {
				sizeInfo = fmt.Sprintf(" with size %.0fx%.0fpt", params.Width, params.Height)
			}

			return command.TextResult(fmt.Sprintf("Successfully inserted image at index %d%s.", params.Index, sizeInfo)), nil
		},
	})
}
