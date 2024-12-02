main.go

package main

import (
	"flag"
	"fmt"
	"log"

	"your_module_name/database"
)

func main() {
	// Define a command-line flag for the config file
	configPath := flag.String("f", "config.yaml", "Path to the database configuration file")
	flag.Parse()

	// Validate that a config file path is provided
	if *configPath == "" {
		log.Fatal("Please provide a configuration file path using the -f flag")
	}

	// Create database connection
	dbConn, err := database.NewDatabaseConnection(*configPath)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer dbConn.Close()

	// Example query
	rows, err := dbConn.Query("SELECT * FROM users LIMIT 5")
	if err != nil {
		dbConn.Logger.Error().Err(err).Msg("Query execution failed")
		return
	}
	defer rows.Close()

	// Process rows
	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			dbConn.Logger.Error().Err(err).Msg("Row scan failed")
			continue
		}
		fmt.Printf("ID: %d, Name: %s\n", id, name)
	}
}
