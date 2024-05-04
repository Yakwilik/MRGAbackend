package retext

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	"net/http"
)

type client interface {
	PostTask(ctx context.Context, task taskRequest) (postTaskResponse, error)
	GetTaskStatus(ctx context.Context, taskID string) (getTaskStatusResponse, error)
}

type task string
type apiMethod string
type lang string
type status string

const (
	methodPostTask  = apiMethod("queue_task")
	methodGetStatus = apiMethod("queue_check")
	url             = "https://api.retext.ai/api/v1/"
	taskSummarize   = task("headlines")
	langRU          = lang("ru")
	statusOK        = status("ok")
)

// taskRequest – тело запроса на создание таски
type taskRequest struct {
	Task       task       `json:"task"`
	SourceText string     `json:"source_text"`
	Additional additional `json:"additional"`
}

// additional – дополнительные поля для тела запроса на создание таски
type additional struct {
	Lang      lang   `json:"lang"`
	MaxLength uint32 `json:"max_length"`
}

// postTaskResponse – тело ответа на запрос создания таски
type postTaskResponse struct {
	Status status               `json:"status"`
	Data   postTaskResponseData `json:"data"`
}

type postTaskResponseData struct {
	TaskId     string `json:"taskId"`
	SourceLang lang   `json:"source_lang"`
}

type clientImpl struct {
	client *http.Client
}

func NewClient() client {
	return &clientImpl{
		client: &http.Client{},
	}
}

func (c *clientImpl) PostTask(ctx context.Context, task taskRequest) (postTaskResponse, error) {
	bodyBytes, err := json.Marshal(task)
	if err != nil {
		return postTaskResponse{}, model.WrapErrorWithMethodName(err, "PostTask")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", url, methodPostTask), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return postTaskResponse{}, model.WrapErrorWithMethodName(err, "PostTask")
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return postTaskResponse{}, model.WrapErrorWithMethodName(err, "PostTask")
	}

	decodedResp := postTaskResponse{}
	if resp.StatusCode == http.StatusForbidden {
		return postTaskResponse{}, errors.New("forbidden")
	}
	err = json.NewDecoder(resp.Body).Decode(&decodedResp)
	if err != nil {
		return postTaskResponse{}, model.WrapErrorWithMethodName(err, "SendPromptWithContext")
	}

	return decodedResp, nil
}

type getTaskStatusResponse struct {
	Status status                    `json:"status"`
	Data   getTaskStatusResponseData `json:"data"`
}

type getTaskStatusResponseData struct {
	Ready      bool     `json:"ready"`
	Successful bool     `json:"successful"`
	Result     []string `json:"result"`
}

func (c *clientImpl) GetTaskStatus(ctx context.Context, taskID string) (getTaskStatusResponse, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s%s?taskId=%s", url, methodGetStatus, taskID))
	if err != nil {
		return getTaskStatusResponse{}, model.WrapErrorWithMethodName(err, "GetTaskStatus")
	}

	decodedResp := getTaskStatusResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&decodedResp); err != nil {
		return getTaskStatusResponse{}, model.WrapErrorWithMethodName(err, "SendPromptWithContext")
	}
	logger.Info(ctx, "Summarize", "responseBody", decodedResp)
	return decodedResp, nil
}
