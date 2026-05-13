package minio

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"eventor/internal/platform/types"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const avatarBucket = "avatars"

type FileStorage struct {
	client *minio.Client
	host   string
}

func NewMinioStorage(
	endpoint string,
	accessKey string,
	secretKey string,
	useSSL bool,
) (*FileStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			accessKey,
			secretKey,
			"",
		),
		Secure: useSSL,
	})

	if err != nil {
		return nil, err
	}

	return &FileStorage{
		client: client,
		host:   endpoint,
	}, nil
}

func (storage *FileStorage) ensureBucket(ctx context.Context) error {
	exists, err := storage.client.BucketExists(ctx, avatarBucket)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return storage.client.MakeBucket(ctx, avatarBucket, minio.MakeBucketOptions{})
}

func (storage *FileStorage) UploadUserAvatar(ctx context.Context, userID types.IdType, content io.Reader, size int64, contentType string) (string, error) {
	if err := storage.ensureBucket(ctx); err != nil {
		return "", err
	}

	objectName := strconv.FormatUint(uint64(userID), 10)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := storage.client.PutObject(ctx, avatarBucket, objectName, content, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}

	return objectName, nil
}

func (storage *FileStorage) DeleteUserAvatar(ctx context.Context, userID types.IdType) error {
	if err := storage.ensureBucket(ctx); err != nil {
		return err
	}

	objectName := strconv.FormatUint(uint64(userID), 10)
	return storage.client.RemoveObject(ctx, avatarBucket, objectName, minio.RemoveObjectOptions{})
}

func (storage *FileStorage) GetUserAvatarURL(ctx context.Context, userID types.IdType, expiry time.Duration) (string, error) {
	if err := storage.ensureBucket(ctx); err != nil {
		return "", err
	}

	objectName := strconv.FormatUint(uint64(userID), 10)
	url, err := storage.client.PresignedGetObject(ctx, avatarBucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presign avatar url: %w", err)
	}

	return url.String(), nil
}
