package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration settings
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Logging  LoggingConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// RedisConfig holds Redis-related configuration
type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	SecretKey  string
	Issuer     string
	Expiration time.Duration
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level       string
	Development bool
}

// LoadConfig loads configuration from environment variables and defaults
func LoadConfig(configPath string) (*Config, error) { // configPath is kept for potential future use but ignored here
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Configure Viper to read environment variables
	v.SetEnvPrefix("API_GATEWAY")                      // Match the prefix used in docker-compose.yml
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // Allows nested env vars like SERVER_PORT
	v.AutomaticEnv()

	// No file reading logic needed here

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config from env: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.readTimeout", "30s")
	v.SetDefault("server.writeTimeout", "30s")
	v.SetDefault("server.shutdownTimeout", "30s")

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.database", "api_gateway")
	v.SetDefault("database.sslmode", "disable")

	// Redis defaults
	v.SetDefault("redis.address", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// Auth defaults
	v.SetDefault("auth.secretKey", "your-secret-key")
	v.SetDefault("auth.issuer", "api-gateway")
	v.SetDefault("auth.expiration", "24h")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.development", false)
}
