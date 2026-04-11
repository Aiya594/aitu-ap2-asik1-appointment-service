package main

import (
	"log"

	"github.com/Aiya594/appointment-services/internal/app"
	cfg "github.com/Aiya594/appointment-services/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	config := cfg.LoadCfg()

	app := app.NewApp(config)
	err = app.Run(config.Port)
	if err != nil {
		log.Fatal(err)
	}
}
