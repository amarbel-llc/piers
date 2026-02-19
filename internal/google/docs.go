package google

type Document struct {
	DocumentID string        `json:"documentId"`
	Title      string        `json:"title"`
	Body       *DocumentBody `json:"body,omitempty"`
	Tabs       []any         `json:"tabs,omitempty"`
}

type DocumentBody struct {
	Content []ContentElement `json:"content,omitempty"`
}

type ContentElement struct {
	Paragraph *Paragraph `json:"paragraph,omitempty"`
	Table     *Table     `json:"table,omitempty"`
}

type Paragraph struct {
	Elements []ParagraphElement `json:"elements,omitempty"`
}

type ParagraphElement struct {
	TextRun *TextRun `json:"textRun,omitempty"`
}

type TextRun struct {
	Content string `json:"content"`
}

type Table struct {
	TableRows []TableRow `json:"tableRows,omitempty"`
}

type TableRow struct {
	TableCells []TableCell `json:"tableCells,omitempty"`
}

type TableCell struct {
	Content []ContentElement `json:"content,omitempty"`
}

type DocsService interface {
	Get(documentID string) (*Document, error)
	BatchUpdate(documentID string, requests any) error
	Create(title string) (*Document, error)
}
