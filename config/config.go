package config

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	migrations "github.com/lewvy/gopk/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

func InitDB() (*sql.DB, error) {

	home_dir, err1 := os.UserHomeDir()

	path := filepath.Join(home_dir, ".local/share/gopk")
	err2 := os.MkdirAll(path, 0700)

	if err := errors.Join(err1, err2); err != nil {
		return nil, err
	}

	db_str := filepath.Join(path, "packages.db")
	if path == "" {
		return nil, errors.New("db dsn not found")
	}

	goose.SetBaseFS(migrations.FS)
	if os.Getenv("DEBUG") == "true" {
	} else {
		goose.SetLogger(log.New(io.Discard, "", 0))
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", db_str)
	if err != nil {
		return nil, err
	}

	if err := goose.Up(db, "schema"); err != nil {
		return nil, fmt.Errorf("error running migrations: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil

}
