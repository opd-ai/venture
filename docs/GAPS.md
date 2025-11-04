# TASK DESCRIPTION:
Conduct a comprehensive audit of the Go codebase following the methodology in AUDIT_ME.md, identify all issues and discrepancies in README.md and ROADMAP*.md files, then automatically implement repairs for high-priority issues with production-ready code.

## CONTEXT:
You are an autonomous software audit and repair agent. Your task is three-fold:
1. **Audit the codebase**: Follow the comprehensive audit methodology specified in `docs/AUDIT_ME.md` to identify bugs, performance issues, missing features, anti-patterns, and API design flaws. Create `AUDIT.md` in the repository root with your findings.
2. **Fix audit findings**: Implement repairs for the highest-priority issues identified in the audit.
3. **Fix documentation discrepancies**: Identify and fix all discrepancies, inconsistencies, and errors in README.md, ROADMAP.md, and ROADMAP_V2.md (version numbers, status markers, duplicate content, stale information, etc.).

## INSTRUCTIONS:

### Phase 1: Conduct Comprehensive Audit (Following AUDIT_ME.md)

**Step 1.1: Review Audit Methodology**
- Read `docs/AUDIT_ME.md` completely to understand the audit approach and categories
- The audit covers 5 main categories:
  1. CORRECTNESS & RELIABILITY (race conditions, nil dereferences, error handling, logic errors, resource leaks)
  2. PERFORMANCE & EFFICIENCY (memory allocation, algorithmic complexity, concurrency bottlenecks, I/O inefficiencies)
  3. API DESIGN & USABILITY (safety, completeness, clarity, consistency, documentation)
  4. CODE QUALITY & MAINTAINABILITY (anti-patterns, code organization, naming, documentation, test coverage, duplication)
  5. COMPLETENESS & PRODUCTION READINESS (missing functionality, observability, configuration, error messages, graceful degradation)

**Step 1.2: Systematic Code Audit**
- Analyze the entire codebase following AUDIT_ME.md categories
- For each file in `pkg/`, `cmd/`, and `examples/`:
  - Check for race conditions (concurrent map access, unprotected shared state)
  - Identify nil pointer dereferences and missing nil checks
  - Find ignored errors and improper error handling
  - Spot logic errors and edge case failures
  - Detect resource leaks (unclosed files, goroutine leaks, memory leaks)
  - Assess performance issues (O(n²) where O(n) possible, excessive allocations)
  - Evaluate API design (can APIs be misused? Are invariants enforced?)
  - Check code quality (duplicated code, complex functions, poor naming, missing tests)
  - Verify completeness (missing observability, inadequate error messages, no graceful degradation)

**Step 1.3: Create AUDIT.md Report**
- Generate `AUDIT.md` in the repository root using the format specified in AUDIT_ME.md
- Include:
  - Executive summary with counts by severity (Critical, High, Medium, Low, Info)
  - Detailed findings for each issue with:
    - Severity rating
    - File location and line numbers
    - Description of the issue
    - Current code showing the problem
    - Impact assessment
    - Recommended fix with code example
    - Justification for the recommendation
  - Summary section with:
    - Critical/High/Medium/Low issue counts
    - Convention assessment
    - Overall health assessment
    - Quick wins (easy fixes with high impact)

### Phase 2: Fix Documentation Discrepancies

**Step 2.1: Analyze README.md**
- Check for:
  - Version number inconsistencies (compare with ROADMAP*.md)
  - Stale status information (e.g., "in development" vs "complete")
  - Broken links or references to non-existent files
  - Outdated feature descriptions
  - Incorrect commands or examples
  - Missing prerequisites or setup steps
  - Duplicate content

**Step 2.2: Analyze ROADMAP.md and ROADMAP_V2.md**
- Check for:
  - Version number mismatches between files
  - Inconsistent phase completion markers (marked complete but not implemented, or vice versa)
  - Duplicate sections or content
  - Stale timeline information
  - Conflicting status statements
  - Incomplete phase descriptions (future tense when should be past tense for completed work)
  - Missing completion dates

**Step 2.3: Fix All Documentation Issues**
- Update version numbers to be consistent across all files
- Remove duplicate content
- Update completion markers and status descriptions
- Fix stale timeline information
- Ensure past tense for completed features
- Add missing completion dates
- Document changes made in a summary file

### Phase 3: Prioritize and Fix Issues

**Step 3.1: Priority Calculation**
For each issue identified in the audit (AUDIT.md), calculate priority score:

**Severity Multipliers:**
- Critical (crashes, data loss, security vulnerabilities, severe race conditions) = 15
- High (reliability issues, significant performance problems, major API safety gaps, resource leaks) = 10
- Medium (moderate performance issues, API usability problems, maintainability concerns) = 5
- Low (minor inefficiencies, style inconsistencies, documentation gaps) = 2

**Impact Factors:**
- User impact: Number of affected code paths × 2
- Production risk: Data corruption (20), Security breach (18), Service outage (15), Silent failure (10), Performance degradation (7), User confusion (4)
- Blast radius: System-wide (5), Multiple packages (3), Single package (2), Single function (1)

**Complexity Penalty:**
- Lines of code to fix ÷ 50
- Cross-package dependencies × 3
- Breaking changes × 10

**Formula:**
`Priority Score = (Severity × User Impact × Production Risk × Blast Radius) - (Complexity Penalty × 0.2)`

**Step 3.2: Select Top Issues**
- Rank all issues by priority score (descending)
- Select the top 5 highest-priority issues for immediate repair
- Ensure at least one documentation fix is included if any exist

**Step 3.3: Implement Fixes**
For each selected issue:

**A. Code Analysis**
- Review existing code patterns, naming conventions, error handling styles
- Identify integration points and dependencies
- Document test coverage patterns

**B. Design Fix**
- Design minimal surgical changes that resolve the issue
- Ensure changes integrate with existing patterns
- Plan backward compatibility where needed
- Document all files requiring modification

**C. Implement Production-Ready Code**
- Generate complete, executable Go code that fixes the issue
- Include proper error handling matching existing patterns
- Add input validation and boundary checks
- Implement logging consistent with codebase
- Add inline documentation for complex logic
- Format code with `gofmt -w -s`

**D. Create Tests**
- Write unit tests covering normal operation
- Add integration tests for cross-package functionality
- Include edge case and error condition tests
- Add regression tests to prevent issue recurrence
- Ensure all tests pass

**E. Verify Fix**
- Compile code: `go build ./...`
- Run tests: `go test ./...`
- Run linter: `golangci-lint run` (if available)
- Verify no regressions in existing functionality
- Check that documentation updates are accurate

## OUTPUT REQUIREMENTS:

### 1. AUDIT.md (Repository Root)
Create `AUDIT.md` in the repository root following the format in AUDIT_ME.md:

```markdown
# Code Audit Report
Generated: [ISO 8601 timestamp]
Codebase Version: [commit hash]

## Executive Summary
- **Critical Issues**: [count]
- **High Priority**: [count]
- **Medium Priority**: [count]
- **Low Priority/Info**: [count]

## Findings by Category

### [CATEGORY]: [Issue Description]
**Severity**: [Critical | High | Medium | Low | Info]
**Location**: `path/to/file.go` (lines X-Y)

**Description**: [Clear explanation of the issue]

**Current Code**:
```go
// Code showing the problem
```

**Impact**: [What problems this causes]

**Recommendation**:
```go
// Suggested fix
```

**Justification**: [Why this fix improves the code]

---

## Summary
**Critical Issues**: [count] - [brief list if any]
**High Priority**: [count] - [brief list]
**Quick Wins**: [List easy-to-fix issues with high impact]
**Overall Health**: [Assessment paragraph]
```

### 2. DOCUMENTATION-FIXES.md (Created for tracking)
Create a summary of all documentation changes:

```markdown
# Documentation Discrepancy Fixes
Date: [ISO 8601 timestamp]

## Issues Fixed in README.md
1. [Description of fix] - Line X
2. [Description of fix] - Line Y

## Issues Fixed in ROADMAP.md
1. [Description of fix] - Line X
2. [Description of fix] - Line Y

## Issues Fixed in ROADMAP_V2.md
1. [Description of fix] - Line X
2. [Description of fix] - Line Y

## Summary
- Total documentation issues fixed: [count]
- Version consistency established: [version number]
- Duplicate content removed: [count] instances
- Completion markers updated: [count] phases
```

### 3. FIXES-IMPLEMENTED.md (Created for tracking)
Document all code fixes implemented:

```markdown
# Code Fixes Implemented
Date: [ISO 8601 timestamp]

## Fix #1: [Issue Description]
**Original Priority Score**: [score]
**Severity**: [Critical/High/Medium/Low]
**Files Modified**: [count]
**Lines Changed**: +[additions] -[deletions]

**Issue**: [Description]
**Solution**: [What was done]
**Testing**: [Tests added/updated]
**Verification**: [How it was verified]

---

[Repeat for each fix]

## Summary
- Total issues fixed: [count]
- Critical issues resolved: [count]
- High priority issues resolved: [count]
- Test coverage added: [percentage or count]
- All tests passing: [Yes/No]
```

## QUALITY CHECKS:
Before finalizing, verify:
1. ✓ AUDIT.md created in repository root following AUDIT_ME.md format
2. ✓ All audit categories from AUDIT_ME.md have been assessed
3. ✓ Each audit finding includes file location, code evidence, and recommended fix
4. ✓ Priority scores calculated correctly using the formula
5. ✓ Top 5 issues selected for fixing based on priority
6. ✓ All code fixes compile without errors (`go build ./...`)
7. ✓ All tests pass (`go test ./...`)
8. ✓ README.md checked for discrepancies and fixed if found
9. ✓ ROADMAP.md checked for discrepancies and fixed if found
10. ✓ ROADMAP_V2.md checked for discrepancies and fixed if found
11. ✓ Version numbers consistent across all documentation files
12. ✓ Duplicate content removed from documentation
13. ✓ Completion markers accurate and in past tense for completed work
14. ✓ DOCUMENTATION-FIXES.md created summarizing doc changes
15. ✓ FIXES-IMPLEMENTED.md created documenting code repairs

## EXECUTION ORDER:
1. **Phase 1**: Conduct audit, create AUDIT.md (follow AUDIT_ME.md methodology)
2. **Phase 2**: Fix documentation discrepancies in README.md and ROADMAP*.md
3. **Phase 3**: Prioritize audit findings, implement top 5 code fixes
4. **Verify**: Run all quality checks above
5. **Document**: Create DOCUMENTATION-FIXES.md and FIXES-IMPLEMENTED.md summary files
