# Character Avatar Enhancement Plan
## Directional Facing & Aerial-View Perspective Migration

**Status**: ✅ APPROVED FOR IMMEDIATE IMPLEMENTATION - BLOCKING TASK  
**Date**: October 26, 2025 (Approved)  
**Priority**: CRITICAL - Next blocking task for Phase 9.1 completion  
**Objective**: Enhance procedural character sprite generation with (1) directional facing indicators and (2) aerial-view perspective compatible with top-down gameplay

---

## 1. Architecture Analysis

### Current System Overview
The avatar generation system resides in `pkg/rendering/sprites/` with three key files:
- **`generator.go`**: Main sprite generation orchestrator with `generateEntity()` and `generateEntityWithTemplate()` methods
- **`anatomy_template.go`**: Template-based anatomical structure system with directional support (Phase 5.2 complete)
- **`animation.go`**: Frame generation with state-based transformations

### Critical Finding: Directional System Already Implemented
**The system already supports 4-directional facing** via `Direction` enum (`DirUp`, `DirDown`, `DirLeft`, `DirRight`) and `HumanoidDirectionalTemplate()` function. Templates adjust arm positioning, head offset, and weapon/shield placement based on facing direction passed through `Config.Custom["facing"]`.

### Current Limitations
1. **Side-view perspective**: Templates use vertical body proportions (head 30%, torso 40%, legs 30%) suitable for side-scrolling but **not optimal for aerial gameplay camera**
2. **No integration with movement system**: Direction parameter exists but isn't connected to entity velocity in `pkg/engine/movement.go`
3. **Static sprites in game**: Client (`cmd/client/`) creates sprites without passing facing direction, resulting in unchanging character orientation

### Integration Points
- **Movement System** (`pkg/engine/movement.go`): Updates velocity, applies friction, triggers animation states
- **Render System** (`pkg/engine/render_system.go`): `drawEntity()` renders sprites at world positions via camera transformation
- **Animation System** (`pkg/engine/animation.go`): `AnimationComponent` tracks states (idle, walk, run, attack) but not facing direction
- **Server** (`cmd/server/main.go`): `createPlayerEntity()` generates player sprites at 28×28 pixels without facing parameter

---

## 2. Technical Design

### 2.1 Aerial-View Template Architecture

**Design Principle**: Maintain existing template system, add aerial-optimized variants.

**Aerial Template Specifications**:
```go
// New template function in anatomy_template.go
func HumanoidAerialTemplate(direction Direction) AnatomicalTemplate
```

**Key Proportions** (28×28px player sprite):
- **Head**: 35% of height, positioned at Y=0.20 (more prominent from above)
- **Torso**: 50% of height, positioned at Y=0.50 (compressed vertical, wider horizontal)
- **Legs**: Minimal visibility (15%), positioned at Y=0.80 (mostly obscured)
- **Arms**: Extend laterally, width=0.70, positioned at torso level
- **Shadow**: Ellipse at Y=0.90, width=0.50, height=0.15, opacity=0.35

**Directional Indicators**:
- **DirUp**: Arms symmetrical, head centered, weapon behind torso (ZIndex=9)
- **DirDown**: Arms asymmetric (forward reach), head centered, weapon in front (ZIndex=13)
- **DirLeft**: Head shift to X=0.42, left arm visible (X=0.35), right arm obscured
- **DirRight**: Head shift to X=0.58, right arm visible (X=0.65), left arm obscured

**Shape Selection** (aerial perspective):
- Head: `ShapeCircle`, `ShapeEllipse` (avoid `ShapeSkull` - too detailed)
- Torso: `ShapeEllipse` (horizontal orientation), `ShapeBean`
- Arms: `ShapeCapsule` with rotation based on direction (0°, 45°, 90°)
- Legs: `ShapeEllipse` (compressed, low opacity=0.8)

### 2.2 Sprite Generation Strategy

**Option A: Sprite Sheet (Recommended)**
- Generate 4 sprites (one per direction) at entity creation
- Store in `EbitenSprite` component as cached images
- Switch based on `AnimationComponent.Facing` field (new)
- **Performance**: 4× memory per entity, zero runtime regeneration cost
- **Fits constraint**: 28×28×4 = 3.1KB per entity sprite sheet

**Option B: Runtime Rotation**
- Generate single forward-facing sprite
- Apply rotation transform in `render_system.go`
- **Limitation**: Rotation doesn't create asymmetry (weapon position, arm visibility)
- **Rejected**: Insufficient visual clarity for gameplay

**Selected Strategy**: Sprite sheet with lazy loading (generate direction on first use).

### 2.3 Component Architecture Changes

**New Field in AnimationComponent** (`pkg/engine/animation.go`):
```go
type AnimationComponent struct {
    // ... existing fields ...
    Facing Direction // New: current facing direction
}

// Direction enum (compatible with sprites.Direction)
type Direction int
const (
    DirUp Direction = iota
    DirDown
    DirLeft
    DirRight
)
```

**Extended EbitenSprite** (`pkg/engine/render_system.go`):
```go
type EbitenSprite struct {
    // ... existing fields ...
    DirectionalImages map[sprites.Direction]*ebiten.Image // Sprite sheet
    CurrentDirection  sprites.Direction                    // Active direction
}
```

### 2.4 Movement-to-Direction Integration

**Algorithm** (in `pkg/engine/movement.go::Update()`):
```go
// After velocity update, before animation state change
if animComp, hasAnim := entity.GetComponent("animation"); hasAnim {
    anim := animComp.(*AnimationComponent)
    
    // Determine facing from velocity (8-directional with threshold)
    if math.Abs(vel.VX) > 0.1 || math.Abs(vel.VY) > 0.1 {
        // Prioritize cardinal directions
        if math.Abs(vel.VX) > math.Abs(vel.VY) {
            if vel.VX > 0 { anim.Facing = DirRight }
            else          { anim.Facing = DirLeft }
        } else {
            if vel.VY > 0 { anim.Facing = DirDown }  // Y increases downward
            else          { anim.Facing = DirUp }
        }
    }
    // Else: retain last facing direction (stationary entities)
}
```

**Threshold**: 0.1 pixels/frame to ignore jitter from friction decay.

### 2.5 Performance Impact Assessment

**Generation Cost** (deterministic, one-time per direction):
- Current: 28ms average for template-based sprite (28×28px)
- Aerial templates: +5ms overhead (4 directions × template selection logic)
- **Total**: 33ms per entity initial generation
- **Within constraint**: <65ms target ✓

**Memory Impact**:
- Current: 3.1KB per 28×28 RGBA sprite
- Sprite sheet: 12.4KB (4 directions)
- 100 entities: 1.24MB (acceptable for client <500MB target)

**Render Performance**:
- Zero impact: same `DrawImage()` call, different source image
- Direction switch: map lookup O(1), negligible

---

## 3. Implementation Roadmap

⚠️ **CRITICAL: APPROVED FOR IMMEDIATE IMPLEMENTATION** ⚠️  
All phases below are approved and ready for development. Begin Phase 1 immediately.  
Total estimated time: 16-20 hours over 3-5 days (November 2-5, 2025 target completion).

### Phase 1: Aerial Template Foundation (3-4 hours) - START IMMEDIATELY
**Files**: `pkg/rendering/sprites/anatomy_template.go`

**Tasks**:
1. ✅ Create `HumanoidAerialTemplate(direction Direction)` function
   - Define body part positions with aerial proportions
   - Implement directional asymmetry (head offset, arm visibility)
   - Add shadow configuration
2. ✅ Create `SelectAerialTemplate(entityType, genre string, direction Direction)` dispatcher
   - Route humanoid types to aerial templates
   - Fallback to existing templates for non-humanoid entities
3. ✅ Add genre-specific aerial variants:
   - `FantasyHumanoidAerial()`, `SciFiHumanoidAerial()`, `HorrorHumanoidAerial()`
   - Adapt proportions while maintaining genre aesthetic

**Testing**:
```bash
go test -run TestHumanoidAerialTemplate ./pkg/rendering/sprites/
go test -run TestAerialDirectionalAsymmetry ./pkg/rendering/sprites/
```

**Validation Criteria**:
- ✓ Template generates 28×28px sprite in <35ms
- ✓ All 4 directions visually distinct (head offset ±0.08, arm positions differ)
- ✓ Maintains seed determinism (same seed → same sprite for each direction)

### Phase 2: Engine Component Integration (2-3 hours)
**Files**: `pkg/engine/animation.go`, `pkg/engine/render_system.go`

**Tasks**:
1. ✅ Add `Facing Direction` field to `AnimationComponent`
   - Update `NewAnimationComponent()` constructor (default: `DirDown`)
   - Add `SetFacing(dir Direction)` method
2. ✅ Extend `EbitenSprite` with directional storage
   - Add `DirectionalImages map[sprites.Direction]*ebiten.Image`
   - Add `CurrentDirection sprites.Direction`
   - Update `NewSpriteComponent()` to initialize map
3. ✅ Modify `render_system.go::drawEntity()`
   - Select sprite from `DirectionalImages[sprite.CurrentDirection]`
   - Fallback to `sprite.Image` if directional not available (backward compatible)

**Testing**:
```bash
go test -run TestAnimationComponent_Facing ./pkg/engine/
go test -run TestEbitenSprite_DirectionalImages ./pkg/engine/
```

### Phase 3: Movement System Integration (2 hours)
**Files**: `pkg/engine/movement.go`

**Tasks**:
1. ✅ Add direction calculation logic in `MovementSystem::Update()`
   - Insert after velocity application, before animation state update
   - Use velocity vector to determine cardinal direction
   - Apply 0.1 threshold to filter noise
2. ✅ Update `AnimationComponent.Facing` based on velocity
3. ✅ Preserve facing when entity is stationary (use last direction)

**Testing**:
```bash
go test -run TestMovementSystem_DirectionUpdate ./pkg/engine/
# Test scenarios: moving N/S/E/W, diagonal (prioritize horizontal), stationary
```

### Phase 4: Sprite Generation Pipeline (3 hours)
**Files**: `pkg/rendering/sprites/generator.go`, `cmd/server/main.go`, `cmd/client/`

**Tasks**:
1. ✅ Update `generator.go::generateEntityWithTemplate()`
   - Check for `config.Custom["useAerial"]` flag
   - Route to `SelectAerialTemplate()` when flag is true
   - Maintain backward compatibility (default to side-view)
2. ✅ Modify `cmd/server/main.go::createPlayerEntity()`
   - Generate 4-directional sprite sheet at entity creation
   - Pass `useAerial: true` and `facing` to sprite generator
   - Store all 4 images in `sprite.DirectionalImages`
3. ✅ Update client sprite generation calls
   - Add aerial flag for player entities
   - Generate on first use for NPCs/enemies (lazy loading)

**Testing**:
```bash
go run ./cmd/entitytest -entityType humanoid -useAerial -seed 12345
# Visual inspection: verify 4 distinct directional sprites
```

### Phase 5: Visual Consistency Refinement (2-3 hours)
**Files**: `pkg/rendering/sprites/anatomy_template.go`

**Tasks**:
1. ✅ Audit all aerial templates for proportion consistency
   - Verify head-to-torso ratio across genres (35:50)
   - Ensure shadow size matches body dimensions
2. ✅ Add color coherence checks
   - Arms use `secondary` color, torso uses `primary`
   - Weapons use `accent1`, shields use `accent2`
3. ✅ Implement boss aerial scaling
   - Create `BossAerialTemplate(base, scale float64)`
   - Apply 2.5× scaling while maintaining directional asymmetry

### Phase 6: Testing & Validation (3 hours)

**Unit Tests** (create `pkg/rendering/sprites/aerial_test.go`):
```go
func TestAerialTemplate_Determinism(t *testing.T)
func TestAerialTemplate_Performance(t *testing.T)
func TestAerialDirection_Asymmetry(t *testing.T)
func TestAerialGenreVariants(t *testing.T)
```

**Integration Tests** (create `pkg/engine/directional_rendering_test.go`):
```go
func TestMovementToFacing_CardinalDirections(t *testing.T)
func TestMovementToFacing_DiagonalPriority(t *testing.T)
func TestRenderSystem_DirectionalSprites(t *testing.T)
```

**Manual Testing**:
```bash
go run ./cmd/client -seed 42 -width 800 -height 600
# WASD movement: verify player sprite faces movement direction
# Stop moving: verify sprite retains last direction
# Attack while moving: verify weapon orientation matches facing
```

**Performance Benchmarks**:
```bash
go test -bench=BenchmarkAerialSpriteGeneration ./pkg/rendering/sprites/
go test -bench=BenchmarkDirectionalSwitch ./pkg/engine/
```

**Success Criteria**:
- ✓ All tests pass with >80% coverage for new code
- ✓ Aerial sprite generation <35ms (avg)
- ✓ Direction switching <0.1ms (render loop)
- ✓ Visual validation: 4 distinct directions visible in gameplay

### Phase 7: Documentation & Migration (1-2 hours)

**Update Documentation**:
- Add aerial template section to `docs/API_REFERENCE.md`
- Update `pkg/rendering/sprites/doc.go` with aerial examples
- Add migration guide: "Converting Side-View to Aerial Sprites"

**Backward Compatibility**:
- Existing code defaults to `useAerial: false` (side-view)
- Server startup flag: `--aerial-sprites` to enable globally
- Client config option in menu system

---

## 4. Quality Improvements

### 4.1 Anatomy Proportion Consistency

**Current Issue**: Side-view templates use 30/40/30 proportions unsuitable for aerial perspective.

**Enhancement**:
- Standardize aerial proportions: **35/50/15** (head/torso/legs) across all genres
- Add validation in `anatomy_template_test.go`:
  ```go
  func TestAerialProportions_Standard(t *testing.T) {
      // Assert head RelativeHeight ≈ 0.35 ± 0.02
      // Assert torso RelativeHeight ≈ 0.50 ± 0.03
  }
  ```

### 4.2 Genre-Specific Aerial Adjustments

**Fantasy Aerial** (`FantasyHumanoidAerial`):
- Broader shoulders (torso width: 0.60 vs 0.50)
- Visible helmet shape on head (add `PartHelmet` with `ShapeHexagon`)
- Cape/cloak shadow behind torso (opacity: 0.4)

**Sci-Fi Aerial** (`SciFiHumanoidAerial`):
- Angular head shapes (`ShapeOctagon`, `ShapeHexagon`)
- Glowing accent on torso (add overlay with `accent3`, opacity: 0.7)
- Jetpack indicator for `DirUp` (small rectangle behind torso)

**Horror Aerial** (`HorrorHumanoidAerial`):
- Elongated head (height: 0.40, width: 0.28)
- Irregular torso shape (`ShapeOrganic`)
- Reduced shadow opacity (0.2) for ghostly effect

**Cyberpunk Aerial** (`CyberpunkHumanoidAerial`):
- Neon glow outlines (add `PartArmor` layer with `accent1`, low opacity)
- Asymmetric head (tech implant on one side)

**Post-Apoc Aerial** (`PostApocHumanoidAerial`):
- Ragged edges (use `ShapeOrganic` for torso/limbs)
- Makeshift armor plates (random rectangles on torso)

### 4.3 Animation Frame Considerations

**Current**: `animation.go::GenerateAnimationFrame()` applies transformations (offset, rotation, scale) to base sprite.

**Enhancement for Directional**:
- Walk/run animations: generate frames for **current facing direction only**
- Attack animation: add weapon swing arc based on facing
  - `DirRight`: weapon rotates 270°→0° (right sweep)
  - `DirDown`: weapon rotates 180°→90° (downward strike)
- Hit animation: apply knockback offset opposite to `Facing`

**Implementation**:
```go
// In animation.go::GenerateAnimationFrame()
config.Custom["facing"] = state.Facing // Pass current direction to generator
baseSprite, err := g.Generate(config)  // Generates directional sprite
```

### 4.4 Color Coherence System

**Issue**: Random color selection from palette can create inconsistent character appearance.

**Solution**: Role-based color assignment in templates
- `primary`: Torso/legs (main body color)
- `secondary`: Head/arms (skin/clothing)
- `accent1`: Weapons (metallic/magical)
- `accent2`: Shields/armor (defensive equipment)
- `accent3`: Special effects (glows, trails)

**Validation**: Add test ensuring all body parts reference valid color roles:
```go
func TestTemplate_ValidColorRoles(t *testing.T) {
    validRoles := []string{"primary", "secondary", "accent1", "accent2", "accent3", "shadow"}
    // Assert all PartSpec.ColorRole in validRoles
}
```

---

## Success Criteria Summary

✅ **Functional Requirements**:
- 4-directional sprites (Up/Down/Left/Right) with visual asymmetry
- Automatic facing update based on movement velocity
- Aerial perspective compatible with top-down camera

✅ **Performance Requirements**:
- Aerial sprite generation: <35ms per direction (meets <65ms constraint)
- Direction switching: <0.1ms per frame (negligible overhead)
- Memory: <15KB per entity sprite sheet (within <500MB client target)

✅ **Quality Requirements**:
- Maintains seed-based determinism (same seed → same sprites)
- Passes all existing sprite tests (backward compatible)
- Visual consistency across genres (standardized proportions)

✅ **Integration Requirements**:
- Works with current ECS architecture (no structural changes)
- Zero external assets (100% procedural)
- Backward compatible (side-view templates remain available)

---

## Risks & Mitigations

**Risk 1**: Aerial perspective reduces character distinctiveness (all characters look similar from above)  
**Mitigation**: Emphasize head shapes, weapon/equipment visibility, genre-specific overlays (helmets, glows, cloaks)

**Risk 2**: 4× memory usage per entity impacts low-end devices  
**Mitigation**: Lazy loading (generate direction on first use), configurable quality setting (2-direction mode for mobile)

**Risk 3**: Movement jitter causes rapid direction changes (visual flicker)  
**Mitigation**: 0.1 velocity threshold + direction persistence when stationary

---

## Next Steps - IMMEDIATE ACTION REQUIRED

**Implementation Authorization**: ✅ APPROVED  
**Start Date**: October 26, 2025  
**Target Completion**: November 2-5, 2025 (16-20 hours development time)

**Approved Implementation Sequence**:
1. **Phase 1-2** (templates + components): 5-7 hours - START IMMEDIATELY
2. **Phase 3-4** (movement + generation): 5 hours
3. **Phase 5-7** (polish + testing): 6-8 hours

**Blocking Status**: This task blocks:
- Phase 9.1 completion (critical for production readiness)
- Visual consistency improvements in Phase 9.2+
- Player experience enhancements requiring directional clarity

**Implementation Priority**: CRITICAL  
**Resource Assignment**: Primary developer focus until completion

**Success Gate**: All 7 phases must complete with:
- ✓ All tests passing (>80% coverage for new code)
- ✓ Performance targets met (<35ms aerial sprite generation)
- ✓ Visual validation confirmed (4 distinct directions)
- ✓ Zero regressions in existing systems

---

**Document Version**: 1.1 - APPROVED FOR IMPLEMENTATION  
**Author**: GitHub Copilot  
**Review Status**: ✅ APPROVED - Ready for Implementation  
**Approver**: Project Lead  
**Next Milestone**: Phase 1 completion within 24-48 hours
