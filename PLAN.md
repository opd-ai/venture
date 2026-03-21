# Implementation Plan: CI/CD Quality Gates & Code Maintainability

## Project Context
- **What it does**: A fully procedural multiplayer action-RPG where all graphics, audio, and gameplay content are generated at runtime from a single binary with no external assets.
- **Current goal**: Add automated CI gates for FPS and memory benchmarks to prevent performance regressions
- **Estimated Scope**: Medium

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|----------------|---------------------|
| 60 FPS minimum performance | ✅ Achieved (~7,800 FPS measured) | Yes - CI automation |
| <500MB client memory | ✅ Achieved (~2.5MB for 2000 entities) | Yes - CI automation |
| ≥40% test coverage per package | ✅ Achieved (all packages pass) | No |
| High-latency multiplayer (200-5000ms) | ✅ Achieved | No |
| 100% procedural content | ✅ Achieved | No |
| Single binary distribution | ✅ Achieved | No |
| Cross-server federation | ✅ Achieved | No |
| Genre-based theming | ✅ Achieved | No |
| VR/stereoscopic support | ⚠️ Experimental (mock adapters) | No |
| Code maintainability (complexity) | ⚠️ 27 functions ≥15 complexity | Yes - refactoring |
| Code deduplication | ⚠️ 8.0% duplication ratio | Yes - consolidation |

## Metrics Summary

| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Total Lines of Code | 214,752 | - | - |
| Total Functions/Methods | 16,755 | - | - |
| Total Packages | 106 | - | - |
| Functions ≥15 cyclomatic complexity | 27 | <15 recommended | ⚠️ Medium |
| Functions ≥10 cyclomatic complexity | 152 | - | Informational |
| Duplication ratio | 8.0% | <10% | ✅ Acceptable |
| Duplicated lines | 32,085 | - | - |
| Documentation coverage (overall) | 90.0% | ≥80% | ✅ Good |
| Exported functions undocumented | 144 | 0 ideal | ⚠️ Low priority |

### High-Complexity Hotspots (≥20 cyclomatic complexity)

| File | Function | Complexity |
|------|----------|------------|
| `cmd/client/handlers.go:2429` | `initializeTerrainCollision` | 37 |
| `pkg/rendering/sprites/generator.go:229` | `generateEntityWithTemplate` | 29 |
| `pkg/rendering/sprites/equipment_renderer.go:580` | `applyMaterialTexture` | 24 |
| `pkg/engine/build_template_system.go:295` | `resetAndApplySkills` | 23 |
| `pkg/engine/behavior_tree_advanced_nodes.go:1167` | `Tick` | 23 |
| `pkg/engine/entity_target_lock_indicator_system.go:155` | `acquireTargets` | 22 |
| `pkg/engine/build_template_system.go:435` | `SaveCurrentBuild` | 21 |
| `pkg/engine/behavior_tree_advanced_nodes.go:556` | `Tick` | 20 |
| `pkg/engine/skill_mutation_system.go:106` | `GenerateMutation` | 20 |

### Largest Duplication Clusters

| Lines | Location | Suggestion |
|-------|----------|------------|
| 79 | `pkg/engine/system_init.go` (2 locations) | Extract system initialization into shared helper |
| 77 | `pkg/engine/system_init.go` (2 locations) | Extract system initialization into shared helper |
| 73 | `cmd/client/handlers.go` (2 locations) | Extract handler pattern into shared helper |
| 65 | `cmd/client/handlers.go` (4 locations) | Extract repeated handler logic |

---

## Implementation Steps

### Step 1: Add FPS Benchmark CI Gate

- **Deliverable**: CI workflow step that runs FPS benchmark and fails on regression ≥20%
- **Dependencies**: None (baseline infrastructure exists)
- **Goal Impact**: Prevents FPS performance regressions from shipping undetected
- **Files**:
  - `.github/workflows/test.yml` — Add benchmark step after tests
  - `scripts/benchmark-regression.sh` — Verify outputs threshold comparison
  - `scripts/benchmark-baseline.json` — Store baseline metrics (create if missing)

**Acceptance Criteria**:
1. CI runs `xvfb-run go test -bench=BenchmarkFPS2000Entities ./pkg/benchmark/fps/`
2. Benchmark results are compared against baseline
3. CI fails if ns/op increases by ≥20% (i.e., FPS drops ≥20%)
4. Baseline can be updated via manual workflow dispatch

**Validation**:
```bash
# Run benchmark and verify output format
xvfb-run go test -bench=BenchmarkFPS2000Entities -benchtime=5s ./pkg/benchmark/fps/ | grep -E "ns/op"

# Verify baseline comparison works
./scripts/benchmark-regression.sh --check
```

---

### Step 2: Add Memory Budget CI Gate

- **Deliverable**: CI workflow step that validates memory usage stays under 500MB
- **Dependencies**: None
- **Goal Impact**: Prevents memory bloat from shipping undetected
- **Files**:
  - `.github/workflows/test.yml` — Add memory check step
  - `scripts/benchmark-memory.sh` — Verify script is headless-compatible

**Acceptance Criteria**:
1. CI runs memory benchmark in headless mode
2. Fails if heap allocation exceeds 500MB budget
3. Reports actual allocation in failure message

**Validation**:
```bash
# Run memory benchmark
xvfb-run go test -bench=BenchmarkMemory -benchmem ./pkg/benchmark/memory/ | grep -E "allocs/op|B/op"

# Verify against budget
./scripts/benchmark-memory.sh --headless --max-heap 500MB
```

---

### Step 3: Refactor `initializeTerrainCollision` (Complexity 37 → <15)

- **Deliverable**: Split into multiple focused functions with single responsibilities
- **Dependencies**: None
- **Goal Impact**: Reduces defect probability on critical rendering path; improves maintainability
- **Files**:
  - `cmd/client/handlers.go:2429` — Refactor function
  - `cmd/client/terrain_collision.go` — New file for extracted helpers (optional)

**Acceptance Criteria**:
1. Original functionality preserved (all existing tests pass)
2. Function complexity reduced to <15
3. Each extracted helper has a clear, documented purpose
4. No new duplication introduced

**Validation**:
```bash
# Verify complexity reduction
go-stats-generator analyze ./cmd/client --skip-tests --format json --sections functions | \
  jq '[.functions[] | select(.name == "initializeTerrainCollision")] | .[0].complexity.cyclomatic'
# Expected: < 15

# Verify tests still pass
xvfb-run go test -v ./cmd/client/...
```

---

### Step 4: Refactor `generateEntityWithTemplate` (Complexity 29 → <15)

- **Deliverable**: Extract template application logic into helper functions
- **Dependencies**: None
- **Goal Impact**: Sprite generation is a critical path; reducing complexity improves reliability
- **Files**:
  - `pkg/rendering/sprites/generator.go:229` — Refactor function
  - `pkg/rendering/sprites/template_helpers.go` — New file for extracted logic (optional)

**Acceptance Criteria**:
1. All sprite generation tests pass
2. Function complexity reduced to <15
3. Template application logic is now unit-testable in isolation

**Validation**:
```bash
# Verify complexity reduction
go-stats-generator analyze ./pkg/rendering/sprites --skip-tests --format json --sections functions | \
  jq '[.functions[] | select(.name == "generateEntityWithTemplate")] | .[0].complexity.cyclomatic'
# Expected: < 15

# Verify tests still pass
xvfb-run go test -v ./pkg/rendering/sprites/...
```

---

### Step 5: Consolidate `system_init.go` Duplication (79+77+73+72+69 lines)

- **Deliverable**: Extract repeated system initialization patterns into helper functions
- **Dependencies**: None
- **Goal Impact**: Reduces maintenance burden; easier to add new systems consistently
- **Files**:
  - `pkg/engine/system_init.go` — Consolidate duplicate blocks
  - `pkg/engine/system_init_helpers.go` — New file for shared initialization logic

**Acceptance Criteria**:
1. Duplication reduced by ≥50% in `system_init.go`
2. All system initialization tests pass
3. System registration pattern is documented

**Validation**:
```bash
# Count duplication before and after
go-stats-generator analyze ./pkg/engine --skip-tests --format json --sections duplication | \
  jq '[.duplication.clones[] | select(.instances[].file | contains("system_init.go"))] | length'
# Expected: Significant reduction

# Verify all systems still initialize
xvfb-run go test -v ./pkg/engine/... -run TestInitialize
```

---

### Step 6: Consolidate `handlers.go` Duplication (73+73+67+65+63 lines)

- **Deliverable**: Extract repeated handler patterns into shared helper functions
- **Dependencies**: None
- **Goal Impact**: Reduces client code complexity; easier handler maintenance
- **Files**:
  - `cmd/client/handlers.go` — Consolidate duplicate blocks
  - `cmd/client/handler_helpers.go` — New file for shared handler logic

**Acceptance Criteria**:
1. Duplication reduced by ≥40% in `handlers.go`
2. All client tests pass
3. Handler pattern is documented for future handlers

**Validation**:
```bash
# Count duplication before and after
go-stats-generator analyze ./cmd/client --skip-tests --format json --sections duplication | \
  jq '[.duplication.clones[] | select(.instances[].file | contains("handlers.go"))] | length'
# Expected: Significant reduction

# Verify client still works
xvfb-run go test -v ./cmd/client/...
```

---

### Step 7: Refactor `applyMaterialTexture` (Complexity 24 → <15)

- **Deliverable**: Split material application logic by material type
- **Dependencies**: Step 4 (both in sprites package, avoid conflicts)
- **Goal Impact**: Equipment rendering is performance-critical; simplifies debugging
- **Files**:
  - `pkg/rendering/sprites/equipment_renderer.go:580` — Refactor function

**Acceptance Criteria**:
1. All equipment rendering tests pass
2. Function complexity reduced to <15
3. Material-specific logic is isolated and testable

**Validation**:
```bash
# Verify complexity reduction
go-stats-generator analyze ./pkg/rendering/sprites --skip-tests --format json --sections functions | \
  jq '[.functions[] | select(.name == "applyMaterialTexture")] | .[0].complexity.cyclomatic'
# Expected: < 15

# Verify tests
xvfb-run go test -v ./pkg/rendering/sprites/... -run TestEquipment
```

---

### Step 8: Create Performance Tuning Documentation

- **Deliverable**: User-facing guide for optimizing game performance on various hardware
- **Dependencies**: Steps 1-2 (ensures documented metrics are accurate)
- **Goal Impact**: Improves user experience; reduces support burden
- **Files**:
  - `docs/PERFORMANCE_TUNING.md` — New comprehensive guide
  - `README.md` — Add link to new guide
  - `docs/FAQ.md` — Add performance FAQ entries (if exists)

**Acceptance Criteria**:
1. Documents minimum hardware specifications
2. Lists all performance-related CLI flags with examples
3. Explains how to interpret profiling output
4. Covers common bottlenecks and solutions
5. Includes hardware-specific optimization recommendations

**Validation**:
```bash
# Verify links are valid
grep -l "PERFORMANCE_TUNING" README.md docs/*.md

# Verify all documented flags exist
grep -oE -- '-[a-z-]+' docs/PERFORMANCE_TUNING.md | sort -u | while read flag; do
  grep -q "$flag" cmd/client/*.go cmd/server/*.go || echo "Missing: $flag"
done
```

---

## Step Dependencies

```
Step 1 (FPS CI) ─────────────────────────────────────┐
                                                      ├──→ Step 8 (Docs)
Step 2 (Memory CI) ──────────────────────────────────┘

Step 3 (handlers complexity) ───────────────────────────→ Step 6 (handlers duplication)

Step 4 (sprites/generator complexity) ──────────────────→ Step 7 (sprites/equipment complexity)

Step 5 (system_init duplication) ─────────────────────── [independent]
```

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Benchmark CI adds false positives | Medium | Low | Use 20% threshold (conservative given 130x headroom) |
| Refactoring introduces bugs | Low | Medium | Require all existing tests to pass; add tests for extracted helpers |
| Duplication consolidation changes behavior | Low | Medium | Keep extracted helpers internal; test each pattern |
| Performance tuning docs become stale | Medium | Low | Link to existing benchmarks; document process, not just values |

---

## External Considerations

### Ebiten v2.9 Deprecations
The project uses Ebiten v2.9.3. Key deprecations to monitor:
- `AppendVerticesAndIndicesForFilling/Stroke` → Use `vector.FillPath()`/`vector.StrokePath()`
- `ebiten.FillRule` deprecated
- `DrawTrianglesOptions.AntiAlias` deprecated

**Action**: Grep codebase for deprecated APIs before next major version bump:
```bash
grep -rn "AppendVerticesAndIndices\|FillRule\|AntiAlias" --include='*.go' pkg/
```

### Go Version
Project requires Go 1.24.5+ and CI tests on 1.23.x and 1.24.x. Ebiten 2.9 requires Go 1.24+.
**Action**: Consider removing Go 1.23.x from CI matrix as it's incompatible with Ebiten 2.9 requirements.

---

## Success Metrics

| Indicator | Current | Target | Measurement |
|-----------|---------|--------|-------------|
| FPS CI gate | ❌ Not implemented | ✅ Automated | CI workflow output |
| Memory CI gate | ❌ Not implemented | ✅ Automated | CI workflow output |
| Functions ≥20 complexity | 9 | ≤3 | `go-stats-generator --sections functions` |
| Functions ≥15 complexity | 27 | ≤10 | `go-stats-generator --sections functions` |
| Duplication ratio | 8.0% | ≤6% | `go-stats-generator --sections duplication` |
| Performance docs | ❌ Missing | ✅ Complete | `docs/PERFORMANCE_TUNING.md` exists |

---

## Appendix: Validation Commands

### Full Quality Check
```bash
# Run all validations
make quality

# Specific checks
go-stats-generator analyze . --skip-tests --format json --sections functions | \
  jq '[.functions[] | select(.complexity.cyclomatic >= 15)] | length'

go-stats-generator analyze . --skip-tests --format json --sections duplication | \
  jq '.duplication.duplication_ratio'
```

### Post-Implementation Verification
```bash
# After all steps complete, run full analysis
go-stats-generator analyze . --skip-tests --format json --sections functions,duplication,documentation > /tmp/final-metrics.json

# Compare against this plan's baseline
echo "Complexity ≥15: $(jq '[.functions[] | select(.complexity.cyclomatic >= 15)] | length' /tmp/final-metrics.json)"
echo "Duplication: $(jq '.duplication.duplication_ratio' /tmp/final-metrics.json)"
echo "Doc coverage: $(jq '.documentation_coverage.coverage.overall' /tmp/final-metrics.json)"
```

---

*Generated: 2026-03-21 | Tool: go-stats-generator v1.0.0 | Analysis time: ~11 minutes*
