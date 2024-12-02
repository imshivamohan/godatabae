Main Application Example

```go
package main

import (
	"fmt"
	"log"

	"your-module/database"
)

func main() {
	// Create database connection
	dbConn, err := database.NewDatabaseConnection("config.yaml")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// Example query
	rows, err := dbConn.Query("SELECT * FROM users LIMIT 5")
	if err != nil {
		dbConn.Logger.Error().Err(err).Msg("Query failed")
		return
	}
	defer rows.Close()

	// Process rows
	for rows.Next() {
		// Scan row data
		var id int
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			dbConn.Logger.Error().Err(err).Msg("Row scan failed")
			continue
		}
		fmt.Printf("ID: %d, Name: %s\n", id, name)
	}

	// Check for errors after iterating
	if err = rows.Err(); err != nil {
		dbConn.Logger.Error().Err(err).Msg("Error occurred during row iteration")
	}
}
```

