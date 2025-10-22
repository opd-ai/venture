# Entity Generation System

The entity generation system creates procedural monsters, NPCs, and other game entities with varied stats, abilities, and behaviors. All generation is deterministic based on seed values, ensuring consistency across clients in multiplayer games.

## Features

- **Diverse Entity Types**: Monsters, NPCs, Bosses, and Minions
- **Stat System**: Health, Damage, Defense, Speed, and Level
- **Rarity System**: Common, Uncommon, Rare, Epic, Legendary
- **Size Classifications**: Tiny, Small, Medium, Large, Huge
- **Genre Support**: Fantasy and Sci-Fi templates (extensible)
- **Deterministic**: Same seed always produces same entities
- **Level Scaling**: Stats scale with depth and difficulty
- **High Performance**: ~14.5μs per 10-entity generation

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/opd-ai/venture/pkg/procgen"
    "github.com/opd-ai/venture/pkg/procgen/entity"
)

func main() {
    // Create generator
    gen := entity.NewEntityGenerator()
    
    // Set parameters
    params := procgen.GenerationParams{
        Difficulty: 0.5,  // 0.0-1.0
        Depth:      5,    // Dungeon level/depth
        GenreID:    "fantasy",
        Custom: map[string]interface{}{
            "count": 10,  // Number of entities to generate
        },
    }
    
    // Generate entities
    result, err := gen.Generate(12345, params)
    if err != nil {
        panic(err)
    }
    
    entities := result.([]*entity.Entity)
    
    // Use the entities
    for _, e := range entities {
        fmt.Printf("%s (Lv.%d): HP=%d, DMG=%d\n", 
            e.Name, e.Stats.Level, e.Stats.MaxHealth, e.Stats.Damage)
    }
}
```

### CLI Tool

The `entitytest` command-line tool lets you generate and visualize entities without writing code:

```bash
# Generate fantasy entities
go run ./cmd/entitytest -genre fantasy -count 20 -depth 5 -seed 12345

# Generate sci-fi entities with verbose output
go run ./cmd/entitytest -genre scifi -count 15 -depth 10 -verbose

# Export to file
go run ./cmd/entitytest -genre fantasy -count 100 -output entities.txt

# Show all options
go run ./cmd/entitytest -help
```

## Entity Types

### Monster
Standard hostile entities with balanced stats. Make up the majority of encounters.

**Example Templates:**
- Fantasy: Orcs, Skeletons, Zombies, Ghouls
- Sci-Fi: Androids, Cyborgs, Battle Mechs

### Boss
Rare, powerful entities with enhanced stats (2-3x normal). Usually one per major area.

**Example Templates:**
- Fantasy: Ancient Dragons, Demon Lords, Lich Kings
- Sci-Fi: Titan Mechs, Omega Units, Prime Destroyers

### Minion
Weak, common entities often found in groups. Fast but fragile.

**Example Templates:**
- Fantasy: Goblins, Kobolds, Imps
- Sci-Fi: Scout Drones, Bots, Probes

### NPC
Non-hostile characters for trading, quests, and lore.

**Example Templates:**
- Fantasy: Merchants, Guards, Priests, Wizards
- Sci-Fi: (Future implementation)

## Stat System

### Core Stats

- **Health/MaxHealth**: Hit points (10-500+)
- **Damage**: Base attack damage (2-60+)
- **Defense**: Damage reduction (0-30+)
- **Speed**: Movement/attack rate (0.5-2.0)
- **Level**: Power level (scales with depth)

### Stat Modifiers

Stats are modified by:
1. **Template Base Range**: Each template has min/max ranges
2. **Level Scaling**: +15% per level above 1
3. **Rarity Multiplier**: 1.0x (Common) to 3.0x (Legendary)
4. **Difficulty**: Affects level calculation

## Rarity System

Rarity determines stat multipliers and drop chances:

| Rarity | Multiplier | Symbol | Spawn Chance |
|--------|------------|--------|--------------|
| Common | 1.0x | ● | ~60% |
| Uncommon | 1.2x | ◆ | ~25% |
| Rare | 1.5x | ★ | ~10% |
| Epic | 2.0x | ◈ | ~4% |
| Legendary | 3.0x | ♛ | ~1% |

Bosses are always Rare or higher. Rarity chance increases with depth.

## Genre Templates

### Fantasy Genre

Includes traditional fantasy creatures and characters:
- Minions: Goblins, Kobolds, Imps, Sprites
- Monsters: Orcs, Skeletons, Zombies, Ghouls, Ogres, Trolls, Minotaurs
- Bosses: Ancient Dragons, Demon Lords, Lich Kings, Elder Wyrms
- NPCs: Merchants, Guards, Priests, Wizards

### Sci-Fi Genre

Includes robotic and futuristic entities:
- Minions: Scout Drones, Bots, Probes
- Monsters: Combat Androids, Security Cyborgs, Battle Mechs
- Bosses: Titan Mechs, Colossus Units, Omega Destroyers

### Adding Custom Genres

```go
// Define custom templates
customTemplates := []entity.EntityTemplate{
    {
        BaseType:     entity.TypeMonster,
        BaseSize:     entity.SizeMedium,
        NamePrefixes: []string{"Alien", "Mutant", "Beast"},
        NameSuffixes: []string{"Hunter", "Warrior", "Predator"},
        Tags:         []string{"hostile", "organic"},
        HealthRange:  [2]int{50, 100},
        DamageRange:  [2]int{10, 20},
        DefenseRange: [2]int{5, 10},
        SpeedRange:   [2]float64{0.8, 1.2},
    },
    // ... more templates
}

// Register with generator
gen := entity.NewEntityGenerator()
gen.templates["custom"] = customTemplates
```

## Entity Properties

### Methods

```go
entity := entities[0]

// Check if hostile to player
if entity.IsHostile() {
    // Attack logic
}

// Check if boss
if entity.IsBoss() {
    // Special boss mechanics
}

// Get threat level (0-100)
threat := entity.GetThreatLevel()
```

### Fields

```go
type Entity struct {
    Name    string        // "Ancient Dragon"
    Type    EntityType    // TypeBoss
    Size    EntitySize    // SizeHuge
    Rarity  Rarity        // RarityLegendary
    Stats   Stats         // Health, Damage, Defense, Speed, Level
    Seed    int64         // Generation seed
    Tags    []string      // ["boss", "elite", "legendary"]
}
```

## Integration Examples

### Spawn Entities in Terrain

```go
// Generate terrain
terrainGen := terrain.NewBSPGenerator()
result, _ := terrainGen.Generate(seed, terrainParams)
terr := result.(*terrain.Terrain)

// Generate entities for terrain
entityGen := entity.NewEntityGenerator()
entityParams := procgen.GenerationParams{
    Difficulty: 0.5,
    Depth:      5,
    GenreID:    "fantasy",
    Custom:     map[string]interface{}{"count": len(terr.Rooms)},
}
result, _ = entityGen.Generate(seed+1, entityParams)
entities := result.([]*entity.Entity)

// Place one entity per room
for i, room := range terr.Rooms {
    if i < len(entities) {
        cx, cy := room.Center()
        // Place entities[i] at (cx, cy)
    }
}
```

### Dynamic Encounter Generation

```go
func generateEncounter(depth int, difficulty float64) []*entity.Entity {
    gen := entity.NewEntityGenerator()
    
    // More entities at higher depths
    count := 3 + depth/2
    
    params := procgen.GenerationParams{
        Difficulty: difficulty,
        Depth:      depth,
        GenreID:    "fantasy",
        Custom:     map[string]interface{}{"count": count},
    }
    
    seed := time.Now().UnixNano()
    result, _ := gen.Generate(seed, params)
    
    return result.([]*entity.Entity)
}
```

## Testing

Run the test suite:

```bash
# Run tests
go test ./pkg/procgen/entity/...

# With coverage
go test -cover ./pkg/procgen/entity/...

# With race detection
go test -race ./pkg/procgen/entity/...

# Benchmarks
go test -bench=. ./pkg/procgen/entity/...
```

Test results:
- ✅ 14 tests, all passing
- ✅ 95.9% code coverage
- ✅ Deterministic generation verified
- ✅ ~14.5μs per 10-entity batch

## Performance

Benchmarks on AMD EPYC 7763:
- **Generation**: ~14.5μs per 10 entities (1.45μs per entity)
- **Memory**: Minimal allocations
- **Determinism**: 100% reproducible with same seed

For 1000 entities:
- Generation time: ~1.5ms
- Memory usage: ~150KB

## Best Practices

### Seed Management

```go
// Use different seed categories for different systems
seedGen := procgen.NewSeedGenerator(worldSeed)

terrainSeed := seedGen.GetSeed("terrain", roomIndex)
entitySeed := seedGen.GetSeed("entity", roomIndex)
lootSeed := seedGen.GetSeed("loot", roomIndex)
```

### Level Scaling

```go
// Entities scale with dungeon depth
for depth := 1; depth <= 20; depth++ {
    params.Depth = depth
    // Entities will be level 1-3 at depth 1
    // Entities will be level 18-22 at depth 20
}
```

### Difficulty Balance

```go
// Adjust difficulty for player progression
params.Difficulty = 0.3  // Easy: 30% level scaling
params.Difficulty = 0.5  // Normal: 50% level scaling
params.Difficulty = 1.0  // Hard: 100% level scaling
```

## Future Enhancements

Planned improvements:
- [ ] Equipment/loot integration
- [ ] AI behavior patterns
- [ ] Special abilities/skills
- [ ] Elemental affinities
- [ ] More genre templates (horror, cyberpunk, post-apocalyptic)
- [ ] Elite/champion variants
- [ ] Faction/alignment system

## Related Systems

- **Terrain Generation**: `pkg/procgen/terrain` - Generate dungeons and caves
- **Item Generation**: `pkg/procgen/item` - Generate weapons and armor
- **ECS System**: `pkg/engine` - Entity-Component-System framework

## API Reference

See [GoDoc](https://pkg.go.dev/github.com/opd-ai/venture/pkg/procgen/entity) for complete API documentation.
