package ai_bot

import (
	"context"
	"github.com/Yakwilik/MRGAbackend/internal/core"
	"github.com/Yakwilik/MRGAbackend/internal/model"
)

type Interface interface {
	GetPromptAnswer(ctx context.Context, prompt string) (string, error)
	GetPromptAnswerWithChatHistory(ctx context.Context, chatID uint32) (string, error)
}

type adapter struct {
	client  client
	useCase core.UseCase
}

func New(useCase core.UseCase, serverHost string) Interface {
	return &adapter{useCase: useCase, client: newClient(serverHost)}
}

func (a *adapter) GetPromptAnswer(ctx context.Context, prompt string) (string, error) {
	resp, err := a.client.SendPromptWithContext(ctx, promptModel{
		UserID:      0,
		ChatID:      0,
		UserQuery:   prompt,
		ChatHistory: []historyMessage{},
	})
	if err != nil {
		return "", err
	}
	return resp.Output, nil
}

func (a *adapter) GetPromptAnswerWithChatHistory(ctx context.Context, chatID uint32) (string, error) {
	messages, err := a.useCase.GetConversation(ctx, chatID)
	if err != nil {
		return "", err
	}
	prompt, history := encodeMessagesToPromptAndHistory(messages)
	resp, err := a.client.SendPromptWithContext(ctx, promptModel{
		UserID:      0,
		ChatID:      chatID,
		UserQuery:   prompt,
		ChatHistory: history,
	})
	if err != nil {
		return "", err
	}

	return resp.Output, nil
}

func encodeMessagesToPromptAndHistory(messages []model.Message) (string, []historyMessage) {
	lastMessageID := len(messages) - 1
	historyMessages := make([]historyMessage, 0, len(messages)-1)
	prompt := ""
	for index, message := range messages {
		if index == lastMessageID {
			prompt = message.Message
		} else {
			historyMessages = append(historyMessages, historyMessage{
				Role: getRole(message.FromChatBot),
				Text: message.Message,
			})
		}
	}

	return prompt, historyMessages
}

func getRole(fromBot bool) string {
	if fromBot {
		return "assistant"
	}
	return "user"
}
