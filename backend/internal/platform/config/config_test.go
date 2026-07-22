package config

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setRequiredEnvVars(t *testing.T) {
	t.Helper()
	t.Setenv("KINDE_ISSUER_URL", "https://example.kinde.com")
	t.Setenv("KINDE_AUDIENCE", "bookleaf-api")
	t.Setenv("KINDE_M2M_CLIENT_ID", "m2m-client-id")
	t.Setenv("KINDE_M2M_CLIENT_SECRET", "m2m-client-secret")
	t.Setenv("KINDE_M2M_TOKEN_URL", "https://example.kinde.com/oauth2/token")
	t.Setenv("KINDE_MANAGEMENT_AUDIENCE", "https://example.kinde.com/api")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,https://app.example.com")
	t.Setenv("DATABASE_HOST", "localhost")
	t.Setenv("DATABASE_NAME", "bookleaf")
	t.Setenv("DATABASE_PORT", "5432")
	t.Setenv("DATABASE_USER", "user")
	t.Setenv("DATABASE_PASSWORD", "pass")
	t.Setenv("DATABASE_SSLMODE", "disable")
	t.Setenv("R2_ACCOUNT_ID", "account-id")
	t.Setenv("R2_ACCESS_KEY_ID", "access-key-id")
	t.Setenv("R2_SECRET_ACCESS_KEY", "secret-access-key")
	t.Setenv("R2_BUCKET_NAME", "bucket-name")
	t.Setenv("OTEL_EXPORTER", "jaeger")
	t.Setenv("OTEL_METRICS_EXPORTER", "prometheus")
	t.Setenv("INTERNAL_API_SECRET", "nyaa")
}

func TestLoad_AllRequiredVarsSet(t *testing.T) {
	t.Chdir(t.TempDir())
	setRequiredEnvVars(t)
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("GOOGLE_VISION_API_KEY", "abc123")
	t.Setenv("PORT", "9090")

	cfg, err := Load()

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "https://example.kinde.com", cfg.Kinde.IssuerURL)
	assert.Equal(t, "bookleaf-api", cfg.Kinde.Audience)
	assert.Equal(t, "m2m-client-id", cfg.Kinde.M2MClientID)
	assert.Equal(t, "m2m-client-secret", cfg.Kinde.M2MClientSecret)
	assert.Equal(t, "https://example.kinde.com/oauth2/token", cfg.Kinde.M2MTokenURL)
	assert.Equal(t, "https://example.kinde.com/api", cfg.Kinde.ManagementAudience)
	assert.Equal(t, []string{"http://localhost:5173", "https://app.example.com"}, cfg.CORSAllowedOrigins)
	assert.Equal(t, "postgres://user:pass@localhost:5432/bookleaf?sslmode=disable", cfg.DB.URL)
	assert.Equal(t, "localhost", cfg.DB.Host)
	assert.Equal(t, "bookleaf", cfg.DB.Name)
	assert.Equal(t, "5432", cfg.DB.Port)
	assert.Equal(t, "user", cfg.DB.User)
	assert.Equal(t, "pass", cfg.DB.Password)
	assert.Equal(t, "disable", cfg.DB.SSLMode)
	assert.Equal(t, "account-id", cfg.R2.AccountID)
	assert.Equal(t, "access-key-id", cfg.R2.AccessKeyID)
	assert.Equal(t, "secret-access-key", cfg.R2.SecretAccessKey)
	assert.Equal(t, "bucket-name", cfg.R2.BucketName)
	assert.True(t, cfg.Obs.OTELEnabled)
	assert.Equal(t, "jaeger", cfg.Obs.OTELExporter)
	assert.Equal(t, "prometheus", cfg.Obs.OTELMetricsExporter)
	assert.Equal(t, 0.1, cfg.Obs.SampleRatio)
	assert.Equal(t, "abc123", cfg.Vision.APIKey)
	assert.Equal(t, "9090", cfg.Port)
}

func TestLoad_VisionAPIKeyUnset(t *testing.T) {
	t.Chdir(t.TempDir())
	setRequiredEnvVars(t)

	cfg, err := Load()

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "", cfg.Vision.APIKey)
}

func TestLoad_MissingRequiredVar(t *testing.T) {
	tests := []struct {
		name       string
		missingVar string
	}{
		{"missing KINDE_ISSUER_URL", "KINDE_ISSUER_URL"},
		{"missing KINDE_AUDIENCE", "KINDE_AUDIENCE"},
		{"missing KINDE_M2M_CLIENT_ID", "KINDE_M2M_CLIENT_ID"},
		{"missing KINDE_M2M_CLIENT_SECRET", "KINDE_M2M_CLIENT_SECRET"},
		{"missing KINDE_M2M_TOKEN_URL", "KINDE_M2M_TOKEN_URL"},
		{"missing KINDE_MANAGEMENT_AUDIENCE", "KINDE_MANAGEMENT_AUDIENCE"},
		{"missing CORS_ALLOWED_ORIGINS", "CORS_ALLOWED_ORIGINS"},
		{"missing DATABASE_HOST", "DATABASE_HOST"},
		{"missing DATABASE_NAME", "DATABASE_NAME"},
		{"missing DATABASE_PORT", "DATABASE_PORT"},
		{"missing DATABASE_USER", "DATABASE_USER"},
		{"missing DATABASE_PASSWORD", "DATABASE_PASSWORD"},
		{"missing DATABASE_SSLMODE", "DATABASE_SSLMODE"},
		{"missing R2_ACCOUNT_ID", "R2_ACCOUNT_ID"},
		{"missing R2_ACCESS_KEY_ID", "R2_ACCESS_KEY_ID"},
		{"missing R2_SECRET_ACCESS_KEY", "R2_SECRET_ACCESS_KEY"},
		{"missing R2_BUCKET_NAME", "R2_BUCKET_NAME"},
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

func TestLoad_OTELDefaultsDisabledWhenUnset(t *testing.T) {
	t.Chdir(t.TempDir())
	setRequiredEnvVars(t)

	cfg, err := Load()

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.False(t, cfg.Obs.OTELEnabled)
}

func TestLoad_OTELDisabledAllowsMissingExporter(t *testing.T) {
	t.Chdir(t.TempDir())
	setRequiredEnvVars(t)
	t.Setenv("OTEL_ENABLED", "false")
	t.Setenv("OTEL_EXPORTER", "")

	cfg, err := Load()

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "", cfg.Obs.OTELExporter)
}

func TestLoad_OTELEnabledRequiresExporter(t *testing.T) {
	t.Chdir(t.TempDir())
	setRequiredEnvVars(t)
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_EXPORTER", "")

	cfg, err := Load()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "OTEL_EXPORTER")
}

func TestLoad_OTELEnabledRequiresMetricsExporter(t *testing.T) {
	t.Chdir(t.TempDir())
	setRequiredEnvVars(t)
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_METRICS_EXPORTER", "")

	cfg, err := Load()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "OTEL_METRICS_EXPORTER")
}

func TestLoad_OTELEnabledWithExportersSet(t *testing.T) {
	t.Chdir(t.TempDir())
	setRequiredEnvVars(t)
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_EXPORTER", "tempo")
	t.Setenv("OTEL_METRICS_EXPORTER", "prometheus")

	cfg, err := Load()

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.True(t, cfg.Obs.OTELEnabled)
	assert.Equal(t, "tempo", cfg.Obs.OTELExporter)
	assert.Equal(t, "prometheus", cfg.Obs.OTELMetricsExporter)
}

func TestLoad_SampleRatio(t *testing.T) {
	tests := []struct {
		name      string
		ratioEnv  *string
		wantRatio float64
		wantErr   bool
	}{
		{"unset defaults to 0.1", nil, 0.1, false},
		{"explicit value", func() *string { s := "0.5"; return &s }(), 0.5, false},
		{"invalid value returns error", func() *string { s := "not-a-float"; return &s }(), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			setRequiredEnvVars(t)
			if tt.ratioEnv != nil {
				t.Setenv("OTEL_SAMPLE_RATIO", *tt.ratioEnv)
			}

			cfg, err := Load()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "OTEL_SAMPLE_RATIO")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantRatio, cfg.Obs.SampleRatio)
		})
	}
}

func TestLoad_MaintenanceMode(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name        string
		modeEnv     *string
		wantEnabled bool
	}{
		{"unset defaults to disabled", nil, false},
		{"true enables maintenance mode", strPtr("true"), true},
		{"invalid value defaults to disabled", strPtr("notabool"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			setRequiredEnvVars(t)
			if tt.modeEnv != nil {
				t.Setenv("MAINTENANCE_MODE", *tt.modeEnv)
			}

			cfg, err := Load()

			require.NoError(t, err)
			require.NotNil(t, cfg)
			assert.Equal(t, tt.wantEnabled, cfg.Maintenance.Enabled)
		})
	}
}

func TestLoad_MaintenanceBypassTokenUnsetDefaultsToEmpty(t *testing.T) {
	t.Chdir(t.TempDir())
	setRequiredEnvVars(t)

	cfg, err := Load()

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "", cfg.Maintenance.BypassToken)
}

func TestLoad_LogFormat(t *testing.T) {
	tests := []struct {
		name       string
		logFormat  *string
		wantFormat string
	}{
		{"explicit log format", func() *string { s := "json"; return &s }(), "json"},
		{"unset defaults to json", nil, "json"},
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

func TestLoad_DatabaseOptions(t *testing.T) {
	t.Chdir(t.TempDir())
	setRequiredEnvVars(t)
	t.Setenv("DATABASE_OPTIONS", "connect_timeout=10&application_name=bookleaf")

	cfg, err := Load()

	require.NoError(t, err)
	require.NotNil(t, cfg)

	parsed, parseErr := url.Parse(cfg.DB.URL)
	require.NoError(t, parseErr)
	query := parsed.Query()
	assert.Equal(t, "disable", query.Get("sslmode"))
	assert.Equal(t, "10", query.Get("connect_timeout"))
	assert.Equal(t, "bookleaf", query.Get("application_name"))
}

func TestLoad_InvalidDatabaseOptions(t *testing.T) {
	t.Chdir(t.TempDir())
	setRequiredEnvVars(t)
	t.Setenv("DATABASE_OPTIONS", "%")

	cfg, err := Load()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "DATABASE_OPTIONS")
}
