package cli

import (
	"go/parser"
	"go/token"
	"testing"
)

// 测试目标：验证 toSnakeCase 函数能正确处理各种输入
func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 正常情况
		{"普通小写", "create", "create"},
		{"驼峰命名", "CreateUser", "create_user"},
		{"多个大写", "CreateUserTable", "create_user_table"},

		// 版本号（已修复的问题）
		{"版本号v1.4.6", "v1.4.6", "v1_4_6"},
		{"版本号v2.0.0", "v2.0.0", "v2_0_0"},
		{"带v前缀版本", "V1.2.3", "v1_2_3"},
		{"带连字符版本", "v2.0.0-beta", "v2_0_0_beta"},

		// 特殊字符
		{"包含连字符", "create-user", "create_user"},
		{"包含下划线", "create_user", "create_user"},
		{"包含数字", "user123", "user123"},
		{"纯数字", "123456", "123456"},
		{"包含空格", "create user", "create_user"},

		// 边界情况
		{"空字符串", "", ""},
		{"单个字符", "a", "a"},
		{"单个大写", "A", "a"},

		// 中文和Unicode字符
		{"中文名称", "用户表", "用户表"},
		{"中英混合", "create用户", "create用户"},
		{"日文", "ユーザー", "ユーザー"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// 测试目标：验证 toCamelCase 函数能正确处理各种输入
func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 正常情况
		{"snake_case", "create_user", "CreateUser"},
		{"已是驼峰", "CreateUser", "CreateUser"},
		{"单词", "user", "User"},

		// 版本号 - 数字部分合并是正确行为，生成的标识符有效
		{"版本号v1.4.6", "v1.4.6", "V146"},
		{"版本号v2.0.0", "v2.0.0", "V200"},
		{"带连字符版本", "v2.0.0-beta", "V200Beta"},

		// 边界情况
		{"空字符串", "", ""},
		{"单个字符", "a", "A"},
		{"带连续下划线", "a__b", "AB"},

		// 中文
		{"中文", "用户表", "用户表"},
		{"中英混合snake", "create_用户_table", "Create用户Table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// 测试目标：验证生成的迁移模板是否为合法的 Go 语法
// 这是核心测试 - 确保生成的代码不会有语法错误
func TestGenerateMigrationTemplate_ValidGoSyntax(t *testing.T) {
	tests := []struct {
		name      string
		migName   string
		tableName string
		timestamp string
		wantValid bool
	}{
		// 正常情况 - 应该生成合法Go代码
		{"普通名称", "create_users", "users", "20260204120000", true},
		{"驼峰名称", "CreateUsers", "users", "20260204120000", true},

		// 版本号 - 修复后应该生成合法Go代码
		{"版本号v1.4.6", "v1.4.6", "table_name", "20260204120000", true},
		{"版本号v2.0.0-beta", "v2.0.0-beta", "table_name", "20260204120000", true},

		// 纯数字 - Go标识符不能以数字开头
		{"纯数字", "123", "table_name", "20260204120000", false},
		{"数字开头", "123_migration", "table_name", "20260204120000", false},

		// 中文 - Go支持Unicode标识符
		{"中文名称", "用户迁移", "users", "20260204120000", true},
		{"中英混合", "create用户表", "users", "20260204120000", true},

		// 特殊字符 - 修复后应该生成合法Go代码
		{"包含连字符", "create-user", "users", "20260204120000", true},
		{"包含空格", "create user", "users", "20260204120000", true},

		// 这些仍然会有问题（生成空标识符或无效字符）
		{"包含斜杠", "create/user", "users", "20260204120000", true},
		{"包含星号", "create*user", "users", "20260204120000", true},
		{"包含括号", "create(user)", "users", "20260204120000", true},

		// 空字符串
		{"空名称", "", "users", "20260204120000", false},
	}

	fset := token.NewFileSet()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := generateMigrationTemplate(tt.migName, tt.tableName, tt.timestamp)

			// 尝试解析生成的Go代码
			_, err := parser.ParseFile(fset, "test.go", content, parser.AllErrors)

			isValid := err == nil

			if isValid != tt.wantValid {
				if tt.wantValid {
					t.Errorf("generateMigrationTemplate(%q) 生成了无效的Go代码: %v\n生成的代码:\n%s",
						tt.migName, err, content)
				} else {
					t.Errorf("generateMigrationTemplate(%q) 预期生成无效Go代码，但实际生成了有效代码\n生成的代码:\n%s",
						tt.migName, content)
				}
			}
		})
	}
}

// 测试目标：验证 extractTableName 函数
func TestExtractTableName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"create_xxx_table格式", "create_users_table", "users"},
		{"add_xxx_to_yyy格式", "add_email_to_users", "users"},
		{"remove_xxx_from_yyy格式", "remove_email_from_users", "users"},
		{"普通名称", "migration", "table_name"},
		{"版本号", "v1.4.6", "table_name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTableName(tt.input)
			if result != tt.expected {
				t.Errorf("extractTableName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// 测试目标：验证生成的结构体名称是否为合法的 Go 标识符
func TestGeneratedStructNameIsValidIdentifier(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldBeOK  bool
		description string
	}{
		{"正常名称", "create_users", true, "普通snake_case名称"},
		{"版本号", "v1.4.6", true, "版本号应该转换为V1_4_6"},
		{"中文", "用户表", true, "Go支持Unicode标识符"},
		{"数字开头", "123test", false, "Go标识符不能以数字开头"},
		{"连字符", "create-user", true, "连字符应该被转换为下划线"},
		{"空格", "create user", true, "空格应该被转换为下划线"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			structName := toCamelCase(tt.input)

			// 检查是否为合法Go标识符
			isValid := isValidGoIdentifier(structName)

			if isValid != tt.shouldBeOK {
				t.Errorf("toCamelCase(%q) = %q, isValidIdentifier = %v, want %v (%s)",
					tt.input, structName, isValid, tt.shouldBeOK, tt.description)
			}
		})
	}
}

// isValidGoIdentifier 检查字符串是否为合法的Go标识符
func isValidGoIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	for i, r := range s {
		if i == 0 {
			// 首字符必须是字母或下划线
			if !isLetter(r) && r != '_' {
				return false
			}
		} else {
			// 后续字符可以是字母、数字或下划线
			if !isLetter(r) && !isDigit(r) && r != '_' {
				return false
			}
		}
	}

	// 检查是否为Go关键字
	keywords := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}

	return !keywords[s]
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r >= 0x80 // Unicode letters
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
