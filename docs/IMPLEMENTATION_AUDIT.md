**Objective**: Execute a two-phase integration audit: (1) Generate comprehensive report identifying integration gaps across ALL systems in ROADMAP_V3.md and ROADMAP_V4.md, (2) Autonomously implement fixes for discovered issues.

**Execution Mode**: Hybrid - Report generation followed by autonomous correction.

**Phase 1: Audit Report Generation**

**Scope**: Systematically audit integration status for every system across the codebase:

**Procedural Generation** (`pkg/procgen/`):
- Content: terrain, entity, item, magic, skills, quest, recipe, station, environment, narrative, puzzle generators
- V4.0 Additions: vehicle, companion, book generators
- Genre: genre registry, blending system

**Rendering Pipeline** (`pkg/rendering/`):
- Core: sprites (anatomy, silhouette, animation, equipment), tiles (patterns, transitions, parallax), particles (weather, physics)
- Visual Effects: lighting (shadows, bloom, AO), post-processing (blur, color grading, vignette), patterns (textures), quality tiers
- UI: inventory, character, skills, quests, map, crafting, shop menus with decorations and gradients
- Optimization: cache system, object pooling

**Game Systems** (`pkg/engine/`):
- Core ECS: movement, collision, combat, inventory, progression, death/revival systems
- AI: behavior trees, squad tactics, faction reputation, companion AI
- V4.0 Systems: vehicle (mounting, durability, combat), companion (loyalty, progression, skills), class (specialization, stat growth), expression, mini-games
- Audio: audio manager, synthesis integration

**Multiplayer Architecture** (`pkg/network/`):
- Core: client-server, protocol, message handlers
- Optimization: client-side prediction, entity interpolation, lag compensation (200-5000ms)
- Synchronization: state sync for all entity types, animated sprites, vehicle mounts, companions

**Supporting Infrastructure**:
- Audio: synthesis (waveforms), music (composition), SFX generation
- Persistence: save/load system, world state management
- Platform: mobile touch input, WebAssembly support
- Utilities: logging, performance profiling

**Analysis Method**:
- Search `cmd/client/main.go` and `cmd/server/main.go` for system initialization
- Verify ECS system registration in world setup
- Check component usage in entity creation flows
- Validate network message handlers for all entity types
- Identify TODO/FIXME/XXX comments in critical paths
- Cross-reference roadmap completion claims against actual code

**Report Output** (write to `AUDIT.md`):
```markdown
# Venture Integration Audit Report
Generated: [timestamp]

## Executive Summary
Status: [COMPLETE | INCOMPLETE]
Systems Audited: [count]
Issues Found: [count]

## Integration Matrix
| System Category | System Name | Status | Evidence | Issues |
|----------------|-------------|--------|----------|---------|
| Procedural Gen | [system] | [✓/✗/⚠] | [file:func] | [description] |
| Rendering | [system] | [✓/✗/⚠] | [file:func] | [description] |
| Game Systems | [system] | [✓/✗/⚠] | [file:func] | [description] |
| Multiplayer | [system] | [✓/✗/⚠] | [file:func] | [description] |
| Infrastructure | [system] | [✓/✗/⚠] | [file:func] | [description] |

## Detailed Findings

### [System Name]
- **Status**: [✓ Integrated / ⚠ Partial / ✗ Missing]
- **Evidence**: [file paths and function names]
- **Integration Points**: [where system connects]
- **Issues**: [specific gaps if any]
- **Impact**: [functionality affected]

## Critical Issues (Priority Fixes)
1. **[System]**: [Issue description]
   - Location: [file:line]
   - Required Action: [specific fix needed]
   - Impact: [severity]

## Verification Checklist
- [ ] All systems initialized in cmd/client/main.go
- [ ] ECS systems registered in engine world
- [ ] Components used in entity creation
- [ ] Network sync covers all entity types
- [ ] TODO/FIXME count in critical paths: [X]

## Recommendations
[Prioritized list of integration improvements]
```

**Phase 2: Autonomous Correction**

After generating AUDIT.md, automatically implement fixes for all identified issues:

**Fix Categories**:
1. **Missing System Registration**: Add initialization calls to `cmd/client/main.go`
2. **Inactive ECS Systems**: Register systems in world setup
3. **Component Integration**: Connect components to entity creation flows
4. **Network Sync Gaps**: Add message handlers for unsynced entity types
5. **Broken Connections**: Restore component-system communication paths

**Fix Verification**:
- Run `go build ./cmd/client` and `go build ./cmd/server` after each fix
- Execute `go test ./...` to ensure no regressions
- Update AUDIT.md with "FIXED" status for resolved issues

**Success Criteria**:
- ✓ = Fully integrated (active code paths, proper ECS patterns, network sync operational)
- ⚠ = Partially integrated (code exists but missing connections)
- ✗ = Not integrated (code present but inactive/disconnected)

**Deliverables**:
1. `AUDIT.md` - Complete integration audit report
2. Code fixes for all identified integration gaps
3. Verification that all systems are operational
4. Updated AUDIT.md showing final status: "COMPLETE - All systems integrated"