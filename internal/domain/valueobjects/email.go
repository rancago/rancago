package valueobjects

import (
	"fmt"
	"net/mail"
)

type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	addr, err := mail.ParseAddress(raw)
	if err != nil {
		return Email{}, fmt.Errorf("invalid email format: %w", err)
	}
	if addr.Address == "" {
		return Email{}, fmt.Errorf("email cannot be empty")
	}
	return Email{value: addr.Address}, nil
}

func MustEmail(raw string) Email {
	e, err := NewEmail(raw)
	if err != nil {
		panic(err)
	}
	return e
}

func (e Email) String() string  { return e.value }
func (e Email) IsEmpty() bool   { return e.value == "" }
func (e Email) Domain() string {
	for i := len(e.value) - 1; i >= 0; i-- {
		if e.value[i] == '@' {
			return e.value[i+1:]
		}
	}
	return ""
}
