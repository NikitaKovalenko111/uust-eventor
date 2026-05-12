package repositories

import (
	"database/sql"
	example_storage "example-service/internal/storage/repositories/example"
)

type Repos struct {
	ExampleRepo *example_storage.ExampleRepo
	// Here are repos
}

func Init(db *sql.DB) *Repos {
	return &Repos{
		ExampleRepo: example_storage.Init(db),
	}
}
