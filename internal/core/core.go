package core

import (
	"context"
	"errors"
	"github.com/Yakwilik/MRGAbackend/internal/client/ai_botV2"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	storagePkg "github.com/Yakwilik/MRGAbackend/internal/storage"
	"strings"
	"time"
)

type UseCase interface {
	SignUp(ctx context.Context, user model.User) error
	Login(ctx context.Context, user model.User) (string, error)
	CheckLogin(ctx context.Context, sessionID string) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
	BeginConversation(ctx context.Context, userEmail string) (uint32, error)
	BeginConversationV2(ctx context.Context, userEmail string, userMessage string) (<-chan *model.ChatResponseChunk, error)
	SendMessage(ctx context.Context, data model.CreateMessageData) error
	GetConversations(ctx context.Context, userEmail string) ([]model.ConversationData, error)
	GetConversation(ctx context.Context, chatID uint32) ([]model.Message, error)
}

type usecase struct {
	storage storagePkg.Interface
	aiBot   ai_botV2.Interface
}

func New(session storagePkg.Interface, aiBot ai_botV2.Interface) UseCase {
	return &usecase{
		storage: session,
		aiBot:   aiBot,
	}
}

func (a *usecase) SignUp(ctx context.Context, user model.User) error {
	// TODO проверять/валидировать email, пароль
	err := a.storage.CreateUser(ctx, user)

	if err != nil {
		if errors.Is(err, model.ErrAlreadyExists) {
			return model.NewValidationError("email", "email already exists", "Пользователь с такой почтой уже существует")
		}

		return err
	}
	return nil
}

func (a *usecase) Login(ctx context.Context, user model.User) (string, error) {
	err := a.storage.CheckCredentials(ctx, user)
	if err != nil {
		if errors.Is(err, model.ErrBadCredentials) {
			return "", model.NewValidationError("credentials", "bad login data", "Неверный логин или пароль")
		}
		return "", err
	}

	session, err := a.storage.CreateSession(ctx, user)
	if err != nil {
		if errors.Is(err, model.ErrBadCredentials) {
			return "", model.NewValidationError("credentials", "bad login data", "Неверный логин или пароль")
		}
		return "", err
	}
	return session, nil
}

func (a *usecase) CheckLogin(ctx context.Context, sessionID string) (string, error) {
	email, err := a.storage.GetEmailBySession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return "", model.NewValidationError("sessionID", "no sessionID", "Вы не авторизованы")
		}
		return "", err
	}
	return email, nil
}

func (a *usecase) BeginConversation(ctx context.Context, userEmail string) (uint32, error) {
	return a.storage.CreateConversation(ctx, userEmail)
}

func (a *usecase) SendMessage(ctx context.Context, data model.CreateMessageData) error {
	if strings.TrimSpace(data.Message) == "" {
		return model.NewValidationError("message", "message must not be empy", "Сообщение не может быть пустым")
	}
	return a.storage.CreateMessage(ctx, data)
}

func (a *usecase) GetConversations(ctx context.Context, userEmail string) ([]model.ConversationData, error) {
	return a.storage.GetConversations(ctx, userEmail)
}

func (a *usecase) GetConversation(ctx context.Context, chatID uint32) ([]model.Message, error) {
	return a.storage.GetConversation(ctx, chatID)
}

func (a *usecase) BeginConversationV2(ctx context.Context, userEmail string, userMessage string) (<-chan *model.ChatResponseChunk, error) {
	chatID, err := a.storage.CreateConversation(ctx, userEmail)
	if err != nil {
		return nil, err
	}

	if err := a.SendMessage(ctx, model.CreateMessageData{
		ChatID:  chatID,
		SentAt:  time.Now(),
		FromBot: false,
		Message: userMessage,
	}); err != nil {
		return nil, err
	}

	respChan, err := a.aiBot.RespondToUserQuery(ctx, model.ChatRequest{
		ChatID:      chatID,
		UserQuery:   userMessage,
		ChatHistory: []model.HistoryMessage{},
	})
	if err != nil {
		return nil, err
	}

	resultChan := make(chan *model.ChatResponseChunk)

	go func() {
		result := strings.Builder{}
		defer close(resultChan)
		defer func() {
			a.SendMessage(ctx, model.CreateMessageData{
				ChatID:  chatID,
				SentAt:  time.Now(),
				FromBot: true,
				Message: result.String(),
			})
		}()

		for part := range respChan {
			select {
			case <-ctx.Done():
				return
			case resultChan <- part:
				if part.MessageStatus == model.StatusOk && part.Role == model.RoleAssistant {
					result.WriteString(part.Chunk)
				}
			}
		}

	}()

	return resultChan, nil
}

func (a *usecase) DeleteSession(ctx context.Context, sessionID string) error {
	return a.storage.DeleteSession(ctx, sessionID)
}
