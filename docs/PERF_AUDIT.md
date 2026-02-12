TASK: Conduct a focused visual performance audit of the Venture game codebase, specifically targeting visual jank, jitter, frame pacing issues, and rendering lag that impact gameplay smoothness.

CONTEXT:
You are analyzing a Go-based game codebase experiencing visual performance problems:
- **Visual jank/jitter**: Inconsistent frame timing causing stuttering despite acceptable average FPS
- **Rendering lag**: Visible delays between input and visual feedback
- **Frame pacing issues**: Uneven frame delivery (e.g., 16ms, 16ms, 35ms pattern instead of consistent 16.67ms)

The rendering pipeline includes:
- Entity-Component-System (ECS) pattern in `pkg/engine/`
- Frame-based rendering loop calling System.Update() and Draw() methods
- Real-time rendering pipeline in `pkg/rendering/` (sprites, particles, lighting, post-processing)
- Procedural sprite/texture generation in `pkg/rendering/sprites/`
- Animation system in `pkg/rendering/animation/`

TARGET: Achieve consistent 60 FPS with <2ms frame time variance (14.67ms - 18.67ms range).

INVESTIGATION REQUIREMENTS:

**Phase 1: Frame Timing Analysis**
1. Profile frame-to-frame timing variance over 5000+ frames during active gameplay
2. Identify frame time spikes (any frame >20ms)
3. Measure 1%, 5%, and 95% percentile frame times
4. Detect frame pacing patterns (consistent vs. irregular delivery)
5. Correlate frame spikes with specific game events (combat, particle spawns, UI updates)

**Phase 2: Rendering Pipeline Analysis**
6. Profile Draw() call timing for each rendering system
7. Measure GPU state changes (texture binds, shader switches, blend mode changes)
8. Identify draw call batching efficiency
9. Check for synchronous sprite/texture generation during render frames
10. Profile post-processing pass overhead
11. Measure particle system draw overhead (particle count vs. frame time correlation)
12. Check lighting and shadow system render costs

**Phase 3: Visual Subsystem Reviews**

For each rendering subsystem, investigate:
- **Sprite rendering** (`pkg/rendering/sprites/`): Cache hits vs. regeneration, texture atlas usage, GPU upload frequency
- **Animation** (`pkg/rendering/animation/`): Frame interpolation smoothness, articulation calculations per frame
- **Particles** (`pkg/rendering/particles/`): Particle count limits, sorting overhead, alpha blending cost
- **Lighting** (`pkg/rendering/lighting/`): Light source count, shadow map updates, bloom calculation cost
- **Post-processing** (`pkg/rendering/postprocess/`): Full-screen pass count, shader complexity
- **UI rendering** (`pkg/rendering/ui/`): Text rendering cost, UI element redraw frequency, layering overhead

**Critical Visual Performance Questions:**
- Are sprites being regenerated mid-frame instead of cached?
- Is texture memory exhausted, causing thrashing?
- Are draw calls sorted to minimize state changes?
- Is vsync causing input-to-render latency?
- Are particles using expensive blend modes or excessive overdraw?
- Is the lighting system recalculating every frame vs. dirty marking?
- Are post-processing shaders compiling or uploading uniforms every frame?
- Is UI redrawing static elements unnecessarily?
- Are animations interpolating smoothly or snapping between frames?
- Is the render thread blocking on Update() logic?

**Visual Jank Patterns to Detect:**
- **Allocation jank**: GC pauses during render (check heap allocations in Draw() paths)
- **Texture jank**: Sprite cache misses forcing mid-frame generation
- **Shader jank**: Shader compilation or uniform uploads in hot path
- **Particle jank**: Particle system spawning/sorting causing frame spikes
- **UI jank**: Complex UI redraws blocking render thread
- **State jank**: Excessive GPU state changes breaking batching

PROFILING METHODOLOGY:

Use these Go and graphics profiling techniques:
- Frame timing: Instrument render loop with high-precision timing (`time.Now()` before/after each Draw() call)
- CPU profiling: `go test -bench=BenchmarkRender -cpuprofile=cpu.prof` focusing on render paths
- Memory profiling: `go test -bench=BenchmarkRender -memprofile=mem.prof` checking Draw() allocations
- GPU profiling: Log draw call count, texture binds, shader switches per frame
- Visual profiling: Add on-screen frame time graph and frame time histogram

Document actual measurements with frame-by-frame data, not assumptions.

OUTPUT FORMAT:

Create `AUDIT.md` with this exact structure:

---
# Visual Performance Audit
**Date:** [YYYY-MM-DD]  
**Investigator:** Claude Visual Performance Audit  
**Codebase:** Venture Game Engine - Rendering Pipeline

## Executive Summary
[2-3 sentences: severity of visual jank, primary rendering bottlenecks, measured frame time variance vs. target]

---

## FRAME TIMING ANALYSIS

### Overall Frame Metrics (5000 frame sample)
- **Min frame time:** XXms
- **Average frame time:** XXms
- **Max frame time:** XXms
- **1st percentile:** XXms (best 1% of frames)
- **5th percentile:** XXms
- **95th percentile:** XXms
- **99th percentile:** XXms (worst 1% of frames)
- **Frames >20ms:** XX% (jank threshold)
- **Frames >30ms:** XX% (severe jank)
- **Frame time variance (stddev):** XXms (target: <2ms)

### Frame Pacing Pattern
[Describe timing pattern: consistent, periodic spikes, random variance, etc.]
[Include visualization if possible: frame time graph or histogram]

---

## VISUAL JANK ISSUES

### Issue V1: [Descriptive Name - e.g., "Sprite Cache Miss Causing Mid-Frame Generation"]
- **Location:** `path/to/file.go:lineNumber` or `SystemName.Draw()`
- **Severity:** Critical | High | Medium | Low
- **Visual Impact:**
  - Frame time spike: +XXms (baseline XXms → XXms)
  - Visible jitter: Yes/No
  - Frequency: Every Nth frame | On event | Random
  - User perception: Severe stutter | Noticeable hitch | Subtle jank
- **Measured Frame Impact:**
  - Affected frames per second: XX
  - % of total jank events: XX%
  - Correlation: [What triggers this? Combat? Particle spawns? UI actions?]
- **Root Cause:** [Technical explanation with code evidence focusing on rendering/drawing]
- **Evidence:**
  ```
  Frame 1247: 16.2ms (normal)
  Frame 1248: 42.8ms (+26.6ms spike) ← Sprite generation in ParticleSystem.Draw()
  Frame 1249: 17.1ms (recovered)
  
  CPU Profile: 26.3ms in generateParticleSprite() called from Draw() path
  ```
- **Rendering Target Gap:** XXms over 16.67ms target, causing XX FPS dip

[Repeat for each visual jank issue - aim for 5-10 distinct rendering issues]

**Total Visual Jank Issues Found:** X  
**Combined Impact:** Reduces effective FPS from 60 to XX during jank events

---

## RENDERING SUBSYSTEM BREAKDOWN

### Sprite Rendering Performance
- **Cache hit rate:** XX%
- **Cache miss penalty:** +XXms per miss
- **Regenerations per second:** XX
- **Texture uploads per frame:** XX
- **Atlas utilization:** XX%
- **Issues found:** X

### Animation System Performance
- **Articulation calculations per frame:** XXX
- **Interpolation overhead:** XXms
- **Frame blending cost:** XXms
- **Issues found:** X

### Particle System Performance
- **Average particle count:** XXX
- **Max particle count observed:** XXX
- **Draw overhead per 100 particles:** XXms
- **Sorting cost:** XXms
- **Blend mode changes:** XX per frame
- **Issues found:** X

### Lighting & Shadows Performance
- **Active light sources:** XX
- **Shadow map updates per frame:** XX
- **Lighting calculation cost:** XXms
- **Bloom pass cost:** XXms
- **Issues found:** X

### Post-Processing Performance
- **Full-screen passes per frame:** XX
- **Shader switches per frame:** XX
- **Total post-processing cost:** XXms
- **Issues found:** X

### UI Rendering Performance
- **UI elements drawn per frame:** XXX
- **Text rendering cost:** XXms
- **Static element redraws:** XX% (should be 0% when cached)
- **Issues found:** X

---

## GPU STATE ANALYSIS

### Draw Call Efficiency
- **Total draw calls per frame:** XXX
- **Batched draws:** XX%
- **Texture binds per frame:** XXX (target: minimize)
- **Shader switches per frame:** XX (target: <10)
- **Blend mode changes:** XX
- **Render target switches:** XX

### State Change Overhead
- **Estimated cost of state changes:** XXms per frame
- **Batching opportunities missed:** XX%
- **Sort-by-texture efficiency:** XX%

---

## ISSUE CATEGORIZATION

**By Visual Impact:**
- Severe jank (>30ms spikes): X issues
- Noticeable jank (20-30ms spikes): X issues
- Subtle jank (17-20ms variance): X issues
- Micro-stutter (<17ms but frequent): X issues

**By Rendering Stage:**
- Sprite/texture issues: X
- Particle rendering: X
- Lighting/shadows: X
- Post-processing: X
- UI rendering: X
- State management: X

**By Root Cause:**
- Mid-frame generation/allocation: X
- Excessive state changes: X
- Unoptimized draw calls: X
- Cache thrashing: X
- Shader compilation/uploads: X
- Overdraw/fill rate: X

---

## PRIORITIZED VISUAL FIX RECOMMENDATIONS

### Priority 1 (Eliminates Severe Jank)
1. **[Issue Name]**: [One-sentence fix description]
   - Expected improvement: Eliminates XX% of >30ms frame spikes
   - Visual quality: [Smoothness improvement description]
   - Implementation complexity: Low | Medium | High
   
### Priority 2 (Reduces Noticeable Jitter)
[Same format]

### Priority 3 (Improves Frame Pacing)
[Same format]

### Priority 4 (Micro-Optimizations)
[Same format]

---

## RENDERING BEST PRACTICES VIOLATIONS

[List any violations of graphics programming best practices found:]
- ☐ Allocations in Draw() paths
- ☐ Synchronous texture generation during render
- ☐ Unsorted draw calls (not batched by texture/shader)
- ☐ Shader compilation in hot path
- ☐ Excessive state changes
- ☐ Overdraw without depth sorting
- ☐ Full-screen post-processing every frame without dirty checking
- ☐ UI redrawing static content

---

## METHODOLOGY NOTES

**Tools Used:**
- [List profiling tools and commands used for visual analysis]

**Test Conditions:**
- Entity count: XXX
- Active particles: XXX
- Light sources: XX
- UI elements visible: XX
- Map complexity: [Description]

**Measurement Approach:**
- [Explain frame timing instrumentation]
- [Describe visual jank detection methodology]

**Limitations:**
- [Any rendering aspects not profiled]
- [Platform-specific caveats]

---

QUALITY CRITERIA:

Before submitting the report, verify:
✓ **Minimum 5 distinct visual/rendering issues identified**
✓ **Each issue includes:**
  - Exact file location OR specific rendering system/method
  - Quantified frame time impact with spike measurements
  - Visual perception description (severe stutter vs. subtle jank)
  - Frequency and trigger correlation
  - Code evidence from rendering pipeline
✓ **Frame timing data** with percentiles and variance measurements
✓ **Rendering subsystem breakdown** with per-system metrics
✓ **GPU state analysis** covering draw calls and state changes
✓ **Clear focus on visual smoothness**, not general performance
✓ **Recommendations prioritized** by visual impact reduction
✓ **Report is actionable** for graphics optimization

CONSTRAINTS:
- **FOCUS ON VISUALS** - This audit covers rendering/drawing performance only, not gameplay logic
- **REPORT ONLY** - Do not implement fixes or modify code
- **EVIDENCE-BASED** - Every claim must reference rendering code or frame timing data
- **QUANTITATIVE** - Use frame time measurements and draw call counts
- **JANK-FOCUSED** - Prioritize frame time variance and spikes over average FPS

EXAMPLE ISSUE ENTRY:

### Issue V2: Particle Sprite Regeneration Every Frame
- **Location:** `pkg/rendering/particles/particle_system.go:289-312` (`ParticleSystem.Draw()`)
- **Severity:** Critical
- **Visual Impact:**
  - Frame time spike: +12.4ms (16.8ms → 29.2ms)
  - Visible jitter: Yes - severe stutter during particle-heavy combat
  - Frequency: Every frame with >50 active particles
  - User perception: Game appears to "freeze" for 1 frame every 3-4 frames
- **Measured Frame Impact:**
  - Affected frames per second: ~20 FPS (33% of frames)
  - % of total jank events: 48%
  - Correlation: Triggered by combat effects, spell casts, explosions
- **Root Cause:** Draw() method calls `generateParticleSprite()` for each particle instead of using cached sprites. This causes texture allocation, procedural generation, and GPU upload every frame.
- **Evidence:**
  ```
  Frame timing with 200 particles:
  Frame 1001: 16.9ms
  Frame 1002: 31.2ms (+14.3ms spike) ← 200 calls to generateParticleSprite()
  Frame 1003: 17.3ms
  Frame 1004: 29.8ms (+12.5ms spike)
  
  CPU Profile: generateParticleSprite() = 14.1s total (47.2% of CPU time)
  Memory Profile: 18.4MB allocated per frame in sprite generation
  Draw call count: 200 individual draws (not batched)
  ```
- **Rendering Target Gap:** 12.5ms average over 16.67ms target during particle events = 34.5ms effective frame time (29 FPS)