package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"your_module_path/database" // Replace with the actual path of your `database` package
)

func main() {
	// Command-line flags
	configPath := flag.String("config", "config.yaml", "Path to the database configuration file")
	filePath := flag.String("file", "", "Path to the security.list file")
	flag.Parse()

	// Validate configPath flag
	if *configPath == "" {
		log.Fatal("Please provide a configuration file path using the -config flag")
	}

	// Validate filePath flag
	if *filePath == "" {
		fmt.Println("Please provide a file path with the -file flag.")
		return
	}

	// Open the file
	file, err := os.Open(*filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Read and parse the file
	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.FieldsPerRecord = -1

	identities := make(map[string]struct{})
	domains := make(map[string]struct{})
	mappings := make(map[string]map[string]struct{})

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	for _, record := range records {
		if len(record) != 2 {
			continue
		}
		domain := strings.TrimSpace(record[0])
		identityList := strings.Split(strings.TrimSpace(record[1]), ":")

		domains[domain] = struct{}{}
		if _, exists := mappings[domain]; !exists {
			mappings[domain] = make(map[string]struct{})
		}

		for _, identity := range identityList {
			identity = strings.TrimSpace(identity)
			identities[identity] = struct{}{}
			mappings[domain][identity] = struct{}{}
		}
	}

	// Create database connection
	dbConn, err := database.NewDatabaseConnection(*configPath)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer dbConn.Close()

	fmt.Println("Successfully connected to the database!")

	// Insert identities
	for identity := range identities {
		_, err := dbConn.DB.Exec(
			`INSERT INTO technical_identities (identity) VALUES ($1)
			ON CONFLICT (identity) DO NOTHING`,
			identity,
		)
		if err != nil {
			fmt.Println("Error inserting identity:", err)
		}
	}

	// Insert domains
	for domain := range domains {
		_, err := dbConn.DB.Exec(
			`INSERT INTO data_domains (domain_name) VALUES ($1)
			ON CONFLICT (domain_name) DO NOTHING`,
			domain,
		)
		if err != nil {
			fmt.Println("Error inserting domain:", err)
		}
	}

	// Insert mappings
	for domain, identitySet := range mappings {
		for identity := range identitySet {
			_, err := dbConn.DB.Exec(
				`INSERT INTO data_domain_identities (identity, domain_name)
				VALUES ($1, $2)
				ON CONFLICT (identity, domain_name) DO NOTHING`,
				identity, domain,
			)
			if err != nil {
				fmt.Println("Error inserting mapping:", err)
			}
		}
	}

	fmt.Println("Data insertion completed successfully!")
}
