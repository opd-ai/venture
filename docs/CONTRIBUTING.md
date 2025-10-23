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

### Development Setup

1. Fork the repository on GitHub
2. Clone your fork and set up development environment:
   ```bash
   git clone https://github.com/YOUR_USERNAME/venture.git
   cd venture
   git remote add upstream https://github.com/opd-ai/venture.git
   ```

**For complete development environment setup, dependencies, build instructions, testing workflows, profiling, and debugging, see [Development Guide](DEVELOPMENT.md).**

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

Write table-driven tests with good coverage. Always test that procedural generation is deterministic (same seed = same output). Add benchmarks for performance-critical code.

**For detailed testing examples, patterns, and best practices, see [Development Guide](DEVELOPMENT.md).**

---

## Code Style

Follow standard Go conventions: use `go fmt`, pass `go vet`, check all errors, document exported items.

**Key requirements:**
- Deterministic generation (same seed = same output)
- ECS architecture (separate entities, components, systems)
- 80% test coverage minimum
- No external assets (100% procedural)

**For detailed code style guidelines, documentation standards, and examples, see [Development Guide](DEVELOPMENT.md).**

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

See [Development Guide](DEVELOPMENT.md) for detailed project structure, package organization, and architectural guidelines.

---

## Performance Guidelines

When contributing performance improvements, always profile first and focus on correctness before optimization. See [Development Guide](DEVELOPMENT.md) for detailed profiling instructions and performance targets.

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
