# RANCAGO Framework

> **Resilient, Agnostic, & Native Clean-Architecture GO Framework**

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![Module](https://img.shields.io/badge/module-github.com%2Francago%2Fframework-blue?style=flat-square)](https://github.com/rancago/rancago)
[![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)](LICENSE)

SOLID Principles · IoC Service Container · Multi-API Transport · pgvector Semantic Search · Redis WebSocket Hub · Google Ecosystem · OAuth Socialite

---

## The Name

**RANCAGO** carries two layers of meaning.

**Official acronym:**

> **R**esilient, **A**gnostic, & **N**ative **C**lean-**A**rchitecture **G**O Framework

Built to be resilient under load, transport-agnostic (REST + gRPC + WebSocket from one service definition), and natively idiomatic to Go's clean-architecture patterns.

**Sundanese / local roots:**

| Word | Meaning |
|---|---|
| **Rancagé** | *Skilled, precise, and structured craftsmanship* - reflects how Rancago enforces SOLID principles and clean hexagonal architecture step by step. |
| **Ranca** | *A fertile, expansive wetland ecosystem* - mirrors the rich built-in feature set: pgvector, Redis, MinIO, Google Drive, Meet, Calendar, OAuth, RBAC, all growing from one foundation. |

The name is both a technical acronym and a tribute to local culture - a framework engineered with care (*rancagé*) and designed to grow like a fertile ecosystem (*ranca*).

---

## Why Rancago?

| Dimension | Vanilla Go | Laravel PHP | **Rancago** |
|---|---|---|---|
| **Speed** | ⚡⚡⚡ Native | 🐢 Moderate | ⚡⚡⚡ Native Go |
| **DX / Productivity** | Manual wiring, days of setup | Artisan, Facades out of the box | CLI generators + Contracts-First |
| **SOLID / Clean Arch** | Unconstrained, often over-coupled | Facades break DIP | **Contracts-First enforced** |
| **Realtime & Scale** | Manual Redis PubSub | Pusher (paid external) | Built-in multi-node WebSocket Hub |
| **Multi-API Transport** | REST only | REST + queued jobs | **1 Service = REST + gRPC + WebSocket** |
| **AI-Ready** | Manual pgvector setup | Third-party packages needed | Semantic search built-in |

---

## Features

| Category | What's included |
|---|---|
| 🏛 **Architecture** | SOLID Principles · IoC Service Container · ServiceProvider lifecycle · Hexagonal Ports & Adapters |
| 💾 **Database** | PostgreSQL + GORM · **pgvector** semantic search (Cosine / L2 / Inner Product) · Generic Repository · Pagination · Transaction Manager |
| 📦 **Storage** | **MinIO** · **Amazon S3** · **Google Drive** · Lazy disk init · Functional Options · Temporary presigned URLs · OCP `RegisterFactory` extension point |
| 📅 **Google Ecosystem** | Calendar API (CRUD + attendees + reminders) · Google Meet auto-link · **MeetingScheduler** facade (1 call = event + Meet link) |
| 🔐 **Auth** | Socialite-style OAuth: **Google / GitHub / Facebook / Custom OIDC** · Redis-backed RBAC (roles + permissions + policy middleware) · Bearer token middleware |
| ⚡ **Cache & Realtime** | Redis Manager (Get/Set/SAdd/PubSub) · Rate limiter · **Scalable WebSocket Hub** (multi-node via Redis Pub/Sub - broadcast / channel / direct) |
| 🎯 **Transport** | **1 service = 3 endpoints**: REST HTTP + gRPC + WebSocket - zero code duplication |
| 🛠 **CLI** | `serve · migrate · scaffold · make:entity/value-object/port/usecase/adapter/model/migration · key:generate · storage:link · route:list · tinker` |

---

## Project Structure

```
rancago/
├── app/
│   ├── Contracts/          # ⭐ All interface definitions (DIP compliance)
│   ├── Models/             # GORM models + pgvector Vector type
│   ├── Providers/          # ServiceProvider implementations (Register + Boot)
│   ├── Services/           # Transport-agnostic business logic
│   ├── Repositories/       # Data access layer
│   └── Http/               # Controllers, Middleware, Requests
├── framework/              # ⭐ Extractable core framework
│   ├── Container/          # IoC Service Container
│   ├── Auth/Providers/     # OAuth Socialite + RBAC + Policy Enforcer
│   ├── Cache/              # Redis Manager + PubSub + Rate Limiter
│   ├── Database/           # Generic Repository + pgvector + Transaction
│   ├── Google/             # Calendar + Meet + MeetingScheduler
│   ├── Storage/Drivers/    # MinIO · S3 · Google Drive · Memory
│   ├── Transport/          # REST / gRPC / WebSocket adapters
│   └── WebSocket/          # Scalable Hub (Redis Pub/Sub)
├── internal/               # Hexagonal ports & adapters (domain layer)
│   ├── domain/             # Entities, value objects, domain errors
│   ├── ports/              # Driven + driving port interfaces
│   ├── application/        # Use case interactors
│   ├── adapters/           # HTTP, gRPC, CLI, cache, persistence adapters
│   ├── bootstrap/          # Internal wiring
│   └── kernel/             # IoC container + config (internal)
├── config/                 # Typed configuration (no map[string]interface{})
├── routes/                 # Declarative route builder
├── bootstrap/              # Application kernel (ServiceProvider wiring)
├── database/migrations/    # Migration files
├── cmd/rancago/            # CLI entry point
└── main.go                 # Server entry point
```

---

## Quick Start

### Prerequisites

- Go 1.23+
- PostgreSQL 14+ with the `vector` extension (`CREATE EXTENSION IF NOT EXISTS vector;`)
- Redis 6+
- MinIO _(optional, for object storage)_
- Google Service Account JSON _(optional, for Calendar / Meet / Drive)_

### 1. Clone & install

```bash
git clone https://github.com/rancago/rancago.git
cd rancago
go mod tidy
```

### 2. Configure

Edit `config/config.go` and update the defaults, or override with environment variables:

```go
Database: DatabaseConfig{
    Host:     "localhost",
    Port:     5432,
    User:     "rancago",
    Password: "rancago",
    DBName:   "rancago_db",
},
Redis: RedisConfig{Host: "localhost", Port: 6379},
Auth: AuthConfig{
    Providers: map[string]Auth.OAuthConfig{
        "google": {
            ClientID:    "your-client-id",
            ClientSecret: "your-secret",
            RedirectURL: "http://localhost:8080/auth/google/callback",
        },
    },
},
```

### 3. Run

```bash
# HTTP + gRPC + WebSocket in one process
go run .

# Or via the CLI
go run ./cmd/rancago serve --port 8080
go run ./cmd/rancago serve --grpc --port 8080
```

### 4. Verify

```bash
curl http://localhost:8080/api/v1/health
# {"status":"healthy","service":"rancago-api"}

curl http://localhost:8080/
# {"message":"Welcome to Rancago Framework 🚀","version":"1.0.0"}
```

---

## Core Concepts

### The Golden Rule: Contracts → Providers → Services

```
app/Contracts/*.go   - define all interfaces first (DIP, ISP compliant)
     ↓
app/Providers/*.go   - bind concrete implementations into the container
     ↓
app/Services/*.go    - transport-agnostic business logic, depend on Contracts only
```

### IoC Container

```go
// Bind types
c.Singleton("service.payment", func(c *Container.Container) (interface{}, error) {
    storage, _ := c.Resolve("Contracts.StorageDriver")
    return Services.NewPaymentService(storage.(Contracts.StorageDriver)), nil
})
c.Alias("service.payment", "Contracts.PaymentService")

// Resolve
svc := container.MustResolve("Contracts.PaymentService").(Contracts.PaymentService)
```

### ServiceProvider pattern

```go
type PaymentServiceProvider struct{}

func (p *PaymentServiceProvider) Register(c *Container.Container) error {
    c.Singleton("service.payment", func(c *Container.Container) (interface{}, error) {
        return Services.NewPaymentService(...), nil
    })
    c.Alias("service.payment", "Contracts.PaymentService")
    return nil
}

func (p *PaymentServiceProvider) Boot(c *Container.Container) error {
    // Register drivers, seed data, run migrations
    return nil
}
```

Register in `bootstrap/app.go`:

```go
app.RegisterProviders(
    Providers.NewStorageServiceProvider(...),
    Providers.NewPaymentServiceProvider(),
)
```

---

## Module Guides

### Storage - MinIO · S3 · Google Drive

Depend on the interface, not the concrete driver:

```go
type FileService struct {
    Storage Contracts.StorageDriver // injected via container
}

func (s *FileService) Upload(ctx context.Context, userID string, r io.Reader) (string, error) {
    path := fmt.Sprintf("avatars/%s.jpg", userID)
    err := s.Storage.Put(ctx, path, r,
        Contracts.WithContentType("image/jpeg"),
        Contracts.WithACL("public-read"),
    )
    if err != nil {
        return "", err
    }
    return s.Storage.URL(ctx, path)
}
```

Add a new driver without touching the Manager (100% OCP):

```go
mgr.RegisterFactory("azure_blob", func(cfg Contracts.StorageDiskConfig) (Contracts.StorageDriver, error) {
    return NewAzureBlobDriver(cfg), nil
})
disk, _ := mgr.Disk("azure_blob")
```

### Google Calendar + Meet

One method schedules the event and creates the Meet link:

```go
result, err := scheduler.ScheduleWithMeet(ctx, &Contracts.CalendarEvent{
    Summary:  "Sprint Review",
    Start:    time.Now().Add(24 * time.Hour),
    End:      time.Now().Add(25 * time.Hour),
    Timezone: "Asia/Jakarta",
    Attendees: []Contracts.Attendee{
        {Email: "team@example.com", DisplayName: "Team"},
    },
    ConferenceData: &Contracts.ConferenceRequest{
        Type:       "hangoutsMeet",
        CreateLink: true,
    },
})
fmt.Println(result.MeetSpace.JoinURL) // https://meet.google.com/xxx-yyy-zzz
```

### OAuth Socialite

```go
// Redirect user to provider
authURL, state, _ := socialite.Redirect(ctx, "google")
http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)

// Handle callback - returns a unified SocialUser regardless of provider
user, _ := socialite.Callback(ctx, "google", code, state)
fmt.Println(user.Email, user.Name, user.AvatarURL)
```

Add a custom provider without changing SocialiteManager:

```go
mgr.RegisterDriver("keycloak", func() (Contracts.AuthProvider, error) {
    return Auth.NewGenericOAuthProvider("keycloak", Auth.OAuthConfig{
        AuthURL:     "https://sso.company.com/auth",
        TokenURL:    "https://sso.company.com/token",
        UserInfoURL: "https://sso.company.com/userinfo",
        Scopes:      []string{"openid", "email", "profile"},
    }), nil
})
```

### Redis RBAC

```go
rbac.AssignRole(ctx, "user-123", "admin")
rbac.GivePermissionToRole(ctx, "admin", "delete:user")

// HTTP middleware - blocks requests without the required permission
mux.Handle("/admin/users",
    rbac.Middleware("delete:user")(http.HandlerFunc(handler)))

// Role-based middleware
mux.Handle("/admin/dashboard",
    rbac.RoleMiddleware("admin", "superadmin")(http.HandlerFunc(handler)))
```

### Notifications - 1 service, 3 transports

Write the business logic once in `app/Services/NotificationService.go`. The same code is exposed via REST, gRPC, and WebSocket automatically.

```bash
# REST
curl -X POST http://localhost:8080/api/v1/notifications/send \
  -H "Content-Type: application/json" \
  -d '{"user_id":"123","title":"Order shipped","body":"Your order #8872 is on the way","channel":"push"}'

# WebSocket (wscat)
wscat -c "ws://localhost:8080/ws?user_id=123"
> {"action":"notification:send","payload":{"user_id":"456","title":"Hello","body":"Test"}}
< {"type":"notification:new","channel":"user:456","payload":{...}}
```

### Multi-node WebSocket (Redis Pub/Sub)

Every `PublishChannel` call publishes to `rancago:ws:{channel}` in Redis. All running instances subscribe and relay messages to their local clients - no sticky sessions required.

```bash
# Node 1
go run . # HTTP :8080

# Node 2 (different terminal, change HTTPPort to 8081)
go run . # HTTP :8081
```

User A on `:8080` sends to `room:123` → User B on `:8081` receives it automatically.

### pgvector Semantic Search

```go
// Cosine similarity search - returns documents most semantically similar to the query
threshold := float64(0.75)
results, _ := docRepo.SimilaritySearch(ctx, queryEmbedding, 10, &threshold)
for _, hit := range results {
    fmt.Printf("[%.1f%%] %s\n", hit.Score*100, hit.Item.Title)
}
```

---

## CLI Reference

```
rancago [command] [flags]

SERVE
  serve                        Start HTTP (+ optional gRPC) server
    --port, -p  int            HTTP port (default: 8080)
    --grpc                     Also start gRPC stub server

GENERATORS
  make:entity     [name]       Domain entity
  make:value-object [name]     Value object
  make:port       [name]       Driving/driven port interface (--driving flag)
  make:usecase    [name]       Use case interactor
  make:adapter    [name]       Infrastructure adapter (--direction driven|driving)
  make:model      [name] [-m]  GORM model (+ migration with -m)
  make:migration  [name]       Migration file stub

SCAFFOLD
  scaffold [name]              Interactive bounded-context scaffolder
                               (entity + port + usecase + adapter in one go)

MIGRATIONS
  migrate                      Run pending migrations
  migrate --rollback           Rollback last batch

UTILITIES
  key:generate                 Generate a secure APP_KEY (base64:...)
  storage:link                 Symlink public/storage → storage/app/public
  route:list                   Print all registered routes (HTTP + gRPC + WS)
  tinker                       Interactive REPL with container access

  help                         Show help
  version / -v                 Show version
```

---

## SOLID Compliance Audit

| Module | SRP | OCP | LSP | ISP | DIP |
|---|---|---|---|---|---|
| **Storage** | ✅ Manager / Driver / Provider each have one job | ✅ `RegisterFactory` - add drivers without changing Manager | ✅ MinIO ≡ S3 ≡ GDrive, fully interchangeable | ✅ 13-method interface, nothing irrelevant | ✅ Depend on `Contracts.StorageDriver` |
| **OAuth** | ✅ Generic provider + named wrappers separated | ✅ `RegisterDriver` - add OIDC/Keycloak without touching SocialiteManager | ✅ All return `SocialUser` | ✅ 4-method slim interface | ✅ `Contracts.AuthProvider` |
| **Transport** | ✅ REST / gRPC / WS adapters each handle one format | ✅ Add GraphQL = new adapter file | ✅ All call `Contracts.NotificationService` | ✅ No REST methods on WS adapter | ✅ All depend on service contracts |
| **RBAC** | ✅ Auth, roles, and permissions separated | ✅ Add permission checks without changing RBACService | ✅ Redis-backed ≡ in-memory, swappable | ✅ Minimal interface per concern | ✅ `Contracts.RBACService` |

---

## Production Checklist

1. **Config** - override via env vars (`APP_KEY`, `DB_*`, `REDIS_*`). Rotate `APP_KEY` from default.
2. **PostgreSQL** - set `SetMaxOpenConns(100)`, `SetMaxIdleConns(25)`. Enable `pg_stat_statements`. Create pgvector HNSW index before going live.
3. **Redis** - enable RDB/AOF persistence for RBAC data. Use Cluster Mode above 1M WebSocket connections.
4. **WebSocket** - put Nginx/HAProxy in front. `rancago:ws:*` pub/sub means sticky sessions are optional.
5. **Storage** - enforce HTTPS for S3/MinIO. Set presigned URL expiry < 15 minutes for private files.
6. **OAuth** - use HTTPS redirect URLs in production. Store `oauth_state` in Redis/signed cookie.
7. **Google APIs** - enable Domain-Wide Delegation on the Service Account for Calendar impersonation.
8. **Observability** - export Prometheus metrics from Redis `INFO stats`, `pg_stat_statements`, and WebSocket `connected_count`.

---

## Key File Reference

| Concept | File |
|---|---|
| IoC Container | [`framework/Container/container.go`](framework/Container/container.go) |
| ServiceProvider interface | [`app/Contracts/provider.go`](app/Contracts/provider.go) |
| StorageDriver interface + Manager | [`app/Contracts/storage.go`](app/Contracts/storage.go) · [`framework/Storage/manager.go`](framework/Storage/manager.go) |
| OAuth + SocialiteManager | [`app/Contracts/auth.go`](app/Contracts/auth.go) · [`framework/Auth/socialite.go`](framework/Auth/socialite.go) |
| Redis RBAC + Middleware | [`framework/Auth/rbac.go`](framework/Auth/rbac.go) |
| pgvector model + Vector type | [`app/Models/models.go`](app/Models/models.go) · [`app/Models/vector.go`](app/Models/vector.go) |
| Google Calendar + Meet | [`framework/Google/calendar.go`](framework/Google/calendar.go) · [`framework/Google/meet.go`](framework/Google/meet.go) |
| WebSocket Hub | [`framework/WebSocket/hub.go`](framework/WebSocket/hub.go) |
| 3-in-1 Transport adapters | [`framework/Transport/NotificationTransport.go`](framework/Transport/NotificationTransport.go) |
| Application bootstrap | [`bootstrap/app.go`](bootstrap/app.go) |
| Typed config | [`config/config.go`](config/config.go) |
| Domain entities | [`internal/domain/entities/`](internal/domain/entities/) |
| Ports (interfaces) | [`internal/ports/`](internal/ports/) |

---

## Contributing

Please read [CONTRIBUTING.md](https://github.com/rancago/rancago/blob/main/CONTRIBUTING.md) before opening an issue or pull request.

---

## License

MIT License - Copyright (c) 2026 [Muhammad Ikhwan Fathulloh](https://github.com/ikhwanfathulloh)

See [LICENSE](LICENSE) for full text.

---

> **RANCAGO**: Go + Laravel productivity - idiomatic, typed, and built to scale.
