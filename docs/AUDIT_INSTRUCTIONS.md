# AUDIT INSTRUCTIONS

**NOTE:** This file contains instructions and template for conducting a comprehensive code audit. This is NOT the audit results themselves. To conduct an audit following these instructions, the output should be placed in `AUDIT.md` in the repository root.

---

## TASK

Conduct a comprehensive audit of a Go codebase to identify bugs, performance issues, missing features, anti-patterns, and API design flaws while respecting existing codebase conventions where they meet minimum Go idiomacy standards. Place the output into a file called `AUDIT.md` directly in the repository root.

## CONTEXT 
This audit targets Go codebases of any maturity level. The audit should surface actionable issues while respecting established conventions in the codebase unless they violate fundamental Go principles. Balance between consistency with existing patterns and adherence to Go best practices.

AUDIT APPROACH:

When evaluating code:
1. **Assess existing conventions**: Identify the codebase's established patterns (naming, structure, error handling style, etc.)
2. **Apply minimum idiomacy threshold**: Respect conventions that are "Go-like enough" even if not textbook perfect
3. **Flag serious violations**: Highlight patterns that materially harm correctness, safety, or maintainability
4. **Suggest improvements within context**: Recommend fixes that align with codebase style where possible

AUDIT CATEGORIES:

## 1. CORRECTNESS & RELIABILITY
Examine for:
- Race conditions and concurrency bugs
- Nil pointer dereferences and invalid memory access
- Error handling problems (ignored errors, lost context)
- Logic errors and edge case failures
- Resource leaks (goroutines, file handles, connections, memory)
- Panic risks and missing recovery where appropriate
- Data corruption possibilities
- Type safety issues (unsafe operations, unchecked assertions)

## 2. PERFORMANCE & EFFICIENCY
Identify:
- Memory allocation inefficiencies
- Algorithmic complexity issues
- Concurrency bottlenecks
- I/O inefficiencies
- Unnecessary computational work
- Lock contention and synchronization overhead
- Resource pooling opportunities
- Optimization opportunities in hot paths

## 3. API DESIGN & USABILITY
Evaluate:
- Safety: Can the API be misused? Are invariants enforced?
- Completeness: Is necessary functionality exposed?
- Clarity: Are function signatures and behaviors obvious?
- Consistency: Does the API follow consistent patterns?
- Footguns: What can go wrong? How can it be prevented?
- Documentation: Are contracts and expectations clear?
- Versioning: Is the API stable and evolution-friendly?

## 4. CODE QUALITY & MAINTAINABILITY
Check for:
- Anti-patterns and problematic idioms
- Code organization and structure issues
- Naming clarity and consistency
- Documentation gaps
- Test coverage and quality
- Code duplication
- Complexity that could be simplified
- Dependency management issues

## 5. COMPLETENESS & PRODUCTION READINESS
Assess:
- Missing critical functionality
- Observability (logging, metrics, tracing)
- Configuration and parameterization
- Error messages and debugging support
- Graceful degradation and fault tolerance
- Lifecycle management and cleanup
- Security considerations
- Platform compatibility

OUTPUT FORMAT:

Provide findings in this structure:

```
## [CATEGORY]: [Concise Description]

**Severity**: [Critical | High | Medium | Low | Info]
**Location**: `path/to/file.go` (lines X-Y) or `package/component`
**Convention Note**: [If applicable: how this relates to codebase conventions]

**Description**:
[Clear explanation of the issue and why it matters]

**Current Code**:
```go
// Relevant code snippet showing the issue
```

**Impact**:
[What problems this causes or could cause]

**Recommendation**:
```go
// Suggested improvement
```

[OR, for broader issues without specific code:]

**Recommendation**:
[Description of suggested approach or changes]

**Justification**:
[Why this recommendation improves the code, respecting codebase style where appropriate]

---
```

End with:

```
## SUMMARY

**Critical Issues**: [count]
[Brief list if any]

**High Priority**: [count]
[Brief list]

**Medium Priority**: [count]

**Low Priority/Info**: [count]

**Convention Assessment**:
[Summary of codebase conventions and where they align/conflict with Go idioms]

**Overall Health**:
[2-3 paragraph assessment covering:
- Primary concerns and risk areas
- Strengths and positive patterns observed
- Recommended prioritization for addressing issues
- Any systemic patterns requiring architectural attention]

**Quick Wins**:
[List of easy-to-fix issues with high impact]
```

SEVERITY GUIDELINES:

- **Critical**: Crashes, data loss, security vulnerabilities, severe race conditions
- **High**: Reliability issues, significant performance problems, major API safety gaps, resource leaks
- **Medium**: Moderate performance issues, API usability problems, maintainability concerns
- **Low**: Minor inefficiencies, style inconsistencies, documentation gaps
- **Info**: Observations, suggestions, questions for maintainers

CONVENTION HANDLING:

When encountering non-standard but consistent patterns:
1. Note the pattern exists and is used consistently
2. Evaluate if it violates fundamental Go principles (safety, clarity, correctness)
3. If it's merely "different" but functional: respect it, perhaps note it as "Info"
4. If it causes real problems: flag it with explanation of why, suggest migration path
5. For recommendations: provide options that work within the existing style when possible

QUALITY CRITERIA:
✓ Issues are specific and verifiable
✓ Severity ratings reflect actual impact
✓ Recommendations are actionable and consider codebase context
✓ Code examples accurately represent the problem
✓ Suggestions respect existing conventions where reasonable
✓ Critical issues are distinguished from stylistic preferences
✓ Summary provides clear prioritization guidance

VALIDATION STEPS:

Before finalizing audit:
- Verify all file paths and line numbers are accurate
- Ensure code examples are syntactically valid
- Confirm recommendations actually address the identified issues
- Check that severity ratings are consistent across similar issues
- Validate that convention assessments are fair and context-aware
- Review that suggestions are practical given codebase constraints

EXAMPLE OUTPUT:

## CORRECTNESS: Unprotected Concurrent Map Access

**Severity**: Critical
**Location**: `internal/store/cache.go` (lines 34-51)
**Convention Note**: Codebase uses minimal locking; this pattern appears in multiple components

**Description**:
The cache implementation has concurrent reads and writes to a map without synchronization. While the codebase appears to favor lightweight concurrency patterns, maps in Go require explicit synchronization for concurrent access.

**Current Code**:
```go
type Cache struct {
    data map[string][]byte
}

func (c *Cache) Get(key string) []byte {
    return c.data[key]  // Concurrent read
}

func (c *Cache) Put(key string, val []byte) {
    c.data[key] = val  // Concurrent write
}
```

**Impact**:
- Runtime panics: "concurrent map read and map write"
- Undefined behavior under load
- Fails `go test -race`
- Production stability risk

**Recommendation**:
```go
type Cache struct {
    mu   sync.RWMutex
    data map[string][]byte
}

func (c *Cache) Get(key string) []byte {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.data[key]
}

func (c *Cache) Put(key string, val []byte) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = val
}
```

**Justification**:
This is non-negotiable per Go's memory model. sync.RWMutex allows concurrent reads while protecting writes. The defer pattern aligns with error handling style used elsewhere in the codebase.

---

## API DESIGN: Exported Mutable State

**Severity**: High
**Location**: `pkg/config/config.go` (lines 12-15)
**Convention Note**: Codebase uses package-level variables for configuration; suggesting safer pattern

**Description**:
Configuration is exposed as exported, mutable package variables. While this matches the codebase's pattern of package-level state, it enables race conditions and makes testing difficult.

**Current Code**:
```go
package config

var (
    MaxRetries = 3
    Timeout    = 30 * time.Second
)
```

**Impact**:
- Tests can interfere with each other
- Race conditions if changed at runtime
- No validation of configuration values
- Difficult to use multiple configurations

**Recommendation**:
```go
package config

type Config struct {
    MaxRetries int
    Timeout    time.Duration
}

func Default() *Config {
    return &Config{
        MaxRetries: 3,
        Timeout:    30 * time.Second,
    }
}
```

**Justification**:
Encapsulation prevents misuse while maintaining simplicity. If package-level state is preferred for consistency, consider using sync.RWMutex and getter functions at minimum.

---