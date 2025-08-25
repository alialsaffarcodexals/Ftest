package models

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(path string, schemaPath string) error {
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	DB.SetConnMaxLifetime(0)
	DB.SetMaxIdleConns(5)
	DB.SetMaxOpenConns(1)

	if err := ping(); err != nil {
		return err
	}

	if err := execSchema(schemaPath); err != nil {
		return err
	}
	return nil
}

func ping() error {
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("sqlite ping: %w", err)
	}
	// ensure foreign keys
	if _, err := DB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return err
	}
	return nil
}

func execSchema(path string) error {
	b, err := os.ReadFile(path)
	if err != nil { return err }
	_, err = DB.Exec(string(b))
	return err
}

// --- small helpers ---

func Now() time.Time { return time.Now().UTC().Truncate(time.Second) }

var ErrNotFound = errors.New("not found")
