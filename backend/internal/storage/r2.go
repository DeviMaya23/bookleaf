package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/devi/bookleaf/internal/config"
	"github.com/devi/bookleaf/internal/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

type r2Storage struct {
	client             *s3.Client
	presign            *s3.PresignClient
	bucket             string
	publicURL          string
	tel                *observability.Telemetry
	presignURLDuration metric.Float64Histogram
}

func NewR2Storage(cfg config.R2Config, tel *observability.Telemetry) StorageService {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	client := s3.New(s3.Options{
		Region:       "auto",
		BaseEndpoint: aws.String(endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	})

	presignURLDuration, _ := tel.Meter.Float64Histogram(
		"r2.presigned_url.duration",
		metric.WithUnit("ms"),
		metric.WithDescription("Duration of R2 presigned URL generation in milliseconds"),
	)

	return &r2Storage{
		client:             client,
		presign:            s3.NewPresignClient(client),
		bucket:             cfg.BucketName,
		publicURL:          cfg.PublicURL,
		tel:                tel,
		presignURLDuration: presignURLDuration,
	}
}

func (r *r2Storage) CDNUrl(key string) string {
	return r.publicURL + "/" + key
}

func (r *r2Storage) GeneratePresignedPutURL(ctx context.Context, key, contentType string, ttl time.Duration) (string, error) {
	ctx, span := r.tel.Tracer.Start(ctx, "storage.GeneratePresignedPutURL")
	defer span.End()

	logger := observability.LoggerFromContext(ctx, r.tel.Logger)
	start := time.Now()

	resp, err := r.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(ttl))

	status := "success"
	if err != nil {
		status = "error"
	}
	r.presignURLDuration.Record(ctx, float64(time.Since(start).Milliseconds()),
		metric.WithAttributes(
			attribute.String("r2.operation", "presigned_put"),
			attribute.String("r2.status", status),
		),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error("presigned put URL generation failed",
			zap.String("event", "r2.presigned_put.failed"),
			zap.String("r2_key", key),
			zap.Error(err),
		)
		return "", fmt.Errorf("presign put %s: %w", key, err)
	}

	logger.Info("presigned put URL generated",
		zap.String("event", "r2.presigned_put.success"),
		zap.String("r2_key", key),
	)
	return resp.URL, nil
}

func (r *r2Storage) GeneratePresignedGetURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	ctx, span := r.tel.Tracer.Start(ctx, "storage.GeneratePresignedGetURL")
	defer span.End()

	logger := observability.LoggerFromContext(ctx, r.tel.Logger)
	start := time.Now()

	resp, err := r.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))

	status := "success"
	if err != nil {
		status = "error"
	}
	r.presignURLDuration.Record(ctx, float64(time.Since(start).Milliseconds()),
		metric.WithAttributes(
			attribute.String("r2.operation", "presigned_get"),
			attribute.String("r2.status", status),
		),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error("presigned get URL generation failed",
			zap.String("event", "r2.presigned_get.failed"),
			zap.String("r2_key", key),
			zap.Error(err),
		)
		return "", fmt.Errorf("presign get %s: %w", key, err)
	}

	logger.Info("presigned get URL generated",
		zap.String("event", "r2.presigned_get.success"),
		zap.String("r2_key", key),
	)
	return resp.URL, nil
}

func (r *r2Storage) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	ctx, span := r.tel.Tracer.Start(ctx, "storage.GetObject")
	defer span.End()

	resp, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("get object %s: %w", key, err)
	}
	return resp.Body, nil
}

func (r *r2Storage) PutObject(ctx context.Context, key string, body io.Reader, contentType string) error {
	ctx, span := r.tel.Tracer.Start(ctx, "storage.PutObject")
	defer span.End()

	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("put object %s: %w", key, err)
	}
	return nil
}

func (r *r2Storage) DeleteObject(ctx context.Context, key string) error {
	ctx, span := r.tel.Tracer.Start(ctx, "storage.DeleteObject")
	defer span.End()

	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("delete object %s: %w", key, err)
	}
	return nil
}

func (r *r2Storage) Ping(ctx context.Context) error {
	_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(r.bucket),
	})
	if err != nil {
		return fmt.Errorf("head bucket %s: %w", r.bucket, err)
	}
	return nil
}

var _ StorageService = (*r2Storage)(nil)
