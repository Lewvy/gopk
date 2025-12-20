package config

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	migrations "github.com/lewvy/gopk/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

func getDataDir() (string, error) {

	if custom := os.Getenv("GOPK_DB_DIR"); custom != "" {
		return custom, nil
	}

	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "gopk"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "gopk"), nil
}

func InitDB() (*sql.DB, error) {

	path, err := getDataDir()
	if err != nil {
		return nil, fmt.Errorf("failed to determine data dir: %w", err)
	}

	if err := os.MkdirAll(path, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}

	dbPath := filepath.Join(path, "packages.db")

	goose.SetBaseFS(migrations.FS)

	if os.Getenv("DEBUG") != "true" {
		goose.SetLogger(log.New(io.Discard, "", 0))
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := goose.Up(db, "schema"); err != nil {
		return nil, fmt.Errorf("error running migrations: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func ResetDB(db *sql.DB) error {

	path, err := getDataDir()
	if err != nil {
		return err
	}
	dbPath := filepath.Join(path, "packages.db")

	if err := backupFile(dbPath); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}
	fmt.Printf("Backup saved to: %s.bak\n", dbPath)

	goose.SetBaseFS(migrations.FS)
	if err := goose.Reset(db, "schema"); err != nil {
		return fmt.Errorf("reset failed: %w", err)
	}

	if err := goose.Up(db, "schema"); err != nil {
		return fmt.Errorf("migration up failed: %w", err)
	}

	return nil
}

func backupFile(src string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(src + ".bak")
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	return err
}
