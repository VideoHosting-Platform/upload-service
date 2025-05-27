package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/minio/minio-go/v7"
)

type Handler struct {
	bucket string
	mc     *minio.Client
}

func New(mc *minio.Client, bucketName string) *Handler {
	return &Handler{mc: mc, bucket: bucketName}
}

func (h *Handler) Init() *echo.Echo {

	router := echo.New()

	router.Use(middleware.CORS())

	router.GET("/ping", func(c echo.Context) error {
		return c.JSON(200, struct {
			Status string
		}{Status: "ok"})
	})

	router.POST("/upload", h.uploadVideo)

	return router
}

func (h *Handler) uploadVideo(c echo.Context) error {
	reader, err := c.Request().MultipartReader()
	if err != nil {
		return c.String(http.StatusBadRequest, "Not a multipart request")
	}

	pr, pw := io.Pipe()

	var title, filename string
	titleReady := make(chan struct{})

	done := make(chan error)
	go func() {
		<-titleReady
		_, err := h.mc.PutObject(
			c.Request().Context(),
			h.bucket,
			title+"_"+filename, // ! change
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
			filename = part.FileName()
			fmt.Println("Original filename:", part.FileName())
			titleReady <- struct{}{}
			io.Copy(pw, part)
		}
	}
	pw.Close()

	if err := <-done; err != nil {
		fmt.Println(err.Error())
		return c.String(http.StatusInternalServerError, fmt.Sprintf("error while uploading to s3 - %s", err.Error()))
	}

	return c.String(http.StatusOK, "File uploaded to MinIO!")
}
