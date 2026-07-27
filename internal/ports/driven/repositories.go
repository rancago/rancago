package driven

import (
	"context"

	"github.com/rancago/framework/internal/domain/entities"
	"github.com/rancago/framework/internal/domain/valueobjects"
)

type NotificationRepository interface {
	Repository[entities.Notification]
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*entities.Notification, valueobjects.PaginationMeta, error)
	MarkRead(ctx context.Context, id valueobjects.ID, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int64, error)
}

type UserRepository interface {
	Repository[entities.User]
	FindByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error)
	FindByProvider(ctx context.Context, provider, providerID string) (*entities.User, error)
	AttachRole(ctx context.Context, userID, roleID valueobjects.ID) error
	AttachPermission(ctx context.Context, userID, permID valueobjects.ID) error
}

type RoleRepository interface {
	Repository[entities.Role]
	FindByName(ctx context.Context, name string) (*entities.Role, error)
	AttachPermission(ctx context.Context, roleID, permID valueobjects.ID) error
}

type PermissionRepository interface {
	Repository[entities.Permission]
	FindByName(ctx context.Context, name string) (*entities.Permission, error)
}

type DocumentRepository interface {
	Repository[entities.Document]
	VectorRepository[entities.Document]
	SearchByContent(ctx context.Context, query string, limit int) ([]*entities.Document, error)
}
