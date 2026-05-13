package services

import (
	"eventor/internal/platform/config"
	file_storage "eventor/internal/platform/storage/files"
	user_service "eventor/internal/user/services/usecase/user"
	"eventor/internal/user/storage/repositories"
	//"github.com/redis/go-redis/v9"
	//"gopkg.in/gomail.v2"
)

type Services struct {
	UserService *user_service.UserService
}

func Init(repos *repositories.Repos, cfg *config.Config, fileStorage *file_storage.FileStorage) *Services {
	return &Services{
		UserService: user_service.Init(repos.UserRepo, fileStorage),
	}
}
