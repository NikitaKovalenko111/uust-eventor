package helpers

// ErrorResponse универсальный формат ошибки для всех контроллеров
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    int               `json:"code"`
	Details map[string]string `json:"details,omitempty"`
}

// SuccessResponse универсальный ответ с данными
type SuccessResponse[T any] struct {
	Data    T      `json:"data"`
	Message string `json:"message,omitempty"`
}
