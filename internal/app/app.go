package app

import (
	"context"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/config"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/handlers"
	mwr "github.com/ArtemSarafannikov/AvitoTestTask/internal/middleware"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/repository"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	config  *config.Config
	server  *echo.Echo
	handler *handlers.Handler
}

func New(config *config.Config) *App {
	s := echo.New()
	repo, err := repository.NewPostgresRepository(config.Storage)
	if err != nil {
		panic(err)
	}
	userService := service.NewUserService(repo)
	transactionService := service.NewTransactionService(repo)
	return &App{
		config:  config,
		server:  s,
		handler: handlers.NewHandler(s.Logger, userService, transactionService),
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "App.Run"

	a.SetupHandlers()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := a.server.Start(":8080"); err != nil && err != http.ErrServerClosed {
			a.server.Logger.Fatalf("%s: %w", op, err)
			return
		}
	}()

	<-quit
	a.server.Logger.Infof("%s: %s", op, "Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.server.Logger.Fatalf("%s: %w", op, err)
		return err
	}

	a.server.Logger.Infof("%s: %s", op, "graceful shutdown complete")
	return nil
}

func (a *App) SetupHandlers() {
	a.server.Use(middleware.Logger())
	a.server.Use(middleware.Recover())
	a.server.Logger.SetLevel(log.INFO)

	a.server.POST("/api/auth", a.handler.AuthHandler)

	withAuthGroup := a.server.Group("/api")

	JWTSecret, exist := os.LookupEnv("JWT_SECRET")
	if !exist {
		a.server.Logger.Fatal("JWT_SECRET environment variable not set")
	}
	withAuthGroup.Use(mwr.JWTMiddleware(JWTSecret))
	withAuthGroup.Use(mwr.AuthMiddleware)
	withAuthGroup.GET("/info", a.handler.GetInfo)
	withAuthGroup.POST("/sendCoin", a.handler.SendCoin)
	withAuthGroup.GET("/buy/:item", a.handler.BuyItem)
}
