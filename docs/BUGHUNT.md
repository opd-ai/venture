**OBJECTIVE:** Perform exhaustive analysis of the entire Venture Go codebase to identify all bugs, anti-patterns, and violations of project-specific best practices.

**EXECUTION MODE:** Report generation only - no automatic code changes.

**CRITICAL REQUIREMENT:** Analyze EVERY Go file in the workspace. Do not sample or skip packages. Use systematic search tools (grep_search, file_search, semantic_search) to examine the complete codebase including all packages, examples, tests, and subdirectories.

**ANALYSIS SCOPE:**

1. **ECS Architecture Violations**
   - Logic/methods in components beyond Type()
   - Systems modifying components directly vs returning new state
   - Missing system registration in World
   - Debug: `grep -r "func.*Component.*\)" --include="*.go" pkg/ | grep -v "Type()"`, verify all systems in cmd/client/main.go, examine every component file

2. **Non-Deterministic Generation (CRITICAL)**
   - time.Now(), time.Since() in generators or game logic
   - Global rand.Intn/Float64 instead of seeded rand.New()
   - OS-dependent operations (file iteration order, map iteration)
   - Debug: `grep -rn "time\.Now\|time\.Since\|rand\.[IF]" pkg/`, scan ALL procgen/ subdirectories, run identical seeds twice and diff outputs

3. **Race Conditions & Concurrency**
   - Unsynchronized map/slice access across goroutines
   - Missing mutex locks on shared state
   - Goroutine leaks (no cancellation context)
   - Channel deadlocks (unbuffered sends without receivers)
   - Debug: `go test -race ./...` on ALL packages, `grep -rn "go func\|make(chan" pkg/`, check runtime.NumGoroutine() growth

4. **Memory Issues**
   - Sprite cache not freeing old entries
   - Deferred Close() in loops causing descriptor leaks
   - Slice append in hot paths causing excessive GC
   - Unclosed file handles, network connections
   - Debug: `go test -memprofile mem.out ./...`, examine EVERY defer statement, profile all long-running systems

5. **Network Anti-Patterns**
   - Type assertions to *net.UDPAddr/*net.TCPConn (use net.Addr/net.Conn interfaces)
   - Missing timeout on Dial/Listen
   - Unbounded read buffers (DoS risk)
   - Debug: `grep -rn "\.(\*net\.(UDP|TCP)" pkg/network/`, search ALL network code for type switches, test timeout behavior

6. **Error Handling**
   - Ignored errors: `_, err := ...; // TODO` or `_ = ...`
   - Panic in library code instead of returning errors
   - Missing error wrapping (use fmt.Errorf with %w)
   - Silent failures without logging
   - Debug: `errcheck ./...`, `grep -rn "panic\(" pkg/`, scan EVERY error return for proper handling

7. **Performance (60 FPS Target)**
   - O(n²) collision checks without spatial partitioning
   - String concatenation in loops (use strings.Builder)
   - JSON marshal/unmarshal in hot paths
   - Allocating slices without capacity hints
   - Reflection in Update() loops
   - Debug: `go test -bench . -cpuprofile cpu.out ./...`, analyze ALL system Update() methods, profile rendering pipeline end-to-end

8. **Logging Issues**
   - String formatting instead of logrus.Fields
   - Logging in tight loops (>1000 logs/sec)
   - Missing context fields (entityID, seed, genre, system_name)
   - Debug: `grep -rn 'log\.\(Info\|Debug\|Warn\|Error\)f\(' pkg/`, examine ALL log statements across codebase

9. **Testing Gaps**
   - Packages below 40% coverage threshold (below 30% for X11/Wayland/Ebiten-dependent packages)
   - Missing table-driven tests for generators
   - No benchmarks for performance-critical code
   - Test files using non-deterministic data
   - Debug: `go test -cover ./... | grep -E "coverage: [0-6]?[0-9]\.[0-9]%"`, review ALL *_test.go files

10. **Additional Code Quality Issues**
    - Exported symbols without godoc comments
    - Improper interface usage (accepting concrete types)
    - Magic numbers without named constants
    - Dead code, unused variables/imports
    - Debug: `go vet ./...`, `golangci-lint run`, review ALL exported declarations

**COMPREHENSIVE SEARCH STRATEGY:**
1. Use `file_search` to enumerate ALL .go files in workspace
2. Use `grep_search` with broad patterns across entire codebase
3. Read complete files for context, not just snippets
4. Check ALL subdirectories: pkg/*, cmd/*, examples/*, mods/*
5. Examine test files (*_test.go) for coverage and quality
6. Review EVERY system in pkg/engine/
7. Review EVERY generator in pkg/procgen/
8. Scan ALL network code in pkg/network/
9. Verify ALL rendering systems in pkg/rendering/

**OUTPUT FORMAT:**
```
## Critical (Fix Now)
- [file.go:L123](file.go#L123) Issue: Non-deterministic rand.Intn() in generator
  Debug: `grep -n "rand\.[IF]" file.go`, compare outputs with seed=42
  Fix: Use rng := rand.New(rand.NewSource(seed))

## Architecture Violations  
- [system.go:L456](system.go#L456) Component has Move() method (violates pure data)
  Debug: Check component interface - should only have Type()
  Fix: Move logic to MovementSystem.Update()

## Performance
- [collision.go:L789](collision.go#L789) O(n²) collision loop without partitioning
  Debug: `go test -bench BenchmarkCollision -cpuprofile cpu.out`
  Fix: Implement quadtree spatial partitioning

## Summary
Critical: X | Architecture: X | Performance: X | Other: X
Packages analyzed: X/97 total
Files examined: X total Go files
Top priority: [Most severe issue]
```

**DEBUGGING COMMANDS:**
```bash
# Comprehensive race detection
go test -race -count=1 ./...

# Full coverage analysis
go test -cover ./... > coverage.txt
awk '/coverage:/ {if ($3 < "65.0%") print $2, $3}' coverage.txt

# Static analysis on entire codebase
go vet ./...
golangci-lint run --enable-all --disable wsl,nlreturn,exhaustruct

# Determinism validation across all generators
for seed in 12345 67890 99999; do
  go run examples/genretest -seed $seed > "out_${seed}_1.txt"
  go run examples/genretest -seed $seed > "out_${seed}_2.txt"  
  diff "out_${seed}_1.txt" "out_${seed}_2.txt" || echo "FAIL: seed $seed non-deterministic"
done

# Performance profiling all packages
for pkg in $(go list ./pkg/...); do
  go test -bench . -cpuprofile "cpu_${pkg##*/}.out" -memprofile "mem_${pkg##*/}.out" "$pkg"
done

# Find all panics
grep -rn "panic\(" pkg/ cmd/ examples/

# Find all ignored errors
grep -rn "_ =" pkg/ | grep -i err
```

**CONSTRAINTS:**
- Examine ENTIRE codebase systematically - do not skip packages
- Report top 30 most severe issues (increase from 25)
- Include file counts and package coverage in summary
- Provide specific line numbers for ALL findings
- Use file link format: [file.go:L123](file.go#L123)

**SUCCESS CRITERIA:**
- Confirmation that ALL 97 packages were analyzed
- Total Go file count reported in summary
- All non-deterministic code identified with reproduction commands
- ECS violations flagged across ALL components/systems
- Race conditions verified with `go test -race ./...` (not selective)
- Performance issues include profiler data or measured FPS impact
- Each critical bug includes: file location, debug command, expected vs actual behavior, specific fix
- Coverage gaps identified for EVERY package below 40% (30% for X11/Wayland/Ebiten-dependent packages)
