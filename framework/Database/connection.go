// Package Database provides multi-driver database connection management for Rancago.
// Supported drivers: mysql, postgres.
// Add gorm.io/driver/mysql and gorm.io/driver/postgres to go.mod to use.
package Database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConfig holds the configuration for a database connection.
type DBConfig struct {
	Driver   string // "mysql" | "postgres"
	Host     string
	Port     int
	User     string
	Password string
	DBName   string

	// MySQL / MariaDB specific
	Charset   string // default: utf8mb4
	ParseTime bool   // default: true
	Loc       string // default: Local

	// PostgreSQL specific
	SSLMode  string // default: disable
	Timezone string // default: Asia/Jakarta

	// Connection pool settings
	MaxOpenConns    int           // default: 25
	MaxIdleConns    int           // default: 10
	ConnMaxLifetime time.Duration // default: 1h
	ConnMaxIdleTime time.Duration // default: 30m

	// GORM settings
	Debug bool // enable GORM query log
}

// Manager wraps a *gorm.DB and implements Contracts.DatabaseConnection.
type Manager struct {
	mu  sync.RWMutex
	db  *gorm.DB
	cfg *DBConfig
}

// NewManager creates and connects a new database Manager.
// Call Connect() to open the actual connection, or use NewConnected() for convenience.
func NewManager(cfg *DBConfig) *Manager {
	applyDefaults(cfg)
	return &Manager{cfg: cfg}
}

// NewConnected creates a Manager and immediately opens a connection.
// Returns an error if the connection fails.
func NewConnected(cfg *DBConfig) (*Manager, error) {
	m := NewManager(cfg)
	return m, m.Connect()
}

// Connect opens the database connection based on cfg.Driver.
func (m *Manager) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	logLevel := logger.Silent
	if m.cfg.Debug {
		logLevel = logger.Info
	}
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	var (
		db  *gorm.DB
		err error
	)

	switch m.cfg.Driver {
	case "mysql", "mariadb":
		db, err = openMySQL(m.cfg, gormCfg)
	case "postgres", "postgresql":
		db, err = openPostgres(m.cfg, gormCfg)
	default:
		return fmt.Errorf("database: unsupported driver %q (supported: mysql, postgres)", m.cfg.Driver)
	}
	if err != nil {
		return fmt.Errorf("database: connect [%s] failed: %w", m.cfg.Driver, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("database: get sql.DB failed: %w", err)
	}
	sqlDB.SetMaxOpenConns(m.cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(m.cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(m.cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(m.cfg.ConnMaxIdleTime)

	m.db = db
	return nil
}

// DB returns the underlying *gorm.DB.
// Panics if Connect() has not been called.
func (m *Manager) DB() *gorm.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.db == nil {
		panic("database: Manager not connected — call Connect() first")
	}
	return m.db
}

// SqlDB returns the underlying *sql.DB for low-level access.
func (m *Manager) SqlDB() (*sql.DB, error) {
	return m.DB().DB()
}

// Ping verifies the database connection is alive.
func (m *Manager) Ping(ctx context.Context) error {
	sqlDB, err := m.SqlDB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// GetDialect returns the driver name.
func (m *Manager) GetDialect() string {
	return m.cfg.Driver
}

// Migrate runs raw SQL migration statements sequentially.
func (m *Manager) Migrate(ctx context.Context, migrations []string) error {
	db := m.DB().WithContext(ctx)
	for _, stmt := range migrations {
		if err := db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("database: migration failed: %w", err)
		}
	}
	return nil
}

// AutoMigrate runs GORM AutoMigrate for the given model types.
func (m *Manager) AutoMigrate(models ...interface{}) error {
	return m.DB().AutoMigrate(models...)
}

// Close closes the underlying sql.DB connection pool.
func (m *Manager) Close() error {
	sqlDB, err := m.SqlDB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// IsConnected returns true if the connection has been established.
func (m *Manager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.db != nil
}

// applyDefaults fills in zero-value fields with sensible defaults.
func applyDefaults(cfg *DBConfig) {
	if cfg.Charset == "" {
		cfg.Charset = "utf8mb4"
	}
	if cfg.Loc == "" {
		cfg.Loc = "Local"
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	if cfg.Timezone == "" {
		cfg.Timezone = "Asia/Jakarta"
	}
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 25
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 10
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = time.Hour
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = 30 * time.Minute
	}
	// ParseTime defaults to true for MySQL
	// (zero value bool == false, so we can't default to true here without
	//  a separate flag — callers should set ParseTime: true explicitly or
	//  use MySQLDSN which always sets it)
}
