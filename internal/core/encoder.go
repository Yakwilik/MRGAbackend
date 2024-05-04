package core

import "github.com/Yakwilik/MRGAbackend/internal/model"

func encodeToChatHistory(messages []model.Message) []model.HistoryMessage {
	lastMessageID := len(messages) - 1
	historyMessages := make([]model.HistoryMessage, 0, len(messages)-1)
	for index, message := range messages {
		if index == lastMessageID {
			continue
		} else {
			historyMessages = append(historyMessages, model.HistoryMessage{
				Role: model.GetRole(message.FromChatBot),
				Text: message.Message,
			})
		}
	}

	return historyMessages
}
