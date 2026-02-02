package schema

import (
	"testing"
)

// 测试目标需求: Index 构建器 API
// 来源: Producer.md - 模块 2: 迁移文件 DSL

func TestNewIndex(t *testing.T) {
	t.Run("creates index with columns", func(t *testing.T) {
		idx := NewIndex("email")

		if len(idx.Columns) != 1 {
			t.Fatalf("expected 1 column, got %d", len(idx.Columns))
		}
		if idx.Columns[0] != "email" {
			t.Errorf("expected column 'email', got '%s'", idx.Columns[0])
		}
		if idx.Type != IndexTypeIndex {
			t.Errorf("expected IndexTypeIndex, got %d", idx.Type)
		}
	})

	t.Run("creates composite index", func(t *testing.T) {
		idx := NewIndex("first_name", "last_name")

		if len(idx.Columns) != 2 {
			t.Fatalf("expected 2 columns, got %d", len(idx.Columns))
		}
	})
}

func TestIndex_Named(t *testing.T) {
	t.Run("sets index name", func(t *testing.T) {
		idx := NewIndex("email")
		result := idx.Named("idx_users_email")

		if idx.Name != "idx_users_email" {
			t.Errorf("expected name 'idx_users_email', got '%s'", idx.Name)
		}
		if result != idx {
			t.Error("expected Named() to return the same index for chaining")
		}
	})
}

func TestIndex_Unique(t *testing.T) {
	t.Run("sets index type to unique", func(t *testing.T) {
		idx := NewIndex("email")
		result := idx.Unique()

		if idx.Type != IndexTypeUnique {
			t.Errorf("expected IndexTypeUnique, got %d", idx.Type)
		}
		if result != idx {
			t.Error("expected Unique() to return the same index for chaining")
		}
	})
}

func TestIndex_Primary(t *testing.T) {
	t.Run("sets index type to primary", func(t *testing.T) {
		idx := NewIndex("id")
		result := idx.Primary()

		if idx.Type != IndexTypePrimary {
			t.Errorf("expected IndexTypePrimary, got %d", idx.Type)
		}
		if result != idx {
			t.Error("expected Primary() to return the same index for chaining")
		}
	})
}

func TestIndex_Fulltext(t *testing.T) {
	t.Run("sets index type to fulltext", func(t *testing.T) {
		idx := NewIndex("content")
		result := idx.Fulltext()

		if idx.Type != IndexTypeFulltext {
			t.Errorf("expected IndexTypeFulltext, got %d", idx.Type)
		}
		if result != idx {
			t.Error("expected Fulltext() to return the same index for chaining")
		}
	})
}

func TestIndexType_Values(t *testing.T) {
	tests := []struct {
		name     string
		idxType  IndexType
		expected int
	}{
		{"IndexTypeIndex", IndexTypeIndex, 0},
		{"IndexTypeUnique", IndexTypeUnique, 1},
		{"IndexTypePrimary", IndexTypePrimary, 2},
		{"IndexTypeFulltext", IndexTypeFulltext, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.idxType) != tt.expected {
				t.Errorf("expected %s to be %d, got %d", tt.name, tt.expected, int(tt.idxType))
			}
		})
	}
}

// 测试链式调用
func TestIndex_ChainedMethods(t *testing.T) {
	t.Run("chained methods work correctly", func(t *testing.T) {
		idx := NewIndex("email").Named("idx_email").Unique()

		if idx.Name != "idx_email" {
			t.Errorf("expected name 'idx_email', got '%s'", idx.Name)
		}
		if idx.Type != IndexTypeUnique {
			t.Errorf("expected IndexTypeUnique, got %d", idx.Type)
		}
	})
}
