# Producer 需求分析文档

## 任务状态
- **状态**: 进行中
- **负责人**: Producer (Claude Sonnet 4)
- **开始时间**: 2026-02-02
- **最后更新**: 2026-02-02

---

## 项目背景
**项目名称**: Migro - Golang 数据库迁移工具
**项目愿景**: 对标 PHP Laravel 框架的数据库迁移工具，实现极优雅、极大方便、极度简化的数据库迁移功能
**目标用户**: Go 开发者、DevOps 工程师、数据库管理员

---

## Laravel 迁移工具核心特性分析

基于 Laravel 12.x 官方文档和实践经验，Laravel 迁移工具的核心特性包括：

### 1. 迁移文件管理
- **生成迁移文件**: `php artisan make:migration create_users_table`
- **迁移文件命名**: 时间戳前缀 + 描述性名称（如 `2024_01_01_000000_create_users_table.php`）
- **迁移文件结构**: `up()` 和 `down()` 方法，支持正向迁移和回滚
- **迁移压缩**: 将多个迁移文件合并为单个 schema 文件

### 2. 迁移执行
- **执行所有待执行迁移**: `php artisan migrate`
- **执行指定步数**: `php artisan migrate --step=1`
- **强制执行（生产环境）**: `php artisan migrate --force`
- **预览 SQL**: `php artisan migrate --pretend`
- **隔离执行**: 支持租户隔离的多数据库迁移

### 3. 回滚操作
- **回滚最后一批迁移**: `php artisan migrate:rollback`
- **回滚指定步数**: `php artisan migrate:rollback --step=5`
- **回滚所有迁移**: `php artisan migrate:reset`
- **回滚并重新执行**: `php artisan migrate:refresh`
- **删除所有表并重新迁移**: `php artisan migrate:fresh`

### 4. 状态查询
- **查看迁移状态**: `php artisan migrate:status`
- **显示已执行和待执行的迁移列表**

### 5. 表操作
- **创建表**: `Schema::create('users', function (Blueprint $table) {...})`
- **修改表**: `Schema::table('users', function (Blueprint $table) {...})`
- **重命名表**: `Schema::rename('from', 'to')`
- **删除表**: `Schema::drop('users')` / `Schema::dropIfExists('users')`
- **检查表是否存在**: `Schema::hasTable('users')`

### 6. 列操作
- **创建列**: 支持 50+ 种列类型（string, integer, text, json, timestamp 等）
- **修改列**: 修改列类型、长度、默认值等
- **重命名列**: `$table->renameColumn('from', 'to')`
- **删除列**: `$table->dropColumn('column_name')`
- **列修饰符**: nullable, default, unsigned, autoIncrement, comment 等

### 7. 索引操作
- **创建索引**: `$table->index('email')` / `$table->unique('email')`
- **复合索引**: `$table->index(['first_name', 'last_name'])`
- **重命名索引**: `$table->renameIndex('from', 'to')`
- **删除索引**: `$table->dropIndex('users_email_index')`

### 8. 外键约束
- **创建外键**: `$table->foreign('user_id')->references('id')->on('users')`
- **级联操作**: `onDelete('cascade')` / `onUpdate('cascade')`
- **删除外键**: `$table->dropForeign('posts_user_id_foreign')`

### 9. 迁移事件
- **迁移前后钩子**: 支持在迁移执行前后触发事件

---

## Migro 功能清单设计

基于 Laravel 迁移工具的特性，结合 Go 语言生态和"极优雅、极大方便、极度简化"的目标，Migro 的功能清单如下：

### 核心功能模块

#### 模块 1: CLI 命令行工具
| 功能 | 命令示例 | 优先级 | 说明 |
|------|---------|--------|------|
| 初始化项目 | `migro init` | P0 | 初始化迁移配置文件和目录结构 |
| 创建迁移文件 | `migro create create_users_table` | P0 | 生成带时间戳的迁移文件 |
| 执行迁移 | `migro up` | P0 | 执行所有待执行的迁移 |
| 回滚迁移 | `migro down` | P0 | 回滚最后一批迁移 |
| 查看状态 | `migro status` | P0 | 显示迁移执行状态 |
| 重置迁移 | `migro reset` | P1 | 回滚所有迁移 |
| 刷新迁移 | `migro refresh` | P1 | 回滚并重新执行所有迁移 |
| 指定步数回滚 | `migro down --step=3` | P1 | 回滚指定步数的迁移 |
| 预览 SQL | `migro up --dry-run` | P2 | 预览将要执行的 SQL 语句 |
| 强制执行 | `migro up --force` | P2 | 跳过确认提示强制执行 |

#### 模块 2: 迁移文件 DSL（Go 语言风格）
| 功能 | 代码示例 | 优先级 | 说明 |
|------|---------|--------|------|
| 创建表 | `m.CreateTable("users", func(t *Table) {...})` | P0 | 流畅的表创建 API |
| 修改表 | `m.AlterTable("users", func(t *Table) {...})` | P0 | 流畅的表修改 API |
| 删除表 | `m.DropTable("users")` | P0 | 删除表 |
| 列定义 | `t.String("name", 100).Nullable()` | P0 | 链式调用定义列 |
| 索引定义 | `t.Index("email").Unique()` | P0 | 链式调用定义索引 |
| 外键定义 | `t.Foreign("user_id").References("users", "id").OnDelete("cascade")` | P0 | 链式调用定义外键 |
| 时间戳列 | `t.Timestamps()` | P1 | 自动添加 created_at 和 updated_at |
| 软删除列 | `t.SoftDeletes()` | P1 | 自动添加 deleted_at |

#### 模块 3: 数据库驱动支持
| 数据库 | 优先级 | 说明 |
|--------|--------|------|
| MySQL | P0 | 支持 MySQL 5.7+ |
| PostgreSQL | P0 | 支持 PostgreSQL 10+ |
| SQLite | P0 | 支持 SQLite 3 |
| SQL Server | P1 | 支持 SQL Server 2017+ |

#### 模块 4: 配置管理
| 功能 | 优先级 | 说明 |
|------|--------|------|
| YAML 配置文件 | P0 | 支持 `migro.yaml` 配置数据库连接 |
| 环境变量支持 | P0 | 支持从环境变量读取配置 |
| 多环境配置 | P1 | 支持 dev/test/prod 多环境配置 |

#### 模块 5: 迁移历史管理
| 功能 | 优先级 | 说明 |
|------|--------|------|
| 迁移记录表 | P0 | 自动创建 `migrations` 表记录执行历史 |
| 批次管理 | P0 | 记录每次迁移的批次号，支持批量回滚 |
| 执行时间记录 | P1 | 记录每个迁移的执行时间 |

---

## 用户故事

### 故事 1: 快速初始化项目
**作为** Go 开发者
**我想要** 通过一条命令初始化数据库迁移环境
**以便于** 快速开始使用迁移工具，无需手动配置

**验收标准**:
- 执行 `migro init` 后自动生成配置文件 `migro.yaml`
- 自动创建 `migrations/` 目录
- 配置文件包含数据库连接示例

---

### 故事 2: 优雅创建迁移文件
**作为** Go 开发者
**我想要** 通过命令快速生成迁移文件模板
**以便于** 专注于编写迁移逻辑，而不是手动创建文件

**验收标准**:
- 执行 `migro create create_users_table` 生成带时间戳的文件
- 文件包含 `Up()` 和 `Down()` 方法模板
- 文件名格式: `20260202150405_create_users_table.go`

---

### 故事 3: 流畅的表定义 API
**作为** Go 开发者
**我想要** 使用链式调用的方式定义表结构
**以便于** 代码更简洁、可读性更强

**验收标准**:
```go
func (m *Migration) Up() error {
    return m.CreateTable("users", func(t *migro.Table) {
        t.ID()
        t.String("name", 100).Nullable()
        t.String("email", 100).Unique()
        t.Timestamps()
    })
}
```

---

### 故事 4: 安全的迁移回滚
**作为** DevOps 工程师
**我想要** 在生产环境出现问题时快速回滚迁移
**以便于** 恢复数据库到稳定状态

**验收标准**:
- 执行 `migro down` 回滚最后一批迁移
- 执行 `migro down --step=3` 回滚指定步数
- 回滚前显示确认提示（生产环境）

---

### 故事 5: 清晰的迁移状态查询
**作为** 数据库管理员
**我想要** 查看当前数据库的迁移状态
**以便于** 了解哪些迁移已执行、哪些待执行

**验收标准**:
- 执行 `migro status` 显示表格形式的迁移列表
- 显示迁移文件名、批次号、执行时间、状态（已执行/待执行）

---

### 故事 6: 多数据库支持
**作为** Go 开发者
**我想要** 在不同项目中使用不同的数据库（MySQL/PostgreSQL/SQLite）
**以便于** 工具适配不同的技术栈

**验收标准**:
- 配置文件中指定 `driver: mysql` 即可切换数据库
- 相同的迁移代码在不同数据库上都能正常执行
- 自动处理不同数据库的 SQL 方言差异

---

## 子任务拆解

| 子任务ID | 子任务名称 | 负责角色 | 优先级 | 预期输出 | 依赖 |
|---------|----------|---------|--------|---------|------|
| T2.1 | 系统架构设计 | Architect | P0 | 架构设计文档、模块划分、接口定义 | T1 |
| T3.1 | CLI 命令行框架搭建 | Engineer | P0 | 实现 init/create/up/down/status 命令 | T2.1 |
| T3.2 | 迁移文件 DSL 设计与实现 | Engineer | P0 | 实现 Table/Column/Index/Foreign API | T2.1 |
| T3.3 | MySQL 驱动实现 | Engineer | P0 | MySQL 数据库驱动 | T3.2 |
| T3.4 | PostgreSQL 驱动实现 | Engineer | P0 | PostgreSQL 数据库驱动 | T3.2 |
| T3.5 | SQLite 驱动实现 | Engineer | P0 | SQLite 数据库驱动 | T3.2 |
| T3.6 | 配置管理模块 | Engineer | P0 | YAML 配置解析、环境变量支持 | T3.1 |
| T3.7 | 迁移历史管理 | Engineer | P0 | migrations 表管理、批次记录 | T3.1 |
| T4.1 | 代码质量审查 | Code Reviewer | P0 | 代码审查报告 | T3.* |
| T5.1 | 单元测试 | Tester | P0 | 单元测试用例和报告 | T3.* |
| T5.2 | 集成测试 | Tester | P0 | 集成测试用例和报告 | T3.* |
| T6.1 | API 文档编写 | API Doc | P1 | API 使用文档和示例 | T3.* |
| T7.1 | Git 版本管理 | Git Tool | P1 | 初始化仓库、提交代码 | T3.* |

---

## 技术要求

### 代码风格
- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 使用 `golangci-lint` 进行静态检查

### 依赖管理
- 使用 Go Modules 管理依赖
- 最小化外部依赖，优先使用标准库

### 测试覆盖率
- 单元测试覆盖率 >= 80%
- 核心模块测试覆盖率 >= 90%

### 文档要求
- 每个公开函数必须有注释
- 提供完整的 README 和使用示例
- 提供 API 文档

---

## 风险与挑战

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 不同数据库 SQL 方言差异 | 高 | 设计统一的抽象层，每个驱动实现特定方言 |
| 迁移文件冲突（多人协作） | 中 | 使用时间戳前缀，提供冲突检测机制 |
| 回滚失败导致数据不一致 | 高 | 使用事务包裹迁移，失败自动回滚 |
| 性能问题（大表迁移） | 中 | 提供批量操作 API，支持分批迁移 |

---

## 下一步行动

1. **Architect** 开始系统架构设计（T2.1）
2. **Engineer** 等待架构设计完成后开始开发
3. **Producer** 持续跟进各角色进度，更新本文档

---

## 参考资料

- [Laravel Database Migrations Documentation](https://laravel.com/docs/12.x/migrations)
- [Laravel Migrate: A Complete Database Migration Guide](https://www.moontechnolabs.com/blog/laravel-migrate-practical-guide/)

---

**任务完成标志**: 本文档已完成需求分析和功能设计，等待 Architect 开始架构设计。

**下一步**: 显式调用 `/team` 继续任务。
