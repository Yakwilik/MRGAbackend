package ai_bot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	"log"
	"net/http"
	"os"
)

type client interface {
	SendPromptWithContext(ctx context.Context, prompt promptModel) (response, error)
}

type promptModel struct {
	UserID      uint32           `json:"user_id"`
	ChatID      uint32           `json:"chat_id"`
	UserQuery   string           `json:"user_query"`
	ChatHistory []historyMessage `json:"chat_history"`
}
type historyMessage struct {
	Role string `json:"role"`
	Text string `json:"text"`
}
type response struct {
	UserId uint32 `json:"user_id"`
	ChatId uint32 `json:"chat_id"`
	Output string `json:"output"`
}

type clientImpl struct {
	serverHost  string
	clientToken string
	client      *http.Client
}

func newClient(serverHost string) client {
	return &clientImpl{serverHost: serverHost, client: &http.Client{}, clientToken: os.Getenv("X_APP_BOT_AUTH_TOKEN")}
}

func (c *clientImpl) SendPromptWithContext(ctx context.Context, prompt promptModel) (response, error) {
	bodyBytes, err := json.Marshal(prompt)
	log.Println(string(bodyBytes))
	if err != nil {
		return response{}, model.WrapErrorWithMethodName(err, "SendPromptWithContext")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/query", c.serverHost), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return response{}, model.WrapErrorWithMethodName(err, "SendPromptWithContext")
	}

	req.Header.Set("x-app-bot-auth-token", c.clientToken)
	resp, err := c.client.Do(req)
	if err != nil {
		return response{}, model.WrapErrorWithMethodName(err, "SendPromptWithContext")
	}
	//respLog, _ := json.Marshal(resp)

	log.Println("statusCode: ", resp.StatusCode, "err:", err)
	decodedResp := response{}
	if resp.StatusCode == http.StatusForbidden {
		return response{}, errors.New("forbidden")
	}
	err = json.NewDecoder(resp.Body).Decode(&decodedResp)
	if err != nil {
		return response{}, model.WrapErrorWithMethodName(err, "SendPromptWithContext")
	}

	return decodedResp, nil
}
