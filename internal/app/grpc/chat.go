package grpc

import (
	"context"
	"errors"
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	pb "github.com/Yakwilik/MRGAbackend/internal/pb/ailawyer"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/helper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"time"
)

func (a *Implementation) BeginConversation(ctx context.Context, request *pb.BeginConversationRequest) (*pb.BeginConversationResponse, error) {
	email, err := a.getUserEmail(ctx)
	if err != nil {
		return nil, err
	}

	chatID, err := a.useCase.BeginConversation(ctx, email)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	err = a.useCase.SendMessage(ctx, decodeSendMessageRequest(chatID, request.GetMsg()))
	if err != nil {
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			return nil, errValidation.WithDetails(codes.InvalidArgument)
		}
		return nil, err
	}

	go func() {
		answer, err := a.aiBotApi.GetPromptAnswer(ctx, request.GetMsg().GetMessage())
		if err != nil {
			log.Printf("error getting answer from AI: %v", err)
			answer = "Не удалось обработать последний запрос. Повторите его"
		}
		err = a.useCase.SendMessage(ctx, model.CreateMessageData{
			ChatID:  chatID,
			SentAt:  time.Now(),
			FromBot: true,
			Message: answer,
		})
		if err != nil {
			log.Printf("error sending answer to DB: %v", err)
			return
		}
	}()

	return &pb.BeginConversationResponse{
		ChatId: chatID,
	}, nil
}

func encodeConversations(data []model.ConversationData) []*pb.Conversation {
	result := make([]*pb.Conversation, 0, len(data))

	for _, conv := range data {
		result = append(result, &pb.Conversation{
			ChatName:    conv.ChatName,
			LastMessage: conv.LastMessage,
			FromChatbot: conv.FromChatBot,
			Role:        protoRole[conv.Role],
			SentAt:      timestamppb.New(conv.SentAt),
			ChatId:      conv.ChatID,
		})
	}

	return result
}
func (a *Implementation) GetConversations(ctx context.Context, request *pb.GetConversationsRequest) (*pb.GetConversationsResponse, error) {
	email, err := a.getUserEmail(ctx)
	if err != nil {
		return nil, err
	}

	conversations, err := a.useCase.GetConversations(ctx, email)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &pb.GetConversationsResponse{
		Conversations: encodeConversations(conversations),
	}, nil
}

func encodeMessages(data []model.Message) []*pb.Message {
	result := make([]*pb.Message, 0, len(data))

	for _, conv := range data {
		result = append(result, &pb.Message{
			Message:     conv.Message,
			FromChatbot: conv.FromChatBot,
			SentAt:      timestamppb.New(conv.SentAt),
			Role:        protoRole[conv.Role],
		})
	}

	return result
}

func (a *Implementation) GetConversation(ctx context.Context, request *pb.GetConversationRequest) (*pb.GetConversationResponse, error) {
	_, err := a.getUserEmail(ctx)
	if err != nil {
		return nil, err
	}

	chatName, messages, err := a.useCase.GetConversation(ctx, request.GetId())
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &pb.GetConversationResponse{
		Messages: encodeMessages(messages),
		ChatName: chatName,
	}, nil
}

func decodeSendMessageRequest(chatID uint32, newMessage *pb.NewMessage) model.CreateMessageData {
	return model.CreateMessageData{
		ChatID:  chatID,
		SentAt:  helper.ConvertProtoTimestampOrNow(newMessage.GetSentAt()),
		FromBot: false,
		Message: newMessage.GetMessage(),
	}
}
func (a *Implementation) SendMessage(ctx context.Context, request *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	_, err := a.getUserEmail(ctx)
	if err != nil {
		return nil, err
	}

	err = a.useCase.SendMessage(ctx, decodeSendMessageRequest(request.GetChatId(), request.GetMsg()))
	if err != nil {
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			return nil, errValidation.WithDetails(codes.InvalidArgument)
		}
		return nil, err
	}

	go func() {
		answer, err := a.aiBotApi.GetPromptAnswerWithChatHistory(ctx, request.GetChatId())
		if err != nil {
			logger.Info(ctx, "error getting answer from AI: %v", err)
			answer = "Не удалось обработать последний запрос. Повторите его"
		}
		err = a.useCase.SendMessage(ctx, model.CreateMessageData{
			ChatID:  request.GetChatId(),
			SentAt:  time.Now(),
			FromBot: true,
			Message: answer,
		})
		if err != nil {
			log.Printf("error sending answer to DB: %v", err)
			return
		}
	}()

	return &pb.SendMessageResponse{}, nil
}

func (a *Implementation) GetHotThemes(ctx context.Context, request *pb.GetHotThemesRequest) (*pb.GetHotThemesResponse, error) {
	return &pb.GetHotThemesResponse{ActualThemes: []string{
		//"Как отсудить свое имущество при развод",
		//"Что делать, если жена не дает развод",
		"Как уйти с работы по собственному желанию?",
		"Как продать квартиру?",
	}}, nil
}
