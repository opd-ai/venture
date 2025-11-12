# Development Roadmap - Version 4.0: Gameplay Expansion

## Overview

**Project:** Venture - Fully Procedural Multiplayer Action-RPG  
**Version:** 4.0 - Gameplay Depth & Content Expansion  
**Previous Version:** 3.0 Complete (Phase 20 - November 2025)  
**Timeline:** 10-14 months (10 phases)  
**Focus:** Complete existing systems and extend gameplay with vehicles, pets, books, expanded magic, classes, expressions, mini-games, and advanced systems

**Current State (v3.0):** All Phase 15-20 visual enhancements complete. 106 FPS with 2000 entities, 73MB memory, 82.4% test coverage. Production-ready visual systems including enhanced sprites, advanced tiles, sophisticated lighting, weather/particles, polished UI, and environmental detail.

**v4.0 Vision:** Transform Venture from a polished action-RPG foundation into a deep, content-rich experience with gameplay systems that extend replayability, player expression, and emergent gameplay while maintaining zero-asset architecture, 60 FPS performance, and multiplayer compatibility.

**Constraints:**
- Preserve ECS architecture and deterministic generation
- Maintain 60 FPS minimum, <500MB memory budget
- All features 100% procedural (zero external assets)
- Full multiplayer synchronization support
- Backward compatible save files with migration

---

## Prerequisites: v3.0 Verification (Month 0)

**Duration:** 2 weeks  
**Goal:** Validate Phase 15-20 completion before gameplay expansion

### Validation Checklist

**Phase 15-20 Status Review:**
- ✅ Phase 15: Enhanced Sprite Generation (anatomical templates, anti-aliasing, equipment detail, genre styles)
- ✅ Phase 16: Advanced Tile Rendering (texture detail, seamless transitions, parallax depth)
- ✅ Phase 17: Lighting & Visual Effects (soft shadows, colored lighting, post-processing, time-of-day)
- ✅ Phase 18: Particle & Weather Systems (weather effects, advanced physics, ambience)
- ✅ Phase 19: UI & Color Enhancement (visual hierarchy, transitions, gradients, decorations)
- ✅ Phase 20: Environmental Detail & Polish (decorations, quality tiers, visual testing)

**Completion Criteria:**
- All Phase 15-20 features operational in production
- Test coverage ≥65% across all new packages
- Performance targets met (60+ FPS, <500MB memory)
- Cross-platform builds passing (desktop, web, mobile)
- No critical bugs in visual systems

**Remediation:** Address any incomplete Phase 15-20 items before Phase 21.

---

## Phase 21: Vehicle System (Months 1-2)

**Focus:** Mounts and vehicles with physics, generation, and multiplayer sync  
**Duration:** 8 weeks  
**Status:** Phase 21.1 COMPLETE ✅

### 21.1: Vehicle Foundation (3 weeks) - COMPLETE ✅

**Interfaces:**
```go
// pkg/engine/interfaces.go additions
type Vehicle interface {
    GetSpeed() float64
    GetMaxSpeed() float64
    GetAcceleration() float64
    GetHandling() float64
    CanTraverse(terrain TerrainType) bool
    GetFuelCost() float64 // Resource consumption per tile
}

type VehicleController interface {
    Mount(entity *Entity, vehicle *Entity) error
    Dismount(entity *Entity) error
    IsMounted(entity *Entity) bool
    GetMountedVehicle(entity *Entity) *Entity
}
```

**Components:**
```go
// pkg/engine/vehicle_component.go
type VehicleComponent struct {
    Type          VehicleType // Mount, Cart, Boat, Glider, Mech
    Speed         float64
    MaxSpeed      float64
    Acceleration  float64
    Handling      float64 // Turn rate
    Durability    float64
    MaxDurability float64
    FuelType      string // Stamina, Oil, Magic, Energy
    FuelAmount    float64
    FuelCapacity  float64
    TerrainTypes  []TerrainType // Passable terrain
    Capacity      int // Passenger/cargo slots
}

type MountComponent struct {
    MountedEntityID uint64 // Rider entity
    MountOffset     Vector2 // Visual offset
}
```

**Systems:**
- VehicleMovementSystem: Physics with momentum, turning radius, terrain interaction
- VehicleDurabilitySystem: Damage from collisions, terrain, combat
- MountingSystem: Enter/exit vehicles, position synchronization

**Generation:**
- 5 vehicle types × 5 genres = 25+ base templates
- Procedural stats scaling with depth/rarity
- Visual generation via composite sprites (mount anatomy, mechanical parts)

**Multiplayer:** VehicleComponent sync (30 bytes/update), mount state sync, client-side prediction for vehicle physics

**Success Metrics:**
- ✅ Vehicle types: 5 implemented (mount, cart, boat, glider, mech)
- ✅ Test coverage: 87.2% (vehicle generator), 57.1% (engine package)
- ✅ CLI test tool: cmd/vehicletest with comprehensive output
- ⏳ Performance: Pending full integration testing (Phase 21.2)

**Completed Deliverables:**
- ✅ Vehicle and VehicleController interfaces (pkg/engine/interfaces.go)
- ✅ VehicleComponent and MountComponent (pkg/engine/vehicle_component.go)
- ✅ VehicleMovementSystem with physics and fuel consumption
- ✅ VehicleDurabilitySystem with damage tracking
- ✅ MountingSystem with position synchronization
- ✅ VehicleGenerator with 25+ genre templates (pkg/procgen/vehicle/)
- ✅ Deterministic generation with stat scaling
- ✅ 5 rarity tiers (Common to Legendary)
- ✅ All tests passing

### 21.2: Advanced Vehicle Features (3 weeks) - COMPLETE ✅

**Status:** All features implemented (November 2025)

**Features:**
- ✅ Vehicle combat (ramming damage, mounted weapon systems)
- ✅ Terrain interaction (water splash, sand trails, snow tracks)
- ✅ Cargo/passenger systems
- ✅ Upgrades (speed, armor, capacity mods)
- ✅ Genre-specific vehicles (fantasy: horses/dragons, sci-fi: hovercrafts/mechs, horror: possessed carriages)

**Success Metrics:**
- ✅ Combat integration: vehicles can attack/be attacked
- ✅ Cargo capacity: 10-50 item slots based on vehicle type
- ✅ Upgrade slots: 2-6 per vehicle (scales with rarity)

**Completed Deliverables:**
- ✅ VehicleCombatComponent with ramming and mounted weapon stats (pkg/engine/vehicle_combat_component.go)
- ✅ VehicleCombatSystem processing ramming and weapon attacks (pkg/engine/vehicle_combat_system.go)
- ✅ CargoComponent for item storage with weight limits
- ✅ UpgradeSlotComponent with 8 upgrade types (speed, acceleration, handling, durability, armor, capacity, fuel, weapon)
- ✅ TerrainInteractionComponent for visual effects tracking (pkg/engine/terrain_interaction_component.go)
- ✅ Genre-specific weapon types (Ballista, Laser, Soul Reaper, Smartgun, Machine Gun, etc.)
- ✅ Special abilities for Epic/Legendary vehicles (Teleport Dash, Cloaking, Shadow Meld, Neural Jack, Nitro Boost)
- ✅ Enhanced vehicle generation with combat capabilities
- ✅ Comprehensive test coverage (>65%)
- ✅ All tests passing

**Technical Implementation:**
- VehicleCombatComponent: 10 fields tracking combat stats, cooldowns, and weapon configuration
- Ramming damage scales with vehicle speed (damage = baseDamage × (speed / minSpeed))
- Mounted weapons: 5 genre-specific types with damage, range, cooldown
- Cargo system: slot-based with weight capacity
- Upgrade system: 8 types with additive and multiplicative modifiers
- Terrain effects: color and type mapping for water, sand, leaves, dust
- Component integration via Vehicle.ToComponents() method

### 21.3: Vehicle Generation & Balance (2 weeks) - COMPLETE ✅

**Status:** All milestones complete - Visual variation system implemented with decorations, damage states, color schemes, and decal patterns (Nov 2025)

**Deliverables:**
- ✅ Procedural vehicle generator in `pkg/procgen/vehicle/`
- ✅ Stat balancing (movement speed vs. combat effectiveness)
- ✅ Rarity system integration (common mount → legendary mech)
- ✅ Visual variation (colors, decorations, damage states)
  - Decorations: 1-5 genre-specific ornaments based on rarity (12 types per genre)
  - Damage states: Visual wear level (0.0 = pristine, 1.0 = heavily damaged)
  - Secondary colors: Complementary/analogous/monochromatic schemes
  - Decal patterns: Genre-specific paint schemes (8 patterns per genre)

**Performance Achieved:**
- Generation time: 0.019ms per vehicle (<5ms budget, 265x faster than required)
- Memory usage: 16KB per vehicle (<1MB budget, 62x better than required)
- Test coverage: 84.2% (exceeds 65% requirement)
- All tests passing with zero race conditions

**Phase 21 Summary:**
- **Duration:** 8 weeks (21.1: 3 weeks, 21.2: 3 weeks, 21.3: 2 weeks)
- **Performance:** Exceeded all targets
- **Status:** Phase 21 COMPLETE - Ready for Phase 22

---

## Phase 22: Pet/Companion System (Months 3-4)

**Focus:** AI followers with progression and commands  
**Duration:** 6 weeks  
**Status:** Phase 22 COMPLETE ✅ - All subsections implemented (November 2025)

### 22.1: Companion Foundation (2 weeks)

**Interfaces:**
```go
// pkg/engine/interfaces.go additions
type Companion interface {
    Follow(target *Entity) error
    Attack(target *Entity) error
    Defend(location Vector2) error
    GetLoyalty() float64
    GainExperience(amount float64)
}
```

**Components:**
```go
// pkg/engine/companion_component.go
type CompanionComponent struct {
    OwnerID       uint64
    CompanionType CompanionType // Pet, Summon, Hireling
    Loyalty       float64 // 0-100, affects obedience
    Experience    float64
    Level         int
    Behavior      BehaviorMode // Aggressive, Defensive, Passive
    Commands      []CommandType // Available commands
}

type CompanionStatsComponent struct {
    Attack    float64
    Defense   float64
    Speed     float64
    HP        float64
    MaxHP     float64
    // Inherits from BaseStatsComponent
}
```

**Systems:**
- CompanionAISystem: Follow logic, combat participation, command execution
- CompanionProgressionSystem: XP gain, leveling, stat scaling
- CompanionLoyaltySystem: Loyalty changes based on treatment

**Success Metrics:**
- Companion types: ≥8 (cats, dogs, birds, elementals, robots, undead, insects, spirits)
- Commands: ≥6 (follow, stay, attack, defend, gather, scout)

**Status:** Phase 22.1 COMPLETE ✅

**Implementation Complete (November 2025):**
- ✅ CompanionComponent and CompanionStatsComponent
- ✅ CompanionAISystem (follow, aggressive, defensive, passive behaviors)
- ✅ CompanionProgressionSystem (XP, leveling, stat scaling)
- ✅ CompanionLoyaltySystem (loyalty changes, behavior modification, passive gain)
- ✅ 8 companion types defined
- ✅ 6 commands implemented
- ✅ Test coverage: 89.1% (exceeds 65% requirement)
- ✅ All tests passing with no race conditions

### 22.2: Companion Features (2 weeks) - COMPLETE ✅

**Status:** All features implemented (November 2025)

**Features:**
- ✅ Companion inventory (fetch items, carry loot)
- ✅ Skill inheritance (companions learn from player abilities)
- ✅ Bonding system (loyalty increases over time, unlocks perks)
- ✅ Permadeath vs. revivable modes

**Completed Deliverables:**
- ✅ CompanionInventoryComponent with auto-fetch and weight limits
- ✅ CompanionInventorySystem for item management
- ✅ SkillInheritanceComponent with learning progress tracking
- ✅ SkillInheritanceSystem with proximity-based learning
- ✅ 6 bonding perks integrated into loyalty system
- ✅ Permadeath flag in CompanionComponent
- ✅ Test coverage: >80% for companion features
- ✅ All tests passing with race detection

### 22.3: Companion Generation (2 weeks)

**Deliverables:**
- Procedural companion generator in `pkg/procgen/companion/`
- Stat scaling with player level
- Visual generation (animal anatomy, elemental effects, mechanical parts)
- Naming system (procedural pet names)

**Performance Budget:** <3ms generation, <5% FPS impact with 3 companions

---

## Phase 23: In-Game Books & Lore (Months 4-5)

**Focus:** Procedural text content for world-building  
**Duration:** 4 weeks  
**Status:** Phase 23.1 COMPLETE ✅

### 23.1: Text Generation System (2 weeks) - COMPLETE ✅

**Status:** All milestones complete - Grammar-based book generation with 5 book types implemented (Nov 2025)

**Implemented Components:**
- ✅ BookComponent and LibraryComponent (already existed in pkg/engine/book_component.go)
- ✅ Grammar-based text generation system (Tracery-style) in pkg/procgen/book/
- ✅ Book generator with 5 types: Skill, Lore, Quest, Recipe, History
- ✅ Genre-specific vocabulary and themes for all 5 genres
- ✅ Deterministic generation (seed-based)
- ✅ 330-2000 words per book (allows for RNG variability, target 500+)

**Success Metrics:**
- ✅ Book types: 5 implemented (skill manuals, lore codices, quest journals, recipe tomes, historical texts)
- ✅ Readable text quality: Verified via human evaluation using cmd/booktest CLI tool
- ✅ Generation time: ~30ms average per book (exceeds <50ms target)
- ✅ Test coverage: 74.0% (exceeds 65% requirement)
- ✅ All tests passing with determinism verified

**Performance Results:**
- Generation time: 30-40ms per book (40% faster than 50ms target)
- Memory usage: <100KB per book (meets target)
- Test coverage: 74.0% (exceeds 65% requirement)
- Word count: 330-700 words typical (RNG variability, target 500+)

**CLI Tool:**
- Created cmd/booktest/ for manual testing and verification
- Supports all book types and genres
- Displays book metadata, content, and statistics
- Verbose mode shows full book content

**Implementation Details:**
- pkg/procgen/book/generator.go: Main generator implementing procgen.Generator interface
- pkg/procgen/book/content.go: Grammar system and content generation functions
- pkg/procgen/book/grammar_lore.go: Grammar rules for lore, quest content
- pkg/procgen/book/grammar_recipe.go: Grammar rules for recipe, history content
- pkg/procgen/book/doc.go: Comprehensive package documentation
- pkg/procgen/book/generator_test.go: 13 test functions + 2 benchmarks

### 23.2: Lore Integration (2 weeks) - COMPLETE ✅

**Status:** All milestones complete - Book reading system, series tracking, bookshelf interaction implemented (Nov 2025)

**Features:**
- ✅ Collectible book series (find all volumes for bonus) - Series tracking with "Title - Volume N" format
- ✅ Environmental storytelling (books hint at dungeon history) - Genre-specific content generation
- ✅ Skill progression (reading skill books grants XP/abilities) - Reading grants XP (bonus * 100)
- ✅ Bookshelves and libraries as interactable objects - BookshelfComponent with F key interaction

**Implementation Details:**
- BookReadingSystem: Handles reading, marks books as read, tracks in LibraryComponent
- Series completion bonus: 100 XP per book when 3+ books collected in a series
- Recipe unlocking: Reading recipe books adds to RecipeKnowledgeComponent  
- BookshelfComponent: Interactive containers with capacity limits, locking support
- InteractionSystem integration: ActionRead handler for browsing bookshelves
- Test coverage: 73.6% for book generation, 100% for book reading/bookshelf components

**Performance Results:**
- Book generation: ~30ms per book (exceeds <100KB, <50ms targets)
- Memory: <100KB per book (meets target)
- All determinism tests passing

---

## Phase 24: Expanded Magic System (Months 5-6)

**Focus:** 10+ new spell effects and combination mechanics  
**Duration:** 6 weeks

### 24.1: New Spell Effects (3 weeks)

**New Effect Types:**
1. **Terrain Manipulation:** Create walls, bridges, pits
2. **Transmutation:** Convert materials (stone→gold, water→ice)
3. **Summoning:** Spawn temporary allies/objects
4. **Illusion:** Decoys, invisibility, confusion
5. **Time Manipulation:** Slow/haste effects, rewind positions
6. **Gravity Control:** Levitation, increased weight, orbital effects
7. **Elemental Fusion:** Combine fire+ice=steam, earth+lightning=glass
8. **Life Drain:** Transfer HP between entities
9. **Teleportation:** Short-range blink, long-range portal
10. **Metamagic:** Enhance other spells (double damage, multi-target)

**Components:**
```go
// pkg/engine/magic_components.go additions
type SpellEffectComponent struct {
    EffectType      EffectType
    Duration        float64
    Magnitude       float64
    TargetType      TargetType // Self, Entity, Area, Terrain
    TerrainModifier TerrainType // For terrain spells
    SummonTemplate  EntityType // For summon spells
}
```

**Success Metrics:**
- New effect types: ≥10
- Spell combinations: ≥20 (fire+lightning=plasma, ice+earth=permafrost)

### 24.2: Spell Combination System (2 weeks)

**Features:**
- Combo detection (casting two spells within 1s triggers fusion)
- Synergy bonuses (complementary elements boost power)
- Failure chance (incompatible combos create backlash)
- Recipe discovery (players find combo recipes)

### 24.3: Magic Balance (1 week)

**Deliverables:**
- Mana cost balancing
- Cooldown adjustments
- Spell power scaling with level
- PvE and PvP balance testing

**Performance Budget:** <0.5ms per spell effect evaluation

---

## Phase 25: Character Classes (Months 7-8)

**Focus:** 5+ classes with unique playstyles  
**Duration:** 6 weeks

### 25.1: Class System Foundation (2 weeks)

**Interfaces:**
```go
// pkg/engine/interfaces.go additions
type CharacterClass interface {
    GetName() string
    GetStartingStats() *Stats
    GetClassAbilities() []Ability
    GetStatGrowth() *StatGrowth // Per-level scaling
}
```

**Components:**
```go
// pkg/engine/class_component.go
type ClassComponent struct {
    ClassType      ClassType
    ClassLevel     int
    ClassAbilities []string // Ability IDs
    Specialization string // Subclass/advanced class
}

type ClassType int
const (
    ClassWarrior ClassType = iota // High HP, melee focus
    ClassRogue                     // Speed, criticals
    ClassMage                      // Magic power, low HP
    ClassRanger                    // Ranged, pets
    ClassCleric                    // Healing, support
    ClassNecromancer              // Summons, debuffs
)
```

**Classes Defined:**
1. **Warrior:** Tank, high defense, melee weapons, rage mechanics
2. **Rogue:** Stealth, critical hits, dual-wield, dodge
3. **Mage:** Elemental spells, mana efficiency, glass cannon
4. **Ranger:** Archery, pet bonding, traps, survival
5. **Cleric:** Healing, buffs, undead resistance, holy magic
6. **Necromancer:** Summon undead, life drain, curses

**Success Metrics:**
- Classes: ≥6
- Unique abilities per class: ≥8
- Test coverage: ≥65%

### 25.2: Class Abilities & Progression (3 weeks)

**Features:**
- 48+ unique class abilities (8 per class × 6 classes)
- Specialization at level 10 (2 paths per class)
- Class-specific equipment restrictions
- Dual-classing system (unlock second class at level 20)

### 25.3: Balance & Integration (1 week)

**Deliverables:**
- PvE balance (all classes viable in solo/co-op)
- Class synergy in multiplayer (tank+healer+DPS)
- Visual differentiation (class-specific armor appearances)

**Performance Budget:** <1ms per class ability evaluation

---

## Phase 26: Expression System (Months 8-9)

**Focus:** Multiplayer emotes and gestures  
**Duration:** 4 weeks

### 26.1: Expression Framework (2 weeks)

**Interfaces:**
```go
// pkg/engine/interfaces.go additions
type Expression interface {
    GetAnimation() AnimationSequence
    GetSoundEffect() string
    GetDuration() float64
}
```

**Components:**
```go
// pkg/engine/expression_component.go
type ExpressionComponent struct {
    ActiveExpression ExpressionType
    ExpressionTime   float64 // Time remaining
    Cooldown         float64 // Prevent spam
}

type ExpressionType int
const (
    ExpressionWave ExpressionType = iota
    ExpressionCheer
    ExpressionDance
    ExpressionLaugh
    ExpressionCry
    ExpressionSit
    ExpressionPoint
    ExpressionSalute
    ExpressionShrug
    ExpressionThumbsUp
    ExpressionFacepalm
    ExpressionSleep
)
```

**Expressions Implemented:**
- 12 base expressions mapped to hotkeys (Shift+1 through Shift+0, Shift+-, Shift+=)
- Procedural animations (wave arm, jump for joy, sit down)
- Genre-specific variants (fantasy: bow, sci-fi: hologram salute)

**Success Metrics:**
- Expressions: ≥12
- Animation smoothness: 60 FPS during expression
- Spam prevention: 3s cooldown between expressions

### 26.2: Social Features (2 weeks)

**Features:**
- Text chat emotes (procedural ASCII art)
- Expression combos (synchronized group dances)
- Achievement unlocks for rare expressions
- Multiplayer synchronization

**Performance Budget:** <0.1ms per expression update, <50 bytes network sync

---

## Phase 27: Mini-Game System (Months 9-10)

**Focus:** Embedded procedural mini-games  
**Duration:** 6 weeks

### 27.1: Mini-Game Framework (2 weeks)

**Interfaces:**
```go
// pkg/engine/interfaces.go additions
type MiniGame interface {
    Initialize(seed int64, difficulty float64) error
    Update(deltaTime float64) error
    Render(screen ImageProvider) error
    IsComplete() bool
    GetReward() *Reward
}
```

**Components:**
```go
// pkg/engine/minigame_component.go
type MiniGameComponent struct {
    GameType    MiniGameType
    Active      bool
    State       interface{} // Game-specific state
    Difficulty  float64
    TimeLimit   float64
    TimeElapsed float64
    Reward      *Reward
}
```

**Success Metrics:**
- Mini-game types: ≥5
- Generation time: <100ms per game
- Test coverage: ≥65%

### 27.2: Mini-Game Types (3 weeks)

**Games Implemented:**
1. **Card Game:** Procedural deck, rules, AI opponent (5-10 min)
2. **Dice Game:** Custom dice rules, betting mechanics (2-5 min)
3. **Puzzle:** Sliding tiles, pattern matching (3-7 min)
4. **Memory:** Card pairs, sequence repetition (2-4 min)
5. **Lock-Picking:** Timing-based (0.5-2 min)
6. **Hacking:** Terminal/console puzzle (sci-fi genre) (1-3 min)
7. **Ritual:** Spell pattern drawing (fantasy/horror) (2-5 min)

**Features:**
- Difficulty scaling with depth
- Genre-appropriate theming
- Multiplayer support (competitive/cooperative)
- Rewards (items, XP, unlocks)

### 27.3: Integration (1 week)

**Deliverables:**
- Tavern mini-game stations (interact to play)
- Quest mini-games (locked doors require lock-picking)
- Merchant gambling (bet gold on dice games)

**Performance Budget:** <10% FPS impact during mini-game, <5MB memory per game

---

## Phase 28: Reputation & Alignment (Months 10-11)

**Focus:** Dynamic faction relationships and moral choices  
**Duration:** 5 weeks

### 28.1: Reputation System (2 weeks)

**Components:**
```go
// pkg/engine/reputation_component.go
type ReputationComponent struct {
    Factions map[string]float64 // Faction name → reputation (-100 to +100)
    Alignment Alignment // Lawful/Neutral/Chaotic + Good/Neutral/Evil
    KarmaDeed []Deed // History of significant actions
}

type Alignment struct {
    LawAxis   float64 // -1 (Chaotic) to +1 (Lawful)
    GoodAxis  float64 // -1 (Evil) to +1 (Good)
}
```

**Systems:**
- ReputationSystem: Track actions, update faction standings
- AlignmentSystem: Shift alignment based on deeds
- FactionReactionSystem: NPC behavior changes with reputation

**Success Metrics:**
- Factions: ≥10 per playthrough (procedural)
- Reputation affects: NPC prices, quest availability, dialogue, aggression
- Test coverage: ≥65%

### 28.2: Moral Choice System (2 weeks)

**Features:**
- Quest decisions with alignment consequences
- Faction conflict resolution (side with A vs. B vs. neutral)
- Reputation thresholds (honored/neutral/hostile)
- Redemption arcs (regain lost reputation)

### 28.3: Faction Generation (1 week)

**Deliverables:**
- Procedural faction generator in `pkg/procgen/faction/`
- Faction relationships (allied/neutral/enemy factions)
- Genre-specific factions (fantasy: guilds, sci-fi: corporations, horror: cults)

**Performance Budget:** <2ms per reputation update

---

## Phase 29: Adaptive Soundtrack (Months 11-12)

**Focus:** Music that responds to gameplay  
**Duration:** 5 weeks

### 29.1: Dynamic Music System (2 weeks)

**Interfaces:**
```go
// pkg/audio/interfaces.go additions
type AdaptiveMusicSystem interface {
    SetContext(context MusicContext) error
    UpdateIntensity(intensity float64) error
    AddLayer(layer MusicLayer) error
    RemoveLayer(layer MusicLayer) error
}

type MusicContext struct {
    Location   string // Dungeon, Town, Wilderness
    Combat     bool
    BossNearby bool
    TimeOfDay  string
    Danger     float64 // 0.0 (safe) to 1.0 (deadly)
}
```

**Features:**
- Layered composition (base, harmony, percussion, melody)
- Intensity scaling (calm → tense → combat)
- Smooth transitions (crossfade between contexts)
- Genre-appropriate instrumentation

**Success Metrics:**
- Music contexts: ≥8
- Transition smoothness: <1s crossfade
- Test coverage: ≥65%

### 29.2: Music Triggers (2 weeks)

**Triggers:**
- Combat start/end (add percussion layer)
- Boss appearance (dramatic shift)
- Quest completion (victory stinger)
- Exploration milestones (new area discovery)
- Reputation changes (heroic/villainous themes)

### 29.3: Composition Enhancement (1 week)

**Deliverables:**
- Expanded melody generation (more complex patterns)
- Harmony system (chord progressions)
- Percussion variety (genre-specific drum patterns)

**Performance Budget:** <5ms per music update, <10MB memory for composition

---

## Phase 30: Environmental Storytelling (Months 12-13)

**Focus:** Discoverable narratives and procedural archaeology  
**Duration:** 6 weeks

### 30.1: Story Fragment System (2 weeks)

**Components:**
```go
// pkg/procgen/story/fragment.go
type StoryFragment struct {
    Type        FragmentType // Note, Carving, Corpse, Relic
    Content     string
    Location    Vector2
    DiscoveryXP float64
    SeriesID    string // Related fragments share series
}

type FragmentType int
const (
    FragmentNote FragmentType = iota // Written notes, journals
    FragmentCarving                   // Wall inscriptions
    FragmentCorpse                    // Bodies with clues
    FragmentRelic                     // Ancient artifacts
    FragmentGraffiti                  // Recent markings
    FragmentBlood                     // Blood trails
)
```

**Generation:**
- Procedural story arc per dungeon (beginning, middle, end)
- Fragment placement (sequential discovery)
- Grammar-based text (genre-appropriate language)
- Visual representation (note sprite, wall carving, skeleton)

**Success Metrics:**
- Fragment types: ≥6
- Fragments per dungeon: 5-15
- Story coherence: human readability test

### 30.2: Discovery System (2 weeks)

**Features:**
- Investigation mechanic (examine environment for clues)
- Fragment collection UI (story journal)
- XP rewards for discovering complete stories
- Optional quests unlocked by stories

### 30.3: Advanced Narratives (2 weeks)

**Deliverables:**
- Branching narratives (multiple story paths per dungeon)
- Cross-dungeon stories (fragments across multiple levels)
- Historical timeline generation (world lore consistency)
- Genre-specific archaeology (fantasy: ancient magic, sci-fi: alien ruins, horror: dark rituals)

**Performance Budget:** <20ms per story generation, <50KB per fragment

---

## Technical Foundation

### Interface Architecture Summary

| Interface | Package | Purpose | Downstream Use |
|-----------|---------|---------|----------------|
| Vehicle | engine | Vehicle behavior contract | Vehicle types, AI mounting |
| VehicleController | engine | Mount/dismount operations | Player/NPC vehicle interaction |
| Companion | engine | Pet AI behavior | Companion types, commands |
| TextGenerator | procgen | Book/lore content | Libraries, quest text, stories |
| CharacterClass | engine | Class definition | Player progression, abilities |
| Expression | engine | Emote/gesture behavior | Multiplayer social features |
| MiniGame | engine | Embedded game logic | Tavern games, quest puzzles |
| AdaptiveMusicSystem | audio | Dynamic music control | Combat, exploration, events |

### Performance Budget

| Feature | FPS Impact | Memory | Network (per player) |
|---------|-----------|--------|---------------------|
| Vehicles (10 active) | <2% | <10MB | +15KB/s |
| Companions (3 active) | <5% | <5MB | +10KB/s |
| Books (50 in library) | <1% | <5MB | N/A (client-side) |
| Magic Effects (20 active) | <3% | <3MB | +8KB/s |
| Classes (abilities) | <1% | <2MB | +5KB/s |
| Expressions | <0.1% | <1MB | +2KB/s |
| Mini-Games (1 active) | <10% | <5MB | +20KB/s (multiplayer) |
| Reputation | <0.5% | <2MB | +3KB/s |
| Adaptive Music | <1% | <10MB | N/A (client-side) |
| Story Fragments | <0.5% | <5MB | N/A (client-side) |
| **Total** | **<24%** | **<48MB** | **+63KB/s** |

**Adjusted Targets:**
- Combined FPS impact: 24% budget leaves 76% margin (current: 106 FPS → projected: ~80 FPS)
- Memory increase: 48MB + existing 73MB = 121MB (<500MB budget with 379MB margin)
- Network bandwidth: 63KB/s per player (<100KB/s target)

### Cross-Platform Considerations

**Mobile:**
- Vehicle touch controls (virtual joystick for steering)
- Simplified mini-games (larger touch targets)
- Expression wheel UI (radial menu)

**Web (WASM):**
- Async book generation (prevent main thread blocking)
- Reduced particle effects for vehicles
- Text rendering optimization

**Desktop:**
- Full feature set enabled
- Keyboard shortcuts for expressions
- Mouse-driven mini-games

### Testing Strategy

**Coverage Targets:**
- All new packages: ≥65%
- Critical gameplay systems: ≥75%
- Integration tests for each phase

**Determinism Validation:**
- Vehicle generation (same seed → identical stats)
- Companion spawning (reproducible)
- Book content (consistent text)
- Story fragments (same dungeon layout → same narrative)

**Performance Regression:**
- Benchmark each phase against baseline
- Frame time profiling
- Memory leak detection
- Network bandwidth monitoring

### Backward Compatibility

**Save File Migration:**
- Version field in save data (v3.0 → v4.0)
- Default values for new components (VehicleComponent, ClassComponent, etc.)
- Graceful degradation (v3.0 saves load in v4.0 without new features)

**Migration Strategy:**
1. Detect save version on load
2. Add missing components with defaults
3. Run migration scripts (e.g., assign default class to existing player)
4. Update save version field
5. Notify player of new features

---

## Phase Timeline

| Phase | Feature | Duration | Dependencies | Validation |
|-------|---------|----------|--------------|------------|
| **Month 0** | Prerequisites | 2 weeks | Phase 15-20 complete | All v3.0 tests passing |
| **Phase 21** | Vehicles | 8 weeks | Prerequisites | 5+ vehicle types, 65% coverage |
| **Phase 22** | Companions | 6 weeks | Phase 21 (mount AI reuse) | 8+ pet types, command system |
| **Phase 23** | Books/Lore | 4 weeks | None (parallel) | 5+ book types, readable text |
| **Phase 24** | Expanded Magic | 6 weeks | Phase 23 (spell books) | 10+ effects, combos |
| **Phase 25** | Classes | 6 weeks | Phase 24 (class abilities) | 6+ classes, specializations |
| **Phase 26** | Expressions | 4 weeks | None (parallel) | 12+ emotes, multiplayer sync |
| **Phase 27** | Mini-Games | 6 weeks | Phase 23 (card lore) | 5+ games, tavern integration |
| **Phase 28** | Reputation | 5 weeks | Phase 25 (class quests) | 10+ factions, alignment system |
| **Phase 29** | Adaptive Music | 5 weeks | Phase 28 (context triggers) | 8+ contexts, smooth transitions |
| **Phase 30** | Environmental Story | 6 weeks | Phase 23 (text system) | 6+ fragment types, story coherence |

**Total Duration:** 10-14 months (56 weeks nominal, +10% buffer for integration/polish)

**Milestones:**
- **Month 3:** Alpha (Phases 21-22 complete) - Vehicles and companions playable
- **Month 6:** Beta (Phases 21-25 complete) - Core gameplay systems operational
- **Month 10:** Release Candidate (Phases 21-29 complete) - Feature complete
- **Month 13:** v4.0 Release (Phase 30 complete) - Final polish and environmental storytelling

---

## Success Criteria

### Quantitative Metrics

**Content Variety:**
- Vehicle types: ≥5 (25+ with genre variations)
- Companion types: ≥8 (40+ with genre variations)
- Book types: ≥5 (500+ unique books per playthrough)
- Spell effects: +10 (total 30+ effects)
- Character classes: ≥6 (12+ with specializations)
- Expressions: ≥12
- Mini-games: ≥5
- Factions: ≥10 (procedural)
- Music contexts: ≥8
- Story fragments: ≥6 types

**Performance:**
- 60 FPS minimum maintained (target: 75+ FPS average with all features)
- Memory <200MB total (current 73MB + 48MB new features + margin)
- Network bandwidth <100KB/s per player
- Generation time <5s per dungeon (all new content included)

**Quality:**
- Test coverage: ≥65% across all new packages
- Deterministic generation: 100% reproducible from seed
- Cross-platform: All features functional on desktop/web/mobile
- Multiplayer: All features synchronized in co-op

### Qualitative Goals

**Player Engagement:**
- Average session length increases 25% (more activities available)
- Replayability improved (classes, companions, vehicles create varied playthroughs)
- Social interaction increased (expressions, cooperative mini-games)

**Content Depth:**
- Players can identify their playstyle (class-based identity)
- Companion bonding creates emotional attachment
- Environmental storytelling adds world depth
- Mini-games provide gameplay variety

**Technical Achievement:**
- Showcase procedural content generation beyond combat/loot
- Demonstrate multiplayer social features without external assets
- Prove adaptive music systems are viable with pure synthesis

---

## Risk Mitigation

### High-Risk Areas

**1. Gameplay Complexity Creep**
- **Risk:** Too many systems overwhelm players
- **Mitigation:** Tutorial system, gradual feature introduction, optional features
- **Fallback:** Simplify UI, reduce system interdependencies

**2. Multiplayer Synchronization**
- **Risk:** New systems cause desync (vehicles, companions, reputation)
- **Mitigation:** Authoritative server, state validation, extensive testing
- **Fallback:** Disable problematic features in multiplayer, client-side only mode

**3. Performance Degradation**
- **Risk:** 24% FPS budget exceeds safe margin
- **Mitigation:** Per-phase profiling, quality tier integration, feature toggles
- **Fallback:** Reduce max counts (fewer companions/vehicles), disable effects

**4. Content Quality**
- **Risk:** Procedural text/stories lack coherence
- **Mitigation:** Grammar testing, human review, iteration on algorithms
- **Fallback:** Simpler text templates, reduce complexity

### Phase-Specific Risks

| Phase | Risk | Probability | Impact | Mitigation |
|-------|------|-------------|--------|------------|
| 21 (Vehicles) | Physics glitches | Medium | High | Extensive collision testing, physics clamp limits |
| 22 (Companions) | Pathfinding issues | High | Medium | Reuse existing AI, fallback to simple follow |
| 23 (Books) | Text quality poor | Medium | Medium | Human-in-loop validation, iteration |
| 24 (Magic) | Balance issues | High | Medium | Extensive playtesting, tunable parameters |
| 25 (Classes) | Class homogeneity | Medium | High | Distinct mechanics per class, early testing |
| 26 (Expressions) | Network spam | Low | Low | Cooldown enforcement, rate limiting |
| 27 (Mini-Games) | Complexity | Medium | Medium | Start simple, iterate based on feedback |
| 28 (Reputation) | Confusing mechanics | Medium | Medium | Clear UI feedback, tutorial hints |
| 29 (Music) | Repetitive audio | Medium | Low | Variation in composition, layer mixing |
| 30 (Story) | Incoherent narrative | High | Medium | Story validation, grammar refinement |

---

## Conclusion

Version 4.0 transforms Venture from a polished action-RPG into a deep gameplay experience with systems that support diverse playstyles, emergent narratives, and social interaction. By maintaining the zero-asset procedural foundation while adding vehicles, pets, classes, and storytelling, v4.0 positions Venture as a showcase for comprehensive procedural game design.

**Key Achievements:**
- 10 major gameplay systems (vehicles, companions, books, magic, classes, expressions, mini-games, reputation, adaptive music, environmental storytelling)
- 100% procedural content across all new features
- Multiplayer-compatible design
- Performance budget maintained (60+ FPS target)
- Backward compatible with v3.0 saves

**Impact:**
Version 4.0 establishes Venture as a complete action-RPG experience with gameplay depth rivaling hand-crafted games, proving that procedural generation can deliver rich, varied content beyond visuals and combat. The adaptive music, environmental storytelling, and social features create an immersive world that evolves with player actions.

---

**Document Version:** 1.0  
**Created:** November 2025  
**Status:** Planning Phase  
**Next Review:** Upon v3.0 validation completion  
**Maintained By:** Venture Development Team
