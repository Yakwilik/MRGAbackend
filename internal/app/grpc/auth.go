package grpc

import (
	"context"
	"errors"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	pb "github.com/Yakwilik/MRGAbackend/internal/pb/ailawyer"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/helper"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"time"
)

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

	c := helper.GetSessionCookie(a.cfg.Domain(ctx), sessionID, time.Now().Add(time.Hour*24))
	_ = grpc.SendHeader(ctx, metadata.New(map[string]string{
		"Set-cookie": c.String(),
	}))

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

	c := helper.GetSessionCookie(a.cfg.Domain(ctx), sessionID, time.Now().Add(time.Hour*24))
	_ = grpc.SendHeader(ctx, metadata.New(map[string]string{
		"Set-cookie": c.String(),
	}))

	return &pb.LoginResponse{Email: req.GetEmail().GetValue()}, nil
}
