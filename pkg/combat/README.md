# Combat Package

## Overview

The `combat` package provides core combat mechanics for the Venture game engine, including damage calculation, entity statistics, and combat resolution interfaces.

## Package Structure

```
pkg/combat/
├── constants.go       - Damage type enumeration (DamageType) and constants
├── types.go          - Core data structures (Damage, Stats) and constructors
├── interfaces.go     - Combat resolution interface (CombatResolver)
├── doc.go            - Package documentation
├── interfaces_test.go - Comprehensive test suite (~98% coverage)
└── AUDIT.md          - Implementation gap audit and recommendations
```

## Core Types

### DamageType
Enumeration of damage types supported by the combat system:
- `DamagePhysical` - Physical/melee damage
- `DamageMagical` - Magical damage
- `DamageFire` - Fire elemental damage
- `DamageIce` - Ice elemental damage
- `DamageLightning` - Lightning elemental damage
- `DamagePoison` - Poison/toxic damage

### Damage
Represents a single damage event:
```go
type Damage struct {
    Amount   float64    // Damage amount
    Type     DamageType // Type of damage
    SourceID uint64     // Entity causing damage
    TargetID uint64     // Entity receiving damage
}
```

### Stats
Represents entity combat statistics:
```go
type Stats struct {
    HP, MaxHP         float64               // Health
    Mana, MaxMana     float64               // Mana/energy
    Attack, MagicPower float64              // Offensive stats
    CritChance, CritDamage float64          // Critical hit stats
    Defense, MagicDefense, Evasion float64  // Defensive stats
    Speed             float64               // Movement speed
    Resistances       map[DamageType]float64 // Damage resistances
}
```

## Interfaces

### CombatResolver
Contract for combat calculation implementations:
```go
type CombatResolver interface {
    CalculateDamage(damage Damage, targetStats *Stats) float64
    ResolveCombat(attackerID, defenderID uint64) []Damage
}
```

**Note**: This interface is designed for external implementation. Concrete implementations should be provided by game system packages (e.g., `pkg/engine`).

## Usage Examples

### Creating Stats
```go
// Create default stats
stats := combat.NewStats()

// Customize stats
stats.Attack = 25
stats.Defense = 15
stats.Resistances[combat.DamageFire] = 0.5 // 50% fire resistance
```

### Working with Damage
```go
// Create a damage event
damage := combat.Damage{
    Amount:   100,
    Type:     combat.DamagePhysical,
    SourceID: attackerID,
    TargetID: defenderID,
}
```

## Testing

Run package tests:
```bash
go test ./pkg/combat/...
```

Current test coverage: **~98%**

## Implementation Status

✅ **Complete**: Core data types and interfaces  
⚠️  **Pending**: CombatResolver concrete implementations (intentionally in other packages)

See [AUDIT.md](./AUDIT.md) for detailed implementation gap analysis and recommendations.

## Related Packages

- `pkg/engine` - Main game engine with combat systems
- `pkg/procgen` - Procedural generation of enemies and encounters

## Design Philosophy

This package follows the principle of **separation of concerns**:
- **Data structures** (Damage, Stats) are defined here
- **Interfaces** (CombatResolver) provide contracts for implementations
- **Concrete implementations** live in game system packages
- **Game logic** is decoupled from data definitions

This design allows:
- Easy testing with mock implementations
- Flexible combat system variations
- Clear dependency boundaries
- Type-safe combat interactions across the codebase
