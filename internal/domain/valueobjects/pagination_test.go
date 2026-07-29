package valueobjects

import "testing"

func TestNewPaginationMeta_Defaults(t *testing.T) {
	p := NewPaginationMeta(0, 0, 0)
	if p.Page != 1 {
		t.Fatalf("expected Page=1, got %d", p.Page)
	}
	if p.PerPage != 25 {
		t.Fatalf("expected PerPage=25, got %d", p.PerPage)
	}
	if p.Total != 0 {
		t.Fatalf("expected Total=0, got %d", p.Total)
	}
	if p.TotalPages != 0 {
		t.Fatalf("expected TotalPages=0, got %d", p.TotalPages)
	}
	if p.HasNext {
		t.Fatalf("expected HasNext=false")
	}
	if p.HasPrev {
		t.Fatalf("expected HasPrev=false")
	}
	if p.Offset() != 0 {
		t.Fatalf("expected Offset=0, got %d", p.Offset())
	}
	if p.Limit() != 25 {
		t.Fatalf("expected Limit=25, got %d", p.Limit())
	}
}

func TestNewPaginationMeta_Computation(t *testing.T) {
	p := NewPaginationMeta(2, 10, 35)
	if p.Page != 2 || p.PerPage != 10 || p.Total != 35 {
		t.Fatalf("unexpected meta: %+v", p)
	}
	if p.TotalPages != 4 {
		t.Fatalf("expected TotalPages=4, got %d", p.TotalPages)
	}
	if !p.HasNext {
		t.Fatalf("expected HasNext=true")
	}
	if !p.HasPrev {
		t.Fatalf("expected HasPrev=true")
	}
	if p.Offset() != 10 {
		t.Fatalf("expected Offset=10, got %d", p.Offset())
	}
	if p.Limit() != 10 {
		t.Fatalf("expected Limit=10, got %d", p.Limit())
	}
}

