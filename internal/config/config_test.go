package config

import (
	"os"
	"path/filepath"
	"testing"
)

// 测试目标需求: 配置管理模块
// 来源: Producer.md - 模块 4: 配置管理, Architect.md - Config 模块

func TestDefaultConfig(t *testing.T) {
	t.Run("returns valid default config", func(t *testing.T) {
		cfg := DefaultConfig()

		if cfg.Driver != "mysql" {
			t.Errorf("expected default driver to be 'mysql', got '%s'", cfg.Driver)
		}
		if cfg.Connection.Host != "localhost" {
			t.Errorf("expected default host to be 'localhost', got '%s'", cfg.Connection.Host)
		}
		if cfg.Connection.Port != 3306 {
			t.Errorf("expected default port to be 3306, got %d", cfg.Connection.Port)
		}
		if cfg.Connection.Database != "myapp" {
			t.Errorf("expected default database to be 'myapp', got '%s'", cfg.Connection.Database)
		}
		if cfg.Connection.Username != "root" {
			t.Errorf("expected default username to be 'root', got '%s'", cfg.Connection.Username)
		}
		if cfg.Connection.Charset != "utf8mb4" {
			t.Errorf("expected default charset to be 'utf8mb4', got '%s'", cfg.Connection.Charset)
		}
		if cfg.Migrations.Path != "./migrations" {
			t.Errorf("expected default migrations path to be './migrations', got '%s'", cfg.Migrations.Path)
		}
		if cfg.Migrations.Table != "migrations" {
			t.Errorf("expected default migrations table to be 'migrations', got '%s'", cfg.Migrations.Table)
		}
		if cfg.Connection.Options == nil {
			t.Error("expected Options to be initialized")
		}
	})
}

func TestGetDefaultPort(t *testing.T) {
	tests := []struct {
		driver   string
		expected int
	}{
		{"mysql", 3306},
		{"postgres", 5432},
		{"sqlite", 0},
		{"unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			port := GetDefaultPort(tt.driver)
			if port != tt.expected {
				t.Errorf("expected port %d for %s, got %d", tt.expected, tt.driver, port)
			}
		})
	}
}

func TestConfig_ToDriverConfig(t *testing.T) {
	t.Run("converts config correctly", func(t *testing.T) {
		cfg := &Config{
			Driver: "postgres",
			Connection: ConnectionConfig{
				Host:     "db.example.com",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
				Password: "testpass",
				Charset:  "utf8",
				Options:  map[string]string{"sslmode": "require"},
			},
		}

		drvCfg := cfg.ToDriverConfig()

		if drvCfg.Driver != "postgres" {
			t.Errorf("expected driver 'postgres', got '%s'", drvCfg.Driver)
		}
		if drvCfg.Host != "db.example.com" {
			t.Errorf("expected host 'db.example.com', got '%s'", drvCfg.Host)
		}
		if drvCfg.Port != 5432 {
			t.Errorf("expected port 5432, got %d", drvCfg.Port)
		}
		if drvCfg.Database != "testdb" {
			t.Errorf("expected database 'testdb', got '%s'", drvCfg.Database)
		}
		if drvCfg.Username != "testuser" {
			t.Errorf("expected username 'testuser', got '%s'", drvCfg.Username)
		}
		if drvCfg.Password != "testpass" {
			t.Errorf("expected password 'testpass', got '%s'", drvCfg.Password)
		}
		if drvCfg.Options["sslmode"] != "require" {
			t.Error("expected sslmode option to be 'require'")
		}
	})
}

func TestNewLoader(t *testing.T) {
	t.Run("uses default config file when empty", func(t *testing.T) {
		loader := NewLoader("")
		if loader.configFile != DefaultConfigFile {
			t.Errorf("expected default config file '%s', got '%s'", DefaultConfigFile, loader.configFile)
		}
	})

	t.Run("uses provided config file", func(t *testing.T) {
		loader := NewLoader("custom.yaml")
		if loader.configFile != "custom.yaml" {
			t.Errorf("expected config file 'custom.yaml', got '%s'", loader.configFile)
		}
	})
}

func TestLoader_Exists(t *testing.T) {
	t.Run("returns false for non-existent file", func(t *testing.T) {
		loader := NewLoader("non_existent_file.yaml")
		if loader.Exists() {
			t.Error("expected Exists() to return false for non-existent file")
		}
	})

	t.Run("returns true for existing file", func(t *testing.T) {
		// 创建临时文件
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.yaml")
		if err := os.WriteFile(tmpFile, []byte("driver: mysql"), 0644); err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		loader := NewLoader(tmpFile)
		if !loader.Exists() {
			t.Error("expected Exists() to return true for existing file")
		}
	})
}

func TestLoader_Load(t *testing.T) {
	t.Run("loads valid config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "migro.yaml")

		content := `driver: postgres
connection:
  host: localhost
  port: 5432
  database: testdb
  username: testuser
  password: testpass
migrations:
  path: ./db/migrations
  table: schema_migrations
`
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		loader := NewLoader(tmpFile)
		cfg, err := loader.Load()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Driver != "postgres" {
			t.Errorf("expected driver 'postgres', got '%s'", cfg.Driver)
		}
		if cfg.Connection.Host != "localhost" {
			t.Errorf("expected host 'localhost', got '%s'", cfg.Connection.Host)
		}
		if cfg.Connection.Port != 5432 {
			t.Errorf("expected port 5432, got %d", cfg.Connection.Port)
		}
		if cfg.Migrations.Path != "./db/migrations" {
			t.Errorf("expected migrations path './db/migrations', got '%s'", cfg.Migrations.Path)
		}
		if cfg.Migrations.Table != "schema_migrations" {
			t.Errorf("expected migrations table 'schema_migrations', got '%s'", cfg.Migrations.Table)
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		loader := NewLoader("non_existent.yaml")
		_, err := loader.Load()

		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "invalid.yaml")

		content := `driver: mysql
connection:
  host: localhost
  port: not_a_number
`
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		loader := NewLoader(tmpFile)
		_, err := loader.Load()

		if err == nil {
			t.Error("expected error for invalid YAML")
		}
	})

	t.Run("expands environment variables", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "env.yaml")

		content := `driver: mysql
connection:
  host: ${TEST_DB_HOST:localhost}
  port: 3306
  database: ${TEST_DB_NAME:default_db}
  username: ${TEST_DB_USER}
  password: ${TEST_DB_PASS:}
`
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		// 设置环境变量
		os.Setenv("TEST_DB_HOST", "custom-host")
		os.Setenv("TEST_DB_USER", "custom-user")
		defer func() {
			os.Unsetenv("TEST_DB_HOST")
			os.Unsetenv("TEST_DB_USER")
		}()

		loader := NewLoader(tmpFile)
		cfg, err := loader.Load()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Connection.Host != "custom-host" {
			t.Errorf("expected host 'custom-host', got '%s'", cfg.Connection.Host)
		}
		if cfg.Connection.Database != "default_db" {
			t.Errorf("expected database 'default_db' (default value), got '%s'", cfg.Connection.Database)
		}
		if cfg.Connection.Username != "custom-user" {
			t.Errorf("expected username 'custom-user', got '%s'", cfg.Connection.Username)
		}
		if cfg.Connection.Password != "" {
			t.Errorf("expected empty password (default), got '%s'", cfg.Connection.Password)
		}
	})

	t.Run("applies defaults for missing values", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "minimal.yaml")

		content := `connection:
  database: mydb
`
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		loader := NewLoader(tmpFile)
		cfg, err := loader.Load()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// 检查默认值
		if cfg.Driver != "mysql" {
			t.Errorf("expected default driver 'mysql', got '%s'", cfg.Driver)
		}
		if cfg.Connection.Host != "localhost" {
			t.Errorf("expected default host 'localhost', got '%s'", cfg.Connection.Host)
		}
		if cfg.Connection.Port != 3306 {
			t.Errorf("expected default port 3306, got %d", cfg.Connection.Port)
		}
		if cfg.Migrations.Path != "./migrations" {
			t.Errorf("expected default migrations path, got '%s'", cfg.Migrations.Path)
		}
		if cfg.Migrations.Table != "migrations" {
			t.Errorf("expected default migrations table, got '%s'", cfg.Migrations.Table)
		}
	})
}

func TestLoader_Save(t *testing.T) {
	t.Run("saves config to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "output.yaml")

		cfg := &Config{
			Driver: "postgres",
			Connection: ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
				Password: "testpass",
			},
			Migrations: MigrationsConfig{
				Path:  "./migrations",
				Table: "migrations",
			},
		}

		loader := NewLoader(tmpFile)
		err := loader.Save(cfg)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// 验证文件已创建
		if !loader.Exists() {
			t.Error("expected file to exist after save")
		}

		// 重新加载并验证
		loadedCfg, err := loader.Load()
		if err != nil {
			t.Fatalf("failed to reload config: %v", err)
		}
		if loadedCfg.Driver != "postgres" {
			t.Errorf("expected driver 'postgres', got '%s'", loadedCfg.Driver)
		}
	})
}

func TestGenerateConfigTemplate(t *testing.T) {
	tests := []struct {
		driver   string
		contains []string
	}{
		{
			"mysql",
			[]string{"driver: mysql", "port: ${DB_PORT:3306}", "charset: utf8mb4"},
		},
		{
			"postgres",
			[]string{"driver: postgres", "port: ${DB_PORT:5432}"},
		},
		{
			"sqlite",
			[]string{"driver: sqlite", "database: ${DB_PATH:./database.db}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			template := GenerateConfigTemplate(tt.driver)

			for _, expected := range tt.contains {
				if !containsString(template, expected) {
					t.Errorf("expected template to contain '%s'", expected)
				}
			}
		})
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"3306", 3306},
		{"5432", 5432},
		{"0", 0},
		{"invalid", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParsePort(tt.input)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
