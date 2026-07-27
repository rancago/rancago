# RANCAGO Framework

> **Resilient, Agnostic, & Native Clean-Architecture GO Framework**

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![Module](https://img.shields.io/badge/module-github.com%2Francago%2Fframework-blue?style=flat-square)](https://github.com/rancago/rancago)
[![License](https://img.shields.io/badge/license-Proprietary-red?style=flat-square)](LICENSE)

SOLID Principles · IoC Service Container · Multi-API Transport · pgvector Semantic Search · Redis WebSocket Hub · Google Ecosystem · OAuth Socialite

---

## Arti Nama

**RANCAGO** membawa dua lapisan makna sekaligus.

**Akronim resmi:**

> **R**esilient, **A**gnostic, & **N**ative **C**lean-**A**rchitecture **G**O Framework

Dirancang untuk tangguh di bawah beban (*resilient*), bebas dari ketergantungan transport tertentu (*agnostic* — REST + gRPC + WebSocket dari satu definisi service), dan sepenuhnya idiomatik terhadap pola clean architecture di Go (*native*).

**Akar Sunda / lokal:**

| Kata | Makna |
|---|---|
| **Rancagé** | *Kecakapan, keterampilan, dan kerapian dalam merancang sesuatu secara bertahap dan terstruktur* — mencerminkan cara Rancago menegakkan prinsip SOLID dan arsitektur hexagonal langkah demi langkah. |
| **Ranca** | *Daerah perairan/lahan luas yang subur* — melambangkan ekosistem framework yang kaya fitur bawaan: pgvector, Redis, MinIO, Google Drive, Meet, Calendar, OAuth, RBAC, semuanya tumbuh dari satu fondasi. |

Namanya sekaligus akronim teknis dan penghormatan terhadap budaya lokal — framework yang dibangun dengan kecakapan (*rancagé*) dan dirancang tumbuh seperti ekosistem subur (*ranca*).

---

## Kenapa Rancago?

| Dimensi | Vanilla Go | Laravel PHP | **Rancago** |
|---|---|---|---|
| **Kecepatan** | ⚡⚡⚡ Native | 🐢 Moderate | ⚡⚡⚡ Native Go |
| **DX / Produktivitas** | Wiring manual, setup berhari-hari | Artisan, Facades siap pakai | CLI generator + Contracts-First |
| **SOLID / Clean Arch** | Bebas, sering over-coupled | Facades melanggar DIP | **Contracts-First di-enforce** |
| **Realtime & Scale** | Redis PubSub manual | Pusher (berbayar) | Built-in multi-node WebSocket Hub |
| **Multi-API Transport** | REST saja | REST + queued jobs | **1 Service = REST + gRPC + WebSocket** |
| **AI-Ready** | Setup pgvector manual | Butuh package eksternal | Semantic search sudah built-in |

---

## Fitur Utama

| Kategori | Yang disertakan |
|---|---|
| 🏛 **Arsitektur** | SOLID Principles · IoC Service Container · Lifecycle ServiceProvider · Hexagonal Ports & Adapters |
| 💾 **Database** | PostgreSQL + GORM · **pgvector** semantic search (Cosine / L2 / Inner Product) · Generic Repository · Pagination · Transaction Manager |
| 📦 **Storage** | **MinIO** · **Amazon S3** · **Google Drive** · Lazy disk init · Functional Options · Temporary presigned URL · Extension point `RegisterFactory` (OCP) |
| 📅 **Google Ecosystem** | Calendar API (CRUD + attendee + reminder) · Google Meet auto-link · Facade **MeetingScheduler** (1 panggilan = event + Meet link) |
| 🔐 **Auth** | OAuth Socialite: **Google / GitHub / Facebook / Custom OIDC** · RBAC berbasis Redis (role + permission + policy middleware) · Auth middleware Bearer token |
| ⚡ **Cache & Realtime** | Redis Manager (Get/Set/SAdd/PubSub) · Rate limiter · **Scalable WebSocket Hub** (multi-node via Redis Pub/Sub - broadcast / channel / direct) |
| 🎯 **Transport** | **1 service = 3 endpoint**: REST HTTP + gRPC + WebSocket - nol duplikasi kode |
| 🛠 **CLI** | `serve · migrate · scaffold · make:entity/value-object/port/usecase/adapter/model/migration · key:generate · storage:link · route:list · tinker` |

---

## Struktur Proyek

```
rancago/
├── app/
│   ├── Contracts/          # ⭐ Semua definisi interface (DIP compliance)
│   ├── Models/             # Model GORM + tipe Vector untuk pgvector
│   ├── Providers/          # Implementasi ServiceProvider (Register + Boot)
│   ├── Services/           # Logika bisnis transport-agnostic
│   ├── Repositories/       # Layer akses data
│   └── Http/               # Controller, Middleware, Request
├── framework/              # ⭐ Core framework yang bisa diekstrak
│   ├── Container/          # IoC Service Container
│   ├── Auth/Providers/     # OAuth Socialite + RBAC + Policy Enforcer
│   ├── Cache/              # Redis Manager + PubSub + Rate Limiter
│   ├── Database/           # Generic Repository + pgvector + Transaction
│   ├── Google/             # Calendar + Meet + MeetingScheduler
│   ├── Storage/Drivers/    # MinIO · S3 · Google Drive · Memory
│   ├── Transport/          # Adapter REST / gRPC / WebSocket
│   └── WebSocket/          # Scalable Hub (Redis Pub/Sub)
├── internal/               # Port & adapter hexagonal (domain layer)
│   ├── domain/             # Entity, value object, domain error
│   ├── ports/              # Interface port driven + driving
│   ├── application/        # Use case interactor
│   ├── adapters/           # Adapter HTTP, gRPC, CLI, cache, persistence
│   ├── bootstrap/          # Wiring internal
│   └── kernel/             # IoC container + config (internal)
├── config/                 # Konfigurasi typed (tanpa map[string]interface{})
├── routes/                 # Declarative route builder
├── bootstrap/              # Kernel aplikasi (wiring ServiceProvider)
├── database/migrations/    # File migrasi database
├── cmd/rancago/            # Entry point CLI
└── main.go                 # Entry point server
```

---

## Quick Start

### Prasyarat

- Go 1.23+
- PostgreSQL 14+ dengan extension `vector` (`CREATE EXTENSION IF NOT EXISTS vector;`)
- Redis 6+
- MinIO _(opsional, untuk object storage)_
- Google Service Account JSON _(opsional, untuk Calendar / Meet / Drive)_

### 1. Clone & install

```bash
git clone https://github.com/rancago/rancago.git
cd rancago
go mod tidy
```

### 2. Konfigurasi

Edit `config/config.go` dan sesuaikan nilai default, atau override via environment variable:

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
            ClientID:     "your-client-id",
            ClientSecret: "your-secret",
            RedirectURL:  "http://localhost:8080/auth/google/callback",
        },
    },
},
```

### 3. Jalankan

```bash
# HTTP + gRPC + WebSocket dalam satu proses
go run .

# Atau via CLI
go run ./cmd/rancago serve --port 8080
go run ./cmd/rancago serve --grpc --port 8080
```

### 4. Verifikasi

```bash
curl http://localhost:8080/api/v1/health
# {"status":"healthy","service":"rancago-api"}

curl http://localhost:8080/
# {"message":"Welcome to Rancago Framework 🚀","version":"1.0.0"}
```

---

## Konsep Inti

### Aturan Emas: Contracts → Providers → Services

```
app/Contracts/*.go   - definisikan semua interface terlebih dahulu (DIP, ISP compliant)
     ↓
app/Providers/*.go   - ikat implementasi konkret ke dalam container
     ↓
app/Services/*.go    - logika bisnis transport-agnostic, hanya bergantung pada Contracts
```

### IoC Container

```go
// Ikat tipe
c.Singleton("service.payment", func(c *Container.Container) (interface{}, error) {
    storage, _ := c.Resolve("Contracts.StorageDriver")
    return Services.NewPaymentService(storage.(Contracts.StorageDriver)), nil
})
c.Alias("service.payment", "Contracts.PaymentService")

// Resolve
svc := container.MustResolve("Contracts.PaymentService").(Contracts.PaymentService)
```

### Pola ServiceProvider

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
    // Daftarkan driver, seed data, jalankan migrasi
    return nil
}
```

Daftarkan di `bootstrap/app.go`:

```go
app.RegisterProviders(
    Providers.NewStorageServiceProvider(...),
    Providers.NewPaymentServiceProvider(),
)
```

---

## Panduan Modul

### Storage - MinIO · S3 · Google Drive

Bergantung pada interface, bukan driver konkret:

```go
type FileService struct {
    Storage Contracts.StorageDriver // disuntikkan via container
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

Tambah driver baru tanpa menyentuh Manager (100% OCP):

```go
mgr.RegisterFactory("azure_blob", func(cfg Contracts.StorageDiskConfig) (Contracts.StorageDriver, error) {
    return NewAzureBlobDriver(cfg), nil
})
disk, _ := mgr.Disk("azure_blob")
```

### Google Calendar + Meet

Satu method untuk menjadwalkan event sekaligus membuat link Meet:

```go
result, err := scheduler.ScheduleWithMeet(ctx, &Contracts.CalendarEvent{
    Summary:  "Sprint Review",
    Start:    time.Now().Add(24 * time.Hour),
    End:      time.Now().Add(25 * time.Hour),
    Timezone: "Asia/Jakarta",
    Attendees: []Contracts.Attendee{
        {Email: "team@example.com", DisplayName: "Tim Dev"},
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
// Redirect ke provider
authURL, state, _ := socialite.Redirect(ctx, "google")
http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)

// Handle callback - mengembalikan SocialUser yang seragam untuk semua provider
user, _ := socialite.Callback(ctx, "google", code, state)
fmt.Println(user.Email, user.Name, user.AvatarURL)
```

Tambah provider custom tanpa mengubah SocialiteManager:

```go
mgr.RegisterDriver("keycloak", func() (Contracts.AuthProvider, error) {
    return Auth.NewGenericOAuthProvider("keycloak", Auth.OAuthConfig{
        AuthURL:     "https://sso.perusahaan.com/auth",
        TokenURL:    "https://sso.perusahaan.com/token",
        UserInfoURL: "https://sso.perusahaan.com/userinfo",
        Scopes:      []string{"openid", "email", "profile"},
    }), nil
})
```

### RBAC Redis-backed

```go
rbac.AssignRole(ctx, "user-123", "admin")
rbac.GivePermissionToRole(ctx, "admin", "delete:user")

// Middleware HTTP - memblokir request yang tidak punya permission yang dibutuhkan
mux.Handle("/admin/users",
    rbac.Middleware("delete:user")(http.HandlerFunc(handler)))

// Middleware berbasis role
mux.Handle("/admin/dashboard",
    rbac.RoleMiddleware("admin", "superadmin")(http.HandlerFunc(handler)))
```

### Notifikasi - 1 service, 3 transport

Tulis logika bisnis sekali saja di `app/Services/NotificationService.go`. Kode yang sama diekspos melalui REST, gRPC, dan WebSocket secara otomatis.

```bash
# REST
curl -X POST http://localhost:8080/api/v1/notifications/send \
  -H "Content-Type: application/json" \
  -d '{"user_id":"123","title":"Pesanan dikirim","body":"Pesanan #8872 sedang dalam perjalanan","channel":"push"}'

# WebSocket (wscat)
wscat -c "ws://localhost:8080/ws?user_id=123"
> {"action":"notification:send","payload":{"user_id":"456","title":"Halo","body":"Test"}}
< {"type":"notification:new","channel":"user:456","payload":{...}}
```

### WebSocket Multi-node (Redis Pub/Sub)

Setiap panggilan `PublishChannel` otomatis mempublikasikan ke `rancago:ws:{channel}` di Redis. Semua instance yang berjalan berlangganan dan meneruskan pesan ke client lokal mereka - tanpa sticky session.

```bash
# Node 1
go run . # HTTP :8080

# Node 2 (terminal lain, ubah HTTPPort ke 8081)
go run . # HTTP :8081
```

User A di `:8080` mengirim ke `room:123` → User B di `:8081` menerimanya secara otomatis.

### pgvector Semantic Search

```go
// Cosine similarity search - mengembalikan dokumen yang paling mirip secara semantik dengan query
threshold := float64(0.75)
results, _ := docRepo.SimilaritySearch(ctx, queryEmbedding, 10, &threshold)
for _, hit := range results {
    fmt.Printf("[%.1f%%] %s\n", hit.Score*100, hit.Item.Title)
}
```

---

## Referensi CLI

```
rancago [command] [flags]

SERVER
  serve                        Jalankan HTTP (+ opsional gRPC) server
    --port, -p  int            Port HTTP (default: 8080)
    --grpc                     Jalankan juga gRPC stub server

GENERATOR
  make:entity     [name]       Domain entity
  make:value-object [name]     Value object
  make:port       [name]       Interface port driving/driven (flag --driving)
  make:usecase    [name]       Use case interactor
  make:adapter    [name]       Adapter infrastruktur (--direction driven|driving)
  make:model      [name] [-m]  Model GORM (+ migrasi dengan -m)
  make:migration  [name]       Stub file migrasi

SCAFFOLD
  scaffold [name]              Scaffolder interaktif bounded context
                               (entity + port + usecase + adapter sekaligus)

MIGRASI
  migrate                      Jalankan migrasi pending
  migrate --rollback           Rollback batch terakhir

UTILITAS
  key:generate                 Generate APP_KEY aman (base64:...)
  storage:link                 Symlink public/storage → storage/app/public
  route:list                   Tampilkan semua route (HTTP + gRPC + WS)
  tinker                       REPL interaktif dengan akses container

  help                         Tampilkan bantuan
  version / -v                 Tampilkan versi
```

---

## Audit Kepatuhan SOLID

| Modul | SRP | OCP | LSP | ISP | DIP |
|---|---|---|---|---|---|
| **Storage** | ✅ Manager / Driver / Provider masing-masing satu tugas | ✅ `RegisterFactory` - tambah driver tanpa ubah Manager | ✅ MinIO ≡ S3 ≡ GDrive, sepenuhnya bisa saling tukar | ✅ Interface 13 method, tidak ada yang tidak relevan | ✅ Bergantung pada `Contracts.StorageDriver` |
| **OAuth** | ✅ Generic provider + wrapper named terpisah | ✅ `RegisterDriver` - tambah OIDC/Keycloak tanpa ubah SocialiteManager | ✅ Semua mengembalikan `SocialUser` | ✅ Interface 4 method yang ramping | ✅ `Contracts.AuthProvider` |
| **Transport** | ✅ Adapter REST / gRPC / WS masing-masing satu format | ✅ Tambah GraphQL = file adapter baru | ✅ Semua memanggil `Contracts.NotificationService` | ✅ Tidak ada method REST di adapter WS | ✅ Semua bergantung pada contract service |
| **RBAC** | ✅ Auth, role, dan permission terpisah | ✅ Tambah pengecekan permission tanpa ubah RBACService | ✅ Berbasis Redis ≡ in-memory, bisa saling tukar | ✅ Interface minimal per concern | ✅ `Contracts.RBACService` |

---

## Checklist Produksi

1. **Config** - override via env var (`APP_KEY`, `DB_*`, `REDIS_*`). Ganti `APP_KEY` dari nilai default.
2. **PostgreSQL** - set `SetMaxOpenConns(100)`, `SetMaxIdleConns(25)`. Aktifkan `pg_stat_statements`. Buat HNSW index pgvector sebelum live.
3. **Redis** - aktifkan persistensi RDB/AOF untuk data RBAC. Gunakan Cluster Mode di atas 1 juta koneksi WebSocket.
4. **WebSocket** - pasang Nginx/HAProxy di depan. Pub/Sub `rancago:ws:*` membuat sticky session menjadi opsional.
5. **Storage** - gunakan HTTPS untuk S3/MinIO. Set expiry presigned URL < 15 menit untuk file privat.
6. **OAuth** - gunakan redirect URL HTTPS di produksi. Simpan `oauth_state` di Redis atau signed cookie.
7. **Google API** - aktifkan Domain-Wide Delegation pada Service Account untuk impersonasi Calendar.
8. **Observabilitas** - ekspor metrik Prometheus dari Redis `INFO stats`, `pg_stat_statements`, dan `connected_count` WebSocket.

---

## Referensi File Penting

| Konsep | File |
|---|---|
| IoC Container | [`framework/Container/container.go`](framework/Container/container.go) |
| Interface ServiceProvider | [`app/Contracts/provider.go`](app/Contracts/provider.go) |
| Interface StorageDriver + Manager | [`app/Contracts/storage.go`](app/Contracts/storage.go) · [`framework/Storage/manager.go`](framework/Storage/manager.go) |
| OAuth + SocialiteManager | [`app/Contracts/auth.go`](app/Contracts/auth.go) · [`framework/Auth/socialite.go`](framework/Auth/socialite.go) |
| Redis RBAC + Middleware | [`framework/Auth/rbac.go`](framework/Auth/rbac.go) |
| Model GORM + tipe Vector | [`app/Models/models.go`](app/Models/models.go) · [`app/Models/vector.go`](app/Models/vector.go) |
| Google Calendar + Meet | [`framework/Google/calendar.go`](framework/Google/calendar.go) · [`framework/Google/meet.go`](framework/Google/meet.go) |
| WebSocket Hub | [`framework/WebSocket/hub.go`](framework/WebSocket/hub.go) |
| Adapter transport 3-in-1 | [`framework/Transport/NotificationTransport.go`](framework/Transport/NotificationTransport.go) |
| Bootstrap aplikasi | [`bootstrap/app.go`](bootstrap/app.go) |
| Konfigurasi typed | [`config/config.go`](config/config.go) |
| Domain entities | [`internal/domain/entities/`](internal/domain/entities/) |
| Port (interface) | [`internal/ports/`](internal/ports/) |

---

## Lisensi

Proprietary - Muhammad Ikhwan Fathulloh © 2026. Rancago Framework 1.0.0.

---

> **Rancago**: Produktivitas Laravel, idiom Go - typed, solid, dan siap scale.
