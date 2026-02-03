# Migro API 文档 - 外部连接支持

## 概述

本文档描述 Migro 数据库迁移工具新增的两个 API，用于支持用户直接传入已有的数据库连接实例。

**版本**: v1.1.0
**日期**: 2026-02-03

---

## 新增 API 列表

| API | 包路径 | 功能 |
|-----|--------|------|
| `ConnectWithDB` | `pkg/driver/{mysql,postgres,sqlite}` | 使用已有的 `*sql.DB` 连接 |
| `ConnectDriver` | `pkg/driver/gorm` | 使用 GORM 实例连接 |

---

## API 1: ConnectWithDB

### 功能描述

允许用户将已有的 `*sql.DB` 数据库连接传入 Migro 驱动，而非通过配置创建新连接。调用方保留连接的所有权，负责连接的生命周期管理。

### 适用场景

- 项目已有数据库连接管理，希望 Migro 复用现有连接池
- 需要在同一事务中执行迁移和业务逻辑
- 测试场景中使用 mock 或内存数据库
- 微服务架构中共享数据库连接

### 函数签名

```go
// MySQL 驱动
func (d *mysql.Driver) ConnectWithDB(db *sql.DB) error

// PostgreSQL 驱动
func (d *postgres.Driver) ConnectWithDB(db *sql.DB) error

// SQLite 驱动
func (d *sqlite.Driver) ConnectWithDB(db *sql.DB) error
```

### 参数说明

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| db | `*sql.DB` | 是 | 已建立的数据库连接实例 |

### 返回值

| 类型 | 说明 |
|------|------|
| `error` | 成功返回 `nil`，失败返回错误信息 |

### 错误情况

| 错误消息 | 原因 |
|---------|------|
| `{driver}: database connection is nil` | 传入的 db 参数为 nil |
| `{driver}: failed to ping database: {err}` | 连接无效或已关闭，Ping 失败 |

其中 `{driver}` 为驱动名称：`mysql`、`postgres` 或 `sqlite`。

### 使用示例

```go
package main

import (
    "context"
    "database/sql"
    "log"

    "github.com/flyits/migro/pkg/driver/mysql"
    "github.com/flyits/migro/pkg/migrator"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    ctx := context.Background()

    // 用户已有的数据库连接
    dsn := "user:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=true"
    existingDB, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer existingDB.Close()

    // 创建 Migro 驱动并传入连接
    drv := mysql.NewDriver()
    if err := drv.ConnectWithDB(existingDB); err != nil {
        log.Fatal(err)
    }
    defer drv.Close() // 不会关闭 existingDB

    // 创建 Migrator 并执行迁移
    m := migrator.NewMigrator(drv, "./migrations", "migrations")
    if err := m.Up(ctx, 0); err != nil {
        log.Fatal(err)
    }

    // existingDB 仍可继续使用
    rows, _ := existingDB.Query("SELECT * FROM users")
    defer rows.Close()
    // ...
}
```

---

## API 2: ConnectDriver (GORM 适配)

### 功能描述

允许用户将 GORM 的 `*gorm.DB` 实例传入 Migro 驱动。内部通过 `gormDB.DB()` 获取底层 `*sql.DB`，然后调用驱动的 `ConnectWithDB` 方法。

### 适用场景

- 项目使用 GORM 作为 ORM，希望迁移工具与 GORM 共享连接
- 利用 GORM 的连接配置和连接池
- 在 GORM 项目中无缝集成数据库迁移功能

### 包路径

```go
import migrogorm "github.com/flyits/migro/pkg/driver/gorm"
```

### 函数签名

```go
func ConnectDriver(drv DBConnector, gormDB *gorm.DB) error
```

### 接口定义

```go
// DBConnector 是驱动必须实现的接口
type DBConnector interface {
    ConnectWithDB(db *sql.DB) error
}
```

所有 Migro 驱动（mysql、postgres、sqlite）均实现此接口。

### 参数说明

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| drv | `DBConnector` | 是 | Migro 驱动实例（mysql.Driver、postgres.Driver 或 sqlite.Driver） |
| gormDB | `*gorm.DB` | 是 | GORM 数据库实例 |

### 返回值

| 类型 | 说明 |
|------|------|
| `error` | 成功返回 `nil`，失败返回错误信息 |

### 错误情况

| 错误消息 | 原因 |
|---------|------|
| `gorm: driver is nil` | 传入的 drv 参数为 nil |
| `gorm: gorm.DB is nil` | 传入的 gormDB 参数为 nil |
| `gorm: failed to get underlying *sql.DB: {err}` | 无法从 GORM 获取底层连接 |
| `{driver}: ...` | ConnectWithDB 返回的错误 |

### 使用示例

```go
package main

import (
    "context"
    "log"

    migrogorm "github.com/flyits/migro/pkg/driver/gorm"
    "github.com/flyits/migro/pkg/driver/mysql"
    "github.com/flyits/migro/pkg/migrator"
    gormmysql "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func main() {
    ctx := context.Background()

    // 用户已有的 GORM 实例
    dsn := "user:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=true"
    gormDB, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    // 创建 Migro 驱动并通过 GORM 连接
    drv := mysql.NewDriver()
    if err := migrogorm.ConnectDriver(drv, gormDB); err != nil {
        log.Fatal(err)
    }
    defer drv.Close() // 不会关闭 GORM 的连接

    // 创建 Migrator 并执行迁移
    m := migrator.NewMigrator(drv, "./migrations", "migrations")
    if err := m.Up(ctx, 0); err != nil {
        log.Fatal(err)
    }

    // GORM 仍可继续使用
    type User struct {
        ID   uint
        Name string
    }
    var users []User
    gormDB.Find(&users)
    // ...
}
```

---

## 连接所有权语义

### ownsConnection 字段

驱动内部使用 `ownsConnection bool` 字段标记连接所有权：

| 连接方式 | ownsConnection | 含义 |
|---------|----------------|------|
| `Connect(config)` | `true` | 驱动拥有连接，负责关闭 |
| `ConnectWithDB(db)` | `false` | 调用方拥有连接，驱动不关闭 |

### Close() 行为差异

```go
func (d *Driver) Close() error {
    if d.db != nil && d.ownsConnection {
        return d.db.Close()  // 仅关闭自有连接
    }
    return nil  // 外部连接不关闭
}
```

| 场景 | Close() 行为 |
|------|-------------|
| 通过 `Connect(config)` 创建 | 关闭数据库连接 |
| 通过 `ConnectWithDB(db)` 传入 | 不关闭，返回 nil |
| 通过 `ConnectDriver(drv, gormDB)` 传入 | 不关闭，返回 nil |

### 资源管理最佳实践

```go
// 场景 1: 使用 Connect() - 驱动管理连接
drv := mysql.NewDriver()
drv.Connect(config)
defer drv.Close()  // 会关闭连接

// 场景 2: 使用 ConnectWithDB() - 调用方管理连接
existingDB, _ := sql.Open("mysql", dsn)
defer existingDB.Close()  // 调用方负责关闭

drv := mysql.NewDriver()
drv.ConnectWithDB(existingDB)
defer drv.Close()  // 不会关闭 existingDB

// 场景 3: 使用 GORM - GORM 管理连接
gormDB, _ := gorm.Open(...)
sqlDB, _ := gormDB.DB()
defer sqlDB.Close()  // 或让 GORM 管理

drv := mysql.NewDriver()
migrogorm.ConnectDriver(drv, gormDB)
defer drv.Close()  // 不会关闭 GORM 连接
```

---

## 向后兼容性

| 变更点 | 影响 |
|-------|------|
| Driver 接口 | 未修改，完全兼容 |
| Connect() 方法 | 行为不变，完全兼容 |
| Close() 方法 | 原有场景行为不变 |
| 新增 ConnectWithDB | 纯新增，不影响现有代码 |
| 新增 gorm 包 | 独立包，按需引入 |

---

## 注意事项

### 1. 驱动类型匹配

传入的 `*sql.DB` 必须与驱动类型匹配：

- `mysql.Driver` 只接受 MySQL 连接
- `postgres.Driver` 只接受 PostgreSQL 连接
- `sqlite.Driver` 只接受 SQLite 连接

类型不匹配会在首次执行 SQL 时因语法错误失败。

### 2. 连接有效性

`ConnectWithDB` 会执行 `Ping()` 验证连接有效性。传入已关闭的连接会返回错误。

### 3. 并发安全

- `*sql.DB` 本身是并发安全的
- `ownsConnection` 字段仅在初始化时设置，运行时只读
- 无新增并发风险

### 4. 每个驱动实例只应调用一次连接方法

不要在同一个 Driver 实例上多次调用 `Connect()` 或 `ConnectWithDB()`，这可能导致资源泄漏。

---

## 依赖说明

### 核心包

无新增依赖，`ConnectWithDB` 使用标准库 `database/sql`。

### GORM 适配包

使用 `pkg/driver/gorm` 包需要引入 GORM 依赖：

```go
import migrogorm "github.com/flyits/migro/pkg/driver/gorm"
```

go.mod 新增依赖：
- `gorm.io/gorm v1.31.1`

---

## 状态

- [x] API 文档编写完成
- [x] 函数签名说明完成
- [x] 使用示例完成
- [x] 连接所有权语义说明完成
- [x] Close() 行为差异说明完成
- [x] 错误处理说明完成

---

*API Doc 完成时间: 2026-02-03*
