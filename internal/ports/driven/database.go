package driven

import (
	"context"
	"database/sql"

	"github.com/rancago/framework/internal/domain/valueobjects"
)

type Repository[T any] interface {
	FindByID(ctx context.Context, id valueobjects.ID) (*T, error)
	FindAll(ctx context.Context) ([]*T, error)
	FindPaginated(ctx context.Context, page, perPage int) ([]*T, valueobjects.PaginationMeta, error)
	Create(ctx context.Context, entity *T) (*T, error)
	Update(ctx context.Context, entity *T) (*T, error)
	Delete(ctx context.Context, id valueobjects.ID) error
	FindBy(ctx context.Context, conditions map[string]interface{}) ([]*T, error)
	FirstOrCreate(ctx context.Context, conditions map[string]interface{}, defaults map[string]interface{}) (*T, bool, error)
}

type VectorSearchResult[T any] struct {
	Item      *T
	ID        interface{}
	Score     float64
	Distance  float64
	Embedding []float32
	Metadata  map[string]interface{}
}

type VectorRepository[T any] interface {
	SimilaritySearch(ctx context.Context, queryEmbedding []float32, limit int, threshold *float64) ([]*VectorSearchResult[T], error)
	CosineSimilarity(ctx context.Context, queryEmbedding []float32, limit int) ([]*VectorSearchResult[T], error)
	L2Distance(ctx context.Context, queryEmbedding []float32, limit int) ([]*VectorSearchResult[T], error)
	InnerProduct(ctx context.Context, queryEmbedding []float32, limit int) ([]*VectorSearchResult[T], error)
	UpsertVector(ctx context.Context, id interface{}, embedding []float32, metadata map[string]interface{}) error
	DeleteVector(ctx context.Context, id interface{}) error
}

type TransactionPort interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type DatabasePort interface {
	DB() *sql.DB
	GetDialect() string
	Ping(ctx context.Context) error
	Migrate(ctx context.Context, migrations []string) error
	Close() error
}
