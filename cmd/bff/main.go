package main

import (
	"context"
	"github.com/Yakwilik/MRGAbackend/internal/app/grpc"
	"github.com/Yakwilik/MRGAbackend/internal/app/rest"
	"github.com/Yakwilik/MRGAbackend/internal/client/ai_bot"
	"github.com/Yakwilik/MRGAbackend/internal/client/ai_botV2"
	"github.com/Yakwilik/MRGAbackend/internal/config"
	"github.com/Yakwilik/MRGAbackend/internal/core"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/helper"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/scratch"
	storagePkg "github.com/Yakwilik/MRGAbackend/internal/storage"
	"log"
	"net/http"
)

func main() {
	ctx := context.Background()
	cfg := config.ParseConfig()

	app, err := scratch.InitApp(
		scratch.WithAppName(cfg.AppName(ctx)),
		scratch.WithEnvironment(cfg.Environment(ctx)),
		scratch.WithLoggerOutput(cfg.LoggerAddr(ctx)),
		scratch.WithLogLevel(cfg.LogLevel(ctx)),
	)
	if err != nil {
		log.Fatalf("can't init app: %s", err)
	}

	db, err := helper.NewPostgresDB(helper.PGConfig{
		Host:     cfg.DBHost(ctx),
		Port:     cfg.DBPort(ctx),
		Username: cfg.DBUsername(ctx),
		Password: cfg.DBPassword(ctx),
		DBName:   cfg.DBName(ctx),
	})
	if err != nil {
		log.Fatalln(err)
	}
	storage := storagePkg.New(db)

	aiBotService := ai_botV2.MustNew(ai_botV2.Config{
		ServiceAddr: cfg.ChatBotServiceAddr(ctx),
	})
	core := core.New(storage, aiBotService)
	app.WithCustomRestHandler(rest.New(core, aiBotService).Init()).
		WithPublicMuxMiddleware(scratch.LoggingMiddleware).
		WithCustomRestMiddleware(func(handler http.Handler) http.Handler {
			return helper.AuthMiddleware(core, handler)
		}).WithPublicMuxMiddleware(scratch.CorsMiddleware)

	if err := app.Run(grpc.NewBackend(grpc.Config{
		UseCase:       core,
		BotApi:        ai_bot.New(core, cfg.ChatBotAddr(ctx)),
		BotApiService: aiBotService,
		Config:        cfg,
	})); err != nil {
		log.Fatalf("can't run app: %s", err)
	}
}
