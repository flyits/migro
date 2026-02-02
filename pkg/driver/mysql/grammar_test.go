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

	t.Run("modify column", func(t *testing.T) {
		table := schema.NewTable("users")
		table.IsAlter = true
		col := table.String("name", 200)
		col.Change = true

		sqls := g.CompileAlter(table)

		found := false
		for _, sql := range sqls {
			if strings.Contains(sql, "MODIFY COLUMN") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected MODIFY COLUMN statement")
		}
	})

	t.Run("add column after", func(t *testing.T) {
		table := schema.NewTable("users")
		table.IsAlter = true
		col := table.String("phone", 20)
		col.After = "name"

		sqls := g.CompileAlter(table)

		found := false
		for _, sql := range sqls {
			if strings.Contains(sql, "AFTER `name`") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected AFTER clause")
		}
	})

	t.Run("drop foreign key", func(t *testing.T) {
		table := schema.NewTable("posts")
		table.IsAlter = true
		table.DropForeign("posts_user_id_fk")

		sqls := g.CompileAlter(table)

		found := false
		for _, sql := range sqls {
			if strings.Contains(sql, "DROP FOREIGN KEY") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected DROP FOREIGN KEY statement")
		}
	})

	t.Run("drop index", func(t *testing.T) {
		table := schema.NewTable("users")
		table.IsAlter = true
		table.DropIndex("users_email_idx")

		sqls := g.CompileAlter(table)

		found := false
		for _, sql := range sqls {
			if strings.Contains(sql, "DROP INDEX") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected DROP INDEX statement")
		}
	})

	t.Run("add index", func(t *testing.T) {
		table := schema.NewTable("users")
		table.IsAlter = true
		table.Index("email")

		sqls := g.CompileAlter(table)

		found := false
		for _, sql := range sqls {
			if strings.Contains(sql, "CREATE INDEX") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected CREATE INDEX statement")
		}
	})

	t.Run("add foreign key", func(t *testing.T) {
		table := schema.NewTable("posts")
		table.IsAlter = true
		table.Foreign("user_id").References("users", "id")

		sqls := g.CompileAlter(table)

		found := false
		for _, sql := range sqls {
			if strings.Contains(sql, "ADD CONSTRAINT") && strings.Contains(sql, "FOREIGN KEY") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected ADD CONSTRAINT FOREIGN KEY statement")
		}
	})
}

func TestGrammar_CompileDropIndex(t *testing.T) {
	g := NewGrammar()

	sql := g.CompileDropIndex("users", "users_email_idx")
	expected := "DROP INDEX `users_email_idx` ON `users`"

	if sql != expected {
		t.Errorf("expected %s, got %s", expected, sql)
	}
}

func TestGrammar_CompileDropForeignKey(t *testing.T) {
	g := NewGrammar()

	sql := g.CompileDropForeignKey("posts", "posts_user_id_fk")
	expected := "ALTER TABLE `posts` DROP FOREIGN KEY `posts_user_id_fk`"

	if sql != expected {
		t.Errorf("expected %s, got %s", expected, sql)
	}
}

func TestGrammar_MigrationTableOperations(t *testing.T) {
	g := NewGrammar()

	t.Run("CompileGetMigrations", func(t *testing.T) {
		sql := g.CompileGetMigrations("migrations")
		if !strings.Contains(sql, "SELECT") && !strings.Contains(sql, "`migrations`") {
			t.Error("expected SELECT from migrations table")
		}
	})

	t.Run("CompileInsertMigration", func(t *testing.T) {
		sql := g.CompileInsertMigration("migrations")
		if !strings.Contains(sql, "INSERT INTO") {
			t.Error("expected INSERT INTO statement")
		}
	})

	t.Run("CompileDeleteMigration", func(t *testing.T) {
		sql := g.CompileDeleteMigration("migrations")
		if !strings.Contains(sql, "DELETE FROM") {
			t.Error("expected DELETE FROM statement")
		}
	})

	t.Run("CompileGetLastBatch", func(t *testing.T) {
		sql := g.CompileGetLastBatch("migrations")
		if !strings.Contains(sql, "MAX(batch)") {
			t.Error("expected MAX(batch) in query")
		}
	})
}

func TestGrammar_CompileColumn_AllTypes(t *testing.T) {
	g := NewGrammar()

	tests := []struct {
		name     string
		col      *schema.Column
		contains []string
	}{
		{
			name:     "text column",
			col:      &schema.Column{Name: "content", Type: schema.TypeText},
			contains: []string{"`content`", "TEXT"},
		},
		{
			name:     "bigint column",
			col:      &schema.Column{Name: "big_id", Type: schema.TypeBigInteger},
			contains: []string{"`big_id`", "BIGINT"},
		},
		{
			name:     "smallint column",
			col:      &schema.Column{Name: "small_num", Type: schema.TypeSmallInteger},
			contains: []string{"`small_num`", "SMALLINT"},
		},
		{
			name:     "tinyint column",
			col:      &schema.Column{Name: "tiny_num", Type: schema.TypeTinyInteger},
			contains: []string{"`tiny_num`", "TINYINT"},
		},
		{
			name:     "float column",
			col:      &schema.Column{Name: "price", Type: schema.TypeFloat},
			contains: []string{"`price`", "FLOAT"},
		},
		{
			name:     "double column",
			col:      &schema.Column{Name: "amount", Type: schema.TypeDouble},
			contains: []string{"`amount`", "DOUBLE"},
		},
		{
			name:     "decimal column",
			col:      &schema.Column{Name: "total", Type: schema.TypeDecimal, Precision: 10, Scale: 2},
			contains: []string{"`total`", "DECIMAL(10,2)"},
		},
		{
			name:     "boolean column",
			col:      &schema.Column{Name: "active", Type: schema.TypeBoolean},
			contains: []string{"`active`", "TINYINT(1)"},
		},
		{
			name:     "date column",
			col:      &schema.Column{Name: "birth_date", Type: schema.TypeDate},
			contains: []string{"`birth_date`", "DATE"},
		},
		{
			name:     "datetime column",
			col:      &schema.Column{Name: "created_at", Type: schema.TypeDateTime},
			contains: []string{"`created_at`", "DATETIME"},
		},
		{
			name:     "timestamp column",
			col:      &schema.Column{Name: "updated_at", Type: schema.TypeTimestamp},
			contains: []string{"`updated_at`", "TIMESTAMP"},
		},
		{
			name:     "time column",
			col:      &schema.Column{Name: "start_time", Type: schema.TypeTime},
			contains: []string{"`start_time`", "TIME"},
		},
		{
			name:     "json column",
			col:      &schema.Column{Name: "metadata", Type: schema.TypeJSON},
			contains: []string{"`metadata`", "JSON"},
		},
		{
			name:     "binary column",
			col:      &schema.Column{Name: "data", Type: schema.TypeBinary},
			contains: []string{"`data`", "BLOB"},
		},
		{
			name:     "uuid column",
			col:      &schema.Column{Name: "uuid", Type: schema.TypeUUID},
			contains: []string{"`uuid`", "CHAR(36)"},
		},
		{
			name:     "column with comment",
			col:      &schema.Column{Name: "status", Type: schema.TypeString, Length: 20, ColumnComment: "User status"},
			contains: []string{"`status`", "COMMENT 'User status'"},
		},
		{
			name:     "column with default bool true",
			col:      &schema.Column{Name: "active", Type: schema.TypeBoolean, DefaultValue: true},
			contains: []string{"DEFAULT 1"},
		},
		{
			name:     "column with default bool false",
			col:      &schema.Column{Name: "active", Type: schema.TypeBoolean, DefaultValue: false},
			contains: []string{"DEFAULT 0"},
		},
		{
			name:     "column with explicit default nil",
			col:      &schema.Column{Name: "deleted_at", Type: schema.TypeTimestamp, IsNullable: true, DefaultValue: "NULL"},
			contains: []string{"DEFAULT 'NULL'"},
		},
		{
			name:     "column with default number",
			col:      &schema.Column{Name: "count", Type: schema.TypeInteger, DefaultValue: 0},
			contains: []string{"DEFAULT 0"},
		},
		{
			name:     "unsigned bigint",
			col:      &schema.Column{Name: "id", Type: schema.TypeBigInteger, IsUnsigned: true},
			contains: []string{"UNSIGNED"},
		},
		{
			name:     "unsigned float",
			col:      &schema.Column{Name: "price", Type: schema.TypeFloat, IsUnsigned: true},
			contains: []string{"UNSIGNED"},
		},
		{
			name:     "unsigned double",
			col:      &schema.Column{Name: "amount", Type: schema.TypeDouble, IsUnsigned: true},
			contains: []string{"UNSIGNED"},
		},
		{
			name:     "unsigned decimal",
			col:      &schema.Column{Name: "total", Type: schema.TypeDecimal, Precision: 10, Scale: 2, IsUnsigned: true},
			contains: []string{"UNSIGNED"},
		},
		{
			name:     "unsigned smallint",
			col:      &schema.Column{Name: "small", Type: schema.TypeSmallInteger, IsUnsigned: true},
			contains: []string{"UNSIGNED"},
		},
		{
			name:     "unsigned tinyint",
			col:      &schema.Column{Name: "tiny", Type: schema.TypeTinyInteger, IsUnsigned: true},
			contains: []string{"UNSIGNED"},
		},
		{
			name:     "unknown type defaults to varchar",
			col:      &schema.Column{Name: "unknown", Type: schema.ColumnType(999)},
			contains: []string{"VARCHAR(255)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql := g.CompileColumn(tt.col)
			for _, expected := range tt.contains {
				if !strings.Contains(sql, expected) {
					t.Errorf("expected SQL to contain '%s', got: %s", expected, sql)
				}
			}
		})
	}
}

func TestGrammar_CompileCreate_WithIndexes(t *testing.T) {
	g := NewGrammar()

	t.Run("table with unique index", func(t *testing.T) {
		table := schema.NewTable("users")
		table.ID()
		table.String("email", 100)
		table.Unique("email")

		sql := g.CompileCreate(table)

		if !strings.Contains(sql, "UNIQUE KEY") {
			t.Error("expected UNIQUE KEY in CREATE TABLE")
		}
	})

	t.Run("table with regular index", func(t *testing.T) {
		table := schema.NewTable("users")
		table.ID()
		table.String("name", 100)
		table.Index("name")

		sql := g.CompileCreate(table)

		if !strings.Contains(sql, "KEY") {
			t.Error("expected KEY in CREATE TABLE")
		}
	})

	t.Run("table with fulltext index", func(t *testing.T) {
		table := schema.NewTable("posts")
		table.ID()
		table.Text("content")
		table.Index("content").Fulltext()

		sql := g.CompileCreate(table)

		if !strings.Contains(sql, "FULLTEXT KEY") {
			t.Error("expected FULLTEXT KEY in CREATE TABLE")
		}
	})
}

func TestGrammar_CompileCreate_WithForeignKey(t *testing.T) {
	g := NewGrammar()

	t.Run("table with foreign key", func(t *testing.T) {
		table := schema.NewTable("posts")
		table.ID()
		table.BigInteger("user_id").Unsigned()
		table.Foreign("user_id").References("users", "id").OnDeleteCascade()

		sql := g.CompileCreate(table)

		if !strings.Contains(sql, "CONSTRAINT") {
			t.Error("expected CONSTRAINT in CREATE TABLE")
		}
		if !strings.Contains(sql, "FOREIGN KEY") {
			t.Error("expected FOREIGN KEY in CREATE TABLE")
		}
		if !strings.Contains(sql, "REFERENCES `users`") {
			t.Error("expected REFERENCES in CREATE TABLE")
		}
		if !strings.Contains(sql, "ON DELETE CASCADE") {
			t.Error("expected ON DELETE CASCADE in CREATE TABLE")
		}
	})

	t.Run("foreign key with on update", func(t *testing.T) {
		table := schema.NewTable("posts")
		table.ID()
		table.BigInteger("user_id").Unsigned()
		table.Foreign("user_id").References("users", "id").OnUpdateCascade()

		sql := g.CompileCreate(table)

		if !strings.Contains(sql, "ON UPDATE CASCADE") {
			t.Error("expected ON UPDATE CASCADE in CREATE TABLE")
		}
	})
}

func TestGrammar_CompileIndex_WithCustomName(t *testing.T) {
	g := NewGrammar()

	idx := schema.NewIndex("email")
	idx.Name = "custom_email_index"

	sql := g.CompileIndex("users", idx)

	if !strings.Contains(sql, "`custom_email_index`") {
		t.Error("expected custom index name")
	}
}

func TestGrammar_CompileForeignKey_WithCustomName(t *testing.T) {
	g := NewGrammar()

	fk := schema.NewForeignKey("user_id")
	fk.Name = "custom_fk_name"
	fk.References("users", "id")

	sql := g.CompileForeignKey("posts", fk)

	if !strings.Contains(sql, "`custom_fk_name`") {
		t.Error("expected custom foreign key name")
	}
}

func TestGrammar_EscapeString(t *testing.T) {
	g := NewGrammar()

	col := &schema.Column{
		Name:          "comment",
		Type:          schema.TypeString,
		Length:        100,
		ColumnComment: "User's comment with 'quotes'",
	}

	sql := g.CompileColumn(col)

	// 验证单引号被正确转义
	if strings.Contains(sql, "User's") {
		t.Error("expected single quotes to be escaped")
	}
	if !strings.Contains(sql, "User''s") {
		t.Error("expected escaped single quotes")
	}
}
