# Migro

[![Version](https://img.shields.io/badge/version-v0.0.3-blue.svg)](https://github.com/flyits/migro/releases/tag/v0.0.3)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

**Migro** 是一个 Go 语言数据库迁移工具，对标 PHP Laravel 框架的迁移功能，提供极优雅、极大方便、极度简化的数据库迁移体验。

> 📋 查看 [CHANGELOG.md](CHANGELOG.md) 了解版本更新历史

## 特性

- **流畅的链式 API** - Laravel 风格的 Schema DSL，代码简洁易读
- **多数据库支持** - MySQL、PostgreSQL、SQLite
- **事务保护** - PostgreSQL/SQLite 支持事务 DDL，确保迁移原子性
- **环境变量支持** - 配置文件支持 `${VAR:default}` 语法
- **批次管理** - 支持按批次回滚迁移
- **Dry Run 模式** - 预览 SQL 而不执行

## 安装

```bash
go install github.com/flyits/migro/cmd/migro@latest
```

或者从源码构建：

```bash
git clone https://github.com/flyits/migro.git
cd migro
go build -o migro ./cmd/migro
```

## 快速入门

### 1. 初始化项目

```bash
migro init --driver mysql
```

这将创建：
- `migro.yaml` - 配置文件
- `migrations/` - 迁移文件目录

### 2. 创建迁移文件

```bash
migro create create_users_table
```

生成文件：`migrations/20260202150405_create_users_table.go`

### 3. 编写迁移

```go
package migrations

import (
    "context"
    "github.com/flyits/migro/internal/migrator"
    "github.com/flyits/migro/pkg/schema"
)

type CreateUsersTable struct{}

func (m *CreateUsersTable) Name() string {
    return "20260202150405_create_users_table"
}

func (m *CreateUsersTable) Up(ctx context.Context, e *migrator.Executor) error {
    return e.CreateTable(ctx, "users", func(t *schema.Table) {
        t.ID()
        t.String("name", 100)
        t.String("email", 100).Unique()
        t.String("password", 255)
        t.Timestamps()
    })
}

func (m *CreateUsersTable) Down(ctx context.Context, e *migrator.Executor) error {
    return e.DropTableIfExists(ctx, "users")
}
```

### 4. 执行迁移

```bash
migro up
```

### 5. 查看状态

```bash
migro status
```

---

## CLI 命令参考

### migro init

初始化迁移环境。

```bash
migro init [flags]
```

**参数：**
| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--driver` | 数据库驱动 (mysql/postgres/sqlite) | mysql |

**示例：**
```bash
migro init --driver postgres
```

---

### migro create

创建新的迁移文件。

```bash
migro create <name> [flags]
```

**参数：**
| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--table` | 指定表名（用于生成模板） | - |

**示例：**
```bash
migro create create_users_table
migro create add_phone_to_users --table users
```

---

### migro up

执行待执行的迁移。

```bash
migro up [flags]
```

**参数：**
| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--step` | 执行指定数量的迁移 | 0 (全部) |
| `--dry-run` | 预览 SQL 而不执行 | false |
| `--force` | 跳过确认提示 | false |

**示例：**
```bash
migro up              # 执行所有待执行迁移
migro up --step=1     # 只执行一个迁移
migro up --dry-run    # 预览 SQL
```

---

### migro down

回滚迁移。

```bash
migro down [flags]
```

**参数：**
| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--step` | 回滚指定数量的迁移 | 0 (最后一批) |
| `--force` | 跳过确认提示 | false |

**示例：**
```bash
migro down            # 回滚最后一批迁移
migro down --step=3   # 回滚最近 3 个迁移
```

---

### migro status

显示迁移状态。

```bash
migro status
```

**输出示例：**
```
+----+----------------------------------------+-------+---------------------+
| #  | Migration                              | Batch | Executed At         |
+----+----------------------------------------+-------+---------------------+
| 1  | 20260202150405_create_users_table      | 1     | 2026-02-02 15:04:05 |
| 2  | 20260202150410_create_posts_table      | 1     | 2026-02-02 15:04:10 |
| 3  | 20260202150415_add_phone_to_users      | -     | Pending             |
+----+----------------------------------------+-------+---------------------+
```

---

### migro reset

回滚所有迁移。

```bash
migro reset [flags]
```

**参数：**
| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--force` | 跳过确认提示 | false |

---

### migro refresh

回滚所有迁移并重新执行。

```bash
migro refresh [flags]
```

**参数：**
| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--force` | 跳过确认提示 | false |

---

### migro version

显示版本信息。

```bash
migro version
```

**输出示例：**
```
migro version v0.0.3
  Git commit: 7737a98
  Build date: 2026-02-04T10:30:00Z
  Go version: go1.21.0
  OS/Arch:    linux/amd64
```

---

### migro upgrade

检查并升级到最新版本。

```bash
migro upgrade [flags]
```

**参数：**
| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--check` | 仅检查更新，不安装 | false |

**示例：**
```bash
migro upgrade           # 检查并升级到最新版本
migro upgrade --check   # 仅检查是否有新版本
```

---

## Schema DSL API

### Table 构建器

#### 创建表

```go
e.CreateTable(ctx, "users", func(t *schema.Table) {
    // 定义列、索引、外键
})
```

#### 修改表

```go
e.AlterTable(ctx, "users", func(t *schema.Table) {
    // 添加、修改、删除列
})
```

#### 删除表

```go
e.DropTable(ctx, "users")
e.DropTableIfExists(ctx, "users")
```

#### 重命名表

```go
e.RenameTable(ctx, "old_name", "new_name")
```

#### 检查表是否存在

```go
exists, err := e.HasTable(ctx, "users")
```

---

### 列类型

| 方法 | 说明 | MySQL | PostgreSQL | SQLite |
|------|------|-------|------------|--------|
| `ID()` | 自增主键 | BIGINT UNSIGNED AUTO_INCREMENT | BIGSERIAL | INTEGER PRIMARY KEY AUTOINCREMENT |
| `String(name, length)` | 字符串 | VARCHAR(length) | VARCHAR(length) | TEXT |
| `Text(name)` | 长文本 | TEXT | TEXT | TEXT |
| `Integer(name)` | 整数 | INT | INTEGER | INTEGER |
| `BigInteger(name)` | 大整数 | BIGINT | BIGINT | INTEGER |
| `SmallInteger(name)` | 小整数 | SMALLINT | SMALLINT | INTEGER |
| `TinyInteger(name)` | 微整数 | TINYINT | SMALLINT | INTEGER |
| `Float(name)` | 浮点数 | FLOAT | REAL | REAL |
| `Double(name)` | 双精度 | DOUBLE | DOUBLE PRECISION | REAL |
| `Decimal(name, p, s)` | 定点数 | DECIMAL(p,s) | DECIMAL(p,s) | REAL |
| `Boolean(name)` | 布尔值 | TINYINT(1) | BOOLEAN | INTEGER |
| `Date(name)` | 日期 | DATE | DATE | TEXT |
| `DateTime(name)` | 日期时间 | DATETIME | TIMESTAMP | TEXT |
| `Timestamp(name)` | 时间戳 | TIMESTAMP | TIMESTAMP | TEXT |
| `Time(name)` | 时间 | TIME | TIME | TEXT |
| `JSON(name)` | JSON | JSON | JSONB | TEXT |
| `Binary(name)` | 二进制 | BLOB | BYTEA | BLOB |
| `UUID(name)` | UUID | CHAR(36) | UUID | TEXT |

#### 便捷方法

```go
// 添加 created_at 和 updated_at 列
t.Timestamps()

// 添加 deleted_at 列（软删除）
t.SoftDeletes()
```

---

### 列修饰符

所有列方法返回 `*Column`，支持链式调用：

```go
t.String("email", 100).Nullable().Unique().Default("").Comment("用户邮箱")
```

| 方法 | 说明 |
|------|------|
| `Nullable()` | 允许 NULL 值 |
| `Default(value)` | 设置默认值 |
| `Unsigned()` | 无符号（仅数值类型） |
| `AutoIncrement()` | 自增 |
| `Primary()` | 主键 |
| `Unique()` | 唯一约束 |
| `Comment(text)` | 列注释 |
| `PlaceAfter(column)` | 放在指定列后（MySQL） |

---

### 索引

```go
// 普通索引
t.Index("email")
t.Index("first_name", "last_name")  // 复合索引

// 唯一索引
t.Unique("email")

// 主键（复合）
t.Primary("user_id", "role_id")

// 命名索引
t.Index("email").Named("idx_users_email")

// 全文索引（MySQL）
t.Index("content").Fulltext()
```

#### 删除索引

```go
e.AlterTable(ctx, "users", func(t *schema.Table) {
    t.DropIndex("idx_users_email")
})
```

---

### 外键

```go
// 基本外键
t.Foreign("user_id").References("users", "id")

// 带级联操作
t.Foreign("user_id").
    References("users", "id").
    OnDeleteCascade().
    OnUpdateCascade()

// 命名外键
t.Foreign("user_id").
    Named("fk_posts_user").
    References("users", "id")
```

#### 外键动作

| 方法 | 说明 |
|------|------|
| `OnDeleteCascade()` | 删除时级联删除 |
| `OnDeleteSetNull()` | 删除时设为 NULL |
| `OnDeleteRestrict()` | 删除时限制（默认） |
| `OnUpdateCascade()` | 更新时级联更新 |
| `OnUpdateSetNull()` | 更新时设为 NULL |
| `OnUpdateRestrict()` | 更新时限制（默认） |

#### 删除外键

```go
e.AlterTable(ctx, "posts", func(t *schema.Table) {
    t.DropForeign("fk_posts_user")
})
```

---

### ALTER TABLE 操作

```go
e.AlterTable(ctx, "users", func(t *schema.Table) {
    // 添加列
    t.String("phone", 20).Nullable()

    // 删除列
    t.DropColumn("old_column")

    // 重命名列
    t.RenameColumn("old_name", "new_name")

    // 修改列类型
    t.ChangeText("description")  // 将 description 列改为 TEXT 类型

    // 添加索引
    t.Index("phone")

    // 删除索引
    t.DropIndex("idx_old")

    // 添加外键
    t.Foreign("department_id").References("departments", "id")

    // 删除外键
    t.DropForeign("fk_old")
})
```

---

### 修改列类型

使用 `Change*` 系列方法修改现有列的类型。这些方法会生成 `ALTER TABLE ... MODIFY COLUMN` 语句。

#### 基本用法

```go
e.AlterTable(ctx, "users", func(t *schema.Table) {
    // 将 VARCHAR 改为 TEXT
    t.ChangeText("bio")

    // 修改 VARCHAR 长度
    t.ChangeString("email", 320)

    // 将 INT 改为 BIGINT
    t.ChangeBigInteger("view_count").Unsigned()

    // 修改 DECIMAL 精度
    t.ChangeDecimal("price", 10, 2).Default(0.00)
})
```

#### 修改列类型方法

| 方法 | 说明 | 生成的 SQL (MySQL) |
|------|------|-------------------|
| `ChangeColumn(name, type)` | 通用列类型修改 | `MODIFY COLUMN name TYPE` |
| `ChangeString(name, length)` | 修改为 VARCHAR | `MODIFY COLUMN name VARCHAR(length)` |
| `ChangeText(name)` | 修改为 TEXT | `MODIFY COLUMN name TEXT` |
| `ChangeInteger(name)` | 修改为 INT | `MODIFY COLUMN name INT` |
| `ChangeBigInteger(name)` | 修改为 BIGINT | `MODIFY COLUMN name BIGINT` |
| `ChangeSmallInteger(name)` | 修改为 SMALLINT | `MODIFY COLUMN name SMALLINT` |
| `ChangeTinyInteger(name)` | 修改为 TINYINT | `MODIFY COLUMN name TINYINT` |
| `ChangeFloat(name)` | 修改为 FLOAT | `MODIFY COLUMN name FLOAT` |
| `ChangeDouble(name)` | 修改为 DOUBLE | `MODIFY COLUMN name DOUBLE` |
| `ChangeDecimal(name, p, s)` | 修改为 DECIMAL | `MODIFY COLUMN name DECIMAL(p,s)` |
| `ChangeBoolean(name)` | 修改为 BOOLEAN | `MODIFY COLUMN name TINYINT(1)` |
| `ChangeDate(name)` | 修改为 DATE | `MODIFY COLUMN name DATE` |
| `ChangeDateTime(name)` | 修改为 DATETIME | `MODIFY COLUMN name DATETIME` |
| `ChangeTimestamp(name)` | 修改为 TIMESTAMP | `MODIFY COLUMN name TIMESTAMP` |
| `ChangeTime(name)` | 修改为 TIME | `MODIFY COLUMN name TIME` |
| `ChangeJSON(name)` | 修改为 JSON | `MODIFY COLUMN name JSON` |
| `ChangeBinary(name)` | 修改为 BINARY/BLOB | `MODIFY COLUMN name BLOB` |
| `ChangeUUID(name)` | 修改为 UUID | `MODIFY COLUMN name CHAR(36)` |

#### 链式调用

所有 `Change*` 方法返回 `*Column`，支持链式调用修饰符：

```go
t.ChangeBigInteger("user_id").Unsigned().Nullable().Comment("用户ID")
t.ChangeString("email", 320).Unique().Default("")
t.ChangeDecimal("amount", 12, 4).Nullable()
```

#### 完整迁移示例

```go
// Up: 修改列类型
func (m *ModifyColumnsInUsers) Up(ctx context.Context, e *migrator.Executor) error {
    return e.AlterTable(ctx, "users", func(t *schema.Table) {
        t.ChangeText("bio")                              // VARCHAR -> TEXT
        t.ChangeBigInteger("follower_count").Unsigned()  // INT -> BIGINT UNSIGNED
        t.ChangeString("username", 100).Unique()         // 修改长度并添加唯一约束
    })
}

// Down: 恢复原类型
func (m *ModifyColumnsInUsers) Down(ctx context.Context, e *migrator.Executor) error {
    return e.AlterTable(ctx, "users", func(t *schema.Table) {
        t.ChangeString("bio", 500)           // TEXT -> VARCHAR(500)
        t.ChangeInteger("follower_count")    // BIGINT -> INT
        t.ChangeString("username", 50)       // 恢复原长度
    })
}
```

#### 注意事项

1. **数据兼容性**：修改列类型时确保现有数据与新类型兼容
2. **SQLite 限制**：SQLite 不支持 `MODIFY COLUMN`，需要重建表
3. **大表操作**：对于大表，考虑使用在线 DDL 避免锁表：
   ```go
   e.Raw(ctx, "ALTER TABLE users MODIFY COLUMN bio TEXT, ALGORITHM=INPLACE, LOCK=NONE")
   ```

---

### MySQL 专用选项

```go
e.CreateTable(ctx, "users", func(t *schema.Table) {
    t.ID()
    t.String("name", 100)
}).SetEngine("InnoDB").SetCharset("utf8mb4").SetCollation("utf8mb4_unicode_ci")
```

---

### 原生 SQL

```go
e.Raw(ctx, "CREATE INDEX CONCURRENTLY idx_users_email ON users(email)")
```

---

## 配置文件

### migro.yaml

```yaml
# 数据库驱动: mysql, postgres, sqlite
driver: mysql

# 数据库连接配置
connection:
  host: ${DB_HOST:localhost}
  port: ${DB_PORT:3306}
  database: ${DB_NAME:myapp}
  username: ${DB_USER:root}
  password: ${DB_PASS:}
  charset: utf8mb4

# 迁移配置
migrations:
  path: ./migrations
  table: migrations
```

### 环境变量

配置文件支持环境变量占位符：

```yaml
host: ${DB_HOST:localhost}  # 使用 DB_HOST 环境变量，默认 localhost
password: ${DB_PASS:}       # 使用 DB_PASS 环境变量，默认空
```

### 各数据库配置示例

#### MySQL

```yaml
driver: mysql
connection:
  host: localhost
  port: 3306
  database: myapp
  username: root
  password: secret
  charset: utf8mb4
```

#### PostgreSQL

```yaml
driver: postgres
connection:
  host: localhost
  port: 5432
  database: myapp
  username: postgres
  password: secret
```

#### SQLite

```yaml
driver: sqlite
connection:
  database: ./database.db
```

---

## 数据库差异

### 事务 DDL

| 数据库 | 支持事务 DDL |
|--------|-------------|
| MySQL | 否（DDL 会隐式提交） |
| PostgreSQL | 是 |
| SQLite | 是 |

Migro 会自动检测数据库类型，对支持事务 DDL 的数据库使用事务包裹迁移。

### 类型映射

不同数据库的类型映射有所不同，Migro 会自动处理：

- **Boolean**: MySQL 使用 `TINYINT(1)`，PostgreSQL 使用 `BOOLEAN`，SQLite 使用 `INTEGER`
- **JSON**: MySQL 使用 `JSON`，PostgreSQL 使用 `JSONB`，SQLite 使用 `TEXT`
- **自增**: MySQL 使用 `AUTO_INCREMENT`，PostgreSQL 使用 `SERIAL/BIGSERIAL`，SQLite 使用 `AUTOINCREMENT`

### ALTER TABLE 限制

SQLite 对 ALTER TABLE 支持有限：
- 支持：ADD COLUMN、RENAME COLUMN
- 不支持：DROP COLUMN、MODIFY COLUMN（需要重建表）

对于复杂的 SQLite 表修改，建议使用 `Raw()` 方法手动处理。

---

## 完整迁移示例

### 创建用户表

```go
func (m *CreateUsersTable) Up(ctx context.Context, e *migrator.Executor) error {
    return e.CreateTable(ctx, "users", func(t *schema.Table) {
        t.ID()
        t.String("name", 100)
        t.String("email", 100).Unique()
        t.String("password", 255)
        t.String("phone", 20).Nullable()
        t.Boolean("is_active").Default(true)
        t.Timestamps()
        t.SoftDeletes()

        t.Index("email")
        t.Index("phone")
    })
}

func (m *CreateUsersTable) Down(ctx context.Context, e *migrator.Executor) error {
    return e.DropTableIfExists(ctx, "users")
}
```

### 创建文章表（带外键）

```go
func (m *CreatePostsTable) Up(ctx context.Context, e *migrator.Executor) error {
    return e.CreateTable(ctx, "posts", func(t *schema.Table) {
        t.ID()
        t.BigInteger("user_id").Unsigned()
        t.String("title", 200)
        t.Text("content")
        t.String("status", 20).Default("draft")
        t.Timestamps()

        t.Index("user_id")
        t.Index("status")
        t.Foreign("user_id").References("users", "id").OnDeleteCascade()
    })
}

func (m *CreatePostsTable) Down(ctx context.Context, e *migrator.Executor) error {
    return e.DropTableIfExists(ctx, "posts")
}
```

### 修改表结构

```go
func (m *AddAvatarToUsers) Up(ctx context.Context, e *migrator.Executor) error {
    return e.AlterTable(ctx, "users", func(t *schema.Table) {
        t.String("avatar", 255).Nullable()
    })
}

func (m *AddAvatarToUsers) Down(ctx context.Context, e *migrator.Executor) error {
    return e.AlterTable(ctx, "users", func(t *schema.Table) {
        t.DropColumn("avatar")
    })
}
```

---

## 最佳实践

### 1. 迁移命名

使用描述性名称，清晰表达迁移目的：

```bash
migro create create_users_table      # 创建表
migro create add_phone_to_users      # 添加列
migro create drop_legacy_columns     # 删除列
migro create add_index_to_posts      # 添加索引
```

### 2. 始终编写 Down 方法

确保每个迁移都可以回滚：

```go
func (m *Migration) Down(ctx context.Context, e *migrator.Executor) error {
    // 不要留空！
    return e.DropTableIfExists(ctx, "users")
}
```

### 3. 使用事务（PostgreSQL/SQLite）

Migro 会自动为支持事务 DDL 的数据库使用事务，无需手动处理。

### 4. 大表迁移

对于大表，避免锁表操作：

```go
// 使用 Raw SQL 执行在线 DDL
e.Raw(ctx, "ALTER TABLE users ADD COLUMN phone VARCHAR(20), ALGORITHM=INPLACE, LOCK=NONE")
```

### 5. 环境变量

生产环境使用环境变量存储敏感信息：

```bash
export DB_HOST=prod-db.example.com
export DB_PASS=secure_password
migro up
```

---

## 依赖

```
github.com/go-sql-driver/mysql v1.9.3
github.com/lib/pq v1.11.1
github.com/mattn/go-sqlite3 v1.14.33
github.com/spf13/cobra v1.10.2
gopkg.in/yaml.v3 v3.0.1
```

---

## 许可证

MIT License
