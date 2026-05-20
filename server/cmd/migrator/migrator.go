package main

import (
	"database/sql"
	"errors"
	"eventor/internal/platform/config"
	"eventor/internal/platform/storage"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const defaultMigrationsPath = "./internal/platform/storage/migrations"

type migrationFile struct {
	Name string
	Path string
}

func main() {
	migrationsPath := flag.String("migrations-path", defaultMigrationsPath, "path to migration files")
	flag.Parse()

	db, err := openDB()

	if err != nil {
		log.Fatalf("migrator: open db: %v", err)
	}

	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("migrator: close db: %v", closeErr)
		}
	}()

	if err = db.Ping(); err != nil {
		log.Fatalf("migrator: ping db: %v", err)
	}

	if err = ensureMigrationsTable(db); err != nil {
		log.Fatalf("migrator: ensure schema_migrations: %v", err)
	}

	files, err := collectMigrationFiles(*migrationsPath)
	if err != nil {
		log.Fatalf("migrator: collect files: %v", err)
	}

	if len(files) == 0 {
		log.Printf("migrator: no *_up.sql files found in %s", *migrationsPath)
		return
	}

	applied, err := appliedMigrations(db)
	if err != nil {
		log.Fatalf("migrator: load applied migrations: %v", err)
	}

	appliedCount := 0
	for _, file := range files {
		if applied[file.Name] {
			log.Printf("skip: %s", file.Name)
			continue
		}

		if err = applySingleMigration(db, file); err != nil {
			log.Fatalf("migrator: apply %s: %v", file.Name, err)
		}

		appliedCount++
		log.Printf("applied: %s", file.Name)
	}

	log.Printf("migrator finished, applied=%d", appliedCount)
}

func openDB() (*sql.DB, error) {
	cfg := config.MustLoad()
	st := storage.Init(cfg)
	db := st.Connect()

	if db == nil {
		return nil, errors.New("database is nil after connect")
	}

	return db, nil
}

func ensureMigrationsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		id BIGSERIAL PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	`

	_, err := db.Exec(query)
	return err
}

func collectMigrationFiles(migrationsPath string) ([]migrationFile, error) {
	info, err := os.Stat(migrationsPath)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", migrationsPath)
	}

	items := make([]migrationFile, 0)
	err = filepath.WalkDir(migrationsPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			return nil
		}

		name := d.Name()
		if !strings.HasSuffix(name, "_up.sql") {
			return nil
		}

		items = append(items, migrationFile{Name: name, Path: path})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return items, nil
}

func appliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query(`SELECT name FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]bool)
	for rows.Next() {
		var name string
		if scanErr := rows.Scan(&name); scanErr != nil {
			return nil, scanErr
		}
		result[name] = true
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}

	return result, nil
}

func applySingleMigration(db *sql.DB, file migrationFile) error {
	data, err := os.ReadFile(file.Path)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec(string(data)); err != nil {
		return fmt.Errorf("exec migration SQL: %w", err)
	}

	if _, err = tx.Exec(`INSERT INTO schema_migrations(name, applied_at) VALUES ($1, $2)`, file.Name, time.Now().UTC()); err != nil {
		return fmt.Errorf("register migration: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
