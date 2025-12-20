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

### Step 1: Update Default Dimensions in Types

**Files to modify:**
- `pkg/rendering/sprites/types.go` (lines 81-82)
- `pkg/rendering/tiles/types.go` (lines 115-116)
- `pkg/rendering/shapes/types.go` (lines 168-169)

**Changes required:**
1. Update `DefaultConfig()` in `sprites/types.go`:
   - Change `Width: 32` → `Width: 64`
   - Change `Height: 32` → `Height: 64`
2. Update `DefaultConfig()` in `tiles/types.go`:
   - Change `Width: 32` → `Width: 64`
   - Change `Height: 32` → `Height: 64`
3. Update `DefaultConfig()` in `shapes/types.go`:
   - Change `Width: 32` → `Width: 64`
   - Change `Height: 32` → `Height: 64`

**Testing checkpoint:** Run `go test ./pkg/rendering/sprites/... ./pkg/rendering/tiles/... ./pkg/rendering/shapes/...` to verify dimension changes don't break generation.

---

### Step 2: Migrate Sprite Anatomy Templates to Top-Down

**Files to modify:**
- `pkg/rendering/sprites/anatomy_template.go`

**Changes required:**
1. Update `HumanoidTemplate()` (lines 168-229) for 64×64 base:
   - Change head `RelativeY: 0.25` → `RelativeY: 0.18`
   - Change head `RelativeHeight: 0.35` → `RelativeHeight: 0.125`
   - Add `PreferredPixelSize{Width: 8, Height: 8}` for pixel-perfect head
   - Update torso to bean shape with shoulders
   - Adjust legs to 48% height proportion

2. Leverage existing `Enhanced64HumanoidTemplate()` (line 2457+):
   - This template already implements Phase 45 proportions (12%/40%/48%)
   - Make it the default for 64×64 sprite generation
   - Update `SelectTemplate64()` to use Enhanced64 for all 64+ sprites

3. Update `QuadrupedTemplate()` (lines 371-435):
   - Adjust body to horizontal ellipse (top-down view of back)
   - Add visible leg positions (4 corner points)
   - Reduce head prominence from above

4. Update all creature templates similarly:
   - `BlobTemplate()`: Circular from above
   - `FlyingTemplate()`: Wing spread visible from above
   - `MechanicalTemplate()`: Angular top-down silhouette

**Testing checkpoint:** Generate sprites at 64×64, validate body part positions are correct for top-down view.

---

### Step 3: Update Tile Generation for Top-Down

**Files to modify:**
- `pkg/rendering/tiles/generator.go`
- `pkg/rendering/tiles/walls.go`
- `pkg/rendering/tiles/parallax.go`

**Changes required:**
1. In `generator.go`:
   - Scale pattern sizes for 64×64 tiles
   - Update `brickWidth: 16` → `brickWidth: 32`
   - Update `brickHeight: 8` → `brickHeight: 16`
   - Update dot spacing from 6 → 12

2. In `walls.go`:
   - Add 3D projection for wall tops (visible from above)
   - Implement shadow casting on adjacent floor tiles
   - Create corner/edge transition tiles

3. Add wall height indicator:
   - Darker top edge (player-facing)
   - Lighter bottom edge (visible top surface)

**Testing checkpoint:** Generate all tile types at 64×64, verify patterns scale correctly.

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
- [ ] All sprites/tiles/shapes generate at 64×64 by default
- [ ] Anti-aliasing: Medium (4× super-sampling)
- [ ] Silhouette recognition ≥ 0.85 for humanoids
- [ ] Shadow opacity at 0.3 for ground-level shadows

### Top-Down Perspective
- [ ] Entity heads in upper 20%, legs at 48% height
- [ ] Shadows at ground level (RelativeY ≥ 0.90)
- [ ] Walls show height with edge gradients

### Performance & Build
- [ ] Sprite generation <2ms, tile <1ms for 64×64
- [ ] Memory <300MB, frame rate ≥60 FPS with 2000 entities
- [ ] `go build ./...` and `go test ./pkg/rendering/...` pass
- [ ] All packages maintain ≥65% test coverage

## 6. Migration Order Summary

```
Phase 1: Types & Defaults (Step 1)
    └── Update default dimensions in all types.go files

Phase 2: Core Generation (Steps 2-4)
    ├── Sprite anatomy templates → top-down
    ├── Tile patterns → scaled + 3D hints
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
