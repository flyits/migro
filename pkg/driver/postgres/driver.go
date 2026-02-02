package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/migro/migro/pkg/driver"
	"github.com/migro/migro/pkg/schema"
)

func init() {
	driver.Register("postgres", func() driver.Driver {
		return NewDriver()
	})
}

// Driver implements the PostgreSQL database driver
type Driver struct {
	db      *sql.DB
	grammar *Grammar
}

// NewDriver creates a new PostgreSQL driver instance
func NewDriver() *Driver {
	return &Driver{
		grammar: NewGrammar(),
	}
}

// Connect establishes a connection to the PostgreSQL database
func (d *Driver) Connect(config *driver.Config) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.Database,
	)

	// Add additional options
	for key, value := range config.Options {
		dsn += " " + key + "=" + value
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("postgres: failed to open connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("postgres: failed to ping database: %w", err)
	}

	d.db = db
	return nil
}

// Close closes the database connection
func (d *Driver) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// DB returns the underlying sql.DB instance
func (d *Driver) DB() *sql.DB {
	return d.db
}

// Begin starts a new transaction
func (d *Driver) Begin(ctx context.Context) (driver.Transaction, error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to begin transaction: %w", err)
	}
	return &transaction{tx: tx}, nil
}

// Exec executes a query without returning rows
func (d *Driver) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows
func (d *Driver) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (d *Driver) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

// CreateTable creates a new table
func (d *Driver) CreateTable(ctx context.Context, table *schema.Table) error {
	sql := d.grammar.CompileCreate(table)
	_, err := d.db.ExecContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("postgres: failed to create table %s: %w", table.Name, err)
	}

	// Create indexes separately (PostgreSQL doesn't support inline index definitions)
	for _, idx := range table.Indexes {
		if idx.Type != schema.IndexTypePrimary {
			idxSQL := d.grammar.CompileIndex(table.Name, idx)
			if _, err := d.db.ExecContext(ctx, idxSQL); err != nil {
				return fmt.Errorf("postgres: failed to create index: %w", err)
			}
		}
	}

	return nil
}

// AlterTable modifies an existing table
func (d *Driver) AlterTable(ctx context.Context, table *schema.Table) error {
	statements := d.grammar.CompileAlter(table)
	for _, stmt := range statements {
		if _, err := d.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("postgres: failed to alter table %s: %w", table.Name, err)
		}
	}
	return nil
}

// DropTable drops a table
func (d *Driver) DropTable(ctx context.Context, name string) error {
	sql := d.grammar.CompileDrop(name)
	_, err := d.db.ExecContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("postgres: failed to drop table %s: %w", name, err)
	}
	return nil
}

// DropTableIfExists drops a table if it exists
func (d *Driver) DropTableIfExists(ctx context.Context, name string) error {
	sql := d.grammar.CompileDropIfExists(name)
	_, err := d.db.ExecContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("postgres: failed to drop table %s: %w", name, err)
	}
	return nil
}

// HasTable checks if a table exists
func (d *Driver) HasTable(ctx context.Context, name string) (bool, error) {
	sql, err := d.grammar.CompileHasTable(name)
	if err != nil {
		return false, fmt.Errorf("postgres: %w", err)
	}
	var count int
	err = d.db.QueryRowContext(ctx, sql).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("postgres: failed to check table existence: %w", err)
	}
	return count > 0, nil
}

// RenameTable renames a table
func (d *Driver) RenameTable(ctx context.Context, from, to string) error {
	sql := d.grammar.CompileRename(from, to)
	_, err := d.db.ExecContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("postgres: failed to rename table %s to %s: %w", from, to, err)
	}
	return nil
}

// CreateMigrationsTable creates the migrations tracking table
func (d *Driver) CreateMigrationsTable(ctx context.Context, tableName string) error {
	sql := d.grammar.CompileCreateMigrationsTable(tableName)
	_, err := d.db.ExecContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("postgres: failed to create migrations table: %w", err)
	}
	return nil
}

// GetExecutedMigrations returns all executed migrations
func (d *Driver) GetExecutedMigrations(ctx context.Context, tableName string) ([]driver.MigrationRecord, error) {
	sql := d.grammar.CompileGetMigrations(tableName)
	rows, err := d.db.QueryContext(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to get migrations: %w", err)
	}
	defer rows.Close()

	var records []driver.MigrationRecord
	for rows.Next() {
		var r driver.MigrationRecord
		if err := rows.Scan(&r.ID, &r.Migration, &r.Batch, &r.ExecutedAt); err != nil {
			return nil, fmt.Errorf("postgres: failed to scan migration record: %w", err)
		}
		records = append(records, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: error iterating migrations: %w", err)
	}

	return records, nil
}

// RecordMigration records a migration execution
func (d *Driver) RecordMigration(ctx context.Context, tableName, migration string, batch int) error {
	sql := d.grammar.CompileInsertMigration(tableName)
	_, err := d.db.ExecContext(ctx, sql, migration, batch)
	if err != nil {
		return fmt.Errorf("postgres: failed to record migration: %w", err)
	}
	return nil
}

// DeleteMigration removes a migration record
func (d *Driver) DeleteMigration(ctx context.Context, tableName, migration string) error {
	sql := d.grammar.CompileDeleteMigration(tableName)
	_, err := d.db.ExecContext(ctx, sql, migration)
	if err != nil {
		return fmt.Errorf("postgres: failed to delete migration: %w", err)
	}
	return nil
}

// GetLastBatch returns the last batch number
func (d *Driver) GetLastBatch(ctx context.Context, tableName string) (int, error) {
	sql := d.grammar.CompileGetLastBatch(tableName)
	var batch int
	err := d.db.QueryRowContext(ctx, sql).Scan(&batch)
	if err != nil {
		return 0, fmt.Errorf("postgres: failed to get last batch: %w", err)
	}
	return batch, nil
}

// Grammar returns the PostgreSQL grammar instance
func (d *Driver) Grammar() driver.Grammar {
	return d.grammar
}

// Name returns the driver name
func (d *Driver) Name() string {
	return "postgres"
}

// transaction wraps sql.Tx to implement driver.Transaction
type transaction struct {
	tx *sql.Tx
}

func (t *transaction) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *transaction) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *transaction) Commit() error {
	return t.tx.Commit()
}

func (t *transaction) Rollback() error {
	return t.tx.Rollback()
}
