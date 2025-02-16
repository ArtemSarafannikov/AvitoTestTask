package middleware

import (
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/model"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

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

func JWTMiddleware(secret string) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey:  []byte(secret),
		TokenLookup: "header:Authorization",
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return jwt.MapClaims{}
		},
		ParseTokenFunc: func(c echo.Context, auth string) (interface{}, error) {
			auth = strings.TrimPrefix(auth, "Bearer ")
			return jwt.Parse(auth, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
		},
	})
}
