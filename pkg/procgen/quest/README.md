# Quest Generation System

Procedural quest generation system for the Venture action-RPG. Generates quests with objectives, rewards, and thematic descriptions based on genre and difficulty.

## Features

- **6 Quest Types**: Kill, Collect, Escort, Explore, Talk, Boss
- **6 Difficulty Levels**: Trivial, Easy, Normal, Hard, Elite, Legendary
- **Genre Support**: Fantasy, Sci-Fi, Horror, Cyberpunk, Post-Apocalyptic
- **Deterministic Generation**: Same seed produces identical quests
- **Depth Scaling**: Rewards and difficulty scale with dungeon depth
- **Rich Rewards**: XP, gold, items, and skill points
- **Quest Tracking**: Status tracking (Not Started, Active, Complete, Failed)
- **Progress Monitoring**: Track objective completion

## Quest Types

### Kill Quests
Defeat a specific number of enemies.

**Fantasy Example:**
- Name: "Slay the Undead"
- Objective: Defeat 10 Skeletons
- Reward: 150 XP, 30 gold

**Sci-Fi Example:**
- Name: "Terminate the Rogue Bots"
- Objective: Eliminate 15 Combat Drones
- Reward: 180 XP, 40 gold

### Collect Quests
Gather items from the world.

**Fantasy Example:**
- Name: "Gather Herbs"
- Objective: Collect 8 Moonflowers
- Reward: 120 XP, 25 gold, Potion

**Sci-Fi Example:**
- Name: "Salvage Data Cores"
- Objective: Recover 5 Data Cores
- Reward: 140 XP, 35 gold

### Boss Quests
Defeat a powerful unique enemy.

**Fantasy Example:**
- Name: "Defeat the Dragon Lord"
- Objective: Slay the Dragon Lord
- Reward: 1000 XP, 500 gold, Epic Weapon, 2 Skill Points

**Sci-Fi Example:**
- Name: "Eliminate the Titan Mech"
- Objective: Destroy the Titan Mech
- Reward: 1200 XP, 600 gold, Legendary Armor, 1 Skill Point

### Explore Quests
Discover a new location.

**Fantasy Example:**
- Name: "Explore the Ancient Ruins"
- Objective: Discover Ancient Ruins
- Reward: 100 XP, 50 gold

### Escort Quests
Protect an NPC to a destination.

**Fantasy Example:**
- Name: "Escort the Merchant"
- Objective: Safely escort the merchant
- Reward: 200 XP, 100 gold

### Talk Quests
Interact with an NPC.

**Fantasy Example:**
- Name: "Speak with the Elder"
- Objective: Talk to the Village Elder
- Reward: 50 XP, 20 gold

## Usage

### Basic Generation

```go
import (
    "github.com/opd-ai/venture/pkg/procgen"
    "github.com/opd-ai/venture/pkg/procgen/quest"
)

// Create generator
generator := quest.NewQuestGenerator()

// Set parameters
params := procgen.GenerationParams{
    Difficulty: 0.5,     // 0.0-1.0
    Depth:      5,       // Dungeon level
    GenreID:    "fantasy",
    Custom: map[string]interface{}{
        "count": 10,     // Number of quests
    },
}

// Generate quests
result, err := generator.Generate(12345, params)
if err != nil {
    log.Fatal(err)
}

quests := result.([]*quest.Quest)
```

### Working with Quests

```go
for _, q := range quests {
    fmt.Printf("Quest: %s\n", q.Name)
    fmt.Printf("Type: %s, Difficulty: %s\n", q.Type, q.Difficulty)
    
    // Check objectives
    for _, obj := range q.Objectives {
        fmt.Printf("  - %s: %d/%d (%.0f%%)\n",
            obj.Description, obj.Current, obj.Required, obj.Progress()*100)
    }
    
    // Check completion
    if q.IsComplete() {
        fmt.Println("Quest complete!")
    }
    
    // Get rewards
    fmt.Printf("Rewards: %d XP, %d gold\n", q.Reward.XP, q.Reward.Gold)
}
```

### Quest Management

```go
// Accept a quest
quest.Status = quest.StatusActive

// Update progress
quest.Objectives[0].Current += 5

// Check progress
progress := quest.Progress() // Returns 0.0-1.0

// Complete quest
if quest.IsComplete() {
    quest.Status = quest.StatusComplete
    
    // Turn in for rewards
    player.XP += quest.Reward.XP
    player.Gold += quest.Reward.Gold
    quest.Status = quest.StatusTurnedIn
}
```

## Scaling

### Depth Scaling
- **Depth 1-3**: Trivial/Easy quests, low rewards
- **Depth 4-6**: Normal quests, moderate rewards  
- **Depth 7-9**: Hard quests, good rewards
- **Depth 10+**: Elite/Legendary quests, excellent rewards

### Difficulty Parameter
- **0.0**: Easier objectives, lower rewards
- **0.5**: Balanced objectives and rewards
- **1.0**: Harder objectives, higher rewards

### Reward Formulas

**XP Reward:**
```
base_xp * (1 + depth * 0.15) * (1 + difficulty * 0.3)
```

**Gold Reward:**
```
base_gold * (1 + depth * 0.15) * (1 + difficulty * 0.3)
```

**Item Rewards:**
- Common quests: 30% chance
- Boss quests: 90% chance
- Rarity scales with difficulty

**Skill Point Rewards:**
- Boss quests: 50% chance for 1-2 points
- Elite+ quests: Higher chance

## Quest Components

### Quest Structure
```go
type Quest struct {
    ID            string        // Unique identifier
    Name          string        // Display name
    Type          QuestType     // Kill, Collect, etc.
    Difficulty    Difficulty    // Trivial to Legendary
    Description   string        // Flavor text
    Objectives    []Objective   // What to accomplish
    Reward        Reward        // What you get
    RequiredLevel int           // Minimum level
    Status        QuestStatus   // Current state
    Seed          int64         // Generation seed
    Tags          []string      // Metadata
    GiverNPC      string        // Quest giver
    Location      string        // Where it takes place
}
```

### Objective Structure
```go
type Objective struct {
    Description string  // "Defeat 10 Goblins"
    Target      string  // "Goblin"
    Required    int     // 10
    Current     int     // Player progress
}
```

### Reward Structure
```go
type Reward struct {
    XP          int       // Experience points
    Gold        int       // Currency
    Items       []string  // Item IDs or types
    SkillPoints int       // Skill points
}
```

## Genre Templates

### Fantasy
- **Kill Targets**: Goblins, Skeletons, Orcs, Wolves, Bandits, Zombies, Spiders
- **Collect Items**: Moonflowers, Mana Crystals, Ancient Runes, Dragon Scales, Phoenix Feathers
- **Boss Enemies**: Dragon Lord, Lich King, Dark Sorcerer, Demon Prince, Ancient Wyrm
- **Locations**: Ancient Ruins, Dark Forest, Forgotten Temple, Mountain Pass, Lost City
- **Themes**: Medieval, magic, dungeons, monsters

### Sci-Fi
- **Kill Targets**: Combat Drones, Alien Warriors, Mutants, Space Pirates, Rogue AIs
- **Collect Items**: Data Cores, Power Cells, Tech Modules, Mineral Samples, Alien Artifacts
- **Boss Enemies**: Titan Mech, Alien Queen, AI Overlord, Warlord, Omega Unit
- **Themes**: Technology, space, aliens, robots

## Testing

### Unit Tests
```bash
# Run tests
go test -tags test ./pkg/procgen/quest/

# With coverage
go test -tags test -cover ./pkg/procgen/quest/

# Verbose
go test -tags test -v ./pkg/procgen/quest/
```

Coverage: 96.6%

### CLI Tool
```bash
# Build the tool
go build -o questtest ./cmd/questtest

# Generate fantasy quests
./questtest -genre fantasy -depth 5 -count 10

# Generate sci-fi quests
./questtest -genre scifi -depth 10 -difficulty 0.8

# Custom seed
./questtest -seed 42 -count 5
```

## Performance

**Generation Speed:**
- 10 quests: ~0.1 ms
- 100 quests: ~1 ms
- 1000 quests: ~10 ms

**Memory Usage:**
- Quest struct: ~400 bytes
- 100 quests: ~40 KB
- 1000 quests: ~400 KB

**Benchmarks:**
```
BenchmarkQuestGeneration-8     50000    30 µs/op    10 quests
BenchmarkQuestValidation-8    200000     8 µs/op    10 quests
```

## Integration

### With Entity Generator
```go
// Generate enemies at appropriate level
targetLevel := quest.RequiredLevel
entity := entityGenerator.Generate(seed, procgen.GenerationParams{
    Depth: targetLevel,
})
```

### With Item Generator
```go
// Generate quest reward items
for _, itemID := range quest.Reward.Items {
    item := itemGenerator.Generate(seed, params)
    player.Inventory.Add(item)
}
```

### With Progression System
```go
// Award XP on quest completion
if quest.IsComplete() && quest.Status == quest.StatusComplete {
    progressionSystem.AwardXP(player, quest.Reward.XP)
    player.Gold += quest.Reward.Gold
    player.SkillPoints += quest.Reward.SkillPoints
    quest.Status = quest.StatusTurnedIn
}
```

### With Combat System
```go
// Track kill quest progress
combatSystem.SetDeathCallback(func(victim *Entity) {
    for _, quest := range player.ActiveQuests {
        if quest.Type == quest.TypeKill {
            for i := range quest.Objectives {
                if quest.Objectives[i].Target == victim.Type {
                    quest.Objectives[i].Current++
                }
            }
        }
    }
})
```

## Best Practices

### Quest Design
1. **Balance objectives**: Don't make quests too long or tedious
2. **Scale rewards**: Higher difficulty = better rewards
3. **Vary types**: Mix combat, exploration, and collection
4. **Provide context**: Use description text for immersion

### Implementation
1. **Check prerequisites**: Verify player level before offering quests
2. **Track progress**: Update objectives in real-time
3. **Save state**: Persist quest status between sessions
4. **Limit active quests**: Don't overwhelm players (5-10 active)

### Testing
1. **Verify determinism**: Same seed = same quests
2. **Test all types**: Ensure all quest types generate correctly
3. **Check scaling**: Verify depth/difficulty affect rewards
4. **Validate output**: Always check generated quests are valid

## Examples

### Quest Board
```go
func GenerateQuestBoard(depth, playerLevel int) []*quest.Quest {
    generator := quest.NewQuestGenerator()
    params := procgen.GenerationParams{
        Difficulty: 0.5,
        Depth:      depth,
        GenreID:    "fantasy",
        Custom:     map[string]interface{}{"count": 5},
    }
    
    result, _ := generator.Generate(time.Now().Unix(), params)
    quests := result.([]*quest.Quest)
    
    // Filter by player level
    available := make([]*quest.Quest, 0)
    for _, q := range quests {
        if q.RequiredLevel <= playerLevel {
            available = append(available, q)
        }
    }
    
    return available
}
```

### Daily Quests
```go
func GenerateDailyQuests(date time.Time) []*quest.Quest {
    // Use date as seed for consistent daily quests
    seed := date.Unix() / 86400 // One day in seconds
    
    generator := quest.NewQuestGenerator()
    params := procgen.GenerationParams{
        Difficulty: 0.6,
        Depth:      5,
        GenreID:    "fantasy",
        Custom:     map[string]interface{}{"count": 3},
    }
    
    result, _ := generator.Generate(seed, params)
    return result.([]*quest.Quest)
}
```

### Quest Chains
```go
func GenerateQuestChain(baseSeed int64, length int) []*quest.Quest {
    chain := make([]*quest.Quest, length)
    generator := quest.NewQuestGenerator()
    
    for i := 0; i < length; i++ {
        params := procgen.GenerationParams{
            Difficulty: 0.5 + float64(i)*0.1,
            Depth:      5 + i,
            GenreID:    "fantasy",
            Custom:     map[string]interface{}{"count": 1},
        }
        
        result, _ := generator.Generate(baseSeed+int64(i), params)
        quests := result.([]*quest.Quest)
        chain[i] = quests[0]
    }
    
    return chain
}
```

## Future Enhancements

### Planned Features
- [ ] Multi-objective quests (defeat X AND collect Y)
- [ ] Time-limited quests
- [ ] Repeatable quests
- [ ] Quest prerequisites (quest chains)
- [ ] Procedural quest dialogue
- [ ] Dynamic quest generation (based on player actions)
- [ ] Quest reputation system
- [ ] Faction-specific quests

### Advanced Features
- [ ] Story-driven quest chains
- [ ] World state affecting quests
- [ ] Player choice affecting outcomes
- [ ] Procedural quest NPCs
- [ ] Quest failure consequences
- [ ] Hidden/secret quests

## Architecture

The quest generator follows the established patterns from other procgen systems:

1. **Types**: Enums, structs, and templates in `types.go`
2. **Generator**: Main generation logic in `generator.go`
3. **Tests**: Comprehensive tests in `quest_test.go`
4. **Documentation**: Package docs in `doc.go` and this README

**Dependencies:**
- `pkg/procgen`: Base generator interface and parameters
- Standard library only (no external dependencies)

**Integration Points:**
- Entity generator (for kill quests)
- Item generator (for collect/reward items)
- Progression system (for XP rewards)
- Combat system (for tracking kills)
- World state (for explore quests)

## License

Part of the Venture project. See main LICENSE file.
