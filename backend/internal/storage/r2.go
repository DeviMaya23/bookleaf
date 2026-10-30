package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/devi/bookleaf/internal/platform/config"
	"github.com/devi/bookleaf/internal/platform/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

const (
	thumbnailPresignWindowSize = 24 * time.Hour
	thumbnailPresignExpiry     = 48 * time.Hour
)

type r2Storage struct {
	client             *s3.Client
	presign            *s3.PresignClient
	bucket             string
	tel                *observability.Telemetry
	presignURLDuration metric.Float64Histogram
	now                func() time.Time
}

// deterministicSigner replaces the signing time chosen by the SDK with a fixed
// one so that presigned URLs are identical for every request in the same window.
type deterministicSigner struct {
	inner       *v4.Signer
	signingTime time.Time
}

func (d deterministicSigner) PresignHTTP(
	ctx context.Context, credentials aws.Credentials, r *http.Request,
	payloadHash string, service string, region string, _ time.Time,
	optFns ...func(*v4.SignerOptions),
) (string, http.Header, error) {
	return d.inner.PresignHTTP(ctx, credentials, r, payloadHash, service, region, d.signingTime, optFns...)
}

func NewR2Storage(cfg config.R2Config, tel *observability.Telemetry) *r2Storage {
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
		tel:                tel,
		presignURLDuration: presignURLDuration,
		now:                time.Now,
	}
}

func (r *r2Storage) GeneratePresignedPutURL(ctx context.Context, key, contentType string, contentLength int64, ttl time.Duration) (string, error) {
	ctx, span := r.tel.Tracer.Start(ctx, "storage.GeneratePresignedPutURL")
	defer span.End()

	logger := observability.LoggerFromContext(ctx, r.tel.Logger)
	start := time.Now()

	input := &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}
	if contentLength > 0 {
		input.ContentLength = aws.Int64(contentLength)
	}

	resp, err := r.presign.PresignPutObject(ctx, input, s3.WithPresignExpires(ttl))

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

// GenerateDeterministicPresignedGetURL signs with the start of the current UTC
// window instead of the current time, so the URL is stable for the whole window.
func (r *r2Storage) GenerateDeterministicPresignedGetURL(ctx context.Context, key string) (string, error) {
	ctx, span := r.tel.Tracer.Start(ctx, "storage.GenerateDeterministicPresignedGetURL")
	defer span.End()

	logger := observability.LoggerFromContext(ctx, r.tel.Logger)
	start := time.Now()

	windowStart := r.now().UTC().Truncate(thumbnailPresignWindowSize)
	signer := deterministicSigner{inner: v4.NewSigner(), signingTime: windowStart}

	resp, err := r.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(thumbnailPresignExpiry), func(o *s3.PresignOptions) {
		o.Presigner = signer
	})

	status := "success"
	if err != nil {
		status = "error"
	}
	r.presignURLDuration.Record(ctx, float64(time.Since(start).Milliseconds()),
		metric.WithAttributes(
			attribute.String("r2.operation", "presigned_get_deterministic"),
			attribute.String("r2.status", status),
		),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error("deterministic presigned get URL generation failed",
			zap.String("event", "r2.presigned_get_deterministic.failed"),
			zap.String("r2_key", key),
			zap.Error(err),
		)
		return "", fmt.Errorf("presign deterministic get %s: %w", key, err)
	}

	return resp.URL, nil
}

func (r *r2Storage) GeneratePresignedDownloadURL(ctx context.Context, key, filename string, ttl time.Duration) (string, error) {
	ctx, span := r.tel.Tracer.Start(ctx, "storage.GeneratePresignedDownloadURL")
	defer span.End()

	logger := observability.LoggerFromContext(ctx, r.tel.Logger)
	start := time.Now()

	resp, err := r.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(r.bucket),
		Key:                        aws.String(key),
		ResponseContentDisposition: aws.String(fmt.Sprintf("attachment; filename=\"%s\"", filename)),
	}, s3.WithPresignExpires(ttl))

	status := "success"
	if err != nil {
		status = "error"
	}
	r.presignURLDuration.Record(ctx, float64(time.Since(start).Milliseconds()),
		metric.WithAttributes(
			attribute.String("r2.operation", "presigned_download"),
			attribute.String("r2.status", status),
		),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error("presigned download URL generation failed",
			zap.String("event", "r2.presigned_download.failed"),
			zap.String("r2_key", key),
			zap.String("filename", filename),
			zap.Error(err),
		)
		return "", fmt.Errorf("presign download %s: %w", key, err)
	}

	logger.Info("presigned download URL generated",
		zap.String("event", "r2.presigned_download.success"),
		zap.String("r2_key", key),
		zap.String("filename", filename),
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

func (r *r2Storage) PutObject(ctx context.Context, key string, body io.Reader, contentType, cacheControl string) error {
	ctx, span := r.tel.Tracer.Start(ctx, "storage.PutObject")
	defer span.End()

	input := &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	}
	if cacheControl != "" {
		input.CacheControl = aws.String(cacheControl)
	}
	_, err := r.client.PutObject(ctx, input)
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
