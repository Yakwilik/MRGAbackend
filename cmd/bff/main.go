package main

import (
	"github.com/Yakwilik/MRGAbackend/internal/app/authorization"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/scratch"
	"log"
)

func main() {
	app, err := scratch.InitApp()
	if err != nil {
		log.Fatalf("can't init app: %s", err)
	}

	if err := app.Run(authorization.NewAuthorization(authorization.Config{})); err != nil {
		log.Fatalf("can't run app: %s", err)
	}
}
