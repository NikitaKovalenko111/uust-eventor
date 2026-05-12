package example_service

import example_storage "example-service/internal/storage/repositories/example"

type ExampleService struct {
	ExampleRepo *example_storage.ExampleRepo
	// Here are repositories associated with this service
}

func Init(exampleRepo *example_storage.ExampleRepo) *ExampleService {
	return &ExampleService{
		ExampleRepo: exampleRepo,
	}
}
