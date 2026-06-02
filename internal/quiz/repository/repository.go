package repository

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"med_book/internal/quiz/models"
)

type SessionStoreInterface interface {
	Save(session *models.Session) error
	Get(sessionID string) (*models.Session, error)
}

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal=WAL")
	if err != nil {
		return nil, err
	}

	// Создание таблицы (обновлённая схема)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			participant TEXT,
			question_ids TEXT,
			status TEXT,
			answers TEXT,
			current_index INTEGER,
			started_at DATETIME,
			completed_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	store := &SQLiteStore{db: db}
	go store.cleanupExpired()

	return store, nil
}

func (s *SQLiteStore) Save(session *models.Session) error {
	// Сериализация данных
	participantJSON, _ := json.Marshal(session.Participant)
	questionIDsJSON, _ := json.Marshal(session.QuestionIDs)
	answersJSON, _ := json.Marshal(session.GetAllAnswers())

	var completedAt interface{}
	if session.CompletedAt != nil {
		completedAt = *session.CompletedAt
	}

	// UPSERT
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO sessions (
			id, user_id, participant, question_ids, status, answers,
			current_index, started_at, completed_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`,
		session.ID,
		session.UserID,
		string(participantJSON),
		string(questionIDsJSON),
		session.Status,
		string(answersJSON),
		session.CurrentIndex,
		session.StartedAt,
		completedAt,
	)
	return err
}

func (s *SQLiteStore) Get(sessionID string) (*models.Session, error) {
	var session models.Session
	var participantJSON, questionIDsJSON, answersJSON string
	var completedAt sql.NullTime

	err := s.db.QueryRow(`
		SELECT 
			id, user_id, participant, question_ids, status, answers,
			current_index, started_at, completed_at
		FROM sessions 
		WHERE id = ?
	`, sessionID).Scan(
		&session.ID,
		&session.UserID,
		&participantJSON,
		&questionIDsJSON,
		&session.Status,
		&answersJSON,
		&session.CurrentIndex,
		&session.StartedAt,
		&completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Десериализация
	if err := json.Unmarshal([]byte(participantJSON), &session.Participant); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(questionIDsJSON), &session.QuestionIDs); err != nil {
		return nil, err
	}

	var answersMap map[int][]int
	if err := json.Unmarshal([]byte(answersJSON), &answersMap); err != nil {
		return nil, err
	}
	session.RestoreAnswers(answersMap)

	if completedAt.Valid {
		session.CompletedAt = &completedAt.Time
	}

	return &session, nil
}

// cleanupExpired удаляет просроченные сессии (где время истекло)
func (s *SQLiteStore) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		// Удаляем сессии, у которых started_at + 1 час < now()
		result, err := s.db.Exec(`
			DELETE FROM sessions 
			WHERE status = 'active' 
			AND datetime(started_at, '+1 hour') < datetime('now')
		`)
		if err != nil {
			log.Printf("Cleanup error: %v", err)
			continue
		}
		if rows, _ := result.RowsAffected(); rows > 0 {
			log.Printf("Cleaned up %d expired sessions", rows)
		}
	}
}

// Close закрывает соединение с БД
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
