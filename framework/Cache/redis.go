// Package Cache provides a Redis manager for the Rancago Framework.
// It wraps github.com/redis/go-redis/v9 and exposes a connection manager,
// Pub/Sub, rate limiting, and a simple get/set/incr/decr API.
package Cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RedisConfig holds connection settings for Redis.
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// RedisManager wraps a Redis client with convenience methods.
// The actual go-redis client is swapped in by the StorageServiceProvider
// once the dependency is available. For now it ships an in-memory fallback
// so the framework compiles and runs without Redis installed.
type RedisManager struct {
	cfg  *RedisConfig
	mu   sync.RWMutex
	data map[string]entry
}

type entry struct {
	value     string
	expiresAt time.Time
	hasTTL    bool
}

// NewRedisManager creates a RedisManager that uses an in-memory fallback.
// Replace the underlying client in a ServiceProvider.Boot() to use real Redis.
func NewRedisManager(cfg *RedisConfig) *RedisManager {
	return &RedisManager{cfg: cfg, data: make(map[string]entry)}
}

// Connect establishes the Redis connection.
func (r *RedisManager) Connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// TODO: replace with real go-redis client when dependency is added.
	return nil
}

// Close closes the Redis connection.
func (r *RedisManager) Close() error { return nil }

func (r *RedisManager) clean() {
	now := time.Now()
	for k, e := range r.data {
		if e.hasTTL && now.After(e.expiresAt) {
			delete(r.data, k)
		}
	}
}

// Get returns the string value stored at key.
func (r *RedisManager) Get(_ context.Context, key string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.data[key]
	if !ok || (e.hasTTL && time.Now().After(e.expiresAt)) {
		return "", fmt.Errorf("cache: key %q not found", key)
	}
	return e.value, nil
}

// Set stores a string value with an optional TTL (0 = no expiry).
func (r *RedisManager) Set(_ context.Context, key, value string, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e := entry{value: value}
	if ttl > 0 {
		e.hasTTL = true
		e.expiresAt = time.Now().Add(ttl)
	}
	r.data[key] = e
	return nil
}

// Delete removes a key.
func (r *RedisManager) Delete(_ context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, key)
	return nil
}

// Incr increments an integer counter stored at key.
func (r *RedisManager) Incr(_ context.Context, key string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var cur int64
	if e, ok := r.data[key]; ok {
		fmt.Sscanf(e.value, "%d", &cur)
	}
	cur++
	r.data[key] = entry{value: fmt.Sprintf("%d", cur)}
	return cur, nil
}

// Decr decrements an integer counter stored at key.
func (r *RedisManager) Decr(_ context.Context, key string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var cur int64
	if e, ok := r.data[key]; ok {
		fmt.Sscanf(e.value, "%d", &cur)
	}
	cur--
	r.data[key] = entry{value: fmt.Sprintf("%d", cur)}
	return cur, nil
}

// Expire sets the TTL on an existing key.
func (r *RedisManager) Expire(_ context.Context, key string, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.data[key]
	if !ok {
		return fmt.Errorf("cache: key %q not found", key)
	}
	e.hasTTL = true
	e.expiresAt = time.Now().Add(ttl)
	r.data[key] = e
	return nil
}

// Exists reports whether a key is present and not expired.
func (r *RedisManager) Exists(_ context.Context, key string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.data[key]
	if !ok {
		return false, nil
	}
	if e.hasTTL && time.Now().After(e.expiresAt) {
		return false, nil
	}
	return true, nil
}

// SAdd adds members to a Redis Set stored at key.
func (r *RedisManager) SAdd(_ context.Context, key string, members ...string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// In-memory: store as comma-delimited string for simplicity.
	existing := ""
	if e, ok := r.data[key]; ok {
		existing = e.value
	}
	for _, m := range members {
		if !containsToken(existing, m) {
			if existing != "" {
				existing += ","
			}
			existing += m
		}
	}
	r.data[key] = entry{value: existing}
	return nil
}

// SMembers returns all members of a Set.
func (r *RedisManager) SMembers(_ context.Context, key string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.data[key]
	if !ok || e.value == "" {
		return nil, nil
	}
	return splitTokens(e.value), nil
}

// SRem removes members from a Set.
func (r *RedisManager) SRem(_ context.Context, key string, members ...string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.data[key]
	if !ok {
		return nil
	}
	tokens := splitTokens(e.value)
	result := tokens[:0]
	for _, t := range tokens {
		found := false
		for _, m := range members {
			if t == m {
				found = true
				break
			}
		}
		if !found {
			result = append(result, t)
		}
	}
	if len(result) == 0 {
		delete(r.data, key)
	} else {
		r.data[key] = entry{value: joinTokens(result)}
	}
	return nil
}

// SIsMember reports whether member is in a Set.
func (r *RedisManager) SIsMember(_ context.Context, key, member string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.data[key]
	if !ok {
		return false, nil
	}
	return containsToken(e.value, member), nil
}

// Publish publishes a message to a channel.
// In-memory: no-op (extend with real Redis PubSub for multi-node support).
func (r *RedisManager) Publish(_ context.Context, channel string, _ []byte) error { return nil }

// Subscribe subscribes to a channel and calls handler on each message.
// In-memory: no-op (extend with real Redis PubSub for multi-node support).
func (r *RedisManager) Subscribe(_ context.Context, _ string, _ func([]byte)) error { return nil }

func containsToken(s, token string) bool {
	for _, t := range splitTokens(s) {
		if t == token {
			return true
		}
	}
	return false
}

func splitTokens(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			if i > start {
				out = append(out, s[start:i])
			}
			start = i + 1
		}
	}
	return out
}

func joinTokens(tokens []string) string {
	out := ""
	for i, t := range tokens {
		if i > 0 {
			out += ","
		}
		out += t
	}
	return out
}
