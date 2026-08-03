package config

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type KindeConfig struct {
	IssuerURL          string
	Audience           string
	M2MClientID        string
	M2MClientSecret    string
	M2MTokenURL        string
	ManagementAudience string
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
}

type ObsConfig struct {
	OTELEnabled         bool
	OTELExporter        string
	OTELMetricsExporter string
	LogFormat           string
	SampleRatio         float64
}

type VisionConfig struct {
	APIKey string
}

type MaintenanceConfig struct {
	Enabled     bool
	BypassToken string
}

type BookletConfig struct {
	BaseURL        string
	InternalSecret string
}

type Config struct {
	Kinde              KindeConfig
	DB                 DBConfig
	R2                 R2Config
	Obs                ObsConfig
	Vision             VisionConfig
	Maintenance        MaintenanceConfig
	Booklet            BookletConfig
	Port               string
	CORSAllowedOrigins []string
	AnthropicAPIKey    string
	AnthropicModel     string
	InternalAPISecret  string
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

	kindeM2MClientID, err := requireEnv("KINDE_M2M_CLIENT_ID")
	if err != nil {
		return nil, err
	}

	kindeM2MClientSecret, err := requireEnv("KINDE_M2M_CLIENT_SECRET")
	if err != nil {
		return nil, err
	}

	kindeM2MTokenURL, err := requireEnv("KINDE_M2M_TOKEN_URL")
	if err != nil {
		return nil, err
	}

	kindeManagementAudience, err := requireEnv("KINDE_MANAGEMENT_AUDIENCE")
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

	corsAllowedOriginsRaw, err := requireEnv("CORS_ALLOWED_ORIGINS")
	if err != nil {
		return nil, err
	}

	internalAPISecret, err := requireEnv("INTERNAL_API_SECRET")
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

	otelEnabled := envWithDefault("OTEL_ENABLED", "false") == "true"
	otelExporter := envWithDefault("OTEL_EXPORTER", "")
	otelMetricsExporter := envWithDefault("OTEL_METRICS_EXPORTER", "")
	if otelEnabled && otelExporter == "" {
		return nil, fmt.Errorf("OTEL_EXPORTER is required when OTEL_ENABLED=true")
	}
	if otelEnabled && otelMetricsExporter == "" {
		return nil, fmt.Errorf("OTEL_METRICS_EXPORTER is required when OTEL_ENABLED=true")
	}

	logFormat := envWithDefault("LOG_FORMAT", "json")
	visionAPIKey := envWithDefault("GOOGLE_VISION_API_KEY", "")
	port := envWithDefault("PORT", "8080")

	maintenanceEnabled := envWithDefault("MAINTENANCE_MODE", "false") == "true"
	maintenanceBypassToken := envWithDefault("MAINTENANCE_BYPASS_TOKEN", "")

	sampleRatioStr := envWithDefault("OTEL_SAMPLE_RATIO", "0.1")
	sampleRatio, err := strconv.ParseFloat(sampleRatioStr, 64)
	if err != nil {
		return nil, fmt.Errorf("OTEL_SAMPLE_RATIO must be a float: %w", err)
	}

	return &Config{
		Kinde: KindeConfig{
			IssuerURL:          kindeIssuerURL,
			Audience:           kindeAudience,
			M2MClientID:        kindeM2MClientID,
			M2MClientSecret:    kindeM2MClientSecret,
			M2MTokenURL:        kindeM2MTokenURL,
			ManagementAudience: kindeManagementAudience,
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
		},
		Obs: ObsConfig{
			OTELEnabled:         otelEnabled,
			OTELExporter:        otelExporter,
			OTELMetricsExporter: otelMetricsExporter,
			LogFormat:           logFormat,
			SampleRatio:         sampleRatio,
		},
		Vision: VisionConfig{
			APIKey: visionAPIKey,
		},
		Maintenance: MaintenanceConfig{
			Enabled:     maintenanceEnabled,
			BypassToken: maintenanceBypassToken,
		},
		Booklet: BookletConfig{
			BaseURL:        envWithDefault("BOOKLET_BASE_URL", ""),
			InternalSecret: envWithDefault("BOOKLET_INTERNAL_SECRET", ""),
		},
		AnthropicAPIKey:    envWithDefault("ANTHROPIC_API_KEY", ""),
		AnthropicModel:     envWithDefault("ANTHROPIC_MODEL", "claude-haiku-4-5-20251001"),
		InternalAPISecret:  internalAPISecret,
		Port:               port,
		CORSAllowedOrigins: strings.Split(corsAllowedOriginsRaw, ","),
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
