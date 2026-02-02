package schema

import "testing"

func TestTable_ChangeColumn(t *testing.T) {
	t.Run("creates column with Change flag set", func(t *testing.T) {
		table := NewTable("users")
		col := table.ChangeColumn("email", TypeString)

		if len(table.Columns) != 1 {
			t.Errorf("expected 1 column, got %d", len(table.Columns))
		}
		if col.Name != "email" {
			t.Errorf("expected column name 'email', got '%s'", col.Name)
		}
		if col.Type != TypeString {
			t.Errorf("expected TypeString, got %v", col.Type)
		}
		if !col.Change {
			t.Error("expected Change flag to be true")
		}
	})
}

func TestTable_ChangeString(t *testing.T) {
	t.Run("changes column to VARCHAR with length", func(t *testing.T) {
		table := NewTable("users")
		col := table.ChangeString("email", 320)

		if col.Type != TypeString {
			t.Errorf("expected TypeString, got %v", col.Type)
		}
		if col.Length != 320 {
			t.Errorf("expected length 320, got %d", col.Length)
		}
		if !col.Change {
			t.Error("expected Change flag to be true")
		}
	})
}

func TestTable_ChangeInteger(t *testing.T) {
	t.Run("changes column to INT", func(t *testing.T) {
		table := NewTable("posts")
		col := table.ChangeInteger("view_count")

		if col.Type != TypeInteger {
			t.Errorf("expected TypeInteger, got %v", col.Type)
		}
		if !col.Change {
			t.Error("expected Change flag to be true")
		}
	})
}

func TestTable_ChangeBigInteger(t *testing.T) {
	t.Run("changes column to BIGINT", func(t *testing.T) {
		table := NewTable("posts")
		col := table.ChangeBigInteger("user_id")

		if col.Type != TypeBigInteger {
			t.Errorf("expected TypeBigInteger, got %v", col.Type)
		}
		if !col.Change {
			t.Error("expected Change flag to be true")
		}
	})
}

func TestTable_ChangeText(t *testing.T) {
	t.Run("changes column to TEXT", func(t *testing.T) {
		table := NewTable("articles")
		col := table.ChangeText("content")

		if col.Type != TypeText {
			t.Errorf("expected TypeText, got %v", col.Type)
		}
		if !col.Change {
			t.Error("expected Change flag to be true")
		}
	})
}

func TestTable_ChangeDecimal(t *testing.T) {
	t.Run("changes column to DECIMAL with precision and scale", func(t *testing.T) {
		table := NewTable("products")
		col := table.ChangeDecimal("price", 10, 2)

		if col.Type != TypeDecimal {
			t.Errorf("expected TypeDecimal, got %v", col.Type)
		}
		if col.Precision != 10 {
			t.Errorf("expected precision 10, got %d", col.Precision)
		}
		if col.Scale != 2 {
			t.Errorf("expected scale 2, got %d", col.Scale)
		}
		if !col.Change {
			t.Error("expected Change flag to be true")
		}
	})
}

func TestTable_ChangeBoolean(t *testing.T) {
	t.Run("changes column to BOOLEAN", func(t *testing.T) {
		table := NewTable("users")
		col := table.ChangeBoolean("is_active")

		if col.Type != TypeBoolean {
			t.Errorf("expected TypeBoolean, got %v", col.Type)
		}
		if !col.Change {
			t.Error("expected Change flag to be true")
		}
	})
}

func TestTable_ChangeTimestamp(t *testing.T) {
	t.Run("changes column to TIMESTAMP", func(t *testing.T) {
		table := NewTable("users")
		col := table.ChangeTimestamp("last_login")

		if col.Type != TypeTimestamp {
			t.Errorf("expected TypeTimestamp, got %v", col.Type)
		}
		if !col.Change {
			t.Error("expected Change flag to be true")
		}
	})
}

func TestTable_ChangeJSON(t *testing.T) {
	t.Run("changes column to JSON", func(t *testing.T) {
		table := NewTable("settings")
		col := table.ChangeJSON("preferences")

		if col.Type != TypeJSON {
			t.Errorf("expected TypeJSON, got %v", col.Type)
		}
		if !col.Change {
			t.Error("expected Change flag to be true")
		}
	})
}

func TestTable_ChangeColumnWithModifiers(t *testing.T) {
	t.Run("changes column with chained modifiers", func(t *testing.T) {
		table := NewTable("users")
		col := table.ChangeBigInteger("user_id").Unsigned().Nullable().Comment("User ID")

		if col.Type != TypeBigInteger {
			t.Errorf("expected TypeBigInteger, got %v", col.Type)
		}
		if !col.Change {
			t.Error("expected Change flag to be true")
		}
		if !col.IsUnsigned {
			t.Error("expected IsUnsigned to be true")
		}
		if !col.IsNullable {
			t.Error("expected IsNullable to be true")
		}
		if col.ColumnComment != "User ID" {
			t.Errorf("expected comment 'User ID', got '%s'", col.ColumnComment)
		}
	})
}

func TestTable_MultipleColumnChanges(t *testing.T) {
	t.Run("changes multiple columns in one table", func(t *testing.T) {
		table := NewTable("users")
		table.ChangeString("email", 320)
		table.ChangeBigInteger("user_id").Unsigned()
		table.ChangeText("bio")

		if len(table.Columns) != 3 {
			t.Errorf("expected 3 columns, got %d", len(table.Columns))
		}

		for _, col := range table.Columns {
			if !col.Change {
				t.Errorf("expected Change flag to be true for column '%s'", col.Name)
			}
		}
	})
}
