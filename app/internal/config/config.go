package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server        ServerConfig
	Log           LogConfig
	Auth          AuthConfig
	Observability ObservabilityConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Address         string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, text
}

// AuthConfig holds API authentication configuration
type AuthConfig struct {
	ZitadelOIDC bool // Master toggle for Zitadel JWT authentication
	JWT         JWTConfig
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Issuer   string        // JWT issuer URL (e.g., "https://zitadel.example.com")
	Audience string        // JWT audience (e.g., project ID)
	JWKSURL  string        // JWKS endpoint URL
	CacheTTL time.Duration // JWKS cache TTL (default: 1 hour)
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	Enabled          bool   // Toggle for Prometheus metrics + OpenTelemetry
	OTLPGRPCEndpoint string // OTLP gRPC endpoint for traces
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	zitadelDomain := getEnv("ZITADEL_DOMAIN", "")

	cfg := &Config{
		Server: ServerConfig{
			Address:         getEnv("SERVER_ADDRESS", "0.0.0.0"),
			Port:            getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:     getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			ShutdownTimeout: getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "text"),
		},
		Auth: AuthConfig{
			ZitadelOIDC: getEnvAsBool("ZITADEL_OIDC", false),
			JWT: JWTConfig{
				Issuer:   getEnv("JWT_ISSUER", deriveJWTIssuer(zitadelDomain)),
				Audience: getEnv("JWT_AUDIENCE", ""),
				JWKSURL:  getEnv("JWT_JWKS_URL", deriveJWKSURL(zitadelDomain)),
				CacheTTL: getEnvAsDuration("JWT_CACHE_TTL", 1*time.Hour),
			},
		},
		Observability: ObservabilityConfig{
			Enabled:          getEnvAsBool("OBSERVABILITY", false),
			OTLPGRPCEndpoint: getEnv("OTLP_GRPC_ENDPOINT", "http://tempo:4317"),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func deriveJWTIssuer(domain string) string {
	if domain == "" {
		return ""
	}
	return "https://" + domain
}

func deriveJWKSURL(domain string) string {
	if domain == "" {
		return ""
	}
	return "https://" + domain + "/oauth/v2/keys"
}
