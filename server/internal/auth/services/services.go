package services

import (
	"eventor/internal/auth/contracts/user_provider"
	auth_service "eventor/internal/auth/services/usecase/auth"
	token_service "eventor/internal/auth/services/usecase/token"
	"eventor/internal/auth/storage/repositories"
	"eventor/internal/platform/config"
)

//"github.com/redis/go-redis/v9"
//"gopkg.in/gomail.v2"

type Services struct {
	AuthService  *auth_service.AuthService
	TokenService *token_service.TokenService
}

func Init(repos *repositories.Repos, cfg *config.Config, userProvider user_provider.UserProvider) *Services {
	tokenService := token_service.Init(repos.TokenRepo, &cfg.JWT)
	authService := auth_service.Init(tokenService, userProvider)

	return &Services{
		AuthService:  authService,
		TokenService: tokenService,
	}
}
