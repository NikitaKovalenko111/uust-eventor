package example_storage

import "database/sql"

type ExampleRepo struct {
	db *sql.DB
	// ...
}

func Init(db *sql.DB) *ExampleRepo {
	return &ExampleRepo{
		db: db,
	}
}
