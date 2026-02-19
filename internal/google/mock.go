package google

func newMockClient() *Client {
	return &Client{
		Docs:   &mockDocsService{},
		Drive:  &mockDriveService{},
		Sheets: &mockSheetsService{},
	}
}

type mockDocsService struct{}

func (m *mockDocsService) Get(documentID string) (*Document, error) {
	return &Document{
		DocumentID: "mock-doc-id-123",
		Title:      "Mock Document",
		Body: &DocumentBody{
			Content: []ContentElement{
				{Paragraph: &Paragraph{
					Elements: []ParagraphElement{
						{TextRun: &TextRun{Content: "Hello from the mock document.\n"}},
					},
				}},
			},
		},
		Tabs: []any{},
	}, nil
}

func (m *mockDocsService) BatchUpdate(documentID string, requests any) error { return nil }

func (m *mockDocsService) Create(title string) (*Document, error) {
	doc, _ := m.Get("")
	return doc, nil
}

type mockDriveService struct{}

var mockFiles = []DriveFile{
	{
		ID: "mock-doc-id-123", Name: "Mock Document",
		MimeType: "application/vnd.google-apps.document",
		ModifiedTime: "2025-01-15T10:30:00.000Z", CreatedTime: "2025-01-01T08:00:00.000Z",
		WebViewLink: "https://docs.google.com/document/d/mock-doc-id-123/edit",
		Owners:       []FileOwner{{DisplayName: "Test User", EmailAddress: "test@example.com"}},
	},
	{
		ID: "mock-sheet-id-456", Name: "Mock Spreadsheet",
		MimeType: "application/vnd.google-apps.spreadsheet",
		ModifiedTime: "2025-01-14T09:00:00.000Z", CreatedTime: "2025-01-02T08:00:00.000Z",
		WebViewLink: "https://docs.google.com/spreadsheets/d/mock-sheet-id-456/edit",
		Owners:       []FileOwner{{DisplayName: "Test User", EmailAddress: "test@example.com"}},
	},
}

func (m *mockDriveService) ListFiles(q string, ps int, ob string) ([]DriveFile, error) {
	return mockFiles, nil
}
func (m *mockDriveService) GetFile(id string) (*DriveFile, error) { return &mockFiles[0], nil }
func (m *mockDriveService) CreateFile(n, mt, p string) (*DriveFile, error) {
	return &mockFiles[0], nil
}
func (m *mockDriveService) UpdateFile(id, n, ap, rp string) (*DriveFile, error) {
	return &mockFiles[0], nil
}
func (m *mockDriveService) CopyFile(id, n string) (*DriveFile, error) {
	f := mockFiles[0]
	f.ID = "mock-copy-id"
	return &f, nil
}
func (m *mockDriveService) DeleteFile(id string) error { return nil }
func (m *mockDriveService) ListComments(id string) ([]Comment, error) {
	return []Comment{}, nil
}
func (m *mockDriveService) GetComment(fid, cid string) (*Comment, error) {
	return &Comment{ID: "mock-comment-id", Content: "Mock comment"}, nil
}
func (m *mockDriveService) CreateComment(fid, c, q string) (*Comment, error) {
	return &Comment{ID: "mock-comment-id", Content: "Mock comment"}, nil
}
func (m *mockDriveService) DeleteComment(fid, cid string) error { return nil }
func (m *mockDriveService) ReplyToComment(fid, cid, c string) (*CommentReply, error) {
	return &CommentReply{ID: "mock-reply-id", Content: "Mock reply"}, nil
}
func (m *mockDriveService) ResolveComment(fid, cid string) error { return nil }

type mockSheetsService struct{}

func (m *mockSheetsService) GetValues(sid, r string) (*ValueRange, error) {
	return &ValueRange{
		Range:  "Sheet1!A1:B3",
		Values: [][]any{{"Name", "Score"}, {"Alice", "95"}, {"Bob", "87"}},
	}, nil
}
func (m *mockSheetsService) UpdateValues(sid, r string, v [][]any) (*UpdateResult, error) {
	return &UpdateResult{UpdatedCells: 6, UpdatedRows: 3}, nil
}
func (m *mockSheetsService) AppendValues(sid, r string, v [][]any) (*UpdateResult, error) {
	return &UpdateResult{UpdatedCells: 2, UpdatedRows: 1}, nil
}
func (m *mockSheetsService) ClearValues(sid, r string) (string, error) {
	return "Sheet1!A1:B3", nil
}
func (m *mockSheetsService) GetSpreadsheet(sid string) (*Spreadsheet, error) {
	return &Spreadsheet{
		SpreadsheetID: "mock-sheet-id-456",
		Properties:    SpreadsheetProps{Title: "Mock Spreadsheet"},
		Sheets: []SpreadsheetSheet{
			{Properties: SheetProperties{SheetID: 0, Title: "Sheet1", Index: 0}},
		},
	}, nil
}
func (m *mockSheetsService) CreateSpreadsheet(title string) (*Spreadsheet, error) {
	return &Spreadsheet{
		SpreadsheetID: "mock-new-sheet-id",
		Properties:    SpreadsheetProps{Title: title},
	}, nil
}
func (m *mockSheetsService) AddSheet(sid, title string) error    { return nil }
func (m *mockSheetsService) BatchUpdate(sid string, req any) error { return nil }
