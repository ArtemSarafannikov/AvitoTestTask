package main

import (
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/app"
)

func main() {
	server := app.New()
	server.MustRun()
}
