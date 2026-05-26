// internal/database/migrate.go
package database

import (
	"log"
	"med_book/internal/model"
)

func Migrate(db *Database) error {
	// Автомиграция
	err := db.GetDB().AutoMigrate(
		&model.User{},
		&model.Session{},
		&model.Book{},
		&model.BookPage{},
		&model.Question{},
		&model.UserAnswer{},
		// &model.Session{},
		// &model.Book{},
		// &model.Question{},
		// &model.UserAnswer{},

	)

	if err != nil {
		return err
	}

	log.Println("Database migrated successfully!")
	return nil
}
