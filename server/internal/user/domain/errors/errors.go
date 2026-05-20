// internal/user/domain/errors/errors.go
package errors

import "errors"

// ==========================================
// Domain errors (бизнес-ошибки)
// ==========================================

var (
	// Validation errors
	ErrValidation       = errors.New("validation error")
	ErrNameRequired     = errors.New("name is required")
	ErrNameTooShort     = errors.New("name must be at least 2 characters")
	ErrNameTooLong      = errors.New("name must not exceed 255 characters")
	ErrNameInvalidChars = errors.New("name contains invalid characters")

	ErrEmailRequired = errors.New("email is required")
	ErrEmailInvalid  = errors.New("email has invalid format")
	ErrEmailTooLong  = errors.New("email must not exceed 255 characters")

	ErrPasswordRequired = errors.New("password is required")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong  = errors.New("password must not exceed 128 characters")

	ErrCityRequired = errors.New("city is required")
	ErrCityTooShort = errors.New("city must be at least 2 characters")
	ErrCityTooLong  = errors.New("city must not exceed 128 characters")

	ErrRoleRequired = errors.New("role is required")
	ErrRoleInvalid  = errors.New("role must be 'user' or 'moderator'")

	ErrFacultyInvalid  = errors.New("faculty has invalid format or too long")
	ErrCourseInvalid   = errors.New("course has invalid format or too long")
	ErrAboutTooLong    = errors.New("about must not exceed 1000 characters")
	ErrAvatarIDInvalid = errors.New("avatar_image_id has invalid format")
	ErrAvatarNotFound  = errors.New("avatar not found")
	ErrFileStorage     = errors.New("file storage error")

	// Business logic errors
	ErrNotFound         = errors.New("not found")
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailExists      = errors.New("email already exists")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrPasswordMismatch = errors.New("password does not match")

	// Infrastructure errors
	ErrDatabase = errors.New("database error")
	ErrHashing  = errors.New("failed to hash password")
)

// ==========================================
// Error helpers
// ==========================================

// IsValidation проверяет, является ли ошибка ошибкой валидации
func IsValidation(err error) bool {
	return errors.Is(err, ErrValidation)
}

// IsNotFound проверяет, является ли ошибка "не найдено"
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) || errors.Is(err, ErrUserNotFound)
}

// IsConflict проверяет конфликт ресурсов (дубликат email)
func IsConflict(err error) bool {
	return errors.Is(err, ErrEmailExists)
}

// WrapDB оборачивает ошибку БД в доменную
func WrapDB(err error) error {
	if err == nil {
		return nil
	}
	return &DBError{Err: err}
}

// DBError обёртка для ошибок базы данных
type DBError struct {
	Err error
}

func (e *DBError) Error() string { return e.Err.Error() }
func (e *DBError) Unwrap() error { return e.Err }
func (e *DBError) Is(target error) bool {
	return errors.Is(target, ErrDatabase)
}
