package usecase

import (
	"context"
	"io"
	"time"
)

type StorageService interface {
	GeneratePresignedPutURL(ctx context.Context, key, contentType string, ttl time.Duration) (string, error)
	GeneratePresignedGetURL(ctx context.Context, key string, ttl time.Duration) (string, error)
	GeneratePresignedDownloadURL(ctx context.Context, key, filename string, ttl time.Duration) (string, error)
	GetObject(ctx context.Context, key string) (io.ReadCloser, error)
PutObject(ctx context.Context, key string, body io.Reader, contentType string) error
	DeleteObject(ctx context.Context, key string) error
	Ping(ctx context.Context) error
}
