package token_provider

import (
	"eventor/internal/auth/domain/models"
)

type TokenProvider interface {
	VerifyAccessToken(tokenString string) (*models.AccessTokenClaims, error)
}
