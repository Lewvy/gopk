package config

import (
	"database/sql"
	"errors"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, errors.New("error loading env")
	}

	db_str := os.Getenv("GOOSE_DBSTRING")
	if db_str == "" {
		return nil, errors.New("db dsn not found")
	}

	db, err := sql.Open("sqlite3", db_str)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil

}
