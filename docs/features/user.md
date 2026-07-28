<!-- CONTEXT_START
module: github.com/rancago/framework
feature: User
generated: 2026-07-28
arch: hexagonal (ports-and-adapters)
CONTEXT_END -->

# Feature: User

> User authentication, OAuth/Socialite login, and RBAC (role/permission) management.

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
8. Email must be created via valueobjects.NewEmail() — validates format.
INSTRUCTION -->

## 📁 Files

| Layer | File | Role |
|-------|------|------|
| Domain Entity | `internal/domain/entities/User.go` | User with email VO, roles, permissions, OAuth fields |
| Domain Entity | `internal/domain/entities/Role.go` | Role with permission list |
| Domain Entity | `internal/domain/entities/Permission.go` | Permission — name-based guard |
| Value Object | `internal/domain/valueobjects/email.go` | Email VO with validation |
| Driven Port | `internal/ports/driven/auth.go` | SocialitePort, AuthProviderPort, OAuthToken, OAuthUser |
| Driven Port | `internal/ports/driven/user.go` | UserRepository, RoleRepository, PermissionRepository |
| Driving Port | `internal/ports/driving/user.go` | UserUseCase — Register, Login, FindByID, GetAuthURL, LoginWithProvider, AssignRole, HasPermission |
| Use Case | `internal/application/usecases/user_usecase.go` | UserInteractor — auth + RBAC orchestration |
| In-Memory Repo | `internal/adapters/driven/persistence/inmemory/user_repo.go` | In-memory with role/permission attachment |
| Auth Adapter | `internal/adapters/driven/auth/socialite.go` | SocialiteManager — multi-provider OAuth |

## 🏗️ Layer Flow

```
POST /auth/register
  └─ (HTTP Handler — not yet wired, add to BuildHTTPServer)
       └─ UserUseCase.Register(ctx, name, email, password)
            └─ UserInteractor.Register()
                 ├─ valueobjects.NewEmail(email)          — validates
                 ├─ UserRepository.FindByEmail()          — check duplicate
                 └─ UserRepository.Create()               — persist

GET /auth/google/callback
  └─ UserUseCase.LoginWithProvider(ctx, "google", code)
       └─ SocialitePort.Provider("google")
            ├─ AuthProviderPort.ExchangeCode(code)
            ├─ AuthProviderPort.GetUserInfo(token)
            └─ UserRepository.FindByProvider() or Create()
```

## 🔌 Bootstrap Keys

| Key | Type |
|-----|------|
| `repo.user` | `driven.UserRepository` |
| `repo.role` | `driven.RoleRepository` |
| `repo.permission` | `driven.PermissionRepository` |
| `socialite` | `driven.SocialitePort` |
| `uc.user` | `driving.UserUseCase` |

Alias: `"driving.UserUseCase"` → `"uc.user"`

## ⚡ Quick Tasks

<!-- OUTPUT_HINTS
When asked to add OAuth provider:
  1. Add provider config to internal/kernel/config.go AuthConfig.Providers map
  2. Register in app/Providers/AuthServiceProvider.go or bootstrap RegisterCore()
  3. No code changes needed in use case or domain

When asked to add RBAC permission check to HTTP route:
  1. Get user ID from context (X-User-ID header via AuthMiddleware)
  2. Call UserUseCase.HasPermission(ctx, userID, "perm.name")
  3. Return 403 if false

When asked to add a field to User entity:
  1. Edit internal/domain/entities/User.go
  2. Update NewUser() if it's a required field
  3. rancago make:migration add_field_to_users
OUTPUT_HINTS -->

| Task | Where |
|------|-------|
| Add OAuth provider | `internal/kernel/config.go` AuthConfig + bootstrap wiring |
| Add role to user | `UserUseCase.AssignRole(ctx, userID, roleName)` |
| Check permission | `UserUseCase.HasPermission(ctx, userID, permName)` |
| Add User field | `internal/domain/entities/User.go` + migration |
| HTTP routes | Add `UserHandler` in `internal/adapters/driving/http/` — not yet generated |

## 🚨 Domain Errors

```go
derrors.New("user.register", derrors.ErrAlreadyExists, "email already registered")
derrors.New("user.login",    derrors.ErrUnauthorized,  "invalid credentials")
derrors.New("user.socialite", derrors.ErrNotFound,     "provider not found")
```

## 🔐 RBAC

```go
// Entity helpers (no repo call needed if user already loaded)
user.HasRole("admin")           // bool
user.HasPermission("post.edit") // checks user perms + role perms

// Use case (loads fresh from repo)
ok, err := uc.user.HasPermission(ctx, userID, "post.edit")
```
