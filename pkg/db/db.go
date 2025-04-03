package db

import (
	"database/sql"
	"fmt"
	"net/url"

	// Blank import for postgres driver
	_ "github.com/lib/pq"
)

// Config represents the database connection settings.
type Config struct {
	// Host specifies the database server's hostname or IP address.
	// Example: "localhost" or "192.168.1.100".
	Host string `toml:"host" mapstructure:"host"`

	// Port defines the port on which the database server is listening.
	// Default PostgreSQL port: 5432.
	Port int `toml:"port" mapstructure:"port"`

	// User represents the database username for authentication.
	// Example: "postgres" for a default PostgreSQL setup.
	User string `toml:"user" mapstructure:"user"`

	// Password is the database user's password.
	Password string `toml:"password" mapstructure:"password"`

	// DBName specifies the name of the PostgreSQL database to connect to.
	DBName string `toml:"dbname" mapstructure:"dbname"`

	// SSLMode determines whether to use SSL/TLS for database connections.
	// Example values: "disable", "require", "verify-full".
	// If empty, it defaults to "disable".
	SSLMode string `toml:"sslmode" mapstructure:"sslmode"`
}

// Validate ensures required fields are set and applies default values where necessary.
func (cfg *Config) Validate() error {
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == 0 {
		cfg.Port = 5432
	}
	if cfg.User == "" {
		cfg.User = "postgres"
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable" // Default to disabling SSL
	}
	// Ensure DBName is explicitly set (no default value)
	if cfg.DBName == "" {
		return ErrMissingDBName
	}

	return nil
}

// DSN constructs the PostgreSQL connection string.
func (cfg Config) DSN() (string, error) {
	if err := cfg.Validate(); err != nil {
		return "", err
	}

	q := url.Values{}
	q.Add("sslmode", cfg.SSLMode)

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:     cfg.DBName,
		RawQuery: q.Encode(),
	}

	return u.String(), nil
}

func NewDB(cfg Config) (*sql.DB, error) {
	connStr, err := cfg.DSN()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedConnect, err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedPing, err)
	}

	return db, nil
}
