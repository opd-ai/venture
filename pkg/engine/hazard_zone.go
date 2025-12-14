// Package engine provides hazard zone tracking for environmental hazards.
//
// This file implements HazardZoneTracker which maintains a spatial registry of
// active hazard zones with efficient spatial queries, lifecycle management, and
// proper cleanup to prevent memory leaks.
package engine

import (
	"math"
	"sync"
	"time"
)

// HazardZone represents a tracked environmental hazard zone.
type HazardZone struct {
	// ID is the unique identifier for this zone (matches entity ID)
	ID uint64

	// Position is the center of the hazard zone
	X, Y float64

	// Radius is the affected area in pixels
	Radius float64

	// HazardType identifies the hazard effect
	HazardType HazardType

	// DamagePerSecond for damaging hazards
	DamagePerSecond float64

	// MovementMultiplier for movement-affecting hazards
	MovementMultiplier float64

	// RemainingDuration in seconds (-1 for permanent)
	RemainingDuration float64

	// CreatedAt timestamp for debugging/analytics
	CreatedAt time.Time

	// Intensity is the current strength (1.0 = full, 0.0 = expired)
	// Used for fading effects before removal
	Intensity float64
}

// HazardZoneTracker manages active hazard zones with spatial indexing.
type HazardZoneTracker struct {
	zones        map[uint64]*HazardZone
	mu           sync.RWMutex
	fadeTime     float64  // Duration of fade-out effect in seconds
	maxZones     int      // Maximum zones to prevent unbounded growth
	zoneCount    int      // Current active zone count
	totalZones   uint64   // Lifetime zone creation count
	removeBuffer []uint64 // Reusable buffer for expired zone IDs
}

// NewHazardZoneTracker creates a new hazard zone tracker.
func NewHazardZoneTracker(maxZones int) *HazardZoneTracker {
	if maxZones <= 0 {
		maxZones = 1000 // Default limit
	}

	return &HazardZoneTracker{
		zones:        make(map[uint64]*HazardZone, maxZones),
		fadeTime:     1.0, // 1 second fade-out
		maxZones:     maxZones,
		removeBuffer: make([]uint64, 0, 16), // Pre-allocate for typical usage
	}
}

// AddZone registers a new hazard zone.
// Returns false if max zones reached.
func (t *HazardZoneTracker) AddZone(zone *HazardZone) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check zone limit
	if t.zoneCount >= t.maxZones {
		return false
	}

	// Initialize zone state
	zone.Intensity = 1.0
	zone.CreatedAt = time.Now()

	t.zones[zone.ID] = zone
	t.zoneCount++
	t.totalZones++

	return true
}

// RemoveZone removes a hazard zone by ID.
func (t *HazardZoneTracker) RemoveZone(id uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.zones[id]; exists {
		delete(t.zones, id)
		t.zoneCount--
	}
}

// GetZone retrieves a zone by ID.
func (t *HazardZoneTracker) GetZone(id uint64) (*HazardZone, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	zone, exists := t.zones[id]
	return zone, exists
}

// GetZonesAt returns all hazard zones affecting a position.
// This is the primary spatial query method used by entity damage processing.
func (t *HazardZoneTracker) GetZonesAt(x, y float64) []*HazardZone {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]*HazardZone, 0, 4) // Most positions have 0-4 overlapping zones

	for _, zone := range t.zones {
		// Calculate distance to zone center
		dx := x - zone.X
		dy := y - zone.Y
		distSq := dx*dx + dy*dy
		radiusSq := zone.Radius * zone.Radius

		// Check if position is within zone radius
		if distSq <= radiusSq {
			result = append(result, zone)
		}
	}

	return result
}

// GetZonesInRadius returns all hazard zones overlapping a circular area.
// Used for area queries (e.g., explosions checking for chain reactions).
func (t *HazardZoneTracker) GetZonesInRadius(x, y, radius float64) []*HazardZone {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]*HazardZone, 0, 8)

	for _, zone := range t.zones {
		// Calculate distance between centers
		dx := x - zone.X
		dy := y - zone.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		// Check if zones overlap (distance < sum of radii)
		if dist < (radius + zone.Radius) {
			result = append(result, zone)
		}
	}

	return result
}

// Update advances zone timers, fades expiring zones, and removes expired zones.
// Returns the number of zones removed.
func (t *HazardZoneTracker) Update(deltaTime float64) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	removed := 0
	// Reuse buffer to reduce allocations
	t.removeBuffer = t.removeBuffer[:0]

	for id, zone := range t.zones {
		// Update duration for non-permanent zones
		if zone.RemainingDuration >= 0 {
			zone.RemainingDuration -= deltaTime

			// Handle fade-out phase
			if zone.RemainingDuration < t.fadeTime {
				// Calculate fade intensity (1.0 → 0.0 over fadeTime)
				zone.Intensity = math.Max(0, zone.RemainingDuration/t.fadeTime)
			}

			// Mark expired zones for removal
			if zone.RemainingDuration < 0 {
				t.removeBuffer = append(t.removeBuffer, id)
			}
		}
	}

	// Remove expired zones
	for _, id := range t.removeBuffer {
		delete(t.zones, id)
		t.zoneCount--
		removed++
	}

	return removed
}

// GetActiveZoneCount returns the current number of active zones.
func (t *HazardZoneTracker) GetActiveZoneCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.zoneCount
}

// GetTotalZoneCount returns the lifetime total of zones created.
func (t *HazardZoneTracker) GetTotalZoneCount() uint64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.totalZones
}

// Clear removes all zones (used for world reset/cleanup).
func (t *HazardZoneTracker) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.zones = make(map[uint64]*HazardZone, t.maxZones)
	t.zoneCount = 0
}

// GetAllZones returns a snapshot of all active zones.
// Used for rendering and debugging.
func (t *HazardZoneTracker) GetAllZones() []*HazardZone {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]*HazardZone, 0, len(t.zones))
	for _, zone := range t.zones {
		result = append(result, zone)
	}

	return result
}

// SetFadeTime configures the fade-out duration for expiring zones.
func (t *HazardZoneTracker) SetFadeTime(seconds float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if seconds >= 0 && seconds <= 5.0 {
		t.fadeTime = seconds
	}
}

// GetFadeTime returns the current fade-out duration.
func (t *HazardZoneTracker) GetFadeTime() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.fadeTime
}
