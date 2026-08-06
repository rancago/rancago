package Database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresDSN builds a PostgreSQL DSN (connection string) from DBConfig.
// Format: host=... user=... password=... dbname=... port=... sslmode=... TimeZone=...
func PostgresDSN(cfg *DBConfig) string {
	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	tz := cfg.Timezone
	if tz == "" {
		tz = "Asia/Jakarta"
	}
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.Port,
		sslMode,
		tz,
	)
}

// openPostgres opens a GORM connection to PostgreSQL.
func openPostgres(cfg *DBConfig, gormCfg *gorm.Config) (*gorm.DB, error) {
	dsn := PostgresDSN(cfg)
	return gorm.Open(postgres.Open(dsn), gormCfg)
}

// NewPostgresManager is a convenience constructor for a PostgreSQL Manager.
//
// Example:
//
//	db, err := Database.NewPostgresManager(&Database.DBConfig{
//	    Host:     "localhost",
//	    Port:     5432,
//	    User:     "rancago",
//	    Password: "secret",
//	    DBName:   "rancago_db",
//	    SSLMode:  "disable",
//	})
func NewPostgresManager(cfg *DBConfig) (*Manager, error) {
	cfg.Driver = "postgres"
	return NewConnected(cfg)
}
