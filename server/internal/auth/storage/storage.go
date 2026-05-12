package storage

import (
	"database/sql"
	"example-service/internal/config"
	"example-service/internal/storage/repositories"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage struct {
	Db           *sql.DB
	Repositories *repositories.Repos
}

func Connect(cfg *config.Config) *sql.DB {
	var connString = fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPass, cfg.DbName,
	)

	db, err := sql.Open("postgres", connString)

	if err != nil {
		panic(fmt.Sprintf("Couldn't connect to database! Error: %s", err.Error()))
	}

	return db
}

func (storage *Storage) Prepare() {
	// SQL Queries to create tables
}

func Init(db *sql.DB) *Storage {
	storage := Storage{
		Db:           db,
		Repositories: repositories.Init(),
	}

	return &storage
}
