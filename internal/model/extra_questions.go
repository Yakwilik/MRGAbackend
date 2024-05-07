package model

import (
	"encoding/json"
	"fmt"
)

// ExtraQuestionsV1 – в таком формате приходит список вопросов: [["вопрос?", "1"]], 1, если на вопрос можно ответить да или нет
type ExtraQuestionsV1 [][]string

type ExtraQuestion struct {
	Question     string `json:"question"`
	MonoQuestion bool   `json:"monoQuestion"`
}

func ParseQuestions(jsonData string) ([]ExtraQuestion, error) {
	// Промежуточная структура для декодирования JSON
	var inputData []ExtraQuestion

	// Декодирование JSON в структуру данных
	if err := json.Unmarshal([]byte(jsonData), &inputData); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	return inputData, nil
}
