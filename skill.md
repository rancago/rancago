# 🧠 Rancago Framework — Skill Matrix & Vibe Code Guide

> **Vibe Code Philosophy**: Produktif seperti Laravel, idiomatik seperti Go, skala enterprise tanpa ribet.
> Rancago bukan sekadar framework — ini adalah *state of mind*: **Contracts First, Providers Second, Services Third**.

---

## 🎯 Core Skills & Capabilities Matrix

| # | Skill Domain | Tingkat Penguasaan | Key Files Reference |
|---|---|---|---|
| 1 | **IoC Service Container** | ★★★★★ | [container.go](file:///d:/rancago/framework/Container/container.go) |
| 2 | **ServiceProvider Pattern** | ★★★★★ | [provider.go](file:///d:/Rancago/app/Contracts/provider.go) |
| 3 | **SOLID Principles Implementation** | ★★★★★ | `app/Contracts/*` + `framework/*` |
| 4 | **Multi-Storage OCP (MinIO/S3/GDrive)** | ★★★★★ | [storage.go](file:///d:/Rancago/app/Contracts/storage.go) + [manager.go](file:///d:/rancago/framework/Storage/manager.go) |
| 5 | **OAuth Socialite-style Registry** | ★★★★★ | [auth.go](file:///d:/Rancago/app/Contracts/auth.go) + `framework/Auth/socialite.go` |
| 6 | **RBAC Policy Enforcer (Redis-backed)** | ★★★★☆ | `framework/Auth/rbac.go` |
| 7 | **pgvector Semantic Search** | ★★★★☆ | [vector.go](file:///d:/rancago/framework/Database/vector.go) + [repository.go](file:///d:/rancago/framework/Database/repository.go) |
| 8 | **Google Calendar + Meet Integration** | ★★★★☆ | [google.go](file:///d:/Rancago/app/Contracts/google.go) + `framework/Google/*` |
| 9 | **Multi-API Transport (REST/gRPC/WS)** | ★★★★★ | [notification.go](file:///d:/Rancago/app/Contracts/notification.go) + `framework/Transport/*` |
| 10 | **Scalable WebSocket Hub** | ★★★★☆ | `framework/WebSocket/hub.go` |
| 11 | **Generic Repository Pattern** | ★★★★☆ | [database.go](file:///d:/Rancago/app/Contracts/database.go) |
| 12 | **CLI Artisan-style (Cobra)** | ★★★★☆ | [main.go](file:///d:/Rancago/cmd/Rancago/main.go) + `cmd/Rancago/commands/*` |

---

## 🏗 Architecture Vibe Code — The Rancago Way

### The Golden Trinity: Contracts → Providers → Services

```
┌─────────────────────────────────────────────────────────────┐
│  1. CONTRACTS (app/Contracts/*.go)                           │
│     └── Define ALL interfaces FIRST — ISP & DIP compliance    │
├─────────────────────────────────────────────────────────────┤
│  2. PROVIDERS (app/Providers/*.go)                           │
│     └── Register(Bind/Singleton) ke Container → Boot() wire   │
├─────────────────────────────────────────────────────────────┤
│  3. SERVICES (app/Services/*.go)                             │
│     └── Business logic, Transport-agnostic, depend Contracts  │
└─────────────────────────────────────────────────────────────┘
```

### Rule #1 — Contracts Always Come First

> ❌ **Bad Vibes**: Struct concrete di-import langsung
> ✅ **Good Vibes**: Struct menerima interface via constructor injection

```go
// ❌ Bad — ketergantungan langsung ke MinIO
type FileService struct {
    Minio *minio.Client
}

// ✅ Good — depend ke abstraksi, siap swap driver apapun
type FileService struct {
    Storage Contracts.StorageDriver
}
```

### Rule #2 — ServiceProvider is The Only Registration Point

> Setiap modul baru WAJIB punya ServiceProvider. Jangan scatter bind di bootstrap.

```go
// app/Providers/PaymentServiceProvider.go — POLA WAJIB
type PaymentServiceProvider struct{}

func (p *PaymentServiceProvider) Register(c *Container.Container) error {
    c.Singleton("service.payment", func(c *Container.Container) (interface{}, error) {
        storage, _ := c.Resolve("Contracts.StorageDriver")
        return Services.NewPaymentService(storage.(Contracts.StorageDriver)), nil
    })
    return nil
}

func (p *PaymentServiceProvider) Boot(c *Container.Container) error {
    // Runtime wiring: register drivers, seed data, etc.
    return nil
}
```

Lalu register di [bootstrap/app.go](file:///d:/Rancago/bootstrap/app.go#L55-L60):
```go
func (a *Application) RegisterCore() {
    a.RegisterProviders(
        Providers.NewStorageServiceProvider(),
        Providers.NewPaymentServiceProvider(), // ⭐ Tambahkan disini
    )
}
```

### Rule #3 — Container Naming Convention

| Tipe Binding | Format Nama | Contoh (Actual di Codebase) |
|---|---|---|
| **Singleton Service** | `service.{name}` | `service.notification` |
| **Manager** | `{domain}` (singular) | `storage` (bukan storage.mgr — lihat StorageServiceProvider) |
| **Driver Instance** | `{domain}.{driver}` | `storage.minio` (pattern via `Disk(name)` call) |
| **Config** | `config` | `config` (Instance, bound di `bootstrap.New()`) |
| **Contract Alias** | `Contracts.{InterfaceName}` | `Contracts.NotificationService`, `Contracts.StorageDriver` |
| **Type Alias** | `{Package}.{TypeName}` | `Storage.StorageManager` (alias untuk concrete struct) |
| **Infra** | `{infra}` | `redis`, `ws.hub` |

> ⚠️ **Actual Binding Reference**: Lihat [StorageServiceProvider.go](file:///d:/Rancago/app/Providers/StorageServiceProvider.go#L17-L38) — key `"storage"` dengan 2 alias: `"Storage.StorageManager"` (concrete type) dan `"Contracts.StorageDriver"` (interface).

---

## 📁 Folder-by-Folder Coding Conventions

### `app/Contracts/` — The Interface Bible

> 📜 **RULE**: Semua dependency antar-modul HANYA lewat Contracts. Jangan pernah cross-import concrete struct.

| File | Tanggung Jawab | Pattern |
|---|---|---|
| [provider.go](file:///d:/Rancago/app/Contracts/provider.go) | ServiceProvider lifecycle (Register + Boot) | Template Method |
| [storage.go](file:///d:/Rancago/app/Contracts/storage.go) | StorageDriver abstraction (13 methods) + Options | Strategy + Functional Options |
| [auth.go](file:///d:/Rancago/app/Contracts/auth.go) | AuthProvider + SocialManager + RBACService | Plugin Registry |
| [google.go](file:///d:/Rancago/app/Contracts/google.go) | Calendar + Meet + Scheduler | Facade |
| [database.go](file:///d:/Rancago/app/Contracts/database.go) | Generic Repository + Vector + Transaction | Generics + Unit of Work |
| [notification.go](file:///d:/Rancago/app/Contracts/notification.go) | NotificationService transport-agnostic | Interface Segregation |

**Vibe Check**: Jika menambah fitur baru → buat Contract terlebih dahulu di folder ini.

---

### `app/Providers/` — Wiring Central

Setiap Provider WAJIB:
1. Implement `Contracts.ServiceProvider` (2 method: Register + Boot)
2. Nama file: `{Domain}ServiceProvider.go`
3. Di Register: hanya bind ke Container, JANGAN ada logic berat
4. Di Boot: wire-up runtime (register drivers, seed data, run migrations)

```go
// POLA: app/Providers/XxxServiceProvider.go
type XxxServiceProvider struct{}

func NewXxxServiceProvider() Contracts.ServiceProvider {
    return &XxxServiceProvider{}
}

func (p *XxxServiceProvider) Register(c *Container.Container) error {
    // ✅ Hanya bind disini
    c.Singleton("service.xxx", func(c *Container.Container) (interface{}, error) {
        dep, _ := c.Resolve("dependency.key")
        return Services.NewXxxService(dep.(DepType)), nil
    })
    // ✅ Alias ke Contract untuk DIP
    c.Alias("service.xxx", "Contracts.XxxService")
    return nil
}

func (p *XxxServiceProvider) Boot(c *Container.Container) error {
    // ✅ Runtime registration, lazy init disini
    return nil
}
```

---

### `app/Services/` — Business Logic Utopia

> 🚫 **DILARANG**: Import `net/http`, `grpc`, `*websocket.Conn` di Services.
> Services harus Transport-Agnostic. Logic yang sama bisa dipanggil dari REST, gRPC, WebSocket, CLI, Queue Worker.

**Pola Constructor dengan Dependency Injection**:
```go
// app/Services/NotificationService.go — POLA BENAR
type NotificationService struct {
    redis *Cache.RedisManager
    hub   *WebSocket.Hub
}

// ⭐ Constructor: semua dependency EXPLICIT, tidak ada global var
func NewNotificationService(redis *Cache.RedisManager, hub *WebSocket.Hub) Contracts.NotificationService {
    return &NotificationService{redis: redis, hub: hub}
}

// ⭐ Semua method terima context.Context parameter pertama
func (s *NotificationService) Send(ctx context.Context, n *Contracts.Notification) (*Contracts.Notification, error) {
    // Business logic disini — tidak tahu apakah dipanggil via REST/gRPC/WS
}
```

---

### `framework/` — The Extracted Core

Isinya adalah framework reusable yang bisa diekstrak ke module terpisah. Setiap sub-folder di `framework/` mewakili 1 Bounded Context:

| Folder | Responsibility | Key Pattern |
|---|---|---|
| `Container/` | IoC Service Container | Singleton/Transient/Instance/Alias |
| `Storage/` + `Drivers/` | Multi-driver storage manager | Factory + Strategy + OCP |
| `Auth/` + `Providers/` | OAuth Socialite + RBAC + Policy | Registry + Middleware Chain |
| `Database/` | Generic Repository + pgvector + Transaction | Repository + Unit of Work |
| `Google/` | Calendar API + Meet API + Scheduler | Adapter + Facade |
| `Cache/` | Redis Manager + RateLimiter | Connection Manager |
| `WebSocket/` | Scalable Hub with Redis Pub/Sub | Pub/Sub + Hub |
| `Transport/` | REST/gRPC/WebSocket Adapters | Adapter |

---

### `config/` — Strict Struct Configuration

Tidak ada `map[string]interface{}` di config. Semua typed struct dengan default value yang production-ready di [config.go](file:///d:/Rancago/config/config.go#L84-L185).

---

### `routes/` — Declarative Route Builder

Router menggunakan pola chaining method dengan middleware string-names. Lihat [routes.go](file:///d:/Rancago/routes/routes.go) untuk pola dasar:

```go
func Api() *Router {
    r := NewRouter()
    r.GET("/api/v1/users", handler, "auth", "rbac:read:users")
    r.POST("/api/v1/orders", createOrder, "auth", "rate:10/min")
    return r
}
```

---

### `cmd/Rancago/commands/` — CLI Generators

Setiap command Cobra punya naming convention:
- `New{Name}Command() *cobra.Command`
- Generator commands = file `generators.go`
- Migration commands = file `migrate.go`

---

## 🎨 Rancago Design Patterns — Signature Vibe

### Pattern #1 — Functional Options Pattern

Digunakan di Storage, Calendar, Meet untuk parameter opsional yang readable:

```go
// app/Contracts/storage.go — POLA OPSI FUNGSIONAL
type StorageOption func(*StorageOptions)

func WithContentType(ct string) StorageOption { ... }
func WithACL(acl string) StorageOption { ... }
func WithMetadata(md map[string]string) StorageOption { ... }

// ✅ Usage — super readable, extensible (tambah opsi = tambah 1 fungsi)
storage.Put(ctx, "avatar.jpg", file,
    WithContentType("image/jpeg"),
    WithACL("public-read"),
    WithMetadata(map[string]string{"user": "123"}),
)
```

### Pattern #2 — Proxy Driver Pattern

StorageManager bisa di-wrap menjadi ProxyDriver yang mengimplement StorageDriver interface, sehingga bisa inject ke struct manapun tanpa tahu disk mana yang dipakai:

```go
// framework/Storage/manager.go — ProxyDriver Pattern
gdrive := storageMgr.Proxy("google_drive")
// gdrive TIPE NYA = Contracts.StorageDriver, siap inject!
```

### Pattern #3 — Multi-Adapter Transport Pattern

1 Service Logic = 3 Endpoint otomatis. Lihat `framework/Transport/NotificationTransport.go`:

```
NotificationService (app/Services)
    ├── NewRESTAdapter()    → /api/v1/notifications/*
    ├── NewGRPCAdapter()    → gRPC :9090
    └── NewWebSocketAction() → action:"notification:*" di /ws
```

TIDAK PERLU duplikat logic. Service hanya 1, adapters yang menyesuaikan format transport.

### Pattern #4 — Plugin Registry (OCP pada Steroid)

Contoh di StorageManager dan SocialiteManager:
```go
// ✅ Tambah driver BARU tanpa ubah Manager code
mgr.RegisterFactory("azure_blob", func(cfg StorageDiskConfig) (StorageDriver, error) {
    return NewAzureBlobDriver(cfg), nil
})

// ✅ Langsung pakai — OCP 100% compliant
disk, _ := mgr.Disk("my_azure_disk")
```

---

## ✍️ Coding Style & Naming Conventions

### Package Organization

```go
// ✅ Satu file = satu interface utama + struct pendukungnya
// app/Contracts/storage.go → StorageDriver, StorageFile, StorageOptions

// ✅ Nama file = nama interface (PascalCase → snake_case)
// Contracts.NotificationService → notification.go
```

### Struct & Constructor

```go
// ✅ CONVENTION ACTUAL: Concrete struct bisa exported + constructor return concrete pointer
//    (kemudian di-bind ke Container + Alias ke Interface untuk DIP compliance)
// Contoh actual: app/Services/NotificationService.go
type InMemoryNotificationService struct { /* fields */ }

// Constructor return concrete type — Container yang handle interface binding
func NewNotificationService(redis *Cache.RedisManager, hub *WebSocket.Hub) *InMemoryNotificationService {
    return &InMemoryNotificationService{redis: redis, wsHub: hub}
}

// ✅ Di ServiceProvider: bind concrete ke interface via Alias
c.Singleton("service.notification", func(c *Container.Container) (interface{}, error) {
    redis, _ := c.Resolve("redis")
    hub, _ := c.Resolve("ws.hub")
    // ⭐ Constructor call return concrete → Container resolve interface via Alias
    return Services.NewNotificationService(
        redis.(*Cache.RedisManager),
        hub.(*WebSocket.Hub),
    ), nil
})
c.Alias("service.notification", "Contracts.NotificationService") // ⭐ Penting!

// ✅ Di Service Provider Constructor — bisa return concrete, nanti Container terima
// app/Providers/StorageServiceProvider.go — actual:
func NewStorageServiceProvider() *StorageServiceProvider {
    return &StorageServiceProvider{}
}
// RegisterProviders() di app.go menerima Contracts.ServiceProvider — struct otomatis satisfy karena punya Register() + Boot()

### Error Handling

```go
// ✅ Wrap error dengan context deskriptif
if err := d.Put(ctx, path, content); err != nil {
    return fmt.Errorf("storage put [%s] failed: %w", path, err)
}

// ✅ Di provider boot/register, log.Fatalf untuk early failure
if err := p.Register(c); err != nil {
    log.Fatalf("[Rancago] Failed to register provider: %v", err)
}
```

### Context Handling

> RULE: Setiap exported function/method yang berinteraksi dengan I/O (DB, Redis, HTTP, Storage) **WAJIB** menerima `context.Context` sebagai parameter pertama.

```go
// ✅ Correct
func (s *Service) ProcessOrder(ctx context.Context, orderID string) error { ... }

// ❌ Wrong — tidak ada context
func (s *Service) ProcessOrder(orderID string) error { ... }
```

---

## 🚫 Anti-Patterns Yang Wajib Dihindari

| # | Anti-Pattern | Konsekuensi | Solusi Rancago |
|---|---|---|---|
| 1 | **Global variables untuk dependency** | Susah test, susah ganti impl | Gunakan Container + Constructor Injection |
| 2 | **Service import struct concrete langsung** | Coupling tinggi, susah refactor | Depend ke `Contracts.*` interface |
| 3 | **Business logic di HTTP Handler** | Tidak bisa dipakai gRPC/WS | Pindahkan ke `app/Services/` |
| 4 | **Bind tanpa Alias ke Contract** | Tidak bisa resolve via interface name | Selalu `c.Alias("service.xxx", "Contracts.XxxService")` |
| 5 | **Driver registration di bootstrap** | OCP violation — harus ubah kernel | Register via Provider.Boot() |
| 6 | **Config pakai map[string]interface{}** | Tidak type-safe, runtime panic | Tambah typed struct di `config/config.go` |
| 7 | **Skip context.Context** | Tidak bisa cancel/timeout propagation | Parameter pertama SELALU ctx |

---

## 🧪 Workflow: Add New Feature — Vibe Check

Misal: Tambah fitur **Payment Gateway** (Midtrans/Xendit). Step-by-step sesuai vibe Rancago:

### Step 1 — Contracts First 📜
```go
// app/Contracts/payment.go — BUAT PERTAMA
type PaymentGateway interface {
    Charge(ctx context.Context, amount int64, orderID string) (*PaymentResult, error)
    Refund(ctx context.Context, txID string, amount int64) error
    Name() string
}
type PaymentService interface {
    CreateInvoice(ctx context.Context, order *Order) (*Invoice, error)
}
```

### Step 2 — Implement Drivers (OCP-ready) 🏭
```go
// framework/Payment/Drivers/midtrans.go
type MidtransDriver struct { ... }
func (m *MidtransDriver) Charge(ctx context.Context, ...) { /* Midtrans API */ }
func (m *MidtransDriver) Name() string { return "midtrans" }
```

### Step 3 — Manager with Registry 🎛
```go
// framework/Payment/manager.go — pola sama dengan StorageManager
type PaymentManager struct {
    factories map[string]func(cfg) (Contracts.PaymentGateway, error)
    gateways  map[string]Contracts.PaymentGateway
}
func (pm *PaymentManager) RegisterFactory(name string, f ...) { ... }
func (pm *PaymentManager) Gateway(name string) (Contracts.PaymentGateway, error) { ... }
```

### Step 4 — Business Service 🧠
```go
// app/Services/PaymentService.go
type paymentService struct {
    payments *Payment.PaymentManager
}
func NewPaymentService(pm *Payment.PaymentManager) Contracts.PaymentService {
    return &paymentService{payments: pm}
}
// ⭐ Logic disini — transport agnostic
```

### Step 5 — ServiceProvider Wiring 🔌
```go
// app/Providers/PaymentServiceProvider.go
func (p *PaymentServiceProvider) Register(c *Container.Container) error {
    c.Singleton("payment.mgr", func(c *Container.Container) (interface{}, error) {
        pm := Payment.NewManager(&a.Config.Payment)
        pm.RegisterFactory("midtrans", Drivers.NewMidtransDriver)
        pm.RegisterFactory("xendit", Drivers.NewXenditDriver)
        return pm, nil
    })
    c.Singleton("service.payment", func(c *Container.Container) (interface{}, error) {
        pm, _ := c.Resolve("payment.mgr")
        return Services.NewPaymentService(pm.(*Payment.PaymentManager)), nil
    })
    c.Alias("service.payment", "Contracts.PaymentService")
    return nil
}
```

### Step 6 — Register di Bootstrap Kernel 🚀
Tambahkan ke `RegisterCore()` di [bootstrap/app.go](file:///d:/Rancago/bootstrap/app.go#L55-L60).

### Step 7 — Multi-Transport Adapters 🔗
```go
// framework/Transport/PaymentTransport.go
func NewPaymentRESTAdapter(svc Contracts.PaymentService) *RESTAdapter { ... }
func NewPaymentGRPCAdapter(svc Contracts.PaymentService) *GRPCAdapter { ... }
// Lalu wire-up di BuildHTTPServer & BuildGRPCServer
```

---

## ✅ Vibe Code Checklist — Sebelum Commit

- [ ] Semua dependency baru punya Contract di `app/Contracts/`?
- [ ] Service menggunakan constructor injection (tidak ada global)?
- [ ] ServiceProvider terpisah dan terdaftar di bootstrap?
- [ ] Tidak ada business logic di HTTP/gRPC/WS handler?
- [ ] Setiap method I/O menerima `context.Context` parameter pertama?
- [ ] Error di-wrap dengan `fmt.Errorf("...: %w", err)`?
- [ ] Container binding punya Alias ke `Contracts.*`?
- [ ] Nama binding sesuai convention (`service.xxx`, `xxx.mgr`)?
- [ ] Config pakai typed struct, bukan `map[string]interface{}`?

---

## 📌 Quick Reference Cheatsheet

```go
// ====== CONTAINER ======
c.Bind(...)           // Transient: resolve sekali = instance baru
c.Singleton(...)      // 1 instance selamanya
c.Instance(...)       // Pre-existing instance langsung dipakai
c.Alias("a", "b")     // "b" → resolve ke "a"
c.Resolve("key")      // Ambil instance
c.MustResolve("key")  // Ambil, panic jika gagal
c.Has("key")          // Cek binding ada

// ====== STORAGE (Actual: lihat StorageServiceProvider.go) ======
// ⚠️ KEY NAME = "storage" (bukan storage.mgr), 2 alias tersedia
sm := container.MustResolve("storage").(*Storage.StorageManager)          // Cara 1: concrete via alias Storage.StorageManager
sd := container.MustResolve("Contracts.StorageDriver").(Contracts.StorageDriver) // Cara 2: interface via Proxy default
disk, _ := sm.Disk()              // Default disk (minio sesuai config)
disk, _ := sm.Disk("google_drive")
disk, _ := sm.Disk("s3")
proxy := sm.Proxy("minio")        // StorageDriver interface langsung, bisa inject ke struct

// ====== NOTIFICATION ======
// ⚠️ Resolve via interface (recommended untuk DIP), alias ke service.notification
notif := container.MustResolve("Contracts.NotificationService").(Contracts.NotificationService)
notif.Send(ctx, &Contracts.Notification{UserID: "123", Title: "Halo"})
notif.Broadcast(ctx, "Pengumuman", "Server maintenance", nil)

// ====== INFRASTRUCTURE ======
redis := container.MustResolve("redis").(*Cache.RedisManager)
hub := container.MustResolve("ws.hub").(*WebSocket.Hub)
cfg := container.MustResolve("config").(*config.Config)

// ====== STORAGE DRIVER REGISTRATION (di Provider.Boot atau Register) ======
sm.RegisterFactory("azure_blob", func(dc config.StorageDiskConfig) (Contracts.StorageDriver, error) {
    return Drivers.NewAzureBlobDriver(dc), nil // Register driver baru TANPA ubah Manager!
})

---

> 🚀 **Remember the Vibe**: "Think in Contracts, Build with Providers, Scale with Services."
> Rancago bukan tentang seberapa banyak fitur — tapi seberapa mudah fitur itu ditambah tanpa merusak yang ada. SOLID itu bukan teori di Rancago, itu adalah default workflow sehari-hari.
