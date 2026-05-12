package storage

import (
	"database/sql"
	"eventor/internal/platform/config"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage struct {
	Db  *sql.DB
	cfg *config.Storage
}

func (storage *Storage) Connect() *sql.DB {
	conn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		storage.cfg.DbHost, storage.cfg.DbPort, storage.cfg.DbUser, storage.cfg.DbPass, storage.cfg.DbName,
	)

	db, err := sql.Open("postgres", conn)

	if err != nil {
		panic("Couldn't connect to db!")
	}

	storage.Db = db

	return db
}

func Init(cfg *config.Storage) *Storage {
	return &Storage{
		Db:  nil,
		cfg: cfg,
	}
}
