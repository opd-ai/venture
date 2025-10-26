# Getting Started with Venture

Welcome to Venture, a fully procedural multiplayer action-RPG! This guide will help you get up and running in just a few minutes.

## Quick Start (5 Minutes)

### 1. Installation

**Prerequisites:**
- Go 1.24.5 or later
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
- **Space** - Attack
- **E** - Use item
- **I** - Open inventory
- **J** - Open quest log
- **K** - Open skill tree
- **M** - Open map
- **Esc** - Pause menu/Help
- **F5** - Quick save
- **F9** - Quick load
- **1-5** - Cast spells

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

**Client:** `-width`, `-height`, `-seed`, `-genre`, `-verbose`, `-multiplayer`, `-server`
**Server:** `-port`, `-max-players`, `-tick-rate`, `-seed`, `-genre`, `-verbose`

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
