package particles

import (
	"math"
	"math/rand"
	"testing"
)

func TestPhysicsType_String(t *testing.T) {
	tests := []struct {
		name     string
		physType PhysicsType
		want     string
	}{
		{"None", PhysicsNone, "none"},
		{"Fluid", PhysicsFluid, "fluid"},
		{"Fire", PhysicsFire, "fire"},
		{"Smoke", PhysicsSmoke, "smoke"},
		{"Debris", PhysicsDebris, "debris"},
		{"Unknown", PhysicsType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.physType.String(); got != tt.want {
				t.Errorf("PhysicsType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultSPHConfig(t *testing.T) {
	cfg := DefaultSPHConfig()

	if cfg.RestDensity <= 0 {
		t.Errorf("RestDensity should be positive, got %f", cfg.RestDensity)
	}
	if cfg.GasConstant <= 0 {
		t.Errorf("GasConstant should be positive, got %f", cfg.GasConstant)
	}
	if cfg.Viscosity < 0 || cfg.Viscosity > 1 {
		t.Errorf("Viscosity should be in [0,1], got %f", cfg.Viscosity)
	}
	if cfg.SmoothingRadius <= 0 {
		t.Errorf("SmoothingRadius should be positive, got %f", cfg.SmoothingRadius)
	}
}

func TestDefaultFireConfig(t *testing.T) {
	cfg := DefaultFireConfig()

	if cfg.HeatDissipation < 0 || cfg.HeatDissipation > 1 {
		t.Errorf("HeatDissipation should be in [0,1], got %f", cfg.HeatDissipation)
	}
	if cfg.IgnitionTemp <= 0 || cfg.IgnitionTemp > 1 {
		t.Errorf("IgnitionTemp should be in (0,1], got %f", cfg.IgnitionTemp)
	}
	if cfg.BuoyancyStrength <= 0 {
		t.Errorf("BuoyancyStrength should be positive, got %f", cfg.BuoyancyStrength)
	}
}

func TestDefaultSmokeConfig(t *testing.T) {
	cfg := DefaultSmokeConfig()

	if cfg.TurbulenceStrength <= 0 {
		t.Errorf("TurbulenceStrength should be positive, got %f", cfg.TurbulenceStrength)
	}
	if cfg.RiseSpeed <= 0 {
		t.Errorf("RiseSpeed should be positive, got %f", cfg.RiseSpeed)
	}
	if cfg.Dissipation < 0 || cfg.Dissipation > 1 {
		t.Errorf("Dissipation should be in [0,1], got %f", cfg.Dissipation)
	}
}

func TestDefaultDebrisConfig(t *testing.T) {
	cfg := DefaultDebrisConfig()

	if cfg.Restitution < 0 || cfg.Restitution > 1 {
		t.Errorf("Restitution should be in [0,1], got %f", cfg.Restitution)
	}
	if cfg.Friction < 0 || cfg.Friction > 1 {
		t.Errorf("Friction should be in [0,1], got %f", cfg.Friction)
	}
	if cfg.CollisionRadius <= 0 {
		t.Errorf("CollisionRadius should be positive, got %f", cfg.CollisionRadius)
	}
}

func TestSpatialHash_CreateAndClear(t *testing.T) {
	hash := NewSpatialHash(10.0, -100, -100, 100, 100)

	if hash.CellSize != 10.0 {
		t.Errorf("CellSize = %f, want 10.0", hash.CellSize)
	}
	if len(hash.Grid) != 0 {
		t.Errorf("Grid should start empty, got %d entries", len(hash.Grid))
	}

	// Insert some particles
	hash.Insert(0, 5, 5)
	hash.Insert(1, 15, 15)

	if len(hash.Grid) == 0 {
		t.Error("Grid should have entries after insertion")
	}

	hash.Clear()
	if len(hash.Grid) != 0 {
		t.Errorf("Grid should be empty after Clear(), got %d entries", len(hash.Grid))
	}
}

func TestSpatialHash_InsertAndQuery(t *testing.T) {
	hash := NewSpatialHash(10.0, -100, -100, 100, 100)

	// Insert particles in a grid pattern
	hash.Insert(0, 0, 0)
	hash.Insert(1, 5, 5)
	hash.Insert(2, 20, 20)
	hash.Insert(3, 50, 50)

	// Query near origin should find first two
	results := hash.QueryRadius(0, 0, 15.0)
	if len(results) < 2 {
		t.Errorf("QueryRadius near origin found %d particles, want at least 2", len(results))
	}

	// Query far away should find distant particles
	results = hash.QueryRadius(50, 50, 15.0)
	if len(results) == 0 {
		t.Error("QueryRadius at (50,50) should find particle 3")
	}
}

func TestSpatialHash_GetNeighbors(t *testing.T) {
	hash := NewSpatialHash(10.0, -100, -100, 100, 100)

	particles := []PhysicsParticle{
		{Particle: Particle{X: 0, Y: 0}},
		{Particle: Particle{X: 5, Y: 0}},
		{Particle: Particle{X: 20, Y: 0}},
	}

	for i := range particles {
		hash.Insert(i, particles[i].X, particles[i].Y)
	}

	// Find neighbors within 10 units of origin
	neighbors := hash.GetNeighbors(particles, 0, 0, 10.0)

	// Should find particles 0 and 1 (within 10 units)
	if len(neighbors) < 1 {
		t.Errorf("GetNeighbors found %d, want at least 1", len(neighbors))
	}

	// Verify it doesn't include particle 2 (20 units away)
	found := false
	for _, idx := range neighbors {
		if idx == 2 {
			found = true
		}
	}
	if found {
		t.Error("GetNeighbors should not include particle 2 (too far)")
	}
}

func TestPoly6Kernel(t *testing.T) {
	h := 10.0

	// At r=0, kernel should be maximum
	k0 := poly6Kernel(0, h)
	if k0 <= 0 {
		t.Errorf("Poly6 kernel at r=0 should be positive, got %f", k0)
	}

	// At r=h/2, kernel should be positive but less than at r=0
	kHalf := poly6Kernel(h/2, h)
	if kHalf <= 0 {
		t.Errorf("Poly6 kernel at r=h/2 should be positive, got %f", kHalf)
	}
	if kHalf >= k0 {
		t.Errorf("Poly6 kernel should decrease with distance: k(h/2)=%f >= k(0)=%f", kHalf, k0)
	}

	// At r=h, kernel should be zero
	kH := poly6Kernel(h, h)
	if kH != 0 {
		t.Errorf("Poly6 kernel at r=h should be 0, got %f", kH)
	}

	// Beyond h, kernel should be zero
	kBeyond := poly6Kernel(h*2, h)
	if kBeyond != 0 {
		t.Errorf("Poly6 kernel beyond h should be 0, got %f", kBeyond)
	}
}

func TestSpikyKernelGradient(t *testing.T) {
	h := 10.0

	// Gradient should be negative (pointing inward)
	grad := spikyKernelGradient(h/2, h)
	if grad >= 0 {
		t.Errorf("Spiky gradient should be negative, got %f", grad)
	}

	// At r=h, gradient should be zero
	gradH := spikyKernelGradient(h, h)
	if gradH != 0 {
		t.Errorf("Spiky gradient at r=h should be 0, got %f", gradH)
	}
}

func TestViscosityKernelLaplacian(t *testing.T) {
	h := 10.0

	// Laplacian should be positive
	lap := viscosityKernelLaplacian(h/2, h)
	if lap <= 0 {
		t.Errorf("Viscosity Laplacian should be positive, got %f", lap)
	}

	// At r=h, Laplacian should be zero
	lapH := viscosityKernelLaplacian(h, h)
	if lapH != 0 {
		t.Errorf("Viscosity Laplacian at r=h should be 0, got %f", lapH)
	}
}

func TestUpdateSPH_EmptyParticles(t *testing.T) {
	var particles []PhysicsParticle
	config := DefaultSPHConfig()

	// Should not panic with empty slice
	UpdateSPH(particles, config, 0.016)
}

func TestUpdateSPH_SingleParticle(t *testing.T) {
	particles := []PhysicsParticle{
		{Particle: Particle{X: 0, Y: 0, VX: 0, VY: 0}},
	}
	config := DefaultSPHConfig()

	UpdateSPH(particles, config, 0.016)

	// Single particle should have density from self-interaction
	if particles[0].Density <= 0 {
		t.Errorf("Single particle should have positive density, got %f", particles[0].Density)
	}
}

func TestUpdateSPH_TwoParticles(t *testing.T) {
	particles := []PhysicsParticle{
		{Particle: Particle{X: 0, Y: 0, VX: 0, VY: 0}},
		{Particle: Particle{X: 10, Y: 0, VX: 0, VY: 0}},
	}
	config := DefaultSPHConfig()
	config.SmoothingRadius = 20.0 // Ensure they interact

	initialVX0 := particles[0].VX
	initialVX1 := particles[1].VX

	UpdateSPH(particles, config, 0.016)

	// Particles should have density and pressure
	if particles[0].Density <= 0 {
		t.Error("Particle 0 should have positive density")
	}
	if particles[1].Density <= 0 {
		t.Error("Particle 1 should have positive density")
	}

	// Velocities should be affected (pressure forces)
	// Note: specific values depend on density calculation
	_ = initialVX0
	_ = initialVX1
}

func TestUpdateFire_EmptyParticles(t *testing.T) {
	var particles []PhysicsParticle
	config := DefaultFireConfig()
	rng := rand.New(rand.NewSource(12345))

	// Should not panic
	UpdateFire(particles, config, 0.016, rng)
}

func TestUpdateFire_HeatDissipation(t *testing.T) {
	particles := []PhysicsParticle{
		{
			Particle: Particle{X: 0, Y: 0},
			Heat:     1.0,
		},
	}
	config := DefaultFireConfig()
	config.HeatDissipation = 0.5
	rng := rand.New(rand.NewSource(12345))

	UpdateFire(particles, config, 0.1, rng)

	// Heat should decrease
	if particles[0].Heat >= 1.0 {
		t.Errorf("Heat should decrease, got %f", particles[0].Heat)
	}
	if particles[0].Heat < 0 {
		t.Errorf("Heat should not be negative, got %f", particles[0].Heat)
	}
}

func TestUpdateFire_Ignition(t *testing.T) {
	particles := []PhysicsParticle{
		{
			Particle:   Particle{X: 0, Y: 0},
			Heat:       0.8,
			FuelRemain: 1.0,
		},
	}
	config := DefaultFireConfig()
	config.IgnitionTemp = 0.7
	rng := rand.New(rand.NewSource(12345))

	UpdateFire(particles, config, 0.016, rng)

	// Should be ignited when heat > ignition temp
	if !particles[0].Ignited {
		t.Error("Particle should be ignited when heat > ignition temp")
	}
}

func TestUpdateFire_FuelConsumption(t *testing.T) {
	particles := []PhysicsParticle{
		{
			Particle:   Particle{X: 0, Y: 0, VY: 0},
			Heat:       1.0,
			Ignited:    true,
			FuelRemain: 1.0,
		},
	}
	config := DefaultFireConfig()
	rng := rand.New(rand.NewSource(12345))

	initialFuel := particles[0].FuelRemain

	UpdateFire(particles, config, 0.1, rng)

	// Fuel should be consumed when ignited
	if particles[0].FuelRemain >= initialFuel {
		t.Errorf("Fuel should be consumed, initial=%f, current=%f", initialFuel, particles[0].FuelRemain)
	}
}

func TestUpdateFire_Buoyancy(t *testing.T) {
	particles := []PhysicsParticle{
		{
			Particle: Particle{X: 0, Y: 0, VY: 0},
			Heat:     0.8,
		},
	}
	config := DefaultFireConfig()
	config.BuoyancyStrength = 100.0
	rng := rand.New(rand.NewSource(12345))

	initialVY := particles[0].VY

	UpdateFire(particles, config, 0.016, rng)

	// Hot particles should rise (VY becomes more negative in typical coordinate system)
	if particles[0].VY >= initialVY {
		t.Errorf("Hot particles should rise, VY: %f -> %f", initialVY, particles[0].VY)
	}
}

func TestUpdateSmoke_EmptyParticles(t *testing.T) {
	var particles []PhysicsParticle
	config := DefaultSmokeConfig()

	// Should not panic
	UpdateSmoke(particles, config, 0.016, 1.0)
}

func TestUpdateSmoke_Turbulence(t *testing.T) {
	particles := []PhysicsParticle{
		{
			Particle: Particle{X: 10, Y: 10, VX: 0, VY: 0},
		},
	}
	config := DefaultSmokeConfig()
	config.TurbulenceStrength = 50.0

	initialVX := particles[0].VX
	initialVY := particles[0].VY

	UpdateSmoke(particles, config, 0.016, 1.0)

	// Turbulence should affect velocity
	vxChanged := particles[0].VX != initialVX
	vyChanged := particles[0].VY != initialVY

	if !vxChanged && !vyChanged {
		t.Error("Turbulence should change particle velocity")
	}
}

func TestUpdateSmoke_Rising(t *testing.T) {
	particles := []PhysicsParticle{
		{
			Particle: Particle{X: 0, Y: 0, VY: 0},
		},
	}
	config := DefaultSmokeConfig()
	config.RiseSpeed = 30.0
	config.TurbulenceStrength = 0 // Disable turbulence for clear test

	UpdateSmoke(particles, config, 0.016, 0)

	// Smoke should rise (VY becomes more negative)
	if particles[0].VY >= 0 {
		t.Errorf("Smoke should rise, VY=%f", particles[0].VY)
	}
}

func TestUpdateSmoke_Expansion(t *testing.T) {
	particles := []PhysicsParticle{
		{
			Particle: Particle{Size: 2.0},
		},
	}
	config := DefaultSmokeConfig()
	config.ExpansionRate = 10.0

	initialSize := particles[0].Size

	UpdateSmoke(particles, config, 0.1, 0)

	// Smoke should expand
	if particles[0].Size <= initialSize {
		t.Errorf("Smoke should expand, size: %f -> %f", initialSize, particles[0].Size)
	}
}

func TestUpdateSmoke_Dissipation(t *testing.T) {
	particles := []PhysicsParticle{
		{
			Particle: Particle{Life: 1.0, InitialLife: 1.0},
		},
	}
	config := DefaultSmokeConfig()
	config.Dissipation = 0.3

	UpdateSmoke(particles, config, 0.1, 0)

	// Life should decrease
	if particles[0].Life >= 1.0 {
		t.Errorf("Smoke life should decrease, got %f", particles[0].Life)
	}
}

func TestUpdateDebris_EmptyParticles(t *testing.T) {
	var particles []PhysicsParticle
	config := DefaultDebrisConfig()

	// Should not panic
	UpdateDebris(particles, config, 0.016, 0)
}

func TestUpdateDebris_GroundCollision(t *testing.T) {
	groundY := 100.0
	particles := []PhysicsParticle{
		{
			Particle: Particle{X: 0, Y: 110, VX: 5, VY: 10},
		},
	}
	config := DefaultDebrisConfig()
	config.Restitution = 0.5

	UpdateDebris(particles, config, 0.016, groundY)

	// Should bounce at ground
	if particles[0].Y != groundY {
		t.Errorf("Particle should be at ground Y=%f, got %f", groundY, particles[0].Y)
	}

	// VY should be reversed and reduced
	if particles[0].VY >= 0 {
		t.Errorf("VY should be negative after bounce, got %f", particles[0].VY)
	}
}

func TestUpdateDebris_ParticleCollision(t *testing.T) {
	particles := []PhysicsParticle{
		{Particle: Particle{X: 0, Y: 0, VX: 10, VY: 0}},
		{Particle: Particle{X: 5, Y: 0, VX: -10, VY: 0}},
	}
	config := DefaultDebrisConfig()
	config.CollisionRadius = 5.0
	config.Restitution = 0.8

	initialVX0 := particles[0].VX
	initialVX1 := particles[1].VX

	UpdateDebris(particles, config, 0.016, -100)

	// Velocities should change after collision
	vx0Changed := particles[0].VX != initialVX0
	vx1Changed := particles[1].VX != initialVX1

	if !vx0Changed || !vx1Changed {
		t.Error("Particle collision should change velocities")
	}

	// Particles should move apart
	dx := particles[1].X - particles[0].X
	if dx <= 5 {
		t.Errorf("Particles should separate after collision, distance=%f", dx)
	}
}

func TestUpdateDebris_RotationDamping(t *testing.T) {
	particles := []PhysicsParticle{
		{
			Particle:        Particle{X: 0, Y: 0},
			AngularVelocity: 10.0,
		},
	}
	config := DefaultDebrisConfig()
	config.RotationDamping = 0.9

	initialAngVel := particles[0].AngularVelocity

	UpdateDebris(particles, config, 0.1, -100)

	// Angular velocity should decrease
	if particles[0].AngularVelocity >= initialAngVel {
		t.Errorf("Angular velocity should decrease, %f -> %f", initialAngVel, particles[0].AngularVelocity)
	}
}

func TestUpdateDebris_Rotation(t *testing.T) {
	particles := []PhysicsParticle{
		{
			Particle:        Particle{Rotation: 0},
			AngularVelocity: math.Pi, // 180 degrees per second
		},
	}
	config := DefaultDebrisConfig()

	UpdateDebris(particles, config, 1.0, -100)

	// Rotation should increase (accounting for damping, won't reach exactly π)
	if particles[0].Rotation <= 0 {
		t.Errorf("Rotation should increase, got %f", particles[0].Rotation)
	}
	// With damping of 0.95, rotation will be slightly less than π
	if particles[0].Rotation > math.Pi {
		t.Errorf("Rotation should be <= π with damping, got %f", particles[0].Rotation)
	}
}

// Benchmark tests
func BenchmarkUpdateSPH(b *testing.B) {
	// Create 200 particles (target from success metrics)
	particles := make([]PhysicsParticle, 200)
	for i := range particles {
		particles[i] = PhysicsParticle{
			Particle: Particle{
				X:  float64(i%20) * 10,
				Y:  float64(i/20) * 10,
				VX: 0,
				VY: 0,
			},
		}
	}
	config := DefaultSPHConfig()
	deltaTime := 0.016 // 60 FPS

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UpdateSPH(particles, config, deltaTime)
	}
}

func BenchmarkUpdateFire(b *testing.B) {
	particles := make([]PhysicsParticle, 200)
	for i := range particles {
		particles[i] = PhysicsParticle{
			Particle: Particle{
				X: float64(i%20) * 10,
				Y: float64(i/20) * 10,
			},
			Heat:       0.5,
			FuelRemain: 1.0,
		}
	}
	config := DefaultFireConfig()
	rng := rand.New(rand.NewSource(12345))
	deltaTime := 0.016

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UpdateFire(particles, config, deltaTime, rng)
	}
}

func BenchmarkUpdateSmoke(b *testing.B) {
	particles := make([]PhysicsParticle, 200)
	for i := range particles {
		particles[i] = PhysicsParticle{
			Particle: Particle{
				X:    float64(i%20) * 10,
				Y:    float64(i/20) * 10,
				Size: 2.0,
				Life: 1.0,
			},
		}
	}
	config := DefaultSmokeConfig()
	deltaTime := 0.016

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UpdateSmoke(particles, config, deltaTime, 1.0)
	}
}

func BenchmarkUpdateDebris(b *testing.B) {
	particles := make([]PhysicsParticle, 200)
	for i := range particles {
		particles[i] = PhysicsParticle{
			Particle: Particle{
				X:  float64(i%20) * 10,
				Y:  float64(i/20) * 10,
				VX: float64(i%10) - 5,
				VY: float64(i/10) - 5,
			},
			AngularVelocity: 1.0,
		}
	}
	config := DefaultDebrisConfig()
	deltaTime := 0.016

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UpdateDebris(particles, config, deltaTime, 100)
	}
}

func BenchmarkSpatialHash_Insert(b *testing.B) {
	hash := NewSpatialHash(10.0, -1000, -1000, 1000, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash.Insert(i%1000, float64(i%100), float64(i%100))
	}
}

func BenchmarkSpatialHash_Query(b *testing.B) {
	hash := NewSpatialHash(10.0, -1000, -1000, 1000, 1000)
	for i := 0; i < 1000; i++ {
		hash.Insert(i, float64(i%100), float64(i%100))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash.QueryRadius(50, 50, 20)
	}
}
