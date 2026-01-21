// Package particles provides advanced particle physics simulations.
// This file implements SPH (Smoothed Particle Hydrodynamics) for fluids,
// fire propagation, smoke turbulence, and particle-particle interactions.
package particles

import (
	"math"
	"math/rand"
)

// PhysicsType defines different physics simulation types.
type PhysicsType int

const (
	// PhysicsNone represents no physics simulation
	PhysicsNone PhysicsType = iota
	// PhysicsFluid represents SPH fluid simulation
	PhysicsFluid
	// PhysicsFire represents fire propagation with heat
	PhysicsFire
	// PhysicsSmoke represents smoke with turbulence
	PhysicsSmoke
	// PhysicsDebris represents bouncing debris with collisions
	PhysicsDebris
)

// String returns the string representation of physics type.
func (p PhysicsType) String() string {
	switch p {
	case PhysicsNone:
		return "none"
	case PhysicsFluid:
		return "fluid"
	case PhysicsFire:
		return "fire"
	case PhysicsSmoke:
		return "smoke"
	case PhysicsDebris:
		return "debris"
	default:
		return "unknown"
	}
}

// SPHConfig contains parameters for Smoothed Particle Hydrodynamics fluid simulation.
type SPHConfig struct {
	// RestDensity is the target density for fluid (kg/m³)
	RestDensity float64

	// GasConstant affects pressure calculation
	GasConstant float64

	// Viscosity controls flow resistance (0.0-1.0)
	Viscosity float64

	// SmoothingRadius defines neighbor interaction distance
	SmoothingRadius float64

	// ParticleMass mass of each particle
	ParticleMass float64

	// SurfaceTension for cohesion effects
	SurfaceTension float64
}

// DefaultSPHConfig returns sensible defaults for fluid simulation.
func DefaultSPHConfig() SPHConfig {
	return SPHConfig{
		RestDensity:     1000.0, // Water density
		GasConstant:     2000.0, // Moderate pressure response
		Viscosity:       0.3,    // Low viscosity (water-like)
		SmoothingRadius: 16.0,   // Moderate interaction radius
		ParticleMass:    1.0,    // Unit mass
		SurfaceTension:  0.1,    // Weak cohesion
	}
}

// FireConfig contains parameters for fire propagation simulation.
type FireConfig struct {
	// HeatDissipation rate heat decreases (0.0-1.0 per second)
	HeatDissipation float64

	// IgnitionTemp temperature threshold to ignite neighbors
	IgnitionTemp float64

	// HeatTransferRate how fast heat spreads to neighbors
	HeatTransferRate float64

	// BuoyancyStrength upward force from heat
	BuoyancyStrength float64

	// EmberChance probability of spawning ember particles (0.0-1.0)
	EmberChance float64

	// FuelConsumptionRate how fast fuel is consumed when ignited (per second)
	FuelConsumptionRate float64

	// HeatTransferRadius distance at which heat spreads to neighbors
	HeatTransferRadius float64

	// HeatTransferFraction fraction of heat transferred to neighbors
	HeatTransferFraction float64

	// ExtinguishHeatMultiplier heat reduction when fuel runs out
	ExtinguishHeatMultiplier float64

	// MinActiveHeat minimum heat threshold for buoyancy and heat transfer
	MinActiveHeat float64
}

// DefaultFireConfig returns sensible defaults for fire simulation.
func DefaultFireConfig() FireConfig {
	return FireConfig{
		HeatDissipation:          0.5,   // Moderate cooling
		IgnitionTemp:             0.7,   // 70% heat to ignite
		HeatTransferRate:         0.3,   // Moderate spread
		BuoyancyStrength:         100.0, // Strong upward force
		EmberChance:              0.05,  // 5% chance per frame
		FuelConsumptionRate:      0.2,   // Moderate fuel burn rate
		HeatTransferRadius:       20.0,  // Heat spreads within 20 pixels
		HeatTransferFraction:     0.5,   // Transfer 50% of heat to neighbors
		ExtinguishHeatMultiplier: 0.5,   // Reduce to 50% when fuel runs out
		MinActiveHeat:            0.1,   // Heat threshold for effects
	}
}

// SmokeConfig contains parameters for smoke turbulence simulation.
type SmokeConfig struct {
	// TurbulenceStrength intensity of chaotic motion
	TurbulenceStrength float64

	// TurbulenceFreq frequency of turbulence changes
	TurbulenceFreq float64

	// RiseSpeed base upward velocity
	RiseSpeed float64

	// Dissipation rate smoke fades (0.0-1.0 per second)
	Dissipation float64

	// ExpansionRate how fast smoke expands
	ExpansionRate float64

	// TurbulenceNoiseScale scaling factor for noise function (controls detail)
	TurbulenceNoiseScale float64
}

// DefaultSmokeConfig returns sensible defaults for smoke simulation.
func DefaultSmokeConfig() SmokeConfig {
	return SmokeConfig{
		TurbulenceStrength:   50.0, // Moderate turbulence
		TurbulenceFreq:       2.0,  // 2 Hz turbulence
		RiseSpeed:            30.0, // Slow rise
		Dissipation:          0.3,  // Slow fade
		ExpansionRate:        10.0, // Moderate expansion
		TurbulenceNoiseScale: 0.1,  // Detail level for turbulence
	}
}

// DebrisConfig contains parameters for debris collision simulation.
type DebrisConfig struct {
	// Restitution coefficient for bouncing (0.0-1.0)
	Restitution float64

	// Friction coefficient for sliding (0.0-1.0)
	Friction float64

	// RotationDamping angular velocity damping
	RotationDamping float64

	// CollisionRadius for particle-particle collisions
	CollisionRadius float64
}

// DefaultDebrisConfig returns sensible defaults for debris simulation.
func DefaultDebrisConfig() DebrisConfig {
	return DebrisConfig{
		Restitution:     0.4,  // Moderate bounce
		Friction:        0.7,  // High friction
		RotationDamping: 0.95, // Light damping
		CollisionRadius: 4.0,  // Small collision radius
	}
}

// PhysicsParticle extends Particle with physics simulation data.
type PhysicsParticle struct {
	Particle

	// Physics type for this particle
	PhysicsType PhysicsType

	// SPH-specific fields
	Density  float64 // Current density (for fluids)
	Pressure float64 // Current pressure (for fluids)

	// Fire-specific fields
	Heat       float64 // Current heat (0.0-1.0)
	Ignited    bool    // Whether particle is on fire
	FuelRemain float64 // Remaining fuel (0.0-1.0)
	EmberTimer float64 // Time until next ember spawn

	// Smoke-specific fields
	TurbulencePhase float64 // Phase for turbulence noise
	ExpansionTime   float64 // Time spent expanding

	// Debris-specific fields
	AngularVelocity float64 // Rotation speed (radians/second)
	LastCollision   float64 // Time since last collision
}

// SpatialHash provides efficient neighbor queries using spatial hashing.
type SpatialHash struct {
	// CellSize size of each grid cell
	CellSize float64

	// Grid maps cell coordinates to particle indices
	Grid map[int64][]int

	// WorldBounds for wrapping (minX, minY, maxX, maxY)
	MinX, MinY, MaxX, MaxY float64
}

// NewSpatialHash creates a spatial hash grid.
func NewSpatialHash(cellSize, minX, minY, maxX, maxY float64) *SpatialHash {
	return &SpatialHash{
		CellSize: cellSize,
		Grid:     make(map[int64][]int),
		MinX:     minX,
		MinY:     minY,
		MaxX:     maxX,
		MaxY:     maxY,
	}
}

// hashKey computes a unique key for a grid cell.
func (sh *SpatialHash) hashKey(x, y int) int64 {
	// Use prime numbers to reduce collisions
	const prime1 = 73856093
	const prime2 = 19349663
	return int64(x)*prime1 ^ int64(y)*prime2
}

// getCellCoords returns the grid cell coordinates for a position.
func (sh *SpatialHash) getCellCoords(x, y float64) (int, int) {
	cellX := int(math.Floor(x / sh.CellSize))
	cellY := int(math.Floor(y / sh.CellSize))
	return cellX, cellY
}

// Clear removes all particles from the grid.
func (sh *SpatialHash) Clear() {
	sh.Grid = make(map[int64][]int)
}

// Insert adds a particle to the spatial hash.
func (sh *SpatialHash) Insert(index int, x, y float64) {
	cellX, cellY := sh.getCellCoords(x, y)
	key := sh.hashKey(cellX, cellY)
	sh.Grid[key] = append(sh.Grid[key], index)
}

// QueryRadius returns all particle indices within radius of position.
func (sh *SpatialHash) QueryRadius(x, y, radius float64) []int {
	return sh.QueryRadiusInto(x, y, radius, nil)
}

// QueryRadiusInto appends particle indices within radius to the provided buffer.
// If buffer is nil, a new slice is allocated. Returns the updated buffer.
// This method enables zero-allocation queries when the buffer is pre-allocated and reused.
func (sh *SpatialHash) QueryRadiusInto(x, y, radius float64, buffer []int) []int {
	// Determine which cells to check
	minCellX := int(math.Floor((x - radius) / sh.CellSize))
	maxCellX := int(math.Floor((x + radius) / sh.CellSize))
	minCellY := int(math.Floor((y - radius) / sh.CellSize))
	maxCellY := int(math.Floor((y + radius) / sh.CellSize))

	result := buffer

	// Check all nearby cells
	for cellX := minCellX; cellX <= maxCellX; cellX++ {
		for cellY := minCellY; cellY <= maxCellY; cellY++ {
			key := sh.hashKey(cellX, cellY)
			if indices, ok := sh.Grid[key]; ok {
				result = append(result, indices...)
			}
		}
	}

	return result
}

// GetNeighbors returns particles within radius, with actual distance check.
func (sh *SpatialHash) GetNeighbors(particles []PhysicsParticle, x, y, radius float64) []int {
	return sh.GetNeighborsInto(particles, x, y, radius, nil, nil)
}

// GetNeighborsInto appends particles within radius to the neighbors buffer.
// candidateBuffer is used internally for QueryRadius results to avoid allocation.
// Both buffers should be reset ([:0]) before calling for zero-allocation operation.
// Returns the updated neighbors buffer.
func (sh *SpatialHash) GetNeighborsInto(particles []PhysicsParticle, x, y, radius float64, candidateBuffer, neighborsBuffer []int) []int {
	candidates := sh.QueryRadiusInto(x, y, radius, candidateBuffer)
	radiusSq := radius * radius

	neighbors := neighborsBuffer
	for _, idx := range candidates {
		if idx >= len(particles) {
			continue
		}
		p := &particles[idx]
		dx := p.X - x
		dy := p.Y - y
		distSq := dx*dx + dy*dy
		if distSq <= radiusSq {
			neighbors = append(neighbors, idx)
		}
	}

	return neighbors
}

// buildSpatialHashForSPH creates a spatial hash for efficient neighbor queries.
func buildSpatialHashForSPH(particles []PhysicsParticle, smoothingRadius float64) *SpatialHash {
	hash := NewSpatialHash(smoothingRadius, -1000, -1000, 1000, 1000)
	for i := range particles {
		hash.Insert(i, particles[i].X, particles[i].Y)
	}
	return hash
}

// computeDensityAndPressure calculates density and pressure for all particles.
func computeDensityAndPressure(particles []PhysicsParticle, hash *SpatialHash, config SPHConfig) {
	// Pre-allocate buffers for zero-allocation neighbor queries
	// Typical neighbor count is ~20-50 particles within smoothing radius
	candidateBuffer := make([]int, 0, 64)
	neighborsBuffer := make([]int, 0, 64)

	for i := range particles {
		p := &particles[i]
		// Reset buffers to reuse capacity
		candidateBuffer = candidateBuffer[:0]
		neighborsBuffer = neighborsBuffer[:0]
		neighbors := hash.GetNeighborsInto(particles, p.X, p.Y, config.SmoothingRadius, candidateBuffer, neighborsBuffer)

		density := 0.0
		for _, nIdx := range neighbors {
			if nIdx == i {
				density += config.ParticleMass * poly6Kernel(0, config.SmoothingRadius)
			} else {
				n := &particles[nIdx]
				dx := p.X - n.X
				dy := p.Y - n.Y
				r := math.Sqrt(dx*dx + dy*dy)
				density += config.ParticleMass * poly6Kernel(r, config.SmoothingRadius)
			}
		}

		p.Density = density
		p.Pressure = config.GasConstant * (density - config.RestDensity)
	}
}

// computePressureAndViscosityForces calculates pressure and viscosity forces for a particle.
func computePressureAndViscosityForces(p *PhysicsParticle, particles []PhysicsParticle, neighbors []int, i int, config SPHConfig) (float64, float64, float64, float64) {
	pressureForceX := 0.0
	pressureForceY := 0.0
	viscosityForceX := 0.0
	viscosityForceY := 0.0

	for _, nIdx := range neighbors {
		if nIdx == i {
			continue
		}
		n := &particles[nIdx]
		dx := p.X - n.X
		dy := p.Y - n.Y
		r := math.Sqrt(dx*dx + dy*dy)

		if r > 0.0001 && n.Density > 0.0001 {
			pressureTerm := (p.Pressure + n.Pressure) / (2.0 * n.Density)
			gradSpiky := spikyKernelGradient(r, config.SmoothingRadius)
			pressureForceX -= config.ParticleMass * pressureTerm * (dx / r) * gradSpiky
			pressureForceY -= config.ParticleMass * pressureTerm * (dy / r) * gradSpiky

			viscosityTerm := config.ParticleMass / n.Density
			lapViscosity := viscosityKernelLaplacian(r, config.SmoothingRadius)
			viscosityForceX += viscosityTerm * (n.VX - p.VX) * lapViscosity
			viscosityForceY += viscosityTerm * (n.VY - p.VY) * lapViscosity
		}
	}

	return pressureForceX, pressureForceY, viscosityForceX, viscosityForceY
}

// applySurfaceTension applies surface tension forces to a particle.
func applySurfaceTension(p *PhysicsParticle, particles []PhysicsParticle, neighbors []int, i int, config SPHConfig, deltaTime float64) {
	if config.SurfaceTension <= 0 || len(neighbors) <= 1 {
		return
	}

	cohesionX := 0.0
	cohesionY := 0.0
	for _, nIdx := range neighbors {
		if nIdx == i {
			continue
		}
		n := &particles[nIdx]
		dx := n.X - p.X
		dy := n.Y - p.Y
		r := math.Sqrt(dx*dx + dy*dy)
		if r > 0.0001 && r < config.SmoothingRadius {
			cohesionX += dx / r
			cohesionY += dy / r
		}
	}
	p.VX += cohesionX * config.SurfaceTension * deltaTime
	p.VY += cohesionY * config.SurfaceTension * deltaTime
}

// UpdateSPH updates fluid particles using SPH algorithm.
func UpdateSPH(particles []PhysicsParticle, config SPHConfig, deltaTime float64) {
	if len(particles) == 0 {
		return
	}

	hash := buildSpatialHashForSPH(particles, config.SmoothingRadius)
	computeDensityAndPressure(particles, hash, config)

	// Pre-allocate buffers for zero-allocation neighbor queries
	candidateBuffer := make([]int, 0, 64)
	neighborsBuffer := make([]int, 0, 64)

	for i := range particles {
		p := &particles[i]
		// Reset buffers to reuse capacity
		candidateBuffer = candidateBuffer[:0]
		neighborsBuffer = neighborsBuffer[:0]
		neighbors := hash.GetNeighborsInto(particles, p.X, p.Y, config.SmoothingRadius, candidateBuffer, neighborsBuffer)

		pressureForceX, pressureForceY, viscosityForceX, viscosityForceY := computePressureAndViscosityForces(p, particles, neighbors, i, config)

		if p.Density > 0 {
			accelX := pressureForceX/p.Density + config.Viscosity*viscosityForceX
			accelY := pressureForceY/p.Density + config.Viscosity*viscosityForceY
			p.VX += accelX * deltaTime
			p.VY += accelY * deltaTime
		}

		applySurfaceTension(p, particles, neighbors, i, config, deltaTime)
	}
}

// poly6Kernel computes the Poly6 SPH kernel function.
func poly6Kernel(r, h float64) float64 {
	if r >= h {
		return 0.0
	}
	hSq := h * h
	rSq := r * r
	term := hSq - rSq
	return (315.0 / (64.0 * math.Pi * math.Pow(h, 9))) * term * term * term
}

// spikyKernelGradient computes the gradient magnitude of the Spiky kernel.
func spikyKernelGradient(r, h float64) float64 {
	if r >= h {
		return 0.0
	}
	term := h - r
	return -(45.0 / (math.Pi * math.Pow(h, 6))) * term * term
}

// viscosityKernelLaplacian computes the Laplacian of the viscosity kernel.
func viscosityKernelLaplacian(r, h float64) float64 {
	if r >= h {
		return 0.0
	}
	return (45.0 / (math.Pi * math.Pow(h, 6))) * (h - r)
}

// UpdateFire updates fire particles with heat propagation.
func UpdateFire(particles []PhysicsParticle, config FireConfig, deltaTime float64, rng *rand.Rand) {
	if len(particles) == 0 || rng == nil {
		return
	}

	hash := buildHeatSpatialHash(particles, config)

	// Pre-allocate buffers for neighbor lookups (reused across particles)
	candidateBuffer := make([]int, 0, 64)
	neighborsBuffer := make([]int, 0, 64)

	for i := range particles {
		updateFireParticle(&particles[i], particles, config, deltaTime, rng, hash, i, candidateBuffer, neighborsBuffer)
	}
}

// buildHeatSpatialHash creates a spatial hash of all particles with heat.
func buildHeatSpatialHash(particles []PhysicsParticle, config FireConfig) *SpatialHash {
	hash := NewSpatialHash(config.HeatTransferRadius, -1000, -1000, 1000, 1000)
	for i := range particles {
		if particles[i].Heat > 0 {
			hash.Insert(i, particles[i].X, particles[i].Y)
		}
	}
	return hash
}

// updateFireParticle updates a single fire particle's heat, ignition, and movement.
func updateFireParticle(p *PhysicsParticle, allParticles []PhysicsParticle, config FireConfig,
	deltaTime float64, rng *rand.Rand, hash *SpatialHash, index int, candidateBuffer, neighborsBuffer []int,
) {
	dissipateHeat(p, config, deltaTime)
	updateIgnitionStatus(p, config)
	consumeFuel(p, config, deltaTime)
	transferHeatToNeighbors(p, allParticles, config, deltaTime, hash, index, candidateBuffer, neighborsBuffer)
	applyBuoyancy(p, config, deltaTime)
	spawnEmbers(p, config, deltaTime, rng)
}

// dissipateHeat reduces particle heat over time.
func dissipateHeat(p *PhysicsParticle, config FireConfig, deltaTime float64) {
	p.Heat -= config.HeatDissipation * deltaTime
	if p.Heat < 0 {
		p.Heat = 0
	}
}

// updateIgnitionStatus checks if particle should ignite based on heat.
func updateIgnitionStatus(p *PhysicsParticle, config FireConfig) {
	if p.Heat >= config.IgnitionTemp {
		p.Ignited = true
	}
}

// consumeFuel consumes fuel if particle is ignited.
func consumeFuel(p *PhysicsParticle, config FireConfig, deltaTime float64) {
	if p.Ignited && p.FuelRemain > 0 {
		p.FuelRemain -= config.FuelConsumptionRate * deltaTime
		if p.FuelRemain <= 0 {
			p.Ignited = false
			p.Heat *= config.ExtinguishHeatMultiplier
		}
	}
}

// transferHeatToNeighbors distributes heat to nearby particles.
// candidateBuffer and neighborsBuffer should be pre-allocated and reset before calling.
func transferHeatToNeighbors(p *PhysicsParticle, allParticles []PhysicsParticle, config FireConfig,
	deltaTime float64, hash *SpatialHash, index int, candidateBuffer, neighborsBuffer []int,
) []int {
	if p.Heat > config.MinActiveHeat {
		neighbors := hash.GetNeighborsInto(allParticles, p.X, p.Y, config.HeatTransferRadius, candidateBuffer, neighborsBuffer)
		for _, nIdx := range neighbors {
			if nIdx == index {
				continue
			}
			n := &allParticles[nIdx]
			heatTransfer := p.Heat * config.HeatTransferRate * deltaTime
			n.Heat += heatTransfer * config.HeatTransferFraction
		}
		return neighbors
	}
	return neighborsBuffer
}

// applyBuoyancy makes heated particles rise.
func applyBuoyancy(p *PhysicsParticle, config FireConfig, deltaTime float64) {
	if p.Heat > config.MinActiveHeat {
		p.VY -= config.BuoyancyStrength * p.Heat * deltaTime
	}
}

// spawnEmbers generates ember particles from ignited particles.
func spawnEmbers(p *PhysicsParticle, config FireConfig, deltaTime float64, rng *rand.Rand) {
	if p.Ignited && rng.Float64() < config.EmberChance*deltaTime {
		p.EmberTimer += deltaTime
	}
}

// UpdateSmoke updates smoke particles with turbulence.
func UpdateSmoke(particles []PhysicsParticle, config SmokeConfig, deltaTime, time float64) {
	for i := range particles {
		p := &particles[i]

		// Update turbulence phase
		p.TurbulencePhase += config.TurbulenceFreq * deltaTime

		// Apply turbulence using Perlin-like noise approximation
		// Use particle position and phase for deterministic turbulence
		noiseX := math.Sin(p.X*config.TurbulenceNoiseScale+p.TurbulencePhase) * math.Cos(p.Y*config.TurbulenceNoiseScale)
		noiseY := math.Cos(p.X*config.TurbulenceNoiseScale) * math.Sin(p.Y*config.TurbulenceNoiseScale+p.TurbulencePhase)

		turbulenceX := noiseX * config.TurbulenceStrength
		turbulenceY := noiseY * config.TurbulenceStrength

		p.VX += turbulenceX * deltaTime
		p.VY += turbulenceY * deltaTime

		// Apply rising motion
		p.VY -= config.RiseSpeed * deltaTime

		// Expand over time
		p.ExpansionTime += deltaTime
		p.Size += config.ExpansionRate * deltaTime

		// Dissipate (fade)
		p.Life -= config.Dissipation * deltaTime / p.InitialLife
	}
}

// UpdateDebris updates debris particles with collision detection.
func UpdateDebris(particles []PhysicsParticle, config DebrisConfig, deltaTime, groundY float64) {
	if len(particles) == 0 {
		return
	}

	// Build spatial hash for collision detection
	hash := NewSpatialHash(config.CollisionRadius*2, -1000, -1000, 1000, 1000)
	for i := range particles {
		hash.Insert(i, particles[i].X, particles[i].Y)
	}

	// Particle-particle collisions
	for i := range particles {
		p := &particles[i]

		// Find nearby particles for collision
		neighbors := hash.GetNeighbors(particles, p.X, p.Y, config.CollisionRadius*2)

		for _, nIdx := range neighbors {
			if nIdx <= i { // Avoid duplicate checks
				continue
			}
			n := &particles[nIdx]

			dx := n.X - p.X
			dy := n.Y - p.Y
			distSq := dx*dx + dy*dy
			minDist := config.CollisionRadius * 2

			if distSq < minDist*minDist && distSq > 0 {
				dist := math.Sqrt(distSq)

				// Normalize collision vector
				nx := dx / dist
				ny := dy / dist

				// Relative velocity
				dvx := n.VX - p.VX
				dvy := n.VY - p.VY

				// Velocity along collision normal
				dvn := dvx*nx + dvy*ny

				// Only resolve if particles are moving toward each other
				if dvn < 0 {
					// Impulse magnitude (equal mass assumed)
					impulse := -(1.0 + config.Restitution) * dvn / 2.0

					// Apply impulse
					p.VX -= impulse * nx
					p.VY -= impulse * ny
					n.VX += impulse * nx
					n.VY += impulse * ny

					// Separate particles to avoid overlap
					overlap := minDist - dist
					separationX := nx * overlap * 0.5
					separationY := ny * overlap * 0.5
					p.X -= separationX
					p.Y -= separationY
					n.X += separationX
					n.Y += separationY

					// Transfer angular velocity
					p.AngularVelocity += dvn * 10.0
					n.AngularVelocity -= dvn * 10.0

					p.LastCollision = 0
					n.LastCollision = 0
				}
			}
		}

		// Ground collision
		if p.Y > groundY && p.VY > 0 {
			p.Y = groundY
			p.VY = -p.VY * config.Restitution

			// Apply friction to horizontal velocity
			p.VX *= (1.0 - config.Friction*deltaTime)

			// Add rotation from impact
			p.AngularVelocity += p.VX * 0.5

			p.LastCollision = 0
		}

		// Apply angular damping
		p.AngularVelocity *= math.Pow(config.RotationDamping, deltaTime)

		// Update rotation
		p.Rotation += p.AngularVelocity * deltaTime

		// Update collision timer
		p.LastCollision += deltaTime
	}
}

// UpdateDebrisPooled updates debris particles using a pooled context for zero-allocation updates.
// Use AcquireDebrisContext before calling and ReleaseDebrisContext when done.
//
// This function is 2-3x faster than UpdateDebris due to eliminated allocations:
//   - No spatial hash allocation per call
//   - No neighbor buffer allocations per particle
//
// Example usage:
//
//	ctx := AcquireDebrisContext()
//	defer ReleaseDebrisContext(ctx)
//	for frame := 0; frame < 1000; frame++ {
//	    UpdateDebrisPooled(particles, config, deltaTime, groundY, ctx)
//	}
func UpdateDebrisPooled(particles []PhysicsParticle, config DebrisConfig, deltaTime, groundY float64, ctx *DebrisContext) {
	if len(particles) == 0 || ctx == nil {
		return
	}

	// Reset and configure spatial hash
	ctx.SpatialHash.Clear()
	ctx.SpatialHash.CellSize = config.CollisionRadius * 2

	// Build spatial hash for collision detection
	for i := range particles {
		ctx.SpatialHash.Insert(i, particles[i].X, particles[i].Y)
	}

	// Particle-particle collisions with reused buffers
	for i := range particles {
		p := &particles[i]

		// Reset buffers for each particle
		ctx.candidateBuffer = ctx.candidateBuffer[:0]
		ctx.neighborsBuffer = ctx.neighborsBuffer[:0]

		// Find nearby particles using pooled buffers
		neighbors := ctx.SpatialHash.GetNeighborsInto(
			particles, p.X, p.Y,
			config.CollisionRadius*2,
			ctx.candidateBuffer, ctx.neighborsBuffer,
		)

		for _, nIdx := range neighbors {
			if nIdx <= i { // Avoid duplicate checks
				continue
			}
			n := &particles[nIdx]

			dx := n.X - p.X
			dy := n.Y - p.Y
			distSq := dx*dx + dy*dy
			minDist := config.CollisionRadius * 2

			if distSq < minDist*minDist && distSq > 0 {
				dist := math.Sqrt(distSq)

				// Normalize collision vector
				nx := dx / dist
				ny := dy / dist

				// Relative velocity
				dvx := n.VX - p.VX
				dvy := n.VY - p.VY

				// Velocity along collision normal
				dvn := dvx*nx + dvy*ny

				// Only resolve if particles are moving toward each other
				if dvn < 0 {
					// Impulse magnitude (equal mass assumed)
					impulse := -(1.0 + config.Restitution) * dvn / 2.0

					// Apply impulse
					p.VX -= impulse * nx
					p.VY -= impulse * ny
					n.VX += impulse * nx
					n.VY += impulse * ny

					// Separate particles to avoid overlap
					overlap := minDist - dist
					separationX := nx * overlap * 0.5
					separationY := ny * overlap * 0.5
					p.X -= separationX
					p.Y -= separationY
					n.X += separationX
					n.Y += separationY

					// Transfer angular velocity
					p.AngularVelocity += dvn * 10.0
					n.AngularVelocity -= dvn * 10.0

					p.LastCollision = 0
					n.LastCollision = 0
				}
			}
		}

		// Ground collision
		if p.Y > groundY && p.VY > 0 {
			p.Y = groundY
			p.VY = -p.VY * config.Restitution

			// Apply friction to horizontal velocity
			p.VX *= (1.0 - config.Friction*deltaTime)

			// Add rotation from impact
			p.AngularVelocity += p.VX * 0.5

			p.LastCollision = 0
		}

		// Apply angular damping
		p.AngularVelocity *= math.Pow(config.RotationDamping, deltaTime)

		// Update rotation
		p.Rotation += p.AngularVelocity * deltaTime

		// Update collision timer
		p.LastCollision += deltaTime
	}
}
