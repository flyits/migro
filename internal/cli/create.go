package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/migro/migro/internal/config"
	"github.com/spf13/cobra"
)

var createTable string

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new migration file",
	Long:  `Creates a new migration file with the given name.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

func init() {
	createCmd.Flags().StringVar(&createTable, "table", "", "table name for the migration")
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Load config to get migrations path
	loader := config.NewLoader(cfgFile)
	cfg, err := loader.Load()
	if err != nil {
		// Use default if config doesn't exist
		cfg = config.DefaultConfig()
	}

	// Generate timestamp
	timestamp := time.Now().Format("20060102150405")

	// Generate filename
	filename := fmt.Sprintf("%s_%s.go", timestamp, toSnakeCase(name))
	filepath := filepath.Join(cfg.Migrations.Path, filename)

	// Ensure migrations directory exists
	if err := os.MkdirAll(cfg.Migrations.Path, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Determine table name
	tableName := createTable
	if tableName == "" {
		tableName = extractTableName(name)
	}

	// Generate migration content
	content := generateMigrationTemplate(name, tableName, timestamp)

	// Write file
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	fmt.Printf("Created migration: %s\n", filepath)

	return nil
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func extractTableName(name string) string {
	// Try to extract table name from migration name
	// e.g., "create_users_table" -> "users"
	// e.g., "add_email_to_users" -> "users"

	name = toSnakeCase(name)

	if strings.HasPrefix(name, "create_") && strings.HasSuffix(name, "_table") {
		return strings.TrimSuffix(strings.TrimPrefix(name, "create_"), "_table")
	}

	if strings.Contains(name, "_to_") {
		parts := strings.Split(name, "_to_")
		if len(parts) > 1 {
			return parts[len(parts)-1]
		}
	}

	if strings.Contains(name, "_from_") {
		parts := strings.Split(name, "_from_")
		if len(parts) > 1 {
			return parts[len(parts)-1]
		}
	}

	return "table_name"
}

func generateMigrationTemplate(name, tableName, timestamp string) string {
	structName := toCamelCase(name)

	return fmt.Sprintf(`package migrations

import (
	"github.com/migro/migro/internal/migrator"
	"github.com/migro/migro/pkg/schema"
)

// %s migration
type %s struct{}

// Name returns the migration name
func (m *%s) Name() string {
	return "%s_%s"
}

// Up runs the migration
func (m *%s) Up(e *migrator.Executor) error {
	return e.CreateTable("%s", func(t *schema.Table) {
		t.ID()
		// Add your columns here
		// t.String("name", 100)
		// t.String("email", 100).Unique()
		t.Timestamps()
	})
}

// Down reverses the migration
func (m *%s) Down(e *migrator.Executor) error {
	return e.DropTableIfExists("%s")
}

func init() {
	// Register this migration
	// migrator.Register(&%s{})
}
`, structName, structName, structName, timestamp, toSnakeCase(name),
		structName, tableName, structName, tableName, structName)
}

func toCamelCase(s string) string {
	s = toSnakeCase(s)
	parts := strings.Split(s, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(string(part[0])))
			result.WriteString(part[1:])
		}
	}
	return result.String()
}
