package minio

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"time"

	"eventor/internal/platform/types"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const avatarBucket = "avatars"
const eventImageBucket = "event-images"

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
	return storage.ensureBucketNamed(ctx, avatarBucket)
}

func (storage *FileStorage) ensureBucketNamed(ctx context.Context, bucketName string) error {
	exists, err := storage.client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return storage.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
}

func (storage *FileStorage) UploadUserAvatar(ctx context.Context, userID types.IdType, content io.Reader, size int64, contentType string) (string, error) {
	if err := storage.ensureBucketNamed(ctx, avatarBucket); err != nil {
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

func (storage *FileStorage) UploadEventImage(ctx context.Context, content io.Reader, size int64, contentType string) (string, error) {
	if err := storage.ensureBucketNamed(ctx, eventImageBucket); err != nil {
		return "", err
	}

	objectName, err := randomStorageKey("event-")
	if err != nil {
		return "", err
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = storage.client.PutObject(ctx, eventImageBucket, objectName, content, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}

	return objectName, nil
}

func (storage *FileStorage) DeleteUserAvatar(ctx context.Context, userID types.IdType) error {
	if err := storage.ensureBucketNamed(ctx, avatarBucket); err != nil {
		return err
	}

	objectName := strconv.FormatUint(uint64(userID), 10)
	return storage.client.RemoveObject(ctx, avatarBucket, objectName, minio.RemoveObjectOptions{})
}

func (storage *FileStorage) DeleteEventImage(ctx context.Context, imageID string) error {
	if err := storage.ensureBucketNamed(ctx, eventImageBucket); err != nil {
		return err
	}

	return storage.client.RemoveObject(ctx, eventImageBucket, imageID, minio.RemoveObjectOptions{})
}

func (storage *FileStorage) GetUserAvatarURL(ctx context.Context, userID types.IdType, expiry time.Duration) (string, error) {
	if err := storage.ensureBucketNamed(ctx, avatarBucket); err != nil {
		return "", err
	}

	objectName := strconv.FormatUint(uint64(userID), 10)
	url, err := storage.client.PresignedGetObject(ctx, avatarBucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presign avatar url: %w", err)
	}

	return url.String(), nil
}

func (storage *FileStorage) OpenEventImage(ctx context.Context, imageID string) (io.ReadCloser, string, error) {
	if err := storage.ensureBucketNamed(ctx, eventImageBucket); err != nil {
		return nil, "", err
	}

	stat, err := storage.client.StatObject(ctx, eventImageBucket, imageID, minio.StatObjectOptions{})
	if err != nil {
		return nil, "", err
	}

	object, err := storage.client.GetObject(ctx, eventImageBucket, imageID, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", err
	}

	contentType := stat.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return object, contentType, nil
}

func (storage *FileStorage) OpenUserAvatar(ctx context.Context, userID types.IdType) (io.ReadCloser, string, error) {
	if err := storage.ensureBucketNamed(ctx, avatarBucket); err != nil {
		return nil, "", err
	}

	objectName := strconv.FormatUint(uint64(userID), 10)
	stat, err := storage.client.StatObject(ctx, avatarBucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, "", err
	}

	object, err := storage.client.GetObject(ctx, avatarBucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", err
	}

	contentType := stat.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return object, contentType, nil
}

func randomStorageKey(prefix string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(bytes), nil
}
