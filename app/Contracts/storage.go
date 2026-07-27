package Contracts

import (
	"context"
	"io"
	"time"
)

// StorageFile represents a file entry from a storage driver listing.
type StorageFile struct {
	Path         string
	Size         int64
	LastModified time.Time
	IsDir        bool
	MimeType     string
}

// StorageOptions holds resolved options for a storage operation.
type StorageOptions struct {
	ContentType string
	ACL         string
	Metadata    map[string]string
	Visibility  string
}

// StorageOption is a functional option for storage operations.
type StorageOption func(*StorageOptions)

// WithContentType sets the MIME content-type for a storage operation.
func WithContentType(ct string) StorageOption {
	return func(o *StorageOptions) { o.ContentType = ct }
}

// WithACL sets the access-control list (e.g. "public-read") for a storage operation.
func WithACL(acl string) StorageOption {
	return func(o *StorageOptions) { o.ACL = acl }
}

// WithMetadata attaches arbitrary key-value metadata to a storage operation.
func WithMetadata(md map[string]string) StorageOption {
	return func(o *StorageOptions) { o.Metadata = md }
}

// WithVisibility sets the file visibility ("public" | "private").
func WithVisibility(v string) StorageOption {
	return func(o *StorageOptions) { o.Visibility = v }
}

// ApplyStorageOptions merges functional options into a concrete StorageOptions struct.
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

// StorageDriver is the core abstraction for any file storage backend.
// Implementations must be interchangeable (LSP compliant).
type StorageDriver interface {
	// CRUD
	Put(ctx context.Context, path string, content io.Reader, opts ...StorageOption) error
	Get(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
	Exists(ctx context.Context, path string) (bool, error)
	Size(ctx context.Context, path string) (int64, error)
	LastModified(ctx context.Context, path string) (time.Time, error)
	// Bulk
	Copy(ctx context.Context, from, to string) error
	Move(ctx context.Context, from, to string) error
	List(ctx context.Context, prefix string) ([]StorageFile, error)
	// URLs
	URL(ctx context.Context, path string) (string, error)
	TemporaryURL(ctx context.Context, path string, expires time.Duration) (string, error)
	// Meta
	Name() string
}

// StorageManager manages multiple named disks and exposes a factory/registry API.
type StorageManager interface {
	// Disk returns the driver for the named disk (lazy-initialised).
	Disk(name ...string) (StorageDriver, error)
	// DefaultDisk returns the driver for the configured default disk.
	DefaultDisk() StorageDriver
	// Register adds or overrides a named driver instance.
	Register(name string, driver StorageDriver)
	// RegisterFactory adds an OCP-compliant driver factory for a driver type.
	RegisterFactory(driverType string, factory func(cfg StorageDiskConfig) (StorageDriver, error))
	// Proxy returns a StorageDriver that proxies to the named disk (for constructor injection).
	Proxy(name string) StorageDriver
	// AvailableDisks returns the list of configured disk names.
	AvailableDisks() []string
}

// StorageDiskConfig holds configuration for a single named disk.
type StorageDiskConfig struct {
	Driver      string
	Endpoint    string
	Bucket      string
	Region      string
	AccessKey   string
	SecretKey   string
	UseSSL      bool
	Credentials string
	FolderID    string
}
