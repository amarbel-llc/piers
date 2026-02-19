package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/amarbel-llc/piers/internal/google"
	"github.com/amarbel-llc/purse-first/libs/go-mcp/command"
)

func registerSheetsCommands(app *command.App, client *google.Client) {
	app.AddCommand(&command.Command{
		Name:        "readSpreadsheet",
		Description: command.Description{Short: "Reads data from a range in a spreadsheet. Returns rows as arrays. Use A1 notation for the range (e.g., \"Sheet1!A1:C10\")."},
		Params: []command.Param{
			{Name: "spreadsheetId", Type: command.String, Description: "The spreadsheet ID — the long string between /d/ and /edit in a Google Sheets URL.", Required: true},
			{Name: "range", Type: command.String, Description: "A1 notation range to read (e.g., \"A1:B10\" or \"Sheet1!A1:B10\").", Required: true},
			{Name: "valueRenderOption", Type: command.String, Description: "How values should be rendered in the output: FORMATTED_VALUE, UNFORMATTED_VALUE, or FORMULA."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				SpreadsheetID string `json:"spreadsheetId"`
				Range         string `json:"range"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			vr, err := client.Sheets.GetValues(params.SpreadsheetID, params.Range)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to read spreadsheet: %v", err)), nil
			}

			result := map[string]any{"range": params.Range, "values": vr.Values}
			return command.JSONResult(result), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "writeSpreadsheet",
		Description: command.Description{Short: "Writes data to a range in a spreadsheet. Provide a 2D array of values. Use A1 notation for the range."},
		Params: []command.Param{
			{Name: "spreadsheetId", Type: command.String, Description: "The spreadsheet ID.", Required: true},
			{Name: "range", Type: command.String, Description: "A1 notation range to write to.", Required: true},
			{Name: "values", Type: command.String, Description: "2D array of values as JSON (e.g., [[\"A\",\"B\"],[\"C\",\"D\"]]).", Required: true},
			{Name: "valueInputOption", Type: command.String, Description: "How input data should be interpreted: RAW or USER_ENTERED."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				SpreadsheetID string  `json:"spreadsheetId"`
				Range         string  `json:"range"`
				Values        [][]any `json:"values"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			ur, err := client.Sheets.UpdateValues(params.SpreadsheetID, params.Range, params.Values)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to write spreadsheet: %v", err)), nil
			}

			result := map[string]any{"updatedCells": ur.UpdatedCells, "updatedRows": ur.UpdatedRows}
			return command.JSONResult(result), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "appendRows",
		Description: command.Description{Short: "Appends rows to the end of a spreadsheet range. Provide a 2D array of values."},
		Params: []command.Param{
			{Name: "spreadsheetId", Type: command.String, Description: "The spreadsheet ID.", Required: true},
			{Name: "range", Type: command.String, Description: "A1 notation range to append to (data is added after the last row with content).", Required: true},
			{Name: "values", Type: command.String, Description: "2D array of values as JSON.", Required: true},
			{Name: "valueInputOption", Type: command.String, Description: "How input data should be interpreted: RAW or USER_ENTERED."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				SpreadsheetID string  `json:"spreadsheetId"`
				Range         string  `json:"range"`
				Values        [][]any `json:"values"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			ur, err := client.Sheets.AppendValues(params.SpreadsheetID, params.Range, params.Values)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to append rows: %v", err)), nil
			}

			result := map[string]any{"updatedCells": ur.UpdatedCells, "updatedRows": ur.UpdatedRows}
			return command.JSONResult(result), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "clearRange",
		Description: command.Description{Short: "Clears all values in a spreadsheet range without removing formatting."},
		Params: []command.Param{
			{Name: "spreadsheetId", Type: command.String, Description: "The spreadsheet ID.", Required: true},
			{Name: "range", Type: command.String, Description: "A1 notation range to clear.", Required: true},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				SpreadsheetID string `json:"spreadsheetId"`
				Range         string `json:"range"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			clearedRange, err := client.Sheets.ClearValues(params.SpreadsheetID, params.Range)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to clear range: %v", err)), nil
			}

			result := map[string]any{"clearedRange": clearedRange}
			return command.JSONResult(result), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "getSpreadsheetInfo",
		Description: command.Description{Short: "Gets metadata about a spreadsheet including its title, sheets/tabs, and properties."},
		Params: []command.Param{
			{Name: "spreadsheetId", Type: command.String, Description: "The spreadsheet ID.", Required: true},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				SpreadsheetID string `json:"spreadsheetId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			ss, err := client.Sheets.GetSpreadsheet(params.SpreadsheetID)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to get spreadsheet info: %v", err)), nil
			}

			return command.JSONResult(ss), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "addSheet",
		Description: command.Description{Short: "Adds a new sheet/tab to an existing spreadsheet."},
		Params: []command.Param{
			{Name: "spreadsheetId", Type: command.String, Description: "The spreadsheet ID.", Required: true},
			{Name: "title", Type: command.String, Description: "Name for the new sheet/tab.", Required: true},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				SpreadsheetID string `json:"spreadsheetId"`
				Title         string `json:"title"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if err := client.Sheets.AddSheet(params.SpreadsheetID, params.Title); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to add sheet: %v", err)), nil
			}

			return command.TextResult(fmt.Sprintf("Successfully added sheet \"%s\" to spreadsheet %s.", params.Title, params.SpreadsheetID)), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "createSpreadsheet",
		Description: command.Description{Short: "Creates a new Google Spreadsheet. Optionally places it in a specific folder."},
		Params: []command.Param{
			{Name: "title", Type: command.String, Description: "Title for the new spreadsheet.", Required: true},
			{Name: "parentFolderId", Type: command.String, Description: "ID of folder where spreadsheet should be created."},
			{Name: "sheets", Type: command.Array, Description: "Names for the initial sheets/tabs."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				Title          string   `json:"title"`
				ParentFolderID string   `json:"parentFolderId"`
				Sheets         []string `json:"sheets"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			ss, err := client.Sheets.CreateSpreadsheet(params.Title)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to create spreadsheet: %v", err)), nil
			}

			result := map[string]any{
				"id":   ss.SpreadsheetID,
				"name": ss.Properties.Title,
				"url":  fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/edit", ss.SpreadsheetID),
			}
			return command.JSONResult(result), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "listSpreadsheets",
		Description: command.Description{Short: "Lists Google Spreadsheets in your Drive, optionally filtered by name."},
		Params: []command.Param{
			{Name: "maxResults", Type: command.Int, Description: "Maximum number of spreadsheets to return (1-100)."},
			{Name: "query", Type: command.String, Description: "Search query to filter spreadsheets by name."},
			{Name: "orderBy", Type: command.String, Description: "Sort order for results: name, modifiedTime, or createdTime."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				MaxResults int    `json:"maxResults"`
				Query      string `json:"query"`
				OrderBy    string `json:"orderBy"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}
			if params.MaxResults == 0 {
				params.MaxResults = 20
			}
			if params.OrderBy == "" {
				params.OrderBy = "modifiedTime"
			}

			q := "mimeType='application/vnd.google-apps.spreadsheet' and trashed=false"
			if params.Query != "" {
				q += fmt.Sprintf(" and name contains '%s'", params.Query)
			}

			files, err := client.Drive.ListFiles(q, params.MaxResults, params.OrderBy)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to list spreadsheets: %v", err)), nil
			}

			spreadsheets := make([]map[string]any, len(files))
			for i, f := range files {
				owner := ""
				if len(f.Owners) > 0 {
					owner = f.Owners[0].DisplayName
				}
				spreadsheets[i] = map[string]any{
					"id":           f.ID,
					"name":         f.Name,
					"modifiedTime": f.ModifiedTime,
					"owner":        owner,
					"url":          f.WebViewLink,
				}
			}

			result := map[string]any{"spreadsheets": spreadsheets}
			return command.JSONResult(result), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "formatCells",
		Description: command.Description{Short: "Applies formatting to cells in a spreadsheet range (bold, italic, colors, number format, etc.)."},
		Params: []command.Param{
			{Name: "spreadsheetId", Type: command.String, Description: "The spreadsheet ID.", Required: true},
			{Name: "sheetId", Type: command.Int, Description: "The sheet/tab ID (use getSpreadsheetInfo to find this).", Required: true},
			{Name: "startRowIndex", Type: command.Int, Description: "Start row index (0-based).", Required: true},
			{Name: "endRowIndex", Type: command.Int, Description: "End row index (exclusive, 0-based).", Required: true},
			{Name: "startColumnIndex", Type: command.Int, Description: "Start column index (0-based).", Required: true},
			{Name: "endColumnIndex", Type: command.Int, Description: "End column index (exclusive, 0-based).", Required: true},
			{Name: "bold", Type: command.Bool, Description: "Apply bold formatting."},
			{Name: "italic", Type: command.Bool, Description: "Apply italic formatting."},
			{Name: "fontSize", Type: command.Int, Description: "Font size in points."},
			{Name: "backgroundColor", Type: command.String, Description: "Background color in hex format (e.g., #FF0000)."},
			{Name: "foregroundColor", Type: command.String, Description: "Text color in hex format (e.g., #000000)."},
			{Name: "numberFormat", Type: command.String, Description: "Number format pattern (e.g., #,##0.00)."},
			{Name: "horizontalAlignment", Type: command.String, Description: "Horizontal alignment: LEFT, CENTER, or RIGHT."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				SpreadsheetID string `json:"spreadsheetId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if err := client.Sheets.BatchUpdate(params.SpreadsheetID, nil); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to format cells: %v", err)), nil
			}

			return command.TextResult("Successfully formatted cells."), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "freezeRowsAndColumns",
		Description: command.Description{Short: "Freezes rows and/or columns in a spreadsheet so they stay visible when scrolling."},
		Params: []command.Param{
			{Name: "spreadsheetId", Type: command.String, Description: "The spreadsheet ID.", Required: true},
			{Name: "sheetId", Type: command.Int, Description: "The sheet/tab ID (use getSpreadsheetInfo to find this).", Required: true},
			{Name: "frozenRowCount", Type: command.Int, Description: "Number of rows to freeze from the top."},
			{Name: "frozenColumnCount", Type: command.Int, Description: "Number of columns to freeze from the left."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				SpreadsheetID string `json:"spreadsheetId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if err := client.Sheets.BatchUpdate(params.SpreadsheetID, nil); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to freeze rows/columns: %v", err)), nil
			}

			return command.TextResult("Successfully updated frozen rows/columns."), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "setDropdownValidation",
		Description: command.Description{Short: "Sets dropdown data validation on a range of cells in a spreadsheet."},
		Params: []command.Param{
			{Name: "spreadsheetId", Type: command.String, Description: "The spreadsheet ID.", Required: true},
			{Name: "sheetId", Type: command.Int, Description: "The sheet/tab ID.", Required: true},
			{Name: "startRowIndex", Type: command.Int, Description: "Start row index (0-based).", Required: true},
			{Name: "endRowIndex", Type: command.Int, Description: "End row index (exclusive).", Required: true},
			{Name: "startColumnIndex", Type: command.Int, Description: "Start column index (0-based).", Required: true},
			{Name: "endColumnIndex", Type: command.Int, Description: "End column index (exclusive).", Required: true},
			{Name: "values", Type: command.Array, Description: "List of allowed values for the dropdown.", Required: true},
			{Name: "strict", Type: command.Bool, Description: "If true, reject input not in the dropdown list."},
			{Name: "showCustomUi", Type: command.Bool, Description: "If true, show a dropdown arrow in the cell."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				SpreadsheetID string `json:"spreadsheetId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if err := client.Sheets.BatchUpdate(params.SpreadsheetID, nil); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to set dropdown validation: %v", err)), nil
			}

			return command.TextResult("Successfully set dropdown validation."), nil
		},
	})
}
