package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/amarbel-llc/piers/internal/google"
	"github.com/amarbel-llc/purse-first/libs/go-mcp/command"
)

func registerCommentCommands(app *command.App, client *google.Client) {
	app.AddCommand(&command.Command{
		Name:        "listComments",
		Description: command.Description{Short: "Lists all comments in a document with their IDs, authors, status, and quoted text. Returns data needed to call getComment, replyToComment, resolveComment, or deleteComment."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string `json:"documentId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			comments, err := client.Drive.ListComments(params.DocumentID)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to list comments: %v", err)), nil
			}

			result := map[string]any{"comments": comments}
			return command.JSONResult(result), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "getComment",
		Description: command.Description{Short: "Gets a specific comment and its full reply thread. Use listComments first to find the comment ID."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "commentId", Type: command.String, Description: "The ID of the comment to retrieve.", Required: true},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string `json:"documentId"`
				CommentID  string `json:"commentId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			comment, err := client.Drive.GetComment(params.DocumentID, params.CommentID)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to get comment: %v", err)), nil
			}

			return command.JSONResult(comment), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "addComment",
		Description: command.Description{Short: "Adds a comment to the document at the specified text range. Note: programmatically created comments appear in the comments panel but may not show as anchored highlights in the document UI."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "startIndex", Type: command.Int, Description: "The starting index of the text range (inclusive, starts from 1).", Required: true},
			{Name: "endIndex", Type: command.Int, Description: "The ending index of the text range (exclusive).", Required: true},
			{Name: "content", Type: command.String, Description: "The text content of the comment.", Required: true},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string `json:"documentId"`
				StartIndex int    `json:"startIndex"`
				EndIndex   int    `json:"endIndex"`
				Content    string `json:"content"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if params.EndIndex <= params.StartIndex {
				return command.TextErrorResult("endIndex must be greater than startIndex"), nil
			}

			comment, err := client.Drive.CreateComment(params.DocumentID, params.Content, "")
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to add comment: %v", err)), nil
			}

			return command.TextResult(fmt.Sprintf("Comment added successfully. Comment ID: %s", comment.ID)), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "replyToComment",
		Description: command.Description{Short: "Adds a reply to an existing comment thread. Use listComments or getComment to find the comment ID."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "commentId", Type: command.String, Description: "The ID of the comment to reply to.", Required: true},
			{Name: "content", Type: command.String, Description: "The text content of the reply.", Required: true},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string `json:"documentId"`
				CommentID  string `json:"commentId"`
				Content    string `json:"content"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			reply, err := client.Drive.ReplyToComment(params.DocumentID, params.CommentID, params.Content)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to add reply: %v", err)), nil
			}

			return command.TextResult(fmt.Sprintf("Reply added successfully. Reply ID: %s", reply.ID)), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "resolveComment",
		Description: command.Description{Short: "Marks a comment as resolved. Note: resolved status may not persist in the Google Docs UI due to a Drive API limitation."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "commentId", Type: command.String, Description: "The ID of the comment to resolve.", Required: true},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string `json:"documentId"`
				CommentID  string `json:"commentId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if err := client.Drive.ResolveComment(params.DocumentID, params.CommentID); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to resolve comment: %v", err)), nil
			}

			return command.TextResult(fmt.Sprintf("Comment %s has been marked as resolved.", params.CommentID)), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "deleteComment",
		Description: command.Description{Short: "Permanently deletes a comment and all its replies from the document."},
		Params: []command.Param{
			{Name: "documentId", Type: command.String, Description: "The document ID — the long string between /d/ and /edit in a Google Docs URL.", Required: true},
			{Name: "commentId", Type: command.String, Description: "The ID of the comment to delete.", Required: true},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				DocumentID string `json:"documentId"`
				CommentID  string `json:"commentId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if err := client.Drive.DeleteComment(params.DocumentID, params.CommentID); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to delete comment: %v", err)), nil
			}

			return command.TextResult(fmt.Sprintf("Comment %s has been deleted.", params.CommentID)), nil
		},
	})
}
