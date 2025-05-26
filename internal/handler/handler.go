package handler

import "github.com/labstack/echo/v4"

type Handler struct {
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Init() *echo.Echo {
	return nil
}
