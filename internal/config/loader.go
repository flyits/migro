package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultConfigFile is the default configuration file name
	DefaultConfigFile = "migro.yaml"
)

// Loader handles configuration loading
type Loader struct {
	configFile string
}

// NewLoader creates a new configuration loader
func NewLoader(configFile string) *Loader {
	if configFile == "" {
		configFile = DefaultConfigFile
	}
	return &Loader{configFile: configFile}
}

// Load loads the configuration from file
func (l *Loader) Load() (*Config, error) {
	data, err := os.ReadFile(l.configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	content := expandEnvVars(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults
	applyDefaults(&cfg)

	return &cfg, nil
}

// Exists checks if the configuration file exists
func (l *Loader) Exists() bool {
	_, err := os.Stat(l.configFile)
	return err == nil
}

// Save saves the configuration to file
func (l *Loader) Save(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(l.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// expandEnvVars expands environment variables in the format ${VAR:default}
func expandEnvVars(content string) string {
	// Match ${VAR} or ${VAR:default}
	re := regexp.MustCompile(`\$\{([^}:]+)(?::([^}]*))?\}`)

	return re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		varName := parts[1]
		defaultValue := ""
		if len(parts) >= 3 {
			defaultValue = parts[2]
		}

		if value := os.Getenv(varName); value != "" {
			return value
		}
		return defaultValue
	})
}

// applyDefaults applies default values to the configuration
func applyDefaults(cfg *Config) {
	if cfg.Driver == "" {
		cfg.Driver = "mysql"
	}

	if cfg.Connection.Host == "" {
		cfg.Connection.Host = "localhost"
	}

	if cfg.Connection.Port == 0 {
		cfg.Connection.Port = GetDefaultPort(cfg.Driver)
	}

	if cfg.Connection.Charset == "" && cfg.Driver == "mysql" {
		cfg.Connection.Charset = "utf8mb4"
	}

	if cfg.Migrations.Path == "" {
		cfg.Migrations.Path = "./migrations"
	}

	if cfg.Migrations.Table == "" {
		cfg.Migrations.Table = "migrations"
	}

	if cfg.Connection.Options == nil {
		cfg.Connection.Options = make(map[string]string)
	}
}

// GenerateConfigTemplate generates a configuration file template
func GenerateConfigTemplate(driverName string) string {
	var sb strings.Builder

	sb.WriteString("# Migro Configuration\n")
	sb.WriteString("# Database migration tool for Go\n\n")

	sb.WriteString("driver: ")
	sb.WriteString(driverName)
	sb.WriteString("\n\n")

	sb.WriteString("connection:\n")

	switch driverName {
	case "mysql":
		sb.WriteString("  host: ${DB_HOST:localhost}\n")
		sb.WriteString("  port: ${DB_PORT:3306}\n")
		sb.WriteString("  database: ${DB_NAME:myapp}\n")
		sb.WriteString("  username: ${DB_USER:root}\n")
		sb.WriteString("  password: ${DB_PASS:}\n")
		sb.WriteString("  charset: utf8mb4\n")
	case "postgres":
		sb.WriteString("  host: ${DB_HOST:localhost}\n")
		sb.WriteString("  port: ${DB_PORT:5432}\n")
		sb.WriteString("  database: ${DB_NAME:myapp}\n")
		sb.WriteString("  username: ${DB_USER:postgres}\n")
		sb.WriteString("  password: ${DB_PASS:}\n")
	case "sqlite":
		sb.WriteString("  database: ${DB_PATH:./database.db}\n")
	}

	sb.WriteString("\nmigrations:\n")
	sb.WriteString("  path: ./migrations\n")
	sb.WriteString("  table: migrations\n")

	return sb.String()
}

// ParsePort parses a port string to int
func ParsePort(s string) int {
	port, _ := strconv.Atoi(s)
	return port
}
