package signal

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the signal server
type Config struct {
	Port           int
	Host           string
	BufferSize     int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	ClientTTL      int
	LogLevel       string
	LogFormat      string
	RepositoryType string
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
}

// DefaultConfig returns configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		Port:           3986,
		Host:           "0.0.0.0",
		BufferSize:     1024,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		ClientTTL:      60,
		LogLevel:       "info",
		LogFormat:      "text",
		RepositoryType: "memory",
		RedisAddr:      "localhost:6379",
		RedisPassword:  "",
		RedisDB:        0,
	}
}

// LoadConfigFromEnv loads configuration from environment variables,
// falling back to defaults.
func LoadConfigFromEnv() (Config, error) {
	cfg := DefaultConfig()

	if port := os.Getenv("SERVER_PORT"); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			return cfg, fmt.Errorf("invalid SERVER_PORT: %w", err)
		}
		if p < 1 || p > 65535 {
			return cfg, fmt.Errorf("SERVER_PORT must be between 1 and 65535")
		}
		cfg.Port = p
	}

	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Host = host
	}

	if bufSize := os.Getenv("BUFFER_SIZE"); bufSize != "" {
		size, err := strconv.Atoi(bufSize)
		if err != nil {
			return cfg, fmt.Errorf("invalid BUFFER_SIZE: %w", err)
		}
		if size < 512 || size > 65536 {
			return cfg, fmt.Errorf("BUFFER_SIZE must be between 512 and 65536")
		}
		cfg.BufferSize = size
	}

	if ttl := os.Getenv("CLIENT_TTL"); ttl != "" {
		t, err := strconv.Atoi(ttl)
		if err != nil {
			return cfg, fmt.Errorf("invalid CLIENT_TTL: %w", err)
		}
		if t < 10 || t > 3600 {
			return cfg, fmt.Errorf("CLIENT_TTL must be between 10 and 3600 seconds")
		}
		cfg.ClientTTL = t
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}
	if logFormat := os.Getenv("LOG_FORMAT"); logFormat != "" {
		cfg.LogFormat = logFormat
	}

	if repoType := os.Getenv("REPOSITORY_TYPE"); repoType != "" {
		if repoType != "memory" && repoType != "redis" {
			return cfg, fmt.Errorf("REPOSITORY_TYPE must be 'memory' or 'redis'")
		}
		cfg.RepositoryType = repoType
	}

	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		cfg.RedisAddr = redisAddr
	}
	if redisPass := os.Getenv("REDIS_PASSWORD"); redisPass != "" {
		cfg.RedisPassword = redisPass
	}
	if redisDB := os.Getenv("REDIS_DB"); redisDB != "" {
		db, err := strconv.Atoi(redisDB)
		if err != nil {
			return cfg, fmt.Errorf("invalid REDIS_DB: %w", err)
		}
		if db < 0 || db > 15 {
			return cfg, fmt.Errorf("REDIS_DB must be between 0 and 15")
		}
		cfg.RedisDB = db
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Port)
	}
	if c.BufferSize < 512 {
		return fmt.Errorf("buffer size too small: %d", c.BufferSize)
	}
	if c.ClientTTL < 10 {
		return fmt.Errorf("client TTL too small: %d", c.ClientTTL)
	}
	if c.RepositoryType != "memory" && c.RepositoryType != "redis" {
		return fmt.Errorf("invalid repository type: %s", c.RepositoryType)
	}
	return nil
}
