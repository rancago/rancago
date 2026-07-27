package entities

import (
	"time"

	"github.com/rancago/framework/internal/domain/valueobjects"
)

type Role struct {
	ID          valueobjects.ID
	Name        string
	Label       string
	Permissions []*Permission
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewRole(name, label string) *Role {
	now := time.Now()
	return &Role{
		Name:      name,
		Label:     label,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (r *Role) AttachPermission(p *Permission) {
	for _, ep := range r.Permissions {
		if ep.Name == p.Name {
			return
		}
	}
	r.Permissions = append(r.Permissions, p)
	r.UpdatedAt = time.Now()
}
