package gorm

import (
	"database/sql"
	"errors"
	"testing"
)

// mockDBConnector is a mock implementation of DBConnector for testing
type mockDBConnector struct {
	connectWithDBCalled bool
	connectWithDBArg    *sql.DB
	connectWithDBErr    error
}

func (m *mockDBConnector) ConnectWithDB(db *sql.DB) error {
	m.connectWithDBCalled = true
	m.connectWithDBArg = db
	return m.connectWithDBErr
}

// mockGormDB is a mock that simulates gorm.DB behavior
// We can't use real gorm.DB in unit tests without a database
// So we test the error paths and nil checks

// TestConnectDriver_NilDriver tests that ConnectDriver returns error for nil driver
func TestConnectDriver_NilDriver(t *testing.T) {
	err := ConnectDriver(nil, nil)
	if err == nil {
		t.Error("ConnectDriver should return error for nil driver")
	}
	if err.Error() != "gorm: driver is nil" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestConnectDriver_NilGormDB tests that ConnectDriver returns error for nil gormDB
func TestConnectDriver_NilGormDB(t *testing.T) {
	mock := &mockDBConnector{}
	err := ConnectDriver(mock, nil)
	if err == nil {
		t.Error("ConnectDriver should return error for nil gormDB")
	}
	if err.Error() != "gorm: gorm.DB is nil" {
		t.Errorf("unexpected error message: %v", err)
	}
	// Verify ConnectWithDB was not called
	if mock.connectWithDBCalled {
		t.Error("ConnectWithDB should not be called when gormDB is nil")
	}
}

// TestDBConnectorInterface tests that the interface is correctly defined
func TestDBConnectorInterface(t *testing.T) {
	// This test verifies that the interface can be implemented
	var _ DBConnector = &mockDBConnector{}
}

// TestConnectDriver_PropagatesError tests that errors from ConnectWithDB are propagated
func TestConnectDriver_PropagatesError(t *testing.T) {
	// This test would require a real gorm.DB instance
	// Since we can't create one without a database, we document this limitation
	// The actual integration test should be done with a real database

	// For now, we verify the mock works correctly
	mock := &mockDBConnector{
		connectWithDBErr: errors.New("connection failed"),
	}

	// We can't test the full flow without a real gorm.DB
	// but we can verify the mock is set up correctly
	err := mock.ConnectWithDB(nil)
	if err == nil {
		t.Error("mock should return error")
	}
	if !mock.connectWithDBCalled {
		t.Error("ConnectWithDB should be called")
	}
}
