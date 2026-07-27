package kernel

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Storage  StorageConfig
	Google   GoogleConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Server   ServerConfig
}

type AppConfig struct {
	Name string
	Env  string
	Key  string
	URL  string
}

type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	Timezone string
}

type StorageConfig struct {
	Default string
	Disks   map[string]StorageDiskConfig
}

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

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	Credentials  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type AuthConfig struct {
	Providers map[string]OAuthProviderConfig
}

type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
}

type ServerConfig struct {
	HTTPPort int
	GRPCPort int
	WSPort   int
	Debug    bool
}

func LoadConfig() *Config {
	return &Config{
		App: AppConfig{
			Name: "Rancago Framework",
			Env:  "local",
			Key:  "base64:your-app-key-here-change-in-production",
			URL:  "http://localhost:8080",
		},
		Database: DatabaseConfig{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     5432,
			User:     "rancago",
			Password: "rancago",
			DBName:   "rancago_db",
			SSLMode:  "disable",
			Timezone: "Asia/Jakarta",
		},
		Storage: StorageConfig{
			Default: "minio",
			Disks: map[string]StorageDiskConfig{
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
			},
		},
		Google: GoogleConfig{
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
			Providers: map[string]OAuthProviderConfig{
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
