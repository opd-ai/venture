# Implementation Gap Analysis
Generated: 2026-02-07T18:17:21.346Z
Codebase Version: f4593ab935961ca53ed735fe3dafebc801a4c582

## Executive Summary
Total Gaps Found: 5
- Critical: 0 (1 completed)
- Moderate: 3
- Minor: 1

## Completed Gaps

### ✅ Gap #1: FishingSystem Uses Non-Deterministic Random in Core Game Logic [COMPLETED]
**Status:** Fixed on 2026-02-07

**Documentation Reference:**
> "All procedural generation MUST use seed-based deterministic algorithms. Never use `time.Now()`, global `math/rand` functions, or system-dependent randomness. Always use `rand.New(rand.NewSource(seed))` to ensure same seed = same output." (Project Overview - Code Assistance Guidelines, Rule #2)

**Implementation Location:** `pkg/engine/fishing_system.go:556-571`, `pkg/engine/fishing_system.go:574-585`, `pkg/engine/fishing_system.go:607-610`

**Solution Implemented:**
1. Added `seed int64` and `rng *rand.Rand` fields to `FishingSystem` struct
2. Modified `NewFishingSystem(world *World)` to `NewFishingSystem(world *World, seed int64)`
3. Replaced all three instances of `time.Now().UnixNano()` with deterministic `fs.rng` usage
4. Added deterministic sorting to `buildEligibleFishList()` to ensure consistent fish ordering from map iteration
5. Updated all tests to use deterministic seeds
6. Created comprehensive determinism tests in `fishing_determinism_test.go`

**Changes Made:**
- `pkg/engine/fishing_system.go`: Added seed/rng fields, updated constructor, fixed 3 non-deterministic calls, added fish list sorting
- `pkg/engine/fishing_system_test.go`: Updated 34 test calls to include seed parameter
- `pkg/engine/fishing_component_test.go`: Updated 4 test calls to include seed parameter
- `pkg/engine/fishing_determinism_test.go`: Created new test file with determinism verification tests

**Verification:**
- All determinism tests pass ✓
- Project compiles successfully ✓
- No regressions in existing tests ✓

---

### Gap #1: FishingSystem Uses Non-Deterministic Random in Core Game Logic
**Documentation Reference:**
> "All procedural generation MUST use seed-based deterministic algorithms. Never use `time.Now()`, global `math/rand` functions, or system-dependent randomness. Always use `rand.New(rand.NewSource(seed))` to ensure same seed = same output." (Project Overview - Code Assistance Guidelines, Rule #2)

**Implementation Location:** `pkg/engine/fishing_system.go:556-571`, `pkg/engine/fishing_system.go:574-585`, `pkg/engine/fishing_system.go:607-610`

**Expected Behavior:** FishingSystem should use seed-based deterministic random for reproducible fish selection and weight calculation.

**Actual Implementation:** Three methods in FishingSystem use `time.Now().UnixNano()` as seed for random generation, making fish catches and weights non-deterministic and non-reproducible across save/load or multiplayer sessions.

**Gap Details:** The FishingSystem violates the core determinism requirement in three critical places:

1. `selectRandomFish()` at line 557: Uses `time.Now().UnixNano()` to seed RNG for fish selection
2. `calculateFishWeight()` at line 575: Uses `time.Now().UnixNano()` for weight randomization  
3. `processReeling()` at line 608: Uses `time.Now().UnixNano()` for fish struggle updates

This breaks multiplayer synchronization (server and client get different results) and save/load reproducibility.

**Reproduction:**
```go
// Call FishingSystem methods twice in sequence
// Results differ because time.Now() changes between calls
fish1 := fs.selectRandomFish(eligibleList)
fish2 := fs.selectRandomFish(eligibleList)
// fish1 != fish2 even with identical inputs
```

**Production Impact:** Critical - Causes desync between server and clients in multiplayer fishing, and non-reproducible gameplay after loading saves.

**Evidence:**
```go
// pkg/engine/fishing_system.go:556-558
func (fs *FishingSystem) selectRandomFish(eligible eligibleFishList) *FishType {
	seed := time.Now().UnixNano()  // Non-deterministic!
	rng := rand.New(rand.NewSource(seed))
```

---

## Remaining Gaps

### Gap #2: Mobile Build Uses Non-Deterministic World Seed Generation
**Documentation Reference:**
> "All procedural generation MUST use seed-based deterministic algorithms. Never use `time.Now()`, global `math/rand` functions, or system-dependent randomness." (Project Overview - Code Assistance Guidelines, Rule #2)

**Implementation Location:** `cmd/mobile/mobile.go:34`

**Expected Behavior:** Mobile builds should have a configurable or reproducible world seed for testing and debugging.

**Actual Implementation:** The mobile entry point unconditionally sets `worldSeed = time.Now().UnixNano()` with no option to override.

**Gap Details:** While the desktop client allows `-seed` flag for reproducible world generation, the mobile build hardcodes a time-based seed. This prevents:
- Reproducible bug reports from mobile testers
- Deterministic world comparison between mobile and desktop
- Save/load testing with consistent world state

**Reproduction:**
```go
// In cmd/mobile/mobile.go init()
// No way to set a specific seed for testing
worldSeed = time.Now().UnixNano()  // Always different
```

**Production Impact:** Moderate - Debugging and QA testing on mobile platforms is hampered by non-reproducible world states.

**Evidence:**
```go
// cmd/mobile/mobile.go:33-34
func init() {
	// ...
	worldSeed = time.Now().UnixNano()
```

---

### Gap #3: Client Missing `-high-latency` Flag Documented in Server
**Documentation Reference:**
> "The multiplayer networking layer supports high-latency connections (200–5000ms) suitable for Tor/onion service routing, with client-side prediction, lag compensation, and snapshot synchronization." (README.md:9)
>
> "| `-high-latency` | `false` | Optimize for Tor/high-latency connections (200–5000ms) |" (README.md:244 - Server Flags)

**Implementation Location:** `cmd/client/util.go:44-81` (flag definitions), `cmd/server/main.go:40` (server has flag)

**Expected Behavior:** Both client and server should have `-high-latency` flag for configuring high-latency mode, as both endpoints need optimized settings.

**Actual Implementation:** Only the server implements the `-high-latency` flag. The client has no corresponding flag to enable high-latency optimizations for its network client configuration.

**Gap Details:** The server correctly implements:
```go
highLatency = flag.Bool("high-latency", false, "...")
```

But the client's `util.go` has no matching flag. The client documentation (`cmd/client/doc.go:124`) mentions "The client supports high-latency connections" but provides no mechanism to enable it.

**Reproduction:**
```bash
# Server can enable high-latency mode
./venture-server -high-latency

# Client has no equivalent - must connect without optimization
./venture-client --multiplayer --server example.onion:8080
# No -high-latency flag available
```

**Production Impact:** Moderate - Users connecting over Tor/high-latency networks cannot optimize client settings to match server configuration.

**Evidence:**
```go
// cmd/server/main.go:40 - EXISTS
highLatency = flag.Bool("high-latency", false, "Use high-latency configuration optimized for Tor/onion services (200-5000ms latency)")

// cmd/client/util.go:44-81 - MISSING
// No high-latency flag defined
```

---

### Gap #4: NetworkSimulator Uses Non-Deterministic RNG for Testing
**Documentation Reference:**
> "Use `rand.New(rand.NewSource(seed))` to ensure same seed = same output." (Project Overview - Code Assistance Guidelines, Rule #2)

**Implementation Location:** `pkg/network/resilience/simulator.go:58`

**Expected Behavior:** NetworkSimulator used for resilience testing should allow seeded random for reproducible test scenarios.

**Actual Implementation:** `NewNetworkSimulator()` hardcodes `time.Now().UnixNano()` as seed, making network simulation tests non-reproducible.

**Gap Details:** The NetworkSimulator is used for testing network conditions (packet loss, latency, jitter). Non-deterministic random makes:
- Intermittent test failures hard to reproduce
- CI/CD flaky tests possible
- Debugging network behavior difficult

**Reproduction:**
```go
// Run the same test twice
sim1 := resilience.NewNetworkSimulator()
sim2 := resilience.NewNetworkSimulator()
// sim1 and sim2 have different random sequences
// Same packet might be "dropped" in one but not the other
```

**Production Impact:** Moderate - Affects test reliability and debugging, not production gameplay directly.

**Evidence:**
```go
// pkg/network/resilience/simulator.go:58
func NewNetworkSimulator() *NetworkSimulator {
	return &NetworkSimulator{
		// ...
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
```

---

### Gap #5: LOG_LEVEL Environment Variable Default Behavior Underdocumented
**Documentation Reference:**
> "| `LOG_LEVEL` | `debug`, `info`, `warn`, `error`, `fatal` | Logging verbosity (unknown values default to `info`) |" (README.md:251)

**Implementation Location:** `pkg/logging/logger.go:117-133`

**Expected Behavior:** Documentation states "unknown values default to `info`"

**Actual Implementation:** This is correctly implemented, but the behavior varies based on `--verbose` flag interaction:

1. When `LOG_LEVEL` is set to unknown value → defaults to `info` ✓
2. When `LOG_LEVEL` is unset AND `--verbose=true` (default) → uses `debug`, not `info`

**Gap Details:** The README states unknown LOG_LEVEL values default to `info`, which is correct. However, users may be confused because:
- Client default `--verbose=true` overrides to `debug` when LOG_LEVEL is unset
- Server default `--verbose=true` also overrides to `debug`

The documentation accurately describes the LOG_LEVEL fallback but doesn't mention the verbose flag interaction.

**Reproduction:**
```bash
# With no LOG_LEVEL set, client uses debug (due to --verbose=true default)
./venture-client
# Logs at debug level

# With invalid LOG_LEVEL, correctly falls back to info
LOG_LEVEL=invalid ./venture-client
# Logs at info level (verbose flag is evaluated after LOG_LEVEL)
```

**Production Impact:** Minor - Accurate documentation, but interaction with `--verbose` could confuse users expecting `info` as default.

**Evidence:**
```go
// pkg/logging/logger.go:117-133
func parseLogLevel(level LogLevel) logrus.Level {
	switch level {
	// ...
	default:
		return logrus.InfoLevel  // Correctly defaults to info
	}
}

// cmd/client/util.go:96-102
if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
	logConfig.Level = logging.LogLevel(logLevel)
} else if *verbose {  // verbose=true by default
	logConfig.Level = logging.DebugLevel  // Overrides to debug
}
```

---

## Recommendations

### Critical (Address Immediately)
1. **FishingSystem Determinism** - Refactor to accept seed parameter or use system-wide game seed for all random operations

### Moderate (Address Before Production)
2. **Mobile Seed Configuration** - Add seed configuration mechanism for mobile builds (environment variable or config file)
3. **Client High-Latency Flag** - Add `-high-latency` flag to client matching server implementation
4. **NetworkSimulator Seeding** - Add constructor variant accepting explicit seed for reproducible testing

### Minor (Documentation Enhancement)
5. **LOG_LEVEL Documentation** - Add note about `--verbose` flag interaction with LOG_LEVEL environment variable
