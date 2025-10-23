# Contributing to Venture

Thank you for your interest in contributing to Venture! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Getting Started](#getting-started)
3. [Development Setup](#development-setup)
4. [Making Changes](#making-changes)
5. [Testing](#testing)
6. [Code Style](#code-style)
7. [Pull Request Process](#pull-request-process)
8. [Reporting Bugs](#reporting-bugs)
9. [Suggesting Features](#suggesting-features)
10. [Project Structure](#project-structure)

---

## Code of Conduct

### Our Pledge

We pledge to make participation in our project a harassment-free experience for everyone, regardless of age, body size, disability, ethnicity, gender identity and expression, level of experience, nationality, personal appearance, race, religion, or sexual identity and orientation.

### Our Standards

**Positive behavior includes:**
- Using welcoming and inclusive language
- Being respectful of differing viewpoints
- Accepting constructive criticism gracefully
- Focusing on what is best for the community
- Showing empathy towards other community members

**Unacceptable behavior includes:**
- Harassment, trolling, or insulting comments
- Publishing others' private information
- Other conduct which could be considered inappropriate

---

## Getting Started

### Prerequisites

- Go 1.24.7 or later
- Git for version control
- See [Development Guide](DEVELOPMENT.md) for complete setup instructions

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/venture.git
   cd venture
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/opd-ai/venture.git
   ```

### Quick Development Setup

```bash
# Download dependencies
go mod download

# Verify setup
go test -tags test ./...
go build ./cmd/client
go build ./cmd/server
```

**For detailed development environment setup, build instructions, profiling, and debugging, see [Development Guide](DEVELOPMENT.md).**

---

## Making Changes

### Project Philosophy

Venture follows these core principles:

1. **Deterministic Generation**: All procedural generation must use seed-based deterministic algorithms
2. **ECS Architecture**: Maintain separation between entities, components, and systems
3. **Performance First**: Target 60 FPS with 2000+ entities
4. **No External Assets**: Everything generated at runtime
5. **Multiplayer Support**: Consider network synchronization

### Types of Contributions

**Bug Fixes:**
- Fix broken functionality
- Resolve crashes or errors
- Improve stability

**Features:**
- New game mechanics
- Additional content generators
- UI improvements
- Performance optimizations

**Documentation:**
- Improve existing docs
- Add examples
- Write tutorials
- Fix typos

**Testing:**
- Add unit tests
- Create integration tests
- Improve test coverage
- Add benchmarks

### Deterministic Generation Rule

**CRITICAL:** All procedural generation must be deterministic!

```go
// ❌ BAD: Non-deterministic
func Generate() {
    value := rand.Intn(100) // Uses global random state
}

// ✅ GOOD: Deterministic
func Generate(seed int64) {
    rng := rand.New(rand.NewSource(seed))
    value := rng.Intn(100)
}
```

Never use:
- `time.Now()` for randomness
- Global `math/rand` functions
- System-dependent randomness

Always use:
- Seeded `rand.New(rand.NewSource(seed))`
- Deterministic algorithms
- Same seed = same result

### ECS Guidelines

**Components:** Pure data structures

```go
// ✅ GOOD: Only data
type PositionComponent struct {
    X, Y float64
}

func (p PositionComponent) Type() string {
    return "position"
}

// ❌ BAD: Logic in component
type PositionComponent struct {
    X, Y float64
}

func (p *PositionComponent) Move(dx, dy float64) {
    p.X += dx
    p.Y += dy
}
```

**Systems:** Contain all logic

```go
// ✅ GOOD: Logic in system
type MovementSystem struct{}

func (s *MovementSystem) Update(entities []*Entity, dt float64) {
    for _, entity := range entities {
        if !entity.HasComponent("position") || !entity.HasComponent("velocity") {
            continue
        }
        
        pos := entity.GetComponent("position").(*PositionComponent)
        vel := entity.GetComponent("velocity").(*VelocityComponent)
        
        pos.X += vel.VX * dt
        pos.Y += vel.VY * dt
    }
}
```

---

## Testing

### Test Requirements

- **Coverage Target**: 80% minimum per package
- **Test Tags**: Use `-tags test` for all tests
- **Table-Driven Tests**: For multiple scenarios
- **Benchmarks**: For performance-critical code

### Writing Tests

```go
func TestMyFeature(t *testing.T) {
    // Table-driven test
    tests := []struct {
        name    string
        input   int
        want    int
        wantErr bool
    }{
        {"positive", 5, 10, false},
        {"negative", -5, 0, true},
        {"zero", 0, 0, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Determinism Tests

Always test that generation is deterministic:

```go
func TestDeterministicGeneration(t *testing.T) {
    gen := NewGenerator()
    seed := int64(12345)
    params := procgen.GenerationParams{...}
    
    // Generate twice with same seed
    result1, err := gen.Generate(seed, params)
    if err != nil {
        t.Fatal(err)
    }
    
    result2, err := gen.Generate(seed, params)
    if err != nil {
        t.Fatal(err)
    }
    
    // Results must be identical
    if !reflect.DeepEqual(result1, result2) {
        t.Error("Generation is not deterministic")
    }
}
```

### Benchmarks

Add benchmarks for performance-critical code:

```go
func BenchmarkGenerate(b *testing.B) {
    gen := NewGenerator()
    seed := int64(12345)
    params := procgen.GenerationParams{...}
    
    for i := 0; i < b.N; i++ {
        gen.Generate(seed, params)
    }
}
```

Run benchmarks:
```bash
go test -tags test -bench=. -benchmem ./...
```

---

## Code Style

### Go Conventions

Follow standard Go conventions:

1. **Formatting**: Use `go fmt`
2. **Linting**: Pass `go vet`
3. **Naming**: Use `MixedCaps`, not `snake_case`
4. **Error handling**: Always check errors
5. **Comments**: Document all exported items

### Documentation

**Package Documentation:**

Every package needs a `doc.go` file:

```go
// Package mypackage provides functionality for X.
//
// This package implements Y using Z algorithm.
//
// Example usage:
//     gen := mypackage.NewGenerator()
//     result, err := gen.Generate(seed, params)
//
package mypackage
```

**Function Documentation:**

```go
// GenerateTerrain creates a procedural dungeon using BSP algorithm.
//
// The seed parameter ensures deterministic generation. The same seed
// with the same params will always produce identical terrain.
//
// Parameters:
//   - seed: Random seed for generation
//   - params: Configuration including width, height, difficulty
//
// Returns the generated terrain or an error if validation fails.
func GenerateTerrain(seed int64, params GenerationParams) (*Terrain, error) {
    // ...
}
```

### Error Handling

```go
// ✅ GOOD: Check all errors
result, err := DoSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// ❌ BAD: Ignore errors
result, _ := DoSomething()

// ✅ GOOD: Wrap errors with context
if err != nil {
    return fmt.Errorf("generating terrain at depth %d: %w", depth, err)
}

// ❌ BAD: Lose error context
if err != nil {
    return err
}
```

### File Organization

```go
// 1. Package declaration and imports
package mypackage

import (
    "fmt"
    "math/rand"
    
    "github.com/opd-ai/venture/pkg/procgen"
)

// 2. Constants
const (
    MaxWidth  = 100
    MaxHeight = 100
)

// 3. Type definitions
type Generator struct {
    // fields
}

// 4. Constructor
func NewGenerator() *Generator {
    return &Generator{}
}

// 5. Methods
func (g *Generator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
    // implementation
}

// 6. Helper functions (unexported)
func helper() {
    // ...
}
```

---

## Pull Request Process

### Before Submitting

1. **Update your branch** with latest upstream:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all checks**:
   ```bash
   go test -tags test ./...
   go vet ./...
   go fmt ./...
   ```

3. **Update documentation** if needed

4. **Add tests** for new features

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
Describe testing done

## Checklist
- [ ] Code follows project style
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] All tests pass
- [ ] No new warnings from go vet
```

### Review Process

1. Maintainer will review your PR
2. Address any requested changes
3. Once approved, PR will be merged
4. Your contribution will be credited!

### Commit Message Guidelines

Use clear, descriptive commit messages:

```
Good:
- "Add terrain generation validation tests"
- "Fix collision detection for diagonal movement"
- "Improve performance of spatial queries by 50%"

Bad:
- "fix bug"
- "update"
- "changes"
```

Format:
```
<type>: <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `test`: Add/update tests
- `perf`: Performance improvement
- `refactor`: Code restructuring
- `style`: Formatting changes

---

## Reporting Bugs

### Before Reporting

1. **Search existing issues** to avoid duplicates
2. **Test with latest version** to see if bug is fixed
3. **Gather information** about the bug

### Bug Report Template

Create a new issue with:

```markdown
**Description:**
Clear description of the bug

**Steps to Reproduce:**
1. Step 1
2. Step 2
3. Step 3

**Expected Behavior:**
What should happen

**Actual Behavior:**
What actually happens

**Environment:**
- OS: [e.g., Ubuntu 22.04]
- Go Version: [e.g., 1.24.7]
- Commit/Version: [e.g., abc1234]

**Additional Context:**
- Error messages
- Screenshots
- Logs
```

---

## Suggesting Features

### Feature Request Template

```markdown
**Feature Description:**
What feature would you like?

**Use Case:**
Why is this feature needed?

**Proposed Implementation:**
How might this work?

**Alternatives Considered:**
Other approaches you've thought about

**Additional Context:**
Examples, mockups, etc.
```

### Feature Criteria

Good features:
- Align with project goals
- Maintain performance targets
- Don't break existing functionality
- Have clear use cases
- Are feasible to implement

---

## Project Structure

See [Architecture](ARCHITECTURE.md) for detailed architectural decisions and [Technical Specification](TECHNICAL_SPEC.md) for complete system architecture.

**Key directories:**
- `cmd/` - Executable applications (client, server, test tools)
- `pkg/` - Reusable library packages (engine, procgen, rendering, audio, network)
- `docs/` - Project documentation
- `examples/` - Standalone demonstration programs

**Package Guidelines:**
- Lower layers don't depend on upper layers
- Use interfaces to break circular dependencies
- Keep packages loosely coupled
- See [Development Guide](DEVELOPMENT.md) for detailed package organization

---

## Performance Guidelines

**Optimization Priorities:**
1. Correctness first (make it work)
2. Clarity second (make it clear)
3. Performance third (make it fast)

**Performance Targets:** 60+ FPS with 2000 entities, <500MB client memory, <2s generation time

**Before optimizing:** Profile using `go test -cpuprofile` and `go test -memprofile`

**For detailed profiling instructions, benchmarking, and optimization techniques, see [Development Guide](DEVELOPMENT.md).**

---

## Getting Help

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **Pull Requests**: Code contributions and reviews
- **Discussions**: General questions and ideas

### Additional Resources

- [Development Guide](DEVELOPMENT.md) - Complete development environment setup and workflow
- [API Reference](API_REFERENCE.md) - API documentation with code examples
- [Architecture](ARCHITECTURE.md) - Architectural decisions and patterns
- [Technical Specification](TECHNICAL_SPEC.md) - Complete technical details
- [Roadmap](ROADMAP.md) - Development phases and progress

---

## License

By contributing to Venture, you agree that your contributions will be licensed under the same license as the project (see [LICENSE](../LICENSE) file).

---

## Thank You!

Thank you for contributing to Venture! Every contribution, no matter how small, helps make the project better. We appreciate your time and effort! 🎮✨

---

**Questions?** Open an issue or discussion on GitHub!
