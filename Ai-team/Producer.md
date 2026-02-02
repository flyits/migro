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

---

# T8: API 文档网页需求分析

## 任务状态
- **状态**: ✅ 已完成
- **负责人**: Producer (Claude Sonnet 4)
- **开始时间**: 2026-02-02
- **完成时间**: 2026-02-02

---

## 1. 目标用户分析

### 1.1 主要用户群体

| 用户类型 | 特征 | 核心需求 | 使用场景 |
|---------|------|---------|---------|
| **Go 初学者** | 刚接触 Go 和数据库迁移 | 快速入门指南、完整示例 | 学习如何使用 Migro |
| **Go 开发者** | 有经验的开发者 | API 参考、最佳实践 | 日常开发中查阅 API |
| **DevOps 工程师** | 负责部署和运维 | CLI 命令参考、配置说明 | 生产环境迁移操作 |
| **技术评估者** | 评估是否采用 Migro | 特性概览、与竞品对比 | 技术选型决策 |

### 1.2 用户使用场景

1. **快速入门**: 新用户需要在 5 分钟内了解 Migro 的核心功能
2. **API 查阅**: 开发者需要快速找到特定 API 的用法和参数
3. **问题排查**: 遇到问题时需要查找错误处理和注意事项
4. **最佳实践**: 了解生产环境的推荐用法

---

## 2. 文档内容结构设计

### 2.1 一级导航结构

```
首页 (Home)
├── 快速入门 (Getting Started)
├── CLI 命令参考 (CLI Reference)
├── Schema DSL API (Schema API)
├── 配置指南 (Configuration)
├── 数据库支持 (Database Support)
├── 最佳实践 (Best Practices)
└── 示例代码 (Examples)
```

### 2.2 详细内容规划

#### 2.2.1 首页 (Home)
- 项目简介和核心特性
- 安装方式
- 快速示例代码
- 导航入口

#### 2.2.2 快速入门 (Getting Started)
- 安装指南
- 初始化项目
- 创建第一个迁移
- 执行迁移
- 查看状态

#### 2.2.3 CLI 命令参考 (CLI Reference)
| 命令 | 说明 |
|------|------|
| `migro init` | 初始化迁移环境 |
| `migro create` | 创建迁移文件 |
| `migro up` | 执行迁移 |
| `migro down` | 回滚迁移 |
| `migro status` | 查看状态 |
| `migro reset` | 重置所有迁移 |
| `migro refresh` | 刷新迁移 |

#### 2.2.4 Schema DSL API
- **Table 构建器**
  - CreateTable / AlterTable / DropTable
  - RenameTable / HasTable
- **列类型**
  - 基础类型: ID, String, Text, Integer, BigInteger...
  - 时间类型: Date, DateTime, Timestamp, Time
  - 特殊类型: JSON, Binary, UUID, Boolean
  - 便捷方法: Timestamps, SoftDeletes
- **列修饰符**
  - Nullable, Default, Unsigned, AutoIncrement
  - Primary, Unique, Comment, PlaceAfter
- **索引操作**
  - Index, Unique, Primary, Fulltext
  - Named, DropIndex
- **外键操作**
  - Foreign, References, Named
  - OnDeleteCascade, OnUpdateCascade...
  - DropForeign
- **ALTER TABLE 操作**
  - 添加/删除/重命名列
  - Change* 系列方法

#### 2.2.5 配置指南 (Configuration)
- migro.yaml 配置文件格式
- 环境变量支持
- MySQL/PostgreSQL/SQLite 配置示例

#### 2.2.6 数据库支持 (Database Support)
- 支持的数据库版本
- 类型映射表
- 事务 DDL 支持
- ALTER TABLE 限制
- 数据库差异说明

#### 2.2.7 最佳实践 (Best Practices)
- 迁移命名规范
- Down 方法编写
- 大表迁移策略
- 环境变量管理
- 生产环境注意事项

#### 2.2.8 示例代码 (Examples)
- 创建用户表
- 创建文章表（带外键）
- 修改表结构
- 完整迁移示例

---

## 3. 页面功能需求

### 3.1 核心功能

| 功能 | 优先级 | 说明 |
|------|--------|------|
| **侧边栏导航** | P0 | 固定侧边栏，支持多级展开 |
| **代码高亮** | P0 | Go 和 YAML 代码语法高亮 |
| **代码复制** | P0 | 一键复制代码块 |
| **响应式布局** | P0 | 适配桌面和移动设备 |
| **锚点导航** | P1 | 页内标题锚点跳转 |
| **搜索功能** | P2 | 全文搜索（可选） |
| **暗色模式** | P2 | 支持深色主题（可选） |

### 3.2 页面组件

| 组件 | 说明 |
|------|------|
| **Header** | 项目名称、GitHub 链接 |
| **Sidebar** | 导航菜单，支持折叠 |
| **Content** | 主内容区域 |
| **Code Block** | 代码展示，带复制按钮 |
| **Table** | API 参数表格 |
| **Callout** | 提示/警告/注意框 |
| **Footer** | 版权信息、许可证 |

---

## 4. 用户体验需求

### 4.1 设计原则

1. **简洁清晰**: 页面布局简洁，信息层次分明
2. **快速定位**: 用户能在 3 次点击内找到目标内容
3. **代码优先**: 代码示例突出显示，易于阅读和复制
4. **一致性**: 页面风格、组件样式保持一致

### 4.2 视觉设计要求

| 要素 | 要求 |
|------|------|
| **配色** | 专业技术文档风格，蓝/灰为主色调 |
| **字体** | 正文使用无衬线字体，代码使用等宽字体 |
| **间距** | 充足的留白，提高可读性 |
| **代码块** | 深色背景，语法高亮 |

### 4.3 交互设计要求

| 交互 | 要求 |
|------|------|
| **导航** | 当前页面高亮，支持键盘导航 |
| **滚动** | 平滑滚动，返回顶部按钮 |
| **代码复制** | 点击后显示"已复制"反馈 |
| **响应式** | 移动端侧边栏可折叠 |

---

## 5. 技术约束

### 5.1 目录结构

```
/doc/web/
├── index.html          # 首页
├── getting-started.html # 快速入门
├── cli-reference.html   # CLI 命令参考
├── schema-api.html      # Schema DSL API
├── configuration.html   # 配置指南
├── database-support.html # 数据库支持
├── best-practices.html  # 最佳实践
├── examples.html        # 示例代码
├── css/
│   └── style.css       # 样式文件
├── js/
│   └── main.js         # 交互脚本
└── assets/
    └── logo.svg        # Logo 图片
```

### 5.2 技术选型建议

| 方面 | 建议 |
|------|------|
| **HTML** | 语义化 HTML5 |
| **CSS** | 原生 CSS 或轻量级框架 |
| **JS** | 原生 JavaScript，无框架依赖 |
| **代码高亮** | Prism.js 或 Highlight.js |
| **构建** | 纯静态文件，无需构建工具 |

### 5.3 兼容性要求

- 支持 Chrome、Firefox、Safari、Edge 最新版本
- 支持移动端浏览器
- 无需支持 IE

---

## 6. 子任务拆解

| 子任务ID | 子任务名称 | 负责角色 | 优先级 | 预期输出 |
|---------|----------|---------|--------|---------|
| T9 | 文档网页架构设计 | Architect | P0 | 技术架构、组件设计、目录结构 |
| T10.1 | 基础框架搭建 | Engineer | P0 | HTML 模板、CSS 样式、JS 交互 |
| T10.2 | 首页开发 | Engineer | P0 | index.html |
| T10.3 | 快速入门页面 | Engineer | P0 | getting-started.html |
| T10.4 | CLI 命令参考页面 | Engineer | P0 | cli-reference.html |
| T10.5 | Schema API 页面 | Engineer | P0 | schema-api.html |
| T10.6 | 配置指南页面 | Engineer | P1 | configuration.html |
| T10.7 | 数据库支持页面 | Engineer | P1 | database-support.html |
| T10.8 | 最佳实践页面 | Engineer | P1 | best-practices.html |
| T10.9 | 示例代码页面 | Engineer | P1 | examples.html |
| T11 | 代码质量审查 | Code Reviewer | P0 | 审查报告 |
| T12 | 功能测试验证 | Tester | P0 | 测试报告 |
| T13 | Git 版本管理 | Git Tool | P0 | 提交代码 |

---

## 7. 用户故事

### 故事 1: 快速了解 Migro
**作为** 技术评估者
**我想要** 在首页快速了解 Migro 的核心特性和优势
**以便于** 决定是否在项目中采用 Migro

**验收标准**:
- 首页展示项目简介和核心特性
- 提供快速示例代码
- 有清晰的导航入口

### 故事 2: 查阅 API 文档
**作为** Go 开发者
**我想要** 快速找到特定 API 的用法和参数说明
**以便于** 在开发中正确使用 Migro API

**验收标准**:
- 侧边栏导航清晰，支持多级展开
- API 文档包含参数说明和代码示例
- 代码块支持一键复制

### 故事 3: 学习最佳实践
**作为** DevOps 工程师
**我想要** 了解生产环境的最佳实践和注意事项
**以便于** 安全地在生产环境执行迁移

**验收标准**:
- 最佳实践页面包含生产环境建议
- 提供大表迁移策略
- 说明环境变量管理方式

---

## 8. 风险与挑战

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 内容过多导致页面臃肿 | 中 | 合理拆分页面，使用锚点导航 |
| 代码示例与实际 API 不一致 | 高 | 从 README.md 提取内容，保持一致 |
| 移动端体验不佳 | 中 | 响应式设计，移动端优先测试 |

---

## 9. 验收标准

1. ✅ 所有页面内容完整，与 README.md 保持一致
2. ✅ 代码高亮正常，支持一键复制
3. ✅ 响应式布局，移动端可用
4. ✅ 侧边栏导航正常，支持多级展开
5. ✅ 页面加载速度 < 2 秒
6. ✅ 无 JavaScript 错误
7. ✅ 兼容主流浏览器

---

**任务完成标志**: T8 需求分析已完成，等待 Architect 开始架构设计 (T9)。

**下一步**: 显式调用 `/team` 继续任务流程。
