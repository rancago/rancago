package valueobjects

type PaginationMeta struct {
	Page       int
	PerPage    int
	Total      int64
	TotalPages int64
	HasNext    bool
	HasPrev    bool
}

func NewPaginationMeta(page, perPage int, total int64) PaginationMeta {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 25
	}
	totalPages := int64(0)
	if perPage > 0 {
		totalPages = (total + int64(perPage) - 1) / int64(perPage)
	}
	offset := (page - 1) * perPage
	return PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    int64(offset+perPage) < total,
		HasPrev:    offset > 0,
	}
}

func (p PaginationMeta) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func (p PaginationMeta) Limit() int {
	return p.PerPage
}
