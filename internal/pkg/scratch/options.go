package scratch

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"log"
	"log/slog"
)

var (
	PublicPort  uint = 7001
	GrpcPort    uint = 7002
	BindAddress      = ""
)

type Options struct {
	PortHTTP uint
	PortGRPC uint

	BindAddress string

	ServeMuxOpts []runtime.ServeMuxOption
}

type Option interface {
	Apply(options *Options) error
}

type optionFn func(options *Options) error

func (fn optionFn) Apply(opts *Options) error {
	return fn(opts)
}

func evaluateOptions(opts []Option) (*Options, error) {
	oo := &Options{
		PortHTTP:    PublicPort,
		PortGRPC:    GrpcPort,
		BindAddress: BindAddress,
	}

	for _, o := range opts {
		if err := o.Apply(oo); err != nil {
			return nil, fmt.Errorf("invalid option: %w", err)
		}
	}

	return oo, nil
}

// logInterceptor это UnaryInterceptor который логгирует детали запроса и ответа.
func logInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Логгирование начала обработки запроса
	log.Printf("Received request: %v", req)

	// Обработка запроса
	resp, err := handler(ctx, req)

	// Логгирование ответа
	if err != nil {
		slog.Error("Request completed with error: %v", err)
	} else {
		slog.Info("Request completed successfully, response: %v", resp)
	}

	return resp, err
}
