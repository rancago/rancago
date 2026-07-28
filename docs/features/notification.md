<!-- CONTEXT_START
module: github.com/rancago/framework
feature: Notification
generated: 2026-07-28
arch: hexagonal (ports-and-adapters)
CONTEXT_END -->

# Feature: Notification

> Real-time in-app and broadcast notification system with Redis-backed unread counter.

## 🤖 AI Instructions

<!-- INSTRUCTION
Read this file FIRST before any edit. Follow hexagonal rules:
1. Domain entities MUST NOT import ports or adapters.
2. Ports are Go interfaces only — no implementations.
3. Use cases depend on driven ports (injected via constructor).
4. Adapters depend on driving ports — never on use case structs directly.
5. Wire new bindings in internal/bootstrap/app.go (Container.Singleton).
6. Use derrors.New(op, sentinel, msg) for domain errors.
7. IDs use valueobjects.ID — call valueobjects.NewIDStr() or NewIDUint().
INSTRUCTION -->

## 📁 Files

| Layer | File | Role |
|-------|------|------|
| Domain Entity | `internal/domain/entities/Notification.go` | Notification with channel enum, read-state, data map |
| Driven Port | `internal/ports/driven/notification.go` | NotificationRepository — CRUD + FindByUserID + MarkRead + GetUnreadCount |
| Driving Port | `internal/ports/driving/notification.go` | NotificationUseCase — Send, Broadcast, ListUserNotifications, MarkRead, GetUnreadCount |
| Use Case | `internal/application/usecases/notification_usecase.go` | Orchestrates repo + cache (Redis unread counter) + WebSocket push |
| HTTP Adapter | `internal/adapters/driving/http/handlers.go` | NotificationHandler — /send /broadcast /list /count /read |
| gRPC Adapter | `internal/adapters/driving/grpc/notification_adapter.go` | GRPCNotificationAdapter stub |
| In-Memory Repo | `internal/adapters/driven/persistence/inmemory/notification_repo.go` | In-memory implementation for dev/test |

## 🏗️ Layer Flow

```
POST /api/v1/notifications/send
  └─ NotificationHandler.handleSend()          (adapters/driving/http)
       └─ NotificationUseCase.Send()           (ports/driving)
            └─ NotificationInteractor.Send()   (application/usecases)
                 ├─ NotificationRepository.Create()  (ports/driven → inmemory)
                 ├─ CachePort.Incr()                 (ports/driven → redis)
                 └─ WebSocketPort.SendDirect()       (ports/driven → hub)
```

## 🔌 Bootstrap Keys

| Key | Type |
|-----|------|
| `repo.notification` | `driven.NotificationRepository` |
| `uc.notification` | `driving.NotificationUseCase` |

Alias: `"driving.NotificationUseCase"` → `"uc.notification"`

## ⚡ Quick Tasks

<!-- OUTPUT_HINTS
When asked to add a method to this feature:
  1. Add method signature to internal/ports/driving/notification.go
  2. Implement in internal/application/usecases/notification_usecase.go
  3. Add HTTP route in NotificationHandler.RegisterRoutes()
  4. Update this file's Quick Tasks table

When asked to add a field to the entity:
  1. Edit internal/domain/entities/Notification.go
  2. Update NewNotification() constructor
  3. Create a migration: rancago make:migration add_field_to_notifications
OUTPUT_HINTS -->

| Task | Where |
|------|-------|
| Add channel type | `internal/domain/entities/Notification.go` — `NotificationChannel` const |
| Add use case method | Port: `internal/ports/driving/notification.go` → Interactor: `notification_usecase.go` |
| Add HTTP route | `internal/adapters/driving/http/handlers.go` `RegisterRoutes()` |
| Swap Redis → real impl | Create `internal/adapters/driven/cache/redis_adapter.go`, wire in bootstrap |
| Add pagination | Already supported via `FindByUserID(ctx, userID, limit, offset)` |

## 🚨 Domain Errors

```go
derrors.New("notification.send", derrors.ErrValidation, "title is required")
derrors.New("notification.mark_read", derrors.ErrNotFound, "notification not found")
```

## 📡 HTTP API

| Method | Path | Handler |
|--------|------|---------|
| POST | `/api/v1/notifications/send` | `handleSend` — body: `{user_id, title, body, channel, data}` |
| POST | `/api/v1/notifications/broadcast` | `handleBroadcast` — body: `{title, body, data}` |
| GET | `/api/v1/notifications/list?user_id=&page=&per_page=` | `handleList` |
| GET | `/api/v1/notifications/count?user_id=` | `handleCount` |
| POST | `/api/v1/notifications/read` | `handleMarkRead` — body: `{id, user_id}` |

## 🏷️ Channels

```go
entities.ChannelDatabase  // default — persisted in DB
entities.ChannelEmail     // email transport
entities.ChannelPush      // push notification
entities.ChannelSMS       // SMS
entities.ChannelBroadcast // WebSocket broadcast to all
```
