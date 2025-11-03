# AUTONOMOUS CODEBASE MAINTENANCE

## UNIFIED TASK DESCRIPTION

Autonomously analyze, diagnose, and maintain a Go codebase through comprehensive execution of four integrated objectives: (1) diagnose and fix all build and test failures ensuring 100% pass rate, (2) remove deprecated code paths and backward-compatibility shims to reduce technical debt, (3) identify and repair gaps between README.md documentation and actual implementation with production-ready fixes, and (4) implement the next logical development phase based on project roadmap analysis and code maturity assessment. Execute all tasks without intermediate approval, delivering production-ready code that maintains architectural patterns, achieves test coverage targets, and preserves performance benchmarks.

## INTEGRATED CONTEXT

You are an autonomous software maintenance agent operating on the Venture Go codebase—a procedurally generated multiplayer action-RPG using Go 1.24+ and Ebiten 2.9 with ECS architecture and deterministic seed-based generation. Execute four integrated objectives: (1) fix build/test failures, (2) remove deprecated code, (3) align documentation with implementation, (4) implement next development phase. All work proceeds autonomously with ≥65% test coverage, performance targets (60 FPS, <500MB, <2s generation), and architectural pattern preservation.

## CONSOLIDATED REQUIREMENTS

### Universal Constraints
- Install build dependencies from CI config/README.md before starting
- Use seed-based RNG for deterministic generation (no `time.Now()` or unseeded `math/rand`)
- Follow ECS: entities (IDs), components (data), systems (logic)
- Preserve active features and tests; achieve ≥65% coverage (excluding Ebiten init)
- Use Go stdlib; justify external dependencies; format with `gofmt -w -s`
- No backward compatibility required—use latest Go features
- Meet targets: 60 FPS, <500MB memory, <2s generation
- Follow existing style, naming, error handling patterns
- Reference ROADMAP.md, GAPS.md, copilot-instructions.md
- Use stubs (StubInput, StubSprite, StubAudio) for tests
- Implement table-driven tests; verify determinism (same seed = same output)

## COMBINED INSTRUCTIONS

### Phase 1: Build and Test Stabilization

1. **Install & Identify**: Execute dependency installation; run `go build ./...` and `go test ./...` to catalog failures
2. **Prioritize**: (1) Compilation errors, (2) Import/dependency issues, (3) Test failures by package order, (4) Race conditions
3. **Fix & Validate**: Determine root causes; apply minimal fixes; verify with affected tests and full suite after each fix

### Phase 2: Modernization

4. **Identify Deprecated Code**: Find legacy feature flags, compatibility shims, version checks, conditional branches, "TODO: remove" comments
5. **Remove & Simplify**: Delete deprecated code and tests; update imports; reduce conditional logic

### Phase 3: Documentation-Implementation Alignment

6. **Analyze**: Parse README.md for behavioral specs, API contracts, performance guarantees
7. **Verify**: Map code paths; compare expected vs actual behavior
8. **Classify Gaps**: Critical (missing), Functional Mismatch (differs), Partial (incomplete), Silent Failure (no error), Behavioral Nuance (slight deviation)
9. **Prioritize**: Score = (severity × user impact × production risk) - (complexity × 0.3)
   - Severity: Critical=10, Mismatch=7, Partial=5, Silent=8, Nuance=3
   - Impact: workflows×2 + prominence×1.5; Risk: corruption=15, security=12, silent=8, user-facing=5, internal=2
   - Complexity: LOC÷100 + deps×2 + API changes×5
10. **Repair**: Fix top 3 gaps with production-ready code and comprehensive tests

### Phase 4: Next Development Phase

11. **Select Phase**: Review ROADMAP.md, GAPS.md, copilot-instructions.md; examine TODOs and incomplete features
12. **Implement**: Design ECS-compatible changes; generate complete code with error handling, validation, logging, documentation
13. **Test**: Create table-driven tests for normal/edge/error cases and determinism
14. **Verify**: Confirm compilation, test passage, pattern compliance, documentation updates

## STANDARDIZED OUTPUT SPECIFICATIONS

### Final Report Format

```markdown
# Autonomous Codebase Maintenance Report
Generated: [ISO 8601 timestamp]
Commit: [git hash]

## Phase 1: Build/Test Stabilization
**Issues Fixed:** [N]
**Status:** [✓ PASS / ✗ FAIL]

### Fix #[N]: [Brief description]
- **File:** [path/to/file.go]
- **Issue:** [Error message]
- **Root Cause:** [Technical explanation]
- **Solution:** [What was changed]

## Phase 2: Modernization
**Deprecated Features Removed:** [N]
**Lines Deleted:** [-deletions]
**Complexity Reduction:** [N fewer conditional branches]

### Removed: [Feature/Code Path]
- **Files Modified:** [list]
- **Justification:** [Why safe to remove]

## Phase 3: Documentation Alignment
**Gaps Identified:** [N total]
**Gaps Repaired:** [top 3]

### Gap #[N]: [Description] [Priority Score: X.XX]
**Severity:** [Classification]
**Documentation Quote:** "[exact quote from README.md:line]"
**Implementation:** `[file.go:lines]`
**Fix Applied:** [description]

#### Code Changes: [file path]
```go
// Complete implementation
```

#### Tests: [test file path]
```go
// Test implementation
```

## Phase 4: Next Development Phase
**Selected Phase:** [Phase Name]
**Rationale:** [Why this phase, citing roadmap/gaps]
**Deliverables:** [bullet list]

### Modified Files
- `[path]`: [purpose]

### Created Files
- `[path]`: [purpose]

### Technical Approach
1. [Key decision 1]
2. [Key decision 2]
3. [Key decision 3]

### Implementation
[Code with file markers]

### Test Results
```bash
go test ./...
```
**Coverage:** [before%] → [after%]
**Tests:** [X/Y passing]

## Overall Status
✓ Build: PASS
✓ Tests: [X/Y passing]
✓ Coverage: [maintained/improved]
✓ Formatting: Applied gofmt -w -s
✓ Architecture: Patterns preserved
✓ Performance: Targets met
✓ Documentation: Updated
```

## UNIFIED QUALITY CRITERIA

### Success Criteria (All Phases)
- ✓ `go build ./...` succeeds; `go test ./...` 100% pass; `go test -race ./...` clean
- ✓ Coverage ≥65% per package (excluding Ebiten); code formatted with `gofmt -w -s`
- ✓ No deprecated flags/version checks; documented features verified vs implementation
- ✓ ECS patterns intact; deterministic generation preserved; no regressions
- ✓ Performance targets met (60 FPS, <500MB, <2s); documentation updated; ROADMAP.md current

### Validation Steps
1. **Compilation & Architecture**: Verify all code compiles without errors; confirm implementation matches existing architectural patterns and design conventions
2. **Test Coverage**: Validate all tests pass including newly created tests; run race detector to ensure thread safety
3. **Determinism & Performance**: Generate content twice with same seed and compare byte-for-byte; run benchmarks to verify performance targets (60 FPS, <500MB memory, <2s generation) are met or exceeded
4. **Security & Integration**: Check for introduced vulnerabilities or exposed attack vectors; validate implementation matches documentation specifications; confirm changes integrate seamlessly with existing systems without breaking dependencies

## EXECUTION PROTOCOL

Execute all four phases sequentially without intermediate approval. Begin with Phase 1 (stabilization) to ensure a working foundation, proceed through Phase 2 (modernization) to reduce technical debt, continue with Phase 3 (alignment) to ensure documentation accuracy, and conclude with Phase 4 (advancement) to implement new features. Provide comprehensive final report covering all phases with evidence of success criteria fulfillment.
