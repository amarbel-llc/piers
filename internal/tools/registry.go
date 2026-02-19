package tools

import (
	"github.com/amarbel-llc/piers/internal/google"
	"github.com/amarbel-llc/purse-first/libs/go-mcp/command"
)

func RegisterAll(client *google.Client) *command.App {
	app := command.NewApp("piers", "MCP server for Google Docs, Sheets, and Drive")
	app.Version = "1.0.0"

	registerDocsCommands(app, client)
	registerDriveCommands(app, client)
	registerSheetsCommands(app, client)
	registerCommentCommands(app, client)
	registerDocsStructureCommands(app, client)
	registerDocsFormattingCommands(app, client)
	registerDocsMarkdownCommands(app, client)

	return app
}
