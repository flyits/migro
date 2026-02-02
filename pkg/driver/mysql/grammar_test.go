package mysql

import (
	"strings"
	"testing"

	"github.com/migro/migro/pkg/schema"
)

// 测试目标需求: MySQL Grammar SQL 生成正确性
// 来源: Architect.md - Grammar 接口, CodeReviewer.md - SQL 注入修复验证

func TestGrammar_TypeMappings(t *testing.T) {
	g := NewGrammar()

	tests := []struct {
		name     string
		method   func() string
		expected string
	}{
		{"TypeText", g.TypeText, "TEXT"},
		{"TypeInteger", g.TypeInteger, "INT"},
		{"TypeBigInteger", g.TypeBigInteger, "BIGINT"},
		{"TypeSmallInteger", g.TypeSmallInteger, "SMALLINT"},
		{"TypeTinyInteger", g.TypeTinyInteger, "TINYINT"},
		{"TypeFloat", g.TypeFloat, "FLOAT"},
		{"TypeDouble", g.TypeDouble, "DOUBLE"},
		{"TypeBoolean", g.TypeBoolean, "TINYINT(1)"},
		{"TypeDate", g.TypeDate, "DATE"},
		{"TypeDateTime", g.TypeDateTime, "DATETIME"},
		{"TypeTimestamp", g.TypeTimestamp, "TIMESTAMP"},
		{"TypeTime", g.TypeTime, "TIME"},
		{"TypeJSON", g.TypeJSON, "JSON"},
		{"TypeBinary", g.TypeBinary, "BLOB"},
		{"TypeUUID", g.TypeUUID, "CHAR(36)"},
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

func TestGrammar_TypeDecimal(t *testing.T) {
	g := NewGrammar()

	result := g.TypeDecimal(10, 2)
	expected := "DECIMAL(10,2)"

	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestGrammar_CompileCreate(t *testing.T) {
	g := NewGrammar()

	t.Run("simple table", func(t *testing.T) {
		table := schema.NewTable("users")
		table.ID()
		table.String("name", 100)

		sql := g.CompileCreate(table)

		if !strings.Contains(sql, "CREATE TABLE `users`") {
			t.Error("expected CREATE TABLE statement")
		}
		if !strings.Contains(sql, "`id`") {
			t.Error("expected id column")
		}
		if !strings.Contains(sql, "`name`") {
			t.Error("expected name column")
		}
		if !strings.Contains(sql, "AUTO_INCREMENT") {
			t.Error("expected AUTO_INCREMENT for id")
		}
		if !strings.Contains(sql, "PRIMARY KEY") {
			t.Error("expected PRIMARY KEY")
		}
	})

	t.Run("table with engine and charset", func(t *testing.T) {
		table := schema.NewTable("users")
		table.ID()
		table.SetEngine("InnoDB")
		table.SetCharset("utf8mb4")
		table.SetCollation("utf8mb4_unicode_ci")

		sql := g.CompileCreate(table)

		if !strings.Contains(sql, "ENGINE=InnoDB") {
			t.Error("expected ENGINE=InnoDB")
		}
		if !strings.Contains(sql, "DEFAULT CHARSET=utf8mb4") {
			t.Error("expected DEFAULT CHARSET=utf8mb4")
		}
		if !strings.Contains(sql, "COLLATE=utf8mb4_unicode_ci") {
			t.Error("expected COLLATE=utf8mb4_unicode_ci")
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

	t.Run("nullable column", func(t *testing.T) {
		col := &schema.Column{Name: "email", Type: schema.TypeString, Length: 100, IsNullable: true}
		sql := g.CompileColumn(col)

		if !strings.Contains(sql, "NULL") {
			t.Error("expected NULL for nullable column")
		}
	})

	t.Run("not null column", func(t *testing.T) {
		col := &schema.Column{Name: "email", Type: schema.TypeString, Length: 100, IsNullable: false}
		sql := g.CompileColumn(col)

		if !strings.Contains(sql, "NOT NULL") {
			t.Error("expected NOT NULL")
		}
	})

	t.Run("column with default", func(t *testing.T) {
		col := &schema.Column{Name: "status", Type: schema.TypeString, Length: 20, DefaultValue: "active"}
		sql := g.CompileColumn(col)

		if !strings.Contains(sql, "DEFAULT 'active'") {
			t.Error("expected DEFAULT 'active'")
		}
	})

	t.Run("unsigned integer", func(t *testing.T) {
		col := &schema.Column{Name: "count", Type: schema.TypeInteger, IsUnsigned: true}
		sql := g.CompileColumn(col)

		if !strings.Contains(sql, "UNSIGNED") {
			t.Error("expected UNSIGNED")
		}
	})

	t.Run("auto increment", func(t *testing.T) {
		col := &schema.Column{Name: "id", Type: schema.TypeBigInteger, IsAutoIncrement: true}
		sql := g.CompileColumn(col)

		if !strings.Contains(sql, "AUTO_INCREMENT") {
			t.Error("expected AUTO_INCREMENT")
		}
	})
}

func TestGrammar_CompileDrop(t *testing.T) {
	g := NewGrammar()

	t.Run("drop table", func(t *testing.T) {
		sql := g.CompileDrop("users")
		expected := "DROP TABLE `users`"

		if sql != expected {
			t.Errorf("expected %s, got %s", expected, sql)
		}
	})
}

func TestGrammar_CompileDropIfExists(t *testing.T) {
	g := NewGrammar()

	t.Run("drop table if exists", func(t *testing.T) {
		sql := g.CompileDropIfExists("users")
		expected := "DROP TABLE IF EXISTS `users`"

		if sql != expected {
			t.Errorf("expected %s, got %s", expected, sql)
		}
	})
}

func TestGrammar_CompileRename(t *testing.T) {
	g := NewGrammar()

	t.Run("rename table", func(t *testing.T) {
		sql := g.CompileRename("old_users", "users")
		expected := "RENAME TABLE `old_users` TO `users`"

		if sql != expected {
			t.Errorf("expected %s, got %s", expected, sql)
		}
	})
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

	t.Run("rejects table name with special characters", func(t *testing.T) {
		_, err := g.CompileHasTable("users-table")

		if err == nil {
			t.Error("expected error for table name with hyphen")
		}
	})

	t.Run("rejects table name starting with number", func(t *testing.T) {
		_, err := g.CompileHasTable("123users")

		if err == nil {
			t.Error("expected error for table name starting with number")
		}
	})

	t.Run("rejects table name exceeding max length", func(t *testing.T) {
		longName := strings.Repeat("a", 65) // MySQL limit is 64
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

	t.Run("accepts identifier starting with underscore", func(t *testing.T) {
		sql, err := g.CompileHasTable("_temp_table")

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
		if !strings.Contains(sql, "ON `users`") {
			t.Error("expected ON `users`")
		}
	})

	t.Run("unique index", func(t *testing.T) {
		idx := schema.NewIndex("email").Unique()
		sql := g.CompileIndex("users", idx)

		if !strings.Contains(sql, "CREATE UNIQUE INDEX") {
			t.Error("expected CREATE UNIQUE INDEX")
		}
	})

	t.Run("fulltext index", func(t *testing.T) {
		idx := schema.NewIndex("content").Fulltext()
		sql := g.CompileIndex("posts", idx)

		if !strings.Contains(sql, "CREATE FULLTEXT INDEX") {
			t.Error("expected CREATE FULLTEXT INDEX")
		}
	})
}

func TestGrammar_CompileForeignKey(t *testing.T) {
	g := NewGrammar()

	t.Run("foreign key with cascade", func(t *testing.T) {
		fk := schema.NewForeignKey("user_id").
			References("users", "id").
			OnDeleteCascade()

		sql := g.CompileForeignKey("posts", fk)

		if !strings.Contains(sql, "ALTER TABLE `posts`") {
			t.Error("expected ALTER TABLE")
		}
		if !strings.Contains(sql, "FOREIGN KEY") {
			t.Error("expected FOREIGN KEY")
		}
		if !strings.Contains(sql, "REFERENCES `users`") {
			t.Error("expected REFERENCES")
		}
		if !strings.Contains(sql, "ON DELETE CASCADE") {
			t.Error("expected ON DELETE CASCADE")
		}
	})
}

func TestGrammar_CompileCreateMigrationsTable(t *testing.T) {
	g := NewGrammar()

	sql := g.CompileCreateMigrationsTable("migrations")

	if !strings.Contains(sql, "CREATE TABLE IF NOT EXISTS `migrations`") {
		t.Error("expected CREATE TABLE IF NOT EXISTS")
	}
	if !strings.Contains(sql, "migration VARCHAR(255)") {
		t.Error("expected migration column")
	}
	if !strings.Contains(sql, "batch INT") {
		t.Error("expected batch column")
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

	t.Run("drop column", func(t *testing.T) {
		table := schema.NewTable("users")
		table.IsAlter = true
		table.DropColumn("old_column")

		sqls := g.CompileAlter(table)

		found := false
		for _, sql := range sqls {
			if strings.Contains(sql, "DROP COLUMN") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected DROP COLUMN statement")
		}
	})

	t.Run("rename column", func(t *testing.T) {
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
}
