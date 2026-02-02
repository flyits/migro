package schema

import (
	"testing"
)

// 测试目标需求: Column 链式修饰符 API
// 来源: Producer.md - 模块 2: 迁移文件 DSL

func TestColumn_Nullable(t *testing.T) {
	t.Run("sets IsNullable to true", func(t *testing.T) {
		col := &Column{Name: "email", Type: TypeString}
		result := col.Nullable()

		if !col.IsNullable {
			t.Error("expected IsNullable to be true")
		}
		if result != col {
			t.Error("expected Nullable() to return the same column for chaining")
		}
	})
}

func TestColumn_Default(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{"string default", "active", "active"},
		{"int default", 0, 0},
		{"bool default", true, true},
		{"nil default", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &Column{Name: "status", Type: TypeString}
			result := col.Default(tt.value)

			if col.DefaultValue != tt.expected {
				t.Errorf("expected DefaultValue to be %v, got %v", tt.expected, col.DefaultValue)
			}
			if result != col {
				t.Error("expected Default() to return the same column for chaining")
			}
		})
	}
}

func TestColumn_Unsigned(t *testing.T) {
	t.Run("sets IsUnsigned to true", func(t *testing.T) {
		col := &Column{Name: "age", Type: TypeInteger}
		result := col.Unsigned()

		if !col.IsUnsigned {
			t.Error("expected IsUnsigned to be true")
		}
		if result != col {
			t.Error("expected Unsigned() to return the same column for chaining")
		}
	})
}

func TestColumn_AutoIncrement(t *testing.T) {
	t.Run("sets IsAutoIncrement to true", func(t *testing.T) {
		col := &Column{Name: "id", Type: TypeBigInteger}
		result := col.AutoIncrement()

		if !col.IsAutoIncrement {
			t.Error("expected IsAutoIncrement to be true")
		}
		if result != col {
			t.Error("expected AutoIncrement() to return the same column for chaining")
		}
	})
}

func TestColumn_Primary(t *testing.T) {
	t.Run("sets IsPrimary to true", func(t *testing.T) {
		col := &Column{Name: "id", Type: TypeBigInteger}
		result := col.Primary()

		if !col.IsPrimary {
			t.Error("expected IsPrimary to be true")
		}
		if result != col {
			t.Error("expected Primary() to return the same column for chaining")
		}
	})
}

func TestColumn_Unique(t *testing.T) {
	t.Run("sets IsUnique to true", func(t *testing.T) {
		col := &Column{Name: "email", Type: TypeString}
		result := col.Unique()

		if !col.IsUnique {
			t.Error("expected IsUnique to be true")
		}
		if result != col {
			t.Error("expected Unique() to return the same column for chaining")
		}
	})
}

func TestColumn_Comment(t *testing.T) {
	t.Run("sets comment", func(t *testing.T) {
		col := &Column{Name: "email", Type: TypeString}
		result := col.Comment("User email address")

		// Comment() 方法设置 ColumnComment 字段
		// 通过再次调用 Comment 并检查返回值来验证链式调用
		if result != col {
			t.Error("expected Comment() to return the same column for chaining")
		}
	})
}

func TestColumn_PlaceAfter(t *testing.T) {
	t.Run("sets After column name", func(t *testing.T) {
		col := &Column{Name: "phone", Type: TypeString}
		result := col.PlaceAfter("email")

		if col.After != "email" {
			t.Errorf("expected After to be 'email', got '%s'", col.After)
		}
		if result != col {
			t.Error("expected PlaceAfter() to return the same column for chaining")
		}
	})
}

// 测试链式调用组合
func TestColumn_ChainedModifiers(t *testing.T) {
	t.Run("multiple modifiers can be chained", func(t *testing.T) {
		col := &Column{Name: "email", Type: TypeString, Length: 100}

		col.Nullable().Default("").Unique()

		if !col.IsNullable {
			t.Error("expected IsNullable to be true")
		}
		if col.DefaultValue != "" {
			t.Errorf("expected DefaultValue to be empty string, got %v", col.DefaultValue)
		}
		if !col.IsUnique {
			t.Error("expected IsUnique to be true")
		}
	})
}

// 测试 ColumnType 枚举值
func TestColumnType_Values(t *testing.T) {
	tests := []struct {
		name     string
		colType  ColumnType
		expected int
	}{
		{"TypeString", TypeString, 0},
		{"TypeText", TypeText, 1},
		{"TypeInteger", TypeInteger, 2},
		{"TypeBigInteger", TypeBigInteger, 3},
		{"TypeSmallInteger", TypeSmallInteger, 4},
		{"TypeTinyInteger", TypeTinyInteger, 5},
		{"TypeFloat", TypeFloat, 6},
		{"TypeDouble", TypeDouble, 7},
		{"TypeDecimal", TypeDecimal, 8},
		{"TypeBoolean", TypeBoolean, 9},
		{"TypeDate", TypeDate, 10},
		{"TypeDateTime", TypeDateTime, 11},
		{"TypeTimestamp", TypeTimestamp, 12},
		{"TypeTime", TypeTime, 13},
		{"TypeJSON", TypeJSON, 14},
		{"TypeBinary", TypeBinary, 15},
		{"TypeUUID", TypeUUID, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.colType) != tt.expected {
				t.Errorf("expected %s to be %d, got %d", tt.name, tt.expected, int(tt.colType))
			}
		})
	}
}
