package services

import (
	"eventor/internal/event/contracts/user_provider"
	event_service "eventor/internal/event/services/usecase/event"
	"eventor/internal/event/storage/repositories"
	"eventor/internal/platform/config"
	file_storage "eventor/internal/platform/storage/files"
	//"github.com/redis/go-redis/v9"
	//"gopkg.in/gomail.v2"
)

type Services struct {
	EventService *event_service.EventService
}

func Init(repos *repositories.Repos, cfg *config.Config, userProvider user_provider.UserProvider, fileStorage *file_storage.FileStorage) *Services {
	return &Services{
		EventService: event_service.Init(repos.EventRepo, userProvider, fileStorage),
	}
}
