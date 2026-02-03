// Package gorm provides GORM integration for migro database drivers.
package gorm

import (
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

// DBConnector is the interface that drivers must implement to support ConnectWithDB.
type DBConnector interface {
	ConnectWithDB(db *sql.DB) error
}

// ConnectDriver connects a migro driver using an existing GORM database instance.
// The driver will use the underlying *sql.DB from GORM but will not close it
// when the driver's Close() method is called.
//
// Example usage:
//
//	gormDB, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
//	drv := mysql.NewDriver()
//	if err := migrogorm.ConnectDriver(drv, gormDB); err != nil {
//	    log.Fatal(err)
//	}
//	migrator := migrator.NewMigrator(drv, "./migrations", "migrations")
func ConnectDriver(drv DBConnector, gormDB *gorm.DB) error {
	if drv == nil {
		return fmt.Errorf("gorm: driver is nil")
	}
	if gormDB == nil {
		return fmt.Errorf("gorm: gorm.DB is nil")
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("gorm: failed to get underlying *sql.DB: %w", err)
	}

	return drv.ConnectWithDB(sqlDB)
}
