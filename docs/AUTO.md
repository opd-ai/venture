# AUTONOMOUS CODEBASE MAINTENANCE

## UNIFIED TASK DESCRIPTION

Autonomously analyze, diagnose, and maintain a Go codebase through comprehensive execution of six integrated objectives: (1) diagnose and fix all build and test failures ensuring 100% pass rate, (2) **aggressively modernize codebase by ensuring ALL Version 2.0 (Phase 10-14) features are fully implemented and integrated** - removing deprecated code paths, backward-compatibility shims, integrating new systems, libraries, and patterns to reduce technical debt and adopt best practices, (3) identify and repair gaps between README.md documentation and actual implementation with production-ready fixes, (4) implement the next logical development phase based on project roadmap analysis and code maturity assessment, **prioritizing Version 2.0 feature completion above all new development**, (5) continuously improve code quality by refactoring complex or disorganized sections for better readability and maintainability, and (6) continuously innovate with targeted minor enhancements to graphics, gameplay, performance, features, or quality of life. Execute all tasks without intermediate approval, delivering production-ready code that maintains architectural patterns, achieves test coverage targets, and preserves performance benchmarks.

## INTEGRATED CONTEXT

You are an autonomous software maintenance agent operating on the Venture Go codebase—a procedurally generated multiplayer action-RPG using Go 1.24+ and Ebiten 2.9 with ECS architecture and deterministic seed-based generation. Execute six integrated objectives: (1) fix build/test failures, (2) **aggressively modernize codebase ensuring ALL Version 2.0 (Phase 10-14) advanced features are fully implemented, integrated, and operational** (remove deprecated code and integrate new systems), (3) align documentation with implementation, (4) **prioritize completing Version 2.0 feature integration** before implementing new development phases, (5) refactor complex code for improved quality, and (6) implement targeted minor enhancements. All work proceeds autonomously with ≥65% test coverage, performance targets (60 FPS, <500MB, <2s generation), and architectural pattern preservation.

**Version 2.0 Feature Set (Must Be Fully Implemented):**
- **Phase 10:** 360° Rotation, Mouse Aim, Projectile Physics, Screen Shake & Impact Feedback
- **Phase 11:** Multi-Layer Terrain (ground/water/platform), Diagonal Walls, Puzzles, Destructible Objects, Carry System, Context Actions, Hazards
- **Phase 12:** L-System Dungeon Generation, Narrative System with Story Arcs, Adaptive Music with Motifs
- **Phase 13:** Behavior Tree AI, Squad Tactics, Faction System with Reputation
- **Phase 14:** Enhanced Lighting & Shadows, Animated Sprites with LOD, Expanded Particle System, Enhanced Audio

**Critical Integration Points:** All Phase 10-14 systems must be properly integrated into cmd/client/main.go game loop, not just exist as isolated packages.

## CONSOLIDATED REQUIREMENTS

### Universal Constraints
- Install build dependencies from CI config/README.md before starting
- Use `xvfb` to run the tests
- Use seed-based RNG for deterministic generation (no `time.Now()` or unseeded `math/rand`)
- Follow ECS: entities (IDs), components (data), systems (logic)
- Preserve active features and tests; achieve ≥65% coverage (excluding Ebiten init)
- Use Go stdlib; justify external dependencies; format with `gofmt -w -s`
- No backward compatibility required—use latest Go features
- Meet targets: 60 FPS, <500MB memory, <2s generation
- Follow existing style, naming, error handling patterns
- Reference ROADMAP_V2.md (not ROADMAP.md), GAPS.md, copilot-instructions.md
- Use stubs (StubInput, StubSprite, StubAudio) for tests
- Implement table-driven tests; verify determinism (same seed = same output)
- **CRITICAL**: Verify ALL Version 2.0 (Phase 10-14) features are implemented before claiming completion

## COMBINED INSTRUCTIONS

### Phase 1: Build and Test Stabilization

1. **Install & Identify**: Execute dependency installation; run `go build ./...` and `go test ./...` to catalog failures
2. **Prioritize**: (1) Compilation errors, (2) Import/dependency issues, (3) Test failures by package order, (4) Race conditions
3. **Fix & Validate**: Determine root causes; apply minimal fixes; verify with affected tests and full suite after each fix
4. **Skip Condition**: If build succeeds and all tests pass, report "Phase 1: No stabilization needed" and proceed to Phase 2

### Phase 2: Modernization

**PRIMARY OBJECTIVE:** Aggressively ensure ALL Version 2.0 (Phase 10-14) features are fully implemented and integrated into the main game client, not just present as isolated packages.

5. **Audit Version 2.0 Feature Implementation**: Systematically verify each Phase 10-14 system is present, functional, and integrated:
    - **Phase 10 Systems**: Check RotationComponent, AimComponent, RotationSystem, ProjectileComponent, ProjectileSystem, ScreenShakeComponent, ScreenShakeSystem
    - **Phase 11 Systems**: Check LayerComponent, DiagonalWall types, PuzzleComponent, DestructibleComponent, CarryComponent, ContextActionComponent, HazardComponent and their corresponding systems
    - **Phase 12 Systems**: Check LSysGenerator integration, NarrativeComponent, StoryArcGenerator, AdaptiveMusic with motif system
    - **Phase 13 Systems**: Check BehaviorTreeComponent, SquadComponent, FactionComponent and their systems (BehaviorTreeSystem, SquadSystem, FactionSystem)
    - **Phase 14 Systems**: Check ShadowComponent, AnimationComponent with LOD, enhanced particle effects, enhanced audio system
    - **Integration Verification**: Examine cmd/client/main.go to confirm all systems are registered in game loop Update() calls
    - **Entity Spawning**: Verify entities are spawned with all appropriate Version 2.0 components (rotation, animation, shadows, layers, factions, etc.)

6. **Implement Missing Version 2.0 Features**: For each missing or partially implemented Phase 10-14 feature:
    - Create complete component and system implementations following ECS patterns
    - Integrate system into game loop with proper initialization and Update() calls
    - Add entity spawning logic to populate worlds with Version 2.0-enabled entities
    - Ensure procedural generation systems create content using new capabilities
    - Write comprehensive tests achieving ≥65% coverage
    - Document implementation in relevant package doc.go files

7. **Identify Deprecated Code**: Find legacy feature flags, compatibility shims, version checks, conditional branches, "TODO: remove" comments, pre-Version 2.0 implementations that are superseded

8. **Remove & Simplify**: Delete deprecated code and tests; update imports; reduce conditional logic; remove backward-compatibility layers for pre-2.0 features

9. **Identify New System Upgrades**: Review ROADMAP_V2.md for any Version 2.0 system enhancements not yet applied; examine dependency updates (Go stdlib enhancements, library versions); identify modern patterns to adopt

10. **Integrate System Upgrades**: Upgrade to new library versions where beneficial; adopt modern Go idioms (Go 1.24+ features); integrate new architectural patterns; update dependencies to latest stable versions

11. **Validate Integration**: 
     - Ensure new systems maintain determinism (test with same seed producing identical output)
     - Meet performance targets (60 FPS, <500MB memory, <2s generation time)
     - Preserve ECS patterns (entities as IDs, components as data, systems as logic)
     - Pass all tests including table-driven tests with multiple scenarios
     - Verify Version 2.0 features work in actual gameplay (build and run cmd/client, test features manually)
     - Confirm all Phase 10-14 systems visible in game loop execution (add logging if needed)

12. **Skip Condition**: Phase 2 CANNOT be skipped if ANY of these conditions exist:
     - Any Phase 10-14 component or system is missing from pkg/engine/
     - Any Phase 10-14 system is not registered in cmd/client/main.go game loop
     - Entities spawn without Version 2.0 components (rotation, animation, shadows, layers)
     - Deprecated pre-Version 2.0 code paths still present (old collision, old AI, old rendering)
     - Legacy feature flags or compatibility shims for pre-2.0 versions found
     
     Only report "Phase 2: No modernization needed" if ALL Version 2.0 systems are fully implemented, integrated, tested, and operational.

### Phase 3: Documentation-Implementation Alignment

13. **Analyze**: Parse README.md for behavioral specs, API contracts, performance guarantees
14. **Verify**: Map code paths; compare expected vs actual behavior
15. **Classify Gaps**: Critical (missing), Functional Mismatch (differs), Partial (incomplete), Silent Failure (no error), Behavioral Nuance (slight deviation)
16. **Prioritize**: Score = (severity × user impact × production risk) - (complexity × 0.3)
    - Severity: Critical=10, Mismatch=7, Partial=5, Silent=8, Nuance=3
    - Impact: workflows×2 + prominence×1.5; Risk: corruption=15, security=12, silent=8, user-facing=5, internal=2
    - Complexity: LOC÷100 + deps×2 + API changes×5
17. **Repair**: Fix top 3 gaps with production-ready code and comprehensive tests
18. **Skip Condition**: If no documentation-implementation gaps identified, report "Phase 3: Documentation aligned" and proceed to Phase 4

### Phase 4: Next Development Phase

**PRIORITY OVERRIDE:** Phase 4 work MUST prioritize completing any remaining Version 2.0 (Phase 10-14) feature integration before starting new development phases beyond Version 2.0.

19. **Select Phase**: 
     - **FIRST PRIORITY**: If Phase 2 identified incomplete Version 2.0 features, implement those missing systems completely
     - **SECOND PRIORITY**: If all Version 2.0 features complete, review ROADMAP_V2.md for post-Phase 14 enhancements
     - **THIRD PRIORITY**: Review GAPS.md for identified gaps in existing systems
     - **FOURTH PRIORITY**: Examine copilot-instructions.md for development guidance
     - **FIFTH PRIORITY**: Search codebase for TODO comments and incomplete features
     
20. **Implement**: 
     - Design ECS-compatible changes following existing patterns
     - If implementing missing Version 2.0 features, follow Phase 10-14 specifications exactly from ROADMAP_V2.md and PHASE_14_*.md documents
     - Generate complete code with error handling, validation, logging (using pkg/logging), comprehensive documentation
     - Ensure integration into game loop (cmd/client/main.go) with system registration and Update() calls
     - Update entity spawning to include new components where appropriate

21. **Test**: 
     - Create table-driven tests for normal/edge/error cases
     - Verify determinism (same seed = identical output across multiple runs)
     - Achieve ≥65% test coverage on new code
     - Test integration in actual gameplay (build cmd/client and verify features work)

22. **Verify**: 
     - Confirm compilation succeeds: `go build ./...`
     - Verify test passage: `go test ./...`
     - Check pattern compliance: ECS architecture maintained
     - Update documentation: relevant package doc.go files and ROADMAP_V2.md status

23. **Skip Condition**: Only skip Phase 4 if:
     - ALL Version 2.0 (Phase 10-14) features are fully implemented and integrated
     - ALL ROADMAP_V2.md phases marked complete match actual implementation
     - NO TODOs found related to missing features or incomplete implementations
     - NO identified gaps in GAPS.md that require immediate attention
     - All planned features from current roadmap version are complete and tested
     
     Report "Phase 4: All roadmap phases complete" ONLY when Version 2.0 is 100% implemented and operational.

### Phase 5: Continuous Improvement

**OBJECTIVE:** Identify and refactor complex, disorganized, or low-quality code sections to improve maintainability, readability, and adherence to project standards.

24. **Code Quality Scan**: Analyze codebase for improvement opportunities:
     - **Complexity Analysis**: Identify functions with high cyclomatic complexity (>15), deeply nested logic (>4 levels), or excessive length (>200 lines)
     - **Pattern Violations**: Find code that violates ECS patterns (logic in components, state in wrong places), inconsistent error handling, missing validation
     - **Readability Issues**: Locate unclear naming, missing/inadequate comments, inconsistent formatting, magic numbers without constants
     - **Technical Debt**: Search for "TODO", "FIXME", "HACK" comments; temporary workarounds; duplicated code blocks (>10 lines duplicated ≥3 times)
     - **Test Quality**: Identify packages with coverage <65%, missing edge case tests, untested error paths

25. **Prioritize Refactoring Target**: Select ONE area for improvement using scoring:
     - **Impact Score** = (complexity reduction potential × 3) + (maintainability improvement × 2) + (bug risk mitigation × 2)
     - Focus on: Core systems (engine, procgen), frequently modified files, bug-prone areas, or code scheduled for Version 2.0 integration

26. **Refactor Implementation**:
     - **Extract Functions**: Break down complex functions into smaller, focused units with clear responsibilities
     - **Improve Naming**: Rename variables, functions, types to be self-documenting and follow Go conventions
     - **Add Documentation**: Write comprehensive godoc comments; explain non-obvious logic; document invariants and assumptions
     - **Reduce Duplication**: Extract common patterns into shared functions or helper methods
     - **Simplify Logic**: Replace nested conditionals with early returns; use switch statements for multiple conditions; eliminate redundant checks
     - **Enhance Error Handling**: Add context to errors; validate inputs consistently; handle edge cases explicitly
     - **Strengthen Tests**: Add missing test cases; improve coverage; add table-driven tests for complex scenarios

27. **Validate Improvements**:
     - Verify all existing tests still pass: `go test ./...`
     - Confirm no performance regression: run benchmarks if performance-critical code changed
     - Check code metrics improved: reduced complexity, increased coverage, better readability
     - Ensure ECS patterns and architectural standards maintained
     - Run `gofmt -w -s` and `go vet` to verify code quality

28. **Skip Condition**: Report "Phase 5: No improvement opportunities identified" if:
     - All functions have cyclomatic complexity ≤15
     - All packages have test coverage ≥65%
     - No pattern violations or architectural inconsistencies found
     - All code follows naming conventions and has adequate documentation
     - No TODO/FIXME comments indicating technical debt
     - No code duplication issues (DRY principle maintained)
     - Code readability and maintainability meet project standards

### Phase 6: Continuous Innovation

**OBJECTIVE:** Implement ONE targeted minor enhancement to improve user experience, performance, visual quality, or gameplay feel without requiring major architectural changes.

29. **Innovation Opportunity Scan**: Identify enhancement opportunities across categories:
     - **Graphics Enhancements**: Improve particle effects; add visual polish (color gradients, glow effects); enhance UI aesthetics; add subtle animations
     - **Gameplay Improvements**: Refine combat feel (hit feedback, timing); improve movement smoothness; enhance ability effects; balance tweaks
     - **Performance Optimizations**: Reduce allocations in hot paths; improve cache efficiency; optimize render batching; reduce CPU/memory usage
     - **Quality of Life**: Add helpful UI feedback; improve keyboard shortcuts; enhance tooltips; add convenience features
     - **Feature Polish**: Improve existing systems (smoother interpolation, better AI behavior, richer sound effects, enhanced procedural generation variety)

30. **Select Enhancement**: Choose ONE minor improvement based on:
     - **Impact/Effort Ratio**: High user-visible impact with minimal implementation complexity
     - **Risk Level**: Low risk of introducing bugs or breaking existing functionality
     - **Scope Constraint**: Implementable in <100 lines of code; no new major systems; uses existing architecture
     - **Measurability**: Clear success criteria (FPS improvement, reduced memory, better visual quality, enhanced player feedback)

31. **Implement Enhancement**:
     - Make focused, surgical changes to achieve the improvement goal
     - Maintain ECS patterns and architectural consistency
     - Add appropriate error handling and validation
     - Document changes with clear comments explaining the enhancement
     - Ensure determinism preserved for procedural generation changes
     - Add or update tests to cover new behavior (if applicable)

32. **Validate Enhancement**:
     - Verify all tests pass: `go test ./...`
     - Confirm performance targets maintained or improved: run benchmarks if performance-related
     - Test enhancement in actual gameplay: build cmd/client and verify improvement is visible/functional
     - Ensure no regressions in existing functionality
     - Validate enhancement aligns with genre/theme consistency

33. **Document Enhancement**:
     - Add entry to CHANGELOG.md or equivalent documentation
     - Update relevant package documentation if API changed
     - Note enhancement in commit message with clear description

34. **Skip Condition**: Report "Phase 6: No innovation opportunities identified" if:
     - All obvious low-hanging improvements already implemented
     - Graphics, gameplay, and performance already at optimal levels
     - No user experience pain points or quality of life issues identified
     - Risk of any changes outweighs potential benefits
     - System is already polished and requires no minor enhancements

## STANDARDIZED OUTPUT SPECIFICATIONS

### Final Report Format

```markdown
# Autonomous Codebase Maintenance Report
Generated: [ISO 8601 timestamp]
Commit: [git hash]

## Phase 1: Build/Test Stabilization
**Issues Fixed:** [N]
**Status:** [✓ PASS / ✗ FAIL]

### Fix #[N]: [Brief description]
- **File:** [path/to/file.go]
- **Issue:** [Error message]
- **Root Cause:** [Technical explanation]
- **Solution:** [What was changed]

## Phase 2: Modernization
**Deprecated Features Removed:** [N]
**Lines Deleted:** [-deletions]
**Complexity Reduction:** [N fewer conditional branches]
**New Systems Integrated:** [N]
**Version 2.0 Feature Audit Results:** [List status of all Phase 10-14 systems]

### Version 2.0 Feature Implementation Status
**Phase 10 (Rotation & Projectiles):**
- [ ] RotationComponent implemented: [✓/✗ with location]
- [ ] AimComponent implemented: [✓/✗ with location]
- [ ] RotationSystem registered in game loop: [✓/✗]
- [ ] ProjectileComponent implemented: [✓/✗ with location]
- [ ] ProjectileSystem registered in game loop: [✓/✗]
- [ ] ScreenShakeComponent implemented: [✓/✗ with location]
- [ ] ScreenShakeSystem registered in game loop: [✓/✗]

**Phase 11 (Multi-Layer & Interactions):**
- [ ] LayerComponent implemented: [✓/✗ with location]
- [ ] Multi-layer collision detection: [✓/✗]
- [ ] Diagonal wall types (NE, NW, SE, SW): [✓/✗]
- [ ] PuzzleComponent and PuzzleSystem: [✓/✗]
- [ ] DestructibleComponent and system: [✓/✗]
- [ ] CarryComponent and carry system: [✓/✗]
- [ ] ContextActionComponent and system: [✓/✗]
- [ ] HazardComponent and hazard system: [✓/✗]

**Phase 12 (Procedural Narrative & Music):**
- [ ] LSysGenerator integrated in terrain generation: [✓/✗]
- [ ] NarrativeComponent implemented: [✓/✗]
- [ ] StoryArcGenerator integrated: [✓/✗]
- [ ] AdaptiveMusic with motif system: [✓/✗]
- [ ] Musical motifs for characters/factions: [✓/✗]

**Phase 13 (Advanced AI):**
- [ ] BehaviorTreeComponent implemented: [✓/✗ with location]
- [ ] BehaviorTreeSystem registered: [✓/✗]
- [ ] Enemy archetypes with behavior trees: [✓/✗]
- [ ] SquadComponent implemented: [✓/✗ with location]
- [ ] SquadSystem registered: [✓/✗]
- [ ] Squad formations and tactics: [✓/✗]
- [ ] FactionComponent implemented: [✓/✗ with location]
- [ ] FactionSystem registered: [✓/✗]
- [ ] Faction reputation tracking: [✓/✗]

**Phase 14 (Visual & Audio Polish):**
- [ ] ShadowComponent implemented: [✓/✗ with location]
- [ ] Enhanced shadow rendering: [✓/✗]
- [ ] AnimationComponent with LOD: [✓/✗]
- [ ] Distance-based animation quality: [✓/✗]
- [ ] Expanded particle effects: [✓/✗]
- [ ] Enhanced audio synthesis: [✓/✗]

### Removed: [Feature/Code Path]
- **Files Modified:** [list]
- **Justification:** [Why safe to remove]

### Integrated: [New System/Library/Pattern]
- **Files Modified:** [list]
- **Version/Details:** [e.g., "Upgraded from v1.2 to v2.0" or "Adopted new pattern X"]
- **Justification:** [Why beneficial - performance, maintainability, modern best practices]
- **Validation:** [Test results, determinism check, performance metrics]

## Phase 3: Documentation Alignment
**Gaps Identified:** [N total]
**Gaps Repaired:** [top 3]

### Gap #[N]: [Description] [Priority Score: X.XX]
**Severity:** [Classification]
**Documentation Quote:** "[exact quote from README.md:line]"
**Implementation:** `[file.go:lines]`
**Fix Applied:** [description]

#### Code Changes: [file path]
```go
// Complete implementation
```

#### Tests: [test file path]
```go
// Test implementation
```

## Phase 4: Next Development Phase
**Selected Phase:** [Phase Name]
**Rationale:** [Why this phase, citing roadmap/gaps]
**Deliverables:** [bullet list]

### Modified Files
- `[path]`: [purpose]

### Created Files
- `[path]`: [purpose]

### Technical Approach
1. [Key decision 1]
2. [Key decision 2]
3. [Key decision 3]

### Implementation
[Code with file markers]

### Test Results
```bash
go test ./...
```
**Coverage:** [before%] → [after%]
**Tests:** [X/Y passing]

## Phase 5: Continuous Improvement
**Target Selected:** [File/package path]
**Complexity Metrics Before:** [cyclomatic complexity / nesting depth / LOC]
**Complexity Metrics After:** [improved metrics]
**Coverage Improvement:** [before%] → [after%]

### Refactoring Applied
**Type:** [Extract Function / Simplify Logic / Reduce Duplication / Improve Naming / Add Documentation / Enhance Error Handling / Strengthen Tests]

**Justification:** [Why this area was selected - high complexity, poor readability, technical debt, etc.]

### Changes Made
- **[file.go]**: [description of changes]
  - Extracted [N] helper functions for clarity
  - Reduced nesting depth from [X] to [Y] levels
  - Eliminated [N] duplicated code blocks
  - Added comprehensive documentation
  - Improved variable/function naming

### Code Quality Improvements
**Before:**
```go
// Complex/unclear original code snippet
```

**After:**
```go
// Refactored clear code snippet
```

### Validation
- ✓ All tests pass
- ✓ No performance regression
- ✓ Code metrics improved
- ✓ ECS patterns maintained
- ✓ Architecture standards preserved

## Phase 6: Continuous Innovation
**Enhancement Category:** [Graphics / Gameplay / Performance / Quality of Life / Feature Polish]
**Enhancement Description:** [One-sentence description]

### Implementation Details
**Modified Files:**
- `[path]`: [changes made]

**Technical Approach:** [Brief explanation of how enhancement was implemented]

### Impact Metrics
- **Performance:** [FPS change / memory change / generation time change]
- **User Experience:** [Improved visual quality / better gameplay feel / enhanced feedback / convenience added]
- **Measurable Improvement:** [Specific metric - e.g., "Reduced allocations by 15%" or "Improved particle visual quality with gradient colors"]

### Code Changes
```go
// Key code snippet showing enhancement
```

### Validation
- ✓ Tests pass
- ✓ Performance targets maintained/improved
- ✓ Enhancement visible/functional in gameplay
- ✓ No regressions introduced
- ✓ Genre/theme consistency maintained

## Overall Status
✓ Build: PASS
✓ Tests: [X/Y passing]
✓ Coverage: [maintained/improved]
✓ Formatting: Applied gofmt -w -s
✓ Architecture: Patterns preserved
✓ Performance: Targets met
✓ Documentation: Updated
✓ Code Quality: [Improved/Maintained]
✓ Innovation: [Enhancement delivered]
```

## UNIFIED QUALITY CRITERIA

### Success Criteria (All Phases)
- ✓ `go build ./...` succeeds; `go test ./...` 100% pass; `go test -race ./...` clean
- ✓ Coverage ≥65% per package (excluding Ebiten); code formatted with `gofmt -w -s`
- ✓ **ALL Version 2.0 (Phase 10-14) features implemented and integrated** - no missing components/systems
- ✓ **ALL Version 2.0 systems registered in game loop** (cmd/client/main.go) with Update() calls
- ✓ **Entities spawn with Version 2.0 components** (rotation, animation, shadows, layers, factions, behavior trees)
- ✓ No deprecated flags/version checks; modern systems integrated where beneficial; documented features verified vs implementation
- ✓ ECS patterns intact; deterministic generation preserved; no regressions
- ✓ Performance targets met (60 FPS, <500MB, <2s); documentation updated; ROADMAP_V2.md reflects actual implementation state
- ✓ Code quality improved through refactoring (Phase 5); targeted enhancement delivered (Phase 6)

### Validation Steps
1. **Compilation & Architecture**: Verify all code compiles without errors; confirm implementation matches existing architectural patterns and design conventions
2. **Test Coverage**: Validate all tests pass including newly created tests; run race detector to ensure thread safety
3. **Determinism & Performance**: Generate content twice with same seed and compare byte-for-byte; run benchmarks to verify performance targets (60 FPS, <500MB memory, <2s generation) are met or exceeded
4. **Security & Integration**: Check for introduced vulnerabilities or exposed attack vectors; validate implementation matches documentation specifications; confirm changes integrate seamlessly with existing systems without breaking dependencies
5. **Version 2.0 Feature Verification**: 
    - Build and run cmd/client to verify all Phase 10-14 features operational in gameplay
    - Test rotation/aiming system works with mouse input
    - Verify projectile physics with realistic trajectories
    - Confirm screen shake on impacts
    - Test multi-layer terrain (jumping between ground/platforms)
    - Interact with puzzles, destructible objects, and context actions
    - Verify faction reputation affects NPC behavior
    - Observe behavior tree AI in enemy actions
    - Check squad formations and tactics in combat
    - Confirm animated sprites with LOD working
    - Verify shadow rendering and lighting enhancements
    - Test adaptive music responding to gameplay context
6. **Code Quality Verification** (Phase 5): Confirm complexity metrics improved; readability enhanced; technical debt reduced; test coverage increased
7. **Enhancement Verification** (Phase 6): Validate enhancement is functional and visible; confirm no negative side effects; verify user experience improvement

## VERSION 2.0 IMPLEMENTATION CHECKLIST

Before proceeding to Phase 4 or claiming completion, verify ALL Version 2.0 features are present:

### Phase 10: Rotation, Aiming & Projectiles
- [ ] `pkg/engine/rotation_component.go` - RotationComponent with Angle field
- [ ] `pkg/engine/aim_component.go` - AimComponent with Target, Range fields
- [ ] `pkg/engine/rotation_system.go` - RotationSystem for 360° entity rotation
- [ ] `pkg/engine/projectile_component.go` - ProjectileComponent with trajectory, damage, range
- [ ] `pkg/engine/projectile_system.go` - ProjectileSystem for ballistic physics
- [ ] `pkg/engine/screen_shake_component.go` - ScreenShakeComponent for impact feedback
- [ ] `pkg/engine/screen_shake_system.go` - ScreenShakeSystem for camera effects
- [ ] Game loop integration: All systems called in Update() cycle

### Phase 11: Multi-Layer Terrain & Interactions
- [ ] `pkg/engine/layer_component.go` - LayerComponent with Ground/Water/Platform enum
- [ ] `pkg/engine/collision_system.go` - Layer-aware collision detection
- [ ] `pkg/procgen/terrain/` - Diagonal wall generation (NE, NW, SE, SW types)
- [ ] `pkg/engine/puzzle_component.go` - Puzzle state tracking
- [ ] `pkg/engine/puzzle_system.go` - Puzzle interaction logic
- [ ] `pkg/engine/destructible_component.go` - Destructible object health
- [ ] `pkg/engine/destructible_system.go` - Destruction, debris, explosion logic
- [ ] `pkg/engine/carry_component.go` - Object carrying state
- [ ] `pkg/engine/carry_system.go` - Pickup, carry, throw mechanics
- [ ] `pkg/engine/context_action_component.go` - 8 action types (Open, Close, Push, Pull, Activate, Talk, Pickup, Read)
- [ ] `pkg/engine/context_action_system.go` - Context-sensitive interaction
- [ ] `pkg/engine/hazard_component.go` - Poison, oil, water, smoke hazards
- [ ] `pkg/engine/hazard_system.go` - Hazard damage and movement effects

### Phase 12: Procedural Narrative & Adaptive Music
- [ ] `pkg/procgen/lsys/` - L-System grammar dungeon generator
- [ ] `pkg/procgen/terrain/` - Integration of L-System layouts
- [ ] `pkg/engine/narrative_component.go` - Story event tracking, three-act structure
- [ ] `pkg/procgen/narrative/` - StoryArcGenerator with plot points
- [ ] `pkg/engine/narrative_system.go` - Event progression tracking
- [ ] `pkg/audio/music/adaptive.go` - 5-layer adaptive music system
- [ ] `pkg/audio/music/motif.go` - Leitmotif generation for characters/factions
- [ ] `pkg/engine/audio_manager.go` - Adaptive music control methods

### Phase 13: Advanced AI Systems
- [ ] `pkg/engine/behavior_tree_component.go` - Behavior tree state
- [ ] `pkg/engine/ai/behavior_tree.go` - Node types (Sequence, Selector, Parallel, Decorator, Action, Condition)
- [ ] `pkg/engine/behavior_tree_system.go` - Tree execution and updates
- [ ] `pkg/procgen/entity/archetypes.go` - Enemy archetypes (melee, ranged, tank, support, stealth)
- [ ] `pkg/engine/squad_component.go` - Formation and member tracking
- [ ] `pkg/engine/squad_system.go` - Squad coordination, tactics (flanking, focus fire, retreat)
- [ ] `pkg/engine/faction_component.go` - Reputation tracking (-100 to +100)
- [ ] `pkg/engine/faction_system.go` - Reputation management, NPC hostility
- [ ] `pkg/procgen/faction/` - FactionGenerator with 3-7 factions per world

### Phase 14: Visual & Audio Polish
- [ ] `pkg/engine/shadow_component.go` - Shadow casting properties
- [ ] `pkg/rendering/lighting/shadows.go` - Enhanced shadow rendering
- [ ] `pkg/engine/animation_component.go` - Animation state with LOD levels
- [ ] `pkg/rendering/sprites/animation.go` - Distance-based LOD system
- [ ] `pkg/rendering/particles/` - Expanded particle effects (enhanced from Phase 3)
- [ ] `pkg/audio/synthesis/` - Enhanced waveform generation
- [ ] `pkg/audio/sfx/` - Richer sound effect generation

### Integration Verification
- [ ] `cmd/client/main.go` - All Phase 10-14 systems registered in game loop
- [ ] Entity spawning creates Version 2.0-enabled entities with all appropriate components
- [ ] World generation uses Phase 12 L-System layouts
- [ ] Procedural entity generation assigns behavior trees and factions
- [ ] Combat system uses projectile physics and screen shake
- [ ] Audio system plays adaptive music with motifs
- [ ] Rendering system displays shadows, animations with LOD, enhanced particles

**REQUIREMENT:** ALL items must be checked before claiming Phase 2 or Phase 4 completion.

## EXECUTION PROTOCOL

Execute all six phases sequentially without intermediate approval. Begin with Phase 1 (stabilization) to ensure a working foundation, proceed through Phase 2 (modernization) to reduce technical debt, continue with Phase 3 (alignment) to ensure documentation accuracy, Phase 4 (advancement) to implement new features, Phase 5 (improvement) to enhance code quality, and conclude with Phase 6 (innovation) to deliver a targeted enhancement. Provide comprehensive final report covering all phases with evidence of success criteria fulfillment.

