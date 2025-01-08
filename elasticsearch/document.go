package hamfts

import "time"

type Document struct {
	ID        string
	Content   string
	CreatedAt time.Time
	Metadata  map[string]interface{}
}

func NewDocument(id string, content string) *Document {
	return &Document{
		ID:        id,
		Content:   content,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}
