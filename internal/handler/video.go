package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handler) uploadVideo(c echo.Context) error {

	reader, err := c.Request().MultipartReader()
	if err != nil {
		return c.String(http.StatusBadRequest, "Not a multipart request")
	}

	var title string
	videoID := uuid.New()
	pr, pw := io.Pipe()
	done := make(chan error)

	go h.uploadToMinio(c.Request().Context(), videoID.String(), pr, done)

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if part.FormName() == "title" {
			buf := new(bytes.Buffer)
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

	err = h.publishEvent(c.Request().Context(), videoID, 1, title) // ! change user id

	if err != nil {
		c.String(http.StatusInternalServerError, "Can not upload video for processing")
	}

	return c.String(http.StatusOK, "File uploaded to MinIO!")
}

func (h *Handler) uploadToMinio(c context.Context, objectName string, reader io.Reader, done chan<- error) {
	err := h.mc.PutObject(
		c,
		objectName,
		reader,
	)

	done <- err
}

func (h *Handler) publishEvent(c context.Context, videoID uuid.UUID, userID int64, videoTitle string) error {
	event := VideoEvent{
		VideoID:    videoID,
		UserID:     userID,
		VideoTitle: videoTitle,
	}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err) // ! change
	}

	return h.q.Publish(c, body)
}
