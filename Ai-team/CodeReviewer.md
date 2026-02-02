# Code Review 报告 - Migro 数据库迁移工具

## 审查状态
- **状态**: 已完成
- **负责人**: Code Reviewer
- **审查时间**: 2026-02-02
- **审查范围**: 全部核心代码

---

## 审查总结

整体代码质量良好，架构设计清晰，符合 Go 语言规范。以下是详细的审查结果。

---

## 【必须修改】

### 1. SQL 注入风险 - 高优先级

**文件**: `pkg/driver/mysql/grammar.go:145-147`

```go
func (g *Grammar) CompileHasTable(name string) string {
    return fmt.Sprintf("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = '%s'", name)
}
```

**问题**: 表名直接拼接到 SQL 字符串中，存在 SQL 注入风险。虽然表名通常来自内部配置，但如果用户可以控制表名，可能导致安全问题。

**同样问题存在于**:
- `pkg/driver/postgres/grammar.go` 的 `CompileHasTable`
- `pkg/driver/sqlite/grammar.go` 的 `CompileHasTable`

**建议**: 使用参数化查询或对表名进行严格验证（只允许字母、数字、下划线）。

---

### 2. 迁移执行缺少事务保护 - 高优先级

**文件**: `internal/migrator/migrator.go:100-116`

```go
for _, migration := range pending {
    executor := NewExecutor(m.driver, m.dryRun)
    if err := migration.Up(executor); err != nil {
        return executed_names, fmt.Errorf("migration %s failed: %w", migration.Name, err)
    }
    // ...
}
```

**问题**: 迁移执行没有使用事务包裹。如果迁移中途失败，已执行的 DDL 语句无法回滚（尤其是 MySQL 的 DDL 会隐式提交）。

**建议**:
- 对于支持事务 DDL 的数据库（PostgreSQL），应使用事务包裹每个迁移
- 对于 MySQL，应在文档中明确说明 DDL 不支持事务回滚
- 考虑添加迁移锁机制防止并发执行

---

### 3. Executor 使用 context.Background() - 中优先级

**文件**: `internal/migrator/migrator.go:292, 307, 318, 329, 334, 345, 355`

```go
func (e *Executor) CreateTable(name string, fn func(*schema.Table)) error {
    // ...
    return e.driver.CreateTable(context.Background(), table)
}
```

**问题**: Executor 的所有方法都使用 `context.Background()`，忽略了调用方传入的 context。这会导致：
- 无法取消长时间运行的迁移
- 无法设置超时
- 无法传递 trace 信息

**建议**: Executor 应该接收并传递 context：
```go
func (e *Executor) CreateTable(ctx context.Context, name string, fn func(*schema.Table)) error
```

---

## 【潜在风险】

### 1. 并发安全 - 驱动注册

**文件**: `pkg/driver/registry.go`

```go
var (
    driversMu sync.RWMutex
    drivers   = make(map[string]Factory)
)
```

**分析**: 驱动注册使用了读写锁保护，这是正确的。但 `init()` 函数中的注册通常在程序启动时单线程执行，实际运行时不会有并发注册的场景。当前实现是安全的。

**状态**: ✅ 无需修改

---

### 2. 数据库连接池配置缺失

**文件**: `pkg/driver/mysql/driver.go:48-58`

```go
db, err := sql.Open("mysql", dsn)
if err != nil {
    return fmt.Errorf("mysql: failed to open connection: %w", err)
}
```

**问题**: 没有配置连接池参数（MaxOpenConns, MaxIdleConns, ConnMaxLifetime）。在高并发场景下可能导致连接耗尽或连接泄漏。

**建议**: 添加连接池配置选项：
```go
db.SetMaxOpenConns(10)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(time.Hour)
```

---

### 3. 错误处理 - 时间解析忽略错误

**文件**: `pkg/driver/sqlite/driver.go:213`

```go
r.ExecutedAt, _ = time.Parse("2006-01-02 15:04:05", executedAt)
```

**问题**: 时间解析错误被忽略。如果数据库中的时间格式不正确，会导致 `ExecutedAt` 为零值。

**建议**: 记录警告日志或返回错误。

---

### 4. 配置文件权限

**文件**: `internal/config/loader.go:65`

```go
if err := os.WriteFile(l.configFile, data, 0644); err != nil {
```

**问题**: 配置文件可能包含数据库密码，使用 0644 权限允许其他用户读取。

**建议**: 使用 0600 权限：
```go
if err := os.WriteFile(l.configFile, data, 0600); err != nil {
```

---

## 【优化建议】

### 1. 代码风格 - 变量命名

**文件**: `internal/migrator/migrator.go:101`

```go
var executed_names []string
```

**建议**: Go 语言推荐使用驼峰命名法，应改为 `executedNames`。

---

### 2. 接口设计 - Grammar 接口过大

**文件**: `pkg/driver/driver.go:39-85`

Grammar 接口包含 30+ 个方法，违反了接口隔离原则。

**建议**: 考虑拆分为更小的接口：
- `TableGrammar` - 表操作
- `ColumnGrammar` - 列类型映射
- `IndexGrammar` - 索引操作
- `MigrationGrammar` - 迁移表操作

---

### 3. 错误信息国际化

当前所有错误信息都是英文硬编码。如果需要支持多语言，建议使用错误码或 i18n 库。

---

### 4. 日志记录缺失

整个项目没有日志记录。建议添加结构化日志（如 slog）用于：
- 记录执行的 SQL 语句
- 记录迁移执行时间
- 记录错误详情

---

### 5. 测试覆盖

当前项目没有单元测试。建议优先为以下模块添加测试：
- `pkg/schema/*` - Schema DSL 构建逻辑
- `pkg/driver/*/grammar.go` - SQL 生成正确性
- `internal/config/loader.go` - 配置解析

---

## 【代码亮点】

1. **清晰的分层架构**: CLI → Migrator → Driver，职责分明
2. **良好的接口抽象**: Driver 和 Grammar 接口设计合理，易于扩展新数据库
3. **链式 API 设计**: Schema DSL 使用流畅的链式调用，用户体验好
4. **错误包装**: 使用 `%w` 正确包装错误，保留错误链
5. **资源管理**: 数据库连接使用 defer 正确关闭
6. **Context 支持**: Driver 接口方法都接受 context 参数

---

## 【是否可以合并 + 原因】

**结论**: ⚠️ **建议修复必须修改项后合并**

**原因**:
1. SQL 注入风险需要修复，这是安全问题
2. 迁移缺少事务保护可能导致数据库状态不一致
3. Executor 忽略 context 会影响生产环境的可控性

**建议优先级**:
1. 【P0】修复 SQL 注入风险
2. 【P0】Executor 传递 context
3. 【P1】添加迁移事务/锁机制
4. 【P2】其他优化建议

---

## 下一步

1. **Engineer** 根据审查意见修复必须修改项
2. **Tester** 开始编写测试用例
3. 修复完成后进行二次审查

---

**任务完成标志**: 代码审查完成，等待 Engineer 修复问题。

**下一步**: 显式调用 `/team` 继续任务。
