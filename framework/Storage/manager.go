// Package Storage provides a multi-driver storage manager for Rancago.
// Design: OCP (add driver = RegisterFactory, no Manager code changes),
//         LSP (all drivers implement Contracts.StorageDriver),
//         Proxy pattern for constructor injection.
package Storage

import (
	"fmt"
	"sync"

	"github.com/rancago/framework/app/Contracts"
)

// Manager is the StorageManager implementation.
// It satisfies Contracts.StorageManager.
type Manager struct {
	defaultDisk string
	mu          sync.RWMutex
	disks       map[string]Contracts.StorageDriver
	diskCfgs    map[string]Contracts.StorageDiskConfig
	factories   map[string]func(Contracts.StorageDiskConfig) (Contracts.StorageDriver, error)
}

// NewManager creates a StorageManager with the given disk configurations and default disk name.
func NewManager(defaultDisk string, diskCfgs map[string]Contracts.StorageDiskConfig) *Manager {
	m := &Manager{
		defaultDisk: defaultDisk,
		disks:       make(map[string]Contracts.StorageDriver),
		diskCfgs:    diskCfgs,
		factories:   make(map[string]func(Contracts.StorageDiskConfig) (Contracts.StorageDriver, error)),
	}
	// Register the built-in in-memory driver as a fallback for every disk.
	m.RegisterFactory("memory", func(cfg Contracts.StorageDiskConfig) (Contracts.StorageDriver, error) {
		return NewMemoryDriver(cfg.Driver), nil
	})
	return m
}

// RegisterFactory registers a driver factory for a given driver type.
// This is the OCP extension point: add a new driver type without touching Manager.
func (m *Manager) RegisterFactory(driverType string, factory func(Contracts.StorageDiskConfig) (Contracts.StorageDriver, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.factories[driverType] = factory
}

// Disk returns the StorageDriver for the named disk (lazy-initialised from config).
// If name is omitted the default disk is used.
func (m *Manager) Disk(name ...string) (Contracts.StorageDriver, error) {
	diskName := m.defaultDisk
	if len(name) > 0 && name[0] != "" {
		diskName = name[0]
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if d, ok := m.disks[diskName]; ok {
		return d, nil
	}
	cfg, ok := m.diskCfgs[diskName]
	if !ok {
		return nil, fmt.Errorf("storage: disk %q is not configured", diskName)
	}
	factory, ok := m.factories[cfg.Driver]
	if !ok {
		// Fall back to in-memory driver.
		factory = func(c Contracts.StorageDiskConfig) (Contracts.StorageDriver, error) {
			return NewMemoryDriver(diskName), nil
		}
	}
	d, err := factory(cfg)
	if err != nil {
		return nil, fmt.Errorf("storage: init disk %q: %w", diskName, err)
	}
	m.disks[diskName] = d
	return d, nil
}

// DefaultDisk returns the driver for the configured default disk.
func (m *Manager) DefaultDisk() Contracts.StorageDriver {
	d, _ := m.Disk()
	if d == nil {
		return NewMemoryDriver("default")
	}
	return d
}

// Register adds or overrides a named driver instance directly (bypasses factory).
func (m *Manager) Register(name string, driver Contracts.StorageDriver) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.disks[name] = driver
}

// Proxy returns a StorageDriver that proxies to the named disk.
// Use this for constructor injection so the consumer doesn't need the Manager.
func (m *Manager) Proxy(name string) Contracts.StorageDriver {
	return &proxyDriver{mgr: m, name: name}
}

// AvailableDisks returns the list of configured disk names.
func (m *Manager) AvailableDisks() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, 0, len(m.diskCfgs))
	for k := range m.diskCfgs {
		names = append(names, k)
	}
	return names
}
