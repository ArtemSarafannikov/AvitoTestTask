package app

import (
	cstErrors "github.com/ArtemSarafannikov/AvitoTestTask/internal/error"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/repository"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/service"
	"github.com/labstack/echo/v4"
	"net/http"
)

type App struct {
	server      *echo.Echo
	userService *service.UserService
}

func New() *App {
	s := echo.New()
	repo, err := repository.NewPostgresRepository()
	if err != nil {
		panic(err)
	}
	userService := service.NewUserService(repo)
	return &App{
		server:      s,
		userService: userService,
	}
}

func (a *App) Run() {
	a.SetupHandlers()
	a.server.Logger.Fatal(a.server.Start(":8080"))
}

func (a *App) SetupHandlers() {
	a.server.GET("/api/info", getInfo)
	a.server.POST("/api/sendCoin", sendCoin)
	a.server.GET("/api/buy/:item", buyItem)
	a.server.POST("/api/auth", a.authHandler)
}

func getInfo(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Info")
}

func sendCoin(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "SendCoin")
}

func buyItem(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "BuyItem")
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (a *App) authHandler(ctx echo.Context) error {
	var req LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": cstErrors.BadRequestDataError.Error()})
	}

	token, err := a.userService.Login(ctx.Request().Context(), req.Username, req.Password)
	if err != nil {
		return ctx.String(http.StatusUnauthorized, err.Error())
	}
	return ctx.String(http.StatusOK, token)
}
