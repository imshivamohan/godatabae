package main

import (
	"flag"
	"fmt"
	"log"

	"your_module_name/database"
)

func main() {
	// Define command-line flags
	configPath := flag.String("f", "config.yaml", "Path to the database configuration file")
	pingFlag := flag.Bool("ping", false, "Test database connection")
	flag.Parse()

	// Validate that a config file path is provided
	if *configPath == "" {
		log.Fatal("Please provide a configuration file path using the -f flag")
	}

	// If ping flag is set, only perform ping test
	if *pingFlag {
		fmt.Println("Testing database connection...")
		err := database.PingDatabase(*configPath)
		if err != nil {
			log.Fatalf("Database connection test failed: %v", err)
		}
		fmt.Println("Database connection successful!")
		return
	}

	// Regular database connection and query logic
	dbConn, err := database.NewDatabaseConnection(*configPath)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer dbConn.Close()

	// Example query (modify as per your database schema)
	rows, err := dbConn.Query("SELECT * FROM users LIMIT 5")
	if err != nil {
		dbConn.Logger.Error().Err(err).Msg("Query execution failed")
		return
	}
	defer rows.Close()

	// Process rows (adjust scanning based on your table structure)
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
