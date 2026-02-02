package postgres

import (
	"strings"
	"testing"

	"github.com/migro/migro/pkg/schema"
)

// 测试目标需求: PostgreSQL Grammar SQL 生成正确性
// 来源: Architect.md - Grammar 接口, CodeReviewer.md - SQL 注入修复验证

func TestGrammar_TypeMappings(t *testing.T) {
	g := NewGrammar()

	tests := []struct {
		name     string
		method   func() string
		expected string
	}{
		{"TypeText", g.TypeText, "TEXT"},
		{"TypeInteger", g.TypeInteger, "INTEGER"},
		{"TypeBigInteger", g.TypeBigInteger, "BIGINT"},
		{"TypeSmallInteger", g.TypeSmallInteger, "SMALLINT"},
		{"TypeTinyInteger", g.TypeTinyInteger, "SMALLINT"}, // PostgreSQL 没有 TINYINT
		{"TypeFloat", g.TypeFloat, "REAL"},
		{"TypeDouble", g.TypeDouble, "DOUBLE PRECISION"},
		{"TypeBoolean", g.TypeBoolean, "BOOLEAN"},
		{"TypeDate", g.TypeDate, "DATE"},
		{"TypeDateTime", g.TypeDateTime, "TIMESTAMP"},
		{"TypeTimestamp", g.TypeTimestamp, "TIMESTAMP"},
		{"TypeTime", g.TypeTime, "TIME"},
		{"TypeJSON", g.TypeJSON, "JSONB"},
		{"TypeBinary", g.TypeBinary, "BYTEA"},
		{"TypeUUID", g.TypeUUID, "UUID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGrammar_TypeString(t *testing.T) {
	g := NewGrammar()

	tests := []struct {
		name     string
		length   int
		expected string
	}{
		{"with length", 100, "VARCHAR(100)"},
		{"zero length defaults to 255", 0, "VARCHAR(255)"},
		{"negative length defaults to 255", -1, "VARCHAR(255)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.TypeString(tt.length)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGrammar_CompileCreate(t *testing.T) {
	g := NewGrammar()

	t.Run("simple table", func(t *testing.T) {
		table := schema.NewTable("users")
		table.ID()
		table.String("name", 100)

		sql := g.CompileCreate(table)

		if !strings.Contains(sql, "CREATE TABLE \"users\"") {
			t.Error("expected CREATE TABLE statement")
		}
		if !strings.Contains(sql, "\"id\"") {
			t.Error("expected id column")
		}
		if !strings.Contains(sql, "BIGSERIAL PRIMARY KEY") {
			t.Error("expected BIGSERIAL PRIMARY KEY for auto-increment id")
		}
	})

	t.Run("table with IF NOT EXISTS", func(t *testing.T) {
		table := schema.NewTable("users")
		table.IfNotExists = true
		table.ID()

		sql := g.CompileCreate(table)

		if !strings.Contains(sql, "IF NOT EXISTS") {
			t.Error("expected IF NOT EXISTS")
		}
	})
}

func TestGrammar_CompileColumn(t *testing.T) {
	g := NewGrammar()

	t.Run("auto increment uses SERIAL", func(t *testing.T) {
		col := &schema.Column{Name: "id", Type: schema.TypeInteger, IsAutoIncrement: true}
		sql := g.CompileColumn(col)

		if !strings.Contains(sql, "SERIAL PRIMARY KEY") {
			t.Error("expected SERIAL PRIMARY KEY")
		}
	})

	t.Run("auto increment bigint uses BIGSERIAL", func(t *testing.T) {
		col := &schema.Column{Name: "id", Type: schema.TypeBigInteger, IsAutoIncrement: true}
		sql := g.CompileColumn(col)

		if !strings.Contains(sql, "BIGSERIAL PRIMARY KEY") {
			t.Error("expected BIGSERIAL PRIMARY KEY")
		}
	})

	t.Run("boolean default uses TRUE/FALSE", func(t *testing.T) {
		col := &schema.Column{Name: "is_active", Type: schema.TypeBoolean, DefaultValue: true}
		sql := g.CompileColumn(col)

		if !strings.Contains(sql, "DEFAULT TRUE") {
			t.Error("expected DEFAULT TRUE")
		}
	})
}

func TestGrammar_CompileDrop(t *testing.T) {
	g := NewGrammar()

	sql := g.CompileDrop("users")
	expected := "DROP TABLE \"users\""

	if sql != expected {
		t.Errorf("expected %s, got %s", expected, sql)
	}
}

func TestGrammar_CompileDropIfExists(t *testing.T) {
	g := NewGrammar()

	sql := g.CompileDropIfExists("users")
	expected := "DROP TABLE IF EXISTS \"users\""

	if sql != expected {
		t.Errorf("expected %s, got %s", expected, sql)
	}
}

func TestGrammar_CompileRename(t *testing.T) {
	g := NewGrammar()

	sql := g.CompileRename("old_users", "users")
	expected := "ALTER TABLE \"old_users\" RENAME TO \"users\""

	if sql != expected {
		t.Errorf("expected %s, got %s", expected, sql)
	}
}

// 测试 SQL 注入防护 (来自 CodeReviewer.md - P0 修复)
func TestGrammar_CompileHasTable_SQLInjectionPrevention(t *testing.T) {
	g := NewGrammar()

	t.Run("valid table name", func(t *testing.T) {
		sql, err := g.CompileHasTable("users")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !strings.Contains(sql, "table_name = 'users'") {
			t.Error("expected valid SQL")
		}
	})

	t.Run("rejects SQL injection attempt", func(t *testing.T) {
		_, err := g.CompileHasTable("users'; DROP TABLE users; --")

		if err == nil {
			t.Error("expected error for SQL injection attempt")
		}
	})

	t.Run("rejects empty table name", func(t *testing.T) {
		_, err := g.CompileHasTable("")

		if err == nil {
			t.Error("expected error for empty table name")
		}
	})

	t.Run("rejects table name exceeding max length", func(t *testing.T) {
		longName := strings.Repeat("a", 64) // PostgreSQL limit is 63
		_, err := g.CompileHasTable(longName)

		if err == nil {
			t.Error("expected error for table name exceeding max length")
		}
	})

	t.Run("accepts valid identifier with underscore", func(t *testing.T) {
		sql, err := g.CompileHasTable("user_profiles")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if sql == "" {
			t.Error("expected valid SQL")
		}
	})
}

func TestGrammar_CompileIndex(t *testing.T) {
	g := NewGrammar()

	t.Run("regular index", func(t *testing.T) {
		idx := schema.NewIndex("email")
		sql := g.CompileIndex("users", idx)

		if !strings.Contains(sql, "CREATE INDEX") {
			t.Error("expected CREATE INDEX")
		}
		if !strings.Contains(sql, "ON \"users\"") {
			t.Error("expected ON \"users\"")
		}
	})

	t.Run("unique index", func(t *testing.T) {
		idx := schema.NewIndex("email").Unique()
		sql := g.CompileIndex("users", idx)

		if !strings.Contains(sql, "CREATE UNIQUE INDEX") {
			t.Error("expected CREATE UNIQUE INDEX")
		}
	})
}

func TestGrammar_CompileForeignKey(t *testing.T) {
	g := NewGrammar()

	fk := schema.NewForeignKey("user_id").
		References("users", "id").
		OnDeleteCascade()

	sql := g.CompileForeignKey("posts", fk)

	if !strings.Contains(sql, "ALTER TABLE \"posts\"") {
		t.Error("expected ALTER TABLE")
	}
	if !strings.Contains(sql, "FOREIGN KEY") {
		t.Error("expected FOREIGN KEY")
	}
	if !strings.Contains(sql, "REFERENCES \"users\"") {
		t.Error("expected REFERENCES")
	}
	if !strings.Contains(sql, "ON DELETE CASCADE") {
		t.Error("expected ON DELETE CASCADE")
	}
}

func TestGrammar_CompileDropForeignKey(t *testing.T) {
	g := NewGrammar()

	sql := g.CompileDropForeignKey("posts", "fk_posts_user_id")

	// PostgreSQL 使用 DROP CONSTRAINT
	if !strings.Contains(sql, "DROP CONSTRAINT") {
		t.Error("expected DROP CONSTRAINT")
	}
}

func TestGrammar_CompileCreateMigrationsTable(t *testing.T) {
	g := NewGrammar()

	sql := g.CompileCreateMigrationsTable("migrations")

	if !strings.Contains(sql, "CREATE TABLE IF NOT EXISTS \"migrations\"") {
		t.Error("expected CREATE TABLE IF NOT EXISTS")
	}
	if !strings.Contains(sql, "SERIAL PRIMARY KEY") {
		t.Error("expected SERIAL PRIMARY KEY")
	}
}

func TestGrammar_CompileInsertMigration(t *testing.T) {
	g := NewGrammar()

	sql := g.CompileInsertMigration("migrations")

	// PostgreSQL 使用 $1, $2 占位符
	if !strings.Contains(sql, "$1") || !strings.Contains(sql, "$2") {
		t.Error("expected PostgreSQL-style placeholders ($1, $2)")
	}
}

func TestGrammar_CompileAlter(t *testing.T) {
	g := NewGrammar()

	t.Run("add column", func(t *testing.T) {
		table := schema.NewTable("users")
		table.IsAlter = true
		table.String("phone", 20)

		sqls := g.CompileAlter(table)

		if len(sqls) == 0 {
			t.Fatal("expected at least one SQL statement")
		}
		if !strings.Contains(sqls[0], "ADD COLUMN") {
			t.Error("expected ADD COLUMN")
		}
	})

	t.Run("modify column generates multiple statements", func(t *testing.T) {
		table := schema.NewTable("users")
		table.IsAlter = true
		col := table.String("name", 200)
		col.Change = true
		col.IsNullable = true
		col.DefaultValue = "unknown"

		sqls := g.CompileAlter(table)

		// PostgreSQL 需要多个语句来修改列
		hasTypeChange := false
		hasNullChange := false
		hasDefaultChange := false

		for _, sql := range sqls {
			if strings.Contains(sql, "ALTER COLUMN") && strings.Contains(sql, "TYPE") {
				hasTypeChange = true
			}
			if strings.Contains(sql, "DROP NOT NULL") {
				hasNullChange = true
			}
			if strings.Contains(sql, "SET DEFAULT") {
				hasDefaultChange = true
			}
		}

		if !hasTypeChange {
			t.Error("expected TYPE change statement")
		}
		if !hasNullChange {
			t.Error("expected NULL change statement")
		}
		if !hasDefaultChange {
			t.Error("expected DEFAULT change statement")
		}
	})
}
