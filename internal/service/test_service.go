package service

import (
	"encoding/json"
	"fmt"
	"med_book/internal/model"
	"os"
	"strings"
)

type TestService struct {
	test *model.Test
}

func NewTestService(path string) (*TestService, error) {
	if path == "" {
		path = "pkg/questions/questions.json"
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	var loader model.Test
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&loader)
	if err != nil {
		return nil, fmt.Errorf("cannot decode file: %w", err)
	}
	fmt.Printf("Найдено в файле %d книг\n", len(loader.Books))

	return &TestService{test: &loader}, nil
}

func (ts TestService) GetQuestion(question_id int) (*model.Question, error) {
	for _, book := range ts.test.Books {
		for _, question := range book.Questions {
			if question.ID == question_id {
				return &question, nil
			}
		}
	}

	return &model.Question{}, fmt.Errorf("question with id %d not found", question_id)
}

// func (ts TestService) GetQuestionFromBook(book_id, question_id int) (model.Question, error) {
// 	for _, book := range ts.test.Books {
// 		if book.ID == book_id {
// 			for _, question := range book.Questions {
// 				if question.ID == question_id {
// 					return question, nil
// 				}
// 			}
// 			return model.Question{}, fmt.Errorf("question with id %d not found", question_id)
// 		}
// 	}

// 	return model.Question{}, fmt.Errorf("book with id %d not found", book_id)
// }

func (ts *TestService) GetBook(bookID int) (*model.Book, error) {
	for _, book := range ts.test.Books {
		if book.ID == bookID {
			return &book, nil
		}
	}

	return nil, fmt.Errorf("book with id %d not found", bookID)
}

func (ts *TestService) GetBooks() []model.Book {
	return ts.test.Books
}

func (ts TestService) NumOfQuestioinsInTheBook(bookID int) (int, error) {
	for _, book := range ts.test.Books {
		if book.ID == bookID {
			return len(book.Questions), nil
		}
	}
	return 0, fmt.Errorf("book with id %d not found", bookID)
}

func (ts *TestService) GetPagesByBookID(bookID int) ([]model.Page, error) {
	for _, book := range ts.test.Books {
		if book.ID == bookID {
			dirPath := strings.TrimPrefix(book.BookPath, "/")

			files, err := os.ReadDir(dirPath)
			if err != nil {
				return nil, fmt.Errorf("cannot read directory %s: %w", dirPath, err)
			}

			pages := make([]model.Page, 0)
			pageNum := 1
			for _, file := range files {
				// Берём только jpg файлы
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".jpg") {
					pages = append(pages, model.Page{
						ID:     pageNum,
						BookID: bookID,
						Number: pageNum,
						Path:   fmt.Sprintf("%s/%s", book.BookPath, file.Name()),
					})
					pageNum++
				}
			}

			if len(pages) == 0 {
				return nil, fmt.Errorf("no jpg files found in %s", dirPath)
			}

			return pages, nil
		}
	}
	return nil, fmt.Errorf("book with id %d not found", bookID)
}
