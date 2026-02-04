# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v0.0.3] - 2026-02-04

### Added

- **CLI version command**: Display version, git commit, build date, Go version and OS/Arch information
  - Support version injection via `-ldflags` at build time
  - Show runtime environment details

- **CLI upgrade command**: Check and upgrade to the latest version from GitHub releases
  - `--check` flag to only check for updates without installing
  - Automatic version comparison (with/without `v` prefix)
  - Uses `go install` for seamless upgrades

### Changed

- Improved CLI help documentation

---

## [v0.0.2] - 2026-02-03

### Added

- **ConnectWithDB API**: Allow passing existing `*sql.DB` connections to drivers
  - MySQL, PostgreSQL, SQLite drivers all support `ConnectWithDB(*sql.DB) error`
  - Connection ownership tracking with `ownsConnection` field
  - `Close()` only closes self-owned connections

- **GORM Adapter**: New `pkg/driver/gorm` package for GORM integration
  - `ConnectDriver(drv DBConnector, gormDB *gorm.DB) error` helper function
  - Seamless integration with existing GORM projects

- **ChangeColumn API**: Support for modifying column types in ALTER TABLE
  - `ChangeString`, `ChangeText`, `ChangeInteger`, `ChangeBigInteger`, etc.
  - Full chain method support for column modifiers

### Fixed

- Fixed syntax errors in migration file generation with special characters

---

## [v0.0.1] - 2026-02-02

### Added

- **Initial release** of Migro database migration tool

- **Core Features**:
  - Laravel-style fluent Schema DSL API
  - Multi-database support: MySQL, PostgreSQL, SQLite
  - Transaction-protected migrations (PostgreSQL/SQLite)
  - Environment variable support in config files (`${VAR:default}`)
  - Batch-based migration management
  - Dry-run mode for SQL preview

- **CLI Commands**:
  - `migro init` - Initialize migration environment
  - `migro create` - Create new migration files
  - `migro up` - Run pending migrations
  - `migro down` - Rollback migrations
  - `migro status` - Show migration status
  - `migro reset` - Rollback all migrations
  - `migro refresh` - Reset and re-run all migrations

- **Schema DSL**:
  - Table operations: create, alter, drop, rename
  - Column types: ID, String, Text, Integer, BigInteger, Float, Double, Decimal, Boolean, Date, DateTime, Timestamp, Time, JSON, Binary, UUID
  - Column modifiers: Nullable, Default, Unsigned, AutoIncrement, Primary, Unique, Comment
  - Index support: regular, unique, composite, fulltext (MySQL)
  - Foreign key support with cascade actions
  - Convenience methods: Timestamps(), SoftDeletes()

- **Database Support**:
  - MySQL with InnoDB engine, charset, collation options
  - PostgreSQL with transactional DDL
  - SQLite with file-based storage

- **Documentation**:
  - Comprehensive README with API reference
  - Web-based API documentation site

---

## Version History Summary

| Version | Date | Highlights |
|---------|------|------------|
| v0.0.3 | 2026-02-04 | CLI version/upgrade commands |
| v0.0.2 | 2026-02-03 | ConnectWithDB, GORM adapter, ChangeColumn API |
| v0.0.1 | 2026-02-02 | Initial release |

---

## Upgrade Guide

### From v0.0.2 to v0.0.3

No breaking changes. Simply upgrade:

```bash
migro upgrade
# or
go install github.com/flyits/migro/cmd/migro@latest
```

### From v0.0.1 to v0.0.2

No breaking changes. The new `ConnectWithDB` API is additive.

---

## Build with Version Info

To build with version information embedded:

```bash
go build -ldflags "\
  -X github.com/flyits/migro/internal/cli.Version=v0.0.3 \
  -X github.com/flyits/migro/internal/cli.GitCommit=$(git rev-parse HEAD) \
  -X github.com/flyits/migro/internal/cli.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o migro ./cmd/migro
```

---

[v0.0.3]: https://github.com/flyits/migro/releases/tag/v0.0.3
[v0.0.2]: https://github.com/flyits/migro/releases/tag/v0.0.2
[v0.0.1]: https://github.com/flyits/migro/releases/tag/v0.0.1
