package storage

import (
	"context"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestR2Storage(now func() time.Time) *r2Storage {
	client := s3.New(s3.Options{
		Region:       "auto",
		BaseEndpoint: aws.String("https://fake.r2.cloudflarestorage.com"),
		Credentials:  credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
	})
	tel := observability.NewTelemetry(nil, nil, nil)
	presignURLDuration, _ := tel.Meter.Float64Histogram("r2.presigned_url.duration")
	return &r2Storage{
		client:             client,
		presign:            s3.NewPresignClient(client),
		bucket:             "test-bucket",
		tel:                tel,
		presignURLDuration: presignURLDuration,
		now:                now,
	}
}

func TestGenerateDeterministicPresignedGetURL_SameURLWithinWindow(t *testing.T) {
	// Two calls at different times within the same 24 h UTC window must produce identical URLs.
	windowStart := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	earlyInWindow := windowStart.Add(3 * time.Hour)
	lateInWindow := windowStart.Add(23*time.Hour + 59*time.Minute)

	storeEarly := newTestR2Storage(func() time.Time { return earlyInWindow })
	storeLate := newTestR2Storage(func() time.Time { return lateInWindow })

	url1, err := storeEarly.GenerateDeterministicPresignedGetURL(context.Background(), "users/u1/thumbnails/img1.jpg")
	require.NoError(t, err)

	url2, err := storeLate.GenerateDeterministicPresignedGetURL(context.Background(), "users/u1/thumbnails/img1.jpg")
	require.NoError(t, err)

	assert.Equal(t, url1, url2, "URLs within the same 24 h window must be identical")
}

func TestGenerateDeterministicPresignedGetURL_DifferentURLAtWindowRollover(t *testing.T) {
	// A call just before midnight UTC and one just after must produce different URLs.
	beforeMidnight := time.Date(2024, 6, 15, 23, 59, 59, 0, time.UTC)
	afterMidnight := time.Date(2024, 6, 16, 0, 0, 1, 0, time.UTC)

	storeBefore := newTestR2Storage(func() time.Time { return beforeMidnight })
	storeAfter := newTestR2Storage(func() time.Time { return afterMidnight })

	url1, err := storeBefore.GenerateDeterministicPresignedGetURL(context.Background(), "users/u1/thumbnails/img1.jpg")
	require.NoError(t, err)

	url2, err := storeAfter.GenerateDeterministicPresignedGetURL(context.Background(), "users/u1/thumbnails/img1.jpg")
	require.NoError(t, err)

	assert.NotEqual(t, url1, url2, "URLs across a window boundary must differ")
}

func TestGenerateDeterministicPresignedGetURL_ExpiryIsAtLeast48Hours(t *testing.T) {
	// The X-Amz-Expires query parameter must be >= 48 h (172800 seconds).
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	store := newTestR2Storage(func() time.Time { return now })

	rawURL, err := store.GenerateDeterministicPresignedGetURL(context.Background(), "users/u1/thumbnails/img1.jpg")
	require.NoError(t, err)

	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)

	expiresStr := parsed.Query().Get("X-Amz-Expires")
	require.NotEmpty(t, expiresStr, "presigned URL must include X-Amz-Expires")

	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, expires, int64(48*60*60), "X-Amz-Expires must be >= 172800 seconds (48 h)")
}
