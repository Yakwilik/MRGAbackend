package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	"github.com/blockloop/scan"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"time"
)

type Interface interface {
	CreateSession(ctx context.Context, user model.User) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
	GetEmailBySession(ctx context.Context, sessionID string) (string, error)
	CreateUser(ctx context.Context, user model.User) error
	CreateConversation(ctx context.Context, userEmail string) (uint32, error)
	CreateMessage(ctx context.Context, data model.CreateMessageData) error
	GetConversations(ctx context.Context, userEmail string) ([]model.ConversationData, error)
	GetConversation(ctx context.Context, chatID uint32) ([]model.Message, error)
	CheckCredentials(ctx context.Context, user model.User) error
	SetChatName(ctx context.Context, chatID uint32, name string) error
}

type storage struct {
	db *sql.DB
}

func New(db *sql.DB) Interface {
	return &storage{db: db}
}

const (
	UNIQUE_VIOLATION      = "unique_violation"
	FOREIGN_KEY_VIOLATION = "foreign_key_violation"
	NOT_NULL_VIOLATION    = "not_null_violation"
)

func (s *storage) CreateSession(ctx context.Context, user model.User) (string, error) {
	sessionID := uuid.NewString()
	_, err := s.db.Exec("INSERT INTO session (session_id, email) values ($1, $2)", sessionID, user.Email)
	if err != nil {
		if pqErr := new(pq.Error); errors.As(err, &pqErr) {
			if pqErr.Code.Name() == FOREIGN_KEY_VIOLATION {
				return "", model.ErrBadCredentials
			}
		}
		return "", fmt.Errorf("error executing query [CreateSession]: %w", err)
	}

	return sessionID, nil
}

func (s *storage) GetEmailBySession(ctx context.Context, sessionID string) (string, error) {
	row := s.db.QueryRow("SELECT email from session where session_id = $1", sessionID)
	if err := row.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", model.ErrNotFound
		}
		return "", fmt.Errorf("error executing query [GetEmailBySession]: %w", err)
	}

	var email string
	if err := row.Scan(&email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", model.ErrNotFound
		}
		return "", fmt.Errorf("error executing query [GetEmailBySession]: %w", err)
	}

	return email, nil
}

func (s *storage) CreateUser(ctx context.Context, user model.User) error {
	_, err := s.db.Exec("INSERT INTO users (email, password_hash) values ($1, $2)", user.Email, user.Password)

	if err != nil {
		if pqErr := new(pq.Error); errors.As(err, &pqErr) {
			if pqErr.Code.Name() == UNIQUE_VIOLATION {
				return model.ErrAlreadyExists
			}
		}
		return fmt.Errorf("error executing query [CreateUser]: %w", err)
	}

	return nil
}

func (s *storage) CreateConversation(ctx context.Context, userEmail string) (uint32, error) {
	var insertedID uint32
	if err := s.db.QueryRow("insert into chats (email) values ($1) returning chat_id;", userEmail).
		Scan(&insertedID); err != nil {
		return 0, fmt.Errorf("error executing query [CreateConversation]: %w", err)
	}

	return insertedID, nil
}

func (s *storage) CreateMessage(ctx context.Context, data model.CreateMessageData) error {
	_, err := s.db.Exec("insert into messages (chat_id, sent_at, message, from_bot) VALUES ($1, $2, $3, $4)", data.ChatID, data.SentAt, data.Message, data.FromBot)
	if err != nil {
		return fmt.Errorf("error executing query [CreateMessage]: %w", err)
	}

	return nil
}

type conversationData struct {
	ChatName    string    `db:"chat_name"`
	LastMessage string    `db:"last_message"`
	FromChatBot bool      `db:"from_bot"`
	SentAt      time.Time `db:"sent_at"`
	ChatID      uint32    `db:"chat_id"`
}

func (s *storage) GetConversations(ctx context.Context, userEmail string) ([]model.ConversationData, error) {
	rows, err := s.db.Query(`
SELECT DISTINCT c.chat_id,
                c.chat_name as chat_name,
                m.message AS last_message,
                m.sent_at AS sent_at,
				m.from_bot AS from_bot
FROM chats c
         INNER JOIN
     (SELECT chat_id,
             message,
             sent_at,
             from_bot
      FROM messages
      WHERE (chat_id, sent_at) IN (SELECT chat_id, MAX(sent_at) AS sent_at
                                   FROM messages
                                   GROUP BY chat_id)) m ON c.chat_id = m.chat_id
WHERE c.email = $1
ORDER BY m.sent_at DESC;`, userEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.ConversationData{}, nil
		}
		return nil, fmt.Errorf("error executing query [GetConversations]: %w", err)
	}
	var result []conversationData
	err = scan.Rows(&result, rows)
	if err != nil {
		return nil, fmt.Errorf("error executing query [GetConversations]: %w", err)
	}

	return decodeConversations(result), nil
}

func decodeConversations(dbData []conversationData) []model.ConversationData {
	result := make([]model.ConversationData, 0, len(dbData))
	for _, dbModel := range dbData {
		result = append(result, model.ConversationData{
			ChatName:    dbModel.ChatName,
			LastMessage: dbModel.LastMessage,
			FromChatBot: dbModel.FromChatBot,
			SentAt:      dbModel.SentAt,
			ChatID:      dbModel.ChatID,
		})
	}

	return result
}

type message struct {
	Message     string    `db:"message"`
	FromChatBot bool      `db:"from_bot"`
	SentAt      time.Time `db:"sent_at"`
}

func (s *storage) GetConversation(ctx context.Context, chatID uint32) ([]model.Message, error) {
	rows, err := s.db.Query(`
SELECT message,
       sent_at,
       from_bot
FROM messages
WHERE chat_id = $1
ORDER BY sent_at ASC;`, chatID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.Message{}, nil
		}
		return nil, fmt.Errorf("error executing query [GetConversation]: %w", err)
	}

	var result []message
	err = scan.Rows(&result, rows)
	if err != nil {
		return nil, fmt.Errorf("error executing query [GetConversation]: %w", err)
	}

	return decodeMessages(result), nil
}

func decodeMessages(dbMessages []message) []model.Message {
	result := make([]model.Message, 0, len(dbMessages))
	for _, dbModel := range dbMessages {
		result = append(result, model.Message{
			Message:     dbModel.Message,
			FromChatBot: dbModel.FromChatBot,
			SentAt:      dbModel.SentAt,
		})
	}

	return result
}

func (s *storage) CheckCredentials(ctx context.Context, user model.User) error {
	exists := false
	if err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND password_hash = $2)", user.Email, user.Password).Scan(&exists); err != nil {
		return fmt.Errorf("error executing query [GetEmailBySession]: %w", err)
	}

	if !exists {
		return model.ErrBadCredentials
	}

	return nil
}

func (s *storage) DeleteSession(ctx context.Context, sessionID string) error {
	if _, err := s.db.Exec("DELETE FROM session WHERE session_id = $1", sessionID); err != nil {
		return fmt.Errorf("error executing query [DeleteSession]: %w", err)
	}

	return nil
}

func (s *storage) SetChatName(ctx context.Context, chatID uint32, name string) error {
	if _, err := s.db.Exec("UPDATE chats SET chat_name = $1 WHERE chat_id = $2", name, chatID); err != nil {
		return fmt.Errorf("error executing query [SetChatName]: %w", err)
	}

	return nil
}
