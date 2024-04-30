package rest

import (
	"encoding/json"
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/helper"
	"net/http"
	"time"
)

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
