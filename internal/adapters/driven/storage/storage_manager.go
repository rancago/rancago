package storage

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/rancago/framework/internal/kernel"
	"github.com/rancago/framework/internal/ports/driven"
)

type memoryFile struct {
	content    []byte
	modifiedAt time.Time
}

type MemoryStorageAdapter struct {
	name  string
	mu    sync.RWMutex
	files map[string]*memoryFile
}

func NewMemoryStorageAdapter(name string) driven.StorageDriver {
	return &MemoryStorageAdapter{
		name:  name,
		files: make(map[string]*memoryFile),
	}
}

func (m *MemoryStorageAdapter) Name() string { return m.name }

func (m *MemoryStorageAdapter) Put(_ context.Context, path string, content io.Reader, _ ...driven.StorageOption) error {
	data, err := io.ReadAll(content)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[path] = &memoryFile{
		content:    data,
		modifiedAt: time.Now(),
	}
	return nil
}

type memReadCloser struct {
	data   []byte
	offset int
}

func (r *memReadCloser) Read(p []byte) (int, error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

func (r *memReadCloser) Close() error { return nil }

func (m *MemoryStorageAdapter) Get(_ context.Context, path string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	f, ok := m.files[path]
	if !ok {
		return nil, fmt.Errorf("storage[%s]: file not found %s", m.name, path)
	}
	return &memReadCloser{data: append([]byte(nil), f.content...)}, nil
}

func (m *MemoryStorageAdapter) Delete(_ context.Context, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.files, path)
	return nil
}

func (m *MemoryStorageAdapter) Exists(_ context.Context, path string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.files[path]
	return ok, nil
}

func (m *MemoryStorageAdapter) Size(_ context.Context, path string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	f, ok := m.files[path]
	if !ok {
		return 0, fmt.Errorf("storage[%s]: file not found %s", m.name, path)
	}
	return int64(len(f.content)), nil
}

func (m *MemoryStorageAdapter) LastModified(_ context.Context, path string) (time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	f, ok := m.files[path]
	if !ok {
		return time.Time{}, fmt.Errorf("storage[%s]: file not found %s", m.name, path)
	}
	return f.modifiedAt, nil
}

func (m *MemoryStorageAdapter) Copy(ctx context.Context, from, to string) error {
	rc, err := m.Get(ctx, from)
	if err != nil {
		return err
	}
	defer rc.Close()
	return m.Put(ctx, to, rc)
}

func (m *MemoryStorageAdapter) Move(ctx context.Context, from, to string) error {
	if err := m.Copy(ctx, from, to); err != nil {
		return err
	}
	return m.Delete(ctx, from)
}

func (m *MemoryStorageAdapter) List(_ context.Context, prefix string) ([]driven.StorageFile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []driven.StorageFile
	for p, f := range m.files {
		if prefix == "" || hasPrefix(p, prefix) {
			out = append(out, driven.StorageFile{
				Path:         p,
				Size:         int64(len(f.content)),
				LastModified: f.modifiedAt,
			})
		}
	}
	return out, nil
}

func (m *MemoryStorageAdapter) URL(_ context.Context, path string) (string, error) {
	return fmt.Sprintf("memory://%s/%s", m.name, path), nil
}

func (m *MemoryStorageAdapter) TemporaryURL(_ context.Context, path string, _ time.Duration) (string, error) {
	return fmt.Sprintf("memory://%s/%s?temp=1", m.name, path), nil
}

func hasPrefix(s, p string) bool {
	if len(p) > len(s) {
		return false
	}
	for i := 0; i < len(p); i++ {
		if s[i] != p[i] {
			return false
		}
	}
	return true
}

type StorageManagerAdapter struct {
	cfg      *kernel.StorageConfig
	mu       sync.RWMutex
	disks    map[string]driven.StorageDriver
}

func NewStorageManagerAdapter(cfg *kernel.StorageConfig) driven.StorageManagerPort {
	mgr := &StorageManagerAdapter{
		cfg:   cfg,
		disks: make(map[string]driven.StorageDriver),
	}
	for name := range cfg.Disks {
		mgr.disks[name] = NewMemoryStorageAdapter(name)
	}
	return mgr
}

func (m *StorageManagerAdapter) Disk(name string) (driven.StorageDriver, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.disks[name]
	if !ok {
		return nil, fmt.Errorf("storage disk %s not configured", name)
	}
	return d, nil
}

func (m *StorageManagerAdapter) DefaultDisk() driven.StorageDriver {
	d, _ := m.Disk(m.cfg.Default)
	if d == nil {
		return NewMemoryStorageAdapter("default")
	}
	return d
}

func (m *StorageManagerAdapter) Register(name string, driver driven.StorageDriver) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.disks[name] = driver
}
