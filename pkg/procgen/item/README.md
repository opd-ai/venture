# Item Generation Package

The `item` package provides procedural generation of game items for the Venture RPG, including weapons, armor, consumables, and accessories.

## Features

- **Multiple Item Types**: Weapons, armor, consumables, and accessories
- **Rarity System**: Five rarity levels from common to legendary
- **Stat Scaling**: Items scale with dungeon depth and difficulty
- **Genre Support**: Fantasy and sci-fi item templates
- **Deterministic Generation**: Same seed produces same items
- **Comprehensive Stats**: Damage, defense, attack speed, durability, value, and more

## Item Types

### Weapons
- **Sword**: Balanced melee weapon
- **Axe**: Heavy, powerful weapon
- **Bow**: Ranged weapon
- **Staff**: Magical weapon
- **Dagger**: Fast, light weapon
- **Spear**: Reach weapon

### Armor
- **Helmet**: Head protection
- **Chest**: Torso protection
- **Legs**: Leg protection
- **Boots**: Foot protection
- **Gloves**: Hand protection
- **Shield**: Additional defense

### Consumables
- **Potion**: Health/buff restoration
- **Scroll**: One-time spell effects
- **Food**: Health over time
- **Bomb**: Area damage

## Rarity System

| Rarity | Stat Multiplier | Drop Chance (Depth 1) | Drop Chance (Depth 20) |
|--------|----------------|----------------------|------------------------|
| Common | 1.0x | 50% | 20% |
| Uncommon | 1.2x | 30% | 30% |
| Rare | 1.5x | 13% | 25% |
| Epic | 2.0x | 5% | 15% |
| Legendary | 3.0x | 2% | 10% |

Higher dungeon depths increase the chance of rarer items.

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/opd-ai/venture/pkg/procgen"
    "github.com/opd-ai/venture/pkg/procgen/item"
)

func main() {
    // Create generator
    generator := item.NewItemGenerator()
    
    // Set up generation parameters
    params := procgen.GenerationParams{
        Depth:      5,           // Dungeon level 5
        Difficulty: 0.5,         // Medium difficulty
        GenreID:    "fantasy",   // Fantasy genre
        Custom: map[string]interface{}{
            "count": 20,         // Generate 20 items
            "type":  "weapon",   // Filter to weapons only
        },
    }
    
    // Generate items
    result, err := generator.Generate(12345, params)
    if err != nil {
        panic(err)
    }
    
    items := result.([]*item.Item)
    
    // Display items
    for _, itm := range items {
        fmt.Printf("%s (%s)\n", itm.Name, itm.Rarity)
        fmt.Printf("  Type: %s\n", itm.Type)
        if itm.Type == item.TypeWeapon {
            fmt.Printf("  Damage: %d\n", itm.Stats.Damage)
            fmt.Printf("  Speed: %.2f\n", itm.Stats.AttackSpeed)
        }
        fmt.Printf("  Value: %d gold\n", itm.Stats.Value)
        fmt.Println()
    }
}
```

## CLI Tool

The `itemtest` command-line tool allows testing item generation:

```bash
# Build the tool
go build -o itemtest ./cmd/itemtest

# Generate fantasy weapons
./itemtest -genre fantasy -count 20 -type weapon -seed 12345

# Generate sci-fi armor at high depth
./itemtest -genre scifi -count 15 -type armor -depth 10

# Show detailed information
./itemtest -genre fantasy -count 10 -verbose

# Save to file
./itemtest -genre fantasy -count 100 -output items.txt
```

### CLI Options

- `-genre`: Item genre (fantasy, scifi) - default: fantasy
- `-count`: Number of items to generate - default: 20
- `-depth`: Dungeon depth (affects level and rarity) - default: 5
- `-type`: Filter by type (weapon, armor, consumable) - optional
- `-seed`: Random seed (0 for random) - default: 0
- `-verbose`: Show detailed information - default: false
- `-output`: Output file (default: stdout)

## Stat Generation

### Damage (Weapons)
- Base damage from template
- Scaled by depth (+10% per level)
- Scaled by rarity (1.2x to 3.0x)
- Modified by difficulty (0.8x to 1.2x)

### Defense (Armor)
- Base defense from template
- Scaled by depth (+10% per level)
- Scaled by rarity (1.2x to 3.0x)
- Modified by difficulty (0.8x to 1.2x)

### Attack Speed (Weapons)
- Base speed from template
- Slightly increased for rare items (+0.05 per rarity level)

### Durability
- Base durability from template
- Increased for rare items (+20 per rarity level)
- Starts at maximum

### Value
- Base value from template
- Scaled by depth and rarity
- Reduced if item is damaged

## Genre Templates

### Fantasy
- **Weapons**: Iron Sword, Battle Axe, Hunter's Bow, Wizard's Staff, Shadow Dagger
- **Armor**: Chain Mail, Steel Helmet, Iron Shield
- **Consumables**: Health Potion, Scroll of Fireball

### Sci-Fi
- **Weapons**: Plasma Blade, Laser Rifle, Ion Cannon
- **Armor**: Combat Suit, HUD Helmet, Tactical Vest
- **Consumables**: (Uses fantasy templates)

## Testing

Run the test suite:

```bash
# Run all tests
go test ./pkg/procgen/item/

# Run with coverage
go test -cover ./pkg/procgen/item/

# Run with verbose output
go test -v ./pkg/procgen/item/
```

Current test coverage: **93.8%**

## Integration

The item generator implements the `procgen.Generator` interface and can be used alongside other procedural generation systems:

```go
// Generate terrain
terrainGen := terrain.NewBSPGenerator()
terrainResult, _ := terrainGen.Generate(seed, terrainParams)

// Generate entities for the terrain
entityGen := entity.NewEntityGenerator()
entityResult, _ := entityGen.Generate(seed+1, entityParams)

// Generate loot items
itemGen := item.NewItemGenerator()
itemResult, _ := itemGen.Generate(seed+2, itemParams)
```

## Future Enhancements

- Item sets with bonuses
- Enchantments and modifiers
- Crafting system integration
- Item quality levels beyond rarity
- More genre templates (cyberpunk, horror, post-apocalyptic)
- Unique/artifact items with special properties
