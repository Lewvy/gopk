package data

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pressly/goose/v3"
)

func ResetDB(db *sql.DB, dbPath string) error {
	if err := backupFile(dbPath); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}
	fmt.Println("Backup created successfully.")

	migrationDir := "./db/migrations"

	if err := goose.Reset(db, migrationDir); err != nil {
		return fmt.Errorf("goose reset failed: %w", err)
	}
	fmt.Println("Database reset (all tables dropped).")

	if err := goose.Up(db, migrationDir); err != nil {
		return fmt.Errorf("goose up failed: %w", err)
	}
	fmt.Println("Database migrated up (fresh schema).")

	return nil
}

func backupFile(src string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	backupName := fmt.Sprintf("%s.bak.%d", src, time.Now().Unix())
	destFile, err := os.Create(backupName)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
