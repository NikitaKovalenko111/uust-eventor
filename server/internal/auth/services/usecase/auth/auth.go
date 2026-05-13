package auth_service

import (
	"context"
	"errors"
	"eventor/internal/auth/contracts/user_provider"
	domain_errors "eventor/internal/auth/domain/errors"
	"eventor/internal/auth/domain/models"
	token_service "eventor/internal/auth/services/usecase/token"
	user_errors "eventor/internal/user/domain/errors"
	user_dto "eventor/internal/user/transport/http/dto/user"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	TokenService *token_service.TokenService
	UserProvider user_provider.UserProvider
}

func Init(tokenService *token_service.TokenService, userProvider user_provider.UserProvider) *AuthService {
	return &AuthService{
		TokenService: tokenService,
		UserProvider: userProvider,
	}
}

func (s *AuthService) Register(city string, name string, email string, password string, role string) (*models.TokenPair, error) {
	existingUser, err := s.UserProvider.GetByEmail(context.Background(), email)

	if err != nil {
		if !errors.Is(err, user_errors.ErrUserNotFound) {
			return nil, err
		}
	}

	if existingUser != nil {
		return nil, domain_errors.ErrEmailAlreadyExists
	}

	tx, err := s.TokenService.TokenRepo.Db.Begin()

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	user, err := s.UserProvider.Create(context.Background(), &user_dto.CreateUserRequest{
		City:     city,
		Role:     role,
		Name:     name,
		Email:    email,
		Password: password,
	}, tx)

	if err != nil {
		return nil, err
	}

	tokenPair, err := s.TokenService.CreateTokenPair(user.ID, user.Email, user.Role, tx)
	if err != nil {
		return nil, err
	}

	tx.Commit()

	return tokenPair, nil
}

func (s *AuthService) Login(email string, password string) (*models.TokenPair, error) {
	user, err := s.UserProvider.GetByEmail(context.Background(), email)

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domain_errors.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, domain_errors.ErrInvalidCredentials
	}

	tx, err := s.TokenService.TokenRepo.Db.Begin()

	if err != nil {
		return nil, err
	}

	tokenPair, err := s.TokenService.CreateTokenPair(user.ID, user.Email, user.Role, tx)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}
