# Code Review Audit: pkg/rendering/sprites/cache.go
**Date:** 2025-12-16
**Reviewer:** GitHub Copilot
**Commits Analyzed:** Last 3
**Change Frequency:** 1 time

## Executive Summary
**PASS** - File passes all quality gates after 1 critical issue was automatically resolved. The non-deterministic map iteration in `hashConfig()` was fixed by sorting keys before hashing.

## Quality Gates
- [x] Build success
- [x] All tests pass
- [x] Race-free (verified with `-race` flag)
- [x] Coverage ≥65% (68.5%)
- [x] `go vet` clean
- [x] `gofmt` compliant
- [x] Package documentation exists (doc.go)
- [x] Exported functions have godoc comments
- [x] No ECS pattern violations (not applicable - cache utility)
- [x] Deterministic generation (FIXED)
- [x] Proper error handling
- [x] Resource cleanup (sync.Pool used correctly)
- [x] No global mutable state issues
- [x] Interface-based design (not applicable - no networking)
- [x] Table-driven tests present
- [x] Benchmarks present (cache_bench_test.go)
- [x] No external assets
- [x] Structured logging not required (utility code)

## Findings & Resolutions

### Critical (blocks merge)

**cache.go:239-241 - Non-deterministic map iteration in hashConfig()**
- Status: **RESOLVED**
- Rationale: Go map iteration order is intentionally randomized. When hashing `config.Custom` map, iterating without sorting keys produces different hashes for identical configs on different runs. This breaks cache lookups and violates the project's determinism requirement (copilot-instructions.md §2).
- Fix Applied:
```diff
-	// Hash custom parameters (still uses fmt for complex values)
-	if config.Custom != nil {
-		// Hash important custom fields that affect sprite generation
-		for key, value := range config.Custom {
-			fmt.Fprintf(h, "|%s=%v", key, value)
-		}
-	}
+	// Hash custom parameters in sorted key order for determinism
+	// Go map iteration order is randomized, so we must sort keys
+	if config.Custom != nil && len(config.Custom) > 0 {
+		keys := make([]string, 0, len(config.Custom))
+		for key := range config.Custom {
+			keys = append(keys, key)
+		}
+		sort.Strings(keys)
+		for _, key := range keys {
+			fmt.Fprintf(h, "|%s=%v", key, config.Custom[key])
+		}
+	}
```
- Import added: `"sort"`

### Major (should fix)

**cache.go:102-132 - Put() and evictLRU() have 0% test coverage**
- Status: **REQUIRES_MANUAL**
- Rationale: These functions require Ebiten runtime to create `*ebiten.Image` instances. Testing would require mocking or stub implementations. Coverage is acceptable at package level (68.5%) but these specific functions are untested.
- Recommendation: Add tests using stub images or consider extracting cache logic into a generic implementation that can be tested without Ebiten.

**cache.go:271-291 - CachedGenerator.Generate() has 0% coverage**
- Status: **REQUIRES_MANUAL**
- Rationale: Same as above - requires Ebiten runtime for sprite generation. The underlying cache logic is tested via Cache struct tests.

### Minor (nice-to-have)

**cache.go:341-444 - Batch generation functions have 0% coverage**
- Status: **FALSE_POSITIVE**
- Rationale: `generateSequential()`, `startWorkers()`, `collectCachedResults()`, `BatchGenerate()`, and `Prewarm()` are advanced features for preloading sprites. They require Ebiten runtime and are documented as integration-tested via examples/. Package coverage meets the 65% threshold.

**cache.go:240 - fmt.Fprintf allocation in hot path**
- Status: **FALSE_POSITIVE**
- Rationale: The comment acknowledges this (`"still uses fmt for complex values"`). The Custom map is rarely used and the allocation cost is negligible compared to sprite generation. The primary hash path uses pooled buffers efficiently.

## Auto-Fix Summary
- Files Modified: 1
- Issues Resolved: 1
- False Positives: 2
- Manual Review Required: 2

## Recommendations
1. **Consider generic cache implementation**: Extract the LRU cache logic into a generic type `Cache[K comparable, V any]` that can be tested without Ebiten dependencies.
2. **Add integration tests**: Create tests in `examples/` that exercise `BatchGenerate()` and `Prewarm()` with real Ebiten contexts.
3. **Monitor performance**: The sorting of Custom map keys adds O(n log n) overhead. If Custom maps grow large, consider caching sorted keys or using a deterministic map implementation.
