package services

import (
	"eventor/internal/platform/config"
	user_service "eventor/internal/user/services/usecase/user"
	"eventor/internal/user/storage/repositories"
	//"github.com/redis/go-redis/v9"
	//"gopkg.in/gomail.v2"
)

type Services struct {
	UserService *user_service.UserService
}

func Init(repos *repositories.Repos, cfg *config.Config) *Services {
	return &Services{
		UserService: user_service.Init(repos.UserRepo),
	}
}
