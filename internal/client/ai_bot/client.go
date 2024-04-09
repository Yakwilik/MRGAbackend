package ai_bot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	"log"
	"net/http"
	"strings"
)

type client interface {
	SendPromptWithContext(ctx context.Context, prompt promptModel) (response, error)
}

type promptModel struct {
	UserID      string           `json:"user_id"`
	ChatID      string           `json:"chat_id"`
	UserQuery   string           `json:"user_query"`
	ChatHistory []historyMessage `json:"chat_history"`
}
type historyMessage struct {
	Role string `json:"role"`
	Text string `json:"text"`
}
type response struct {
	UserId string `json:"user_id"`
	ChatId string `json:"chat_id"`
	Output string `json:"output"`
}

type clientImpl struct {
	serverHost string
	client     *http.Client
}

func newClient(serverHost string) client {
	return &clientImpl{serverHost: serverHost, client: &http.Client{}}
}

func (c *clientImpl) SendPromptWithContext(ctx context.Context, prompt promptModel) (response, error) {
	bodyBytes, err := json.Marshal(prompt)
	log.Println(string(bodyBytes))
	if err != nil {
		return response{}, model.WrapErrorWithMethodName(err, "SendPromptWithContext")
	}
	resp, err := c.client.Post(fmt.Sprintf("%s/query", c.serverHost), "application/json", strings.NewReader(string(bodyBytes)))
	if err != nil {
		return response{}, model.WrapErrorWithMethodName(err, "SendPromptWithContext")
	}

	decodedResp := response{}
	err = json.NewDecoder(resp.Body).Decode(&decodedResp)
	if err != nil {
		return response{}, model.WrapErrorWithMethodName(err, "SendPromptWithContext")
	}

	return decodedResp, nil
}
