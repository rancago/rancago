package Contracts

import (
	"context"
	"time"
)

// NotificationChannel represents the delivery channel for a notification.
type NotificationChannel string

const (
	ChannelEmail     NotificationChannel = "email"
	ChannelPush      NotificationChannel = "push"
	ChannelSMS       NotificationChannel = "sms"
	ChannelBroadcast NotificationChannel = "broadcast"
	ChannelDatabase  NotificationChannel = "database"
)

// Notification is the transport-agnostic notification entity used across all adapters.
type Notification struct {
	ID        string
	UserID    string
	Title     string
	Body      string
	Channel   NotificationChannel
	Data      map[string]string
	Read      bool
	ReadAt    *time.Time
	CreatedAt time.Time
}

// NotificationService is the transport-agnostic business contract.
// It must NOT import net/http, grpc, or websocket packages.
// The same implementation is called from REST, gRPC, WebSocket, and CLI adapters.
type NotificationService interface {
	// Send sends a notification to a specific user.
	Send(ctx context.Context, n *Notification) (*Notification, error)
	// Broadcast sends a notification to all connected WebSocket clients.
	Broadcast(ctx context.Context, title, body string, data map[string]string) error
	// List returns paginated notifications for a user (includes broadcast).
	List(ctx context.Context, userID string, limit, offset int) ([]*Notification, PaginationMeta, error)
	// MarkRead marks a notification as read.
	MarkRead(ctx context.Context, id, userID string) error
	// GetUnreadCount returns the unread notification count for a user.
	GetUnreadCount(ctx context.Context, userID string) (int64, error)
}
