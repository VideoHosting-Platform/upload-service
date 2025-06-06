package minio_connection

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint   string `env:"MINIO_ENDPOINT"`
	BucketName string `env:"MINIO_BUCKET_NAME"`
	AccessKey  string `env:"MINIO_ACESS_KEY"`
	SecretKey  string `env:"MINIO_SECRET_KEY"`
	UseSSL     bool   `env:"MINIO_USE_SSL" env-default:"false"`
}

type MinioClient struct {
	mc         *minio.Client
	bucketName string
}

func NewClient(cfg *Config) (*MinioClient, error) {

	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to MinIO error : %w", err)
	}
	_, err = mc.ListBuckets(context.Background())
	if err != nil {
		return nil, fmt.Errorf("connect to MinIO error : %w", err)
	}
	return &MinioClient{mc: mc, bucketName: cfg.BucketName}, err
}

func (m *MinioClient) PutObject(
	ctx context.Context,
	objectName string,
	reader io.Reader,
) error {
	_, err := m.mc.PutObject(
		ctx,
		m.bucketName,
		objectName,
		reader,
		-1,
		minio.PutObjectOptions{
			ContentType: "video/mp4",
		},
	)
	return err
}
