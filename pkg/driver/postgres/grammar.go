package postgres

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/flyits/migro/pkg/schema"
)

// validIdentifier validates that a name contains only safe characters
// to prevent SQL injection. Allows letters, numbers, and underscores.
var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// validateIdentifier checks if the given name is a valid SQL identifier
func validateIdentifier(name string) error {
	if name == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	if len(name) > 63 { // PostgreSQL limit is 63 characters
		return fmt.Errorf("identifier too long: max 63 characters")
	}
	if !validIdentifier.MatchString(name) {
		return fmt.Errorf("invalid identifier %q: must contain only letters, numbers, and underscores, and start with a letter or underscore", name)
	}
	return nil
}

// Grammar implements the PostgreSQL SQL dialect
type Grammar struct{}

// NewGrammar creates a new PostgreSQL grammar instance
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

	// Primary key from columns marked as primary
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
func (g *Grammar) CompileAlter(table *schema.Table) []string {
	var statements []string
	tableName := g.wrapTable(table.Name)

	// Drop foreign keys first
	for _, fkName := range table.DropForeignKeys {
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", tableName, g.wrap(fkName)))
	}

	// Drop indexes
	for _, idxName := range table.DropIndexes {
		statements = append(statements, fmt.Sprintf("DROP INDEX %s", g.wrap(idxName)))
	}

	// Drop columns
	for _, colName := range table.DropColumns {
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, g.wrap(colName)))
	}

	// Rename columns
	for oldName, newName := range table.RenameColumns {
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", tableName, g.wrap(oldName), g.wrap(newName)))
	}

	// Add/modify columns
	for _, col := range table.Columns {
		if col.Change {
			// PostgreSQL requires separate statements for each change
			statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s", tableName, g.wrap(col.Name), g.getColumnType(col)))
			if col.IsNullable {
				statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL", tableName, g.wrap(col.Name)))
			} else {
				statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL", tableName, g.wrap(col.Name)))
			}
			if col.DefaultValue != nil {
				statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s", tableName, g.wrap(col.Name), g.formatDefault(col.DefaultValue)))
			}
		} else {
			statements = append(statements, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, g.CompileColumn(col)))
		}
	}

	// Add indexes
	for _, idx := range table.Indexes {
		statements = append(statements, g.CompileIndex(table.Name, idx))
	}

	// Add foreign keys
	for _, fk := range table.ForeignKeys {
		statements = append(statements, g.CompileForeignKey(table.Name, fk))
	}

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
	return fmt.Sprintf("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '%s'", name), nil
}

// Type mappings

func (g *Grammar) TypeString(length int) string {
	if length <= 0 {
		length = 255
	}
	return fmt.Sprintf("VARCHAR(%d)", length)
}

func (g *Grammar) TypeText() string {
	return "TEXT"
}

func (g *Grammar) TypeInteger() string {
	return "INTEGER"
}

func (g *Grammar) TypeBigInteger() string {
	return "BIGINT"
}

func (g *Grammar) TypeSmallInteger() string {
	return "SMALLINT"
}

func (g *Grammar) TypeTinyInteger() string {
	return "SMALLINT" // PostgreSQL doesn't have TINYINT
}

func (g *Grammar) TypeFloat() string {
	return "REAL"
}

func (g *Grammar) TypeDouble() string {
	return "DOUBLE PRECISION"
}

func (g *Grammar) TypeDecimal(precision, scale int) string {
	return fmt.Sprintf("DECIMAL(%d,%d)", precision, scale)
}

func (g *Grammar) TypeBoolean() string {
	return "BOOLEAN"
}

func (g *Grammar) TypeDate() string {
	return "DATE"
}

func (g *Grammar) TypeDateTime() string {
	return "TIMESTAMP"
}

func (g *Grammar) TypeTimestamp() string {
	return "TIMESTAMP"
}

func (g *Grammar) TypeTime() string {
	return "TIME"
}

func (g *Grammar) TypeJSON() string {
	return "JSONB"
}

func (g *Grammar) TypeBinary() string {
	return "BYTEA"
}

func (g *Grammar) TypeUUID() string {
	return "UUID"
}

// CompileColumn generates column definition SQL
func (g *Grammar) CompileColumn(col *schema.Column) string {
	var sb strings.Builder

	sb.WriteString(g.wrap(col.Name))
	sb.WriteString(" ")

	// Handle auto-increment with SERIAL types
	if col.IsAutoIncrement {
		if col.Type == schema.TypeBigInteger {
			sb.WriteString("BIGSERIAL PRIMARY KEY")
		} else {
			sb.WriteString("SERIAL PRIMARY KEY")
		}
		return sb.String()
	}

	sb.WriteString(g.getColumnType(col))

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
	case schema.TypeString:
		return g.TypeString(col.Length)
	case schema.TypeText:
		return g.TypeText()
	case schema.TypeInteger:
		return g.TypeInteger()
	case schema.TypeBigInteger:
		return g.TypeBigInteger()
	case schema.TypeSmallInteger:
		return g.TypeSmallInteger()
	case schema.TypeTinyInteger:
		return g.TypeTinyInteger()
	case schema.TypeFloat:
		return g.TypeFloat()
	case schema.TypeDouble:
		return g.TypeDouble()
	case schema.TypeDecimal:
		return g.TypeDecimal(col.Precision, col.Scale)
	case schema.TypeBoolean:
		return g.TypeBoolean()
	case schema.TypeDate:
		return g.TypeDate()
	case schema.TypeDateTime:
		return g.TypeDateTime()
	case schema.TypeTimestamp:
		return g.TypeTimestamp()
	case schema.TypeTime:
		return g.TypeTime()
	case schema.TypeJSON:
		return g.TypeJSON()
	case schema.TypeBinary:
		return g.TypeBinary()
	case schema.TypeUUID:
		return g.TypeUUID()
	default:
		return "VARCHAR(255)"
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

// CompileForeignKey generates ADD FOREIGN KEY SQL
func (g *Grammar) CompileForeignKey(tableName string, fk *schema.ForeignKey) string {
	fkName := fk.Name
	if fkName == "" {
		fkName = g.generateForeignKeyName(tableName, fk.Columns)
	}

	cols := make([]string, len(fk.Columns))
	for i, col := range fk.Columns {
		cols[i] = g.wrap(col)
	}

	sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
		g.wrapTable(tableName),
		g.wrap(fkName),
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

// CompileDropForeignKey generates DROP CONSTRAINT SQL
func (g *Grammar) CompileDropForeignKey(tableName, fkName string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", g.wrapTable(tableName), g.wrap(fkName))
}

// Migrations table operations

func (g *Grammar) CompileCreateMigrationsTable(tableName string) string {
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
  id SERIAL PRIMARY KEY,
  migration VARCHAR(255) NOT NULL,
  batch INTEGER NOT NULL,
  executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`, g.wrapTable(tableName))
}

func (g *Grammar) CompileGetMigrations(tableName string) string {
	return fmt.Sprintf("SELECT id, migration, batch, executed_at FROM %s ORDER BY batch, migration", g.wrapTable(tableName))
}

func (g *Grammar) CompileInsertMigration(tableName string) string {
	return fmt.Sprintf("INSERT INTO %s (migration, batch) VALUES ($1, $2)", g.wrapTable(tableName))
}

func (g *Grammar) CompileDeleteMigration(tableName string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE migration = $1", g.wrapTable(tableName))
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
	fkName := fk.Name
	if fkName == "" {
		fkName = strings.Join(fk.Columns, "_") + "_fk"
	}

	cols := make([]string, len(fk.Columns))
	for i, col := range fk.Columns {
		cols[i] = g.wrap(col)
	}

	sql := fmt.Sprintf("CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
		g.wrap(fkName),
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

func (g *Grammar) generateForeignKeyName(tableName string, columns []string) string {
	return fmt.Sprintf("%s_%s_fk", tableName, strings.Join(columns, "_"))
}

func (g *Grammar) formatDefault(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("'%s'", escapeString(v))
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	case nil:
		return "NULL"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func escapeString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
