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

var (
	downStep  int
	downForce bool
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback migrations",
	Long:  `Rolls back the last batch of migrations or a specified number of migrations.`,
	RunE:  runDown,
}

func init() {
	downCmd.Flags().IntVar(&downStep, "step", 0, "number of migrations to rollback (0 = last batch)")
	downCmd.Flags().BoolVar(&downForce, "force", false, "force rollback without confirmation")
	rootCmd.AddCommand(downCmd)
}

func runDown(cmd *cobra.Command, args []string) error {
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

	// Run rollback
	rolledBack, err := m.Down(ctx, downStep)
	if err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	if len(rolledBack) == 0 {
		fmt.Println("Nothing to rollback.")
		return nil
	}

	fmt.Println("Migrations rolled back:")
	for _, name := range rolledBack {
		fmt.Printf("  - %s\n", name)
	}

	return nil
}
