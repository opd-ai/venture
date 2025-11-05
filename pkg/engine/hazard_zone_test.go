package engine

import (
	"testing"
	"time"
)

func TestNewHazardZoneTracker(t *testing.T) {
	tests := []struct {
		name     string
		maxZones int
		wantMax  int
	}{
		{"valid max zones", 500, 500},
		{"zero max zones (uses default)", 0, 1000},
		{"negative max zones (uses default)", -10, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewHazardZoneTracker(tt.maxZones)
			if tracker == nil {
				t.Fatal("NewHazardZoneTracker returned nil")
			}
			if tracker.maxZones != tt.wantMax {
				t.Errorf("maxZones = %d, want %d", tracker.maxZones, tt.wantMax)
			}
			if tracker.GetActiveZoneCount() != 0 {
				t.Error("new tracker should have 0 active zones")
			}
		})
	}
}

func TestHazardZoneTracker_AddRemoveZone(t *testing.T) {
	tracker := NewHazardZoneTracker(100)

	// Create test zone
	zone := &HazardZone{
		ID:                 1,
		X:                  100,
		Y:                  200,
		Radius:             50,
		HazardType:         HazardPoison,
		DamagePerSecond:    5.0,
		RemainingDuration:  10.0,
		MovementMultiplier: 1.0,
	}

	// Add zone
	if !tracker.AddZone(zone) {
		t.Fatal("AddZone failed")
	}
	if tracker.GetActiveZoneCount() != 1 {
		t.Errorf("active count = %d, want 1", tracker.GetActiveZoneCount())
	}

	// Verify zone was added
	retrieved, exists := tracker.GetZone(1)
	if !exists {
		t.Fatal("zone not found after adding")
	}
	if retrieved.ID != zone.ID {
		t.Error("retrieved zone has wrong ID")
	}
	if retrieved.Intensity != 1.0 {
		t.Errorf("intensity = %f, want 1.0", retrieved.Intensity)
	}

	// Remove zone
	tracker.RemoveZone(1)
	if tracker.GetActiveZoneCount() != 0 {
		t.Errorf("active count = %d, want 0 after removal", tracker.GetActiveZoneCount())
	}

	// Verify zone was removed
	_, exists = tracker.GetZone(1)
	if exists {
		t.Error("zone still exists after removal")
	}
}

func TestHazardZoneTracker_MaxZonesLimit(t *testing.T) {
	maxZones := 5
	tracker := NewHazardZoneTracker(maxZones)

	// Add zones up to limit
	for i := 1; i <= maxZones; i++ {
		zone := &HazardZone{
			ID:                uint64(i),
			X:                 float64(i * 100),
			Y:                 float64(i * 100),
			Radius:            50,
			HazardType:        HazardPoison,
			RemainingDuration: 10.0,
		}
		if !tracker.AddZone(zone) {
			t.Fatalf("AddZone failed for zone %d (within limit)", i)
		}
	}

	if tracker.GetActiveZoneCount() != maxZones {
		t.Errorf("active count = %d, want %d", tracker.GetActiveZoneCount(), maxZones)
	}

	// Try to add beyond limit
	extraZone := &HazardZone{
		ID:                uint64(maxZones + 1),
		X:                 1000,
		Y:                 1000,
		Radius:            50,
		HazardType:        HazardOil,
		RemainingDuration: 10.0,
	}
	if tracker.AddZone(extraZone) {
		t.Error("AddZone should fail when max zones reached")
	}

	if tracker.GetActiveZoneCount() != maxZones {
		t.Error("zone count changed after failed add")
	}
}

func TestHazardZoneTracker_GetZonesAt(t *testing.T) {
	tracker := NewHazardZoneTracker(100)

	// Add zones at different positions
	zones := []*HazardZone{
		{ID: 1, X: 100, Y: 100, Radius: 50, HazardType: HazardPoison, RemainingDuration: 10},
		{ID: 2, X: 300, Y: 100, Radius: 50, HazardType: HazardOil, RemainingDuration: 10},
		{ID: 3, X: 100, Y: 300, Radius: 50, HazardType: HazardWater, RemainingDuration: 10},
	}

	for _, zone := range zones {
		tracker.AddZone(zone)
	}

	tests := []struct {
		name      string
		x, y      float64
		wantCount int
		wantIDs   []uint64
	}{
		{"center of zone 1", 100, 100, 1, []uint64{1}},
		{"edge of zone 1", 150, 100, 1, []uint64{1}},
		{"outside all zones", 500, 500, 0, nil},
		{"between zones", 200, 100, 0, nil},
		{"center of zone 2", 300, 100, 1, []uint64{2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tracker.GetZonesAt(tt.x, tt.y)
			if len(result) != tt.wantCount {
				t.Errorf("GetZonesAt(%f, %f) returned %d zones, want %d",
					tt.x, tt.y, len(result), tt.wantCount)
			}

			// Verify specific zone IDs if expected
			if tt.wantIDs != nil {
				foundIDs := make(map[uint64]bool)
				for _, zone := range result {
					foundIDs[zone.ID] = true
				}
				for _, wantID := range tt.wantIDs {
					if !foundIDs[wantID] {
						t.Errorf("expected zone ID %d not found", wantID)
					}
				}
			}
		})
	}
}

func TestHazardZoneTracker_GetZonesInRadius(t *testing.T) {
	tracker := NewHazardZoneTracker(100)

	// Add zones
	zones := []*HazardZone{
		{ID: 1, X: 100, Y: 100, Radius: 50, HazardType: HazardPoison, RemainingDuration: 10},
		{ID: 2, X: 200, Y: 100, Radius: 50, HazardType: HazardOil, RemainingDuration: 10},
		{ID: 3, X: 500, Y: 500, Radius: 50, HazardType: HazardWater, RemainingDuration: 10},
	}

	for _, zone := range zones {
		tracker.AddZone(zone)
	}

	tests := []struct {
		name      string
		x, y      float64
		radius    float64
		wantCount int
	}{
		{"overlaps zone 1 only", 80, 100, 50, 1},
		{"overlaps zones 1 and 2", 150, 100, 80, 2},
		{"large radius overlaps all", 250, 250, 500, 3},
		{"overlaps none", 1000, 1000, 50, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tracker.GetZonesInRadius(tt.x, tt.y, tt.radius)
			if len(result) != tt.wantCount {
				t.Errorf("GetZonesInRadius returned %d zones, want %d",
					len(result), tt.wantCount)
			}
		})
	}
}

func TestHazardZoneTracker_Update(t *testing.T) {
	tracker := NewHazardZoneTracker(100)

	// Add zone with 2 second duration
	zone := &HazardZone{
		ID:                1,
		X:                 100,
		Y:                 100,
		Radius:            50,
		HazardType:        HazardPoison,
		RemainingDuration: 2.0,
	}
	tracker.AddZone(zone)

	// Update 1 second
	removed := tracker.Update(1.0)
	if removed != 0 {
		t.Error("zone should not be removed yet")
	}
	if tracker.GetActiveZoneCount() != 1 {
		t.Error("zone count changed unexpectedly")
	}

	// Verify duration decreased
	updated, exists := tracker.GetZone(1)
	if !exists {
		t.Fatal("zone disappeared after update")
	}
	if updated.RemainingDuration >= 2.0 {
		t.Error("duration not decremented")
	}

	// Update past expiration
	removed = tracker.Update(2.0)
	if removed != 1 {
		t.Errorf("removed %d zones, want 1", removed)
	}
	if tracker.GetActiveZoneCount() != 0 {
		t.Error("expired zone not removed")
	}
}

func TestHazardZoneTracker_UpdateFadeIntensity(t *testing.T) {
	tracker := NewHazardZoneTracker(100)
	tracker.SetFadeTime(1.0)

	// Add zone with 1.5 second duration (will fade in last 1 second)
	zone := &HazardZone{
		ID:                1,
		X:                 100,
		Y:                 100,
		Radius:            50,
		HazardType:        HazardPoison,
		RemainingDuration: 1.5,
	}
	tracker.AddZone(zone)

	// Update 0.6 seconds (remaining: 0.9s, should start fading)
	tracker.Update(0.6)
	updated, _ := tracker.GetZone(1)
	if updated.Intensity >= 1.0 {
		t.Error("intensity should be fading")
	}
	if updated.Intensity <= 0.0 {
		t.Error("intensity should not be zero yet")
	}

	// Update to near expiration
	tracker.Update(0.8)
	updated, _ = tracker.GetZone(1)
	if updated.Intensity >= 0.5 {
		t.Errorf("intensity = %f, should be very low", updated.Intensity)
	}
}

func TestHazardZoneTracker_PermanentZones(t *testing.T) {
	tracker := NewHazardZoneTracker(100)

	// Add permanent zone (duration -1)
	zone := &HazardZone{
		ID:                1,
		X:                 100,
		Y:                 100,
		Radius:            50,
		HazardType:        HazardOil,
		RemainingDuration: -1,
	}
	tracker.AddZone(zone)

	// Update many times
	for i := 0; i < 10; i++ {
		removed := tracker.Update(1.0)
		if removed != 0 {
			t.Fatal("permanent zone was removed")
		}
	}

	// Verify zone still exists
	if tracker.GetActiveZoneCount() != 1 {
		t.Error("permanent zone disappeared")
	}

	updated, _ := tracker.GetZone(1)
	if updated.RemainingDuration != -1 {
		t.Error("permanent zone duration changed")
	}
}

func TestHazardZoneTracker_Clear(t *testing.T) {
	tracker := NewHazardZoneTracker(100)

	// Add multiple zones
	for i := 1; i <= 10; i++ {
		zone := &HazardZone{
			ID:                uint64(i),
			X:                 float64(i * 100),
			Y:                 float64(i * 100),
			Radius:            50,
			HazardType:        HazardPoison,
			RemainingDuration: 10.0,
		}
		tracker.AddZone(zone)
	}

	if tracker.GetActiveZoneCount() != 10 {
		t.Fatal("zones not added correctly")
	}

	// Clear all zones
	tracker.Clear()

	if tracker.GetActiveZoneCount() != 0 {
		t.Error("zones not cleared")
	}

	// Verify no zones remain
	zones := tracker.GetAllZones()
	if len(zones) != 0 {
		t.Error("GetAllZones returned zones after clear")
	}
}

func TestHazardZoneTracker_GetAllZones(t *testing.T) {
	tracker := NewHazardZoneTracker(100)

	// Add zones
	count := 5
	for i := 1; i <= count; i++ {
		zone := &HazardZone{
			ID:                uint64(i),
			X:                 float64(i * 100),
			Y:                 float64(i * 100),
			Radius:            50,
			HazardType:        HazardPoison,
			RemainingDuration: 10.0,
		}
		tracker.AddZone(zone)
	}

	zones := tracker.GetAllZones()
	if len(zones) != count {
		t.Errorf("GetAllZones returned %d zones, want %d", len(zones), count)
	}

	// Verify all zones present
	foundIDs := make(map[uint64]bool)
	for _, zone := range zones {
		foundIDs[zone.ID] = true
	}

	for i := 1; i <= count; i++ {
		if !foundIDs[uint64(i)] {
			t.Errorf("zone ID %d not found in GetAllZones result", i)
		}
	}
}

func TestHazardZoneTracker_SetGetFadeTime(t *testing.T) {
	tracker := NewHazardZoneTracker(100)

	// Test default fade time
	if tracker.GetFadeTime() != 1.0 {
		t.Errorf("default fade time = %f, want 1.0", tracker.GetFadeTime())
	}

	tests := []struct {
		name     string
		fadeTime float64
		wantSet  float64
	}{
		{"valid fade time", 2.5, 2.5},
		{"zero fade time", 0.0, 0.0},
		{"max fade time", 5.0, 5.0},
		{"negative fade time (ignored)", -1.0, 5.0}, // Should keep previous value
		{"too large (ignored)", 10.0, 5.0},          // Should keep previous value
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker.SetFadeTime(tt.fadeTime)
			got := tracker.GetFadeTime()
			if got != tt.wantSet {
				t.Errorf("fade time = %f, want %f", got, tt.wantSet)
			}
		})
	}
}

func TestHazardZoneTracker_ConcurrentAccess(t *testing.T) {
	tracker := NewHazardZoneTracker(1000)

	// Concurrent add operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(start int) {
			for j := 0; j < 10; j++ {
				zone := &HazardZone{
					ID:                uint64(start*10 + j),
					X:                 float64(j * 50),
					Y:                 float64(j * 50),
					Radius:            25,
					HazardType:        HazardPoison,
					RemainingDuration: 5.0,
				}
				tracker.AddZone(zone)
			}
			done <- true
		}(i)
	}

	// Wait for all adds
	for i := 0; i < 10; i++ {
		<-done
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			tracker.GetZonesAt(100, 100)
			tracker.GetActiveZoneCount()
			done <- true
		}()
	}

	// Wait for all reads
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic or corrupt data
	count := tracker.GetActiveZoneCount()
	if count <= 0 || count > 1000 {
		t.Errorf("zone count = %d, expected reasonable value", count)
	}
}

func TestHazardZoneTracker_TotalZoneCount(t *testing.T) {
	tracker := NewHazardZoneTracker(100)

	// Add and remove zones
	for i := 1; i <= 5; i++ {
		zone := &HazardZone{
			ID:                uint64(i),
			X:                 100,
			Y:                 100,
			Radius:            50,
			HazardType:        HazardPoison,
			RemainingDuration: 0.1, // Will expire quickly
		}
		tracker.AddZone(zone)
	}

	if tracker.GetTotalZoneCount() != 5 {
		t.Errorf("total zones = %d, want 5", tracker.GetTotalZoneCount())
	}

	// Expire zones
	tracker.Update(1.0)

	if tracker.GetActiveZoneCount() != 0 {
		t.Error("zones not expired")
	}

	// Total should remain 5
	if tracker.GetTotalZoneCount() != 5 {
		t.Errorf("total zones = %d, want 5 (lifetime count)", tracker.GetTotalZoneCount())
	}

	// Add more zones
	for i := 6; i <= 8; i++ {
		zone := &HazardZone{
			ID:                uint64(i),
			X:                 100,
			Y:                 100,
			Radius:            50,
			HazardType:        HazardOil,
			RemainingDuration: 10.0,
		}
		tracker.AddZone(zone)
	}

	if tracker.GetTotalZoneCount() != 8 {
		t.Errorf("total zones = %d, want 8", tracker.GetTotalZoneCount())
	}
}

func TestHazardZone_CreatedAtTimestamp(t *testing.T) {
	tracker := NewHazardZoneTracker(100)

	before := time.Now()
	zone := &HazardZone{
		ID:                1,
		X:                 100,
		Y:                 100,
		Radius:            50,
		HazardType:        HazardPoison,
		RemainingDuration: 10.0,
	}
	tracker.AddZone(zone)
	after := time.Now()

	retrieved, exists := tracker.GetZone(1)
	if !exists {
		t.Fatal("zone not found")
	}

	// Verify CreatedAt timestamp was set
	if retrieved.CreatedAt.IsZero() {
		t.Error("CreatedAt timestamp not set")
	}

	// Verify timestamp is reasonable
	if retrieved.CreatedAt.Before(before) || retrieved.CreatedAt.After(after) {
		t.Error("CreatedAt timestamp out of expected range")
	}
}
