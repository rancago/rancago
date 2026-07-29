<!-- CONTEXT_START
module: github.com/rancago/framework
feature: Document
generated: 2026-07-28
arch: hexagonal (ports-and-adapters)
CONTEXT_END -->

# Feature: Document

> Document storage and vector-search enabled document repository.

## 🤖 AI Instructions

<!-- INSTRUCTION
Read this file FIRST before any edit. Follow hexagonal rules:
1. Domain entities MUST NOT import ports or adapters.
2. Ports are Go interfaces only - no implementations.
3. Use cases depend on driven ports (injected via constructor).
4. Adapters depend on driving ports - never on use case structs directly.
5. Wire new bindings in internal/bootstrap/app.go (Container.Singleton).
6. Use derrors.New(op, sentinel, msg) for domain errors.
7. IDs use valueobjects.ID - call valueobjects.NewIDStr() or NewIDUint().
8. Vector search uses driven.VectorRepository[entities.Document] - float32 embeddings.
INSTRUCTION -->

## 📁 Files

| Layer | File | Role |
|-------|------|------|
| Domain Entity | `internal/domain/entities/Document.go` | Document with content, embedding, metadata |
| Domain Model | `app/Models/vector.go` | GORM-style ORM model with pgvector column |
| Driven Port | `internal/ports/driven/document.go` | DocumentRepository extends Repository[Document] + vector search |
| Driving Port | `internal/ports/driving/document.go` | DocumentUseCase |
| Use Case | `internal/application/usecases/document_usecase.go` | DocumentInteractor |
| In-Memory Repo | `internal/adapters/driven/persistence/inmemory/document_repo.go` | In-memory with linear similarity stub |

## 🏗️ Layer Flow

```
HTTP/gRPC
  └─ DocumentUseCase (driving port)
       └─ DocumentInteractor
            ├─ DocumentRepository.Create()          - persist
            ├─ DocumentRepository.SimilaritySearch() - vector nearest-neighbor
            └─ StorageDriver.Put()                  - file upload (optional)
```

## 🔌 Bootstrap Keys

| Key | Type |
|-----|------|
| `repo.document` | `driven.DocumentRepository` |
| `uc.document` | `driving.DocumentUseCase` |

Alias: `"driving.DocumentUseCase"` → `"uc.document"`

## ⚡ Quick Tasks

<!-- OUTPUT_HINTS
When asked to add vector search endpoint:
  1. Add method to internal/ports/driving/document.go
  2. Implement in document_usecase.go calling repo.SimilaritySearch()
  3. Add HTTP route in a new DocumentHandler driving adapter

When asked to connect real pgvector:
  1. Create internal/adapters/driven/persistence/postgres/document_repo.go
  2. Implement driven.DocumentRepository using pgx + pgvector
  3. Swap binding in bootstrap RegisterCore()
OUTPUT_HINTS -->

| Task | Where |
|------|-------|
| Add document field | `internal/domain/entities/Document.go` + migration |
| Add vector search | Use case method → `DocumentRepository.SimilaritySearch()` |
| Swap to real DB | New adapter in `internal/adapters/driven/persistence/postgres/` |
| Add HTTP handler | `rancago make:adapter DocumentHandler --direction driving` |

## 🚨 Domain Errors

```go
derrors.New("document.create", derrors.ErrValidation, "content is required")
derrors.New("document.find",   derrors.ErrNotFound,   "document not found")
```

## 🔍 Vector Search

```go
// Driven port method signature:
SimilaritySearch(ctx, queryEmbedding []float32, limit int, threshold *float64) ([]*VectorSearchResult[Document], error)

// VectorSearchResult fields: Item *Document, Score float64, Distance float64
```
