// Package Repositories contains the data-access layer for Rancago Framework.
// Implementations here satisfy the Contracts.Repository[T] interface.
// Swap out the in-memory implementations for GORM-backed ones once
// gorm.io/gorm is added to go.mod.
package Repositories

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rancago/framework/app/Contracts"
)

// ---- Generic in-memory store ----

type record struct {
	id        string
	data      interface{}
	createdAt time.Time
	updatedAt time.Time
}

// MemoryStore is a simple thread-safe in-memory key-value store.
// It backs the in-memory repository implementations below.
type MemoryStore struct {
	mu   sync.RWMutex
	rows map[string]*record
	seq  uint64
}

// NewMemoryStore creates a new MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{rows: make(map[string]*record)}
}

// NextID returns the next auto-incremented string ID.
func (s *MemoryStore) NextID() string {
	s.mu.Lock()
	s.seq++
	id := fmt.Sprintf("%d", s.seq)
	s.mu.Unlock()
	return id
}

// Set stores a value.
func (s *MemoryStore) Set(id string, v interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rows[id]; ok {
		s.rows[id].data = v
		s.rows[id].updatedAt = time.Now()
	} else {
		now := time.Now()
		s.rows[id] = &record{id: id, data: v, createdAt: now, updatedAt: now}
	}
}

// Get retrieves a value.
func (s *MemoryStore) Get(id string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rows[id]
	if !ok {
		return nil, false
	}
	return r.data, true
}

// Delete removes a value.
func (s *MemoryStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rows, id)
}

// All returns all stored values in insertion order.
func (s *MemoryStore) All() []interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]interface{}, 0, len(s.rows))
	for _, r := range s.rows {
		out = append(out, r.data)
	}
	return out
}

// Count returns the number of stored records.
func (s *MemoryStore) Count() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return int64(len(s.rows))
}

// paginate is a generic helper that slices a slice and returns PaginationMeta.
func paginate[T any](items []*T, page, perPage int) ([]*T, Contracts.PaginationMeta, error) {
	if perPage <= 0 {
		perPage = 25
	}
	if page < 1 {
		page = 1
	}
	total := int64(len(items))
	offset := (page - 1) * perPage
	end := offset + perPage
	if end > len(items) {
		end = len(items)
	}
	totalPages := (total + int64(perPage) - 1) / int64(perPage)
	meta := Contracts.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    int64(offset+perPage) < total,
		HasPrev:    offset > 0,
	}
	if offset >= len(items) {
		return []*T{}, meta, nil
	}
	return items[offset:end], meta, nil
}

// ---- Repository placeholder (GORM implementations go here) ----

// BaseRepository is a placeholder that documents where GORM implementations live.
// Example usage once gorm.io/gorm is added:
//
//	type UserRepository struct {
//	    db *gorm.DB
//	}
//
//	func (r *UserRepository) FindByID(ctx context.Context, id interface{}) (*Models.User, error) {
//	    var u Models.User
//	    return &u, r.db.WithContext(ctx).First(&u, id).Error
//	}
type BaseRepository struct{}

// Ensure context is used even in stub methods.
var _ context.Context = context.Background()
