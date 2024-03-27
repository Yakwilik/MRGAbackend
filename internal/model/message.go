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
