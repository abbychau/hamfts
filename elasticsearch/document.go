package hamfts

import (
	"encoding/gob"
	"time"
)

func init() {
	// Register types for gob encoding
	gob.Register(map[string]interface{}{})
	gob.Register(time.Time{})
}

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
