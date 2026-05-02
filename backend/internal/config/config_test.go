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
