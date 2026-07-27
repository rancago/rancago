package entities

import (
	"fmt"
	"time"

	"github.com/rancago/framework/internal/domain/valueobjects"
)

type NotificationChannel string

const (
	ChannelEmail     NotificationChannel = "email"
	ChannelPush      NotificationChannel = "push"
	ChannelSMS       NotificationChannel = "sms"
	ChannelBroadcast NotificationChannel = "broadcast"
	ChannelDatabase  NotificationChannel = "database"
)

type Notification struct {
	ID        valueobjects.ID
	UserID    string
	Title     string
	Body      string
	Channel   NotificationChannel
	Data      map[string]string
	Read      bool
	ReadAt    *time.Time
	CreatedAt time.Time
}

func NewNotification(userID, title, body string, ch NotificationChannel) (*Notification, error) {
	if userID == "" && ch != ChannelBroadcast {
		return nil, fmt.Errorf("notification: user_id is required for channel %s", ch)
	}
	if title == "" && body == "" {
		return nil, fmt.Errorf("notification: title and body cannot both be empty")
	}
	now := time.Now()
	return &Notification{
		UserID:    userID,
		Title:     title,
		Body:      body,
		Channel:   ch,
		Data:      make(map[string]string),
		Read:      false,
		CreatedAt: now,
	}, nil
}

func (n *Notification) MarkRead() {
	if n.Read {
		return
	}
	now := time.Now()
	n.Read = true
	n.ReadAt = &now
}

func (n *Notification) IsVisibleTo(userID string) bool {
	if n.Channel == ChannelBroadcast {
		return true
	}
	return n.UserID == userID
}
