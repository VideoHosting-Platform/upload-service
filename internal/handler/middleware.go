package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func (h *Handler) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, "no authoriztion header founded")
		}
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			return c.JSON(http.StatusBadRequest, "not a bearer token")
		}

		userID, err := h.tm.ParseAccessToken(headerParts[1])
		if err != nil {
			fmt.Println(err.Error()) // ! change
			return c.JSON(http.StatusUnauthorized, "token expired or not valid")
		}
		c.Set("userID", userID)
		slog.Info("middleware", "id", userID)
		return next(c)
	}
}
