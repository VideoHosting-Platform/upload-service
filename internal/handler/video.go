package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/VideoHosting-Platform/upload-service/pkg/logger"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

func (h *Handler) uploadVideo(c echo.Context) error {

	requestID := uuid.New().String()
	log := logger.WithSource("handler").With(
		"request_id", requestID,
		"method", c.Request().Method,
		"path", c.Request().URL.Path,
	)

	// metrics
	uploadRequests.Inc()
	timer := prometheus.NewTimer(uploadDuration) // Замер времени
	defer timer.ObserveDuration()

	reader, err := c.Request().MultipartReader()
	if err != nil {
		uploadFailedTotal.WithLabelValues(reasonInvalidFormat).Inc()
		log.Warn("invalid request format", "error", err)
		return c.String(http.StatusBadRequest, "Not a multipart request")
	}

	userIDstr := c.Request().Header.Get("X-User-Id")
	userID, err := strconv.ParseInt(userIDstr, 10, 64)
	if err != nil {
		log.Warn("invalid user id", "error", err)
		return c.String(http.StatusBadRequest, "Invalid user id")
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
		defer part.Close()

		switch part.FormName() {
		case "title":
			buf := new(bytes.Buffer)
			_, err := io.Copy(buf, part)
			if err != nil {
				uploadFailedTotal.WithLabelValues(reasonReadTitleError).Inc()
				log.Error("failed to read title", "error", err)
				return c.String(http.StatusInternalServerError, "read title error")
			}
			title = buf.String()
		case "video":
			_, err := io.Copy(pw, part)
			if err != nil {
				log.Error("failed to copy video data", "error", err)
				return c.String(http.StatusInternalServerError, "video copy error")
			}
		default:
			uploadFailedTotal.WithLabelValues(reasonInvalidFormat).Inc()
			log.Warn("invalid form field", "field", part.FormName())
			return c.String(http.StatusBadRequest, "not a valid form title")
		}
	}
	pw.Close()

	if err := <-done; err != nil {
		uploadFailedTotal.WithLabelValues(reasonMinioError).Inc()
		log.Error("failed to upload to MinIO", "error", err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("error while uploading to s3 - %s", err.Error()))
	}

	err = h.publishEvent(c.Request().Context(), videoID, userID, title)
	if err != nil {
		uploadFailedTotal.WithLabelValues(reasonPublishEventError).Inc()
		log.Error("failed to publish event", "error", err)
		c.String(http.StatusInternalServerError, "Can not upload video for processing")
	}

	uploadSuccessTotal.Inc()
	log.Info("video upload completed successfully")
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
