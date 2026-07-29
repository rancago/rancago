package valueobjects

import "testing"

func TestNewEmail(t *testing.T) {
	t.Run("valid address", func(t *testing.T) {
		e, err := NewEmail("user@example.com")
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if e.String() != "user@example.com" {
			t.Fatalf("expected value user@example.com, got %q", e.String())
		}
		if e.IsEmpty() {
			t.Fatalf("expected non-empty email")
		}
		if got := e.Domain(); got != "example.com" {
			t.Fatalf("expected domain example.com, got %q", got)
		}
	})

	t.Run("valid address with display name", func(t *testing.T) {
		e, err := NewEmail(`User Name <user@example.com>`)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if e.String() != "user@example.com" {
			t.Fatalf("expected parsed address user@example.com, got %q", e.String())
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		_, err := NewEmail("not-an-email")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestMustEmail(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		e := MustEmail("user@example.com")
		if e.String() != "user@example.com" {
			t.Fatalf("expected user@example.com, got %q", e.String())
		}
	})

	t.Run("panic on invalid", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("expected panic, got nil")
			}
		}()
		_ = MustEmail("bad")
	})
}

