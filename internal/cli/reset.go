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

var resetForce bool

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Rollback all migrations",
	Long:  `Rolls back all executed migrations.`,
	RunE:  runReset,
}

func init() {
	resetCmd.Flags().BoolVar(&resetForce, "force", false, "force reset without confirmation")
	rootCmd.AddCommand(resetCmd)
}

func runReset(cmd *cobra.Command, args []string) error {
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

	// Run reset
	rolledBack, err := m.Reset(ctx)
	if err != nil {
		return fmt.Errorf("reset failed: %w", err)
	}

	if len(rolledBack) == 0 {
		fmt.Println("Nothing to reset.")
		return nil
	}

	fmt.Println("Migrations rolled back:")
	for _, name := range rolledBack {
		fmt.Printf("  - %s\n", name)
	}

	return nil
}
