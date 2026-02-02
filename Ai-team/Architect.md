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

---

# T9: API 文档网页架构设计

## 任务状态
- **状态**: ✅ 已完成
- **负责人**: Architect (Claude Opus 4.5)
- **开始时间**: 2026-02-02
- **完成时间**: 2026-02-02

---

## 1. 技术架构概述

### 1.1 设计目标

| 目标 | 说明 |
|------|------|
| **零构建依赖** | 纯静态文件，无需 Node.js、Webpack 等构建工具 |
| **快速加载** | 首屏加载 < 2 秒，关键资源优先加载 |
| **易于维护** | 组件化设计，内容与样式分离 |
| **响应式** | 桌面端与移动端自适应布局 |

### 1.2 技术选型决策

| 技术领域 | 选型 | 决策理由 |
|---------|------|---------|
| **HTML** | 语义化 HTML5 | 标准规范，SEO 友好，无需编译 |
| **CSS** | 原生 CSS + CSS Variables | 无框架依赖，变量系统支持主题切换 |
| **JavaScript** | 原生 ES6+ | 现代浏览器原生支持，无需 polyfill |
| **代码高亮** | Prism.js | 轻量（核心 2KB），支持 Go/YAML/Bash，CDN 加载 |
| **图标** | 内联 SVG | 无外部依赖，可通过 CSS 控制颜色 |

### 1.3 架构分层

```
┌─────────────────────────────────────────────────────────────┐
│                      表现层 (HTML)                           │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────────┐   │
│  │ Header  │ │ Sidebar │ │ Content │ │     Footer      │   │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────────┬────────┘   │
└───────┼───────────┼───────────┼────────────────┼────────────┘
        │           │           │                │
        ▼           ▼           ▼                ▼
┌─────────────────────────────────────────────────────────────┐
│                      样式层 (CSS)                            │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐   │
│  │ Variables│ │  Layout  │ │Components│ │  Responsive  │   │
│  │ (主题)   │ │  (布局)  │ │ (组件)   │ │  (响应式)    │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────────┘   │
└─────────────────────────────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────────────────────────────┐
│                      交互层 (JavaScript)                     │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐   │
│  │ Sidebar  │ │   Code   │ │  Scroll  │ │   Mobile     │   │
│  │ Toggle   │ │   Copy   │ │  Smooth  │ │   Menu       │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. 目录结构设计

```
/doc/web/
├── index.html              # 首页
├── getting-started.html    # 快速入门
├── cli-reference.html      # CLI 命令参考
├── schema-api.html         # Schema DSL API（最大页面）
├── configuration.html      # 配置指南
├── database-support.html   # 数据库支持
├── best-practices.html     # 最佳实践
├── examples.html           # 示例代码
│
├── css/
│   ├── variables.css       # CSS 变量（颜色、字体、间距）
│   ├── base.css            # 基础样式（reset、typography）
│   ├── layout.css          # 布局样式（header、sidebar、content）
│   ├── components.css      # 组件样式（code-block、table、callout）
│   └── responsive.css      # 响应式断点
│
├── js/
│   └── main.js             # 所有交互逻辑（单文件，< 5KB）
│
└── assets/
    ├── logo.svg            # Migro Logo
    └── favicon.ico         # 网站图标
```

### 2.1 文件职责说明

| 文件 | 职责 | 预估大小 |
|------|------|---------|
| `variables.css` | 定义全局 CSS 变量，支持主题切换 | < 1KB |
| `base.css` | CSS Reset + 基础排版样式 | < 2KB |
| `layout.css` | 三栏布局（Header/Sidebar/Content） | < 3KB |
| `components.css` | 可复用组件样式 | < 4KB |
| `responsive.css` | 媒体查询断点 | < 2KB |
| `main.js` | 侧边栏折叠、代码复制、平滑滚动 | < 5KB |

**总计**: CSS ~12KB + JS ~5KB + Prism.js ~15KB (CDN) = **~32KB**

---

## 3. 组件设计

### 3.1 Header 组件

```
┌─────────────────────────────────────────────────────────────┐
│  [Logo] Migro                              [GitHub] [Menu]  │
└─────────────────────────────────────────────────────────────┘
```

**结构说明**（非代码，仅用于设计说明）:
- 固定在页面顶部（`position: fixed`）
- 高度: 60px
- 左侧: Logo + 项目名称
- 右侧: GitHub 链接 + 移动端菜单按钮
- 背景: 白色，底部 1px 边框

### 3.2 Sidebar 组件

```
┌──────────────────┐
│ 快速入门          │
│ CLI 命令参考      │
│ ▼ Schema API     │
│   ├─ Table 构建器 │
│   ├─ 列类型       │
│   ├─ 列修饰符     │
│   ├─ 索引        │
│   └─ 外键        │
│ 配置指南          │
│ 数据库支持        │
│ 最佳实践          │
│ 示例代码          │
└──────────────────┘
```

**结构说明**:
- 固定在左侧（`position: fixed`）
- 宽度: 260px
- 支持二级菜单展开/折叠
- 当前页面高亮显示
- 移动端: 默认隐藏，点击菜单按钮滑出

**交互行为**:
1. 点击一级菜单: 跳转页面或展开子菜单
2. 点击二级菜单: 跳转到页面锚点
3. 当前页面自动高亮
4. 移动端点击遮罩层关闭侧边栏

### 3.3 Content 组件

```
┌─────────────────────────────────────────────────────────────┐
│ # 页面标题                                                   │
│                                                              │
│ 正文内容...                                                  │
│                                                              │
│ ## 二级标题                                                  │
│                                                              │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ ```go                                        [Copy]     │ │
│ │ func main() {                                           │ │
│ │     fmt.Println("Hello")                                │ │
│ │ }                                                       │ │
│ │ ```                                                     │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                              │
│ | 列1 | 列2 | 列3 |                                         │
│ |-----|-----|-----|                                         │
│ | A   | B   | C   |                                         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**结构说明**:
- 主内容区域，左侧留出 Sidebar 宽度
- 最大宽度: 900px，居中显示
- 内边距: 40px（桌面端）/ 20px（移动端）

### 3.4 Code Block 组件

**结构说明**:
- 深色背景（#1e1e1e 或 #282c34）
- 右上角复制按钮
- 左上角语言标签（Go / YAML / Bash）
- 使用 Prism.js 语法高亮
- 支持横向滚动（长代码行）

**复制交互**:
1. 点击复制按钮
2. 按钮文字变为 "Copied!"
3. 2 秒后恢复为 "Copy"

### 3.5 Table 组件

**结构说明**:
- 全宽表格，响应式横向滚动
- 表头背景: 浅灰色
- 斑马纹行背景
- 边框: 1px 浅灰色

### 3.6 Callout 组件

```
┌─────────────────────────────────────────────────────────────┐
│ ⚠️ 注意                                                      │
│ SQLite 不支持 MODIFY COLUMN，需要重建表。                    │
└─────────────────────────────────────────────────────────────┘
```

**类型**:
- `info`: 蓝色边框，信息提示
- `warning`: 黄色边框，警告提示
- `danger`: 红色边框，危险提示
- `tip`: 绿色边框，技巧提示

---

## 4. CSS 架构设计

### 4.1 CSS 变量定义

```
非代码，仅用于设计说明

:root {
  /* 颜色系统 */
  --color-primary: #3b82f6;        /* 主色调：蓝色 */
  --color-primary-dark: #2563eb;
  --color-text: #1f2937;           /* 正文颜色 */
  --color-text-light: #6b7280;     /* 次要文字 */
  --color-bg: #ffffff;             /* 背景色 */
  --color-bg-secondary: #f9fafb;   /* 次要背景 */
  --color-border: #e5e7eb;         /* 边框色 */
  --color-code-bg: #1e1e1e;        /* 代码块背景 */

  /* 字体系统 */
  --font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  --font-mono: "SF Mono", "Fira Code", Consolas, monospace;
  --font-size-base: 16px;
  --font-size-sm: 14px;
  --font-size-lg: 18px;
  --font-size-xl: 24px;
  --font-size-2xl: 32px;

  /* 间距系统 */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;
  --spacing-2xl: 48px;

  /* 布局尺寸 */
  --header-height: 60px;
  --sidebar-width: 260px;
  --content-max-width: 900px;

  /* 圆角 */
  --radius-sm: 4px;
  --radius-md: 8px;

  /* 阴影 */
  --shadow-sm: 0 1px 2px rgba(0,0,0,0.05);
  --shadow-md: 0 4px 6px rgba(0,0,0,0.1);
}
```

### 4.2 布局系统

**三栏布局结构**:
```
┌─────────────────────────────────────────────────────────────┐
│                        Header (fixed)                        │
├──────────────┬──────────────────────────────────────────────┤
│              │                                               │
│   Sidebar    │                  Content                      │
│   (fixed)    │                  (scrollable)                 │
│              │                                               │
│              │                                               │
│              ├──────────────────────────────────────────────┤
│              │                  Footer                       │
└──────────────┴──────────────────────────────────────────────┘
```

**布局实现要点**:
- Header: `position: fixed; top: 0; width: 100%; z-index: 100;`
- Sidebar: `position: fixed; left: 0; top: 60px; width: 260px; height: calc(100vh - 60px);`
- Content: `margin-left: 260px; margin-top: 60px; padding: 40px;`

### 4.3 响应式断点

| 断点 | 宽度 | 布局变化 |
|------|------|---------|
| Desktop | > 1024px | 三栏布局，Sidebar 固定显示 |
| Tablet | 768px - 1024px | Sidebar 收窄至 200px |
| Mobile | < 768px | Sidebar 隐藏，点击菜单按钮显示 |

**移动端适配要点**:
- Sidebar 默认隐藏，通过 `transform: translateX(-100%)` 实现
- 点击菜单按钮添加 `.sidebar-open` 类，Sidebar 滑入
- 显示遮罩层，点击遮罩关闭 Sidebar
- Content 区域全宽显示

---

## 5. JavaScript 架构设计

### 5.1 模块划分

```
main.js
├── initSidebar()        # 侧边栏交互
│   ├── toggleSubmenu()  # 展开/折叠子菜单
│   ├── highlightCurrent() # 高亮当前页面
│   └── mobileToggle()   # 移动端开关
│
├── initCodeCopy()       # 代码复制功能
│   └── copyToClipboard() # 复制到剪贴板
│
├── initSmoothScroll()   # 平滑滚动
│   └── scrollToAnchor() # 锚点跳转
│
└── initBackToTop()      # 返回顶部按钮
```

### 5.2 核心交互逻辑

**侧边栏子菜单展开**:
```
非代码，仅用于设计说明

1. 监听一级菜单点击事件
2. 如果有子菜单:
   - 切换 .expanded 类
   - 子菜单高度从 0 过渡到 auto（使用 max-height 技巧）
3. 如果无子菜单:
   - 正常跳转页面
```

**代码复制**:
```
非代码，仅用于设计说明

1. 为每个 <pre><code> 块动态添加复制按钮
2. 点击按钮时:
   - 获取 <code> 元素的 textContent
   - 调用 navigator.clipboard.writeText()
   - 按钮文字变为 "Copied!"
   - 2 秒后恢复
3. 降级处理: 不支持 Clipboard API 时使用 execCommand
```

**移动端侧边栏**:
```
非代码，仅用于设计说明

1. 点击菜单按钮:
   - 给 body 添加 .sidebar-open 类
   - Sidebar 通过 CSS transform 滑入
   - 显示遮罩层
2. 点击遮罩层或关闭按钮:
   - 移除 .sidebar-open 类
   - Sidebar 滑出
   - 隐藏遮罩层
```

### 5.3 Prism.js 集成

**CDN 加载**:
```
非代码，仅用于设计说明

<!-- 在 </body> 前加载 -->
<script src="https://cdn.jsdelivr.net/npm/prismjs@1.29.0/prism.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/prismjs@1.29.0/components/prism-go.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/prismjs@1.29.0/components/prism-yaml.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/prismjs@1.29.0/components/prism-bash.min.js"></script>
<link href="https://cdn.jsdelivr.net/npm/prismjs@1.29.0/themes/prism-tomorrow.min.css" rel="stylesheet">
```

**支持的语言**:
- `go`: Go 代码
- `yaml`: 配置文件
- `bash`: 命令行
- `sql`: SQL 语句（可选）

---

## 6. 页面模板设计

### 6.1 HTML 基础结构

```
非代码，仅用于设计说明

<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>页面标题 - Migro 文档</title>

    <!-- CSS -->
    <link rel="stylesheet" href="css/variables.css">
    <link rel="stylesheet" href="css/base.css">
    <link rel="stylesheet" href="css/layout.css">
    <link rel="stylesheet" href="css/components.css">
    <link rel="stylesheet" href="css/responsive.css">

    <!-- Prism.js 主题 -->
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/prismjs@1.29.0/themes/prism-tomorrow.min.css">
</head>
<body>
    <!-- Header -->
    <header class="header">...</header>

    <!-- Sidebar -->
    <aside class="sidebar">...</aside>

    <!-- Overlay (移动端) -->
    <div class="overlay"></div>

    <!-- Main Content -->
    <main class="content">
        <article class="article">
            <!-- 页面内容 -->
        </article>

        <!-- Footer -->
        <footer class="footer">...</footer>
    </main>

    <!-- Prism.js -->
    <script src="https://cdn.jsdelivr.net/npm/prismjs@1.29.0/prism.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/prismjs@1.29.0/components/prism-go.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/prismjs@1.29.0/components/prism-yaml.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/prismjs@1.29.0/components/prism-bash.min.js"></script>

    <!-- Main JS -->
    <script src="js/main.js"></script>
</body>
</html>
```

### 6.2 Sidebar 导航结构

```
非代码，仅用于设计说明

<nav class="nav">
    <ul class="nav-list">
        <li class="nav-item">
            <a href="index.html" class="nav-link">首页</a>
        </li>
        <li class="nav-item">
            <a href="getting-started.html" class="nav-link">快速入门</a>
        </li>
        <li class="nav-item">
            <a href="cli-reference.html" class="nav-link">CLI 命令参考</a>
        </li>
        <li class="nav-item has-submenu">
            <a href="schema-api.html" class="nav-link">
                Schema API
                <span class="nav-arrow">▼</span>
            </a>
            <ul class="nav-submenu">
                <li><a href="schema-api.html#table-builder">Table 构建器</a></li>
                <li><a href="schema-api.html#column-types">列类型</a></li>
                <li><a href="schema-api.html#column-modifiers">列修饰符</a></li>
                <li><a href="schema-api.html#indexes">索引</a></li>
                <li><a href="schema-api.html#foreign-keys">外键</a></li>
                <li><a href="schema-api.html#alter-table">ALTER TABLE</a></li>
            </ul>
        </li>
        <!-- 更多菜单项... -->
    </ul>
</nav>
```

### 6.3 代码块结构

```
非代码，仅用于设计说明

<div class="code-block">
    <div class="code-header">
        <span class="code-lang">go</span>
        <button class="code-copy" data-clipboard-target="#code-1">Copy</button>
    </div>
    <pre><code id="code-1" class="language-go">
func main() {
    fmt.Println("Hello, Migro!")
}
    </code></pre>
</div>
```

---

## 7. 响应式设计方案

### 7.1 断点策略

```
非代码，仅用于设计说明

/* 移动端优先 */
.sidebar { display: none; }
.content { margin-left: 0; }

/* 平板端 */
@media (min-width: 768px) {
    .sidebar {
        display: block;
        width: 200px;
    }
    .content {
        margin-left: 200px;
    }
}

/* 桌面端 */
@media (min-width: 1024px) {
    .sidebar {
        width: 260px;
    }
    .content {
        margin-left: 260px;
    }
}
```

### 7.2 移动端适配要点

| 组件 | 桌面端 | 移动端 |
|------|--------|--------|
| Header | Logo + 导航链接 | Logo + 菜单按钮 |
| Sidebar | 固定显示 | 隐藏，点击滑出 |
| Content | 左侧留白 260px | 全宽显示 |
| Code Block | 横向滚动 | 横向滚动 + 字体缩小 |
| Table | 正常显示 | 横向滚动容器 |

### 7.3 触摸优化

- 点击区域最小 44x44px
- 按钮间距足够，避免误触
- 侧边栏支持滑动手势关闭（可选）

---

## 8. 性能优化策略

### 8.1 资源加载优化

| 策略 | 实现方式 |
|------|---------|
| CSS 内联关键样式 | 将 `variables.css` 和 `base.css` 内联到 `<head>` |
| JS 延迟加载 | 将 `<script>` 放在 `</body>` 前 |
| Prism.js CDN | 使用 jsDelivr CDN，支持 HTTP/2 |
| 图片懒加载 | Logo 使用内联 SVG，无需额外请求 |

### 8.2 缓存策略

- 静态资源添加版本号: `style.css?v=1.0.0`
- 设置长期缓存头（由服务器配置）
- CDN 资源自带缓存

### 8.3 预期性能指标

| 指标 | 目标值 |
|------|--------|
| 首次内容绘制 (FCP) | < 1s |
| 最大内容绘制 (LCP) | < 2s |
| 累积布局偏移 (CLS) | < 0.1 |
| 总资源大小 | < 100KB (不含 CDN) |

---

## 9. 开发指南

### 9.1 Engineer 开发顺序建议

1. **第一阶段: 基础框架**
   - 创建目录结构
   - 编写 CSS 变量和基础样式
   - 实现三栏布局
   - 完成 Header 和 Sidebar 组件

2. **第二阶段: 核心页面**
   - 首页 (index.html)
   - 快速入门 (getting-started.html)
   - CLI 命令参考 (cli-reference.html)

3. **第三阶段: API 文档**
   - Schema API (schema-api.html) - 最大页面
   - 配置指南 (configuration.html)
   - 数据库支持 (database-support.html)

4. **第四阶段: 辅助页面**
   - 最佳实践 (best-practices.html)
   - 示例代码 (examples.html)

5. **第五阶段: 交互完善**
   - 代码复制功能
   - 移动端适配
   - 响应式测试

### 9.2 内容来源

所有文档内容从 `README.md` 提取，确保一致性:

| 页面 | README.md 对应章节 |
|------|-------------------|
| 首页 | 特性 + 安装 + 快速入门 |
| 快速入门 | 快速入门 |
| CLI 命令参考 | CLI 命令参考 |
| Schema API | Schema DSL API |
| 配置指南 | 配置文件 |
| 数据库支持 | 数据库差异 |
| 最佳实践 | 最佳实践 |
| 示例代码 | 完整迁移示例 |

---

## 10. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| Prism.js CDN 不可用 | 代码无高亮 | 提供本地 fallback 文件 |
| 移动端布局错乱 | 用户体验差 | 移动端优先开发，充分测试 |
| 内容与 README 不同步 | 文档不准确 | 建立内容更新流程 |
| 浏览器兼容性问题 | 部分功能失效 | 使用 CSS/JS 特性检测 |

---

## 11. 验收检查清单

- [ ] 所有 8 个页面内容完整
- [ ] 侧边栏导航正常，支持二级菜单
- [ ] 代码高亮正常（Go/YAML/Bash）
- [ ] 代码复制功能正常
- [ ] 响应式布局正常（桌面/平板/移动）
- [ ] 移动端侧边栏滑动正常
- [ ] 页面加载速度 < 2 秒
- [ ] 无 JavaScript 控制台错误
- [ ] Chrome/Firefox/Safari/Edge 兼容

---

**任务完成标志**: T9 架构设计已完成，等待 Engineer 开始开发 (T10)。

**下一步**: 显式调用 `/team` 继续任务流程。
