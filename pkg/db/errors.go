package db

import "errors"

var (
	// Validation Errors
	ErrMissingHost   = errors.New("database host is required")
	ErrInvalidPort   = errors.New("database port must be greater than zero")
	ErrMissingUser   = errors.New("database user is required")
	ErrMissingDBName = errors.New("database name is required")

	// Connection Errors
	ErrInvalidConfig = errors.New("invalid database configuration")
	ErrFailedConnect = errors.New("failed to connect to database")
	ErrFailedPing    = errors.New("failed to ping database")
	ErrNilDatabase   = errors.New("database connection is nil")
	ErrFailedClose   = errors.New("failed to close database connection")
)
