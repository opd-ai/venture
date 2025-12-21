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

### Step 4: Update Shape Generator

**Files to modify:**
- `pkg/rendering/shapes/generator.go`
- `pkg/rendering/shapes/types.go`

**Changes required:**
1. Update all shape radius calculations:
   - Replace `radius := math.Min(cx, cy) * 0.8` with dynamic scaling
   - Add top-down perspective distortion for 3D shapes

2. Enable anti-aliasing conditionally:
   - Default: `AntiAlias: AntiAliasLow` (2× super-sampling for performance)
   - Quality mode: `AntiAlias: AntiAliasMedium` (4× super-sampling)
   - Add performance toggle in `shapes/types.go:Config.Quality` field

3. Add new top-down specific shapes:
   - `ShapeFootprint`: Humanoid feet visible from above
   - `ShapeShoulders`: Shoulder/torso silhouette
   - `ShapeArmReach`: Extended arm positions

**Testing checkpoint:** Generate all 24 shape types at 64×64, verify edge quality.

---

### Step 5: Update Particle System

**Files to modify:**
- `pkg/rendering/particles/generator.go`
- `pkg/rendering/particles/types.go`

**Changes required:**
1. Scale particle sizes:
   - Update `MinSize` and `MaxSize` defaults × 2
   - Adjust spread values for larger viewport

2. Add Z-layer awareness:
   - Particles above entities render with higher Z-index
   - Ground-level particles (dust, debris) use lower Z-index

3. Update behavior for top-down:
   - Rising particles (smoke, embers): Shrink + fade as they rise
   - Falling particles: Grow slightly before impact

**Testing checkpoint:** Generate all 10 particle types, verify visual quality at new scale.

---

### Step 6: Update Animation System

**Files to modify:**
- `pkg/rendering/animation/controller.go`
- `pkg/rendering/animation/direction.go`

**Changes required:**
1. Update frame dimensions for 64×64:
   - Adjust animation frame positioning
   - Scale walk cycle pixel offsets

2. Enhance directional sprites:
   - Add subtle head rotation for direction indication
   - Arm position changes per direction

**Testing checkpoint:** Test 4-directional animation at 64×64.

---

### Step 7: Update UI Components

**Files to modify:**
- `pkg/rendering/ui/generator.go`
- `pkg/rendering/ui/types.go`

**Changes required:**
1. Scale UI element sizes:
   - Icons: 16×16 → 32×32
   - Borders: 1px → 2px
   - Padding: proportional increase

2. Update UI sprite generation for clarity at new scale.

**Testing checkpoint:** Verify UI elements render correctly at new dimensions.

---

### Step 8: Update Cache and Pool Systems

**Files to modify:**
- `pkg/rendering/sprites/cache.go`
- `pkg/rendering/sprites/pool.go`
- `pkg/rendering/cache/sprite_cache.go`

**Changes required:**
1. Update pool size categories:
   - Add 64×64 pool tier
   - Update memory limits for larger sprites

2. Adjust cache capacity:
   - 64×64 RGBA = 16KB per sprite (64×64×4 bytes)
   - 300MB limit ÷ 16KB = ~18,750 max cached 64×64 sprites
   - Reduce max cached sprites from current limit to maintain <300MB

**Testing checkpoint:** Run cache benchmarks, verify memory usage stays under limits.

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
- [ ] Anti-aliasing: Medium (4× super-sampling)
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

Phase 2: Core Generation (Steps 2-4)
    ├── ✅ Sprite anatomy templates → top-down (Step 2 COMPLETE)
    ├── ✅ Tile patterns → scaled + 3D hints (Step 3 COMPLETE)
    └── Shape primitives → anti-aliased + scaled

Phase 3: Effects & Animation (Steps 5-6)
    ├── Particle scaling + Z-layers
    └── Animation frame updates

Phase 4: UI & Infrastructure (Steps 7-8)
    ├── UI component scaling
    └── Cache/pool capacity adjustments

Phase 5: Polish & Testing (Steps 9-10)
    ├── Lighting + post-process updates
    └── Full test suite updates
```

Each phase maintains a working build. Run tests after each step to ensure no regressions.
