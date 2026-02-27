# Testing Guide

**Version:** 1.0.0  
**Last Updated:** February 2026

Comprehensive testing strategy, infrastructure, and best practices.

---

## Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Test Infrastructure](#test-infrastructure)
3. [Writing Tests](#writing-tests)
4. [Running Tests](#running-tests)
5. [Coverage](#coverage)

---

## Testing Philosophy

**Principles:**
- Deterministic: Same seed = same result
- Fast: Tests run in <10s total
- Isolated: No dependencies between tests
- Table-driven: Multiple scenarios per test

**Coverage Target:** ≥40% per package (≥30% for packages depending on X11/Wayland/Ebiten)

**Current Coverage:** 82.4% average (engine 50%, procgen 73%, rendering 92%, audio high, saveload 71%, combat 100%, world 100%)

---

## Test Infrastructure

### Go Testing

Use Go's built-in testing package (no build tags required).

```go
func TestFunction(t *testing.T) {
    // Test implementation
}

func BenchmarkFunction(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Benchmark code
    }
}
```

### Stub Implementations

For testing without Ebiten runtime:

```go
// Production
sprite := ebiten.NewImage(32, 32)

// Testing
sprite := &StubSprite{Width: 32, Height: 32}
```

**Stubs:** StubInput, StubSprite, StubAudio, StubRenderer

### Table-Driven Tests

```go
func TestGenerator(t *testing.T) {
    tests := []struct {
        name    string
        seed    int64
        params  GenerationParams
        wantErr bool
    }{
        {"valid", 12345, validParams, false},
        {"invalid depth", 12345, invalidParams, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := gen.Generate(tt.seed, tt.params)
            if (err != nil) != tt.wantErr {
                t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## Writing Tests

### Unit Tests

Test single functions/methods in isolation.

```go
func TestPositionComponent(t *testing.T) {
    comp := &PositionComponent{X: 10, Y: 20}
    if comp.Type() != "position" {
        t.Errorf("Type() = %v, want position", comp.Type())
    }
}
```

### Integration Tests

Test multiple components working together.

```go
func TestMovementSystem(t *testing.T) {
    world := NewWorld()
    world.AddSystem(NewMovementSystem(200.0))
    entity := world.CreateEntity()
    entity.AddComponent(&PositionComponent{X: 0, Y: 0})
    entity.AddComponent(&VelocityComponent{VX: 10, VY: 0})
    world.Update(1.0)
    pos := entity.GetComponent("position").(*PositionComponent)
    if pos.X != 10 {
        t.Errorf("X = %v, want 10", pos.X)
    }
}
```

### Determinism Tests

Verify same seed produces same output.

```go
func TestDeterminism(t *testing.T) {
    seed := int64(12345)
    result1, _ := gen.Generate(seed, params)
    result2, _ := gen.Generate(seed, params)
    if !reflect.DeepEqual(result1, result2) {
        t.Error("Generator not deterministic")
    }
}
```

### Benchmark Tests

Measure performance of critical paths.

```go
func BenchmarkGenerate(b *testing.B) {
    gen := NewGenerator()
    params := GenerationParams{Difficulty: 0.5, Depth: 5}
    for i := 0; i < b.N; i++ {
        gen.Generate(12345, params)
    }
}
```

---

## Running Tests

### All Tests

```bash
go test ./...
```

### Package Tests

```bash
go test ./pkg/engine
go test ./pkg/procgen/terrain
```

### Specific Test

```bash
go test -run TestMovementSystem ./pkg/engine
```

### With Coverage

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Benchmarks

```bash
go test -bench=. ./...
go test -bench=BenchmarkGenerate -benchmem ./pkg/procgen/terrain
```

### Race Detection

```bash
go test -race ./...
```

### Headless Testing (X11/Display Required Packages)

Some packages (`cmd/server`, `pkg/world/housing`, `pkg/integration/*`) require a display due to Ebiten dependencies. On headless systems (CI, Docker, SSH sessions), use Xvfb (virtual framebuffer):

**Install Xvfb:**
```bash
# Ubuntu/Debian
sudo apt-get install xvfb

# Fedora/RHEL
sudo dnf install xorg-x11-server-Xvfb

# macOS (not typically needed - has native display)
brew install --cask xquartz
```

**Run tests with Xvfb:**
```bash
# Single package
xvfb-run -s "-screen 0 1920x1080x24" go test -v ./cmd/server/...

# All packages
xvfb-run -s "-screen 0 1920x1080x24" go test ./...

# With race detector and coverage
xvfb-run -s "-screen 0 1920x1080x24" go test -race -cover ./pkg/world/housing/... ./pkg/integration/...
```

**Configure persistent Xvfb (for interactive development):**
```bash
# Start virtual display on :99
Xvfb :99 -screen 0 1920x1080x24 &

# Set environment variable
export DISPLAY=:99

# Now run tests normally
go test ./cmd/server/...
```

**Note:** The CI pipeline (`.github/workflows/test.yml`) automatically uses `xvfb-run` for all test executions on Linux runners.

### Makefile Targets

```bash
make test           # Run all tests
make test-coverage  # Generate coverage report
make test-race      # Run with race detector
make bench          # Run benchmarks
```

---

## Coverage

### Current Coverage (by package)

| Package | Coverage | Notes |
|---------|----------|-------|
| engine | 50.0% | Ebiten-dependent code excluded |
| procgen | 73.1% | Core generation package |
| procgen/entity | 92.1% | High coverage |
| procgen/environment | 95.0% | High coverage |
| procgen/genre | 100% | Full coverage |
| procgen/item | 91.2% | High coverage |
| procgen/magic | 89.1% | High coverage |
| procgen/narrative | 93.7% | High coverage |
| procgen/puzzle | 93.4% | High coverage |
| procgen/quest | 91.3% | High coverage |
| procgen/skills | 86.1% | High coverage |
| procgen/station | 94.2% | High coverage |
| procgen/terrain | 93.3% | High coverage |
| rendering/lighting | 96.7% | Excellent coverage |
| rendering/palette | 97.0% | Excellent coverage |
| rendering/particles | 94.4% | High coverage |
| rendering/patterns | 78.7% | Good coverage |
| rendering/postprocess | 84.4% | Good coverage |
| rendering/quality | 96.6% | Excellent coverage |
| rendering/tiles | 92.5% | High coverage |
| rendering/ui | 94.9% | High coverage |
| audio/music | 96.6% | Excellent coverage |
| audio/sfx | 89.3% | High coverage |
| audio/synthesis | 94.2% | High coverage |
| saveload | 70.9% | Core functionality covered |
| combat | 100% | Full coverage |
| world | 100% | Full coverage |
| logging | 77.8% | Good coverage |
| **Average** | **82.4%** | Exceeds 40% target |

### Excluded from Coverage

Functions requiring Ebiten runtime (cannot run in CI without X11):
- `ebiten.NewImage()`
- Rendering operations (DrawImage, etc.)
- Audio playback
- Input handling (Ebiten-specific)

**Strategy:** Isolate Ebiten dependencies, use stub implementations for testing.

---

## Best Practices

**DO:**
- Write tests for all exported functions
- Use table-driven tests for multiple scenarios
- Test both success and error paths
- Verify determinism for generators
- Include benchmarks for performance-critical code
- Use descriptive test names

**DON'T:**
- Skip error checking in tests
- Use global state
- Depend on test execution order
- Hard-code paths (use relative or temp)
- Test Ebiten initialization in unit tests

---

## CI/CD Integration

GitHub Actions runs tests automatically:

```yaml
# .github/workflows/test.yml
- name: Run tests
  run: go test -race -cover ./...
```

**On:** Push to main, pull requests  
**Requirements:** Go 1.24+, xvfb (for display tests)

---

## Additional Resources

- [Development Guide](DEVELOPMENT.md) - Testing workflow
- [Contributing Guide](CONTRIBUTING.md) - Code quality standards
- [CI/CD Guide](CI_CD.md) - Continuous integration setup

**Test Examples:** See `*_test.go` files throughout codebase.

---

**Version:** 1.0.0  
**Last Updated:** February 2026  
**Maintained By:** Venture Development Team
