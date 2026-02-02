package sqlite

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/migro/migro/pkg/schema"
)

// validIdentifier validates that a name contains only safe characters
// to prevent SQL injection. Allows letters, numbers, and underscores.
var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// validateIdentifier checks if the given name is a valid SQL identifier
func validateIdentifier(name string) error {
	if name == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	if len(name) > 128 { // SQLite doesn't have a strict limit, but we set a reasonable one
		return fmt.Errorf("identifier too long: max 128 characters")
	}
	if !validIdentifier.MatchString(name) {
		return fmt.Errorf("invalid identifier %q: must contain only letters, numbers, and underscores, and start with a letter or underscore", name)
	}
	return nil
}

// Grammar implements the SQLite SQL dialect
type Grammar struct{}

// NewGrammar creates a new SQLite grammar instance
func NewGrammar() *Grammar {
	return &Grammar{}
}

// CompileCreate generates CREATE TABLE SQL
func (g *Grammar) CompileCreate(table *schema.Table) string {
	var sb strings.Builder

	sb.WriteString("CREATE TABLE ")
	if table.IfNotExists {
		sb.WriteString("IF NOT EXISTS ")
	}
	sb.WriteString(g.wrapTable(table.Name))
	sb.WriteString(" (\n")

	// Columns
	columns := make([]string, 0, len(table.Columns))
	for _, col := range table.Columns {
		columns = append(columns, "  "+g.CompileColumn(col))
	}

	// Primary key from columns marked as primary (if not auto-increment)
	var primaryCols []string
	for _, col := range table.Columns {
		if col.IsPrimary && !col.IsAutoIncrement {
			primaryCols = append(primaryCols, g.wrap(col.Name))
		}
	}
	if len(primaryCols) > 0 {
		columns = append(columns, fmt.Sprintf("  PRIMARY KEY (%s)", strings.Join(primaryCols, ", ")))
	}

	// Foreign keys inline
	for _, fk := range table.ForeignKeys {
		columns = append(columns, "  "+g.compileForeignKeyInline(fk))
	}

	sb.WriteString(strings.Join(columns, ",\n"))
	sb.WriteString("\n)")

	return sb.String()
}

// CompileAlter generates ALTER TABLE SQL statements
// Note: SQLite has limited ALTER TABLE support
func (g *Grammar) CompileAlter(table *schema.Table) []string {
	var statements []string
	tableName := g.wrapTable(table.Name)

	// SQLite only supports ADD COLUMN and RENAME COLUMN
	// For other operations, we need to recreate the table

	// Rename columns (SQLite 3.25.0+)
	for oldName, newName := range table.RenameColumns {
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", tableName, g.wrap(oldName), g.wrap(newName)))
	}

	// Add columns
	for _, col := range table.Columns {
		if !col.Change {
			statements = append(statements, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, g.CompileColumn(col)))
		}
	}

	// Add indexes
	for _, idx := range table.Indexes {
		statements = append(statements, g.CompileIndex(table.Name, idx))
	}

	// Note: DROP COLUMN, MODIFY COLUMN, DROP INDEX, DROP FOREIGN KEY
	// require table recreation in SQLite, which is not implemented here
	// Users should use Raw SQL for complex alterations

	return statements
}

// CompileDrop generates DROP TABLE SQL
func (g *Grammar) CompileDrop(name string) string {
	return fmt.Sprintf("DROP TABLE %s", g.wrapTable(name))
}

// CompileDropIfExists generates DROP TABLE IF EXISTS SQL
func (g *Grammar) CompileDropIfExists(name string) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s", g.wrapTable(name))
}

// CompileRename generates ALTER TABLE RENAME SQL
func (g *Grammar) CompileRename(from, to string) string {
	return fmt.Sprintf("ALTER TABLE %s RENAME TO %s", g.wrapTable(from), g.wrapTable(to))
}

// CompileHasTable generates SQL to check if table exists
// The table name is validated to prevent SQL injection
func (g *Grammar) CompileHasTable(name string) (string, error) {
	if err := validateIdentifier(name); err != nil {
		return "", fmt.Errorf("invalid table name: %w", err)
	}
	return fmt.Sprintf("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='%s'", name), nil
}

// Type mappings - SQLite uses type affinity

func (g *Grammar) TypeString(length int) string {
	return "TEXT"
}

func (g *Grammar) TypeText() string {
	return "TEXT"
}

func (g *Grammar) TypeInteger() string {
	return "INTEGER"
}

func (g *Grammar) TypeBigInteger() string {
	return "INTEGER"
}

func (g *Grammar) TypeSmallInteger() string {
	return "INTEGER"
}

func (g *Grammar) TypeTinyInteger() string {
	return "INTEGER"
}

func (g *Grammar) TypeFloat() string {
	return "REAL"
}

func (g *Grammar) TypeDouble() string {
	return "REAL"
}

func (g *Grammar) TypeDecimal(precision, scale int) string {
	return "REAL"
}

func (g *Grammar) TypeBoolean() string {
	return "INTEGER"
}

func (g *Grammar) TypeDate() string {
	return "TEXT"
}

func (g *Grammar) TypeDateTime() string {
	return "TEXT"
}

func (g *Grammar) TypeTimestamp() string {
	return "TEXT"
}

func (g *Grammar) TypeTime() string {
	return "TEXT"
}

func (g *Grammar) TypeJSON() string {
	return "TEXT"
}

func (g *Grammar) TypeBinary() string {
	return "BLOB"
}

func (g *Grammar) TypeUUID() string {
	return "TEXT"
}

// CompileColumn generates column definition SQL
func (g *Grammar) CompileColumn(col *schema.Column) string {
	var sb strings.Builder

	sb.WriteString(g.wrap(col.Name))
	sb.WriteString(" ")

	// Handle auto-increment with INTEGER PRIMARY KEY
	if col.IsAutoIncrement {
		sb.WriteString("INTEGER PRIMARY KEY AUTOINCREMENT")
		return sb.String()
	}

	sb.WriteString(g.getColumnType(col))

	if col.IsPrimary {
		sb.WriteString(" PRIMARY KEY")
	}

	if !col.IsNullable {
		sb.WriteString(" NOT NULL")
	}

	if col.DefaultValue != nil {
		sb.WriteString(" DEFAULT ")
		sb.WriteString(g.formatDefault(col.DefaultValue))
	}

	if col.IsUnique {
		sb.WriteString(" UNIQUE")
	}

	return sb.String()
}

func (g *Grammar) getColumnType(col *schema.Column) string {
	switch col.Type {
	case schema.TypeString, schema.TypeText, schema.TypeDate, schema.TypeDateTime,
		schema.TypeTimestamp, schema.TypeTime, schema.TypeJSON, schema.TypeUUID:
		return "TEXT"
	case schema.TypeInteger, schema.TypeBigInteger, schema.TypeSmallInteger,
		schema.TypeTinyInteger, schema.TypeBoolean:
		return "INTEGER"
	case schema.TypeFloat, schema.TypeDouble, schema.TypeDecimal:
		return "REAL"
	case schema.TypeBinary:
		return "BLOB"
	default:
		return "TEXT"
	}
}

// CompileIndex generates CREATE INDEX SQL
func (g *Grammar) CompileIndex(tableName string, idx *schema.Index) string {
	indexName := idx.Name
	if indexName == "" {
		indexName = g.generateIndexName(tableName, idx.Columns, idx.Type)
	}

	cols := make([]string, len(idx.Columns))
	for i, col := range idx.Columns {
		cols[i] = g.wrap(col)
	}

	switch idx.Type {
	case schema.IndexTypeUnique:
		return fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)", g.wrap(indexName), g.wrapTable(tableName), strings.Join(cols, ", "))
	default:
		return fmt.Sprintf("CREATE INDEX %s ON %s (%s)", g.wrap(indexName), g.wrapTable(tableName), strings.Join(cols, ", "))
	}
}

// CompileDropIndex generates DROP INDEX SQL
func (g *Grammar) CompileDropIndex(tableName, indexName string) string {
	return fmt.Sprintf("DROP INDEX %s", g.wrap(indexName))
}

// CompileForeignKey generates foreign key SQL (not supported as ALTER in SQLite)
func (g *Grammar) CompileForeignKey(tableName string, fk *schema.ForeignKey) string {
	// SQLite doesn't support adding foreign keys via ALTER TABLE
	// This is only used for inline definition in CREATE TABLE
	return ""
}

// CompileDropForeignKey - SQLite doesn't support dropping foreign keys
func (g *Grammar) CompileDropForeignKey(tableName, fkName string) string {
	return ""
}

// Migrations table operations

func (g *Grammar) CompileCreateMigrationsTable(tableName string) string {
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  migration TEXT NOT NULL,
  batch INTEGER NOT NULL,
  executed_at TEXT DEFAULT CURRENT_TIMESTAMP
)`, g.wrapTable(tableName))
}

func (g *Grammar) CompileGetMigrations(tableName string) string {
	return fmt.Sprintf("SELECT id, migration, batch, executed_at FROM %s ORDER BY batch, migration", g.wrapTable(tableName))
}

func (g *Grammar) CompileInsertMigration(tableName string) string {
	return fmt.Sprintf("INSERT INTO %s (migration, batch) VALUES (?, ?)", g.wrapTable(tableName))
}

func (g *Grammar) CompileDeleteMigration(tableName string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE migration = ?", g.wrapTable(tableName))
}

func (g *Grammar) CompileGetLastBatch(tableName string) string {
	return fmt.Sprintf("SELECT COALESCE(MAX(batch), 0) FROM %s", g.wrapTable(tableName))
}

// Helper methods

func (g *Grammar) wrap(name string) string {
	return "\"" + name + "\""
}

func (g *Grammar) wrapTable(name string) string {
	return "\"" + name + "\""
}

func (g *Grammar) compileForeignKeyInline(fk *schema.ForeignKey) string {
	cols := make([]string, len(fk.Columns))
	for i, col := range fk.Columns {
		cols[i] = g.wrap(col)
	}

	sql := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s)",
		strings.Join(cols, ", "),
		g.wrapTable(fk.ReferenceTable),
		g.wrap(fk.ReferenceColumn))

	if fk.OnDelete != "" {
		sql += " ON DELETE " + string(fk.OnDelete)
	}
	if fk.OnUpdate != "" {
		sql += " ON UPDATE " + string(fk.OnUpdate)
	}

	return sql
}

func (g *Grammar) generateIndexName(tableName string, columns []string, indexType schema.IndexType) string {
	suffix := "idx"
	if indexType == schema.IndexTypeUnique {
		suffix = "unique"
	}
	return fmt.Sprintf("%s_%s_%s", tableName, strings.Join(columns, "_"), suffix)
}

func (g *Grammar) formatDefault(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("'%s'", escapeString(v))
	case bool:
		if v {
			return "1"
		}
		return "0"
	case nil:
		return "NULL"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func escapeString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
