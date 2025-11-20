package engine

import (
	"sync"
	"time"
)

// GameClock provides time tracking for game simulation.
// Supports both deterministic (simulation-based) and real-time clocks.
type GameClock interface {
	Now() time.Time
	Advance(deltaTime float64)
	Reset(startTime time.Time)
}

// SimulationClock provides deterministic time for reproducible gameplay.
// Time only advances when explicitly advanced via Advance().
type SimulationClock struct {
	mu          sync.RWMutex
	currentTime time.Time
}

// NewSimulationClock creates a new deterministic simulation clock.
// seed parameter is kept for API compatibility but currentTime is initialized to Unix epoch.
func NewSimulationClock(seed int64) *SimulationClock {
	return &SimulationClock{
		currentTime: time.Unix(0, 0), // Start at epoch for determinism
	}
}

// Now returns the current simulation time.
func (c *SimulationClock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentTime
}

// Advance moves the simulation time forward by deltaTime seconds.
func (c *SimulationClock) Advance(deltaTime float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentTime = c.currentTime.Add(time.Duration(deltaTime * float64(time.Second)))
}

// Reset sets the simulation time to the specified start time.
func (c *SimulationClock) Reset(startTime time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentTime = startTime
}

// RealTimeClock provides actual wall-clock time.
// Used for non-deterministic features like UI timestamps.
type RealTimeClock struct{}

// NewRealTimeClock creates a new real-time clock.
func NewRealTimeClock() *RealTimeClock {
	return &RealTimeClock{}
}

// Now returns the current real time.
func (c *RealTimeClock) Now() time.Time {
	return time.Now()
}

// Advance is a no-op for real-time clocks.
func (c *RealTimeClock) Advance(deltaTime float64) {
	// Real-time clock is not affected by simulation advancement
}

// Reset is a no-op for real-time clocks.
func (c *RealTimeClock) Reset(startTime time.Time) {
	// Real-time clock cannot be reset
}
