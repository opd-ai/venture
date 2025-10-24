# Test Suite Refactoring Analysis

**Date:** October 24, 2025  
**Branch:** interfaces  
**Goal:** Eliminate build tag conflicts through interface-based dependency injection

## Executive Summary

The Venture repository currently uses `//go:build test` and `// +build test` tags to swap production implementations with test stubs. This creates mutual exclusivity where production code and tests cannot be built simultaneously. This analysis documents the current state and outlines the refactoring strategy.

## Current State Analysis

### Build Tag Usage Overview

**Total files with build tags:** 84 files
- **pkg/engine:** 28 stub files
- **examples:** 8 demo programs
- **pkg/procgen:** 2 test files
- **pkg/saveload:** 1 test file

### Critical Issue

The build tags create conflicts:
```bash
# This fails because stub files conflict with production files
go build ./...

# This works but only builds test versions
go test -tags test ./...
```

**Build Failure Example:**
```
pkg/engine/game_test_stub.go:23:23: undefined: TerrainRenderSystem
pkg/engine/game_test_stub.go:25:23: undefined: TutorialSystem
pkg/engine/game_test_stub.go:26:23: undefined: HelpSystem
```

The issue: `game_test_stub.go` references types that are themselves defined with build tags, creating circular dependency issues.

## Type Inventory - Build Tag Stubs

### pkg/engine - Primary Problem Area

| Stub File | Production File | Type(s) | Reason for Stubbing |
|-----------|----------------|---------|---------------------|
| `game_test_stub.go` | `game.go` | `Game` | Ebiten dependency (`ebiten.Game` interface) |
| `components_test_stub.go` | `components.go` | `InputComponent`, `SpriteComponent` | Ebiten types (`*ebiten.Image`, color handling) |
| `render_system_test_stub.go` | `render_system.go` | `RenderSystem` | Ebiten rendering (`*ebiten.Image` drawing) |
| `hud_system_test_stub.go` | `hud_system.go` | `HUDSystem` | Ebiten text/image rendering |
| `menu_system_test_stub.go` | `menu_system.go` | `MenuSystem` | Ebiten UI rendering |
| `character_ui_test_stub.go` | `character_ui.go` | `CharacterUI` | Ebiten UI rendering |
| `skills_ui_test_stub.go` | `skills_ui.go` | `SkillsUI` | Ebiten UI rendering |
| `map_ui_test_stub.go` | `map_ui.go` | `MapUI` | Ebiten UI rendering |
| `ui_systems_test_stub.go` | `ui_systems.go` | `InventoryUI`, `QuestUI` | Ebiten UI rendering |

**Additional Types (defined in test files, not production):**
- `terrain_render_system_test.go` - `TerrainRenderSystem` test stub
- `input_system_test.go` - `InputSystem` test stub (but has production version)
- `tutorial_system_test.go` - `TutorialSystem` test stub
- `help_system_test.go` - Missing stub definition causing build failures

### Dependency Graph

```
Game (stub)
├─ World (no stub needed - pure data)
├─ TerrainRenderSystem (has stub in test file)
├─ TutorialSystem (has stub in test file)  
├─ HelpSystem (MISSING STUB - causes build failure)
├─ MenuSystem (has stub)
├─ CameraSystem (has production version, used in stubs)
├─ RenderSystem (has stub)
├─ HUDSystem (has stub)
└─ UI Systems (all have stubs)
    ├─ InventoryUI
    ├─ QuestUI
    ├─ CharacterUI
    ├─ SkillsUI
    └─ MapUI
```

## Coverage Baseline

**Attempted Test Run:** `go test -tags test -cover ./...`

**Result:** Build failures in pkg/engine prevent full coverage measurement

**Successful Package Coverage:**
- `pkg/audio/music`: 100.0%
- `pkg/audio/sfx`: 85.3%
- `pkg/audio/synthesis`: 94.2%
- `pkg/combat`: 100.0%
- `pkg/network`: 66.1%
- `pkg/procgen`: 100.0%
- `pkg/procgen/entity`: 96.1%
- `pkg/procgen/item`: 94.8%
- `pkg/procgen/magic`: 91.9%
- `pkg/procgen/quest`: 96.6%
- `pkg/procgen/skills`: 90.6%
- `pkg/procgen/terrain`: 97.4%
- `pkg/rendering/palette`: 98.4%
- `pkg/rendering/particles`: 98.0%
- `pkg/rendering/shapes`: 100.0%
- `pkg/rendering/sprites`: 100.0%
- `pkg/rendering/tiles`: 92.6%
- `pkg/rendering/ui`: 88.2%
- `pkg/saveload`: 71.0%
- `pkg/world`: 100.0%

**Failed Packages:**
- `pkg/engine`: Build failed (circular stub dependencies)
- `cmd/client`: Build failed
- `cmd/server`: Build failed
- Several cmd/* and examples/* packages: Build failed

## Refactoring Strategy

### Phase 1: Fix Immediate Build Issues

**Problem:** Circular stub dependencies in pkg/engine

**Solution:** Consolidate all engine stubs into proper test-only files

**Actions:**
1. Create `test_stubs_test.go` with all stub type definitions
2. Remove individual `*_test_stub.go` files
3. Ensure all referenced types have stub definitions

### Phase 2: Extract Core Interfaces

**Target Types for Interface Extraction:**

#### 2.1 Game Loop Interface
```go
// Interface: GameRunner
// Production: EbitenGame (implements ebiten.Game + GameRunner)
// Test: StubGame (implements GameRunner only)
```

Purpose: Separate game loop from Ebiten dependency

#### 2.2 Rendering Interfaces
```go
// Interface: Renderer
// Production: EbitenRenderer (uses *ebiten.Image)
// Test: StubRenderer (no-op or mock rendering)

// Interface: ImageProvider
// Production: EbitenImageProvider
// Test: StubImageProvider
```

Purpose: Abstract away Ebiten rendering types

#### 2.3 Component Interfaces
```go
// Interface: InputProvider
// Production: EbitenInputComponent (reads ebiten.Keys)
// Test: StubInputComponent (controllable state)

// Interface: SpriteProvider
// Production: EbitenSpriteComponent (has *ebiten.Image)
// Test: StubSpriteComponent (no image, just properties)
```

Purpose: Allow components to work without Ebiten types

#### 2.4 UI System Interfaces
```go
// Interface: UIRenderer
// Production: EbitenUIRenderer
// Test: StubUIRenderer

// Apply to: HUDSystem, MenuSystem, CharacterUI, SkillsUI, MapUI, InventoryUI, QuestUI
```

Purpose: Separate UI logic from rendering implementation

### Phase 3: Eliminate Build Tags

**Target:** Zero `//go:build test` tags in production code paths

**Exceptions (documented):**
- Example programs (`examples/*`) - Keep tags for CI/CD (legitimate use case)
- Integration tests that explicitly test Ebiten integration (if any)

### Phase 4: Reorganize Test Structure

**New Structure:**
```
pkg/engine/
├── interfaces.go           # All interfaces
├── game.go                 # Production Game implementation
├── components.go           # Production components
├── render_system.go        # Production rendering
├── ui_systems.go           # Production UI systems
├── game_test.go            # Tests using stub implementations
├── components_test.go      # Component tests
├── stubs_test.go           # All test stubs (automatic test-only via *_test.go)
└── ... (other files)
```

**Benefits:**
- No build tags needed (`*_test.go` is automatically test-only)
- Production and test code can coexist
- Clear separation of concerns

## Risk Assessment

### High Risk
- **Engine package refactoring:** 28 stub files, complex dependency graph
- **Breaking existing tests:** Many tests depend on current stub behavior

### Medium Risk
- **Performance impact:** Interface calls vs direct calls (likely negligible)
- **Increased complexity:** More interfaces to maintain

### Low Risk
- **Coverage regression:** Should maintain or improve with better test isolation
- **Build system:** Standard Go patterns, no custom tooling

## Validation Checklist

- [ ] All production builds succeed: `go build ./...`
- [ ] All tests pass: `go test ./...`
- [ ] No build tags remain: `grep -r "//go:build test" pkg/ | wc -l` = 0
- [ ] Coverage >= baseline
- [ ] `go vet ./...` passes
- [ ] Documentation updated

## Redundant Test Analysis

**To be completed in Phase 1:** Identify duplicate/obsolete tests

**Initial candidates for review:**
- Tests in `*_test_stub.go` files that only test stub behavior (not real functionality)
- Duplicate integration tests across examples/ and pkg/engine/
- Tests for deprecated features (if any)

## Next Steps

1. **Fix immediate build failures** - Consolidate engine stubs
2. **Establish working baseline** - Get `go test -tags test ./...` passing
3. **Document redundant tests** - Create archive strategy
4. **Begin interface extraction** - Start with Game interface
5. **Iterate per component** - One type at a time, commit frequently

## Notes

- The project already has some interface usage (e.g., `Component` interface)
- ECS architecture is well-suited for interface-based design
- Most stubbing is due to Ebiten dependency, not complex business logic
- Consider: Can we use type aliases or wrapper types to reduce interface count?

