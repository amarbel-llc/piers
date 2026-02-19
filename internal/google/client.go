package google

import (
	"context"
	"fmt"
	"os"
)

type Client struct {
	Docs   DocsService
	Drive  DriveService
	Sheets SheetsService
}

func NewClient(ctx context.Context) (*Client, error) {
	if os.Getenv("MOCK_AUTH") == "1" {
		return newMockClient(), nil
	}
	return nil, fmt.Errorf("real OAuth2 not yet implemented; set MOCK_AUTH=1 for testing")
}
