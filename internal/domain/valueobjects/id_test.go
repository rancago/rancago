package valueobjects

import "testing"

func TestID_UintBackAndForth(t *testing.T) {
	id := NewIDUint(42)

	if id.IsString() {
		t.Fatalf("expected IsString=false")
	}
	if id.IsZero() {
		t.Fatalf("expected IsZero=false")
	}
	if id.String() != "42" {
		t.Fatalf("expected String=42, got %q", id.String())
	}

	u, err := id.Uint()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if u != 42 {
		t.Fatalf("expected Uint=42, got %d", u)
	}
}

func TestID_StringBackAndForth(t *testing.T) {
	t.Run("numeric string can convert to uint", func(t *testing.T) {
		id := NewIDStr("7")
		if !id.IsString() {
			t.Fatalf("expected IsString=true")
		}
		if id.String() != "7" {
			t.Fatalf("expected String=7, got %q", id.String())
		}

		u, err := id.Uint()
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if u != 7 {
			t.Fatalf("expected Uint=7, got %d", u)
		}
	})

	t.Run("non-numeric string fails to convert to uint", func(t *testing.T) {
		id := NewIDStr("abc")
		_, err := id.Uint()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("empty string is zero", func(t *testing.T) {
		id := NewIDStr("")
		if !id.IsZero() {
			t.Fatalf("expected IsZero=true")
		}
	})
}

