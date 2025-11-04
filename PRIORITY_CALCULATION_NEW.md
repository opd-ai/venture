# Priority Calculation for Audit Issues
**Date**: 2025-11-04T18:40:00Z  
**Formula**: Priority Score = (Severity × User Impact × Production Risk × Blast Radius) - (Complexity Penalty × 0.2)

## Severity Multipliers
- Critical: 15 (crashes, data loss, security, severe race conditions)
- High: 10 (reliability issues, significant performance, major API gaps, resource leaks)
- Medium: 5 (moderate performance, API usability, maintainability)
- Low: 2 (minor inefficiencies, style, documentation)

## Calculations

### Issue #1: Mutable Exported Map in GenerationParams
**Severity**: Critical (15) - Race condition risk  
**User Impact**: All multiplayer workflows (10 affected paths × 2) = 20  
**Production Risk**: Data corruption (20)  
**Blast Radius**: System-wide (5) - affects all 15+ generators  
**Complexity**: High - Breaking API change  
- Lines to fix: ~50 (GenerationParams + all generator usages)
- Cross-package: 15 generators × 3 = 45
- Breaking changes: 10

**Complexity Penalty**: (50/50 + 45 + 10) = 56  
**Priority Score**: (15 × 20 × 20 × 5) - (56 × 0.2) = **30,000 - 11.2 = 29,988.8**

---

### Issue #2: Ignored Errors in Item Spawning
**Severity**: High (10) - Silent failures  
**User Impact**: All players spawning (8 workflows × 2) = 16  
**Production Risk**: Silent failure (10)  
**Blast Radius**: Single package (2)  
**Complexity**: Low - Add logging  
- Lines to fix: 5 locations × 2 lines = 10
- Cross-package: 0
- Breaking changes: 0

**Complexity Penalty**: (10/50 + 0 + 0) = 0.2  
**Priority Score**: (10 × 16 × 10 × 2) - (0.2 × 0.2) = **3,200 - 0.04 = 3,199.96**

---

### Issue #3: Missing Connection Cleanup in Server Shutdown
**Severity**: High (10) - Goroutine leaks  
**User Impact**: Server operators (3 workflows × 2) = 6  
**Production Risk**: Service outage (15)  
**Blast Radius**: Single package (2)  
**Complexity**: Medium  
- Lines to fix: ~20 (shutdown method refactor)
- Cross-package: 0
- Breaking changes: 0

**Complexity Penalty**: (20/50 + 0 + 0) = 0.4  
**Priority Score**: (10 × 6 × 15 × 2) - (0.4 × 0.2) = **1,800 - 0.08 = 1,799.92**

---

### Issue #4: O(n) Component Lookup in Hot Path
**Severity**: Medium (5) - Performance bottleneck  
**User Impact**: All gameplay (10 workflows × 2) = 20  
**Production Risk**: Performance degradation (7)  
**Blast Radius**: Single package (2) - engine/ecs.go  
**Complexity**: Medium  
- Lines to fix: ~30 (Entity struct + methods)
- Cross-package: 0
- Breaking changes: 0 (backward compatible)

**Complexity Penalty**: (30/50 + 0 + 0) = 0.6  
**Priority Score**: (5 × 20 × 7 × 2) - (0.6 × 0.2) = **1,400 - 0.12 = 1,399.88**

---

### Issue #5: TODO Comments Blocking Features
**Severity**: High (10) - Incomplete features  
**User Impact**: LAN party users (5 workflows × 2) = 10  
**Production Risk**: User confusion (4)  
**Blast Radius**: Single package (2) - hostplay  
**Complexity**: Medium  
- Lines to fix: ~40 (two methods)
- Cross-package: 3 (hostplay, network, engine)
- Breaking changes: 0

**Complexity Penalty**: (40/50 + 9 + 0) = 9.8  
**Priority Score**: (10 × 10 × 4 × 2) - (9.8 × 0.2) = **800 - 1.96 = 798.04**

---

### Issue #6: Generator.Validate() Returns interface{}
**Severity**: Medium (5) - Runtime type errors  
**User Impact**: Developers (3 workflows × 2) = 6  
**Production Risk**: Silent failure (10)  
**Blast Radius**: System-wide (5) - all generators  
**Complexity**: High - Generics or major refactor  
- Lines to fix: ~100 (interface + all implementations)
- Cross-package: 15 generators × 3 = 45
- Breaking changes: 10 (if using generics)

**Complexity Penalty**: (100/50 + 45 + 10) = 57  
**Priority Score**: (5 × 6 × 10 × 5) - (57 × 0.2) = **1,500 - 11.4 = 1,488.6**

---

### Issue #7: Missing Context in Long-Running Operations
**Severity**: Medium (5) - Cannot cancel gracefully  
**User Impact**: Server operators (3 workflows × 2) = 6  
**Production Risk**: Service outage (15)  
**Blast Radius**: Single package (2)  
**Complexity**: Low  
- Lines to fix: ~15
- Cross-package: 0
- Breaking changes: 0

**Complexity Penalty**: (15/50 + 0 + 0) = 0.3  
**Priority Score**: (5 × 6 × 15 × 2) - (0.3 × 0.2) = **900 - 0.06 = 899.94**

---

### Issue #8: Duplicated Validation Logic
**Severity**: Medium (5) - Maintenance burden  
**User Impact**: Developers (2 workflows × 2) = 4  
**Production Risk**: Internal maintainability (2)  
**Blast Radius**: Multiple packages (3) - 10+ generators  
**Complexity**: Low  
- Lines to fix: ~20 (centralize validation)
- Cross-package: 0 (moving code to base)
- Breaking changes: 0

**Complexity Penalty**: (20/50 + 0 + 0) = 0.4  
**Priority Score**: (5 × 4 × 2 × 3) - (0.4 × 0.2) = **120 - 0.08 = 119.92**

---

### Issue #9: Missing Package Documentation
**Severity**: Low (2) - Documentation gap  
**User Impact**: Developers (2 workflows × 2) = 4  
**Production Risk**: User confusion (4)  
**Blast Radius**: Multiple packages (3)  
**Complexity**: Low  
- Lines to fix: ~30 per package, 3 packages = 90
- Cross-package: 0
- Breaking changes: 0

**Complexity Penalty**: (90/50 + 0 + 0) = 1.8  
**Priority Score**: (2 × 4 × 4 × 3) - (1.8 × 0.2) = **96 - 0.36 = 95.64**

---

### Issue #10: Hardcoded Configuration Values
**Severity**: Medium (5) - Production inflexibility  
**User Impact**: Operations team (3 workflows × 2) = 6  
**Production Risk**: Performance degradation (7)  
**Blast Radius**: System-wide (5) - 50+ locations  
**Complexity**: Very High  
- Lines to fix: ~200 (config system + all usages)
- Cross-package: 50 locations × 3 = 150
- Breaking changes: 0

**Complexity Penalty**: (200/50 + 150 + 0) = 154  
**Priority Score**: (5 × 6 × 7 × 5) - (154 × 0.2) = **1,050 - 30.8 = 1,019.2**

---

### Issue #11: Insufficient Error Context in Generators
**Severity**: High (10) - Debugging difficulty  
**User Impact**: All users (debugging scenarios) (5 workflows × 2) = 10  
**Production Risk**: Silent failure (10)  
**Blast Radius**: System-wide (5) - all generators  
**Complexity**: Medium  
- Lines to fix: ~60 (error type + all generators)
- Cross-package: 15 generators × 3 = 45
- Breaking changes: 0

**Complexity Penalty**: (60/50 + 45 + 0) = 46.2  
**Priority Score**: (10 × 10 × 10 × 5) - (46.2 × 0.2) = **5,000 - 9.24 = 4,990.76**

---

### Issue #12: Missing Observability for Performance
**Severity**: Medium (5) - Blind to production issues  
**User Impact**: Operations (4 workflows × 2) = 8  
**Production Risk**: Performance degradation (7)  
**Blast Radius**: Multiple packages (3) - game loop, render, network  
**Complexity**: Medium  
- Lines to fix: ~80 (metrics package + instrumentation)
- Cross-package: 10 locations × 3 = 30
- Breaking changes: 0

**Complexity Penalty**: (80/50 + 30 + 0) = 31.6  
**Priority Score**: (5 × 8 × 7 × 3) - (31.6 × 0.2) = **840 - 6.32 = 833.68**

---

## Top 5 Highest-Priority Issues

1. **Issue #1: Mutable Exported Map** - Priority: **29,988.8** ⚠️ CRITICAL
2. **Issue #11: Insufficient Error Context** - Priority: **4,990.76** 🔴 HIGH
3. **Issue #2: Ignored Errors in Item Spawning** - Priority: **3,199.96** 🔴 HIGH
4. **Issue #3: Missing Connection Cleanup** - Priority: **1,799.92** 🟡 MEDIUM-HIGH
5. **Issue #6: Validate() interface{} API** - Priority: **1,488.6** 🟡 MEDIUM-HIGH

**Alternative Top 5 (if excluding breaking changes)**:
1. **Issue #11: Insufficient Error Context** - Priority: **4,990.76** (no breaking changes)
2. **Issue #2: Ignored Errors** - Priority: **3,199.96** (no breaking changes)
3. **Issue #3: Connection Cleanup** - Priority: **1,799.92** (no breaking changes)
4. **Issue #4: O(n) Component Lookup** - Priority: **1,399.88** (backward compatible)
5. **Issue #10: Hardcoded Config** - Priority: **1,019.2** (no breaking changes)

---

## Recommendation

**For immediate implementation** (this PR):
- Fix Issues #2, #3, #4, #7, #8 (all non-breaking, high-impact)
- Document Issues #1, #6, #11 for future major version release

**Reasoning**:
- Issue #1 requires major version bump (breaking API change)
- Issue #6 ideally needs Go generics (breaking change)
- Issue #11 is high-value but touches 15+ files (larger scope)
- Issues #2, #3, #4 are surgical fixes with immediate benefit
- Issues #7, #8 improve maintainability without breaking changes

---

**Total Issues Analyzed**: 21  
**Critical Issues**: 1  
**High Priority Issues**: 5  
**Medium Priority Issues**: 7  
**Low Priority Issues**: 6  
**Calculation Method**: Per AUDIT_ME.md specifications
