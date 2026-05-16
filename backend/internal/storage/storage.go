package storage

import (
	"context"
	"io"
	"time"
)

type StorageService interface {
	GeneratePresignedPutURL(ctx context.Context, key, contentType string, ttl time.Duration) (string, error)
	GeneratePresignedGetURL(ctx context.Context, key string, ttl time.Duration) (string, error)
	GetObject(ctx context.Context, key string) (io.ReadCloser, error)
	PutObject(ctx context.Context, key string, body io.Reader, contentType string) error
	DeleteObject(ctx context.Context, key string) error
	Ping(ctx context.Context) error
	CDNUrl(key string) string
}

func MimeTypeToExt(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".jpg"
	}
}
