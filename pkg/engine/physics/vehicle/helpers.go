package vehicle

// Helper functions for vehicle physics components.
// These functions operate on component data without being methods,
// maintaining ECS purity where components are pure data structures.

// Suspension helpers

// GetWheelLoad returns the load (force) on a specific wheel.
// Returns 0.0 if wheelIndex is out of bounds.
func GetWheelLoad(s *SuspensionComponent, wheelIndex int) float64 {
	if wheelIndex < 0 || wheelIndex >= len(s.Wheels) {
		return 0.0
	}
	return s.Wheels[wheelIndex].Load
}

// GetWheelCompression returns the compression ratio [0.0, 1.0] for a specific wheel.
// Returns 0.0 if wheelIndex is out of bounds.
func GetWheelCompression(s *SuspensionComponent, wheelIndex int) float64 {
	if wheelIndex < 0 || wheelIndex >= len(s.Wheels) {
		return 0.0
	}
	return s.Wheels[wheelIndex].Compression
}

// IsWheelGrounded checks if a specific wheel is in contact with terrain.
// Returns false if wheelIndex is out of bounds.
func IsWheelGrounded(s *SuspensionComponent, wheelIndex int) bool {
	if wheelIndex < 0 || wheelIndex >= len(s.Wheels) {
		return false
	}
	return s.Wheels[wheelIndex].IsGrounded
}

// GetGroundedWheelCount returns the number of wheels currently touching terrain.
func GetGroundedWheelCount(s *SuspensionComponent) int {
	count := 0
	for i := range s.Wheels {
		if s.Wheels[i].IsGrounded {
			count++
		}
	}
	return count
}

// SetWheelLoad sets the load force on a specific wheel.
// Used by the weight transfer system to modify individual wheel loading.
// No-op if wheelIndex is out of bounds.
func SetWheelLoad(s *SuspensionComponent, wheelIndex int, load float64) {
	if wheelIndex >= 0 && wheelIndex < len(s.Wheels) {
		s.Wheels[wheelIndex].Load = load
	}
}

// Weight transfer helpers

// GetWheelWeights returns the current weight distribution for all wheels.
// Returns array with [FrontLeft, FrontRight, RearLeft, RearRight].
func GetWheelWeights(w *WeightTransferComponent) [4]float64 {
	return [4]float64{
		w.FrontLeftWeight,
		w.FrontRightWeight,
		w.RearLeftWeight,
		w.RearRightWeight,
	}
}

// GetFrontAxleWeight returns the percentage of weight on the front axle.
func GetFrontAxleWeight(w *WeightTransferComponent) float64 {
	return w.FrontLeftWeight + w.FrontRightWeight
}

// GetRearAxleWeight returns the percentage of weight on the rear axle.
func GetRearAxleWeight(w *WeightTransferComponent) float64 {
	return w.RearLeftWeight + w.RearRightWeight
}

// GetLeftSideWeight returns the percentage of weight on the left side.
func GetLeftSideWeight(w *WeightTransferComponent) float64 {
	return w.FrontLeftWeight + w.RearLeftWeight
}

// GetRightSideWeight returns the percentage of weight on the right side.
func GetRightSideWeight(w *WeightTransferComponent) float64 {
	return w.FrontRightWeight + w.RearRightWeight
}

// GetTransferMagnitude returns the magnitude of the last weight transfer.
// Useful for visual feedback (vehicle leaning in turns, nose-diving during braking).
func GetTransferMagnitude(w *WeightTransferComponent) float64 {
	return w.LastTransferMagnitude
}

// ResetWeightTransfer resets the component to static weight distribution.
func ResetWeightTransfer(w *WeightTransferComponent) {
	w.AccelerationX = 0.0
	w.AccelerationY = 0.0
	w.AngularAccel = 0.0
	w.PrevVelocityX = 0.0
	w.PrevVelocityY = 0.0
	w.PrevAngularVel = 0.0
	w.FrontLeftWeight = 0.25
	w.FrontRightWeight = 0.25
	w.RearLeftWeight = 0.25
	w.RearRightWeight = 0.25
	w.LastTransferMagnitude = 0.0
}

// Collision response helpers

// GetDamageMultiplier returns a multiplier based on structural integrity.
// At 100% integrity: 1.0x performance
// At 50% integrity: 0.75x performance
// At 0% integrity: 0.5x performance (minimum)
func GetDamageMultiplier(c *CollisionResponseComponent) float64 {
	return 0.5 + (c.StructuralIntegrity * 0.5)
}

// IsDestroyed checks if structural integrity is depleted.
func IsDestroyed(c *CollisionResponseComponent) bool {
	return c.StructuralIntegrity <= 0.0
}

// RepairVehicle increases structural integrity.
// Caps integrity at 1.0.
func RepairVehicle(c *CollisionResponseComponent, amount float64) {
	c.StructuralIntegrity += amount
	if c.StructuralIntegrity > 1.0 {
		c.StructuralIntegrity = 1.0
	}
}

// ResetCollisionResponse resets collision tracking (used when respawning vehicle).
func ResetCollisionResponse(c *CollisionResponseComponent) {
	c.LastImpactVelocity = 0.0
	c.LastImpactForce = 0.0
	c.LastImpactAngle = 0.0
	c.TotalImpactDamage = 0.0
	c.StructuralIntegrity = 1.0
	c.CollisionCount = 0
}

// ShouldCauseDamage checks if the given impact speed exceeds the damage threshold.
func ShouldCauseDamage(c *CollisionResponseComponent, impactSpeed float64) bool {
	return impactSpeed >= c.DamageThreshold
}

// Terrain deformation helpers

// GetVisibleTracks returns all tracks that should be rendered within the given bounds.
// Uses AABB culling for small track counts (<= 200).
// For large track counts (> 200), uses spatial hash grid for O(k) instead of O(n) culling.
func GetVisibleTracks(t *TerrainDeformationComponent, minX, minY, maxX, maxY float64) []TrackMark {
	// Reuse buffer if available, otherwise allocate
	if t.visibleBuffer == nil {
		t.visibleBuffer = make([]TrackMark, 0, t.MaxTracks)
	}
	// Reset buffer length while preserving capacity
	t.visibleBuffer = t.visibleBuffer[:0]

	// For small track counts, simple AABB culling is optimal
	if len(t.Tracks) <= 200 {
		for i := range t.Tracks {
			track := &t.Tracks[i]

			// Simple AABB culling
			if track.X >= minX && track.X <= maxX && track.Y >= minY && track.Y <= maxY {
				t.visibleBuffer = append(t.visibleBuffer, *track)
			}
		}
		return t.visibleBuffer
	}

	// For large track counts, use spatial partitioning
	// Grid cell size in pixels (larger cells = fewer buckets to check)
	const cellSize = 64.0

	// Calculate grid bounds
	minGridX := int(minX / cellSize)
	maxGridX := int(maxX / cellSize)
	minGridY := int(minY / cellSize)
	maxGridY := int(maxY / cellSize)

	// Build spatial hash on-demand (track positions don't change after creation)
	grid := make(map[[2]int][]int) // grid cell -> track indices
	for i := range t.Tracks {
		track := &t.Tracks[i]
		gx := int(track.X / cellSize)
		gy := int(track.Y / cellSize)
		key := [2]int{gx, gy}
		grid[key] = append(grid[key], i)
	}

	// Query only cells intersecting the viewport
	checked := make(map[int]bool, len(t.Tracks)/4) // dedup tracks at cell boundaries
	for gx := minGridX; gx <= maxGridX; gx++ {
		for gy := minGridY; gy <= maxGridY; gy++ {
			key := [2]int{gx, gy}
			indices := grid[key]
			for _, idx := range indices {
				if checked[idx] {
					continue
				}
				checked[idx] = true

				track := &t.Tracks[idx]
				// Final AABB check (grid cells may partially overlap viewport)
				if track.X >= minX && track.X <= maxX && track.Y >= minY && track.Y <= maxY {
					t.visibleBuffer = append(t.visibleBuffer, *track)
				}
			}
		}
	}

	return t.visibleBuffer
}

// GetTrackAlpha calculates the opacity of a track based on age and fade time.
// Returns value in range [0.0, 1.0] where 0.0 is fully faded, 1.0 is fresh.
func GetTrackAlpha(track *TrackMark) float64 {
	if track.FadeTime <= 0 {
		return 0.0
	}

	// Linear fade
	alpha := 1.0 - (track.Age / track.FadeTime)
	if alpha < 0.0 {
		alpha = 0.0
	}

	return alpha
}

// ClearTracks removes all track marks.
func ClearTracks(t *TerrainDeformationComponent) {
	t.Tracks = t.Tracks[:0]
}
