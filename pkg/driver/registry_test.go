package driver

import (
	"context"
	"database/sql"
	"sync"
	"testing"

	"github.com/migro/migro/pkg/schema"
)

// 测试目标需求: 驱动注册表功能正确性
// 覆盖: Register, Get, Drivers 函数

// mockDriver 用于测试的模拟驱动
type mockDriver struct {
	name string
}

func (m *mockDriver) Connect(config *Config) error                                   { return nil }
func (m *mockDriver) Close() error                                                   { return nil }
func (m *mockDriver) DB() *sql.DB                                                    { return nil }
func (m *mockDriver) Begin(ctx context.Context) (Transaction, error)                 { return nil, nil }
func (m *mockDriver) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (m *mockDriver) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (m *mockDriver) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}
func (m *mockDriver) CreateTable(ctx context.Context, table *schema.Table) error  { return nil }
func (m *mockDriver) AlterTable(ctx context.Context, table *schema.Table) error   { return nil }
func (m *mockDriver) DropTable(ctx context.Context, name string) error            { return nil }
func (m *mockDriver) DropTableIfExists(ctx context.Context, name string) error    { return nil }
func (m *mockDriver) HasTable(ctx context.Context, name string) (bool, error)     { return false, nil }
func (m *mockDriver) RenameTable(ctx context.Context, from, to string) error      { return nil }
func (m *mockDriver) CreateMigrationsTable(ctx context.Context, tableName string) error {
	return nil
}
func (m *mockDriver) GetExecutedMigrations(ctx context.Context, tableName string) ([]MigrationRecord, error) {
	return nil, nil
}
func (m *mockDriver) RecordMigration(ctx context.Context, tableName, migration string, batch int) error {
	return nil
}
func (m *mockDriver) DeleteMigration(ctx context.Context, tableName, migration string) error {
	return nil
}
func (m *mockDriver) GetLastBatch(ctx context.Context, tableName string) (int, error) { return 0, nil }
func (m *mockDriver) Grammar() Grammar                                                { return nil }
func (m *mockDriver) Name() string                                                    { return m.name }

// resetDrivers 重置驱动注册表（用于测试隔离）
func resetDrivers() {
	driversMu.Lock()
	defer driversMu.Unlock()
	drivers = make(map[string]Factory)
}

func TestRegister(t *testing.T) {
	t.Run("register valid driver", func(t *testing.T) {
		resetDrivers()

		Register("test_driver", func() Driver {
			return &mockDriver{name: "test_driver"}
		})

		drv, err := Get("test_driver")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if drv == nil {
			t.Error("expected driver, got nil")
		}
		if drv.Name() != "test_driver" {
			t.Errorf("expected name 'test_driver', got '%s'", drv.Name())
		}
	})

	t.Run("register nil factory panics", func(t *testing.T) {
		resetDrivers()

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for nil factory")
			}
		}()

		Register("nil_driver", nil)
	})

	t.Run("register duplicate driver panics", func(t *testing.T) {
		resetDrivers()

		Register("dup_driver", func() Driver {
			return &mockDriver{name: "dup_driver"}
		})

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for duplicate registration")
			}
		}()

		Register("dup_driver", func() Driver {
			return &mockDriver{name: "dup_driver"}
		})
	})
}

func TestGet(t *testing.T) {
	t.Run("get registered driver", func(t *testing.T) {
		resetDrivers()

		Register("get_test", func() Driver {
			return &mockDriver{name: "get_test"}
		})

		drv, err := Get("get_test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if drv == nil {
			t.Error("expected driver, got nil")
		}
	})

	t.Run("get unregistered driver returns error", func(t *testing.T) {
		resetDrivers()

		_, err := Get("nonexistent")
		if err == nil {
			t.Error("expected error for unregistered driver")
		}
	})

	t.Run("get returns new instance each time", func(t *testing.T) {
		resetDrivers()

		callCount := 0
		Register("instance_test", func() Driver {
			callCount++
			return &mockDriver{name: "instance_test"}
		})

		_, _ = Get("instance_test")
		_, _ = Get("instance_test")

		if callCount != 2 {
			t.Errorf("expected factory called 2 times, got %d", callCount)
		}
	})
}

func TestDrivers(t *testing.T) {
	t.Run("returns empty list when no drivers registered", func(t *testing.T) {
		resetDrivers()

		list := Drivers()
		if len(list) != 0 {
			t.Errorf("expected empty list, got %d drivers", len(list))
		}
	})

	t.Run("returns all registered drivers", func(t *testing.T) {
		resetDrivers()

		Register("driver_a", func() Driver { return &mockDriver{name: "a"} })
		Register("driver_b", func() Driver { return &mockDriver{name: "b"} })
		Register("driver_c", func() Driver { return &mockDriver{name: "c"} })

		list := Drivers()
		if len(list) != 3 {
			t.Errorf("expected 3 drivers, got %d", len(list))
		}

		// 验证所有驱动都在列表中
		driverSet := make(map[string]bool)
		for _, name := range list {
			driverSet[name] = true
		}

		for _, expected := range []string{"driver_a", "driver_b", "driver_c"} {
			if !driverSet[expected] {
				t.Errorf("expected driver '%s' in list", expected)
			}
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	t.Run("concurrent register and get", func(t *testing.T) {
		resetDrivers()

		var wg sync.WaitGroup
		errChan := make(chan error, 100)

		// 先注册一些驱动
		for i := 0; i < 10; i++ {
			name := "concurrent_" + string(rune('a'+i))
			Register(name, func() Driver {
				return &mockDriver{name: name}
			})
		}

		// 并发获取驱动
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				name := "concurrent_" + string(rune('a'+(idx%10)))
				_, err := Get(name)
				if err != nil {
					errChan <- err
				}
			}(i)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			t.Errorf("concurrent access error: %v", err)
		}
	})

	t.Run("concurrent Drivers call", func(t *testing.T) {
		resetDrivers()

		Register("concurrent_list", func() Driver {
			return &mockDriver{name: "concurrent_list"}
		})

		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				list := Drivers()
				if len(list) == 0 {
					t.Error("expected at least one driver")
				}
			}()
		}

		wg.Wait()
	})
}
