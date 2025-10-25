# Visual Generation Enhancement Plan

**Version:** 1.0  
**Created:** October 24, 2025  
**Timeline:** 8-10 weeks  
**Status:** Planning Phase

## Executive Summary

Enhance Venture's procedural visual generation to provide 3-5x visual variety, smooth character animation, and genre-distinct aesthetics while maintaining deterministic generation, zero external assets, and the fixed 28x28 pixel player sprite size.

### Success Metrics
- **Visual Variety:** 3-5x increase in unique visual elements
- **Animation:** 30+ FPS smooth character animation with 10+ states per entity type
- **Performance:** Maintain 60+ FPS (current: 106 FPS), <500MB memory
- **Player Sprite:** Fixed at 28x28 pixels (NON-NEGOTIABLE)
- **Test Coverage:** Maintain 80%+ for all modified packages
- **Determinism:** 100% reproducible (same seed = identical output)

---

## Phase 1: Animation System Foundation (Week 1-2)

### Objectives
1. Implement multi-frame sprite animation system
2. Add animation state machine for entities
3. Create procedural frame generation
4. Optimize for batch rendering

### Implementation

**New Components:**
```go
// pkg/engine/animation_component.go
type AnimationComponent struct {
    CurrentState    AnimationState  // idle, walk, attack, death, etc.
    Frames          []*ebiten.Image // Frame cache
    FrameIndex      int
    FrameTime       float64         // Time per frame
    TimeAccumulator float64
    Loop            bool
    OnComplete      func()          // Callback for one-shot animations
}

type AnimationState string
const (
    StateIdle   AnimationState = "idle"
    StateWalk                  = "walk"
    StateAttack                = "attack"
    StateDeath                 = "death"
    StateCast                  = "cast"
)
```

**New System:**
```go
// pkg/engine/animation_system.go
type AnimationSystem struct {
    spriteGenerator *sprites.Generator
    frameCache      map[string][]*ebiten.Image // Cache by seed+state
}

func (s *AnimationSystem) Update(entities []*Entity, deltaTime float64) {
    // Update frame timers, transition states, regenerate frames if needed
}
```

**Sprite Generator Extension:**
```go
// pkg/rendering/sprites/animation.go
func (g *Generator) GenerateAnimationFrames(config Config, state AnimationState, frameCount int) ([]*ebiten.Image, error) {
    // Generate frames with deterministic variations per frame
    // Use seed + frame index for unique but consistent frames
}
```

### Files Modified/Created
- `pkg/engine/animation_component.go` (new, ~120 lines)
- `pkg/engine/animation_system.go` (new, ~180 lines)
- `pkg/rendering/sprites/animation.go` (new, ~200 lines)
- `pkg/engine/render_system.go` (modify: integrate animation frames)

### Testing
- Animation state transitions (table-driven tests)
- Frame generation determinism (same seed = same frames)
- Performance benchmark: 2000 animated entities at 60 FPS
- Memory test: frame cache size limits

### Performance Targets
- Frame generation: <5ms per entity
- Memory: <2MB for 100 animated entities (with caching)
- No frame drops with 500+ animated entities

---

## Phase 2: Character Visual Expansion (Week 3-4)

### Objectives
1. Multi-layer sprite composition (head, body, limbs, accessories)
2. Visual equipment display on sprites
3. Status effect visual indicators
4. Entity variation system

### Implementation

**Enhanced Sprite Config:**
```go
// pkg/rendering/sprites/types.go (extend)
type CompositeConfig struct {
    BaseSeed    int64
    Layers      []LayerConfig
    Equipment   []EquipmentVisual
    StatusFX    []StatusEffect
}

type LayerConfig struct {
    Type     LayerType  // head, body, legs, weapon, armor, accessory
    ZIndex   int
    Offset   Point
    Scale    float64
    Color    color.Color
}
```

**Equipment Visualization:**
```go
// pkg/engine/equipment_visual_component.go
type EquipmentVisualComponent struct {
    WeaponLayer  *ebiten.Image
    ArmorLayer   *ebiten.Image
    AccessoryLayers []*ebiten.Image
    Dirty        bool  // Regenerate on equipment change
}
```

### Files Modified/Created
- `pkg/rendering/sprites/composite.go` (new, ~250 lines)
- `pkg/rendering/sprites/equipment.go` (new, ~150 lines)
- `pkg/engine/equipment_visual_component.go` (new, ~80 lines)
- `pkg/engine/equipment_visual_system.go` (new, ~120 lines)
- `pkg/rendering/sprites/generator.go` (modify: add composition)

### Testing
- Composite sprite generation correctness
- Equipment layer ordering
- Status effect overlay rendering
- Variation system (100+ unique entities from same template)

---

## Phase 3: Shape & Pattern Library (Week 5)

### Objectives
1. Add 10+ new procedural shapes
2. Implement pattern generation (stripes, dots, gradients, noise)
3. Genre-specific shape libraries
4. Composite shape templates

### Implementation

**New Shapes:**
```go
// pkg/rendering/shapes/types.go (extend ShapeType)
const (
    // Existing: Circle, Rectangle, Triangle, Polygon, Star, Ring
    ShapeHexagon ShapeType = 6
    ShapeOctagon           = 7
    ShapeCross             = 8
    ShapeHeart             = 9
    ShapeCrescent          = 10
    ShapeGear              = 11
    ShapeCrystal           = 12
    ShapeLightning         = 13
    ShapeWave              = 14
    ShapeSpiral            = 15
    ShapeOrganic           = 16  // Blob-like using noise
)
```

**Pattern System:**
```go
// pkg/rendering/patterns/generator.go (new package)
type PatternType int
const (
    PatternStripes PatternType = iota
    PatternDots
    PatternGradient
    PatternNoise
    PatternCheckerboard
    PatternCircles
)

func (g *Generator) ApplyPattern(img *ebiten.Image, pattern PatternType, params PatternParams) error
```

### Files Modified/Created
- `pkg/rendering/shapes/generator.go` (extend: +200 lines)
- `pkg/rendering/patterns/generator.go` (new, ~300 lines)
- `pkg/rendering/patterns/types.go` (new, ~80 lines)

### Testing
- Each new shape renders correctly
- Pattern determinism
- Genre-specific shape selection

---

## Phase 4: Color System Enhancement (Week 6)

### Objectives
1. Expand palettes to 12+ colors per genre
2. Implement color harmony rules
3. Add palette mood variations
4. Rarity-based color schemes

### Implementation

**Extended Palette:**
```go
// pkg/rendering/palette/types.go (extend)
type Palette struct {
    // Core colors (existing)
    Primary, Secondary, Background, Text color.Color
    Accent1, Accent2, Danger, Success color.Color
    
    // Extended colors (new)
    Highlight, Shadow color.Color
    Neutral1, Neutral2, Neutral3 color.Color
    
    // Rarity colors
    RarityColors map[item.Rarity]color.Color
    
    // Mood variant
    Mood PaletteMood  // bright, dark, saturated, muted
}

type ColorHarmony int
const (
    HarmonyComplementary ColorHarmony = iota
    HarmonyAnalogous
    HarmonyTriadic
    HarmonyTetradic
)
```

### Files Modified/Created
- `pkg/rendering/palette/generator.go` (extend: +150 lines)
- `pkg/rendering/palette/harmony.go` (new, ~120 lines)
- `pkg/rendering/palette/types.go` (modify)

### Testing
- Color harmony validation (hue relationships)
- Rarity color distinctness (perceptual distance tests)
- Mood variations maintain genre consistency

---

## Phase 5: Environment Visual Enhancement (Week 7)

### Objectives
1. Advanced tile pattern generation
2. Environmental object sprites (furniture, decorations)
3. Lighting/shadow color modulation
4. Weather particle effects

### Implementation

**Tile Variations:**
```go
// pkg/rendering/tiles/generator.go (extend)
func (g *Generator) GenerateTileSet(tileType TileType, variations int) []*ebiten.Image {
    // Generate N variations of same tile type using seed offsets
}
```

**Environmental Objects:**
```go
// pkg/procgen/environment/generator.go (new package)
type EnvironmentObject struct {
    Type     ObjectType  // furniture, decoration, obstacle, hazard
    Sprite   *ebiten.Image
    Collidable bool
    Interactable bool
}
```

**Lighting System:**
```go
// pkg/rendering/lighting/system.go (new package)
type LightingComponent struct {
    AmbientColor color.Color
    LightSources []LightSource
}

type LightSource struct {
    Position Point
    Color    color.Color
    Radius   float64
    Intensity float64
}
```

### Files Modified/Created
- `pkg/rendering/tiles/variations.go` (new, ~150 lines)
- `pkg/procgen/environment/generator.go` (new, ~200 lines)
- `pkg/rendering/lighting/system.go` (new, ~180 lines)
- `pkg/engine/lighting_component.go` (new, ~60 lines)

### Testing
- Tile variation uniqueness
- Object placement collision-free
- Lighting performance (many light sources)

---

## Phase 6: Performance Optimization (Week 8)

### Objectives
1. Sprite caching and pooling
2. Render culling optimization
3. Batch rendering for animated entities
4. Memory profiling and reduction

### Implementation

**Sprite Cache:**
```go
// pkg/rendering/cache/sprite_cache.go (new package)
type SpriteCache struct {
    cache    sync.Map  // map[cacheKey]*ebiten.Image
    maxSize  int
    eviction EvictionPolicy
}

func (c *SpriteCache) Get(key CacheKey) (*ebiten.Image, bool)
func (c *SpriteCache) Put(key CacheKey, sprite *ebiten.Image)
```

**Render Culling:**
```go
// pkg/engine/render_system.go (modify)
func (r *EbitenRenderSystem) cullEntities(entities []*Entity, viewport Rect) []*Entity {
    // Spatial partitioning + viewport culling
}
```

**Object Pool:**
```go
// pkg/rendering/pool/image_pool.go (new)
var ImagePool = sync.Pool{
    New: func() interface{} {
        return ebiten.NewImage(32, 32)
    },
}
```

### Files Modified/Created
- `pkg/rendering/cache/sprite_cache.go` (new, ~200 lines)
- `pkg/rendering/pool/image_pool.go` (new, ~80 lines)
- `pkg/engine/render_system.go` (optimize: +100 lines)

### Testing
- Cache hit rate benchmarks (target: >70%)
- Memory usage profiling (before/after)
- Frame time benchmarks (2000 entities)
- Stress test: 5000 animated entities

### Performance Targets
- Frame time: <16ms (60 FPS) with 2000 entities
- Memory: <400MB total (current: ~350MB baseline)
- Cache efficiency: 70%+ hit rate

---

## Phase 7: Integration & Polish (Week 9-10)

### Objectives
1. Integrate all systems with ECS
2. Network synchronization for animation states
3. Save/load system updates
4. Visual regression testing

### Implementation

**Network Sync:**
```go
// pkg/network/animation_sync.go (new)
type AnimationStatePacket struct {
    EntityID uint64
    State    AnimationState
    Frame    int
}
```

**Save/Load:**
```go
// pkg/saveload/visual_state.go (extend)
type VisualState struct {
    AnimationState AnimationState
    EquipmentVisuals []EquipmentID
    CurrentFrame   int
}
```

### Files Modified/Created
- `pkg/network/animation_sync.go` (new, ~100 lines)
- `pkg/saveload/visual_state.go` (extend, +80 lines)
- Integration glue code across systems (~150 lines total)

### Testing
- End-to-end animation playback
- Multiplayer animation sync validation
- Save/load preserves visual state
- Genre consistency across all systems
- Backward compatibility with existing saves

---

## Technical Design Details

### Animation State Machine
```
[Idle] ─┬→ [Walk] ──→ [Idle]
        ├→ [Attack] ─→ [Idle]
        ├→ [Cast] ───→ [Idle]
        └→ [Death] (terminal)
        
Transitions triggered by:
- Velocity change (walk)
- Combat system events (attack)
- Spell casting (cast)
- Health depletion (death)
```

### Sprite Composition System
```
Base Sprite (28x28 player, scaled for enemies)
├─ Body Layer (always present)
├─ Head Layer (facial features)
├─ Equipment Layers (0-5)
│  ├─ Weapon (hand position)
│  ├─ Armor (body overlay)
│  └─ Accessories (various positions)
└─ Effect Layers (status effects, buffs)
   ├─ Buff glow
   ├─ Debuff particles
   └─ Environmental (wet, burning, frozen)
```

### Frame Generation Algorithm
```go
// Deterministic frame generation
func generateFrame(baseSeed int64, state AnimationState, frameIndex int) *ebiten.Image {
    seed := baseSeed + hash(state) + int64(frameIndex)
    rng := rand.New(rand.NewSource(seed))
    
    // Apply frame-specific transformations
    offset := calculateOffset(state, frameIndex)      // e.g., walk cycle
    rotation := calculateRotation(state, frameIndex)  // e.g., attack swing
    scale := calculateScale(state, frameIndex)        // e.g., squash/stretch
    
    return compositeFrame(seed, offset, rotation, scale)
}
```

### Cache Strategy
- **Level 1:** Recent frames (LRU cache, 100 entries)
- **Level 2:** Common animations (idle, walk - permanent)
- **Level 3:** On-demand generation (rare states)

### Performance Optimization Strategies
1. **Batch Rendering:** Group entities by sprite/animation state
2. **Spatial Partitioning:** Quadtree for viewport culling (cell size: 256x256)
3. **LOD System:** Simplify distant entities (>800px from camera)
4. **Frame Skip:** Generate every 2nd frame for far entities
5. **Lazy Generation:** Generate sprites on first visibility

---

## Risk Mitigation

### Risk 1: Animation Performance Impact
**Mitigation:** Aggressive caching, frame pooling, benchmark gates (no merge if <60 FPS)

### Risk 2: Memory Explosion (Frame Cache)
**Mitigation:** LRU eviction, cache size limits (max 200MB), lazy generation

### Risk 3: Network Bandwidth (Animation Sync)
**Mitigation:** Sync state only (not frames), delta compression, 200ms interpolation buffer

### Risk 4: Backward Compatibility
**Mitigation:** Versioned save format, fallback to static sprites, migration tool

### Risk 5: Determinism Breaking
**Mitigation:** Comprehensive determinism tests, seed isolation per system

---

## Migration & Rollout

### Phase Rollout Strategy
1. **Behind Feature Flag:** `ENABLE_ANIMATIONS=true`
2. **Opt-in Beta:** CLI flag `-experimental-visuals`
3. **Gradual Rollout:** Per-genre enablement
4. **Full Deployment:** After 2 weeks of beta testing

### Backward Compatibility
- Old saves auto-migrate (add default animation state)
- Static sprite fallback for old entity data
- Server version negotiation (animation-aware clients only)

### Fallback Mechanisms
```go
if animationSystem.Failed() {
    log.Warn("Falling back to static sprites")
    useStaticRenderer()
}
```

---

## Validation & Testing

### Automated Tests Per Phase
| Phase | Unit Tests | Integration Tests | Benchmarks |
|-------|-----------|-------------------|------------|
| 1     | 15        | 5                 | 3          |
| 2     | 20        | 8                 | 2          |
| 3     | 18        | 4                 | 2          |
| 4     | 12        | 3                 | 1          |
| 5     | 16        | 6                 | 2          |
| 6     | 10        | 8                 | 5          |
| 7     | 8         | 12                | 3          |

### Visual Regression Testing
```bash
# Generate reference images
go test -tags visual ./pkg/rendering/... -update-golden

# Compare against baseline
go test -tags visual ./pkg/rendering/...
```

### Performance Benchmarks
```bash
# Baseline (must pass before merge)
go test -bench=. ./pkg/rendering/sprites -benchtime=10s
# Target: <5ms per sprite generation
# Target: 2000 entities at 60+ FPS
```

### Genre Consistency Validation
```go
func TestGenreVisualConsistency(t *testing.T) {
    genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
    for _, genre := range genres {
        validatePalette(genre)
        validateShapeSet(genre)
        validatePatterns(genre)
    }
}
```

---

## Code Estimates

### Total Lines of Code by Phase
| Phase | New Code | Modified Code | Test Code | Total |
|-------|----------|---------------|-----------|-------|
| 1     | 500      | 200           | 400       | 1100  |
| 2     | 600      | 150           | 500       | 1250  |
| 3     | 580      | 100           | 450       | 1130  |
| 4     | 270      | 150           | 300       | 720   |
| 5     | 590      | 100           | 400       | 1090  |
| 6     | 380      | 300           | 350       | 1030  |
| 7     | 180      | 200           | 400       | 780   |
| **Total** | **3100** | **1200** | **2800** | **7100** |

### Package Coverage Targets
- `pkg/rendering/sprites`: 100% → 100% (maintain)
- `pkg/rendering/shapes`: 100% → 100% (maintain)
- `pkg/rendering/palette`: 98.4% → 100%
- `pkg/rendering/patterns`: N/A → 90%+
- `pkg/rendering/cache`: N/A → 85%+
- `pkg/engine` (animation): N/A → 90%+

---

## Success Criteria Checklist

### Visual Variety (3-5x increase)
- [ ] 10+ animation states per entity type
- [ ] 16 shape types (up from 6)
- [ ] 12+ colors per palette (up from 8)
- [ ] 5+ tile variations per type
- [ ] 20+ environmental object types

### Performance (maintain/improve)
- [ ] 60+ FPS with 2000 animated entities
- [ ] <500MB memory usage
- [ ] <16ms frame time (95th percentile)
- [ ] <5ms sprite generation time

### Quality
- [ ] 80%+ test coverage maintained
- [ ] 100% deterministic (10 test runs, identical output)
- [ ] Zero visual regressions
- [ ] Genre-distinct styles clearly recognizable

### Compatibility
- [ ] Player sprite remains 28x28 pixels
- [ ] Backward compatible save files
- [ ] Network protocol version negotiation
- [ ] Graceful fallback to static sprites

---

## Timeline

```
Week 1-2:  Phase 1 (Animation Foundation)
Week 3-4:  Phase 2 (Character Expansion)
Week 5:    Phase 3 (Shapes & Patterns)
Week 6:    Phase 4 (Color Enhancement)
Week 7:    Phase 5 (Environment)
Week 8:    Phase 6 (Optimization)
Week 9-10: Phase 7 (Integration & Polish)
```

**Total Duration:** 10 weeks  
**Estimated Effort:** 200-250 hours

---

## Appendix: Key Algorithms

### Frame Interpolation (Walk Cycle)
```go
func calculateWalkOffset(frameIndex, totalFrames int) (dx, dy float64) {
    t := float64(frameIndex) / float64(totalFrames)
    cycle := math.Sin(t * 2 * math.Pi)
    dx = 0
    dy = cycle * 2.0  // 2 pixel bob
    return
}
```

### Color Harmony (Triadic)
```go
func generateTriadic(baseHue float64) []float64 {
    return []float64{
        baseHue,
        math.Mod(baseHue+120, 360),
        math.Mod(baseHue+240, 360),
    }
}
```

### Spatial Hash (Render Culling)
```go
func spatialHash(x, y, cellSize float64) (int, int) {
    return int(x / cellSize), int(y / cellSize)
}
```

---

**Document End**
