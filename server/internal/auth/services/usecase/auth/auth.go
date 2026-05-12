package user_service

import (
	"errors"
	domain_errors "eventor/internal/auth/domain/errors"
	"eventor/internal/auth/domain/models"
	token_service "eventor/internal/auth/services/usecase/token"
	user_repo "eventor/internal/auth/storage/repositories/user"
	"eventor/internal/platform/types"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	UserRepo     *user_repo.UserRepo
	TokenService *token_service.TokenService
}

func Init(userRepo *user_repo.UserRepo, tokenService *token_service.TokenService) *UserService {
	return &UserService{
		UserRepo:     userRepo,
		TokenService: tokenService,
	}
}

// Register creates a new user and returns a token pair
func (us *UserService) Register(email string, password string, role string) (*models.TokenPair, error) {
	// Check if user already exists
	existingUser, err := us.UserRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		return nil, domain_errors.ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Join(domain_errors.ErrTokenGenerationFailed, err)
	}

	// Create user
	user, err := us.UserRepo.Create(email, string(hashedPassword), role)
	if err != nil {
		return nil, err
	}

	// Generate token pair
	tokenPair, err := us.TokenService.CreateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// Login authenticates a user and returns a token pair
func (us *UserService) Login(email string, password string) (*models.TokenPair, error) {
	// Find user by email
	user, err := us.UserRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domain_errors.ErrInvalidCredentials
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, domain_errors.ErrInvalidCredentials
	}

	// Generate token pair
	tokenPair, err := us.TokenService.CreateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// GetUser retrieves user by ID
func (us *UserService) GetUser(userID types.IdType) (*models.User, error) {
	return us.UserRepo.FindByID(userID)
}
