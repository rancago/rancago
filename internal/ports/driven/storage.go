package driven

import (
	"context"
	"io"
	"time"
)

type StorageDriver interface {
	Put(ctx context.Context, path string, content io.Reader, opts ...StorageOption) error
	Get(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
	Exists(ctx context.Context, path string) (bool, error)
	Size(ctx context.Context, path string) (int64, error)
	LastModified(ctx context.Context, path string) (time.Time, error)
	Copy(ctx context.Context, from, to string) error
	Move(ctx context.Context, from, to string) error
	List(ctx context.Context, prefix string) ([]StorageFile, error)
	URL(ctx context.Context, path string) (string, error)
	TemporaryURL(ctx context.Context, path string, expires time.Duration) (string, error)
	Name() string
}

type StorageFile struct {
	Path         string
	Size         int64
	LastModified time.Time
	IsDir        bool
	MimeType     string
}

type StorageOptions struct {
	ContentType string
	ACL         string
	Metadata    map[string]string
	Visibility  string
}

type StorageOption func(*StorageOptions)

func WithContentType(ct string) StorageOption {
	return func(o *StorageOptions) { o.ContentType = ct }
}

func WithACL(acl string) StorageOption {
	return func(o *StorageOptions) { o.ACL = acl }
}

func WithMetadata(md map[string]string) StorageOption {
	return func(o *StorageOptions) { o.Metadata = md }
}

func WithVisibility(v string) StorageOption {
	return func(o *StorageOptions) { o.Visibility = v }
}

func ApplyStorageOptions(opts []StorageOption) *StorageOptions {
	o := &StorageOptions{
		Visibility: "private",
		Metadata:   make(map[string]string),
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

type StorageManagerPort interface {
	Disk(name string) (StorageDriver, error)
	DefaultDisk() StorageDriver
	Register(name string, driver StorageDriver)
}
