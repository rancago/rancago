package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rancago/framework/internal/domain/entities"
	"github.com/rancago/framework/internal/domain/valueobjects"
	"github.com/rancago/framework/internal/ports/driven"
	"github.com/rancago/framework/internal/ports/driving"
)

type NotificationInteractor struct {
	notifRepo driven.NotificationRepository
	cache     driven.CachePort
	ws        driven.WebSocketPort
}

func NewNotificationInteractor(
	repo driven.NotificationRepository,
	cache driven.CachePort,
	ws driven.WebSocketPort,
) driving.NotificationUseCase {
	return &NotificationInteractor{
		notifRepo: repo,
		cache:     cache,
		ws:        ws,
	}
}

func (uc *NotificationInteractor) Send(
	ctx context.Context,
	userID, title, body string,
	channel entities.NotificationChannel,
	data map[string]string,
) (*entities.Notification, error) {
	n, err := entities.NewNotification(userID, title, body, channel)
	if err != nil {
		return nil, fmt.Errorf("notification send: %w", err)
	}
	if data != nil {
		n.Data = data
	}
	saved, err := uc.notifRepo.Create(ctx, n)
	if err != nil {
		return nil, fmt.Errorf("notification save: %w", err)
	}
	if uc.cache != nil {
		cacheKey := fmt.Sprintf("notif:unread:%s", userID)
		_, _ = uc.cache.Incr(ctx, cacheKey)
		_ = uc.cache.Expire(ctx, cacheKey, 7*24*time.Hour)
	}
	if uc.ws != nil {
		payload := map[string]interface{}{
			"id":         saved.ID.String(),
			"user_id":    saved.UserID,
			"title":      saved.Title,
			"body":       saved.Body,
			"channel":    string(saved.Channel),
			"data":       saved.Data,
			"created_at": saved.CreatedAt.Unix(),
		}
		msg := driven.WebSocketMessage{
			Type:      "notification:new",
			Channel:   fmt.Sprintf("user:%s", saved.UserID),
			Payload:   payload,
			Timestamp: saved.CreatedAt.Unix(),
		}
		raw, _ := json.Marshal(msg)
		uc.ws.SendDirect(userID, raw)
		uc.ws.PublishChannel(msg.Channel, raw)
	}
	return saved, nil
}

func (uc *NotificationInteractor) Broadcast(
	ctx context.Context,
	title, body string,
	data map[string]string,
) error {
	n, err := entities.NewNotification("", title, body, entities.ChannelBroadcast)
	if err != nil {
		return fmt.Errorf("notification broadcast: %w", err)
	}
	if data != nil {
		n.Data = data
	}
	if _, err := uc.notifRepo.Create(ctx, n); err != nil {
		return fmt.Errorf("notification broadcast save: %w", err)
	}
	if uc.ws != nil {
		msg := driven.WebSocketMessage{
			Type:      "notification:broadcast",
			Channel:   "broadcast",
			Payload:   n,
			Timestamp: n.CreatedAt.Unix(),
		}
		raw, _ := json.Marshal(msg)
		uc.ws.Broadcast(raw)
	}
	return nil
}

func (uc *NotificationInteractor) ListUserNotifications(
	ctx context.Context,
	userID string,
	page, perPage int,
) ([]*entities.Notification, valueobjects.PaginationMeta, error) {
	return uc.notifRepo.FindByUserID(ctx, userID, perPage, (page-1)*perPage)
}

func (uc *NotificationInteractor) MarkRead(
	ctx context.Context,
	id valueobjects.ID,
	userID string,
) error {
	if err := uc.notifRepo.MarkRead(ctx, id, userID); err != nil {
		return err
	}
	if uc.cache != nil {
		cacheKey := fmt.Sprintf("notif:unread:%s", userID)
		_, _ = uc.cache.Decr(ctx, cacheKey)
	}
	return nil
}

func (uc *NotificationInteractor) GetUnreadCount(
	ctx context.Context,
	userID string,
) (int64, error) {
	if uc.cache != nil {
		cacheKey := fmt.Sprintf("notif:unread:%s", userID)
		str, err := uc.cache.Get(ctx, cacheKey)
		if err == nil {
			var cnt int64
			_, scanErr := fmt.Sscanf(str, "%d", &cnt)
			if scanErr == nil {
				return cnt, nil
			}
		}
	}
	return uc.notifRepo.GetUnreadCount(ctx, userID)
}
