package grpc

import (
	"context"
	"errors"
	"github.com/Yakwilik/MRGAbackend/internal/client/ai_bot"
	"github.com/Yakwilik/MRGAbackend/internal/client/ai_botV2"
	"github.com/Yakwilik/MRGAbackend/internal/config"
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
	pb.UnimplementedBackendServer
	cfg          *config.Config
	useCase      core.UseCase
	aiBotApi     ai_bot.Interface
	aiBotService ai_botV2.Interface
}

type BackendServiceDesc struct {
	svc pb.BackendServer
}

func (a *BackendServiceDesc) RegisterGRPC(server *grpc.Server) {
	pb.RegisterBackendServer(server, a.svc)
}

func (a *BackendServiceDesc) RegisterGateway(ctx context.Context, mux *runtime.ServeMux) error {
	return pb.RegisterBackendHandlerServer(ctx, mux, a.svc)
}

func NewBackendServiceDesc(implementation *Implementation) scratch.ServiceDesc {
	return &BackendServiceDesc{svc: implementation}
}

func (a *Implementation) GetDescription() scratch.ServiceDesc {
	return NewBackendServiceDesc(a)
}

type Config struct {
	UseCase core.UseCase
	// почему не в usecase?
	BotApi        ai_bot.Interface
	BotApiService ai_botV2.Interface
	Config        *config.Config
}

func NewBackend(cfg Config) *Implementation {
	return &Implementation{
		useCase:      cfg.UseCase,
		aiBotApi:     cfg.BotApi,
		aiBotService: cfg.BotApiService,
		cfg:          cfg.Config,
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
