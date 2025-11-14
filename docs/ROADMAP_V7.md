# Development Roadmap - Version 7.0: Advanced Visual Improvements

## Overview

**Project:** Venture - Fully Procedural Multiplayer Action-RPG  
**Version:** 7.0 - Advanced Visual Improvements  
**Previous Version:** 6.0 Complete (Phase 42 - Projected 2028)  
**Timeline:** 8-10 months (6 phases)  
**Date:** November 2025  
**Focus:** Enhanced visual fidelity through display scaling, improved sprite anatomy, and refined collision rendering

---

## Objective & Scope

**What V7 Delivers:**
- Larger default display resolution (1280x720 → 1920x1080) with scalable UI
- Enhanced aerial-view sprite anatomy with 40%+ improved silhouette recognition
- Smoothed wall rendering with anti-aliased edges and corner blending
- Pixel-perfect wall collision detection with sub-pixel precision
- Improved visual clarity for high-DPI displays
- Performance optimizations for larger viewports

**What V7 Does NOT Include:**
- 3D graphics or perspective rendering
- External asset dependencies (maintains 100% procedural generation)
- Breaking changes to existing sprite/tile generation APIs
- New gameplay mechanics or content systems
- VR/AR support

**Completion Criteria:** All features functional with ≥65% test coverage, 60+ FPS maintained at 1920x1080, <500MB memory, backward compatible with v6.0.

---

## Key Constraints & Compatibility

### Performance Targets

**Resolution Scaling:**
- Default: 1920x1080 (2.25x pixel count vs 1280x720)
- Maintain 60 FPS minimum (current: 106 FPS at 1280x720, target: 60+ FPS at 1920x1080)
- Memory budget: <500MB (current: 73MB, expect ~150-200MB with larger sprites/tiles)
- Sprite cache efficiency: Maintain ≥90% hit rate (current: 95.9%)
- UI scaling: Dynamic element sizing based on resolution

**Backward Compatibility:**
- Support legacy resolutions via `-width` and `-height` flags
- Existing save files load without migration
- Multiplayer sync with mixed-resolution clients
- All v6.0 features preserved

### Visual Quality Targets

**Sprite Improvements:**
- Silhouette recognition score: 0.75 → 0.85+ (15.6% improvement)
- Anatomical detail: Leverage Phase 15.1 enhanced templates
- Anti-aliasing: Super-sampling for smooth edges
- Genre-specific variations: Use existing genre anatomy system
- Animation fluidity: Maintain existing 8-direction support

**Wall Rendering:**
- Edge smoothing: Anti-aliased transitions between wall/floor
- Corner blending: Seamless diagonal and 90° corner rendering
- Pattern continuity: Texture alignment across wall segments
- Shadow integration: Proper shadow casting on smoothed walls

**Collision Precision:**
- Sub-pixel accuracy: 0.1 pixel tolerance (current: 1 pixel)
- Wall edge detection: Smooth collision boundaries matching visual rendering
- Performance: Maintain quadtree spatial partitioning efficiency

---

## Architecture Overview

### Display Scaling System

```
Resolution Pipeline:
  User Config (-width/-height flags)
       ↓
  Display Manager (pkg/rendering/display/)
       ↓
  ┌─────────────┴─────────────┐
  ↓                           ↓
Viewport Scaler          UI Scaler
(sprites/tiles)      (menus/HUD)
  ↓                           ↓
Sprite Generator         Font Renderer
(larger base size)     (dynamic sizing)
  ↓                           ↓
  └─────────────┬─────────────┘
                ↓
        Final Framebuffer
           (1920x1080)
```

**Key Components:**
- `pkg/rendering/display/manager.go` - Resolution management
- `pkg/rendering/display/scaler.go` - Viewport/UI scaling logic
- `pkg/rendering/sprites/generator.go` - Enhanced for larger base sprites
- `pkg/rendering/ui/scaler.go` - Dynamic UI element sizing

### Enhanced Sprite Anatomy

**Existing Systems (Phase 15.1):**
- `pkg/rendering/sprites/anatomy_template.go` - EnhancedHumanoidTemplate, DetailedHumanoidTemplate
- `pkg/rendering/sprites/silhouette.go` - Silhouette analysis and scoring
- `pkg/rendering/sprites/composite.go` - Multi-layer sprite composition

**V7 Enhancements:**
- Increased base sprite size: 32x32 → 64x64 pixels (4x detail)
- Refined body part proportions (head, torso, limbs)
- Genre-specific anatomical variations (fantasy vs sci-fi vs horror)
- Improved color/pattern application from `pkg/rendering/palette/` and `pkg/rendering/patterns/`

### Wall Rendering & Collision

**Rendering Improvements:**
- Use `pkg/rendering/patterns/generator.go` for edge anti-aliasing
- Enhance `pkg/rendering/tiles/generator.go` with smoothing filters
- Corner detection and blending (L-joints, T-joints, cross-joints)
- Shadow-aware edge rendering (darker edges in shadowed areas)

**Collision Enhancements:**
- Refine `pkg/engine/collision.go` with sub-pixel precision
- Edge normal calculation for smooth wall sliding
- Quadtree optimization for larger worlds (maintain O(log n) queries)
- Integration with smoothed visual rendering

---

## Phase 37: Display Resolution Foundation (Weeks 1-4)

**Status:** Not Started  
**Duration:** 4 weeks  
**Focus:** Increase default resolution and implement scalable UI framework

### 37.1: Resolution Management System

**Deliverables:**
- [ ] Create `pkg/rendering/display/` package with manager, scaler, config
- [ ] Implement dynamic resolution detection (system default, user override)
- [ ] Add `-width`, `-height`, `-fullscreen` CLI flags with validation
- [ ] Support common resolutions: 1280x720, 1920x1080, 2560x1440, 3840x2160
- [ ] Backward compatibility: Default to 1920x1080, allow legacy 1280x720

**Technical Approach:**
- Display manager reads system capabilities via Ebiten API
- Config struct stores target resolution, scaling mode (fit/fill/stretch)
- Resolution change triggers viewport recalculation and cache invalidation
- Testing: Unit tests for scaling calculations, integration tests for resolution changes

**Success Metrics:**
- All resolutions render without distortion
- Aspect ratio preservation in fit/fill modes
- <50ms resolution switch time (cache rebuild)

### 37.2: UI Scaling Framework

**Deliverables:**
- [ ] Create `pkg/rendering/ui/scaler.go` for dynamic element sizing
- [ ] Scale fonts: 12pt → dynamic based on resolution (16pt at 1920x1080)
- [ ] Scale UI elements: buttons, panels, health bars, inventory slots
- [ ] Maintain readability: Min font size 10pt, max 24pt
- [ ] Update all 7 UI menus: inventory, character, skills, quests, map, crafting, shop

**Technical Approach:**
- Base UI elements on reference resolution (1920x1080)
- Scaling factor = current_height / reference_height
- Apply to font sizes, button dimensions, spacing, borders
- Cache scaled UI elements per resolution

**Success Metrics:**
- UI readable at all supported resolutions
- Element proportions consistent across resolutions
- <100ms UI initialization time per menu

**Performance Target:** 60 FPS at 1920x1080, <150MB memory overhead

**Testing Requirements:**
- [ ] Unit tests: Scaling calculations, boundary conditions
- [ ] Integration tests: UI rendering at multiple resolutions
- [ ] Visual regression tests: Compare screenshots across resolutions
- [ ] Coverage: ≥65% for `pkg/rendering/display/` and `pkg/rendering/ui/scaler.go`

---

## Phase 38: Viewport & Cache Optimization (Weeks 5-8)

**Status:** Not Started  
**Duration:** 4 weeks  
**Focus:** Optimize rendering pipeline for larger viewports and sprite sizes

### 38.1: Viewport Culling Enhancement

**Deliverables:**
- [ ] Update viewport culling for 2.25x larger screen area (1280x720 → 1920x1080)
- [ ] Enhance quadtree queries in `pkg/engine/` for efficient spatial partitioning
- [ ] Implement frustum culling with margin (render off-screen entities within 1 tile)
- [ ] Optimize batch rendering for larger entity counts per frame

**Technical Approach:**
- Viewport bounds calculated from camera position + resolution
- Quadtree range query returns entities within viewport + margin
- Batch renderer groups entities by sprite type/texture for fewer draw calls
- Existing Phase 3 optimizations: culling (1,635x speedup), batching (1,667x speedup)

**Success Metrics:**
- Viewport culling efficiency: <5% entities rendered outside view
- Quadtree query time: <0.1ms for 2000 entities
- Draw call reduction: <50 batches per frame (down from unbatched ~500)

### 38.2: Sprite Cache Scaling

**Deliverables:**
- [ ] Enhance `pkg/rendering/sprites/cache.go` for larger sprite sizes (32x32 → 64x64)
- [ ] Implement LRU eviction policy with size-based limits (max 1000 cached sprites)
- [ ] Add memory monitoring: Evict least-used sprites when >300MB cache size
- [ ] Maintain ≥90% cache hit rate (current: 95.9%)

**Technical Approach:**
- Cache key includes resolution + sprite params (type, genre, seed)
- LRU tracking with access timestamps
- Memory estimation: 64x64 RGBA = 16KB per sprite, 1000 sprites = 16MB baseline
- Batch pre-generation for common sprites (player, basic monsters)

**Success Metrics:**
- Cache hit rate: ≥90% during normal gameplay
- Memory usage: <300MB sprite cache at steady state
- Cache lookup time: <0.01ms (hash map access)

**Performance Target:** Maintain 60 FPS at 1920x1080 with 2000 entities

**Testing Requirements:**
- [ ] Benchmark tests: Viewport culling, cache performance
- [ ] Load tests: 5000 entities, cache thrashing scenarios
- [ ] Memory profiling: Verify <500MB total usage
- [ ] Coverage: ≥65% for optimized systems

---

## Phase 39: Enhanced Sprite Anatomy (Weeks 9-13)

**Status:** Not Started  
**Duration:** 5 weeks  
**Focus:** Improve aerial-view sprite anatomical detail and recognition

### 39.1: Refined Anatomical Templates

**Deliverables:**
- [ ] Enhance `pkg/rendering/sprites/anatomy_template.go` with higher-detail templates
- [ ] Increase base sprite size: 32x32 → 64x64 pixels for 4x detail capacity
- [ ] Refine body part proportions: Head (12% body height), torso (40%), legs (48%)
- [ ] Add secondary details: Shoulders, neck, feet definition
- [ ] Genre-specific variations: Bulky (fantasy warrior), sleek (cyberpunk), twisted (horror)

**Technical Approach:**
- Leverage existing EnhancedHumanoidTemplate and DetailedHumanoidTemplate
- Scale proportions for 64x64 canvas (head: 8px, torso: 26px, legs: 30px)
- Use Phase 15.1 super-sampling for anti-aliased edges
- Apply genre-specific anatomy functions for thematic variation

**Success Metrics:**
- Silhouette recognition score: 0.75 → 0.85+ (measured via `pkg/rendering/sprites/silhouette.go`)
- Body part distinctiveness: Head, torso, limbs identifiable at 64x64
- Genre variation: ≥5 distinct templates per genre (fantasy, sci-fi, horror, cyberpunk, post-apoc)

### 39.2: Color & Pattern Integration

**Deliverables:**
- [ ] Apply genre-specific palettes from `pkg/rendering/palette/` to 64x64 sprites
- [ ] Use `pkg/rendering/patterns/generator.go` for armor/clothing texture detail
- [ ] Implement multi-layer composition: Base skin → clothing → armor → accessories
- [ ] Add shading gradients for depth perception (lighter top, darker bottom)

**Technical Approach:**
- Composite pipeline: Anatomical base → color fill → pattern overlay → shading
- Pattern scaling: Adjust pattern frequency for 64x64 canvas
- Genre palettes define primary/secondary/accent colors
- Shading: 20% luminance gradient from top to bottom

**Success Metrics:**
- Visual clarity: Sprites distinguishable at 64x64 without labels
- Pattern visibility: Textures readable at base size, no aliasing
- Genre theming: 90%+ player recognition of genre from sprite alone

**Performance Target:** <2ms sprite generation time (64x64 vs <1ms for 32x32)

**Testing Requirements:**
- [ ] Silhouette tests: Verify ≥0.85 recognition score
- [ ] Visual quality tests: Manual inspection of generated sprites
- [ ] Genre variation tests: Generate 100 sprites per genre, verify diversity
- [ ] Performance benchmarks: Sprite generation time, cache hit rate
- [ ] Coverage: ≥65% for sprite generation pipeline

---

## Phase 40: Animation Fluidity (Weeks 14-17)

**Status:** Not Started  
**Duration:** 4 weeks  
**Focus:** Improve sprite animation quality at larger sizes

### 40.1: Enhanced Animation Frames

**Deliverables:**
- [ ] Refine `pkg/rendering/sprites/animation.go` for 64x64 sprite animations
- [ ] Maintain 8-direction support (N, NE, E, SE, S, SW, W, NW)
- [ ] Add subtle body part articulation (arm swing, leg stride)
- [ ] Implement frame interpolation for smoother transitions (4 frames → 8 frames)

**Technical Approach:**
- Animation system uses existing walk/attack/idle states
- Body part offset calculations: Arms ±3px, legs ±4px per frame
- Frame interpolation: Linear blend between keyframes
- Direction-specific mirroring for E/W symmetry

**Success Metrics:**
- Animation smoothness: 8 frames per cycle (up from 4)
- Movement fluidity: No visible "popping" between frames
- Direction consistency: All 8 directions have equivalent quality

### 40.2: Performance Optimization

**Deliverables:**
- [ ] Implement sprite pooling for animation frames in `pkg/rendering/sprites/pool.go`
- [ ] Cache animated sequences (not just static sprites)
- [ ] Optimize frame generation: Pre-compute common animations
- [ ] Maintain 60 FPS with 100 animated entities on screen

**Technical Approach:**
- Pool reuses image buffers for animation frames
- Animation cache key: entity_type + state + direction + frame
- Pre-generation: Player + 10 common monsters at startup
- Frame delta optimization: Only redraw changed pixels

**Success Metrics:**
- Animation cache hit rate: ≥85%
- Frame generation time: <1ms per frame (vs <0.5ms for static)
- Memory overhead: <50MB for animation frame cache

**Performance Target:** 60 FPS at 1920x1080 with 100 animated entities

**Testing Requirements:**
- [ ] Animation quality tests: Visual inspection of all states/directions
- [ ] Performance benchmarks: Frame generation, cache efficiency
- [ ] Integration tests: Animations in gameplay context
- [ ] Coverage: ≥65% for animation systems

---

## Phase 41: Wall Rendering Improvements (Weeks 18-21)

**Status:** Not Started  
**Duration:** 4 weeks  
**Focus:** Smooth wall edges and enhance corner rendering

### 41.1: Edge Anti-Aliasing

**Deliverables:**
- [ ] Implement edge smoothing in `pkg/rendering/tiles/generator.go`
- [ ] Use `pkg/rendering/patterns/generator.go` for texture-aware anti-aliasing
- [ ] Apply 2x2 super-sampling for wall edges (render at 2x, downsample)
- [ ] Blend wall/floor boundary pixels (50% wall color, 50% floor color)

**Technical Approach:**
- Detect wall edges via neighbor tile analysis (wall adjacent to floor)
- Super-sampling: Render wall at 128x128, downsample to 64x64
- Gaussian blur on edge pixels (1-pixel radius)
- Pattern-aware: Preserve texture detail while smoothing edges

**Success Metrics:**
- Edge smoothness: No visible jagged pixels at 1920x1080
- Performance: <0.5ms per wall tile (vs <0.3ms unsmoothed)
- Visual quality: 90%+ player preference vs unsmoothed in blind test

### 41.2: Corner Blending

**Deliverables:**
- [ ] Detect corner types: L-joints (90°), T-joints (3-way), cross-joints (4-way)
- [ ] Generate corner blend tiles for seamless transitions
- [ ] Support diagonal walls (existing Phase 11 system) with smoothed corners
- [ ] Integrate with existing shadow system from Phase 17

**Technical Approach:**
- Corner detection: Count neighboring wall tiles (2 = L, 3 = T, 4 = cross)
- Blend radius: 4-pixel corner rounding for smooth transitions
- Diagonal support: 45° angle blending with adjacent orthogonal walls
- Shadow casting: Update normal calculations for smoothed corners

**Success Metrics:**
- Corner continuity: No visible seams at wall intersections
- Diagonal quality: Smooth transitions at 45° angles
- Shadow accuracy: Proper shadow casting from smoothed geometry

**Performance Target:** <1ms per corner tile, 60 FPS maintained

**Testing Requirements:**
- [ ] Visual quality tests: Manual inspection of wall/corner rendering
- [ ] Edge case tests: All corner types (L, T, cross, diagonal combinations)
- [ ] Performance benchmarks: Tile generation time with smoothing
- [ ] Coverage: ≥65% for tile rendering enhancements

---

## Phase 42: Pixel-Perfect Collision (Weeks 22-26)

**Status:** Not Started  
**Duration:** 5 weeks  
**Focus:** Sub-pixel collision precision matching visual rendering

### 42.1: Sub-Pixel Collision Detection

**Deliverables:**
- [ ] Enhance `pkg/engine/collision.go` with 0.1-pixel precision (vs current 1-pixel)
- [ ] Implement smooth wall sliding along collision boundaries
- [ ] Calculate edge normals for realistic bounce/slide physics
- [ ] Integrate with smoothed wall rendering from Phase 41

**Technical Approach:**
- Collision bounds: Float64 positions (vs previous Int rounding)
- Edge normals: Perpendicular vectors from wall surface
- Wall sliding: Velocity projection along wall surface (dot product)
- Quadtree precision: Store float64 bounds, 0.1-unit query tolerance

**Success Metrics:**
- Collision accuracy: ±0.1 pixel tolerance (vs ±1 pixel)
- Wall sliding smoothness: No "sticky" corners or abrupt stops
- Performance: Maintain O(log n) collision queries via quadtree

### 42.2: Rendering Alignment

**Deliverables:**
- [ ] Align collision boundaries exactly with visual wall edges
- [ ] Update collision shapes for smoothed corners (rounded rectangles)
- [ ] Implement collision layer system (wall, entity, projectile, trigger)
- [ ] Visual debug mode: Render collision bounds overlay (dev flag `-debug-collision`)

**Technical Approach:**
- Collision bounds derived from tile rendering output
- Rounded corners: Use circle segments for corner collision
- Layer filtering: Quadtree segregation by collision layer
- Debug rendering: Draw collision polygons in red outline over game view

**Success Metrics:**
- Visual/collision alignment: <0.5 pixel discrepancy
- Layer separation: 0% false collisions between incompatible layers
- Debug mode performance: <5ms overhead per frame

**Performance Target:** 60 FPS with 2000 entities and 10,000 wall tiles

**Testing Requirements:**
- [ ] Collision accuracy tests: Sub-pixel boundary validation
- [ ] Integration tests: Movement system + enhanced collision
- [ ] Performance benchmarks: Quadtree queries with float64 precision
- [ ] Visual regression: Compare collision debug overlay with wall rendering
- [ ] Coverage: ≥65% for collision system enhancements

---

## Success Criteria & Acceptance Tests

### Visual Quality Validation

**Sprite Quality:**
- [ ] Silhouette recognition ≥0.85 (measured via automated scoring)
- [ ] Genre distinctiveness: 90%+ correct genre identification in user testing
- [ ] Animation smoothness: 8 frames per cycle, 60 FPS playback

**Wall Rendering:**
- [ ] Edge smoothness: No jagged pixels at 1920x1080
- [ ] Corner continuity: Seamless blending at all junction types
- [ ] Shadow integration: Proper casting on smoothed walls

**UI Scaling:**
- [ ] Readability: All text legible at supported resolutions (1280x720 to 3840x2160)
- [ ] Element proportions: Consistent across resolutions
- [ ] No clipping or overflow at any resolution

### Performance Validation

**Frame Rate:**
- [ ] 60 FPS minimum at 1920x1080 with 2000 entities
- [ ] 30 FPS minimum at 3840x2160 with 1000 entities
- [ ] Viewport culling efficiency: ≥95% off-screen entities excluded

**Memory Usage:**
- [ ] Total: <500MB at 1920x1080 during normal gameplay
- [ ] Sprite cache: <300MB with ≥90% hit rate
- [ ] Animation frames: <50MB overhead

**Load Times:**
- [ ] Resolution change: <50ms
- [ ] Sprite cache warm-up: <2s for common sprites
- [ ] UI initialization: <100ms per menu

### Compatibility Testing

**Backward Compatibility:**
- [ ] v6.0 save files load without errors
- [ ] Legacy resolution (1280x720) renders correctly
- [ ] Multiplayer sync with mixed-resolution clients
- [ ] All existing CLI flags functional

**Cross-Platform:**
- [ ] Desktop: Linux, macOS, Windows at 1920x1080
- [ ] WebAssembly: 1920x1080 in browser (Chrome, Firefox, Safari)
- [ ] Mobile: Adaptive scaling on iOS/Android (720p to 1440p)

**Determinism:**
- [ ] Sprite generation: Same seed produces identical sprites at all resolutions
- [ ] World generation: Resolution changes don't affect terrain/entities
- [ ] Multiplayer: Clients at different resolutions stay synchronized

---

## Risk Mitigation

### Performance Risks

**Risk:** 1920x1080 (2.25x pixels) causes FPS drop below 60  
**Mitigation:**
- Implement viewport culling early (Phase 38.1)
- Profile CPU/GPU bottlenecks before sprite enhancements
- Add quality settings: Low (1280x720), Medium (1920x1080), High (2560x1440)
- Fallback: Auto-reduce resolution if FPS <60 for 5 consecutive seconds

**Risk:** Sprite cache memory exceeds 500MB budget  
**Mitigation:**
- LRU eviction in Phase 38.2
- Memory monitoring with automatic cache trimming
- Lazy generation: Only create sprites when first needed
- Quality tiers: Cache low-res sprites for distant entities

### Visual Quality Risks

**Risk:** 64x64 sprites lack detail improvements  
**Mitigation:**
- User testing at Phase 39 midpoint (week 11)
- A/B comparison with 32x32 baseline
- Iterate on anatomy templates based on feedback
- Fallback: Keep 32x32 as "low quality" option

**Risk:** Smoothed walls increase rendering complexity  
**Mitigation:**
- Profile tile rendering before anti-aliasing implementation
- Cache smoothed wall tiles aggressively
- Implement LOD: Smooth walls only within 10-tile radius
- Toggle: `-smooth-walls=false` flag for performance mode

### Development Risks

**Risk:** Timeline exceeds 10 months  
**Mitigation:**
- Phase 37-38 are critical path (foundation), prioritize early
- Phases 39-40 (sprites) parallelizable with 41-42 (walls/collision)
- Reduce scope if needed: Skip animation fluidity (Phase 40)
- Maintain 2-week buffer for integration and polish

**Risk:** Breaking changes to sprite/tile APIs  
**Mitigation:**
- Maintain v6.0 API compatibility via wrapper functions
- Version sprite generation: `GenerateV6()` and `GenerateV7()` methods
- Gradual migration: Keep both systems until v7 stable
- Deprecation warnings: Log when using v6 APIs in v7 context

---

## Dependencies & Prerequisites

### Technical Dependencies

**Existing Systems (Must Remain Functional):**
- Phase 3: Viewport culling (1,635x speedup), batch rendering (1,667x speedup)
- Phase 15: Enhanced sprite anatomy templates, silhouette analysis
- Phase 17: Dynamic lighting and shadow system
- Phase 11: Diagonal wall rendering
- Quadtree spatial partitioning in `pkg/engine/`

**New Package Requirements:**
- `pkg/rendering/display/` - Resolution management (Phase 37)
- `pkg/rendering/ui/scaler.go` - UI scaling framework (Phase 37)
- Enhanced `pkg/rendering/sprites/anatomy_template.go` (Phase 39)
- Enhanced `pkg/rendering/tiles/generator.go` (Phase 41)
- Enhanced `pkg/engine/collision.go` (Phase 42)

### Development Environment

**Build Requirements:**
- Go 1.24.5+ (current: 1.24.7)
- Ebiten v2.9.3+ (for higher resolution support)
- 16GB RAM for development (large sprite compilation)
- 1920x1080 minimum display for testing

**Testing Infrastructure:**
- Visual regression testing framework (screenshot comparison)
- Performance profiling: `go test -cpuprofile`, `go test -memprofile`
- Multi-resolution test suite (1280x720, 1920x1080, 2560x1440, 3840x2160)
- Automated silhouette scoring in CI

---

## Timeline & Milestones

### Month 1-2: Foundation (Phases 37-38)
- Week 1-4: Display resolution system, UI scaling
- Week 5-8: Viewport optimization, sprite cache scaling
- **Milestone 1:** 1920x1080 rendering functional with 60 FPS

### Month 3-4: Sprites (Phases 39-40)
- Week 9-13: Enhanced anatomy templates, 64x64 sprite generation
- Week 14-17: Animation fluidity, performance optimization
- **Milestone 2:** Silhouette score ≥0.85, animations smooth

### Month 5-7: Walls & Collision (Phases 41-42)
- Week 18-21: Wall edge smoothing, corner blending
- Week 22-26: Sub-pixel collision, rendering alignment
- **Milestone 3:** Smooth walls with pixel-perfect collision

### Month 8-10: Integration & Polish
- Week 27-30: Cross-platform testing, performance tuning
- Week 31-34: Bug fixes, edge case handling
- Week 35-40: Documentation, user testing, final release
- **Milestone 4:** Version 7.0 release-ready

---

## Version 7.0 Completion Checklist

### Core Features
- [ ] 1920x1080 default resolution with scalable UI
- [ ] 64x64 sprite generation with ≥0.85 silhouette score
- [ ] Smoothed wall rendering with anti-aliased edges
- [ ] Sub-pixel collision detection (0.1-pixel precision)
- [ ] Performance: 60 FPS at 1920x1080, <500MB memory

### Quality Assurance
- [ ] Test coverage ≥65% per package (average 82.4% maintained)
- [ ] All acceptance tests passing
- [ ] Cross-platform validation (Linux, macOS, Windows, Web, Mobile)
- [ ] Backward compatibility with v6.0 confirmed
- [ ] Multiplayer sync with mixed resolutions verified

### Documentation
- [ ] Update ARCHITECTURE.md with display scaling system
- [ ] Update TECHNICAL_SPEC.md with sprite/collision enhancements
- [ ] Update USER_MANUAL.md with resolution options
- [ ] Create RELEASE_NOTES_V7.0.md
- [ ] Update API_REFERENCE.md with new display/scaling APIs

### Release Preparation
- [ ] Version bump: 6.0 → 7.0 in all manifests
- [ ] Build binaries for all platforms
- [ ] WebAssembly deployment to GitHub Pages
- [ ] Mobile builds (Android APK/AAB, iOS IPA)
- [ ] Tag release: `v7.0.0`

---

## Post-V7 Outlook

**Potential V8.0 Features (Not Committed):**
- Advanced particle physics (fluid simulation, cloth dynamics)
- Procedural music composition enhancements
- VR/AR experimental support
- Advanced AI behaviors (emergent squad tactics)
- Expanded crafting systems

**Long-term Vision:**
- Maintain zero-asset architecture indefinitely
- Continue performance leadership (target 120 FPS at 1920x1080)
- Expand to new genres (steampunk, medieval, western)
- Community content: Shareable world seeds, custom rulesets

---

**Document Status:** Draft  
**Last Updated:** November 14, 2025  
**Next Review:** Upon Phase 36 completion (V6.0 release)
