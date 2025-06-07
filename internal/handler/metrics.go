package handler

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// upload metrics
	uploadRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "upload_service_upload_total",
		Help: "Total video upload requests",
	})
	uploadSuccessTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "upload_service_success_total",
		Help: "Total video successfuly uploaded",
	})
	uploadFailedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "upload_service_failed_total",
			Help: "Total video upload failed",
		},
		[]string{"reason"},
	)
	uploadDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "upload_service_upload_duration_seconds",
		Help:    "Duration of video upload processing.",
		Buckets: []float64{0.1, 0.5, 1, 5, 10},
	})
)
