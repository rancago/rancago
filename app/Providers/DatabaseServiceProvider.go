package Providers

import (
	"context"
	"fmt"
	"log"

	"github.com/rancago/framework/app/Contracts"
	"github.com/rancago/framework/framework/Container"
	"github.com/rancago/framework/framework/Database"
)

// DatabaseServiceProvider wires the database connection into the container.
//
// Container bindings registered:
//   - "db"                  → *Database.Manager  (concrete)
//   - "Database.Manager"    → alias to "db"
//   - "Contracts.Database"  → alias to "db"
//
// Usage in bootstrap/app.go:
//
//	app.RegisterProviders(
//	    Providers.NewDatabaseServiceProvider(Providers.DBProviderConfig{
//	        Driver:   cfg.Database.Driver,  // "mysql" or "postgres"
//	        Host:     cfg.Database.Host,
//	        Port:     cfg.Database.Port,
//	        User:     cfg.Database.User,
//	        Password: cfg.Database.Password,
//	        DBName:   cfg.Database.DBName,
//	        SSLMode:  cfg.Database.SSLMode,
//	        Debug:    cfg.App.Env != "production",
//	    }),
//	)
type DatabaseServiceProvider struct {
	cfg DBProviderConfig
}

// DBProviderConfig holds all database connection parameters for the provider.
type DBProviderConfig struct {
	Driver   string // "mysql" | "postgres"
	Host     string
	Port     int
	User     string
	Password string
	DBName   string

	// MySQL specific
	Charset string // default: utf8mb4
	Loc     string // default: Local

	// PostgreSQL specific
	SSLMode  string // default: disable
	Timezone string // default: Asia/Jakarta

	Debug bool
}

// NewDatabaseServiceProvider creates a DatabaseServiceProvider.
func NewDatabaseServiceProvider(cfg DBProviderConfig) Contracts.ServiceProvider {
	return &DatabaseServiceProvider{cfg: cfg}
}

// Register binds the database manager as a Singleton.
// The connection is lazy — opened on first Resolve("db").
func (p *DatabaseServiceProvider) Register(c *Container.Container) error {
	cfg := p.cfg // capture for closure

	c.Singleton("db", func(_ *Container.Container) (interface{}, error) {
		dbCfg := &Database.DBConfig{
			Driver:    cfg.Driver,
			Host:      cfg.Host,
			Port:      cfg.Port,
			User:      cfg.User,
			Password:  cfg.Password,
			DBName:    cfg.DBName,
			Charset:   cfg.Charset,
			Loc:       cfg.Loc,
			SSLMode:   cfg.SSLMode,
			Timezone:  cfg.Timezone,
			Debug:     cfg.Debug,
			ParseTime: true,
		}
		mgr, err := Database.NewConnected(dbCfg)
		if err != nil {
			return nil, fmt.Errorf("DatabaseServiceProvider: %w", err)
		}
		log.Printf("[rancago] Database connected: driver=%s host=%s dbname=%s",
			cfg.Driver, cfg.Host, cfg.DBName)
		return mgr, nil
	})

	c.Alias("db", "Database.Manager")
	c.Alias("db", "Contracts.Database")
	return nil
}

// Boot runs a connectivity ping after all providers have registered.
func (p *DatabaseServiceProvider) Boot(c *Container.Container) error {
	if !c.Has("db") {
		return nil
	}
	raw, err := c.Resolve("db")
	if err != nil {
		return fmt.Errorf("DatabaseServiceProvider boot: %w", err)
	}
	mgr, ok := raw.(*Database.Manager)
	if !ok || !mgr.IsConnected() {
		return nil
	}
	// Warm ping to surface connectivity issues early
	if err := mgr.Ping(context.Background()); err != nil {
		log.Printf("[rancago] ⚠️  Database ping failed: %v", err)
		// Non-fatal — app can still start; handle at route level
	}
	return nil
}
