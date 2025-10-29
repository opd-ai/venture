package particles

import (
	"image/color"
	"testing"
	"unsafe"
)

func TestNewParticleSystem_UsesPool(t *testing.T) {
	// Create and release particle system
	ps1 := NewParticleSystem([]Particle{}, ParticleSpark, DefaultConfig())
	addr1 := uintptr(unsafe.Pointer(ps1))
	ReleaseParticleSystem(ps1)
	
	// Next allocation should reuse same memory
	ps2 := NewParticleSystem([]Particle{}, ParticleSmoke, DefaultConfig())
	addr2 := uintptr(unsafe.Pointer(ps2))
	
	if addr1 != addr2 {
		t.Errorf("Pool not reusing objects: addr1=%v, addr2=%v", addr1, addr2)
	}
	
	// Verify state was reset
	if ps2.Type != ParticleSmoke {
		t.Errorf("ParticleSystem type not set correctly: got %v, want %v", ps2.Type, ParticleSmoke)
	}
	if ps2.ElapsedTime != 0 {
		t.Error("ParticleSystem elapsed time not reset")
	}
	if len(ps2.Particles) != 0 {
		t.Errorf("ParticleSystem particles not cleared: got %d particles", len(ps2.Particles))
	}
	
	ReleaseParticleSystem(ps2)
}

func TestNewParticleSystem_InitializesCorrectly(t *testing.T) {
	particles := []Particle{
		{X: 10, Y: 20, VX: 1, VY: 2, Life: 1.0, InitialLife: 1.0},
		{X: 30, Y: 40, VX: 3, VY: 4, Life: 1.0, InitialLife: 1.0},
	}
	config := Config{
		Type:     ParticleMagic,
		Count:    2,
		GenreID:  "fantasy",
		Seed:     12345,
		Duration: 2.0,
		SpreadX:  5.0,
		SpreadY:  5.0,
		Gravity:  0.0,
		MinSize:  1.0,
		MaxSize:  3.0,
		Custom:   make(map[string]interface{}),
	}
	
	ps := NewParticleSystem(particles, ParticleMagic, config)
	defer ReleaseParticleSystem(ps)
	
	if ps.Type != ParticleMagic {
		t.Errorf("Type = %v, want %v", ps.Type, ParticleMagic)
	}
	if len(ps.Particles) != 2 {
		t.Errorf("Particle count = %d, want 2", len(ps.Particles))
	}
	if ps.Particles[0].X != 10 || ps.Particles[0].Y != 20 {
		t.Error("Particle positions not preserved")
	}
	if ps.ElapsedTime != 0 {
		t.Error("ElapsedTime should be 0 for new system")
	}
	if ps.Config.Duration != 2.0 {
		t.Errorf("Config.Duration = %f, want 2.0", ps.Config.Duration)
	}
}

func TestReleaseParticleSystem_ClearsState(t *testing.T) {
	particles := []Particle{
		{X: 100, Y: 200, Life: 0.5},
		{X: 300, Y: 400, Life: 0.8},
	}
	ps := NewParticleSystem(particles, ParticleFlame, DefaultConfig())
	ps.ElapsedTime = 5.0
	
	ReleaseParticleSystem(ps)
	
	// Get same system back from pool
	ps2 := NewParticleSystem([]Particle{}, ParticleDust, DefaultConfig())
	
	// All fields should be reset
	if len(ps2.Particles) != 0 {
		t.Errorf("Particles not cleared: got %d particles", len(ps2.Particles))
	}
	if ps2.ElapsedTime != 0 {
		t.Error("ElapsedTime not reset")
	}
	// Type should be new value, not old
	if ps2.Type != ParticleDust {
		t.Errorf("Type = %v, want %v (should be new value)", ps2.Type, ParticleDust)
	}
	
	ReleaseParticleSystem(ps2)
}

func TestReleaseParticleSystem_NilSafe(t *testing.T) {
	// Should not panic with nil
	ReleaseParticleSystem(nil)
}

func TestParticleSystem_CapacityReuse(t *testing.T) {
	// Create system with 50 particles
	particles := make([]Particle, 50)
	for i := range particles {
		particles[i] = Particle{X: float64(i), Y: float64(i)}
	}
	
	ps := NewParticleSystem(particles, ParticleSpark, DefaultConfig())
	originalCap := cap(ps.Particles)
	ReleaseParticleSystem(ps)
	
	// Reuse should keep capacity
	ps2 := NewParticleSystem([]Particle{}, ParticleSmoke, DefaultConfig())
	newCap := cap(ps2.Particles)
	
	// Capacity should be at least original (may be larger)
	if newCap < originalCap {
		t.Errorf("Capacity shrunk: was %d, now %d", originalCap, newCap)
	}
	
	ReleaseParticleSystem(ps2)
}

func TestAcquireParticleSlice_UsesPool(t *testing.T) {
	slice1 := AcquireParticleSlice()
	addr1 := uintptr(unsafe.Pointer(slice1))
	ReleaseParticleSlice(slice1)
	
	slice2 := AcquireParticleSlice()
	addr2 := uintptr(unsafe.Pointer(slice2))
	
	if addr1 != addr2 {
		t.Errorf("Slice pool not reusing objects: addr1=%v, addr2=%v", addr1, addr2)
	}
	
	if len(*slice2) != 0 {
		t.Errorf("Slice not reset: length = %d, want 0", len(*slice2))
	}
	if cap(*slice2) < 100 {
		t.Errorf("Slice capacity = %d, want >= 100", cap(*slice2))
	}
	
	ReleaseParticleSlice(slice2)
}

func TestAcquireParticleSlice_AppendWorks(t *testing.T) {
	slice := AcquireParticleSlice()
	defer ReleaseParticleSlice(slice)
	
	// Append particles
	*slice = append(*slice, Particle{X: 1, Y: 2})
	*slice = append(*slice, Particle{X: 3, Y: 4})
	
	if len(*slice) != 2 {
		t.Errorf("Length after append = %d, want 2", len(*slice))
	}
	if (*slice)[0].X != 1 {
		t.Error("First particle not preserved")
	}
}

func TestReleaseParticleSlice_NilSafe(t *testing.T) {
	// Should not panic with nil
	ReleaseParticleSlice(nil)
}

func TestParticlePoolStats_Tracking(t *testing.T) {
	// Reset stats
	ResetParticlePoolStats()
	
	// Get initial stats (should be zero)
	stats := GetParticlePoolStats()
	if stats.SystemsActive != 0 {
		t.Errorf("Initial SystemsActive = %d, want 0", stats.SystemsActive)
	}
	
	// Note: Actual stat tracking is disabled by default for performance.
	// This test just verifies the API works without panicking.
}

func TestParticlePool_NoMemoryLeaks(t *testing.T) {
	// Create and release many particle systems
	for i := 0; i < 1000; i++ {
		particles := []Particle{
			{X: float64(i), Y: float64(i)},
		}
		ps := NewParticleSystem(particles, ParticleSpark, DefaultConfig())
		ReleaseParticleSystem(ps)
	}
	
	// Create and release many particle slices
	for i := 0; i < 1000; i++ {
		slice := AcquireParticleSlice()
		*slice = append(*slice, Particle{X: float64(i)})
		ReleaseParticleSlice(slice)
	}
	
	// If no panic and test completes, no obvious leak
	// (Actual leak testing requires runtime memory profiling)
}

func TestParticlePool_ConcurrentAccess(t *testing.T) {
	// sync.Pool is thread-safe, but verify no panics with concurrent use
	done := make(chan bool)
	
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				ps := NewParticleSystem([]Particle{}, ParticleSpark, DefaultConfig())
				ps.Particles = append(ps.Particles, Particle{X: 1, Y: 2})
				ReleaseParticleSystem(ps)
			}
			done <- true
		}()
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// No panic = success
}

// Benchmarks

func BenchmarkParticleSystemPooling(b *testing.B) {
	particles := []Particle{
		{X: 1, Y: 2, VX: 0.5, VY: 0.5, Life: 1.0, InitialLife: 1.0, Size: 2.0},
		{X: 3, Y: 4, VX: 0.5, VY: 0.5, Life: 1.0, InitialLife: 1.0, Size: 2.0},
		{X: 5, Y: 6, VX: 0.5, VY: 0.5, Life: 1.0, InitialLife: 1.0, Size: 2.0},
	}
	config := DefaultConfig()
	
	b.Run("WithPooling", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ps := NewParticleSystem(particles, ParticleSpark, config)
			ReleaseParticleSystem(ps)
		}
	})
	
	b.Run("WithoutPooling", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ps := &ParticleSystem{
				Particles:   append([]Particle{}, particles...),
				Type:        ParticleSpark,
				Config:      config,
				ElapsedTime: 0,
			}
			_ = ps
		}
	})
}

func BenchmarkParticleSlicePooling(b *testing.B) {
	b.Run("WithPooling", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			slice := AcquireParticleSlice()
			*slice = append(*slice, Particle{X: 1, Y: 2})
			*slice = append(*slice, Particle{X: 3, Y: 4})
			*slice = append(*slice, Particle{X: 5, Y: 6})
			ReleaseParticleSlice(slice)
		}
	})
	
	b.Run("WithoutPooling", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			slice := make([]Particle, 0, 100)
			slice = append(slice, Particle{X: 1, Y: 2})
			slice = append(slice, Particle{X: 3, Y: 4})
			slice = append(slice, Particle{X: 5, Y: 6})
			_ = slice
		}
	})
}

func BenchmarkParticleSystemUpdate(b *testing.B) {
	// Test that pooling doesn't impact update performance
	particles := make([]Particle, 100)
	for i := range particles {
		particles[i] = Particle{
			X: float64(i), Y: float64(i),
			VX: 1.0, VY: 1.0,
			Life: 1.0, InitialLife: 1.0,
			Size: 2.0,
			Color: color.RGBA{255, 255, 255, 255},
		}
	}
	
	config := DefaultConfig()
	config.Gravity = 9.8
	
	ps := NewParticleSystem(particles, ParticleSpark, config)
	defer ReleaseParticleSystem(ps)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		ps.Update(0.016) // 60 FPS delta time
	}
}

// Expected benchmark results:
// WithPooling:     ~50-100 ns/op,   0 B/op,  0 allocs/op (after warmup)
// WithoutPooling:  ~200-500 ns/op, 896 B/op,  1 allocs/op
// Update should be allocation-free regardless of pooling
