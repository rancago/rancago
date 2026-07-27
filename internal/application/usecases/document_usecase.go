package usecases

import (
	"context"
	"fmt"

	"github.com/rancago/framework/internal/domain/entities"
	"github.com/rancago/framework/internal/domain/valueobjects"
	derrors "github.com/rancago/framework/internal/domain/errors"
	"github.com/rancago/framework/internal/ports/driven"
	"github.com/rancago/framework/internal/ports/driving"
)

type DocumentInteractor struct {
	docs driven.DocumentRepository
}

func NewDocumentInteractor(docs driven.DocumentRepository) driving.DocumentUseCase {
	return &DocumentInteractor{docs: docs}
}

func (uc *DocumentInteractor) Create(ctx context.Context, title, content string, userID *valueobjects.ID) (*entities.Document, error) {
	if title == "" {
		return nil, derrors.New("document.create", derrors.ErrValidation, "title is required")
	}
	d := entities.NewDocument(title, content)
	d.UserID = userID
	return uc.docs.Create(ctx, d)
}

func (uc *DocumentInteractor) GetByID(ctx context.Context, id valueobjects.ID) (*entities.Document, error) {
	d, err := uc.docs.FindByID(ctx, id)
	if err != nil || d == nil {
		return nil, derrors.New("document.get", derrors.ErrNotFound, fmt.Sprintf("document %s not found", id.String()))
	}
	return d, nil
}

func (uc *DocumentInteractor) Update(ctx context.Context, id valueobjects.ID, title, content string) (*entities.Document, error) {
	d, err := uc.docs.FindByID(ctx, id)
	if err != nil || d == nil {
		return nil, derrors.New("document.update", derrors.ErrNotFound, fmt.Sprintf("document %s not found", id.String()))
	}
	if title != "" {
		d.Title = title
	}
	if content != "" {
		d.UpdateContent(content)
	}
	return uc.docs.Update(ctx, d)
}

func (uc *DocumentInteractor) Delete(ctx context.Context, id valueobjects.ID) error {
	return uc.docs.Delete(ctx, id)
}

func (uc *DocumentInteractor) List(ctx context.Context, page, perPage int) ([]*entities.Document, valueobjects.PaginationMeta, error) {
	return uc.docs.FindPaginated(ctx, page, perPage)
}

func (uc *DocumentInteractor) VectorSearch(
	ctx context.Context,
	queryEmbedding []float32,
	limit int,
) ([]*entities.Document, []float64, error) {
	if len(queryEmbedding) == 0 {
		return nil, nil, derrors.New("document.vector_search", derrors.ErrValidation, "query embedding is empty")
	}
	results, err := uc.docs.CosineSimilarity(ctx, queryEmbedding, limit)
	if err != nil {
		return nil, nil, err
	}
	docs := make([]*entities.Document, 0, len(results))
	scores := make([]float64, 0, len(results))
	for _, r := range results {
		docs = append(docs, r.Item)
		scores = append(scores, r.Score)
	}
	return docs, scores, nil
}

func (uc *DocumentInteractor) TextSearch(ctx context.Context, query string, limit int) ([]*entities.Document, error) {
	return uc.docs.SearchByContent(ctx, query, limit)
}

func (uc *DocumentInteractor) SetEmbedding(ctx context.Context, id valueobjects.ID, embedding []float32) error {
	d, err := uc.docs.FindByID(ctx, id)
	if err != nil || d == nil {
		return derrors.New("document.embedding", derrors.ErrNotFound, fmt.Sprintf("document %s not found", id.String()))
	}
	d.SetEmbedding(embedding)
	_, err = uc.docs.Update(ctx, d)
	return err
}
