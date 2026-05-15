package token_service

import (
	"database/sql"
	"errors"
	domain_errors "eventor/internal/auth/domain/errors"
	"eventor/internal/auth/domain/models"
	token_repo "eventor/internal/auth/storage/repositories/token"
	"eventor/internal/platform/config"
	"eventor/internal/platform/types"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	TokenRepo *token_repo.TokenRepo
	cfg       *config.JWT
}

func Init(tokenRepo *token_repo.TokenRepo, cfg *config.JWT) *TokenService {
	return &TokenService{
		TokenRepo: tokenRepo,
		cfg:       cfg,
	}
}

func (s *TokenService) GenerateAccessToken(userID types.IdType, email string, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.cfg.AccessTokenTTL)

	claims := &models.AccessTokenClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.cfg.Secret))

	if err != nil {
		return "", time.Time{}, domain_errors.ErrTokenGenerationFailed
	}

	return tokenString, expiresAt, nil
}

func (s *TokenService) GenerateRefreshToken(userID types.IdType, tokenID types.IdType) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.cfg.RefreshTokenTTL)

	claims := &models.RefreshTokenClaims{
		UserID:  userID,
		TokenID: tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.Secret))

	if err != nil {
		return "", time.Time{}, domain_errors.ErrTokenGenerationFailed
	}

	return tokenString, expiresAt, nil
}

func (s *TokenService) VerifyAccessToken(tokenString string) (*models.AccessTokenClaims, error) {
	claims := &models.AccessTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.Secret), nil
	})

	if err != nil {
		return nil, domain_errors.ErrInvalidToken
	}

	if !token.Valid {
		return nil, domain_errors.ErrInvalidToken
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, domain_errors.ErrTokenExpired
	}

	return claims, nil
}

func (s *TokenService) VerifyRefreshToken(tokenString string) (*models.RefreshTokenClaims, error) {
	claims := &models.RefreshTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.Secret), nil
	})

	if err != nil {
		return nil, domain_errors.ErrInvalidToken
	}

	if !token.Valid {
		return nil, domain_errors.ErrInvalidToken
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, domain_errors.ErrRefreshTokenExpired
	}

	return claims, nil
}

func (s *TokenService) CreateTokenPair(userID types.IdType, email string, role string, tx *sql.Tx) (*models.TokenPair, error) {
	accessToken, accessExpiresAt, err := s.GenerateAccessToken(userID, email, role)
	if err != nil {
		return nil, errors.Join(domain_errors.ErrTokenGenerationFailed, err)
	}

	refreshToken, _, err := s.GenerateRefreshToken(userID, userID)
	if err != nil {
		return nil, errors.Join(domain_errors.ErrTokenGenerationFailed, err)
	}

	token := models.CreateToken(userID, refreshToken, s.cfg.RefreshTokenTTL)

	_, err = s.TokenRepo.Create(token, tx)

	if err != nil {
		return nil, errors.Join(domain_errors.ErrTokenGenerationFailed, err)
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiresAt,
	}, nil
}

func (s *TokenService) RefreshTokenPair(refreshTokenString string, userID types.IdType, email string, role string, tx *sql.Tx) (*models.TokenPair, error) {
	var stx *sql.Tx

	if tx == nil {
		tx, err := s.TokenRepo.Db.Begin()

		if err != nil {
			return nil, err
		}

		stx = tx
		defer stx.Rollback()
	} else {
		stx = tx
	}

	claims, err := s.VerifyRefreshToken(refreshTokenString)

	if err != nil {
		if errors.Is(err, domain_errors.ErrTokenExpired) {
			return nil, domain_errors.ErrRefreshTokenExpired
		}
		return nil, domain_errors.ErrInvalidToken
	}

	if claims.UserID != userID {
		return nil, domain_errors.ErrInvalidToken
	}

	tokenPair, err := s.CreateTokenPair(userID, email, role, stx)
	if err != nil {
		return nil, err
	}

	if tx == nil {
		if err := stx.Commit(); err != nil {
			return nil, err
		}
	}

	return tokenPair, nil
}
