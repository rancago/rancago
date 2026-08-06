// Package bootstrap is the application kernel for Rancago Framework.
// It wires all ServiceProviders, adapters, and servers together.
package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/rancago/framework/app/Contracts"
	appProviders "github.com/rancago/framework/app/Providers"
	"github.com/rancago/framework/app/Services"
	"github.com/rancago/framework/config"
	"github.com/rancago/framework/framework/Auth"
	"github.com/rancago/framework/framework/Cache"
	"github.com/rancago/framework/framework/Container"
	"github.com/rancago/framework/framework/Google"
	"github.com/rancago/framework/framework/WebSocket"
	dhttp "github.com/rancago/framework/internal/adapters/driving/http"
	dgrpc "github.com/rancago/framework/internal/adapters/driving/grpc"
	"github.com/rancago/framework/internal/ports/driving"

	// internal adapters still used for legacy hexagonal layer
	"github.com/rancago/framework/internal/adapters/driven/auth"
	icache "github.com/rancago/framework/internal/adapters/driven/cache"
	"github.com/rancago/framework/internal/adapters/driven/persistence/inmemory"
	"github.com/rancago/framework/internal/adapters/driven/storage"
	iws "github.com/rancago/framework/internal/adapters/driven/websocket"
	"github.com/rancago/framework/internal/application/usecases"
	"github.com/rancago/framework/internal/kernel"
	idriven "github.com/rancago/framework/internal/ports/driven"
)

// MinimalGRPCServer is a stub gRPC server used until google.golang.org/grpc is wired in.
type MinimalGRPCServer struct{}

func NewMinimalGRPCServer() *MinimalGRPCServer { return &MinimalGRPCServer{} }

func (s *MinimalGRPCServer) Serve(lis net.Listener) error {
	log.Printf("[rancago] gRPC stub server: listening on %s", lis.Addr())
	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}
		_ = conn.Close()
	}
}

func (s *MinimalGRPCServer) GracefulStop() {}

// contextKey is the package-local context key type for RBAC user ID injection.
type contextKey string

const userIDCtxKey contextKey = "rancago_user_id"

// Application is the root application struct.
type Application struct {
	// Container is the IoC service container (framework layer).
	Container *Container.Container
	// Config is the typed application configuration.
	Config *config.Config

	// Legacy internal kernel (hexagonal adapters still use this).
	internalContainer *kernel.Container
	internalConfig    *kernel.Config
}

// New creates and bootstraps a new Application.
func New() *Application {
	cfg := config.Load()
	app := &Application{
		Container:         Container.NewContainer(),
		Config:            cfg,
		internalContainer: kernel.NewContainer(),
		internalConfig:    kernel.LoadConfig(),
	}
	app.Container.Instance("config", cfg)
	app.internalContainer.Instance("config", app.internalConfig)
	return app
}

// RegisterProviders runs Register() on all providers then Boot() on all providers.
// This ensures all bindings are available before any Boot() runs.
func (a *Application) RegisterProviders(providers ...Contracts.ServiceProvider) {
	for _, p := range providers {
		if err := p.Register(a.Container); err != nil {
			log.Fatalf("[rancago] provider Register failed: %v", err)
		}
	}
	for _, p := range providers {
		if err := p.Boot(a.Container); err != nil {
			log.Fatalf("[rancago] provider Boot failed: %v", err)
		}
	}
}

// RegisterCore wires the default set of infrastructure providers and the
// legacy hexagonal adapters used by the internal application layer.
func (a *Application) RegisterCore() {
	// ---- Framework-layer providers ----
	a.RegisterProviders(
		// Database (MySQL / PostgreSQL via GORM)
		appProviders.NewDatabaseServiceProvider(appProviders.DBProviderConfig{
			Driver:   a.Config.Database.Driver,
			Host:     a.Config.Database.Host,
			Port:     a.Config.Database.Port,
			User:     a.Config.Database.User,
			Password: a.Config.Database.Password,
			DBName:   a.Config.Database.DBName,
			Charset:  a.Config.Database.Charset,
			Loc:      a.Config.Database.Loc,
			SSLMode:  a.Config.Database.SSLMode,
			Timezone: a.Config.Database.Timezone,
			Debug:    a.Config.App.Env != "production",
		}),
		appProviders.NewStorageServiceProvider(
			a.Config.Storage.Default,
			a.Config.Storage.Disks,
		),
		appProviders.NewGoogleServiceProvider(Google.GoogleConfig{
			ClientID:     a.Config.Google.ClientID,
			ClientSecret: a.Config.Google.ClientSecret,
			RedirectURL:  a.Config.Google.RedirectURL,
			Scopes:       a.Config.Google.Scopes,
			Credentials:  a.Config.Google.Credentials,
		}),
		appProviders.NewAuthServiceProvider(
			a.Config.Auth.Providers,
			userIDCtxKey,
		),
	)

	// ---- Redis & WebSocket (framework layer) ----
	a.Container.Singleton("redis", func(_ *Container.Container) (interface{}, error) {
		mgr := Cache.NewRedisManager(&Cache.RedisConfig{
			Host:     a.Config.Redis.Host,
			Port:     a.Config.Redis.Port,
			Password: a.Config.Redis.Password,
			DB:       a.Config.Redis.DB,
		})
		_ = mgr.Connect()
		return mgr, nil
	})
	a.Container.Singleton("ws.hub", func(c *Container.Container) (interface{}, error) {
		redisRaw, _ := c.Resolve("redis")
		hub := WebSocket.NewHub(redisRaw.(*Cache.RedisManager))
		hub.StartRedisListener()
		go hub.Run()
		return hub, nil
	})

	// ---- NotificationService (framework Services layer) ----
	a.Container.Singleton("service.notification", func(c *Container.Container) (interface{}, error) {
		redisRaw, _ := c.Resolve("redis")
		hubRaw, _ := c.Resolve("ws.hub")
		return Services.NewNotificationService(
			redisRaw.(*Cache.RedisManager),
			hubRaw.(*WebSocket.Hub),
		), nil
	})
	a.Container.Alias("service.notification", "Contracts.NotificationService")

	// ---- Legacy hexagonal internal layer (keeps existing adapters running) ----
	a.internalContainer.Singleton("redis", func(_ *kernel.Container) (interface{}, error) {
		mgr := icache.NewRedisManagerAdapter(&a.internalConfig.Redis)
		_ = mgr.Connect()
		return mgr, nil
	})
	a.internalContainer.Singleton("ws.hub", func(c *kernel.Container) (interface{}, error) {
		redisRaw, _ := c.Resolve("redis")
		hub := iws.NewHubAdapter(redisRaw.(idriven.CachePort))
		hub.StartRedisListener()
		go hub.Run()
		return hub, nil
	})
	a.internalContainer.Singleton("storage", func(_ *kernel.Container) (interface{}, error) {
		return storage.NewStorageManagerAdapter(&a.internalConfig.Storage), nil
	})
	a.internalContainer.Singleton("socialite", func(_ *kernel.Container) (interface{}, error) {
		return auth.NewSocialiteManager(&a.internalConfig.Auth), nil
	})
	a.internalContainer.Singleton("repo.notification", func(_ *kernel.Container) (interface{}, error) {
		return inmemory.NewInMemoryNotificationRepo(), nil
	})
	a.internalContainer.Singleton("repo.user", func(_ *kernel.Container) (interface{}, error) {
		return inmemory.NewInMemoryUserRepo(), nil
	})
	a.internalContainer.Singleton("repo.document", func(_ *kernel.Container) (interface{}, error) {
		return inmemory.NewInMemoryDocumentRepo(), nil
	})
	a.internalContainer.Singleton("repo.role", func(_ *kernel.Container) (interface{}, error) {
		return inmemory.NewInMemoryRoleRepo(), nil
	})
	a.internalContainer.Singleton("repo.permission", func(_ *kernel.Container) (interface{}, error) {
		return inmemory.NewInMemoryPermissionRepo(), nil
	})
	a.internalContainer.Singleton("uc.notification", func(c *kernel.Container) (interface{}, error) {
		repoRaw, _ := c.Resolve("repo.notification")
		redisRaw, _ := c.Resolve("redis")
		wsRaw, _ := c.Resolve("ws.hub")
		return usecases.NewNotificationInteractor(
			repoRaw.(idriven.NotificationRepository),
			redisRaw.(idriven.CachePort),
			wsRaw.(idriven.WebSocketPort),
		), nil
	})
	a.internalContainer.Singleton("uc.user", func(c *kernel.Container) (interface{}, error) {
		userRepoRaw, _ := c.Resolve("repo.user")
		roleRepoRaw, _ := c.Resolve("repo.role")
		permRepoRaw, _ := c.Resolve("repo.permission")
		socRaw, _ := c.Resolve("socialite")
		return usecases.NewUserInteractor(
			userRepoRaw.(idriven.UserRepository),
			roleRepoRaw.(idriven.RoleRepository),
			permRepoRaw.(idriven.PermissionRepository),
			socRaw.(idriven.SocialitePort),
		), nil
	})
	a.internalContainer.Singleton("uc.document", func(c *kernel.Container) (interface{}, error) {
		docRepoRaw, _ := c.Resolve("repo.document")
		return usecases.NewDocumentInteractor(docRepoRaw.(idriven.DocumentRepository)), nil
	})
	a.internalContainer.Alias("uc.notification", "driving.NotificationUseCase")
	a.internalContainer.Alias("uc.user", "driving.UserUseCase")
	a.internalContainer.Alias("uc.document", "driving.DocumentUseCase")
}

// Boot logs the bootstrap completion.
func (a *Application) Boot() {
	log.Println("[rancago] Core providers booted. Framework + internal adapters wired.")
	log.Printf("[rancago] Auth providers: %v", a.Config.Auth.Providers)
	log.Printf("[rancago] Storage disks: %v (default: %s)", a.Config.Storage.Disks, a.Config.Storage.Default)
}

// BuildHTTPServer builds and returns a configured *http.Server.
func (a *Application) BuildHTTPServer() *http.Server {
	mux := http.NewServeMux()

	// Health + welcome (internal adapters router)
	web := dhttp.NewRouter()
	hh := dhttp.NewHealthHandler(a.internalConfig.App.Name)
	web.GET("/", hh.Welcome, "web")
	web.GET("/api/v1/health", hh.Health, "api")
	for _, r := range web.All() {
		mux.HandleFunc(r.Path, r.Handler)
	}

	// Notification REST adapter (internal hexagonal)
	ucRaw, _ := a.internalContainer.Resolve("uc.notification")
	nh := dhttp.NewNotificationHandler(ucRaw.(driving.NotificationUseCase))
	nh.RegisterRoutes(mux, "/api/v1/notifications")

	// WebSocket endpoint (internal hub)
	wsRaw, _ := a.internalContainer.Resolve("ws.hub")
	mux.HandleFunc("/ws", wsRaw.(*iws.HubAdapter).Handler)

	addr := fmt.Sprintf(":%d", a.internalConfig.Server.HTTPPort)
	return &http.Server{Addr: addr, Handler: mux}
}

// BuildGRPCServer builds and returns a MinimalGRPCServer.
func (a *Application) BuildGRPCServer() *MinimalGRPCServer {
	s := NewMinimalGRPCServer()
	ucRaw, _ := a.internalContainer.Resolve("uc.notification")
	grpcAdapter := dgrpc.NewGRPCAdapter(ucRaw.(driving.NotificationUseCase))
	grpcAdapter.RegisterGRPC(s)
	return s
}

// StartHTTP starts the HTTP server and blocks until ctx is cancelled.
func (a *Application) StartHTTP(ctx context.Context) error {
	srv := a.BuildHTTPServer()
	log.Printf("[rancago] HTTP server listening on :%d", a.internalConfig.Server.HTTPPort)
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	select {
	case <-ctx.Done():
		return srv.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

// StartGRPC starts the gRPC stub server and blocks until ctx is cancelled.
func (a *Application) StartGRPC(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", a.internalConfig.Server.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("grpc listen failed: %w", err)
	}
	srv := a.BuildGRPCServer()
	log.Printf("[rancago] gRPC server listening on %s", addr)
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(lis) }()
	select {
	case <-ctx.Done():
		srv.GracefulStop()
		return nil
	case err := <-errCh:
		return err
	}
}

// AuthMiddleware is an example Bearer token auth middleware.
// It injects the user ID into the request context for RBAC middleware.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: validate Bearer token from Authorization header.
		// For demo: accept X-User-ID header as user identity.
		userID := r.Header.Get("X-User-ID")
		if userID != "" {
			ctx := context.WithValue(r.Context(), userIDCtxKey, userID)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

// withAuth wraps a handler with the AuthMiddleware.
func withAuth(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AuthMiddleware(h).ServeHTTP(w, r)
	})
}

// unused suppresses compiler warnings for withAuth until it is wired to real routes.
var _ = withAuth

// OAuthConfig is a re-export for use outside this package.
type OAuthConfig = Auth.OAuthConfig
