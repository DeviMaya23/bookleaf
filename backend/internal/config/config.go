package config

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type KindeConfig struct {
	IssuerURL string
	Audience  string
}

type DBConfig struct {
	Host     string
	Name     string
	Port     string
	User     string
	Password string
	SSLMode  string
	URL      string
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	PublicURL       string
}

type ObsConfig struct {
	OTELExporter        string
	OTELMetricsExporter string
	LogFormat           string
}

type VisionConfig struct {
	APIKey string
}

type Config struct {
	Kinde KindeConfig
	DB    DBConfig
	R2    R2Config
	Obs   ObsConfig
	Vision VisionConfig
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

	databaseHost, err := requireEnv("DATABASE_HOST")
	if err != nil {
		return nil, err
	}

	databaseName, err := requireEnv("DATABASE_NAME")
	if err != nil {
		return nil, err
	}

	databasePort, err := requireEnv("DATABASE_PORT")
	if err != nil {
		return nil, err
	}

	databaseUser, err := requireEnv("DATABASE_USER")
	if err != nil {
		return nil, err
	}

	databasePassword, err := requireEnv("DATABASE_PASSWORD")
	if err != nil {
		return nil, err
	}

	databaseSSLMode, err := requireEnv("DATABASE_SSLMODE")
	if err != nil {
		return nil, err
	}

	databaseOptions := os.Getenv("DATABASE_OPTIONS")
	databaseURL, err := buildDatabaseURL(DBConfig{
		Host:     databaseHost,
		Name:     databaseName,
		Port:     databasePort,
		User:     databaseUser,
		Password: databasePassword,
		SSLMode:  databaseSSLMode,
	}, databaseOptions)
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

	otelExporter, err := requireEnv("OTEL_EXPORTER")
	if err != nil {
		return nil, err
	}

	otelMetricsExporter, err := requireEnv("OTEL_METRICS_EXPORTER")
	if err != nil {
		return nil, err
	}

	logFormat := envWithDefault("LOG_FORMAT", "json")
	visionAPIKey := envWithDefault("GOOGLE_VISION_API_KEY", "")
	port := envWithDefault("PORT", "8080")

	return &Config{
		Kinde: KindeConfig{
			IssuerURL: kindeIssuerURL,
			Audience:  kindeAudience,
		},
		DB: DBConfig{
			Host:     databaseHost,
			Name:     databaseName,
			Port:     databasePort,
			User:     databaseUser,
			Password: databasePassword,
			SSLMode:  databaseSSLMode,
			URL:      databaseURL,
		},
		R2: R2Config{
			AccountID:       r2AccountID,
			AccessKeyID:     r2AccessKeyID,
			SecretAccessKey: r2SecretAccessKey,
			BucketName:      r2BucketName,
			PublicURL:       r2PublicURL,
		},
		Obs: ObsConfig{
			OTELExporter:        otelExporter,
			OTELMetricsExporter: otelMetricsExporter,
			LogFormat:           logFormat,
		},
		Vision: VisionConfig{
			APIKey: visionAPIKey,
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

func buildDatabaseURL(cfg DBConfig, optionsRaw string) (string, error) {
	query, err := url.ParseQuery(optionsRaw)
	if err != nil {
		return "", fmt.Errorf("DATABASE_OPTIONS is invalid: %w", err)
	}
	query.Set("sslmode", cfg.SSLMode)

	dbURL := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, cfg.Port),
		Path:     "/" + cfg.Name,
		RawQuery: query.Encode(),
	}

	return dbURL.String(), nil
}
