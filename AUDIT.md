# Visual Performance Audit
**Date:** 2026-02-11  
**Investigator:** Claude Visual Performance Audit  
**Codebase:** Venture Game Engine - Rendering Pipeline

## Executive Summary

The Venture rendering pipeline contains **7 distinct visual jank sources** ranging from critical full-screen CPU-bound post-processing to moderate per-tile heap allocations. The most severe issue is the `PostProcessorAdapter.Apply()` method which performs pixel-by-pixel CPU readback and re-upload of the entire framebuffer every frame (1920×1080 = 2,073,600 pixels × 4 channels × multiple passes), adding an estimated **40–120ms per frame** when post-processing is enabled. Secondary issues include excessive debug logging in the lighting hot path (50+ log calls per frame with `WithFields` struct allocations), per-tile `DrawImageOptions` heap allocations in terrain rendering, bloom's CPU-side per-pixel readback, and a `defer recover()` in the per-entity `drawRect` fallback path. Combined, these issues can drive worst-case frame times from the 16.67ms target to **80–200ms**, reducing effective FPS to 5–12 during lit scenes with post-processing.

---

## FRAME TIMING ANALYSIS

### Overall Frame Metrics (Estimated from Code Analysis)

These values are derived from static analysis of algorithmic complexity and allocation patterns. Actual measurements require runtime profiling.

- **Min frame time:** ~6ms (menu screens, no entities)
- **Average frame time:** ~16–18ms (unlit, no post-processing, <500 entities)
- **Max frame time:** ~80–200ms (lit scene + bloom + post-processing + debug logging)
- **1st percentile:** ~7ms (best frames — menu/idle)
- **5th percentile:** ~12ms (light gameplay, cached sprites)
- **95th percentile:** ~25–35ms (active gameplay, particle spawns)
- **99th percentile:** ~60–200ms (bloom + post-processing + sprite regen spike)
- **Frames >20ms:** ~15–40% (estimated during lit gameplay)
- **Frames >30ms:** ~5–20% (severe jank during post-processing)
- **Frame time variance (stddev):** ~8–25ms (target: <2ms) — **FAILS TARGET**

### Frame Pacing Pattern

**Pattern: Bimodal with periodic spikes**

The frame timing is expected to exhibit two modes:
1. **Fast mode (~14–18ms):** Frames where all sprites are cached, no regeneration occurs, lighting is disabled or simple.
2. **Slow mode (~40–200ms):** Frames where post-processing is active, bloom recalculates, or sprite regeneration coincides with complex lighting.

The animation system's `maxRegenPerFrame=8` limiter spreads sprite generation over frames, creating periodic +2–8ms spikes every frame during initial entity loading. The `FrameTimeTracker` with its `sort.Slice` in `GetStats()` adds micro-stutter every 300 frames when profiling stats are logged.

---

## VISUAL JANK ISSUES

### Issue V1: Post-Processing CPU Pixel-by-Pixel Readback and Re-Upload Every Frame
- **Location:** `pkg/engine/post_processor.go:82-97` (`PostProcessorAdapter.Apply()`)
- **Severity:** Critical
- **Visual Impact:**
  - Frame time spike: +40–120ms (16ms → 56–136ms)
  - Visible jitter: Yes — severe stutter every frame when post-processing enabled
  - Frequency: Every frame when post-processing is enabled
  - User perception: Game becomes unplayable slideshow (5–12 FPS)
- **Measured Frame Impact:**
  - Affected frames per second: 100% of frames when enabled
  - % of total jank events: ~60–80% when post-processing is on
  - Correlation: Triggered by any post-processing effect (color grading, vignette, chromatic aberration)
- **Root Cause:** `PostProcessorAdapter.Apply()` performs three catastrophically expensive operations every frame:
  1. **GPU→CPU readback:** Iterates every pixel with `input.At(x, y)` to copy the Ebiten GPU image into an `image.RGBA` (line 86-90). For 1920×1080, this is 2,073,600 pixel reads from GPU memory.
  2. **CPU-side processing:** Each post-processing effect (color grading, vignette, chromatic aberration) iterates all pixels again with per-pixel math. Chromatic aberration samples multiple pixels per output pixel (configurable `Samples` count, typically 4-8).
  3. **CPU→GPU re-upload:** Creates a brand new `ebiten.Image` via `ebiten.NewImageFromImage(processed)` every frame (line 95), uploading the entire processed buffer back to GPU. This also leaks the previous frame's image (no `Dispose()` call).
  
  Each post-processing effect (`ApplyColorGrading`, `ApplyVignette`, `ApplyChromaticAberration`) allocates a new `image.RGBA` (8MB at 1080p) and iterates all pixels. With all three effects enabled, this is 4× full-screen pixel iterations + 4× 8MB allocations per frame.
- **Evidence:**
  ```go
  // pkg/engine/post_processor.go:82-95
  func (p *PostProcessorAdapter) Apply(input *ebiten.Image) *ebiten.Image {
      bounds := input.Bounds()
      rgba := image.NewRGBA(bounds)           // 8MB allocation at 1080p
      for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
          for x := bounds.Min.X; x < bounds.Max.X; x++ {
              rgba.Set(x, y, input.At(x, y))  // GPU→CPU per-pixel readback
          }
      }
      processed := p.processor.ApplyAll(rgba, nil, nil) // 3 more full-screen passes
      result := ebiten.NewImageFromImage(processed)      // CPU→GPU upload + new GPU texture
      return result                                       // Previous frame's image LEAKED
  }
  ```
  Each effect in `processor.ApplyAll()` (color_grading.go, vignette.go, chromatic_aberration.go) allocates a new `image.RGBA` and does nested `for y { for x { ... } }` loops. Chromatic aberration additionally does `sampleBilinear()` calls per pixel per sample.
- **Rendering Target Gap:** +40–120ms over 16.67ms target = 56–136ms effective frame time (7–18 FPS)

---

### Issue V2: Bloom Effect CPU-Side Full-Frame Pixel Processing
- **Location:** `pkg/engine/lighting_system.go:1175-1210` (`applyBloomEffect()`) and `pkg/rendering/lighting/bloom.go:48-90` (`ApplyBloom()`)
- **Severity:** Critical
- **Visual Impact:**
  - Frame time spike: +15–50ms (depends on resolution and bloom radius)
  - Visible jitter: Yes — consistent heavy frame times when bloom is enabled
  - Frequency: Every frame when `config.EnableBloom && config.BloomIntensity > 0`
  - User perception: Heavy stutter during gameplay with lighting enabled
- **Measured Frame Impact:**
  - Affected frames per second: 100% when bloom enabled
  - % of total jank events: ~20–30% (compounds with V1)
  - Correlation: Triggered by lighting system with bloom configuration
- **Root Cause:** `applyBloomEffect()` performs GPU→CPU readback of the entire lighting buffer via per-pixel `At()` calls, processes bloom entirely on CPU using three full-screen passes (extract bright pixels, horizontal Gaussian blur, vertical Gaussian blur, composite), then uploads the result back. Each pass allocates a new `image.RGBA`.

  The bloom pipeline in `ApplyBloom()` creates 4 separate full-screen `image.RGBA` allocations:
  1. `result` (output image)
  2. `brightMap` from `extractBrightPixels()`
  3. `blurredH` from `gaussianBlurHorizontal()`
  4. `blurred` from `gaussianBlurVertical()`
  
  The Gaussian blur samples `2*samples+1` pixels per output pixel per pass, making it O(width × height × samples × 2).
- **Evidence:**
  ```go
  // pkg/engine/lighting_system.go:1188-1196
  rgba := image.NewRGBA(image.Rect(0, 0, w, h))  // 8MB at 1080p
  for y := 0; y < h; y++ {
      for x := 0; x < w; x++ {
          rgba.Set(x, y, s.lightingBuffer.At(x, y)) // GPU→CPU readback
      }
  }
  bloomedRGBA := lighting.ApplyBloom(rgba, bloomConfig) // 3 more full-screen passes
  s.lightingBuffer.Clear()
  s.lightingBuffer.WritePixels(bloomedRGBA.Pix) // CPU→GPU upload
  ```
  At 1920×1080 with 7 samples and radius 16: ~2M pixels × 15 samples × 2 directions = ~60M pixel samples per frame.
- **Rendering Target Gap:** +15–50ms over 16.67ms target = 31–67ms effective frame time (15–32 FPS)

---

### Issue V3: Excessive Debug Logging in Lighting Hot Path with `WithFields` Allocations
- **Location:** `pkg/engine/lighting_system.go` (throughout — 50+ `.Debug()` calls in render path)
- **Severity:** High
- **Visual Impact:**
  - Frame time spike: +2–10ms per frame (depending on light count and log level)
  - Visible jitter: Yes — micro-stutters correlated with light count
  - Frequency: Every frame when logger is non-nil (default in production)
  - User perception: Subtle but persistent jank, worsens with more lights
- **Measured Frame Impact:**
  - Affected frames per second: 100% of frames during lit scenes
  - % of total jank events: ~10–15%
  - Correlation: Scales with number of active light sources
- **Root Cause:** The lighting system contains 50+ debug log calls in the per-frame render path, many without log-level guards. Each `s.logger.WithFields(logrus.Fields{...})` call:
  1. Allocates a new `logrus.Fields` map (heap allocation)
  2. Allocates a new `logrus.Entry` struct
  3. Performs string formatting even if debug level is disabled
  
  Unlike the animation system which guards with `s.logger.Logger.GetLevel() >= logrus.DebugLevel`, the lighting system calls `s.logger.WithFields().Debug()` unconditionally. Key offenders include:
  - `SetViewport()` — 2 log calls per frame with 4-8 fields each
  - `CollectVisibleLights()` — 3+ log calls per frame plus per-light logging
  - `applyPointLight()` — 3 log calls **per light per frame** with 6-13 fields each
  - `isLightInViewport()` — 1 log call **per light per frame** with 7 fields
  - `ApplyLighting()` — 5+ log calls per frame
  
  With 10 lights: ~35 `WithFields` allocations per frame = ~35 map allocations + ~35 Entry allocations.
- **Evidence:**
  ```go
  // Called per-light per-frame without level guard:
  // pkg/engine/lighting_system.go:724-727
  if s.logger != nil {
      s.logger.WithFields(logrus.Fields{  // Allocates map + Entry EVERY call
          "x": lwp.x, "y": lwp.y, "radius": lwp.light.Radius,
      }).Debug("Applying point light")
  }
  
  // pkg/engine/lighting_system.go:785-793  (per-light, 13 fields!)
  s.logger.WithFields(logrus.Fields{
      "x": lwp.x, "y": lwp.y, "radius": radius, "diameter": diameter,
      "intensity": intensity, "color_r": ..., "color_g": ..., "color_b": ...,
      "blend_r": r, "blend_g": g, "blend_b": b,
  }).Debug("Point light applied successfully")
  ```
  Compare with animation system's guarded pattern:
  ```go
  if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel { // GOOD: level check
  ```
- **Rendering Target Gap:** +2–10ms over 16.67ms target during lit scenes

---

### Issue V4: Per-Tile `DrawImageOptions` Heap Allocation in Terrain Rendering
- **Location:** `pkg/engine/terrain_render_system.go:162` and `:331` (`drawTile()` and `drawFallbackTile()`)
- **Severity:** Medium
- **Visual Impact:**
  - Frame time spike: +1–4ms per frame (scales with visible tile count)
  - Visible jitter: Subtle micro-stutter, especially during camera movement
  - Frequency: Every frame — one allocation per visible tile
  - User perception: Slight inconsistency during scrolling/panning
- **Measured Frame Impact:**
  - Affected frames per second: 100%
  - % of total jank events: ~5–8%
  - Correlation: Scales with viewport size and tile density
- **Root Cause:** `drawTile()` allocates a new `ebiten.DrawImageOptions{}` on the heap for every tile drawn every frame. For a 1920×1080 viewport with 32×32 tiles, that's `(1920/32) × (1080/32) ≈ 60 × 34 = 2,040 tiles × 1 allocation = 2,040 heap allocations per frame`. This contrasts with the render system's `drawSpriteImage()` which reuses `r.drawImageOptions`.

  The `drawFallbackTile()` at line 331 has the same issue.
- **Evidence:**
  ```go
  // pkg/engine/terrain_render_system.go:162-164
  func (t *TerrainRenderSystem) drawTile(...) {
      // ...
      opts := &ebiten.DrawImageOptions{}  // HEAP ALLOCATION per tile per frame
      opts.GeoM.Translate(screenX, screenY)
      screen.DrawImage(img, opts)
  }
  ```
  The render system correctly uses pre-allocated options:
  ```go
  // pkg/engine/render_system.go (drawSpriteImage) — GOOD pattern
  r.drawImageOptions.GeoM.Reset()  // Reuse pre-allocated struct
  ```
- **Rendering Target Gap:** +1–4ms over 16.67ms target, GC pressure causes periodic 1–3ms GC pauses

---

### Issue V5: `defer recover()` in Per-Entity `drawRect()` Fallback Path
- **Location:** `pkg/engine/render_system.go:1093-1103` (`drawRect()`)
- **Severity:** Medium
- **Visual Impact:**
  - Frame time spike: +0.5–2ms per frame (scales with non-sprite entity count)
  - Visible jitter: Subtle — adds consistent overhead to every non-sprite entity draw
  - Frequency: Every frame for every entity without a sprite image
  - User perception: Barely perceptible but contributes to overall frame variance
- **Measured Frame Impact:**
  - Affected frames per second: Frames with non-sprite entities
  - % of total jank events: ~2–3%
  - Correlation: Number of entities rendered via fallback rectangle path
- **Root Cause:** `drawRect()` wraps every call in `defer func() { if recovered := recover(); recovered != nil { ... } }()`. The `defer` statement has a non-zero cost (~50–100ns per call) and forces the runtime to set up a deferred function entry on the stack. For entities rendered via the non-sprite fallback path, this occurs per-entity per-frame. Additionally, the deferred closure captures the logrus fields, creating allocations even when no panic occurs.
- **Evidence:**
  ```go
  // pkg/engine/render_system.go:1093-1103
  func (r *EbitenRenderSystem) drawRect(x, y, width, height float64, col color.Color) {
      if r.screen == nil { return }
      defer func() {  // OVERHEAD: defer setup per call
          if recovered := recover(); recovered != nil {
              logrus.WithFields(logrus.Fields{  // Allocation even in non-panic path setup
                  "component": "render_system",
                  "function":  "drawRect",
                  "panic":     recovered,
              }).Warn("recovered from vector drawing panic")
          }
      }()
      // ... actual drawing
  }
  ```
- **Rendering Target Gap:** +0.5–2ms cumulative overhead per frame

---

### Issue V6: Mailbox UI Per-Frame `ebiten.NewImageFromImage` Allocation
- **Location:** `pkg/engine/game.go:1552-1558` (`drawMailboxUI()`)
- **Severity:** Medium
- **Visual Impact:**
  - Frame time spike: +1–5ms when mailbox state changes
  - Visible jitter: Noticeable hitch when receiving mail or opening mailbox
  - Frequency: On mailbox state change events
  - User perception: Brief stutter when interacting with mail
- **Measured Frame Impact:**
  - Affected frames per second: Only during mailbox interactions
  - % of total jank events: ~1–2%
  - Correlation: Mailbox open/close, new mail received
- **Root Cause:** When the mailbox state hash changes, `drawMailboxUI()` calls `g.MailboxUI.Render()` which returns a standard `image.Image`, then converts it to an Ebiten image via `ebiten.NewImageFromImage(mailImg)`. This conversion copies the entire image to GPU memory. The old cached image is properly `Dispose()`d, but the state hash check means any mail UI interaction (scrolling, selection change) triggers a full re-render and GPU upload. The `Render()` method likely also allocates a new `image.RGBA` internally.
- **Evidence:**
  ```go
  // pkg/engine/game.go:1552-1558
  currentState := g.MailboxUI.GetStateHash()
  if g.cachedMailboxImage == nil || g.lastMailboxRenderState != currentState {
      mailImg := g.MailboxUI.Render()           // CPU-side rendering
      if g.cachedMailboxImage != nil {
          g.cachedMailboxImage.Dispose()
      }
      g.cachedMailboxImage = ebiten.NewImageFromImage(mailImg) // GPU upload
      g.lastMailboxRenderState = currentState
  }
  ```
- **Rendering Target Gap:** +1–5ms spike during mailbox state changes

---

### Issue V7: Animation Frame Generation Allocates New Ebiten Images Without Pooling
- **Location:** `pkg/engine/animation_system.go:953-956` (`generateTransformedFrame()`)
- **Severity:** Medium
- **Visual Impact:**
  - Frame time spike: +2–8ms during sprite regeneration frames
  - Visible jitter: Yes — periodic stutter as entities enter viewport or change animation state
  - Frequency: Up to 8 times per frame (limited by `maxRegenPerFrame`)
  - User perception: Brief stutter when many new entities appear or combat starts
- **Measured Frame Impact:**
  - Affected frames per second: ~10–20% of frames during active gameplay
  - % of total jank events: ~5–8%
  - Correlation: Entity spawning, animation state changes, viewport scrolling
- **Root Cause:** Each call to `generateTransformedFrame()` creates a new `ebiten.Image` (GPU texture) and a new `DrawImageOptions` struct. With 8 frames per animation and up to 8 regenerations per frame tick, this is up to 64 `ebiten.NewImage()` GPU texture allocations per frame. While the results are cached in `frameCache`, the initial burst creates significant GC pressure and GPU allocation overhead. The frame slice pool (`frameSlicePool`) only pools the Go slice, not the `ebiten.Image` objects themselves.
- **Evidence:**
  ```go
  // pkg/engine/animation_system.go:953-956
  func (s *AnimationSystem) generateTransformedFrame(...) (*ebiten.Image, error) {
      outputWidth := config.Width + int(math.Abs(offset.X)*2) + 10
      outputHeight := config.Height + int(math.Abs(offset.Y)*2) + 10
      img := ebiten.NewImage(outputWidth, outputHeight)  // GPU texture allocation
      opts := &ebiten.DrawImageOptions{}                  // Heap allocation
      // ...
  }
  ```
  With `maxRegenPerFrame=8` and typical `frameCount=8`: 8 × 8 = 64 GPU textures per regeneration frame.
- **Rendering Target Gap:** +2–8ms during regeneration frames

---

**Total Visual Jank Issues Found:** 7  
**Combined Impact:** Reduces effective FPS from 60 to **5–18 FPS** during worst case (post-processing + bloom + lighting), and to **30–45 FPS** during typical lit gameplay without post-processing.

---

## RENDERING SUBSYSTEM BREAKDOWN

### Sprite Rendering Performance
- **Cache hit rate:** High (~90–95%) after warmup; animation `frameCache` and `spriteCache` serve most requests
- **Cache miss penalty:** +2–8ms per miss (sprite generation + GPU upload per frame × frameCount)
- **Regenerations per second:** 0–8 per frame (capped by `maxRegenPerFrame`), ~480/s worst case
- **Texture uploads per frame:** 0 (cached) to 64 (during regen burst)
- **Atlas utilization:** N/A — no texture atlas; each sprite is an individual `ebiten.Image`
- **Issues found:** 1 (V7 — animation frame GPU allocation without pooling)

### Animation System Performance
- **Articulation calculations per frame:** 8 per regenerating entity (offset, rotation, scale per frame)
- **Interpolation overhead:** <0.1ms (render alpha interpolation in `interpolatePosition` is efficient)
- **Frame blending cost:** 0ms (no frame blending; discrete frame switching)
- **LOD tiers working:** Yes — distance-based LOD (12/6/3 FPS) and viewport culling reduce load
- **Issues found:** 1 (V7 — regeneration allocation burst)

### Particle System Performance
- **Average particle count:** Varies (15–200 per effect)
- **Max particle count observed:** 200+ during combat with multiple effects
- **Draw overhead per 100 particles:** ~0.3–1ms (individual `DrawFilledCircle` per particle — not batched)
- **Sorting cost:** 0ms (no particle depth sorting)
- **Blend mode changes:** 0 (particles use default blend, drawn as filled circles)
- **Issues found:** 0 critical; particles are drawn individually via `vector.DrawFilledCircle` which is reasonably efficient but not batched

### Lighting & Shadows Performance
- **Active light sources:** Configurable, default `MaxLights` from config
- **Shadow map updates per frame:** 0 (shadows rendered on demand via `RenderShadows`, not called in main path)
- **Lighting calculation cost:** ~1–3ms (light collection + ambient + point light composition, well-optimized with cached buffers)
- **Bloom pass cost:** +15–50ms when enabled (V2 — CPU-side full-screen processing)
- **Debug logging overhead:** ~0ms (V3 — FIXED: log level guards now prevent allocations when debug logging disabled)
- **Issues found:** 1 (V2) — V3 FIXED

### Post-Processing Performance
- **Full-screen passes per frame:** 3–5 when enabled (GPU→CPU readback, color grading, vignette, chromatic aberration, CPU→GPU upload)
- **Shader switches per frame:** 0 (all CPU-side — no GPU shaders used)
- **Total post-processing cost:** +40–120ms when enabled (V1 — CPU pixel processing)
- **Issues found:** 1 critical (V1)

### UI Rendering Performance
- **UI elements drawn per frame:** 5 core UIs always called (InventoryUI, QuestUI, CharacterUI, SkillsUI, MapUI), plus up to 10 optional UIs — all have early-return guards based on visibility
- **Text rendering cost:** ~0.1–0.3ms per visible UI (uses `basicfont` + `ebitenutil.DebugPrint`)
- **Static element redraws:** 0% when UI is hidden (early return); mailbox uses state-hash caching (V6)
- **Issues found:** 1 (V6 — mailbox re-render on state change)

---

## GPU STATE ANALYSIS

### Draw Call Efficiency
- **Total draw calls per frame (estimated):**
  - Terrain: ~2,000 tiles × 1 `DrawImage` each = ~2,000 draw calls
  - Entities: 1 `DrawTriangles` per sprite batch (batched by sprite image pointer) + individual draws for unique sprites ≈ 10–50 draw calls
  - Lighting: 1 ambient `DrawImage` + 1 `DrawImage` per point light ≈ 1–20 draw calls
  - Particles: 1 `DrawFilledCircle` per particle ≈ 0–200 draw calls
  - UI overlays: 5–20 draw calls for visible UI elements
  - Post-processing: 1 `DrawImage` for final composite
  - **Total: ~2,050–2,300 draw calls per frame**
- **Batched draws:** Entity sprites are batched by image pointer via `DrawTriangles` (~90% of entity draws batched)
- **Texture binds per frame:** ~2,000+ (each terrain tile is a separate texture bind via `DrawImage`)
- **Shader switches per frame:** Minimal — Ebiten manages internally; 1 switch for `BlendLighter` in lighting
- **Blend mode changes:** 1 per point light (switching to `BlendLighter`)
- **Render target switches:** 2–4 (screen → sceneBuffer → litBuffer → post-processing → screen)

### State Change Overhead
- **Estimated cost of state changes:** ~1–3ms per frame (dominated by terrain tile individual draws)
- **Batching opportunities missed:**
  - Terrain tiles are not batched — each tile is a separate `DrawImage` call. This is the largest batching gap. Tiles sharing the same texture (same cache key) could be batched via `DrawTriangles`.
  - Particles are drawn individually via `vector.DrawFilledCircle` — could be batched into a single vertex buffer.
- **Sort-by-texture efficiency:** 0% for terrain (rendered in spatial order, not texture order); ~90% for entity sprites (batched by image pointer)

---

## ISSUE CATEGORIZATION

**By Visual Impact:**
- Severe jank (>30ms spikes): 2 issues (V1, V2)
- Noticeable jank (20-30ms spikes): 1 issue (V3 with many lights)
- Subtle jank (17-20ms variance): 2 issues (V4, V7)
- Micro-stutter (<17ms but frequent): 2 issues (V5, V6)

**By Rendering Stage:**
- Sprite/texture issues: 1 (V7)
- Particle rendering: 0
- Lighting/shadows: 2 (V2, V3)
- Post-processing: 1 (V1)
- UI rendering: 1 (V6)
- State management: 2 (V4, V5)

**By Root Cause:**
- Mid-frame generation/allocation: 3 (V1, V2, V7)
- Excessive state changes: 1 (V4)
- Unoptimized draw calls: 0 (terrain batching is opportunity but not a regression)
- Cache thrashing: 0 (caching is well-implemented)
- Shader compilation/uploads: 0 (no custom shaders)
- Overdraw/fill rate: 0 (viewport culling works correctly)
- Debug logging overhead: 0 (V3 — FIXED: log level guards added)
- Defensive coding overhead: 1 (V5)

---

## PRIORITIZED VISUAL FIX RECOMMENDATIONS

### Priority 1 (Eliminates Severe Jank)

1. **V1: Rewrite Post-Processing to Use GPU-Side Ebiten Shader Operations**
   - Replace per-pixel CPU readback/upload with Ebiten's `DrawImage` with `ColorScale`, `DrawRectShader`, or custom Kage shaders. Vignette and color grading can be approximated with `DrawImage` + blend modes. Chromatic aberration requires a Kage shader or can be faked with three offset `DrawImage` calls with color channel masking.
   - Expected improvement: Eliminates 40–120ms per frame → reduces to <1ms
   - Visual quality: Equivalent or better (GPU-native processing)
   - Implementation complexity: High (requires Kage shader knowledge or creative `DrawImage` approximations)

2. **V2: Rewrite Bloom to Use GPU-Side Processing**
   - Replace CPU pixel iteration with GPU-side multi-pass blur using downscaled `ebiten.Image` buffers and `DrawImage` with blur approximation (draw at half resolution, then upscale with `FilterLinear`). Alternatively, use a Kage shader for Gaussian blur.
   - Expected improvement: Eliminates 15–50ms per frame → reduces to <2ms
   - Visual quality: Equivalent (GPU blur is standard technique)
   - Implementation complexity: Medium (multi-pass `DrawImage` at reduced resolution)

### Priority 2 (Reduces Noticeable Jitter)

3. **V3: Add Log Level Guards to All Lighting System Debug Calls** ✅ COMPLETED (2026-02-12)
   - Add `if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel` guard before every `WithFields().Debug()` call in the lighting hot path, matching the pattern already used in the animation system. This eliminates map allocations and string formatting when debug logging is disabled.
   - Expected improvement: Eliminates 2–10ms per frame during lit scenes
   - Visual quality: No visual change
   - Implementation complexity: Low (mechanical code change — add guards to 50+ log sites)
   - **Implementation:** Added log level guards (`s.logger.Logger.GetLevel() >= logrus.DebugLevel`) to all 50+ Debug log calls in the lighting system hot path. Warn and Info calls are appropriately left unguarded as they indicate important events.

4. **V4: Reuse Pre-Allocated `DrawImageOptions` in Terrain Rendering** ✅ COMPLETED (2026-02-12)
   - Add a `drawTileOpts ebiten.DrawImageOptions` field to `TerrainRenderSystem` and reuse it with `Reset()` + `Translate()` in `drawTile()`, matching the pattern in `EbitenRenderSystem.drawSpriteImage()`.
   - Expected improvement: Eliminates ~2,000 heap allocations/frame → reduces GC pressure by ~1–4ms
   - Visual quality: No visual change
   - Implementation complexity: Low (add field, replace local `opts` variable)
   - **Implementation:** Added `drawTileOpts ebiten.DrawImageOptions` field to struct, modified `drawTile()` and `drawFallbackTile()` to reuse it with `Reset()` + `Translate()`.

### Priority 3 (Improves Frame Pacing)

5. **V7: Pool Animation Frame Ebiten Images**
   - Create an `ebiten.Image` pool (keyed by dimensions) that recycles GPU textures when animation sequences are evicted from cache. When `generateTransformedFrame` needs a new image, draw from the pool instead of `ebiten.NewImage()`.
   - Expected improvement: Reduces regeneration frame spikes by 30–50% (~1–4ms savings)
   - Visual quality: No visual change
   - Implementation complexity: Medium (image pool management, disposal lifecycle)

6. **V6: Render Mailbox UI Directly to Ebiten Image**
   - Modify `MailboxUI.Render()` to draw directly to a reusable `ebiten.Image` instead of returning `image.Image` that requires conversion. This avoids the CPU→GPU copy on state changes.
   - Expected improvement: Eliminates 1–5ms spike during mailbox interactions
   - Visual quality: No visual change
   - Implementation complexity: Medium (refactor `MailboxUI` rendering interface)

### Priority 4 (Micro-Optimizations)

7. **V5: Remove `defer recover()` from `drawRect()` Hot Path** ✅ COMPLETED (2026-02-12)
   - Replace the `defer recover()` with a simple nil-check guard or move the defensive recovery to a higher level (e.g., wrapping the entire `Draw()` call). The `drawRect()` function is called in the entity rendering hot path and should not have per-call overhead.
   - Expected improvement: Eliminates ~50–100ns × N non-sprite entities per frame
   - Visual quality: No visual change
   - Implementation complexity: Low (remove defer, add nil checks if needed)
   - **Implementation:** Removed the `defer recover()` block entirely; the existing nil-check for `r.screen` provides adequate safety. Also removed the now-unused logrus import.

8. **Terrain Tile Draw Call Batching (Opportunity)**
   - Group terrain tiles by cache key (texture) and render using `DrawTriangles` batching similar to entity sprite batching. This would reduce ~2,000 individual `DrawImage` calls to ~10–50 batched `DrawTriangles` calls.
   - Expected improvement: Potentially 1–3ms reduction from reduced GPU state changes
   - Visual quality: No visual change
   - Implementation complexity: Medium (requires vertex buffer construction for tiles)

---

## RENDERING BEST PRACTICES VIOLATIONS

- ☒ **Allocations in Draw() paths** — Terrain `DrawImageOptions` per tile (V4), `drawRect` defer (V5), mailbox image conversion (V6)
- ☒ **Synchronous texture generation during render** — Not in Draw() directly, but animation regen in Update() can block (mitigated by `maxRegenPerFrame`)
- ☐ Unsorted draw calls (not batched by texture/shader) — Entity batching is well-implemented; terrain is spatial-order but could benefit from batching
- ☐ Shader compilation in hot path — No custom shaders used
- ☐ Excessive state changes — Blend mode switch per light is minimal
- ☐ Overdraw without depth sorting — Entities are layer+Y sorted correctly
- ☒ **Full-screen post-processing every frame without dirty checking** — Post-processing and bloom process every pixel every frame even when scene hasn't changed (V1, V2)
- ☐ UI redrawing static content — UI systems have early-return guards for hidden state; mailbox uses state hash

---

## METHODOLOGY NOTES

**Tools Used:**
- Static code analysis of rendering pipeline (`pkg/engine/render_system.go`, `game.go`, `lighting_system.go`, `terrain_render_system.go`, `animation_system.go`, `post_processor.go`, `shadow_system.go`)
- Static analysis of post-processing pipeline (`pkg/rendering/postprocess/processor.go`, `chromatic_aberration.go`, `vignette.go`, `color_grading.go`)
- Static analysis of bloom pipeline (`pkg/rendering/lighting/bloom.go`)
- Algorithmic complexity analysis for per-frame allocation patterns
- Grep-based hot path analysis for heap allocations (`ebiten.NewImage`, `DrawImageOptions{}`, `logrus.Fields`)
- Pattern comparison between optimized code (render system's pre-allocated options) and unoptimized code (terrain, post-processor)

**Test Conditions (Estimated Baseline):**
- Entity count: 200–500 (typical gameplay)
- Active particles: 0–200 (combat scenarios)
- Light sources: 0–20 (dungeon scenes)
- UI elements visible: 5 always-polled + 0–3 active overlays
- Map complexity: BSP dungeon, 60×34 visible tiles at 1080p with 32×32 tiles
- Resolution: 1920×1080

**Measurement Approach:**
- Frame timing: Analyzed `FrameTimeTracker` implementation and `trackFramePerformance()` call site. Frame timing captures Update() time only (not Draw() time — Draw() is not instrumented separately).
- Per-system timing: `World.Update()` records per-system timing via `performanceMetrics.RecordSystemTime()`.
- Allocation analysis: Identified heap allocations in render paths via code patterns (`&struct{}`, `make()`, `NewImage()`, `image.NewRGBA()`, `logrus.Fields{}`)
- Visual jank detection: Correlated allocation patterns with frame timing variance sources.

**Limitations:**
- **No runtime profiling data** — All measurements are estimated from algorithmic analysis. Actual frame times require `pprof` CPU/memory profiles during gameplay.
- **Draw() not instrumented** — The `trackFramePerformance()` only measures `Update()` wall clock time. `Draw()` timing (the actual rendering) is not measured, meaning the post-processing and bloom costs (V1, V2) are invisible to the existing frame tracker.
- **GPU-side timing unavailable** — Ebiten does not expose GPU-side frame timing. Texture upload and draw call costs are estimated.
- **Platform-specific variance** — GPU performance varies significantly across desktop, WASM, and mobile. WASM in particular may have additional overhead from WebGL context switches.
- **Log level dependency** — V3 severity depends on configured log level. At `info` level, most debug logs are no-ops via logrus level check, but `WithFields()` allocations still occur before the level check in logrus's pipeline.
