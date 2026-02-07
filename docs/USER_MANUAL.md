# Venture User Manual

Complete guide to gameplay mechanics, systems, and features.

**Version:** 1.0.0  
**Last Updated:** February 2026

**New to Venture?** Start with [Getting Started Guide](GETTING_STARTED.md).

---

## Table of Contents

1. [Introduction](#introduction)
2. [Game Controls](#game-controls)
3. [Character System](#character-system)
4. [Combat & Inventory](#combat--inventory)
5. [Magic, Skills & Quests](#magic-skills--quests)
6. [World & Multiplayer](#world--multiplayer)
7. [Advanced Features](#advanced-features)

---

## Introduction

Venture is a procedurally generated action-RPG where everything—dungeons, enemies, items, abilities, music, graphics—is generated at runtime.

**Core Principles:** Procedural everything, deterministic generation (same seed = same world), real-time action, co-op multiplayer, multiple genres.

---

## Game Controls

### Keyboard & Mouse

**Movement:** WASD (diagonal supported)  
**Actions:** Space (attack/interact), E (use item), F (NPC interaction), 1-5 (spells)  
**Interface:** I (inventory), C (character), K (skills), J (quests), M (map), R (crafting), G (gallery), H (housing), Tab (cycle targets), F1 (help), Esc (close/pause)  
**Saving:** F5 (quick save), F9 (quick load)  
**Mouse:** Move (aim), Left Click (select/attack), Right Click (cancel), Scroll (zoom)

### Menu Navigation

All menus support dual-exit: original key OR Esc. Click close button or press hotkey/Esc to exit.

### Touch Controls (Mobile/WASM)

Virtual D-pad (bottom-left), Action buttons A/B (bottom-right), Menu button (top-right).

---

## Character System

### Stats

**Primary:** Health (survivability), Mana (magic resource), Attack (physical damage), Defense (damage reduction), Magic Power (spell damage), Magic Defense (spell resistance)  
**Secondary:** Crit Chance/Damage, Evasion, Resistances (Fire, Ice, Lightning, Poison, Dark)

**Formula:** `FinalDamage = BaseDamage × (1 - TargetDefense/100) × CritMultiplier`

### Experience & Leveling

Kill enemies, complete quests, explore to gain XP. Each level grants stat points and skill points.

**Level Rewards:** +5 stat points, +1 skill point

### Death System

On death: lose 10% gold (min 100, max 10,000), respawn at safe point, keep items/XP. Hardcore mode: permadeath.

---

## Combat & Inventory

### Combat

**Attack Types:** Physical (melee/ranged), Magical (elemental/arcane)  
**Status Effects:** Burn (DOT), Freeze (slow/stun), Poison (DOT), Stun (immobilize), Bleed (DOT)  
**Resistances:** Reduce incoming elemental damage

### Inventory & Equipment

**Capacity:** 30 items, 100 units weight  
**Equipment Slots:** Weapon, Armor (Head, Chest, Legs, Hands, Feet), Accessories (2)  
**Item Rarity:** Common (1.0x), Uncommon (1.2x), Rare (1.5x), Epic (2.0x), Legendary (3.0x)

**Auto-Pickup:** Items collected automatically when touched.

---

## Magic, Skills & Quests

### Magic System

**Spell Slots:** 5 quick-cast slots (1-5 keys)  
**Mana:** Required for casting, regenerates over time  
**Elemental Types:** Fire, Ice, Lightning, Poison, Dark, Light, Arcane

**Formula:** `Damage = BaseDamage × (1 + MagicPower/100) × (1 - TargetMagicDefense/100)`

### Skill Trees

**Trees:** Combat, Magic, Defense, Utility  
**Progression:** Unlock skills with skill points, prerequisites required  
**Reset:** Visit trainers to reset skills (gold cost)

### Quest System

**Types:** Kill (defeat enemies), Collect (gather items), Explore (discover areas), Escort (protect NPC), Boss (defeat boss)  
**Rewards:** XP, gold, items, skill points  
**Max Active:** 5 quests

---

## World & Multiplayer

### World Generation

**Seed-Based:** Enter seed for reproducible worlds  
**Dungeon Levels:** Difficulty/loot increases with depth  
**Room Types:** Start, Boss, Treasure, Shop, Shrine, Library, Trap, Standard

### Genre System

**Genres:** Fantasy (magic/dragons), Sci-Fi (tech/lasers), Horror (dark/monsters), Cyberpunk (neon/hackers), Post-Apocalyptic (wasteland/survivors)

Each genre has unique themes, color palettes, entity types, and items.

### Multiplayer

**Mode:** Co-op (2-4 players)  
**Connection:** Host server or join via IP:port  
**Features:** Shared world (same seed), independent inventories, XP sharing, loot instancing

**Lag Compensation:** Supports 200-5000ms latency with client prediction.

---

## Advanced Features

### Visual Quality (V3.0)

**Enhanced Graphics:**
- **Sprite Detail:** 40% increase with anatomical accuracy, facial features (eyes, mouth)
- **Silhouette Quality:** Improved from 0.65 to 0.75 average score
- **Anti-Aliasing:** 4 quality levels (Off, Low, Medium, High) with 2x2 to 8x8 super-sampling
- **Genre Variations:** Each genre has unique visual style (organic, geometric, distorted, augmented, weathered)

**Lighting System (V3.0 Enhanced):**
- Soft shadows with gradient edges (no harsh pixels)
- Colored lighting matching light sources (red fire, blue magic, neon colors)
- Bloom effects for magical and technological lights
- Advanced ambient occlusion for depth perception
- Dynamic flickering torches and environmental lights
- Genre-specific lighting presets (warm fantasy, cool sci-fi, dim horror, neon cyberpunk, harsh post-apocalyptic)
- Performance: <5% frame time overhead

**Weather Systems (V3.0 New):**
- Comprehensive weather types: rain, snow, fog, dust, ash
- Genre-specific variations (neon rain in cyberpunk, blood rain in horror, radiation dust in post-apocalyptic)
- Fluid simulation for realistic particle behavior
- Intensity levels: light, medium, heavy, extreme
- Environmental interactions (particles affected by wind and gravity)
- Weather transitions between types

**Tile Rendering (V3.0 Enhanced):**
- Procedural texture patterns: stone (granite, marble, cobblestone), wood (oak, rough logs, weathered boards), metal (steel, rusted iron, tech plating), organic (grass, dirt, moss, coral)
- 50+ unique patterns per genre via seed-based generation
- Smooth transitions with automated edge detection
- Multi-layer depth effects for visual richness
- Detail layers for texture complexity
- Normal mapping simulation for depth perception

**UI Enhancements (V3.0):**
- Dynamic color palettes that adapt to genre theme
- Improved visual hierarchy for better readability
- Smooth menu transitions and animations
- Procedural UI decorations matching genre style

**Post-Processing (V3.0 New):**
- Parallax backgrounds with multi-layer depth (2-5 layers)
- Time-of-day system with dynamic lighting shifts (optional day/night cycles)
- Screen-space enhancements and visual polish
- Genre-specific visual filters (warm fantasy, cool sci-fi, desaturated horror, high-contrast cyberpunk, washed-out post-apocalyptic)

**Performance Control:**
```bash
# Configure weather for performance
-weather ""              # Disable weather effects (empty string)
-weather-intensity light # Use light weather for better performance
```

**Performance (V3.0):**
- 106 FPS maintained with 2000 entities (70% above 60 FPS target)
- 73MB memory usage (86% below 500MB budget)
- Sprite generation: 3-5ms per sprite (with 40% more detail)
- Cache hit rate: 95.9% (maintained from V2.0)

### Crafting System

**Stations:** Forge, Alchemy Lab, Enchanting Table, Workbench  
**Process:** Recipes + materials = crafted items  
**Unlock:** Find recipes in world or purchase from merchants

### Rotation & Camera

**Rotation:** Q/E keys for 360° view  
**Camera:** Follow player, scroll to zoom, shake on impacts

### Lighting & Shadows (V3.0 Enhanced)

Dynamic lighting with soft shadows, colored lights, and bloom effects. Genre-specific lighting presets create appropriate atmosphere for each theme.

### Save System

**Auto-Save:** On level transitions, quest completion  
**Manual Save:** F5 (quick save), Pause Menu → Save Game  
**Format:** JSON files in `saves/` directory

**Cloud Sync:** Not currently supported (local saves only).

**V3.0 Compatibility:** V2.0 save files load with enhanced V3.0 visuals automatically.

---

## Tips & Strategies

**Combat:** Use elemental advantages, dodge enemy attacks, manage mana/health  
**Exploration:** Check every corner for secrets, use map to track progress  
**Build:** Balance offense/defense, specialize in combat or magic  
**Multiplayer:** Coordinate with team, share resources, revive fallen allies

---

## Troubleshooting

**Performance:** Lower resolution, reduce particles, adjust graphics settings  
**Crashes:** Check logs, update drivers, report bugs with save files  
**Multiplayer:** Verify firewall settings, check server IP/port, test connectivity

---

## Additional Resources

**Documentation:** [Getting Started](GETTING_STARTED.md), [Development](DEVELOPMENT.md), [Architecture](ARCHITECTURE.md)  
**Community:** [GitHub Issues](https://github.com/opd-ai/venture/issues), [Discussions](https://github.com/opd-ai/venture/discussions)

---

**Version:** 1.0.0  
**Last Updated:** February 2026  
**Maintained By:** Venture Development Team
