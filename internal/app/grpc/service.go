package grpc

import (
	"context"
	"errors"
	"github.com/Yakwilik/MRGAbackend/internal/client/ai_bot"
	"github.com/Yakwilik/MRGAbackend/internal/core"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	pb "github.com/Yakwilik/MRGAbackend/internal/pb/ailawyer"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/helper"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/scratch"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type Implementation struct {
	pb.UnimplementedAuthorizationServer

	useCase  core.UseCase
	aiBotApi ai_bot.Interface
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
	Auth   core.UseCase
	BotApi ai_bot.Interface
}

func NewAuthorization(cfg Config) *Implementation {
	return &Implementation{
		useCase:  cfg.Auth,
		aiBotApi: cfg.BotApi,
	}
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
