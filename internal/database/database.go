package database

import (
    "database/sql"
    "fmt"
    "log/slog"
    _ "github.com/mattn/go-sqlite3"
    "med_book/internal/config"
)

var DB *sql.DB

func InitDB(cfg *config.DatabaseConfig) error {
	
}

func CloseDB() error {
	if DB != nil {
		slog.Info("Closing database connection")
		return DB.Close()
	}
	return nil
}

func createTables() error {
	createUserTable := `
	CREATE TABLE IF NOT EXIST users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		firstName TEXT NOT NULL,
		lastName TEXT NOT NULL,
		patronymic TEXT NOT NULL,
	);`

	if _, err := DB.Exec(createUserTable); err != nil {
		return err
	}

	slog.Info("Tables created successfully")
	return nil
}

func createUsersTable() error {
	createUserTable := `
	CREATE TABLE IF NOT EXIST users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		firstName TEXT NOT NULL,
		lastName TEXT NOT NULL,
		patronymic TEXT NOT NULL,
	);`

	if _, err := DB.Exec(createUserTable); err != nil {
		return err
	}

	slog.Info("Tables created successfully")
	return nil
}