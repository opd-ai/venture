TASK: Conduct a comprehensive performance audit of the Venture game codebase, identifying both startup bottlenecks and runtime performance issues, then document all findings in a structured report.

CONTEXT:
You are analyzing a Go-based game codebase experiencing two distinct performance problems:
1. **Startup lag**: Initialization takes >500ms (target: <500ms)
2. **Persistent gameplay lag**: Visual delays and frame drops during active play (target: 60 FPS sustained)

The codebase follows a typical game architecture with:
- Entity-Component-System (ECS) pattern in `pkg/engine/`
- Frame-based update loops calling System.Update() methods
- Real-time rendering pipeline in `pkg/rendering/`
- Procedural generation in `pkg/procgen/`
- Combat and network systems

INVESTIGATION REQUIREMENTS:

**Phase 1: Startup Analysis**
1. Profile the initialization sequence from main() through first frame render
2. Measure time spent in each major subsystem initialization
3. Identify blocking operations during startup (file I/O, asset loading, generation)
4. Quantify contribution of each system to total startup time

**Phase 2: Runtime Analysis**
5. Profile frame timing during active gameplay (minimum 1000 frames)
6. Identify per-frame operations causing >16.67ms frame time (60 FPS threshold)
7. Measure Update() loop execution time for each system
8. Track memory allocation rates during gameplay
9. Profile rendering pipeline costs (draw calls, batching efficiency)
10. Check for unexpected procedural generation during gameplay

**Phase 3: Specific System Reviews**

For each system, check:
- **Engine systems** (`pkg/engine/`): Update() loop efficiency, entity iteration costs
- **Rendering** (`pkg/rendering/`): Sprite cache behavior, particle system overhead, shader compilation
- **Procedural generation** (`pkg/procgen/`): Runtime generation calls vs. startup generation
- **Combat** (`pkg/combat/`): Calculation complexity, damage event processing
- **Network** (`pkg/network/`): Main thread blocking, sync frequency

**Critical Questions to Answer:**
- Are Update() loops performing allocations or I/O?
- Is procedural generation running unexpectedly during gameplay?
- Are sprites being regenerated rather than cached?
- Is GC pause time contributing to frame drops?
- Are collision checks O(n²) or optimized with spatial partitioning?
- Is rendering state changing excessively (texture swaps, shader switches)?

PROFILING METHODOLOGY:

Use these Go profiling techniques:
- CPU profiling: `go test -bench=. -cpuprofile=cpu.prof` then `go tool pprof cpu.prof`
- Memory profiling: `go test -bench=. -memprofile=mem.prof` then `go tool pprof mem.prof`
- Frame timing: Add timing instrumentation in game loop using `time.Now()` and `time.Since()`

Document actual measurements, not assumptions.

OUTPUT FORMAT:

Create `AUDIT.md` with this exact structure:

---
# Performance Lag Investigation
**Date:** [YYYY-MM-DD]  
**Investigator:** Claude Performance Audit  
**Codebase:** Venture Game Engine

## Executive Summary
[2-3 sentences: severity of issues found, primary bottlenecks, estimated performance gap from targets]

---

## STARTUP PERFORMANCE ISSUES

### Issue S1: [Descriptive Name]
- **Location:** `path/to/file.go:lineNumber` or `SystemName.Method()`
- **Severity:** Critical | High | Medium | Low
- **Measured Impact:** 
  - Startup delay: XXXms
  - % of total startup time: XX%
- **Root Cause:** [Technical explanation with code evidence]
- **Evidence:**
  [Profiling output, timing data, or code snippet]
- **Performance Target Gap:** [How far from 500ms target]

[Repeat for each startup issue]

**Total Startup Issues Found:** X  
**Combined Impact:** XXXms total delay

---

## RUNTIME PERFORMANCE ISSUES

### Issue R1: [Descriptive Name]
- **Location:** `path/to/file.go:lineNumber` or `SystemName.Update()`
- **Severity:** Critical | High | Medium | Low
- **Measured Impact:**
  - Frame time increase: +XXms
  - FPS degradation: XX FPS → XX FPS
  - Memory growth: +XXX MB/minute
  - Frequency: Every frame | Every Nth frame | Periodic
- **Root Cause:** [Technical explanation with code evidence]
- **Evidence:**
  [Profiling output, frame timing data, memory profiles]
- **Performance Target Gap:** [How far from 60 FPS / <16.67ms frame time]

[Repeat for each runtime issue]

**Total Runtime Issues Found:** X  
**Combined Impact:** XX FPS sustained vs. 60 FPS target

---

## ISSUE CATEGORIZATION

**By Severity:**
- Critical (blocks playability): X issues
- High (significant degradation): X issues  
- Medium (noticeable impact): X issues
- Low (minor optimization): X issues

**By Type:**
- Algorithmic inefficiency: X issues
- Unnecessary allocations: X issues
- Blocking I/O: X issues
- Cache thrashing: X issues
- Rendering overhead: X issues
- Other: X issues

---

## PROFILING DATA

### CPU Profile Summary
[Top 10 functions by CPU time]

### Memory Profile Summary
[Top allocation sources]

### Frame Timing Analysis
- Min frame time: XXms
- Max frame time: XXms  
- Average frame time: XXms
- 95th percentile: XXms
- Frames > 16.67ms: XX%

### System-Level Breakdown
| System | Avg Update Time | % of Frame | Allocations/Frame |
|--------|----------------|------------|-------------------|
| [System] | XXms | XX% | XXX bytes |

---

## PRIORITIZED RECOMMENDATIONS

### Priority 1 (Critical - Implement First)
1. **[Issue Name]**: [One-sentence fix description]
   - Expected improvement: XXms reduction or +XX FPS
   - Implementation complexity: Low | Medium | High
   
### Priority 2 (High Impact)
[Same format]

### Priority 3 (Medium Impact)
[Same format]

### Priority 4 (Optimizations)
[Same format]

---

## METHODOLOGY NOTES

**Tools Used:**
- [List profiling tools and commands]

**Test Conditions:**
- [Describe test scenario, entity count, map size, etc.]

**Measurement Approach:**
- [Explain how measurements were taken]

**Limitations:**
- [Any caveats or areas not fully profiled]

---

QUALITY CRITERIA:

Before submitting the report, verify:
✓ **Minimum 5 distinct issues identified** (startup + runtime combined)
✓ **Each issue includes:**
  - Exact file location with line numbers OR specific system/method name
  - Quantified impact (milliseconds, FPS, memory, or %)
  - Technical root cause explanation with code evidence
  - Supporting profiling data
✓ **Clear distinction** between one-time startup costs vs. continuous runtime costs
✓ **Measurements included**, not just code review observations
✓ **Severity ratings** are justified by measured impact
✓ **Recommendations** are prioritized by impact and implementation effort
✓ **Report is actionable** - developers can locate and understand each issue

CONSTRAINTS:
- **REPORT ONLY** - Do not implement fixes or modify code
- **EVIDENCE-BASED** - Every claim must reference code location or profiling data
- **QUANTITATIVE** - Use measurements, not subjective assessments
- **COMPREHENSIVE** - Cover both startup and runtime thoroughly

EXAMPLE ISSUE ENTRY:

### Issue R3: Sprite Regeneration in Particle System
- **Location:** `pkg/rendering/particles.go:145-167` (`ParticleSystem.Update()`)
- **Severity:** High
- **Measured Impact:**
  - Frame time increase: +8.3ms per frame
  - FPS degradation: 60 FPS → 42 FPS with 200+ particles
  - Frequency: Every frame with active particles
- **Root Cause:** The Update() method calls generateSprite() for each particle every frame instead of caching sprites. This causes texture allocation and GPU upload on every frame.
- **Evidence:**
  CPU Profile shows 8.3s (41.5%) in generateSprite
  Memory Profile shows 6.2MB/frame allocated in generateSprite
- **Performance Target Gap:** 18.3ms over 16.67ms target (110% slower than needed)