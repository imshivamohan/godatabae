Database Connection Module

package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	// Import database drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"os"
	"path/filepath"
)

// DatabaseConfig represents the structure of database configuration
type DatabaseConfig struct {
	Database struct {
		Driver   string `yaml:"driver"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"database"`

	ConnectionPool struct {
		MaxOpenConnections    int           `yaml:"max_open_connections"`
		MaxIdleConnections    int           `yaml:"max_idle_connections"`
		MaxConnectionLifetime time.Duration `yaml:"max_connection_lifetime"`
	} `yaml:"connection_pool"`

	Logger struct {
		Level      string `yaml:"level"`
		OutputPath string `yaml:"output_path"`
	} `yaml:"logger"`
}

// DatabaseConnection holds the database connection and logger
type DatabaseConnection struct {
	DB     *sql.DB
	Config *DatabaseConfig
	Logger zerolog.Logger
}

// NewDatabaseConnection creates a new database connection
func NewDatabaseConnection(configPath string) (*DatabaseConnection, error) {
	// Read config file
	config, err := readConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %v", err)
	}

	// Setup logger
	logger := setupLogger(config)

	// Create connection string based on driver
	connectionString := buildConnectionString(config)

	// Open database connection
	db, err := sql.Open(config.Database.Driver, connectionString)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to database")
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.ConnectionPool.MaxOpenConnections)
	db.SetMaxIdleConns(config.ConnectionPool.MaxIdleConnections)
	db.SetConnMaxLifetime(config.ConnectionPool.MaxConnectionLifetime)

	// Test connection
	if err := db.Ping(); err != nil {
		logger.Error().Err(err).Msg("Failed to ping database")
		return nil, err
	}

	logger.Info().Msg("Database connection established successfully")

	return &DatabaseConnection{
		DB:     db,
		Config: config,
		Logger: logger,
	}, nil
}

// readConfig reads the YAML configuration file
func readConfig(configPath string) (*DatabaseConfig, error) {
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
	var config DatabaseConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// setupLogger configures zerolog based on config
func setupLogger(config *DatabaseConfig) zerolog.Logger {
	// Set logging level
	switch config.Logger.Level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Create log file if output path is specified
	if config.Logger.OutputPath != "" {
		// Ensure directory exists
		os.MkdirAll(filepath.Dir(config.Logger.OutputPath), os.ModePerm)
		
		// Open log file
		logFile, err := os.OpenFile(
			config.Logger.OutputPath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0666,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to open log file")
		}

		return zerolog.New(logFile).With().Timestamp().Logger()
	}

	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}

// buildConnectionString creates connection string based on driver
func buildConnectionString(config *DatabaseConfig) string {
	switch config.Database.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			config.Database.Username,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.Database,
		)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Database.Host,
			config.Database.Port,
			config.Database.Username,
			config.Database.Password,
			config.Database.Database,
		)
	case "sqlite3":
		return config.Database.Database // SQLite uses file path
	default:
		panic(fmt.Sprintf("Unsupported database driver: %s", config.Database.Driver))
	}
}

// Close closes the database connection
func (dc *DatabaseConnection) Close() error {
	dc.Logger.Info().Msg("Closing database connection")
	return dc.DB.Close()
}

// Query executes a query that returns rows
func (dc *DatabaseConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	dc.Logger.Debug().Str("query", query).Msg("Executing query")
	return dc.DB.Query(query, args...)
}

// QueryRow executes a query that is expected to return at most one row
func (dc *DatabaseConnection) QueryRow(query string, args ...interface{}) *sql.Row {
	dc.Logger.Debug().Str("query", query).Msg("Executing query row")
	return dc.DB.QueryRow(query, args...)
}

// Exec executes a query without returning any rows
func (dc *DatabaseConnection) Exec(query string, args ...interface{}) (sql.Result, error) {
	dc.Logger.Debug().Str("query", query).Msg("Executing exec")
	return dc.DB.Exec(query, args...)
}
```
