package domain_errors

import "errors"

var (
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrTokenExpired          = errors.New("token expired")
	ErrInvalidToken          = errors.New("invalid token")
	ErrUserNotFound          = errors.New("user not found")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrInvalidTokenSignature = errors.New("invalid token signature")
	ErrTokenGenerationFailed = errors.New("token generation failed")
	ErrRefreshTokenNotFound  = errors.New("refresh token not found")
	ErrRefreshTokenExpired   = errors.New("refresh token expired")
)
