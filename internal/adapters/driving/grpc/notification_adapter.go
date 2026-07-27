package grpc

import (
	"encoding/json"

	"github.com/rancago/framework/internal/domain/entities"
	"github.com/rancago/framework/internal/domain/valueobjects"
	"github.com/rancago/framework/internal/ports/driving"
)

type GRPCServer interface{}

type GRPCNotificationAdapter struct {
	uc driving.NotificationUseCase
}

func NewGRPCAdapter(uc driving.NotificationUseCase) *GRPCNotificationAdapter {
	return &GRPCNotificationAdapter{uc: uc}
}

func (a *GRPCNotificationAdapter) RegisterGRPC(s GRPCServer) {
}

type GRPCNotification struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	Channel   string            `json:"channel"`
	Data      map[string]string `json:"data"`
	Read      bool              `json:"read"`
	ReadAt    string            `json:"read_at,omitempty"`
	CreatedAt string            `json:"created_at"`
}

func toGRPCNotification(n *entities.Notification) *GRPCNotification {
	gn := &GRPCNotification{
		ID:        n.ID.String(),
		UserID:    n.UserID,
		Title:     n.Title,
		Body:      n.Body,
		Channel:   string(n.Channel),
		Data:      n.Data,
		Read:      n.Read,
		CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if n.ReadAt != nil {
		gn.ReadAt = n.ReadAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return gn
}

func (a *GRPCNotificationAdapter) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func parseGRPCID(s string) valueobjects.ID {
	return valueobjects.NewIDStr(s)
}
