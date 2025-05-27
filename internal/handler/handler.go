package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/VideoHosting-Platform/upload-service/pkg/tokenutil"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/minio/minio-go/v7"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Handler struct {
	tm *tokenutil.TokenManager

	bucket string
	mc     *minio.Client

	ch        *amqp.Channel
	queueName string
	chMutex   sync.Mutex
}

type VideoEvent struct {
	VideoID    uuid.UUID `json:"video_id"`
	UserID     int64     `json:"user_id"`
	VideoTitle string    `json:"video_title"`
}

func New(tm *tokenutil.TokenManager, mc *minio.Client, bucketName string, ch *amqp.Channel, qn string) *Handler {
	return &Handler{
		mc:        mc,
		bucket:    bucketName,
		ch:        ch,
		queueName: qn,
	}
}

func (h *Handler) Init() *echo.Echo {

	router := echo.New()

	router.Use(middleware.CORS())
	router.Use(h.AuthMiddleware)

	router.GET("/ping", func(c echo.Context) error {
		return c.JSON(200, struct {
			Status string
		}{Status: "ok"})
	})

	router.POST("/upload", h.uploadVideo)

	return router
}

func (h *Handler) uploadVideo(c echo.Context) error {

	// userID := c.Get("userID").(int)
	// fmt.Println(userID)

	reader, err := c.Request().MultipartReader()
	if err != nil {
		return c.String(http.StatusBadRequest, "Not a multipart request")
	}

	var title string
	videoID := uuid.New()
	pr, pw := io.Pipe()
	done := make(chan error)

	go func() {
		_, err := h.mc.PutObject(
			c.Request().Context(),
			h.bucket,
			videoID.String(), // ! change
			pr,
			-1,
			minio.PutObjectOptions{},
		)

		done <- err
	}()

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		fmt.Println(part.FormName())
		if part.FormName() == "title" {
			buf := new(bytes.Buffer)
			io.Copy(buf, part)
			_, err := io.Copy(buf, part)
			if err != nil {
				return c.String(http.StatusInternalServerError, "read title error")
			}
			title = buf.String()
		} else if part.FormName() == "video" {
			io.Copy(pw, part)
		}
	}
	pw.Close()

	if err := <-done; err != nil {
		fmt.Println(err.Error())
		return c.String(http.StatusInternalServerError, fmt.Sprintf("error while uploading to s3 - %s", err.Error()))
	}

	event := VideoEvent{
		VideoID:    videoID,
		UserID:     1, // ! change
		VideoTitle: title,
	}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err) // ! change
	}

	h.chMutex.Lock()
	defer h.chMutex.Unlock()

	err = h.ch.Publish( // ! change
		"", // exchange
		h.queueName,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		c.String(http.StatusInternalServerError, "Can not upload video for processing")
	}

	return c.String(http.StatusOK, "File uploaded to MinIO!")
}
