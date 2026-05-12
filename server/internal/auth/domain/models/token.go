package models

import (
	"crypto/sha256"
	"eventor/internal/platform/types"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Token struct {
	Id        types.IdType
	UserId    types.IdType
	Hash      string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type AccessTokenClaims struct {
	UserID types.IdType `json:"user_id"`
	Email  string       `json:"email"`
	Role   string       `json:"role"`
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	UserID  types.IdType `json:"user_id"`
	TokenID types.IdType `json:"token_id"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func CreateToken(userId types.IdType, token string, tokenTTL time.Duration) *Token {
	hash := sha256.Sum256([]byte(token))
	hashString := fmt.Sprintf("%x", hash)

	return &Token{
		Id:        0,
		UserId:    userId,
		Hash:      hashString,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(tokenTTL),
	}
}

func (t *Token) UpdateToken(token string, tokenTTL time.Duration) error {
	hash := sha256.Sum256([]byte(token))
	hashString := fmt.Sprintf("%x", hash)

	t.Hash = hashString
	t.CreatedAt = time.Now()
	t.ExpiresAt = time.Now().Add(tokenTTL)

	return nil
}

func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}
