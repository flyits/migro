# Producer 需求分析文档

## 任务概述

为 Migro 数据库迁移工具新增两个 API，支持用户直接传入已有的数据库连接实例，而非通过配置创建新连接。

---

## 需求场景分析

### 场景 1：传入 `*sql.DB` 实例

**用户故事**：
> 作为一个 Go 开发者，我希望能够将已有的 `*sql.DB` 数据库连接传入 Migro，以便在现有项目中复用数据库连接池，避免创建重复连接。

**典型使用场景**：
- 项目已有数据库连接管理，不希望 Migro 单独创建连接
- 需要在同一事务中执行迁移和业务逻辑
- 使用连接池管理，希望 Migro 共享连接池
- 测试场景中使用 mock 或内存数据库

### 场景 2：传入 GORM `*gorm.DB` 实例

**用户故事**：
> 作为一个使用 GORM 的 Go 开发者，我希望能够直接传入 GORM 实例给 Migro，以便在 GORM 项目中无缝集成数据库迁移功能。

**典型使用场景**：
- 项目使用 GORM 作为 ORM，希望迁移工具与 GORM 共享连接
- 利用 GORM 的连接配置和连接池
- 在 GORM 事务中执行迁移

---

## 现有架构分析

### 当前连接方式

```go
// 1. 通过 Config 创建连接
drv, _ := driver.Get("mysql")
drv.Connect(&driver.Config{
    Host:     "localhost",
    Port:     3306,
    Database: "mydb",
    Username: "root",
    Password: "password",
})

// 2. 创建 Migrator
migrator := migrator.NewMigrator(drv, "./migrations", "migrations")
```

### 架构特点

1. **Driver 接口**：定义了 `Connect(config *Config) error` 方法
2. **各驱动实现**：MySQL、PostgreSQL、SQLite 都实现了 Driver 接口
3. **驱动内部**：都持有 `*sql.DB` 实例（`d.db` 字段）
4. **Migrator**：接收 `driver.Driver` 接口

### 关键发现

- 所有驱动内部都使用 `*sql.DB`
- 驱动已有 `DB() *sql.DB` 方法返回底层连接
- 需要新增方法允许从外部设置 `*sql.DB`

---

## API 设计方案

### API 1：`ConnectWithDB` - 传入 `*sql.DB`

**函数签名**：
```go
// 在 Driver 接口中新增
ConnectWithDB(db *sql.DB) error

// 各驱动实现
func (d *Driver) ConnectWithDB(db *sql.DB) error
```

**使用示例**：
```go
// 用户已有的数据库连接
existingDB, _ := sql.Open("mysql", dsn)

// 创建驱动并传入连接
drv := mysql.NewDriver()
drv.ConnectWithDB(existingDB)

// 创建 Migrator
migrator := migrator.NewMigrator(drv, "./migrations", "migrations")
```

### API 2：`ConnectWithGorm` - 传入 GORM 实例

**函数签名**：
```go
// 在 Driver 接口中新增
ConnectWithGorm(gormDB *gorm.DB) error

// 各驱动实现
func (d *Driver) ConnectWithGorm(gormDB *gorm.DB) error
```

**使用示例**：
```go
// 用户已有的 GORM 实例
gormDB, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})

// 创建驱动并传入 GORM 连接
drv := mysql.NewDriver()
drv.ConnectWithGorm(gormDB)

// 创建 Migrator
migrator := migrator.NewMigrator(drv, "./migrations", "migrations")
```

---

## 子任务拆解

| 序号 | 子任务名称 | 用户故事 | 功能描述 | 预期输出 | 负责人 | 优先级 |
|------|-----------|---------|---------|---------|--------|--------|
| 1 | 架构设计 | 作为架构师，我需要设计 API 的接口和实现方式 | 设计 `ConnectWithDB` 和 `ConnectWithGorm` 的接口定义和实现策略 | 架构设计文档 | Architect | P0 |
| 2 | 修改 Driver 接口 | 作为开发者，我需要在 Driver 接口中新增方法 | 在 `pkg/driver/driver.go` 中新增 `ConnectWithDB` 和 `ConnectWithGorm` 方法定义 | 更新后的接口定义 | Engineer | P0 |
| 3 | 实现 MySQL 驱动 | 作为开发者，我需要为 MySQL 驱动实现新 API | 在 `pkg/driver/mysql/driver.go` 中实现两个新方法 | MySQL 驱动代码 | Engineer | P0 |
| 4 | 实现 PostgreSQL 驱动 | 作为开发者，我需要为 PostgreSQL 驱动实现新 API | 在 `pkg/driver/postgres/driver.go` 中实现两个新方法 | PostgreSQL 驱动代码 | Engineer | P0 |
| 5 | 实现 SQLite 驱动 | 作为开发者，我需要为 SQLite 驱动实现新 API | 在 `pkg/driver/sqlite/driver.go` 中实现两个新方法 | SQLite 驱动代码 | Engineer | P0 |
| 6 | 单元测试 | 作为测试工程师，我需要验证新 API 功能正确 | 为三个驱动的新方法编写单元测试 | 测试代码和报告 | Tester | P1 |
| 7 | 代码审查 | 作为审查员，我需要确保代码质量和安全性 | 审查所有新增代码的质量、风格和安全性 | 审查报告 | Code Reviewer | P1 |
| 8 | API 文档 | 作为文档工程师，我需要编写 API 使用文档 | 编写新 API 的使用说明和示例 | API 文档 | API Doc | P2 |

---

## 技术要点

### 1. GORM 依赖处理

GORM 是外部依赖，需要考虑：
- 是否将 GORM 作为可选依赖
- 可通过 `gormDB.DB()` 获取底层 `*sql.DB`

**建议方案**：
- `ConnectWithGorm` 内部调用 `gormDB.DB()` 获取 `*sql.DB`
- 然后复用 `ConnectWithDB` 的逻辑
- GORM 作为可选依赖，仅在使用时引入

### 2. 连接所有权

需要明确：
- 传入的连接由调用方管理生命周期
- `Close()` 方法不应关闭外部传入的连接

**建议方案**：
- 新增 `ownsConnection bool` 字段标记连接所有权
- `Close()` 仅在 `ownsConnection=true` 时关闭连接

### 3. 驱动类型检测

传入 `*sql.DB` 时需要确保驱动类型匹配：
- MySQL 驱动只接受 MySQL 连接
- PostgreSQL 驱动只接受 PostgreSQL 连接

**建议方案**：
- 通过 Ping 验证连接有效性
- 信任调用方传入正确类型的连接

---

## 验收标准

1. **功能完整**：两个 API 在三个驱动中都能正常工作
2. **向后兼容**：现有 `Connect(config)` 方式不受影响
3. **连接管理**：正确处理连接所有权，避免资源泄漏
4. **测试覆盖**：新增代码有对应的单元测试
5. **文档完善**：API 有清晰的使用说明和示例

---

## 状态

- [x] 需求分析完成
- [x] API 设计完成
- [ ] 等待架构师确认设计方案
- [ ] 等待工程师实现
- [ ] 等待测试验证
- [ ] 等待代码审查
- [ ] 等待文档编写

---

*Producer 完成时间: 2026-02-03*
