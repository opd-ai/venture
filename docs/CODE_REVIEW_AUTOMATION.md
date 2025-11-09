# Automated Code Review Tools

This directory contains tools for automated package selection and code review auditing based on the methodology defined in `docs/CODE_REVIEW_PLAN.md`.

## Tools

### 1. Package Selector (`cmd/auditselect`)

Analyzes package dependencies and selects the next package for audit based on dependency depth.

**Usage:**
```bash
# Basic usage - prints selected package path
go run cmd/auditselect/main.go

# Verbose mode - shows dependency analysis
go run cmd/auditselect/main.go -v

# Custom package directory
go run cmd/auditselect/main.go -pkg custom/pkg/path
```

**Selection Algorithm:**
1. Scans `pkg/` directory recursively
2. Identifies all Go packages
3. Excludes packages with existing `AUDIT.md` files
4. Analyzes Go imports to count internal dependencies
5. Sorts by dependency depth (ascending)
6. Prioritizes by path: `engine` → `procgen/*` → `rendering/*` → others
7. Selects package with lowest dependency depth

**Example Output:**
```
Found 40 packages
34 packages need auditing
pkg/procgen/genre
Selected: pkg/procgen/genre (depth: 0 dependencies)
```

### 2. Code Review Script (`scripts/automated-code-review.sh`)

Performs comprehensive automated code review on a selected package.

**Usage:**
```bash
# Review a specific package
./scripts/automated-code-review.sh -p pkg/procgen/genre

# Verbose mode with detailed output
./scripts/automated-code-review.sh -p pkg/procgen/genre -v
```

**Review Phases:**

**Phase 1: Static Analysis**
- `go vet` - Identifies suspicious constructs
- `gofmt -l` - Checks code formatting
- `go build` - Verifies compilation

**Phase 2: Testing**
- `go test` - Runs all tests
- `go test -race` - Detects race conditions
- Coverage analysis (target ≥65%)

**Phase 3: Documentation**
- Checks for `doc.go` file
- Verifies godoc coverage for exported identifiers

**Phase 4: Pattern Compliance**
- Deterministic generation (no `time.Now()`, global `rand`)
- RNG isolation (uses seeded instances)
- Error handling (all errors checked and wrapped)

**Example Output:**
```
===================================================================
Automated Code Review for: pkg/procgen/genre
===================================================================

[INFO] Phase 1: Static Analysis & Structure Review
  Checking go_vet... PASS
  Checking go_fmt... PASS
  Checking build... PASS
[INFO] Phase 2: Testing & Coverage
  Checking tests... PASS
  Checking race... PASS
  Checking coverage... PASS (100.0%)
[INFO] Phase 3: Documentation & API Review
  Package documentation: PASS
  Checking godoc coverage... WARN
[INFO] Phase 4: Pattern Compliance Checks
  Checking deterministic generation... PASS
  Checking RNG isolation... PASS
  Checking error handling... PASS

Quality Gates:
  [✓] go_vet
  [✓] go_fmt
  [✓] build
  [✓] tests
  [✓] race
  [✓] coverage
  [✓] doc_go
  [!] godoc
  [✓] determinism
  [✓] rng_isolation
  [✓] error_handling

Issues Found:
  Critical: 0
  Major: 0
  Minor: 0
```

## Workflow

### Complete Audit Workflow

1. **Select Package:**
   ```bash
   PACKAGE=$(go run cmd/auditselect/main.go)
   echo "Selected: $PACKAGE"
   ```

2. **Run Automated Review:**
   ```bash
   ./scripts/automated-code-review.sh -p $PACKAGE -v
   ```

3. **Review Results:**
   - Check quality gate results
   - Note any failing gates or warnings
   - Review coverage percentage
   - Examine pattern compliance

4. **Generate AUDIT.md:**
   - Use review results to create comprehensive audit report
   - Include all sections from CODE_REVIEW_PLAN.md template
   - Document findings with file:line references
   - Provide code snippets for issues
   - Add recommendations

5. **Save Audit Report:**
   ```bash
   # Create AUDIT.md in the package directory
   # Example: pkg/procgen/genre/AUDIT.md
   ```

6. **Verify and Commit:**
   ```bash
   git add $PACKAGE/AUDIT.md
   git commit -m "docs: Add code review audit for $PACKAGE"
   ```

## AUDIT.md Format

Each `AUDIT.md` file follows this structure:

```markdown
# Code Review Audit: [package name]
**Date:** [ISO date]
**Reviewer:** GitHub Copilot
**Dependency Depth:** [N]

## Executive Summary
[Pass/Fail with brief assessment]

## Quality Gates
- [x] Build success
- [x] All tests pass
- [x] Race-free
- [x] Coverage ≥65%
[... all 18 gates from CODE_REVIEW_PLAN.md]

## Findings
### Critical (blocks merge)
[file:line - issue - fix]

### Major (should fix)
[file:line - issue - fix]

### Minor (nice-to-have)
[file:line - issue - fix]

## Recommendations
[Actionable next steps]
```

## Package Status

### Audited Packages (6)
- ✅ `pkg/audio` - Previously audited
- ✅ `pkg/combat` - Previously audited
- ✅ `pkg/logging` - Previously audited
- ✅ `pkg/mobile` - Previously audited
- ✅ `pkg/world` - Previously audited
- ✅ `pkg/procgen/genre` - **Newly audited (100% coverage, all gates pass)**

### Remaining Packages (34)
Run `go run cmd/auditselect/main.go -v` to see the next package to audit.

## Quality Gate Checklist

From `docs/CODE_REVIEW_PLAN.md`, all packages must meet these 18 quality gates:

1. ✓ Build Success
2. ✓ Test Pass
3. ✓ Race Freedom
4. ✓ Code Coverage ≥65%
5. ✓ Static Analysis Clean
6. ✓ Code Formatting
7. ✓ Documentation Complete
8. ✓ Package Docs Present
9. ✓ No Circular Dependencies
10. ✓ Performance Targets Met
11. ✓ Determinism Verified
12. ✓ ECS Pattern Compliance
13. ✓ Error Handling
14. ✓ Input Validation
15. ✓ Resource Cleanup
16. ✓ API Documentation
17. ✓ Multiplayer Sync
18. ✓ Genre Compatibility

## Troubleshooting

### Common Issues

**Issue:** `go vet` fails
- **Fix:** Run `go vet ./pkg/[package]` manually and address reported issues

**Issue:** Tests fail
- **Fix:** Run `go test -v ./pkg/[package]` to see detailed failure output

**Issue:** Race conditions detected
- **Fix:** Review concurrent code, add proper synchronization (mutexes, channels)

**Issue:** Coverage below 65%
- **Note:** Functions requiring Ebiten runtime initialization (e.g., `ebiten.NewImage()`) are excluded from coverage requirements
- **Fix:** Add tests for business logic, use stub implementations for Ebiten dependencies

**Issue:** Non-deterministic generation
- **Fix:** Replace `time.Now()` and global `rand` with seeded RNG: `rand.New(rand.NewSource(seed))`

## References

- **CODE_REVIEW_PLAN.md** - Comprehensive review methodology
- **ARCHITECTURE.md** - System architecture and patterns
- **TESTING.md** - Testing guidelines and coverage targets
- **CONTRIBUTING.md** - Code style and quality standards

## Future Enhancements

- [ ] Automated AUDIT.md generation from review results
- [ ] Integration with CI/CD pipeline
- [ ] HTML report generation
- [ ] Trend analysis (coverage over time, quality improvements)
- [ ] Package dependency visualization
