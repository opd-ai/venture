# Testing Guide

**Version:** 1.0  
**Last Updated:** October 2025

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

**Coverage Target:** ≥65% per package (excluding Ebiten initialization)

**Current Coverage:** 82.4% average (engine 50%, procgen 100%, rendering 85%, audio high, saveload 67%, combat 100%, world 100%)

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
| procgen | 100% | Full coverage |
| procgen/* | 85-100% | High coverage across subpackages |
| rendering/sprites | 63.8% | Ebiten initialization excluded |
| rendering/tiles | 92.2% | Excellent coverage |
| audio/* | High | Implicit via music/sfx tests |
| saveload | 66.9% | Core functionality covered |
| combat | 100% | Full coverage |
| world | 100% | Full coverage |
| **Average** | **82.4%** | Exceeds 65% target |

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

**Version:** 1.0  
**Last Updated:** October 2025  
**Maintained By:** Venture Development Team
