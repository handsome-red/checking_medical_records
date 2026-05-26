package _import

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"med_book/internal/model"
)

type ImportService struct {
	db *gorm.DB
}

func NewImportService(db *gorm.DB) *ImportService {
	return &ImportService{db: db}
}

// ImportFromFile импортирует данные из JSON файла
func (s *ImportService) ImportFromFile(filePath string) error {
	// 1. Читаем файл
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	// 2. Парсим JSON
	var importData model.ImportData
	if err := json.Unmarshal(data, &importData); err != nil {
		return fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	fmt.Printf("📖 Найдено книг для импорта: %d\n\n", len(importData.Books))

	// 3. Импортируем каждую книгу
	for idx, importBook := range importData.Books {
		fmt.Printf("[%d/%d] Обработка книги: %s\n", idx+1, len(importData.Books), importBook.Name)

		if err := s.importBook(importBook); err != nil {
			return fmt.Errorf("ошибка импорта книги '%s': %w", importBook.Name, err)
		}

		fmt.Printf("  ✅ Книга успешно импортирована\n\n")
	}

	return nil
}

// importBook импортирует одну книгу
func (s *ImportService) importBook(importBook model.ImportBook) error {
	// Начинаем транзакцию
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Обработка паники и откат при ошибке
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Создаём книгу
	bookID := uuid.New()
	book := &model.Book{
		ID:       bookID,
		Title:    importBook.Name,
		BookPath: importBook.BookPath,
	}

	if err := tx.Create(book).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("ошибка создания книги: %w", err)
	}
	fmt.Printf("  📚 Создана книга (ID: %s)\n", bookID)

	// 2. Импортируем вопросы
	questionCount := 0
	optionCount := 0

	for _, importQ := range importBook.Questions {
		questionID := uuid.New()
		question := &model.Question{
			ID:        questionID,
			BookID:    bookID,
			Text:      importQ.Text,
			SortOrder: importQ.ID, // используем ID из JSON как порядок
			Points:    1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := tx.Create(question).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("ошибка создания вопроса: %w", err)
		}
		questionCount++

		// 3. Создаём варианты ответов
		for optIdx, importOpt := range importQ.Options {
			optionID := uuid.New()
			option := &model.Option{
				ID:         optionID,
				QuestionID: questionID,
				Text:       importOpt.Text,
				IsCorrect:  importOpt.Correct,
				SortOrder:  optIdx + 1,
			}

			if err := tx.Create(option).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("ошибка создания варианта ответа: %w", err)
			}
			optionCount++
		}
	}

	// Фиксируем транзакцию
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("ошибка коммита транзакции: %w", err)
	}

	fmt.Printf("  📝 Создано вопросов: %d, вариантов ответов: %d\n", questionCount, optionCount)

	return nil
}

// ImportBooksIfNotExists импортирует книги только если их ещё нет
func (s *ImportService) ImportBooksIfNotExists(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	var importData model.ImportData
	if err := json.Unmarshal(data, &importData); err != nil {
		return fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	for _, importBook := range importData.Books {
		// Проверяем, существует ли книга с таким названием или путём
		var existingBook model.Book
		err := s.db.Where("title = ? OR book_path = ?", importBook.Name, importBook.BookPath).
			First(&existingBook).Error

		if err == nil {
			fmt.Printf("⏭️  Книга '%s' уже существует, пропускаем\n", importBook.Name)
			continue
		}

		if err != gorm.ErrRecordNotFound {
			return err
		}

		// Книги нет - импортируем
		if err := s.importBook(importBook); err != nil {
			return err
		}
	}

	return nil
}

// ClearBooks очищает все книги и связанные данные
func (s *ImportService) ClearBooks() error {
	// Удаляем в правильном порядке (из-за внешних ключей)
	if err := s.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.UserAnswer{}).Error; err != nil {
		return err
	}
	if err := s.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Session{}).Error; err != nil {
		return err
	}
	if err := s.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Option{}).Error; err != nil {
		return err
	}
	if err := s.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Question{}).Error; err != nil {
		return err
	}
	if err := s.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Image{}).Error; err != nil {
		return err
	}
	if err := s.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.BookPage{}).Error; err != nil {
		return err
	}
	if err := s.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Book{}).Error; err != nil {
		return err
	}

	fmt.Println("🗑️  Все книги и связанные данные удалены")
	return nil
}
