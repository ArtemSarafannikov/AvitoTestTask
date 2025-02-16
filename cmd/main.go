package main

import (
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/app"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/config"
)

func main() {
	cfg := config.MustLoad()

	server := app.New(cfg)
	server.MustRun()
}
