# Social Systems Guide

**Version:** 1.0.0  
**Date:** February 2026  
**Status:** Production

## Overview

Venture v5.0 introduces comprehensive social systems for player-to-player communication and interaction in multiplayer mode. These features enhance the cooperative gameplay experience while maintaining the game's core focus on procedural generation and action-RPG mechanics.

## Features

### 1. Text Chat System

**Channels:**
- **Global**: Visible to all players on the server (3-second cooldown)
- **Local**: Players within 10-tile radius (1-second cooldown)
- **Party**: Party members only (0.5-second cooldown)
- **Whisper**: Direct message to specific player (0.5-second cooldown)

**Security:**
- End-to-end encryption using Diffie-Hellman key exchange
- Server cannot read message content (privacy-first design)
- Client-side profanity filter (optional, configurable)

**Rate Limiting:**
- Per-channel cooldowns prevent spam
- Violation triggers progressive mute: 30s → 60s → 120s (max 10 minutes)
- Automatic unmute after expiry

**Range Extension Items:**
- **Megaphone**: Extends local chat radius to 30 tiles (10 uses, consumable)
- **Walkie-Talkie**: Unlimited range for local chat (equippable)

### 2. NPC Dialog System

**Dynamic Conversations:**
- NPCs generate responses using Markov chain models
- Genre-specific vocabulary (fantasy: "thee/thou", sci-fi: "protocol/system")
- Personality traits influence word selection
- Conversation history provides context for responses

**Deterministic Mode:**
- Set `DeterministicMode` on the `NPCDialogComponent` for reproducible dialog
- Useful for testing and debugging
- Fallback to template-based responses

**Multi-Player Dialogs:**
- Each player sees personalized NPC responses

### 3. Image Sharing

**Upload & Share:**
- Supported formats: PNG, JPEG, GIF (non-animated)
- Maximum size: 500KB per image
- Maximum dimensions: 2048×2048 pixels
- Automatic thumbnail generation (128×128)

**Privacy Controls:**
- Manual acceptance required for full image download
- Thumbnails auto-download for preview
- Images expire after 10 minutes or sender disconnect
- Report images via right-click menu

**Moderation:**
- Server-side validation (size, type, dimensions)
- Moderation hooks for future ML-based NSFW detection
- Rate limit: 1 image per 60 seconds

### 4. Item Trading

**Two-Phase Commit Protocol:**
1. **Propose**: Player A sends trade offer (items offered + items requested)
2. **Review**: Player B reviews and accepts/rejects/counter-proposes
3. **Validate**: Server checks ownership, proximity, trust, tradability
4. **Commit**: Server atomically transfers items between inventories
5. **Acknowledge**: Both players receive confirmation or rollback notification

**Proximity Requirements:**
- Players must be within 5 tiles to initiate trade
- Trade auto-cancels if distance exceeds 10 tiles during negotiation
- Server uses lag compensation for fair proximity checks

**Trust Mechanics:**
- Trust score ranges from 0.0 (untrusted) to 1.0 (fully trusted)
- Default: 0.5 for new trading partners
- Successful trades increase trust (+0.05)
- Failed trades decrease trust (-0.1)
- High trust (>0.8): Can trade rare/legendary items
- Low trust (<0.3): Limited to common/uncommon items, max 5 items per trade

**Failure Handling:**
- Atomic rollback on any failure
- Detailed error messages: proximity, trust, ownership, item moved
- Trade history tracks successes and failures

## Chat Commands

### Global Chat
```
/g <message>  - Send global chat message
/global <message>
```

### Local Chat
```
/l <message>  - Send local chat message (10-tile radius)
/local <message>
```

### Party Chat
```
/p <message>  - Send party chat message
/party <message>
```

### Whisper (Direct Message)
```
/w <player> <message>  - Send whisper to specific player
/whisper <player> <message>
/tell <player> <message>
```

### Channel Management
```
/join <channel>  - Subscribe to channel (global, local, party)
/leave <channel> - Unsubscribe from channel
/channels        - List active channels
```

### Utility Commands
```
/mute <player>   - Ignore messages from player
/unmute <player> - Unignore player
/clear           - Clear chat history
/help chat       - Show chat help
```

## Trading Commands

### Initiate Trade
```
/trade <player>  - Request trade with nearby player
```

### Trade Actions (in trade UI)
- **Accept**: Confirm trade proposal
- **Reject**: Decline trade proposal
- **Cancel**: Cancel your own trade proposal
- **Counter**: Modify proposal (add/remove items)

### Trade History
```
/trades          - View recent trade history
/trust <player>  - View trust score with player
```

## Configuration

Social systems are enabled by default when running in multiplayer mode. Configuration is handled programmatically through the ECS components and system initialization.

### NPC Dialog
Dialog behavior is controlled via the `NPCDialogComponent`:
- `DeterministicMode`: Set to `true` for template-based dialog (useful for testing)
- `NPCPersonality`: Adjusts friendliness, formality, and verbosity traits
- Markov chain order is set when creating the generator (2-3 recommended)

## Best Practices

### Chat Etiquette
1. **Respect cooldowns**: Don't spam messages
2. **Use appropriate channels**: Global for server-wide, Local for nearby players
3. **Whisper for private**: Use whispers for one-on-one conversations
4. **Report abuse**: Right-click messages to report violations

### Safe Trading
1. **Review carefully**: Double-check items before accepting
2. **Build trust gradually**: Start with small trades
3. **Stay close**: Don't move away during active trades
4. **Check trust scores**: `/trust <player>` before major trades
5. **Report scams**: Use `/report <player>` for suspicious behavior

### NPC Interactions
1. **Provide context**: NPCs use conversation history for better responses
2. **Be patient**: Response generation takes <50ms but may seem instant
3. **Experiment**: Different inputs yield different responses (non-deterministic)

## Troubleshooting

### Chat Issues

**Message not delivered:**
- Check cooldown timer (bottom of chat window)
- Verify recipient is online (for whispers)
- Ensure within range (for local chat)
- Check if muted (rate limit violation)

**Can't read encrypted messages:**
- Reconnect to server (triggers key exchange)

**Profanity filter issues:**
- The profanity filter is opt-in and disabled by default
- Enable it programmatically via `ProfanityFilter.Enable()` (see `pkg/network/profanity.go`)
- Custom word lists can be loaded with `LoadWordListFromFile()`
- Report filter issues on GitHub for improvements

### Trading Issues

**Trade rejected automatically:**
- **Proximity**: Move closer (within 5 tiles)
- **Trust**: Build trust with smaller trades first
- **Ownership**: Item may have been moved/dropped
- **Rarity**: High trust required for rare/legendary items

**Trade stuck "pending":**
- Wait for partner to accept/reject
- Auto-cancels after 30 seconds
- Manual cancel: Click "Cancel" button

**Items disappeared after failed trade:**
- Items always rollback to original owner on failure
- Check inventory tabs (may be in different category)
- Reconnect if inventory desync occurs

### NPC Dialog Issues

**Repetitive responses:**
- Provide varied inputs (conversation history matters)
- Different NPCs have different personalities

**Inappropriate/nonsensical dialog:**
- Set `DeterministicMode = true` on the `NPCDialogComponent` for template-based fallback
- Report issue with seed/NPC ID for debugging

## Performance Considerations

### Chat Bandwidth
- Text messages: <1KB each
- Average usage: <10KB/s per player
- Images: 500KB max, 1 per minute
- Total overhead: ~25KB/s (well within 100KB/s budget)

### Generation Times
- NPC dialog: <50ms per response
- Chat encryption: ~50µs per message
- Chat decryption: ~40µs per message
- Trade validation: <10ms

### Memory Usage
- Chat history: ~100 messages × 1KB = 100KB per player
- Trade state: ~10KB per active trade
- Dialog cache: ~50KB per NPC
- Total: <200KB per player (minimal overhead)

## Privacy & Security

### What Server Can See
- Image metadata (sender, timestamp, size)
- Trade proposals (item IDs, quantities, participants)
- NPC dialog (server-generated)
- Player positions (for proximity checks)

### What Server Cannot See
- Chat message content (E2E encrypted)
- Player passwords or auth tokens
- Image content (stored ephemerally)

### User Controls
- Block list: `/mute <player>`
- Report system: `/report <player> <reason>`
- Manual image acceptance (default)

## Future Enhancements (v5.1+)

**Planned Features:**
- ML-based image moderation (NSFW detection)
- Advanced spam filters (pattern recognition)
- Guild/clan systems
- Persistent chat rooms
- Voice chat (long-term consideration)
- Trade marketplace/auction house

**Community Requests:**
- Image history/scrollback
- Trust score persistence
- Trading achievements/statistics
- Custom emojis
- Chat log export

## Support & Feedback

**Bug Reports:**
- GitHub Issues: `github.com/opd-ai/venture/issues`
- In-game: `/report <description>`

**Feature Requests:**
- Discussions: `github.com/opd-ai/venture/discussions`
- Discord: Link in README.md

**Documentation:**
- API Reference: `docs/API_REFERENCE.md`
- Technical Spec: `docs/TECHNICAL_SPEC.md`
- Migration Guide: `docs/MIGRATION_V5.md`

---

**Last Updated:** February 2026  
**Maintained By:** Venture Development Team  
**Version:** 1.0.0 Production
