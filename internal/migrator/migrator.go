package migrator

import (
	"context"
	"fmt"
	"sort"

	"github.com/migro/migro/pkg/driver"
	"github.com/migro/migro/pkg/schema"
)

// Migration represents a migration definition
type Migration struct {
	Name string
	Up   func(context.Context, *Executor) error
	Down func(context.Context, *Executor) error
}

// Migrator handles migration execution
type Migrator struct {
	driver         driver.Driver
	migrationsPath string
	tableName      string
	migrations     []Migration
	dryRun         bool
}

// NewMigrator creates a new migrator instance
func NewMigrator(drv driver.Driver, migrationsPath, tableName string) *Migrator {
	return &Migrator{
		driver:         drv,
		migrationsPath: migrationsPath,
		tableName:      tableName,
		migrations:     make([]Migration, 0),
	}
}

// SetDryRun enables or disables dry run mode
func (m *Migrator) SetDryRun(dryRun bool) {
	m.dryRun = dryRun
}

// Register registers a migration
func (m *Migrator) Register(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// RegisterAll registers multiple migrations
func (m *Migrator) RegisterAll(migrations []Migration) {
	m.migrations = append(m.migrations, migrations...)
}

// supportsTransactionalDDL returns true if the database supports transactional DDL
// PostgreSQL and SQLite support transactional DDL, MySQL does not
func (m *Migrator) supportsTransactionalDDL() bool {
	name := m.driver.Name()
	return name == "postgres" || name == "sqlite"
}

// executeMigrationInTransaction executes a migration within a transaction
// isUp: true for Up migration, false for Down migration
func (m *Migrator) executeMigrationInTransaction(ctx context.Context, migration Migration, batch int, isUp bool) error {
	tx, err := m.driver.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for migration %s: %w", migration.Name, err)
	}

	// Create a transaction-aware executor
	executor := NewTransactionExecutor(m.driver, tx)

	var execErr error
	if isUp {
		execErr = migration.Up(ctx, executor)
	} else {
		execErr = migration.Down(ctx, executor)
	}

	if execErr != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("migration %s failed: %w (rollback also failed: %v)", migration.Name, execErr, rbErr)
		}
		if isUp {
			return fmt.Errorf("migration %s failed (rolled back): %w", migration.Name, execErr)
		}
		return fmt.Errorf("rollback of %s failed (transaction rolled back): %w", migration.Name, execErr)
	}

	// Record or delete migration within the same transaction
	if isUp {
		sql := m.driver.Grammar().CompileInsertMigration(m.tableName)
		if _, err := tx.Exec(ctx, sql, migration.Name, batch); err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return fmt.Errorf("failed to record migration %s: %w (rollback also failed: %v)", migration.Name, err, rbErr)
			}
			return fmt.Errorf("failed to record migration %s (rolled back): %w", migration.Name, err)
		}
	} else {
		sql := m.driver.Grammar().CompileDeleteMigration(m.tableName)
		if _, err := tx.Exec(ctx, sql, migration.Name); err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return fmt.Errorf("failed to delete migration record %s: %w (rollback also failed: %v)", migration.Name, err, rbErr)
			}
			return fmt.Errorf("failed to delete migration record %s (rolled back): %w", migration.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration %s: %w", migration.Name, err)
	}

	return nil
}

// Up runs all pending migrations
func (m *Migrator) Up(ctx context.Context, step int) ([]string, error) {
	// Ensure migrations table exists
	if err := m.driver.CreateMigrationsTable(ctx, m.tableName); err != nil {
		return nil, fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get executed migrations
	executed, err := m.driver.GetExecutedMigrations(ctx, m.tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get executed migrations: %w", err)
	}

	executedMap := make(map[string]bool)
	for _, r := range executed {
		executedMap[r.Migration] = true
	}

	// Find pending migrations
	var pending []Migration
	for _, migration := range m.migrations {
		if !executedMap[migration.Name] {
			pending = append(pending, migration)
		}
	}

	// Sort by name (which includes timestamp)
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Name < pending[j].Name
	})

	// Apply step limit
	if step > 0 && step < len(pending) {
		pending = pending[:step]
	}

	if len(pending) == 0 {
		return nil, nil
	}

	// Get next batch number
	lastBatch, err := m.driver.GetLastBatch(ctx, m.tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get last batch: %w", err)
	}
	batch := lastBatch + 1

	// Execute migrations
	var executedNames []string
	for _, migration := range pending {
		// Use transaction for databases that support transactional DDL
		// PostgreSQL and SQLite support transactional DDL, MySQL does not
		if m.supportsTransactionalDDL() && !m.dryRun {
			if err := m.executeMigrationInTransaction(ctx, migration, batch, true); err != nil {
				return executedNames, err
			}
		} else {
			executor := NewExecutor(m.driver, m.dryRun)
			if err := migration.Up(ctx, executor); err != nil {
				return executedNames, fmt.Errorf("migration %s failed: %w", migration.Name, err)
			}
			if !m.dryRun {
				if err := m.driver.RecordMigration(ctx, m.tableName, migration.Name, batch); err != nil {
					return executedNames, fmt.Errorf("failed to record migration %s: %w", migration.Name, err)
				}
			}
		}

		executedNames = append(executedNames, migration.Name)
	}

	return executedNames, nil
}

// Down rolls back migrations
func (m *Migrator) Down(ctx context.Context, step int) ([]string, error) {
	// Get executed migrations
	executed, err := m.driver.GetExecutedMigrations(ctx, m.tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get executed migrations: %w", err)
	}

	if len(executed) == 0 {
		return nil, nil
	}

	// Get last batch
	lastBatch, err := m.driver.GetLastBatch(ctx, m.tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get last batch: %w", err)
	}

	// Find migrations to rollback
	var toRollback []driver.MigrationRecord
	if step > 0 {
		// Rollback specific number of migrations
		count := 0
		for i := len(executed) - 1; i >= 0 && count < step; i-- {
			toRollback = append(toRollback, executed[i])
			count++
		}
	} else {
		// Rollback last batch
		for _, r := range executed {
			if r.Batch == lastBatch {
				toRollback = append(toRollback, r)
			}
		}
	}

	// Reverse order for rollback
	sort.Slice(toRollback, func(i, j int) bool {
		return toRollback[i].Migration > toRollback[j].Migration
	})

	// Create migration map for lookup
	migrationMap := make(map[string]Migration)
	for _, migration := range m.migrations {
		migrationMap[migration.Name] = migration
	}

	// Execute rollbacks
	var rolledBack []string
	for _, record := range toRollback {
		migration, ok := migrationMap[record.Migration]
		if !ok {
			return rolledBack, fmt.Errorf("migration %s not found in registered migrations", record.Migration)
		}

		// Use transaction for databases that support transactional DDL
		if m.supportsTransactionalDDL() && !m.dryRun {
			if err := m.executeMigrationInTransaction(ctx, migration, 0, false); err != nil {
				return rolledBack, err
			}
		} else {
			executor := NewExecutor(m.driver, m.dryRun)
			if err := migration.Down(ctx, executor); err != nil {
				return rolledBack, fmt.Errorf("rollback of %s failed: %w", migration.Name, err)
			}
			if !m.dryRun {
				if err := m.driver.DeleteMigration(ctx, m.tableName, migration.Name); err != nil {
					return rolledBack, fmt.Errorf("failed to delete migration record %s: %w", migration.Name, err)
				}
			}
		}

		rolledBack = append(rolledBack, migration.Name)
	}

	return rolledBack, nil
}

// Reset rolls back all migrations
func (m *Migrator) Reset(ctx context.Context) ([]string, error) {
	executed, err := m.driver.GetExecutedMigrations(ctx, m.tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get executed migrations: %w", err)
	}

	return m.Down(ctx, len(executed))
}

// Refresh rolls back all migrations and re-runs them
func (m *Migrator) Refresh(ctx context.Context) ([]string, []string, error) {
	rolledBack, err := m.Reset(ctx)
	if err != nil {
		return rolledBack, nil, err
	}

	executed, err := m.Up(ctx, 0)
	return rolledBack, executed, err
}

// Status returns the status of all migrations
func (m *Migrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	// Ensure migrations table exists
	if err := m.driver.CreateMigrationsTable(ctx, m.tableName); err != nil {
		return nil, fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get executed migrations
	executed, err := m.driver.GetExecutedMigrations(ctx, m.tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get executed migrations: %w", err)
	}

	executedMap := make(map[string]driver.MigrationRecord)
	for _, r := range executed {
		executedMap[r.Migration] = r
	}

	// Build status list
	var statuses []MigrationStatus
	for _, migration := range m.migrations {
		status := MigrationStatus{
			Name: migration.Name,
		}

		if record, ok := executedMap[migration.Name]; ok {
			status.Ran = true
			status.Batch = record.Batch
			status.ExecutedAt = record.ExecutedAt.Format("2006-01-02 15:04:05")
		}

		statuses = append(statuses, status)
	}

	// Sort by name
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Name < statuses[j].Name
	})

	return statuses, nil
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Name       string
	Ran        bool
	Batch      int
	ExecutedAt string
}

// Executor provides the API for migration operations
type Executor struct {
	driver driver.Driver
	tx     driver.Transaction // optional transaction for transactional DDL
	dryRun bool
	sqls   []string
}

// NewExecutor creates a new executor
func NewExecutor(drv driver.Driver, dryRun bool) *Executor {
	return &Executor{
		driver: drv,
		tx:     nil,
		dryRun: dryRun,
		sqls:   make([]string, 0),
	}
}

// CreateTable creates a new table
func (e *Executor) CreateTable(ctx context.Context, name string, fn func(*schema.Table)) error {
	table := schema.NewTable(name)
	fn(table)

	sql := e.driver.Grammar().CompileCreate(table)

	if e.dryRun {
		e.sqls = append(e.sqls, sql)
		return nil
	}

	// Use transaction if available
	if e.tx != nil {
		if _, err := e.tx.Exec(ctx, sql); err != nil {
			return fmt.Errorf("failed to create table %s: %w", name, err)
		}
		// Create indexes separately for transaction mode
		for _, idx := range table.Indexes {
			if idx.Type != schema.IndexTypePrimary {
				idxSQL := e.driver.Grammar().CompileIndex(table.Name, idx)
				if _, err := e.tx.Exec(ctx, idxSQL); err != nil {
					return fmt.Errorf("failed to create index: %w", err)
				}
			}
		}
		return nil
	}

	return e.driver.CreateTable(ctx, table)
}

// AlterTable modifies an existing table
func (e *Executor) AlterTable(ctx context.Context, name string, fn func(*schema.Table)) error {
	table := schema.NewTable(name)
	table.IsAlter = true
	fn(table)

	sqls := e.driver.Grammar().CompileAlter(table)

	if e.dryRun {
		e.sqls = append(e.sqls, sqls...)
		return nil
	}

	// Use transaction if available
	if e.tx != nil {
		for _, sql := range sqls {
			if sql == "" {
				continue
			}
			if _, err := e.tx.Exec(ctx, sql); err != nil {
				return fmt.Errorf("failed to alter table %s: %w", name, err)
			}
		}
		return nil
	}

	return e.driver.AlterTable(ctx, table)
}

// DropTable drops a table
func (e *Executor) DropTable(ctx context.Context, name string) error {
	sql := e.driver.Grammar().CompileDrop(name)

	if e.dryRun {
		e.sqls = append(e.sqls, sql)
		return nil
	}

	// Use transaction if available
	if e.tx != nil {
		if _, err := e.tx.Exec(ctx, sql); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", name, err)
		}
		return nil
	}

	return e.driver.DropTable(ctx, name)
}

// DropTableIfExists drops a table if it exists
func (e *Executor) DropTableIfExists(ctx context.Context, name string) error {
	sql := e.driver.Grammar().CompileDropIfExists(name)

	if e.dryRun {
		e.sqls = append(e.sqls, sql)
		return nil
	}

	// Use transaction if available
	if e.tx != nil {
		if _, err := e.tx.Exec(ctx, sql); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", name, err)
		}
		return nil
	}

	return e.driver.DropTableIfExists(ctx, name)
}

// HasTable checks if a table exists
func (e *Executor) HasTable(ctx context.Context, name string) (bool, error) {
	// HasTable always uses the driver directly (read operation)
	return e.driver.HasTable(ctx, name)
}

// RenameTable renames a table
func (e *Executor) RenameTable(ctx context.Context, from, to string) error {
	sql := e.driver.Grammar().CompileRename(from, to)

	if e.dryRun {
		e.sqls = append(e.sqls, sql)
		return nil
	}

	// Use transaction if available
	if e.tx != nil {
		if _, err := e.tx.Exec(ctx, sql); err != nil {
			return fmt.Errorf("failed to rename table %s to %s: %w", from, to, err)
		}
		return nil
	}

	return e.driver.RenameTable(ctx, from, to)
}

// Raw executes raw SQL
func (e *Executor) Raw(ctx context.Context, sql string) error {
	if e.dryRun {
		e.sqls = append(e.sqls, sql)
		return nil
	}

	// Use transaction if available
	if e.tx != nil {
		_, err := e.tx.Exec(ctx, sql)
		return err
	}

	_, err := e.driver.Exec(ctx, sql)
	return err
}

// GetSQL returns the collected SQL statements (for dry run)
func (e *Executor) GetSQL() []string {
	return e.sqls
}

// NewTransactionExecutor creates a new executor that uses a transaction
func NewTransactionExecutor(drv driver.Driver, tx driver.Transaction) *Executor {
	return &Executor{
		driver: drv,
		tx:     tx,
		dryRun: false,
		sqls:   make([]string, 0),
	}
}
