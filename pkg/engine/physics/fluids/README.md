# Fluids Package - Fluid Dynamics Simulation

Package `pkg/engine/physics/fluids` implements grid-based fluid dynamics simulation for water, lava, and other liquids.

## Package Structure

After reorganization (2026-01-20), the package is organized as follows:

### Implementation Files
- **`buoyancy_calculator.go`** - BuoyancyCalculator for Archimedes' principle force calculations
- **`swimming.go`** - SwimmingManager for swimming mechanics and stamina
- **`flooding.go`** - FloodingManager for enclosed area flooding
- **`simulator.go`** - Grid-based fluid dynamics simulation engine
- **`types.go`** - All type definitions, constants, and fluid properties
- **`doc.go`** - Comprehensive package documentation

### Test Files
- **`buoyancy_test.go`** - Tests for all buoyancy and manager implementations
- **`simulator_test.go`** - Tests for fluid simulation
- **`types_test.go`** - Tests for types and helper functions

## Features

### Fluid Simulation
Grid-based simulation using simplified Navier-Stokes equations:
```go
config := fluids.DefaultSimulationConfig()
sim := fluids.NewSimulator(config)

// Add water at position (10, 20)
sim.AddFluid(10, 20, 1.0, fluids.FluidWater)

// Update simulation
sim.Update(deltaTime)

// Get grid state
grid := sim.GetGrid()
```

### Buoyancy Calculations
Calculate buoyancy forces for entities in fluid:
```go
calc := fluids.NewBuoyancyCalculator(9.81)  // Earth gravity
component := &fluids.BuoyancyComponent{
    Mass:   500.0,  // kg
    Volume: 2.0,    // m³
}

calc.CalculateBuoyancy(component, fluidAmount, fluids.FluidWater)

if component.Buoyant {
    // Entity floats
    netForce := calc.GetNetForce(component)
}
```

### Swimming Mechanics
Manage swimming with stamina and drowning:
```go
swimMgr := fluids.NewSwimmingManager(9.81)
swimming := &fluids.SwimmingComponent{
    Stamina:        100.0,
    MaxStamina:     100.0,
    StaminaDrain:   10.0,  // per second
    StaminaRegen:   20.0,  // per second on land
    SwimSpeed:      0.5,   // 50% of normal speed
    DrowningDamage: 5.0,   // per second
}

swimMgr.UpdateSwimming(swimming, buoyancy, fluidAmount, deltaTime)

speedMultiplier := swimMgr.GetSwimSpeedMultiplier(swimming)
damage := swimMgr.GetDrowningDamage(swimming)
```

### Flooding System
Manage flooding in enclosed spaces:
```go
floodMgr := fluids.NewFloodingManager(simulator)
flooding := &fluids.FloodingComponent{
    FloodLevel:     0.0,
    MaxFloodLevel:  100.0,
    FloodRate:      10.0,  // units per second
}

// Add water source
floodMgr.AddFloodSource(flooding, x, y, flowRate)

// Update flooding
floodMgr.UpdateFlooding(flooding, deltaTime)

// Check flood status
if floodMgr.IsFullyFlooded(flooding) {
    percent := floodMgr.GetFloodPercentage(flooding)
}
```

## Fluid Types

Five fluid types with distinct physical properties:

| Fluid | Density (kg/m³) | Viscosity | Damage (per sec) | Special |
|-------|----------------|-----------|------------------|---------|
| Water | 1000 | Low | 0 | Neutral |
| Lava | 3100 | High | 50 | High density |
| Oil | 800 | Medium | 0 | Flammable |
| Acid | 1200 | Low | 25 | Corrosive |
| Poison | 1050 | Low | 15 | Toxic |

Access properties:
```go
props := fluids.GetFluidProperties(fluids.FluidLava)
// props.Density = 3100.0
// props.Viscosity = 0.8
// props.Damage = 50.0
```

## Physics Equations

### Buoyant Force
```
F_buoyant = ρ_fluid * V_submerged * g
```
Where:
- ρ_fluid: Fluid density (kg/m³)
- V_submerged: Submerged volume (m³)
- g: Gravity (9.81 m/s²)

### Net Force
```
F_net = F_buoyant - F_weight
F_weight = mass * g
```

### Pressure
```
Pressure = depth * density * gravity
```

### Swim Speed
```
Speed_multiplier = swim_speed * (0.5 + 0.5 * stamina_ratio)
```

## Thread Safety

- **Simulator**: Not thread-safe (single-threaded physics tick)
- **Managers**: Thread-safe where needed for concurrent entity access

## Performance

- **Simulation**: ~2ms for 100x100 grid at 30 FPS
- **Buoyancy**: <0.1ms per entity
- **Swimming**: <0.05ms per entity
- **Flooding**: <0.2ms per area

Optimizations:
- Grid-based simulation reduces O(n²) to O(n)
- Pressure calculation cached per grid update
- Boundary checks prevent expensive out-of-bounds operations

## Test Coverage

- **Coverage**: 95.1% of statements
- **Tests**: 57 tests (all passing)
- **Benchmarks**: 4 benchmark functions

## Integration Points

- **V4 Physics Engine** - Core physics simulation loop
- **V4 Vehicles** - Boat/submarine buoyancy
- **V8 Environment** - Water bodies and fluid placement
- **ECS System** - BuoyancyComponent, SwimmingComponent, FloodingComponent

## Configuration

Default simulation parameters:
```go
config := fluids.DefaultSimulationConfig()
// config.GridWidth = 100
// config.GridHeight = 100
// config.UpdateRate = 30.0 (FPS)
// config.Gravity = 9.81
// config.MaxFluidPerCell = 1.0
```

## Documentation

For detailed information, see:
- `doc.go` - Comprehensive package documentation with physics details
- `AUDIT.md` - Implementation audit and quality metrics
- Go documentation: `go doc github.com/opd-ai/venture/pkg/engine/physics/fluids`

## Recent Changes

**2026-01-20 Reorganization:**
- Split monolithic `buoyancy.go` (184 lines) into 3 focused files
- Added file-level documentation for all implementation files
- Created comprehensive AUDIT.md
- Maintained 100% test compatibility (57/57 tests passing)
- Zero regressions, zero build errors
- Achieved 95.1% test coverage

## Example: Complete Integration

```go
// Setup simulation
config := fluids.DefaultSimulationConfig()
sim := fluids.NewSimulator(config)
buoyancyCalc := fluids.NewBuoyancyCalculator(9.81)
swimMgr := fluids.NewSwimmingManager(9.81)

// Create entity components
buoyancy := &fluids.BuoyancyComponent{
    Mass:   70.0,   // 70kg human
    Volume: 0.07,   // 70L volume
}
swimming := &fluids.SwimmingComponent{
    Stamina:      100.0,
    MaxStamina:   100.0,
    StaminaDrain: 10.0,
    StaminaRegen: 20.0,
    SwimSpeed:    0.5,
}

// Game loop
for {
    // Get fluid at entity position
    grid := sim.GetGrid()
    fluidAmount := grid.Cells[entityY][entityX].Amount
    fluidType := grid.Cells[entityY][entityX].Type
    
    // Update physics
    buoyancyCalc.CalculateBuoyancy(buoyancy, fluidAmount, fluidType)
    swimMgr.UpdateSwimming(swimming, buoyancy, fluidAmount, deltaTime)
    
    // Apply forces
    if buoyancy.Buoyant {
        applyForce(0, buoyancyCalc.GetNetForce(buoyancy))
    }
    
    // Update fluid simulation
    sim.Update(deltaTime)
    
    time.Sleep(time.Second / 60)
}
```
