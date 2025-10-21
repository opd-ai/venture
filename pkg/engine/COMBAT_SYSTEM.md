# Combat System

This document describes the combat system implemented in Phase 5 of the Venture project.

## Overview

The combat system provides a complete framework for handling combat interactions between entities, including:

- **Health management** - Track entity health and death states
- **Attack mechanics** - Range-based attacks with cooldowns
- **Damage calculation** - Apply stats, resistances, critical hits
- **Status effects** - Buffs and debuffs with duration and tick mechanics
- **Team system** - Identify allies and enemies

## Components

### HealthComponent

Tracks an entity's current and maximum health.

```go
type HealthComponent struct {
    Current float64
    Max     float64
}
```

**Methods:**
- `IsAlive() bool` - Returns true if health > 0
- `IsDead() bool` - Returns true if health <= 0
- `Heal(amount float64)` - Increase health (capped at max)
- `TakeDamage(amount float64)` - Decrease health (minimum 0)

**Example:**
```go
health := &engine.HealthComponent{Current: 100, Max: 100}
health.TakeDamage(30)  // Current = 70
health.Heal(20)        // Current = 90
```

### StatsComponent

Contains combat statistics for an entity.

```go
type StatsComponent struct {
    Attack       float64
    Defense      float64
    MagicPower   float64
    MagicDefense float64
    CritChance   float64  // 0.0 to 1.0
    CritDamage   float64  // Multiplier (e.g., 2.0 = 200%)
    Evasion      float64  // 0.0 to 1.0
    Resistances  map[combat.DamageType]float64
}
```

**Methods:**
- `NewStatsComponent()` - Create with default values
- `GetResistance(damageType) float64` - Get resistance for a damage type

**Example:**
```go
stats := engine.NewStatsComponent()
stats.Attack = 25
stats.CritChance = 0.15  // 15% crit chance
stats.CritDamage = 2.0   // 200% damage on crit
stats.Resistances[combat.DamageFire] = 0.5  // 50% fire resistance
```

### AttackComponent

Defines an entity's attack capabilities.

```go
type AttackComponent struct {
    Damage        float64
    DamageType    combat.DamageType
    Range         float64
    Cooldown      float64
    CooldownTimer float64
}
```

**Damage Types:**
- `DamagePhysical` - Physical damage (uses Attack/Defense)
- `DamageMagical` - Magic damage (uses MagicPower/MagicDefense)
- `DamageFire`, `DamageIce`, `DamageLightning`, `DamagePoison` - Elemental damage

**Methods:**
- `CanAttack() bool` - Check if cooldown is ready
- `ResetCooldown()` - Start cooldown timer
- `UpdateCooldown(deltaTime)` - Update cooldown (handled by CombatSystem)

**Example:**
```go
attack := &engine.AttackComponent{
    Damage:     25,
    DamageType: combat.DamagePhysical,
    Range:      50,
    Cooldown:   1.0,  // 1 second cooldown
}
```

### StatusEffectComponent

Represents a temporary buff or debuff.

```go
type StatusEffectComponent struct {
    EffectType   string   // "poison", "burn", "regeneration", etc.
    Duration     float64  // Seconds remaining
    Magnitude    float64  // Effect strength
    TickInterval float64  // Seconds between ticks (0 = one-time)
    NextTick     float64  // Time until next tick
}
```

**Built-in Effect Types:**
- `"poison"` - Damage over time
- `"burn"` - Damage over time
- `"regeneration"` - Healing over time

**Methods:**
- `IsExpired() bool` - Check if duration has expired
- `Update(deltaTime) bool` - Update duration and ticks

**Example:**
```go
// Apply poison: 10 damage per second for 5 seconds
combatSystem.ApplyStatusEffect(entity, "poison", 5.0, 10.0, 1.0)
```

### TeamComponent

Identifies which team an entity belongs to.

```go
type TeamComponent struct {
    TeamID int  // 0 = neutral, 1+ = team ID
}
```

**Methods:**
- `IsAlly(otherTeam int) bool` - Check if same team
- `IsEnemy(otherTeam int) bool` - Check if different teams (excludes neutral)

**Example:**
```go
player := &engine.TeamComponent{TeamID: 1}
enemy := &engine.TeamComponent{TeamID: 2}
neutral := &engine.TeamComponent{TeamID: 0}

player.IsEnemy(2)    // true
player.IsEnemy(0)    // false (neutral)
neutral.IsEnemy(1)   // false (neutral)
```

## Combat System

The `CombatSystem` handles all combat logic and status effect processing.

### Creating a Combat System

```go
combatSystem := engine.NewCombatSystem(seed int64)
world.AddSystem(combatSystem)
```

The seed ensures deterministic combat outcomes (critical hits, evasion).

### Attacking

```go
hit := combatSystem.Attack(attacker, target *Entity) bool
```

Returns `true` if the attack hit, `false` if it missed or was invalid.

**Attack Requirements:**
1. Attacker has `AttackComponent` and attack is ready (cooldown expired)
2. Target has `HealthComponent` and is alive
3. Target is within attack range (if both have positions)

**Attack Process:**
1. Check evasion - target may dodge based on `Evasion` stat
2. Calculate base damage - `AttackComponent.Damage` + attacker stats
3. Apply critical hit - random chance based on `CritChance`
4. Apply defense - subtract `Defense` or `MagicDefense`
5. Apply resistance - multiply by `(1.0 - resistance)`
6. Apply damage to target (minimum 1 damage)
7. Reset attacker's cooldown

**Example:**
```go
if combatSystem.Attack(warrior, goblin) {
    fmt.Println("Hit!")
} else {
    fmt.Println("Miss!")
}
```

### Status Effects

```go
combatSystem.ApplyStatusEffect(target, effectType, duration, magnitude, tickInterval)
```

**Parameters:**
- `target` - Entity to apply effect to
- `effectType` - Effect identifier ("poison", "burn", "regeneration", etc.)
- `duration` - Total duration in seconds
- `magnitude` - Effect strength (damage/heal per tick)
- `tickInterval` - Seconds between ticks (0 for one-time effects)

**Example:**
```go
// Poison: 10 damage per second for 5 seconds
combatSystem.ApplyStatusEffect(enemy, "poison", 5.0, 10.0, 1.0)

// Buff: instant effect (no ticking)
combatSystem.ApplyStatusEffect(player, "speed_boost", 10.0, 50.0, 0)
```

### Healing

```go
combatSystem.Heal(target *Entity, amount float64)
```

Heals the target by the specified amount (capped at max health).

**Example:**
```go
combatSystem.Heal(player, 50)
```

### Callbacks

The combat system supports callbacks for damage and death events.

```go
// Damage callback
combatSystem.SetDamageCallback(func(attacker, target *Entity, damage float64) {
    fmt.Printf("%d dealt %0.f damage to %d\n", attacker.ID, damage, target.ID)
})

// Death callback
combatSystem.SetDeathCallback(func(entity *Entity) {
    fmt.Printf("Entity %d died!\n", entity.ID)
    world.RemoveEntity(entity.ID)
})
```

### Helper Functions

#### Find Enemies in Range

```go
enemies := engine.FindEnemiesInRange(world, attacker, maxRange float64) []*Entity
```

Returns all enemy entities within `maxRange` of the attacker.

**Requirements:**
- Attacker must have `PositionComponent` and `TeamComponent`
- Enemies must have different `TeamID` (excludes neutral team 0)
- Enemies must have `HealthComponent` and be alive
- Enemies must have `PositionComponent`

**Example:**
```go
enemies := engine.FindEnemiesInRange(world, player, 100)
fmt.Printf("Found %d enemies in range\n", len(enemies))
```

#### Find Nearest Enemy

```go
nearest := engine.FindNearestEnemy(world, attacker, maxRange float64) *Entity
```

Returns the closest enemy entity within `maxRange`, or `nil` if none found.

**Example:**
```go
nearest := engine.FindNearestEnemy(world, player, 150)
if nearest != nil {
    combatSystem.Attack(player, nearest)
}
```

## Damage Calculation

The combat system calculates damage using the following formula:

```
baseDamage = AttackComponent.Damage

// Add attacker stats
if DamageType == Magical:
    baseDamage += MagicPower
else:
    baseDamage += Attack

// Apply critical hit
if randomRoll() < CritChance:
    baseDamage *= CritDamage

// Apply target defense
if DamageType == Magical:
    finalDamage = baseDamage - MagicDefense
else:
    finalDamage = baseDamage - Defense

// Apply resistance
resistance = GetResistance(DamageType)
finalDamage *= (1.0 - resistance)

// Minimum damage
finalDamage = max(1.0, finalDamage)
```

## Usage Examples

### Basic Combat

```go
// Create attacker
warrior := world.CreateEntity()
warrior.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
warrior.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
warrior.AddComponent(&engine.AttackComponent{
    Damage:     25,
    DamageType: combat.DamagePhysical,
    Range:      50,
    Cooldown:   1.0,
})
warrior.AddComponent(engine.NewStatsComponent())

// Create target
goblin := world.CreateEntity()
goblin.AddComponent(&engine.PositionComponent{X: 30, Y: 0})
goblin.AddComponent(&engine.HealthComponent{Current: 50, Max: 50})
goblin.AddComponent(engine.NewStatsComponent())

world.Update(0)

// Attack
if combatSystem.Attack(warrior, goblin) {
    healthComp, _ := goblin.GetComponent("health")
    health := healthComp.(*engine.HealthComponent)
    fmt.Printf("Goblin health: %.0f\n", health.Current)
}
```

### Magic Combat with Resistances

```go
// Fire mage
mage := world.CreateEntity()
mage.AddComponent(&engine.AttackComponent{
    Damage:     30,
    DamageType: combat.DamageFire,
    Range:      100,
    Cooldown:   2.0,
})
mageStats := engine.NewStatsComponent()
mageStats.MagicPower = 20
mage.AddComponent(mageStats)

// Fire-resistant enemy
elemental := world.CreateEntity()
elemental.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
elementalStats := engine.NewStatsComponent()
elementalStats.Resistances[combat.DamageFire] = 0.75  // 75% fire resistance
elemental.AddComponent(elementalStats)

world.Update(0)

// Attack (damage will be reduced by resistance)
combatSystem.Attack(mage, elemental)
```

### Status Effects

```go
// Apply poison
combatSystem.ApplyStatusEffect(enemy, "poison", 5.0, 10.0, 1.0)

// Simulate over time
for i := 0; i < 6; i++ {
    world.Update(1.0)  // Update 1 second
    healthComp, _ := enemy.GetComponent("health")
    health := healthComp.(*engine.HealthComponent)
    fmt.Printf("HP: %.0f\n", health.Current)
}
```

### Team-Based Combat

```go
// Create player team
player1 := world.CreateEntity()
player1.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
player1.AddComponent(&engine.TeamComponent{TeamID: 1})
player1.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})

player2 := world.CreateEntity()
player2.AddComponent(&engine.PositionComponent{X: 20, Y: 0})
player2.AddComponent(&engine.TeamComponent{TeamID: 1})
player2.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})

// Create enemy team
enemy1 := world.CreateEntity()
enemy1.AddComponent(&engine.PositionComponent{X: 100, Y: 0})
enemy1.AddComponent(&engine.TeamComponent{TeamID: 2})
enemy1.AddComponent(&engine.HealthComponent{Current: 80, Max: 80})

world.Update(0)

// Find enemies
enemies := engine.FindEnemiesInRange(world, player1, 150)
fmt.Printf("Found %d enemies\n", len(enemies))

// Target nearest
nearest := engine.FindNearestEnemy(world, player1, 150)
if nearest != nil {
    combatSystem.Attack(player1, nearest)
}
```

## Integration with Other Systems

### Collision System

The combat system integrates with the collision system through the `PositionComponent`:

```go
// Check range using distance calculation
distance := engine.GetDistance(attacker, target)
if distance <= attack.Range {
    combatSystem.Attack(attacker, target)
}
```

### Procedural Generation

Combat stats can be generated using the entity generator:

```go
// Generate entity with procgen
entity := entityGenerator.Generate(seed, params)

// Apply generated stats to combat components
combatEntity := world.CreateEntity()
combatEntity.AddComponent(&engine.HealthComponent{
    Current: entity.Stats.MaxHP,
    Max:     entity.Stats.MaxHP,
})

stats := engine.NewStatsComponent()
stats.Attack = entity.Stats.Attack
stats.Defense = entity.Stats.Defense
combatEntity.AddComponent(stats)
```

## Performance Considerations

- **Cooldown Updates**: O(n) with number of entities with attacks
- **Status Effect Updates**: O(n) with number of entities with effects
- **Attack Calculation**: O(1) per attack
- **Find Enemies in Range**: O(n) with number of entities (consider spatial partitioning for large worlds)

## Testing

The combat system includes comprehensive tests with 90.1% coverage:

```bash
go test -tags test -cover ./pkg/engine/...
```

Test categories:
- Health management
- Attack mechanics and cooldowns
- Evasion and critical hits
- Damage calculation with resistances
- Status effect processing
- Team-based enemy detection
- Event callbacks

## Future Enhancements

Potential future additions to the combat system:

- **Damage types**: More elemental types (arcane, holy, shadow)
- **Status effects**: More effect types (stun, slow, silence, etc.)
- **Combo system**: Chain attacks for bonus damage
- **Block/parry**: Defensive mechanics
- **Area-of-effect**: Attacks hitting multiple targets
- **Projectile system**: Physical projectiles with travel time
- **Damage reflection**: Reflect damage back to attacker
- **Lifesteal**: Heal based on damage dealt
- **Armor penetration**: Bypass a percentage of defense

## See Also

- [Movement and Collision System](MOVEMENT_COLLISION.md)
- [Entity Generator](../procgen/entity/README.md)
- [Combat Demo](../../examples/combat_demo.go)
