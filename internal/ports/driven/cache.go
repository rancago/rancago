package driven

import (
	"context"
	"time"
)

type CachePort interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Exists(ctx context.Context, key string) (bool, error)
	Publish(ctx context.Context, channel string, message []byte) error
	Subscribe(ctx context.Context, channel string, handler func(message []byte)) error
	Connect() error
	Close() error
}
