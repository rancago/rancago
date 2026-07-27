package Contracts

import (
	"context"
	"net/http"
	"time"
)

// SocialUser is the unified user struct returned by every OAuth provider.
// All providers (Google, GitHub, Facebook, custom OIDC) normalise to this struct.
type SocialUser struct {
	Provider      string
	ID            string
	Email         string
	Name          string
	Nickname      string // GitHub username, etc.
	AvatarURL     string
	Token         string
	RefreshToken  string
	ExpiresAt     *time.Time
	RawAttributes map[string]interface{}
}

// AuthProvider is the contract every OAuth driver must implement.
// Register a new provider via SocialiteManager.RegisterDriver (OCP compliant).
type AuthProvider interface {
	// Name returns the provider slug (e.g. "google", "github").
	Name() string
	// Redirect returns the OAuth authorization URL and a CSRF state token.
	Redirect(ctx context.Context) (authURL, state string, err error)
	// Callback exchanges the code+state for a SocialUser.
	Callback(ctx context.Context, code, state string) (*SocialUser, error)
	// UserFromToken returns a SocialUser from an existing access token.
	UserFromToken(ctx context.Context, token string) (*SocialUser, error)
}

// SocialiteManager manages OAuth driver registration and dispatching.
type SocialiteManager interface {
	// Redirect delegates to the named provider's Redirect().
	Redirect(ctx context.Context, driver string) (authURL, state string, err error)
	// Callback delegates to the named provider's Callback().
	Callback(ctx context.Context, driver, code, state string) (*SocialUser, error)
	// Driver returns the named AuthProvider.
	Driver(name string) (AuthProvider, error)
	// RegisterDriver adds or overrides an OAuth driver factory (OCP).
	RegisterDriver(name string, factory func() (AuthProvider, error))
	// Drivers returns all registered driver names.
	Drivers() []string
}

// RBACService provides role and permission management backed by Redis.
type RBACService interface {
	// Role management
	AssignRole(ctx context.Context, userID, roleName string) error
	RemoveRole(ctx context.Context, userID, roleName string) error
	HasRole(ctx context.Context, userID, roleName string) (bool, error)
	HasAnyRole(ctx context.Context, userID string, roles ...string) (bool, error)
	HasAllRoles(ctx context.Context, userID string, roles ...string) (bool, error)
	GetRoles(ctx context.Context, userID string) ([]string, error)

	// Permission management
	GivePermissionToRole(ctx context.Context, roleName, permission string) error
	RevokePermissionFromRole(ctx context.Context, roleName, permission string) error
	GivePermissionToUser(ctx context.Context, userID, permission string) error
	RevokePermissionFromUser(ctx context.Context, userID, permission string) error
	HasPermission(ctx context.Context, userID, permission string) (bool, error)

	// Middleware
	Middleware(requiredPermission string) func(http.Handler) http.Handler
	RoleMiddleware(roles ...string) func(http.Handler) http.Handler
}
