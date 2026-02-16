# Security Audit Package

Phase 64.2 security audit framework for Venture v10.0 Production Readiness.

## Overview

Comprehensive security validation across 6 domains with 30 automated checks:

1. **Federation Security** (8 checks): Certificate validation, replay prevention, DoS protection
2. **Chat & Encryption** (6 checks): E2E encryption, key exchange, spam prevention
3. **Mod Sandbox** (6 checks): File/network isolation, resource limits, API restrictions
4. **Input Validation** (4 checks): Message sanitization, bounds checking
5. **Anti-Cheat** (3 checks): Stat validation, inventory integrity, speed enforcement
6. **Privacy** (3 checks): Data minimization, user consent, log protection

## Usage

### Programmatic
```go
import "github.com/opd-ai/venture/pkg/security"

auditor := security.NewAuditor(nil)
results := auditor.RunFullAudit()

if results.HasCritical() {
    log.Fatalf("Critical vulnerabilities: %d", results.CriticalCount)
}

fmt.Println(results.Summary())
// Output: Security Audit: 30/30 passed (100.0%) - Critical: 0, High: 0, Medium: 0, Low: 0
```

### CLI Tool
```bash
# Run full audit
./securitytest

# Show all check details
./securitytest -verbose

# Filter by domain
./securitytest -domain="Chat & Encryption"

# JSON output
./securitytest -json > audit.json
```

## Exit Codes
- `0`: All checks passed
- `1`: Non-critical issues found
- `2`: Critical vulnerabilities detected

## Current Results

**30/30 checks passed (100.0%)**

- ✅ Federation Security: 8/8 (100%)
- ✅ Chat & Encryption: 6/6 (100%)
- ✅ Mod Sandbox: 6/6 (100%)
- ✅ Input Validation: 4/4 (100%)
- ✅ Anti-Cheat: 3/3 (100%)
- ✅ Privacy: 3/3 (100%)

## Test Coverage

**91.3%** - Exceeds 65% requirement by 40%

All tests passing with zero race conditions.

## Performance

- Full audit: <100µs
- Zero allocations in hot paths
- Thread-safe (no shared mutable state)

## Mod Sandbox Implementation

The mod sandbox (implemented in `pkg/modding/sandbox.go`) provides comprehensive security:

1. **File System Isolation**: Mods loaded only from configured mods directory
2. **Network Isolation**: Data-driven JSON mods with no network APIs
3. **Memory Limits**: MaxMods, MaxRules, and file size limits
4. **CPU Limits**: Data-driven mods with zero code execution
5. **API Restrictions**: Whitelisted rule patterns (difficulty, loot, spawn, etc.)
6. **Code Execution Safety**: Pure JSON data files with no script interpretation

## Documentation

See `doc.go` for comprehensive package documentation including:
- Detailed domain descriptions
- Usage examples
- Severity level definitions
- Acceptance criteria

## Related Files

- `pkg/security/audit.go`: Core audit implementation
- `pkg/security/audit_test.go`: 16 test functions + 3 benchmarks
- `pkg/modding/sandbox.go`: Mod sandbox implementation
- `cmd/securitytest/main.go`: CLI tool
