# Venture User Manual

Complete guide to gameplay mechanics, systems, and features.

**Version:** 2.0  
**Last Updated:** October 2025

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
**Interface:** I (inventory), C (character), K (skills), J (quests), M (map), R (crafting), Tab (cycle targets), Esc (close/pause)  
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

### Crafting System

**Stations:** Forge, Alchemy Lab, Enchanting Table, Workbench  
**Process:** Recipes + materials = crafted items  
**Unlock:** Find recipes in world or purchase from merchants

### Rotation & Camera

**Rotation:** Q/E keys for 360° view  
**Camera:** Follow player, scroll to zoom, shake on impacts

### Lighting & Shadows

Dynamic lighting with intensity/color, shadow projection from entities/terrain.

### Save System

**Auto-Save:** On level transitions, quest completion  
**Manual Save:** F5 (quick save), Pause Menu → Save Game  
**Format:** JSON files in `saves/` directory

**Cloud Sync:** Not currently supported (local saves only).

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

**Version:** 2.0  
**Last Updated:** October 2025  
**Maintained By:** Venture Development Team
