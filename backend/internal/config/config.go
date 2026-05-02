package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type KindeConfig struct {
	IssuerURL string
	Audience  string
}

type DBConfig struct {
	URL string
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	PublicURL       string
}

type Config struct {
	Kinde KindeConfig
	DB    DBConfig
	R2    R2Config
	Port  string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		if os.IsNotExist(err) || errors.Is(err, os.ErrNotExist) {
			log.Printf("warning: .env file not found, using existing environment: %v", err)
		} else {
			return nil, fmt.Errorf("load .env: %w", err)
		}
	}

	return loadFromEnv()
}

func loadFromEnv() (*Config, error) {
	kindeIssuerURL, err := requireEnv("KINDE_ISSUER_URL")
	if err != nil {
		return nil, err
	}

	kindeAudience, err := requireEnv("KINDE_AUDIENCE")
	if err != nil {
		return nil, err
	}

	databaseURL, err := requireEnv("DATABASE_URL")
	if err != nil {
		return nil, err
	}

	r2AccountID, err := requireEnv("R2_ACCOUNT_ID")
	if err != nil {
		return nil, err
	}

	r2AccessKeyID, err := requireEnv("R2_ACCESS_KEY_ID")
	if err != nil {
		return nil, err
	}

	r2SecretAccessKey, err := requireEnv("R2_SECRET_ACCESS_KEY")
	if err != nil {
		return nil, err
	}

	r2BucketName, err := requireEnv("R2_BUCKET_NAME")
	if err != nil {
		return nil, err
	}

	r2PublicURL, err := requireEnv("R2_PUBLIC_URL")
	if err != nil {
		return nil, err
	}

	port := envWithDefault("PORT", "8080")

	return &Config{
		Kinde: KindeConfig{
			IssuerURL: kindeIssuerURL,
			Audience:  kindeAudience,
		},
		DB: DBConfig{
			URL: databaseURL,
		},
		R2: R2Config{
			AccountID:       r2AccountID,
			AccessKeyID:     r2AccessKeyID,
			SecretAccessKey: r2SecretAccessKey,
			BucketName:      r2BucketName,
			PublicURL:       r2PublicURL,
		},
		Port: port,
	}, nil
}

func requireEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}

func envWithDefault(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}
