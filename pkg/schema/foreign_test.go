package schema

import (
	"testing"
)

// 测试目标需求: ForeignKey 构建器 API
// 来源: Producer.md - 模块 2: 迁移文件 DSL

func TestNewForeignKey(t *testing.T) {
	t.Run("creates foreign key with column", func(t *testing.T) {
		fk := NewForeignKey("user_id")

		if len(fk.Columns) != 1 {
			t.Fatalf("expected 1 column, got %d", len(fk.Columns))
		}
		if fk.Columns[0] != "user_id" {
			t.Errorf("expected column 'user_id', got '%s'", fk.Columns[0])
		}
		// 默认动作应该是 RESTRICT
		if fk.OnDelete != ActionRestrict {
			t.Errorf("expected default OnDelete to be ActionRestrict, got %s", fk.OnDelete)
		}
		if fk.OnUpdate != ActionRestrict {
			t.Errorf("expected default OnUpdate to be ActionRestrict, got %s", fk.OnUpdate)
		}
	})
}

func TestForeignKey_Named(t *testing.T) {
	t.Run("sets foreign key name", func(t *testing.T) {
		fk := NewForeignKey("user_id")
		result := fk.Named("fk_posts_user_id")

		if fk.Name != "fk_posts_user_id" {
			t.Errorf("expected name 'fk_posts_user_id', got '%s'", fk.Name)
		}
		if result != fk {
			t.Error("expected Named() to return the same foreign key for chaining")
		}
	})
}

func TestForeignKey_References(t *testing.T) {
	t.Run("sets reference table and column", func(t *testing.T) {
		fk := NewForeignKey("user_id")
		result := fk.References("users", "id")

		if fk.ReferenceTable != "users" {
			t.Errorf("expected reference table 'users', got '%s'", fk.ReferenceTable)
		}
		if fk.ReferenceColumn != "id" {
			t.Errorf("expected reference column 'id', got '%s'", fk.ReferenceColumn)
		}
		if result != fk {
			t.Error("expected References() to return the same foreign key for chaining")
		}
	})
}

func TestForeignKey_OnDeleteActions(t *testing.T) {
	tests := []struct {
		name     string
		method   func(*ForeignKey) *ForeignKey
		expected ForeignKeyAction
	}{
		{"OnDeleteCascade", func(fk *ForeignKey) *ForeignKey { return fk.OnDeleteCascade() }, ActionCascade},
		{"OnDeleteSetNull", func(fk *ForeignKey) *ForeignKey { return fk.OnDeleteSetNull() }, ActionSetNull},
		{"OnDeleteRestrict", func(fk *ForeignKey) *ForeignKey { return fk.OnDeleteRestrict() }, ActionRestrict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fk := NewForeignKey("user_id")
			result := tt.method(fk)

			if fk.OnDelete != tt.expected {
				t.Errorf("expected OnDelete to be %s, got %s", tt.expected, fk.OnDelete)
			}
			if result != fk {
				t.Error("expected method to return the same foreign key for chaining")
			}
		})
	}
}

func TestForeignKey_OnUpdateActions(t *testing.T) {
	tests := []struct {
		name     string
		method   func(*ForeignKey) *ForeignKey
		expected ForeignKeyAction
	}{
		{"OnUpdateCascade", func(fk *ForeignKey) *ForeignKey { return fk.OnUpdateCascade() }, ActionCascade},
		{"OnUpdateSetNull", func(fk *ForeignKey) *ForeignKey { return fk.OnUpdateSetNull() }, ActionSetNull},
		{"OnUpdateRestrict", func(fk *ForeignKey) *ForeignKey { return fk.OnUpdateRestrict() }, ActionRestrict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fk := NewForeignKey("user_id")
			result := tt.method(fk)

			if fk.OnUpdate != tt.expected {
				t.Errorf("expected OnUpdate to be %s, got %s", tt.expected, fk.OnUpdate)
			}
			if result != fk {
				t.Error("expected method to return the same foreign key for chaining")
			}
		})
	}
}

func TestForeignKeyAction_Values(t *testing.T) {
	tests := []struct {
		name     string
		action   ForeignKeyAction
		expected string
	}{
		{"ActionCascade", ActionCascade, "CASCADE"},
		{"ActionRestrict", ActionRestrict, "RESTRICT"},
		{"ActionSetNull", ActionSetNull, "SET NULL"},
		{"ActionNoAction", ActionNoAction, "NO ACTION"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.action) != tt.expected {
				t.Errorf("expected %s to be '%s', got '%s'", tt.name, tt.expected, string(tt.action))
			}
		})
	}
}

// 测试完整的外键定义场景 (来自 Producer.md)
func TestForeignKey_CompleteDefinition(t *testing.T) {
	t.Run("creates complete foreign key with cascade", func(t *testing.T) {
		fk := NewForeignKey("user_id").
			Named("fk_posts_user").
			References("users", "id").
			OnDeleteCascade().
			OnUpdateCascade()

		if fk.Name != "fk_posts_user" {
			t.Errorf("expected name 'fk_posts_user', got '%s'", fk.Name)
		}
		if fk.ReferenceTable != "users" {
			t.Errorf("expected reference table 'users', got '%s'", fk.ReferenceTable)
		}
		if fk.ReferenceColumn != "id" {
			t.Errorf("expected reference column 'id', got '%s'", fk.ReferenceColumn)
		}
		if fk.OnDelete != ActionCascade {
			t.Errorf("expected OnDelete to be CASCADE, got %s", fk.OnDelete)
		}
		if fk.OnUpdate != ActionCascade {
			t.Errorf("expected OnUpdate to be CASCADE, got %s", fk.OnUpdate)
		}
	})
}
