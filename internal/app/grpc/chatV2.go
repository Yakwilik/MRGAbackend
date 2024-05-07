package grpc

import (
	"errors"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	pb "github.com/Yakwilik/MRGAbackend/internal/pb/ailawyer"
	"google.golang.org/grpc/codes"
	"strings"
)

func (a *Implementation) SendMessageV2(request *pb.SendMessageRequest, server pb.Backend_SendMessageV2Server) error {
	_, err := a.getUserEmail(server.Context())
	if err != nil {
		return err
	}

	err = a.useCase.SendMessage(server.Context(), decodeSendMessageRequest(request.GetChatId(), request.GetMsg()))
	if err != nil {
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			return errValidation.WithDetails(codes.InvalidArgument)
		}
		return err
	}

	_, history, err := a.useCase.GetConversation(server.Context(), request.GetChatId())
	if err != nil {
		return err
	}

	responseChan, err := a.aiBotService.RespondToUserQuery(server.Context(), model.ChatRequest{
		ChatID:      request.GetChatId(),
		UserQuery:   request.GetMsg().GetMessage(),
		ChatHistory: encodeToChatHistory(history),
	})

	if err != nil {
		return err
	}

	result := strings.Builder{}
	for data := range responseChan {
		if data.MessageStatus == model.StatusOk {
			result.WriteString(data.Chunk)
		}
		err := server.Send(&pb.MessageV2{
			Text:         data.Chunk,
			Role:         protoRole[data.Role],
			ChatId:       &data.ChatID,
			Status:       protoStatus[data.MessageStatus],
			ErrorDetails: data.ErrorDetails,
		})
		if err != nil {
			return err
		}
	}

	err = a.useCase.SendMessage(server.Context(), decodeSendMessageRequest(request.GetChatId(), &pb.NewMessage{
		Message: result.String(),
	}))
	if err != nil {
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			return errValidation.WithDetails(codes.InvalidArgument)
		}
		return err
	}

	return nil
}

func encodeToChatHistory(messages []model.Message) []model.HistoryMessage {
	lastMessageID := len(messages) - 1
	historyMessages := make([]model.HistoryMessage, 0, len(messages)-1)
	for index, message := range messages {
		if index == lastMessageID {
			continue
		} else {
			historyMessages = append(historyMessages, model.HistoryMessage{
				Role: getRole(message.FromChatBot),
				Text: message.Message,
			})
		}
	}

	return historyMessages
}

func getRole(fromChatBot bool) model.Role {
	if fromChatBot {
		return model.RoleAssistant
	}
	return model.RoleUser
}

var protoRole = map[model.Role]pb.MessageRole{
	model.RoleUser:             pb.MessageRole_user,
	model.RoleAssistant:        pb.MessageRole_assistant,
	model.RoleTechnical:        pb.MessageRole_technical,
	model.RoleDocumentRedirect: pb.MessageRole_document_redirect,
	model.RoleExtraQuestions:   pb.MessageRole_extra_questions,
}

var protoStatus = map[model.MessageStatus]pb.MessageStatus{
	model.StatusOk:   pb.MessageStatus_OK,
	model.StatusFail: pb.MessageStatus_ERROR,
}
