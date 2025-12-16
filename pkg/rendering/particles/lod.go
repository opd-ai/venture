// Package particles provides LOD (Level of Detail) system for particle optimization.
// This file implements distance-based particle reduction and viewport culling
// to maintain 60 FPS with 1000+ active particles (Phase 14.3).
package particles

import (
	"math"
)

// LODConfig contains parameters for particle LOD system.
type LODConfig struct {
	// Enabled flag for LOD system
	Enabled bool

	// MaxParticles global limit across all systems
	MaxParticles int

	// DistanceClose threshold for full detail (pixels)
	DistanceClose float64

	// DistanceMid threshold for reduced detail (pixels)
	DistanceMid float64

	// DistanceFar threshold for culling (pixels)
	DistanceFar float64

	// CloseReduction fraction to keep at close distance (1.0 = all)
	CloseReduction float64

	// MidReduction fraction to keep at mid distance (0.5 = half)
	MidReduction float64

	// FarReduction fraction to keep at far distance (0.1 = 10%)
	FarReduction float64
}

// DefaultLODConfig returns sensible default LOD parameters.
func DefaultLODConfig() LODConfig {
	return LODConfig{
		Enabled:        true,
		MaxParticles:   1000,
		DistanceClose:  200.0,
		DistanceMid:    400.0,
		DistanceFar:    800.0,
		CloseReduction: 1.0,  // Keep all particles
		MidReduction:   0.5,  // Keep half
		FarReduction:   0.25, // Keep quarter
	}
}

// ViewportCullingConfig contains parameters for viewport culling.
type ViewportCullingConfig struct {
	// Enabled flag for viewport culling
	Enabled bool

	// Margin in pixels around viewport to start rendering
	Margin float64

	// ViewportX, ViewportY top-left corner of viewport
	ViewportX float64
	ViewportY float64

	// ViewportWidth, ViewportHeight size of viewport
	ViewportWidth  float64
	ViewportHeight float64
}

// DefaultViewportCullingConfig returns sensible default culling parameters.
func DefaultViewportCullingConfig() ViewportCullingConfig {
	return ViewportCullingConfig{
		Enabled:        true,
		Margin:         100.0, // Start rendering 100px before entering viewport
		ViewportX:      0,
		ViewportY:      0,
		ViewportWidth:  800.0,
		ViewportHeight: 600.0,
	}
}

// LODStats tracks LOD system performance.
type LODStats struct {
	// TotalParticles before LOD
	TotalParticles int

	// VisibleParticles after viewport culling
	VisibleParticles int

	// RenderedParticles after distance LOD
	RenderedParticles int

	// CulledByViewport particles outside viewport
	CulledByViewport int

	// CulledByDistance particles too far away
	CulledByDistance int

	// ReducedByLOD particles reduced by distance tiers
	ReducedByLOD int
}

// ApplyViewportCulling filters particles to only those visible in viewport.
// Returns indices of visible particles.
func ApplyViewportCulling(particles []Particle, config ViewportCullingConfig) []int {
	if !config.Enabled {
		// Return all indices if culling disabled
		indices := make([]int, len(particles))
		for i := range indices {
			indices[i] = i
		}
		return indices
	}

	// Calculate viewport bounds with margin
	minX := config.ViewportX - config.Margin
	maxX := config.ViewportX + config.ViewportWidth + config.Margin
	minY := config.ViewportY - config.Margin
	maxY := config.ViewportY + config.ViewportHeight + config.Margin

	// Filter visible particles
	visible := make([]int, 0, len(particles))
	for i, p := range particles {
		if p.X >= minX && p.X <= maxX && p.Y >= minY && p.Y <= maxY {
			visible = append(visible, i)
		}
	}

	return visible
}

// ApplyDistanceLOD filters particles based on distance from camera/player.
// Returns indices of particles to render.
func ApplyDistanceLOD(particles []Particle, visibleIndices []int, cameraX, cameraY float64, config LODConfig) []int {
	if !config.Enabled {
		return visibleIndices
	}

	n := len(visibleIndices)
	if n == 0 {
		return visibleIndices
	}

	// Pre-allocate tier slices with estimated capacity to reduce allocations
	// Typical distribution: 30% close, 30% mid, 30% far, 10% culled
	estimatedPerTier := (n + 2) / 3
	close := make([]int, 0, estimatedPerTier)
	mid := make([]int, 0, estimatedPerTier)
	far := make([]int, 0, estimatedPerTier)

	// Pre-compute squared distances to avoid sqrt in hot loop
	distCloseSq := config.DistanceClose * config.DistanceClose
	distMidSq := config.DistanceMid * config.DistanceMid
	distFarSq := config.DistanceFar * config.DistanceFar

	for _, idx := range visibleIndices {
		p := particles[idx]
		dx := p.X - cameraX
		dy := p.Y - cameraY
		distSq := dx*dx + dy*dy

		if distSq < distCloseSq {
			close = append(close, idx)
		} else if distSq < distMidSq {
			mid = append(mid, idx)
		} else if distSq < distFarSq {
			far = append(far, idx)
		}
		// Beyond DistanceFar: culled entirely
	}

	// Apply reduction factors to each tier
	keepClose := int(float64(len(close)) * config.CloseReduction)
	keepMid := int(float64(len(mid)) * config.MidReduction)
	keepFar := int(float64(len(far)) * config.FarReduction)

	// Pre-allocate result with exact capacity needed
	result := make([]int, 0, keepClose+keepMid+keepFar)
	result = append(result, close[:keepClose]...)
	result = append(result, mid[:keepMid]...)
	result = append(result, far[:keepFar]...)

	return result
}

// EnforceLODLimit enforces global particle limit across all systems.
// Returns indices to render, prioritizing closer particles.
func EnforceLODLimit(particles []Particle, renderIndices []int, cameraX, cameraY float64, maxParticles int) []int {
	if len(renderIndices) <= maxParticles {
		return renderIndices
	}

	// Sort by distance (closer particles have priority)
	type particleDistance struct {
		index    int
		distance float64
	}

	distances := make([]particleDistance, len(renderIndices))
	for i, idx := range renderIndices {
		p := particles[idx]
		dx := p.X - cameraX
		dy := p.Y - cameraY
		dist := math.Sqrt(dx*dx + dy*dy)
		distances[i] = particleDistance{index: idx, distance: dist}
	}

	// Simple bubble sort (fine for small arrays)
	// For production, use sort.Slice with custom comparator
	for i := 0; i < len(distances)-1; i++ {
		for j := 0; j < len(distances)-i-1; j++ {
			if distances[j].distance > distances[j+1].distance {
				distances[j], distances[j+1] = distances[j+1], distances[j]
			}
		}
	}

	// Take first N (closest)
	result := make([]int, maxParticles)
	for i := 0; i < maxParticles; i++ {
		result[i] = distances[i].index
	}

	return result
}

// CalculateLODStats computes LOD statistics for debugging/monitoring.
func CalculateLODStats(totalParticles int, visibleIndices, renderIndices []int) LODStats {
	return LODStats{
		TotalParticles:    totalParticles,
		VisibleParticles:  len(visibleIndices),
		RenderedParticles: len(renderIndices),
		CulledByViewport:  totalParticles - len(visibleIndices),
		ReducedByLOD:      len(visibleIndices) - len(renderIndices),
		CulledByDistance:  totalParticles - len(visibleIndices) - (len(visibleIndices) - len(renderIndices)),
	}
}
