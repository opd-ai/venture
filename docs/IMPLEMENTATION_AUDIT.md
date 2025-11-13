**Objective**: Execute a two-phase integration audit: (1) Generate comprehensive report identifying integration gaps across ALL systems in ROADMAP_V3.md and ROADMAP_V4.md, (2) Autonomously implement fixes for discovered issues.

**Execution Mode**: Hybrid - Report generation followed by autonomous correction.

**Phase 1: Audit Report Generation**

**Scope**: Systematically audit integration status for every system across the codebase:

**Procedural Generation** (`pkg/procgen/`):
- V3.0 Content: terrain, entity, item, magic, skills, quest, recipe, station, environment, narrative, puzzle generators
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
- V4.0 Core Systems: 
   - Vehicle system (mounting, durability, combat integration)
   - Companion system (loyalty, progression, skills, AI)
   - Class system (specialization, stat growth, abilities)
   - Expression system (emotes, reactions)
   - Mini-games system (puzzles, challenges)
- Audio: audio manager, synthesis integration

**V4.0 Advanced Features** (verify implementation across systems):
- **Humanoid/Monster Sprite Refinement**: Enhanced anatomical accuracy, animation fluidity, genre-specific style variations
- **Tile Rendering Enhancement**: Texture detail generation, seamless transitions, multi-layer depth effects, parallax scrolling
- **Advanced Lighting**: Soft shadows, colored lighting, bloom effects, ambient occlusion, dynamic light sources
- **Rich Particle Systems**: Weather effects (rain, snow, fog), fluid simulation, environmental interactions
- **Visual Post-Processing**: Screen-space effects (blur, DOF), color grading per genre, vignette, chromatic aberration
- **UI Polish**: Visual hierarchy improvements, smooth transitions, procedural decorations, contextual tooltips
- **Enhanced Audio**: Environmental ambience, dynamic music transitions, positional audio, audio occlusion
- **Performance Optimization**: Frame pacing, memory pooling, asset streaming, LOD systems
- **Advanced Narrative**: Branching dialogue trees, moral choice tracking, consequence systems, lore generation
- **Expanded Crafting**: Multi-step recipes, quality tiers, experimentation mechanics, recipe discovery
- **Dynamic Events**: Random world events, timed challenges, seasonal variations, catastrophic events
- **Achievement System**: Milestone tracking, hidden achievements, progress rewards, statistics
- **Mod Support**: Lua scripting integration, custom content loading, save compatibility
- **Accessibility**: Colorblind modes, scalable UI, audio cues, input remapping, difficulty presets

**Multiplayer Architecture** (`pkg/network/`):
- Core: client-server, protocol, message handlers
- Optimization: client-side prediction, entity interpolation, lag compensation (200-5000ms)
- Synchronization: state sync for all entity types, animated sprites, vehicle mounts, companions, expressions, mini-game state

**Supporting Infrastructure**:
- Audio: synthesis (waveforms), music (composition, dynamic transitions), SFX generation, positional audio
- Persistence: save/load system, world state management, mod data handling
- Platform: mobile touch input, WebAssembly support, controller support
- Utilities: logging, performance profiling, achievement tracking, modding API

**Analysis Method**:
- Search `cmd/client/main.go` and `cmd/server/main.go` for system initialization
- Verify ECS system registration in world setup
- Check component usage in entity creation flows
- Validate network message handlers for all entity types (including V4.0: vehicles, companions, expressions)
- Identify TODO/FIXME/XXX comments in critical paths
- Cross-reference ROADMAP_V4.md completion claims against actual code
- Verify all V4.0 advanced features have corresponding implementations
- Check for stub/placeholder code that lacks full implementation

**Report Output** (write to `AUDIT.md`):
```markdown
# Venture Integration Audit Report
Generated: [timestamp]

## Executive Summary
Status: [COMPLETE | INCOMPLETE]
Systems Audited: [count]
V3.0 Systems: [count ✓/✗]
V4.0 Systems: [count ✓/✗]
V4.0 Advanced Features: [count ✓/✗]
Issues Found: [count]

## Integration Matrix

### V3.0 Systems
| System Category | System Name | Status | Evidence | Issues |
|----------------|-------------|--------|----------|---------|
| Procedural Gen | [system] | [✓/✗/⚠] | [file:func] | [description] |
| Rendering | [system] | [✓/✗/⚠] | [file:func] | [description] |
| Game Systems | [system] | [✓/✗/⚠] | [file:func] | [description] |
| Multiplayer | [system] | [✓/✗/⚠] | [file:func] | [description] |
| Infrastructure | [system] | [✓/✗/⚠] | [file:func] | [description] |

### V4.0 Core Systems
| System Category | System Name | Status | Evidence | Issues |
|----------------|-------------|--------|----------|---------|
| Procedural Gen | Vehicle Generator | [✓/✗/⚠] | [file:func] | [description] |
| Procedural Gen | Companion Generator | [✓/✗/⚠] | [file:func] | [description] |
| Procedural Gen | Book Generator | [✓/✗/⚠] | [file:func] | [description] |
| Game Systems | Vehicle System | [✓/✗/⚠] | [file:func] | [description] |
| Game Systems | Companion System | [✓/✗/⚠] | [file:func] | [description] |
| Game Systems | Class System | [✓/✗/⚠] | [file:func] | [description] |
| Game Systems | Expression System | [✓/✗/⚠] | [file:func] | [description] |
| Game Systems | Mini-games System | [✓/✗/⚠] | [file:func] | [description] |

### V4.0 Advanced Features
| Feature Category | Feature Name | Status | Evidence | Issues |
|-----------------|--------------|--------|----------|---------|
| Rendering | Humanoid/Monster Refinement | [✓/✗/⚠] | [file:func] | [description] |
| Rendering | Tile Enhancement | [✓/✗/⚠] | [file:func] | [description] |
| Rendering | Advanced Lighting | [✓/✗/⚠] | [file:func] | [description] |
| Rendering | Rich Particle Systems | [✓/✗/⚠] | [file:func] | [description] |
| Rendering | Visual Post-Processing | [✓/✗/⚠] | [file:func] | [description] |
| Rendering | UI Polish | [✓/✗/⚠] | [file:func] | [description] |
| Audio | Enhanced Audio | [✓/✗/⚠] | [file:func] | [description] |
| Performance | Optimization Systems | [✓/✗/⚠] | [file:func] | [description] |
| Narrative | Advanced Narrative | [✓/✗/⚠] | [file:func] | [description] |
| Crafting | Expanded Crafting | [✓/✗/⚠] | [file:func] | [description] |
| Events | Dynamic Events | [✓/✗/⚠] | [file:func] | [description] |
| Progression | Achievement System | [✓/✗/⚠] | [file:func] | [description] |
| Extensibility | Mod Support | [✓/✗/⚠] | [file:func] | [description] |
| Accessibility | Accessibility Features | [✓/✗/⚠] | [file:func] | [description] |

## Detailed Findings

### [System Name]
- **Status**: [✓ Integrated / ⚠ Partial / ✗ Missing]
- **Evidence**: [file paths and function names]
- **Integration Points**: [where system connects]
- **Issues**: [specific gaps if any]
- **Impact**: [functionality affected]
- **V4.0 Specific**: [notes on V4.0 feature implementation if applicable]

## Critical Issues (Priority Fixes)
1. **[System]**: [Issue description]
    - Location: [file:line]
    - Required Action: [specific fix needed]
    - Impact: [severity]
    - V4.0 Blocker: [yes/no]

## V4.0 Completion Analysis
- Total V4.0 Features: [count]
- Fully Implemented: [count] ([percentage]%)
- Partially Implemented: [count] ([percentage]%)
- Not Implemented: [count] ([percentage]%)
- Blockers: [list of critical missing features]

## Verification Checklist
- [ ] All V3.0 systems initialized in cmd/client/main.go
- [ ] All V4.0 systems initialized in cmd/client/main.go
- [ ] ECS systems registered in engine world
- [ ] Components used in entity creation
- [ ] Network sync covers all entity types (including vehicles, companions)
- [ ] V4.0 advanced features have active code paths
- [ ] TODO/FIXME count in critical paths: [X]
- [ ] Stub/placeholder implementations identified: [X]

## Recommendations
[Prioritized list of integration improvements, V4.0 features first]
```

**Phase 2: Autonomous Correction**

After generating AUDIT.md, automatically implement fixes for all identified issues:

**Fix Categories**:
1. **Missing System Registration**: Add initialization calls to `cmd/client/main.go` and `cmd/server/main.go`
2. **Inactive ECS Systems**: Register systems in world setup
3. **Component Integration**: Connect components to entity creation flows
4. **Network Sync Gaps**: Add message handlers for unsynced entity types (vehicles, companions, expressions)
5. **Broken Connections**: Restore component-system communication paths
6. **V4.0 Stub Implementations**: Replace placeholder code with full implementations
7. **Advanced Feature Integration**: Connect V4.0 advanced features to core systems

**Fix Verification**:
- Run `go build ./cmd/client` and `go build ./cmd/server` after each fix
- Execute `go test ./...` to ensure no regressions
- Verify V4.0 features are accessible in gameplay
- Update AUDIT.md with "FIXED" status for resolved issues

**Success Criteria**:
- ✓ = Fully integrated (active code paths, proper ECS patterns, network sync operational, V4.0 features functional)
- ⚠ = Partially integrated (code exists but missing connections or incomplete V4.0 implementation)
- ✗ = Not integrated (code present but inactive/disconnected or V4.0 feature missing)

**Deliverables**:
1. `AUDIT.md` - Complete integration audit report covering ALL V3.0 and V4.0 systems
2. Code fixes for all identified integration gaps
3. Verification that all V3.0 and V4.0 systems are operational
4. Updated AUDIT.md showing final status: "COMPLETE - All V3.0 and V4.0 systems integrated"