package minio_connection

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint   string `yaml:"endpoint"`
	BucketName string `yaml:"bucket_name"`
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
	UseSSL     bool   `yaml:"use_ssl"`
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
