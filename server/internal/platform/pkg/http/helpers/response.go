package helpers

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

// FormatValidationErrors форматирует ошибки validator в map[field]message
// Используется в контроллерах для возврата деталей валидации в ответе
func FormatValidationErrors(err error) map[string]string {
	details := make(map[string]string)
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			field := e.Field()
			details[field] = validationMessage(e)
		}
	}
	return details
}

// ExtractValidationErrorDetails распаковывает обёрнутые ошибки валидации
// Например: "validation error: email: invalid format" → {"email": "invalid format"}
func ExtractValidationErrorDetails(err error) map[string]string {
	for err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			return FormatValidationErrors(errs)
		}
		err = errors.Unwrap(err)
	}
	return nil
}

// validationMessage возвращает человеко-читаемое сообщение по тегу валидатора
func validationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be valid email"
	case "min":
		return "must be at least " + e.Param() + " characters"
	case "max":
		return "must not exceed " + e.Param() + " characters"
	case "oneof":
		return "must be one of: " + e.Param()
	case "len":
		return "must be exactly " + e.Param() + " characters"
	case "numeric":
		return "must be a number"
	case "url":
		return "must be valid URL"
	case "datetime":
		return "must be valid RFC3339 datetime"
	default:
		return "invalid value"
	}
}
