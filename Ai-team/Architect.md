# 架构设计文档

## 1. 现有架构分析

### 1.1 Driver 结构体模式

三个驱动（MySQL、PostgreSQL、SQLite）采用相同的结构模式：

```
（非代码，仅用于设计说明）

type Driver struct {
    db      *sql.DB      // 底层数据库连接
    grammar *Grammar     // SQL 方言生成器
}
```

### 1.2 连接生命周期

当前连接生命周期：

```
NewDriver() → Connect(config) → [使用] → Close()
     ↓              ↓                        ↓
  创建实例     创建 *sql.DB              关闭连接
```

**关键行为**：
- `Connect()` 内部调用 `sql.Open()` 创建连接
- `Connect()` 执行 `Ping()` 验证连接有效性
- `Close()` 无条件关闭 `d.db`
- 驱动拥有连接的完整生命周期控制权

### 1.3 风险点识别

当前 `Close()` 实现存在隐患：

```go
func (d *Driver) Close() error {
    if d.db != nil {
        return d.db.Close()  // 无条件关闭
    }
    return nil
}
```

若引入外部传入连接，此行为将导致：
- 调用方连接被意外关闭
- 连接池被破坏
- 其他依赖该连接的组件失效

---

## 2. 架构设计方案

### 2.1 核心设计决策

#### 决策 1：不修改 Driver 接口

**理由**：
- `Driver` 接口已有 20+ 方法，职责边界清晰
- 新增方法会破坏所有现有实现的兼容性
- 连接方式属于"构造阶段"行为，不属于"运行时"接口契约

**方案**：在具体驱动实现上新增方法，不修改接口定义

#### 决策 2：GORM 支持独立包

**理由**：
- GORM 是重量级依赖（引入大量间接依赖）
- 核心库不应强制依赖 ORM
- 用户可能只使用 `*sql.DB` 方式

**方案**：
- `ConnectWithDB(*sql.DB)` 放在各驱动包内
- GORM 支持放在独立子包 `pkg/driver/gorm/`
- 通过 `gormDB.DB()` 获取 `*sql.DB` 后调用 `ConnectWithDB`

#### 决策 3：连接所有权标记

**理由**：
- 外部传入的连接不应被 `Close()` 关闭
- 需要区分"自有连接"与"借用连接"

**方案**：新增 `ownsConnection bool` 字段

---

### 2.2 结构变更设计

#### 驱动结构体变更

```
（非代码，仅用于设计说明）

type Driver struct {
    db             *sql.DB
    grammar        *Grammar
    ownsConnection bool      // 新增：标记连接所有权
}
```

#### 连接所有权语义

| 连接方式 | ownsConnection | Close() 行为 |
|---------|----------------|-------------|
| `Connect(config)` | true | 关闭连接 |
| `ConnectWithDB(db)` | false | 不关闭连接 |

---

### 2.3 API 设计

#### API 1: ConnectWithDB

```
（非代码，仅用于设计说明）

// 在各驱动包中实现（mysql/postgres/sqlite）
func (d *Driver) ConnectWithDB(db *sql.DB) error

行为规范：
1. 验证 db 非 nil
2. 执行 Ping() 验证连接有效性
3. 设置 d.db = db
4. 设置 d.ownsConnection = false
5. 返回 nil 或错误
```

#### API 2: ConnectWithGorm（独立包）

```
（非代码，仅用于设计说明）

// 在 pkg/driver/gorm/ 包中提供辅助函数
package gorm

func ConnectDriver(drv interface{ ConnectWithDB(*sql.DB) error }, gormDB *gorm.DB) error

行为规范：
1. 调用 gormDB.DB() 获取 *sql.DB
2. 调用 drv.ConnectWithDB(sqlDB)
3. 返回结果
```

---

### 2.4 Close() 行为变更

```
（非代码，仅用于设计说明）

func (d *Driver) Close() error {
    if d.db != nil && d.ownsConnection {
        return d.db.Close()
    }
    return nil
}
```

**行为矩阵**：

| d.db | ownsConnection | 结果 |
|------|----------------|------|
| nil | - | 返回 nil |
| 非 nil | true | 关闭连接 |
| 非 nil | false | 不关闭，返回 nil |

---

## 3. 包结构设计

```
pkg/driver/
├── driver.go           # Driver 接口（不修改）
├── registry.go         # 驱动注册（不修改）
├── mysql/
│   └── driver.go       # 新增 ConnectWithDB 方法
├── postgres/
│   └── driver.go       # 新增 ConnectWithDB 方法
├── sqlite/
│   └── driver.go       # 新增 ConnectWithDB 方法
└── gorm/               # 新增：GORM 适配包
    └── adapter.go      # ConnectDriver 辅助函数
```

---

## 4. 使用场景示例

### 4.1 传入 *sql.DB

```
（非代码，仅用于设计说明）

// 用户代码
existingDB, _ := sql.Open("mysql", dsn)

drv := mysql.NewDriver()
drv.ConnectWithDB(existingDB)

migrator := migrator.NewMigrator(drv, "./migrations", "migrations")
migrator.Up(ctx, 0)

// drv.Close() 不会关闭 existingDB
drv.Close()

// existingDB 仍可继续使用
existingDB.Query(...)
```

### 4.2 传入 GORM 实例

```
（非代码，仅用于设计说明）

import migrogorm "github.com/flyits/migro/pkg/driver/gorm"

// 用户代码
gormDB, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})

drv := mysql.NewDriver()
migrogorm.ConnectDriver(drv, gormDB)

migrator := migrator.NewMigrator(drv, "./migrations", "migrations")
migrator.Up(ctx, 0)
```

---

## 5. 风险评估与缓解

### 5.1 连接状态不一致风险

**场景**：用户传入已关闭的连接

**缓解**：`ConnectWithDB` 执行 `Ping()` 验证

### 5.2 驱动类型不匹配风险

**场景**：MySQL 驱动传入 PostgreSQL 连接

**评估**：
- Go 的 `database/sql` 是抽象层，底层驱动不同会导致 SQL 语法错误
- 无法在连接层面检测驱动类型

**缓解**：
- 文档明确说明责任边界
- 信任调用方传入正确类型
- 首次执行迁移时会因 SQL 语法错误快速失败

### 5.3 并发安全

**评估**：
- `*sql.DB` 本身是并发安全的
- `ownsConnection` 字段仅在初始化时设置，运行时只读
- 无新增并发风险

### 5.4 资源泄漏风险

**场景**：用户忘记关闭自有连接

**评估**：
- 这是调用方责任，与 Migro 无关
- 文档需明确说明所有权语义

---

## 6. 向后兼容性分析

| 变更点 | 影响 | 兼容性 |
|-------|------|--------|
| Driver 接口 | 不修改 | ✅ 完全兼容 |
| 驱动结构体新增字段 | 内部实现 | ✅ 完全兼容 |
| Connect() 行为 | 设置 ownsConnection=true | ✅ 完全兼容 |
| Close() 行为 | 条件关闭 | ✅ 完全兼容（原有场景不变） |
| 新增 ConnectWithDB | 新方法 | ✅ 完全兼容 |
| 新增 gorm 包 | 新包 | ✅ 完全兼容 |

---

## 7. 实现检查清单

### 7.1 各驱动需实现

- [ ] 新增 `ownsConnection bool` 字段
- [ ] 实现 `ConnectWithDB(*sql.DB) error` 方法
- [ ] 修改 `Connect()` 设置 `ownsConnection = true`
- [ ] 修改 `Close()` 检查 `ownsConnection`

### 7.2 新增 GORM 适配包

- [ ] 创建 `pkg/driver/gorm/adapter.go`
- [ ] 实现 `ConnectDriver` 函数
- [ ] 添加 GORM 依赖到 go.mod（仅 gorm 包）

### 7.3 测试要点

- [ ] `ConnectWithDB` 正常连接
- [ ] `ConnectWithDB` 传入 nil 返回错误
- [ ] `ConnectWithDB` 传入无效连接返回错误
- [ ] `Close()` 不关闭外部连接
- [ ] `Close()` 关闭自有连接
- [ ] GORM 适配正常工作

---

## 8. 技术决策记录 (ADR)

### ADR-001: 不修改 Driver 接口

**状态**：已采纳

**背景**：需要新增连接方式，可选择修改接口或在实现层新增方法

**决策**：在具体实现上新增方法，不修改接口

**理由**：
- 接口稳定性优先
- 连接方式是构造行为，不是运行时契约
- 避免破坏现有实现

### ADR-002: GORM 支持独立包

**状态**：已采纳

**背景**：GORM 是重量级依赖

**决策**：GORM 适配放在独立子包

**理由**：
- 核心库保持轻量
- 按需引入依赖
- 符合 Go 依赖管理最佳实践

### ADR-003: 连接所有权标记

**状态**：已采纳

**背景**：需要区分自有连接和借用连接

**决策**：使用 `ownsConnection bool` 字段

**理由**：
- 简单直接
- 运行时零开销
- 语义清晰

---

*架构师完成时间: 2026-02-03*
