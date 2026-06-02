package models

import (
	"errors"
	"time"
)

const (
	constDuration time.Duration = 1 * time.Hour
)

// OptionBackend — вариант ответа (для бэкенда)
type OptionBackend struct {
	ID        int
	Text      string
	IsCorrect bool
}

// OptionFrontend — вариант ответа (для фронтенда)
type OptionFrontend struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

// QuestionBackend — вопрос (для бэкенда)
type QuestionBackend struct {
	ID      int
	Text    string
	Options []OptionBackend
}

// QuestionFrontend — вопрос (для фронтенда)
type QuestionFrontend struct {
	ID      int              `json:"id"`
	Text    string           `json:"text"`
	Options []OptionFrontend `json:"options"`
}

// ToFrontend конвертирует вопрос для отправки на фронт
func (q *QuestionBackend) ToFrontend() *QuestionFrontend {
	opts := make([]OptionFrontend, len(q.Options))
	for i, opt := range q.Options {
		opts[i] = OptionFrontend{ID: opt.ID, Text: opt.Text}
	}
	return &QuestionFrontend{
		ID:      q.ID,
		Text:    q.Text,
		Options: opts,
	}
}

// Participant — данные участника
type Participant struct {
	LastName    string `json:"last_name" validate:"required,min=2"`
	FirstName   string `json:"first_name" validate:"required,min=2"`
	MiddleName  string `json:"middle_name"`
	District    string `json:"district" validate:"required"`
	City        string `json:"city" validate:"required"`
	School      string `json:"school" validate:"required"`
	Grade       int    `json:"grade" validate:"min=1,max=11"`
	VideoAgreed bool   `json:"video_agreed" validate:"required"`
}

// Validate проверяет корректность данных участника
func (p *Participant) Validate() error {
	if p.LastName == "" {
		return errors.New("last name is required")
	}
	if p.FirstName == "" {
		return errors.New("first name is required")
	}
	if p.District == "" {
		return errors.New("district is required")
	}
	if p.City == "" {
		return errors.New("city is required")
	}
	if p.School == "" {
		return errors.New("school is required")
	}
	if p.Grade < 1 || p.Grade > 11 {
		return errors.New("grade must be between 1 and 11")
	}
	if !p.VideoAgreed {
		return errors.New("video agreement is required")
	}
	return nil
}

// SessionStatus — статус сессии
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusAbandoned SessionStatus = "abandoned"
)

// Session — данные сессии пользователя
type Session struct {
	ID           string        `json:"id"`
	UserID       string        `json:"user_id"` // для связи с User
	Participant  Participant   `json:"participant"`
	QuestionIDs  []int         `json:"question_ids"` // порядок вопросов (ID)
	Status       SessionStatus `json:"status"`
	answers      map[int][]int // приватное поле (questionID -> []optionID)
	CurrentIndex int           `json:"current_index"`
	StartedAt    time.Time     `json:"started_at"`
	CompletedAt  *time.Time    `json:"completed_at,omitempty"`
}

// NewSession создаёт новую сессию
func NewSession(id, userID string, participant Participant, questionIDs []int) *Session {
	now := time.Now()
	return &Session{
		ID:           id,
		UserID:       userID,
		Participant:  participant,
		QuestionIDs:  questionIDs,
		Status:       SessionStatusActive,
		answers:      make(map[int][]int),
		CurrentIndex: 0,
		StartedAt:    now,
	}
}

// AddAnswer добавляет ответ на вопрос
func (s *Session) AddAnswer(questionID, optionID int) {
	if s.answers == nil {
		s.answers = make(map[int][]int)
	}

	// Проверяем, не выбран ли уже этот вариант
	for _, id := range s.answers[questionID] {
		if id == optionID {
			return // уже выбран
		}
	}

	s.answers[questionID] = append(s.answers[questionID], optionID)
}

// RemoveAnswer удаляет ответ на вопрос
func (s *Session) RemoveAnswer(questionID, optionID int) {
	if s.answers == nil {
		return
	}

	options := s.answers[questionID]
	for i, id := range options {
		if id == optionID {
			s.answers[questionID] = append(options[:i], options[i+1:]...)
			return
		}
	}
}

// GetAnswers возвращает ответы на вопрос (копия)
func (s *Session) GetAnswers(questionID int) []int {
	if s.answers == nil {
		return []int{}
	}

	answers := s.answers[questionID]
	if answers == nil {
		return []int{}
	}

	// Возвращаем копию
	result := make([]int, len(answers))
	copy(result, answers)
	return result
}

// GetAllAnswers возвращает все ответы (копия)
func (s *Session) GetAllAnswers() map[int][]int {
	if s.answers == nil {
		return make(map[int][]int)
	}

	result := make(map[int][]int)
	for qID, opts := range s.answers {
		optsCopy := make([]int, len(opts))
		copy(optsCopy, opts)
		result[qID] = optsCopy
	}
	return result
}

// IsCompleted проверяет, завершена ли сессия
func (s *Session) IsCompleted() bool {
	return s.Status == SessionStatusCompleted
}

// Complete завершает сессию
func (s *Session) Complete() {
	now := time.Now()
	s.Status = SessionStatusCompleted
	s.CompletedAt = &now
}

// IsExpired проверяет, не истекла ли сессия (таймаут 30 минут)
func (s *Session) IsExpired() bool {
	if s.IsCompleted() {
		return false
	}
	return time.Since(s.StartedAt) > constDuration
}

// Progress возвращает прогресс прохождения (0.0 - 1.0)
func (s *Session) Progress() float64 {
	if len(s.QuestionIDs) == 0 {
		return 0
	}
	return float64(s.CurrentIndex) / float64(len(s.QuestionIDs))
}

// CurrentQuestionID возвращает ID текущего вопроса
func (s *Session) CurrentQuestionID() int {
	if s.CurrentIndex >= len(s.QuestionIDs) {
		return 0
	}
	return s.QuestionIDs[s.CurrentIndex]
}

// NextQuestion переходит к следующему вопросу
func (s *Session) NextQuestion() bool {
	if s.CurrentIndex+1 < len(s.QuestionIDs) {
		s.CurrentIndex++
		return true
	}
	return false
}

// PrevQuestion возвращается к предыдущему вопросу
func (s *Session) PrevQuestion() bool {
	if s.CurrentIndex > 0 {
		s.CurrentIndex--
		return true
	}
	return false
}

// RestoreAnswers восстанавливает answers из map (для загрузки из БД)
func (s *Session) RestoreAnswers(answers map[int][]int) {
	if answers == nil {
		s.answers = make(map[int][]int)
		return
	}
	s.answers = make(map[int][]int)
	for qID, opts := range answers {
		optsCopy := make([]int, len(opts))
		copy(optsCopy, opts)
		s.answers[qID] = optsCopy
	}
}

// QuestionsData — корневая структура для загрузки вопросов из JSON
type QuestionsData struct {
	Questions []QuestionBackend `json:"questions"`
}
