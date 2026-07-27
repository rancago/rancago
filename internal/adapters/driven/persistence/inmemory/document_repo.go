package inmemory

import (
	"context"
	"math"

	"github.com/rancago/framework/internal/domain/entities"
	"github.com/rancago/framework/internal/domain/valueobjects"
	"github.com/rancago/framework/internal/ports/driven"
)

type InMemoryDocumentRepo struct {
	InMemoryUserRepo
	docs map[string]*entities.Document
}

func NewInMemoryDocumentRepo() driven.DocumentRepository {
	return &InMemoryDocumentRepo{
		docs: make(map[string]*entities.Document),
	}
}

func (r *InMemoryDocumentRepo) cloneDoc(d *entities.Document) *entities.Document {
	cp := *d
	if d.Metadata != nil {
		cp.Metadata = make(map[string]string, len(d.Metadata))
		for k, v := range d.Metadata {
			cp.Metadata[k] = v
		}
	}
	if d.Embedding != nil {
		cp.Embedding = append([]float32(nil), d.Embedding...)
	}
	return &cp
}

func (r *InMemoryDocumentRepo) FindByID(_ context.Context, id valueobjects.ID) (*entities.Document, error) {
	d, ok := r.docs[id.String()]
	if !ok {
		return nil, nil
	}
	return r.cloneDoc(d), nil
}

func (r *InMemoryDocumentRepo) FindAll(_ context.Context) ([]*entities.Document, error) {
	out := make([]*entities.Document, 0, len(r.docs))
	for _, d := range r.docs {
		out = append(out, r.cloneDoc(d))
	}
	return out, nil
}

func (r *InMemoryDocumentRepo) FindPaginated(
	_ context.Context,
	page, perPage int,
) ([]*entities.Document, valueobjects.PaginationMeta, error) {
	all, _ := r.FindAll(context.Background())
	return paginate(all, page, perPage)
}

func (r *InMemoryDocumentRepo) Create(_ context.Context, d *entities.Document) (*entities.Document, error) {
	if d.ID.IsZero() {
		d.ID = valueobjects.NewIDStr("doc_" + d.CreatedAt.Format("20060102150405"))
	}
	r.docs[d.ID.String()] = r.cloneDoc(d)
	return r.cloneDoc(d), nil
}

func (r *InMemoryDocumentRepo) Update(_ context.Context, d *entities.Document) (*entities.Document, error) {
	r.docs[d.ID.String()] = r.cloneDoc(d)
	return r.cloneDoc(d), nil
}

func (r *InMemoryDocumentRepo) Delete(_ context.Context, id valueobjects.ID) error {
	delete(r.docs, id.String())
	return nil
}

func (r *InMemoryDocumentRepo) FindBy(_ context.Context, _ map[string]interface{}) ([]*entities.Document, error) {
	return r.FindAll(context.Background())
}

func (r *InMemoryDocumentRepo) FirstOrCreate(
	ctx context.Context,
	conds map[string]interface{},
	defaults map[string]interface{},
) (*entities.Document, bool, error) {
	title := "Untitled"
	if v, ok := defaults["title"]; ok {
		title = v.(string)
	}
	content := ""
	if v, ok := defaults["content"]; ok {
		content = v.(string)
	}
	d := entities.NewDocument(title, content)
	created, err := r.Create(ctx, d)
	return created, true, err
}

func cosineSim(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func (r *InMemoryDocumentRepo) SimilaritySearch(
	_ context.Context,
	queryEmbedding []float32,
	limit int,
	threshold *float64,
) ([]*driven.VectorSearchResult[entities.Document], error) {
	type scored struct {
		d     *entities.Document
		score float64
	}
	var list []scored
	for _, d := range r.docs {
		if len(d.Embedding) == 0 {
			continue
		}
		s := cosineSim(queryEmbedding, d.Embedding)
		if threshold != nil && s < *threshold {
			continue
		}
		list = append(list, scored{d: r.cloneDoc(d), score: s})
	}
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].score > list[i].score {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	out := make([]*driven.VectorSearchResult[entities.Document], 0, len(list))
	for _, s := range list {
		out = append(out, &driven.VectorSearchResult[entities.Document]{
			Item:     s.d,
			ID:       s.d.ID.String(),
			Score:    s.score,
			Distance: 1 - s.score,
		})
	}
	return out, nil
}

func (r *InMemoryDocumentRepo) CosineSimilarity(
	ctx context.Context,
	queryEmbedding []float32,
	limit int,
) ([]*driven.VectorSearchResult[entities.Document], error) {
	return r.SimilaritySearch(ctx, queryEmbedding, limit, nil)
}

func (r *InMemoryDocumentRepo) L2Distance(
	_ context.Context,
	queryEmbedding []float32,
	limit int,
) ([]*driven.VectorSearchResult[entities.Document], error) {
	type scored struct {
		d   *entities.Document
		dist float64
	}
	var list []scored
	for _, d := range r.docs {
		if len(d.Embedding) == 0 {
			continue
		}
		var dist float64
		for i := range queryEmbedding {
			if i >= len(d.Embedding) {
				break
			}
			diff := float64(queryEmbedding[i]) - float64(d.Embedding[i])
			dist += diff * diff
		}
		dist = math.Sqrt(dist)
		list = append(list, scored{d: r.cloneDoc(d), dist: dist})
	}
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].dist < list[i].dist {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	out := make([]*driven.VectorSearchResult[entities.Document], 0, len(list))
	for _, s := range list {
		out = append(out, &driven.VectorSearchResult[entities.Document]{
			Item:     s.d,
			ID:       s.d.ID.String(),
			Distance: s.dist,
		})
	}
	return out, nil
}

func (r *InMemoryDocumentRepo) InnerProduct(
	_ context.Context,
	queryEmbedding []float32,
	limit int,
) ([]*driven.VectorSearchResult[entities.Document], error) {
	type scored struct {
		d     *entities.Document
		score float64
	}
	var list []scored
	for _, d := range r.docs {
		if len(d.Embedding) == 0 {
			continue
		}
		var score float64
		for i := range queryEmbedding {
			if i >= len(d.Embedding) {
				break
			}
			score += float64(queryEmbedding[i]) * float64(d.Embedding[i])
		}
		list = append(list, scored{d: r.cloneDoc(d), score: score})
	}
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].score > list[i].score {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	out := make([]*driven.VectorSearchResult[entities.Document], 0, len(list))
	for _, s := range list {
		out = append(out, &driven.VectorSearchResult[entities.Document]{
			Item:  s.d,
			ID:    s.d.ID.String(),
			Score: s.score,
		})
	}
	return out, nil
}

func (r *InMemoryDocumentRepo) UpsertVector(_ context.Context, id interface{}, embedding []float32, _ map[string]interface{}) error {
	key := id.(string)
	if d, ok := r.docs[key]; ok {
		d.Embedding = embedding
	}
	return nil
}

func (r *InMemoryDocumentRepo) DeleteVector(_ context.Context, id interface{}) error {
	key := id.(string)
	if d, ok := r.docs[key]; ok {
		d.Embedding = nil
	}
	return nil
}

func (r *InMemoryDocumentRepo) SearchByContent(_ context.Context, query string, limit int) ([]*entities.Document, error) {
	var out []*entities.Document
	for _, d := range r.docs {
		if contains(d.Title, query) || contains(d.Content, query) {
			out = append(out, r.cloneDoc(d))
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func contains(s, substr string) bool {
	if substr == "" {
		return true
	}
	n := len(s)
	m := len(substr)
	if m > n {
		return false
	}
	for i := 0; i <= n-m; i++ {
		if s[i:i+m] == substr {
			return true
		}
	}
	return false
}
