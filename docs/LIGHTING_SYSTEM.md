# Dynamic Lighting System - Implementation Guide

## Overview

The Dynamic Lighting System adds atmospheric lighting effects to Venture, enhancing visual immersion with point lights, ambient lighting, and dynamic effects. The system integrates seamlessly with the existing ECS architecture and rendering pipeline.

**Status**: Phase 5.3 - IN PROGRESS  
**Version**: 1.1 Beta  
**Implementation Date**: October 2025

## Architecture

### Components

#### LightComponent
Point light source attached to entities (torches, spells, environmental lights).

**Fields**:
- `Color` (color.RGBA): Light color (RGB)
- `Radius` (float64): Maximum light reach in pixels
- `Intensity` (float64): Brightness multiplier (1.0 = full)
- `Falloff` (LightFalloffType): Distance attenuation curve
- `Enabled` (bool): Toggle light on/off
- `Flickering` (bool): Enable random intensity variation
- `Pulsing` (bool): Enable smooth periodic intensity changes

**Falloff Types**:
- `FalloffLinear`: Decreases linearly with distance
- `FalloffQuadratic`: Decreases with distance² (realistic, recommended)
- `FalloffInverseSquare`: Physical 1/d² falloff
- `FalloffConstant`: No falloff until radius cutoff

#### AmbientLightComponent
Global scene lighting (typically one per world/area).

**Fields**:
- `Color` (color.RGBA): Ambient light color
- `Intensity` (float64): Global brightness (0.0-1.0)

#### LightingConfig
System-wide configuration and performance settings.

**Fields**:
- `Enabled` (bool): Master toggle
- `MaxLights` (int): Maximum lights per frame (default 16)
- `GammaCorrection` (bool): Apply gamma correction
- `AmbientIntensity` (float64): Default ambient level
- `AmbientColor` (color.RGBA): Default ambient color

### Systems

#### LightingSystem
Processes lights and applies lighting to rendered scenes.

**Methods**:
- `Update(entities, deltaTime)`: Updates light animations
- `CollectVisibleLights(entities)`: Gathers lights in viewport
- `ApplyLighting(screen, scene, entities)`: Post-processing pass
- `CalculateLightIntensityAt(x, y, entities)`: Query light at point
- `SetViewport(x, y, w, h)`: Configure culling viewport

## Usage Examples

### Basic Setup

```go
// Create lighting configuration
config := engine.NewLightingConfig()
config.SetGenrePreset("fantasy") // Warm tones, moderate ambient

// Create lighting system
lightingSystem := engine.NewLightingSystem(world, config)

// Set viewport for culling
lightingSystem.SetViewport(cameraX, cameraY, screenWidth, screenHeight)
```

### Adding Lights to Entities

```go
// Player with torch
player := world.CreateEntity()
player.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
torchLight := engine.NewTorchLight(200) // 200-pixel radius
player.AddComponent(torchLight)

// Spell projectile with colored light
spell := world.CreateEntity()
spell.AddComponent(&engine.PositionComponent{X: 150, Y: 150})
spellLight := engine.NewSpellLight(80, color.RGBA{0, 100, 255, 255}) // Blue light
spell.AddComponent(spellLight)

// Environmental crystal
crystal := world.CreateEntity()
crystal.AddComponent(&engine.PositionComponent{X: 200, Y: 200})
crystalLight := engine.NewCrystalLight(120, color.RGBA{255, 0, 255, 255}) // Purple
crystal.AddComponent(crystalLight)
```

### Custom Lights

```go
// Custom light with specific settings
light := engine.NewLightComponent(
    150,                              // radius
    color.RGBA{255, 180, 100, 255},  // warm orange
    1.2,                              // intensity
)
light.Falloff = engine.FalloffQuadratic
light.Flickering = true
light.FlickerSpeed = 3.0
light.FlickerAmount = 0.15

entity.AddComponent(light)
```

### Ambient Lighting

```go
// Create ambient light entity (one per world/area)
ambientEntity := world.CreateEntity()
ambient := engine.NewAmbientLightComponent(
    color.RGBA{100, 110, 120, 255}, // Cool blue tint
    0.3,                            // Low ambient (dark dungeon)
)
ambientEntity.AddComponent(ambient)
```

### Integration with Game Loop

```go
func (g *Game) Update() error {
    deltaTime := 1.0 / 60.0
    
    // Update systems
    entities := world.GetAllEntities()
    lightingSystem.Update(entities, deltaTime)
    
    // ... other systems
    
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    // Render scene to buffer
    sceneBuffer := ebiten.NewImage(screenWidth, screenHeight)
    renderSystem.Draw(sceneBuffer, entities)
    
    // Apply lighting as post-processing
    lightingSystem.ApplyLighting(screen, sceneBuffer, entities)
}
```

### Genre-Specific Configuration

```go
// Fantasy: warm, bright
config.SetGenrePreset("fantasy")
// AmbientIntensity = 0.4, AmbientColor = warm tone

// Horror: dark, cold
config.SetGenrePreset("horror")
// AmbientIntensity = 0.15, AmbientColor = cold tone

// Sci-Fi: cool, moderate
config.SetGenrePreset("sci-fi")
// AmbientIntensity = 0.35, AmbientColor = cool blue

// Cyberpunk: dark with purple tint
config.SetGenrePreset("cyberpunk")
// AmbientIntensity = 0.25, AmbientColor = purple

// Post-Apocalyptic: dusty, harsh
config.SetGenrePreset("post-apocalyptic")
// AmbientIntensity = 0.3, AmbientColor = dusty yellow
```

### Querying Light Intensity

```go
// Check light level at a position (useful for stealth mechanics)
intensity := lightingSystem.CalculateLightIntensityAt(playerX, playerY, entities)

if intensity < 0.2 {
    // Player is in shadows (stealth bonus)
} else if intensity > 0.8 {
    // Player is well-lit (visible to enemies)
}
```

### Performance Optimization

```go
// Configure for performance
config.MaxLights = 8            // Reduce for lower-end hardware
config.GammaCorrection = false  // Disable if performance critical

// Enable/disable dynamically
lightingSystem.SetEnabled(false) // Turn off lighting
lightingSystem.SetEnabled(true)  // Turn on lighting

// Command-line flag
if *enableLighting {
    lightingSystem.SetEnabled(true)
}
```

## Performance Characteristics

### Benchmarks (Target)

- **Frame Rate**: 60+ FPS with 16 lights
- **Light Culling**: O(n) where n = total lights, only processes visible lights
- **Light Limit**: Max 16 lights per frame (configurable)
- **Memory**: ~100 bytes per light, ~10KB for lighting system
- **CPU**: ~0.5ms per frame for 16 lights (target)

### Optimization Strategies

1. **Viewport Culling**: Only process lights visible on screen
2. **Light Limits**: Hard cap on lights per frame
3. **Lazy Updates**: Light animations update only once per frame
4. **Deferred Rendering**: Lighting calculated in separate pass
5. **Feature Toggle**: Can be disabled entirely for low-end systems

## Genre-Specific Recommendations

### Fantasy
- **Ambient**: 0.4 intensity, warm tones (120, 110, 90)
- **Torches**: Flickering orange lights (200-pixel radius)
- **Magic**: Pulsing colored lights based on element
- **Crystals**: Gentle pulsing lights in dungeons

### Sci-Fi
- **Ambient**: 0.35 intensity, cool blue (90, 110, 140)
- **Lights**: Constant white/cyan lights (no flicker)
- **Screens**: Pulsing blue/green lights on terminals
- **Weapons**: Bright, sharp lights on energy weapons

### Horror
- **Ambient**: 0.15 intensity, very dark (80, 75, 90)
- **Flashlight**: Player torch with 150-pixel radius
- **Environmental**: Sparse, flickering lights
- **Effects**: Very low intensity for tension

### Cyberpunk
- **Ambient**: 0.25 intensity, purple tint (100, 80, 120)
- **Neon**: Bright pulsing lights (pink, blue, green)
- **Streets**: Flickering streetlights
- **Holograms**: Constant cyan lights

### Post-Apocalyptic
- **Ambient**: 0.3 intensity, dusty (130, 120, 100)
- **Fires**: Flickering orange lights from burning debris
- **Lanterns**: Dim, flickering lights on survivors
- **Radiation**: Pulsing green lights in hazard zones

## Integration Checklist

- [x] Create LightComponent and AmbientLightComponent
- [x] Implement LightingSystem with culling and limits
- [x] Add animation support (flicker, pulse)
- [x] Write comprehensive tests (85%+ coverage)
- [ ] Integrate with RenderSystem
- [ ] Add player torch by default
- [ ] Generate spell lights based on element
- [ ] Spawn environmental lights in terrain generation
- [ ] Add command-line flag `-enable-lighting`
- [ ] Performance profiling and optimization
- [ ] Update user manual with lighting controls
- [ ] Add lighting section to TECHNICAL_SPEC.md

## Known Limitations

1. **Shader Support**: Current implementation uses simplified lighting without shaders
2. **Shadows**: No shadow casting in initial implementation
3. **Occlusion**: Lights don't account for walls/obstacles
4. **Dynamic Quality**: No automatic quality adjustment yet

## Future Enhancements (Phase 10+)

- Shadow casting from light sources
- Light occlusion by terrain
- Dynamic quality adjustment
- Colored shadows
- Light bloom effects
- Volumetric lighting (fog/dust)
- HDR lighting support

## References

- **Components**: `pkg/engine/lighting_components.go`
- **System**: `pkg/engine/lighting_system.go`
- **Tests**: `pkg/engine/lighting_components_test.go`, `pkg/engine/lighting_system_test.go`
- **Roadmap**: `docs/ROADMAP.md` (Section 5.3)
- **Architecture**: `docs/ARCHITECTURE.md`

---

**Implementation Status**: Components and System Complete  
**Next Steps**: Render pipeline integration and entity spawning  
**Target Completion**: 2-3 weeks from start date  
**Maintainer**: Venture Development Team
