# Shadow System Documentation

## Overview

The Shadow System is part of Phase 14.1 (Enhanced Lighting & Shadows) and provides realistic shadow casting, ambient occlusion, and depth perception for the Venture game engine. It integrates seamlessly with the existing LightingSystem to create atmospheric lighting effects.

## Components

### ShadowComponent

Marks an entity as casting shadows when illuminated by light sources.

**Fields:**
- `Enabled`: Toggle shadow casting on/off
- `ShadowType`: Type of shadow rendering (Hard, Soft, Contact)
- `Opacity`: Shadow darkness (0.0 = invisible, 1.0 = opaque)
- `Radius`: Entity's shadow-casting radius in pixels
- `Height`: Affects shadow size/position for contact shadows
- `CastsShadow`: Whether this entity blocks light
- `ReceivesShadow`: Whether shadows are rendered on this entity
- `SoftEdgeRadius`: Penumbra size for soft shadows (pixels)
- `Color`: Shadow color tint (usually black or dark gray)

**Shadow Types:**
- **Hard Shadows**: Sharp edges, fastest rendering
- **Soft Shadows**: Gradual edge transition with penumbra (future enhancement)
- **Contact Shadows**: Small shadows at ground contact points

**Usage Example:**
```go
// Add shadow to entity
entity := world.CreateEntity()
entity.AddComponent(&PositionComponent{X: 400, Y: 300})
shadow := engine.NewShadowComponent(16) // 16 pixel radius
entity.AddComponent(shadow)

// Soft shadow
softShadow := engine.NewSoftShadow(20, 4) // radius 20, soft edge 4
entity.AddComponent(softShadow)

// Contact shadow (for entities touching ground)
contactShadow := engine.NewContactShadow(24, 8) // radius 24, height 8
entity.AddComponent(contactShadow)
```

### AmbientOcclusionComponent

Creates subtle darkening in corners and at contact points for depth perception.

**Fields:**
- `Enabled`: Toggle ambient occlusion on/off
- `Intensity`: Darkening strength (0.0-1.0)
- `Radius`: How far occlusion extends in pixels
- `Samples`: Quality vs performance (4-16 typical)
- `CornerDarkening`: Extra darkening at convex corners
- `CornerAmount`: Corner darkening intensity (0.0-1.0)

**Usage Example:**
```go
// Add ambient occlusion to entity
ao := engine.NewAmbientOcclusionComponent(0.5, 32) // intensity 0.5, radius 32
entity.AddComponent(ao)

// Adjust settings
ao.CornerDarkening = true
ao.CornerAmount = 0.7
ao.Samples = 12 // Higher quality
```

## Systems

### ShadowSystem

Processes shadow-casting entities and renders shadows based on light sources.

**Key Methods:**
- `NewShadowSystem(world)`: Create shadow system
- `SetEnabled(enabled)`: Toggle shadow rendering
- `SetMaxShadows(max)`: Limit shadow count per frame
- `SetRenderQuality(quality)`: Set rendering quality (0.5-2.0)
- `SetViewport(x, y, width, height)`: Set viewport for culling
- `RenderShadows(screen, lightX, lightY, lightRadius)`: Render shadows from a light
- `RenderAmbientOcclusion(screen)`: Render ambient occlusion

**Usage Example:**
```go
// Create shadow system
shadowSystem := engine.NewShadowSystem(world)

// Configure
shadowSystem.SetMaxShadows(100)
shadowSystem.SetRenderQuality(1.0) // Normal quality
shadowSystem.SetViewport(cameraX, cameraY, 800, 600)

// Render shadows for each light source
for _, light := range lightSources {
    shadowBuffer := shadowSystem.RenderShadows(screen, light.X, light.Y, light.Radius)
    // Composite shadowBuffer with scene
}

// Render ambient occlusion
shadowSystem.RenderAmbientOcclusion(screen)
```

### LightingSystem Integration

The LightingSystem automatically creates and manages a ShadowSystem when shadows are enabled.

**Configuration:**
```go
config := engine.NewLightingConfig()
config.ShadowsEnabled = true
config.ShadowOpacity = 0.6
config.ShadowQuality = 1.0
config.MaxShadows = 100

lightingSystem := engine.NewLightingSystem(world, config)

// Access shadow system
shadowSystem := lightingSystem.GetShadowSystem()

// Toggle shadows at runtime
lightingSystem.EnableShadows(true)
```

## Genre-Specific Settings

Shadow appearance varies by genre for atmospheric consistency:

| Genre | Shadow Opacity | Max Lights | Ambient Intensity | Visual Style |
|-------|---------------|------------|-------------------|--------------|
| Fantasy | 0.5 | 20 | 0.4 | Medium, warm shadows |
| Sci-Fi | 0.6 | 24 | 0.35 | Sharp, technical shadows |
| Horror | 0.8 | 12 | 0.15 | Deep, oppressive shadows |
| Cyberpunk | 0.7 | 28 | 0.25 | Strong neon shadows |
| Post-Apocalyptic | 0.4 | 16 | 0.3 | Diffuse, dusty shadows |

**Apply Genre Preset:**
```go
config := engine.NewLightingConfig()
config.SetGenrePreset("horror") // Applies horror-specific settings
```

## Performance Considerations

### Optimization Strategies

1. **Viewport Culling**: Shadows outside the viewport are automatically skipped
2. **Shadow Limits**: Set `MaxShadows` based on target hardware (50-200 typical)
3. **Quality Scaling**: Use `ShadowQuality` to balance visuals vs performance
   - 0.5 = half resolution (4x faster)
   - 1.0 = normal resolution
   - 2.0 = super-sampled (4x slower)
4. **Shadow Type Selection**: Hard shadows are fastest, soft shadows require more processing

### Performance Targets

- **60 FPS Target**: Maintain with 100 shadow casters at quality 1.0
- **Memory Usage**: <10MB for shadow buffers
- **Frame Time**: <2ms for shadow rendering with 100 entities

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=BenchmarkShadowSystem ./pkg/engine
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=BenchmarkShadowSystem ./pkg/engine
go tool pprof mem.prof
```

## Implementation Details

### Shadow Casting Algorithm

1. **Light Source Detection**: Find all lights in range
2. **Entity Collection**: Gather entities with ShadowComponent
3. **Distance Check**: Skip entities outside light radius
4. **Viewport Culling**: Skip entities outside camera view
5. **Ray Casting**: Calculate shadow direction from light position
6. **Shadow Rendering**: Draw shadow based on type (hard/soft/contact)

### Shadow Types Detail

**Hard Shadows:**
- Use simple ray-casting from light to entity
- Shadow extends away from light at fixed length
- Rendered as stretched rectangle/ellipse
- Fastest performance

**Soft Shadows:**
- Currently uses hard shadow with reduced opacity
- Future: Implement penumbra gradients
- Requires blur/gradient rendering

**Contact Shadows:**
- Small elliptical shadows at entity's "feet"
- Offset slightly based on light direction
- Provides ground contact visual cue
- Performance: similar to hard shadows

## Testing

Shadow components and logic are fully tested. Rendering tests require graphics context (DISPLAY):

```bash
# Component tests (no DISPLAY needed)
go test ./pkg/engine -run TestShadowComponent -v
go test ./pkg/engine -run TestAmbientOcclusion -v

# System tests (requires DISPLAY)
go test ./pkg/engine -run TestShadowSystem

# All tests with short mode (skips rendering)
go test ./pkg/engine -short
```

Test Coverage: 85%+ on testable code (excluding Ebiten rendering functions)

## Integration with Game

### Adding Shadows to Entities

```go
// Player shadow
player := world.CreateEntity()
player.AddComponent(&PositionComponent{X: 400, Y: 300})
player.AddComponent(engine.NewShadowComponent(14)) // Player sprite radius

// Enemy shadow
enemy := world.CreateEntity()
enemy.AddComponent(&PositionComponent{X: 500, Y: 400})
enemyShadow := engine.NewShadowComponent(16)
enemyShadow.ShadowType = engine.ShadowTypeContact // Contact shadow
enemy.AddComponent(enemyShadow)

// Boss with AO
boss := world.CreateEntity()
boss.AddComponent(&PositionComponent{X: 600, Y: 500})
boss.AddComponent(engine.NewShadowComponent(32)) // Large shadow
boss.AddComponent(engine.NewAmbientOcclusionComponent(0.6, 48)) // Strong AO
```

### Render Pipeline Integration

```go
func (g *Game) Draw(screen *ebiten.Image) {
    // 1. Render base scene
    g.renderTerrain(screen)
    g.renderEntities(screen)
    
    // 2. Apply lighting and shadows
    if g.lightingSystem != nil {
        // Shadow system is managed by lighting system
        g.lightingSystem.ApplyLighting(screen)
    }
    
    // 3. Render UI on top
    g.renderUI(screen)
}
```

## Future Enhancements

### Planned Features

1. **Soft Shadow Penumbra**: Implement proper gradient-based soft shadows
2. **Shadow Caching**: Cache static shadow shapes for stationary entities
3. **Shadow Maps**: 2D shadow map generation for more accurate shadows
4. **Colored Shadows**: Support tinted shadows for stained glass effects
5. **Dynamic Shadow Blending**: Multiple overlapping shadows blend naturally
6. **Self-Shadowing**: Entities can cast shadows on themselves

### Performance Improvements

1. **GPU Acceleration**: Investigate shader-based shadow rendering
2. **Spatial Partitioning**: Use quadtree for faster shadow caster queries
3. **LOD System**: Reduce shadow quality for distant entities
4. **Temporal Coherence**: Reuse shadow calculations across frames

## Troubleshooting

### Shadows Not Rendering

1. Check `ShadowsEnabled` in LightingConfig
2. Verify entities have both PositionComponent and ShadowComponent
3. Ensure entities are within light radius
4. Check viewport culling settings
5. Verify `MaxShadows` limit not reached

### Performance Issues

1. Reduce `MaxShadows` (try 50-75)
2. Lower `ShadowQuality` to 0.5 or 0.75
3. Use Hard shadows instead of Soft
4. Enable viewport culling
5. Profile to identify bottlenecks

### Visual Artifacts

1. Check shadow opacity (0.0-1.0 range)
2. Verify shadow radius matches sprite size
3. Adjust soft edge radius for soft shadows
4. Check Z-ordering of shadow rendering

## References

- LIGHTING_SYSTEM.md - Lighting system documentation
- ROADMAP_V3.md - Phase 17 specification (Sophisticated Lighting)
- pkg/engine/shadow_components.go - Component definitions
- pkg/engine/shadow_system.go - System implementation
- pkg/engine/lighting_system.go - Lighting integration

## Version History

- **v2.0 Phase 14.1** (November 4, 2025): Initial shadow system implementation
  - ShadowComponent and AmbientOcclusionComponent
  - ShadowSystem with hard/soft/contact shadow types
  - LightingSystem integration
  - Genre-specific shadow presets
  - Comprehensive tests (85%+ coverage)
