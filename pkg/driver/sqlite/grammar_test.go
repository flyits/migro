package sqlite

import (
	"strings"
	"testing"

	"github.com/migro/migro/pkg/schema"
)

// 测试目标需求: SQLite Grammar SQL 生成正确性
// 来源: Architect.md - Grammar 接口, CodeReviewer.md - SQL 注入修复验证

func TestGrammar_TypeMappings(t *testing.T) {
	g := NewGrammar()

	// SQLite 使用类型亲和性，大多数类型映射到 TEXT, INTEGER, REAL, BLOB
	tests := []struct {
		name     string
		method   func() string
		expected string
	}{
		{"TypeText", g.TypeText, "TEXT"},
		{"TypeInteger", g.TypeInteger, "INTEGER"},
		{"TypeBigInteger", g.TypeBigInteger, "INTEGER"},
		{"TypeSmallInteger", g.TypeSmallInteger, "INTEGER"},
		{"TypeTinyInteger", g.TypeTinyInteger, "INTEGER"},
		{"TypeFloat", g.TypeFloat, "REAL"},
		{"TypeDouble", g.TypeDouble, "REAL"},
		{"TypeBoolean", g.TypeBoolean, "INTEGER"},
		{"TypeDate", g.TypeDate, "TEXT"},
		{"TypeDateTime", g.TypeDateTime, "TEXT"},
		{"TypeTimestamp", g.TypeTimestamp, "TEXT"},
		{"TypeTime", g.TypeTime, "TEXT"},
		{"TypeJSON", g.TypeJSON, "TEXT"},
		{"TypeBinary", g.TypeBinary, "BLOB"},
		{"TypeUUID", g.TypeUUID, "TEXT"},
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

	// SQLite 的 VARCHAR 映射到 TEXT
	result := g.TypeString(100)
	if result != "TEXT" {
		t.Errorf("expected TEXT, got %s", result)
	}
}

func TestGrammar_TypeDecimal(t *testing.T) {
	g := NewGrammar()

	// SQLite 的 DECIMAL 映射到 REAL
	result := g.TypeDecimal(10, 2)
	if result != "REAL" {
		t.Errorf("expected REAL, got %s", result)
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
		if !strings.Contains(sql, "INTEGER PRIMARY KEY AUTOINCREMENT") {
			t.Error("expected INTEGER PRIMARY KEY AUTOINCREMENT for id")
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

	t.Run("table with foreign key", func(t *testing.T) {
		table := schema.NewTable("posts")
		table.ID()
		table.BigInteger("user_id")
		table.Foreign("user_id").References("users", "id").OnDeleteCascade()

		sql := g.CompileCreate(table)

		if !strings.Contains(sql, "FOREIGN KEY") {
			t.Error("expected FOREIGN KEY")
		}
		if !strings.Contains(sql, "REFERENCES \"users\"") {
			t.Error("expected REFERENCES")
		}
		if !strings.Contains(sql, "ON DELETE CASCADE") {
			t.Error("expected ON DELETE CASCADE")
		}
	})
}

func TestGrammar_CompileColumn(t *testing.T) {
	g := NewGrammar()

	t.Run("auto increment uses INTEGER PRIMARY KEY AUTOINCREMENT", func(t *testing.T) {
		col := &schema.Column{Name: "id", Type: schema.TypeBigInteger, IsAutoIncrement: true}
		sql := g.CompileColumn(col)

		if !strings.Contains(sql, "INTEGER PRIMARY KEY AUTOINCREMENT") {
			t.Error("expected INTEGER PRIMARY KEY AUTOINCREMENT")
		}
	})

	t.Run("boolean default uses 0/1", func(t *testing.T) {
		col := &schema.Column{Name: "is_active", Type: schema.TypeBoolean, DefaultValue: true}
		sql := g.CompileColumn(col)

		if !strings.Contains(sql, "DEFAULT 1") {
			t.Error("expected DEFAULT 1 for boolean true")
		}
	})

	t.Run("boolean false default", func(t *testing.T) {
		col := &schema.Column{Name: "is_active", Type: schema.TypeBoolean, DefaultValue: false}
		sql := g.CompileColumn(col)

		if !strings.Contains(sql, "DEFAULT 0") {
			t.Error("expected DEFAULT 0 for boolean false")
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
		if !strings.Contains(sql, "name='users'") {
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
		longName := strings.Repeat("a", 129) // SQLite limit is 128
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

func TestGrammar_CompileDropIndex(t *testing.T) {
	g := NewGrammar()

	sql := g.CompileDropIndex("users", "idx_email")

	// SQLite 的 DROP INDEX 不需要表名
	if !strings.Contains(sql, "DROP INDEX") {
		t.Error("expected DROP INDEX")
	}
}

func TestGrammar_CompileForeignKey(t *testing.T) {
	g := NewGrammar()

	fk := schema.NewForeignKey("user_id").
		References("users", "id").
		OnDeleteCascade()

	sql := g.CompileForeignKey("posts", fk)

	// SQLite 不支持通过 ALTER TABLE 添加外键
	if sql != "" {
		t.Error("expected empty string for SQLite foreign key (not supported via ALTER)")
	}
}

func TestGrammar_CompileDropForeignKey(t *testing.T) {
	g := NewGrammar()

	sql := g.CompileDropForeignKey("posts", "fk_posts_user_id")

	// SQLite 不支持删除外键
	if sql != "" {
		t.Error("expected empty string for SQLite drop foreign key (not supported)")
	}
}

func TestGrammar_CompileCreateMigrationsTable(t *testing.T) {
	g := NewGrammar()

	sql := g.CompileCreateMigrationsTable("migrations")

	if !strings.Contains(sql, "CREATE TABLE IF NOT EXISTS \"migrations\"") {
		t.Error("expected CREATE TABLE IF NOT EXISTS")
	}
	if !strings.Contains(sql, "INTEGER PRIMARY KEY AUTOINCREMENT") {
		t.Error("expected INTEGER PRIMARY KEY AUTOINCREMENT")
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

	t.Run("rename column (SQLite 3.25.0+)", func(t *testing.T) {
		table := schema.NewTable("users")
		table.IsAlter = true
		table.RenameColumn("old_name", "new_name")

		sqls := g.CompileAlter(table)

		found := false
		for _, sql := range sqls {
			if strings.Contains(sql, "RENAME COLUMN") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected RENAME COLUMN statement")
		}
	})

	// 注意: SQLite 不支持 DROP COLUMN 和 MODIFY COLUMN
	// 这些操作需要重建表
}
