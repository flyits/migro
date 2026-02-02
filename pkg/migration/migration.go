package migration

import (
	"github.com/migro/migro/internal/migrator"
)

// Migration defines the interface for database migrations
type Migration interface {
	// Name returns the unique name of the migration
	Name() string

	// Up runs the migration
	Up(e *migrator.Executor) error

	// Down reverses the migration
	Down(e *migrator.Executor) error
}
