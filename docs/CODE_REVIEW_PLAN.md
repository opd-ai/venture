# Code Review Methodology

**Version:** 1.0  
**Target Project:** Venture - Procedural Multiplayer Action-RPG  
**Technology Stack:** Go 1.24+, Ebiten 2.9.3, ECS Architecture  
**Last Updated:** November 2025

## Objective

A successful code review ensures the Venture codebase:
- **Eliminates defects** that could cause crashes, data corruption, or incorrect gameplay
- **Maintains robust API design** with clear contracts, proper error handling, and deterministic behavior
- **Maximizes downstream usability** through maintainable code, comprehensive tests, and clear documentation
- **Preserves architectural integrity** of the ECS pattern and procedural generation system
- **Meets performance targets** of 60 FPS with <500MB memory usage

**Completion Criteria:** All quality gates pass with documented justification for any exceptions.

---

## Pre-Review Setup

### 1. Environment Preparation
**Rationale:** Ensure consistent tooling and ability to validate code locally.

```bash
# Clone and setup repository
git clone https://github.com/opd-ai/venture.git
cd venture
go mod download
go mod verify

# Install development tools
go install mvdan.cc/gofumpt@latest
go install golang.org/x/tools/cmd/staticcheck@latest
# Optional: golangci-lint for comprehensive linting
```

### 2. Install Build Dependencies
**Rationale:** Install platform-specific dependencies required for building and testing Ebiten applications.

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get update
sudo apt-get install libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev \
                     libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev \
                     pkg-config libx11-dev xvfb
```

**macOS:**
```bash
xcode-select --install
```

**Windows:**
No additional dependencies required.

**Note:** `xvfb` (X virtual framebuffer) is required on Linux for running tests in headless environments (CI/CD, servers without displays). Tests should be run with `xvfb-run` on Linux:
```bash
xvfb-run -s "-screen 0 1920x1080x24" go test ./...
```

### 3. Baseline Validation
**Rationale:** Establish current state before reviewing changes.

```bash
# Verify all tests pass (use xvfb-run on Linux)
go test ./...

# Check code compiles without errors
go build ./cmd/client
go build ./cmd/server

# Run linters
go vet ./...
go fmt ./...

# Generate coverage baseline
go test -cover -coverprofile=coverage.out ./...
go tool cover -func=coverage.out > coverage-baseline.txt

# Run automated quality gates validation
make validate-code-review
# Or directly:
./scripts/validate-code-review.sh
```

**Note:** The `validate-code-review.sh` script automates verification of the first 9 quality gates.
Use `SKIP_RACE=true` and `SKIP_COVERAGE=true` environment variables to skip time-consuming checks during
initial validation.

### 4. Documentation Review
**Rationale:** Understand project architecture, patterns, and constraints.

**Required Reading:**
- `docs/ARCHITECTURE.md` - ECS patterns, ADRs, package structure
- `docs/TECHNICAL_SPEC.md` - Performance targets, data structures, core systems
- `docs/TESTING.md` - Testing philosophy, stub implementations, coverage targets
- `docs/CONTRIBUTING.md` - Code style, development workflow, quality standards
- `.github/copilot-instructions.md` - Project-specific guidelines and patterns

### 5. Review Scope Definition
**Rationale:** Focus review effort on changed/new code.

```bash
# Identify changed files
git diff --name-only <base-branch>...HEAD

# Review commit history
git log --oneline --graph <base-branch>..HEAD

# Check for new packages
find pkg/ -type d -newer <baseline-date>
```

---

## Review Process

### Phase 1: Static Analysis & Structure Review

**Goal:** Identify syntax errors, code smells, and structural issues before deeper analysis.

#### Step 1.1: Automated Static Analysis
**Action:** Run all available static analysis tools.

```bash
# Standard Go tooling
go vet ./...
gofmt -s -l .
staticcheck ./...

# Optional comprehensive linting
golangci-lint run
```

**What to check:**
- Unused variables, imports, or functions
- Printf format string mismatches
- Suspicious constructs (e.g., defer in loops)
- Shadowed variables
- Unreachable code

**Pass criteria:**
- Zero errors from `go vet`
- Zero files listed by `gofmt -l` (all code formatted)
- Zero high-priority issues from staticcheck/golangci-lint

**Fail criteria:**
- Any errors reported by `go vet`
- Unformatted code files
- Critical or high-severity linter warnings

#### Step 1.2: Package Structure Inspection
**Action:** Verify package organization follows project conventions.

**What to check:**
- All packages under `pkg/` with clear, single responsibilities
- No circular dependencies between packages
- Dependency flow: `engine` ← `procgen` ← `rendering` ← UI
- Each package has `doc.go` with comprehensive package documentation
- Public APIs in `interfaces.go` files where appropriate

**Pass criteria:**
- Package boundaries align with documented architecture
- Dependencies flow in correct direction (no cycles)
- All public packages have `doc.go` with ≥3 paragraphs of documentation

**Fail criteria:**
- Circular dependencies detected
- Packages mixing concerns (e.g., rendering code in engine package)
- Missing package documentation for any public package

#### Step 1.3: File Organization Review
**Action:** Ensure files follow naming conventions and size limits.

**What to check:**
- Test files named `*_test.go` in same package
- No files exceeding 1000 lines (suggests lack of separation)
- Related functionality grouped in appropriately named files
- `interfaces.go` for interface definitions
- Generator files include both `Generate()` and `Validate()` methods

**Pass criteria:**
- File naming follows Go conventions
- No single file >1000 lines without documented justification
- Test files present for all substantial logic files

**Fail criteria:**
- Test files in wrong location or named incorrectly
- Mega-files (>1500 lines) without clear separation points
- Public interfaces not in `interfaces.go`

---

### Phase 2: API Design & Interface Inspection

**Goal:** Ensure public APIs are well-designed, documented, and maintain backward compatibility.

#### Step 2.1: Interface Completeness Check
**Action:** Review all exported types, functions, and methods.

**What to check:**
- All exported identifiers have godoc comments starting with the identifier name
- Function signatures use appropriate parameter types (avoid `interface{}` when possible)
- Return multiple values appropriately (value + error pattern)
- Methods receiver types are consistent (pointer vs. value)
- Interface definitions are minimal and focused

**Pass criteria:**
- 100% of exported identifiers have godoc comments
- Clear rationale for any `interface{}` parameters
- All error-prone operations return `error` type
- Receiver types documented if mixing pointer/value

**Fail criteria:**
- Any exported identifier without documentation
- Functions that can fail but don't return errors
- Inconsistent receiver types without justification

#### Step 2.2: Generator Interface Compliance
**Action:** Verify all procedural generators follow established patterns.

**What to check:**
- Implements `Generate(seed int64, params GenerationParams) (interface{}, error)`
- Implements `Validate(result interface{}) error`
- Uses deterministic seed-based RNG (no `math/rand` global, no `time.Now()`)
- Returns typed results suitable for type assertion
- Supports `GenreID` parameter in `GenerationParams`

**Pass criteria:**
- All generators implement required interface methods
- All RNG uses `rand.New(rand.NewSource(seed))`
- Validation checks are meaningful (not just `return nil`)
- Genre support verified for all five genres

**Fail criteria:**
- Generator using non-deterministic randomness
- Missing `Validate()` implementation
- Genre parameter ignored
- Type assertion patterns unclear or unsafe

#### Step 2.3: ECS Component Standards
**Action:** Review component implementations for data purity.

**What to check:**
- Components are pure data structures (no business logic)
- All components implement `Type() string` method
- Type identifiers are lowercase, single-word (matching component name sans "Component")
- No methods beyond simple getters and `Type()`
- Fields are appropriately exported/unexported

**Pass criteria:**
- All components are data-only (no logic methods)
- `Type()` returns consistent identifiers
- No computed fields (calculation belongs in systems)

**Fail criteria:**
- Component contains business logic methods
- `Type()` method missing or returns inconsistent values
- Complex initialization logic in component constructors

#### Step 2.4: System Interface Verification
**Action:** Ensure systems follow ECS update pattern.

**What to check:**
- Systems implement `Update(deltaTime float64)` or equivalent
- Systems query entities via `world.GetEntitiesWith(componentTypes...)`
- Systems are stateless or manage minimal state
- No direct entity creation (only through world)
- Proper cleanup of removed entities

**Pass criteria:**
- All systems use world queries for entity access
- State management is explicit and minimal
- Entity lifecycle properly managed
- Systems don't cache entity references across updates

**Fail criteria:**
- Systems storing stale entity references
- Direct component manipulation without queries
- Unbounded state growth in systems

---

### Phase 3: Ebiten Pattern Compliance Check

**Goal:** Ensure code follows Ebiten best practices and avoids common pitfalls.

#### Step 3.1: Image Creation & Management
**Action:** Review all Ebiten image operations.

**What to check:**
- Images created once during initialization, not per-frame
- No `ebiten.NewImage()` calls in game loop or `Draw()` methods
- Proper image disposal if dynamically created
- Image caching for procedurally generated sprites (see `pkg/rendering/cache/`)
- Avoid excessive `DrawImage()` calls (use batching)

**Pass criteria:**
- Zero image allocations in hot paths (Update/Draw)
- Sprite cache hit rate >90% for generated content
- Proper resource pooling for temporary images

**Fail criteria:**
- Image creation in `Draw()` or `Update()` methods
- No caching strategy for generated sprites
- Memory leaks from undisposed images

#### Step 3.2: Input Handling Patterns
**Action:** Verify input processing follows Ebiten conventions.

**What to check:**
- Input checked in `Update()`, not `Draw()`
- No direct `ebiten` package imports in non-rendering code (use interfaces)
- Keyboard/mouse/gamepad state checked frame-perfect
- Mobile touch input handled via `pkg/mobile/` interfaces
- Input buffering for networked games

**Pass criteria:**
- All input processing in `Update()` method
- Input abstracted behind interfaces for testability
- Touch input properly translated to game actions

**Fail criteria:**
- Input checks in `Draw()` method
- Direct Ebiten dependencies in core game logic
- Missing touch input support where required

#### Step 3.3: Audio Synthesis Verification
**Action:** Check audio generation and playback patterns.

**What to check:**
- Audio players created once, not per sound event
- Waveform generation uses `pkg/audio/synthesis/` properly
- Music composition follows `pkg/audio/music/` patterns
- Sound effects use `pkg/audio/sfx/` with pooling
- Audio buffers properly sized (avoid underruns)
- Volume and panning controls implemented

**Pass criteria:**
- Audio player pooling implemented
- No audio generation in game loop
- Procedural audio deterministic with seed

**Fail criteria:**
- Audio player creation per sound event
- Buffer underruns causing audio glitches
- Non-deterministic audio synthesis

#### Step 3.4: Rendering Performance Check
**Action:** Validate rendering optimizations are in place.

**What to check:**
- Viewport culling active (only draw visible entities)
- Batch rendering by layer/texture type
- Sprite caching for procedurally generated content
- Object pooling for frequently created/destroyed objects
- Draw call minimization (target <1000 per frame)

**Pass criteria:**
- Viewport culling verified with >1000 entities
- Sprite cache hit rate >95% during steady gameplay
- Frame time <16.67ms (60 FPS) with 2000 entities
- Memory usage <500MB client target met

**Fail criteria:**
- No viewport culling implementation
- Sprite cache disabled or ineffective
- Frame rate drops below 60 FPS with <500 entities
- Excessive memory allocations per frame

---

### Phase 4: Concurrency & Resource Safety Audit

**Goal:** Identify race conditions, deadlocks, and resource leaks.

#### Step 4.1: Race Detection Testing
**Action:** Run comprehensive race detector tests.

```bash
# Run all tests with race detector
go test -race ./...

# Run race detector on long-running integration tests
go test -race -timeout 5m ./pkg/network/...
go test -race -timeout 5m ./pkg/engine/...
```

**What to check:**
- Zero race conditions detected
- Goroutine launches are controlled and bounded
- Shared state protected by mutexes or channels
- WaitGroup usage for goroutine lifecycle
- Context cancellation for graceful shutdown

**Pass criteria:**
- All tests pass with `-race` flag
- Goroutine counts remain bounded (no leaks)
- Proper synchronization primitives used

**Fail criteria:**
- Any race conditions detected
- Goroutine leaks (count grows unbounded)
- Unprotected shared state access

#### Step 4.2: Multiplayer Synchronization Review
**Action:** Examine network synchronization patterns.

**What to check:**
- Client-side prediction properly reconciles with server state
- Entity interpolation buffers sized appropriately (100-200ms)
- Server is authoritative for all game-critical state
- Lag compensation uses snapshot history correctly
- State delta compression reduces bandwidth
- Network components use proper locking

**Pass criteria:**
- Prediction-reconciliation loop verified
- Interpolation smooth at 200-5000ms latency
- Server authority enforced for critical operations
- Bandwidth <100KB/s per player target met

**Fail criteria:**
- Client can manipulate authoritative state
- Prediction never reconciles (divergence)
- Lag compensation allows cheating
- Network bandwidth exceeds targets

#### Step 4.3: Memory Leak Detection
**Action:** Profile for memory leaks and excessive allocations.

```bash
# Run memory profiling
go test -memprofile=mem.prof -bench=. ./pkg/engine/
go tool pprof -alloc_space mem.prof

# Check for leaks in long-running scenarios
go test -memprofile=mem-long.prof -run TestLongRunning -timeout 10m
```

**What to check:**
- Memory usage stabilizes (no continuous growth)
- Object pooling used for frequent allocations
- Sprite cache size bounded and purges old entries
- Entity deletion properly releases components
- Goroutines cleaned up on connection close

**Pass criteria:**
- Memory usage flat after initialization period
- All pools have maximum size limits
- Zero memory leaks in 10-minute stress test
- Cache eviction policies functional

**Fail criteria:**
- Unbounded memory growth over time
- Pools growing without limits
- Entities not garbage collected when removed

#### Step 4.4: Resource Cleanup Verification
**Action:** Check for proper resource disposal.

**What to check:**
- Images disposed when no longer needed
- Audio players stopped and released
- Network connections closed with timeout
- File handles closed (save/load system)
- Contexts cancelled on shutdown
- Defer statements for cleanup in error paths

**Pass criteria:**
- All resource acquisitions paired with cleanup
- Defer used for error-path cleanup
- No file descriptor leaks
- Clean shutdown verified

**Fail criteria:**
- Resources not cleaned up on error
- File handles remain open after operations
- Network connections not closed

---

### Phase 5: Error Handling & Edge Case Review

**Goal:** Ensure robust error handling and comprehensive edge case coverage.

#### Step 5.1: Error Propagation Analysis
**Action:** Trace error handling from bottom to top of call stack.

**What to check:**
- All error returns checked (no `_ = functionCall()` for error-returning functions)
- Errors wrapped with context using `fmt.Errorf("context: %w", err)`
- Panics only for programmer errors in init, not runtime conditions
- Validation errors returned, not logged and ignored
- Critical errors logged with structured logging (logrus)

**Pass criteria:**
- Zero unchecked errors
- All errors wrapped with meaningful context
- No panic usage for recoverable errors
- Structured logging for all significant errors

**Fail criteria:**
- Unchecked error returns
- Generic error messages without context
- Panic used for user input validation
- Errors swallowed without logging

#### Step 5.2: Input Validation Review
**Action:** Verify all external inputs are validated.

**What to check:**
- Command-line flags validated (ranges, formats)
- Network messages validated before deserialization
- User input sanitized (inventory actions, crafting recipes)
- File format validation in save/load system
- Seed values validated (overflow, special values)
- Generation parameters validated (difficulty 0.0-1.0, depth ≥0)

**Pass criteria:**
- All external inputs validated before use
- Validation failures return clear error messages
- Invalid input cannot crash the application
- Bounds checking on all array/slice accesses

**Fail criteria:**
- Network data deserialized without validation
- User input directly used in calculations
- Array access without bounds checking
- File parsing without format verification

#### Step 5.3: Edge Case Testing Coverage
**Action:** Review test suite for edge case coverage.

**What to check:**
- Boundary value tests (0, 1, max, overflow)
- Empty collection handling (empty world, no entities)
- Maximum capacity tests (2000+ entities)
- Concurrent access patterns in multiplayer
- Network disconnect/reconnect scenarios
- Save/load with corrupted data
- Invalid genre/seed combinations

**Pass criteria:**
- Edge cases have dedicated test cases
- Boundary values explicitly tested
- Error injection tests for failure paths
- Table-driven tests cover ≥5 scenarios per function

**Fail criteria:**
- Only happy-path tests present
- No tests for empty/zero/max values
- Error paths untested
- <3 test scenarios per complex function

#### Step 5.4: Determinism Verification
**Action:** Ensure procedural generation is deterministic.

**What to check:**
- All generators produce identical output with same seed
- No usage of `time.Now()`, `math/rand` global, or system entropy
- RNG instances created with `rand.New(rand.NewSource(seed))`
- Same seed + same parameters = identical results (verified by tests)
- Seed derivation consistent across runs
- Multiplayer clients generate same world from shared seed

**Pass criteria:**
- All generator tests verify determinism (run twice, compare)
- Zero non-deterministic random sources found
- Seed derivation uses `procgen.SeedGenerator` pattern
- Multiplayer sync tests pass at 200ms+ latency

**Fail criteria:**
- Any generator fails determinism test
- `time.Now()` or global `rand` used in generation
- Seed derivation produces different values
- Multiplayer desync due to generation differences

---

## Quality Gates

All gates must pass for code review completion:

- [ ] **Build Success**: `go build ./cmd/client` and `go build ./cmd/server` complete without errors
- [ ] **Test Pass**: `go test ./...` shows 100% pass rate
- [ ] **Race Freedom**: `go test -race ./...` reports zero race conditions
- [ ] **Code Coverage**: Package coverage ≥65% (excluding Ebiten initialization functions—i.e., functions requiring `ebiten.NewImage()`, rendering operations, or audio playback; see [TESTING.md](../docs/TESTING.md) for details)
- [ ] **Static Analysis**: `go vet ./...` reports zero issues
- [ ] **Code Formatting**: `gofmt -l .` returns empty (all files formatted)
- [ ] **Documentation Complete**: All exported identifiers have godoc comments
- [ ] **Package Docs Present**: All public packages have `doc.go` files
- [ ] **No Circular Dependencies**: Package import graph is acyclic
- [ ] **Performance Targets Met**: 60 FPS with 2000 entities, <500MB memory
- [ ] **Determinism Verified**: Generator tests prove same seed = same output
- [ ] **ECS Pattern Compliance**: Components are data-only, systems use queries
- [ ] **Error Handling**: All errors checked, wrapped with context, properly logged
- [ ] **Input Validation**: All external inputs validated before use
- [ ] **Resource Cleanup**: All resources properly disposed (verified with `-race` and memory profiling)
- [ ] **API Documentation**: Public APIs documented with usage examples
- [ ] **Multiplayer Sync**: Client-server synchronization tested at high latency (200-5000ms)
- [ ] **Genre Compatibility**: All generators tested with all five genres

---

## Final Assessment

### If All Quality Gates Pass

**Conclusion:** The code meets Venture's quality standards and is approved for merge.

**Required Actions:**
1. Document any exceptions to quality gates with justification
2. Update `docs/ARCHITECTURE.md` if architectural patterns changed
3. Update `docs/API_REFERENCE.md` if public APIs added/changed
4. Create release notes if user-facing features added
5. Merge to main branch with confidence

### If Any Gate Fails

**Conclusion:** The code requires remediation before approval.

**Required Actions:**
1. Document all failing gates with specific failure details
2. Create actionable tasks for each failure:
   - Link to failing test output
   - Specify exact files/functions requiring fixes
   - Provide example of correct implementation pattern
3. Re-run full review process after fixes implemented
4. Do not merge until all gates pass

---

## Appendix A: Quick Reference Checklist

**Before Review:**
- [ ] Clone repo, run `go mod download`
- [ ] Install build dependencies (see Pre-Review Setup step 2)
- [ ] Run baseline tests: `go test ./...` (Linux: use `xvfb-run -s "-screen 0 1920x1080x24" go test ./...`)
- [ ] Run automated validation: `make validate-code-review` or `./scripts/validate-code-review.sh`
- [ ] Read ARCHITECTURE, TECHNICAL_SPEC, TESTING docs
- [ ] Identify changed files with `git diff`

**Phase 1 - Static Analysis:**
- [ ] `go vet ./...` - zero errors
- [ ] `gofmt -l .` - zero files
- [ ] Package structure follows conventions
- [ ] File sizes reasonable (<1000 lines)

**Phase 2 - API Design:**
- [ ] All exports have godoc comments
- [ ] Generators implement required interface
- [ ] Components are data-only with `Type()` method
- [ ] Systems use world queries

**Phase 3 - Ebiten Patterns:**
- [ ] No image creation in game loop
- [ ] Input handled in `Update()` only
- [ ] Audio uses synthesis packages properly
- [ ] Rendering optimizations active

**Phase 4 - Concurrency:**
- [ ] `go test -race ./...` - zero races
- [ ] Memory profiling shows no leaks
- [ ] Resources cleaned up properly
- [ ] Multiplayer sync verified

**Phase 5 - Error Handling:**
- [ ] All errors checked and wrapped
- [ ] Input validation comprehensive
- [ ] Edge cases tested
- [ ] Determinism verified for generators

**Final Gates:**
- [ ] All 18 quality gates pass
- [ ] Documentation updated
- [ ] Ready for merge or requires remediation documented

---

## Appendix B: Common Failure Patterns & Solutions

### Pattern 1: Non-Deterministic Generation
**Symptom:** Generator produces different output with same seed.  
**Common Causes:**
- Using `time.Now()` for randomness
- Using global `math/rand` functions
- Reading system entropy
- Iterating maps in random order (map iteration is non-deterministic in Go)

**Solution:**
```go
// BAD: Non-deterministic
value := rand.Intn(100)  // Uses global seed
timestamp := time.Now().Unix()

// GOOD: Deterministic
rng := rand.New(rand.NewSource(seed))
value := rng.Intn(100)

// BAD: Map iteration order varies
for key, val := range entityMap {
    processEntity(key, val)  // Non-deterministic order
}

// GOOD: Sort keys first
keys := make([]string, 0, len(entityMap))
for k := range entityMap {
    keys = append(keys, k)
}
sort.Strings(keys)
for _, key := range keys {
    processEntity(key, entityMap[key])  // Deterministic order
}
```

### Pattern 2: Component Logic Violation
**Symptom:** Components contain business logic methods.  
**Solution:**
```go
// BAD: Logic in component
type HealthComponent struct {
    Current, Max int
}
func (h *HealthComponent) TakeDamage(amount int) {
    h.Current -= amount  // Logic doesn't belong here
}

// GOOD: Component is pure data
type HealthComponent struct {
    Current, Max int
}
func (h HealthComponent) Type() string { return "health" }

// Logic in system
func (s *CombatSystem) ApplyDamage(entity *Entity, amount int) {
    health := entity.GetComponent("health").(*HealthComponent)
    health.Current = max(0, health.Current - amount)
}
```

### Pattern 3: Unchecked Errors
**Symptom:** Error returns ignored, potential crashes.  
**Solution:**
```go
// BAD: Unchecked error
data, _ := loadFile(path)  // Ignores error

// GOOD: Proper error handling
data, err := loadFile(path)
if err != nil {
    return fmt.Errorf("failed to load file %s: %w", path, err)
}
```

### Pattern 4: Image Creation in Game Loop
**Symptom:** Frame rate drops, memory usage grows.  
**Solution:**
```go
// BAD: Creates image every frame
func (g *Game) Draw(screen *ebiten.Image) {
    sprite := ebiten.NewImage(32, 32)  // Memory leak!
    // ... draw sprite
}

// GOOD: Create once, cache
type Game struct {
    spriteCache map[string]*ebiten.Image
}

func (g *Game) getSprite(key string) *ebiten.Image {
    if sprite, ok := g.spriteCache[key]; ok {
        return sprite
    }
    sprite := ebiten.NewImage(32, 32)
    // ... generate sprite
    g.spriteCache[key] = sprite
    return sprite
}
```

### Pattern 5: Missing Input Validation
**Symptom:** Crashes on invalid input, security issues.  
**Solution:**
```go
// BAD: No validation
func (g *Generator) Generate(seed int64, params GenerationParams) (interface{}, error) {
    depth := params.Depth  // Could be negative or enormous
    difficulty := params.Difficulty  // Could be outside [0,1]
    // ... generate using unsafe values
}

// GOOD: Validate first
func (g *Generator) Generate(seed int64, params GenerationParams) (interface{}, error) {
    if params.Depth < 0 || params.Depth > 100 {
        return nil, fmt.Errorf("depth %d outside valid range [0,100]", params.Depth)
    }
    if params.Difficulty < 0.0 || params.Difficulty > 1.0 {
        return nil, fmt.Errorf("difficulty %f outside valid range [0.0,1.0]", params.Difficulty)
    }
    // ... safe to proceed
}
```

---

## Appendix C: Performance Profiling Quick Guide

### CPU Profiling
```bash
# Profile a specific benchmark
go test -cpuprofile=cpu.prof -bench=BenchmarkGenerate ./pkg/procgen/terrain/
go tool pprof cpu.prof

# Interactive commands in pprof
(pprof) top20        # Show top 20 CPU consumers
(pprof) list Generate  # Show annotated source for Generate function
(pprof) web          # Generate call graph (requires graphviz)
```

### Memory Profiling
```bash
# Profile memory allocations
go test -memprofile=mem.prof -bench=. ./pkg/rendering/sprites/
go tool pprof -alloc_space mem.prof

# Check for leaks
(pprof) top20
(pprof) list NewSprite
(pprof) traces      # Show allocation traces
```

### Live Application Profiling
```bash
# Add to main.go:
import _ "net/http/pprof"
import "net/http"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

# Then profile running app:
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
go tool pprof http://localhost:6060/debug/pprof/heap
```

### Benchmark Interpretation
Target metrics from requirements:
- **Frame time:** <16.67ms (60 FPS)
- **Memory:** <500MB client, <1GB server (4 players)
- **Generation time:** <2s per world area
- **Bandwidth:** <100KB/s per player

Any benchmark exceeding these targets requires optimization.

---

**End of Code Review Methodology Plan**
