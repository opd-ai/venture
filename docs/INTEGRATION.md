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
     - Check that new features added in recent phases are persisted

3. **Pattern-Based Detection**:
   - Search: `grep -rn "TODO\|FIXME\|HACK\|XXX\|disabled\|commented.*out" pkg/ cmd/`
   - Search: `grep -rn "not.*implemented\|placeholder\|stub" pkg/ cmd/`
   - Find unused systems: Compare system definitions vs. instantiations
   - Find orphaned generators: Compare generator interfaces vs. invocations

4. **Cross-Reference Validation**:
   - For each roadmap feature, trace from:
     - Generation → Entity creation → Component attachment → System processing → UI display → Input handling → Network sync → Save/load
   - Identify breaks in any chain

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
- **Category D**: Add fields to save/load structs, implement marshal/unmarshal, handle backward compatibility
- **Category E**: Add generator invocation in world/level creation, spawn entities with results, attach components
- **Category F**: Complete missing code paths, ensure feature works in all contexts

**INLINE DOCUMENTATION TEMPLATE**:
```go
// INTEGRATION FIX [Category X]: [Feature Name]
// Gap: [Description of what was missing]
// Fix: [What was added/changed]
// Roadmap: [Reference to ROADMAP_V*.md section]
```

**TECHNICAL CONSTRAINTS**:
- Maintain deterministic generation (seed-based RNG only)
- Follow ECS patterns (components are data, systems are logic)
- Preserve performance targets (60 FPS, <500MB memory, <2s generation)
- Keep network bandwidth <100KB/s per player
- All fixes must compile and pass: `make test`
- If all components are integrated, do nothing.

**SUCCESS CRITERIA**:
- Every roadmap feature has complete integration chain
- No systems defined but unused
- No generators that don't spawn entities
- All features accessible via UI or gameplay in both modes
- All player-affecting state is networked and persisted
- Zero TODO/FIXME related to missing integration
- If all components are integrated, do nothing.

**OUTPUT**: Modified source files with inline comments documenting each integration gap found and fixed, organized by category.