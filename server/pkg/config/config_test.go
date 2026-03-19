package config

import (
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Test default values
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"ServerPort", cfg.ServerPort, 3986},
		{"ServerHost", cfg.ServerHost, "0.0.0.0"},
		{"BufferSize", cfg.BufferSize, 1024},
		{"ClientTTL", cfg.ClientTTL, 60},
		{"RepositoryType", cfg.RepositoryType, "memory"},
		{"RedisAddr", cfg.RedisAddr, "localhost:6379"},
		{"RedisDB", cfg.RedisDB, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}

	if cfg.ReadTimeout != 30*time.Second {
		t.Errorf("ReadTimeout = %v, want %v", cfg.ReadTimeout, 30*time.Second)
	}

	if cfg.WriteTimeout != 30*time.Second {
		t.Errorf("WriteTimeout = %v, want %v", cfg.WriteTimeout, 30*time.Second)
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Save original env vars
	originalEnv := make(map[string]string)
	envVars := []string{"SERVER_PORT", "SERVER_HOST", "BUFFER_SIZE", "CLIENT_TTL", "REPOSITORY_TYPE"}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
	}

	// Cleanup after test
	defer func() {
		for key, val := range originalEnv {
			if val == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, val)
			}
		}
	}()

	t.Run("DefaultValues", func(t *testing.T) {
		// Clear all env vars
		for _, key := range envVars {
			os.Unsetenv(key)
		}

		cfg, err := LoadFromEnv()
		if err != nil {
			t.Fatalf("LoadFromEnv() error = %v", err)
		}

		if cfg.ServerPort != 3986 {
			t.Errorf("ServerPort = %d, want 3986", cfg.ServerPort)
		}
	})

	t.Run("CustomPort", func(t *testing.T) {
		os.Setenv("SERVER_PORT", "8080")
		defer os.Unsetenv("SERVER_PORT")

		cfg, err := LoadFromEnv()
		if err != nil {
			t.Fatalf("LoadFromEnv() error = %v", err)
		}

		if cfg.ServerPort != 8080 {
			t.Errorf("ServerPort = %d, want 8080", cfg.ServerPort)
		}
	})

	t.Run("InvalidPort", func(t *testing.T) {
		os.Setenv("SERVER_PORT", "invalid")
		defer os.Unsetenv("SERVER_PORT")

		_, err := LoadFromEnv()
		if err == nil {
			t.Error("LoadFromEnv() expected error for invalid port")
		}
	})

	t.Run("PortOutOfRange", func(t *testing.T) {
		os.Setenv("SERVER_PORT", "99999")
		defer os.Unsetenv("SERVER_PORT")

		_, err := LoadFromEnv()
		if err == nil {
			t.Error("LoadFromEnv() expected error for out of range port")
		}
	})

	t.Run("CustomBufferSize", func(t *testing.T) {
		os.Setenv("BUFFER_SIZE", "2048")
		defer os.Unsetenv("BUFFER_SIZE")

		cfg, err := LoadFromEnv()
		if err != nil {
			t.Fatalf("LoadFromEnv() error = %v", err)
		}

		if cfg.BufferSize != 2048 {
			t.Errorf("BufferSize = %d, want 2048", cfg.BufferSize)
		}
	})

	t.Run("InvalidBufferSize", func(t *testing.T) {
		os.Setenv("BUFFER_SIZE", "100")
		defer os.Unsetenv("BUFFER_SIZE")

		_, err := LoadFromEnv()
		if err == nil {
			t.Error("LoadFromEnv() expected error for buffer size too small")
		}
	})

	t.Run("CustomRepositoryType", func(t *testing.T) {
		os.Setenv("REPOSITORY_TYPE", "redis")
		defer os.Unsetenv("REPOSITORY_TYPE")

		cfg, err := LoadFromEnv()
		if err != nil {
			t.Fatalf("LoadFromEnv() error = %v", err)
		}

		if cfg.RepositoryType != "redis" {
			t.Errorf("RepositoryType = %s, want redis", cfg.RepositoryType)
		}
	})

	t.Run("InvalidRepositoryType", func(t *testing.T) {
		os.Setenv("REPOSITORY_TYPE", "invalid")
		defer os.Unsetenv("REPOSITORY_TYPE")

		_, err := LoadFromEnv()
		if err == nil {
			t.Error("LoadFromEnv() expected error for invalid repository type")
		}
	})
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "ValidConfig",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "InvalidPort",
			config: &Config{
				ServerPort:     99999,
				BufferSize:     1024,
				ClientTTL:      60,
				RepositoryType: "memory",
			},
			wantErr: true,
		},
		{
			name: "InvalidBufferSize",
			config: &Config{
				ServerPort:     3986,
				BufferSize:     100,
				ClientTTL:      60,
				RepositoryType: "memory",
			},
			wantErr: true,
		},
		{
			name: "InvalidClientTTL",
			config: &Config{
				ServerPort:     3986,
				BufferSize:     1024,
				ClientTTL:      5,
				RepositoryType: "memory",
			},
			wantErr: true,
		},
		{
			name: "InvalidRepositoryType",
			config: &Config{
				ServerPort:     3986,
				BufferSize:     1024,
				ClientTTL:      60,
				RepositoryType: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
