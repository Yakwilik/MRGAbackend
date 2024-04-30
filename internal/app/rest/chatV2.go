package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/helper"
	"net/http"
	"strings"
	"time"
)

type newMessage struct {
	Message string    `json:"message"`
	SentAt  time.Time `json:"sent_at"`
}

type sendMessageRequest struct {
	ChatID  uint32     `json:"chat_id"`
	Message newMessage `json:"message"`
}

func (receiver *Handler) sendMsgV2(w http.ResponseWriter, r *http.Request) {
	email, err := helper.GetEmailFromContext(r.Context())
	logger.Info(r.Context(), "sendMsgV2", "email", email, "err", err)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var request sendMessageRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		logger.Error(r.Context(), "error", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	if !request.Message.SentAt.IsZero() {
		now = request.Message.SentAt
	}

	if err = receiver.useCase.SendMessage(r.Context(), model.CreateMessageData{
		ChatID:  request.ChatID,
		SentAt:  now,
		FromBot: false,
		Message: request.Message.Message,
	}); err != nil {
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			http.Error(w, errValidation.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	history, err := receiver.useCase.GetConversation(r.Context(), request.ChatID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := model.ChatRequest{
		ChatID:      request.ChatID,
		UserQuery:   request.Message.Message,
		ChatHistory: encodeToChatHistory(history),
	}

	logger.Info(r.Context(), "request", "body", data, "handler", "sendMsgV2")

	responseChan, err := receiver.aiBotService.RespondToUserQuery(r.Context(), data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	result := strings.Builder{}

	for data := range responseChan {
		if data.MessageStatus == model.StatusOk {
			result.WriteString(data.Chunk)
		}

		bytes, marshalErr := json.Marshal(data)
		if marshalErr != nil {
			fmt.Fprintf(w, "data: %s\n\n", `{"error": "error encoding json"}`)
			flusher.Flush()
			continue
		}
		if _, writeErr := fmt.Fprintf(w, "data: %s\n\n", bytes); writeErr != nil {
			http.Error(w, writeErr.Error(), http.StatusInternalServerError)
			return
		}
		flusher.Flush()
	}
	receiver.useCase.SendMessage(r.Context(), model.CreateMessageData{
		ChatID:  request.ChatID,
		SentAt:  time.Now(),
		FromBot: true,
		Message: result.String(),
	})

	logger.Info(r.Context(), "request", "body", result.String(), "handler", "sendMsgV2")
	fmt.Fprint(w, "event: close\n\n")
	flusher.Flush()
	w.Header().Set("Connection", "close")
}

func encodeToChatHistory(messages []model.Message) []model.HistoryMessage {
	lastMessageID := len(messages) - 1
	historyMessages := make([]model.HistoryMessage, 0, len(messages)-1)
	for index, message := range messages {
		if index == lastMessageID {
			continue
		} else {
			historyMessages = append(historyMessages, model.HistoryMessage{
				Role: getRole(message.FromChatBot),
				Text: message.Message,
			})
		}
	}

	return historyMessages
}

func getRole(fromChatBot bool) model.Role {
	if fromChatBot {
		return model.RoleAssistant
	}
	return model.RoleUser
}

type beginConversationRequest struct {
	Message string `json:"message"`
}

func (receiver *Handler) beginConversationV2(w http.ResponseWriter, r *http.Request) {
	email, err := helper.GetEmailFromContext(r.Context())
	logger.Info(r.Context(), "sendMsgV2", "email", email, "err", err)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var request beginConversationRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		logger.Error(r.Context(), "error", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	respCh, err := receiver.useCase.BeginConversationV2(r.Context(), email, request.Message)
	if err != nil {
		logger.Error(r.Context(), "error", "err", err)
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			http.Error(w, errValidation.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for data := range respCh {
		dataBytes, marshalErr := json.Marshal(data)
		if marshalErr != nil {
			fmt.Fprintf(w, "data: %s\n\n", `{"error": "error encoding json"}`)
			flusher.Flush()
			continue
		}
		if _, writeErr := fmt.Fprintf(w, "data: %s\n\n", dataBytes); writeErr != nil {
			http.Error(w, writeErr.Error(), http.StatusInternalServerError)
			return
		}
		flusher.Flush()
	}

	fmt.Fprint(w, "event: close\n\n")
	flusher.Flush()
	w.Header().Set("Connection", "close")

}
