package schema

import (
	"testing"
)

// 测试目标需求: Table 构建器 API
// 来源: Producer.md - 模块 2: 迁移文件 DSL, Architect.md - Table 构建器

func TestNewTable(t *testing.T) {
	t.Run("creates table with correct name", func(t *testing.T) {
		table := NewTable("users")

		if table.Name != "users" {
			t.Errorf("expected table name to be 'users', got '%s'", table.Name)
		}
		if table.Columns == nil {
			t.Error("expected Columns to be initialized")
		}
		if table.Indexes == nil {
			t.Error("expected Indexes to be initialized")
		}
		if table.ForeignKeys == nil {
			t.Error("expected ForeignKeys to be initialized")
		}
		if table.RenameColumns == nil {
			t.Error("expected RenameColumns to be initialized")
		}
	})
}

func TestTable_ID(t *testing.T) {
	t.Run("creates auto-increment primary key", func(t *testing.T) {
		table := NewTable("users")
		col := table.ID()

		if col.Name != "id" {
			t.Errorf("expected column name to be 'id', got '%s'", col.Name)
		}
		if col.Type != TypeBigInteger {
			t.Errorf("expected column type to be TypeBigInteger, got %d", col.Type)
		}
		if !col.IsAutoIncrement {
			t.Error("expected IsAutoIncrement to be true")
		}
		if !col.IsUnsigned {
			t.Error("expected IsUnsigned to be true")
		}
		if !col.IsPrimary {
			t.Error("expected IsPrimary to be true")
		}
		if len(table.Columns) != 1 {
			t.Errorf("expected 1 column, got %d", len(table.Columns))
		}
	})
}

func TestTable_String(t *testing.T) {
	t.Run("creates VARCHAR column with length", func(t *testing.T) {
		table := NewTable("users")
		col := table.String("name", 100)

		if col.Name != "name" {
			t.Errorf("expected column name to be 'name', got '%s'", col.Name)
		}
		if col.Type != TypeString {
			t.Errorf("expected column type to be TypeString, got %d", col.Type)
		}
		if col.Length != 100 {
			t.Errorf("expected length to be 100, got %d", col.Length)
		}
	})
}

func TestTable_Text(t *testing.T) {
	t.Run("creates TEXT column", func(t *testing.T) {
		table := NewTable("posts")
		col := table.Text("content")

		if col.Name != "content" {
			t.Errorf("expected column name to be 'content', got '%s'", col.Name)
		}
		if col.Type != TypeText {
			t.Errorf("expected column type to be TypeText, got %d", col.Type)
		}
	})
}

func TestTable_IntegerTypes(t *testing.T) {
	tests := []struct {
		name       string
		method     func(*Table, string) *Column
		colName    string
		expected   ColumnType
	}{
		{"Integer", func(t *Table, n string) *Column { return t.Integer(n) }, "count", TypeInteger},
		{"BigInteger", func(t *Table, n string) *Column { return t.BigInteger(n) }, "big_count", TypeBigInteger},
		{"SmallInteger", func(t *Table, n string) *Column { return t.SmallInteger(n) }, "small_count", TypeSmallInteger},
		{"TinyInteger", func(t *Table, n string) *Column { return t.TinyInteger(n) }, "tiny_count", TypeTinyInteger},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := NewTable("test")
			col := tt.method(table, tt.colName)

			if col.Name != tt.colName {
				t.Errorf("expected column name to be '%s', got '%s'", tt.colName, col.Name)
			}
			if col.Type != tt.expected {
				t.Errorf("expected column type to be %d, got %d", tt.expected, col.Type)
			}
		})
	}
}

func TestTable_FloatTypes(t *testing.T) {
	t.Run("Float", func(t *testing.T) {
		table := NewTable("products")
		col := table.Float("price")

		if col.Type != TypeFloat {
			t.Errorf("expected column type to be TypeFloat, got %d", col.Type)
		}
	})

	t.Run("Double", func(t *testing.T) {
		table := NewTable("products")
		col := table.Double("precise_price")

		if col.Type != TypeDouble {
			t.Errorf("expected column type to be TypeDouble, got %d", col.Type)
		}
	})

	t.Run("Decimal", func(t *testing.T) {
		table := NewTable("products")
		col := table.Decimal("amount", 10, 2)

		if col.Type != TypeDecimal {
			t.Errorf("expected column type to be TypeDecimal, got %d", col.Type)
		}
		if col.Precision != 10 {
			t.Errorf("expected precision to be 10, got %d", col.Precision)
		}
		if col.Scale != 2 {
			t.Errorf("expected scale to be 2, got %d", col.Scale)
		}
	})
}

func TestTable_Boolean(t *testing.T) {
	t.Run("creates BOOLEAN column", func(t *testing.T) {
		table := NewTable("users")
		col := table.Boolean("is_active")

		if col.Type != TypeBoolean {
			t.Errorf("expected column type to be TypeBoolean, got %d", col.Type)
		}
	})
}

func TestTable_DateTimeTypes(t *testing.T) {
	tests := []struct {
		name     string
		method   func(*Table, string) *Column
		expected ColumnType
	}{
		{"Date", func(t *Table, n string) *Column { return t.Date(n) }, TypeDate},
		{"DateTime", func(t *Table, n string) *Column { return t.DateTime(n) }, TypeDateTime},
		{"Timestamp", func(t *Table, n string) *Column { return t.Timestamp(n) }, TypeTimestamp},
		{"Time", func(t *Table, n string) *Column { return t.Time(n) }, TypeTime},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := NewTable("events")
			col := tt.method(table, "test_col")

			if col.Type != tt.expected {
				t.Errorf("expected column type to be %d, got %d", tt.expected, col.Type)
			}
		})
	}
}

func TestTable_JSON(t *testing.T) {
	t.Run("creates JSON column", func(t *testing.T) {
		table := NewTable("settings")
		col := table.JSON("data")

		if col.Type != TypeJSON {
			t.Errorf("expected column type to be TypeJSON, got %d", col.Type)
		}
	})
}

func TestTable_Binary(t *testing.T) {
	t.Run("creates BINARY column", func(t *testing.T) {
		table := NewTable("files")
		col := table.Binary("content")

		if col.Type != TypeBinary {
			t.Errorf("expected column type to be TypeBinary, got %d", col.Type)
		}
	})
}

func TestTable_UUID(t *testing.T) {
	t.Run("creates UUID column", func(t *testing.T) {
		table := NewTable("users")
		col := table.UUID("uuid")

		if col.Type != TypeUUID {
			t.Errorf("expected column type to be TypeUUID, got %d", col.Type)
		}
	})
}

func TestTable_Timestamps(t *testing.T) {
	t.Run("creates created_at and updated_at columns", func(t *testing.T) {
		table := NewTable("users")
		table.Timestamps()

		if len(table.Columns) != 2 {
			t.Fatalf("expected 2 columns, got %d", len(table.Columns))
		}

		createdAt := table.Columns[0]
		if createdAt.Name != "created_at" {
			t.Errorf("expected first column to be 'created_at', got '%s'", createdAt.Name)
		}
		if createdAt.Type != TypeTimestamp {
			t.Errorf("expected created_at type to be TypeTimestamp, got %d", createdAt.Type)
		}
		if !createdAt.IsNullable {
			t.Error("expected created_at to be nullable")
		}

		updatedAt := table.Columns[1]
		if updatedAt.Name != "updated_at" {
			t.Errorf("expected second column to be 'updated_at', got '%s'", updatedAt.Name)
		}
		if !updatedAt.IsNullable {
			t.Error("expected updated_at to be nullable")
		}
	})
}

func TestTable_SoftDeletes(t *testing.T) {
	t.Run("creates deleted_at column", func(t *testing.T) {
		table := NewTable("users")
		table.SoftDeletes()

		if len(table.Columns) != 1 {
			t.Fatalf("expected 1 column, got %d", len(table.Columns))
		}

		deletedAt := table.Columns[0]
		if deletedAt.Name != "deleted_at" {
			t.Errorf("expected column to be 'deleted_at', got '%s'", deletedAt.Name)
		}
		if deletedAt.Type != TypeTimestamp {
			t.Errorf("expected type to be TypeTimestamp, got %d", deletedAt.Type)
		}
		if !deletedAt.IsNullable {
			t.Error("expected deleted_at to be nullable")
		}
	})
}

func TestTable_Index(t *testing.T) {
	t.Run("creates index on single column", func(t *testing.T) {
		table := NewTable("users")
		idx := table.Index("email")

		if len(table.Indexes) != 1 {
			t.Fatalf("expected 1 index, got %d", len(table.Indexes))
		}
		if len(idx.Columns) != 1 || idx.Columns[0] != "email" {
			t.Errorf("expected index on 'email', got %v", idx.Columns)
		}
		if idx.Type != IndexTypeIndex {
			t.Errorf("expected IndexTypeIndex, got %d", idx.Type)
		}
	})

	t.Run("creates composite index", func(t *testing.T) {
		table := NewTable("users")
		idx := table.Index("first_name", "last_name")

		if len(idx.Columns) != 2 {
			t.Errorf("expected 2 columns in index, got %d", len(idx.Columns))
		}
	})
}

func TestTable_Unique(t *testing.T) {
	t.Run("creates unique index", func(t *testing.T) {
		table := NewTable("users")
		idx := table.Unique("email")

		if idx.Type != IndexTypeUnique {
			t.Errorf("expected IndexTypeUnique, got %d", idx.Type)
		}
	})
}

func TestTable_Primary(t *testing.T) {
	t.Run("creates primary key index", func(t *testing.T) {
		table := NewTable("user_roles")
		idx := table.Primary("user_id", "role_id")

		if idx.Type != IndexTypePrimary {
			t.Errorf("expected IndexTypePrimary, got %d", idx.Type)
		}
		if len(table.PrimaryKey) != 2 {
			t.Errorf("expected 2 primary key columns, got %d", len(table.PrimaryKey))
		}
	})
}

func TestTable_Foreign(t *testing.T) {
	t.Run("creates foreign key", func(t *testing.T) {
		table := NewTable("posts")
		fk := table.Foreign("user_id")

		if len(table.ForeignKeys) != 1 {
			t.Fatalf("expected 1 foreign key, got %d", len(table.ForeignKeys))
		}
		if len(fk.Columns) != 1 || fk.Columns[0] != "user_id" {
			t.Errorf("expected foreign key on 'user_id', got %v", fk.Columns)
		}
	})
}

func TestTable_DropColumn(t *testing.T) {
	t.Run("marks column for deletion", func(t *testing.T) {
		table := NewTable("users")
		table.DropColumn("old_column")

		if len(table.DropColumns) != 1 {
			t.Fatalf("expected 1 column to drop, got %d", len(table.DropColumns))
		}
		if table.DropColumns[0] != "old_column" {
			t.Errorf("expected 'old_column', got '%s'", table.DropColumns[0])
		}
	})
}

func TestTable_DropIndex(t *testing.T) {
	t.Run("marks index for deletion", func(t *testing.T) {
		table := NewTable("users")
		table.DropIndex("users_email_idx")

		if len(table.DropIndexes) != 1 {
			t.Fatalf("expected 1 index to drop, got %d", len(table.DropIndexes))
		}
		if table.DropIndexes[0] != "users_email_idx" {
			t.Errorf("expected 'users_email_idx', got '%s'", table.DropIndexes[0])
		}
	})
}

func TestTable_DropForeign(t *testing.T) {
	t.Run("marks foreign key for deletion", func(t *testing.T) {
		table := NewTable("posts")
		table.DropForeign("posts_user_id_fk")

		if len(table.DropForeignKeys) != 1 {
			t.Fatalf("expected 1 foreign key to drop, got %d", len(table.DropForeignKeys))
		}
		if table.DropForeignKeys[0] != "posts_user_id_fk" {
			t.Errorf("expected 'posts_user_id_fk', got '%s'", table.DropForeignKeys[0])
		}
	})
}

func TestTable_RenameColumn(t *testing.T) {
	t.Run("records column rename", func(t *testing.T) {
		table := NewTable("users")
		table.RenameColumn("old_name", "new_name")

		if len(table.RenameColumns) != 1 {
			t.Fatalf("expected 1 rename, got %d", len(table.RenameColumns))
		}
		if table.RenameColumns["old_name"] != "new_name" {
			t.Errorf("expected rename from 'old_name' to 'new_name'")
		}
	})
}

func TestTable_SetEngine(t *testing.T) {
	t.Run("sets MySQL engine", func(t *testing.T) {
		table := NewTable("users")
		result := table.SetEngine("InnoDB")

		if table.Engine != "InnoDB" {
			t.Errorf("expected engine to be 'InnoDB', got '%s'", table.Engine)
		}
		if result != table {
			t.Error("expected SetEngine() to return the same table for chaining")
		}
	})
}

func TestTable_SetCharset(t *testing.T) {
	t.Run("sets MySQL charset", func(t *testing.T) {
		table := NewTable("users")
		result := table.SetCharset("utf8mb4")

		if table.Charset != "utf8mb4" {
			t.Errorf("expected charset to be 'utf8mb4', got '%s'", table.Charset)
		}
		if result != table {
			t.Error("expected SetCharset() to return the same table for chaining")
		}
	})
}

func TestTable_SetCollation(t *testing.T) {
	t.Run("sets MySQL collation", func(t *testing.T) {
		table := NewTable("users")
		result := table.SetCollation("utf8mb4_unicode_ci")

		if table.Collation != "utf8mb4_unicode_ci" {
			t.Errorf("expected collation to be 'utf8mb4_unicode_ci', got '%s'", table.Collation)
		}
		if result != table {
			t.Error("expected SetCollation() to return the same table for chaining")
		}
	})
}

// 测试完整的表定义场景 (来自 Producer.md 用户故事 3)
func TestTable_CompleteUserTableDefinition(t *testing.T) {
	t.Run("creates complete users table", func(t *testing.T) {
		table := NewTable("users")
		table.ID()
		table.String("name", 100).Nullable()
		table.String("email", 100).Unique()
		table.String("password", 255)
		table.Timestamps()

		// 验证列数量
		if len(table.Columns) != 6 { // id, name, email, password, created_at, updated_at
			t.Errorf("expected 6 columns, got %d", len(table.Columns))
		}

		// 验证 id 列
		idCol := table.Columns[0]
		if idCol.Name != "id" || !idCol.IsPrimary || !idCol.IsAutoIncrement {
			t.Error("id column not configured correctly")
		}

		// 验证 name 列
		nameCol := table.Columns[1]
		if nameCol.Name != "name" || !nameCol.IsNullable || nameCol.Length != 100 {
			t.Error("name column not configured correctly")
		}

		// 验证 email 列
		emailCol := table.Columns[2]
		if emailCol.Name != "email" || !emailCol.IsUnique || emailCol.Length != 100 {
			t.Error("email column not configured correctly")
		}
	})
}
