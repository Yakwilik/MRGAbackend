package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/helper"
	"net/http"
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

	if request.Message.SentAt.IsZero() {
		request.Message.SentAt = time.Now()
	}

	responseChan, err := receiver.useCase.SendMessageV2(r.Context(), model.CreateMessageData{
		ChatID:  request.ChatID,
		SentAt:  request.Message.SentAt,
		Role:    model.RoleUser,
		FromBot: false,
		Message: request.Message.Message,
	})
	if err != nil {
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			http.Error(w, errValidation.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	streamResponse(w, flusher, responseChan)
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

	respCh, err := receiver.useCase.BeginConversationV2(context.WithoutCancel(r.Context()), email, request.Message)
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

	streamResponse(w, flusher, respCh)
}

func streamResponse[T any](w http.ResponseWriter, flusher http.Flusher, respChan <-chan T) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	defer func() {
		fmt.Fprint(w, "event: close\n\n")
		flusher.Flush()
		w.Header().Set("Connection", "close")
	}()

	for data := range respChan {
		dataBytes, marshalErr := json.Marshal(data)
		if marshalErr != nil {
			fmt.Fprintf(w, "data: %s\n\n", `{"error": "error encoding json"}`)
			flusher.Flush()
			continue
		}
		if _, writeErr := fmt.Fprintf(w, "data: %s\n\n", dataBytes); writeErr != nil {
			return
		}
		flusher.Flush()

	}

}

type retryLastMessageRequest struct {
	ChatID uint32 `json:"chat_id"`
}

func (receiver *Handler) retryLastMessage(w http.ResponseWriter, r *http.Request) {
	email, err := helper.GetEmailFromContext(r.Context())
	logger.Info(r.Context(), "sendMsgV2", "email", email, "err", err)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var request retryLastMessageRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		logger.Error(r.Context(), "error", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	respCh, err := receiver.useCase.RetryLastMessage(context.WithoutCancel(r.Context()), request.ChatID)
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

	streamResponse(w, flusher, respCh)
}
