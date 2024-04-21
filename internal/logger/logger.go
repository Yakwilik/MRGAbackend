package logger

import (
	"context"
	"google.golang.org/grpc/metadata"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
	"io"
	"log"
	"log/slog"
	"os"
)

type Options struct {
	OutputAddr  string
	Environment string
	AppName     string
	LogLevel    slog.Level
}

func InitLogger(opts Options) {
	writers := []io.Writer{os.Stdout}

	if opts.OutputAddr != "" {
		writer, err := gelf.NewTCPWriter(opts.OutputAddr)
		if err != nil {
			log.Fatal(err)
		}
		writers = append(writers, writer)
	}
	var handler slog.Handler = slog.NewJSONHandler(io.MultiWriter(writers...), &slog.HandlerOptions{
		AddSource: opts.LogLevel == slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "level" {
				return slog.Any("log_level", a.Value)
			}
			return a
		},
		Level: opts.LogLevel,
	})

	handler = handler.WithGroup("backend").WithAttrs([]slog.Attr{
		slog.Any("AppName", opts.AppName),
		slog.Any("Environment", opts.Environment),
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.Info("Logger initialized", "output address", opts.OutputAddr, "log level", opts.LogLevel)
	log.Printf("Logger initialized with writers: %+v", writers)
}

type reqIDKey struct{}

var reqIDKeyMD = "request_id"

func WithRequestID(ctx context.Context, requestID string) context.Context {
	existingMD, _ := metadata.FromOutgoingContext(ctx)

	newMD := metadata.New(map[string]string{
		reqIDKeyMD: requestID,
	})

	combinedMD := metadata.Join(existingMD, newMD)

	return metadata.NewOutgoingContext(context.WithValue(ctx, reqIDKey{}, requestID), combinedMD)
}

func Info(ctx context.Context, msg string, v ...interface{}) {
	if reqID, ok := ctx.Value(reqIDKey{}).(string); ok {
		v = append(v, reqIDKeyMD, reqID)
	} else {
		// Пытаемся извлечь request_id из метаданных, если он не найден в контексте
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			slog.Info("metadata", v...)
			if requestIDs, ok := md[reqIDKeyMD]; ok && len(requestIDs) > 0 {
				v = append(v, reqIDKeyMD, requestIDs[0])
			}
		}
	}

	slog.InfoContext(ctx, msg, v...)
}

func Error(ctx context.Context, msg string, v ...interface{}) {
	if reqID, ok := ctx.Value(reqIDKey{}).(string); ok {
		v = append(v, reqIDKeyMD, reqID)
	} else {
		// Пытаемся извлечь request_id из метаданных, если он не найден в контексте
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if requestIDs, ok := md["request-id"]; ok && len(requestIDs) > 0 {
				v = append(v, reqIDKeyMD, requestIDs[0])
			}
		}
	}

	slog.ErrorContext(ctx, msg, v...)
}
