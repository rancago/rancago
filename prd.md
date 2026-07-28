# 📋 Rancago Framework - Product Requirements Document (PRD)

> **Version**: 1.0.0
> **Last Updated**: 2026-07-28
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

### 2.1 Layered Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                     TRANSPORT LAYER (Entry Points)                │
│  REST HTTP (net/http ServeMux) • gRPC • WebSocket Hub (/ws)      │
│  ↓ Adapters: RESTAdapter / GRPCAdapter / WSAction                │
├──────────────────────────────────────────────────────────────────┤
│                       SERVICE LAYER (Use Cases)                   │
│  app/Services/*.go - Business Logic, Transport-Agnostic          │
│  ↓ Dependencies: HANYA pakai Contracts (interfaces)              │
├──────────────────────────────────────────────────────────────────┤
│                     CONTRACTS LAYER (Abstractions)                │
│  app/Contracts/*.go - ServiceProvider, StorageDriver, etc.       │
│  ← No dependency ke luar - ISP & DIP 100% compliant              │
├──────────────────────────────────────────────────────────────────┤
│                  REPOSITORY / DRIVER LAYER (Implementations)      │
│  framework/Storage/Drivers/ • framework/Auth/Providers/          │
│  framework/Database/ • framework/Google/*.go                     │
├──────────────────────────────────────────────────────────────────┤
│                     INFRASTRUCTURE LAYER (External)               │
│  PostgreSQL + pgvector • Redis • MinIO/S3/GDrive • Google APIs   │
└──────────────────────────────────────────────────────────────────┘
```

### 2.2 Physical Folder Structure

```
rancago/
├── app/
│   ├── Contracts/          # Semua interface definitions
│   │   ├── auth.go         # AuthProvider + RBAC + SocialManager
│   │   ├── database.go     # Repository<T> + VectorRepository + Transaction
│   │   ├── google.go       # Calendar + Meet + Scheduler
│   │   ├── notification.go # NotificationService (Transport Agnostic)
│   │   ├── provider.go     # ServiceProvider interface
│   │   └── storage.go      # StorageDriver + Options Pattern
│   ├── Models/
│   │   ├── models.go       # User + Role + Permission + Document (GORM + pgvector HNSW 1536d)
│   │   └── vector.go       # pgvector custom type
│   ├── Providers/
│   │   ├── AuthServiceProvider.go
│   │   ├── GoogleServiceProvider.go
│   │   └── StorageServiceProvider.go
│   ├── Repositories/
│   │   └── repository.go   # Generic GORM repository implementation
│   └── Services/
│       └── NotificationService.go
├── bootstrap/
│   └── app.go              # New() → RegisterCore() → Boot() → StartHTTP/gRPC
├── cmd/rancago/
│   └── main.go             # CLI entry point (hand-rolled dispatcher)
├── config/
│   └── config.go           # Typed struct configuration + Load()
├── database/
│   ├── migrations/         # Timestamped migration files
│   └── seeders/
├── docs/
│   └── features/           # Per-feature context docs (auto-generated)
├── framework/
│   ├── Auth/               # OAuth Socialite + RBAC + Policy
│   │   ├── Providers/      # Generic OAuth provider wrapper
│   │   ├── rbac.go
│   │   └── socialite.go
│   ├── Cache/
│   │   └── redis.go        # Connection + PubSub + RateLimiter
│   ├── Container/
│   │   └── container.go    # Bind/Singleton/Instance/Alias/Resolve
│   ├── Database/           # (legacy layer — use internal/ports for new features)
│   ├── Google/
│   │   ├── calendar.go
│   │   └── meet.go
│   ├── Storage/
│   │   ├── Drivers/
│   │   │   └── memory.go
│   │   ├── manager.go
│   │   └── memory.go
│   ├── Transport/
│   │   └── NotificationTransport.go
│   └── WebSocket/
│       └── hub.go
├── internal/               # Hexagonal architecture layer (canonical new features)
│   ├── adapters/
│   │   ├── driven/         # DB, cache, storage, auth adapters
│   │   └── driving/        # HTTP, gRPC, CLI adapters
│   ├── application/
│   │   └── usecases/       # Use case interactors
│   ├── bootstrap/
│   │   └── app.go
│   ├── domain/
│   │   ├── entities/
│   │   ├── errors/
│   │   └── valueobjects/
│   ├── kernel/
│   │   ├── config.go
│   │   └── container.go
│   └── ports/
│       ├── driven/         # Outbound port interfaces
│       └── driving/        # Inbound port interfaces
├── go.mod                  # module github.com/rancago/framework (Go 1.23.4)
├── prd.md                  # This document
└── skill.md                # Coding conventions & vibe code guide
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
**Reference Files**: `framework/Container/container.go` + `app/Contracts/provider.go`

#### Requirements Functional (FR):
- [x] FR-F1-01: Support 3 binding types: **Singleton**, **Transient**, **Instance**
- [x] FR-F1-02: Support **Alias** mapping (e.g. `Contracts.NotificationService` → `service.notification`)
- [x] FR-F1-03: `Resolve(abstract string) (interface{}, error)` + `MustResolve(abstract string) interface{}`
- [x] FR-F1-04: Thread-safe dengan `sync.RWMutex`
- [x] FR-F1-05: `Call(fn interface{})` — auto-resolve parameter function via reflection
- [x] FR-F1-06: `Has(abstract string) bool`

#### Non-Functional Requirements (NFR):
- [x] NFR-F1-01: Resolve time < 50µs per binding (cached singleton)
- [x] NFR-F1-02: Zero allocation untuk Singleton resolve setelah instantiasi pertama

---

### 3.2 F2: ServiceProvider Registration System

**Priority**: P0 (Must Have)
**Reference Files**: `app/Contracts/provider.go` + `bootstrap/app.go`

#### FR:
- [x] FR-F2-01: Setiap module implement `ServiceProvider` interface:
  ```go
  type ServiceProvider interface {
      Register(app *Container.Container) error
      Boot(app *Container.Container) error
  }
  ```
- [x] FR-F2-02: `Application.RegisterProviders(...)` menerima variadic ServiceProvider
- [x] FR-F2-03: Semua `Register()` dijalankan DULU sebelum semua `Boot()`
- [x] FR-F2-04: `RegisterCore()` bootstrap default 3 providers: Storage, Google, Auth

#### NFR:
- [x] NFR-F2-01: Provider gagal Register/Boot → `log.Fatalf` (early failure)

---

### 3.3 F3: Multi-Storage Manager

**Priority**: P0 (Must Have)
**Reference Files**: `app/Contracts/storage.go` + `framework/Storage/manager.go`

#### FR:
- [x] FR-F3-01: `StorageDriver` interface 13 method: Put/Get/Delete/Exists/Size/LastModified/Copy/Move/List/URL/TemporaryURL/Name
- [x] FR-F3-02: Functional Options Pattern: `WithContentType / WithACL / WithMetadata / WithVisibility`
- [x] FR-F3-03: 3 Drivers: **MinIO** + **Amazon S3** + **Google Drive**
- [x] FR-F3-04: Lazy Disk Initialization
- [x] FR-F3-05: `RegisterFactory(driverType, factory)` untuk tambah driver tanpa ubah Manager
- [x] FR-F3-06: Proxy Pattern — `Proxy(name)` returns `StorageDriver` untuk constructor injection
- [x] FR-F3-07: Default disk configurable via `StorageConfig.Default`
- [x] FR-F3-08: Container binding key: `"storage"`, alias `"Contracts.StorageDriver"`

---

### 3.4 F4: OAuth Socialite + RBAC Policy

**Priority**: P0 (Must Have)
**Reference Files**: `app/Contracts/auth.go` + `framework/Auth/socialite.go` + `framework/Auth/rbac.go`

#### FR (Socialite OAuth):
- [x] FR-F4-01: `AuthProvider` interface: Name/Redirect/Callback/UserFromToken
- [x] FR-F4-02: `SocialUser` struct seragam: Provider/ID/Email/Name/Avatar/Token/RefreshToken/RawAttributes
- [x] FR-F4-03: 3 Providers default: **Google** + **GitHub** + **Facebook**
- [x] FR-F4-04: `RegisterDriver(name, factory)` untuk tambah provider custom

#### FR (RBAC Redis-backed):
- [x] FR-F4-05: Role: `AssignRole / RemoveRole / HasRole / HasAnyRole / HasAllRoles / GetRoles`
- [x] FR-F4-06: Permission: `GivePermission / RevokePermission / HasPermission`
- [x] FR-F4-07: Middleware: `Middleware(permission)` + `RoleMiddleware(roles...)`
- [x] FR-F4-08: Redis key convention: `rbac:user:{id}:roles` (SET), `rbac:role:{name}:permissions` (SET)

---

### 3.5 F5: Google Ecosystem Integration

**Priority**: P1 (Should Have)
**Reference Files**: `app/Contracts/google.go` + `framework/Google/calendar.go` + `framework/Google/meet.go`

#### FR:
- [x] FR-F5-01: `CalendarService`: CRUD Events + AddAttendees + ListEvents
- [x] FR-F5-02: `ConferenceData.CreateLink = true` → auto-create Google Meet link
- [x] FR-F5-03: `MeetService`: CreateSpace / GetSpace / GenerateJoinURL
- [x] FR-F5-04: `MeetingScheduler.ScheduleWithMeet` → 1 call = Calendar event + Meet space
- [x] FR-F5-05: Auth via Google Service Account JSON

---

### 3.6 F6: Database — Generic Repository + pgvector

**Priority**: P0 (Must Have)
**Reference Files**: `app/Contracts/database.go` + `app/Models/vector.go` + `app/Models/models.go`

#### FR:
- [x] FR-F6-01: Generic `Repository[T any, ID any]` interface dengan CRUD + FindBy + FirstOrCreate
- [x] FR-F6-02: `PaginationMeta`: Page/PerPage/Total/TotalPages/HasNext/HasPrev
- [x] FR-F6-03: Custom `Vector` GORM-compatible type (Scan + Value)
- [x] FR-F6-04: `VectorRepository[T]`: CosineSimilarity / L2Distance / InnerProduct / SimilaritySearch
- [x] FR-F6-05: Document model dengan `Embedding vector(1536)` + HNSW index `vector_cosine_ops`
- [x] FR-F6-06: `Transaction.Do(ctx, fn)` dengan auto-rollback

---

### 3.7 F7: Multi-API Transport Layer

**Priority**: P1 (Should Have)
**Reference Files**: `app/Contracts/notification.go` + `framework/Transport/NotificationTransport.go`

#### FR:
- [x] FR-F7-01: Business logic di `app/Services/` TIDAK BOLEH import `net/http`, `grpc`, atau websocket
- [x] FR-F7-02: 3 Adapters untuk service yang sama: RESTAdapter + GRPCAdapter + WebSocketAction
- [x] FR-F7-03: REST via `RegisterRoutes(mux, prefix)`, gRPC via `RegisterGRPC(server)`

---

### 3.8 F8: Scalable WebSocket Hub

**Priority**: P1 (Should Have)
**Reference Files**: `framework/WebSocket/hub.go`

#### FR:
- [x] FR-F8-01: 3 message type: Direct (user) / Channel (room) / Broadcast (all)
- [x] FR-F8-02: Multi-node scale via Redis Pub/Sub — `Rancago:ws:{channelName}`
- [x] FR-F8-03: Handshake: `ws://host/ws?user_id=xxx`
- [x] FR-F8-04: JSON envelope: `{"type":"...","channel":"...","payload":{...}}`
- [x] FR-F8-05: Heartbeat ping/pong + graceful disconnect

#### NFR:
- [x] NFR-F8-01: ≥ 10K concurrent WS connections per node
- [x] NFR-F8-02: Broadcast latency < 50ms via Redis PubSub
- [x] NFR-F8-03: Scale horizontal dengan 0 code change

---

### 3.9 F9: CLI (Rancago-style, hand-rolled)

**Priority**: P1 (Should Have)
**Reference Files**: `cmd/rancago/main.go` + `internal/adapters/driving/cli/`

#### FR:

| Category | Command | Fungsi | Status |
|---|---|---|---|
| **Serve** | `serve [--port] [--grpc]` | HTTP + optional gRPC, graceful shutdown | ✅ |
| **Scaffold** | `make:feature <Name>` | Full hexagonal scaffold + FEATURE.md | ✅ |
| | `scaffold <Name>` | Interactive bounded context scaffolder | ✅ |
| **Generator** | `make:entity <Name>` | Domain entity | ✅ |
| | `make:port <Name> [--driving]` | Port interface | ✅ |
| | `make:usecase <Name>` | Use case interactor | ✅ |
| | `make:adapter <Name> [--direction]` | Driving/driven adapter | ✅ |
| | `make:model <Name> [-m]` | GORM model + optional migration | ✅ |
| | `make:migration <name>` | Timestamped migration file | ✅ |
| **Utility** | `key:generate` | Output `APP_KEY=base64:xxxx` | ✅ |
| | `storage:link` | Symlink `public/storage` → `storage/app/public` | ✅ |
| | `route:list` | Print semua route | ✅ |
| | `tinker` | Mini REPL dengan Container/port info | ✅ |
| | `migrate` | Run migrations (stub, swap adapter untuk real DB) | ✅ |

---

### 3.10 F10: Configuration System

**Priority**: P0 (Must Have)
**Reference Files**: `config/config.go` + `internal/kernel/config.go`

#### FR:
- [x] FR-F10-01: 100% typed struct — tidak ada `map[string]interface{}` untuk config
- [x] FR-F10-02: `Load()` returns `*Config` dengan default value siap pakai
- [x] FR-F10-03: Config structs: AppConfig / DatabaseConfig / StorageConfig / GoogleConfig / RedisConfig / AuthConfig / ServerConfig

---

## 🔌 4. API Contract Specifications

### 4.1 REST Endpoints (HTTP :8080)

| Method | Path | Description |
|---|---|---|
| GET | `/` | Welcome + version |
| GET | `/api/v1/health` | Health check |
| POST | `/api/v1/notifications/send` | `{user_id,title,body,channel,data}` |
| POST | `/api/v1/notifications/broadcast` | `{title,body,data}` |
| GET | `/api/v1/notifications/list?user_id=&page=&per_page=` | Paginated list |
| GET | `/api/v1/notifications/count?user_id=` | Unread count |
| POST | `/api/v1/notifications/read` | `{id,user_id}` |
| GET | `/ws?user_id=` | WebSocket upgrade |

### 4.2 gRPC Service (:9090)
```protobuf
service NotificationService {
    rpc Send(SendRequest) returns (Notification);
    rpc Broadcast(BroadcastRequest) returns (StatusResponse);
    rpc List(ListRequest) returns (ListResponse);
    rpc MarkRead(MarkReadRequest) returns (StatusResponse);
    rpc Count(CountRequest) returns (CountResponse);
}
```

### 4.3 WebSocket Envelope

**Client → Server**:
```json
{"action":"notification:send","payload":{"user_id":"456","title":"Halo","body":"Test"}}
{"action":"notification:broadcast","payload":{"title":"Broadcast"}}
```

**Server → Client**:
```json
{"type":"notification:new","channel":"user:123","payload":{"title":"Order Dikirim!"}}
{"type":"broadcast","channel":"broadcast","payload":{"title":"Maintenance 30 menit"}}
```

---

## 🧪 5. Test Strategy & Acceptance Criteria

### 5.1 Unit Test Coverage Target

| Module | Minimum Coverage | Test Focus |
|---|---|---|
| `framework/Container` | 90% | Bind/Singleton/Alias/Resolve thread safety |
| `framework/Storage` | 85% | Manager factory, ProxyDriver, Options |
| `framework/Auth/rbac` | 90% | Role/permission assign, middleware chain |
| `internal/domain` | 90% | Entity logic, value object validation |
| `internal/application/usecases` | 85% | Use case orchestration via mock ports |
| `framework/WebSocket` | 75% | Hub routing, Redis PubSub bridge |

### 5.2 Acceptance Criteria

#### AC-F3 (Storage):
- [ ] Upload file ke MinIO via default disk
- [ ] Switch ke Google Drive dengan `Disk("google_drive")` tanpa code lain berubah
- [ ] Tambah driver LocalFileSystem via `RegisterFactory` tanpa ubah framework

#### AC-F4 (Auth + RBAC):
- [ ] OAuth redirect → URL valid + state di Redis
- [ ] State mismatch callback → 400
- [ ] Role "admin" tanpa permission "delete:user" → diblock middleware

#### AC-F6 (pgvector):
- [ ] `EnsureExtension` buat extension vector di PostgreSQL
- [ ] SimilaritySearch threshold=0.8 hanya return ≥ 0.8
- [ ] 1000 vectors upsert + search < 100ms

#### AC-F8 (WebSocket):
- [ ] 2 instance paralel: broadcast via channel diterima di kedua instance
- [ ] Disconnect → cleanup channel otomatis

#### AC-F9 (CLI):
- [ ] `make:feature Order` generate semua file hexagonal + `docs/features/order.md`
- [ ] `make:migration create_orders_table` buat file dengan timestamp

---

## 🚀 6. Deployment & Production Requirements

### 6.1 Infrastructure Stack

| Komponen | Minimum Version | Resource |
|---|---|---|
| **Go Runtime** | 1.23.4 | — |
| **PostgreSQL** | 14+ + extension `vector` | 2 vCPU / 4GB RAM |
| **Redis** | 6+ (RDB/AOF persistence) | 1 vCPU / 2GB RAM |
| **MinIO / S3 Compatible** | Latest | 1 vCPU / 2GB RAM |
| **Load Balancer** | Nginx / HAProxy | — |

### 6.2 Production Checklist

- [ ] `APP_KEY` diganti dari default (`rancago key:generate`)
- [ ] OAuth RedirectURL pakai HTTPS
- [ ] Storage ACL default = `"private"`
- [ ] WebSocket `CheckOrigin` strict (bukan `*`)
- [ ] CORS origin strict list
- [ ] PostgreSQL HNSW index di-create sebelum production load
- [ ] Redis `maxmemory-policy allkeys-lru`
- [ ] Structured logging JSON (bukan `log.Println`)

---

## 📊 7. Success Metrics (KPIs)

### Business Metrics
| KPI | Target 6 Bulan | Target 1 Tahun |
|---|---|---|
| Time-to-Market MVP baru | < 3 hari | < 2 hari |
| Feature tanpa regression | < 2 bug / feature | < 0.5 bug / feature |
| Developer Satisfaction | > 4.0 / 5.0 | > 4.5 / 5.0 |

### Technical Metrics
| KPI | Target |
|---|---|
| p95 API Latency | < 200ms |
| WebSocket Broadcast Latency | < 100ms |
| Container Resolve Time | < 100µs |
| Unit Test Coverage | > 85% |
| pgvector Search (10K vectors) | < 50ms |

---

## 🗺 8. Product Roadmap

### Milestone 1.0 (Current ✅)
- [x] IoC Container + ServiceProvider lifecycle
- [x] Storage Manager 3 Drivers (MinIO/S3/GDrive)
- [x] OAuth Socialite 3 Providers (Google/GitHub/Facebook)
- [x] Redis RBAC + Policy Middleware
- [x] Generic Repository + pgvector Semantic Search
- [x] Google Calendar + Meet + MeetingScheduler
- [x] NotificationService 3-in-1 Transport (REST/gRPC/WS)
- [x] Scalable WebSocket Hub Redis PubSub
- [x] Hexagonal architecture layer (`internal/`)
- [x] CLI Generators + Feature Doc auto-generation
- [x] Typed Struct Configuration

### Milestone 1.1 (Q3 2026)
- [ ] Queue Worker System (Redis Streams) — Horizon-style
- [ ] Validation Package (go-playground/validator wrapper)
- [ ] Cursor-based Pagination + JSON:API spec
- [ ] Structured Logging (zerolog) + trace ID middleware
- [ ] CORS Middleware bawaan

### Milestone 1.2 (Q4 2026)
- [ ] GraphQL Adapter (gqlgen) — Transport ke-4
- [ ] SSE (Server-Sent Events) Adapter
- [ ] Rate Limiter Middleware (Redis sliding window)
- [ ] Testing Helpers — Container mock, Storage fake driver

### Milestone 2.0 (2027)
- [ ] Rancago Admin Panel — Auto CRUD UI dari GORM Models
- [ ] Multi-Tenancy Package (schema/column-based)
- [ ] Extract `framework/` ke Go module terpisah
- [ ] `rancago new my-project` skeleton installer

---

## 🔍 9. SOLID Principles Proof Matrix

| Prinsip | Implementasi Nyata | Lokasi |
|---|---|---|
| **S**ingle Responsibility | `StorageManager` hanya registry + lazy disk; `MinIODriver` hanya MinIO API | `framework/Storage/manager.go` vs `Drivers/` |
| **O**pen/Closed | `RegisterFactory()` tambah driver tanpa ubah `StorageManager` | `framework/Storage/manager.go` |
| **L**iskov Substitution | `MinIODriver ≡ GDriveDriver` — implement `StorageDriver`, swap tanpa ubah consumer | `app/Contracts/storage.go` |
| **I**nterface Segregation | `CalendarService` terpisah dari `MeetService` terpisah dari `MeetingScheduler` | `app/Contracts/google.go` |
| **D**ependency Inversion | Use cases menerima driven ports via constructor — tidak ada global state | `internal/application/usecases/` |

---

## 📚 10. Reference Files Quick Index

| Konsep | Path |
|---|---|
| Bootstrap Kernel | `bootstrap/app.go` |
| Internal Bootstrap (hexagonal) | `internal/bootstrap/app.go` |
| Container IoC | `framework/Container/container.go` |
| Internal Container | `internal/kernel/container.go` |
| ServiceProvider Interface | `app/Contracts/provider.go` |
| StorageDriver Interface | `app/Contracts/storage.go` |
| StorageManager | `framework/Storage/manager.go` |
| Auth + RBAC + Socialite | `app/Contracts/auth.go` |
| Google Calendar + Meet | `app/Contracts/google.go` |
| Repository + Vector + Transaction | `app/Contracts/database.go` |
| NotificationService Contract | `app/Contracts/notification.go` |
| pgvector Type | `app/Models/vector.go` |
| WebSocket Hub | `framework/WebSocket/hub.go` |
| Typed Config (framework) | `config/config.go` |
| Typed Config (internal) | `internal/kernel/config.go` |
| CLI Dispatcher | `internal/adapters/driving/cli/cli.go` |
| CLI Commands | `internal/adapters/driving/cli/commands/` |
| Domain Entities | `internal/domain/entities/` |
| Driven Ports | `internal/ports/driven/` |
| Driving Ports | `internal/ports/driving/` |
| Feature Docs | `docs/features/` |
| Vibe Code Guide | `skill.md` |

---

## ✅ Final Sign-off Criteria

- [ ] Semua P0 Features FR checklist completed
- [ ] Unit test coverage > 85% untuk framework core
- [ ] Integration test: Storage 3 driver interchangeable
- [ ] Integration test: 2-instance WebSocket Pub/Sub end-to-end
- [ ] pgvector 10K vector benchmark < 50ms
- [ ] `prd.md` + `skill.md` + `docs/features/` konsisten dengan actual codebase
- [ ] Security: APP_KEY replaced, ACL private default, CheckOrigin strict

---

> 🚀 **Rancago 1.0.0**: Produktivitas tanpa mengorbankan prinsip. Kecepatan tanpa meninggalkan kualitas. Dari MVP 3 hari ke sistem 1 juta user — tanpa rewrite total.
