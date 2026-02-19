package tools

import (
	"github.com/amarbel-llc/purse-first/libs/go-mcp/command"
	"github.com/amarbel-llc/piers/internal/google"
)

func RegisterAll(client *google.Client) *command.App {
	app := command.NewApp("piers", "MCP server for Google Docs, Sheets, and Drive")
	app.Version = "1.0.0"
	return app
}
