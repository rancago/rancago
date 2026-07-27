package entities

import (
	"time"

	"github.com/rancago/framework/internal/domain/valueobjects"
)

type Document struct {
	ID         valueobjects.ID
	Title      string
	Content    string
	Embedding  []float32
	Metadata   map[string]string
	UserID     *valueobjects.ID
	SourceType string
	SourceID   string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewDocument(title, content string) *Document {
	now := time.Now()
	return &Document{
		Title:     title,
		Content:   content,
		Metadata:  make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (d *Document) UpdateContent(content string) {
	d.Content = content
	d.UpdatedAt = time.Now()
}

func (d *Document) SetEmbedding(emb []float32) {
	d.Embedding = emb
	d.UpdatedAt = time.Now()
}

func (d *Document) SetMetadata(key, value string) {
	if d.Metadata == nil {
		d.Metadata = make(map[string]string)
	}
	d.Metadata[key] = value
	d.UpdatedAt = time.Now()
}
