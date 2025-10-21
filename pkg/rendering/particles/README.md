# Particles Package

The `particles` package provides procedural particle effect generation for the Venture game. All particle effects are generated at runtime using deterministic algorithms.

## Features

- **6 Particle Types**: Spark, Smoke, Magic, Flame, Blood, and Dust
- **Deterministic Generation**: Same seed produces identical particle patterns
- **Genre-Aware**: Uses genre-specific color palettes
- **Physics Simulation**: Particles support velocity, gravity, rotation
- **Lifecycle Management**: Automatic particle aging and death
- **High Performance**: Optimized for thousands of particles per second

## Particle Types

### Spark
Bright, quick-moving particles perfect for impacts and explosions.
- **Use cases**: Sword strikes, gun impacts, explosions
- **Behavior**: Fast, radial expansion
- **Colors**: Yellow, orange, white

### Smoke
Soft, slowly rising particles for atmospheric effects.
- **Use cases**: Fire smoke, dust clouds, fog
- **Behavior**: Slow upward drift with rotation
- **Colors**: Grays with transparency

### Magic
Glowing particles using genre-appropriate colors.
- **Use cases**: Spell casting, magical auras, enchantments
- **Behavior**: Swirling motion with medium speed
- **Colors**: From genre palette (primary, secondary)

### Flame
Fire-like particles with upward motion and color gradient.
- **Use cases**: Torches, fire spells, burning objects
- **Behavior**: Rising with slight randomness
- **Colors**: Orange, yellow, red gradient

### Blood
Splatter particles for combat effects.
- **Use cases**: Damage indicators, death effects
- **Behavior**: Radial expansion with gravity
- **Colors**: Dark red tones

### Dust
Small, light particles for environmental effects.
- **Use cases**: Footsteps, wind, debris
- **Behavior**: Slow, gentle motion
- **Colors**: Earthy tones with transparency

## Usage

### Basic Generation

```go
import "github.com/opd-ai/venture/pkg/rendering/particles"

// Create generator
gen := particles.NewGenerator()

// Configure particle system
config := particles.Config{
    Type:     particles.ParticleSpark,
    Count:    50,
    GenreID:  "fantasy",
    Seed:     12345,
    Duration: 1.0,
    SpreadX:  10.0,
    SpreadY:  10.0,
    Gravity:  0.0,
    MinSize:  1.0,
    MaxSize:  3.0,
}

// Generate particle system
system, err := gen.Generate(config)
if err != nil {
    log.Fatal(err)
}
```

### Updating Particles

```go
// In game loop
deltaTime := 0.016 // ~60 FPS

// Update all particles
system.Update(deltaTime)

// Check if any particles are still alive
if system.IsAlive() {
    // Render particles
    for _, p := range system.GetAliveParticles() {
        // Draw particle at (p.X, p.Y) with p.Color and p.Size
        // Use p.Life for fade effects (0.0 = fully faded, 1.0 = full opacity)
    }
}
```

### Advanced Configuration

```go
// Explosion effect with gravity
config := particles.Config{
    Type:     particles.ParticleSpark,
    Count:    200,
    GenreID:  "scifi",
    Seed:     time.Now().UnixNano(), // Random seed
    Duration: 0.5,
    SpreadX:  20.0,
    SpreadY:  20.0,
    Gravity:  50.0, // Strong downward force
    MinSize:  0.5,
    MaxSize:  2.5,
}

// Rising smoke
config := particles.Config{
    Type:     particles.ParticleSmoke,
    Count:    100,
    GenreID:  "postapoc",
    Seed:     12345,
    Duration: 3.0,
    SpreadX:  3.0,
    SpreadY:  15.0, // More vertical spread
    Gravity:  -2.0, // Negative = upward
    MinSize:  2.0,
    MaxSize:  6.0,
}
```

## Configuration Parameters

### Required Parameters

- **Type**: ParticleType (Spark, Smoke, Magic, Flame, Blood, Dust)
- **Count**: Number of particles (1-10000)
- **GenreID**: Genre for color selection ("fantasy", "scifi", etc.)
- **Duration**: Particle lifetime in seconds
- **MinSize/MaxSize**: Particle size range in pixels

### Optional Parameters

- **Seed**: Random seed (default: 0)
- **SpreadX/SpreadY**: Initial velocity spread (default: 5.0)
- **Gravity**: Vertical acceleration (default: 0.0)
  - Positive values = downward
  - Negative values = upward
- **Custom**: Map for additional parameters

## Performance

Particle generation is highly optimized:
- **Generation**: <1ms for systems with 1000 particles
- **Update**: ~0.1ms for 1000 particles at 60 FPS
- **Memory**: ~100 bytes per particle

Benchmarks (on typical hardware):
```
BenchmarkGenerator_Generate-8     100000    10523 ns/op    (100 particles)
BenchmarkParticleSystem_Update-8  50000     25841 ns/op    (1000 particles)
```

## Testing

Run tests with the test build tag:

```bash
go test -tags test ./pkg/rendering/particles/...
```

### Test Coverage

Current coverage: **100%** of statements

## Determinism

All particle generation is deterministic when using the same seed:

```go
config := particles.Config{
    Type:    particles.ParticleMagic,
    Count:   100,
    Seed:    12345, // Fixed seed
    // ... other parameters
}

system1, _ := gen.Generate(config)
system2, _ := gen.Generate(config)

// system1 and system2 will have identical particles
```

This is essential for:
- Multiplayer synchronization
- Replay systems
- Debugging and testing

## Integration with Game Systems

### With ECS

```go
// Create particle component
type ParticleComponent struct {
    System *particles.ParticleSystem
}

func (c *ParticleComponent) Type() string {
    return "particle"
}

// In rendering system
func (s *RenderingSystem) Update(deltaTime float64, world *engine.World) {
    for _, entity := range world.GetEntitiesWithComponent("particle") {
        comp := entity.GetComponent("particle").(*ParticleComponent)
        comp.System.Update(deltaTime)
        
        // Remove entity when particles are dead
        if !comp.System.IsAlive() {
            world.RemoveEntity(entity.ID)
        }
    }
}
```

### With Combat System

```go
// Spawn particles on hit
func spawnHitEffect(x, y float64, damageType string) {
    var pType particles.ParticleType
    switch damageType {
    case "physical":
        pType = particles.ParticleBlood
    case "fire":
        pType = particles.ParticleFlame
    case "magic":
        pType = particles.ParticleMagic
    }
    
    config := particles.Config{
        Type:    pType,
        Count:   30,
        GenreID: currentGenre,
        Seed:    generateSeed(),
        // ... other params
    }
    
    system, _ := particleGen.Generate(config)
    // Add to game world at (x, y)
}
```

## API Reference

### Types

- `ParticleType`: Enum for particle types
- `Config`: Configuration for particle generation
- `Particle`: Individual particle data
- `ParticleSystem`: Collection of particles with update logic
- `Generator`: Factory for creating particle systems

### Key Methods

- `NewGenerator() *Generator`: Create a new particle generator
- `Generate(config Config) (*ParticleSystem, error)`: Generate particle system
- `Update(deltaTime float64)`: Update all particles
- `IsAlive() bool`: Check if any particles are alive
- `GetAliveParticles() []Particle`: Get only living particles
- `Validate(result interface{}) error`: Validate particle system

## Examples

See `/examples/` directory for complete examples:
- `particle_effects.go`: Showcase of all particle types
- `particle_integration.go`: Integration with ECS and rendering

## Contributing

When adding new particle types:
1. Add enum to `ParticleType` in `types.go`
2. Implement generation in `generator.go`
3. Add tests in `generator_test.go`
4. Update this README with usage examples

Maintain:
- Deterministic generation
- 100% test coverage
- Performance targets (<1ms per 1000 particles)
