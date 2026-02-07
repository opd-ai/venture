TASK: Perform a detailed, multi-dimensional implementation gap audit of the venture Go/Ebiten game codebase.

CONTEXT: Venture is a procedural multiplayer action-RPG using Entity-Component-System (ECS) architecture. Key directories:
- `cmd/client/` — Desktop client entrypoint, UI systems, lazy initialization
- `cmd/server/` — Dedicated server, player management, system versioning (v4/v8/v9)
- `pkg/engine/` — Core ECS (World, Entity, Component, System), 100+ game systems
- `pkg/procgen/` — 25+ procedural generators (terrain, items, quests, NPCs, etc.)
- `pkg/rendering/` — Sprite generation, animation, lighting, particles, UI
- `pkg/network/` — Client/server protocol, federation, WebRTC, resilience
- `pkg/world/` — Chunk persistence, housing, economy, territory, raids
- `pkg/integration/` — Cross-system integrations (guild housing, trade routes, etc.)
- `pkg/audio/` — Procedural music, SFX, synthesis

OBJECTIVE: Identify every category of implementation gap in the codebase — not just missing instantiations, but also incomplete integrations, dead code paths, unresolved TODOs, and mismatches between documentation and runtime behavior.

SCOPE: Audit each of the following gap categories systematically.

## 1. Instantiation Gaps
Scan all packages for struct definitions (systems, components, managers) and cross-reference against runtime initialization in `cmd/client/main.go`, `cmd/server/main.go`, and system registration files (`v4_systems.go`, `v8_systems.go`, `v9_systems.go`, `system_wrappers.go`).

Flag items that are:
- Defined but never instantiated outside `_test.go` files
- Instantiated in server but missing from client (or vice versa)
- Conditionally instantiated but with no configuration path to enable them
- Registered in older system versions (v4/v8) but missing from current (v9)

## 2. Interface Compliance Gaps
Check that all interfaces declared in `pkg/engine/interfaces.go` and other package-level interfaces have complete implementations. Flag:
- Interfaces with zero runtime implementations
- Implementations that satisfy the interface signature but contain stub/no-op logic (empty method bodies, hardcoded returns)
- Components missing the required `Type() string` method
- Systems missing the required `Update(entities []*Entity, deltaTime float64)` method

## 3. Integration Wiring Gaps
Examine `pkg/integration/` packages and cross-system interaction points. Flag:
- Integration packages that import their dependencies but never call key methods
- Event emitters with no registered listeners (or listeners with no emitters)
- Systems that reference other systems by name/type but where the referenced system is never instantiated
- Circular or missing dependency chains between subsystems

## 4. Procedural Generation Gaps
Audit `pkg/procgen/` generators for completeness. Flag:
- Generators implementing the `Generator` interface but never invoked outside tests
- Generation parameters (`Difficulty`, `Depth`, `GenreID`) that are accepted but ignored in logic
- Genre-specific code paths that exist for some genres but not others (e.g., fantasy has terrain gen but cyberpunk does not)
- Seeds that are accepted but not propagated to sub-generators (non-determinism risk)

## 5. Network Protocol Gaps
Audit `pkg/network/` for completeness. Flag:
- Packet types defined but never sent or handled
- Client-side prediction logic with no corresponding server reconciliation
- Federation features (discovery, sync, transfer) that are partially wired
- Concrete type usage (`net.UDPConn`, `net.TCPConn`) instead of interfaces (`net.PacketConn`, `net.Conn`)

## 6. Dead Code and TODO Gaps
Scan the entire codebase for:
- Exported functions/methods with zero callers outside their own package and tests
- `TODO`, `FIXME`, `HACK`, `XXX` comments indicating unfinished work
- Commented-out runtime logic or feature code (especially blocks longer than 5 lines) that indicates disabled/dead code paths or abandoned work (exclude purely stylistic or documentation comments)
- Empty function/method bodies in non-test files

## 7. Documentation-to-Code Gaps
Cross-reference feature documentation (`README.md` and user-facing `docs/*.md` guides) and CLI flag definitions against actual code. Audit/meta docs (like this file) are not subject to gap detection. Flag:
- Features described in documentation with no corresponding implementation
- CLI flags defined in code but missing from documentation (or vice versa)
- Performance targets stated in docs (e.g., 60 FPS, <500MB RAM) with no benchmark or validation code
- API contracts or configuration formats documented but not enforced by validation

OUTPUT FORMAT:
```
## Implementation Gap Audit Report

### Executive Summary
- Total gaps found: [N]
- Critical (blocks functionality): [N]
- High (degrades quality): [N]
- Medium (incomplete feature): [N]
- Low (cosmetic/cleanup): [N]

### Gap Category Breakdown

#### 1. Instantiation Gaps
| System/Component | Defined In | Expected Runtime Location | Status | Severity |
|------------------|-----------|--------------------------|--------|----------|
| [Name]           | [file:line] | [entrypoint/init file]  | Missing/Stub/Disabled | Critical/High/Medium/Low |

#### 2. Interface Compliance Gaps
| Interface | Declared In | Implementation | Issue | Severity |
|-----------|------------|----------------|-------|----------|
| [Name]    | [file:line] | [file:line] or None | [stub/missing/incomplete] | [severity] |

#### 3. Integration Wiring Gaps
| Integration | Packages Involved | Issue | Severity |
|-------------|------------------|-------|----------|
| [Name]      | [pkg A] ↔ [pkg B] | [description] | [severity] |

#### 4. Procedural Generation Gaps
| Generator | Package | Issue | Severity |
|-----------|---------|-------|----------|
| [Name]    | [pkg]   | [unused/partial genre support/seed leak] | [severity] |

#### 5. Network Protocol Gaps
| Component | Package | Issue | Severity |
|-----------|---------|-------|----------|
| [Name]    | [pkg]   | [unhandled packet/missing reconciliation] | [severity] |

#### 6. Dead Code / TODO Gaps
| Location | Type | Detail | Severity |
|----------|------|--------|----------|
| [file:line] | [TODO/dead code/empty body] | [description] | [severity] |

#### 7. Documentation-to-Code Gaps
| Document | Claim | Code Location | Issue | Severity |
|----------|-------|---------------|-------|----------|
| [doc file] | [quoted claim] | [file:line] or None | [missing/mismatch] | [severity] |

### Prioritized Recommendations
1. **[Critical]** [Actionable fix with file reference]
2. **[High]** [Actionable fix with file reference]
...
```

ANALYSIS INSTRUCTIONS:
- Work package-by-package through `pkg/` directories, then `cmd/` entrypoints
- Use `grep -r` patterns to locate struct definitions, interface assertions, TODO markers
- Trace initialization call chains from `main()` through to system registration
- For each gap found, verify it is a real gap (not gated behind a legitimate feature flag)
- Assign severity based on impact: Critical = blocks core gameplay, High = degrades player experience, Medium = incomplete feature, Low = cleanup/cosmetic
- Report only verified findings with file paths and line numbers

CONSTRAINTS:
- Exclude `_test.go` files from runtime analysis (but note when something exists only in tests)
- Exclude `examples/` and `docs/` from gap detection in categories 1–6 (they are reference material). For category 7 (Documentation-to-Code Gaps), `docs/` files serve as input sources to cross-reference against code — but audit/meta docs (like this file) are not themselves treated as missing implementations
- Do not report style issues, naming conventions, or formatting concerns
- Focus on functional gaps that affect runtime behavior
- If the audit exceeds response limits, keep the total response under 4000 tokens, prioritize Critical and High severity gaps, and note that lower-severity items were truncated

SUCCESS CRITERIA:
✓ All 7 gap categories audited with findings or explicit "no gaps found"
✓ Every finding includes file path and line number
✓ Severity ratings are consistent and justified
✓ Recommendations are actionable (specific files to modify, methods to wire up)
✓ Zero false positives — every reported gap is verified against the actual code
