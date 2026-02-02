package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyits/migro/internal/config"
	"github.com/flyits/migro/internal/migrator"
	"github.com/flyits/migro/pkg/driver"
	_ "github.com/flyits/migro/pkg/driver/mysql"
	_ "github.com/flyits/migro/pkg/driver/postgres"
	_ "github.com/flyits/migro/pkg/driver/sqlite"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  `Displays the status of all migrations.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Load config
	loader := config.NewLoader(cfgFile)
	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get driver
	drv, err := driver.Get(cfg.Driver)
	if err != nil {
		return fmt.Errorf("failed to get driver: %w", err)
	}

	// Connect to database
	if err := drv.Connect(cfg.ToDriverConfig()); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer drv.Close()

	// Create migrator
	m := migrator.NewMigrator(drv, cfg.Migrations.Path, cfg.Migrations.Table)

	ctx := context.Background()

	// Get status
	statuses, err := m.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if len(statuses) == 0 {
		fmt.Println("No migrations found.")
		return nil
	}

	// Print table header
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("| %-40s | %-6s | %-5s | %-19s |\n", "Migration", "Status", "Batch", "Executed At")
	fmt.Println(strings.Repeat("-", 80))

	// Print migrations
	for _, s := range statuses {
		status := "Pending"
		batch := ""
		executedAt := ""

		if s.Ran {
			status = "Ran"
			batch = fmt.Sprintf("%d", s.Batch)
			executedAt = s.ExecutedAt
		}

		name := s.Name
		if len(name) > 40 {
			name = name[:37] + "..."
		}

		fmt.Printf("| %-40s | %-6s | %-5s | %-19s |\n", name, status, batch, executedAt)
	}

	fmt.Println(strings.Repeat("-", 80))

	return nil
}
