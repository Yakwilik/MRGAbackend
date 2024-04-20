package main

import (
	"github.com/Yakwilik/MRGAbackend/internal/app/grpc"
	"github.com/Yakwilik/MRGAbackend/internal/app/rest"
	"github.com/Yakwilik/MRGAbackend/internal/client/ai_bot"
	"github.com/Yakwilik/MRGAbackend/internal/config"
	"github.com/Yakwilik/MRGAbackend/internal/core"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/helper"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/scratch"
	storagePkg "github.com/Yakwilik/MRGAbackend/internal/storage"
	"log"
	"net/http"
)

func main() {
	cfg := config.ParseConfig()

	db, err := helper.NewPostgresDB(helper.PGConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		Username: cfg.DBUsername,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
	})
	if err != nil {
		log.Fatalln(err)
	}
	storage := storagePkg.New(db)

	core := core.New(storage)

	app, err := scratch.InitApp(
		scratch.WithCustomRestHandler(rest.New(core).Init()),
		scratch.WithMiddleware(func(handler http.Handler) http.Handler {
			return helper.AuthMiddleware(core, handler)
		}),
	)
	if err != nil {
		log.Fatalf("can't init app: %s", err)
	}

	// TODO вынести адрес сервера в конфиг
	if err := app.Run(grpc.NewAuthorization(grpc.Config{
		Auth:   core,
		BotApi: ai_bot.New(core, "http://212.233.96.112:8002"),
	})); err != nil {
		log.Fatalf("can't run app: %s", err)
	}
}
