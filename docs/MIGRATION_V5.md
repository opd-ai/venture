# Migration Guide: v4.0 → v5.0

**Version:** 5.0  
**Date:** November 2025  
**Target Audience:** Players upgrading from v4.0 to v5.0

## Overview

Venture v5.0 introduces social systems (chat, NPC dialog, image sharing, trading) while maintaining full backward compatibility with v4.0 save files. This guide explains the migration process, new features, and potential compatibility considerations.

## Save File Compatibility

### Automatic Migration

**Save files from v4.0 are fully compatible with v5.0.**

When loading a v4.0 save in v5.0:
1. Save file version detected automatically
2. New social components added to existing entities
3. Default settings applied for social features
4. Original game state preserved (world, inventory, progression)

**No manual migration required.**

### Migration Details

**What Gets Migrated:**
- World seed and procedural generation state ✅
- Player position, inventory, equipment ✅
- Quest progress, skill trees, character stats ✅
- NPC states, merchant inventories ✅
- Crafting recipes, station progress ✅
- Companion/pet data (v4.0 feature) ✅

**What Gets Added:**
- ChatComponent for all player entities (new)
- TradeComponent with default trust score 0.5 (new)
- NPCDialogComponent for NPCs (new, replaces simple dialog)
- Social system state (empty message history) (new)

**What Remains Unchanged:**
- World generation (same seed produces same world)
- Entity stats, items, magic, skills (deterministic)
- Multiplayer sync (compatible with v4.0 servers if social features disabled)

### Migration Process

1. **Backup**: Automatically created at `saves/<filename>.v4.0.backup`
2. **Detect**: Version read from save file metadata
3. **Transform**: New components added, existing data preserved
4. **Validate**: Integrity checks ensure no data loss
5. **Save**: New save written with v5.0 version tag

**Migration is one-way**: v5.0 saves cannot be loaded in v4.0.

## New Save File Structure

### v4.0 Save Format
```json
{
  "version": "4.0",
  "worldSeed": 12345,
  "entities": [
    {
      "id": 1,
      "components": {
        "position": {...},
        "inventory": {...},
        "equipment": {...}
      }
    }
  ]
}
```

### v5.0 Save Format
```json
{
  "version": "5.0",
  "worldSeed": 12345,
  "entities": [
    {
      "id": 1,
      "components": {
        "position": {...},
        "inventory": {...},
        "equipment": {...},
        "chat": {
          "messages": [],
          "unreadCount": 0,
          "activeChannels": ["global", "local"],
          "lastMessageTime": {},
          "localRadius": 10.0
        },
        "trade": {
          "activeTrade": null,
          "tradeHistory": [],
          "trustScore": 0.5
        }
      }
    }
  ],
  "socialState": {
    "npcDialogHistory": {},
    "muteList": [],
    "blockList": []
  }
}
```

**Size increase**: ~10-20KB per save file (negligible)

## Feature Comparison

### Chat System (NEW)

**v4.0:**
- No player-to-player chat
- Simple NPC dialog (template-based)

**v5.0:**
- 4 chat channels (Global, Local, Party, Whisper)
- E2E encrypted messaging
- Range extension items (Megaphone, Walkie-Talkie)
- Rate limiting and mute system
- Client-side profanity filter (optional)

**Migration Impact:** None. Chat system is opt-in via multiplayer mode.

### NPC Dialog (ENHANCED)

**v4.0:**
- Template-based responses
- Fixed dialog trees
- Deterministic (same input → same output)

**v5.0:**
- Markov chain generation (dynamic responses)
- Genre-specific vocabulary
- Personality traits
- Conversation history context
- **Fallback**: Use `--deterministic-dialog=true` for v4.0 behavior

**Migration Impact:** NPCs will give varied responses instead of fixed templates. Enable deterministic mode if you prefer v4.0 behavior.

### Image Sharing (NEW)

**v4.0:** Not available

**v5.0:**
- Share images with other players
- Thumbnail preview
- Manual acceptance required
- 10-minute expiry

**Migration Impact:** None. Feature disabled in single-player.

### Trading (NEW)

**v4.0:**
- Drop items on ground for other players
- No atomic transfer (potential for scams)

**v5.0:**
- Two-phase commit protocol
- Atomic transfers (no scams)
- Trust mechanics
- Proximity validation
- Rollback on failure

**Migration Impact:** Safer trading with protection against item loss.

## Configuration Changes

### New Flags (v5.0)

**Social Features:**
```bash
--disable-social       # Disable all social features (v4.0 mode)
--disable-chat         # Disable chat only
--disable-trading      # Disable trading only
--profanity-filter     # Enable profanity filter
--auto-download-images # Auto-download images (default: false)
```

**NPC Dialog:**
```bash
--deterministic-dialog # Use v4.0 template-based dialog
--shared-dialog        # Synchronized NPC dialog in multiplayer
--dialog-order=<2-3>   # Markov chain order (default: 2)
```

**Trading:**
```bash
--trade-radius=<tiles> # Custom proximity (default: 5)
```

### Changed Behavior

**Multiplayer:**
- v4.0: Chat not available
- v5.0: Chat enabled by default (disable with `--disable-chat`)

**NPC Interaction:**
- v4.0: Press E → fixed dialog
- v5.0: Press E → dynamic dialog (use `--deterministic-dialog=true` for v4.0 mode)

**Item Transfers:**
- v4.0: Drop items on ground
- v5.0: Use trade UI (safer, atomic)

## Performance Comparison

### v4.0 Baseline
- 106 FPS with 2000 entities
- 73MB memory usage
- <100KB/s network bandwidth

### v5.0 Overhead
- 105 FPS with 2000 entities + chat (1 FPS decrease)
- 75MB memory usage (+2MB for social state)
- <125KB/s network bandwidth (+25KB/s for social features)

**Impact:** Minimal (<2% performance decrease)

## Multiplayer Compatibility

### Server Version Compatibility

**v5.0 Client + v5.0 Server:**
- ✅ All features available
- ✅ Chat, NPC dialog, trading, image sharing

**v5.0 Client + v4.0 Server:**
- ✅ Compatible with `--disable-social` flag
- ❌ Social features unavailable
- ✅ Core gameplay identical

**v4.0 Client + v5.0 Server:**
- ❌ Incompatible (v4.0 client cannot connect to v5.0 server)
- **Solution:** Upgrade client to v5.0

### Migration Strategy for Servers

**Option 1: Immediate Upgrade**
1. Backup server data
2. Upgrade to v5.0
3. Announce to players (client upgrade required)

**Option 2: Gradual Migration**
1. Run v4.0 server in parallel
2. Launch v5.0 beta server (separate)
3. Allow players to choose version
4. Sunset v4.0 after adoption period

**Recommended:** Option 2 for large communities

## Breaking Changes

### API Changes (for Modders)

**DialogComponent → NPCDialogComponent:**
```go
// v4.0
type DialogComponent struct {
    Choices []DialogChoice
}

// v5.0
type NPCDialogComponent struct {
    NPCPersonality *dialog.Personality
    ConversationHistory []string
    ResponseHistory []string
    Generator *dialog.MarkovGenerator
    // ... (see pkg/engine/npcdialog_component.go)
}
```

**New Components:**
```go
// v5.0 additions
type ChatComponent struct { /* ... */ }
type TradeComponent struct { /* ... */ }
```

**System Changes:**
```go
// v4.0
type DialogSystem struct { /* simple */ }

// v5.0
type NPCDialogSystem struct { /* Markov-based */ }
type ChatSystem struct { /* new */ }
type TradeSystem struct { /* new */ }
```

### Network Protocol Changes

**v4.0 Protocol:**
- Message types: 1-20
- No encryption

**v5.0 Protocol:**
- Message types: 1-30 (added chat, trade, image messages)
- E2E encryption for chat
- ACK/NACK for reliable delivery

**Incompatibility:** v4.0 clients cannot parse v5.0 messages.

## Upgrade Checklist

### For Players

- [ ] Backup save files (`saves/` directory)
- [ ] Download v5.0 client
- [ ] Test save file loading (automatic migration)
- [ ] Review new chat commands (`/help chat`)
- [ ] Configure privacy settings (`--profanity-filter`, `--auto-download-images`)
- [ ] Try NPC dialog (note: responses will vary)
- [ ] Test trading with trusted players (build trust gradually)

### For Server Operators

- [ ] Backup server world data
- [ ] Upgrade server binary to v5.0
- [ ] Update configuration (`--social-beta=true` for testing)
- [ ] Announce breaking change to players (client upgrade required)
- [ ] Monitor chat for spam (adjust rate limits if needed)
- [ ] Enable moderation hooks (image validation)
- [ ] Test multiplayer trade scenarios
- [ ] Document server-specific rules (chat etiquette, trading policies)

### For Modders/Contributors

- [ ] Update ECS component references (`DialogComponent` → `NPCDialogComponent`)
- [ ] Test mods with new social components
- [ ] Add chat integration if applicable
- [ ] Update documentation for API changes
- [ ] Test network protocol compatibility
- [ ] Verify save/load with new components

## Rollback Procedure

**If you need to revert to v4.0:**

1. **Restore backup:** `cp saves/<filename>.v4.0.backup saves/<filename>.sav`
2. **Downgrade binary:** Install v4.0 client
3. **Data loss:** Any chat history, trust scores, or trades from v5.0 will be lost
4. **Warning:** v5.0 saves cannot be loaded in v4.0 (use backup created during migration)

## Troubleshooting

### Migration Errors

**"Save file version incompatible":**
- Error indicates save file corruption, not version issue
- Try loading backup: `saves/<filename>.v4.0.backup`

**"Failed to add social components":**
- Rare migration error
- Report issue with save file seed and error message
- Workaround: Start new game, copy world seed

**"NPC dialog not working":**
- Check flag: `--deterministic-dialog=false` (should be false for v5.0 mode)
- Verify corpus files exist: `pkg/procgen/dialog/corpora/`
- Fallback: Enable `--deterministic-dialog=true`

### Performance Issues

**FPS dropped after upgrade:**
- Disable social features: `--disable-social`
- Reduce chat history size: `--max-chat-messages=50`
- Disable image auto-download: `--auto-download-images=false`

**Network lag increased:**
- Check bandwidth: social adds ~25KB/s overhead
- Reduce image sharing frequency
- Use local chat instead of global

### Gameplay Changes

**NPCs give different responses:**
- Expected behavior in v5.0 (Markov-based dialog)
- Revert to v4.0 behavior: `--deterministic-dialog=true`

**Can't drop items anymore:**
- Use trade system instead (`/trade <player>`)
- Safer than dropping (atomic transfers)

## FAQ

**Q: Do I need to start a new game for v5.0?**  
A: No. v4.0 saves are fully compatible. Migration is automatic.

**Q: Will my world seed still work the same?**  
A: Yes. World generation is unchanged. Same seed produces same world.

**Q: Can I play v5.0 solo without social features?**  
A: Yes. Use `--disable-social` flag for v4.0 experience.

**Q: Are v5.0 servers compatible with v4.0 clients?**  
A: No. v4.0 clients must upgrade to v5.0.

**Q: Can I disable specific social features?**  
A: Yes. Use `--disable-chat`, `--disable-trading`, `--deterministic-dialog`, etc.

**Q: Will my mods still work?**  
A: Most mods will work. Check for `DialogComponent` usage (renamed to `NPCDialogComponent`).

**Q: How do I revert to v4.0?**  
A: Restore backup save file and downgrade binary. v5.0 saves cannot be loaded in v4.0.

**Q: Is chat mandatory in multiplayer?**  
A: No. Chat is optional. Use `--disable-chat` to disable.

**Q: Do NPCs remember conversations?**  
A: Yes (last 10 interactions). Conversation history provides context for responses.

**Q: Is trading safe?**  
A: Yes. Two-phase commit ensures atomic transfers. Items rollback on any failure.

## Support

**Issues:** `github.com/opd-ai/venture/issues`  
**Discussions:** `github.com/opd-ai/venture/discussions`  
**Documentation:** `docs/SOCIAL_SYSTEMS.md`, `docs/API_REFERENCE.md`

---

**Last Updated:** November 2025  
**Maintained By:** Venture Development Team  
**Version:** 5.0 Production
