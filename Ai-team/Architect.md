# Migro 系统架构设计文档

## 任务状态
- **状态**: 进行中
- **负责人**: Architect (Claude Opus 4.5)
- **开始时间**: 2026-02-02
- **最后更新**: 2026-02-02

---

## 1. 架构概述

### 1.1 设计目标
- **极优雅**: 提供流畅的链式 API，代码简洁易读
- **极大方便**: 一条命令完成常见操作，零配置快速上手
- **极度简化**: 隐藏数据库差异，统一操作接口

### 1.2 架构原则
- **分层解耦**: CLI 层、业务逻辑层、数据访问层严格分离
- **接口抽象**: 通过接口定义契约，实现可替换、可测试
- **单一职责**: 每个模块只负责一个明确的功能域
- **依赖倒置**: 高层模块不依赖低层模块，都依赖抽象

### 1.3 整体架构图（文字描述）

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Layer                             │
│  ┌─────┐ ┌────────┐ ┌────┐ ┌──────┐ ┌────────┐ ┌─────────┐  │
│  │init │ │ create │ │ up │ │ down │ │ status │ │ refresh │  │
│  └──┬──┘ └───┬────┘ └─┬──┘ └──┬───┘ └───┬────┘ └────┬────┘  │
└─────┼────────┼────────┼───────┼─────────┼───────────┼────────┘
      │        │        │       │         │           │
      ▼        ▼        ▼       ▼         ▼           ▼
┌─────────────────────────────────────────────────────────────┐
│                    Business Logic Layer                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │   Migrator   │  │ FileManager  │  │ HistoryManager   │   │
│  │  (迁移引擎)  │  │ (文件管理)   │  │   (历史管理)     │   │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘   │
│         │                 │                    │             │
│  ┌──────▼─────────────────▼────────────────────▼──────────┐ │
│  │                    Schema Builder                       │ │
│  │  ┌───────┐ ┌────────┐ ┌───────┐ ┌─────────────────┐    │ │
│  │  │ Table │ │ Column │ │ Index │ │ ForeignKey      │    │ │
│  │  └───────┘ └────────┘ └───────┘ └─────────────────┘    │ │
│  └─────────────────────────┬──────────────────────────────┘ │
└────────────────────────────┼────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                    Data Access Layer                         │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                   Driver Interface                    │   │
│  └──────────────────────────┬───────────────────────────┘   │
│                             │                                │
│    ┌────────────────────────┼────────────────────────┐      │
│    │                        │                        │      │
│    ▼                        ▼                        ▼      │
│  ┌──────────┐         ┌──────────┐          ┌──────────┐   │
│  │  MySQL   │         │ Postgres │          │  SQLite  │   │
│  │  Driver  │         │  Driver  │          │  Driver  │   │
│  └──────────┘         └──────────┘          └──────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. 目录结构设计

```
migro/
├── cmd/                          # CLI 入口
│   └── migro/
│       └── main.go               # 程序入口
├── internal/                     # 内部包（不对外暴露）
│   ├── cli/                      # CLI 命令实现
│   │   ├── root.go               # 根命令
│   │   ├── init.go               # init 命令
│   │   ├── create.go             # create 命令
│   │   ├── up.go                 # up 命令
│   │   ├── down.go               # down 命令
│   │   ├── status.go             # status 命令
│   │   ├── reset.go              # reset 命令
│   │   └── refresh.go            # refresh 命令
│   ├── config/                   # 配置管理
│   │   ├── config.go             # 配置结构定义
│   │   ├── loader.go             # 配置加载器
│   │   └── env.go                # 环境变量处理
│   ├── migrator/                 # 迁移引擎
│   │   ├── migrator.go           # 迁移执行器
│   │   ├── file_manager.go       # 迁移文件管理
│   │   └── history.go            # 迁移历史管理
│   └── generator/                # 代码生成器
│       ├── template.go           # 迁移文件模板
│       └── generator.go          # 文件生成逻辑
├── pkg/                          # 公开包（对外暴露的 API）
│   ├── schema/                   # Schema DSL
│   │   ├── table.go              # Table 构建器
│   │   ├── column.go             # Column 定义
│   │   ├── index.go              # Index 定义
│   │   ├── foreign.go            # ForeignKey 定义
│   │   └── blueprint.go          # Blueprint 接口
│   ├── driver/                   # 数据库驱动
│   │   ├── driver.go             # Driver 接口定义
│   │   ├── mysql/                # MySQL 驱动
│   │   │   ├── driver.go
│   │   │   ├── grammar.go        # SQL 语法生成
│   │   │   └── types.go          # 类型映射
│   │   ├── postgres/             # PostgreSQL 驱动
│   │   │   ├── driver.go
│   │   │   ├── grammar.go
│   │   │   └── types.go
│   │   └── sqlite/               # SQLite 驱动
│   │       ├── driver.go
│   │       ├── grammar.go
│   │       └── types.go
│   └── migration/                # 迁移接口
│       └── migration.go          # Migration 接口定义
├── migrations/                   # 用户迁移文件目录（示例）
├── migro.yaml                    # 配置文件（示例）
├── go.mod
├── go.sum
└── README.md
```

---

## 3. 核心接口定义

### 3.1 Driver 接口（数据访问层核心抽象）

```
非代码，仅用于设计说明

Driver 接口定义了数据库驱动必须实现的契约：

type Driver interface {
    // 连接管理
    Connect(config *Config) error
    Close() error

    // 事务管理
    Begin() (Transaction, error)

    // Schema 操作
    CreateTable(table *schema.Table) error
    AlterTable(table *schema.Table) error
    DropTable(name string) error
    DropTableIfExists(name string) error
    HasTable(name string) (bool, error)

    // 迁移历史
    CreateMigrationsTable() error
    GetExecutedMigrations() ([]MigrationRecord, error)
    RecordMigration(name string, batch int) error
    DeleteMigration(name string) error
    GetLastBatch() (int, error)

    // SQL 生成
    Grammar() Grammar
}
```

**设计动机**:
- 将数据库操作抽象为统一接口，上层业务逻辑无需关心具体数据库类型
- 每个数据库驱动实现自己的 SQL 方言，通过 Grammar 接口隔离差异
- 事务管理独立抽象，确保迁移的原子性

### 3.2 Grammar 接口（SQL 方言抽象）

```
非代码，仅用于设计说明

Grammar 接口负责生成特定数据库的 SQL 语句：

type Grammar interface {
    // 表操作
    CompileCreate(table *schema.Table) string
    CompileAlter(table *schema.Table) []string
    CompileDrop(name string) string
    CompileDropIfExists(name string) string

    // 列类型映射
    TypeString(length int) string
    TypeInteger() string
    TypeBigInteger() string
    TypeText() string
    TypeBoolean() string
    TypeDate() string
    TypeDateTime() string
    TypeTimestamp() string
    TypeJSON() string
    // ... 更多类型

    // 修饰符
    ModifyNullable(column *schema.Column) string
    ModifyDefault(column *schema.Column) string
    ModifyAutoIncrement(column *schema.Column) string
}
```

**设计动机**:
- 不同数据库的 SQL 语法差异通过 Grammar 实现隔离
- 类型映射方法确保 Go 类型到数据库类型的正确转换
- 修饰符方法处理 nullable、default 等列属性的语法差异

### 3.3 Table 构建器（Schema DSL 核心）

```
非代码，仅用于设计说明

Table 结构体提供链式 API 定义表结构：

type Table struct {
    Name       string
    Columns    []*Column
    Indexes    []*Index
    ForeignKeys []*ForeignKey
    Engine     string        // MySQL 专用
    Charset    string        // MySQL 专用
    Collation  string        // MySQL 专用
}

// 列定义方法（返回 *Column 支持链式调用）
func (t *Table) ID() *Column                           // 自增主键
func (t *Table) String(name string, length int) *Column
func (t *Table) Integer(name string) *Column
func (t *Table) BigInteger(name string) *Column
func (t *Table) Text(name string) *Column
func (t *Table) Boolean(name string) *Column
func (t *Table) Date(name string) *Column
func (t *Table) DateTime(name string) *Column
func (t *Table) Timestamp(name string) *Column
func (t *Table) JSON(name string) *Column
func (t *Table) Timestamps()                           // created_at + updated_at
func (t *Table) SoftDeletes()                          // deleted_at

// 索引定义
func (t *Table) Index(columns ...string) *Index
func (t *Table) Unique(columns ...string) *Index
func (t *Table) Primary(columns ...string) *Index

// 外键定义
func (t *Table) Foreign(column string) *ForeignKey
```

### 3.4 Column 构建器（链式修饰符）

```
非代码，仅用于设计说明

type Column struct {
    Name          string
    Type          ColumnType
    Length        int
    Precision     int
    Scale         int
    IsNullable    bool
    DefaultValue  interface{}
    IsAutoIncrement bool
    IsUnsigned    bool
    Comment       string
}

// 链式修饰符方法
func (c *Column) Nullable() *Column
func (c *Column) Default(value interface{}) *Column
func (c *Column) Unsigned() *Column
func (c *Column) AutoIncrement() *Column
func (c *Column) Comment(text string) *Column
func (c *Column) Unique() *Column
func (c *Column) Primary() *Column
```

### 3.5 Migration 接口（用户迁移文件契约）

```
非代码，仅用于设计说明

type Migration interface {
    Up(m *Migrator) error
    Down(m *Migrator) error
}

// Migrator 提供给用户的操作 API
type Migrator struct {
    driver Driver
}

func (m *Migrator) CreateTable(name string, fn func(*Table)) error
func (m *Migrator) AlterTable(name string, fn func(*Table)) error
func (m *Migrator) DropTable(name string) error
func (m *Migrator) DropTableIfExists(name string) error
func (m *Migrator) HasTable(name string) (bool, error)
func (m *Migrator) RenameTable(from, to string) error
func (m *Migrator) Raw(sql string) error
```

---

## 4. 模块详细设计

### 4.1 CLI 模块

**职责**: 解析命令行参数，调用业务逻辑层

**依赖**: 推荐使用 `cobra` 库（Go 生态标准 CLI 框架）

**命令设计**:

| 命令 | 参数 | 功能 |
|------|------|------|
| `migro init` | `--driver` | 初始化项目，生成配置文件和目录 |
| `migro create <name>` | `--table` | 创建迁移文件 |
| `migro up` | `--step`, `--dry-run`, `--force` | 执行迁移 |
| `migro down` | `--step`, `--force` | 回滚迁移 |
| `migro status` | - | 显示迁移状态 |
| `migro reset` | `--force` | 回滚所有迁移 |
| `migro refresh` | `--force` | 回滚并重新执行所有迁移 |

**执行流程**:
1. 解析命令行参数
2. 加载配置文件
3. 初始化数据库驱动
4. 调用 Migrator 执行操作
5. 输出结果

### 4.2 Config 模块

**职责**: 管理配置文件和环境变量

**配置文件结构** (migro.yaml):

```yaml
# 非代码，仅用于设计说明
driver: mysql
connection:
  host: ${DB_HOST:localhost}
  port: ${DB_PORT:3306}
  database: ${DB_NAME:myapp}
  username: ${DB_USER:root}
  password: ${DB_PASS:}
  charset: utf8mb4

migrations:
  path: ./migrations
  table: migrations

environments:
  development:
    connection:
      host: localhost
  production:
    connection:
      host: prod-db.example.com
```

**设计要点**:
- 支持环境变量占位符 `${VAR:default}`
- 支持多环境配置覆盖
- 敏感信息（密码）优先从环境变量读取

### 4.3 Migrator 模块

**职责**: 迁移执行引擎，协调文件管理和历史管理

**核心流程**:

**Up 操作**:
1. 扫描 migrations 目录获取所有迁移文件
2. 查询数据库获取已执行的迁移
3. 计算待执行的迁移列表
4. 按时间戳顺序执行每个迁移的 Up 方法
5. 记录执行结果到 migrations 表

**Down 操作**:
1. 查询最后一批迁移记录
2. 按时间戳倒序执行每个迁移的 Down 方法
3. 从 migrations 表删除记录

**事务边界**:
- 每个迁移文件独立事务
- 单个迁移失败时回滚该迁移，不影响已成功的迁移
- 提供 `--dry-run` 模式预览 SQL 而不执行

### 4.4 History 模块

**职责**: 管理迁移历史记录

**migrations 表结构**:

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INT AUTO_INCREMENT | 主键 |
| migration | VARCHAR(255) | 迁移文件名 |
| batch | INT | 批次号 |
| executed_at | TIMESTAMP | 执行时间 |

**批次管理**:
- 每次 `up` 操作分配新的批次号
- `down` 操作按批次回滚
- 支持 `--step` 参数指定回滚步数

---

## 5. 数据库驱动设计

### 5.1 驱动注册机制

```
非代码，仅用于设计说明

采用工厂模式 + 注册表实现驱动的动态加载：

var drivers = make(map[string]DriverFactory)

type DriverFactory func(config *Config) (Driver, error)

func Register(name string, factory DriverFactory)
func Get(name string) (DriverFactory, error)

// 各驱动在 init() 中自注册
func init() {
    driver.Register("mysql", NewMySQLDriver)
    driver.Register("postgres", NewPostgresDriver)
    driver.Register("sqlite", NewSQLiteDriver)
}
```

### 5.2 MySQL 驱动特性

- 支持 InnoDB/MyISAM 引擎选择
- 支持 charset/collation 设置
- 支持 UNSIGNED 整数类型
- JSON 类型映射到 MySQL JSON

### 5.3 PostgreSQL 驱动特性

- 支持 SERIAL/BIGSERIAL 自增类型
- 支持 JSONB 类型
- 支持数组类型
- 支持 UUID 类型

### 5.4 SQLite 驱动特性

- 类型亲和性处理
- 外键约束需显式启用
- 不支持 ALTER COLUMN（需重建表）

---

## 6. 风险分析与缓解

### 6.1 并发安全

**风险**: 多实例同时执行迁移可能导致冲突

**缓解措施**:
- 使用数据库锁（SELECT ... FOR UPDATE）
- 迁移开始前获取锁，完成后释放
- 提供 `--lock-timeout` 参数

### 6.2 事务与回滚

**风险**: 迁移失败导致数据库状态不一致

**缓解措施**:
- 每个迁移独立事务
- DDL 语句在某些数据库（MySQL）会隐式提交，需特殊处理
- 提供 `--dry-run` 预览模式

### 6.3 大表迁移

**风险**: 大表 DDL 操作可能锁表、超时

**缓解措施**:
- 提供 Raw SQL 接口，用户可使用 pt-online-schema-change 等工具
- 文档中说明大表迁移最佳实践
- 未来版本可考虑集成在线 DDL 工具

### 6.4 类型兼容性

**风险**: 不同数据库类型映射不一致

**缓解措施**:
- 定义明确的类型映射表
- 文档说明各数据库的类型差异
- 提供 Raw 方法让用户编写原生 SQL

---

## 7. 可测试性设计

### 7.1 接口隔离

- 所有外部依赖通过接口注入
- Driver 接口支持 Mock 实现
- 文件系统操作抽象为接口

### 7.2 测试策略

| 层级 | 测试类型 | 覆盖目标 |
|------|---------|---------|
| Schema DSL | 单元测试 | Table/Column/Index 构建逻辑 |
| Grammar | 单元测试 | SQL 生成正确性 |
| Driver | 集成测试 | 数据库操作正确性 |
| Migrator | 集成测试 | 迁移执行流程 |
| CLI | E2E 测试 | 命令行交互 |

### 7.3 测试数据库

- 单元测试使用 Mock Driver
- 集成测试使用 SQLite 内存数据库
- CI 环境使用 Docker 启动 MySQL/PostgreSQL

---

## 8. 扩展性设计

### 8.1 新增数据库驱动

1. 实现 Driver 接口
2. 实现 Grammar 接口
3. 在 init() 中注册驱动
4. 无需修改上层代码

### 8.2 新增列类型

1. 在 ColumnType 枚举中添加类型
2. 在 Table 中添加对应方法
3. 在各 Grammar 中实现类型映射

### 8.3 插件机制（未来）

- 支持迁移前后钩子
- 支持自定义命令
- 支持第三方驱动加载

---

## 9. 依赖选型建议

| 功能 | 推荐库 | 理由 |
|------|-------|------|
| CLI 框架 | cobra | Go 生态标准，功能完善 |
| 配置解析 | viper | 支持多格式、环境变量 |
| MySQL 驱动 | go-sql-driver/mysql | 官方推荐 |
| PostgreSQL 驱动 | lib/pq 或 pgx | pgx 性能更好 |
| SQLite 驱动 | mattn/go-sqlite3 | 成熟稳定 |
| 日志 | zerolog 或 slog | slog 是标准库 |
| 测试 | testify | 断言和 Mock 支持 |

---

## 10. 下一步行动

1. **Engineer** 根据本架构文档开始实现
2. 优先实现 P0 功能：CLI 框架、Schema DSL、MySQL 驱动
3. 建议开发顺序：
   - 第一阶段：项目骨架 + 配置管理 + CLI 框架
   - 第二阶段：Schema DSL + MySQL 驱动
   - 第三阶段：Migrator 引擎 + 历史管理
   - 第四阶段：PostgreSQL/SQLite 驱动
   - 第五阶段：高级功能（dry-run、refresh 等）

---

**任务完成标志**: 本文档已完成系统架构设计，等待 Engineer 开始开发。

**下一步**: 显式调用 `/team` 继续任务。
