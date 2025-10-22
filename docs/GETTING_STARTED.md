# Getting Started with Venture

Welcome to Venture, a fully procedural multiplayer action-RPG! This guide will help you get up and running in just a few minutes.

## Quick Start (5 Minutes)

### 1. Installation

**Prerequisites:**
- Go 1.24.7 or later
- Platform dependencies:
  - **Linux:** `sudo apt-get install libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config`
  - **macOS:** Xcode command line tools (`xcode-select --install`)
  - **Windows:** No additional dependencies needed

**Clone and Build:**
```bash
# Clone the repository
git clone https://github.com/opd-ai/venture.git
cd venture

# Build the client
go build -o venture-client ./cmd/client

# Build the server (optional for multiplayer)
go build -o venture-server ./cmd/server
```

### 2. First Launch

Start the game client:
```bash
./venture-client
```

**Default Controls:**
- **WASD** - Move your character
- **Mouse** - Look around and target
- **Space** - Attack/Interact
- **E** - Use item/Open inventory
- **Tab** - Toggle character stats
- **Esc** - Pause menu

### 3. Your First Game

When you start, you'll spawn in a procedurally generated dungeon. Here's what to do:

1. **Explore**: Move around with WASD to explore the dungeon
2. **Fight**: Encounter enemies and use Space to attack them
3. **Collect**: Pick up items dropped by defeated enemies
4. **Level Up**: Gain experience and unlock new abilities
5. **Progress**: Find the stairs to descend to deeper levels

## Game Overview

### What Makes Venture Unique?

- **100% Procedurally Generated**: Everything—maps, enemies, items, music, graphics—is generated at runtime
- **No Asset Files**: Single binary with zero external dependencies
- **Multiplayer Ready**: Co-op gameplay with high-latency support (even over Tor!)
- **Multiple Genres**: Play in fantasy, sci-fi, horror, cyberpunk, or post-apocalyptic settings

### Core Gameplay Loop

```
Explore → Fight → Collect → Level Up → Progress → Repeat
```

1. **Explore** procedurally generated dungeons with unique layouts every time
2. **Fight** dynamically generated enemies with scaled difficulty
3. **Collect** procedurally created items with randomized stats
4. **Level Up** your character with generated skill trees
5. **Progress** to deeper levels with increasing challenges

## Basic Concepts

### Character Stats

Your character has several key stats:

- **Health (HP)**: Your life points - don't let it reach zero!
- **Attack**: Determines your damage output
- **Defense**: Reduces incoming damage
- **Magic**: Affects spell power and mana
- **Speed**: How fast you move

These stats grow as you level up and equip better items.

### Combat System

Combat is real-time and action-oriented:

1. **Melee Combat**: Get close and press Space to attack
2. **Ranged Combat**: Aim with mouse and attack from distance
3. **Magic Spells**: Use equipped spells for special effects
4. **Dodging**: Move away from enemy attacks to avoid damage
5. **Critical Hits**: Random chance for bonus damage

### Inventory & Equipment

- **Inventory Capacity**: Limited slots, manage wisely
- **Equipment Slots**: Head, Body, Weapon, Accessory
- **Item Rarity**: Common → Uncommon → Rare → Epic → Legendary
- **Auto-sort**: Press 'I' to organize inventory

### Progression System

- **Experience Points (XP)**: Earned by defeating enemies
- **Leveling**: Reach XP threshold to level up
- **Skill Points**: Unlock abilities in your skill tree
- **Stat Growth**: Automatic stat increases per level

## Game Modes

### Single Player

Start the client directly to play solo:
```bash
./venture-client -seed 12345 -genre fantasy
```

**Options:**
- `-seed`: Set world seed (default: random)
- `-genre`: Choose theme (fantasy, scifi, horror, cyberpunk, postapoc)
- `-width`/`-height`: Set screen resolution

### Multiplayer Co-op

Start a dedicated server:
```bash
# Start server
./venture-server -port 8080 -max-players 4

# Connect clients
./venture-client -server localhost:8080
```

**Multiplayer Features:**
- Up to 4 players cooperative
- Shared world with synchronized state
- High-latency support (200-5000ms)
- Client-side prediction for responsiveness

## Customization

### World Generation

Control world generation with seed and genre:

```bash
# Fantasy themed world with specific seed
./venture-client -seed 42 -genre fantasy

# Sci-fi horror hybrid
./venture-client -genre scifi -difficulty 0.7
```

### Display Settings

Customize the window:
```bash
# Fullscreen HD
./venture-client -width 1920 -height 1080

# Windowed mode
./venture-client -width 1280 -height 720
```

### Difficulty Settings

Adjust challenge level (0.0 = easiest, 1.0 = hardest):
```bash
./venture-client -difficulty 0.3  # Easy
./venture-client -difficulty 0.5  # Normal (default)
./venture-client -difficulty 0.8  # Hard
```

## Tips for New Players

### Combat Tips
1. **Don't rush into groups**: Pull enemies one at a time
2. **Use the terrain**: Doorways limit enemy numbers
3. **Watch your health**: Retreat when health is low
4. **Learn enemy patterns**: Each enemy type has predictable behavior
5. **Save healing items**: Use them strategically

### Exploration Tips
1. **Clear each room**: Don't miss loot or easy XP
2. **Look for secrets**: Some rooms hide treasure
3. **Manage inventory**: Don't hoard common items
4. **Check equipment**: Always equip better gear
5. **Save progress**: Use save points regularly (F5 quick save)

### Progression Tips
1. **Focus your build**: Don't spread skill points thin
2. **Complete quests**: Extra XP and rewards
3. **Upgrade regularly**: Equipment scales with depth
4. **Learn spell combos**: Some spells synergize
5. **Experiment with genres**: Each has unique content

## Troubleshooting

### Game won't start
- **Linux**: Install X11 libraries: `sudo apt-get install libc6-dev libgl1-mesa-dev ...`
- **macOS**: Install Xcode tools: `xcode-select --install`
- **Windows**: Ensure Go is properly installed

### Low framerate
- Lower resolution: `./venture-client -width 800 -height 600`
- Reduce entity count in settings
- Check system requirements

### Cannot connect to server
- Verify server is running: `./venture-server -verbose`
- Check firewall settings
- Confirm correct port: default is 8080

### Game crashes
- Update to latest version
- Check console output for errors
- Report issues on GitHub with error logs

## Next Steps

Now that you're familiar with the basics:

1. **Read the [User Manual](USER_MANUAL.md)** for detailed gameplay mechanics
2. **Check [API Reference](API_REFERENCE.md)** if you want to modify or extend the game
3. **Join the community** to share experiences and get help
4. **Try different genres** to experience variety in content generation

## Command Reference

### Client Options
```
-width int         Screen width (default: 800)
-height int        Screen height (default: 600)
-seed int          World seed (default: random)
-genre string      Genre ID (default: "fantasy")
-difficulty float  Difficulty 0.0-1.0 (default: 0.5)
-server string     Server address for multiplayer
-verbose          Enable debug logging
```

### Server Options
```
-port string       Server port (default: "8080")
-max-players int   Maximum players (default: 4)
-tick-rate int     Updates per second (default: 20)
-seed int          World seed (default: random)
-genre string      Genre ID (default: "fantasy")
-verbose          Enable debug logging
```

## Resources

- **Project Repository**: https://github.com/opd-ai/venture
- **Documentation**: [docs/](.)
- **Bug Reports**: GitHub Issues
- **Contributing**: [CONTRIBUTING.md](CONTRIBUTING.md)

---

**Ready to play?** Launch the game and start your adventure!

```bash
./venture-client
```

Have fun exploring the infinite procedurally generated worlds of Venture!
