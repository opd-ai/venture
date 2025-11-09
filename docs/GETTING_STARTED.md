# Getting Started with Venture

Welcome to Venture, a fully procedural multiplayer action-RPG! This guide will help you get up and running in just a few minutes.

## Quick Start (5 Minutes)

### 1. Installation

**Prerequisites:**
- Go 1.24.5 or later
- Platform dependencies:
  - **Linux:** `sudo apt-get install libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config libx11-dev`
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
- **Space** - Attack
- **E** - Use item / Open chest
- **F** - Interact with NPCs / Browse merchant shop
- **1-5** - Cast spells
- **F5** - Quick save
- **F9** - Quick load

**In-Game Menus** (Press key again OR ESC to close):

| Menu | Key | What It Does |
|------|-----|--------------|
| Inventory | I | Manage items, drag-drop equipment |
| Character Stats | C | View health, stats, equipment, attributes |
| Skill Tree | K | Spend skill points, unlock new abilities |
| Quest Log | J | Track active quests, view objectives |
| World Map | M | Navigate explored areas, fog of war |
| Crafting | R | Brew potions, enchant gear, craft items |
| Shop | F | Buy/sell items (when near merchant) |
| Help | H or F1 | Controls reference and game tips |
| Pause | ESC | Pause game (closes any open menu first) |

**Navigation Tip:** All menus use dual-exit—press the menu's letter key to toggle, or ESC to close any menu.

### 3. Your First Game

When you start, you'll spawn in a procedurally generated dungeon. Here's what to do:

1. **Explore**: Move around with WASD to explore the dungeon
2. **Fight**: Encounter enemies and use Space to attack them
3. **Collect**: Pick up items dropped by defeated enemies
4. **Level Up**: Gain experience and unlock new abilities
5. **Progress**: Find the stairs to descend to deeper levels

## Core Gameplay

Venture is a procedurally generated action-RPG where everything is created at runtime. The basic gameplay loop is:

**Explore → Fight → Collect → Level Up → Progress → Repeat**

You'll explore unique dungeons, fight generated enemies, collect randomized loot, and progress deeper into increasingly challenging levels.

## Key Concepts

- **Real-time combat** with movement, attacks, and abilities
- **Character progression** through XP, levels, and skill points  
- **Inventory management** with equipment slots and item rarity

**For detailed mechanics and advanced gameplay, see [User Manual](USER_MANUAL.md).**

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
- `-enable-lighting`: Enable dynamic lighting system (experimental, enhances atmosphere)
- `-verbose`: Enable detailed logging
- `-profile`: Enable performance profiling

### Multiplayer Co-op

#### Quick Start - Host-and-Play (LAN Party Mode)

Perfect for LAN parties and local co-op! The host player starts both server and client with a single command:

```bash
# Host player: start server + client (one command!)
./venture-client --host-and-play

# Other players on the same network: join the host
./venture-client -multiplayer -server <host-ip>:8080
```

**Host Configuration:**
- `--host-lan`: Allow LAN connections (default: localhost only for security)
- `-port 8080`: Starting port (auto-tries 8081-8089 if occupied)
- `-max-players 4`: Maximum players (default: 4)
- `-tick-rate 20`: Server update rate (default: 20 Hz)

**Finding the Host IP:**
- **Linux:** `ip addr show | grep inet`
- **Windows:** `ipconfig`
- **macOS:** `ifconfig | grep inet`

**Security Note:** By default, `--host-and-play` binds to localhost only (127.0.0.1). To allow LAN connections, explicitly add `--host-lan`:

```bash
# Allow LAN connections (other computers on local network can join)
./venture-client --host-and-play --host-lan
```

**Example LAN Party Setup:**
```bash
# Host (192.168.1.100): start server accessible on LAN
./venture-client --host-and-play --host-lan -max-players 4

# Player 2: connect from another computer
./venture-client -multiplayer -server 192.168.1.100:8080

# Player 3: connect
./venture-client -multiplayer -server 192.168.1.100:8080
```

#### Traditional Setup - Dedicated Server

For persistent servers or remote hosting, use a dedicated server:

```bash
# Start server
./venture-server -port 8080 -max-players 4

# Connect clients
./venture-client -multiplayer -server localhost:8080
```

**Multiplayer Features:**
- Up to 4 players cooperative (configurable)
- Shared world with synchronized state
- High-latency support (200-5000ms, including Tor/onion services)
- Client-side prediction for responsiveness
- Automatic port fallback (tries 8080-8089)

## Customization

```bash
# Set world seed and genre
./venture-client -seed 42 -genre fantasy

# Adjust screen size
./venture-client -width 1280 -height 720

# Enable verbose logging
./venture-client -verbose
```

**For complete customization options and advanced settings, see [User Manual](USER_MANUAL.md).**

## Tips for New Players

- **Combat:** Pull enemies one at a time, use terrain for advantage, watch your health
- **Exploration:** Clear each room, look for secrets, manage inventory wisely
- **Progression:** Focus your skill points, complete quests, upgrade equipment regularly

**For detailed strategies, mechanics explanations, and advanced tips, see [User Manual](USER_MANUAL.md).**

## Visual Enhancements (V3.0)

### Enhanced Graphics Quality

Venture V3.0 features significant visual improvements while maintaining 100% procedural generation:

**What's New in V3.0:**
- **40% more sprite detail** with anatomical accuracy and facial features
- **Professional-grade lighting** with soft shadows and bloom effects
- **Rich weather systems** with rain, snow, fog, and genre-specific variations
- **Smooth tile transitions** with 50+ procedural texture patterns per genre
- **Dynamic UI colors** that adapt to the genre theme
- **Parallax backgrounds** for enhanced depth perception

All V3.0 enhancements are **enabled by default** and optimized to maintain excellent performance (106 FPS with 2000 entities).

### Dynamic Lighting & Weather

```bash
# Lighting is enabled by default in V3.0, with genre-specific presets
./venture-client -genre fantasy   # Warm torchlight
./venture-client -genre scifi     # Cool neon lights
./venture-client -genre horror    # Dim, eerie lighting
./venture-client -genre cyberpunk # Vibrant neon glow
./venture-client -genre postapoc  # Harsh, dusty lighting

# Enable weather effects for additional atmosphere
./venture-client -enable-weather -weather rain -weather-intensity medium

# Combined for maximum immersion
./venture-client -genre horror -enable-weather -weather fog -weather-intensity heavy
```

**Weather Types (V3.0):**
- **Rain:** Water droplets with realistic fluid simulation
- **Snow:** Snowflakes with accumulation effects
- **Fog:** Volumetric fog that obscures distant areas
- **Dust:** Swirling dust particles in the air
- **Ash:** Falling ash (perfect for post-apocalyptic genre)

**Weather Intensity:** Light, Medium, Heavy, Extreme

**Genre-Specific Weather:**
- Fantasy: Natural rain and snow
- Sci-Fi: Neon rain, energy fog
- Horror: Blood rain, toxic fog
- Cyberpunk: Acid rain, pollution smog
- Post-Apocalyptic: Radiation dust, ash fall

### Visual Quality Settings (V3.0)

```bash
# Anti-aliasing quality (V3.0 feature)
./venture-client -aa-quality high    # Best quality (8x8 super-sampling)
./venture-client -aa-quality medium  # Balanced (4x4 super-sampling)
./venture-client -aa-quality low     # Performance (2x2 super-sampling)
./venture-client -aa-quality off     # Fastest, no anti-aliasing

# Lighting quality
./venture-client -lighting-quality high    # Soft shadows + bloom effects
./venture-client -lighting-quality medium  # Soft shadows only
./venture-client -lighting-quality low     # Basic lighting (fastest)

# Particle density for weather effects
./venture-client -particle-density high    # Maximum particle count
./venture-client -particle-density medium  # Balanced (default)
./venture-client -particle-density low     # Better performance

# Performance preset (sets all quality options)
./venture-client -preset quality      # Maximum visual quality
./venture-client -preset balanced     # Balanced quality/performance (default)
./venture-client -preset performance  # Maximum performance
```

**Performance Notes:**
- V3.0 maintains 106 FPS with all features enabled (70% above 60 FPS target)
- Sprite cache hit rate: 95.9% (same as V2.0)
- Lighting overhead: <5% frame time
- All enhancements use deterministic generation (same seed = same visuals)

---

## Troubleshooting

**Game won't start:**
- Linux: Install X11 libraries (see prerequisites)
- macOS: Install Xcode command line tools
- Windows: Verify Go installation

**Performance issues:** Lower resolution, reduce settings
**Connection issues:** Check server status and firewall
**Crashes:** Check console for errors, report on GitHub

**For detailed troubleshooting, see [User Manual](USER_MANUAL.md) and [Development Guide](DEVELOPMENT.md).**

## Next Steps

Now that you're familiar with the basics:

1. **Read the [User Manual](USER_MANUAL.md)** for detailed gameplay mechanics
2. **Check [API Reference](API_REFERENCE.md)** if you want to modify or extend the game
3. **Join the community** to share experiences and get help
4. **Try different genres** to experience variety in content generation

## Command Reference

**Client:** `-width`, `-height`, `-seed`, `-genre`, `-enable-lighting`, `-lighting-quality`, `-enable-weather`, `-weather`, `-weather-intensity`, `-particle-density`, `-aa-quality`, `-preset`, `-verbose`, `-profile`, `-multiplayer`, `-server`, `--host-and-play`, `--host-lan`, `-port`, `-max-players`, `-tick-rate`, `-no-tutorial`
**Server:** `-port`, `-max-players`, `-tick-rate`, `-seed`, `-genre`, `-verbose`, `-aerial-sprites`, `-high-latency`

**V3.0 New Options:**
- `-aa-quality`: Anti-aliasing level (off, low, medium, high)
- `-lighting-quality`: Lighting quality (low, medium, high)
- `-particle-density`: Particle count (low, medium, high)
- `-preset`: Overall quality preset (performance, balanced, quality)
- `-weather`: Weather type (rain, snow, fog, dust, ash)
- `-weather-intensity`: Weather strength (light, medium, heavy, extreme)

**For complete command-line options and configuration details, see [User Manual](USER_MANUAL.md).**

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
