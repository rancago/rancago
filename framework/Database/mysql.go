package Database

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQLDSN builds a MySQL/MariaDB DSN from DBConfig.
// Format: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
func MySQLDSN(cfg *DBConfig) string {
	charset := cfg.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	loc := cfg.Loc
	if loc == "" {
		loc = "Local"
	}
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s&multiStatements=true",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		charset,
		loc,
	)
}

// openMySQL opens a GORM connection to MySQL / MariaDB.
func openMySQL(cfg *DBConfig, gormCfg *gorm.Config) (*gorm.DB, error) {
	dsn := MySQLDSN(cfg)
	return gorm.Open(mysql.Open(dsn), gormCfg)
}

// NewMySQLManager is a convenience constructor for a MySQL Manager.
//
// Example:
//
//	db, err := Database.NewMySQLManager(&Database.DBConfig{
//	    Host:     "localhost",
//	    Port:     3306,
//	    User:     "root",
//	    Password: "secret",
//	    DBName:   "rancago_db",
//	})
func NewMySQLManager(cfg *DBConfig) (*Manager, error) {
	cfg.Driver = "mysql"
	return NewConnected(cfg)
}
