package migrator

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/flyits/migro/pkg/driver"
	"github.com/flyits/migro/pkg/schema"
)

// 测试目标需求: Migrator 迁移执行器功能正确性
// 覆盖: NewMigrator, Register, Up, Down, Reset, Refresh, Status

// mockGrammar 模拟 Grammar 接口
type mockGrammar struct{}

func (g *mockGrammar) CompileCreate(table *schema.Table) string  { return "CREATE TABLE test" }
func (g *mockGrammar) CompileAlter(table *schema.Table) []string { return []string{"ALTER TABLE test"} }
func (g *mockGrammar) CompileDrop(name string) string            { return "DROP TABLE " + name }
func (g *mockGrammar) CompileDropIfExists(name string) string    { return "DROP TABLE IF EXISTS " + name }
func (g *mockGrammar) CompileRename(from, to string) string {
	return "RENAME TABLE " + from + " TO " + to
}
func (g *mockGrammar) CompileHasTable(name string) (string, error) { return "SELECT 1", nil }
func (g *mockGrammar) TypeString(length int) string                { return "VARCHAR(255)" }
func (g *mockGrammar) TypeText() string                            { return "TEXT" }
func (g *mockGrammar) TypeInteger() string                         { return "INT" }
func (g *mockGrammar) TypeBigInteger() string                      { return "BIGINT" }
func (g *mockGrammar) TypeSmallInteger() string                    { return "SMALLINT" }
func (g *mockGrammar) TypeTinyInteger() string                     { return "TINYINT" }
func (g *mockGrammar) TypeFloat() string                           { return "FLOAT" }
func (g *mockGrammar) TypeDouble() string                          { return "DOUBLE" }
func (g *mockGrammar) TypeDecimal(precision, scale int) string     { return "DECIMAL(10,2)" }
func (g *mockGrammar) TypeBoolean() string                         { return "BOOLEAN" }
func (g *mockGrammar) TypeDate() string                            { return "DATE" }
func (g *mockGrammar) TypeDateTime() string                        { return "DATETIME" }
func (g *mockGrammar) TypeTimestamp() string                       { return "TIMESTAMP" }
func (g *mockGrammar) TypeTime() string                            { return "TIME" }
func (g *mockGrammar) TypeJSON() string                            { return "JSON" }
func (g *mockGrammar) TypeBinary() string                          { return "BLOB" }
func (g *mockGrammar) TypeUUID() string                            { return "UUID" }
func (g *mockGrammar) CompileColumn(col *schema.Column) string     { return col.Name + " VARCHAR(255)" }
func (g *mockGrammar) CompileIndex(tableName string, idx *schema.Index) string {
	return "CREATE INDEX idx ON " + tableName
}
func (g *mockGrammar) CompileDropIndex(tableName, indexName string) string {
	return "DROP INDEX " + indexName
}
func (g *mockGrammar) CompileForeignKey(tableName string, fk *schema.ForeignKey) string {
	return "ALTER TABLE " + tableName + " ADD FOREIGN KEY"
}
func (g *mockGrammar) CompileDropForeignKey(tableName, fkName string) string {
	return "ALTER TABLE " + tableName + " DROP FOREIGN KEY " + fkName
}
func (g *mockGrammar) CompileCreateMigrationsTable(tableName string) string {
	return "CREATE TABLE " + tableName
}
func (g *mockGrammar) CompileGetMigrations(tableName string) string {
	return "SELECT * FROM " + tableName
}
func (g *mockGrammar) CompileInsertMigration(tableName string) string {
	return "INSERT INTO " + tableName
}
func (g *mockGrammar) CompileDeleteMigration(tableName string) string {
	return "DELETE FROM " + tableName
}
func (g *mockGrammar) CompileGetLastBatch(tableName string) string {
	return "SELECT MAX(batch) FROM " + tableName
}

// mockTransaction 模拟事务
type mockTransaction struct {
	committed  bool
	rolledBack bool
	execErr    error
}

func (t *mockTransaction) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, t.execErr
}
func (t *mockTransaction) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (t *mockTransaction) Commit() error {
	t.committed = true
	return nil
}
func (t *mockTransaction) Rollback() error {
	t.rolledBack = true
	return nil
}

// mockDriver 模拟驱动
type mockDriver struct {
	name               string
	grammar            driver.Grammar
	executedMigrations []driver.MigrationRecord
	lastBatch          int
	createTableErr     error
	execErr            error
	beginErr           error
	tx                 *mockTransaction
}

func newMockDriver(name string) *mockDriver {
	return &mockDriver{
		name:               name,
		grammar:            &mockGrammar{},
		executedMigrations: []driver.MigrationRecord{},
		lastBatch:          0,
		tx:                 &mockTransaction{},
	}
}

func (d *mockDriver) Connect(config *driver.Config) error { return nil }
func (d *mockDriver) Close() error                        { return nil }
func (d *mockDriver) DB() *sql.DB                         { return nil }
func (d *mockDriver) Begin(ctx context.Context) (driver.Transaction, error) {
	if d.beginErr != nil {
		return nil, d.beginErr
	}
	return d.tx, nil
}
func (d *mockDriver) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, d.execErr
}
func (d *mockDriver) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (d *mockDriver) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}
func (d *mockDriver) CreateTable(ctx context.Context, table *schema.Table) error {
	return d.createTableErr
}
func (d *mockDriver) AlterTable(ctx context.Context, table *schema.Table) error { return nil }
func (d *mockDriver) DropTable(ctx context.Context, name string) error          { return nil }
func (d *mockDriver) DropTableIfExists(ctx context.Context, name string) error  { return nil }
func (d *mockDriver) HasTable(ctx context.Context, name string) (bool, error)   { return false, nil }
func (d *mockDriver) RenameTable(ctx context.Context, from, to string) error    { return nil }
func (d *mockDriver) CreateMigrationsTable(ctx context.Context, tableName string) error {
	return nil
}
func (d *mockDriver) GetExecutedMigrations(ctx context.Context, tableName string) ([]driver.MigrationRecord, error) {
	return d.executedMigrations, nil
}
func (d *mockDriver) RecordMigration(ctx context.Context, tableName, migration string, batch int) error {
	d.executedMigrations = append(d.executedMigrations, driver.MigrationRecord{
		Migration:  migration,
		Batch:      batch,
		ExecutedAt: time.Now(),
	})
	d.lastBatch = batch
	return nil
}
func (d *mockDriver) DeleteMigration(ctx context.Context, tableName, migration string) error {
	for i, m := range d.executedMigrations {
		if m.Migration == migration {
			d.executedMigrations = append(d.executedMigrations[:i], d.executedMigrations[i+1:]...)
			break
		}
	}
	return nil
}
func (d *mockDriver) GetLastBatch(ctx context.Context, tableName string) (int, error) {
	return d.lastBatch, nil
}
func (d *mockDriver) Grammar() driver.Grammar { return d.grammar }
func (d *mockDriver) Name() string            { return d.name }

func TestNewMigrator(t *testing.T) {
	drv := newMockDriver("mysql")
	m := NewMigrator(drv, "./migrations", "migrations")

	if m == nil {
		t.Fatal("expected migrator, got nil")
	}
	if m.driver != drv {
		t.Error("driver not set correctly")
	}
	if m.migrationsPath != "./migrations" {
		t.Errorf("expected path './migrations', got '%s'", m.migrationsPath)
	}
	if m.tableName != "migrations" {
		t.Errorf("expected table 'migrations', got '%s'", m.tableName)
	}
}

func TestMigrator_SetDryRun(t *testing.T) {
	drv := newMockDriver("mysql")
	m := NewMigrator(drv, "./migrations", "migrations")

	m.SetDryRun(true)
	if !m.dryRun {
		t.Error("expected dryRun to be true")
	}

	m.SetDryRun(false)
	if m.dryRun {
		t.Error("expected dryRun to be false")
	}
}

func TestMigrator_Register(t *testing.T) {
	drv := newMockDriver("mysql")
	m := NewMigrator(drv, "./migrations", "migrations")

	migration := Migration{
		Name: "001_create_users",
		Up:   func(ctx context.Context, e *Executor) error { return nil },
		Down: func(ctx context.Context, e *Executor) error { return nil },
	}

	m.Register(migration)

	if len(m.migrations) != 1 {
		t.Errorf("expected 1 migration, got %d", len(m.migrations))
	}
	if m.migrations[0].Name != "001_create_users" {
		t.Errorf("expected name '001_create_users', got '%s'", m.migrations[0].Name)
	}
}

func TestMigrator_RegisterAll(t *testing.T) {
	drv := newMockDriver("mysql")
	m := NewMigrator(drv, "./migrations", "migrations")

	migrations := []Migration{
		{Name: "001_create_users", Up: func(ctx context.Context, e *Executor) error { return nil }},
		{Name: "002_create_posts", Up: func(ctx context.Context, e *Executor) error { return nil }},
		{Name: "003_create_comments", Up: func(ctx context.Context, e *Executor) error { return nil }},
	}

	m.RegisterAll(migrations)

	if len(m.migrations) != 3 {
		t.Errorf("expected 3 migrations, got %d", len(m.migrations))
	}
}

func TestMigrator_supportsTransactionalDDL(t *testing.T) {
	tests := []struct {
		driverName string
		expected   bool
	}{
		{"postgres", true},
		{"sqlite", true},
		{"mysql", false},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.driverName, func(t *testing.T) {
			drv := newMockDriver(tt.driverName)
			m := NewMigrator(drv, "./migrations", "migrations")

			result := m.supportsTransactionalDDL()
			if result != tt.expected {
				t.Errorf("expected %v for %s, got %v", tt.expected, tt.driverName, result)
			}
		})
	}
}

func TestMigrator_Up(t *testing.T) {
	t.Run("no pending migrations", func(t *testing.T) {
		drv := newMockDriver("mysql")
		m := NewMigrator(drv, "./migrations", "migrations")

		executed, err := m.Up(context.Background(), 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(executed) != 0 {
			t.Errorf("expected no executed migrations, got %d", len(executed))
		}
	})

	t.Run("run all pending migrations", func(t *testing.T) {
		drv := newMockDriver("mysql")
		m := NewMigrator(drv, "./migrations", "migrations")

		m.Register(Migration{
			Name: "001_create_users",
			Up:   func(ctx context.Context, e *Executor) error { return nil },
		})
		m.Register(Migration{
			Name: "002_create_posts",
			Up:   func(ctx context.Context, e *Executor) error { return nil },
		})

		executed, err := m.Up(context.Background(), 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(executed) != 2 {
			t.Errorf("expected 2 executed migrations, got %d", len(executed))
		}
	})

	t.Run("run with step limit", func(t *testing.T) {
		drv := newMockDriver("mysql")
		m := NewMigrator(drv, "./migrations", "migrations")

		m.Register(Migration{
			Name: "001_create_users",
			Up:   func(ctx context.Context, e *Executor) error { return nil },
		})
		m.Register(Migration{
			Name: "002_create_posts",
			Up:   func(ctx context.Context, e *Executor) error { return nil },
		})
		m.Register(Migration{
			Name: "003_create_comments",
			Up:   func(ctx context.Context, e *Executor) error { return nil },
		})

		executed, err := m.Up(context.Background(), 2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(executed) != 2 {
			t.Errorf("expected 2 executed migrations, got %d", len(executed))
		}
	})

	t.Run("skip already executed migrations", func(t *testing.T) {
		drv := newMockDriver("mysql")
		drv.executedMigrations = []driver.MigrationRecord{
			{Migration: "001_create_users", Batch: 1},
		}
		drv.lastBatch = 1

		m := NewMigrator(drv, "./migrations", "migrations")
		m.Register(Migration{
			Name: "001_create_users",
			Up:   func(ctx context.Context, e *Executor) error { return nil },
		})
		m.Register(Migration{
			Name: "002_create_posts",
			Up:   func(ctx context.Context, e *Executor) error { return nil },
		})

		executed, err := m.Up(context.Background(), 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(executed) != 1 {
			t.Errorf("expected 1 executed migration, got %d", len(executed))
		}
		if executed[0] != "002_create_posts" {
			t.Errorf("expected '002_create_posts', got '%s'", executed[0])
		}
	})

	t.Run("migration failure stops execution", func(t *testing.T) {
		drv := newMockDriver("mysql")
		m := NewMigrator(drv, "./migrations", "migrations")

		m.Register(Migration{
			Name: "001_create_users",
			Up:   func(ctx context.Context, e *Executor) error { return nil },
		})
		m.Register(Migration{
			Name: "002_failing_migration",
			Up:   func(ctx context.Context, e *Executor) error { return errors.New("migration failed") },
		})
		m.Register(Migration{
			Name: "003_create_comments",
			Up:   func(ctx context.Context, e *Executor) error { return nil },
		})

		executed, err := m.Up(context.Background(), 0)
		if err == nil {
			t.Error("expected error for failing migration")
		}
		if len(executed) != 1 {
			t.Errorf("expected 1 executed migration before failure, got %d", len(executed))
		}
	})

	t.Run("dry run mode", func(t *testing.T) {
		drv := newMockDriver("mysql")
		m := NewMigrator(drv, "./migrations", "migrations")
		m.SetDryRun(true)

		m.Register(Migration{
			Name: "001_create_users",
			Up:   func(ctx context.Context, e *Executor) error { return nil },
		})

		executed, err := m.Up(context.Background(), 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(executed) != 1 {
			t.Errorf("expected 1 executed migration, got %d", len(executed))
		}
		// 验证没有实际记录迁移
		if len(drv.executedMigrations) != 0 {
			t.Error("expected no migrations recorded in dry run mode")
		}
	})
}

func TestMigrator_Down(t *testing.T) {
	t.Run("no migrations to rollback", func(t *testing.T) {
		drv := newMockDriver("mysql")
		m := NewMigrator(drv, "./migrations", "migrations")

		rolledBack, err := m.Down(context.Background(), 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(rolledBack) != 0 {
			t.Errorf("expected no rolled back migrations, got %d", len(rolledBack))
		}
	})

	t.Run("rollback last batch", func(t *testing.T) {
		drv := newMockDriver("mysql")
		drv.executedMigrations = []driver.MigrationRecord{
			{Migration: "001_create_users", Batch: 1},
			{Migration: "002_create_posts", Batch: 2},
			{Migration: "003_create_comments", Batch: 2},
		}
		drv.lastBatch = 2

		m := NewMigrator(drv, "./migrations", "migrations")
		m.Register(Migration{
			Name: "001_create_users",
			Down: func(ctx context.Context, e *Executor) error { return nil },
		})
		m.Register(Migration{
			Name: "002_create_posts",
			Down: func(ctx context.Context, e *Executor) error { return nil },
		})
		m.Register(Migration{
			Name: "003_create_comments",
			Down: func(ctx context.Context, e *Executor) error { return nil },
		})

		rolledBack, err := m.Down(context.Background(), 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(rolledBack) != 2 {
			t.Errorf("expected 2 rolled back migrations, got %d", len(rolledBack))
		}
	})

	t.Run("rollback with step limit", func(t *testing.T) {
		drv := newMockDriver("mysql")
		drv.executedMigrations = []driver.MigrationRecord{
			{Migration: "001_create_users", Batch: 1},
			{Migration: "002_create_posts", Batch: 1},
			{Migration: "003_create_comments", Batch: 1},
		}
		drv.lastBatch = 1

		m := NewMigrator(drv, "./migrations", "migrations")
		m.Register(Migration{
			Name: "001_create_users",
			Down: func(ctx context.Context, e *Executor) error { return nil },
		})
		m.Register(Migration{
			Name: "002_create_posts",
			Down: func(ctx context.Context, e *Executor) error { return nil },
		})
		m.Register(Migration{
			Name: "003_create_comments",
			Down: func(ctx context.Context, e *Executor) error { return nil },
		})

		rolledBack, err := m.Down(context.Background(), 2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(rolledBack) != 2 {
			t.Errorf("expected 2 rolled back migrations, got %d", len(rolledBack))
		}
	})

	t.Run("migration not found error", func(t *testing.T) {
		drv := newMockDriver("mysql")
		drv.executedMigrations = []driver.MigrationRecord{
			{Migration: "001_unknown_migration", Batch: 1},
		}
		drv.lastBatch = 1

		m := NewMigrator(drv, "./migrations", "migrations")
		// 不注册任何迁移

		_, err := m.Down(context.Background(), 0)
		if err == nil {
			t.Error("expected error for unknown migration")
		}
	})
}

func TestMigrator_Reset(t *testing.T) {
	t.Run("reset all migrations", func(t *testing.T) {
		drv := newMockDriver("mysql")
		drv.executedMigrations = []driver.MigrationRecord{
			{Migration: "001_create_users", Batch: 1},
			{Migration: "002_create_posts", Batch: 2},
		}
		drv.lastBatch = 2

		m := NewMigrator(drv, "./migrations", "migrations")
		m.Register(Migration{
			Name: "001_create_users",
			Down: func(ctx context.Context, e *Executor) error { return nil },
		})
		m.Register(Migration{
			Name: "002_create_posts",
			Down: func(ctx context.Context, e *Executor) error { return nil },
		})

		rolledBack, err := m.Reset(context.Background())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(rolledBack) != 2 {
			t.Errorf("expected 2 rolled back migrations, got %d", len(rolledBack))
		}
	})
}

func TestMigrator_Refresh(t *testing.T) {
	t.Run("refresh migrations", func(t *testing.T) {
		drv := newMockDriver("mysql")
		drv.executedMigrations = []driver.MigrationRecord{
			{Migration: "001_create_users", Batch: 1},
		}
		drv.lastBatch = 1

		m := NewMigrator(drv, "./migrations", "migrations")
		m.Register(Migration{
			Name: "001_create_users",
			Up:   func(ctx context.Context, e *Executor) error { return nil },
			Down: func(ctx context.Context, e *Executor) error { return nil },
		})
		m.Register(Migration{
			Name: "002_create_posts",
			Up:   func(ctx context.Context, e *Executor) error { return nil },
			Down: func(ctx context.Context, e *Executor) error { return nil },
		})

		rolledBack, executed, err := m.Refresh(context.Background())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(rolledBack) != 1 {
			t.Errorf("expected 1 rolled back migration, got %d", len(rolledBack))
		}
		if len(executed) != 2 {
			t.Errorf("expected 2 executed migrations, got %d", len(executed))
		}
	})
}

func TestMigrator_Status(t *testing.T) {
	t.Run("status with mixed migrations", func(t *testing.T) {
		drv := newMockDriver("mysql")
		drv.executedMigrations = []driver.MigrationRecord{
			{Migration: "001_create_users", Batch: 1, ExecutedAt: time.Now()},
		}

		m := NewMigrator(drv, "./migrations", "migrations")
		m.Register(Migration{Name: "001_create_users"})
		m.Register(Migration{Name: "002_create_posts"})

		statuses, err := m.Status(context.Background())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(statuses) != 2 {
			t.Errorf("expected 2 statuses, got %d", len(statuses))
		}

		// 验证第一个迁移已执行
		if !statuses[0].Ran {
			t.Error("expected first migration to be ran")
		}
		// 验证第二个迁移未执行
		if statuses[1].Ran {
			t.Error("expected second migration to not be ran")
		}
	})
}

func TestExecutor_DryRun(t *testing.T) {
	drv := newMockDriver("mysql")
	e := NewExecutor(drv, true)

	ctx := context.Background()

	t.Run("CreateTable in dry run", func(t *testing.T) {
		err := e.CreateTable(ctx, "users", func(t *schema.Table) {
			t.ID()
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		sqls := e.GetSQL()
		if len(sqls) == 0 {
			t.Error("expected SQL to be collected")
		}
	})

	t.Run("DropTable in dry run", func(t *testing.T) {
		e := NewExecutor(drv, true)
		err := e.DropTable(ctx, "users")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		sqls := e.GetSQL()
		if len(sqls) == 0 {
			t.Error("expected SQL to be collected")
		}
	})

	t.Run("Raw in dry run", func(t *testing.T) {
		e := NewExecutor(drv, true)
		err := e.Raw(ctx, "SELECT 1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		sqls := e.GetSQL()
		if len(sqls) == 0 {
			t.Error("expected SQL to be collected")
		}
	})
}

func TestExecutor_Operations(t *testing.T) {
	drv := newMockDriver("mysql")
	ctx := context.Background()

	t.Run("CreateTable", func(t *testing.T) {
		e := NewExecutor(drv, false)
		err := e.CreateTable(ctx, "users", func(t *schema.Table) {
			t.ID()
			t.String("name", 100)
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("AlterTable", func(t *testing.T) {
		e := NewExecutor(drv, false)
		err := e.AlterTable(ctx, "users", func(t *schema.Table) {
			t.String("email", 100)
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("DropTable", func(t *testing.T) {
		e := NewExecutor(drv, false)
		err := e.DropTable(ctx, "users")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("DropTableIfExists", func(t *testing.T) {
		e := NewExecutor(drv, false)
		err := e.DropTableIfExists(ctx, "users")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("HasTable", func(t *testing.T) {
		e := NewExecutor(drv, false)
		exists, err := e.HasTable(ctx, "users")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if exists {
			t.Error("expected table to not exist")
		}
	})

	t.Run("RenameTable", func(t *testing.T) {
		e := NewExecutor(drv, false)
		err := e.RenameTable(ctx, "old_users", "users")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("Raw", func(t *testing.T) {
		e := NewExecutor(drv, false)
		err := e.Raw(ctx, "SELECT 1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestNewTransactionExecutor(t *testing.T) {
	drv := newMockDriver("postgres")
	tx := &mockTransaction{}

	e := NewTransactionExecutor(drv, tx)

	if e == nil {
		t.Fatal("expected executor, got nil")
	}
	if e.tx != tx {
		t.Error("transaction not set correctly")
	}
	if e.dryRun {
		t.Error("expected dryRun to be false")
	}
}
