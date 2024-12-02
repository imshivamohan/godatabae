package database

import (
	"context"
	"fmt"
	"time"
)

// PingDatabase tests the database connection and returns a detailed result
func PingDatabase(configPath string) error {
	// Create a new database connection
	dbConn, err := NewDatabaseConnection(configPath)
	if err != nil {
		return fmt.Errorf("failed to create database connection: %v", err)
	}
	defer dbConn.Close()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ping the database
	start := time.Now()
	err = dbConn.DB.PingContext(ctx)
	duration := time.Since(start)

	if err != nil {
		dbConn.Logger.Error().
			Err(err).
			Str("driver", dbConn.Config.Database.Driver).
			Str("host", dbConn.Config.Database.Host).
			Str("database", dbConn.Config.Database.DBName).
			Dur("ping_duration", duration).
			Msg("Database ping failed")
		return fmt.Errorf("database ping failed: %v", err)
	}

	// Log successful connection
	dbConn.Logger.Info().
		Str("driver", dbConn.Config.Database.Driver).
		Str("host", dbConn.Config.Database.Host).
		Str("database", dbConn.Config.Database.DBName).
		Dur("ping_duration", duration).
		Msg("Database connection successful")

	return nil
}

// TestDatabaseConnection is a wrapper for PingDatabase that can be used in tests
func TestDatabaseConnection(configPath string) (bool, error) {
	err := PingDatabase(configPath)
	if err != nil {
		return false, err
	}
	return true, nil
}
