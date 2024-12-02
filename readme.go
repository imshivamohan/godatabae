Now you can run the application with:
bashCopygo run main.go -f config.yaml
Key changes:

Used flag package to parse command-line arguments
Added -f flag to specify config file path
Provided a default value of "config.yaml"
Added basic validation to ensure a config path is provided

Additional features:

If no -f flag is provided, it defaults to "config.yaml"
If an empty config path is provided, it logs a fatal error
Flexible configuration file specification

Example usage variations:
bashCopy# Use default config.yaml
go run main.go

# Specify a different config file
go run main.go -f /path/to/custom/config.yaml
Would you like me to make any further modifications to the command-line argument handling?
