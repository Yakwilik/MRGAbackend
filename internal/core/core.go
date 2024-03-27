package core

import (
	"context"
	"errors"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	"github.com/Yakwilik/MRGAbackend/internal/storage"
	"strings"
)

type UseCase interface {
	SignUp(ctx context.Context, user model.User) error
	Login(ctx context.Context, user model.User) (string, error)
	CheckLogin(ctx context.Context, sessionID string) (string, error)
	BeginConversation(ctx context.Context, userEmail string) (uint32, error)
	SendMessage(ctx context.Context, data model.CreateMessageData) error
	GetConversations(ctx context.Context, userEmail string) ([]model.ConversationData, error)
	GetConversation(ctx context.Context, chatID uint32) ([]model.Message, error)
}

type usecase struct {
	storage storage.Interface
}

func New(session storage.Interface) UseCase {
	return &usecase{storage: session}
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
