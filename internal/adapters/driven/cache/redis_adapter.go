package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rancago/framework/internal/kernel"
	"github.com/rancago/framework/internal/ports/driven"
)

type RedisManagerAdapter struct {
	cfg      *kernel.RedisConfig
	mu       sync.RWMutex
	data     map[string]cacheEntry
	connected bool
}

type cacheEntry struct {
	value      string
	expiresAt  time.Time
	hasExpiry  bool
}

func NewRedisManagerAdapter(cfg *kernel.RedisConfig) driven.CachePort {
	return &RedisManagerAdapter{
		cfg:  cfg,
		data: make(map[string]cacheEntry),
	}
}

func (r *RedisManagerAdapter) Connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connected = true
	return nil
}

func (r *RedisManagerAdapter) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connected = false
	return nil
}

func (r *RedisManagerAdapter) cleanExpired() {
	now := time.Now()
	for k, v := range r.data {
		if v.hasExpiry && now.After(v.expiresAt) {
			delete(r.data, k)
		}
	}
}

func (r *RedisManagerAdapter) Get(_ context.Context, key string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	r.cleanExpired()
	e, ok := r.data[key]
	if !ok {
		return "", fmt.Errorf("redis: key %s not found", key)
	}
	if e.hasExpiry && time.Now().After(e.expiresAt) {
		return "", fmt.Errorf("redis: key %s expired", key)
	}
	return e.value, nil
}

func (r *RedisManagerAdapter) Set(_ context.Context, key, value string, expiration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry := cacheEntry{value: value}
	if expiration > 0 {
		entry.hasExpiry = true
		entry.expiresAt = time.Now().Add(expiration)
	}
	r.data[key] = entry
	return nil
}

func (r *RedisManagerAdapter) Delete(_ context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, key)
	return nil
}

func (r *RedisManagerAdapter) Incr(_ context.Context, key string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var curr int64 = 0
	if e, ok := r.data[key]; ok {
		fmt.Sscanf(e.value, "%d", &curr)
	}
	curr++
	r.data[key] = cacheEntry{value: fmt.Sprintf("%d", curr)}
	return curr, nil
}

func (r *RedisManagerAdapter) Decr(_ context.Context, key string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var curr int64 = 0
	if e, ok := r.data[key]; ok {
		fmt.Sscanf(e.value, "%d", &curr)
	}
	curr--
	r.data[key] = cacheEntry{value: fmt.Sprintf("%d", curr)}
	return curr, nil
}

func (r *RedisManagerAdapter) Expire(_ context.Context, key string, expiration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.data[key]
	if !ok {
		return fmt.Errorf("redis: key %s not found", key)
	}
	e.hasExpiry = true
	e.expiresAt = time.Now().Add(expiration)
	r.data[key] = e
	return nil
}

func (r *RedisManagerAdapter) Exists(_ context.Context, key string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	r.cleanExpired()
	_, ok := r.data[key]
	return ok, nil
}

func (r *RedisManagerAdapter) Publish(_ context.Context, _ string, _ []byte) error { return nil }
func (r *RedisManagerAdapter) Subscribe(_ context.Context, _ string, _ func([]byte)) error { return nil }
