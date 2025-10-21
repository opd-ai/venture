# Development Guide

## Getting Started

### Prerequisites

- **Go:** Version 1.21 or later
- **Operating System:** Windows, macOS, or Linux
- **Platform-specific dependencies:**
  - **Linux:** X11 development libraries
    ```bash
    apt-get install libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev \
                    libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev \
                    pkg-config libx11-dev
    ```
  - **macOS:** Xcode command line tools
    ```bash
    xcode-select --install
    ```
  - **Windows:** No additional dependencies required

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/opd-ai/venture.git
cd venture

# Download dependencies
go mod download

# Verify setup by running tests
go test -tags test ./pkg/...

# Build the project
go build ./cmd/client
go build ./cmd/server
```

## Project Structure

The project follows Go best practices with a clear separation between commands and packages:

```
venture/
├── cmd/                    # Command-line applications
│   ├── client/            # Game client
│   └── server/            # Game server
├── pkg/                    # Reusable packages
│   ├── engine/            # ECS framework and game loop
│   ├── procgen/           # Procedural generation
│   ├── rendering/         # Graphics generation
│   ├── audio/             # Audio synthesis
│   ├── network/           # Multiplayer networking
│   ├── combat/            # Combat mechanics
│   └── world/             # World state
├── docs/                   # Documentation
└── go.mod                  # Go module definition
```

## Development Workflow

### 1. Building

```bash
# Build client
go build -o venture-client ./cmd/client

# Build server
go build -o venture-server ./cmd/server

# Build both
make build  # (if Makefile is created)

# Build with optimizations (release)
go build -ldflags="-s -w" ./cmd/client
go build -ldflags="-s -w" ./cmd/server
```

### 2. Testing

```bash
# Run all package tests (excludes Ebiten initialization)
go test -tags test ./pkg/...

# Run tests with coverage
go test -tags test -cover ./pkg/...

# Run tests with race detection
go test -tags test -race ./pkg/...

# Run specific package tests
go test -tags test ./pkg/engine
go test -tags test ./pkg/procgen

# Generate coverage report
go test -tags test -coverprofile=coverage.out ./pkg/...
go tool cover -html=coverage.out
```

**Note:** Tests use the `-tags test` flag to exclude Ebiten initialization which requires a display. This allows running tests in CI/headless environments.

### 3. Running

```bash
# Run client in single-player mode
./venture-client -width 1024 -height 768 -seed 12345

# Run server
./venture-server -port 8080 -max-players 4

# Run client connecting to server (when implemented)
./venture-client -server localhost:8080
```

### 4. Code Quality

```bash
# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Vet code
go vet ./...

# Check for common mistakes
staticcheck ./...
```

### 5. Profiling

```bash
# CPU profiling
go test -tags test -cpuprofile=cpu.prof -bench=. ./pkg/engine
go tool pprof cpu.prof

# Memory profiling
go test -tags test -memprofile=mem.prof -bench=. ./pkg/engine
go tool pprof mem.prof

# Profile running application
go build -o venture-client ./cmd/client
./venture-client &
PID=$!
go tool pprof http://localhost:6060/debug/pprof/profile
kill $PID
```

## Package Development Guidelines

### Creating a New Package

1. Create package directory under `pkg/`
2. Add `doc.go` with package documentation
3. Define public interfaces in `interfaces.go` (if applicable)
4. Implement core functionality
5. Add comprehensive tests
6. Add examples in `example_test.go`

Example package structure:

```
pkg/newpkg/
├── doc.go              # Package documentation
├── interfaces.go       # Public interfaces
├── implementation.go   # Core implementation
├── implementation_test.go  # Unit tests
└── example_test.go     # Example usage
```

### Code Standards

1. **Documentation:**
   - Every exported function, type, and constant must have a doc comment
   - Package must have a `doc.go` file
   - Use complete sentences starting with the element name

2. **Testing:**
   - Target 80%+ code coverage
   - Test edge cases and error conditions
   - Use table-driven tests for multiple scenarios
   - Benchmark performance-critical code

3. **Error Handling:**
   - Return errors, don't panic (except for programmer errors)
   - Wrap errors with context using `fmt.Errorf`
   - Check all errors

4. **Concurrency:**
   - Use goroutines sparingly and document their lifecycle
   - Protect shared state with mutexes or channels
   - Test with `-race` flag

### Adding a System to the ECS

1. Implement the `System` interface:
   ```go
   type MySystem struct {
       // System state
   }
   
   func (s *MySystem) Update(entities []*Entity, deltaTime float64) {
       // Filter entities with required components
       for _, entity := range entities {
           if !entity.HasComponent("required") {
               continue
           }
           // Process entity
       }
   }
   ```

2. Register the system with the world:
   ```go
   world := engine.NewWorld()
   world.AddSystem(&MySystem{})
   ```

### Adding a Generator

1. Implement the `Generator` interface:
   ```go
   type MyGenerator struct{}
   
   func (g *MyGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
       // Use seed for deterministic generation
       rng := rand.New(rand.NewSource(seed))
       // Generate content
       return result, nil
   }
   
   func (g *MyGenerator) Validate(result interface{}) error {
       // Validate generated content
       return nil
   }
   ```

2. Test for determinism:
   ```go
   func TestMyGeneratorDeterminism(t *testing.T) {
       gen := &MyGenerator{}
       seed := int64(12345)
       params := procgen.GenerationParams{}
       
       result1, _ := gen.Generate(seed, params)
       result2, _ := gen.Generate(seed, params)
       
       // Verify results are identical
   }
   ```

## Debugging

### Common Issues

**"undefined: engine.NewGame" when testing:**
- Use `-tags test` flag: `go test -tags test ./...`
- The `NewGame` function requires Ebiten and is excluded from tests

**"DISPLAY environment variable missing":**
- Tests requiring display are automatically skipped in headless environments
- Build tags prevent Ebiten initialization during tests

**Build fails on Linux:**
- Install required X11 development libraries (see Prerequisites)

### Debugging Tools

```bash
# Print variables during test
go test -tags test -v ./pkg/engine

# Run with delve debugger
dlv test -- -tags test ./pkg/engine

# Trace execution
go test -tags test -trace trace.out ./pkg/engine
go tool trace trace.out
```

## Contributing

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make changes following code standards**

3. **Test thoroughly:**
   ```bash
   go test -tags test ./...
   go test -tags test -race ./...
   go vet ./...
   ```

4. **Commit with descriptive messages:**
   ```bash
   git commit -m "Add terrain generation system"
   ```

5. **Push and create pull request**

## Performance Optimization

### Profiling Checklist

- [ ] Profile before optimizing
- [ ] Focus on hot paths (>10% CPU time)
- [ ] Reduce allocations in tight loops
- [ ] Use object pooling for frequently allocated objects
- [ ] Consider sync.Pool for temporary objects
- [ ] Batch operations where possible
- [ ] Use spatial partitioning for entity queries

### Performance Targets

- **Frame Rate:** 60 FPS minimum
- **Memory:** <500MB client, <1GB server
- **Generation Time:** <2 seconds for world areas
- **Network:** <100KB/s per player

## Release Process

1. **Version bump:** Update version in code
2. **Run full test suite:** `go test -tags test ./...`
3. **Build release binaries:**
   ```bash
   GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o venture-client-linux ./cmd/client
   GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o venture-client-windows.exe ./cmd/client
   GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o venture-client-macos ./cmd/client
   ```
4. **Create release notes**
5. **Tag release:** `git tag -a v0.1.0 -m "Release v0.1.0"`
6. **Push tag:** `git push origin v0.1.0`

## Resources

- **Ebiten Documentation:** https://ebiten.org/
- **Go Documentation:** https://golang.org/doc/
- **ECS Pattern:** https://en.wikipedia.org/wiki/Entity_component_system
- **Procedural Generation:** https://en.wikipedia.org/wiki/Procedural_generation
