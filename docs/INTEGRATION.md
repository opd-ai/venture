# Venture Integration Audit — Detailed Execution Plan

## Objective
Generate two documents:
1. `docs/INTEGRATION_AUDIT.md` — Comprehensive inventory of all `pkg/` packages with activation status
2. `docs/PLAN.md` — Phased roadmap to integrate dormant packages as always-on baseline features

## Execution Mode
**Report Generation** — Analyze codebase and produce documents. Do not modify source code.

## Core Principle
**All features are baseline features.** Dormant packages become unconditionally enabled upon integration. No feature flags, no toggles, no optional activation. If a package exists, it should be active.

---

## Phase 1: Establish Baseline (Active Systems)

### Step 1.1: Extract Client Imports
Identify all packages currently used by the game client:
```bash
grep -rh "github.com/opd-ai/venture/pkg/" cmd/client/ --include="*.go" | \
  grep -oP '"[^"]*"' | tr -d '"' | sort -u > /tmp/client_imports.txt
```

### Step 1.2: Extract Server Imports
Identify all packages currently used by the game server:
```bash
grep -rh "github.com/opd-ai/venture/pkg/" cmd/server/ --include="*.go" | \
  grep -oP '"[^"]*"' | tr -d '"' | sort -u > /tmp/server_imports.txt
```

### Step 1.3: Identify Registered ECS Systems
Find all systems added to the game world:
```bash
grep -rn "AddSystem\|RegisterSystem\|world\.Add" cmd/client/ cmd/server/ --include="*.go"
```

### Step 1.4: Find Existing Feature Flags (To Remove)
Identify flags gating functionality that should become always-on:
```bash
grep -rn 'enable\|disable\|toggle' cmd/client/ cmd/server/ --include="*.go" | grep -i flag
```
**Action**: These flags should be removed during integration; features become unconditional.

---

## Phase 2: Inventory All Packages

### Step 2.1: Enumerate Package Directories
```bash
find pkg -type d -mindepth 1 -maxdepth 3 | sort > /tmp/all_packages.txt
```

### Step 2.2: For Each Package, Collect Metadata

| Check | Command | Meaning |
|-------|---------|---------|
| Has Go files | `ls pkg/<path>/*.go 2>/dev/null \| wc -l` | Package exists |
| Has doc.go | `test -f pkg/<path>/doc.go` | Documented |
| Has tests | `ls pkg/<path>/*_test.go 2>/dev/null \| wc -l` | Tested |
| Exports public API | `grep -l '^func [A-Z]' pkg/<path>/*.go` | Usable externally |
| Line count | `wc -l pkg/<path>/*.go 2>/dev/null \| tail -1` | Implementation size |

### Step 2.3: Classify Package Completeness
- **Complete**: Has doc.go, tests, exports, >200 LOC
- **Partial**: Has exports but missing docs or tests
- **Stub**: <100 LOC or only type definitions

---

## Phase 3: Cross-Reference (Find Dormant Packages)

### Step 3.1: Compute Set Difference
```bash
comm -23 /tmp/all_packages.txt <(cat /tmp/client_imports.txt /tmp/server_imports.txt | sort -u)
```

### Step 3.2: Check for Indirect Usage
```bash
grep -r "pkg/<dormant_package>" pkg/ --include="*.go" | grep -v "_test.go"
```
**Classify**: 
- "Truly Dormant" = not imported anywhere
- "Indirectly Used" = imported by other pkg/ but not client/server
- "Test Only" = only in *_test.go files

### Step 3.3: Check Example/Demo Usage
```bash
grep -r "pkg/<dormant_package>" examples/ cmd/*test/ --include="*.go"
```
**Note**: Demos prove functionality works — activation should be straightforward.

---

## Phase 4: Dependency Analysis

### Step 4.1: Build Dependency Graph
```bash
go list -f '{{.ImportPath}}: {{join .Imports ", "}}' ./pkg/<path>/...
```

### Step 4.2: Identify Activation Order
If pkg/A imports pkg/B, then B must be integrated before A.
Build topological order for integration phases.

### Step 4.3: Map to Integration Points
```bash
grep -rn "Component\|System" pkg/<path>/*.go | head -20
```
- Defines Components → Entity integration
- Defines Systems → World.AddSystem() unconditionally
- Defines Generators → Procgen pipeline integration

---

## Phase 5: Determine Integration Steps

### Step 5.1: Categorize Integration Type

| Integration Type | Detection | Integration Pattern |
|------------------|-----------|---------------------|
| **ECS System** | Has `Update(dt)` | Unconditional `AddSystem()` |
| **Procgen Generator** | Implements `Generator` | Add to generation pipeline |
| **Rendering Layer** | Imports `ebiten` | Add to render loop |
| **Network Protocol** | Imports `network/` | Register handlers at startup |
| **UI Component** | In `rendering/ui/` | Add to UI initialization |
| **Audio** | In `audio/` | Connect to AudioManager |
| **Save/Load** | Has serialization | Register with SaveLoad |

### Step 5.2: Write Concrete Integration Steps
For each dormant package, specify exact changes:
```
Example for `pkg/companion/learning/`:
1. Add import in cmd/client/main.go
2. Create LearningSystem: `learningSys := learning.NewSystem(companionSys)`
3. Register unconditionally: `world.AddSystem(learningSys)`
```
**No feature flags. No conditionals. Just add it.**

### Step 5.3: Estimate Effort
- **Small (S)**: Import + 1-2 line registration
- **Medium (M)**: New initialization code or UI wiring
- **Large (L)**: Multiple system integrations

---

## Phase 6: Review Recent Changes

### Step 6.1: Check Roadmap Context
```bash
cat docs/ROADMAP*.md | grep -A5 -B5 "Phase\|TODO\|WIP\|Complete"
```

### Step 6.2: Git History
```bash
git log --since="2 months ago" --name-only --pretty=format: -- pkg/ | sort | uniq -c | sort -rn | head -30
```
Recently modified = likely being activated already.

---

## Phase 7: Generate Output Documents

### Output 1: `docs/INTEGRATION_AUDIT.md`

```markdown
# Integration Audit - December 2025

## Summary
- **Total pkg/ directories**: X
- **Active**: Y  
- **Dormant (requires integration)**: Z

## Active Packages
| Package | Used By | Purpose |
|---------|---------|---------|
| `pkg/engine` | client, server | Core ECS |
| ... | ... | ... |

## Dormant Packages (To Integrate)

### `pkg/companion/learning`
- **Completeness**: Complete
- **Blocker**: Not registered as system
- **Dependencies**: `pkg/companion` (active)
- **Integration**:
  1. Import in `cmd/client/main.go`
  2. `world.AddSystem(learning.NewSystem(companionSys))`
- **Effort**: Small

### `pkg/engine/physics/destruction`
...

## Dependency Graph
[Mermaid diagram showing integration order]
```

### Output 2: `docs/PLAN.md`

```markdown
# Integration Plan - December 2025

## Principle
All packages become unconditionally active. No feature flags. No optional toggles.

## Phase 1: Foundation (Week 1)
Activate core systems that others depend on.

- [ ] `pkg/engine/physics/destruction`
  - File: `cmd/client/main.go`
  - Add: `world.AddSystem(destruction.NewSystem())`
  - Verify: `go run ./cmd/destructiontest`

- [ ] `pkg/class/advanced`
  - File: `cmd/client/main.go`
  - Add: `world.AddSystem(advancedclass.NewSystem())`

## Phase 2: Social Systems (Week 2)
- [ ] `pkg/network/chat`
- [ ] `pkg/social/persistence`
...

## Phase 3: Content Generation (Week 3)
- [ ] `pkg/procgen/companion`
- [ ] `pkg/procgen/dialog`
...

## Verification
After each phase:
- [ ] `go build ./cmd/client && go build ./cmd/server`
- [ ] `go test ./pkg/...`
- [ ] Run client, verify features work
```

---

## Success Criteria
- [ ] All `pkg/` subdirectories inventoried
- [ ] Every dormant package has exact file + code changes
- [ ] Integration order respects dependencies
- [ ] No feature flags in integration steps
- [ ] All features become always-on baseline
- [ ] Excludes `cmd/*test/` (standalone tools, not features)

## Anti-Patterns to Avoid
- ❌ `if enableFeatureX { ... }` — No conditionals
- ❌ `-enable-*` flags — No toggles
- ❌ "Optional" integration — Everything is mandatory
- ✅ Unconditional `AddSystem()` calls
- ✅ Direct imports without guards
- ✅ Always-on initialization