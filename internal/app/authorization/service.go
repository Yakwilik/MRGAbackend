package authorization

import (
	"context"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	pb "github.com/Yakwilik/MRGAbackend/internal/pb/authorization"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/scratch"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net/http"
	"sync"
	"time"
)

type Implementation struct {
	pb.UnimplementedAuthorizationServer

	users   map[string]model.User
	usersMX *sync.Mutex

	sessions   map[string]string
	sessionsMX *sync.Mutex
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
}

func NewAuthorization(cfg Config) *Implementation {
	return &Implementation{
		users:      make(map[string]model.User),
		usersMX:    &sync.Mutex{},
		sessions:   make(map[string]string),
		sessionsMX: &sync.Mutex{},
	}
}

func (a *Implementation) SignUPV1(ctx context.Context, req *pb.SignUPRequest) (*pb.SignUPResponse, error) {
	a.usersMX.Lock()
	defer a.usersMX.Unlock()
	if _, ok := a.users[req.GetEmail().GetValue()]; ok {
		return nil, status.Error(codes.AlreadyExists, "user with such email already exists")
	}

	a.users[req.GetEmail().GetValue()] = model.User{
		Email:    req.GetEmail().GetValue(),
		Password: req.GetPassword().GetValue(),
	}
	session := uuid.NewString()

	a.sessionsMX.Lock()
	defer a.sessionsMX.Unlock()
	a.sessions[session] = req.GetEmail().GetValue()

	c := http.Cookie{
		Name:     "session_id",
		Value:    session,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	}

	_ = grpc.SendHeader(ctx, metadata.New(map[string]string{
		"Set-Cookie": c.String(),
	}))

	return &pb.SignUPResponse{}, nil
}

func (a *Implementation) CheckLogin(ctx context.Context, req *pb.CheckLoginRequest) (*pb.CheckLoginResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	cookie := md.Get("Cookie")

	request := http.Request{Header: http.Header{"Cookie": cookie}}
	sessionCookie, err := request.Cookie("session_id")
	if err != nil {
		return nil, model.NewValidationError("session_id", "no session_id", "Вы не авторизованы").WithDetails(codes.Unauthenticated)
	}
	a.sessionsMX.Lock()
	userEmail, ok := a.sessions[sessionCookie.Value]
	a.sessionsMX.Unlock()
	if !ok {
		return nil, model.NewValidationError("session_id", "no session_id", "Вы не авторизованы").WithDetails(codes.Unauthenticated)
	}

	a.usersMX.Lock()
	user, ok := a.users[userEmail]
	a.usersMX.Unlock()
	if !ok {
		return nil, model.NewValidationError("session_id", "no session_id", "Вы не авторизованы").WithDetails(codes.Unauthenticated)
	}

	return &pb.CheckLoginResponse{
		IsLogged: true,
		Email:    &wrappers.StringValue{Value: user.Email},
	}, nil
}

func (a *Implementation) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return nil, nil
}
