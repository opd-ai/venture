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
- Identify features marked as "complete" that are partially implemented or missing
- Check for documentation inconsistencies (version numbers, status markers, completion dates)
- Verify phase deliverables match actual codebase state

**D. Code Quality Scan**
- Identify bugs: logic errors, race conditions, nil pointer dereferences, resource leaks
- Find security issues: injection vulnerabilities, authentication/authorization gaps, data exposure
- Detect performance problems: inefficient algorithms, memory leaks, unnecessary allocations
- Spot code smells: duplicated code, overly complex functions, poor error handling

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
Execute these automated validations:
1. Confirm all documented features from README.md have assessment status
2. Verify all audit findings from AUDIT.md/AUDIT_ME.md are addressed or documented
3. Validate roadmap claims in ROADMAP*.md match implementation or are flagged as gaps
4. Ensure each issue includes exact source reference (doc quote or code location)
5. Confirm all issues have reproducible evidence with code snippets or steps
6. Validate priority scoring calculations are mathematically correct per formula
7. Verify generated repair code is syntactically valid Go (compiles without errors)
8. Ensure repair code follows existing codebase patterns and conventions
9. Confirm all repairs include comprehensive test coverage (unit + integration + edge cases)
10. Validate repairs maintain backward compatibility where required
11. Check that no new security vulnerabilities are introduced by repairs
12. Verify documentation updates (README.md, ROADMAP*.md) reflect repair changes
13. Confirm all critical and high-severity issues are addressed in top 5 repairs
14. Validate that quick wins (high impact, low complexity) are prioritized appropriately

## EXAMPLES:

### Example 1: Implementation Gap Detection

### Issue #1: Rate Limiter Allows One Extra Request [Priority Score: 47.2]
**Category:** Implementation Gap - Functional Mismatch
**Severity:** Functional Mismatch (10 pts)
**Source:** README.md:147

**Reference/Quote:**
> "The API rate limiter enforces a strict limit of 100 requests per minute per IP address" (README.md:147)

**Location:** `middleware/ratelimit.go:52-67`

**Expected Behavior:** Exactly 100 requests allowed per minute, 101st request blocked

**Actual Behavior:** 101 requests allowed due to off-by-one error in counter comparison

**Issue Details:** The rate limiter uses `<=` comparison instead of `<`, allowing request 101 to proceed before blocking starts. This violates the documented "strict limit" guarantee.

**Reproduction Scenario:**
```go
// Send exactly 101 requests within 59 seconds
// Expected: Request 101 returns 429 Too Many Requests
// Actual: Request 101 returns 200 OK, request 102 returns 429
```

**Impact Assessment:**
- User Impact: Affects all API consumers (workflows: 5)
- Production Risk: Service overload (7 pts), violates SLA
- Blast Radius: System-wide (5 pts)

**Code Evidence:**
```go
if requestCount <= limit { // BUG: Should be < not <=
    return next(ctx)
}
return ErrRateLimitExceeded
```

**Priority Calculation:**
- Severity: 10 × User Impact: 8 × Production Risk: 7 × Blast Radius: 5 - Complexity: 1
- Final Score: 47.2

**Dependencies:** None

---

### Example 2: Security Vulnerability Detection

### Issue #2: Missing Authentication on Admin Endpoints [Priority Score: 540.0]
**Category:** Security Issue - Authentication Gap
**Severity:** Critical Security (15 pts)
**Source:** Code Scan + AUDIT.md

**Reference/Quote:**
> "Admin endpoints at /admin/* lack authentication middleware" (AUDIT.md:45)

**Location:** `api/routes.go:120-135`

**Expected Behavior:** All /admin/* endpoints should require valid admin authentication token

**Actual Behavior:** Admin endpoints accessible without authentication

**Issue Details:** Router configuration directly maps /admin/* paths to handlers without authentication middleware. Any user can access admin functions including user deletion, data export, and system configuration.

**Reproduction Scenario:**
```bash
# No authentication required - direct access works
curl http://localhost:8080/admin/users/delete/123
# Expected: 401 Unauthorized
# Actual: 200 OK, user deleted
```

**Impact Assessment:**
- User Impact: Critical - all users' data at risk (workflows: 10)
- Production Risk: Security breach (18 pts), data corruption (20 pts)
- Blast Radius: System-wide (5 pts)

**Code Evidence:**
```go
// Missing auth middleware!
router.POST("/admin/users/delete/:id", handlers.DeleteUser)
router.GET("/admin/export", handlers.ExportAllData)
```

**Priority Calculation:**
- Severity: 15 × User Impact: 22 × Production Risk: 20 × Blast Radius: 5 - Complexity: 2
- Final Score: 540.0

**Dependencies:** Requires authentication middleware implementation

---

### Example 3: Performance Problem Detection

### Issue #3: O(n²) Search in Hot Path [Priority Score: 126.0]
**Category:** Performance Problem - Algorithmic Complexity
**Severity:** Performance Critical (10 pts)
**Source:** Code Scan (profiling data)

**Location:** `search/finder.go:78-95`

**Expected Behavior:** Search should be O(n log n) or better for 10k+ item catalogs

**Actual Behavior:** Nested loop causes O(n²) behavior, 5+ second searches with 10k items

**Issue Details:** FindMatchingItems uses nested loop instead of hash map lookup. For each query term, iterates through all items. Profiling shows this function consumes 87% of CPU time during search operations.

**Reproduction Scenario:**
```go
// Catalog with 10,000 items, query with 5 terms
// Expected: < 100ms search time
// Actual: 5,200ms search time (52x slower than spec)
```

**Impact Assessment:**
- User Impact: All search operations (workflows: 8)
- Production Risk: Degraded performance (7 pts), user frustration
- Blast Radius: Single module (2 pts)

**Code Evidence:**
```go
// O(n²) nested loop - should use map
for _, term := range queryTerms {
    for _, item := range allItems { // Scans all items for each term!
        if strings.Contains(item.Name, term) {
            matches = append(matches, item)
        }
    }
}
```

**Priority Calculation:**
- Severity: 10 × User Impact: 18 × Production Risk: 7 × Blast Radius: 2 - Complexity: 3
- Final Score: 126.0

---

### Example 4: Documentation Inconsistency Detection

### Issue #4: Version Number Mismatch Across Documentation [Priority Score: 24.0]
**Category:** Documentation Issue - Inconsistency
**Severity:** Documentation Inconsistency (4 pts)
**Source:** README.md + ROADMAP.md + ROADMAP_V2.md

**Reference/Quote:**
- README.md:24: "Version: 1.1 Production"
- ROADMAP.md:11: "Version: 1.1 Production + 2.0 Phase 14 Complete"
- ROADMAP_V2.md:8: "Version: 2.0 - Enhanced Mechanics"

**Expected Behavior:** All documentation files should report consistent version number

**Actual Behavior:** Three different version descriptions cause user confusion about actual release status

**Issue Details:** Version number discrepancy makes it unclear whether project is at v1.1, v2.0, or some hybrid state. Users and contributors cannot determine which features are available.

**Impact Assessment:**
- User Impact: All users reading documentation (workflows: 3)
- Production Risk: User confusion (4 pts)
- Blast Radius: Multiple modules (3 pts) - affects 3 documentation files

**Priority Calculation:**
- Severity: 4 × User Impact: 9 × Production Risk: 4 × Blast Radius: 3 - Complexity: 0.5
- Final Score: 24.0

---

### Example 5: Code Quality Issue Detection

### Issue #5: Duplicated Error Handling Logic [Priority Score: 18.0]
**Category:** Code Quality - Duplication
**Severity:** Code Duplication (3 pts)
**Source:** Code Scan

**Location:** `handlers/*.go` (12 files)

**Expected Behavior:** Error handling should use shared helper function

**Actual Behavior:** Same 15-line error handling pattern duplicated in 12 handlers

**Issue Details:** Error handling logic (logging, status code selection, response formatting) repeated in 12 different handler files. Changes to error handling require updating all 12 locations, increasing maintenance burden and bug risk.

**Code Evidence:**
```go
// Duplicated in 12 files - should be extracted
if err != nil {
    logger.Error("Operation failed", "error", err)
    statusCode := 500
    if errors.Is(err, ErrNotFound) {
        statusCode = 404
    } else if errors.Is(err, ErrUnauthorized) {
        statusCode = 401
    }
    // ... 8 more lines of error handling
}
```

**Impact Assessment:**
- User Impact: Developers maintaining code (workflows: 2)
- Production Risk: Internal maintainability (2 pts)
- Blast Radius: Multiple modules (3 pts) - 12 handler files

**Priority Calculation:**
- Severity: 3 × User Impact: 4 × Production Risk: 2 × Blast Radius: 3 - Complexity: 4
- Final Score: 18.0

### Example Autonomous Repair:

## Repair #1: Rate Limiter Off-By-One Correction
**Original Issue Priority:** 47.2
**Issue Category:** Implementation Gap - Functional Mismatch
**Files Modified:** 2
**Lines Changed:** +15 -3

### Implementation Strategy
Minimal surgical fix changing comparison operator from `<=` to `<` in rate limiter logic. Added comprehensive test coverage to prevent regression and validate exact boundary behavior. No API changes or configuration modifications required. Maintains backward compatibility as the fix aligns implementation with documented behavior.

### Code Changes

#### File: middleware/ratelimit.go
**Action:** Modified
**Change Type:** Bug Fix - Implementation Gap

```go
// Line 52-67: Rate limiter check function
func (rl *RateLimiter) checkLimit(ctx context.Context, key string) error {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    requestCount := rl.counters[key]
    
    // FIXED: Changed <= to < for strict limit enforcement
    // This ensures exactly 'limit' requests are allowed, not 'limit + 1'
    // Aligns implementation with README.md:147 documentation guarantee
    if requestCount < rl.limit {
        rl.counters[key]++
        return nil
    }
    
    return ErrRateLimitExceeded
}
```

### Integration Requirements
- Dependencies: None (existing packages sufficient)
- Configuration: No changes required
- Migration: Not required (backward compatible fix - behavior now matches documentation)
- Documentation Updates: None (README.md already describes correct behavior)

### Validation Tests

#### Unit Tests: middleware/ratelimit_test.go
```go
// Added: Test for exact limit boundary behavior
func TestRateLimiterStrictLimit(t *testing.T) {
    limiter := NewRateLimiter(100, time.Minute)
    key := "test-ip"
    
    // Should allow exactly 100 requests
    for i := 0; i < 100; i++ {
        err := limiter.checkLimit(context.Background(), key)
        if err != nil {
            t.Fatalf("Request %d should succeed, got error: %v", i+1, err)
        }
    }
    
    // Request 101 should fail
    err := limiter.checkLimit(context.Background(), key)
    if err != ErrRateLimitExceeded {
        t.Fatalf("Request 101 should fail with ErrRateLimitExceeded, got: %v", err)
    }
}

// Added: Test concurrent access (race condition check)
func TestRateLimiterConcurrent(t *testing.T) {
    limiter := NewRateLimiter(100, time.Minute)
    key := "test-ip"
    var wg sync.WaitGroup
    successCount := atomic.Int32{}
    
    // Simulate 200 concurrent requests
    for i := 0; i < 200; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            if limiter.checkLimit(context.Background(), key) == nil {
                successCount.Add(1)
            }
        }()
    }
    wg.Wait()
    
    // Exactly 100 should succeed
    if got := successCount.Load(); got != 100 {
        t.Errorf("Expected exactly 100 successful requests, got %d", got)
    }
}
```

### Verification Results
- [✓] Syntax validation passed: `go build ./middleware`
- [✓] Pattern compliance verified: Uses existing mutex pattern, follows codebase conventions
- [✓] Tests pass: 9/9 (added 2 new tests) - `go test ./middleware -v`
- [✓] Linter passed: `golangci-lint run ./middleware` - no warnings
- [✓] Documentation alignment confirmed: Implementation now matches README.md:147
- [✓] No regressions detected: All existing tests pass
- [✓] Security review passed: No new attack vectors, proper synchronization
- [✓] Performance validated: No measurable performance impact (< 1% overhead)

### Deployment Instructions
1. Deploy to staging environment
2. Run full test suite: `go test ./...`
3. Run load test with exactly 100 requests per IP to verify boundary behavior
4. Monitor rate limiter metrics for 24 hours in staging
5. Verify 429 errors occur at exactly 100 requests per IP (not 101)
6. Deploy to production during low-traffic window (recommended: Tuesday 2-4 AM UTC)
7. Alert on-call team of deployment
8. Monitor error rates and rate limiter metrics for first 48 hours
9. Rollback procedure: Revert commit [hash] if error rate increases > 5%
