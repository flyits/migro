package cli

import (
	"context"
	"fmt"

	"github.com/migro/migro/internal/config"
	"github.com/migro/migro/internal/migrator"
	"github.com/migro/migro/pkg/driver"
	_ "github.com/migro/migro/pkg/driver/mysql"
	_ "github.com/migro/migro/pkg/driver/postgres"
	_ "github.com/migro/migro/pkg/driver/sqlite"
	"github.com/spf13/cobra"
)

var (
	upStep   int
	upDryRun bool
	upForce  bool
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Run pending migrations",
	Long:  `Executes all pending database migrations.`,
	RunE:  runUp,
}

func init() {
	upCmd.Flags().IntVar(&upStep, "step", 0, "number of migrations to run")
	upCmd.Flags().BoolVar(&upDryRun, "dry-run", false, "show SQL without executing")
	upCmd.Flags().BoolVar(&upForce, "force", false, "force execution without confirmation")
	rootCmd.AddCommand(upCmd)
}

func runUp(cmd *cobra.Command, args []string) error {
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
	m.SetDryRun(upDryRun)

	// Note: In a real implementation, migrations would be loaded from files
	// For now, we'll show a message about how to register migrations

	ctx := context.Background()

	// Run migrations
	executed, err := m.Up(ctx, upStep)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	if len(executed) == 0 {
		fmt.Println("Nothing to migrate.")
		return nil
	}

	if upDryRun {
		fmt.Println("Dry run - SQL statements that would be executed:")
	} else {
		fmt.Println("Migrations executed:")
	}

	for _, name := range executed {
		fmt.Printf("  - %s\n", name)
	}

	return nil
}
