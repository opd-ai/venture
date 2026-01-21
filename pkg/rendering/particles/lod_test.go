// Package particles provides LOD system tests.
package particles

import (
	"testing"
)

func TestDefaultLODConfig(t *testing.T) {
	config := DefaultLODConfig()

	if !config.Enabled {
		t.Error("Default LOD should be enabled")
	}
	if config.MaxParticles != 1000 {
		t.Errorf("MaxParticles = %d, want 1000", config.MaxParticles)
	}
	if config.CloseReduction != 1.0 {
		t.Errorf("CloseReduction = %f, want 1.0", config.CloseReduction)
	}
	if config.MidReduction != 0.5 {
		t.Errorf("MidReduction = %f, want 0.5", config.MidReduction)
	}
	if config.FarReduction != 0.25 {
		t.Errorf("FarReduction = %f, want 0.25", config.FarReduction)
	}
}

func TestDefaultViewportCullingConfig(t *testing.T) {
	config := DefaultViewportCullingConfig()

	if !config.Enabled {
		t.Error("Default viewport culling should be enabled")
	}
	if config.Margin != 100.0 {
		t.Errorf("Margin = %f, want 100.0", config.Margin)
	}
}

func TestApplyViewportCulling(t *testing.T) {
	particles := []Particle{
		{X: 100, Y: 100},  // Inside viewport
		{X: 500, Y: 500},  // Inside viewport
		{X: -200, Y: 100}, // Outside (left)
		{X: 1200, Y: 100}, // Outside (right)
		{X: 100, Y: -200}, // Outside (top)
		{X: 100, Y: 1200}, // Outside (bottom)
	}

	config := ViewportCullingConfig{
		Enabled:        true,
		Margin:         50.0,
		ViewportX:      0,
		ViewportY:      0,
		ViewportWidth:  800.0,
		ViewportHeight: 600.0,
	}

	visible := ApplyViewportCulling(particles, config)

	// Should have 2 particles inside viewport
	if len(visible) != 2 {
		t.Errorf("Visible particles = %d, want 2", len(visible))
	}

	// Verify correct indices
	if visible[0] != 0 || visible[1] != 1 {
		t.Errorf("Visible indices = %v, want [0, 1]", visible)
	}
}

func TestApplyViewportCulling_WithMargin(t *testing.T) {
	particles := []Particle{
		{X: -50, Y: 100},  // Just outside, but within margin
		{X: 850, Y: 100},  // Just outside, but within margin
		{X: -200, Y: 100}, // Outside margin
	}

	config := ViewportCullingConfig{
		Enabled:        true,
		Margin:         100.0,
		ViewportX:      0,
		ViewportY:      0,
		ViewportWidth:  800.0,
		ViewportHeight: 600.0,
	}

	visible := ApplyViewportCulling(particles, config)

	// Should have 2 particles (first two within margin)
	if len(visible) != 2 {
		t.Errorf("Visible particles = %d, want 2", len(visible))
	}
}

func TestApplyViewportCulling_Disabled(t *testing.T) {
	particles := []Particle{
		{X: 100, Y: 100},
		{X: 5000, Y: 5000}, // Far outside
	}

	config := ViewportCullingConfig{
		Enabled:        false,
		ViewportX:      0,
		ViewportY:      0,
		ViewportWidth:  800.0,
		ViewportHeight: 600.0,
	}

	visible := ApplyViewportCulling(particles, config)

	// Should return all particles when disabled
	if len(visible) != 2 {
		t.Errorf("Visible particles = %d, want 2 (culling disabled)", len(visible))
	}
}

func TestApplyDistanceLOD(t *testing.T) {
	// Create particles at various distances
	particles := []Particle{
		{X: 50, Y: 0},   // Close (50 units)
		{X: 100, Y: 0},  // Close (100 units)
		{X: 300, Y: 0},  // Mid (300 units)
		{X: 350, Y: 0},  // Mid (350 units)
		{X: 600, Y: 0},  // Far (600 units)
		{X: 700, Y: 0},  // Far (700 units)
		{X: 1000, Y: 0}, // Beyond far (culled)
	}

	visibleIndices := []int{0, 1, 2, 3, 4, 5, 6} // All visible initially

	config := LODConfig{
		Enabled:        true,
		MaxParticles:   1000,
		DistanceClose:  200.0,
		DistanceMid:    400.0,
		DistanceFar:    800.0,
		CloseReduction: 1.0, // Keep all close
		MidReduction:   0.5, // Keep half mid
		FarReduction:   0.5, // Keep half far
	}

	rendered := ApplyDistanceLOD(particles, visibleIndices, 0, 0, config)

	// Should have: 2 close + 1 mid (half of 2) + 1 far (half of 2) = 4
	// Note: 1 particle beyond DistanceFar is culled entirely
	if len(rendered) != 4 {
		t.Errorf("Rendered particles = %d, want 4 (2 close + 1 mid + 1 far)", len(rendered))
	}

	// Verify close particles are included
	hasClose := false
	for _, idx := range rendered {
		if idx == 0 || idx == 1 {
			hasClose = true
			break
		}
	}
	if !hasClose {
		t.Error("Should include close particles")
	}
}

func TestApplyDistanceLOD_Disabled(t *testing.T) {
	particles := []Particle{
		{X: 50, Y: 0},
		{X: 500, Y: 0},
		{X: 1000, Y: 0},
	}

	visibleIndices := []int{0, 1, 2}

	config := LODConfig{
		Enabled:        false,
		DistanceClose:  200.0,
		DistanceMid:    400.0,
		DistanceFar:    800.0,
		CloseReduction: 1.0,
		MidReduction:   0.5,
		FarReduction:   0.1,
	}

	rendered := ApplyDistanceLOD(particles, visibleIndices, 0, 0, config)

	// Should return all when disabled
	if len(rendered) != 3 {
		t.Errorf("Rendered particles = %d, want 3 (LOD disabled)", len(rendered))
	}
}

func TestEnforceLODLimit(t *testing.T) {
	particles := []Particle{
		{X: 10, Y: 0},  // Closest
		{X: 50, Y: 0},  // Close
		{X: 100, Y: 0}, // Mid
		{X: 200, Y: 0}, // Far
		{X: 500, Y: 0}, // Farthest
	}

	renderIndices := []int{0, 1, 2, 3, 4} // All 5

	// Limit to 3 particles
	limited := EnforceLODLimit(particles, renderIndices, 0, 0, 3)

	if len(limited) != 3 {
		t.Errorf("Limited particles = %d, want 3", len(limited))
	}

	// Should prioritize closest particles (indices 0, 1, 2)
	for _, idx := range limited {
		if idx > 2 {
			t.Errorf("Should prioritize closer particles, got index %d", idx)
		}
	}
}

func TestEnforceLODLimit_NoLimit(t *testing.T) {
	particles := []Particle{
		{X: 10, Y: 0},
		{X: 50, Y: 0},
	}

	renderIndices := []int{0, 1}

	// Limit of 10 (more than we have)
	limited := EnforceLODLimit(particles, renderIndices, 0, 0, 10)

	// Should return all when under limit
	if len(limited) != 2 {
		t.Errorf("Limited particles = %d, want 2", len(limited))
	}
}

func TestCalculateLODStats(t *testing.T) {
	stats := CalculateLODStats(100, []int{0, 1, 2, 3, 4}, []int{0, 1, 2})

	if stats.TotalParticles != 100 {
		t.Errorf("TotalParticles = %d, want 100", stats.TotalParticles)
	}
	if stats.VisibleParticles != 5 {
		t.Errorf("VisibleParticles = %d, want 5", stats.VisibleParticles)
	}
	if stats.RenderedParticles != 3 {
		t.Errorf("RenderedParticles = %d, want 3", stats.RenderedParticles)
	}
	if stats.CulledByViewport != 95 {
		t.Errorf("CulledByViewport = %d, want 95", stats.CulledByViewport)
	}
	if stats.ReducedByLOD != 2 {
		t.Errorf("ReducedByLOD = %d, want 2", stats.ReducedByLOD)
	}
}

func BenchmarkApplyViewportCulling(b *testing.B) {
	// Create 1000 particles
	particles := make([]Particle, 1000)
	for i := range particles {
		particles[i] = Particle{
			X: float64(i * 10),
			Y: float64(i % 100),
		}
	}

	config := DefaultViewportCullingConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyViewportCulling(particles, config)
	}
}

func BenchmarkApplyDistanceLOD(b *testing.B) {
	// Create 500 particles
	particles := make([]Particle, 500)
	for i := range particles {
		particles[i] = Particle{
			X: float64(i * 5),
			Y: 0,
		}
	}

	visibleIndices := make([]int, 500)
	for i := range visibleIndices {
		visibleIndices[i] = i
	}

	config := DefaultLODConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyDistanceLOD(particles, visibleIndices, 0, 0, config)
	}
}

func BenchmarkEnforceLODLimit(b *testing.B) {
	// Create 1000 particles
	particles := make([]Particle, 1000)
	for i := range particles {
		particles[i] = Particle{
			X: float64(i),
			Y: 0,
		}
	}

	renderIndices := make([]int, 1000)
	for i := range renderIndices {
		renderIndices[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EnforceLODLimit(particles, renderIndices, 0, 0, 500)
	}
}

func TestNewStaggeredLODEnforcer(t *testing.T) {
	enforcer := NewStaggeredLODEnforcer(4, 500)

	if enforcer.staggerFrames != 4 {
		t.Errorf("staggerFrames = %d, want 4", enforcer.staggerFrames)
	}
	if enforcer.maxParticles != 500 {
		t.Errorf("maxParticles = %d, want 500", enforcer.maxParticles)
	}
}

func TestNewStaggeredLODEnforcer_Defaults(t *testing.T) {
	// Zero values should use defaults
	enforcer := NewStaggeredLODEnforcer(0, 0)

	if enforcer.staggerFrames != 1 {
		t.Errorf("staggerFrames = %d, want 1 (min)", enforcer.staggerFrames)
	}
	if enforcer.maxParticles != 1000 {
		t.Errorf("maxParticles = %d, want 1000 (default)", enforcer.maxParticles)
	}
}

func TestStaggeredLODEnforcer_UnderLimit(t *testing.T) {
	enforcer := NewStaggeredLODEnforcer(4, 100)
	particles := make([]Particle, 50)
	for i := range particles {
		particles[i] = Particle{X: float64(i), Y: 0}
	}
	indices := make([]int, 50)
	for i := range indices {
		indices[i] = i
	}

	result := enforcer.Update(particles, indices, 0, 0)
	if len(result) != 50 {
		t.Errorf("result length = %d, want 50 (all under limit)", len(result))
	}
}

func TestStaggeredLODEnforcer_OverLimit(t *testing.T) {
	enforcer := NewStaggeredLODEnforcer(4, 100)
	particles := make([]Particle, 200)
	for i := range particles {
		particles[i] = Particle{X: float64(i * 10), Y: 0}
	}
	indices := make([]int, 200)
	for i := range indices {
		indices[i] = i
	}

	// First frame should return initial slice as fallback
	result := enforcer.Update(particles, indices, 0, 0)
	if len(result) != 100 {
		t.Errorf("first frame result length = %d, want 100", len(result))
	}

	// Run through remaining frames to complete cycle
	for i := 0; i < 3; i++ {
		result = enforcer.Update(particles, indices, 0, 0)
	}

	// After full cycle, should have sorted result with closest particles
	if len(result) != 100 {
		t.Errorf("result length = %d, want 100", len(result))
	}
}

func TestStaggeredLODEnforcer_CycleProgress(t *testing.T) {
	enforcer := NewStaggeredLODEnforcer(4, 100)
	particles := make([]Particle, 200)
	indices := make([]int, 200)
	for i := range indices {
		indices[i] = i
	}

	// Check initial progress
	current, total := enforcer.GetCycleProgress()
	if total != 4 {
		t.Errorf("total = %d, want 4", total)
	}
	if current != 0 {
		t.Errorf("initial current = %d, want 0", current)
	}

	// After one update
	enforcer.Update(particles, indices, 0, 0)
	current, _ = enforcer.GetCycleProgress()
	if current != 1 {
		t.Errorf("after 1 update current = %d, want 1", current)
	}

	// After 4 updates (complete cycle)
	for i := 0; i < 3; i++ {
		enforcer.Update(particles, indices, 0, 0)
	}
	current, _ = enforcer.GetCycleProgress()
	if current != 0 {
		t.Errorf("after cycle current = %d, want 0 (reset)", current)
	}
}

func TestStaggeredLODEnforcer_Reset(t *testing.T) {
	enforcer := NewStaggeredLODEnforcer(4, 100)
	particles := make([]Particle, 200)
	indices := make([]int, 200)
	for i := range indices {
		indices[i] = i
	}

	// Run a few frames
	enforcer.Update(particles, indices, 0, 0)
	enforcer.Update(particles, indices, 0, 0)

	// Reset
	enforcer.Reset()

	current, _ := enforcer.GetCycleProgress()
	if current != 0 {
		t.Errorf("current after reset = %d, want 0", current)
	}
}

func TestStaggeredLODEnforcer_SetMaxParticles(t *testing.T) {
	enforcer := NewStaggeredLODEnforcer(4, 100)
	enforcer.SetMaxParticles(200)

	if enforcer.maxParticles != 200 {
		t.Errorf("maxParticles = %d, want 200", enforcer.maxParticles)
	}

	// Invalid value should be ignored
	enforcer.SetMaxParticles(0)
	if enforcer.maxParticles != 200 {
		t.Errorf("maxParticles after invalid = %d, want 200 (unchanged)", enforcer.maxParticles)
	}
}

func TestStaggeredLODEnforcer_SetStaggerFrames(t *testing.T) {
	enforcer := NewStaggeredLODEnforcer(4, 100)
	enforcer.SetStaggerFrames(8)

	if enforcer.staggerFrames != 8 {
		t.Errorf("staggerFrames = %d, want 8", enforcer.staggerFrames)
	}

	// Invalid value should be ignored
	enforcer.SetStaggerFrames(0)
	if enforcer.staggerFrames != 8 {
		t.Errorf("staggerFrames after invalid = %d, want 8 (unchanged)", enforcer.staggerFrames)
	}
}

func BenchmarkStaggeredLODEnforcer(b *testing.B) {
	enforcer := NewStaggeredLODEnforcer(4, 500)
	particles := make([]Particle, 1000)
	for i := range particles {
		particles[i] = Particle{X: float64(i), Y: 0}
	}
	indices := make([]int, 1000)
	for i := range indices {
		indices[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enforcer.Update(particles, indices, 0, 0)
	}
}
