# Version 2.0 Feature Parity Audit Guide

**Project:** Venture - Procedural Action RPG  
**Purpose:** Audit methodology to verify all game versions (desktop, mobile, WASM) have Version 2.0 features enabled  
**Instructions:** Perform a fresh audit using this guide - do not assume previous findings

---

## Audit Methodology

This document provides a systematic approach to audit Version 2.0 feature parity across all platforms. Perform each step to gather current state data.

### Platforms to Audit

| Platform | Entry Point | Build Command | Build Tags |
|----------|-------------|---------------|------------|
| **Desktop** (Linux/macOS/Windows) | `cmd/client/main.go` | `go build ./cmd/client` | Check for platform exclusions |
| **WASM** (Web browsers) | Check if uses desktop or separate | `GOOS=js GOARCH=wasm go build ./cmd/client` | Check WASM-specific tags |
| **Mobile** (Android/iOS) | `cmd/mobile/` directory | `ebitenmobile bind ./cmd/mobile` | Check for `android` or `ios` tags |

### Version 2.0 Feature Phases

Version 2.0 consists of the following phases (verify current implementation):

- **Phase 10:** Enhanced Controls & Combat (rotation, projectiles, screen shake, visual feedback)
- **Phase 11:** Advanced Level Design (multi-layer terrain, puzzles, environmental destruction)
- **Phase 12:** Next-Generation Content (L-System layouts, dynamic narratives, adaptive music)
- **Phase 13:** Advanced AI (behavior trees, squad tactics, faction systems)
- **Phase 14:** Visual & Audio Polish (enhanced lighting/shadows, animated sprites, particles)


---

## Step 1: Identify System Registration Locations

### Action Items

1. **Locate desktop client system registrations:**
   ```bash
   grep -n "game.World.AddSystem|World.AddSystem" cmd/client/main.go
   ```
   - Count total `AddSystem` calls
   - Note line numbers for reference
   - Identify which systems are registered

2. **Locate mobile client initialization:**
   ```bash
   cat cmd/mobile/mobile.go
   ```
   - Check if `AddSystem` calls are present
   - Verify what initialization occurs
   - Compare with desktop approach

3. **Check WASM build configuration:**
   ```bash
   head -5 cmd/client/main.go  # Check build tags
   ```
   - Determine if WASM uses desktop client code
   - Check for WASM-specific exclusions

### Expected Systems (Reference List)

The following systems are typically present in Version 2.0 implementations. Verify actual presence in code:

**Core Gameplay:**
- InputSystem
- CameraSystem (Phase 10.3)
- RotationSystem (Phase 10.1)
- PlayerCombatSystem
- PlayerItemUseSystem
- PlayerSpellCastingSystem
- MovementSystem
- CollisionSystem
- ProjectileSystem (Phase 10.2)
- CombatSystem
- StatusEffectSystem
- RevivalSystem
- AISystem

**Phase 13: Advanced AI:**
- BehaviorTreeSystem (Phase 13.1)
- SquadSystem (Phase 13.2)
- FactionSystem (Phase 13.3)

**Progression:**
- ProgressionSystem
- SkillProgressionSystem
- VisualFeedbackSystem (Phase 10.3)
- AudioManagerSystem
- ObjectiveTrackerSystem

**Inventory & Economy:**
- ItemPickupSystem
- SpellCastingSystem
- ManaRegenSystem
- InventorySystem
- CommerceSystem
- DialogSystem
- CraftingSystem
- InteractionSystem (Phase 11.2)

**Visual & Environmental:**
- AnimationSystem (Phase 14.2)
- EquipmentVisualSystem
- ParticleSystem (Phase 14.3)
- TutorialSystem
- HelpSystem
- WeatherSystem
- LifetimeSystem
- PuzzleSystem (Phase 11.2)
- FirePropagationSystem (Phase 11.3)
- DestructibleObjectSystem (Phase 11.3)
- CarrySystem (Phase 11.3)
- HazardSystem (Phase 11.3)
- NarrativeSystem (Phase 12.2)
- ShadowSystem (Phase 14.1)
- SpatialSystem (optimization)

---

## Step 2: Compare Platform Implementations

### Desktop Client Audit

**File:** `cmd/client/main.go`

1. Count `AddSystem` calls:
   ```bash
   grep -c "game.World.AddSystem" cmd/client/main.go
   ```

2. List all registered systems:
   ```bash
   grep "game.World.AddSystem" cmd/client/main.go | nl
   ```

3. Check build tags:
   ```bash
   head -5 cmd/client/main.go
   ```

**Record findings:**
- Total systems: ___
- Build tag restrictions: ___
- System registration pattern: ___

### WASM Client Audit

1. Determine WASM entry point:
   ```bash
   grep -r "GOOS=js|GOARCH=wasm" Makefile
   ```

2. Check if WASM uses desktop client:
   - If builds `cmd/client`, inherits all desktop systems
   - If separate entry point, audit independently

**Record findings:**
- Uses desktop client: Yes / No
- Separate WASM code: Yes / No
- Total systems: ___

### Mobile Client Audit

**File:** `cmd/mobile/mobile.go`

1. Check initialization code:
   ```bash
   cat cmd/mobile/mobile.go
   ```

2. Count `AddSystem` calls:
   ```bash
   grep -c "AddSystem" cmd/mobile/mobile.go
   ```

3. Verify game instance creation:
   - Look for `NewEbitenGame` or `NewEbitenGameWithLogger`
   - Check what happens after game creation
   - Identify if systems are registered

**Record findings:**
- Initialization method: ___
- Total systems: ___
- System registration present: Yes / No

---

## Step 3: Feature Parity Analysis

### Platform Comparison Matrix

Create a comparison matrix based on audit findings:

| Feature Category | Desktop | WASM | Mobile | Status |
|-----------------|---------|------|--------|--------|
| Core gameplay (13 systems) | ? | ? | ? | ? |
| Phase 13 AI (3 systems) | ? | ? | ? | ? |
| Progression (5 systems) | ? | ? | ? | ? |
| Inventory (8 systems) | ? | ? | ? | ? |
| Visual/Environmental (14 systems) | ? | ? | ? | ? |
| **Total Systems** | ? | ? | ? | ? |

### Version 2.0 Phase Coverage

| Phase | Desktop | WASM | Mobile | Issues |
|-------|---------|------|--------|--------|
| Phase 10: Enhanced Controls | ? | ? | ? | |
| Phase 11: Advanced Level Design | ? | ? | ? | |
| Phase 12: Next-Gen Content | ? | ? | ? | |
| Phase 13: Advanced AI | ? | ? | ? | |
| Phase 14: Visual & Audio Polish | ? | ? | ? | |

---

## Step 4: Identify Discrepancies

### Gap Analysis

For each platform with missing systems:

1. **List missing systems**
   - System name
   - Associated phase
   - Impact on gameplay

2. **Identify root cause**
   - Missing initialization code?
   - Build tag exclusion?
   - Intentional platform limitation?

3. **Assess player impact**
   - Is game playable without these systems?
   - What features are unavailable?
   - User experience implications

---

## Step 5: Recommendations

Based on audit findings, consider these approaches:

### Option 1: Shared System Initialization

**When to use:** Multiple platforms lack same systems

**Approach:**
- Extract system initialization into shared function
- Create `pkg/engine/system_init.go`
- Implement `InitializeGameSystems(game, seed, genreID, logger)` function
- Call from all platform entry points

**Benefits:**
- Single source of truth
- Automatic parity for new systems
- Easier maintenance

**Implementation:**
```go
// pkg/engine/system_init.go
func InitializeGameSystems(game *EbitenGame, seed int64, genreID string, logger *logrus.Logger) error {
    // Register all systems in dependency order
    // Implementation based on desktop client patterns
    return nil
}
```

### Option 2: Platform-Specific Features

**When to use:** Platforms have different capabilities

**Approach:**
- Document which features work on which platforms
- Implement platform detection
- Disable unsupported features gracefully

**Considerations:**
- User experience differs by platform
- Maintenance complexity increases
- Documentation must be clear

### Option 3: Defer Mobile Implementation

**When to use:** Mobile not production-ready

**Approach:**
- Focus on desktop/WASM completion
- Document mobile as experimental
- Plan mobile feature implementation separately

**Considerations:**
- Set user expectations correctly
- Update README/documentation
- Create mobile roadmap

---

## Step 6: Verification Checklist

After implementing fixes, verify with this checklist:

### Build Verification
- [ ] Desktop builds: `go build ./cmd/client`
- [ ] WASM builds: `make build-wasm`
- [ ] Mobile builds: `make android-aar` or `make ios-framework`
- [ ] No build errors or warnings

### System Registration Verification

For each platform, verify systems are registered:

#### Core Gameplay (13 systems)
- [ ] InputSystem
- [ ] CameraSystem
- [ ] RotationSystem
- [ ] PlayerCombatSystem
- [ ] PlayerItemUseSystem
- [ ] PlayerSpellCastingSystem
- [ ] MovementSystem
- [ ] CollisionSystem
- [ ] ProjectileSystem
- [ ] CombatSystem
- [ ] StatusEffectSystem
- [ ] RevivalSystem
- [ ] AISystem

#### Phase 13: Advanced AI (3 systems)
- [ ] BehaviorTreeSystem
- [ ] SquadSystem
- [ ] FactionSystem

#### Progression (5 systems)
- [ ] ProgressionSystem
- [ ] SkillProgressionSystem
- [ ] VisualFeedbackSystem
- [ ] AudioManagerSystem
- [ ] ObjectiveTrackerSystem

#### Inventory & Economy (8 systems)
- [ ] ItemPickupSystem
- [ ] SpellCastingSystem
- [ ] ManaRegenSystem
- [ ] InventorySystem
- [ ] CommerceSystem
- [ ] DialogSystem
- [ ] CraftingSystem
- [ ] InteractionSystem

#### Visual & Environmental (14+ systems)
- [ ] AnimationSystem
- [ ] EquipmentVisualSystem
- [ ] ParticleSystem
- [ ] TutorialSystem
- [ ] HelpSystem
- [ ] WeatherSystem
- [ ] LifetimeSystem
- [ ] PuzzleSystem
- [ ] FirePropagationSystem
- [ ] DestructibleObjectSystem
- [ ] CarrySystem
- [ ] HazardSystem
- [ ] NarrativeSystem
- [ ] ShadowSystem
- [ ] SpatialSystem (if applicable)

### Functional Testing

Test each Phase on each platform:

**Phase 10: Enhanced Controls & Combat**
- [ ] 360° rotation responds to input
- [ ] Projectiles spawn and travel
- [ ] Screen shake on impacts
- [ ] Visual feedback displays
- [ ] Mouse/touch aim works correctly

**Phase 11: Advanced Level Design**
- [ ] Multi-layer terrain renders
- [ ] Puzzles are interactive
- [ ] Objects can be destroyed
- [ ] Items can be carried/thrown
- [ ] Fire propagates correctly

**Phase 12: Next-Generation Content**
- [ ] L-System layouts generate
- [ ] Narrative events trigger
- [ ] Music adapts to context
- [ ] Genre-specific content appears

**Phase 13: Advanced AI**
- [ ] Enemies use behavior trees
- [ ] Squad coordination visible
- [ ] Faction reputation affects NPCs
- [ ] Tactical behaviors execute

**Phase 14: Visual & Audio Polish**
- [ ] Lighting effects render (if enabled)
- [ ] Shadows cast correctly
- [ ] Sprite animations play
- [ ] Particle effects appear
- [ ] Positional audio works

### Performance Testing
- [ ] Desktop: 60 FPS maintained
- [ ] WASM: 30+ FPS in browser
- [ ] Mobile: 30+ FPS on target devices
- [ ] Memory: <500MB on all platforms
- [ ] Load time: <5 seconds

---

## Audit Report Template

Use this template to document audit findings:

```markdown
# Version 2.0 Feature Parity Audit Report

**Audit Date:** YYYY-MM-DD
**Auditor:** [Name]
**Commit/Branch:** [hash/branch]

## Summary

- Desktop systems: X/Y registered
- WASM systems: X/Y registered  
- Mobile systems: X/Y registered

## Platform Details

### Desktop
- Entry point: cmd/client/main.go
- Build tags: [tags]
- Systems registered: [list]
- Issues: [description]

### WASM
- Entry point: [path]
- Build configuration: [details]
- Systems registered: [list]
- Issues: [description]

### Mobile
- Entry point: cmd/mobile/mobile.go
- Systems registered: [list]
- Issues: [description]

## Gaps Identified

1. [Platform]: Missing [systems]
   - Impact: [description]
   - Severity: Critical/High/Medium/Low
   - Recommendation: [action]

## Recommendations

[Prioritized list of actions]

## Next Steps

[Specific implementation tasks]
```

---

## Technical Notes

### Build Tags and Platform Selection

Build tags control which code compiles for each platform:

```go
//go:build !android && !ios
// Compiles for desktop and WASM (excludes only mobile)

//go:build android || ios  
// Compiles only for mobile

//go:build js && wasm
// WASM-specific (rarely needed for this project)
```

### System Initialization Dependencies

Systems must be registered in specific order due to dependencies:

1. Input → all systems depend on input
2. Camera → must apply before rendering
3. Rotation → needs latest aim direction
4. Physics → AI needs accurate positions
5. Combat → status effects process combat results
6. Visual systems → need complete game state

### ebitenmobile Architecture

Mobile builds use `ebitenmobile bind` which requires:
- Exported functions (Init, Start, Update)
- No main() function in bound package
- Compatible API for Java/Objective-C interop

This is why mobile uses separate entry point from desktop.

---

## Maintenance

### When to Re-Audit

Perform fresh audit when:
- New systems are added to any platform
- Version numbers change (2.0 → 2.1)
- Platform support added/removed
- Major refactoring of initialization code
- Player reports platform-specific issues

### Keeping This Guide Updated

Update this guide when:
- New phases are added
- System architecture changes
- Build process changes
- New platforms are supported

---

**Document Version:** 2.0  
**Last Updated:** 2025-11-05  
**Purpose:** Audit methodology and checklist - perform fresh audit each time
