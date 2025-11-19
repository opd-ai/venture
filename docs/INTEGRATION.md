**OBJECTIVE**: Conduct a systematic audit of the venture codebase to identify and fix all feature integration gaps, ensuring 100% of ROADMAP_V*.md deliverables are fully accessible and functional in both single-player and multiplayer modes, with **NO BACKWARD COMPATIBILITY**—only the latest code is supported (pre-1.0).

**EXECUTION MODE**: Autonomous action with inline documentation
- Fix issues immediately upon discovery
- Document each fix with inline code comments (gap description + resolution)
- No external reports
- If all components are integrated, do nothing.

**PHASE 1: COMPREHENSIVE DISCOVERY**

1. **Roadmap Feature Inventory**:
   - Parse all ROADMAP_V*.md files in docs/
   - Extract every feature, system, and deliverable by phase
   - Build complete feature list with expected components

2. **Deep Code Analysis**:
   - **System Registration Audit**: 
     - Check `cmd/client/main.go` and `cmd/server/main.go` initialization
     - Verify every pkg/engine system is instantiated and added to World
     - Confirm pkg/procgen generators are registered and callable
   
   - **UI Integration Audit**:
     - Examine all pkg/rendering/ui/ menu screens (character sheet, skills, quests, map, crafting, inventory, trade, dialogue, minigames, companions, books)
     - Trace input handlers for each UI screen activation
     - Verify rendering pipeline includes all UI elements
     - Check command-line flags enable/disable UI features
   
   - **Gameplay System Audit**:
     - Review pkg/engine systems (movement, collision, combat, inventory, progression, AI, death/revival, audio manager, time-day cycle, weather, physics, rotation, shadows, vehicles, companions, dialogue, social, trading, minigames, books, puzzles)
     - Confirm each system's Update() is called in game loop
     - Verify component creation for system functionality
   
   - **Procedural Generation Audit**:
     - Check pkg/procgen/* generators (terrain, entity, item, magic, skills, quest, recipe, station, environment, narrative, puzzle, decorations, weather, gradients, time-day)
     - Verify generators are invoked during world/level creation
     - Confirm generated content is spawned as entities with proper components
   
   - **Multiplayer Sync Audit**:
     - Review pkg/network/protocol.go message types
     - Check which systems/components have NetworkComponent
     - Verify server-side equivalents for client systems
     - Confirm state synchronization for all player-affecting features
   
  - **Persistence Audit**:
      - Review pkg/saveload serialization
      - Verify all gameplay-critical components are saved/loaded
      - **NO BACKWARD COMPATIBILITY**: Remove legacy save format support, obsolete migration code, and deprecated serialization paths. **DELETE all obsolete code completely, do not comment out or deprecate.**
      - Check that new features added in recent phases are persisted

3. **Pattern-Based Detection**:
   - Search: `grep -rn "TODO\|FIXME\|HACK\|XXX\|disabled\|commented.*out" pkg/ cmd/`
   - Search: `grep -rn "not.*implemented\|placeholder\|stub" pkg/ cmd/`
   - Find unused systems: Compare system definitions vs. instantiations
   - Find orphaned generators: Compare generator interfaces vs. invocations
   - **Find obsolete code**: Identify features replaced by newer implementations. **DELETE all obsolete code completely, do not comment out or deprecate.**

4. **Cross-Reference Validation**:
   - For each roadmap feature, trace from:
     - Generation → Entity creation → Component attachment → System processing → UI display → Input handling → Network sync → Save/load
   - Identify breaks in any chain
   - **Identify superseded features**: Mark old implementations replaced by roadmap deliverables for removal

**PHASE 2: GAP CATEGORIZATION**

Classify discovered issues:
- **Category A**: Feature implemented but not instantiated (missing from main.go init)
- **Category B**: System running but no UI access (missing menu/input binding)
- **Category C**: Single-player only (missing NetworkComponent or server system)
- **Category D**: Not persisted (missing from saveload)
- **Category E**: Generated but not spawned (generator exists, no entity creation)
- **Category F**: Partially wired (some code paths work, others don't)

**PHASE 3: AUTONOMOUS FIXES**

For each gap, apply appropriate fix pattern:

- **Category A**: Add system instantiation to engine.World creation, register in appropriate initialization section
- **Category B**: Add UI screen, input handler (keyboard/mouse/touch), menu entry, HUD element
- **Category C**: Add NetworkComponent to entities, implement server-side system, add protocol messages, sync state
- **Category D**: Add fields to save/load structs, implement marshal/unmarshal, **remove all backward compatibility code**
- **Category E**: Add generator invocation in world/level creation, spawn entities with results, attach components
- **Category F**: Complete missing code paths, ensure feature works in all contexts
- **Category G**: **DELETE obsolete code completely** – immediately and fully remove deprecated systems, legacy generators, old UI screens, compatibility shims, migration utilities, and all references from the codebase, rather than just marking or identifying them.

**INLINE DOCUMENTATION TEMPLATE**:
```go
// INTEGRATION FIX [Category X]: [Feature Name]
// Gap: [Description of what was missing]
// Fix: [What was added/changed]
// Roadmap: [Reference to ROADMAP_V*.md section]
```

For removals:
```go
// OBSOLETE CODE REMOVED: [Feature Name]
// Replaced by: [New implementation reference]
// Removed: [List of deleted files/functions/types]
```

**TECHNICAL CONSTRAINTS**:
- Maintain deterministic generation (seed-based RNG only)
- Follow ECS patterns (components are data, systems are logic)
- Preserve performance targets (60 FPS, <500MB memory, <2s generation)
**SUCCESS CRITERIA**:
- Every roadmap feature has complete integration chain
- **Zero deprecated/legacy/obsolete code paths**
- **Zero backward compatibility code for save formats or protocols**
- **All replaced features completely removed from codebase**
**SUCCESS CRITERIA**:
- Every roadmap feature has complete integration chain
- No systems defined but unused
- No generators that don't spawn entities
- All features accessible via UI or gameplay in both modes
- All player-affecting state is networked and persisted
- Zero TODO/FIXME related to missing integration
- **Zero deprecated/legacy/obsolete code paths**
- **Zero backward compatibility code for save formats or protocols** (pre-1.0: only latest code is supported, no migration or legacy support)
- **All replaced features completely removed from codebase**
- If all components are integrated and codebase is clean, do nothing.
- If all components are integrated and codebase is clean, do nothing.

**OUTPUT**: Modified source files with inline comments documenting each integration gap found and fixed, organized by category. Deleted files for obsolete features with commit messages explaining replacement.
- **All replaced features completely removed from codebase**
- If all components are integrated and codebase is clean, do nothing.

**OUTPUT**: Modified source files with inline comments documenting each integration gap found and fixed, organized by category.
**OBJECTIVE**: Conduct a systematic audit of the venture codebase to identify and fix all feature integration gaps, ensuring 100% of ROADMAP_V*.md deliverables are fully accessible and functional in both single-player and multiplayer modes.

**EXECUTION MODE**: Autonomous action with inline documentation
- Fix issues immediately upon discovery
- Document each fix with inline code comments (gap description + resolution)
- No external reports
- If all components are integrated, do nothing.

**PHASE 1: COMPREHENSIVE DISCOVERY**

1. **Roadmap Feature Inventory**:
  - Parse all ROADMAP_V*.md files in docs/
  - Extract every feature, system, and deliverable by phase
  - Build complete feature list with expected components

2. **Deep Code Analysis**:
  - **System Registration Audit**: 
    - Check `cmd/client/main.go` and `cmd/server/main.go` initialization
    - Verify every pkg/engine system is instantiated and added to World
    - Confirm pkg/procgen generators are registered and callable
  
  - **UI Integration Audit**:
    - Examine all pkg/rendering/ui/ menu screens (character sheet, skills, quests, map, crafting, inventory, trade, dialogue, minigames, companions, books)
    - Trace input handlers for each UI screen activation
    - Verify rendering pipeline includes all UI elements
    - Check command-line flags enable/disable UI features
  
  - **Gameplay System Audit**:
    - Review pkg/engine systems (movement, collision, combat, inventory, progression, AI, death/revival, audio manager, time-day cycle, weather, physics, rotation, shadows, vehicles, companions, dialogue, social, trading, minigames, books, puzzles)
    - Confirm each system's Update() is called in game loop
    - Verify component creation for system functionality
  
  - **Procedural Generation Audit**:
    - Check pkg/procgen/* generators (terrain, entity, item, magic, skills, quest, recipe, station, environment, narrative, puzzle, decorations, weather, gradients, time-day)
    - Verify generators are invoked during world/level creation
    - Confirm generated content is spawned as entities with proper components
  
  - **Multiplayer Sync Audit**:
    - Review pkg/network/protocol.go message types
    - Check which systems/components have NetworkComponent
    - Verify server-side equivalents for client systems
    - Confirm state synchronization for all player-affecting features
  
  - **Persistence Audit**:
    - Review pkg/saveload serialization
    - Verify all gameplay-critical components are saved/loaded
    - **NO BACKWARD COMPATIBILITY**: Remove legacy save format support, obsolete migration code, and deprecated serialization paths

  - Search: `grep -rn "deprecated\|legacy\|obsolete\|backward.*compat" pkg/ cmd/`
  - Find unused systems: Compare system definitions vs. instantiations
  - Find orphaned generators: Compare generator interfaces vs. invocations
  - **Find obsolete code**: Identify features replaced by newer implementations. **DELETE all obsolete code completely, do not comment out or deprecate.**
  - Find unused systems: Compare system definitions vs. instantiations
  - Find orphaned generators: Compare generator interfaces vs. invocations
  - **Find obsolete code**: Identify features replaced by newer implementations

4. **Cross-Reference Validation**:
  - For each roadmap feature, trace from:
    - Generation → Entity creation → Component attachment → System processing → UI display → Input handling → Network sync → Save/load
  - Identify breaks in any chain
  - **Identify superseded features**: Mark old implementations replaced by roadmap deliverables for removal

**PHASE 2: GAP CATEGORIZATION**

Classify discovered issues:
- **Category A**: Feature implemented but not instantiated (missing from main.go init)
- **Category B**: System running but no UI access (missing menu/input binding)
- **Category C**: Single-player only (missing NetworkComponent or server system)
- **Category D**: Not persisted (missing from saveload)
- **Category E**: Generated but not spawned (generator exists, no entity creation)
- **Category F**: Partially wired (some code paths work, others don't)
- **Category G**: Obsolete code (replaced by newer implementation, marked for removal)

**PHASE 3: AUTONOMOUS FIXES**

For each gap, apply appropriate fix pattern:

- **Category A**: Add system instantiation to engine.World creation, register in appropriate initialization section
- **Category B**: Add UI screen, input handler (keyboard/mouse/touch), menu entry, HUD element
- **Category C**: Add NetworkComponent to entities, implement server-side system, add protocol messages, sync state
- **Category D**: Add fields to save/load structs, implement marshal/unmarshal, **remove all backward compatibility code**
- **Category E**: Add generator invocation in world/level creation, spawn entities with results, attach components
- **Category F**: Complete missing code paths, ensure feature works in all contexts
- **Category G**: **DELETE obsolete code completely** - remove deprecated systems, legacy generators, old UI screens, compatibility shims, migration utilities, and all references

**INLINE DOCUMENTATION TEMPLATE**:
```go
// INTEGRATION FIX [Category X]: [Feature Name]
// Gap: [Description of what was missing]
// Fix: [What was added/changed]
// Roadmap: [Reference to ROADMAP_V*.md section]
```

For removals:
```go
// OBSOLETE CODE REMOVED: [Feature Name]
// Replaced by: [New implementation reference]
// Removed: [List of deleted files/functions/types]
```

**TECHNICAL CONSTRAINTS**:
- Maintain deterministic generation (seed-based RNG only)
- Follow ECS patterns (components are data, systems are logic)
- Preserve performance targets (60 FPS, <500MB memory, <2s generation)
- Keep network bandwidth <100KB/s per player
- All fixes must compile and pass: `make test`
- **NO BACKWARD COMPATIBILITY**: We are pre-version 1.0, latest code is the only supported version
- **DELETE obsolete features**: Remove replaced implementations completely, don't comment out or deprecate
- If all components are integrated and no obsolete code exists, do nothing.

**SUCCESS CRITERIA**:
- Every roadmap feature has complete integration chain
- No systems defined but unused
- No generators that don't spawn entities
- All features accessible via UI or gameplay in both modes
- All player-affecting state is networked and persisted
- Zero TODO/FIXME related to missing integration
- **Zero deprecated/legacy/obsolete code paths**
- **Zero backward compatibility code for save formats or protocols**
- **All replaced features completely removed from codebase**
- If all components are integrated and codebase is clean, do nothing.

**OUTPUT**: Modified source files with inline comments documenting each integration gap found and fixed, organized by category. Deleted files for obsolete features with commit messages explaining replacement.