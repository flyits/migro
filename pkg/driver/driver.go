package driver

import (
	"context"
	"database/sql"
	"time"

	"github.com/migro/migro/pkg/schema"
)

// MigrationRecord represents a record in the migrations table
type MigrationRecord struct {
	ID         int64
	Migration  string
	Batch      int
	ExecutedAt time.Time
}

// Config holds database connection configuration
type Config struct {
	Driver   string
	Host     string
	Port     int
	Database string
	Username string
	Password string
	Charset  string
	Options  map[string]string
}

// Transaction represents a database transaction
type Transaction interface {
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	Commit() error
	Rollback() error
}

// Grammar defines the interface for SQL dialect generation
type Grammar interface {
	// Table operations
	CompileCreate(table *schema.Table) string
	CompileAlter(table *schema.Table) []string
	CompileDrop(name string) string
	CompileDropIfExists(name string) string
	CompileRename(from, to string) string
	CompileHasTable(name string) (string, error)

	// Type mappings
	TypeString(length int) string
	TypeText() string
	TypeInteger() string
	TypeBigInteger() string
	TypeSmallInteger() string
	TypeTinyInteger() string
	TypeFloat() string
	TypeDouble() string
	TypeDecimal(precision, scale int) string
	TypeBoolean() string
	TypeDate() string
	TypeDateTime() string
	TypeTimestamp() string
	TypeTime() string
	TypeJSON() string
	TypeBinary() string
	TypeUUID() string

	// Column modifiers
	CompileColumn(col *schema.Column) string

	// Index operations
	CompileIndex(tableName string, idx *schema.Index) string
	CompileDropIndex(tableName, indexName string) string

	// Foreign key operations
	CompileForeignKey(tableName string, fk *schema.ForeignKey) string
	CompileDropForeignKey(tableName, fkName string) string

	// Migrations table
	CompileCreateMigrationsTable(tableName string) string
	CompileGetMigrations(tableName string) string
	CompileInsertMigration(tableName string) string
	CompileDeleteMigration(tableName string) string
	CompileGetLastBatch(tableName string) string
}

// Driver defines the interface for database drivers
type Driver interface {
	// Connection management
	Connect(config *Config) error
	Close() error
	DB() *sql.DB

	// Transaction management
	Begin(ctx context.Context) (Transaction, error)

	// Direct execution
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row

	// Schema operations
	CreateTable(ctx context.Context, table *schema.Table) error
	AlterTable(ctx context.Context, table *schema.Table) error
	DropTable(ctx context.Context, name string) error
	DropTableIfExists(ctx context.Context, name string) error
	HasTable(ctx context.Context, name string) (bool, error)
	RenameTable(ctx context.Context, from, to string) error

	// Migration history
	CreateMigrationsTable(ctx context.Context, tableName string) error
	GetExecutedMigrations(ctx context.Context, tableName string) ([]MigrationRecord, error)
	RecordMigration(ctx context.Context, tableName, migration string, batch int) error
	DeleteMigration(ctx context.Context, tableName, migration string) error
	GetLastBatch(ctx context.Context, tableName string) (int, error)

	// Grammar access
	Grammar() Grammar

	// Driver name
	Name() string
}
