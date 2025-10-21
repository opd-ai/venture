# Skill Tree Generation System

## Overview

The skill tree generation system provides deterministic, procedural generation of character progression systems for the Venture action-RPG. Each skill tree represents a character archetype (Warrior, Mage, Rogue, etc.) with interconnected skills arranged in tiers.

## Features

- **Multiple Skill Types**: Passive bonuses, Active abilities, Ultimate powers, and Synergy skills
- **Tier-Based Progression**: 7 tiers (0-6) from basic to master/ultimate skills
- **Prerequisite System**: Skills require previous tier skills to unlock
- **Balanced Scaling**: Stats scale with depth and tier for appropriate power levels
- **Genre Support**: Fantasy (Warrior, Mage, Rogue) and Sci-Fi (Soldier, Engineer, Biotic) templates
- **Deterministic Generation**: Same seed always produces identical skill trees
- **Comprehensive Validation**: Ensures generated trees are valid and playable

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/opd-ai/venture/pkg/procgen"
    "github.com/opd-ai/venture/pkg/procgen/skills"
)

func main() {
    // Create generator
    gen := skills.NewSkillTreeGenerator()
    
    // Configure generation parameters
    params := procgen.GenerationParams{
        Depth:      10,
        Difficulty: 0.5,
        GenreID:    "fantasy",
        Custom: map[string]interface{}{
            "count": 3, // Generate 3 trees
        },
    }
    
    // Generate skill trees
    result, err := gen.Generate(12345, params)
    if err != nil {
        log.Fatal(err)
    }
    
    trees := result.([]*skills.SkillTree)
    
    // Validate results
    if err := gen.Validate(trees); err != nil {
        log.Fatal(err)
    }
    
    // Use the skill trees
    for _, tree := range trees {
        fmt.Printf("Tree: %s\n", tree.Name)
        fmt.Printf("Skills: %d\n", len(tree.Nodes))
        fmt.Printf("Max Points: %d\n", tree.MaxPoints)
    }
}
```

### CLI Tool

Test skill tree generation without writing code:

```bash
# Generate fantasy trees
skilltest -genre fantasy -count 3 -depth 5 -seed 12345

# Generate sci-fi trees with detailed output
skilltest -genre scifi -count 3 -depth 10 -verbose

# Save to file
skilltest -genre fantasy -count 5 -output skills.txt
```

## Skill Types

### Passive Skills
- Always-active bonuses
- No activation required
- Can be leveled multiple times (typically 3-5 levels)
- Examples: +10% damage, +5% crit chance, +15% health

### Active Skills
- Player-activated abilities
- Have cooldowns and resource costs
- Can be leveled to improve effectiveness
- Examples: Fireball, Charge Attack, Stealth

### Ultimate Skills
- Powerful, game-changing abilities
- Typically one level only
- Long cooldowns
- Require high player level and prerequisites
- Examples: Meteor, Berserker Rage, Orbital Strike

### Synergy Skills
- Enhance other skills in the tree
- Provide multiplicative bonuses
- Encourage specific playstyles
- Examples: "Melee attacks have 20% chance to cast equipped spell"

## Skill Tree Structure

Each skill tree contains:

```
Tier 0 (Basic):      3 root skills
Tier 1-2 (Inter):    4-5 skills per tier
Tier 3-4 (Advanced): 3-4 skills per tier  
Tier 5 (Master):     2 skills
Tier 6 (Ultimate):   1 ultimate skill
```

Total: ~15-25 skills per tree

### Prerequisites

- Each skill (except tier 0) requires 1-2 skills from the previous tier
- Creates meaningful progression paths
- Players must make strategic choices
- Ultimate skills require the entire tree path

## Fantasy Genre Trees

### Warrior
**Focus**: Melee combat and physical prowess
- **Passives**: Weapon mastery, armor proficiency, health bonuses
- **Actives**: Cleave, Charge, Smash attacks
- **Ultimate**: Berserker Rage (massive damage + survivability)

### Mage
**Focus**: Arcane magic and elemental power
- **Passives**: Spell damage, mana regeneration, cast speed
- **Actives**: Elemental spells, arcane missiles
- **Ultimate**: Cataclysm (devastating AoE magic)

### Rogue
**Focus**: Stealth, speed, and precision
- **Passives**: Critical chance, dodge, movement speed
- **Actives**: Backstab, Shadow Strike, evasion abilities
- **Ultimate**: Blade Dance (rapid attacks with high crit)

## Sci-Fi Genre Trees

### Soldier
**Focus**: Advanced weaponry and explosives
- **Passives**: Weapon proficiency, armor, accuracy
- **Actives**: Grenades, missile strikes, tactical abilities
- **Ultimate**: Orbital Strike (massive AoE destruction)

### Engineer
**Focus**: Technology and gadgets
- **Passives**: Tech bonuses, cooldown reduction, efficiency
- **Actives**: Turrets, drones, deployables
- **Ultimate**: Mech Suit (temporary power armor)

### Biotic
**Focus**: Psionic powers and mental abilities
- **Passives**: Psi power, shields, mental fortitude
- **Actives**: Mind blast, telekinesis, crowd control
- **Ultimate**: Singularity (creates devastating gravity field)

## Parameters

### Depth
- Controls power scaling (1-100)
- Higher depth = more powerful skills
- Affects: player level requirements, stat values, max skill points
- Recommended: 1-5 (early game), 10-20 (mid game), 30+ (end game)

### Difficulty
- Controls challenge level (0.0-1.0)
- Affects: skill point costs, stat requirements
- Recommended: 0.5 for balanced gameplay

### Count
- Number of skill trees to generate (1-10)
- Fantasy generates: Warrior, Mage, Rogue (cycles if count > 3)
- Sci-Fi generates: Soldier, Engineer, Biotic (cycles if count > 3)

## Skill Progression Example

```
Level 1 Player:
  ├─ Can learn Tier 0 skills (3 available)
  └─ Has 3 skill points

Level 10 Player:
  ├─ Can learn Tier 0-2 skills
  ├─ Must have prerequisites from previous tier
  └─ Has 15 skill points

Level 30 Player:
  ├─ Can learn all tiers including ultimate
  ├─ Must have complete path to ultimate
  └─ Has 45 skill points
```

## Skill Requirements

Each skill has requirements that must be met:

```go
type Requirements struct {
    PlayerLevel      int                // Minimum character level
    SkillPoints      int                // Points needed to learn
    PrerequisiteIDs  []string          // Previous skills required
    AttributeMinimums map[string]int   // Stat requirements (optional)
}
```

Check if player can unlock a skill:

```go
skill := tree.GetSkillByID("skill_12345_5")
canUnlock := skill.IsUnlocked(
    playerLevel,
    availableSkillPoints,
    learnedSkills,
    playerAttributes,
)
```

## Effects System

Skills provide effects that modify character stats:

```go
type Effect struct {
    Type        string  // "damage", "defense", "speed", etc.
    Value       float64 // Numeric value
    IsPercent   bool    // Whether value is percentage
    Description string  // Human-readable description
}
```

### Common Effect Types

- **Combat**: `damage`, `crit_chance`, `crit_damage`, `attack_speed`
- **Defense**: `armor`, `health`, `damage_reduction`, `dodge_chance`
- **Magic**: `spell_damage`, `mana_regen`, `cast_speed`, `max_mana`
- **Utility**: `move_speed`, `cooldown_reduction`, `resource_efficiency`

## Deterministic Generation

The system guarantees identical results for the same seed:

```go
seed := int64(12345)

result1, _ := gen.Generate(seed, params)
result2, _ := gen.Generate(seed, params)

// result1 and result2 are identical
trees1 := result1.([]*skills.SkillTree)
trees2 := result2.([]*skills.SkillTree)

// Same tree names, same skill names, same effects
```

This is critical for:
- Multiplayer synchronization
- Reproducible testing
- Balance verification
- Bug reporting

## Validation

The generator validates all generated content:

```go
err := gen.Validate(result)
if err != nil {
    // Something is wrong with generated trees
    log.Fatal(err)
}
```

Validation checks:
- ✓ All skills have valid names and effects
- ✓ Prerequisites reference existing skills
- ✓ Root nodes exist (no orphan skills)
- ✓ Stat values are within valid ranges
- ✓ Tree structure is complete

## Performance

Generation is extremely fast:

```
Fantasy (3 trees):  ~100-200 µs
Sci-Fi (3 trees):   ~100-200 µs  
Large tree (depth 50): ~300 µs
```

Performance targets:
- ✓ <1ms for typical generation
- ✓ <10ms for extreme cases
- ✓ Zero allocations for hot paths

## Testing

The package includes comprehensive tests:

```bash
# Run all tests
go test ./pkg/procgen/skills/

# Run with coverage
go test -cover ./pkg/procgen/skills/

# Run with race detection
go test -race ./pkg/procgen/skills/
```

**Current Coverage**: 90.6%

## Integration

### With Game Systems

```go
// Generate trees at game start
trees := generateSkillTrees(seed, params)

// Store in player data
player.AvailableTrees = trees

// When player levels up
player.SkillPoints += 1

// When player learns skill
skill := tree.GetSkillByID(skillID)
if skill.IsUnlocked(player.Level, player.SkillPoints, player.LearnedSkills, player.Attributes) {
    skill.Level++
    player.LearnedSkills[skillID] = true
    player.SkillPoints -= skill.Requirements.SkillPoints
    
    // Apply skill effects to player
    for _, effect := range skill.Effects {
        player.ApplyEffect(effect)
    }
}
```

### With Save System

```go
// Serialize tree state
type SavedSkillTree struct {
    TreeID        string
    Seed          int64
    LearnedSkills map[string]int // skill ID -> level
}

// Restore tree from save
func RestoreTree(saved SavedSkillTree) *SkillTree {
    // Regenerate tree from seed
    gen := skills.NewSkillTreeGenerator()
    result, _ := gen.Generate(saved.Seed, params)
    trees := result.([]*skills.SkillTree)
    
    tree := findTreeByID(trees, saved.TreeID)
    
    // Restore learned skills
    for skillID, level := range saved.LearnedSkills {
        skill := tree.GetSkillByID(skillID)
        skill.Level = level
    }
    
    return tree
}
```

## Architecture

The skill tree system follows the same patterns as other generators:

```
skills/
├── doc.go          # Package documentation
├── types.go        # Core data structures
├── generator.go    # Generator implementation
├── templates.go    # Genre templates
├── skills_test.go  # Comprehensive tests
└── README.md       # This file
```

### Design Principles

1. **Deterministic**: Same seed = same output
2. **Validated**: All content is validated before use
3. **Extensible**: Easy to add new genres/archetypes
4. **Testable**: High test coverage, clear contracts
5. **Performant**: Fast generation, minimal allocations

## Future Enhancements

Potential improvements for future versions:

- [ ] Cross-tree synergies (multi-class bonuses)
- [ ] Dynamic tree generation based on playstyle
- [ ] Skill mutation/evolution systems
- [ ] Hybrid genre trees (fantasy-scifi fusion)
- [ ] Procedural skill effects (not just templates)
- [ ] Visual tree layout optimization
- [ ] Tree difficulty ratings
- [ ] Achievement-unlocked skills

## Examples

### Finding Skills by Category

```go
// Get all combat skills
for _, node := range tree.Nodes {
    if node.Skill.Category == skills.CategoryCombat {
        fmt.Printf("Combat skill: %s\n", node.Skill.Name)
    }
}
```

### Calculating Total Tree Power

```go
func CalculateTreePower(tree *skills.SkillTree) float64 {
    power := 0.0
    for _, node := range tree.Nodes {
        if node.Skill.Level > 0 {
            for _, effect := range node.Skill.Effects {
                power += effect.Value * float64(node.Skill.Level)
            }
        }
    }
    return power
}
```

### Building Skill Recommendations

```go
func RecommendNextSkill(tree *skills.SkillTree, player *Player) *skills.Skill {
    var best *skills.Skill
    bestValue := 0.0
    
    for _, node := range tree.Nodes {
        skill := node.Skill
        if skill.Level > 0 {
            continue // Already learned
        }
        
        if !skill.IsUnlocked(player.Level, player.SkillPoints, player.LearnedSkills, player.Attributes) {
            continue // Can't learn yet
        }
        
        // Calculate value based on player's playstyle
        value := calculateSkillValue(skill, player.Playstyle)
        if value > bestValue {
            bestValue = value
            best = skill
        }
    }
    
    return best
}
```

## Support

For issues, questions, or contributions:
- GitHub Issues: https://github.com/opd-ai/venture/issues
- Documentation: See [ARCHITECTURE.md](../../../docs/ARCHITECTURE.md)
- Related Systems: [entity](../entity/), [item](../item/), [magic](../magic/)

## License

See repository [LICENSE](../../../LICENSE) file.
