package rest

import (
	"github.com/Yakwilik/MRGAbackend/internal/client/ai_botV2"
	"github.com/Yakwilik/MRGAbackend/internal/config"
	"github.com/Yakwilik/MRGAbackend/internal/core"
	"net/http"
)

type Handler struct {
	mux          *http.ServeMux
	useCase      core.UseCase
	aiBotService ai_botV2.Interface

	cfg *config.Config
}

func New(cfg *config.Config, useCase core.UseCase, aiBot ai_botV2.Interface) *Handler {
	return &Handler{
		mux:          http.NewServeMux(),
		useCase:      useCase,
		aiBotService: aiBot,
		cfg:          cfg,
	}
}

func (receiver *Handler) Init() *http.ServeMux {
	receiver.mux.HandleFunc("GET /document", receiver.document)
	receiver.mux.HandleFunc("POST /connect", receiver.centrifugoConnect)
	receiver.mux.HandleFunc("POST /v2/chats/send", receiver.sendMsgV2)
	receiver.mux.HandleFunc("POST /v2/chats/begin", receiver.beginConversationV2)
	receiver.mux.HandleFunc("POST /v2/chats/retry", receiver.retryLastMessage)
	receiver.mux.HandleFunc("POST /v1/auth/logout", receiver.logout)
	return receiver.mux
}
