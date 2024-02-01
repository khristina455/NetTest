package minio

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
	"mime/multipart"
	"path/filepath"
)

type Minio struct {
	Client     *minio.Client
	Host       string
	BucketName string
}

type Client interface {
	SaveImage(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
	DeleteImage(ctx context.Context, objectName string) error
}

type MinioConfig struct {
	Host            string
	BucketName      string
	AccessKeyID     string
	SecretAccessKey string
	Location        string
}

func InitConfig() MinioConfig {
	viper.AddConfigPath("config")
	viper.SetConfigName("config")

	config := MinioConfig{
		Host:            viper.GetString("minio.host"),
		BucketName:      viper.GetString("minio.bucketName"),
		AccessKeyID:     viper.GetString("minio.accessKeyId"),
		SecretAccessKey: viper.GetString("minio.SecretAccessKey"),
		Location:        viper.GetString("location"),
	}

	return config
}

func NewMinioClient(ctx context.Context, config MinioConfig) (Client, error) {
	minioClient, err := minio.New(config.Host, &minio.Options{Creds: credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, "")})
	if err != nil {
		return nil, err
	}

	if err = minioClient.MakeBucket(ctx, config.BucketName, minio.MakeBucketOptions{Region: config.Location}); err != nil {
		fmt.Println(err)
	}

	return &Minio{Client: minioClient, BucketName: config.BucketName, Host: config.Host}, nil
}

func (m *Minio) SaveImage(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	objectName := uuid.New().String() + filepath.Ext(header.Filename)

	if _, err := m.Client.PutObject(ctx, m.BucketName, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: header.Header.Get("Content-Type"),
	}); err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s/%s/%s", m.Host, m.BucketName, objectName), nil
}

func (m *Minio) DeleteImage(ctx context.Context, objectName string) error {
	return m.Client.RemoveObject(ctx, m.BucketName, objectName, minio.RemoveObjectOptions{})
}
