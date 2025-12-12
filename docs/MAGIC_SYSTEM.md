# Magic System Documentation

**Version**: 2.0 Beta  
**Status**: Complete (Phase 2.2)  
**Date**: December 12, 2025

## Overview

The Magic System provides procedurally generated spells with deterministic generation, balanced stats, and comprehensive spell effects. Players can cast spells using keys 1-5, consuming mana and triggering visual/audio feedback.

## Architecture

### Core Components

1. **Spell Generation** (`pkg/procgen/magic/`)
   - Procedurally generates spells from templates
   - Genre-specific themes (fantasy, sci-fi, horror)
   - Deterministic with seed-based RNG
   - 6 spell types, 9 elements, 5 rarity levels

2. **Spell Casting** (`pkg/engine/spell_casting.go`)
   - ManaComponent: Tracks current/max mana and regeneration
   - SpellSlotComponent: 5 spell slots with cooldowns
   - SpellCastingSystem: Executes spells and applies effects
   - ManaRegenSystem: Passive mana regeneration

3. **Player Input** (`pkg/engine/player_spell_casting.go`)
   - PlayerSpellCastingSystem: Maps keys 1-5 to spell slots
   - Prevents casting while already casting
   - Checks mana availability before starting cast

4. **Spell Effects** (`pkg/engine/spell_effect_component.go`, `spell_effect_system.go`)
   - Advanced effects: terrain manipulation, summoning, illusion
   - Time/gravity control, elemental fusion, life drain
   - Area-of-effect and multi-target support

## Spell Types

### 1. Offensive Spells
- **Target Types**: Single, Area, Cone, Line, All Enemies
- **Effects**: Direct damage with elemental status effects
- **Elements**: Fire (burning), Ice (frozen), Lightning (shocked), Earth (poison)
- **Examples**: Fireball, Ice Shard, Lightning Bolt, Boulder Throw

### 2. Healing Spells
- **Target Types**: Self, Single Ally, Area, All Allies
- **Effects**: Restore HP to caster or allies
- **Mechanics**: Finds nearest injured ally, area healing
- **Examples**: Heal, Heal Field, Greater Heal

### 3. Defensive Spells
- **Effects**: Shield absorption, damage reduction
- **Duration**: Temporary buffs with configurable duration
- **Examples**: Shield, Barrier, Protect

### 4. Buff Spells
- **Effects**: Stat boosts (haste, strength, fortify)
- **Elements**: Wind (haste), Light (strength), Earth (fortify)
- **Duration**: 10-30 seconds based on rarity
- **Examples**: Haste, Strength, Fortify

### 5. Debuff Spells
- **Effects**: Stat reductions (weakness, vulnerability)
- **Target Types**: Single, Area
- **Duration**: 5-15 seconds
- **Examples**: Weakness, Vulnerability, Slow

### 6. Utility Spells
- **Types**: Teleport, Reveal, Speed Boost
- **Teleport**: Blink to new location (max 100 units)
- **Reveal**: Reveals fog of war in radius
- **Speed Boost**: Increases movement speed
- **Examples**: Blink, Reveal, Swift

### 7. Summon Spells (Advanced)
- **Effects**: Spawn temporary allies
- **Duration**: Time-limited summons
- **Examples**: Summon Servant, Call Spirit

## Spell Statistics

### Core Stats
- **Damage**: Offensive spell power
- **Healing**: HP restoration amount
- **Mana Cost**: Mana consumed on cast
- **Cooldown**: Time before spell can be cast again
- **Cast Time**: Channeling time before spell executes
- **Range**: Maximum distance for single-target spells
- **Area Size**: Radius for area-of-effect spells
- **Duration**: How long buffs/debuffs last
- **Required Level**: Minimum level to equip spell

### Rarity Scaling
- **Common** (1.0x): Base stats
- **Uncommon** (1.2x): +20% power
- **Rare** (1.5x): +50% power, shorter cooldowns
- **Epic** (2.0x): +100% power, faster cast time
- **Legendary** (3.0x): +200% power, enhanced effects

### Balance Formula
- **DPS** (Damage Per Second): `Damage / Cooldown`
  - Target: 15-35 DPS at level 1
  - Scales with level: `target * (1 + level * 0.2)`
- **HPS** (Healing Per Second): `Healing / Cooldown`
  - Target: 15-35 HPS at level 1
- **Mana Efficiency**: `Power / ManaCost`
  - Minimum: 1.0 power per mana

## Elemental Effects

### Fire
- **Status**: Burning (10 DPS, 3 seconds)
- **Visuals**: Orange/red flame particles rising upward
- **Audio**: Crackling fire sound

### Ice
- **Status**: Frozen (50% movement slow, 2 seconds)
- **Visuals**: Blue/white crystal particles
- **Audio**: Crystalline impact sound

### Lightning
- **Status**: Shocked (vulnerable to chaining, 2 seconds)
- **Chaining**: Hits 2 additional targets within 15 units
- **Visuals**: Fast yellow/white sparks
- **Audio**: Electric crackle

### Earth
- **Status**: Poisoned (5 DPS, 5 seconds, 30% proc chance)
- **Visuals**: Brown/green dust falling to ground
- **Audio**: Impact thud

### Wind
- **Effect**: High mobility, knockback
- **Visuals**: Fast-moving dust particles
- **Audio**: Whooshing wind

### Light
- **Effect**: Buffs, reveals, healing
- **Visuals**: Bright white/yellow particles rising
- **Audio**: Celestial chime

### Dark
- **Effect**: Debuffs, drains, curses
- **Visuals**: Purple/black smoke particles
- **Audio**: Ominous whisper

### Arcane
- **Effect**: Pure magical damage, no status
- **Visuals**: Magenta energy particles
- **Audio**: Magical hum

## Usage

### Player Controls
- **Key 1-5**: Cast spell from slot 1-5
- **Slots**: Each slot holds one spell
- **Mana Bar**: Displays current/max mana
- **Cooldown**: Visual indicator on spell icons

### Spell Loading
```go
// Generate and load 5 spells for player
engine.LoadPlayerSpells(player, seed, genreID, depth)
```

### Manual Spell Generation
```go
generator := magic.NewSpellGenerator()
params := procgen.GenerationParams{
    Difficulty: 0.5,
    Depth:      5,
    GenreID:    "fantasy",
    Custom: map[string]interface{}{
        "count": 10,
    },
}
result, err := generator.Generate(seed, params)
spells := result.([]*magic.Spell)
```

### Spell Casting Programmatically
```go
// Start casting spell from slot 0
castingSystem.StartCast(player, 0)

// Cancel current cast
castingSystem.CancelCast(player)
```

## Testing

### Test Coverage
- **Magic Generator**: 24 tests (100% passing)
  - Generation with multiple genres
  - Determinism verification
  - Depth scaling
  - Rarity distribution
  - Balance validation

- **Spell Casting**: 37 tests (100% passing)
  - Player spell casting system
  - Mana management
  - Cooldown mechanics
  - Elemental effects
  - Buff/debuff application
  - Healing targeting

### Running Tests
```bash
# Test magic generator
go test ./pkg/procgen/magic/... -v

# Test spell casting systems
go test ./pkg/engine -run "Spell|Mana" -v

# Test all magic-related code
go test ./pkg/procgen/magic/... ./pkg/engine -run "Spell|Mana" -v
```

## Performance

### Benchmarks
- Spell generation: <1ms per 100 spells
- Spell casting: <0.1ms per cast
- Effect processing: <0.05ms per effect per frame
- Mana regeneration: <0.01ms per entity per frame

### Memory Usage
- SpellComponent: ~200 bytes
- ManaComponent: ~24 bytes
- SpellSlotComponent: ~200 bytes
- Total per player: ~500 bytes

## Future Enhancements

### Potential Features
1. **Spell Crafting**: Combine spells to create new effects
2. **Spell Modification**: Apply affixes to spells (faster cast, more damage)
3. **Spell Scrolls**: One-time use spell items
4. **Spell Books**: Learn new spells from found books
5. **Metamagic**: Modify active spells on-the-fly
6. **Spell Combos**: Chain spells for bonus effects
7. **Spell Schools**: Specialize in fire, ice, lightning, etc.
8. **Ultimate Spells**: High-cost, high-impact spells with long cooldowns

## References

- **Generator Implementation**: `pkg/procgen/magic/generator.go`
- **Spell Templates**: `pkg/procgen/magic/advanced_templates.go`
- **Balance Configuration**: `pkg/procgen/magic/balance.go`
- **Casting System**: `pkg/engine/spell_casting.go`
- **Player Input**: `pkg/engine/player_spell_casting.go`
- **Effect System**: `pkg/engine/spell_effect_system.go`
- **Tests**: `pkg/procgen/magic/magic_test.go`, `pkg/engine/spell_casting_test.go`
