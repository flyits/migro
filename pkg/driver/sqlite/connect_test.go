//go:build cgo

package sqlite

import (
	"database/sql"
	"testing"

	"github.com/flyits/migro/pkg/driver"
	_ "github.com/mattn/go-sqlite3"
)

// TestConnectWithDB_Success tests that ConnectWithDB works with a valid connection
func TestConnectWithDB_Success(t *testing.T) {
	// Create an external database connection
	externalDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create external db: %v", err)
	}
	defer externalDB.Close()

	// Verify the external connection works
	if err := externalDB.Ping(); err != nil {
		t.Fatalf("external db ping failed: %v", err)
	}

	// Create driver and connect with external DB
	drv := NewDriver()
	err = drv.ConnectWithDB(externalDB)
	if err != nil {
		t.Fatalf("ConnectWithDB failed: %v", err)
	}

	// Verify the driver is using the external connection
	if drv.DB() != externalDB {
		t.Error("driver should use the external connection")
	}

	// Verify the connection is usable
	var result int
	err = drv.DB().QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if result != 1 {
		t.Errorf("expected 1, got %d", result)
	}
}

// TestConnectWithDB_NilConnection tests that ConnectWithDB returns error for nil connection
func TestConnectWithDB_NilConnection(t *testing.T) {
	drv := NewDriver()
	err := drv.ConnectWithDB(nil)
	if err == nil {
		t.Error("ConnectWithDB should return error for nil connection")
	}
}

// TestConnectWithDB_ClosedConnection tests that ConnectWithDB returns error for closed connection
func TestConnectWithDB_ClosedConnection(t *testing.T) {
	// Create and close a connection
	closedDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	closedDB.Close() // Close it immediately

	drv := NewDriver()
	err = drv.ConnectWithDB(closedDB)
	if err == nil {
		t.Error("ConnectWithDB should return error for closed connection")
	}
}

// TestClose_DoesNotCloseExternalConnection tests that Close() does not close external connection
func TestClose_DoesNotCloseExternalConnection(t *testing.T) {
	// Create an external database connection
	externalDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create external db: %v", err)
	}
	defer externalDB.Close()

	// Create driver and connect with external DB
	drv := NewDriver()
	err = drv.ConnectWithDB(externalDB)
	if err != nil {
		t.Fatalf("ConnectWithDB failed: %v", err)
	}

	// Close the driver
	err = drv.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify the external connection is still usable
	err = externalDB.Ping()
	if err != nil {
		t.Error("external connection should still be usable after driver.Close()")
	}

	// Verify we can still query
	var result int
	err = externalDB.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Errorf("external connection should still work: %v", err)
	}
}

// TestClose_ClosesOwnedConnection tests that Close() closes connection created by Connect()
func TestClose_ClosesOwnedConnection(t *testing.T) {
	drv := NewDriver()

	// Use Connect() to create an owned connection
	err := drv.Connect(&driver.Config{Database: ":memory:"})
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Get reference to the internal db
	internalDB := drv.DB()

	// Close the driver
	err = drv.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify the internal connection is closed
	err = internalDB.Ping()
	if err == nil {
		t.Error("internal connection should be closed after driver.Close()")
	}
}

// TestConnect_SetsOwnsConnectionTrue tests that Connect() sets ownsConnection to true
// by verifying Close() actually closes the connection
func TestConnect_SetsOwnsConnectionTrue(t *testing.T) {
	drv := NewDriver()

	err := drv.Connect(&driver.Config{Database: ":memory:"})
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Get reference to the internal db before closing
	internalDB := drv.DB()

	// Close should close the connection since it's owned
	err = drv.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify the connection is actually closed
	err = internalDB.Ping()
	if err == nil {
		t.Error("connection should be closed after driver.Close() when using Connect()")
	}
}

// TestConnectWithDB_SetsOwnsConnectionFalse tests that ConnectWithDB sets ownsConnection to false
// by verifying Close() does NOT close the connection
func TestConnectWithDB_SetsOwnsConnectionFalse(t *testing.T) {
	// Create an external database connection
	externalDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create external db: %v", err)
	}
	defer externalDB.Close()

	drv := NewDriver()
	err = drv.ConnectWithDB(externalDB)
	if err != nil {
		t.Fatalf("ConnectWithDB failed: %v", err)
	}

	// Close the driver
	err = drv.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify the connection is NOT closed
	err = externalDB.Ping()
	if err != nil {
		t.Error("connection should NOT be closed after driver.Close() when using ConnectWithDB()")
	}
}
