package rest

import (
	"github.com/Yakwilik/MRGAbackend/internal/client/ai_botV2"
	"github.com/Yakwilik/MRGAbackend/internal/core"
	"net/http"
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
	receiver.mux.HandleFunc("POST /v2/chats/begin", receiver.beginConversationV2)
	return receiver.mux
}
