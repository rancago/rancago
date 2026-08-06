// Package config provides fully-typed configuration for Rancago Framework.
// All values are typed structs - no map[string]interface{} here.
package config

import (
	"github.com/rancago/framework/app/Contracts"
	"github.com/rancago/framework/framework/Auth"
	"github.com/rancago/framework/framework/Google"
)

// Config is the root configuration struct for the Rancago application.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Storage  StorageConfig
	Google   Google.GoogleConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Server   ServerConfig
}

// AppConfig holds general application settings.
type AppConfig struct {
	Name string
	Env  string // local | development | staging | production
	Key  string // APP_KEY (base64-encoded 32-byte secret)
	URL  string
}

// DatabaseConfig holds database connection settings.
// Supported drivers: "mysql" (or "mariadb") and "postgres".
type DatabaseConfig struct {
	Driver   string // "mysql" | "postgres"
	Host     string
	Port     int
	User     string
	Password string
	DBName   string

	// PostgreSQL specific
	SSLMode  string
	Timezone string

	// MySQL / MariaDB specific
	Charset string // default: utf8mb4
	Loc     string // default: Local
}

// StorageConfig mirrors the Contracts.StorageManager config.
type StorageConfig struct {
	Default string
	Disks   map[string]Contracts.StorageDiskConfig
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// AuthConfig holds OAuth provider configurations.
type AuthConfig struct {
	Providers map[string]Auth.OAuthConfig
}

// ServerConfig holds HTTP, gRPC, and WebSocket server settings.
type ServerConfig struct {
	HTTPPort int
	GRPCPort int
	WSPort   int
	Debug    bool
}

// Load returns a *Config with sensible production-ready defaults.
// Override individual fields via environment variables in your bootstrap.
func Load() *Config {
	return &Config{
		App: AppConfig{
			Name: "Rancago Framework",
			Env:  "local",
			Key:  "base64:change-me-in-production-use-rancago-key-generate",
			URL:  "http://localhost:8080",
		},
		Database: DatabaseConfig{
			Driver:   "mysql",
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "",
			DBName:   "rancago_db",
			Charset:  "utf8mb4",
			Loc:      "Local",
			// Uncomment below to use PostgreSQL instead:
			// Driver:   "postgres",
			// Port:     5432,
			// SSLMode:  "disable",
			// Timezone: "Asia/Jakarta",
		},
		Storage: StorageConfig{
			Default: "minio",
			Disks: map[string]Contracts.StorageDiskConfig{
				"minio": {
					Driver:    "minio",
					Endpoint:  "localhost:9000",
					Bucket:    "rancago-bucket",
					Region:    "us-east-1",
					AccessKey: "minioadmin",
					SecretKey: "minioadmin",
					UseSSL:    false,
				},
				"s3": {
					Driver:    "s3",
					Endpoint:  "s3.amazonaws.com",
					Bucket:    "rancago-bucket",
					Region:    "ap-southeast-1",
					AccessKey: "your-aws-key",
					SecretKey: "your-aws-secret",
					UseSSL:    true,
				},
				"google_drive": {
					Driver:      "google_drive",
					Credentials: "path/to/service-account.json",
					FolderID:    "root",
				},
				"memory": {
					Driver: "memory",
				},
			},
		},
		Google: Google.GoogleConfig{
			ClientID:     "your-google-client-id",
			ClientSecret: "your-google-client-secret",
			RedirectURL:  "http://localhost:8080/auth/google/callback",
			Scopes: []string{
				"https://www.googleapis.com/auth/calendar",
				"https://www.googleapis.com/auth/drive",
				"https://www.googleapis.com/auth/meetings.space.created",
			},
			Credentials: "path/to/service-account.json",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
		Auth: AuthConfig{
			Providers: map[string]Auth.OAuthConfig{
				"google": {
					ClientID:     "google-client-id",
					ClientSecret: "google-client-secret",
					RedirectURL:  "http://localhost:8080/auth/google/callback",
					AuthURL:      "https://accounts.google.com/o/oauth2/auth",
					TokenURL:     "https://oauth2.googleapis.com/token",
					UserInfoURL:  "https://www.googleapis.com/oauth2/v3/userinfo",
					Scopes:       []string{"email", "profile"},
				},
				"github": {
					ClientID:     "github-client-id",
					ClientSecret: "github-client-secret",
					RedirectURL:  "http://localhost:8080/auth/github/callback",
					AuthURL:      "https://github.com/login/oauth/authorize",
					TokenURL:     "https://github.com/login/oauth/access_token",
					UserInfoURL:  "https://api.github.com/user",
					Scopes:       []string{"user:email", "read:user"},
				},
				"facebook": {
					ClientID:     "fb-client-id",
					ClientSecret: "fb-client-secret",
					RedirectURL:  "http://localhost:8080/auth/facebook/callback",
					AuthURL:      "https://www.facebook.com/v18.0/dialog/oauth",
					TokenURL:     "https://graph.facebook.com/v18.0/oauth/access_token",
					UserInfoURL:  "https://graph.facebook.com/me?fields=id,name,email",
					Scopes:       []string{"email", "public_profile"},
				},
			},
		},
		Server: ServerConfig{
			HTTPPort: 8080,
			GRPCPort: 9090,
			WSPort:   6001,
			Debug:    true,
		},
	}
}
