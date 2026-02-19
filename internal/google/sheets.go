package google

type ValueRange struct {
	Range  string  `json:"range"`
	Values [][]any `json:"values,omitempty"`
}

type UpdateResult struct {
	UpdatedCells int `json:"updatedCells"`
	UpdatedRows  int `json:"updatedRows"`
}

type Spreadsheet struct {
	SpreadsheetID string             `json:"spreadsheetId"`
	Properties    SpreadsheetProps   `json:"properties"`
	Sheets        []SpreadsheetSheet `json:"sheets,omitempty"`
}

type SpreadsheetProps struct {
	Title string `json:"title"`
}

type SpreadsheetSheet struct {
	Properties SheetProperties `json:"properties"`
}

type SheetProperties struct {
	SheetID int    `json:"sheetId"`
	Title   string `json:"title"`
	Index   int    `json:"index"`
}

type SheetsService interface {
	GetValues(spreadsheetID string, rangeStr string) (*ValueRange, error)
	UpdateValues(spreadsheetID string, rangeStr string, values [][]any) (*UpdateResult, error)
	AppendValues(spreadsheetID string, rangeStr string, values [][]any) (*UpdateResult, error)
	ClearValues(spreadsheetID string, rangeStr string) (string, error)
	GetSpreadsheet(spreadsheetID string) (*Spreadsheet, error)
	CreateSpreadsheet(title string) (*Spreadsheet, error)
	AddSheet(spreadsheetID string, title string) error
	BatchUpdate(spreadsheetID string, requests any) error
}
