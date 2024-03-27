package main

import (
	"github.com/Yakwilik/MRGAbackend/internal/app/grpc"
	"github.com/Yakwilik/MRGAbackend/internal/config"
	"github.com/Yakwilik/MRGAbackend/internal/core"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/helper"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/scratch"
	"github.com/Yakwilik/MRGAbackend/internal/storage"
	"log"
)

func main() {
	cfg := config.ParseConfig()
	app, err := scratch.InitApp()
	if err != nil {
		log.Fatalf("can't init app: %s", err)
	}
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
	session := storage.New(db)

	if err := app.Run(grpc.NewAuthorization(grpc.Config{
		Auth: core.New(session),
	})); err != nil {
		log.Fatalf("can't run app: %s", err)
	}
}
