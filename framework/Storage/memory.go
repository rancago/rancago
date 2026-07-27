package Storage

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/rancago/framework/app/Contracts"
)

// MemoryDriver is an in-memory StorageDriver implementation.
// It satisfies Contracts.StorageDriver and is the built-in fallback driver.
type MemoryDriver struct {
	diskName string
	mu       sync.RWMutex
	files    map[string]*memFile
}

type memFile struct {
	data      []byte
	modified  time.Time
	mimeType  string
}

// NewMemoryDriver creates a new MemoryDriver named diskName.
func NewMemoryDriver(diskName string) *MemoryDriver {
	return &MemoryDriver{diskName: diskName, files: make(map[string]*memFile)}
}

func (m *MemoryDriver) Name() string { return m.diskName }

func (m *MemoryDriver) Put(_ context.Context, path string, content io.Reader, opts ...Contracts.StorageOption) error {
	o := Contracts.ApplyStorageOptions(opts)
	data, err := io.ReadAll(content)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[path] = &memFile{data: data, modified: time.Now(), mimeType: o.ContentType}
	return nil
}

func (m *MemoryDriver) Get(_ context.Context, path string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	f, ok := m.files[path]
	if !ok {
		return nil, fmt.Errorf("storage[%s]: file not found: %s", m.diskName, path)
	}
	return io.NopCloser(bytesReader(f.data)), nil
}

func (m *MemoryDriver) Delete(_ context.Context, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.files, path)
	return nil
}

func (m *MemoryDriver) Exists(_ context.Context, path string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.files[path]
	return ok, nil
}

func (m *MemoryDriver) Size(_ context.Context, path string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	f, ok := m.files[path]
	if !ok {
		return 0, fmt.Errorf("storage[%s]: file not found: %s", m.diskName, path)
	}
	return int64(len(f.data)), nil
}

func (m *MemoryDriver) LastModified(_ context.Context, path string) (time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	f, ok := m.files[path]
	if !ok {
		return time.Time{}, fmt.Errorf("storage[%s]: file not found: %s", m.diskName, path)
	}
	return f.modified, nil
}

func (m *MemoryDriver) Copy(ctx context.Context, from, to string) error {
	rc, err := m.Get(ctx, from)
	if err != nil {
		return err
	}
	defer rc.Close()
	return m.Put(ctx, to, rc)
}

func (m *MemoryDriver) Move(ctx context.Context, from, to string) error {
	if err := m.Copy(ctx, from, to); err != nil {
		return err
	}
	return m.Delete(ctx, from)
}

func (m *MemoryDriver) List(_ context.Context, prefix string) ([]Contracts.StorageFile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Contracts.StorageFile
	for p, f := range m.files {
		if prefix == "" || hasPrefix(p, prefix) {
			out = append(out, Contracts.StorageFile{
				Path:         p,
				Size:         int64(len(f.data)),
				LastModified: f.modified,
				MimeType:     f.mimeType,
			})
		}
	}
	return out, nil
}

func (m *MemoryDriver) URL(_ context.Context, path string) (string, error) {
	return fmt.Sprintf("memory://%s/%s", m.diskName, path), nil
}

func (m *MemoryDriver) TemporaryURL(_ context.Context, path string, expires time.Duration) (string, error) {
	exp := time.Now().Add(expires).Unix()
	return fmt.Sprintf("memory://%s/%s?expires=%d", m.diskName, path, exp), nil
}

// ---- helpers ----

type byteReader struct {
	data   []byte
	offset int
}

func bytesReader(data []byte) *byteReader { return &byteReader{data: append([]byte(nil), data...)} }

func (r *byteReader) Read(p []byte) (int, error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

func hasPrefix(s, p string) bool {
	if len(p) > len(s) {
		return false
	}
	return s[:len(p)] == p
}

// ---- proxyDriver ----

type proxyDriver struct {
	mgr  *Manager
	name string
}

func (p *proxyDriver) Name() string { return p.name }

func (p *proxyDriver) Put(ctx context.Context, path string, content io.Reader, opts ...Contracts.StorageOption) error {
	d, err := p.mgr.Disk(p.name)
	if err != nil {
		return err
	}
	return d.Put(ctx, path, content, opts...)
}

func (p *proxyDriver) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	d, err := p.mgr.Disk(p.name)
	if err != nil {
		return nil, err
	}
	return d.Get(ctx, path)
}

func (p *proxyDriver) Delete(ctx context.Context, path string) error {
	d, err := p.mgr.Disk(p.name)
	if err != nil {
		return err
	}
	return d.Delete(ctx, path)
}

func (p *proxyDriver) Exists(ctx context.Context, path string) (bool, error) {
	d, err := p.mgr.Disk(p.name)
	if err != nil {
		return false, err
	}
	return d.Exists(ctx, path)
}

func (p *proxyDriver) Size(ctx context.Context, path string) (int64, error) {
	d, err := p.mgr.Disk(p.name)
	if err != nil {
		return 0, err
	}
	return d.Size(ctx, path)
}

func (p *proxyDriver) LastModified(ctx context.Context, path string) (time.Time, error) {
	d, err := p.mgr.Disk(p.name)
	if err != nil {
		return time.Time{}, err
	}
	return d.LastModified(ctx, path)
}

func (p *proxyDriver) Copy(ctx context.Context, from, to string) error {
	d, err := p.mgr.Disk(p.name)
	if err != nil {
		return err
	}
	return d.Copy(ctx, from, to)
}

func (p *proxyDriver) Move(ctx context.Context, from, to string) error {
	d, err := p.mgr.Disk(p.name)
	if err != nil {
		return err
	}
	return d.Move(ctx, from, to)
}

func (p *proxyDriver) List(ctx context.Context, prefix string) ([]Contracts.StorageFile, error) {
	d, err := p.mgr.Disk(p.name)
	if err != nil {
		return nil, err
	}
	return d.List(ctx, prefix)
}

func (p *proxyDriver) URL(ctx context.Context, path string) (string, error) {
	d, err := p.mgr.Disk(p.name)
	if err != nil {
		return "", err
	}
	return d.URL(ctx, path)
}

func (p *proxyDriver) TemporaryURL(ctx context.Context, path string, expires time.Duration) (string, error) {
	d, err := p.mgr.Disk(p.name)
	if err != nil {
		return "", err
	}
	return d.TemporaryURL(ctx, path, expires)
}
