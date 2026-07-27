package Contracts

import (
	"context"
	"database/sql"
)

// Repository is the generic CRUD contract for any entity type T.
type Repository[T any] interface {
	FindByID(ctx context.Context, id interface{}) (*T, error)
	FindAll(ctx context.Context) ([]*T, error)
	FindPaginated(ctx context.Context, page, perPage int) ([]*T, PaginationMeta, error)
	Create(ctx context.Context, entity *T) (*T, error)
	Update(ctx context.Context, entity *T) (*T, error)
	Delete(ctx context.Context, id interface{}) error
	FindBy(ctx context.Context, conditions map[string]interface{}) ([]*T, error)
	FirstOrCreate(ctx context.Context, conditions, defaults map[string]interface{}) (*T, bool, error)
}

// PaginationMeta holds standard pagination metadata.
type PaginationMeta struct {
	Page       int
	PerPage    int
	Total      int64
	TotalPages int64
	HasNext    bool
	HasPrev    bool
}

// VectorSearchResult wraps a search hit with similarity scores.
type VectorSearchResult[T any] struct {
	Item      *T
	ID        interface{}
	Score     float64
	Distance  float64
	Embedding []float32
	Metadata  map[string]interface{}
}

// VectorRepository extends Repository with pgvector semantic search capabilities.
type VectorRepository[T any] interface {
	SimilaritySearch(ctx context.Context, queryEmbedding []float32, limit int, threshold *float64) ([]*VectorSearchResult[T], error)
	CosineSimilarity(ctx context.Context, queryEmbedding []float32, limit int) ([]*VectorSearchResult[T], error)
	L2Distance(ctx context.Context, queryEmbedding []float32, limit int) ([]*VectorSearchResult[T], error)
	InnerProduct(ctx context.Context, queryEmbedding []float32, limit int) ([]*VectorSearchResult[T], error)
	UpsertVector(ctx context.Context, id interface{}, embedding []float32, metadata map[string]interface{}) error
	DeleteVector(ctx context.Context, id interface{}) error
	// EnsureExtension creates the pgvector extension if it does not exist.
	EnsureExtension(ctx context.Context) error
}

// Transaction is the contract for database transaction management.
type Transaction interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	// Do wraps fn in a transaction, auto-rolling back on error or panic.
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

// DatabaseConnection is the low-level database connection contract.
type DatabaseConnection interface {
	DB() *sql.DB
	GetDialect() string
	Ping(ctx context.Context) error
	Migrate(ctx context.Context, migrations []string) error
	Close() error
}
