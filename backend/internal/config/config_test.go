package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setRequiredEnvVars(t *testing.T) {
	t.Helper()
	t.Setenv("KINDE_ISSUER_URL", "https://example.kinde.com")
	t.Setenv("KINDE_AUDIENCE", "bookleaf-api")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/bookleaf")
	t.Setenv("R2_ACCOUNT_ID", "account-id")
	t.Setenv("R2_ACCESS_KEY_ID", "access-key-id")
	t.Setenv("R2_SECRET_ACCESS_KEY", "secret-access-key")
	t.Setenv("R2_BUCKET_NAME", "bucket-name")
	t.Setenv("R2_PUBLIC_URL", "https://assets.bookleaf.app")
	t.Setenv("OTEL_EXPORTER", "jaeger")
	t.Setenv("OTEL_METRICS_EXPORTER", "prometheus")
}

func TestLoad_AllRequiredVarsSet(t *testing.T) {
	t.Chdir(t.TempDir())
	setRequiredEnvVars(t)
	t.Setenv("PORT", "9090")

	cfg, err := Load()

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "https://example.kinde.com", cfg.Kinde.IssuerURL)
	assert.Equal(t, "bookleaf-api", cfg.Kinde.Audience)
	assert.Equal(t, "postgres://user:pass@localhost:5432/bookleaf", cfg.DB.URL)
	assert.Equal(t, "account-id", cfg.R2.AccountID)
	assert.Equal(t, "access-key-id", cfg.R2.AccessKeyID)
	assert.Equal(t, "secret-access-key", cfg.R2.SecretAccessKey)
	assert.Equal(t, "bucket-name", cfg.R2.BucketName)
	assert.Equal(t, "https://assets.bookleaf.app", cfg.R2.PublicURL)
	assert.Equal(t, "jaeger", cfg.Obs.OTELExporter)
	assert.Equal(t, "prometheus", cfg.Obs.OTELMetricsExporter)
	assert.Equal(t, "9090", cfg.Port)
}

func TestLoad_MissingRequiredVar(t *testing.T) {
	tests := []struct {
		name       string
		missingVar string
	}{
		{"missing KINDE_ISSUER_URL", "KINDE_ISSUER_URL"},
		{"missing KINDE_AUDIENCE", "KINDE_AUDIENCE"},
		{"missing DATABASE_URL", "DATABASE_URL"},
		{"missing R2_ACCOUNT_ID", "R2_ACCOUNT_ID"},
		{"missing R2_ACCESS_KEY_ID", "R2_ACCESS_KEY_ID"},
		{"missing R2_SECRET_ACCESS_KEY", "R2_SECRET_ACCESS_KEY"},
		{"missing R2_BUCKET_NAME", "R2_BUCKET_NAME"},
		{"missing R2_PUBLIC_URL", "R2_PUBLIC_URL"},
		{"missing OTEL_EXPORTER", "OTEL_EXPORTER"},
		{"missing OTEL_METRICS_EXPORTER", "OTEL_METRICS_EXPORTER"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			setRequiredEnvVars(t)
			t.Setenv("PORT", "8080")
			t.Setenv(tt.missingVar, "")

			cfg, err := Load()

			require.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), tt.missingVar)
		})
	}
}

func TestLoad_LogFormat(t *testing.T) {
	tests := []struct {
		name       string
		logFormat  *string
		wantFormat string
	}{
		{"explicit log format", func() *string { s := "json"; return &s }(), "json"},
		{"unset defaults to console", nil, "console"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			setRequiredEnvVars(t)
			if tt.logFormat != nil {
				t.Setenv("LOG_FORMAT", *tt.logFormat)
			}

			cfg, err := Load()

			require.NoError(t, err)
			assert.Equal(t, tt.wantFormat, cfg.Obs.LogFormat)
		})
	}
}

func TestLoad_Port(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name     string
		portEnv  *string // nil = unset, non-nil = set to value (empty string = empty)
		wantPort string
	}{
		{"explicit port value", strPtr("3000"), "3000"},
		{"empty port defaults to 8080", strPtr(""), "8080"},
		{"unset port defaults to 8080", nil, "8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			setRequiredEnvVars(t)
			if tt.portEnv != nil {
				t.Setenv("PORT", *tt.portEnv)
			}

			cfg, err := Load()

			require.NoError(t, err)
			assert.Equal(t, tt.wantPort, cfg.Port)
		})
	}
}
