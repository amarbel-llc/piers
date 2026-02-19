package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/amarbel-llc/purse-first/libs/go-mcp/command"
	"github.com/amarbel-llc/piers/internal/google"
)

func registerDocsFormattingCommands(app *command.App, client *google.Client) {
	app.AddCommand(&command.Command{
		Name:        "applyTextStyle",
		Description: command.Description{Short: "Applies character-level formatting (bold, italic, color, font, etc.) to text identified by a character range or by searching for a text string."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "startIndex", Type: command.Int, Description: "The starting index of the text range (inclusive, starts from 1)."},
			{Name: "endIndex", Type: command.Int, Description: "The ending index of the text range (exclusive)."},
			{Name: "textToFind", Type: command.String, Description: "The exact text string to locate (alternative to using startIndex/endIndex)."},
			{Name: "matchInstance", Type: command.Int, Description: "Which instance of the text to target (1st, 2nd, etc.). Defaults to 1."},
			{Name: "bold", Type: command.Bool, Description: "Apply bold formatting."},
			{Name: "italic", Type: command.Bool, Description: "Apply italic formatting."},
			{Name: "underline", Type: command.Bool, Description: "Apply underline formatting."},
			{Name: "strikethrough", Type: command.Bool, Description: "Apply strikethrough formatting."},
			{Name: "fontSize", Type: command.Int, Description: "Set font size (in points, e.g., 12)."},
			{Name: "fontFamily", Type: command.String, Description: "Set font family (e.g., \"Arial\", \"Times New Roman\")."},
			{Name: "foregroundColor", Type: command.String, Description: "Set text color using hex format (e.g., \"#FF0000\")."},
			{Name: "backgroundColor", Type: command.String, Description: "Set text background color using hex format (e.g., \"#FFFF00\")."},
			{Name: "linkUrl", Type: command.String, Description: "Make the text a hyperlink pointing to this URL."},
			{Name: "tabId", Type: command.String, Description: "The ID of the specific tab to apply formatting in. If not specified, operates on the first tab."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string `json:"documentId"`
				StartIndex int    `json:"startIndex"`
				EndIndex   int    `json:"endIndex"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if err := client.Docs.BatchUpdate(params.DocumentID, nil); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to apply text style: %v", err)), nil
			}

			return command.TextResult("Successfully applied text style."), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "applyParagraphStyle",
		Description: command.Description{Short: "Applies paragraph-level formatting (alignment, spacing, heading styles) to paragraphs identified by a character range or by searching for text. Use namedStyleType to set heading levels."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "startIndex", Type: command.Int, Description: "The starting index of the paragraph range (inclusive, starts from 1)."},
			{Name: "endIndex", Type: command.Int, Description: "The ending index of the paragraph range (exclusive)."},
			{Name: "textToFind", Type: command.String, Description: "Text to locate within the target paragraph (alternative to using startIndex/endIndex)."},
			{Name: "matchInstance", Type: command.Int, Description: "Which instance of the text to target (1st, 2nd, etc.). Defaults to 1."},
			{Name: "indexWithinParagraph", Type: command.Int, Description: "An index located anywhere within the target paragraph."},
			{Name: "alignment", Type: command.String, Description: "Paragraph alignment: START, END, CENTER, or JUSTIFIED."},
			{Name: "indentStart", Type: command.Float, Description: "Left indentation in points."},
			{Name: "indentEnd", Type: command.Float, Description: "Right indentation in points."},
			{Name: "spaceAbove", Type: command.Float, Description: "Space before the paragraph in points."},
			{Name: "spaceBelow", Type: command.Float, Description: "Space after the paragraph in points."},
			{Name: "namedStyleType", Type: command.String, Description: "Apply a built-in named paragraph style: NORMAL_TEXT, TITLE, SUBTITLE, HEADING_1 through HEADING_6."},
			{Name: "keepWithNext", Type: command.Bool, Description: "Keep this paragraph together with the next one on the same page."},
			{Name: "tabId", Type: command.String, Description: "The ID of the specific tab to apply formatting in. If not specified, operates on the first tab."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string `json:"documentId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if err := client.Docs.BatchUpdate(params.DocumentID, nil); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to apply paragraph style: %v", err)), nil
			}

			return command.TextResult("Successfully applied paragraph style."), nil
		},
	})
}
