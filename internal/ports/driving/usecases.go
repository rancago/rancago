package driving

import (
	"context"

	"github.com/rancago/framework/internal/domain/entities"
	"github.com/rancago/framework/internal/domain/valueobjects"
)

type NotificationUseCase interface {
	Send(ctx context.Context, userID, title, body string, channel entities.NotificationChannel, data map[string]string) (*entities.Notification, error)
	Broadcast(ctx context.Context, title, body string, data map[string]string) error
	ListUserNotifications(ctx context.Context, userID string, page, perPage int) ([]*entities.Notification, valueobjects.PaginationMeta, error)
	MarkRead(ctx context.Context, id valueobjects.ID, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int64, error)
}

type UserUseCase interface {
	Register(ctx context.Context, name, email, password string) (*entities.User, error)
	Login(ctx context.Context, email, password string) (*entities.User, error)
	FindByID(ctx context.Context, id valueobjects.ID) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	LoginWithProvider(ctx context.Context, provider, code string) (*entities.User, error)
	GetAuthURL(ctx context.Context, provider, state string) (string, error)
	AssignRole(ctx context.Context, userID valueobjects.ID, roleName string) error
	HasPermission(ctx context.Context, userID valueobjects.ID, permName string) (bool, error)
}

type DocumentUseCase interface {
	Create(ctx context.Context, title, content string, userID *valueobjects.ID) (*entities.Document, error)
	GetByID(ctx context.Context, id valueobjects.ID) (*entities.Document, error)
	Update(ctx context.Context, id valueobjects.ID, title, content string) (*entities.Document, error)
	Delete(ctx context.Context, id valueobjects.ID) error
	List(ctx context.Context, page, perPage int) ([]*entities.Document, valueobjects.PaginationMeta, error)
	VectorSearch(ctx context.Context, queryEmbedding []float32, limit int) ([]*entities.Document, []float64, error)
	TextSearch(ctx context.Context, query string, limit int) ([]*entities.Document, error)
	SetEmbedding(ctx context.Context, id valueobjects.ID, embedding []float32) error
}
