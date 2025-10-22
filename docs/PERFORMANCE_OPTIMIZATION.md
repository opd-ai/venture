# Performance Optimization Guide

This document provides guidance on optimizing performance in the Venture game engine. It covers profiling, optimization techniques, and best practices for meeting the 60 FPS performance target.

## Performance Targets

The Venture engine has the following performance targets:

- **FPS:** 60 minimum on modest hardware
- **Frame Time:** <16.67ms per frame (for 60 FPS)
- **Memory:** <500MB client, <1GB server (4 players)
- **Generation Time:** <2 seconds for new world areas
- **Network Bandwidth:** <100KB/s per player at 20 updates/second

## Performance Monitoring

### Built-in Metrics

The engine includes a built-in performance monitoring system in `pkg/engine/performance.go`:

```go
// Create a performance monitor
world := engine.NewWorld()
monitor := engine.NewPerformanceMonitor(world)

// Update with monitoring
monitor.Update(deltaTime)

// Get metrics
metrics := monitor.GetMetrics()
fmt.Println(metrics.String())
fmt.Println(metrics.DetailedString())

// Check if meeting 60 FPS target
if !metrics.IsPerformanceTarget() {
    log.Println("Performance below target!")
}
```

### Metrics Available

The `PerformanceMetrics` struct tracks:

- **FPS:** Current frames per second
- **Frame Time:** Current, average, min, and max frame times
- **Update Time:** Time spent in world update
- **System Times:** Per-system breakdown of execution time
- **Entity Counts:** Total and active entity counts
- **Memory Stats:** Allocated and in-use memory (when sampled)

### Timing Code Sections

Use the `Timer` helper to measure specific code sections:

```go
timer := engine.NewTimer("terrain generation")
// ... code to time ...
elapsed := timer.Stop()

// Or with logging:
timer := engine.NewTimer("entity spawning")
// ... code to time ...
timer.StopAndLog() // Prints: [PERF] entity spawning: 5.23ms
```

## Spatial Partitioning

### Quadtree System

The engine includes a quadtree for efficient spatial queries in `pkg/engine/spatial_partition.go`:

```go
// Create spatial partition system
sps := engine.NewSpatialPartitionSystem(worldWidth, worldHeight)

// Add to world systems
world.AddSystem(sps)

// Query entities within radius
entities := sps.QueryRadius(x, y, radius)

// Query entities within bounds
bounds := engine.Bounds{X: 0, Y: 0, Width: 100, Height: 100}
entities := sps.QueryBounds(bounds)
```

### When to Use Spatial Partitioning

Use spatial queries instead of iterating all entities when:

- Finding enemies within attack range
- Finding items near the player
- Collision detection within a region
- AI perception checks (what can the entity see?)
- Rendering culling (what's visible on screen?)

### Performance Characteristics

- **Insert:** O(log n) average, O(n) worst case
- **Query:** O(log n + k) where k is number of results
- **Radius Query:** O(log n + k) with circle filtering
- **Rebuild:** O(n log n) - called periodically as entities move

The quadtree automatically rebuilds every 60 frames (1 second at 60 FPS) to account for entity movement.

## Optimization Techniques

### 1. Reduce Allocations

**Problem:** Frequent allocations cause garbage collection pauses.

**Solutions:**

- **Entity List Caching:** The World now caches the entity list instead of rebuilding it every frame:
  ```go
  // Old (allocates every frame):
  entities := make([]*Entity, 0, len(w.entities))
  for _, e := range w.entities { entities = append(entities, e) }
  
  // New (cached, reused):
  return w.cachedEntityList
  ```

- **Object Pooling:** Reuse objects instead of creating new ones:
  ```go
  type ParticlePool struct {
      particles []*Particle
      nextIndex int
  }
  
  func (p *ParticlePool) Get() *Particle {
      if p.nextIndex < len(p.particles) {
          particle := p.particles[p.nextIndex]
          p.nextIndex++
          return particle
      }
      return &Particle{} // Create new only when pool exhausted
  }
  
  func (p *ParticlePool) ReturnAll() {
      p.nextIndex = 0 // Reset without freeing
  }
  ```

- **Pre-allocate Slices:** If you know the capacity, pre-allocate:
  ```go
  // Good:
  entities := make([]*Entity, 0, expectedCount)
  
  // Bad:
  entities := []*Entity{}
  ```

### 2. Avoid Unnecessary Work

**Component Type Checks:**

Cache component type strings:
```go
const (
    ComponentTypePosition = "position"
    ComponentTypeVelocity = "velocity"
    ComponentTypeHealth   = "health"
)

// Use constants instead of string literals
entity.GetComponent(ComponentTypePosition)
```

**Early Returns:**

Exit loops and functions as soon as possible:
```go
// Good:
for _, entity := range entities {
    if !entity.HasComponent("position") {
        continue // Skip immediately
    }
    // Process entity...
}

// Bad:
for _, entity := range entities {
    if entity.HasComponent("position") {
        // Process entity...
    }
}
```

**Lazy Evaluation:**

Don't calculate values until needed:
```go
// Good:
if entity.HasComponent("health") {
    health := entity.GetComponent("health").(*HealthComponent)
    if health.Current <= 0 {
        // Only calculate distance if entity is dead
        distance := calculateDistance(player, entity)
    }
}
```

### 3. Use Efficient Algorithms

**Distance Calculations:**

Use squared distance to avoid expensive sqrt():
```go
// For comparison, use squared distance:
distSq := engine.DistanceSquared(x1, y1, x2, y2)
if distSq < radiusSq { // Compare with squared radius
    // Entity in range
}

// Only calculate actual distance when needed:
distance := math.Sqrt(distSq)
```

**Component Queries:**

Use `GetEntitiesWith()` for filtered queries:
```go
// Efficient - only returns entities with both components:
entities := world.GetEntitiesWith("position", "velocity")

// Less efficient - checks all entities:
for _, entity := range world.GetEntities() {
    if entity.HasComponent("position") && entity.HasComponent("velocity") {
        // ...
    }
}
```

### 4. Optimize Hot Paths

**Identify Hot Paths:**

Use profiling to find code that runs frequently:
```bash
go test -tags test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
# Type 'top' to see top functions by CPU time
```

**Common Hot Paths:**

- Game loop update (called 60 times per second)
- Rendering (called 60 times per second)
- Collision detection (checks many entity pairs)
- AI pathfinding (expensive algorithm)
- Component queries (searches entity lists)

### 5. Parallelize Where Possible

**Use Goroutines for Independent Work:**

```go
// Parallelize independent system updates
var wg sync.WaitGroup
for _, system := range independentSystems {
    wg.Add(1)
    go func(s System) {
        defer wg.Done()
        s.Update(entities, deltaTime)
    }(system)
}
wg.Wait()
```

**Caution:** Only parallelize truly independent systems. Shared state requires careful synchronization.

## Profiling

### CPU Profiling

```bash
# Profile tests
go test -tags test -cpuprofile=cpu.prof -bench=BenchmarkWorld ./pkg/engine
go tool pprof cpu.prof

# Profile the client (add to main.go):
import "runtime/pprof"
f, _ := os.Create("cpu.prof")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()
```

In pprof:
- `top`: Show top functions by CPU time
- `list <function>`: Show annotated source for function
- `web`: Generate visual call graph (requires graphviz)

### Memory Profiling

```bash
# Profile memory allocations
go test -tags test -memprofile=mem.prof -bench=BenchmarkWorld ./pkg/engine
go tool pprof mem.prof

# Profile memory in use
go test -tags test -memprofile=mem.prof -memprofilerate=1 -bench=.
```

### Benchmark Tests

Add benchmark tests for performance-critical code:

```go
func BenchmarkSystemUpdate(b *testing.B) {
    world := NewWorld()
    system := NewMovementSystem()
    
    // Setup
    for i := 0; i < 1000; i++ {
        entity := world.CreateEntity()
        entity.AddComponent(&PositionComponent{X: 0, Y: 0})
        entity.AddComponent(&VelocityComponent{VX: 1, VY: 1})
    }
    world.Update(0)
    entities := world.GetEntities()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        system.Update(entities, 0.016)
    }
}
```

Run benchmarks:
```bash
go test -tags test -bench=. -benchmem ./pkg/engine
```

## Common Performance Issues

### Issue: Low FPS

**Symptoms:**
- `metrics.FPS < 60`
- Frame time > 16.67ms

**Diagnosis:**
```go
metrics := monitor.GetMetrics()
percentages := metrics.GetFrameTimePercent()
for name, percent := range percentages {
    fmt.Printf("%s: %.2f%%\n", name, percent)
}
```

**Solutions:**
1. Identify the slowest system
2. Profile that system
3. Apply optimizations (see above)
4. Use spatial partitioning for entity queries
5. Reduce entity count if necessary

### Issue: High Memory Usage

**Symptoms:**
- Memory > 500MB on client
- Frequent garbage collection

**Diagnosis:**
```bash
go test -tags test -memprofile=mem.prof -bench=.
go tool pprof -alloc_space mem.prof
# Type 'top' to see top allocators
```

**Solutions:**
1. Use object pooling
2. Reduce slice allocations
3. Clear unused data structures
4. Use pointers for large structs
5. Profile and fix specific allocators

### Issue: Stuttering/Hitching

**Symptoms:**
- Occasional frame drops
- Max frame time >> average frame time

**Causes:**
- Garbage collection pauses
- Expensive operations (generation, pathfinding)
- Blocking I/O

**Solutions:**
1. Reduce allocations to reduce GC pressure
2. Spread expensive operations over multiple frames
3. Use async operations for I/O
4. Monitor `metrics.MaxFrameTime` to identify spikes

## Best Practices

### Do:
✅ Profile before optimizing (measure, don't guess)
✅ Use spatial partitioning for proximity queries
✅ Cache expensive calculations
✅ Pre-allocate slices with known capacity
✅ Use squared distance for comparisons
✅ Return early from functions and loops
✅ Run benchmarks to verify optimizations
✅ Monitor performance metrics in production

### Don't:
❌ Optimize without profiling first
❌ Iterate all entities when spatial queries would work
❌ Allocate in hot paths (game loop, rendering)
❌ Use reflection in performance-critical code
❌ Calculate expensive values that won't be used
❌ Ignore performance metrics
❌ Assume an optimization helps without measuring

## Performance Checklist

Before considering performance optimization complete:

- [ ] All systems meet frame time budget (< 16.67ms total)
- [ ] FPS consistently at or above 60 on target hardware
- [ ] Memory usage under target (<500MB client)
- [ ] No stuttering or frame drops during normal gameplay
- [ ] Generation completes in <2 seconds
- [ ] Benchmarks exist for all performance-critical paths
- [ ] Profiling data confirms no obvious bottlenecks
- [ ] Performance metrics are monitored in production

## Example: Optimized Entity Search

**Before (O(n) - iterates all entities):**
```go
func FindNearbyEnemies(world *World, x, y, radius float64) []*Entity {
    enemies := make([]*Entity, 0)
    for _, entity := range world.GetEntities() {
        if !entity.HasComponent("team") || !entity.HasComponent("position") {
            continue
        }
        pos := entity.GetComponent("position").(*PositionComponent)
        dist := Distance(x, y, pos.X, pos.Y)
        if dist <= radius {
            enemies = append(enemies, entity)
        }
    }
    return enemies
}
```

**After (O(log n + k) - uses spatial partitioning):**
```go
func FindNearbyEnemies(sps *SpatialPartitionSystem, world *World, 
                       x, y, radius float64, playerTeam int) []*Entity {
    // Query only entities in radius (spatial partition)
    candidates := sps.QueryRadius(x, y, radius)
    
    enemies := make([]*Entity, 0, len(candidates))
    for _, entity := range candidates {
        // Check team (most will be filtered here)
        teamComp, ok := entity.GetComponent("team")
        if !ok {
            continue
        }
        team := teamComp.(*TeamComponent)
        if team.TeamID == playerTeam {
            continue // Same team, not an enemy
        }
        
        enemies = append(enemies, entity)
    }
    return enemies
}
```

**Performance Improvement:**
- With 1000 entities, 50 in range: 1000 checks → ~50 checks
- With 10000 entities, 50 in range: 10000 checks → ~50 checks
- 20-200x faster depending on entity count and query radius

## Resources

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Performance Tips](https://github.com/dgryski/go-perfbook)
- [Profiling Go Programs](https://blog.golang.org/profiling-go-programs)
- [High Performance Go Workshop](https://dave.cheney.net/high-performance-go-workshop/gopherchina-2019.html)

## Monitoring in Production

Add performance monitoring to your game:

```go
// In main loop
if frameCount % 60 == 0 { // Every second
    metrics := monitor.GetMetrics()
    if !metrics.IsPerformanceTarget() {
        log.Printf("PERFORMANCE WARNING: %s", metrics.String())
    }
}

// Log detailed stats periodically
if frameCount % 600 == 0 { // Every 10 seconds
    log.Printf("\n%s", metrics.DetailedString())
}
```

This helps identify performance issues in the wild and track performance improvements over time.
