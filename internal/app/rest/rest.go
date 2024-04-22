package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	index "github.com/Yakwilik/MRGAbackend/internal/app/templ"
	"github.com/Yakwilik/MRGAbackend/internal/client/ai_botV2"
	"github.com/Yakwilik/MRGAbackend/internal/core"
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	pb "github.com/Yakwilik/MRGAbackend/internal/pb/ailawyer"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/helper"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	mux          *http.ServeMux
	useCase      core.UseCase
	aiBotService ai_botV2.Interface
}

func New(useCase core.UseCase, aiBot ai_botV2.Interface) *Handler {
	return &Handler{mux: http.NewServeMux(), useCase: useCase, aiBotService: aiBot}
}

func (receiver *Handler) Init() *http.ServeMux {
	receiver.mux.HandleFunc("GET /document", receiver.document)
	receiver.mux.HandleFunc("POST /connect", receiver.centrifugoConnect)
	receiver.mux.HandleFunc("POST /v2/chats/send", receiver.sendMsgV2)
	return receiver.mux
}

func (receiver *Handler) document(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	sb := &strings.Builder{}
	index.Index().Render(r.Context(), sb)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.Document{
		FieldsBlocks: nil,
	})

}

type connectRequest struct {
	Client    string         `json:"client"`
	Transport string         `json:"transport"`
	Protocol  string         `json:"protocol"`
	Encoding  string         `json:"encoding"`
	Name      string         `json:"name"`
	Version   string         `json:"version"`
	Data      map[string]any `json:"data"`
	B64data   string         `json:"b64data"`
	Channels  []string       `json:"channels"`
}

type connectResponse struct {
	Result connectResponseData `json:"result"`
}

type connectResponseData struct {
	User     string                    `json:"user"`
	ExpireAt int64                     `json:"expire_at"`
	Info     map[string]any            `json:"info"`
	B64info  string                    `json:"b64info"`
	Data     map[string]any            `json:"data"`
	B64data  string                    `json:"b64data"`
	Channels []string                  `json:"channels"`
	Subs     map[string]map[string]any `json:"subs"`
	Meta     map[string]any            `json:"meta"`
}

func (receiver *Handler) centrifugoConnect(w http.ResponseWriter, r *http.Request) {
	email, err := helper.GetEmailFromContext(r.Context())
	logger.Info(r.Context(), "centrifugoConnect", "email", email, "err", err)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("{\n  \"disconnect\": {\n    \"code\": 4501,\n    \"reason\": \"unauthorized\"\n  }\n}"))
		return
	}
	req := connectRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(r.Context(), "error %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\n  \"disconnect\": {\n    \"code\": 4501,\n    \"reason\": \"unauthorized\"\n  }\n}"))
		return
	}
	bytes, _ := json.Marshal(req)
	logger.Info(r.Context(), "request", "body", string(bytes), "handler", "centrifugoConnect")

	json.NewEncoder(w).Encode(connectResponse{Result: connectResponseData{
		User:     email,
		ExpireAt: time.Now().Add(time.Hour * 24).Unix(),
		Channels: []string{email},
		Subs:     map[string]map[string]any{email: {"allow_publish_for_client": true}},
	}})
}

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

var protoRole = map[model.Role]pb.MessageRole{
	model.RoleUser:      pb.MessageRole_USER,
	model.RoleAssistant: pb.MessageRole_ASSISTANT,
	model.RoleTechnical: pb.MessageRole_TECHNICAL,
}

var protoStatus = map[model.MessageStatus]pb.MessageStatus{
	model.StatusOk:   pb.MessageStatus_OK,
	model.StatusFail: pb.MessageStatus_ERROR,
}
