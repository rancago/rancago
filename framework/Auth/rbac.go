package Auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rancago/framework/app/Contracts"
	"github.com/rancago/framework/framework/Cache"
)

// RBACService provides Redis-backed role and permission management.
// Keys follow the convention:
//
//	rbac:user:{userID}:roles        (Set)
//	rbac:role:{roleName}:permissions (Set)
//	rbac:user:{userID}:permissions  (Set, for direct user grants)
//
// This satisfies Contracts.RBACService.
type RBACService struct {
	redis *Cache.RedisManager
	// userIDKey is a middleware context key for extracting the user ID.
	userIDKey interface{}
}

// NewRBACService creates a new Redis-backed RBAC service.
// userIDKey is the context key used by auth middleware to store the user ID.
func NewRBACService(redis *Cache.RedisManager, userIDKey interface{}) Contracts.RBACService {
	return &RBACService{redis: redis, userIDKey: userIDKey}
}

// ---- Role management ----

func (r *RBACService) AssignRole(ctx context.Context, userID, roleName string) error {
	return r.redis.SAdd(ctx, userRoleKey(userID), roleName)
}

func (r *RBACService) RemoveRole(ctx context.Context, userID, roleName string) error {
	return r.redis.SRem(ctx, userRoleKey(userID), roleName)
}

func (r *RBACService) HasRole(ctx context.Context, userID, roleName string) (bool, error) {
	return r.redis.SIsMember(ctx, userRoleKey(userID), roleName)
}

func (r *RBACService) HasAnyRole(ctx context.Context, userID string, roles ...string) (bool, error) {
	for _, role := range roles {
		ok, err := r.HasRole(ctx, userID, role)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func (r *RBACService) HasAllRoles(ctx context.Context, userID string, roles ...string) (bool, error) {
	for _, role := range roles {
		ok, err := r.HasRole(ctx, userID, role)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func (r *RBACService) GetRoles(ctx context.Context, userID string) ([]string, error) {
	return r.redis.SMembers(ctx, userRoleKey(userID))
}

// ---- Permission management ----

func (r *RBACService) GivePermissionToRole(ctx context.Context, roleName, permission string) error {
	return r.redis.SAdd(ctx, rolePermKey(roleName), permission)
}

func (r *RBACService) RevokePermissionFromRole(ctx context.Context, roleName, permission string) error {
	return r.redis.SRem(ctx, rolePermKey(roleName), permission)
}

func (r *RBACService) GivePermissionToUser(ctx context.Context, userID, permission string) error {
	return r.redis.SAdd(ctx, userPermKey(userID), permission)
}

func (r *RBACService) RevokePermissionFromUser(ctx context.Context, userID, permission string) error {
	return r.redis.SRem(ctx, userPermKey(userID), permission)
}

// HasPermission checks:
//  1. Direct user permission grant.
//  2. Permission via any assigned role.
func (r *RBACService) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	// 1. Direct grant
	ok, err := r.redis.SIsMember(ctx, userPermKey(userID), permission)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	// 2. Via roles
	roles, err := r.GetRoles(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, role := range roles {
		ok, err = r.redis.SIsMember(ctx, rolePermKey(role), permission)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// ---- Middleware ----

// Middleware returns an HTTP middleware that blocks requests from users who
// do not have the required permission. The user ID is read from the request context
// using the userIDKey supplied to NewRBACService.
func (r *RBACService) Middleware(requiredPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			userID, ok := req.Context().Value(r.userIDKey).(string)
			if !ok || userID == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			allowed, err := r.HasPermission(req.Context(), userID, requiredPermission)
			if err != nil || !allowed {
				http.Error(w, fmt.Sprintf(`{"error":"forbidden","required":"%s"}`, requiredPermission), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

// RoleMiddleware returns an HTTP middleware that blocks requests from users who
// do not hold at least one of the required roles.
func (r *RBACService) RoleMiddleware(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			userID, ok := req.Context().Value(r.userIDKey).(string)
			if !ok || userID == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			allowed, err := r.HasAnyRole(req.Context(), userID, roles...)
			if err != nil || !allowed {
				http.Error(w, `{"error":"forbidden","required_roles":"one of `+fmt.Sprintf("%v", roles)+`"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

// ---- key helpers ----

func userRoleKey(userID string) string  { return "rbac:user:" + userID + ":roles" }
func rolePermKey(role string) string    { return "rbac:role:" + role + ":permissions" }
func userPermKey(userID string) string  { return "rbac:user:" + userID + ":permissions" }
