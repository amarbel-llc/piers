package google

type DriveFile struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	MimeType     string      `json:"mimeType,omitempty"`
	ModifiedTime string      `json:"modifiedTime,omitempty"`
	CreatedTime  string      `json:"createdTime,omitempty"`
	WebViewLink  string      `json:"webViewLink,omitempty"`
	Owners       []FileOwner `json:"owners,omitempty"`
	Parents      []string    `json:"parents,omitempty"`
}

type FileOwner struct {
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

type Comment struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type CommentReply struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type DriveService interface {
	ListFiles(query string, pageSize int, orderBy string) ([]DriveFile, error)
	GetFile(fileID string) (*DriveFile, error)
	CreateFile(name string, mimeType string, parentID string) (*DriveFile, error)
	UpdateFile(fileID string, name string, addParents string, removeParents string) (*DriveFile, error)
	CopyFile(fileID string, name string) (*DriveFile, error)
	DeleteFile(fileID string) error
	ListComments(fileID string) ([]Comment, error)
	GetComment(fileID string, commentID string) (*Comment, error)
	CreateComment(fileID string, content string, quotedContent string) (*Comment, error)
	DeleteComment(fileID string, commentID string) error
	ReplyToComment(fileID string, commentID string, content string) (*CommentReply, error)
	ResolveComment(fileID string, commentID string) error
}
