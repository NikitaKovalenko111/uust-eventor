package storage

import (
	"database/sql"
	"eventor/internal/platform/config"
	minio "eventor/internal/platform/storage/files"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type Storage struct {
	Db          *sql.DB
	FileStorage *minio.FileStorage
	cfg         *config.Config
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

	// Ensure the DB is reachable
	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("Couldn't ping db: %v", err))
	}

	// If this is the first time (no users table), run initial migration
	var exists bool
	row := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='users')")
	if err := row.Scan(&exists); err != nil {
		panic(fmt.Sprintf("Couldn't check schema existence: %v", err))
	}

	if !exists {
		// read and execute initial migration
		sqlBytes, err := os.ReadFile("internal/platform/storage/migrations/1_init_up.sql")
		if err != nil {
			panic(fmt.Sprintf("Couldn't read migration file: %v", err))
		}

		if _, err := db.Exec(string(sqlBytes)); err != nil {
			panic(fmt.Sprintf("Failed to apply initial migration: %v", err))
		}
	}

	// Apply idempotent incremental schema updates for existing databases.
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS event_comments (
			id SERIAL PRIMARY KEY,
			event_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			text TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_event_comments_event_created ON event_comments(event_id, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_event_comments_user ON event_comments(user_id);
	`); err != nil {
		panic(fmt.Sprintf("Failed to apply incremental schema updates: %v", err))
	}

	storage.Db = db

	return db
}

func (storage *Storage) NewFileStorage() (*minio.FileStorage, error) {
	fileStorage, err := minio.NewMinioStorage(
		storage.cfg.FileStorage.Endpoint, storage.cfg.FileStorage.AccessKey, storage.cfg.FileStorage.SecretKey, storage.cfg.FileStorage.UseSSL,
	)

	if err != nil {
		return nil, err
	}

	storage.FileStorage = fileStorage

	return fileStorage, nil
}

func Init(cfg *config.Config) *Storage {
	return &Storage{
		Db:          nil,
		FileStorage: nil,
		cfg:         cfg,
	}
}
