package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/amarbel-llc/piers/internal/google"
	"github.com/amarbel-llc/purse-first/libs/go-mcp/command"
)

type documentInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ModifiedTime string `json:"modifiedTime,omitempty"`
	Owner        string `json:"owner,omitempty"`
	URL          string `json:"url,omitempty"`
}

func filesToDocumentInfos(files []google.DriveFile) []documentInfo {
	docs := make([]documentInfo, len(files))
	for i, f := range files {
		owner := ""
		if len(f.Owners) > 0 {
			owner = f.Owners[0].DisplayName
		}
		docs[i] = documentInfo{
			ID:           f.ID,
			Name:         f.Name,
			ModifiedTime: f.ModifiedTime,
			Owner:        owner,
			URL:          f.WebViewLink,
		}
	}
	return docs
}

func registerDriveCommands(app *command.App, client *google.Client) {
	app.AddCommand(&command.Command{
		Name:        "listDocuments",
		Description: command.Description{Short: "Lists Google Documents in your Drive, optionally filtered by name or content. Use modifiedAfter to find recently changed documents."},
		Params: []command.Param{
			{Name: "maxResults", Type: command.Int, Description: "Maximum number of documents to return (1-100)."},
			{Name: "query", Type: command.String, Description: "Search query to filter documents by name or content."},
			{Name: "orderBy", Type: command.String, Description: "Sort order for results: name, modifiedTime, or createdTime."},
			{Name: "modifiedAfter", Type: command.String, Description: "Only return documents modified after this date (ISO 8601 format, e.g., \"2024-01-01\")."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				MaxResults    int    `json:"maxResults"`
				Query         string `json:"query"`
				OrderBy       string `json:"orderBy"`
				ModifiedAfter string `json:"modifiedAfter"`
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

			q := "mimeType='application/vnd.google-apps.document' and trashed=false"
			if params.Query != "" {
				q += fmt.Sprintf(" and (name contains '%s' or fullText contains '%s')", params.Query, params.Query)
			}
			if params.ModifiedAfter != "" {
				q += fmt.Sprintf(" and modifiedTime > '%s'", params.ModifiedAfter)
			}

			files, err := client.Drive.ListFiles(q, params.MaxResults, params.OrderBy)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to list documents: %v", err)), nil
			}

			result := map[string]any{"documents": filesToDocumentInfos(files)}
			return command.JSONResult(result), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "searchDocuments",
		Description: command.Description{Short: "Searches for documents by name, content, or both. Use listDocuments for browsing and this tool for targeted queries."},
		Params: []command.Param{
			{Name: "query", Type: command.String, Description: "Search term to find in document names or content.", Required: true},
			{Name: "searchIn", Type: command.String, Description: "Where to search: name, content, or both."},
			{Name: "maxResults", Type: command.Int, Description: "Maximum number of results to return."},
			{Name: "modifiedAfter", Type: command.String, Description: "Only return documents modified after this date (ISO 8601 format)."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				Query         string `json:"query"`
				SearchIn      string `json:"searchIn"`
				MaxResults    int    `json:"maxResults"`
				ModifiedAfter string `json:"modifiedAfter"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}
			if params.SearchIn == "" {
				params.SearchIn = "both"
			}
			if params.MaxResults == 0 {
				params.MaxResults = 10
			}

			q := "mimeType='application/vnd.google-apps.document' and trashed=false"
			switch params.SearchIn {
			case "name":
				q += fmt.Sprintf(" and name contains '%s'", params.Query)
			case "content":
				q += fmt.Sprintf(" and fullText contains '%s'", params.Query)
			default:
				q += fmt.Sprintf(" and (name contains '%s' or fullText contains '%s')", params.Query, params.Query)
			}
			if params.ModifiedAfter != "" {
				q += fmt.Sprintf(" and modifiedTime > '%s'", params.ModifiedAfter)
			}

			files, err := client.Drive.ListFiles(q, params.MaxResults, "modifiedTime desc")
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to search documents: %v", err)), nil
			}

			result := map[string]any{"documents": filesToDocumentInfos(files)}
			return command.JSONResult(result), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "getDocumentInfo",
		Description: command.Description{Short: "Gets metadata about a document including its name, owner, sharing status, and modification history."},
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

			file, err := client.Drive.GetFile(params.DocumentID)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to get document info: %v", err)), nil
			}

			owner := ""
			if len(file.Owners) > 0 {
				owner = file.Owners[0].DisplayName
			}
			info := map[string]any{
				"id":           file.ID,
				"name":         file.Name,
				"mimeType":     file.MimeType,
				"createdTime":  file.CreatedTime,
				"modifiedTime": file.ModifiedTime,
				"owner":        owner,
				"url":          file.WebViewLink,
			}
			return command.JSONResult(info), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "createFolder",
		Description: command.Description{Short: "Creates a new folder in Google Drive. Optionally places it inside an existing parent folder."},
		Params: []command.Param{
			{Name: "name", Type: command.String, Description: "Name for the new folder.", Required: true},
			{Name: "parentFolderId", Type: command.String, Description: "Parent folder ID. If not provided, creates folder in Drive root."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				Name           string `json:"name"`
				ParentFolderID string `json:"parentFolderId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			file, err := client.Drive.CreateFile(params.Name, "application/vnd.google-apps.folder", params.ParentFolderID)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to create folder: %v", err)), nil
			}

			result := map[string]any{"id": file.ID, "name": file.Name, "url": file.WebViewLink}
			return command.JSONResult(result), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "listFolderContents",
		Description: command.Description{Short: "Lists files and subfolders within a Drive folder. Use folderId='root' to browse the top-level of the Drive."},
		Params: []command.Param{
			{Name: "folderId", Type: command.String, Description: "ID of the folder to list contents of. Use \"root\" for the root Drive folder.", Required: true},
			{Name: "includeSubfolders", Type: command.Bool, Description: "Whether to include subfolders in results."},
			{Name: "includeFiles", Type: command.Bool, Description: "Whether to include files in results."},
			{Name: "maxResults", Type: command.Int, Description: "Maximum number of items to return."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				FolderID          string `json:"folderId"`
				IncludeSubfolders *bool  `json:"includeSubfolders"`
				IncludeFiles      *bool  `json:"includeFiles"`
				MaxResults        int    `json:"maxResults"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}
			if params.MaxResults == 0 {
				params.MaxResults = 50
			}
			inclSubfolders := params.IncludeSubfolders == nil || *params.IncludeSubfolders
			inclFiles := params.IncludeFiles == nil || *params.IncludeFiles

			q := fmt.Sprintf("'%s' in parents and trashed=false", params.FolderID)
			if !inclSubfolders {
				q += " and mimeType!='application/vnd.google-apps.folder'"
			} else if !inclFiles {
				q += " and mimeType='application/vnd.google-apps.folder'"
			}

			files, err := client.Drive.ListFiles(q, params.MaxResults, "folder,name")
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to list folder contents: %v", err)), nil
			}

			var folders, items []map[string]any
			for _, f := range files {
				if f.MimeType == "application/vnd.google-apps.folder" {
					folders = append(folders, map[string]any{"id": f.ID, "name": f.Name, "modifiedTime": f.ModifiedTime})
				} else {
					items = append(items, map[string]any{"id": f.ID, "name": f.Name, "mimeType": f.MimeType, "modifiedTime": f.ModifiedTime})
				}
			}
			result := map[string]any{"folders": folders, "files": items}
			return command.JSONResult(result), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "getFolderInfo",
		Description: command.Description{Short: "Gets metadata about a Drive folder including its name, owner, sharing status, and parent folder."},
		Params: []command.Param{
			{Name: "folderId", Type: command.String, Description: "ID of the folder to get information about.", Required: true},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				FolderID string `json:"folderId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			file, err := client.Drive.GetFile(params.FolderID)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to get folder info: %v", err)), nil
			}

			owner := ""
			if len(file.Owners) > 0 {
				owner = file.Owners[0].DisplayName
			}
			var parentID *string
			if len(file.Parents) > 0 {
				parentID = &file.Parents[0]
			}
			info := map[string]any{
				"id":             file.ID,
				"name":           file.Name,
				"createdTime":    file.CreatedTime,
				"modifiedTime":   file.ModifiedTime,
				"owner":          owner,
				"url":            file.WebViewLink,
				"parentFolderId": parentID,
			}
			return command.JSONResult(info), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "moveFile",
		Description: command.Description{Short: "Moves a file or folder to a different Drive folder. By default adds the new parent while keeping existing parents; set removeFromAllParents=true for a true move."},
		Params: []command.Param{
			{Name: "fileId", Type: command.String, Description: "The file or folder ID from a Google Drive URL or a previous tool result.", Required: true},
			{Name: "newParentId", Type: command.String, Description: "ID of the destination folder. Use \"root\" for Drive root.", Required: true},
			{Name: "removeFromAllParents", Type: command.Bool, Description: "If true, removes from all current parents. If false, adds to new parent while keeping existing parents."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				FileID               string `json:"fileId"`
				NewParentID          string `json:"newParentId"`
				RemoveFromAllParents bool   `json:"removeFromAllParents"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			file, err := client.Drive.UpdateFile(params.FileID, "", params.NewParentID, "")
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to move file: %v", err)), nil
			}

			return command.TextResult(fmt.Sprintf("Successfully moved file to new location.\nFile ID: %s", file.ID)), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "copyFile",
		Description: command.Description{Short: "Creates a copy of a file or document in Google Drive. Returns the new copy's ID and URL."},
		Params: []command.Param{
			{Name: "fileId", Type: command.String, Description: "The file or folder ID from a Google Drive URL or a previous tool result.", Required: true},
			{Name: "newName", Type: command.String, Description: "Name for the copied file. If not provided, will use \"Copy of [original name]\"."},
			{Name: "parentFolderId", Type: command.String, Description: "ID of folder where copy should be placed. If not provided, places in same location as original."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				FileID         string `json:"fileId"`
				NewName        string `json:"newName"`
				ParentFolderID string `json:"parentFolderId"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			file, err := client.Drive.CopyFile(params.FileID, params.NewName)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to copy file: %v", err)), nil
			}

			result := map[string]any{"id": file.ID, "name": file.Name, "url": file.WebViewLink}
			return command.JSONResult(result), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "renameFile",
		Description: command.Description{Short: "Renames a file or folder in Google Drive. Returns the updated file info."},
		Params: []command.Param{
			{Name: "fileId", Type: command.String, Description: "The file or folder ID from a Google Drive URL or a previous tool result.", Required: true},
			{Name: "newName", Type: command.String, Description: "New name for the file or folder.", Required: true},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				FileID  string `json:"fileId"`
				NewName string `json:"newName"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			file, err := client.Drive.UpdateFile(params.FileID, params.NewName, "", "")
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to rename file: %v", err)), nil
			}

			return command.TextResult(fmt.Sprintf("Successfully renamed to \"%s\" (ID: %s)", file.Name, file.ID)), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "deleteFile",
		Description: command.Description{Short: "Moves a file or folder to the trash, or permanently deletes it. Set permanent=true for irreversible deletion."},
		Params: []command.Param{
			{Name: "fileId", Type: command.String, Description: "The file or folder ID from a Google Drive URL or a previous tool result.", Required: true},
			{Name: "permanent", Type: command.Bool, Description: "If true, permanently deletes the file instead of moving it to trash."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				FileID    string `json:"fileId"`
				Permanent bool   `json:"permanent"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			if err := client.Drive.DeleteFile(params.FileID); err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to delete file: %v", err)), nil
			}

			action := "trashed"
			if params.Permanent {
				action = "permanently_deleted"
			}
			result := map[string]any{
				"success": true,
				"action":  action,
				"fileId":  params.FileID,
				"message": fmt.Sprintf("Successfully %s file %s.", action, params.FileID),
			}
			return command.JSONResult(result), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "createDocument",
		Description: command.Description{Short: "Creates a new empty Google Document. Optionally places it in a specific folder and adds initial text content."},
		Params: []command.Param{
			{Name: "title", Type: command.String, Description: "Title for the new document.", Required: true},
			{Name: "parentFolderId", Type: command.String, Description: "ID of folder where document should be created. If not provided, creates in Drive root."},
			{Name: "initialContent", Type: command.String, Description: "Initial content to add to the document."},
			{Name: "contentFormat", Type: command.String, Description: "How to interpret initialContent. 'markdown' (default) converts markdown to formatted Google Docs content. 'raw' inserts the text as-is."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				Title          string `json:"title"`
				ParentFolderID string `json:"parentFolderId"`
				InitialContent string `json:"initialContent"`
				ContentFormat  string `json:"contentFormat"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			doc, err := client.Docs.Create(params.Title)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to create document: %v", err)), nil
			}

			result := map[string]any{
				"id":   doc.DocumentID,
				"name": doc.Title,
				"url":  fmt.Sprintf("https://docs.google.com/document/d/%s/edit", doc.DocumentID),
			}
			return command.JSONResult(result), nil
		},
	})

	app.AddCommand(&command.Command{
		Name:        "createDocumentFromTemplate",
		Description: command.Description{Short: "Creates a new document by copying an existing template and optionally replacing placeholder text."},
		Params: []command.Param{
			{Name: "templateId", Type: command.String, Description: "ID of the template document to copy from.", Required: true},
			{Name: "newTitle", Type: command.String, Description: "Title for the new document.", Required: true},
			{Name: "parentFolderId", Type: command.String, Description: "ID of folder where document should be created. If not provided, creates in Drive root."},
			{Name: "replacements", Type: command.String, Description: "Key-value pairs for text replacements in the template (JSON object, e.g., {\"{{NAME}}\": \"John Doe\"})."},
		},
		Run: func(ctx context.Context, args json.RawMessage, _ command.Prompter) (*command.Result, error) {
			var params struct {
				TemplateID     string            `json:"templateId"`
				NewTitle       string            `json:"newTitle"`
				ParentFolderID string            `json:"parentFolderId"`
				Replacements   map[string]string `json:"replacements"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return command.TextErrorResult(fmt.Sprintf("invalid arguments: %v", err)), nil
			}

			file, err := client.Drive.CopyFile(params.TemplateID, params.NewTitle)
			if err != nil {
				return command.TextErrorResult(fmt.Sprintf("failed to create document from template: %v", err)), nil
			}

			return command.TextResult(fmt.Sprintf("Successfully created document \"%s\" from template (ID: %s)", file.Name, file.ID)), nil
		},
	})
}
