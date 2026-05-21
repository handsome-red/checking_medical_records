// internal/service/question_service.go
package service

import (
	"med_book/internal/database"
	"med_book/internal/model"
)

type QuestionService struct {
	db *database.Database
}

func NewQuestionService(db *database.Database) *QuestionService {
	return &QuestionService{db: db}
}

func (s *QuestionService) GetByID(id int) (*model.Question, error) {
	var question model.Question
	err := s.db.GetDB().
		Preload("Options").
		First(&question, id).Error
	return &question, err
}

func (s *QuestionService) GetMultiple(ids []int) ([]model.Question, error) {
	var questions []model.Question
	err := s.db.GetDB().
		Preload("Options").
		Where("id IN ?", ids).
		Find(&questions).Error
	return questions, err
}

func (s *QuestionService) GetCorrectAnswers(questionID int) ([]int, error) {
	var options []model.Option
	err := s.db.GetDB().
		Where("question_id = ? AND correct = ?", questionID, true).
		Find(&options).Error
	if err != nil {
		return nil, err
	}

	indexes := make([]int, len(options))
	for i, _ := range options {
		indexes[i] = i
	}
	return indexes, nil
}

// GetAllQuestions возвращает количество всех вопросов
func (s *QuestionService) GetQuestions() ([]model.Question, error) {
	var questions []model.Question
	result := s.db.GetDB().
		Model(&model.Question{}).
		Find(questions)

	if result.Error != nil {
		return nil, result.Error
	}

	return questions, nil
}
