# 🧠 Rancago Framework - Skill Matrix & Vibe Code Guide

> **Vibe Code Philosophy**: Produktif seperti Laravel, idiomatik seperti Go, skala enterprise tanpa ribet.
> Rancago bukan sekadar framework - ini adalah *state of mind*: **Contracts First, Providers Second, Services Third**.

---

## 🎯 Core Skills Matrix

| # | Skill Domain | Key Files |
|---|---|---|
| 1 | **IoC Service Container** | `framework/Container/container.go` |
| 2 | **ServiceProvider Pattern** | `app/Contracts/provider.go` |
| 3 | **SOLID Principles Implementation** | `app/Contracts/` + `framework/` |
| 4 | **Multi-Storage OCP (MinIO/S3/GDrive)** | `app/Contracts/storage.go` + `framework/Storage/manager.go` |
| 5 | **OAuth Socialite Registry** | `app/Contracts/auth.go` + `framework/Auth/socialite.go` |
| 6 | **RBAC Policy Enforcer (Redis-backed)** | `framework/Auth/rbac.go` |
| 7 | **pgvector Semantic Search** | `app/Models/vector.go` + `app/Repositories/repository.go` |
| 8 | **Google Calendar + Meet Integration** | `app/Contracts/google.go` + `framework/Google/` |
| 9 | **Multi-API Transport (REST/gRPC/WS)** | `app/Contracts/notification.go` + `framework/Transport/` |
| 10 | **Scalable WebSocket Hub** | `framework/WebSocket/hub.go` |
| 11 | **Generic Repository Pattern** | `app/Contracts/database.go` |
| 12 | **Hexagonal Architecture** | `internal/` (domain → ports → usecases → adapters) |
| 13 | **CLI Generators** | `internal/adapters/driving/cli/commands/` |

---

## 🏗 Architecture Vibe Code - The Rancago Way

### Two Complementary Layers

```
Framework Layer (app/ + framework/)          Hexagonal Layer (internal/)
────────────────────────────────             ──────────────────────────────────
app/Contracts/   - interfaces                internal/domain/         - entities + VOs + errors
app/Providers/   - DI wiring                 internal/ports/          - driven + driving interfaces
app/Services/    - business logic            internal/application/    - use case interactors
framework/       - implementations           internal/adapters/       - HTTP/gRPC/CLI/DB/Cache
bootstrap/app.go - dual-container boot       internal/bootstrap/      - internal wiring
```

New features → use the **hexagonal layer** (`internal/`). Existing framework layer stays for backward compat.

### The Golden Trinity (Framework Layer): Contracts → Providers → Services

```
1. CONTRACTS (app/Contracts/*.go)
   └── Define ALL interfaces FIRST - ISP & DIP compliance

2. PROVIDERS (app/Providers/*.go)
   └── Register(Bind/Singleton) → Boot() wire adapters

3. SERVICES (app/Services/*.go)
   └── Business logic, Transport-agnostic, depend on Contracts only
```

### Hexagonal Layer Flow

```
domain → ports → application/usecases → adapters → bootstrap → cmd
```

Dependency direction is strictly one-way. Never reverse it.

---

## 📐 Rule #1 - Contracts Always Come First

```go
// ❌ Bad - direct coupling to implementation
type FileService struct {
    Minio *minio.Client
}

// ✅ Good - depend on abstraction, swap driver freely
type FileService struct {
    Storage Contracts.StorageDriver
}
```

## 📐 Rule #2 - ServiceProvider is the Only Registration Point

```go
// app/Providers/PaymentServiceProvider.go
type PaymentServiceProvider struct{}

func (p *PaymentServiceProvider) Register(c *Container.Container) error {
    c.Singleton("service.payment", func(c *Container.Container) (interface{}, error) {
        storage, _ := c.Resolve("Contracts.StorageDriver")
        return Services.NewPaymentService(storage.(Contracts.StorageDriver)), nil
    })
    c.Alias("service.payment", "Contracts.PaymentService")
    return nil
}

func (p *PaymentServiceProvider) Boot(c *Container.Container) error {
    return nil
}
```

Register di `bootstrap/app.go`:
```go
func (a *Application) RegisterCore() {
    a.RegisterProviders(
        Providers.NewStorageServiceProvider(...),
        Providers.NewPaymentServiceProvider(), // tambah disini
    )
}
```

## 📐 Rule #3 - Container Naming Convention

| Tipe | Format | Contoh |
|---|---|---|
| Singleton Service | `service.{name}` | `service.notification` |
| Manager | `{domain}` | `storage`, `redis` |
| Infrastructure | `{infra}` | `ws.hub` |
| Contract Alias | `Contracts.{Interface}` | `Contracts.StorageDriver` |
| Config | `config` | `config` |
| Hexagonal repo | `repo.{name}` | `repo.user`, `repo.notification` |
| Hexagonal usecase | `uc.{name}` | `uc.user`, `uc.notification` |
| Driving port alias | `driving.{UseCase}` | `driving.UserUseCase` |

---

## 📁 Folder-by-Folder Conventions

### `app/Contracts/` - The Interface Bible

All cross-module dependencies go through Contracts only. Never cross-import concrete structs.

| File | Responsibility | Pattern |
|---|---|---|
| `provider.go` | ServiceProvider lifecycle | Template Method |
| `storage.go` | StorageDriver 13 methods + Options | Strategy + Functional Options |
| `auth.go` | AuthProvider + SocialManager + RBACService | Plugin Registry |
| `google.go` | Calendar + Meet + Scheduler | Facade |
| `database.go` | Generic Repository + Vector + Transaction | Generics + Unit of Work |
| `notification.go` | NotificationService transport-agnostic | Interface Segregation |

### `app/Providers/` - Wiring Central

Every Provider must:
1. Implement `Contracts.ServiceProvider` (Register + Boot)
2. File name: `{Domain}ServiceProvider.go`
3. Register: bind to Container only - no heavy logic
4. Boot: runtime wire-up (register drivers, seed, migrations)

### `app/Services/` - Business Logic

```go
// ❌ NEVER import these in Services
import "net/http"
import "google.golang.org/grpc"
```

Services must be Transport-Agnostic. Same logic callable from REST, gRPC, WebSocket, CLI.

```go
func NewNotificationService(redis *Cache.RedisManager, hub *WebSocket.Hub) Contracts.NotificationService {
    return &NotificationService{redis: redis, hub: hub}
}

func (s *NotificationService) Send(ctx context.Context, n *Contracts.Notification) (*Contracts.Notification, error) {
    // Pure business logic - no HTTP/gRPC/WS here
}
```

### `internal/` - Hexagonal Architecture (Canonical for New Features)

```go
// Constructor in use case - injects driven ports, returns driving port interface
func NewFooInteractor(repo driven.FooRepository) driving.FooUseCase {
    return &FooInteractor{repo: repo}
}

// Wiring in internal/bootstrap/app.go
a.Container.Singleton("repo.foo", func(c *kernel.Container) (interface{}, error) {
    return inmemory.NewInMemoryFooRepo(), nil
})
a.Container.Singleton("uc.foo", func(c *kernel.Container) (interface{}, error) {
    r, _ := c.Resolve("repo.foo")
    return usecases.NewFooInteractor(r.(driven.FooRepository)), nil
})
a.Container.Alias("uc.foo", "driving.FooUseCase")
```

---

## 🎨 Signature Design Patterns

### Functional Options Pattern

```go
storage.Put(ctx, "avatar.jpg", file,
    WithContentType("image/jpeg"),
    WithACL("public-read"),
    WithMetadata(map[string]string{"user": "123"}),
)
```

### Proxy Driver Pattern

```go
// Returns StorageDriver interface directly - inject anywhere
proxy := storageMgr.Proxy("google_drive")
```

### Multi-Adapter Transport Pattern

```
NotificationService (one implementation)
    ├── NewRESTAdapter()         → /api/v1/notifications/*
    ├── NewGRPCAdapter()         → gRPC :9090
    └── NewWebSocketAction()     → action:"notification:*" at /ws
```

### Plugin Registry (OCP)

```go
// Add driver without touching Manager code
mgr.RegisterFactory("azure_blob", func(cfg StorageDiskConfig) (Contracts.StorageDriver, error) {
    return NewAzureBlobDriver(cfg), nil
})
disk, _ := mgr.Disk("my_azure_disk")
```

---

## ✍️ Coding Style

### Constructor Pattern

```go
// Hexagonal layer: return interface, not struct
func NewFooInteractor(repo driven.FooRepository) driving.FooUseCase {
    return &FooInteractor{repo: repo}
}

// Framework layer: return concrete, Container handles interface via Alias
func NewNotificationService(redis *Cache.RedisManager, hub *WebSocket.Hub) *InMemoryNotificationService {
    return &InMemoryNotificationService{redis: redis, wsHub: hub}
}
// Then in provider: c.Alias("service.notification", "Contracts.NotificationService")
```

### Domain Errors (Hexagonal Layer)

```go
// Always use derrors - never raw fmt.Errorf for business errors
return derrors.New("foo.create", derrors.ErrValidation, "name is required")

// Available sentinels:
// derrors.ErrNotFound / ErrUnauthorized / ErrForbidden
// derrors.ErrValidation / ErrConflict / ErrAlreadyExists / ErrInternal
```

### Error Handling (Framework Layer)

```go
if err := d.Put(ctx, path, content); err != nil {
    return fmt.Errorf("storage put [%s] failed: %w", path, err)
}
```

### Context Rule

Every exported function touching I/O (DB, Redis, HTTP, Storage) **must** accept `context.Context` as first parameter.

```go
// ✅ Correct
func (s *Service) ProcessOrder(ctx context.Context, orderID string) error { ... }

// ❌ Wrong
func (s *Service) ProcessOrder(orderID string) error { ... }
```

---

## 🚫 Anti-Patterns

| # | Anti-Pattern | Solusi |
|---|---|---|
| 1 | Global variables untuk dependency | Container + Constructor Injection |
| 2 | Service import struct concrete langsung | Depend ke `Contracts.*` interface |
| 3 | Business logic di HTTP Handler | Pindahkan ke `app/Services/` atau use case |
| 4 | Bind tanpa Alias ke Contract | Selalu `c.Alias("service.xxx", "Contracts.XxxService")` |
| 5 | Import `internal/` dari `app/` atau `framework/` | Layer boundary: internal is strictly isolated |
| 6 | Adapter import adapter lain | Adapters hanya depend pada port interfaces |
| 7 | Config pakai `map[string]interface{}` | Typed struct di `config/config.go` |
| 8 | Skip `context.Context` | Parameter pertama SELALU ctx |

---

## 🧪 Workflow: Add New Feature - Step by Step

Contoh: tambah fitur **Payment Gateway**.

```
Step 1 - Scaffold (pakai CLI):
  rancago make:feature Payment

Step 2 - Define Contract (framework layer) or edit generated port (hexagonal):
  app/Contracts/payment.go  OR  internal/ports/driving/payment_usecase.go

Step 3 - Implement driver (OCP):
  framework/Payment/Drivers/midtrans.go

Step 4 - Business Service / Use Case:
  app/Services/PaymentService.go  OR  internal/application/usecases/payment_usecase.go

Step 5 - ServiceProvider wiring:
  app/Providers/PaymentServiceProvider.go

Step 6 - Register di bootstrap:
  bootstrap/app.go  →  RegisterCore()  →  tambah PaymentServiceProvider

Step 7 - Transport adapters:
  framework/Transport/PaymentTransport.go  (REST + gRPC)
  OR driving adapter sudah di-generate oleh make:feature
```

---

## ✅ Vibe Code Checklist - Sebelum Commit

- [ ] Semua dependency baru punya Contract di `app/Contracts/` atau port di `internal/ports/`?
- [ ] Service/use case menggunakan constructor injection (tidak ada global)?
- [ ] ServiceProvider terpisah dan terdaftar di bootstrap?
- [ ] Tidak ada business logic di HTTP/gRPC/WS handler?
- [ ] Setiap method I/O menerima `context.Context` parameter pertama?
- [ ] Error di-wrap dengan `derrors.New()` (hexagonal) atau `fmt.Errorf("...: %w", err)` (framework)?
- [ ] Container binding punya Alias ke `Contracts.*` atau `driving.*`?
- [ ] Nama binding sesuai convention?
- [ ] Config pakai typed struct?
- [ ] Feature baru punya `docs/features/<name>.md`?

---

## 📌 Quick Reference Cheatsheet

```go
// ====== CONTAINER (framework layer) ======
c.Bind(...)           // Transient
c.Singleton(...)      // 1 instance
c.Instance(...)       // Pre-existing
c.Alias("a", "b")     // "b" resolves to "a"
c.Resolve("key")      // Get instance
c.MustResolve("key")  // Get, panic on miss
c.Has("key")          // Check binding

// ====== CONTAINER (internal/kernel) - same API ======
a.Container.Singleton("repo.foo", func(c *kernel.Container) (interface{}, error) { ... })
a.Container.Alias("uc.foo", "driving.FooUseCase")

// ====== STORAGE ======
sm := container.MustResolve("storage").(*Storage.StorageManager)
disk, _ := sm.Disk()                    // Default disk
disk, _ := sm.Disk("google_drive")
proxy := sm.Proxy("minio")             // StorageDriver interface, injectable

// ====== NOTIFICATION ======
notif := container.MustResolve("Contracts.NotificationService").(Contracts.NotificationService)
notif.Send(ctx, &Contracts.Notification{UserID: "123", Title: "Halo"})

// ====== INFRASTRUCTURE ======
redis := container.MustResolve("redis").(*Cache.RedisManager)
hub   := container.MustResolve("ws.hub").(*WebSocket.Hub)
cfg   := container.MustResolve("config").(*config.Config)

// ====== DOMAIN ERRORS (hexagonal) ======
derrors.New("order.create", derrors.ErrValidation, "amount must be positive")
derrors.IsNotFound(err)     // true if wraps ErrNotFound
derrors.IsValidation(err)   // true if wraps ErrValidation

// ====== VALUE OBJECTS (hexagonal) ======
id    := valueobjects.NewIDStr("uuid-here")
id    := valueobjects.NewIDUint(42)
email, err := valueobjects.NewEmail("user@example.com")
```

---

> 🚀 **Remember the Vibe**: "Think in Contracts, Build with Providers, Scale with Services."
> Rancago bukan tentang seberapa banyak fitur - tapi seberapa mudah fitur itu ditambah tanpa merusak yang ada.
