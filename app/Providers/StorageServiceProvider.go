// Package Providers contains ServiceProvider implementations for Rancago Framework.
package Providers

import (
	"github.com/rancago/framework/app/Contracts"
	"github.com/rancago/framework/framework/Container"
	"github.com/rancago/framework/framework/Storage"
)

// StorageServiceProvider registers and boots the multi-driver storage manager.
//
// Container bindings (post-registration):
//   - "storage"                → *Storage.Manager (singleton, concrete)
//   - "Storage.Manager"        → alias to "storage" (concrete type alias)
//   - "Contracts.StorageDriver" → alias to "storage" via DefaultDisk proxy (interface)
type StorageServiceProvider struct {
	diskCfgs    map[string]Contracts.StorageDiskConfig
	defaultDisk string
}

// NewStorageServiceProvider creates a StorageServiceProvider.
// diskCfgs maps disk names to their configurations.
// defaultDisk is the name of the disk returned by Manager.DefaultDisk().
func NewStorageServiceProvider(
	defaultDisk string,
	diskCfgs map[string]Contracts.StorageDiskConfig,
) *StorageServiceProvider {
	return &StorageServiceProvider{
		diskCfgs:    diskCfgs,
		defaultDisk: defaultDisk,
	}
}

// Register binds the StorageManager into the container.
func (p *StorageServiceProvider) Register(c *Container.Container) error {
	c.Singleton("storage", func(c *Container.Container) (interface{}, error) {
		mgr := Storage.NewManager(p.defaultDisk, p.diskCfgs)
		return mgr, nil
	})
	c.Alias("storage", "Storage.Manager")
	c.Alias("storage", "Contracts.StorageDriver")
	return nil
}

// Boot runs after all providers have registered.
// Register additional driver factories here (OCP extension point).
func (p *StorageServiceProvider) Boot(_ *Container.Container) error {
	return nil
}
