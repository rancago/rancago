// Package Services contains business logic implementations for Rancago Framework.
// Services are transport-agnostic: they must NOT import net/http, grpc, or websocket packages.
// The same service is consumed by REST, gRPC, WebSocket, and CLI adapters.
package Services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/rancago/framework/app/Contracts"
	"github.com/rancago/framework/framework/Cache"
	"github.com/rancago/framework/framework/WebSocket"
)

// NotificationService is the transport-agnostic notification service implementation.
// Backing store: in-memory map (swap out for a real Repository in production).
// Cache: Redis-backed unread counter (decrement on read, increment on send).
// Realtime: WebSocket Hub push on send and broadcast.
type NotificationService struct {
	mu    sync.RWMutex
	items map[string]*Contracts.Notification
	seq   uint64
	redis *Cache.RedisManager
	hub   *WebSocket.Hub
}

// Ensure NotificationService satisfies the contract at compile time.
var _ Contracts.NotificationService = (*NotificationService)(nil)

// NewNotificationService creates a NotificationService.
// Passing nil for redis or hub is safe — those features are gracefully skipped.
func NewNotificationService(redis *Cache.RedisManager, hub *WebSocket.Hub) *NotificationService {
	return &NotificationService{
		items: make(map[string]*Contracts.Notification),
		redis: redis,
		hub:   hub,
	}
}

func (s *NotificationService) nextID() string {
	s.mu.Lock()
	s.seq++
	id := s.seq
	s.mu.Unlock()
	return fmt.Sprintf("notif_%d_%d", time.Now().UnixNano(), id)
}

func (s *NotificationService) clone(n *Contracts.Notification) *Contracts.Notification {
	cp := *n
	if n.Data != nil {
		cp.Data = make(map[string]string, len(n.Data))
		for k, v := range n.Data {
			cp.Data[k] = v
		}
	}
	if n.ReadAt != nil {
		t := *n.ReadAt
		cp.ReadAt = &t
	}
	return &cp
}

// Send persists a notification, increments the Redis unread counter,
// and pushes the notification to the target user's WebSocket channel.
func (s *NotificationService) Send(ctx context.Context, n *Contracts.Notification) (*Contracts.Notification, error) {
	if n == nil {
		return nil, fmt.Errorf("notification: nil notification")
	}
	if n.UserID == "" && n.Channel != Contracts.ChannelBroadcast {
		return nil, fmt.Errorf("notification: user_id required for channel %s", n.Channel)
	}
	if n.Title == "" && n.Body == "" {
		return nil, fmt.Errorf("notification: title and body cannot both be empty")
	}
	n.ID = s.nextID()
	n.CreatedAt = time.Now()
	if n.Data == nil {
		n.Data = make(map[string]string)
	}
	if n.Channel == "" {
		n.Channel = Contracts.ChannelDatabase
	}
	saved := s.clone(n)
	s.mu.Lock()
	s.items[n.ID] = saved
	s.mu.Unlock()

	// Increment Redis unread counter.
	if s.redis != nil && n.UserID != "" {
		cacheKey := fmt.Sprintf("notif:unread:%s", n.UserID)
		_, _ = s.redis.Incr(ctx, cacheKey)
		_ = s.redis.Expire(ctx, cacheKey, 7*24*time.Hour)
	}

	// Push via WebSocket.
	if s.hub != nil {
		payload := map[string]interface{}{
			"id":         saved.ID,
			"user_id":    saved.UserID,
			"title":      saved.Title,
			"body":       saved.Body,
			"channel":    string(saved.Channel),
			"created_at": saved.CreatedAt.Unix(),
		}
		raw, _ := json.Marshal(WebSocket.Message{
			Type:      "notification:new",
			Channel:   "user:" + saved.UserID,
			Payload:   payload,
			Timestamp: saved.CreatedAt.Unix(),
		})
		s.hub.SendDirect(saved.UserID, raw)
		s.hub.PublishChannel("user:"+saved.UserID, raw)
	}

	return s.clone(saved), nil
}

// Broadcast sends a message to all connected WebSocket clients.
func (s *NotificationService) Broadcast(ctx context.Context, title, body string, data map[string]string) error {
	n := &Contracts.Notification{
		ID:        s.nextID(),
		UserID:    "",
		Title:     title,
		Body:      body,
		Channel:   Contracts.ChannelBroadcast,
		Data:      data,
		CreatedAt: time.Now(),
	}
	if n.Data == nil {
		n.Data = make(map[string]string)
	}
	s.mu.Lock()
	s.items[n.ID] = s.clone(n)
	s.mu.Unlock()

	if s.hub != nil {
		raw, _ := json.Marshal(WebSocket.Message{
			Type:      "notification:broadcast",
			Channel:   "broadcast",
			Payload:   n,
			Timestamp: n.CreatedAt.Unix(),
		})
		s.hub.Broadcast(raw)
	}
	return nil
}

// List returns paginated notifications for userID (includes broadcast channel).
func (s *NotificationService) List(_ context.Context, userID string, limit, offset int) ([]*Contracts.Notification, Contracts.PaginationMeta, error) {
	s.mu.RLock()
	var all []*Contracts.Notification
	for _, n := range s.items {
		if n.UserID == userID || n.Channel == Contracts.ChannelBroadcast {
			all = append(all, s.clone(n))
		}
	}
	s.mu.RUnlock()

	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})

	if limit <= 0 {
		limit = 25
	}
	if offset < 0 {
		offset = 0
	}
	total := int64(len(all))
	page := offset/limit + 1
	meta := Contracts.PaginationMeta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: (total + int64(limit) - 1) / int64(limit),
		HasNext:    int64(offset+limit) < total,
		HasPrev:    offset > 0,
	}
	end := offset + limit
	if offset >= len(all) {
		return []*Contracts.Notification{}, meta, nil
	}
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], meta, nil
}

// MarkRead marks a notification as read and decrements the Redis unread counter.
func (s *NotificationService) MarkRead(ctx context.Context, id, userID string) error {
	s.mu.Lock()
	n, ok := s.items[id]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("notification: %s not found", id)
	}
	if n.UserID != userID && n.Channel != Contracts.ChannelBroadcast {
		s.mu.Unlock()
		return fmt.Errorf("notification: not owned by user %s", userID)
	}
	if !n.Read {
		now := time.Now()
		n.Read = true
		n.ReadAt = &now
		if s.redis != nil {
			_, _ = s.redis.Decr(ctx, fmt.Sprintf("notif:unread:%s", userID))
		}
	}
	s.mu.Unlock()
	return nil
}

// GetUnreadCount returns the unread notification count for userID.
// Redis-first with in-memory fallback.
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	if s.redis != nil {
		key := fmt.Sprintf("notif:unread:%s", userID)
		str, err := s.redis.Get(ctx, key)
		if err == nil {
			var cnt int64
			_, scanErr := fmt.Sscanf(str, "%d", &cnt)
			if scanErr == nil {
				return cnt, nil
			}
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var cnt int64
	for _, n := range s.items {
		if (n.UserID == userID || n.Channel == Contracts.ChannelBroadcast) && !n.Read {
			cnt++
		}
	}
	return cnt, nil
}
