package rest

import (
	"encoding/json"
	"fmt"
	index "github.com/Yakwilik/MRGAbackend/internal/app/templ"
	"github.com/Yakwilik/MRGAbackend/internal/core"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/helper"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	mux     *http.ServeMux
	useCase core.UseCase
}

func New(useCase core.UseCase) *Handler {
	return &Handler{mux: http.NewServeMux(), useCase: useCase}
}

func (receiver *Handler) Init() *http.ServeMux {
	receiver.mux.HandleFunc("GET /document", receiver.document)
	receiver.mux.HandleFunc("POST /connect", receiver.centrifugoConnect)
	return receiver.mux
}

func (receiver *Handler) document(w http.ResponseWriter, r *http.Request) {

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
	slog.Info(fmt.Sprintf("email: %s, err: %v", email, err))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("{\n  \"disconnect\": {\n    \"code\": 4501,\n    \"reason\": \"unauthorized\"\n  }\n}"))
		return
	}
	req := connectRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("error %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\n  \"disconnect\": {\n    \"code\": 4501,\n    \"reason\": \"unauthorized\"\n  }\n}"))
		return
	}
	bytes, _ := json.Marshal(req)
	slog.Info(string(bytes))

	json.NewEncoder(w).Encode(connectResponse{Result: connectResponseData{
		User:     email,
		ExpireAt: time.Now().Add(time.Hour * 24).Unix(),
		Channels: []string{email},
	}})
}
