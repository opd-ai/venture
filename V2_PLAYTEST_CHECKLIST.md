# V2.0 Playtest Checklist

**Purpose:** Identify which v2.0 features actually work vs. appear broken at runtime.  
**Date:** November 5, 2025  
**Build:** `./venture-client` (latest main branch)

---

## How to Use This Checklist

For each feature:
1. **Test it in-game**
2. **Mark status**: ✅ Works | ❌ Broken | ⚠️ Partial | ❓ Unclear
3. **Add notes** describing what you see (or don't see)

---

## Phase 10.1: 360° Rotation & Mouse Aim

### Player Rotation
- [ ] **Player sprite rotates** to face mouse cursor
  - **How to test:** Move mouse around screen, watch player sprite
  - **Expected:** Player visibly rotates smoothly (~1-2 seconds for 180°)
  - **Status:** 
  - **Notes:** 

- [ ] **Aim direction indicator** (white arrow on HUD)
  - **How to test:** Look for white arrow pointing from player center
  - **Expected:** Arrow always points toward mouse cursor
  - **Status:** 
  - **Notes:** 

- [ ] **Combat uses aim direction**
  - **How to test:** Face mouse left, press Space to attack
  - **Expected:** Attack fires toward mouse, not movement direction
  - **Status:** 
  - **Notes:** 

### Enemy Rotation
- [ ] **Enemies rotate** to face player
  - **How to test:** Let enemy chase you, watch if it rotates
  - **Expected:** Enemy sprite rotates to face you during chase/attack
  - **Status:** 
  - **Notes:** 

- [ ] **Enemy attacks** fire toward player
  - **How to test:** Get hit by enemy, note attack direction
  - **Expected:** Attacks come from enemy's facing direction
  - **Status:** 
  - **Notes:** 

---

## Phase 10.2: Projectile Physics

### Ranged Weapons
- [ ] **Projectiles spawn** from ranged weapons
  - **How to test:** Equip bow/gun, attack, look for moving projectile
  - **Expected:** Visible arrow/bullet flies through air
  - **Status:** 
  - **Notes:** 

- [ ] **Projectiles have physics** (speed, arc, bounce)
  - **How to test:** Fire projectile at wall
  - **Expected:** Projectile bounces or stops at wall
  - **Status:** 
  - **Notes:** 

- [ ] **Projectiles hit enemies**
  - **How to test:** Fire projectile at enemy
  - **Expected:** Enemy takes damage when projectile collides
  - **Status:** 
  - **Notes:** 

### Magic Projectiles
- [ ] **Spells spawn projectiles** (fireball, ice shard, etc.)
  - **How to test:** Press 1-5 to cast spell
  - **Expected:** Visible spell projectile flies toward target
  - **Status:** 
  - **Notes:** 

---

## Phase 10.3: Screen Shake & Impact Feedback

### Screen Shake
- [ ] **Screen shakes when taking damage**
  - **How to test:** Let enemy hit you
  - **Expected:** Camera shakes briefly on impact
  - **Status:** 
  - **Notes:** 

- [ ] **Critical hits cause bigger shake**
  - **How to test:** Deal critical hit to enemy
  - **Expected:** More intense camera shake than normal hit
  - **Status:** 
  - **Notes:** 

### Hit-Stop
- [ ] **Frame pause on strong impacts**
  - **How to test:** Land heavy attack
  - **Expected:** Brief freeze-frame effect (0.1-0.2 seconds)
  - **Status:** 
  - **Notes:** 

### Visual Feedback
- [ ] **Impact particles** on hit
  - **How to test:** Hit enemy, look for particle burst
  - **Expected:** Radial particle burst at impact point
  - **Status:** 
  - **Notes:** 

- [ ] **Damage numbers** float up
  - **How to test:** Deal damage, look for floating numbers
  - **Expected:** Damage value floats up and fades
  - **Status:** 
  - **Notes:** 

- [ ] **Hit flash** on damaged entities
  - **How to test:** Damage enemy, watch sprite
  - **Expected:** Sprite flashes white/red briefly
  - **Status:** 
  - **Notes:** 

---

## Phase 11.1: Diagonal Walls & Multi-Layer Terrain

### Diagonal Walls
- [ ] **Diagonal wall tiles** visible in dungeon
  - **How to test:** Explore dungeon rooms
  - **Expected:** Walls at 45° angles (NE/NW/SE/SW)
  - **Status:** 
  - **Notes:** 

- [ ] **Collision works** on diagonal walls
  - **How to test:** Walk into diagonal wall
  - **Expected:** Player slides along diagonal surface
  - **Status:** 
  - **Notes:** 

### Multi-Layer Terrain
- [ ] **Layer transitions** exist (stairs, ladders, holes)
  - **How to test:** Look for vertical level transitions
  - **Expected:** Objects/areas that change layer when used
  - **Status:** 
  - **Notes:** 

- [ ] **Entities on different layers** interact correctly
  - **How to test:** Attack enemy on different layer
  - **Expected:** Can't hit enemies on different layer
  - **Status:** 
  - **Notes:** 

---

## Phase 11.2: Procedural Puzzle Generation

### Puzzle Elements
- [ ] **Pressure plates** spawn in rooms
  - **How to test:** Look for floor switches in rooms
  - **Expected:** Square colored tiles on floor
  - **Status:** 
  - **Notes:** 

- [ ] **Levers** spawn in rooms
  - **How to test:** Look for wall-mounted switches
  - **Expected:** Vertical lever objects on walls
  - **Status:** 
  - **Notes:** 

- [ ] **Puzzle doors** that require activation
  - **How to test:** Find locked door linked to puzzle
  - **Expected:** Door won't open until puzzle solved
  - **Status:** 
  - **Notes:** 

### Puzzle Interaction
- [ ] **Can activate pressure plates** (walk on them)
  - **How to test:** Walk onto pressure plate
  - **Expected:** Plate visibly depresses, triggers effect
  - **Status:** 
  - **Notes:** 

- [ ] **Can toggle levers** (F key)
  - **How to test:** Press F near lever
  - **Expected:** Lever switches position
  - **Status:** 
  - **Notes:** 

- [ ] **Puzzle completion** opens doors/rewards
  - **How to test:** Complete puzzle sequence
  - **Expected:** Door opens or chest unlocks
  - **Status:** 
  - **Notes:** 

---

## Phase 11.3: Environmental Destruction

### Destructible Objects
- [ ] **Crates** spawn in rooms
  - **How to test:** Look for wooden boxes
  - **Expected:** Brown crate objects in dungeon
  - **Status:** 
  - **Notes:** 

- [ ] **Barrels** spawn in rooms
  - **How to test:** Look for barrel objects
  - **Expected:** Cylindrical barrel objects
  - **Status:** 
  - **Notes:** 

- [ ] **Explosive barrels** (red barrels)
  - **How to test:** Look for red-colored barrels
  - **Expected:** Different visual from regular barrels
  - **Status:** 
  - **Notes:** 

### Destruction Mechanics
- [ ] **Can attack objects** to destroy them
  - **How to test:** Attack crate/barrel with Space
  - **Expected:** Object breaks after hits
  - **Status:** 
  - **Notes:** 

- [ ] **Objects drop loot** when destroyed
  - **How to test:** Destroy object, look for item drops
  - **Expected:** Health potion, gold, or items appear
  - **Status:** 
  - **Notes:** 

- [ ] **Explosive barrels** cause chain reactions
  - **How to test:** Destroy explosive barrel near others
  - **Expected:** Explosion damages nearby barrels/enemies
  - **Status:** 
  - **Notes:** 

### Object Interaction
- [ ] **Can pick up objects** (F key)
  - **How to test:** Press F near crate
  - **Expected:** Pick up and carry object
  - **Status:** 
  - **Notes:** 

- [ ] **Can throw carried objects** (F key again)
  - **How to test:** While carrying, press F
  - **Expected:** Object flies in aim direction
  - **Status:** 
  - **Notes:** 

- [ ] **Thrown objects damage enemies**
  - **How to test:** Throw object at enemy
  - **Expected:** Enemy takes damage on impact
  - **Status:** 
  - **Notes:** 

---

## Phase 12.1: Grammar-Based Layout Generation

- [ ] **Dungeon layouts** look different from v1.1 BSP
  - **How to test:** Generate multiple dungeons, observe patterns
  - **Expected:** Non-rectangular rooms, organic shapes
  - **Status:** 
  - **Notes:** 

---

## Phase 12.2: Dynamic Narrative Assembly

### Story Events
- [ ] **Narrative events** trigger during gameplay
  - **How to test:** Play for 10+ minutes, watch for story text
  - **Expected:** Story snippets appear in UI or dialog
  - **Status:** 
  - **Notes:** 

- [ ] **Quests have storylines**
  - **How to test:** Read quest descriptions (J key)
  - **Expected:** Quests reference overarching story
  - **Status:** 
  - **Notes:** 

- [ ] **NPC dialogue** references world story
  - **How to test:** Talk to NPCs (F key near merchant)
  - **Expected:** Dialogue mentions factions, events
  - **Status:** 
  - **Notes:** 

---

## Phase 12.3: Enhanced Procedural Music

### Adaptive Music
- [ ] **Music changes** in combat
  - **How to test:** Enter/exit combat, listen for change
  - **Expected:** Combat music more intense than exploration
  - **Status:** 
  - **Notes:** 

- [ ] **Boss music** plays for boss encounters
  - **How to test:** Fight boss enemy
  - **Expected:** Dramatic boss fight music
  - **Status:** 
  - **Notes:** 

### Musical Motifs
- [ ] **Character/faction themes** audible
  - **How to test:** Interact with faction NPCs
  - **Expected:** Subtle musical theme for each faction
  - **Status:** 
  - **Notes:** 

---

## Phase 13.1: Behavior Tree AI

### Enemy Behavior
- [ ] **Enemies use complex tactics**
  - **How to test:** Observe enemy AI during combat
  - **Expected:** Flanking, retreating, using cover
  - **Status:** 
  - **Notes:** 

- [ ] **AI adapts to player strategy**
  - **How to test:** Use same tactic repeatedly
  - **Expected:** Enemies counter or change behavior
  - **Status:** 
  - **Notes:** 

---

## Phase 13.2: Squad Tactics & Coordination

### Squad Behavior
- [ ] **Enemies coordinate** attacks
  - **How to test:** Fight multiple enemies
  - **Expected:** Simultaneous attacks, flanking patterns
  - **Status:** 
  - **Notes:** 

- [ ] **Squad formations** visible
  - **How to test:** Watch enemy groups move
  - **Expected:** Enemies maintain relative positions
  - **Status:** 
  - **Notes:** 

---

## Phase 13.3: Faction Reputation & Relationships

### Faction System
- [ ] **Faction reputation** shown in UI
  - **How to test:** Check character sheet (C key)
  - **Expected:** Reputation values for factions
  - **Status:** 
  - **Notes:** 

- [ ] **Actions affect reputation**
  - **How to test:** Kill faction member, check reputation
  - **Expected:** Reputation decreases with that faction
  - **Status:** 
  - **Notes:** 

- [ ] **NPCs react** to reputation
  - **How to test:** Talk to NPC with low reputation
  - **Expected:** Hostile dialogue or refusal to trade
  - **Status:** 
  - **Notes:** 

---

## Phase 14.1: Enhanced Lighting & Shadows

**Note:** Requires `-enable-lighting` flag to test

### Dynamic Lighting
- [ ] **Light sources** visible (torches, crystals)
  - **How to test:** Run `./venture-client -enable-lighting`
  - **Expected:** Glowing lights throughout dungeon
  - **Status:** 
  - **Notes:** 

- [ ] **Shadows cast** from entities
  - **How to test:** Stand near light source
  - **Expected:** Shadow extends from player
  - **Status:** 
  - **Notes:** 

- [ ] **Torch flicker** effect
  - **How to test:** Watch wall torches
  - **Expected:** Light intensity varies over time
  - **Status:** 
  - **Notes:** 

---

## Phase 14.2: Animated Sprites

### Animation System
- [ ] **Player sprite animates** during movement
  - **How to test:** Move with WASD
  - **Expected:** Multi-frame walk animation
  - **Status:** 
  - **Notes:** 

- [ ] **Idle animation** when standing still
  - **How to test:** Stop moving
  - **Expected:** Subtle breathing/idle animation
  - **Status:** 
  - **Notes:** 

- [ ] **Enemy sprites animate**
  - **How to test:** Watch enemy movement
  - **Expected:** Enemy walk/attack animations
  - **Status:** 
  - **Notes:** 

---

## Phase 14.3: Particle System Expansion

**Note:** Weather requires `-enable-weather` flag

### Particle Effects
- [ ] **Combat particles** on hits
  - **How to test:** Attack enemy
  - **Expected:** Sparks, blood, or impact particles
  - **Status:** 
  - **Notes:** 

- [ ] **Death particles** on enemy death
  - **How to test:** Kill enemy
  - **Expected:** Particle burst when enemy dies
  - **Status:** 
  - **Notes:** 

- [ ] **Weather effects** (rain, snow, etc.)
  - **How to test:** Run with `-enable-weather -weather rain`
  - **Expected:** Rain particles fall from top
  - **Status:** 
  - **Notes:** 

---

## Phase 14.4: Audio System Enhancement

### Positional Audio
- [ ] **Sounds change** based on distance
  - **How to test:** Walk toward/away from sound source
  - **Expected:** Volume increases when closer
  - **Status:** 
  - **Notes:** 

- [ ] **Sounds pan** left/right based on position
  - **How to test:** Enemy attack from left/right
  - **Expected:** Sound comes from correct direction
  - **Status:** 
  - **Notes:** 

### Reverb
- [ ] **Reverb effect** in large rooms
  - **How to test:** Enter large room, listen for echo
  - **Expected:** Sounds have slight echo/reverb
  - **Status:** 
  - **Notes:** 

---

## General Testing

### Performance
- [ ] **Frame rate** is playable (>30 FPS)
  - **How to test:** Run with `-profile` flag, check console
  - **Expected:** Average frame time < 16ms (60 FPS)
  - **Status:** 
  - **Notes:** 

- [ ] **No stuttering** during gameplay
  - **How to test:** Move around, enter combat
  - **Expected:** Smooth motion, no freezes
  - **Status:** 
  - **Notes:** 

### Stability
- [ ] **No crashes** during 10+ min playtest
  - **How to test:** Play normally for 10 minutes
  - **Expected:** Game remains stable
  - **Status:** 
  - **Notes:** 

- [ ] **No error messages** in console
  - **How to test:** Watch console output
  - **Expected:** Only INFO logs, no ERRORs
  - **Status:** 
  - **Notes:** 

---

## Summary

**Total Features Tested:** [count]  
**✅ Working:** [count]  
**❌ Broken:** [count]  
**⚠️ Partial:** [count]  
**❓ Unclear:** [count]

### Top 5 Most Broken Features
1. 
2. 
3. 
4. 
5. 

### Notes for Developers
[Add any additional observations, patterns, or theories about what's broken]

---

**Tester:** [Your name]  
**Date:** [Test date]  
**Build:** `git rev-parse --short HEAD` output: [commit hash]  
**Platform:** [Linux/macOS/Windows/Web]
