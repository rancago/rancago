# 📋 Rancago Framework — Product Requirements Document (PRD)

> **Version**: 1.0.0  
> **Last Updated**: 2026-07-27  
> **Author**: Rancago Framework Team  
> **Status**: Active Development  
> **Module Path**: `github.com/rancago/framework`

---

## 🎯 1. Executive Summary & Product Vision

### Product Vision

> **"Go + Laravel = Produktif namun tetap idiomatik & kokoh."**

Rancago adalah framework backend berbasis **Go 1.23+** yang mengadopsi DX (Developer Experience) gaya Laravel, namun tetap memegang teguh idiom Go, **SOLID Principles**, dan kesiapan skala enterprise. Rancago menjawab frustasi developer Go yang harus merakit stack dari nol (router, DI, storage abstraction, auth, realtime) sambil menghindari "Laravel PHP ke Go" yang tidak idiomatik.

### Unique Value Proposition (UVP)

| Dimensi | Vanilla Go | Laravel PHP | **Rancago Framework** |
|---|---|---|---|
| **Kecepatan** | ⚡⚡⚡ Sangat Cepat | 🐢 Moderate | ⚡⚡⚡ Go Native Performance |
| **DX / Produktivitas** | 📉 Setup manual berhari-hari | 📈 Artisan, Facades out of the box | 📈 CLI Generators + Contracts-First |
| **SOLID / Clean Arch** | ❌ Bebas, sering over-coupled | ⚠️ Facades = anti-DIP | ✅ **Contracts First** di enforce |
| **Realtime & Scale** | 🛠 Manual implement Redis PubSub | ⛔ Pusher eksternal mahal | ✅ WebSocket Hub multi-node built-in |
| **Multi-API Transport** | 📉 REST doang (lainnya manual) | 📉 REST + queued jobs | ✅ **1 Service = REST/gRPC/WS** auto |
| **AI-Ready (pgvector)** | ❌ Manual setup extension | ⛔ Butuh package eksternal | ✅ Semantic Search 1 line code |

### Target Persona

1. **👨‍💻 Go Backend Engineer** — Ingin produktivitas Laravel tanpa meninggalkan performa & type safety Go.
2. **🏢 Startup CTO** — Butuh framework yang bisa scale dari MVP ke 1M+ user tanpa rewrite total.
3. **🎓 Software Architect** — Mengutamakan SOLID, testability, dan modular design yang enforce pattern.
4. **🔧 Ex-Laravel Developer yang belajar Go** — Ingin familiar workflow tapi idiomatik di Go.

---

## 🏗 2. System Architecture Overview

### 2.1 Layered Architecture (Clean Architecture Style)

```
┌──────────────────────────────────────────────────────────────────┐
│                     TRANSPORT LAYER (Entry Points)                │
│  REST HTTP (net/http ServeMux) • gRPC • WebSocket Hub (/ws)      │
│  ↓ Adapters: RESTAdapter / GRPCAdapter / WSAction                │
├──────────────────────────────────────────────────────────────────┤
│                       SERVICE LAYER (Use Cases)                   │
│  app/Services/*.go — Business Logic, Transport-Agnostic          │
│  ↓ Dependencies: HANYA pakai Contracts (interfaces)              │
├──────────────────────────────────────────────────────────────────┤
│                     CONTRACTS LAYER (Abstractions)                │
│  app/Contracts/*.go — ServiceProvider, StorageDriver, etc.       │
│  ← No dependency ke luar — ISP & DIP 100% compliant              │
├──────────────────────────────────────────────────────────────────┤
│                  REPOSITORY / DRIVER LAYER (Implementations)      │
│  framework/Storage/Drivers/ • framework/Auth/Providers/          │
│  framework/Database/repository.go • framework/Google/*.go        │
├──────────────────────────────────────────────────────────────────┤
│                     INFRASTRUCTURE LAYER (External)               │
│  PostgreSQL + pgvector • Redis • MinIO/S3/GDrive • Google APIs   │
└──────────────────────────────────────────────────────────────────┘
```

### 2.2 Physical Folder Structure (Actual di `d:\Rancago`)

```
d:\Rancago\
├── 📁 app/                          # Application Layer (User Code)
│   ├── 📁 Contracts/                 # ⭐ Semua interface definitions
│   │   ├── auth.go                   # AuthProvider + RBAC + SocialManager
│   │   ├── database.go               # Repository<T> + VectorRepository + Transaction
│   │   ├── google.go                 # Calendar + Meet + Scheduler
│   │   ├── notification.go           # NotificationService (Transport Agnostic)
│   │   ├── provider.go               # ServiceProvider interface
│   │   └── storage.go                # StorageDriver + Options Pattern
│   ├── 📁 Models/                    # GORM Entity Models (actual file content)
│   │   └── models.go                 # User + Role + Permission + Document (with pgvector HNSW index 1536d)
│   ├── 📁 Providers/                 # ServiceProvider implementations
│   │   ├── AuthServiceProvider.go
│   │   ├── GoogleServiceProvider.go
│   │   └── StorageServiceProvider.go
│   └── 📁 Services/                  # Business Logic / Use Cases
│       └── NotificationService.go
├── 📁 bootstrap/                     # Application Kernel
│   └── app.go                        # New() → RegisterCore() → Boot() → StartHTTP/gRPC
├── 📁 cmd/Rancago/                     # Artisan-style CLI
│   ├── 📁 commands/                  # Generators + Migrations
│   │   ├── generators.go             # make:* commands
│   │   └── migrate.go                # migrate:* commands
│   └── main.go                       # Cobra root CLI entry
├── 📁 config/                        # Typed Struct Configuration
│   └── config.go                     # Load() returns *Config with defaults
├── 📁 framework/                     # ⭐ Core Framework (extractable module)
│   ├── 📁 Auth/                      # OAuth Socialite + RBAC + Policy
│   │   ├── 📁 Providers/             # Generic OAuth wrapper
│   │   ├── rbac.go                   # Redis-backed RBAC + Middleware
│   │   └── socialite.go              # SocialiteManager + Driver Registry
│   ├── 📁 Cache/                     # Redis Manager
│   │   └── redis.go                  # Connection + PubSub + RateLimiter
│   ├── 📁 Container/                 # IoC Service Container
│   │   └── container.go              # Bind/Singleton/Instance/Alias/Resolve
│   ├── 📁 Database/                  # GORM + pgvector + Repository
│   │   ├── repository.go             # Generic CRUD + Vector Semantic Search
│   │   └── vector.go                 # pgvector type (Scan/Value + GORM)
│   ├── 📁 Google/                    # Calendar + Meet APIs
│   │   ├── calendar.go               # Google Calendar CRUD events
│   │   └── meet.go                   # Google Meet Space + Scheduler
│   ├── 📁 Storage/                   # Multi-driver Storage Manager
│   │   ├── 📁 Drivers/               # MinIO • S3 • Google Drive
│   │   │   ├── gdrive.go
│   │   │   └── minio.go
│   │   └── manager.go                # Factory + Proxy Pattern
│   ├── 📁 Transport/                 # Multi-API Adapters
│   │   └── NotificationTransport.go  # REST/gRPC/WS 3-in-1
│   └── 📁 WebSocket/                 # Scalable Realtime Hub
│       └── hub.go                    # Redis PubSub multi-node broadcast
├── 📁 routes/                        # Declarative Router Builder
│   └── routes.go                     # Web() + Api() Routers
├── 📄 main.go                        # Entry point (HTTP + gRPC + WS parallel)
├── 📄 go.mod                         # module github.com/rancago/framework (Go 1.23.4)
├── 📄 README.md                      # Public documentation
├── 📄 skill.md                       # ⭐ Coding conventions & vibe code guide
└── 📄 prd.md                         # ⭐ This document
```

### 2.3 Core Workflow: Request Lifecycle

```
Client Request (HTTP/gRPC/WS)
    ↓
Transport Adapter (REST/GRPCAdapter/WSAction)
    ↓ format serialization, extract params
Service Contract method call (e.g. NotificationService.Send)
    ↓ resolve via Container
app/Services/ concrete implementation
    ↓ business logic, call Repository/Driver via Contract
Repository/Driver (PostgreSQL/Redis/Storage/Google API)
    ↓ Infrastructure I/O
Response kembali ke Transport Adapter → serialized → Client
```

---

## 📦 3. Feature Breakdown & Requirements

### 3.1 F1: IoC Service Container

**Priority**: P0 (Must Have)  
**Owner**: Framework Core  
**Reference Files**: [container.go](file:///d:/rancago/framework/Container/container.go) + [provider.go](file:///d:/Rancago/app/Contracts/provider.go)

#### Requirements Functional (FR):
- [x] FR-F1-01: Support 3 binding types: **Singleton** (1 instance), **Transient** (new per resolve), **Instance** (pre-existing)
- [x] FR-F1-02: Support **Alias** mapping (e.g. `Contracts.NotificationService` → `service.notification`)
- [x] FR-F1-03: Method `Resolve(abstract string) (interface{}, error)` + `MustResolve(abstract string) interface{}`
- [x] FR-F1-04: Thread-safe dengan `sync.RWMutex` untuk concurrent resolve
- [x] FR-F1-05: `Call(fn interface{})` — auto-resolve parameter function via reflection
- [x] FR-F1-06: `Has(abstract string) bool` untuk cek binding existence

#### Non-Functional Requirements (NFR):
- [x] NFR-F1-01: Resolve time < 50µs per binding (cached singleton)
- [x] NFR-F1-02: Zero allocation untuk Singleton resolve setelah instantiasi pertama

#### Design Pattern:
- IoC Container dengan Registration-Resolution separation
- ServiceProvider lifecycle: `Register` → `Boot`

---

### 3.2 F2: ServiceProvider Registration System

**Priority**: P0 (Must Have)  
**Reference Files**: [provider.go](file:///d:/Rancago/app/Contracts/provider.go) + [app.go Bootstrap](file:///d:/Rancago/bootstrap/app.go#L38-L60)

#### FR:
- [x] FR-F2-01: Setiap module harus implement `ServiceProvider` interface:
  ```go
  type ServiceProvider interface {
      Register(app *Container.Container) error
      Boot(app *Container.Container) error
  }
  ```
- [x] FR-F2-02: `Application.RegisterProviders(...)` menerima variadic ServiceProvider
- [x] FR-F2-03: Semua `Register()` dijalankan DULU sebelum semua `Boot()` (dependency graph aman)
- [x] FR-F2-04: `RegisterCore()` di app.go bootstrap default 3 providers:
  - `StorageServiceProvider`
  - `GoogleServiceProvider`
  - `AuthServiceProvider`

#### NFR:
- [x] NFR-F2-01: Jika salah satu provider gagal Register/Boot → `log.Fatalf` (early failure principle)

---

### 3.3 F3: Multi-Storage Manager (Flysystem-style OCP)

**Priority**: P0 (Must Have)  
**Reference Files**: [storage.go](file:///d:/Rancago/app/Contracts/storage.go) + [manager.go](file:///d:/rancago/framework/Storage/manager.go)

#### FR:
- [x] FR-F3-01: `StorageDriver` interface 13 method lengkap:
  - CRUD File: `Put / Get / Delete / Exists / Size / LastModified`
  - Operasi Bulk: `Copy / Move / List`
  - URL Generation: `URL / TemporaryURL`
  - Meta: `Name()`
- [x] FR-F3-02: **Functional Options Pattern**: `WithContentType / WithACL / WithMetadata / WithVisibility`
- [x] FR-F3-03: 3 Drivers built-in: **MinIO** + **Amazon S3** + **Google Drive** (via Service Account)
- [x] FR-F3-04: Lazy Disk Initialization — Disk di-instantiate hanya saat `Disk(name)` pertama kali dipanggil
- [x] FR-F3-05: `RegisterFactory(driverType, factory)` untuk tambah driver baru TANPA ubah Manager code
- [x] FR-F3-06: **ProxyDriver Pattern** — `Proxy(name)` returns `StorageDriver` langsung (untuk constructor injection)
- [x] FR-F3-07: Default disk configurable via `StorageConfig.Default`
- [x] FR-F3-08: `AvailableDisks() []string` list semua configured disk
- [x] FR-F3-09: **Container Binding Actual** (lihat [StorageServiceProvider.go](file:///d:/Rancago/app/Providers/StorageServiceProvider.go#L17-L38)):
  - Singleton key: `"storage"` (bukan `"storage.mgr"`)
  - Alias #1: `"Storage.StorageManager"` → resolve ke `*Storage.StorageManager` concrete
  - Alias #2: `"Contracts.StorageDriver"` → resolve ke default Proxy (untuk DIP constructor injection)

#### Config Structure (di [config.go](file:///d:/Rancago/config/config.go#L31-L47)):
```go
type StorageConfig struct {
    Default string
    Disks   map[string]StorageDiskConfig // minio, s3, google_drive
}
```

#### NFR:
- [x] NFR-F3-01: 100% Open/Closed Principle — tambah driver = RegisterFactory, tidak sentuh file framework
- [x] NFR-F3-02: Liskov Substitution — MinIODriver, GDriveDriver, S3Driver interchangeable tanpa kode consumer berubah

---

### 3.4 F4: OAuth Socialite-style Manager + RBAC Policy

**Priority**: P0 (Must Have)  
**Reference Files**: [auth.go](file:///d:/Rancago/app/Contracts/auth.go)

#### FR (Socialite OAuth):
- [x] FR-F4-01: `AuthProvider` interface: `Name / Redirect / Callback / UserFromToken`
- [x] FR-F4-02: `SocialUser` struct seragam untuk SEMUA provider (Provider/ID/Email/Name/Avatar/Nickname/Token/RefreshToken/RawAttributes)
- [x] FR-F4-03: 3 Providers built-in config: **Google** + **GitHub** + **Facebook**
- [x] FR-F4-04: `RegisterDriver(name, factory)` — tambah provider custom (Keycloak/LinkedIn/OIDC) TANPA ubah manager
- [x] FR-F4-05: `Redirect(ctx, driver)` returns `(authURL, state, error)` + auto-generate secure state
- [x] FR-F4-06: `Callback(ctx, driver, code, state)` validates state + exchanges code → `SocialUser`

#### FR (RBAC Redis-backed):
- [x] FR-F4-07: **Role Management**: `AssignRole / RemoveRole / HasRole / HasAnyRole / HasAllRoles / GetRoles`
- [x] FR-F4-08: **Permission Management**: `GivePermission / RevokePermission / HasPermission`
- [x] FR-F4-09: **Policy Enforcer Middleware**:
  - `Middleware(requiredPermission)` → HTTP middleware untuk permission-based auth
  - `RoleMiddleware(requiredRole...)` → HTTP middleware untuk role-based auth
  - Chainable: `Authenticate → Enforce → Handler`
- [x] FR-F4-10: Storage backing: 100% Redis (Set + Hash data structure)
- [x] FR-F4-11: Naming convention keys:
  - Roles user: `rbac:user:{userID}:roles` (SET)
  - Permissions role: `rbac:role:{roleName}:permissions` (SET)
  - Direct permission user: `rbac:user:{userID}:permissions` (SET)

#### NFR:
- [x] NFR-F4-01: RBAC check < 1ms karena Redis in-memory
- [x] NFR-F4-02: OCP — tambah OAuth provider tidak ubah SocialiteManager code

---

### 3.5 F5: Google Ecosystem Integration (Calendar + Meet)

**Priority**: P1 (Should Have)  
**Reference Files**: [google.go](file:///d:/Rancago/app/Contracts/google.go)

#### FR:
- [x] FR-F5-01: `CalendarService` interface: CRUD Events + AddAttendees + ListEvents
- [x] FR-F5-02: Full `CalendarEvent` fields:
  - Summary, Description, Location, Start/End/Timezone
  - Attendees (Email + DisplayName + Optional + ResponseStatus)
  - ConferenceData (ConferenceRequest.Type + CreateLink flag)
  - Reminders (Method: email/popup + Minutes) + Recurrence rules
- [x] FR-F5-03: `ConferenceData.CreateLink = true` → auto create Google Meet link via Calendar API
- [x] FR-F5-04: `CalendarEventResult` includes: HTMLLink, MeetLink, MeetID, ICalUID
- [x] FR-F5-05: `MeetService` standalone: `CreateSpace / GetSpace / GenerateJoinURL`
- [x] FR-F5-06: **⭐ MeetingScheduler Facade**: 1 method = Calendar + Meet dibuatkan bersamaan:
  ```go
  scheduler.ScheduleWithMeet(ctx, &CalendarEvent{ConferenceData: &ConferenceRequest{CreateLink: true}})
  // Returns: MeetingResult{CalendarEvent + MeetSpace (JoinURL + MeetingCode)}
  ```
- [x] FR-F5-07: `RescheduleWithMeet` + `CancelMeeting` for full lifecycle
- [x] FR-F5-08: Authentication via Google Service Account JSON (Domain-Wide Delegation capable)

#### Scopes Required (config default di [config.go](file:///d:/Rancago/config/config.go#L130-L140)):
```
https://www.googleapis.com/auth/calendar
https://www.googleapis.com/auth/drive
https://www.googleapis.com/auth/meetings.space.created
```

---

### 3.6 F6: Database — Generic Repository + pgvector Semantic Search

**Priority**: P0 (Must Have)  
**Reference Files**: [database.go](file:///d:/Rancago/app/Contracts/database.go) + [vector.go](file:///d:/rancago/framework/Database/vector.go) + [models.go](file:///d:/Rancago/app/Models/models.go)

#### FR (Built-in GORM Models — actual di models.go):
- [x] FR-F6-00: 4 Core GORM Models built-in:
  - **User**: ID/Name/Email/Password/AvatarURL/Provider/ProviderID/RememberToken/EmailVerifiedAt + many2many Roles & Permissions
  - **Role**: ID/Name/Label + many2many Permissions
  - **Permission**: ID/Name/Action/Resource (granular: action + resource separation)
  - **Document**: ID/Title/Content + **Embedding Database.Vector (1536d)** dengan **HNSW Index** (opclass: `vector_cosine_ops`) + Metadata JSONB (pgvector production-ready tag)
  ```go
  // Actual GORM tag for pgvector HNSW index (production ready):
  Embedding Database.Vector `gorm:"type:vector(1536);
      index:idx_documents_embedding,type:hnsw,using:hnsw,
      opclass:vector_cosine_ops" json:"-"`
  ```

#### FR (Generic Repository):
- [x] FR-F6-01: Generic interface menggunakan Go 1.23+ Type Parameters:
  ```go
  type Repository[T any, ID any] interface {
      FindByID / FindAll / FindPaginated
      Create / Update / Delete
      FindBy(conditions map[string]interface{})
      FirstOrCreate(conditions, defaults map[string]interface{})
  }
  ```
- [x] FR-F6-02: `PaginationMeta` standard: Page / PerPage / Total / TotalPages / HasNext / HasPrev

#### FR (pgvector Semantic Search):
- [x] FR-F6-03: Custom `Vector` type (GORM-compatible: Scan + Value driver interface)
- [x] FR-F6-04: `VectorRepository[T]` interface — 3 algorithm built-in:
  - **Cosine Similarity** (default untuk semantic search: 0 = tidak mirip, 1 = identik)
  - **L2 Distance (Euclidean)** — untuk geometri/image embedding
  - **Inner Product** — untuk normalized embedding (alias cosine normalized)
- [x] FR-F6-05: `SimilaritySearch` dengan configurable `threshold` (misal 0.75 = filter similarity ≥ 75%)
- [x] FR-F6-06: `UpsertVector(id, embedding, metadata)` dan `DeleteVector(id)` untuk index management
- [x] FR-F6-07: `EnsureExtension()` — otomatis `CREATE EXTENSION IF NOT EXISTS vector;`

#### FR (Transaction Manager):
- [x] FR-F6-08: `Transaction` interface: `Begin / Commit / Rollback / Do(ctx, fn)` dengan auto-rollback on panic/error

#### NFR:
- [x] NFR-F6-01: pgvector index HNSW/IVFFlat support (via raw SQL di repository)
- [x] NFR-F6-02: Similarity search 10k vectors < 50ms (dengan HNSW index)

---

### 3.7 F7: Multi-API Transport Layer (1 Service = 3 Endpoint)

**Priority**: P1 (Should Have)  
**Reference Files**: [notification.go](file:///d:/Rancago/app/Contracts/notification.go) + `framework/Transport/NotificationTransport.go`

#### FR:
- [x] FR-F7-01: **Transport Agnostic Rule**: Business logic di `app/Services/` TIDAK BOLEH import `net/http`, `grpc`, atau `websocket`
- [x] FR-F7-02: 3 Adapters otomatis mengekspos Service yang sama:

| Adapter | Endpoint / Protocol | Format |
|---|---|---|
| **RESTAdapter** | `POST /api/v1/notifications/send` | JSON Body → Service method → JSON Response |
| **GRPCAdapter** | `:9090` gRPC server | Protobuf → Service method → Protobuf |
| **WebSocketAction** | `/ws` connection | JSON Envelope `{action:"notification:send", payload:...}` |

- [x] FR-F7-03: REST routes auto-registrasi via `RegisterRoutes(mux, prefix)`
- [x] FR-F7-04: gRPC auto-registrasi via `RegisterGRPC(server)`
- [x] FR-F7-05: WebSocket auto-dispatch action envelope via Hub

#### Demo Service: NotificationService (app/Services/NotificationService.go)
Concrete struct: `InMemoryNotificationService` — in-memory map store + Redis unread counter cache + WebSocket push.
Container binding (lihat [bootstrap/app.go](file:///d:/Rancago/bootstrap/app.go#L73-L82)):
- Singleton key: `"service.notification"`
- Alias: `"Contracts.NotificationService"` → resolve via interface (DIP compliant)

Methods:
- `Send(ctx, *Notification)` — Kirim ke user tertentu + cache unread count Redis + push WS realtime
- `Broadcast(ctx, title, body, data)` — Broadcast ke ALL WS clients via Hub.Broadcast
- `List(ctx, userID, limit, offset)` — In-memory filter + sort desc by CreatedAt + PaginationMeta
- `MarkRead(ctx, id, userID)` — Update Read flag + decrement Redis unread
- `GetUnreadCount(ctx, userID)` — Redis-first with fallback in-memory count

---

### 3.8 F8: Scalable WebSocket Hub (Multi-Node via Redis PubSub)

**Priority**: P1 (Should Have)  
**Reference Files**: `framework/WebSocket/hub.go`

#### FR:
- [x] FR-F8-01: Hub support 3 tipe message:
  - **Direct**: Kirim ke user ID tertentu
  - **Channel**: Publish/Subscribe ke room channel
  - **Broadcast**: Kirim ke semua connected clients
- [x] FR-F8-02: **Multi-node horizontal scale via Redis Pub/Sub**:
  - Semua `PublishChannel()` otomatis PUBLISH ke Redis key: `Rancago:ws:{channelName}`
  - Setiap instance SUBSCRIBE channel global Redis, lalu re-dispatch ke local clients
  - User A di Server 8080 kirim channel = User B di Server 8081 TERIMA, tanpa sticky session
- [x] FR-F8-03: Handshake parameter: `ws://localhost:8080/ws?user_id=xxx` untuk auto-join user channel
- [x] FR-F8-04: JSON Envelope format standard:
  ```json
  {"type":"channel","channel":"room:123","payload":{"msg":"Halo"}}
  {"type":"notification:new","channel":"user:123","payload":{"title":"Pesan baru"}}
  ```
- [x] FR-F8-05: Heartbeat ping/pong + graceful disconnect
- [x] FR-F8-06: Auto-cleanup user channels saat client disconnect

#### NFR:
- [x] NFR-F8-01: 1 node support ≥ 10K concurrent WS connections
- [x] NFR-F8-02: Latency broadcast < 50ms via Redis PubSub antar node
- [x] NFR-F8-03: Scale to 100 node dengan 0 code change (cuma jalankan instance lebih banyak)

---

### 3.9 F9: CLI Artisan-style (Cobra-based)

**Priority**: P1 (Should Have)  
**Reference Files**: [cmd/Rancago/main.go](file:///d:/Rancago/cmd/Rancago/main.go) + [commands/](file:///d:/Rancago/cmd/Rancago/commands/)

#### FR:

| Category | Command | Fungsi | Status |
|---|---|---|---|
| **Serve** | `serve --port 8080 --grpc` | Jalankan HTTP (opsional gRPC paralel) dengan graceful shutdown | ✅ |
| **Generator** | `make:controller Name [-r]` | Buat controller (flag `-r` = resourceful: Index/Show/Store/Update/Destroy) | ✅ |
| | `make:model Name [-m]` | Buat GORM model (flag `-m` = auto buat migration) | ✅ |
| | `make:middleware Name` | Buat middleware stub dengan log + stopwatch template | ✅ |
| | `make:provider Name` | Buat ServiceProvider stub (Register + Boot) | ✅ |
| **Migration** | `migrate make -n create_xxx` | Generate migration stub di `database/migrations/` | ✅ |
| | `migrate up` | Jalankan semua migration pending (batch tracking) | ✅ |
| | `migrate rollback -s N` | Rollback N batch terakhir | ✅ |
| | `migrate reset` | Drop semua table kecuali migrations | ✅ |
| | `migrate fresh` | Reset + Up ulang (untuk testing/dev) | ✅ |
| | `migrate status` | Tabel status migration ✅ Ran / ❌ Pending | ✅ |
| **Utility** | `key:generate` | Output `APP_KEY=base64:xxxx` untuk env | ✅ |
| | `storage:link` | Symlink `public/storage` → `storage/app/public` | ✅ |
| | `route:list` | Print semua route (REST + gRPC + WS action) | ✅ |
| | `tinker` | Mini REPL dengan Container access | ✅ |

#### CLI Version Info:
```
Version   = "1.0.0"
BuildDate = "2026-07-27"
Module    = github.com/rancago/framework
```

---

### 3.10 F10: Configuration System (Typed Struct)

**Priority**: P0 (Must Have)  
**Reference Files**: [config.go](file:///d:/Rancago/config/config.go)

#### FR:
- [x] FR-F10-01: 100% typed struct configuration — TIDAK ADA `map[string]interface{}` untuk config
- [x] FR-F10-02: `Load()` returns `*Config` dengan default value production-ready
- [x] FR-F10-03: Config Categories:

| Config Struct | Fields Penting | Default Value |
|---|---|---|
| `AppConfig` | Name / Env / Key / URL | Name="Rancago Framework", Env="local" |
| `DatabaseConfig` | Driver/Host/Port/User/Pass/DBName/SSLMode | Postgres localhost:5432 Rancago/Rancago |
| `StorageConfig` | Default + Disks map | default="minio" + 3 disks: minio/s3/google_drive |
| `GoogleConfig` | ClientID/Secret/Redirect/Scopes/Credentials | Calendar/Drive/Meet scopes |
| `RedisConfig` | Host/Port/Password/DB | localhost:6379 DB=0 |
| `AuthConfig` | Providers map[OAuthProviderConfig] | google/github/facebook pre-configured |
| `ServerConfig` | HTTPPort/GRPCPort/WSPort/Debug | 8080/9090/6001 Debug=true |

#### NFR:
- [x] NFR-F10-01: Extensible — tambah config section = tambah typed struct field + default value di `Load()`

---

## 🔌 4. API Contract Specifications

### 4.1 REST API Endpoints (via HTTP :8080)

#### Health & Welcome
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/` | ❌ | Welcome message + version |
| GET | `/api/v1/health` | ❌ | Health check response |

#### Notifications (REST Adapter Registered at `/api/v1/notifications`)
| Method | Path | Auth | Request Body | Response |
|---|---|---|---|---|
| POST | `/send` | ✅ | `{user_id,title,body,channel,data}` | `Notification{id,...}` |
| POST | `/broadcast` | ✅ | `{title,body,data}` | `{status:"ok"}` |
| GET | `/list?user_id=&page=&limit=` | ✅ | Query params | `{data: [...], meta: PaginationMeta}` |
| GET | `/count?user_id=` | ✅ | Query param | `{unread: number}` |
| POST | `/mark_read` | ✅ | `{id,user_id}` | `{status:"ok"}` |

### 4.2 gRPC Service (via :9090)
```protobuf
service NotificationService {
    rpc Send(SendRequest) returns (Notification);
    rpc Broadcast(BroadcastRequest) returns (StatusResponse);
    rpc List(ListRequest) returns (ListResponse);
    rpc MarkRead(MarkReadRequest) returns (StatusResponse);
    rpc Count(CountRequest) returns (CountResponse);
}
```

### 4.3 WebSocket Actions (via /ws)
Connect: `ws://localhost:8080/ws?user_id=123`

**Outgoing (Client → Server)**:
```json
{"action":"notification:send","payload":{"user_id":"456","title":"Halo","body":"Test"}}
{"action":"notification:list","payload":{"limit":5}}
{"action":"notification:count"}
{"action":"notification:mark_read","payload":{"id":"notif_xxx"}}
{"action":"notification:broadcast","payload":{"title":"Broadcast"}}
```

**Incoming (Server → Client Push)**:
```json
{"type":"notification:new","channel":"user:123","payload":{"title":"Order Dikirim!","body":"..."}}
{"type":"broadcast","channel":"broadcast","payload":{"title":"Maintenance 30 menit lagi"}}
```

---

## 🧪 5. Test Strategy & Acceptance Criteria

### 5.1 Unit Test Coverage Target

| Module | Minimum Coverage | Test Focus |
|---|---|---|
| `framework/Container` | 90% | Bind/Singleton/Alias/Resolve thread safety |
| `framework/Storage` | 85% | Manager factory, ProxyDriver, Options pattern |
| `framework/Auth/rbac` | 90% | Role/permission assign, middleware chain |
| `framework/Database` | 80% | Repository CRUD, pgvector similarity |
| `framework/WebSocket` | 75% | Hub channel routing, Redis PubSub bridge |
| `app/Services` | 95% | Business logic, mocking Contracts via interface |

### 5.2 Acceptance Criteria Per Feature

#### AC-F3 (Storage):
- [ ] User dapat upload file ke MinIO via default disk
- [ ] User dapat switch ke Google Drive hanya dengan `Disk("google_drive")` tanpa code lain berubah
- [ ] User dapat tambah LocalFileSystem driver dengan `RegisterFactory` tanpa ubah framework code

#### AC-F4 (Auth + RBAC):
- [ ] Redirect Google OAuth dapat URL valid + state tersimpan di Redis/Cookie
- [ ] Callback dengan state mismatch → error 400
- [ ] User dengan role "admin" tapi tanpa permission "delete:user" → kena block di Middleware

#### AC-F5 (Google Calendar + Meet):
- [ ] `ScheduleWithMeet` dengan `ConferenceData.CreateLink=true` → CalendarEvent terbuat + MeetSpace.JoinURL terisi valid
- [ ] Undangan email attendee otomatis terkirim dari Google Calendar

#### AC-F6 (pgvector):
- [ ] `EnsureExtension` berhasil buat extension vector di PostgreSQL
- [ ] SimilaritySearch dengan threshold=0.8 hanya return similarity ≥ 0.8
- [ ] Upsert 1000 vectors + Search selesai < 100ms

#### AC-F7/F8 (Transport + WS):
- [ ] 1 `NotificationService.Send` call: tercatat di Redis, REST return 200, gRPC return ok, WS push realtime ke user target
- [ ] 2 instance Rancago berjalan paralel: broadcast via channel = user di kedua instance TERIMA message

#### AC-F9 (CLI):
- [ ] `make:controller ProductController -r` generate file dengan 5 method lengkap + imports
- [ ] `migrate fresh` drop + recreate semua table tanpa error

---

## 🚀 6. Deployment & Production Requirements

### 6.1 Infrastructure Stack (Minimum Viable)

| Komponen | Minimum Version | Resource Minimum |
|---|---|---|
| **Go Runtime** | 1.23.4 | — |
| **PostgreSQL** | 14+ dengan extension `vector` | 2 vCPU / 4GB RAM |
| **Redis** | 6+ (enable persistence RDB/AOF) | 1 vCPU / 2GB RAM |
| **MinIO / S3 Compatible** | Latest (opsional tanpa storage feature) | 1 vCPU / 2GB RAM |
| **Load Balancer** | Nginx / HAProxy | — |
| **Google Service Account** | DwD-enabled untuk Calendar impersonation (opsional) | — |

### 6.2 Scaling Checklist Production

1. **PostgreSQL**:
   - Connection Pool: `SetMaxOpenConns(100)`, `SetMaxIdleConns(25)`
   - Enable `pg_stat_statements` extension untuk query analysis
   - pgvector table buat **HNSW index** (bukan IVFFlat) untuk query rendah-latensi

2. **Redis**:
   - Cluster Mode untuk > 1 juta WebSocket clients
   - Enable persistence RDB/AOF untuk RBAC data tidak hilang restart
   - `maxmemory-policy allkeys-lru` untuk cache eviction aman

3. **WebSocket**:
   - Nginx di depan: `ip_hash` sticky session (opsional — Redis PubSub sudah mengatasi tanpa sticky)
   - `Upgrader.CheckOrigin` di-set ketat untuk production (bukan `*`)
   - `gorilla/websocket` `ReadBufferSize`/`WriteBufferSize` disesuaikan

4. **Security**:
   - `APP_KEY` (AppConfig.Key) di-ganti dari default — pakai `Rancago key:generate` output
   - OAuth RedirectURL production pakai HTTPS
   - Storage PresignedURL expiry < 15 menit untuk private file
   - CORS Origin strict list, bukan `*`

5. **Observability**:
   - Export Prometheus metrics:
     - Redis: `INFO stats` connected_clients, used_memory
     - PostgreSQL: `pg_stat_statements` top 10 lambat queries
     - WebSocket: `len(hub.clients)` connection count per instance
   - Structured logging JSON untuk production (bukan default `log.Println`)

---

## 📊 7. Success Metrics (KPIs)

### Business Metrics
| KPI | Target 6 Bulan | Target 1 Tahun | Cara Ukur |
|---|---|---|---|
| Time-to-Market MVP baru | < 3 hari | < 2 hari | Waktu dari idea ke production deploy |
| Feature addition tanpa regression | < 2 bug / feature | < 0.5 bug / feature | Bug tracker ratio |
| Developer Satisfaction Score | > 4.0 / 5.0 | > 4.5 / 5.0 | Survey quarterly |

### Technical Metrics
| KPI | Target | Cara Ukur |
|---|---|---|
| p95 API Latency | < 200ms | APM (New Relic / Datadog) |
| WebSocket Broadcast Latency | < 100ms | Redis PubSub latency + Hub dispatch |
| Container Resolve Time | < 100µs | Benchmark `Container.Resolve` |
| Unit Test Coverage | > 85% | `go test -coverprofile` |
| pgvector Semantic Search (10K vectors) | < 50ms | Benchmark SimilaritySearch |

---

## 🗺 8. Product Roadmap (Prioritized)

### Milestone 1.0 (Current — ✅ Completed di struktur ini)
- [x] IoC Container + ServiceProvider lifecycle
- [x] Storage Manager 3 Drivers (MinIO/S3/GDrive)
- [x] Socialite-style OAuth 3 Providers (Google/GitHub/Facebook)
- [x] Redis RBAC + Policy Middleware
- [x] Generic Repository + pgvector Semantic Search
- [x] Google Calendar + Meet Integration + MeetingScheduler
- [x] NotificationService 3-in-1 Transport (REST/gRPC/WS)
- [x] Scalable WebSocket Hub Redis PubSub
- [x] Artisan-style CLI Generators + Migrations
- [x] Typed Struct Configuration

### Milestone 1.1 (Next — Q3 2026)
- [ ] **Queue Worker System** (Beanstalkd / Redis Streams) — Horizon-style dashboard
- [ ] **Validation Package** — go-playground/validator wrapper dengan Form Request style
- [ ] **Pagination Template** — Cursor-based pagination + JSON:API spec
- [ ] **Logging Package** — Zerolog wrapper structured JSON + trace ID middleware
- [ ] **CORS Middleware** rs/cors terintegrasi default di Router

### Milestone 1.2 (Q4 2026)
- [ ] **GraphQL Adapter** — gqlgen sebagai Transport ke-4 (1 Service = REST/gRPC/WS/GraphQL)
- [ ] **SSE (Server-Sent Events) Adapter** — realtime ringan tanpa WebSocket
- [ ] **Rate Limiter Middleware** — Redis-backed sliding window
- [ ] **Testing Helpers** — Container mock helper, Storage fake driver, WS hub in-memory test

### Milestone 2.0 (2027)
- [ ] **Rancago Admin Panel** — Auto CRUD UI dari GORM Models (seperti Nova/Filament)
- [ ] **Multi-Tenancy Package** — Schema-based atau Column-based multi-tenant ready
- [ ] **Extract `framework/` ke Go Module terpisah** — `go get github.com/rancago/framework`
- [ ] **Rancago Installer** — `Rancago new my-project` clone skeleton + init config

---

## 🔍 9. SOLID Principles Proof Matrix

Rancago DI-DESAIN untuk enforce SOLID, bukan sekadar slogan. Berikut bukti implementasinya di codebase:

| Prinsip | Implementasi Nyata di Rancago | Lokasi File |
|---|---|---|
| **S**ingle Responsibility | `StorageManager` hanya mengatur registry & lazy disk; `MinIODriver` hanya handle MinIO API | [manager.go](file:///d:/rancago/framework/Storage/manager.go) vs [minio.go](file:///d:/rancago/framework/Storage/Drivers/minio.go) |
| **O**pen/Closed | `RegisterFactory()` tambah driver Local/Azure tanpa ubah `StorageManager` 1 baris | [manager.go](file:///d:/rancago/framework/Storage/manager.go#L33-L37) |
| **L**iskov Substitution | `MinIODriver ≡ GDriveDriver` — keduanya implement `StorageDriver` 13 method, bisa swap tanpa code consumer berubah | [storage.go](file:///d:/Rancago/app/Contracts/storage.go#L9-L22) |
| **I**nterface Segregation | `CalendarService` terpisah dari `MeetService` terpisah dari `MeetingScheduler` — tidak ada method yang irrelevant untuk implementor | [google.go](file:///d:/Rancago/app/Contracts/google.go#L8-L99) |
| **D**ependency Inversion | `NotificationService` menerima `*RedisManager` dan `*Hub` lewat constructor, tidak akses global; semua bergantung ke `Contracts.*` interface | `app/Services/NotificationService.go` + [notification.go](file:///d:/Rancago/app/Contracts/notification.go) |

---

## 📚 10. Reference Files Quick Index (for Engineers)

| Konsep | Absolute Path (file://) |
|---|---|
| Application Kernel Bootstrap | [bootstrap/app.go](file:///d:/Rancago/bootstrap/app.go) |
| Container IoC Implementation | [framework/Container/container.go](file:///d:/rancago/framework/Container/container.go) |
| ServiceProvider Interface | [app/Contracts/provider.go](file:///d:/Rancago/app/Contracts/provider.go) |
| StorageDriver Interface (13 method) | [app/Contracts/storage.go](file:///d:/Rancago/app/Contracts/storage.go) |
| StorageManager OCP Registry | [framework/Storage/manager.go](file:///d:/rancago/framework/Storage/manager.go) |
| Auth + RBAC + Socialite Contracts | [app/Contracts/auth.go](file:///d:/Rancago/app/Contracts/auth.go) |
| Calendar + Meet + Scheduler | [app/Contracts/google.go](file:///d:/Rancago/app/Contracts/google.go) |
| Repository + Vector + Transaction | [app/Contracts/database.go](file:///d:/Rancago/app/Contracts/database.go) |
| NotificationService Transport Agnostic | [app/Contracts/notification.go](file:///d:/Rancago/app/Contracts/notification.go) |
| pgvector Type (GORM-compatible) | [framework/Database/vector.go](file:///d:/rancago/framework/Database/vector.go) |
| WebSocket Hub + Redis PubSub | [framework/WebSocket/hub.go](file:///d:/rancago/framework/WebSocket/hub.go) |
| Typed Struct Configuration | [config/config.go](file:///d:/Rancago/config/config.go) |
| Declarative Router Builder | [routes/routes.go](file:///d:/Rancago/routes/routes.go) |
| CLI Root (Cobra Commands) | [cmd/Rancago/main.go](file:///d:/Rancago/cmd/Rancago/main.go) |
| Vibe Code & Convention Guide | [skill.md](file:///d:/Rancago/skill.md) |

---

## ✅ Final Sign-off Criteria (Go-Live Checklist)

- [ ] Semua P0 Features ✅ FR checklist completed
- [ ] Unit test coverage > 85% untuk framework core
- [ ] Integration test passed:
  - Storage MinIO ↔ S3 ↔ GDrive interchangeable test
  - 2-instance WebSocket Pub/Sub end-to-end test
  - OAuth Google → Callback → SocialUser serialization
- [ ] pgvector 10K vector benchmark < 50ms
- [ ] Documentation: README + skill.md + prd.md konsisten dengan actual code structure
- [ ] Security audit:
  - APP_KEY di-replace dari default
  - Storage ACL default = "private"
  - WebSocket CheckOrigin strict
  - RBAC middleware integration test tanpa bypass

---

> 🚀 **Rancago 1.0.0**: Bukan sekadar framework. Ini adalah janji: *produktivitas tanpa mengorbankan prinsip, kecepatan tanpa meninggalkan kualitas.*
> Dari MVP yang selesai 3 hari, ke sistem yang menghandle 1 juta user per hari — tanpa rewrite total. That's the Rancago Promise. 🦫✨
