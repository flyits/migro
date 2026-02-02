# API Doc 文档编写进度

## 任务状态
- **状态**: 已完成
- **负责人**: API Doc (Claude Sonnet 4)
- **开始时间**: 2026-02-02
- **完成时间**: 2026-02-02

---

## 文档编写计划

| 文档 | 内容 | 状态 |
|------|------|------|
| README.md | 主文档（快速入门、CLI、Schema DSL、配置） | ✅ 已完成 |
| ApiDoc.md | 文档编写进度记录 | ✅ 已完成 |

---

## 文档结构

### README.md 包含以下章节：

1. **项目简介** - Migro 概述和特性
2. **快速入门** - 安装、初始化、基本使用
3. **CLI 命令参考** - 所有命令详细说明
4. **Schema DSL API** - Table、Column、Index、ForeignKey API
5. **配置文件** - migro.yaml 格式和环境变量
6. **数据库支持** - MySQL、PostgreSQL、SQLite 差异
7. **迁移文件示例** - 完整的迁移代码示例
8. **最佳实践** - 使用建议和注意事项

---

## 文档来源

- 需求文档: `Ai-team/Producer.md`
- 架构文档: `Ai-team/Architect.md`
- 开发文档: `Ai-team/Engineer.md`
- 源代码: `pkg/schema/*.go`, `internal/migrator/migrator.go`

---

## API 覆盖情况

### Table API ✅
- `NewTable(name)` - 创建表定义
- `ID()` - 自增主键
- `String(name, length)` - VARCHAR 列
- `Text(name)` - TEXT 列
- `Integer(name)` - INT 列
- `BigInteger(name)` - BIGINT 列
- `SmallInteger(name)` - SMALLINT 列
- `TinyInteger(name)` - TINYINT 列
- `Float(name)` - FLOAT 列
- `Double(name)` - DOUBLE 列
- `Decimal(name, precision, scale)` - DECIMAL 列
- `Boolean(name)` - BOOLEAN 列
- `Date(name)` - DATE 列
- `DateTime(name)` - DATETIME 列
- `Timestamp(name)` - TIMESTAMP 列
- `Time(name)` - TIME 列
- `JSON(name)` - JSON 列
- `Binary(name)` - BINARY/BLOB 列
- `UUID(name)` - UUID 列
- `Timestamps()` - created_at + updated_at
- `SoftDeletes()` - deleted_at
- `Index(columns...)` - 普通索引
- `Unique(columns...)` - 唯一索引
- `Primary(columns...)` - 主键
- `Foreign(column)` - 外键
- `DropColumn(name)` - 删除列
- `DropIndex(name)` - 删除索引
- `DropForeign(name)` - 删除外键
- `RenameColumn(from, to)` - 重命名列
- `SetEngine(engine)` - MySQL 引擎
- `SetCharset(charset)` - MySQL 字符集
- `SetCollation(collation)` - MySQL 排序规则

### Column API ✅
- `Nullable()` - 可空
- `Default(value)` - 默认值
- `Unsigned()` - 无符号
- `AutoIncrement()` - 自增
- `Primary()` - 主键
- `Unique()` - 唯一
- `Comment(text)` - 注释
- `PlaceAfter(column)` - 位置（MySQL）

### Index API ✅
- `Named(name)` - 索引名称
- `Unique()` - 唯一索引
- `Primary()` - 主键索引
- `Fulltext()` - 全文索引（MySQL）

### ForeignKey API ✅
- `Named(name)` - 外键名称
- `References(table, column)` - 引用
- `OnDeleteCascade()` - 级联删除
- `OnDeleteSetNull()` - 设为 NULL
- `OnDeleteRestrict()` - 限制删除
- `OnUpdateCascade()` - 级联更新
- `OnUpdateSetNull()` - 设为 NULL
- `OnUpdateRestrict()` - 限制更新

### Executor API ✅
- `CreateTable(ctx, name, fn)` - 创建表
- `AlterTable(ctx, name, fn)` - 修改表
- `DropTable(ctx, name)` - 删除表
- `DropTableIfExists(ctx, name)` - 安全删除表
- `HasTable(ctx, name)` - 检查表存在
- `RenameTable(ctx, from, to)` - 重命名表
- `Raw(ctx, sql)` - 原生 SQL

---

## 输出文件

| 文件 | 路径 | 说明 |
|------|------|------|
| README.md | 项目根目录 | 主文档 |
| ApiDoc.md | Ai-team/ | 进度记录 |

---

**任务完成标志**: API 文档编写完成。

**下一步**: 显式调用 `/team` 继续任务。
