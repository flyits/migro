package config

import (
	"github.com/migro/migro/pkg/driver"
)

// Config represents the migro configuration
type Config struct {
	Driver     string           `yaml:"driver"`
	Connection ConnectionConfig `yaml:"connection"`
	Migrations MigrationsConfig `yaml:"migrations"`
}

// ConnectionConfig holds database connection settings
type ConnectionConfig struct {
	Host     string            `yaml:"host"`
	Port     int               `yaml:"port"`
	Database string            `yaml:"database"`
	Username string            `yaml:"username"`
	Password string            `yaml:"password"`
	Charset  string            `yaml:"charset"`
	Options  map[string]string `yaml:"options"`
}

// MigrationsConfig holds migration settings
type MigrationsConfig struct {
	Path  string `yaml:"path"`
	Table string `yaml:"table"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Driver: "mysql",
		Connection: ConnectionConfig{
			Host:     "localhost",
			Port:     3306,
			Database: "myapp",
			Username: "root",
			Password: "",
			Charset:  "utf8mb4",
			Options:  make(map[string]string),
		},
		Migrations: MigrationsConfig{
			Path:  "./migrations",
			Table: "migrations",
		},
	}
}

// ToDriverConfig converts to driver.Config
func (c *Config) ToDriverConfig() *driver.Config {
	return &driver.Config{
		Driver:   c.Driver,
		Host:     c.Connection.Host,
		Port:     c.Connection.Port,
		Database: c.Connection.Database,
		Username: c.Connection.Username,
		Password: c.Connection.Password,
		Charset:  c.Connection.Charset,
		Options:  c.Connection.Options,
	}
}

// GetDefaultPort returns the default port for the given driver
func GetDefaultPort(driverName string) int {
	switch driverName {
	case "mysql":
		return 3306
	case "postgres":
		return 5432
	case "sqlite":
		return 0
	default:
		return 0
	}
}
