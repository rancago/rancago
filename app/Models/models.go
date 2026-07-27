// Package Models contains the GORM entity models for Rancago Framework.
// These are the actual database representations (ORM layer).
package Models

import (
	"time"

	"gorm.io/gorm"
)

// User is the GORM model for the users table.
// Many-to-many with Role and Permission (direct grants).
type User struct {
	ID              uint           `gorm:"primaryKey"                               json:"id"`
	Name            string         `gorm:"size:255;not null"                        json:"name"`
	Email           string         `gorm:"size:255;uniqueIndex;not null"            json:"email"`
	Password        string         `gorm:"size:255"                                 json:"-"`
	AvatarURL       string         `gorm:"size:1024"                                json:"avatar_url,omitempty"`
	Provider        string         `gorm:"size:50;index"                            json:"provider,omitempty"`
	ProviderID      string         `gorm:"size:255;index"                           json:"provider_id,omitempty"`
	RememberToken   string         `gorm:"size:100"                                 json:"-"`
	EmailVerifiedAt *time.Time     `                                                json:"email_verified_at,omitempty"`
	Roles           []*Role        `gorm:"many2many:user_roles"                     json:"roles,omitempty"`
	Permissions     []*Permission  `gorm:"many2many:user_permissions"               json:"permissions,omitempty"`
	CreatedAt       time.Time      `                                                json:"created_at"`
	UpdatedAt       time.Time      `                                                json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index"                                    json:"deleted_at,omitempty"`
}

func (User) TableName() string { return "users" }

// Role is the GORM model for the roles table.
type Role struct {
	ID          uint          `gorm:"primaryKey"                        json:"id"`
	Name        string        `gorm:"size:100;uniqueIndex;not null"     json:"name"`
	Label       string        `gorm:"size:255"                          json:"label"`
	Permissions []*Permission `gorm:"many2many:role_permissions"        json:"permissions,omitempty"`
	CreatedAt   time.Time     `                                         json:"created_at"`
	UpdatedAt   time.Time     `                                         json:"updated_at"`
}

func (Role) TableName() string { return "roles" }

// Permission is the GORM model for the permissions table.
// Action + Resource enables granular control: e.g. action="delete", resource="user".
type Permission struct {
	ID        uint      `gorm:"primaryKey"                     json:"id"`
	Name      string    `gorm:"size:150;uniqueIndex;not null"  json:"name"`
	Action    string    `gorm:"size:100;not null"              json:"action"`
	Resource  string    `gorm:"size:100;not null"              json:"resource"`
	CreatedAt time.Time `                                      json:"created_at"`
	UpdatedAt time.Time `                                      json:"updated_at"`
}

func (Permission) TableName() string { return "permissions" }

// Document is the GORM model for semantic-search enabled documents.
// The Embedding column is a 1536-dimension pgvector with an HNSW cosine index.
type Document struct {
	ID         uint           `gorm:"primaryKey"                                                                                                                        json:"id"`
	Title      string         `gorm:"size:500;not null"                                                                                                                 json:"title"`
	Content    string         `gorm:"type:text"                                                                                                                        json:"content"`
	UserID     *uint          `gorm:"index"                                                                                                                            json:"user_id,omitempty"`
	SourceType string         `gorm:"size:100"                                                                                                                         json:"source_type,omitempty"`
	SourceID   string         `gorm:"size:255"                                                                                                                         json:"source_id,omitempty"`
	Metadata   []byte         `gorm:"type:jsonb"                                                                                                                       json:"metadata,omitempty"`
	// pgvector HNSW index — production-ready for cosine similarity semantic search.
	Embedding  Vector         `gorm:"type:vector(1536);index:idx_documents_embedding,type:hnsw,using:hnsw,opclass:vector_cosine_ops"                                   json:"-"`
	CreatedAt  time.Time      `                                                                                                                                        json:"created_at"`
	UpdatedAt  time.Time      `                                                                                                                                        json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index"                                                                                                                           json:"deleted_at,omitempty"`
}

func (Document) TableName() string { return "documents" }

// Notification is the GORM model for persisted notifications.
type Notification struct {
	ID        uint       `gorm:"primaryKey"                      json:"id"`
	UserID    string     `gorm:"size:255;index;not null"         json:"user_id"`
	Title     string     `gorm:"size:500;not null"               json:"title"`
	Body      string     `gorm:"type:text"                       json:"body"`
	Channel   string     `gorm:"size:50;not null;default:database" json:"channel"`
	Data      []byte     `gorm:"type:jsonb"                      json:"data,omitempty"`
	Read      bool       `gorm:"default:false"                   json:"read"`
	ReadAt    *time.Time `                                       json:"read_at,omitempty"`
	CreatedAt time.Time  `                                       json:"created_at"`
}

func (Notification) TableName() string { return "notifications" }
