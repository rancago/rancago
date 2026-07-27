package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	dhttp "github.com/rancago/framework/internal/adapters/driving/http"
	dgrpc "github.com/rancago/framework/internal/adapters/driving/grpc"
	"github.com/rancago/framework/internal/adapters/driven/auth"
	"github.com/rancago/framework/internal/adapters/driven/cache"
	"github.com/rancago/framework/internal/adapters/driven/persistence/inmemory"
	"github.com/rancago/framework/internal/adapters/driven/storage"
	"github.com/rancago/framework/internal/adapters/driven/websocket"
	"github.com/rancago/framework/internal/application/usecases"
	"github.com/rancago/framework/internal/kernel"
	"github.com/rancago/framework/internal/ports/driven"
	"github.com/rancago/framework/internal/ports/driving"
)

type MinimalGRPCServer struct {
	handlers map[string]interface{}
}

func NewMinimalGRPCServer() *MinimalGRPCServer {
	return &MinimalGRPCServer{handlers: make(map[string]interface{})}
}

func (s *MinimalGRPCServer) Serve(lis net.Listener) error {
	log.Printf("[rancago] gRPC stub server: listening on %s (extend with google.golang.org/grpc for real RPC)", lis.Addr())
	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}
		_ = conn.Close()
	}
}

func (s *MinimalGRPCServer) GracefulStop() {}

type Application struct {
	Container *kernel.Container
	Config    *kernel.Config
}

func New() *Application {
	app := &Application{
		Container: kernel.NewContainer(),
		Config:    kernel.LoadConfig(),
	}
	app.Container.Instance("config", app.Config)
	return app
}

func (a *Application) RegisterCore() {
	a.Container.Singleton("redis", func(c *kernel.Container) (interface{}, error) {
		mgr := cache.NewRedisManagerAdapter(&a.Config.Redis)
		_ = mgr.Connect()
		return mgr, nil
	})
	a.Container.Singleton("ws.hub", func(c *kernel.Container) (interface{}, error) {
		redisRaw, _ := c.Resolve("redis")
		hub := websocket.NewHubAdapter(redisRaw.(driven.CachePort))
		hub.StartRedisListener()
		go hub.Run()
		return hub, nil
	})
	a.Container.Singleton("storage", func(c *kernel.Container) (interface{}, error) {
		return storage.NewStorageManagerAdapter(&a.Config.Storage), nil
	})
	a.Container.Singleton("socialite", func(c *kernel.Container) (interface{}, error) {
		return auth.NewSocialiteManager(&a.Config.Auth), nil
	})
	a.Container.Singleton("repo.notification", func(c *kernel.Container) (interface{}, error) {
		return inmemory.NewInMemoryNotificationRepo(), nil
	})
	a.Container.Singleton("repo.user", func(c *kernel.Container) (interface{}, error) {
		return inmemory.NewInMemoryUserRepo(), nil
	})
	a.Container.Singleton("repo.document", func(c *kernel.Container) (interface{}, error) {
		return inmemory.NewInMemoryDocumentRepo(), nil
	})
	a.Container.Singleton("uc.notification", func(c *kernel.Container) (interface{}, error) {
		repoRaw, _ := c.Resolve("repo.notification")
		redisRaw, _ := c.Resolve("redis")
		wsRaw, _ := c.Resolve("ws.hub")
		return usecases.NewNotificationInteractor(
			repoRaw.(driven.NotificationRepository),
			redisRaw.(driven.CachePort),
			wsRaw.(driven.WebSocketPort),
		), nil
	})
	a.Container.Singleton("repo.role", func(c *kernel.Container) (interface{}, error) {
		return inmemory.NewInMemoryRoleRepo(), nil
	})
	a.Container.Singleton("repo.permission", func(c *kernel.Container) (interface{}, error) {
		return inmemory.NewInMemoryPermissionRepo(), nil
	})
	a.Container.Singleton("uc.user", func(c *kernel.Container) (interface{}, error) {
		userRepoRaw, _ := c.Resolve("repo.user")
		roleRepoRaw, _ := c.Resolve("repo.role")
		permRepoRaw, _ := c.Resolve("repo.permission")
		socRaw, _ := c.Resolve("socialite")
		return usecases.NewUserInteractor(
			userRepoRaw.(driven.UserRepository),
			roleRepoRaw.(driven.RoleRepository),
			permRepoRaw.(driven.PermissionRepository),
			socRaw.(driven.SocialitePort),
		), nil
	})
	a.Container.Singleton("uc.document", func(c *kernel.Container) (interface{}, error) {
		docRepoRaw, _ := c.Resolve("repo.document")
		return usecases.NewDocumentInteractor(docRepoRaw.(driven.DocumentRepository)), nil
	})
	a.Container.Alias("uc.notification", "driving.NotificationUseCase")
	a.Container.Alias("uc.user", "driving.UserUseCase")
	a.Container.Alias("uc.document", "driving.DocumentUseCase")
	a.Container.Alias("repo.notification", "driven.NotificationRepository")
	a.Container.Alias("redis", "driven.CachePort")
	a.Container.Alias("ws.hub", "driven.WebSocketPort")
	a.Container.Alias("storage", "driven.StorageManagerPort")
	a.Container.Alias("socialite", "driven.SocialitePort")
}

func (a *Application) Boot() {
	log.Println("[rancago] Core drivers wired (in-memory persistence, mock cache, mock storage, mock socialite).")
}

func (a *Application) BuildHTTPServer() *http.Server {
	mux := http.NewServeMux()
	web := dhttp.NewRouter()
	hh := dhttp.NewHealthHandler(a.Config.App.Name)
	web.GET("/", hh.Welcome, "web")
	web.GET("/api/v1/health", hh.Health, "api")
	for _, r := range web.All() {
		mux.HandleFunc(r.Path, r.Handler)
	}
	ucRaw, _ := a.Container.Resolve("uc.notification")
	nh := dhttp.NewNotificationHandler(ucRaw.(driving.NotificationUseCase))
	nh.RegisterRoutes(mux, "/api/v1/notifications")
	wsRaw, _ := a.Container.Resolve("ws.hub")
	mux.HandleFunc("/ws", wsRaw.(*websocket.HubAdapter).Handler)
	addr := fmt.Sprintf(":%d", a.Config.Server.HTTPPort)
	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

func (a *Application) BuildGRPCServer() *MinimalGRPCServer {
	s := NewMinimalGRPCServer()
	ucRaw, _ := a.Container.Resolve("uc.notification")
	grpcAdapter := dgrpc.NewGRPCAdapter(ucRaw.(driving.NotificationUseCase))
	grpcAdapter.RegisterGRPC(s)
	return s
}

func (a *Application) StartGRPC(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", a.Config.Server.GRPCPort)
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

func (a *Application) StartHTTP(ctx context.Context) error {
	srv := a.BuildHTTPServer()
	log.Printf("[rancago] HTTP server listening on :%d", a.Config.Server.HTTPPort)
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	select {
	case <-ctx.Done():
		return srv.Shutdown(ctx)
	case err := <-errCh:
		return err
	}
}
