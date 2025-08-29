package codec

import (
	"fmt"
)

// Common error types for the codec package

// ValidationError represents an error in input validation
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s (%s): %s", e.Field, e.Value, e.Message)
}

// EncodingError represents an error during encoding operations
type EncodingError struct {
	Operation string
	Cause     error
}

func (e EncodingError) Error() string {
	return fmt.Errorf("encoding error during %s: %w", e.Operation, e.Cause).Error()
}

func (e EncodingError) Unwrap() error {
	return e.Cause
}

// DecodingError represents an error during decoding operations
type DecodingError struct {
	Operation string
	Input     string
	Cause     error
}

func (e DecodingError) Error() string {
	return fmt.Errorf("decoding error during %s (input: %s): %w", e.Operation, e.Input, e.Cause).Error()
}

func (e DecodingError) Unwrap() error {
	return e.Cause
}

// NetworkError represents network-related errors
type NetworkError struct {
	Endpoint  string
	Operation string
	Cause     error
}

func (e NetworkError) Error() string {
	return fmt.Errorf("network error for %s during %s: %w", e.Endpoint, e.Operation, e.Cause).Error()
}

func (e NetworkError) Unwrap() error {
	return e.Cause
}
