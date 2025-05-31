package minio

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinioClientFromEnv() (*minio.Client, string, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	user := os.Getenv("MINIO_USER")
	pass := os.Getenv("MINIO_PASSWORD")
	bucket := os.Getenv("MINIO_BUCKET")
	sslStr := os.Getenv("MINIO_USE_SSL")

	if endpoint == "" || user == "" || pass == "" || bucket == "" {
		return nil, "", fmt.Errorf("не хватает переменных окружения для MinIO")
	}

	useSSL, err := strconv.ParseBool(sslStr)
	if err != nil {
		useSSL = false
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(user, pass, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, "", fmt.Errorf("ошибка подключения к MinIO: %w", err)
	}

	ctx := context.Background()
	found, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, "", err
	}
	if !found {
		err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, "", fmt.Errorf("ошибка создания bucket: %w", err)
		}
		log.Println("Создан новый bucket:", bucket)
	} else {
		log.Println("Bucket есть:", bucket)
	}

	return client, bucket, nil
}
