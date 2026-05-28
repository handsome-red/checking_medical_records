// internal/repository/answer_repository.go
package repository

import (
	"context"
	"med_book/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AnswerRepository struct {
	db *gorm.DB
}

func NewAnswerRepository(db *gorm.DB) *AnswerRepository {
	return &AnswerRepository{db: db}
}

// FindBySessionID возвращает все ответы сессии
func (r *AnswerRepository) FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.UserAnswer, error) {
	var answers []model.UserAnswer
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("answered_at ASC").
		Find(&answers).Error
	return answers, err
}

// GetAnsweredQuestionIDs возвращает ID вопросов, на которые уже ответили
func (r *AnswerRepository) GetAnsweredQuestionIDs(ctx context.Context, sessionID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&model.UserAnswer{}).
		Where("session_id = ?", sessionID).
		Pluck("question_id", &ids).Error
	return ids, err
}

// ExistsBySessionAndQuestion проверяет, отвечал ли пользователь на вопрос
func (r *AnswerRepository) ExistsBySessionAndQuestion(ctx context.Context, sessionID uuid.UUID, questionID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserAnswer{}).
		Where("session_id = ? AND question_id = ?", sessionID, questionID).
		Count(&count).Error
	return count > 0, err
}

// CountBySessionID возвращает количество ответов в сессии
func (r *AnswerRepository) CountBySessionID(ctx context.Context, sessionID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserAnswer{}).
		Where("session_id = ?", sessionID).
		Count(&count).Error
	return int(count), err
}
