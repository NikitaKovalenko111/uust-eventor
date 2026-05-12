package services

import (
	"example-service/internal/config"
	example_service "example-service/internal/services/usecase/example"
	"example-service/internal/storage/repositories"
	//"github.com/redis/go-redis/v9"
	//"gopkg.in/gomail.v2"
)

type Services struct {
	ExampleService *example_service.ExampleService
}

func Init(repos *repositories.Repos, cfg *config.Config) *Services {
	return &Services{
		ExampleService: example_service.Init(repos.ExampleRepo),
	}
}
