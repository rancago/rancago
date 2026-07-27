package entities

import (
	"time"

	"github.com/rancago/framework/internal/domain/valueobjects"
)

type User struct {
	ID              valueobjects.ID
	Name            string
	Email           valueobjects.Email
	PasswordHash    string
	AvatarURL       string
	Provider        string
	ProviderID      string
	RememberToken   string
	EmailVerifiedAt *time.Time
	Roles           []*Role
	Permissions     []*Permission
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewUser(name string, email valueobjects.Email) *User {
	now := time.Now()
	return &User{
		Name:      name,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (u *User) HasRole(roleName string) bool {
	for _, r := range u.Roles {
		if r.Name == roleName {
			return true
		}
	}
	return false
}

func (u *User) HasPermission(permName string) bool {
	for _, p := range u.Permissions {
		if p.Name == permName {
			return true
		}
	}
	for _, r := range u.Roles {
		for _, rp := range r.Permissions {
			if rp.Name == permName {
				return true
			}
		}
	}
	return false
}

func (u *User) VerifyEmail() {
	now := time.Now()
	u.EmailVerifiedAt = &now
	u.UpdatedAt = now
}

func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}
