package unit

import (
	"os"
	"testing"

	"github.com/caarlos0/env/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Config represents the application configuration for testing
type Config struct {
	Port               int    `env:"APP_PORT" envDefault:"8080"`
	Host               string `env:"HOST" envDefault:"0.0.0.0"`
	LogLevel           string `env:"LOG_LEVEL" envDefault:"info"`
	ShutdownTimeout    int    `env:"SHUTDOWN_TIMEOUT" envDefault:"30"`
	HealthCheckTimeout int    `env:"HEALTH_CHECK_TIMEOUT" envDefault:"30"`
	SwaggerEnabled     bool   `env:"SWAGGER_ENABLED" envDefault:"true"`
	APIBaseURL         string `env:"API_BASE_URL" envDefault:"http://localhost:8080"`
	APIVersion         string `env:"API_VERSION" envDefault:"v1"`
	CORSOrigins        string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:3000,http://localhost:8080"`
	ServerReadTimeout  int    `env:"SERVER_READ_TIMEOUT" envDefault:"15"`
	ServerWriteTimeout int    `env:"SERVER_WRITE_TIMEOUT" envDefault:"15"`
	ServerIdleTimeout  int    `env:"SERVER_IDLE_TIMEOUT" envDefault:"60"`
}

func TestConfigDefaults(t *testing.T) {
	// Clear environment variables
	envVars := []string{
		"APP_PORT", "HOST", "LOG_LEVEL", "SHUTDOWN_TIMEOUT",
		"HEALTH_CHECK_TIMEOUT", "SWAGGER_ENABLED", "API_BASE_URL",
		"API_VERSION", "CORS_ALLOWED_ORIGINS", "SERVER_READ_TIMEOUT",
		"SERVER_WRITE_TIMEOUT", "SERVER_IDLE_TIMEOUT",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	cfg := &Config{}
	err := env.Parse(cfg)
	require.NoError(t, err)

	// Test default values
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, 30, cfg.ShutdownTimeout)
	assert.Equal(t, 30, cfg.HealthCheckTimeout)
	assert.True(t, cfg.SwaggerEnabled)
	assert.Equal(t, "http://localhost:8080", cfg.APIBaseURL)
	assert.Equal(t, "v1", cfg.APIVersion)
	assert.Equal(t, "http://localhost:3000,http://localhost:8080", cfg.CORSOrigins)
	assert.Equal(t, 15, cfg.ServerReadTimeout)
	assert.Equal(t, 15, cfg.ServerWriteTimeout)
	assert.Equal(t, 60, cfg.ServerIdleTimeout)
}

func TestConfigEnvironmentOverrides(t *testing.T) {
	// Set environment variables
	os.Setenv("APP_PORT", "9090")
	os.Setenv("HOST", "127.0.0.1")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("SHUTDOWN_TIMEOUT", "60")
	os.Setenv("SWAGGER_ENABLED", "false")
	defer func() {
		os.Unsetenv("APP_PORT")
		os.Unsetenv("HOST")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("SHUTDOWN_TIMEOUT")
		os.Unsetenv("SWAGGER_ENABLED")
	}()

	cfg := &Config{}
	err := env.Parse(cfg)
	require.NoError(t, err)

	// Test environment overrides
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, 60, cfg.ShutdownTimeout)
	assert.False(t, cfg.SwaggerEnabled)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		port        string
		expectError bool
	}{
		{
			name:        "Valid Port",
			port:        "8080",
			expectError: false,
		},
		{
			name:        "Invalid Port - Non-numeric",
			port:        "invalid",
			expectError: true,
		},
		{
			name:        "Invalid Port - Out of range",
			port:        "99999",
			expectError: false, // env parsing allows this, validation should be done elsewhere
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("APP_PORT", tt.port)
			defer os.Unsetenv("APP_PORT")

			cfg := &Config{}
			err := env.Parse(cfg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
