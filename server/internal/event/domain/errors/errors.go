// internal/event/domain/errors/errors.go
package errors

import (
	stderrors "errors"
	"fmt"
)

// ==========================================
// Domain errors (бизнес-ошибки событий)
// ==========================================
var (
	// Validation errors
	ErrValidation          = stderrors.New("validation error")
	ErrTitleRequired       = stderrors.New("title is required")
	ErrTitleTooLong        = stderrors.New("title must not exceed 255 characters")
	ErrDescriptionRequired = stderrors.New("description is required")
	ErrDescriptionTooLong  = stderrors.New("description must not exceed 5000 characters")
	ErrDateRequired        = stderrors.New("event_date is required")
	ErrDateInPast          = stderrors.New("event date cannot be in the past")
	ErrLocationRequired    = stderrors.New("location is required")
	ErrLocationTooLong     = stderrors.New("location must not exceed 255 characters")
	ErrImageIDInvalid      = stderrors.New("invalid image_id format or too long")

	// Business logic errors
	ErrNotFound            = stderrors.New("not found")
	ErrEventNotFound       = stderrors.New("event not found")
	ErrCreatorNotFound     = stderrors.New("creator (user) not found")
	ErrUnauthorized        = stderrors.New("unauthorized")
	ErrForbidden           = stderrors.New("forbidden")
	ErrForeignKeyViolation = stderrors.New("foreign key violation (creator_id does not exist)")

	// Infrastructure errors
	ErrDatabase = stderrors.New("database error")
)

// ==========================================
// Error helpers (проверки типов ошибок)
// ==========================================

// IsValidation проверяет, является ли ошибка ошибкой валидации
func IsValidation(err error) bool {
	return stderrors.Is(err, ErrValidation)
}

// IsNotFound проверяет, является ли ошибка "не найдено"
func IsNotFound(err error) bool {
	return stderrors.Is(err, ErrNotFound) ||
		stderrors.Is(err, ErrEventNotFound) ||
		stderrors.Is(err, ErrCreatorNotFound)
}

// IsConflict проверяет конфликт ресурсов (нарушение FK или уникальности)
func IsConflict(err error) bool {
	return stderrors.Is(err, ErrForeignKeyViolation)
}

// WrapDB оборачивает ошибку БД в доменную
func WrapDB(err error) error {
	if err == nil {
		return nil
	}
	return &DBError{Err: err}
}

// NewValidationError удобная обёртка для валидации полей
// Пример: errors.NewValidationError("title", "is required")
func NewValidationError(field, msg string) error {
	return fmt.Errorf("%w: %s: %s", ErrValidation, field, msg)
}

// ==========================================
// DBError (обёртка для ошибок PostgreSQL)
// ==========================================

// DBError обёртка для ошибок базы данных
type DBError struct {
	Err error
}

func (e *DBError) Error() string { return e.Err.Error() }
func (e *DBError) Unwrap() error { return e.Err }
func (e *DBError) Is(target error) bool {
	return stderrors.Is(target, ErrDatabase)
}
