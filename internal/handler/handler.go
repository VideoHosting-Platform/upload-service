package handler

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Интерфейс для Minio клиента
type MinioClient interface {
	PutObject(ctx context.Context, objectName string, reader io.Reader) error
}

// Интерфейс для RabbitMQ
type EventPublisher interface {
	Publish(ctx context.Context, body []byte) error
}

type Handler struct {
	mc MinioClient
	q  EventPublisher
}

type VideoEvent struct {
	VideoID    uuid.UUID `json:"video_id"`
	UserID     int64     `json:"user_id"`
	VideoTitle string    `json:"video_title"`
}

func New(mc MinioClient, q EventPublisher) *Handler {
	return &Handler{
		mc: mc,
		q:  q,
	}
}

func (h *Handler) Init() *echo.Echo {

	router := echo.New()

	router.Use(middleware.CORS())
	router.Use(echoprometheus.NewMiddleware("upload_service"))

	router.GET("/metrics", echoprometheus.NewHandler())

	router.GET("/ping", func(c echo.Context) error {
		return c.JSON(200, struct {
			Status string
		}{Status: "ok"})
	})

	router.POST("/upload", h.uploadVideo)

	return router
}
