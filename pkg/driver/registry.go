package driver

import (
	"fmt"
	"sync"
)

// Factory is a function that creates a new driver instance
type Factory func() Driver

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]Factory)
)

// Register registers a driver factory with the given name
func Register(name string, factory Factory) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if factory == nil {
		panic("driver: Register factory is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("driver: Register called twice for driver " + name)
	}
	drivers[name] = factory
}

// Get returns a new instance of the driver with the given name
func Get(name string) (Driver, error) {
	driversMu.RLock()
	factory, ok := drivers[name]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("driver: unknown driver %q (forgotten import?)", name)
	}
	return factory(), nil
}

// Drivers returns a list of registered driver names
func Drivers() []string {
	driversMu.RLock()
	defer driversMu.RUnlock()
	list := make([]string, 0, len(drivers))
	for name := range drivers {
		list = append(list, name)
	}
	return list
}
