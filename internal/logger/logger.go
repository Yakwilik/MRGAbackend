package logger

import (
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
	"io"
	"log"
	"log/slog"
	"os"
)

func InitLogger(outputAddr string, environment string) {
	writers := []io.Writer{os.Stdout}
	log.Printf("environment: %s", environment)
	if environment == "prod" {
		writer, err := gelf.NewTCPWriter(outputAddr)
		if err != nil {
			log.Fatal(err)
		}
		writers = append(writers, writer)
	}
	handler := slog.NewJSONHandler(io.MultiWriter(writers...), &slog.HandlerOptions{
		AddSource:   true,
		ReplaceAttr: nil,
	})

	slog.SetDefault(slog.New(handler))
	log.Printf("Logger initialized with writers: %+v", writers)
}
