package app

import (
	"fmt"
	cstErrors "github.com/ArtemSarafannikov/AvitoTestTask/internal/error"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/model"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/repository"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/service"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"net/http"
)

type App struct {
	server             *echo.Echo
	userService        *service.UserService
	transactionService *service.TransactionService
}

func New() *App {
	s := echo.New()
	repo, err := repository.NewPostgresRepository()
	if err != nil {
		panic(err)
	}
	userService := service.NewUserService(repo)
	transcationService := service.NewTransactionService(repo)
	return &App{
		server:             s,
		userService:        userService,
		transactionService: transcationService,
	}
}

func (a *App) Run() {
	a.SetupHandlers()
	a.server.Logger.Fatal(a.server.Start(":8080"))
}

func (a *App) SetupHandlers() {
	// TODO: make handlers in other module
	a.server.POST("/api/auth", a.authHandler)

	withAuthGroup := a.server.Group("/api")

	// TODO: make secret key in .env
	// TODO: replace message to error in response
	//withAuthGroup.Use(echojwt.JWT([]byte(utils.JWTSecret)))
	withAuthGroup.Use(JWTMiddleware(utils.JWTSecret))
	withAuthGroup.Use(AuthMiddleware)
	withAuthGroup.GET("/info", a.getInfo)
	withAuthGroup.POST("/sendCoin", a.sendCoin)
	withAuthGroup.GET("/buy/:item", a.buyItem)
}

func (a *App) getInfo(ctx echo.Context) error {
	return ctx.String(http.StatusOK, fmt.Sprintf("Info: %s", ctx.Get("sub").(string)))
}

func (a *App) sendCoin(ctx echo.Context) error {
	var req model.SendCoinRequest
	if err := ctx.Bind(&req); err != nil {
		// TODO: remake error to struct model
		return ctx.JSON(http.StatusBadRequest, map[string]string{"errors": cstErrors.BadRequestDataError.Error()})
	}
	if err := a.transactionService.SendCoin(ctx.Request().Context(), ctx.Get(utils.UserIdCtxKey).(string), req.ToUser, req.Amount); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"errors": err.Error()})
	}
	return ctx.NoContent(http.StatusOK)
}

func (a *App) buyItem(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "BuyItem")
}

func (a *App) authHandler(ctx echo.Context) error {
	var req model.AuthRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": cstErrors.BadRequestDataError.Error()})
	}

	token, err := a.userService.Login(ctx.Request().Context(), req.Username, req.Password)
	if err != nil {
		return ctx.String(http.StatusUnauthorized, err.Error())
	}
	return ctx.String(http.StatusOK, token)
}

func JWTMiddleware(secret string) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey:  []byte(secret),
		TokenLookup: "header:Authorization",
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return jwt.MapClaims{}
		},
	})
}

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		user := ctx.Get("user")
		invalidAuthError := model.ErrorResponse{Errors: "invalid token"}
		if user == nil {
			return ctx.JSON(http.StatusUnauthorized, invalidAuthError)
		}

		token, ok := user.(*jwt.Token)
		if !ok {
			return ctx.JSON(http.StatusUnauthorized, invalidAuthError)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return ctx.JSON(http.StatusUnauthorized, invalidAuthError)
		}

		userId, ok := claims["sub"].(string)
		if !ok || userId == "" {
			return ctx.JSON(http.StatusUnauthorized, invalidAuthError)
		}

		ctx.Set(utils.UserIdCtxKey, userId)

		return next(ctx)
	}
}
