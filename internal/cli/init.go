package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flyits/migro/internal/config"
	"github.com/spf13/cobra"
)

var initDriver string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new migro project",
	Long:  `Creates a migro.yaml configuration file and migrations directory.`,
	RunE:  runInit,
}

func init() {
	initCmd.Flags().StringVar(&initDriver, "driver", "mysql", "database driver (mysql, postgres, sqlite)")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if config already exists
	loader := config.NewLoader(cfgFile)
	if loader.Exists() {
		return fmt.Errorf("configuration file %s already exists", cfgFile)
	}

	// Generate config template
	content := config.GenerateConfigTemplate(initDriver)

	// Write config file
	if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	fmt.Printf("Created configuration file: %s\n", cfgFile)

	// Create migrations directory
	migrationsDir := "./migrations"
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	fmt.Printf("Created migrations directory: %s\n", migrationsDir)

	// Create a .gitkeep file
	gitkeepPath := filepath.Join(migrationsDir, ".gitkeep")
	if err := os.WriteFile(gitkeepPath, []byte{}, 0644); err != nil {
		return fmt.Errorf("failed to create .gitkeep: %w", err)
	}

	fmt.Println("\nMigro initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit migro.yaml with your database credentials")
	fmt.Println("  2. Run 'migro create <name>' to create a new migration")
	fmt.Println("  3. Run 'migro up' to execute migrations")

	return nil
}
