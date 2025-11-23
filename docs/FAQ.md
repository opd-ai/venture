# Frequently Asked Questions (FAQ)

**Venture - Fully Procedural Multiplayer Action-RPG**  
**Version:** 10.0  
**Last Updated:** December 2025

This document answers the 50+ most common questions from players, organized by category.

---

## Table of Contents

1. [General](#general)
2. [Installation & Setup](#installation--setup)
3. [Gameplay](#gameplay)
4. [Multiplayer & Federation](#multiplayer--federation)
5. [Performance & Technical](#performance--technical)
6. [Content & Features](#content--features)
7. [Modding & Customization](#modding--customization)
8. [Troubleshooting](#troubleshooting)

---

## General

### What is Venture?

Venture is a fully procedural multiplayer action-RPG where **100% of content** is generated at runtime—no external art assets, sound files, or text documents. Everything from sprites to dungeons to music to NPC dialog is created algorithmically using deterministic seed-based generation.

### Is Venture free?

Yes, Venture is **open-source** (MIT License) and free to play. You can download binaries from [GitHub Releases](https://github.com/opd-ai/venture/releases) or build from source.

### What platforms are supported?

- **Desktop:** Windows (x64), macOS (Intel + Apple Silicon), Linux (x64, ARM64)
- **Web:** Any modern browser via WebAssembly (Chrome, Firefox, Safari, Edge)
- **Mobile:** Android (ARMv7, ARM64), iOS (ARM64)

### What version should I play?

**Latest stable:** v10.0 (December 2025)  
**Recommended for new players:** Start with v10.0 for full feature set and polish.

### Can I play offline / single-player?

Yes! Start the game without `-multiplayer` flag for local single-player mode. All features except cross-server federation and multiplayer chat work offline.

---

## Installation & Setup

### How do I install Venture?

**Desktop:**
1. Download the appropriate build for your OS from [Releases](https://github.com/opd-ai/venture/releases)
2. Extract the archive (`venture-v10.0-linux-x64.tar.gz`, etc.)
3. Run `./venture-client` (Linux/macOS) or `venture-client.exe` (Windows)

**Web:**
Visit [https://opd-ai.github.io/venture](https://opd-ai.github.io/venture) - no installation required!

**Mobile:**
- **Android:** Download `.apk` from Releases, enable "Install Unknown Apps", install
- **iOS:** Requires TestFlight or building from source (App Store submission in progress)

See [Getting Started](GETTING_STARTED.md) for detailed instructions.

### What are the system requirements?

**Minimum:**
- **CPU:** Dual-core 2.0 GHz
- **RAM:** 2 GB
- **GPU:** OpenGL 2.1 or WebGL 2.0 compatible
- **Storage:** 50 MB
- **OS:** Windows 10, macOS 10.13, Ubuntu 18.04, or equivalent

**Recommended:**
- **CPU:** Quad-core 3.0 GHz
- **RAM:** 4 GB
- **GPU:** Dedicated graphics with OpenGL 3.3+
- **Storage:** 200 MB (for saves and caching)
- **OS:** Latest stable release

**Note:** Mobile requirements vary by device. Most devices from 2020+ should run smoothly.

### How do I update to the latest version?

**Manual:**
1. Download the new version from Releases
2. Extract to a new folder (don't overwrite old version)
3. Copy your `saves/` folder from old version to new version
4. Run the new client

**Automatic (planned for v10.1):** In-game update checker will notify you and auto-download.

### Where are save files stored?

- **Windows:** `%APPDATA%\Venture\saves\`
- **macOS:** `~/Library/Application Support/Venture/saves/`
- **Linux:** `~/.local/share/venture/saves/`
- **Web:** Browser LocalStorage (limited to 10MB)
- **Mobile:** App-specific storage (auto-backed up to iCloud/Google Drive if enabled)

---

## Gameplay

### How do I start a new game?

1. Launch the client
2. Main Menu → "New Game"
3. Choose character class (6 options: Warrior, Rogue, Mage, Ranger, Cleric, Necromancer)
4. Select difficulty (Easy, Normal, Hard, Nightmare)
5. Enter world seed (optional, leave blank for random)
6. Click "Start Adventure"

### What is a world seed?

A **world seed** is a number (e.g., `12345`) that determines the layout of the entire world—terrain, dungeons, NPCs, items, etc. **Same seed = identical world.** Share seeds with friends to explore the same world together!

**Popular seeds:**
- `42`: Balanced dungeon layout, good for beginners
- `2025`: High-difficulty challenges, legendary loot
- `1337`: Easter eggs and secret areas

### How do I save my progress?

**Manual Save:** Press `F5` or Main Menu → Save Game  
**Auto-Save:** Game auto-saves every 10 minutes and on exit  
**Quick Save Slots:** F5 (Quick Save), F9 (Quick Load)

**Important:** Single-player saves are **local only**. Multiplayer progress is saved on the server.

### How do I level up and get stronger?

1. **Gain Experience (XP):** Kill enemies, complete quests, discover locations, read books
2. **Level Up:** When XP bar fills, you gain 1 level (allocate stat points in Character Sheet)
3. **Equipment:** Find/craft better weapons, armor, and consumables
4. **Skills:** Unlock abilities via skill tree (K menu)
5. **Magic:** Learn spells from spell books or teachers
6. **Classes:** Specialize at level 10, dual-class at level 20

### What are the different damage types?

- **Physical:** Melee weapons, arrows (resisted by armor)
- **Fire:** Burns enemies, effective vs. undead/nature
- **Ice:** Slows enemies, effective vs. fire creatures
- **Lightning:** Chain damage, effective vs. mechanical/water
- **Poison:** Damage over time, effective vs. organic enemies
- **Holy:** Extra damage vs. undead/demons
- **Dark:** Life drain, effective vs. living enemies

### How do I unlock new areas?

- **Explore:** Walk to the edge of the map to discover adjacent zones
- **Portals:** Find portal stones to teleport to distant regions
- **Quests:** Some quests unlock new dungeon levels or secret areas
- **Keys:** Find keys to unlock doors, gated areas
- **Progression:** Certain areas require minimum level (e.g., Nightmare Dungeon needs Level 30)

---

## Multiplayer & Federation

### How do I play with friends?

**Option 1: Join Existing Server**
1. Launch client with `-multiplayer -server <IP:PORT>` (e.g., `./venture-client -multiplayer -server 192.168.1.100:8080`)
2. Enter server password if required
3. Your character joins the server's world

**Option 2: Host Your Own Server**
1. Launch server: `./venture-server -port 8080 -max-players 10`
2. Share your IP address with friends (find via `ifconfig` or `ipconfig`)
3. Friends connect using Option 1

**Option 3: LAN Party Mode**
1. Launch client with `-host-and-play` (auto-starts server + connects client)
2. Friends on same network connect to your IP

See [Multiplayer Guide](MULTIPLAYER_GUIDE.md) for detailed setup.

### What is server federation?

**Federation** allows multiple servers to connect, creating a **multiverse** of interconnected worlds. Players can travel between servers via portals, trade items cross-server, and join guilds spanning multiple servers.

**Example:** Server A (seed 1000) ↔ Server B (seed 2000) = Players can portal between worlds!

See [Federation Guide](docs/FEDERATION.md) for technical details.

### Can I play on mobile with desktop players?

Yes! **Cross-platform multiplayer** is fully supported. Mobile players connect to the same servers as desktop/web players with identical gameplay.

**Note:** Touch controls may be less precise than mouse/keyboard. Use auto-aim assist (Settings → Accessibility) for better experience.

### How does lag compensation work?

Venture supports **200-5000ms latency** (including Tor onion services). Key techniques:
- **Client-Side Prediction:** Your actions feel instant, server corrects if needed
- **Entity Interpolation:** Smooth movement of other players despite network jitter
- **Lag Compensation:** Server rewinds time for hit detection (fair combat at high ping)

**Performance Tip:** Latency shown in top-right corner. Green (<100ms) = excellent, Yellow (100-500ms) = good, Red (>500ms) = playable but laggy.

### Can I transfer my character between servers?

**Yes, via portals!** Cross-server portals use a **two-phase commit** protocol:
1. Approach portal, select destination server
2. Client sends transfer request to origin server
3. Origin server validates, freezes character state
4. Destination server receives and validates state
5. Transfer completes (success) or rolls back (failure, character stays on origin)

**Limitations:**
- Destination server must be **trusted** (reputation score >0.5)
- Transfer timeout: 60 seconds (network issues cause rollback)
- Inventory limit: <100KB (drop excess items before transfer)

---

## Performance & Technical

### Why is my FPS low?

**Common causes:**
1. **High Resolution:** Try 1280×720 instead of 1920×1080 (Settings → Video → Resolution)
2. **Too Many Entities:** Reduce particle effects (Settings → Graphics → Particle Quality: Low)
3. **Post-Processing:** Disable bloom, AO, motion blur (Settings → Graphics → Advanced)
4. **Background Apps:** Close memory-intensive programs
5. **Outdated Drivers:** Update GPU drivers

**Performance Mode:** Settings → Graphics → Quality: Low (disables most effects, 2x FPS boost)

See [Performance Guide](PERFORMANCE.md) for profiling and optimization.

### How much memory does Venture use?

**Typical usage:**
- **Client:** 73-200 MB (depends on cache size)
- **Server:** 100-500 MB (depends on player count and world size)

**High memory (>500MB)?** Check Settings → Graphics → Cache Size. Reduce from 150MB to 50MB if needed.

### Does Venture support mods?

**Yes!** v8.0 introduced a **server mod framework**. Mods can customize:
- Game rules (damage multipliers, XP rates, etc.)
- World generation (tweak terrain, entity spawn rates)
- Custom events (seasonal events, special challenges)

**Constraint:** Mods **cannot** add external assets (maintains zero-asset architecture).

See [Modding Guide](MODDING_GUIDE.md) for creating mods.

### Is Venture deterministic?

**Mostly yes:**
- **Deterministic:** Terrain, entity stats, item generation, quest content (same seed = identical)
- **Non-Deterministic:** NPC dialog (Markov chains), player actions, multiplayer interactions

This enables **reproducible worlds** for testing and sharing while allowing dynamic content variety.

---

## Content & Features

### How many dungeons are there?

**Infinite!** Dungeons are procedurally generated. Every world seed creates a unique dungeon network. Typical playthrough visits 50-100 dungeons before reaching endgame.

### What's the max level?

**Level Cap:** 100 (soft cap)  
**Prestige Levels:** Unlimited (reset to level 1, keep equipment, gain prestige bonuses)

### How many character classes are there?

**6 Base Classes:**
- **Warrior:** High HP, melee focus, rage mechanics
- **Rogue:** Speed, critical hits, stealth, dual-wield
- **Mage:** Elemental spells, mana efficiency, glass cannon
- **Ranger:** Archery, pet bonding, traps, survival
- **Cleric:** Healing, buffs, undead resistance, holy magic
- **Necromancer:** Summon undead, life drain, curses

**Specializations:** Each class has 2 specializations (unlocked at level 10)  
**Dual-Classing:** Unlock a second class at level 20 (12 possible combinations)

### What are companions/pets?

**Companions** are AI-controlled allies that follow you. Types:
- **Pets:** Cats, dogs, birds (loyalty-based bonding)
- **Summons:** Elementals, undead (temporary, magic-based)
- **Hirelings:** NPC mercenaries (paid with gold)

**Features:**
- Level up with you (gain stats, learn skills)
- Customizable behavior (aggressive, defensive, passive)
- Inventory (carry loot, fetch items)
- Permadeath mode (hardcore option)

### What vehicles are available?

**5 Vehicle Types:**
- **Mounts:** Horses, dragons (fantasy), hovercrafts (sci-fi)
- **Carts:** Cargo transport, slow but high capacity
- **Boats:** Water travel, fishing
- **Gliders:** Aerial navigation, limited fuel
- **Mechs:** Combat vehicles, heavy armor

**Mechanics:** Fuel consumption, durability, upgrades (speed, armor, capacity), genre-specific (fantasy vs. sci-fi designs)

### Can I build a house?

**Yes!** v8.0 added **player housing**:
- **Claim Territory:** Find unclaimed land, build a house
- **Procedural Buildings:** Generate floor plans, place furniture
- **Customization:** Color schemes, decorations, functional furniture (crafting stations, storage)
- **Persistence:** Houses saved across sessions, visible to other players on the server
- **Guild Halls:** Large multi-floor buildings for guilds (requires 5+ members)

### Are there quests?

**Yes, procedurally generated!** Quest types:
- **Kill Quests:** Defeat X enemies of type Y
- **Fetch Quests:** Find item/NPC and return
- **Escort Quests:** Protect NPC to destination
- **Exploration:** Discover hidden locations
- **Moral Choices:** Branching dialog with alignment consequences
- **Legendary:** Multi-stage epic quests with unique rewards

**Quest Chains:** Some quests unlock follow-up quests (up to 10 stages)

### What is the endgame content?

- **Prestige Levels:** Reset to level 1, gain prestige bonuses, replay with harder enemies
- **Legendary Quests:** Multi-hour epic quests with legendary item rewards
- **Raids:** 10-player dungeon runs with boss mechanics
- **Guild Warfare:** Territory control, siege battles
- **Cross-Server Events:** Server vs. server competitions
- **Min-Maxing:** Perfect builds, speedrun challenges, achievement hunting

---

## Modding & Customization

### How do I install mods?

1. Download mod files (`.venturemod` extension)
2. Place in `mods/` folder (create if doesn't exist)
3. Launch server with `-enable-mods` flag
4. Server auto-loads mods on startup

**Note:** Clients don't need mods installed—mods are server-side only.

### Can I create custom content?

**Yes, within constraints:**
- **Allowed:** Custom game rules, stat multipliers, world generation tweaks, custom events
- **Not Allowed:** External assets (images, sounds, models), executable code injection

Mods are **configuration files (JSON/YAML)** + **scripting (Lua-based, sandboxed)**.

Example: Increase enemy HP by 2x, reduce XP by 1.5x for hardcore mode.

### Where can I share my mods?

- **Community Hub:** [GitHub Discussions](https://github.com/opd-ai/venture/discussions)
- **Mod Repository:** [Venture Mod Hub](https://github.com/opd-ai/venture-mods) (curated collection)
- **Discord:** Share in #modding channel

See [Modding Guide](MODDING_GUIDE.md) for mod creation tutorials.

---

## Troubleshooting

### Game won't start / crashes on launch

**Windows:**
1. Install [Visual C++ Redistributables](https://aka.ms/vs/17/release/vc_redist.x64.exe)
2. Update GPU drivers (NVIDIA/AMD/Intel)
3. Run as Administrator

**macOS:**
1. System Preferences → Security & Privacy → Allow `venture-client`
2. Update to latest macOS version

**Linux:**
1. Install missing libraries: `sudo apt install libc6-dev libgl1-mesa-dev libasound2-dev`
2. Grant execute permission: `chmod +x venture-client`

### Multiplayer connection failed

1. **Check Server Status:** Ensure server is running (`./venture-server -port 8080`)
2. **Firewall:** Allow port 8080 (TCP/UDP) in firewall settings
3. **Correct IP:** Verify IP address (`ifconfig` / `ipconfig`)
4. **Port Forwarding:** For internet play, forward port 8080 in router settings
5. **Timeout:** Increase client timeout: `./venture-client -multiplayer -server <IP:PORT> -timeout 30`

### Screen flickering / visual glitches

1. **Update GPU Drivers:** Essential for OpenGL compatibility
2. **Disable Post-Processing:** Settings → Graphics → Bloom/AO/Motion Blur: Off
3. **Change Resolution:** Try 1280×720 or 1920×1080 (avoid non-standard resolutions)
4. **Fullscreen Mode:** Toggle F11 (some GPUs handle fullscreen differently)

### Audio crackling / no sound

1. **Audio Drivers:** Update OS audio drivers
2. **Sample Rate:** Settings → Audio → Sample Rate: 44100 Hz (standard)
3. **Buffer Size:** Increase buffer size (Settings → Audio → Buffer Size: 2048)
4. **WebAssembly:** Browser audio has 50-100ms latency (expected, no fix)

### Save file corrupted

1. **Check Backups:** `saves/backups/` folder contains last 10 auto-saves
2. **Restore Backup:** Copy `autosave_YYYYMMDD_HHMMSS.save` to `saves/current.save`
3. **Manual Recovery:** Contact support with `saves/current.save` file (we can attempt repair)

**Prevention:** Enable cloud sync (Settings → Saves → Cloud Backup: Enable) for automatic offsite backups.

### Still having issues?

1. **Check Logs:** `logs/venture-client.log` or `logs/venture-server.log`
2. **Search Issues:** [GitHub Issues](https://github.com/opd-ai/venture/issues)
3. **Ask Community:** Discord #help channel
4. **Report Bug:** [New Issue](https://github.com/opd-ai/venture/issues/new) with log file attached

---

## Additional Resources

- **[Getting Started](GETTING_STARTED.md):** 5-minute quickstart guide
- **[User Manual](USER_MANUAL.md):** Comprehensive feature guide
- **[Controls](CONTROLS.md):** Keyboard/mouse/gamepad mappings
- **[Troubleshooting](TROUBLESHOOTING.md):** Detailed problem-solving guide
- **[Gameplay Guide](GAMEPLAY_GUIDE.md):** Strategies, tips, progression advice
- **[Multiplayer Guide](MULTIPLAYER_GUIDE.md):** Hosting servers, joining, federation
- **[Modding Guide](MODDING_GUIDE.md):** Creating mods, API reference
- **[Discord](https://discord.gg/venture):** Community support and discussion
- **[GitHub](https://github.com/opd-ai/venture):** Source code, issues, contributions

---

**Version History:**

- **v10.0 (December 2025):** Added housing, guild warfare, Phase 66 production release
- **v8.0 (November 2025):** Added housing, guilds, modding framework
- **v7.0 (November 2025):** Enhanced sprites, display scaling, 1920×1080 default
- **v6.0 (November 2025):** Server federation, persistent worlds, cross-server travel
- **v5.0 (November 2025):** Social features, chat, image sharing, item trading

---

**Can't find your answer?** Join our [Discord](https://discord.gg/venture) or open a [GitHub Discussion](https://github.com/opd-ai/venture/discussions).

### How do I backup my save files?

**Manual Backup:**
1. Navigate to save directory (see "Where are save files stored?" above)
2. Copy entire `saves/` folder to backup location (external drive, cloud storage)
3. Zip for compression: `zip -r venture-saves-backup.zip saves/`

**Automatic Backup:**
- Enable cloud sync: Settings → Saves → Cloud Backup: Enable
- Auto-backup creates `saves/backups/autosave_YYYYMMDD_HHMMSS.save` every 10 minutes

### Can I play with keyboard only (no mouse)?

Yes! Enable keyboard-only mode: Settings → Controls → Keyboard-Only Mode: Enable

All actions become keyboard-accessible:
- Arrow keys: Menu navigation
- Enter: Confirm/Select
- Tab: Cycle UI elements
- Number keys: Inventory slots
- Escape: Cancel/Back

### What happens when I die?

**Normal Mode:**
- Respawn at last save point (town, checkpoint)
- Lose 10% XP (not levels)
- Equipped items kept, some inventory items drop

**Hardcore Mode:**
- Permadeath (character deleted)
- Can retrieve items from corpse with new character

**Multiplayer:**
- Team Revival: Party members can revive you within 60 seconds
- No XP loss in multiplayer (server configurable)

### How do quests scale with my level?

Quests dynamically scale to your level ±5:
- Quest Level = Player Level + Random(-5, +5)
- Rewards scale with quest difficulty
- Can re-roll quest difficulty at quest giver (costs gold)

**Example:** Level 20 player receives Level 15-25 quests.

### Can I rename my character?

Yes! Visit any town's Name Registry NPC:
- Cost: 1000 gold
- Limit: Once per 7 days (real time)
- Restrictions: 3-20 characters, alphanumeric + spaces

### What are prestige levels?

**Prestige System** (unlocked at Level 100):
- Reset to Level 1, keep equipment
- Gain +10% stats per prestige level
- Unlock prestige-exclusive content
- Maximum: 10 prestige levels

### How does loot sharing work in multiplayer?

**Loot Modes** (configurable by server):
1. **Free-for-All:** First player to pick up gets loot
2. **Round Robin:** Loot rotates through party members
3. **Need/Greed:** Players roll for items they need
4. **Master Looter:** Party leader distributes loot

**Trading:** Can trade items to party members (see Item Trading section).

### Can I change my character class?

**Dual-Classing:** Unlock second class at Level 20 (see "How many character classes are there?" above)

**Class Respec:**
- Visit Class Trainer NPC in major towns
- Cost: 5000 gold + 1 Skill Reset Token
- Keeps current level and equipment
- Reallocate all skill points

### How does the honor/dishonor system work?

**Alignment System** (Good ↔ Evil):
- Actions affect alignment: Help NPCs (Good), Steal/Kill (Evil)
- Visible via Character Sheet → Alignment
- Unlocks alignment-exclusive quests, items, NPCs
- Reputation with factions also affected

**Redemption:** Evil characters can redeem via multi-part quests.

### Where can I find the rarest items?

**Legendary Drop Sources:**
1. **Bosses:** 5% legendary drop chance (end of dungeons)
2. **Legendary Quests:** Guaranteed legendary reward
3. **Raids:** 10-player content, 20% legendary chance
4. **Crafting:** Combine 5 epic items → 1 legendary (random)
5. **World Events:** Server-wide events, limited-time legendaries

---

**Still have questions?** Join our [Discord](https://discord.gg/venture) or check the [User Manual](USER_MANUAL.md) for detailed information.
