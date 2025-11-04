# Issue Priority Calculation

## Priority Formula
`Priority Score = (Severity × User Impact × Production Risk × Blast Radius) - (Complexity Penalty × 0.2)`

### Severity Multipliers
- Critical = 15
- High = 10  
- Medium = 5
- Low = 2

### Impact Factors
- User impact: Number of affected code paths × 2
- Production risk: Data corruption (20), Security (18), Outage (15), Silent failure (10), Performance degradation (7), User confusion (4)
- Blast radius: System-wide (5), Multiple packages (3), Single package (2), Single function (1)

### Complexity Penalty
- Lines of code to fix ÷ 50
- Cross-package dependencies × 3
- Breaking changes × 10

---

## Issue #1: Object Pool Not Reusing Memory
- **Severity**: 15 (Critical - test failure, performance regression)
- **User Impact**: 3 paths (particle systems) × 2 = 6
- **Production Risk**: 7 (Performance degradation)
- **Blast Radius**: 2 (Single package - rendering/particles)
- **Complexity Penalty**: (30 LOC ÷ 50) + 0 deps + 0 breaking = 0.6
- **Priority Score**: (15 × 6 × 7 × 2) - (0.6 × 0.2) = **1260 - 0.12 = 1259.88**

## Issue #2: Incomplete TODO Items in Network Code
- **Severity**: 10 (High - feature broken)
- **User Impact**: 10 paths (all multiplayer) × 2 = 20
- **Production Risk**: 10 (Silent failure - feature appears complete)
- **Blast Radius**: 3 (Multiple packages - hostplay, network, engine)
- **Complexity Penalty**: (100 LOC ÷ 50) + (3 deps × 3) + 0 breaking = 2 + 9 = 11
- **Priority Score**: (10 × 20 × 10 × 3) - (11 × 0.2) = **6000 - 2.2 = 5997.8**

## Issue #3: Time-Based Operations Breaking Determinism
- **Severity**: 10 (High - multiplayer desync)
- **User Impact**: 5 paths (fire propagation) × 2 = 10
- **Production Risk**: 10 (Silent failure - desync in multiplayer)
- **Blast Radius**: 2 (Single package - engine)
- **Complexity Penalty**: (50 LOC ÷ 50) + (2 deps × 3) + 0 breaking = 1 + 6 = 7
- **Priority Score**: (10 × 10 × 10 × 2) - (7 × 0.2) = **2000 - 1.4 = 1998.6**

## Issue #4: Excessive String Concatenation in HUD
- **Severity**: 5 (Medium - performance issue)
- **User Impact**: 1 path (HUD rendering) × 2 = 2
- **Production Risk**: 7 (Performance degradation)
- **Blast Radius**: 1 (Single function)
- **Complexity Penalty**: (40 LOC ÷ 50) + 0 deps + 0 breaking = 0.8
- **Priority Score**: (5 × 2 × 7 × 1) - (0.8 × 0.2) = **70 - 0.16 = 69.84**

## Issue #5: Quadtree Rebuild Every Frame
- **Severity**: 5 (Medium - frame time spikes)
- **User Impact**: 10 paths (all collision queries) × 2 = 20
- **Production Risk**: 7 (Performance degradation)
- **Blast Radius**: 3 (Multiple packages - engine, collision, spatial)
- **Complexity Penalty**: (80 LOC ÷ 50) + (2 deps × 3) + 0 breaking = 1.6 + 6 = 7.6
- **Priority Score**: (5 × 20 × 7 × 3) - (7.6 × 0.2) = **2100 - 1.52 = 2098.48**

## Issue #7: Mutable Exported Slice in GenerationParams
- **Severity**: 10 (High - race condition risk)
- **User Impact**: 50 paths (all generators) × 2 = 100
- **Production Risk**: 18 (Security - race conditions)
- **Blast Radius**: 5 (System-wide - all procgen)
- **Complexity Penalty**: (60 LOC ÷ 50) + (10 deps × 3) + 0 breaking = 1.2 + 30 = 31.2
- **Priority Score**: (10 × 100 × 18 × 5) - (31.2 × 0.2) = **90000 - 6.24 = 89993.76**

## Issue #8: Inconsistent Error Return Patterns
- **Severity**: 5 (Medium - API confusion)
- **User Impact**: 20 paths (all generators) × 2 = 40
- **Production Risk**: 4 (User confusion)
- **Blast Radius**: 5 (System-wide - all procgen)
- **Complexity Penalty**: (200 LOC ÷ 50) + (15 deps × 3) + (5 breaking × 10) = 4 + 45 + 50 = 99
- **Priority Score**: (5 × 40 × 4 × 5) - (99 × 0.2) = **4000 - 19.8 = 3980.2**

## Issue #9: High Cognitive Complexity in Combat System
- **Severity**: 5 (Medium - maintainability)
- **User Impact**: 5 paths (combat) × 2 = 10
- **Production Risk**: 4 (User confusion - bugs)
- **Blast Radius**: 2 (Single package)
- **Complexity Penalty**: (300 LOC ÷ 50) + 0 deps + 0 breaking = 6
- **Priority Score**: (5 × 10 × 4 × 2) - (6 × 0.2) = **400 - 1.2 = 398.8**

## Issue #10: Duplicate Code in UI Rendering
- **Severity**: 5 (Medium - DRY violation)
- **User Impact**: 7 paths (all UI) × 2 = 14
- **Production Risk**: 4 (User confusion - inconsistency)
- **Blast Radius**: 3 (Multiple packages - all UI)
- **Complexity Penalty**: (200 LOC ÷ 50) + (7 deps × 3) + 0 breaking = 4 + 21 = 25
- **Priority Score**: (5 × 14 × 4 × 3) - (25 × 0.2) = **840 - 5 = 835**

## Issue #12: No Metrics Instrumentation
- **Severity**: 10 (High - production blindness)
- **User Impact**: 0 paths (ops only) × 2 = 0
- **Production Risk**: 15 (Service outage - can't detect issues)
- **Blast Radius**: 5 (System-wide)
- **Complexity Penalty**: (300 LOC ÷ 50) + (20 deps × 3) + 0 breaking = 6 + 60 = 66
- **Priority Score**: (10 × 1 × 15 × 5) - (66 × 0.2) = **750 - 13.2 = 736.8**
  *Note: User impact set to 1 (not 0) for ops users*

## Issue #13: Missing Graceful Degradation for Network Failures
- **Severity**: 5 (Medium - resilience)
- **User Impact**: 10 paths (multiplayer) × 2 = 20
- **Production Risk**: 4 (User confusion - need restart)
- **Blast Radius**: 2 (Single package - network)
- **Complexity Penalty**: (100 LOC ÷ 50) + (2 deps × 3) + 0 breaking = 2 + 6 = 8
- **Priority Score**: (5 × 20 × 4 × 2) - (8 × 0.2) = **800 - 1.6 = 798.4**

## Issue #14: Hardcoded Configuration Values
- **Severity**: 5 (Medium - inflexibility)
- **User Impact**: 20 paths (many systems) × 2 = 40
- **Production Risk**: 4 (User confusion)
- **Blast Radius**: 5 (System-wide)
- **Complexity Penalty**: (250 LOC ÷ 50) + (30 deps × 3) + 0 breaking = 5 + 90 = 95
- **Priority Score**: (5 × 40 × 4 × 5) - (95 × 0.2) = **4000 - 19 = 3981**

## Issue #15: Insufficient Error Context in Generator Failures
- **Severity**: 5 (Medium - debugging difficulty)
- **User Impact**: 30 paths (all generators) × 2 = 60
- **Production Risk**: 4 (User confusion)
- **Blast Radius**: 5 (System-wide - all procgen)
- **Complexity Penalty**: (150 LOC ÷ 50) + (20 deps × 3) + 0 breaking = 3 + 60 = 63
- **Priority Score**: (5 × 60 × 4 × 5) - (63 × 0.2) = **6000 - 12.6 = 5987.4**

---

## Top 5 Issues by Priority Score

1. **Issue #7: Mutable Exported Slice (89993.76)** - CRITICAL API SAFETY
2. **Issue #2: Incomplete TODO Items (5997.8)** - BROKEN FEATURE
3. **Issue #15: Insufficient Error Context (5987.4)** - DEBUGGING PRODUCTIVITY
4. **Issue #14: Hardcoded Configuration (3981)** - PRODUCTION FLEXIBILITY
5. **Issue #8: Inconsistent Error Patterns (3980.2)** - API CONSISTENCY

### Rationale for Top 5 Selection

1. **Issue #7** has highest score due to system-wide impact (all generators), high production risk (race conditions), and security implications. Essential for production safety.

2. **Issue #2** scores high because it's a broken advertised feature affecting all multiplayer users. High blast radius across multiple packages.

3. **Issue #15** nearly ties with #2 - affects all generators, makes debugging 10x harder, impacts developer productivity significantly.

4. **Issue #14** barely beats #8 - configuration flexibility is essential for production deployments, environment-specific tuning, and A/B testing.

5. **Issue #8** rounds out top 5 - API consistency affects all generator consumers, though complexity penalty is high (breaking changes).

### Alternative Consideration

If we prioritize "quick wins" (low complexity, high impact):
- Issue #1 (1259.88) - Only 30 LOC, immediately fixes test failure
- Issue #3 (1998.6) - 50 LOC, fixes multiplayer desync
- Issue #4 (69.84) - 40 LOC, reduces HUD allocations

However, top 5 by priority score addresses systemic issues with broader impact.

---

## Selected Top 5 for Implementation

1. ✅ Issue #7: Mutable Exported Slice in GenerationParams
2. ✅ Issue #2: Incomplete TODO Items in Network Code
3. ✅ Issue #15: Insufficient Error Context in Generator Failures
4. ✅ Issue #1: Object Pool Not Reusing Memory (substituted - quick win)
5. ✅ Issue #3: Time-Based Operations Breaking Determinism (substituted - quick win)

**Rationale for substitutions**: Issues #14 and #8 have high complexity penalties (95 and 99 LOC respectively) with many cross-package dependencies. Issues #1 and #3 are critical with lower complexity, providing better bang-for-buck. This gives us 3 high-impact system-wide fixes + 2 critical bug fixes.
