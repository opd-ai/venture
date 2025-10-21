# Magic/Spell Generation System

The magic generation system provides procedural generation of diverse spells and abilities for the Venture action-RPG. It generates balanced, themed spells that scale with game progression.

## Features

- **7 Spell Types**: Offensive, Defensive, Healing, Buff, Debuff, Utility, Summon
- **9 Elements**: Fire, Ice, Lightning, Earth, Wind, Light, Dark, Arcane, None
- **7 Target Patterns**: Self, Single, Area, Cone, Line, All Allies, All Enemies
- **5 Rarity Levels**: Common, Uncommon, Rare, Epic, Legendary
- **Deterministic Generation**: Same seed always produces same spells
- **Genre Support**: Fantasy and Sci-Fi themes
- **Balanced Scaling**: Power scales with depth, difficulty, and rarity

## Quick Start

### Generate Spells Programmatically

```go
import (
    "github.com/opd-ai/venture/pkg/procgen"
    "github.com/opd-ai/venture/pkg/procgen/magic"
)

// Create generator
gen := magic.NewSpellGenerator()

// Set parameters
params := procgen.GenerationParams{
    Difficulty: 0.5,  // 0.0 to 1.0
    Depth:      10,   // Game progression level
    GenreID:    "fantasy",
    Custom: map[string]interface{}{
        "count": 20,  // Number of spells
    },
}

// Generate spells
result, err := gen.Generate(12345, params)
if err != nil {
    log.Fatal(err)
}

spells := result.([]*magic.Spell)

// Use the spells
for _, spell := range spells {
    fmt.Printf("%s: %s (Power: %d)\n", 
        spell.Name, spell.Description, spell.GetPowerLevel())
}
```

### Test with CLI Tool

```bash
# Build the tool
go build -o magictest ./cmd/magictest

# Generate fantasy spells
./magictest -genre fantasy -count 20 -depth 5 -seed 12345

# Generate sci-fi spells with verbose output
./magictest -genre scifi -count 15 -depth 10 -verbose

# Filter by spell type
./magictest -type offensive -count 30

# Save to file
./magictest -count 100 -output spells.txt
```

## Spell Types

### Offensive Spells
Damage-dealing spells for combat.

**Fantasy Examples:**
- Fire Bolt: Single-target fire damage
- Ice Storm: Area-effect ice damage
- Lightning Strike: Line-effect lightning damage
- Stone Barrage: Heavy earth damage
- Shadow Wave: Cone of dark damage

**Sci-Fi Examples:**
- Plasma Beam: Precision energy weapon
- Fusion Blast: Explosive area damage
- Cryo Ray: Freezing beam attack

### Support Spells

**Healing:**
- Fantasy: Heal Touch, Divine Grace
- Sci-Fi: Nano Injection, Medical Field

**Defensive:**
- Fantasy: Mana Shield, Arcane Barrier
- Sci-Fi: Energy Field, Kinetic Shield

**Buffs:**
- Fantasy: Haste Blessing, Swift Enhancement
- Sci-Fi: Combat Stimulant, Tactical Boost

**Debuffs:**
- Fantasy: Weakness Touch, Curse Affliction
- Sci-Fi: System Disruption, Neural Damper

## Spell Statistics

### Core Stats
- **Damage**: Direct damage for offensive spells (0-500+)
- **Healing**: Health restoration for healing spells (0-500+)
- **Mana Cost**: Resource cost to cast (10-100+)
- **Cooldown**: Time before spell can be recast (2-40 seconds)
- **Cast Time**: Casting duration (0.3-2.5 seconds)

### Advanced Stats
- **Range**: Maximum distance spell can reach (5-40 units)
- **Area Size**: Radius for area effects (3-12 units)
- **Duration**: Length of buffs/debuffs (10-70 seconds)
- **Required Level**: Minimum character level (1-50+)

## Scaling System

### Depth Scaling
Spells become more powerful as the player progresses:
```
Power = Base × (1.0 + Depth × 0.1)
```

At depth 10, spells are 2x as powerful as depth 0.

### Difficulty Scaling
Higher difficulty increases challenge and rewards:
```
Scale = 0.8 + Difficulty × 0.4
```

### Rarity Scaling
Higher rarity provides superior stats:

| Rarity    | Power Mult | Cooldown Div | Cost Mult |
|-----------|------------|--------------|-----------|
| Common    | 1.0×       | 1.0×         | 1.0×      |
| Uncommon  | 1.25×      | 1.25×        | 1.25×     |
| Rare      | 1.5×       | 1.5×         | 1.5×      |
| Epic      | 1.75×      | 1.75×        | 1.75×     |
| Legendary | 2.0×       | 2.0×         | 2.0×      |

## Element System

### Fantasy Elements

**Fire** 🔥
- High damage
- Fast cast time
- Medium range
- Effect: Burning damage over time

**Ice** ❄️
- Moderate damage
- Area effects
- Medium-long range
- Effect: Slowing/freezing

**Lightning** ⚡
- High damage
- Very fast cast
- Long range
- Effect: Chain to nearby targets

**Earth** 🗻
- Very high damage
- Slow cast time
- Medium range
- Effect: Stunning/knockback

**Wind** 💨
- Moderate damage
- Fast cast time
- Good for mobility
- Effect: Speed buffs

**Light** ✨
- Healing/holy damage
- Medium cast time
- Short-medium range
- Effect: Healing over time

**Dark** 🌑
- Damage over time
- Medium cast time
- Cone/area effects
- Effect: Fear/debuffs

**Arcane** 🔮
- Pure magic
- Versatile stats
- Good for shields
- Effect: Mana regeneration

### Sci-Fi Element Mapping

Sci-fi spells use technological themes:
- Fire → Explosive/Thermal weapons
- Ice → Cryo/Freeze technology
- Lightning → Plasma/Ion weapons
- Earth → Kinetic/Mass weapons
- Wind → Boost/Propulsion systems
- Light → Medical/Healing tech
- Dark → EMP/Disruption systems
- Arcane → Energy/Force fields

## Target Patterns

### Single Target
Affects one entity (ally or enemy).
- High damage/healing
- Precise targeting
- Low mana cost

### Area of Effect (AoE)
Affects all targets in a radius.
- Moderate damage/healing
- Area Size stat determines radius
- High mana cost
- Best for groups

### Cone
Affects targets in a cone shape.
- Moderate damage
- Medium range
- Good for multiple enemies in front

### Line
Affects targets in a straight line.
- Good damage
- Long range
- Pierces through enemies

### Self
Affects only the caster.
- Buffs and shields
- Instant cast
- Low mana cost

### All Allies / All Enemies
Affects every ally or enemy on the field.
- Low individual effect
- Very high mana cost
- Long cooldown
- Raid/boss mechanics

## Rarity Distribution

Rarity chances are influenced by depth and difficulty:

```
Base Chances (Depth 0, Difficulty 0.0):
- Common:    50%
- Uncommon:  25%
- Rare:      15%
- Epic:       7%
- Legendary:  3%

High Level (Depth 20, Difficulty 1.0):
- Common:    10%
- Uncommon:  20%
- Rare:      30%
- Epic:      25%
- Legendary: 15%
```

## Validation

All generated spells are validated to ensure:
- Non-empty names
- Valid type, element, target, and rarity
- Positive required level (≥1)
- Non-negative stats
- Offensive spells have damage > 0
- Healing spells have healing > 0
- Appropriate stat values for spell type

## Performance

Spell generation is highly optimized:

```
Benchmark Results (100 spells):
- Generation: ~50-100 µs per spell
- Validation: ~1-5 µs per spell
- Memory: ~2 KB per spell
```

Perfect for real-time generation during gameplay.

## Testing

The package includes comprehensive tests:

```bash
# Run tests
go test ./pkg/procgen/magic/

# Run with coverage
go test -cover ./pkg/procgen/magic/
# Coverage: 91.9%

# Run benchmarks
go test -bench=. ./pkg/procgen/magic/
```

## Examples

### Generate Balanced Spells

```go
params := procgen.GenerationParams{
    Difficulty: 0.5,
    Depth:      10,
    GenreID:    "fantasy",
    Custom:     map[string]interface{}{"count": 10},
}

spells, _ := gen.Generate(seed, params)
```

### Find Powerful Spells

```go
for _, spell := range spells {
    if spell.GetPowerLevel() >= 80 && spell.Rarity >= magic.RarityEpic {
        fmt.Printf("Found powerful spell: %s\n", spell.Name)
    }
}
```

### Filter by Element

```go
fireSpells := []*magic.Spell{}
for _, spell := range spells {
    if spell.Element == magic.ElementFire {
        fireSpells = append(fireSpells, spell)
    }
}
```

## Integration with Game Systems

### Player Progression
```go
playerLevel := 15
availableSpells := []*magic.Spell{}

for _, spell := range allSpells {
    if spell.Stats.RequiredLevel <= playerLevel {
        availableSpells = append(availableSpells, spell)
    }
}
```

### Spell Learning
```go
if player.Gold >= 100 && player.Level >= spell.Stats.RequiredLevel {
    player.LearnSpell(spell)
    player.Gold -= 100
}
```

### Combat System
```go
func CastSpell(caster *Entity, spell *magic.Spell, target *Entity) {
    if caster.Mana < spell.Stats.ManaCost {
        return // Not enough mana
    }
    
    caster.Mana -= spell.Stats.ManaCost
    
    if spell.IsOffensive() {
        target.Health -= spell.Stats.Damage
    } else if spell.Type == magic.TypeHealing {
        target.Health += spell.Stats.Healing
        if target.Health > target.MaxHealth {
            target.Health = target.MaxHealth
        }
    }
    
    // Start cooldown timer
    caster.SpellCooldowns[spell.Name] = spell.Stats.Cooldown
}
```

## Architecture

The magic generation system follows these design principles:

1. **Deterministic**: Same seed produces same spells
2. **Scalable**: Smooth power curve with depth
3. **Balanced**: No overpowered combinations
4. **Diverse**: Wide variety of spell types
5. **Validated**: All output is validated
6. **Testable**: 91.9% test coverage
7. **Fast**: Sub-millisecond generation

## Future Enhancements

Potential additions in future phases:
- Spell combinations (multicast)
- Elemental interactions (fire + ice = steam)
- Metamagic modifiers (quickcast, empower)
- Spell schools (necromancy, illusion)
- Custom spell crafting
- Spell mutations/evolution
- Conditional effects (triggers)

## Related Systems

- **Entity Generation**: Monsters that cast spells
- **Item Generation**: Spell scrolls and magic items
- **Skill Tree**: Spell specialization paths
- **Genre System**: Theme-appropriate spells
- **Combat System**: Spell casting mechanics

## Support

For questions or issues:
- Review package documentation: `go doc github.com/opd-ai/venture/pkg/procgen/magic`
- Check test cases for examples
- Use CLI tool for interactive testing
- See main README.md for project overview

## License

Part of the Venture project. See LICENSE file for details.
