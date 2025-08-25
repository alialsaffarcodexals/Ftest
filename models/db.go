package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

const (
	DefaultQueryTimeout = 4 * time.Second
	SchemaTimeout       = 8 * time.Second
	PingTimeout         = 3 * time.Second
)

func InitDB(path, schemaPath string) error {
	dsn := fmt.Sprintf(
		"file:%s?cache=shared&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)",
		path,
	)
	var err error
	DB, err = sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	DB.SetMaxOpenConns(1)
	DB.SetMaxIdleConns(1)
	DB.SetConnMaxLifetime(0)

	{
		ctx, cancel := context.WithTimeout(context.Background(), PingTimeout)
		defer cancel()
		if err := DB.PingContext(ctx); err != nil {
			return fmt.Errorf("sqlite ping: %w", err)
		}
	}
	b, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}
	{
		ctx, cancel := context.WithTimeout(context.Background(), SchemaTimeout)
		defer cancel()
		if _, err := DB.ExecContext(ctx, string(b)); err != nil {
			return fmt.Errorf("apply schema: %w", err)
		}
	}
	return nil
}

func Now() time.Time { return time.Now().UTC().Truncate(time.Second) }

var ErrNotFound = errors.New("not found")
