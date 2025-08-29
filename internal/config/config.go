package config

import (
	"os"
	"strconv"
	"time"

	"github.com/kdhira/age-plugin-tlock/internal/drand"
)

// Config holds all configurable parameters for the age-plugin-tlock
type Config struct {
	// DrandEndpoint is the default drand HTTP API endpoint
	DrandEndpoint string

	// RequestTimeout is the timeout for network requests
	RequestTimeout time.Duration

	// MaxRetries is the maximum number of retries for network requests
	MaxRetries int

	// RetryDelay is the base delay between retries
	RetryDelay time.Duration

	// StrictMode enables strict chain validation by default
	StrictMode bool
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		DrandEndpoint:  drand.DefaultEndpoint,
		RequestTimeout: 10 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		StrictMode:     false,
	}
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	config := DefaultConfig()

	if endpoint := os.Getenv("AGE_PLUGIN_TLOCK_ENDPOINT"); endpoint != "" {
		config.DrandEndpoint = endpoint
	}

	if timeoutStr := os.Getenv("AGE_PLUGIN_TLOCK_TIMEOUT"); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			config.RequestTimeout = timeout
		}
	}

	if retriesStr := os.Getenv("AGE_PLUGIN_TLOCK_MAX_RETRIES"); retriesStr != "" {
		if retries, err := strconv.Atoi(retriesStr); err == nil && retries > 0 {
			config.MaxRetries = retries
		}
	}

	if delayStr := os.Getenv("AGE_PLUGIN_TLOCK_RETRY_DELAY"); delayStr != "" {
		if delay, err := time.ParseDuration(delayStr); err == nil {
			config.RetryDelay = delay
		}
	}

	if strictStr := os.Getenv("AGE_PLUGIN_TLOCK_STRICT_MODE"); strictStr != "" {
		if strict, err := strconv.ParseBool(strictStr); err == nil {
			config.StrictMode = strict
		}
	}

	return config
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.DrandEndpoint == "" {
		return &ValidationError{Field: "DrandEndpoint", Message: "cannot be empty"}
	}
	if c.RequestTimeout <= 0 {
		return &ValidationError{Field: "RequestTimeout", Message: "must be positive"}
	}
	if c.MaxRetries < 0 {
		return &ValidationError{Field: "MaxRetries", Message: "cannot be negative"}
	}
	if c.RetryDelay <= 0 {
		return &ValidationError{Field: "RetryDelay", Message: "must be positive"}
	}
	return nil
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return "config validation error: " + e.Field + " " + e.Message
}
