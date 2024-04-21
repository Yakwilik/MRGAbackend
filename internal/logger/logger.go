package logger

import (
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
		AddSource:   true,
		ReplaceAttr: nil,
		Level:       opts.LogLevel,
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
