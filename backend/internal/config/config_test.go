package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoad_AllRequiredVarsSet(t *testing.T) {
	t.Chdir(t.TempDir())

	t.Setenv("KINDE_ISSUER_URL", "https://example.kinde.com")
	t.Setenv("KINDE_AUDIENCE", "bookleaf-api")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/bookleaf")
	t.Setenv("PORT", "9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	if cfg.Kinde.IssuerURL != "https://example.kinde.com" {
		t.Fatalf("expected kinde issuer URL to be set, got: %s", cfg.Kinde.IssuerURL)
	}
	if cfg.Kinde.Audience != "bookleaf-api" {
		t.Fatalf("expected kinde audience to be set, got: %s", cfg.Kinde.Audience)
	}
	if cfg.DB.URL != "postgres://user:pass@localhost:5432/bookleaf" {
		t.Fatalf("expected database URL to be set, got: %s", cfg.DB.URL)
	}
	if cfg.Port != "9090" {
		t.Fatalf("expected port 9090, got: %s", cfg.Port)
	}
}

func TestLoad_MissingRequiredVar(t *testing.T) {
	requiredVars := []string{
		"KINDE_ISSUER_URL",
		"KINDE_AUDIENCE",
		"DATABASE_URL",
	}

	for _, missingVar := range requiredVars {
		t.Run(missingVar, func(t *testing.T) {
			t.Chdir(t.TempDir())

			t.Setenv("KINDE_ISSUER_URL", "https://example.kinde.com")
			t.Setenv("KINDE_AUDIENCE", "bookleaf-api")
			t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/bookleaf")
			t.Setenv("PORT", "8080")
			t.Setenv(missingVar, "")

			cfg, err := Load()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if cfg != nil {
				t.Fatal("expected nil config when required var is missing")
			}
			if !strings.Contains(err.Error(), missingVar) {
				t.Fatalf("expected error to mention %s, got: %v", missingVar, err)
			}
		})
	}
}

func TestLoad_PortDefaultsWhenUnset(t *testing.T) {
	t.Chdir(t.TempDir())

	t.Setenv("KINDE_ISSUER_URL", "https://example.kinde.com")
	t.Setenv("KINDE_AUDIENCE", "bookleaf-api")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/bookleaf")
	unsetEnv(t, "PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got: %s", cfg.Port)
	}
}

func TestLoad_PortUsesExplicitValue(t *testing.T) {
	t.Chdir(t.TempDir())

	t.Setenv("KINDE_ISSUER_URL", "https://example.kinde.com")
	t.Setenv("KINDE_AUDIENCE", "bookleaf-api")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/bookleaf")
	t.Setenv("PORT", "3000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Port != "3000" {
		t.Fatalf("expected port 3000, got: %s", cfg.Port)
	}
}

func TestLoad_PortDefaultsWhenEmpty(t *testing.T) {
	t.Chdir(t.TempDir())

	t.Setenv("KINDE_ISSUER_URL", "https://example.kinde.com")
	t.Setenv("KINDE_AUDIENCE", "bookleaf-api")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/bookleaf")
	t.Setenv("PORT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got: %s", cfg.Port)
	}
}

func unsetEnv(t *testing.T, name string) {
	t.Helper()

	original, hadOriginal := os.LookupEnv(name)
	if err := os.Unsetenv(name); err != nil {
		t.Fatalf("unset %s: %v", name, err)
	}

	t.Cleanup(func() {
		var err error
		if hadOriginal {
			err = os.Setenv(name, original)
		} else {
			err = os.Unsetenv(name)
		}
		if err != nil {
			t.Fatalf("restore %s: %v", name, err)
		}
	})
}
