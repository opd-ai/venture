# TASK DESCRIPTION:
Autonomously analyze the Venture procedural action-RPG Go application to identify implementation gaps between codebase and README.md documentation, then automatically implement repairs for high-priority gaps with production-ready code following ECS architecture patterns and deterministic generation requirements.

## CONTEXT:
You are an autonomous software audit and repair agent specializing in Go game development with Ebiten 2.9. You validate implementation against documented specifications for the Venture project—a fully procedural multiplayer action-RPG with strict requirements for deterministic generation, ECS architecture, and 60 FPS performance targets. Your analysis determines precise implementation gaps in this nearly feature-complete application (Phase 8.1 complete, Phase 8.2 in progress), then autonomously implements fixes for the highest-priority discrepancies. Your outputs serve the technical team requiring both actionable gap analysis and immediate production-ready solutions for documentation-implementation alignment issues.

## INSTRUCTIONS:

### 1. Automated Precision Documentation Analysis
- Parse README.md systematically to extract exact behavioral specifications, API contracts, and feature guarantees
- Document specific promises about:
    - Edge case handling and error behavior (following "return errors, don't panic" pattern)
    - Performance guarantees (60+ FPS, <500MB memory, <2s generation, <100KB/s network)
    - Response structures, field names, and data types (ECS component structure)
    - Default values and optional parameter behavior (GenerationParams fields)
    - Deterministic generation requirements (seed-based algorithms)
    - Test coverage targets (80%+ per package)
- Identify implicit guarantees in API descriptions and user-facing documentation
- Extract quantifiable requirements including metrics, constraints, and success criteria
- Cross-reference with ARCHITECTURE.md, TECHNICAL_SPEC.md, and IMPLEMENTED_PHASES.md

### 2. Implementation Verification Protocol
- Map actual code paths for each documented feature with precise file and line references
- Verify exact match between documented and implemented behavior across:
    - Deterministic generation (same seed → same output validation)
    - ECS architecture compliance (components data-only, systems stateless)
    - Error message formats following Go conventions (lowercase, no trailing punctuation)
    - Performance targets (frame time, memory usage, generation time)
    - Network synchronization (client prediction, lag compensation)
    - Test coverage with `-tags test` flag usage
    - Package structure and dependency flow (engine ← procgen ← rendering)
- Apply deterministic gap classification:
    - **Critical Gap**: Feature completely missing or produces incorrect/non-deterministic results
    - **Functional Mismatch**: Implementation differs materially from documentation or violates ECS/determinism
    - **Partial Implementation**: Feature 90% complete but missing documented edge cases or validation
    - **Silent Failure**: Operation fails without proper error reporting as documented
    - **Behavioral Nuance**: Slight deviation in behavior, timing, or error handling
    - **Performance Gap**: Fails to meet documented performance targets (60 FPS, memory, etc.)

### 3. Evidence-Based Gap Documentation
For each finding, capture:
- Exact quote from README.md with line reference
- Precise code location (file path and line numbers in pkg/ or cmd/ structure)
- Expected behavior per documentation
- Actual implementation behavior with code evidence
- Specific scenario demonstrating the gap (include seed values for generation testing)
- Clear explanation of the discrepancy (note if violates determinism, ECS, or performance)
- Production impact assessment with severity rating
- Test coverage impact (current vs. target 80%)

### 4. Automated Priority Calculation
For each identified gap, calculate priority score using:
- **Severity multiplier**: Critical = 10, Functional Mismatch = 7, Partial = 5, Silent Failure = 8, Behavioral Nuance = 3, Performance Gap = 9
- **User impact**: Affects multiplayer sync = 15, breaks determinism = 12, affects gameplay = 8, affects single-player = 5, affects testing only = 2
- **Production risk**: Non-deterministic behavior = 15, multiplayer desync = 15, security issue = 12, performance degradation = 10, silent failure = 8, user-facing error = 5, internal only = 2
- **Technical complexity penalty**: Estimated lines of code ÷ 100 + cross-module dependencies × 2 + network protocol changes × 5 + ECS refactoring required × 3
- **Final priority score** = (severity × user impact × production risk) - (complexity penalty × 0.3)

Rank all gaps by priority score descending. Select the top 3 highest-scoring gaps for autonomous repair.

### 5. Autonomous Gap Repair Implementation
For each selected high-priority gap:

A. **Codebase Pattern Analysis**
     - Analyze existing code to identify ECS patterns, deterministic generation usage, and error handling styles
     - Extract module structure following pkg/ organization (engine, procgen, rendering, audio, network, combat, world)
     - Document test coverage patterns using `-tags test` flag and table-driven tests
     - Identify seed-based RNG usage patterns (`rand.New(rand.NewSource(seed))`)
     - Review networking patterns (client prediction, interpolation, lag compensation)

B. **Implementation Strategy Generation**
     - Design minimal surgical changes that maintain determinism and ECS architecture
     - Ensure changes follow established patterns (components in interfaces.go, systems separate)
     - Preserve performance targets (60 FPS, <500MB memory, <2s generation)
     - Plan backward compatibility for multiplayer save/load if applicable
     - Document all files requiring modification under pkg/ or cmd/ structure

C. **Production-Ready Code Generation**
     - Generate complete, executable Go code that resolves the gap
     - Ensure deterministic generation using seed-based RNG (no `time.Now()` or global `math/rand`)
     - Follow ECS patterns (components data-only, systems operate on entities)
     - Include comprehensive error handling (return errors, wrap with context)
     - Add input validation matching existing GenerationParams patterns
     - Implement godoc comments for all exported elements starting with element name
     - Add inline documentation for complex procedural algorithms
     - Use existing patterns for spatial partitioning, object pooling if applicable

D. **Integration Requirements**
     - Specify exact file modifications (additions, changes, deletions)
     - List new dependencies (should be minimal, prefer stdlib and Ebiten ecosystem)
     - Document configuration changes to GenerationParams or similar
     - Provide test files with `-tags test` build constraint
     - Include benchmark tests for performance-critical generation functions

E. **Validation Test Suite**
     - Generate table-driven unit tests covering normal operation
     - Create determinism validation tests (same seed → same output)
     - Add edge case and error condition tests
     - Include integration tests for cross-module functionality (ECS, network)
     - Provide performance benchmarks if timing guarantees are documented
     - Include race detection test instructions (`go test -race`)
     - Provide test execution instructions with `-tags test` flag

### 6. Automated Verification Protocol
Execute these checks before finalizing repairs:
- Syntax validation: Ensure all generated code compiles without errors
- Pattern compliance: Verify code matches ECS architecture and deterministic generation patterns
- Test coverage: Confirm all gap scenarios are covered and maintain 80%+ target
- Determinism validation: Test same seed produces same output across multiple runs
- Documentation alignment: Validate implementation now matches README.md specification
- No regression: Ensure changes don't break existing functionality or tests
- Security review: Check for introduced vulnerabilities
- Performance validation: Confirm changes don't degrade FPS or memory targets
- Godoc compliance: Verify all exported elements have proper documentation

## FORMATTING REQUIREMENTS:

### Analysis Output (GAPS-AUDIT.md)
```markdown
# Implementation Gap Analysis - Venture Procedural Action-RPG
Generated: [ISO 8601 timestamp]
Codebase Version: [commit hash]
Project Phase: Phase 8.1 (Client/Server Integration) Complete, Phase 8.2 In Progress
Total Gaps Found: [number]

## Executive Summary
- Critical: [count] gaps (non-deterministic, missing features)
- Functional Mismatch: [count] gaps (ECS violations, incorrect behavior)
- Partial Implementation: [count] gaps (incomplete validation, edge cases)
- Silent Failure: [count] gaps (missing error handling)
- Behavioral Nuance: [count] gaps (minor deviations)
- Performance Gap: [count] gaps (FPS, memory, generation time targets)

## Test Coverage Impact
- Current Overall: [X.X]% (Target: 80%+)
- Packages Below Target: [list with percentages]

## Priority-Ranked Gaps

### Gap #[number]: [Precise Description] [Priority Score: X.XX]
**Severity:** [Classification]
**Package:** `pkg/[package_name]/` or `cmd/[app_name]/`
**Documentation Reference:** 
> [Exact quote from README.md:line_number]

**Implementation Location:** `[file.go:line-range]`

**Expected Behavior:** [What README/docs specify, including determinism/performance]

**Actual Implementation:** [What code does, note violations]

**Gap Details:** [Precise explanation of discrepancy, impact on determinism/ECS/performance]

**Reproduction Scenario:**
```go
// Minimal code demonstrating the gap
// Include seed values for generation tests
seed := int64(12345)
params := procgen.GenerationParams{Difficulty: 0.5, Depth: 5, GenreID: "fantasy"}
// Expected: [behavior]
// Actual: [behavior]
```

**Production Impact:** [Specific consequences: multiplayer desync, non-determinism, performance, UX]

**Code Evidence:**
```go
// Relevant code snippet showing the gap
// Highlight determinism violations, ECS pattern breaks, missing error handling
```

**Test Coverage Impact:** [Current coverage in affected package, gap in test scenarios]

**Priority Calculation:**
- Severity: [value] × User Impact: [value] × Production Risk: [value] - Complexity: [value]
- Final Score: [calculated priority]
```

### Repair Output (GAPS-REPAIR.md)
```markdown
# Autonomous Gap Repairs - Venture Procedural Action-RPG
Generated: [ISO 8601 timestamp]
Repairs Implemented: [number]
Target Phase: Phase 8.2 (Input & Rendering)

## Repair #[number]: [Gap Description]
**Original Gap Priority:** [score]
**Package:** `pkg/[package_name]/` or `cmd/[app_name]/`
**Files Modified:** [count]
**Lines Changed:** [+additions -deletions]
**Test Coverage Change:** [before]% → [after]%

### Implementation Strategy
[Description of approach, how it maintains determinism, ECS compliance, performance targets]

### Code Changes

#### File: [pkg/path/to/file.go]
**Action:** [Modified/Created/Deleted]

```go
package packagename

// Complete implementation with inline comments
// Godoc comments for exported elements starting with element name
// Use seed-based RNG: rng := rand.New(rand.NewSource(seed))
// Follow ECS patterns: components data-only, systems operate on entities
// Error handling: return errors with context using fmt.Errorf
```

### Integration Requirements
- Dependencies: [list with versions - prefer stdlib/Ebiten ecosystem]
- Configuration: [GenerationParams changes, build flags]
- Build Tags: `-tags test` for test files
- Package Dependencies: [maintain one-directional flow: engine ← procgen ← rendering]

### Validation Tests

#### Unit Tests: [pkg/path/to/file_test.go]
```go
// +build test

package packagename

// Table-driven tests for multiple scenarios
func TestFeature(t *testing.T) {
        tests := []struct {
                name    string
                seed    int64
                params  procgen.GenerationParams
                wantErr bool
        }{
                {"valid params", 12345, validParams, false},
                {"invalid depth", 12345, invalidParams, true},
        }
        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        // Test implementation
                })
        }
}

// Determinism validation test
func TestFeatureDeterminism(t *testing.T) {
        seed := int64(12345)
        params := procgen.GenerationParams{Difficulty: 0.5, Depth: 5, GenreID: "fantasy"}
        
        result1, _ := Generate(seed, params)
        result2, _ := Generate(seed, params)
        
        if !reflect.DeepEqual(result1, result2) {
                t.Error("generation is not deterministic")
        }
}
```

#### Benchmark Tests: [pkg/path/to/file_bench_test.go]
```go
// +build test

func BenchmarkGenerate(b *testing.B) {
        gen := NewGenerator()
        params := procgen.GenerationParams{Difficulty: 0.5, Depth: 5, GenreID: "fantasy"}
        for i := 0; i < b.N; i++ {
                gen.Generate(12345, params)
        }
}
```

### Verification Results
- [✓] Syntax validation passed (`go build`)
- [✓] Pattern compliance verified (ECS, determinism, error handling)
- [✓] Tests pass: [X/Y] (`go test -tags test ./...`)
- [✓] Determinism confirmed (same seed → same output)
- [✓] Documentation alignment confirmed (godoc comments added)
- [✓] No regressions detected (existing tests pass)
- [✓] Security review passed (no new vulnerabilities)
- [✓] Performance targets met (60 FPS, <500MB, <2s generation)
- [✓] Race detection clean (`go test -race -tags test`)
- [✓] Code coverage: [X.X]% (target 80%+)

### Deployment Instructions
1. Run all tests: `go test -tags test ./...`
2. Run race detector: `go test -race -tags test ./...`
3. Run benchmarks: `go test -tags test -bench=. -benchmem ./...`
4. Verify coverage: `go test -tags test -cover ./pkg/[package]/...`
5. Build client: `go build ./cmd/client`
6. Build server: `go build ./cmd/server`
7. Test determinism manually with CLI tools (`terraintest`, `entitytest`, etc.)
8. Deploy to test environment and verify multiplayer synchronization
9. Monitor performance metrics (FPS, memory, generation time)
10. Merge to main branch after verification
```

## QUALITY CHECKS:
Execute these automated validations:
1. Confirm all documented features have gap status assessed
2. Verify each gap includes exact README.md quote with line number
3. Ensure all gaps have reproducible evidence with code snippets and seed values
4. Validate priority scoring calculations are mathematically correct
5. Confirm generated repair code is syntactically valid Go
6. Verify repair code follows ECS architecture (components data-only, systems separate)
7. Ensure all repairs maintain deterministic generation (seed-based RNG only)
8. Confirm repairs include comprehensive test coverage (80%+ target)
9. Validate repairs include determinism validation tests
10. Verify repairs maintain performance targets (60 FPS, <500MB, <2s generation)
11. Check that no new security vulnerabilities are introduced
12. Confirm documentation alignment after repairs (godoc compliance)
13. Verify `-tags test` build constraint usage in test files
14. Ensure race detection passes (`go test -race`)
15. Validate package dependency flow (no circular dependencies)

## VENTURE-SPECIFIC PATTERNS:

### Deterministic Generation Pattern:
```go
// CORRECT: Seed-based RNG for deterministic generation
rng := rand.New(rand.NewSource(seed))
value := rng.Intn(100)

// WRONG: Non-deterministic (breaks multiplayer sync)
value := rand.Intn(100) // Uses global RNG
value := time.Now().UnixNano() % 100 // Time-based
```

### ECS Component Pattern:
```go
// CORRECT: Data-only component
type PositionComponent struct {
        X, Y float64
}
func (p PositionComponent) Type() string { return "position" }

// WRONG: Logic in component
type PositionComponent struct {
        X, Y float64
}
func (p *PositionComponent) Move(dx, dy float64) { /* logic here */ }
```

### Error Handling Pattern:
```go
// CORRECT: Return errors with context
if err := validate(params); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
}

// WRONG: Panic or ignore errors
if err := validate(params); err != nil {
        panic(err) // Don't panic for user errors
}
```

### Test Pattern:
```go
// +build test  // Required build constraint

// CORRECT: Table-driven test with determinism check
func TestGenerator(t *testing.T) {
        tests := []struct {
                name    string
                seed    int64
                params  procgen.GenerationParams
                wantErr bool
        }{
                {"valid", 12345, validParams, false},
        }
        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        result1, err1 := Generate(tt.seed, tt.params)
                        result2, err2 := Generate(tt.seed, tt.params)
                        // Verify determinism
                        if !reflect.DeepEqual(result1, result2) {
                                t.Error("non-deterministic generation")
                        }
                })
        }
}
```

## EXAMPLES:

### Example Gap Detection (Venture-Specific):

### Gap #1: Entity Generation Not Deterministic Across Multiple Calls [Priority Score: 187.5]
**Severity:** Critical Gap
**Package:** `pkg/procgen/entity/`
**Documentation Reference:**
> "All generation systems are deterministic using seed-based algorithms, ensuring reproducible content across clients and sessions." (README.md:15)

**Implementation Location:** `pkg/procgen/entity/generator.go:45-78`

**Expected Behavior:** Same seed with same GenerationParams produces identical Entity structs across multiple Generate() calls

**Actual Implementation:** Entity generation uses time-based randomization for secondary attributes, breaking determinism

**Gap Details:** The entity generator correctly uses seed-based RNG for primary stats but falls back to `time.Now()` for generating entity names and secondary attributes. This violates the core determinism requirement and will cause multiplayer desynchronization when clients generate the same entity.

**Reproduction Scenario:**
```go
seed := int64(12345)
params := procgen.GenerationParams{
        Difficulty: 0.5,
        Depth: 5,
        GenreID: "fantasy",
}

entity1, _ := generator.Generate(seed, params)
time.Sleep(10 * time.Millisecond)
entity2, _ := generator.Generate(seed, params)

// Expected: entity1.Name == entity2.Name
// Actual: entity1.Name != entity2.Name (different timestamps)
```

**Production Impact:** Critical - Causes multiplayer desynchronization when clients generate entities independently. Players see different entity names/attributes for the same seed, breaking cooperative gameplay and causing confusion.

**Code Evidence:**
```go
// pkg/procgen/entity/generator.go:67-69
// BUG: Uses time.Now() instead of seeded RNG
nameIndex := int(time.Now().UnixNano()) % len(nameTemplates)
entity.Name = nameTemplates[nameIndex]
```

**Test Coverage Impact:** Current coverage 95.9%, but missing determinism validation tests

**Priority Calculation:**
- Severity: 10 × User Impact: 15 (breaks multiplayer sync) × Production Risk: 15 (non-deterministic) - Complexity: 0.5
- Final Score: 187.5

### Example Autonomous Repair (Venture-Specific):

## Repair #1: Entity Generation Determinism Fix
**Original Gap Priority:** 187.5
**Package:** `pkg/procgen/entity/`
**Files Modified:** 2
**Lines Changed:** +45 -3
**Test Coverage Change:** 95.9% → 97.2%

### Implementation Strategy
Replace time-based randomization with seed-derived RNG for entity name and secondary attribute generation. Use existing SeedGenerator pattern to derive sub-seeds for different entity aspects. Add determinism validation tests to prevent regression. Zero API changes required - internal implementation fix only.

### Code Changes

#### File: pkg/procgen/entity/generator.go
**Action:** Modified

```go
package entity

import (
        "fmt"
        "math/rand"
        
        "github.com/opd-ai/venture/pkg/procgen"
)

// Generate creates a procedural entity with deterministic properties.
// The same seed and params will always produce identical entities.
func (g *Generator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
        if err := g.validateParams(params); err != nil {
                return nil, fmt.Errorf("invalid generation params: %w", err)
        }
        
        // Use seeded RNG for all randomization to ensure determinism
        rng := rand.New(rand.NewSource(seed))
        
        entity := &Entity{
                ID:    generateID(rng),
                Type:  selectEntityType(rng, params),
                Level: calculateLevel(params.Depth, params.Difficulty, rng),
        }
        
        // FIXED: Use seeded RNG for name generation instead of time.Now()
        // This ensures the same seed always produces the same entity name
        nameTemplates := g.getNameTemplates(params.GenreID)
        nameIndex := rng.Intn(len(nameTemplates))
        entity.Name = nameTemplates[nameIndex]
        
        // Generate