package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_import "med_book/internal/import"
	"med_book/internal/model"
)

func main() {
	var (
		dsn          = flag.String("dsn", "./data/med_book.db?_journal_mode=WAL&_foreign_keys=1", "SQLite DSN")
		filePath     = flag.String("file", "pkg/questions/questions.json", "Путь к JSON файлу")
		dropDB       = flag.Bool("drop", false, "Очистить все книги перед импортом")
		skipIfExists = flag.Bool("skip", false, "Пропускать существующие книги")
	)
	flag.Parse()

	if err := os.MkdirAll("data", 0o755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	db, err := gorm.Open(sqlite.Open(*dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	fmt.Println("Database connected")

	fmt.Println("Running migrations...")
	if err := db.AutoMigrate(
		&model.Book{},
		&model.BookPage{},
		&model.Image{},
		&model.Question{},
		&model.Option{},
		&model.User{},
		&model.Session{},
		&model.UserAnswer{},
	); err != nil {
		log.Fatal("Migration failed:", err)
	}

	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		log.Fatalf("File not found: %s", *filePath)
	}

	service := _import.NewImportService(db)

	if *dropDB {
		fmt.Println("Clearing books...")
		if err := service.ClearBooks(); err != nil {
			log.Fatal("Clear failed:", err)
		}
	}

	fmt.Printf("Import from: %s\n\n", *filePath)

	var importErr error
	if *skipIfExists {
		importErr = service.ImportBooksIfNotExists(*filePath)
	} else {
		importErr = service.ImportFromFile(*filePath)
	}

	if importErr != nil {
		log.Fatal("Import failed:", importErr)
	}

	printStats(db)
	fmt.Println("\nImport completed successfully")
}

func printStats(db *gorm.DB) {
	var bookCount, questionCount, optionCount, userCount, sessionCount, answerCount int64
	db.Model(&model.Book{}).Count(&bookCount)
	db.Model(&model.Question{}).Count(&questionCount)
	db.Model(&model.Option{}).Count(&optionCount)
	db.Model(&model.User{}).Count(&userCount)
	db.Model(&model.Session{}).Count(&sessionCount)
	db.Model(&model.UserAnswer{}).Count(&answerCount)

	fmt.Println("\nDatabase stats:")
	fmt.Printf("  Books: %d\n", bookCount)
	fmt.Printf("  Questions: %d\n", questionCount)
	fmt.Printf("  Options: %d\n", optionCount)
	fmt.Printf("  Users: %d\n", userCount)
	fmt.Printf("  Sessions: %d\n", sessionCount)
	fmt.Printf("  Answers: %d\n", answerCount)
}
