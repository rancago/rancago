package entities

import (
	"time"

	"github.com/rancago/framework/internal/domain/valueobjects"
)

type Permission struct {
	ID        valueobjects.ID
	Name      string
	Action    string
	Resource  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewPermission(name, action, resource string) *Permission {
	now := time.Now()
	return &Permission{
		Name:      name,
		Action:    action,
		Resource:  resource,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
