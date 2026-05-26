package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_import "med_book/internal/import"
	"med_book/internal/model"
)

func main() {
	// Параметры командной строки
	var (
		dsn          = flag.String("dsn", "", "Connection string для PostgreSQL")
		filePath     = flag.String("file", "data.json", "Путь к JSON файлу")
		dropDB       = flag.Bool("drop", false, "Очистить все книги перед импортом")
		skipIfExists = flag.Bool("skip", false, "Пропускать существующие книги")
	)
	flag.Parse()

	// DSN по умолчанию
	if *dsn == "" {
		*dsn = "host=localhost user=atlas password=123 dbname=med_book port=5432 sslmode=disable TimeZone=UTC"
		fmt.Println("⚠️  Используется DSN по умолчанию:", *dsn)
	}

	// Подключаемся к БД
	db, err := gorm.Open(postgres.Open(*dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("❌ Ошибка подключения к БД:", err)
	}
	fmt.Println("✅ Подключение к БД установлено")

	// Выполняем миграцию
	fmt.Println("📦 Выполняем миграцию схемы...")
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
		log.Fatal("❌ Ошибка миграции:", err)
	}
	fmt.Println("✅ Миграция завершена")

	// Проверяем существование файла
	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		log.Fatalf("❌ Файл не найден: %s", *filePath)
	}

	// Создаём сервис импорта
	service := _import.NewImportService(db)

	// Очищаем данные если нужно
	if *dropDB {
		fmt.Println("⚠️  Очистка всех книг...")
		if err := service.ClearBooks(); err != nil {
			log.Fatal("❌ Ошибка очистки:", err)
		}
	}

	// Импортируем данные
	fmt.Printf("📖 Импорт из файла: %s\n\n", *filePath)

	if *skipIfExists {
		err = service.ImportBooksIfNotExists(*filePath)
	} else {
		err = service.ImportFromFile(*filePath)
	}

	if err != nil {
		log.Fatal("❌ Ошибка импорта:", err)
	}

	// Выводим статистику
	printStats(db)

	fmt.Println("\n🎉 Импорт успешно завершён!")
}

func printStats(db *gorm.DB) {
	var (
		bookCount     int64
		questionCount int64
		optionCount   int64
		userCount     int64
		sessionCount  int64
		answerCount   int64
	)

	db.Model(&model.Book{}).Count(&bookCount)
	db.Model(&model.Question{}).Count(&questionCount)
	db.Model(&model.Option{}).Count(&optionCount)
	db.Model(&model.User{}).Count(&userCount)
	db.Model(&model.Session{}).Count(&sessionCount)
	db.Model(&model.UserAnswer{}).Count(&answerCount)

	fmt.Println("\n📊 Статистика базы данных:")
	fmt.Printf("   📚 Книг: %d\n", bookCount)
	fmt.Printf("   ❓ Вопросов: %d\n", questionCount)
	fmt.Printf("   🔘 Вариантов ответов: %d\n", optionCount)
	fmt.Printf("   👥 Пользователей: %d\n", userCount)
	fmt.Printf("   📝 Сессий: %d\n", sessionCount)
	fmt.Printf("   💬 Ответов: %d\n", answerCount)
}
