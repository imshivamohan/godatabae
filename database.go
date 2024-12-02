package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	// Database drivers
	_ "github.com/lib/pq"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

// Config represents the comprehensive database configuration
type Config struct {
	Database struct {
		Driver     string `yaml:"driver"`
		Host       string `yaml:"host"`
		Port       int    `yaml:"port"`
		Username   string `yaml:"username"`
		Password   string `yaml:"password"`
		DBName     string `yaml:"dbname"`
		SSLMode    string `yaml:"sslmode"`
		Filepath   string `yaml:"filepath"`
		LogLevel   string `yaml:"log_level"`
		DBSchema   string `yaml:"dbschema"`
		Pool struct {
			MaxOpenConns    int    `yaml:"max_open_conns"`
			MaxIdleConns    int    `yaml:"max_idle_conns"`
			ConnMaxLifetime string `yaml:"conn_max_lifetime"`
			ConnMaxIdleTime string `yaml:"conn_max_idle_time"`
		} `yaml:"pool"`
	} `yaml:"database"`
}

// DatabaseConnection holds the database connection and configuration
type DatabaseConnection struct {
	DB     *sql.DB
	Config *Config
	Logger zerolog.Logger
}

// NewDatabaseConnection creates a new database connection
func NewDatabaseConnection(configPath string) (*DatabaseConnection, error) {
	// Read configuration
	config, err := readConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %v", err)
	}

	// Setup logger
	logger := setupLogger(config.Database.LogLevel)

	// Build connection string
	dsn := buildConnectionString(config)

	// Open database connection
	db, err := sql.Open(config.Database.Driver, dsn)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to open database connection")
		return nil, err
	}

	// Configure connection pool
	if err := configureConnectionPool(db, config); err != nil {
		logger.Error().Err(err).Msg("Failed to configure connection pool")
		return nil, err
	}

	// Ping database to verify connection
	if err := pingDatabase(db, logger); err != nil {
		logger.Error().Err(err).Msg("Database connection ping failed")
		return nil, err
	}

	logger.Info().
		Str("driver", config.Database.Driver).
		Str("host", config.Database.Host).
		Str("database", config.Database.DBName).
		Msg("Database connection established successfully")

	return &DatabaseConnection{
		DB:     db,
		Config: config,
		Logger: logger,
	}, nil
}

// readConfig reads the YAML configuration file
func readConfig(configPath string) (*Config, error) {
	// Ensure absolute path
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}

	// Read file
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	// Unmarshal YAML
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// buildConnectionString creates connection string based on driver
func buildConnectionString(config *Config) string {
	switch config.Database.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s search_path=%s",
			config.Database.Host,
			config.Database.Port,
			config.Database.Username,
			config.Database.Password,
			config.Database.DBName,
			config.Database.SSLMode,
			config.Database.DBSchema,
		)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Database.Username,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.DBName,
		)
	case "sqlite3":
		return config.Database.Filepath
	default:
		panic(fmt.Sprintf("Unsupported database driver: %s", config.Database.Driver))
	}
}

// configureConnectionPool sets up connection pool settings
func configureConnectionPool(db *sql.DB, config *Config) error {
	// Parse connection pool durations
	connMaxLifetime, err := time.ParseDuration(config.Database.Pool.ConnMaxLifetime)
	if err != nil {
		return fmt.Errorf("invalid conn_max_lifetime: %v", err)
	}

	connMaxIdleTime, err := time.ParseDuration(config.Database.Pool.ConnMaxIdleTime)
	if err != nil {
		return fmt.Errorf("invalid conn_max_idle_time: %v", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(config.Database.Pool.MaxOpenConns)
	db.SetMaxIdleConns(config.Database.Pool.MaxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	return nil
}

// pingDatabase tests the database connection
func pingDatabase(db *sql.DB, logger zerolog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}

	return nil
}

// setupLogger configures zerolog based on log level
func setupLogger(level string) zerolog.Logger {
	// Configure log level
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Create logger with timestamp
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}

// Close closes the database connection
func (dc *DatabaseConnection) Close() error {
	dc.Logger.Info().Msg("Closing database connection")
	return dc.DB.Close()
}

// Query executes a generic query with logging
func (dc *DatabaseConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	// Log the query with debug level
	dc.Logger.Debug().
		Str("query", query).
		Interface("args", args).
		Msg("Executing database query")

	// Execute the query
	rows, err := dc.DB.Query(query, args...)
	if err != nil {
		dc.Logger.Error().
			Err(err).
			Str("query", query).
			Interface("args", args).
			Msg("Query execution failed")
		return nil, err
	}

	return rows, nil
}

// QueryRow executes a query that is expected to return at most one row
func (dc *DatabaseConnection) QueryRow(query string, args ...interface{}) *sql.Row {
	// Log the query with debug level
	dc.Logger.Debug().
		Str("query", query).
		Interface("args", args).
		Msg("Executing single row query")

	return dc.DB.QueryRow(query, args...)
}

// Exec executes a query without returning any rows
func (dc *DatabaseConnection) Exec(query string, args ...interface{}) (sql.Result, error) {
	// Log the query with debug level
	dc.Logger.Debug().
		Str("query", query).
		Interface("args", args).
		Msg("Executing database modification")

	// Execute the query
	result, err := dc.DB.Exec(query, args...)
	if err != nil {
		dc.Logger.Error().
			Err(err).
			Str("query", query).
			Interface("args", args).
			Msg("Query execution failed")
		return nil, err
	}

	return result, nil
}
