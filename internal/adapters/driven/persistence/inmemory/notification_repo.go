package inmemory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/rancago/framework/internal/domain/entities"
	"github.com/rancago/framework/internal/domain/valueobjects"
	"github.com/rancago/framework/internal/ports/driven"
)

type InMemoryNotificationRepo struct {
	mu    sync.RWMutex
	items map[string]*entities.Notification
	seq   uint64
}

func NewInMemoryNotificationRepo() driven.NotificationRepository {
	return &InMemoryNotificationRepo{
		items: make(map[string]*entities.Notification),
	}
}

func (r *InMemoryNotificationRepo) nextID() valueobjects.ID {
	r.mu.Lock()
	r.seq++
	id := r.seq
	r.mu.Unlock()
	return valueobjects.NewIDStr(fmt.Sprintf("notif_%d_%d", time.Now().UnixNano(), id))
}

func (r *InMemoryNotificationRepo) clone(n *entities.Notification) *entities.Notification {
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

func (r *InMemoryNotificationRepo) FindByID(_ context.Context, id valueobjects.ID) (*entities.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, ok := r.items[id.String()]
	if !ok {
		return nil, nil
	}
	return r.clone(n), nil
}

func (r *InMemoryNotificationRepo) FindAll(_ context.Context) ([]*entities.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*entities.Notification, 0, len(r.items))
	for _, n := range r.items {
		out = append(out, r.clone(n))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func (r *InMemoryNotificationRepo) FindPaginated(
	_ context.Context,
	page, perPage int,
) ([]*entities.Notification, valueobjects.PaginationMeta, error) {
	all, _ := r.FindAll(context.Background())
	return paginate(all, page, perPage)
}

func (r *InMemoryNotificationRepo) Create(_ context.Context, e *entities.Notification) (*entities.Notification, error) {
	if e.ID.IsZero() {
		e.ID = r.nextID()
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	if e.Data == nil {
		e.Data = make(map[string]string)
	}
	r.mu.Lock()
	r.items[e.ID.String()] = r.clone(e)
	r.mu.Unlock()
	return r.clone(e), nil
}

func (r *InMemoryNotificationRepo) Update(_ context.Context, e *entities.Notification) (*entities.Notification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[e.ID.String()]; !ok {
		return nil, fmt.Errorf("notification not found: %s", e.ID.String())
	}
	r.items[e.ID.String()] = r.clone(e)
	return r.clone(e), nil
}

func (r *InMemoryNotificationRepo) Delete(_ context.Context, id valueobjects.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id.String())
	return nil
}

func (r *InMemoryNotificationRepo) FindBy(_ context.Context, conds map[string]interface{}) ([]*entities.Notification, error) {
	all, _ := r.FindAll(context.Background())
	out := make([]*entities.Notification, 0)
	for _, n := range all {
		match := true
		for k, v := range conds {
			switch k {
			case "user_id", "UserID":
				if n.UserID != fmt.Sprintf("%v", v) {
					match = false
				}
			case "channel", "Channel":
				if string(n.Channel) != fmt.Sprintf("%v", v) {
					match = false
				}
			}
			if !match {
				break
			}
		}
		if match {
			out = append(out, n)
		}
	}
	return out, nil
}

func (r *InMemoryNotificationRepo) FirstOrCreate(
	ctx context.Context,
	conds map[string]interface{},
	defaults map[string]interface{},
) (*entities.Notification, bool, error) {
	existing, err := r.FindBy(ctx, conds)
	if err != nil {
		return nil, false, err
	}
	if len(existing) > 0 {
		return existing[0], false, nil
	}
	ch := entities.ChannelDatabase
	if v, ok := conds["channel"]; ok {
		ch = entities.NotificationChannel(fmt.Sprintf("%v", v))
	}
	userID := ""
	if v, ok := conds["user_id"]; ok {
		userID = fmt.Sprintf("%v", v)
	}
	title := ""
	if v, ok := defaults["title"]; ok {
		title = fmt.Sprintf("%v", v)
	}
	body := ""
	if v, ok := defaults["body"]; ok {
		body = fmt.Sprintf("%v", v)
	}
	n, err := entities.NewNotification(userID, title, body, ch)
	if err != nil {
		return nil, false, err
	}
	created, err := r.Create(ctx, n)
	return created, true, err
}

func (r *InMemoryNotificationRepo) FindByUserID(
	_ context.Context,
	userID string,
	limit, offset int,
) ([]*entities.Notification, valueobjects.PaginationMeta, error) {
	r.mu.RLock()
	var all []*entities.Notification
	for _, n := range r.items {
		if n.UserID == userID || n.Channel == entities.ChannelBroadcast {
			all = append(all, r.clone(n))
		}
	}
	r.mu.RUnlock()
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
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	page := (offset / limit) + 1
	if limit == 0 {
		page = 1
	}
	meta := valueobjects.NewPaginationMeta(page, limit, total)
	if offset > len(all) {
		return []*entities.Notification{}, meta, nil
	}
	return all[offset:end], meta, nil
}

func (r *InMemoryNotificationRepo) MarkRead(
	_ context.Context,
	id valueobjects.ID,
	userID string,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.items[id.String()]
	if !ok {
		return fmt.Errorf("notification not found: %s", id.String())
	}
	if n.UserID != userID && n.Channel != entities.ChannelBroadcast {
		return fmt.Errorf("notification: not owned")
	}
	n.MarkRead()
	return nil
}

func (r *InMemoryNotificationRepo) GetUnreadCount(_ context.Context, userID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var cnt int64
	for _, n := range r.items {
		if (n.UserID == userID || n.Channel == entities.ChannelBroadcast) && !n.Read {
			cnt++
		}
	}
	return cnt, nil
}

func paginate[T any](items []*T, page, perPage int) ([]*T, valueobjects.PaginationMeta, error) {
	if perPage <= 0 {
		perPage = 25
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage
	total := int64(len(items))
	end := offset + perPage
	if end > len(items) {
		end = len(items)
	}
	meta := valueobjects.NewPaginationMeta(page, perPage, total)
	if offset > len(items) {
		return []*T{}, meta, nil
	}
	return items[offset:end], meta, nil
}
