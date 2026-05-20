package auth_dto

// RefreshTokenRequest
// @Description Request body for refreshing authentication tokens
type RefreshTokenRequest struct {
	// Refresh token obtained from previous login/registration
	// @example "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM..."}
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LoginRequest
// @Description Request body for user authentication
type LoginRequest struct {
	// User's email address (must be valid email format)
	// @example "user@example.com"
	// @MinLength(1)
	Email string `json:"email" validate:"required,email"`
	// User's password (minimum 6 characters)
	// @example "SecurePass123"
	// @MinLength(6)
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest
// @Description Request body for user/moderator registration
type RegisterRequest struct {
	// User's city of residence (optional)
	// @example "Moscow"
	City string `json:"city"`
	// User's full name (minimum 3 characters)
	// @example "Ivan Ivanov"
	// @MinLength(3)
	Name string `json:"name" validate:"required,min=3"`
	// User's email address (must be valid and unique)
	// @example "user@example.com"
	Email string `json:"email" validate:"required,email"`
	// User's password (minimum 6 characters)
	// @example "SecurePass123"
	// @MinLength(6)
	Password string `json:"password" validate:"required,min=6"`
	// User role: "user" (default) or "moderator" (protected endpoint)
	// @enum user moderator
	// @default "user"
	// @example "user"
	Role string `json:"role" validate:"omitempty,oneof=user moderator"`
}

// AuthResponse
// @Description Response containing authentication tokens
type AuthResponse struct {
	// JWT access token for authorized requests
	// @example "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM..."
	AccessToken string `json:"access_token"`
	// JWT refresh token for obtaining new access tokens
	// @example "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyZWZyZXNo..."
	RefreshToken string `json:"refresh_token"`
	// Token expiration time in RFC3339 format
	// @example "2024-12-31T23:59:59Z"
	ExpiresAt string `json:"expires_at"`
}

// ErrorResponse
// @Description Standard error response structure
type ErrorResponse struct {
	// Human-readable error message
	// @example "invalid credentials"
	Error string `json:"error"`
}

// ModeratorRegisterResponse
// @Description Response for moderator registration (includes generated password if applicable)
type ModeratorRegisterResponse struct {
	// JWT access token for the new moderator
	// @example "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
	AccessToken string `json:"access_token"`
	// JWT refresh token for the new moderator
	// @example "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
	RefreshToken string `json:"refresh_token"`
	// Token expiration time in RFC3339 format
	// @example "2024-12-31T23:59:59Z"
	ExpiresAt string `json:"expires_at"`
	// Generated password (only present if not provided in request)
	// @example "aB3xK9mP2nQ7"
	Password string `json:"password,omitempty"`
}
