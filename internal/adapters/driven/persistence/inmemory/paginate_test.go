package inmemory

import "testing"

type paginateItem struct {
	V int
}

func TestPaginate_Defaults(t *testing.T) {
	items := []*paginateItem{{V: 1}, {V: 2}, {V: 3}}

	got, meta, err := paginate(items, 0, 0)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if meta.Page != 1 {
		t.Fatalf("expected Page=1, got %d", meta.Page)
	}
	if meta.PerPage != 25 {
		t.Fatalf("expected PerPage=25, got %d", meta.PerPage)
	}
	if meta.Total != int64(len(items)) {
		t.Fatalf("expected Total=%d, got %d", len(items), meta.Total)
	}
	if meta.TotalPages != 1 {
		t.Fatalf("expected TotalPages=1, got %d", meta.TotalPages)
	}
	if len(got) != len(items) {
		t.Fatalf("expected len=%d, got %d", len(items), len(got))
	}
}

func TestPaginate_PageWindow(t *testing.T) {
	items := []*paginateItem{
		{V: 1}, {V: 2}, {V: 3}, {V: 4}, {V: 5},
		{V: 6}, {V: 7}, {V: 8}, {V: 9}, {V: 10},
	}

	got, meta, err := paginate(items, 2, 3)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if meta.Page != 2 || meta.PerPage != 3 || meta.Total != 10 {
		t.Fatalf("unexpected meta: %+v", meta)
	}
	if meta.TotalPages != 4 {
		t.Fatalf("expected TotalPages=4, got %d", meta.TotalPages)
	}
	if !meta.HasNext || !meta.HasPrev {
		t.Fatalf("expected HasNext=true and HasPrev=true, got next=%v prev=%v", meta.HasNext, meta.HasPrev)
	}
	if len(got) != 3 {
		t.Fatalf("expected len=3, got %d", len(got))
	}
	if got[0].V != 4 || got[1].V != 5 || got[2].V != 6 {
		t.Fatalf("unexpected window values: %v %v %v", got[0].V, got[1].V, got[2].V)
	}
}

func TestPaginate_OffsetBeyondItems(t *testing.T) {
	items := []*paginateItem{{V: 1}, {V: 2}, {V: 3}, {V: 4}}

	got, meta, err := paginate(items, 5, 3)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if meta.Total != int64(len(items)) {
		t.Fatalf("expected Total=%d, got %d", len(items), meta.Total)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result, got len=%d", len(got))
	}
}

