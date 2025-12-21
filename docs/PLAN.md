# Procedural Graphics Improvement Plan

## 1. Current State Analysis

### 1.1 Existing Procedural Generation Functions

| Package | File | Key Functions | Description |
|---------|------|---------------|-------------|
| `pkg/rendering/sprites` | `generator.go` | `Generate()`, `generateEntity()`, `generateItem()`, `generateTile()`, `generateParticle()`, `generateUI()` | Core sprite generation |
| `pkg/rendering/sprites` | `anatomy_template.go` | `HumanoidTemplate()`, `QuadrupedTemplate()`, `BlobTemplate()`, `MechanicalTemplate()`, `FlyingTemplate()`, `SelectTemplate()` | Entity body part templates |
| `pkg/rendering/sprites` | `item_template.go` | `SelectItemTemplate()` | Item visual templates |
| `pkg/rendering/tiles` | `generator.go` | `Generate()`, `generateFloor()`, `generateWall()`, `generateDoor()`, `generateWater()`, `generateLava()`, `generateTrap()`, `generateStairs()` | Tile rendering |
| `pkg/rendering/shapes` | `generator.go` | `Generate()`, `generateShape()`, `isInside()`, `inCircle()`, `inRectangle()`, `inPolygon()` | Geometric primitive generation |
| `pkg/rendering/particles` | `generator.go` | `Generate()`, `generateSparks()`, `generateSmoke()`, `generateMagic()`, `generateFlame()`, `generateEmbers()` | Particle effects |
| `pkg/rendering/palette` | `generator.go` | `Generate()`, `GenerateWithOptions()` | Color palette generation |
| `pkg/rendering/ui` | `generator.go` | UI element generation functions | UI component rendering |
| `pkg/rendering/lighting` | `system.go` | Lighting calculations | Dynamic lighting |

### 1.2 Current Asset Specifications

| Asset Type | Current Size | Format | Location |
|------------|--------------|--------|----------|
| Entity Sprites (player) | 28×28 px | RGBA | `sprites/types.go:81` |
| Entity Sprites (default) | 32×32 px | RGBA | `sprites/types.go:81-82` |
| Tile Sprites | 32×32 px | RGBA | `tiles/types.go:115-116` |
| Shape Primitives | 32×32 px | RGBA | `shapes/types.go:168-169` |
| Particles | 2-8 px | RGBA | `particles/types.go` |
| UI Elements | Variable | RGBA | `ui/types.go` |
| EnhancedHumanoidTemplate | 28×28 px | RGBA | Phase 15.1, `anatomy_template.go:238` |
| Phase 45 Templates | 64×64 px | RGBA | `anatomy_template.go:2436+` |

### 1.3 Key Dependencies

- **Ebiten v2.9.3**: Game engine for rendering
- **image/color**: Color types (RGBA)
- **math/rand**: Deterministic procedural generation
- **pkg/procgen/SeedGenerator**: Seed-based determinism
- **pkg/rendering/palette**: Color palette system
- **pkg/rendering/shapes**: Geometric primitives (24 shape types)

## 2. Design Specifications

### 2.1 Asset Dimensions

| Asset Type | Current Size | New Size | Scale Factor | Rationale |
|------------|--------------|----------|--------------|-----------|
| Entity Sprites (player) | 28×28 px | 64×64 px | 2.29× | Enhanced detail, facial features |
| Entity Sprites (NPC) | 32×32 px | 64×64 px | 2.0× | Consistent entity scale |
| Entity Sprites (boss) | 80×80 px | 128×128 px | 1.6× | Boss prominence |
| Tile Sprites | 32×32 px | 64×64 px | 2.0× | Higher terrain detail |
| Item Sprites | 16-32 px | 32-64 px | 2.0× | Improved item recognition |
| Particle Base | 2-8 px | 4-16 px | 2.0× | Visible particle effects |
| UI Icons | 16×16 px | 32×32 px | 2.0× | Crisp interface elements |

### 2.2 Visual Improvements

| Improvement | Current State | Target State | Measurement |
|-------------|---------------|--------------|-------------|
| Anti-aliasing | Off/Low | Medium (4× super-sampling) | Edge smoothness |
| Color depth | 8-bit RGBA | 8-bit RGBA + HDR effects | Dynamic range |
| Shape complexity | 5-8 body parts | 8-12 body parts | Part count |
| Silhouette score | 0.75 | 0.85+ | Recognition accuracy |
| Shadow detail | Ellipse only | Multi-layer ambient occlusion | Shadow quality |
| Directional sprites | 4 directions | 4 directions + diagonal hints | Direction clarity |
| Palette colors | 8-10 per genre | 12-16 per genre | Color variety |

### 2.3 Top-Down Perspective Requirements

| Visual Element | Current Perspective | Target Top-Down Specs |
|----------------|--------------------|-----------------------|
| **Entity Head** | 30% height, centered | 12-15% height, upper portion, ellipse shape |
| **Entity Torso** | 40% height, centered | 40-45% height, visible shoulders, bean/ellipse shape |
| **Entity Legs** | 30% height, bottom | 40-48% height, visible from above, shortened perspective |
| **Entity Shadow** | Small ellipse | Larger ellipse (40% width), 0.3 opacity, Z-index 0 |
| **Tile Floor** | Flat texture | Top-down grid, edge gradients |
| **Tile Wall** | Side view hints | Shortened height (3D projection), shadow on floor side |
| **Items** | Flat icons | 15° rotation, drop shadow |
| **Particles** | Random spread | Radial from center, Z-layer aware |

#### Body Part Proportions (Top-Down 64×64)
```
Head:   8×8 px (12.5% of sprite height)
        - RelativeY: 0.18 (upper portion)
        - Visible from above with foreshortening

Torso: 10×14 px (21.9% width × height)
        - RelativeY: 0.40 (center-upper)
        - Bean/ellipse shape for organic look

Arms:  12×10 px (extended reach visible)
        - RelativeY: 0.42 (shoulder level)
        - Horizontal orientation from above

Legs:   8×16 px (48% of sprite height)
        - RelativeY: 0.72 (lower portion)
        - Compressed vertical for top-down view

Shadow: 40% width ellipse
        - RelativeY: 0.93 (ground level)
        - Opacity: 0.3
```

## 3. Implementation Steps

### Step 1: Update Default Dimensions in Types ✅ COMPLETED

**Files modified:**
- `pkg/rendering/sprites/types.go` (lines 81-82)
- `pkg/rendering/tiles/types.go` (lines 115-116)
- `pkg/rendering/shapes/types.go` (lines 168-169)

**Changes completed:**
1. ✅ Updated `DefaultConfig()` in `sprites/types.go`:
   - Changed `Width: 32` → `Width: 64`
   - Changed `Height: 32` → `Height: 64`
2. ✅ Updated `DefaultConfig()` in `tiles/types.go`:
   - Changed `Width: 32` → `Width: 64`
   - Changed `Height: 32` → `Height: 64`
3. ✅ Updated `DefaultConfig()` in `shapes/types.go`:
   - Changed `Width: 32` → `Width: 64`
   - Changed `Height: 32` → `Height: 64`

**Tests updated:**
- `pkg/rendering/sprites/generator_test.go`: Updated TestDefaultConfig assertions
- `pkg/rendering/tiles/generator_test.go`: Updated TestDefaultConfig assertions
- `pkg/rendering/tiles/transitions_test.go`: Updated TestDefaultTransitionConfig assertions
- `pkg/rendering/shapes/generator_test.go`: Updated TestDefaultConfig assertions

**Testing checkpoint:** ✅ `go test ./pkg/rendering/sprites/... ./pkg/rendering/tiles/... ./pkg/rendering/shapes/...` all pass.

---

### Step 2: Migrate Sprite Anatomy Templates to Top-Down ✅ COMPLETED

**Files modified:**
- `pkg/rendering/sprites/anatomy_template.go`
- `pkg/rendering/sprites/anatomy_template_test.go`

**Changes completed:**
1. ✅ Updated `HumanoidTemplate()` for top-down 32×32 perspective:
   - Changed head `RelativeY: 0.25` → `RelativeY: 0.12` (upper 12%)
   - Changed head `RelativeHeight: 0.35` → `RelativeHeight: 0.12` (12% proportion)
   - Changed legs `RelativeHeight: 0.35` → `RelativeHeight: 0.48` (48% proportion)
   - Added `PreferredPixelSize` for pixel-perfect dimensions: head 6×4, torso 8×13, legs 6×15, arms 12×8
   - Updated shadow to 0.08 height for slim top-down appearance

2. ✅ Updated `QuadrupedTemplate()` for top-down perspective:
   - Removed rotation (was `Rotation: 90`) for top-down view
   - Head at upper portion (`RelativeY: 0.15`)
   - Body as horizontal ellipse visible from above
   - Legs spread under body with wide stance
   - Added `PreferredPixelSize` for all body parts

3. ✅ Updated `BlobTemplate()` for top-down perspective:
   - Added `PartHead` for facial features visible from above
   - Increased blob mass visibility (`RelativeHeight: 0.60`)
   - Shadow at bottom (`RelativeY: 0.88`)

4. ✅ Updated `FlyingTemplate()` for top-down perspective:
   - Wings spread horizontally (no rotation)
   - Head at top (`RelativeY: 0.20`) facing forward
   - Lighter shadow (`Opacity: 0.20`) for flying creatures
   - Added `PreferredPixelSize` for all body parts

5. ✅ Updated `MechanicalTemplate()` for top-down perspective:
   - Phase 45 proportions: head 12%, torso 40%, legs 48%
   - Head at upper portion (`RelativeY: 0.12`)
   - Shadow at bottom with slim height
   - Added `PreferredPixelSize` for all body parts

**Tests updated:**
- `TestQuadrupedTemplate`: Changed rotation check from 90 to 0, added top-down position checks
- `TestBlobTemplate`: Updated part count check (2-3 parts allowed for optional head)
- `TestMechanicalTemplate`: Added top-down position checks
- `TestFlyingTemplate`: Added top-down position checks
- `TestTemplateProportions`: Updated ranges for Phase 45 proportions (head 8-20%, torso 30-50%, legs 35-55%)
- `TestPhase151EnhancedProportionalScaling`: Updated to expect `PreferredPixelSize` in `HumanoidTemplate()`
- `TestHumanoidTemplate`: Added Phase 45 proportion verification

**Testing checkpoint:** ✅ `go test ./pkg/rendering/sprites/... ./pkg/rendering/...` all pass.

---

### Step 3: Update Tile Generation for Top-Down ✅ COMPLETED

**Files modified:**
- `pkg/rendering/tiles/generator.go`
- `pkg/rendering/tiles/walls.go`
- `pkg/rendering/tiles/generator_test.go`
- `pkg/rendering/tiles/walls_test.go`

**Changes completed:**
1. ✅ In `generator.go`:
   - Scaled `brickWidth: 16` → `brickWidth: 32`
   - Scaled `brickHeight: 8` → `brickHeight: 16`
   - Scaled dot spacing from 6 → 12
   - Scaled dot radius from 1 → 2
   - Scaled checkerboard checkSize from 4 → 8
   - Scaled lines spacing from 4 → 8
   - Scaled grain frequency from 0.3 → 0.15 (halved for wider grain)
   - Scaled door frame thickness from 2 → 4

2. ✅ In `walls.go`:
   - Added `EnableHeightEdges` option to `EnhancedWallConfig`
   - Implemented `applyWallHeightEdges()` function for 3D wall projection:
     - Darker top edge (player-facing, shadow-casting)
     - Lighter bottom edge (visible top surface reflecting light)
     - Left/right edge shading for complete 3D effect
   - Edge thickness scales with tile size (height/16)

**Tests updated/added:**
- `TestGenerateEnhancedWall_WithHeightEdges`: Verifies height edge indicators work correctly
- `TestGenerateEnhancedWall_WithShadows`: Updated to disable height edges and avoid edge effects
- `TestDefaultEnhancedWallConfig`: Added check for EnableHeightEdges default value
- `TestGenerator_Generate64x64`: Verifies all tile types generate at 64×64
- `TestGenerator_PatternScaling`: Verifies patterns have visible variation

**Testing checkpoint:** ✅ `go test ./pkg/rendering/tiles/...` all pass.

---

### Step 4: Update Shape Generator ✅ COMPLETED

**Files modified:**
- `pkg/rendering/shapes/generator.go`
- `pkg/rendering/shapes/types.go`
- `pkg/rendering/shapes/generator_test.go`

**Changes completed:**
1. ✅ Updated default anti-aliasing:
   - Default: `AntiAlias: AntiAliasLow` (2× super-sampling for performance)
   - Added `Quality` field to Config for higher quality mode toggle

2. ✅ Added new top-down specific shapes (Phase 45):
   - `ShapeFootprint`: Humanoid feet visible from above (two elliptical foot shapes)
   - `ShapeShoulders`: Shoulder/torso silhouette (wide upper body tapering to waist)
   - `ShapeArmReach`: Extended arm positions (arms reaching horizontally with hands)

3. ✅ Updated shape type definitions:
   - Added 3 new shape types to ShapeType enum
   - Updated String() method for new types
   - Added cases to isInside() switch statement

**Tests updated:**
- `TestDefaultConfig`: Added checks for AntiAliasLow default and Quality field
- `TestShapeType_String`: Added new shape types (footprint, shoulders, armreach)
- `TestAllShapeTypes`: Updated to include all 27 shape types
- Added `TestNewShapes_Phase45`: Tests generation of new shapes
- Added `TestShapeType_String_Phase45`: Tests string representation of new types
- Added `TestShapeDeterminism_Phase45`: Tests deterministic generation
- Added `BenchmarkNewShapes_Phase45`: Performance benchmarks for new shapes

**Testing checkpoint:** ✅ All 27 shape types build and generate correctly at 64×64.

---

### Step 5: Update Particle System ✅ COMPLETED

**Files modified:**
- `pkg/rendering/particles/generator.go`
- `pkg/rendering/particles/types.go`
- `pkg/rendering/particles/behaviors.go`
- `pkg/rendering/particles/generator_test.go`

**Changes completed:**
1. ✅ Scaled particle sizes:
   - Updated `MinSize` default: 1.0 → 4.0
   - Updated `MaxSize` default: 3.0 → 16.0
   - Updated `SpreadX` and `SpreadY` defaults: 5.0 → 10.0 for larger viewport
   - Scaled particle spawn positions 2× in all generators

2. ✅ Added Z-layer awareness:
   - Added `ZLayer` type with 4 levels: `ZLayerGround`, `ZLayerEntity`, `ZLayerAbove`, `ZLayerSky`
   - Added `ZLayer` field to both `Config` and `Particle` structs
   - Assigned Z-layers per particle type:
     - Ground level (dust, debris, blood): `ZLayerGround`
     - Entity level (sparks): `ZLayerEntity`
     - Above entity (smoke, magic, flame, sparkles): `ZLayerAbove`
     - Sky level (embers, smoke plumes): `ZLayerSky`

3. ✅ Updated behavior for top-down perspective:
   - Added `BehaviorShrinkOnRise` for rising particles (smoke, embers, flames, smoke plumes)
   - Added `BehaviorGrowOnFall` for falling particles (blood, debris)
   - Implemented `applySizeScaling()` in behaviors.go:
     - Rising particles shrink to minimum 25% of initial size
     - Falling particles grow to maximum 150% of initial size
   - Added `InitialSize` field to track original size for scaling

4. ✅ Added helper methods:
   - `Config.toPhysicsConfig()` for converting Config to PhysicsConfig

**Tests added:**
- `TestZLayerAssignment`: Verifies all particle types have correct Z-layers
- `TestInitialSizeTracking`: Verifies InitialSize is set for all particle types
- `TestSizeScalingBehaviors`: Tests shrink-on-rise and grow-on-fall behaviors
- `TestZLayerConstants`: Verifies Z-layer ordering
- Updated `TestDefaultConfig`: Added checks for Phase 45 scaled values

**Testing checkpoint:** ✅ All 10 particle types generate correctly with new scale and Z-layers.
`go test ./pkg/rendering/particles/...` passes.

---

### Step 6: Update Animation System ✅ COMPLETED

**Files modified:**
- `pkg/rendering/animation/articulation.go`
- `pkg/rendering/animation/controller.go`
- `pkg/rendering/animation/articulation_test.go`

**Changes completed:**
1. ✅ Updated articulation configuration for 64×64 sprites:
   - Scaled `ArmOffsetMax: 3.0` → `6.0` (2× for 64×64)
   - Scaled `LegOffsetMax: 4.0` → `8.0` (2× for 64×64)
   - Scaled `HeadOffsetMax: 2.0` → `4.0` (2× for 64×64)
   - Scaled `TailOffsetMax: 5.0` → `10.0` (2× for 64×64)

2. ✅ Scaled all hardcoded pixel offsets in articulation functions:
   - Idle animation: head bob, torso expansion, arm movement (2× scaling)
   - Walk/run animation: head bob, torso bob (2× scaling)
   - Attack animation: arm X offsets for wind-up/strike/follow-through (2× scaling)
   - Death animation: all body part Y offsets (2× scaling)
   - Jump animation: crouch depth, arc height, landing compression (2× scaling)
   - Cast animation: head concentration height (2× scaling)
   - Hit animation: knockback distance (2× scaling)

3. ✅ Added directional head rotation for direction indication:
   - `calculateDirectionalHeadRotation()` function added
   - East/West: full rotation toward direction
   - Diagonal: partial rotation (0.5-0.8× max)
   - North/South: no rotation (facing camera/away)

4. ✅ Added arm position changes per direction:
   - `calculateDirectionalArmOffsets()` function added
   - Arms extend toward movement direction
   - Exaggerated for running (1.3× multiplier)

5. ✅ Updated controller padding for 64×64:
   - `applyArticulation()` padding: 10 → 20 pixels

**Tests updated/added:**
- `TestDefaultArticulationConfig`: Updated for 64×64 scaled values
- `TestCalculateIdleArticulation`: Updated motion limits for 64×64
- `TestDirectionalHeadRotation`: New test for directional head rotation
- `TestDirectionalArmOffsets`: New test for arm position changes
- `TestWalkArticulationDirectional`: New test for directional walk enhancements
- `TestPhase45AnimationScaling`: New test verifying 64×64 scaling

**Testing checkpoint:** ✅ `go test ./pkg/rendering/animation/...` all pass.
`go test ./pkg/rendering/...` all 16 packages pass.

---

### Step 7: Update UI Components ✅ COMPLETED

**Files modified:**
- `pkg/rendering/ui/generator.go`
- `pkg/rendering/ui/hierarchy.go`
- `pkg/rendering/ui/decorations.go`
- `pkg/rendering/ui/generator_test.go`
- `pkg/rendering/ui/hierarchy_test.go`

**Changes completed:**
1. ✅ Scaled UI element sizes (2× for 64×64):
   - Icons: Padding 2 → 4, border 1 → 2
   - Borders: All hierarchy levels scaled 2× (1→2, 2→4, 3→6)
   - Button highlight positions: 2,2 → 4,4
   - Panel border: 1 → 2
   - HealthBar padding: 2 → 4, total 4 → 8
   - Label padding: 1 → 2
   - Frame corner sizes: 4 → 8, border 3 → 6

2. ✅ Updated hierarchy border thicknesses:
   - Primary: 3 → 6
   - Secondary: 2 → 4
   - Tertiary: 1 → 2
   - Quaternary: 1 → 2

3. ✅ Updated decorations for 64×64:
   - Ornate corner size: 8 → 16
   - Center decoration size: 4 → 8
   - Tech angular cut size: 6 → 12
   - Weathered border: 3 → 6, range 5 → 10

4. ✅ Updated separator dimensions:
   - Dashed: dashLength 8→16, gapLength 4→8
   - Dotted: dotSize 2→4, gapSize 4→8
   - Gradient: line thickness 2→4
   - Decorative diamond size: 3 → 6

5. ✅ Updated selectBorderThickness:
   - Frame: 3 → 6
   - Ornate genres (fantasy/horror): 3 → 6
   - Default: 2 → 4

**Tests added:**
- `TestGetHierarchyStyle_Phase45BorderThickness`: Verifies scaled border thicknesses
- `TestPhase45_UIScaling`: Tests icon, health bar, and frame scaling
- `TestPhase45_SelectBorderThickness`: Tests genre-based border selection
- `TestPhase45_GenerateAt64x64`: Verifies all UI elements generate at 64×64

**Testing checkpoint:** ✅ `go test ./pkg/rendering/ui/...` all pass.

---

### Step 8: Update Cache and Pool Systems ✅ COMPLETED

**Files modified:**
- `pkg/rendering/pool/image_pool.go`
- `pkg/rendering/sprites/cache.go`
- `pkg/rendering/sprites/pool.go`
- `pkg/rendering/cache/sprite_cache.go`

**Changes completed:**
1. ✅ Updated pool size constants:
   - Added `SizeDefault = 64` constant (Phase 45 standard)
   - `SizeMedium` is now an alias for `SizeDefault` for compatibility
   - Updated comments to reflect 64×64 as the new default
   - Added memory calculation documentation (32×32=4KB, 64×64=16KB, 128×128=64KB)

2. ✅ Added cache size constants and documentation:
   - Added `BytesPerPixel`, `SpriteSize32`, `SpriteSize64`, `SpriteSize128` constants
   - Added `DefaultCacheSize = 16MB` (~1000 64×64 sprites)
   - Added `MaxCacheSize = 300MB` (~18,000 64×64 sprites)
   - Updated `NewCache()` documentation with Phase 45 memory calculations
   - Updated `NewSpriteCache()` documentation with recommended sizes

3. ✅ Added sprite pool size constants:
   - Added `DefaultSpriteSize = 64`, `SmallSpriteSize = 32`, `LargeSpriteSize = 128`
   - Updated `ImagePool` documentation for Phase 45

**Tests added:**
- `TestPhase45_SizeConstants`: Verifies pool size constants
- `TestPhase45_DefaultSizePooling`: Verifies 64×64 pooling works
- `TestPhase45_MemoryCalculation`: Verifies memory size calculations
- `TestPhase45_SpriteSizeConstants`: Verifies cache size constants
- `TestPhase45_CacheSizeConstants`: Verifies cache capacity constants
- `TestPhase45_CacheWith64x64Sprites`: Verifies cache eviction with 64×64 sprites
- `TestPhase45_DefaultCacheCapacity`: Verifies default cache can hold many sprites
- `BenchmarkImagePool_GetPut_Default64`: Benchmark for 64×64 pooling

**Testing checkpoint:** ✅ All tests pass. Benchmarks show 64×64 pool operations at ~382ns/op with only 3 allocations.

---

### Step 9: Update Lighting and Post-Processing

**Files to modify:**
- `pkg/rendering/lighting/system.go`
- `pkg/rendering/lighting/ambient_occlusion.go`
- `pkg/rendering/postprocess/processor.go`

**Changes required:**
1. Scale lighting calculations for 64×64 tiles
2. Update ambient occlusion radius for larger sprites
3. Adjust bloom and other post-process effects

**Testing checkpoint:** Verify lighting effects work at new scale.

---

### Step 10: Update Tests

**Files to modify:**
- All `*_test.go` files in `pkg/rendering/`

**Changes required:**
1. Update dimension assertions from 32 → 64
2. Add new tests for top-down perspective validation
3. Update benchmark baselines

**Testing checkpoint:** Full test suite passes with new dimensions.

## 4. Legacy Code Removal

**Note:** No functions need deletion; all are enhanced in place.

### References to Update

| Location | Current | New |
|----------|---------|-----|
| `sprites/types.go:81-82` | `Width: 32, Height: 32` | `Width: 64, Height: 64` |
| `tiles/types.go:115-116` | `Width: 32, Height: 32` | `Width: 64, Height: 64` |
| `shapes/types.go:168-169` | `Width: 32, Height: 32` | `Width: 64, Height: 64` |
| All test files | `32x32` assertions | `64x64` assertions |

## 5. Validation Criteria

### Dimension & Visual Quality
- [x] All sprites/tiles/shapes generate at 64×64 by default
- [x] Anti-aliasing: Low (2× super-sampling) by default, Medium (4×) via Quality toggle
- [ ] Silhouette recognition ≥ 0.85 for humanoids
- [ ] Shadow opacity at 0.3 for ground-level shadows

### Top-Down Perspective
- [x] Entity heads in upper 20%, legs at 48% height
- [x] Shadows at ground level (RelativeY ≥ 0.90)
- [x] Walls show height with edge gradients

### Performance & Build
- [ ] Sprite generation <2ms, tile <1ms for 64×64
- [ ] Memory <300MB, frame rate ≥60 FPS with 2000 entities
- [ ] `go build ./...` and `go test ./pkg/rendering/...` pass
- [ ] All packages maintain ≥65% test coverage

## 6. Migration Order Summary

```
Phase 1: Types & Defaults (Step 1) ✅ COMPLETE
    └── ✅ Update default dimensions in all types.go files

Phase 2: Core Generation (Steps 2-4) ✅ COMPLETE
    ├── ✅ Sprite anatomy templates → top-down (Step 2 COMPLETE)
    ├── ✅ Tile patterns → scaled + 3D hints (Step 3 COMPLETE)
    └── ✅ Shape primitives → anti-aliased + scaled (Step 4 COMPLETE)

Phase 3: Effects & Animation (Steps 5-6) ✅ COMPLETE
    ├── ✅ Particle scaling + Z-layers (Step 5 COMPLETE)
    └── ✅ Animation frame updates + directional enhancements (Step 6 COMPLETE)

Phase 4: UI & Infrastructure (Steps 7-8) ✅ COMPLETE
    ├── ✅ UI component scaling (Step 7 COMPLETE)
    └── ✅ Cache/pool capacity adjustments (Step 8 COMPLETE)

Phase 5: Polish & Testing (Steps 9-10)
    ├── Lighting + post-process updates
    └── Full test suite updates
```

Each phase maintains a working build. Run tests after each step to ensure no regressions.
