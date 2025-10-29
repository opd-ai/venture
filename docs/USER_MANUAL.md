# Venture User Manual

Complete guide to gameplay mechanics, systems, and features.

**Version:** 1.0  
**Last Updated:** October 2025

**New to Venture?** Start with the [Getting Started Guide](GETTING_STARTED.md) for a quick 5-minute introduction.

---

## Table of Contents

1. [Introduction](#introduction)
2. [Game Controls](#game-controls)
3. [Character System](#character-system)
4. [Combat Mechanics](#combat-mechanics)
5. [Inventory & Equipment](#inventory--equipment)
6. [Magic & Abilities](#magic--abilities)
7. [Skill Trees](#skill-trees)
8. [Quest System](#quest-system)
9. [World Generation](#world-generation)
10. [Multiplayer](#multiplayer)
11. [Save System](#save-system)
12. [Genre System](#genre-system)
13. [Advanced Mechanics](#advanced-mechanics)

---

## Introduction

Venture is a procedurally generated action-RPG where every playthrough is unique. The game generates everything at runtime—dungeons, enemies, items, abilities, music, and graphics—ensuring infinite replayability.

### Core Design Philosophy

- **Procedural Everything**: No pre-made assets
- **Deterministic Generation**: Same seed = same world
- **Real-time Action**: No turn-based waiting
- **Co-op Multiplayer**: Play with friends
- **Genre Diversity**: Multiple thematic worlds

---

## Game Controls

### Default Keyboard & Mouse

**Movement:**
- `W` - Move Up
- `A` - Move Left
- `S` - Move Down
- `D` - Move Right
- Diagonal movement supported (e.g., W+D = northeast)

**Actions:**
- `Space` - Primary Attack / Interact
- `E` - Use Item / Open Chest
- `F` - Interact with NPCs / Merchants (when nearby)
- `1-5` - Cast Spell (slots 1-5)

**Interface:**
- `I` - Inventory (Press I or ESC to close)
- `C` - Character Stats (Press C or ESC to close)
- `K` - Skill Tree (Press K or ESC to close)
- `J` - Quest Log (Press J or ESC to close)
- `M` - Map (Press M or ESC to close)
- `R` - Crafting Menu (Press R or ESC to close)
- `Tab` - Cycle Targets
- `Esc` - Close any open menu / Pause Menu

**Saving:**
- `F5` - Quick Save
- `F9` - Quick Load

**Mouse:**
- `Move` - Aim / Look direction
- `Left Click` - Confirm / Select / Attack
- `Right Click` - Cancel / Alt action
- `Scroll Wheel` - Zoom camera

### Custom Key Bindings

Edit key bindings in the settings menu (Esc → Settings → Controls).

### Menu Navigation Standard

All in-game menus follow a consistent, user-friendly navigation pattern:

| Menu | Open Key | Close Methods |
|------|----------|---------------|
| Inventory | `I` | Press `I` again OR press `ESC` |
| Character Stats | `C` | Press `C` again OR press `ESC` |
| Skill Tree | `K` | Press `K` again OR press `ESC` |
| Quest Log | `J` | Press `J` again OR press `ESC` |
| World Map | `M` | Press `M` again OR press `ESC` |
| Crafting | `R` | Press `R` again OR press `ESC` |

**Key Navigation Features:**
- **Toggle Key**: Each menu uses its assigned letter key to both open and close
- **Universal Exit**: `ESC` closes any open menu without reopening it
- **No Traps**: You're never stuck in a menu—both exit methods always work simultaneously
- **Visual Hints**: Each menu displays "Press [KEY] or [ESC] to close" at the top or bottom

**Example:** Open inventory with `I`, close it with either `I` or `ESC`. The same pattern works for all menus.

---

## Character System

### Core Stats

Your character has six primary stats:

1. **Health (HP)**: Maximum hit points
   - Base: 100 HP
   - Growth: +20 HP per level
   - Modified by: Equipment, buffs, debuffs

2. **Mana (MP)**: Magic spell resource
   - Base: 50 MP (if magic class)
   - Growth: +10 MP per level
   - Regenerates: 1 MP per second

3. **Attack**: Physical damage output
   - Base: 10
   - Growth: +5 per level
   - Modified by: Weapons, strength

4. **Defense**: Damage reduction
   - Base: 5
   - Growth: +3 per level
   - Modified by: Armor, buffs
   - Formula: `Damage = max(1, BaseDamage - Defense)`

5. **Magic Power**: Spell effectiveness
   - Base: 5
   - Growth: +4 per level
   - Modified by: Intelligence, equipment

6. **Speed**: Movement and action rate
   - Base: 100 units/second
   - Growth: +5 per level
   - Modified by: Buffs, debuffs, equipment

### Secondary Stats

Derived from primary stats and equipment:

- **Critical Chance**: % chance for 2x damage (base: 5%)
- **Critical Damage**: Multiplier for crits (base: 2.0x)
- **Evasion**: % chance to dodge (base: 5%)
- **Resistances**: Reduce elemental damage
  - Fire, Ice, Lightning, Poison, Dark

### Stat Scaling

Stats scale with level using the formula:
```
FinalStat = BaseStat * (1.0 + 0.15 * (Level - 1)) * RarityMultiplier
```

**Rarity Multipliers:**
- Common: 1.0x
- Uncommon: 1.2x
- Rare: 1.5x
- Epic: 2.0x
- Legendary: 3.0x

---

## Combat Mechanics

### Combat Types

**1. Melee Combat**
- Close-range physical attacks
- High damage, short range
- Fast attack speed
- No resource cost

**2. Ranged Combat**
- Distance physical attacks
- Medium damage, long range
- Requires ammunition (or infinite)
- Aim with mouse

**3. Magic Combat**
- Spell-based attacks
- Variable damage and effects
- Consumes mana
- Diverse targeting patterns

### Damage Calculation

**Physical Damage:**
```
BaseDamage = Attack * WeaponMultiplier
CritDamage = BaseDamage * CritMultiplier (if crit)
FinalDamage = max(1, CritDamage - TargetDefense)
```

**Magical Damage:**
```
BaseDamage = MagicPower * SpellPower
Resistance = Target.GetResistance(Element)
FinalDamage = BaseDamage * (1.0 - Resistance)
```

### Status Effects

Combat can apply temporary effects:

**Damage Over Time:**
- Poison: -5 HP/sec for 5 seconds
- Burn: -10 HP/sec for 3 seconds
- Bleed: -3 HP/sec for 10 seconds

**Debuffs:**
- Slow: -50% movement speed
- Weak: -30% attack damage
- Curse: -20% all stats

**Buffs:**
- Haste: +50% speed
- Strength: +30% attack
- Shield: +50 temporary HP

### Combat Tips

1. **Positioning**: Use terrain for advantage
2. **Timing**: Learn enemy attack patterns
3. **Resource Management**: Don't waste mana
4. **Status Effects**: Apply DoTs early
5. **Combos**: Chain abilities for efficiency

---

## Inventory & Equipment

### Inventory System

- **Capacity**: 20 item slots (upgradeable)
- **Weight Limit**: 100 units (upgradeable)
- **Organization**: Auto-sort by type/rarity
- **Stacking**: Consumables stack to 99

### Equipment Slots

Your character has four equipment slots:

1. **Weapon**: Primary damage source
2. **Armor**: Body defense
3. **Helmet**: Head protection
4. **Accessory**: Special effects

### Item Rarity

Items come in five rarity tiers:

| Rarity | Color | Stat Bonus | Drop Rate |
|--------|-------|------------|-----------|
| Common | Gray | 1.0x | 60% |
| Uncommon | Green | 1.2x | 25% |
| Rare | Blue | 1.5x | 10% |
| Epic | Purple | 2.0x | 4% |
| Legendary | Orange | 3.0x | 1% |

### Item Types

**Weapons:**
- Swords: Balanced melee
- Axes: High damage, slow
- Daggers: Fast, low damage
- Bows: Ranged physical
- Staves: Magic casting

**Armor:**
- Light: Low defense, high speed
- Medium: Balanced
- Heavy: High defense, low speed

**Consumables:**
- Health Potions: Restore HP
- Mana Potions: Restore MP
- Scrolls: Single-use spells
- Food: Temporary buffs

**Accessories:**
- Rings: Stat bonuses
- Amulets: Special abilities
- Charms: Resistance bonuses

### Inventory Management

**Picking Up Items:**
- Walk over items to auto-pickup
- Press 'E' to manually pick up
- Inventory full? Drop lowest value item

**Dropping Items:**
- Open inventory (I)
- Right-click item
- Select "Drop"

**Selling Items:**
- Visit NPC merchants
- Sell value = 50% of item value
- Rare items worth more

---

## Magic & Abilities

### Spell Types

1. **Offensive**: Damage enemies
   - Fireball: Single target, fire damage
   - Lightning Bolt: Chain lightning
   - Meteor: Area of effect

2. **Defensive**: Protect yourself
   - Shield: Temporary HP
   - Ice Armor: +Defense, slow attackers
   - Barrier: Block projectiles

3. **Healing**: Restore health
   - Heal: Single target HP restore
   - Regeneration: HP over time
   - Life Steal: Damage → HP

4. **Buff**: Enhance abilities
   - Haste: +Speed temporarily
   - Strength: +Attack damage
   - Focus: +Critical chance

5. **Debuff**: Weaken enemies
   - Slow: Reduce enemy speed
   - Weaken: Lower enemy damage
   - Curse: Reduce all enemy stats

6. **Utility**: Special effects
   - Teleport: Instant movement
   - Invisibility: Hide from enemies
   - Light: Illuminate darkness

7. **Summon**: Create allies
   - Summon Minion: AI companion
   - Spirit: Temporary fighter
   - Familiar: Permanent pet

### Spell Targeting

**Self**: Affects only you
**Single Target**: One enemy/ally
**Area of Effect (AoE)**: Circle around target
**Cone**: 90° arc in front
**Line**: Straight projectile
**All Allies**: Everyone on your team
**All Enemies**: All hostile targets

### Mana Management

- **Base Regen**: 1 MP/second
- **Potion Restore**: 50 MP instantly
- **Wait to Regen**: Avoid overcasting
- **Mana Efficiency**: Use low-cost spells frequently

---

## Skill Trees

### Tree Structure

Each character class has 3-4 skill trees with 20-30 skills each:

**Fantasy Example:**
- **Warrior Tree**: Melee combat, defense, tanking
- **Mage Tree**: Elemental magic, crowd control
- **Rogue Tree**: Stealth, critical hits, mobility

**Sci-Fi Example:**
- **Soldier Tree**: Weapons, tactics, armor
- **Engineer Tree**: Tech abilities, turrets
- **Biotic Tree**: Psychic powers, shields

### Skill Tiers

Skills are organized by tier:

1. **Basic (Tier 1)**: Available from level 1
2. **Intermediate (Tier 2)**: Unlocks at level 10
3. **Advanced (Tier 3)**: Unlocks at level 20
4. **Master (Tier 4)**: Unlocks at level 30

### Skill Points

- **Earn**: 1 point per level
- **Respend**: 100 gold to reset tree
- **Max Level**: Each skill has 5 ranks
- **Prerequisites**: Some skills require others

### Skill Types

1. **Passive**: Always active bonuses
   - +10% HP
   - +5 Attack
   - +15% Crit Chance

2. **Active**: Abilities you trigger
   - Power Strike: Double damage
   - Dash: Quick movement
   - Heal: Restore HP

3. **Ultimate**: Powerful once-per-minute abilities
   - Rage: +100% damage for 10s
   - Time Stop: Freeze enemies
   - Meteor Storm: Massive AoE

4. **Synergy**: Combo with other skills
   - "Fire + Ice = Steam Explosion"
   - "Dash + Attack = Critical Strike"

---

## Quest System

### Quest Types

1. **Main Quests**: Story progression
   - Required for game completion
   - Unlock new areas
   - High rewards

2. **Side Quests**: Optional objectives
   - Extra XP and items
   - Lore and world-building
   - Repeatable (some)

3. **Bounty Quests**: Kill targets
   - Eliminate specific enemies
   - Time-limited
   - Gold rewards

4. **Collection Quests**: Gather items
   - Find X items
   - Return to questgiver
   - Item rewards

5. **Escort Quests**: Protect NPCs
   - NPC follows you
   - Keep them alive
   - XP rewards

### Quest Tracking

- **Quest Log**: Press 'J' to view active quests
- **Objectives**: Track progress for each quest
- **Waypoints**: Map markers for quest locations
- **Notifications**: Alert on quest completion

### Quest Rewards

- **Experience Points**: Scale with quest level
- **Gold**: For purchasing items
- **Items**: Unique equipment or consumables
- **Skill Points**: Rare reward for major quests

---

## World Generation

### Procedural Terrain

Dungeons are generated using two algorithms:

1. **Binary Space Partitioning (BSP)**
   - Creates rectangular rooms
   - Connected by corridors
   - Structured layout

2. **Cellular Automata**
   - Organic cave systems
   - Natural-looking spaces
   - Less predictable

### Depth Scaling

As you descend, difficulty increases:

- **Depth 1-5**: Tutorial area, common enemies
- **Depth 6-10**: Increased difficulty
- **Depth 11-20**: Rare enemies, better loot
- **Depth 21+**: End-game content, legendary items

### Room Types

- **Standard**: Basic empty room
- **Combat**: Filled with enemies
- **Treasure**: Contains loot chests
- **Boss**: Powerful enemy encounter
- **Shop**: NPC merchant
- **Shrine**: Healing or buffs
- **Secret**: Hidden valuable items

### Seed System

World generation is deterministic:
```bash
# Same seed = same world
./venture-client -seed 12345
```

Use seeds to:
- Share worlds with friends
- Replay favorite dungeons
- Speed run competitions
- Test specific scenarios

---

## Multiplayer

### Game Modes

**Cooperative (Co-op):**
- 2-4 players
- Shared world
- Split loot
- Team-based combat

**Server Types:**
- **Dedicated**: Standalone server
- **Listen**: Player hosts
- **Local**: Same computer (splitscreen not supported yet)

### Hosting a Server

Start a dedicated server:
```bash
./venture-server -port 8080 -max-players 4 -seed 12345
```

**Server Options:**
- `-port`: Network port (default: 8080)
- `-max-players`: Player limit (default: 4)
- `-tick-rate`: Updates/sec (default: 20)
- `-seed`: World seed
- `-genre`: World theme

### Joining a Server

Connect as a client:
```bash
./venture-client -server address:port
```

Examples:
```bash
# Local network
./venture-client -server 192.168.1.100:8080

# Internet
./venture-client -server game.example.com:8080

# Localhost
./venture-client -server localhost:8080
```

### Network Features

**Client-Side Prediction:**
- Your inputs are applied immediately
- Server validates and corrects
- Smooth gameplay even with lag

**Entity Interpolation:**
- Other players move smoothly
- 100-200ms delay buffer
- Hides network jitter

**Lag Compensation:**
- Server rewinds for hit detection
- Fair combat with high latency
- Supports 200-5000ms connections

### Multiplayer Tips

1. **Communicate**: Use voice chat or text
2. **Share Loot**: Be fair with items
3. **Revive Allies**: Help downed teammates
4. **Coordinate**: Plan strategies together
5. **Low Latency**: Use wired connections when possible

---

## Save System

### Save Formats

Venture uses JSON save files:
- **Location**: `./saves/` directory
- **Format**: Human-readable JSON
- **Size**: 2-10 KB per save

### What's Saved

**Player State:**
- Position
- Health and mana
- Stats and level
- Experience points
- Inventory and equipment
- Active quests

**World State:**
- World seed (regenerates terrain)
- Current depth
- Time played
- Difficulty setting
- Genre

**Settings:**
- Key bindings
- Screen resolution
- Audio volumes
- Graphics options

### Save Operations

**Quick Save:**
- Press `F5` anytime
- Saves to "quicksave.json"
- Overrides previous quicksave

**Quick Load:**
- Press `F9` to load quicksave
- Returns to last F5 position

**Manual Save:**
- Pause menu → Save Game
- Choose save slot (1-10)
- Enter save name

**Auto-save:**
- Every 5 minutes
- On level transition
- Before boss fights
- "autosave.json"

### Save Management

**List Saves:**
- View all saves in menu
- Shows date, time, character level

**Delete Saves:**
- Remove unwanted saves
- Frees disk space

**Export/Import:**
- Copy save files to share
- Backup important saves

---

## Genre System

### Available Genres

1. **Fantasy** 🧙
   - Medieval theme
   - Magic and swords
   - Dragons and dungeons
   - Colors: Earth tones, magic glows

2. **Sci-Fi** 🚀
   - Futuristic technology
   - Laser weapons
   - Alien enemies
   - Colors: Neon blues, metallics

3. **Horror** 👻
   - Dark atmosphere
   - Scary monsters
   - Limited visibility
   - Colors: Dark, blood red, sickly green

4. **Cyberpunk** 🌆
   - Urban future
   - Hacking abilities
   - Corporate enemies
   - Colors: Neon pink, purple, blue

5. **Post-Apocalyptic** ☢️
   - Wasteland survival
   - Mutant creatures
   - Makeshift weapons
   - Colors: Brown, gray, rust

### Genre Blending

Combine two genres for hybrid worlds:

```bash
# Sci-Fi Horror
./venture-client -genre scifi -secondary horror -blend 0.3

# Dark Fantasy
./venture-client -genre fantasy -secondary horror -blend 0.5
```

**Blend Weight (0.0-1.0):**
- 0.0 = Pure primary genre
- 0.5 = Equal mix
- 1.0 = Pure secondary genre

**Blended Features:**
- Mixed color palettes
- Hybrid enemy types
- Combined item themes
- Merged music styles

---

## Advanced Mechanics

### Damage Types

- **Physical**: Reduced by defense
- **Fire**: Burning damage over time
- **Ice**: Slows enemy movement
- **Lightning**: Chain to nearby enemies
- **Poison**: DoT, ignores armor
- **Dark**: Lifesteal component

### Elemental Interactions

- Fire + Ice = Explosion
- Lightning + Water = AoE shock
- Earth + Wind = Sandstorm
- Light + Dark = Void damage

### Enemy AI Behaviors

**Melee AI:**
- Chase player when in range
- Stop to attack
- Return to spawn when far

**Ranged AI:**
- Keep distance from player
- Circle strafe while shooting
- Flee when player too close

**Mage AI:**
- Cast spells at range
- Teleport when threatened
- Summon minions

**Boss AI:**
- Multiple phases
- Special abilities
- Pattern-based attacks
- Enrage at low HP

### Commerce & Trading

**Merchant NPCs:**
Merchants spawn in dungeons and settlements, offering items for purchase and accepting items for sale.

**Interacting with Merchants:**
1. Approach a merchant (within ~64 pixels)
2. Press `F` to interact
3. Dialog menu appears with options
4. Shop UI opens showing merchant inventory

**Buying Items:**
- Browse merchant inventory in shop UI
- Item prices scale with rarity (Common 1.0x, Uncommon 1.5x, Rare 3.0x, Epic 8.0x, Legendary 25.0x)
- Click item to purchase (requires sufficient gold)
- Item transfers to your inventory

**Selling Items:**
- Switch to "Sell" mode in shop UI
- Select items from your inventory
- Receive gold based on item value
- Useful for clearing inventory space

**Merchant Types:**
- **Fixed Merchants**: Permanent location in settlements
- **Nomadic Merchants**: Spawn periodically in random dungeon rooms

### Crafting System

**Opening Crafting UI:**
Press `R` anywhere to open the crafting menu.

**Recipe-Based Crafting:**
- Discover recipes through gameplay
- Each recipe requires specific materials
- Success chance scales with skill level (50% at level 1, 95% at max)
- Failed crafts consume 50% of materials

**Crafting Categories:**
- **Potions**: Health, mana, buff consumables
- **Enchanting**: Enhance weapon/armor stats
- **Magic Items**: Create wands, rings, amulets

**Material Collection:**
- Gather materials from defeated enemies
- Find materials in treasure chests
- Purchase rare materials from merchants

**Crafting Process:**
1. Open crafting UI (`R` key)
2. Select a known recipe
3. System checks for required materials
4. Crafting takes time (progress bar)
5. Success/failure based on skill level
6. Gain XP on completion

### Advanced Combat Techniques

**Animation Canceling:**
- Start attack, dash to cancel
- Reduces attack cooldown
- Requires precise timing

**Kiting:**
- Move while attacking
- Keep enemies at range
- Essential for ranged builds

**Crowd Control Chains:**
- Stun → Slow → Root
- Prevents enemy action
- Requires coordination in multiplayer

**Burst Damage:**
- Apply all buffs
- Use ultimate ability
- All high-damage skills
- Delete boss quickly

---

## Performance Optimization

### Graphics Settings

**Resolution:**
- Lower = Better FPS
- 800x600 minimum
- 1920x1080 maximum

**Entity Limit:**
- Reduce visible entities
- Improves CPU performance

**Particle Effects:**
- Low/Medium/High
- Affects visual quality

### Troubleshooting

**Low FPS:**
1. Lower resolution
2. Reduce entity limit
3. Disable particles
4. Close background apps

**High Latency (Multiplayer):**
1. Use wired connection
2. Close bandwidth-heavy apps
3. Choose closer servers
4. Check firewall settings

**Crashes:**
1. Update to latest version
2. Check console for errors
3. Verify file integrity
4. Report bugs on GitHub

---

## Appendix

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| WASD | Movement |
| Space | Attack |
| E | Interact/Use |
| I | Inventory |
| C | Character |
| K | Skills |
| J | Quests |
| M | Map |
| Tab | Cycle Targets |
| F5 | Quick Save |
| F9 | Quick Load |
| Esc | Menu |

### Console Commands

Enable developer console with `~` key:

```
/tp x y        - Teleport to coordinates
/give item_id  - Spawn item
/level n       - Set level to n
/god           - Toggle invincibility
/noclip        - Walk through walls
/spawn enemy   - Spawn enemy
/kill_all      - Eliminate all enemies
```

(Note: Achievements disabled with cheats)

### File Locations

- **Saves**: `./saves/`
- **Config**: `./config.json`
- **Logs**: `./logs/venture.log`
- **Screenshots**: `./screenshots/`

---

**Version:** 1.0  
**Last Updated:** October 2025  
**For More Info:** See [GETTING_STARTED.md](GETTING_STARTED.md) and [API_REFERENCE.md](API_REFERENCE.md)
