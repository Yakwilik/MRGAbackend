package grpc

import (
	"context"
	"errors"
	"github.com/Yakwilik/MRGAbackend/internal/core"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	pb "github.com/Yakwilik/MRGAbackend/internal/pb/ailawyer"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/helper"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/scratch"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Implementation struct {
	pb.UnimplementedAuthorizationServer

	useCase core.UseCase
}

type AuthorizationServiceDesc struct {
	svc pb.AuthorizationServer
}

func (a *AuthorizationServiceDesc) RegisterGRPC(server *grpc.Server) {
	pb.RegisterAuthorizationServer(server, a.svc)
}

func (a *AuthorizationServiceDesc) RegisterGateway(ctx context.Context, mux *runtime.ServeMux) error {
	return pb.RegisterAuthorizationHandlerServer(ctx, mux, a.svc)
}

func NewAuthorizationServiceDesc(implementation *Implementation) scratch.ServiceDesc {
	return &AuthorizationServiceDesc{svc: implementation}
}

func (a *Implementation) GetDescription() scratch.ServiceDesc {
	return NewAuthorizationServiceDesc(a)
}

type Config struct {
	Auth core.UseCase
}

func NewAuthorization(cfg Config) *Implementation {
	return &Implementation{
		useCase: cfg.Auth,
	}
}

func (a *Implementation) SignUPV1(ctx context.Context, req *pb.SignUPRequest) (*pb.SignUPResponse, error) {
	err := a.useCase.SignUp(ctx, model.User{
		Email:    req.GetEmail().GetValue(),
		Password: req.GetPassword().GetValue(),
	})

	if err != nil {
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			return nil, errValidation.WithDetails(codes.InvalidArgument)
		}
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	sessionID, err := a.useCase.Login(ctx, model.User{
		Email:    req.GetEmail().GetValue(),
		Password: req.GetPassword().GetValue(),
	})
	if err != nil {
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			return nil, errValidation.WithDetails(codes.InvalidArgument)
		}
		return nil, err
	}

	helper.SetSessionID(ctx, sessionID)

	return &pb.SignUPResponse{Email: req.GetEmail().GetValue()}, nil
}

func (a *Implementation) CheckLogin(ctx context.Context, req *pb.CheckLoginRequest) (*pb.CheckLoginResponse, error) {
	sessionID, ok := helper.SessionIDFromContextMD(ctx)
	if !ok {
		return nil, model.NewValidationError("session_id", "no session_id", "Вы не авторизованы").WithDetails(codes.Unauthenticated)
	}

	email, err := a.useCase.CheckLogin(ctx, sessionID)
	if err != nil {
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			return nil, errValidation.WithDetails(codes.InvalidArgument)
		}
		return nil, err
	}

	return &pb.CheckLoginResponse{
		IsLogged: true,
		Email:    &wrappers.StringValue{Value: email},
	}, nil
}

func (a *Implementation) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	sessionID, err := a.useCase.Login(ctx, model.User{
		Email:    req.GetEmail().GetValue(),
		Password: req.GetPassword().GetValue(),
	})
	if err != nil {
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			return nil, errValidation.WithDetails(codes.InvalidArgument)
		}
		return nil, err
	}

	helper.SetSessionID(ctx, sessionID)

	return &pb.LoginResponse{Email: req.GetEmail().GetValue()}, nil
}

func (a *Implementation) BeginConversation(ctx context.Context, request *pb.BeginConversationRequest) (*pb.BeginConversationResponse, error) {
	email, err := a.getUserEmail(ctx)
	if err != nil {
		return nil, err
	}

	chatID, err := a.useCase.BeginConversation(ctx, email)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &pb.BeginConversationResponse{
		ChatId: chatID,
	}, nil
}

func encodeConversations(data []model.ConversationData) []*pb.Conversation {
	result := make([]*pb.Conversation, 0, len(data))

	for _, conv := range data {
		result = append(result, &pb.Conversation{
			LastMessage: conv.LastMessage,
			FromChatbot: conv.FromChatBot,
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
		})
	}

	return result
}

func (a *Implementation) GetConversation(ctx context.Context, request *pb.GetConversationRequest) (*pb.GetConversationResponse, error) {
	_, err := a.getUserEmail(ctx)
	if err != nil {
		return nil, err
	}

	messages, err := a.useCase.GetConversation(ctx, request.GetId())
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &pb.GetConversationResponse{Messages: encodeMessages(messages)}, nil
}

func decodeSendMessageRequest(request *pb.SendMessageRequest) model.CreateMessageData {
	return model.CreateMessageData{
		ChatID:  request.ChatId,
		SentAt:  helper.ConvertProtoTimestampOrNow(request.SentAt),
		FromBot: request.FromBot,
		Message: request.Message,
	}
}
func (a *Implementation) SendMessage(ctx context.Context, request *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	_, err := a.getUserEmail(ctx)
	if err != nil {
		return nil, err
	}

	err = a.useCase.SendMessage(ctx, decodeSendMessageRequest(request))
	if err != nil {
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			return nil, errValidation.WithDetails(codes.InvalidArgument)
		}
		return nil, err
	}

	return &pb.SendMessageResponse{}, nil
}

func (a *Implementation) getUserEmail(ctx context.Context) (string, error) {
	sessionID, ok := helper.SessionIDFromContextMD(ctx)
	if !ok {
		return "nil", model.NewValidationError("session_id", "no session_id", "Вы не авторизованы").WithDetails(codes.Unauthenticated)
	}

	email, err := a.useCase.CheckLogin(ctx, sessionID)
	if err != nil {
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			return "", errValidation.WithDetails(codes.InvalidArgument)
		}
		return "", err
	}

	return email, nil
}
