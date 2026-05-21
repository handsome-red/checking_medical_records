// internal/database/migrate.go
package database

import (
	"log"
	"med_book/internal/model"
)

func Migrate(db *Database) error {
	// Автомиграция
	err := db.GetDB().AutoMigrate(
		&model.Session{},
		&model.SessionProgress{},
		&model.SessionBook{},
		&model.SessionQuestion{},
		&model.SessionAnswer{},
		&model.User{},
	)

	if err != nil {
		return err
	}

	log.Println("Database migrated successfully!")
	return nil
}
