package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound      = errors.New("resource not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrValidation    = errors.New("validation error")
	ErrConflict      = errors.New("conflict")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidInput  = errors.New("invalid input")
	ErrInternal      = errors.New("internal error")
)

type DomainError struct {
	Op      string
	Err     error
	Message string
}

func (e *DomainError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *DomainError) Unwrap() error { return e.Err }
func (e *DomainError) Is(target error) bool { return errors.Is(e.Err, target) }

func New(op string, err error, msg string) error {
	return &DomainError{Op: op, Err: err, Message: msg}
}

func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return &DomainError{Op: op, Err: err}
}

func IsNotFound(err error) bool      { return errors.Is(err, ErrNotFound) }
func IsUnauthorized(err error) bool  { return errors.Is(err, ErrUnauthorized) }
func IsForbidden(err error) bool     { return errors.Is(err, ErrForbidden) }
func IsValidation(err error) bool    { return errors.Is(err, ErrValidation) }
func IsConflict(err error) bool      { return errors.Is(err, ErrConflict) }
func IsAlreadyExists(err error) bool { return errors.Is(err, ErrAlreadyExists) }
func IsInvalidInput(err error) bool  { return errors.Is(err, ErrInvalidInput) }
