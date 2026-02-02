package cli

import (
	"context"
	"fmt"

	"github.com/flyits/migro/internal/config"
	"github.com/flyits/migro/internal/migrator"
	"github.com/flyits/migro/pkg/driver"
	_ "github.com/flyits/migro/pkg/driver/mysql"
	_ "github.com/flyits/migro/pkg/driver/postgres"
	_ "github.com/flyits/migro/pkg/driver/sqlite"
	"github.com/spf13/cobra"
)

var refreshForce bool

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Reset and re-run all migrations",
	Long:  `Rolls back all migrations and re-runs them.`,
	RunE:  runRefresh,
}

func init() {
	refreshCmd.Flags().BoolVar(&refreshForce, "force", false, "force refresh without confirmation")
	rootCmd.AddCommand(refreshCmd)
}

func runRefresh(cmd *cobra.Command, args []string) error {
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

	// Run refresh
	rolledBack, executed, err := m.Refresh(ctx)
	if err != nil {
		return fmt.Errorf("refresh failed: %w", err)
	}

	if len(rolledBack) > 0 {
		fmt.Println("Migrations rolled back:")
		for _, name := range rolledBack {
			fmt.Printf("  - %s\n", name)
		}
	}

	if len(executed) > 0 {
		fmt.Println("\nMigrations executed:")
		for _, name := range executed {
			fmt.Printf("  - %s\n", name)
		}
	}

	if len(rolledBack) == 0 && len(executed) == 0 {
		fmt.Println("Nothing to refresh.")
	}

	return nil
}
