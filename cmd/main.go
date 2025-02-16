package main

import (
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/app"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/config"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic("No .env file found")
	}
}

func main() {
	cfg := config.MustLoad()

	server := app.New(cfg)
	server.MustRun()
}
