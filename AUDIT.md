# Implementation Gap Analysis
Generated: 2026-01-22T02:29:56.195Z
Codebase Version: 35fc07098576d0c07334ebb7b3801870284c3468

## Executive Summary
Total Gaps Found: 4
- Critical: 0
- Moderate: 3
- Minor: 1

All findings represent documentation drift where README.md claims do not match the actual implementation. The codebase is mature and functional, but documentation has not been updated to reflect implementation decisions.

---

## Detailed Findings

### Gap #1: Animation Frame Count Mismatch (8 documented vs 4 implemented)
**Severity: Moderate**

**Documentation Reference:** 
> "🖥️ **V7.0 Visual Fidelity** - 1920×1080 display, 64×64 sprites, 8-frame animations, anti-aliased walls, pixel-perfect collision" (README.md:19)

> "✅ **V7.0 Complete** (Phases 43-48): 1920×1080 display, 64×64 sprites, 8-frame animations, anti-aliased walls, pixel-perfect collision" (README.md:40)

**Implementation Location:** 
- `cmd/client/consts.go:24`
- `cmd/client/handlers.go:1914`
- `pkg/engine/entity_spawning.go:214,407`
- `pkg/engine/companion_spawning.go:180`
- `pkg/engine/animation_component.go:224`

**Expected Behavior:** All entity animations should use 8 frames per animation cycle as documented in V7.0 release notes.

**Actual Implementation:** All major game entities use 4-frame animations:
- Player: `playerFrameCount = 4` (consts.go:24)
- Enemies: `enemyAnim.FrameCount = 4` (entity_spawning.go:214,407)
- Merchants: `merchantAnim.FrameCount = 4` (merchant_spawn.go:64)
- Companions: `animComp.FrameCount = 4` (companion_spawning.go:180)
- Default component: `FrameCount: 4` (animation_component.go:224)

**Gap Details:** The README and changelog claim that V7.0 introduced 8-frame animations for smoother visual quality. However, all animation component initializations use a 4-frame count. The animation system supports arbitrary frame counts, but the actual entity implementations have not been updated to use 8 frames.

**Reproduction:**
```go
// In cmd/client/consts.go line 24:
playerFrameCount = 4    // frames per animation sequence

// In pkg/engine/animation_component.go line 224:
FrameCount: 4, // Default 4 frames per animation

// All entity spawning uses FrameCount = 4
```

**Production Impact:** Visual quality is lower than documented. Animations appear less smooth than the 8-frame specification would provide. This is a cosmetic issue that does not affect gameplay.

**Evidence:**
```go
// cmd/client/consts.go:23-24
playerFrameTime    = 0.15 // seconds per animation frame (~6.7 FPS)
playerFrameCount   = 4    // frames per animation sequence

// pkg/engine/animation_component.go:213-228
func NewAnimationComponent(seed int64) *AnimationComponent {
    return &AnimationComponent{
        CurrentState:    AnimationStateIdle,
        PreviousState:   AnimationStateIdle,
        Frames:          nil,
        FrameIndex:      0,
        FrameTime:       1.0 / 12.0, // Phase 15.2: 12 FPS for close range
        TimeAccumulator: 0.0,
        Loop:            true,
        Playing:         true,
        Seed:            seed,
        FrameCount:      4, // Default 4 frames per animation
        Dirty:           true,
        Facing:          DirDown,
    }
}
```

---

### Gap #2: Prestige Class Level Requirement Mismatch (30 documented vs 20 implemented)
**Severity: Moderate**

**Documentation Reference:** 
> "**Skills & Progression**: Unlock abilities, multi-class at level 20, prestige classes at level 30, talent trees" (README.md:162)

**Implementation Location:** 
- `pkg/class/advanced/doc.go:9,40`
- `pkg/class/advanced/constants.go:36`
- `pkg/engine/advanced_class_ui.go:418`

**Expected Behavior:** Prestige classes should unlock at level 30 as specified in the README.

**Actual Implementation:** Prestige classes unlock at level 20.

**Gap Details:** The README states prestige classes unlock at level 30, but all code documentation and UI messages indicate level 20 is the unlock threshold. The internal package documentation in `pkg/class/advanced/doc.go` explicitly states "level 20+", and the UI displays "Prestige classes unlock at level 20".

**Reproduction:**
```go
// Check pkg/class/advanced/doc.go:9
//  2. Prestige Classes: Advanced specializations unlocked at level 20+ with specific requirements

// Check pkg/engine/advanced_class_ui.go:418
text.Draw(screen, fmt.Sprintf("Prestige classes unlock at level 20 (current: %d)", advClass.Level), ...)
```

**Production Impact:** Players may reach level 20 and be surprised they can access prestige classes earlier than documented, or may grind to level 30 unnecessarily if they only read the README. Minor gameplay confusion potential.

**Evidence:**
```go
// pkg/class/advanced/doc.go:9
//  2. Prestige Classes: Advanced specializations unlocked at level 20+ with specific requirements

// pkg/class/advanced/doc.go:40
// 20 prestige classes available at level 20+ with specific requirements:

// pkg/class/advanced/constants.go:36
// PrestigeClassID identifies a prestige class (unlocked at level 20+)

// pkg/engine/advanced_class_ui.go:418
text.Draw(screen, fmt.Sprintf("Prestige classes unlock at level 20 (current: %d)", advClass.Level), basicfont.Face7x13, 70, y, color.RGBA{150, 150, 150, 255})
```

---

### Gap #3: Talent Tree Implementation Incomplete (120 implemented vs implied 450)
**Severity: Moderate**

**Documentation Reference:** 
> "🧠 **Deep Gameplay**: Companion AI with 24-skill trees & personality evolution, branching narratives with 6 endings, multi-classing (15 base + 20 prestige), talent trees (120 talents)" (README.md:49)

> "Each base class has 30 talents organized in 3 categories" (pkg/class/advanced/doc.go:55)

**Implementation Location:** `pkg/class/advanced/talents.go:164,438,708,709`

**Expected Behavior:** With 15 base classes and 30 talents per class, there should be 450 total talents. The README claims 120 talents.

**Actual Implementation:** Only 4 classes have implemented talent trees:
1. Warrior (30 talents)
2. Mage (30 talents)
3. Rogue (30 talents)
4. Cleric (30 talents)

Total: 120 talents implemented.

**Gap Details:** The README accurately states "120 talents" which matches implementation. However, the package documentation claims "Each base class has 30 talents" with 15 base classes, implying 450 total talents. This is a documentation inconsistency between the README and package-level docs.

The 11 remaining base classes (Berserker, Paladin, Knight, Assassin, Ranger, Ninja, Elementalist, Necromancer, Enchanter, Bard, Druid) do not have talent trees implemented. Attempting to get their talent trees returns an error.

**Reproduction:**
```go
// List all talent tree assignments in pkg/class/advanced/talents.go
m.talentTrees[ClassWarrior] = &TalentTree{...}  // line 164
m.talentTrees[ClassMage] = &TalentTree{...}     // line 438
m.talentTrees[ClassRogue] = createRogueTalentTree()   // line 708
m.talentTrees[ClassCleric] = createClericTalentTree() // line 709

// Only 4 classes have talent trees
// 11 remaining classes will return error: "no talent tree for class: X"
```

**Production Impact:** Players choosing any of the 11 unimplemented classes (Berserker, Paladin, Knight, Assassin, Ranger, Ninja, Elementalist, Necromancer, Enchanter, Bard, Druid) cannot access talent trees. The UI will display an error when attempting to view talents for these classes.

**Evidence:**
```go
// pkg/class/advanced/talents.go - Only 4 talent trees are assigned
m.talentTrees[ClassWarrior] = &TalentTree{
    Name:    "Warrior Talents",
    ClassID: ClassWarrior,
    Offensive: []TalentDefinition{...}, // 10 talents
    Defensive: []TalentDefinition{...}, // 10 talents
    Utility: []TalentDefinition{...},   // 10 talents
}

m.talentTrees[ClassMage] = &TalentTree{...}
m.talentTrees[ClassRogue] = createRogueTalentTree()
m.talentTrees[ClassCleric] = createClericTalentTree()

// pkg/class/advanced/manager.go:379-382
func (m *Manager) GetTalentTree(classID ClassID) (*TalentTree, error) {
    tree, exists := m.talentTrees[classID]
    if !exists {
        return nil, fmt.Errorf("no talent tree for class: %s", classID)
    }
    return tree, nil
}
```

---

### Gap #4: Default Max Players Inconsistency (4 in README example vs 8 default)
**Severity: Minor**

**Documentation Reference:** 
> "```bash
> # Start a dedicated server (no graphics, 24/7 hosting)
> ./venture-server -port 8080 -max-players 4
> ```" (README.md:188-189)

**Implementation Location:** `cmd/server/main.go:34`

**Expected Behavior:** The README example suggests 4 as a typical max-players value.

**Actual Implementation:** Default max-players is 8, not 4.

**Gap Details:** The README shows `-max-players 4` in the example, but the actual default is 8. This is not a functional issue, just a documentation example that doesn't match the default behavior. The hostplay package correctly uses 4 as its default (for localhost servers), creating an inconsistency between dedicated server mode and host-and-play mode.

**Reproduction:**
```bash
# Run server without specifying max-players
./venture-server -port 8080
# Server logs will show maxPlayers: 8, not 4
```

**Production Impact:** None functionally. Users who run the server without flags will get 8 player capacity instead of 4. The example in README appears to be intentionally conservative.

**Evidence:**
```go
// cmd/server/main.go:34
maxPlayers = flag.Int("max-players", 8, "Maximum number of players")

// pkg/hostplay/server_manager.go:98-99 (for comparison)
if config.MaxPlayers == 0 {
    config.MaxPlayers = 4
}
```

---

## Summary of Recommended Actions

| Gap # | Issue | Recommended Fix |
|-------|-------|-----------------|
| 1 | 8-frame vs 4-frame animations | Update animation implementations to use 8 frames, OR update README to reflect 4-frame reality |
| 2 | Prestige level 30 vs 20 | Update README line 162 from "level 30" to "level 20" |
| 3 | Incomplete talent trees | Add talent trees for remaining 11 classes, OR add note that only 4 classes have full talent trees |
| 4 | Max players example | Consider updating README example to show default value of 8, or add note about defaults |

## Verification Commands

```bash
# Verify animation frame counts
grep -rn "FrameCount\s*=\s*[0-9]" --include="*.go" | grep -v test

# Verify prestige level requirements
grep -rn "prestige.*20\|20.*prestige" --include="*.go" | grep -v test

# Count implemented talent trees
grep -rn "talentTrees\[Class" --include="*.go"

# Check max-players default
grep -rn "max-players.*default\|maxPlayers.*default" --include="*.go"
```
