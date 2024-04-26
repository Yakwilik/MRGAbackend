package scratch

import (
	"context"
	"fmt"
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"log/slog"
	"net/http"
)

var (
	PublicPort  uint = 7001
	GrpcPort    uint = 7002
	BindAddress      = ""

	LogLevel    slog.Level = slog.LevelInfo
	AppName     string     = "app"
	Environment string     = "dev"
)

type Options struct {
	PortHTTP uint
	PortGRPC uint

	BindAddress string

	ServeMuxOpts []runtime.ServeMuxOption

	// Middleware для эндпоинтов для кастомных рест-эндпоинтов
	CustomMuxMiddleware       []func(http.Handler) http.Handler
	EnableCustomMuxMiddleware bool
	EnableCustomHandler       bool
	CustomHandler             http.Handler

	// Middleware для эндпоинтов grpc-gateway
	EnableGatewayMiddleware bool
	GatewayMiddleware       []func(http.Handler) http.Handler

	// Middleware для всех публичных эндпоинтов
	EnablePublicMuxMiddleware bool
	PublicMuxMiddleware       []func(http.Handler) http.Handler

	LogLevel     slog.Level
	LoggerOutput string
	AppName      string
	Environment  string
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
		LogLevel:    LogLevel,
		AppName:     AppName,
		Environment: Environment,
	}

	for _, o := range opts {
		if err := o.Apply(oo); err != nil {
			return nil, fmt.Errorf("invalid option: %w", err)
		}
	}

	return oo, nil
}

// LogInterceptor это UnaryInterceptor который логгирует детали запроса и ответа.
func LogInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Логгирование начала обработки запроса
	slog.Info("Received request", "server", info.Server, "method", info.FullMethod, "request", req)

	// Обработка запроса
	resp, err := handler(ctx, req)

	// Логгирование ответа
	if err != nil {
		logger.Error(ctx, "Request completed with error", "error", err)
	} else {
		logger.Error(ctx, "Request completed successfully", "response", resp)
	}

	return resp, err
}

func WithCustomRestHandler(handler http.Handler) Option {
	return optionFn(func(o *Options) error {
		o.CustomHandler = handler
		o.EnableCustomHandler = true

		return nil
	})
}

func WithPublicMiddleware(f func(handler http.Handler) http.Handler) Option {
	return optionFn(func(o *Options) error {
		o.EnableCustomMuxMiddleware = true
		o.CustomMuxMiddleware = append(o.CustomMuxMiddleware, f)
		return nil
	})
}

// WithLogLevel устанавливает уровень логирования
func WithLogLevel(level slog.Level) Option {
	return optionFn(func(o *Options) error {
		o.LogLevel = level
		return nil
	})
}

// WithLoggerOutput устанавливает выход для логгера
func WithLoggerOutput(output string) Option {
	return optionFn(func(o *Options) error {
		o.LoggerOutput = output
		return nil
	})
}

// WithAppName устанавливает название приложения
func WithAppName(name string) Option {
	return optionFn(func(o *Options) error {
		o.AppName = name
		return nil
	})
}

// WithEnvironment устанавливает окружение, в котором работает приложение
func WithEnvironment(env string) Option {
	return optionFn(func(o *Options) error {
		o.Environment = env
		return nil
	})
}

func WithGatewayMiddleware(f func(handler http.Handler) http.Handler) Option {
	return optionFn(func(o *Options) error {
		o.EnableCustomMuxMiddleware = true
		o.CustomMuxMiddleware = append(o.CustomMuxMiddleware, f)
		return nil
	})
}
