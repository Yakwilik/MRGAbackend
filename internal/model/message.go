package model

import "time"

type CreateMessageData struct {
	ChatID  uint32
	SentAt  time.Time
	FromBot bool
	Message string
}

type ConversationData struct {
	LastMessage string
	FromChatBot bool
	SentAt      time.Time
	ChatID      uint32
}

type Message struct {
	Message     string
	FromChatBot bool
	SentAt      time.Time
}

type HistoryMessage struct {
	Role Role   `json:"role"`
	Text string `json:"text"`
}

type ChatRequest struct {
	UserID      uint32           `json:"user_id"`
	ChatID      uint32           `json:"chat_id"`
	UserQuery   string           `json:"user_query"`
	ChatHistory []HistoryMessage `json:"chat_history"`
}

type ChatResponseChunk struct {
	UserID        uint32
	ChatID        uint32
	Role          Role
	Chunk         string
	MessageStatus MessageStatus
	ErrorDetails  string
}
