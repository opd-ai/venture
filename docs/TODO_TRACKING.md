# TODO Tracking for Future Implementation

This document tracks TODO items found in the codebase as of v4.1 stable release.
These represent future features that should be implemented in upcoming versions.

## Network Systems

### Trade System (`pkg/network/trade/system.go`)
- [ ] Notify recipient when trade is initiated
- [ ] Implement two-phase commit for trade execution
- [ ] Validate item ownership before trade
- [ ] Check trust scores between players
- [ ] Atomic ownership transfer for items
- [ ] Update trust scores after completed trades
- [ ] Implement trade cancellation logic

### Chat System (`pkg/network/chat/system.go`)
- [ ] Get current timestamp for messages (currently using 0)
- [ ] Implement end-to-end encryption for private messages
- [ ] Broadcast messages to recipients based on channel type
- [ ] Implement rate limiting to prevent spam
- [ ] Full E2E encryption implementation

### Federation Protocol (`pkg/network/federation/protocol.go`)
- [ ] Implement TLS handshake for secure server connections
- [ ] Exchange and validate server certificates
- [ ] Negotiate capabilities between federated servers
- [ ] Serialize complete player state for server transfers
- [ ] Two-phase commit for player transfers
- [ ] Rollback mechanism on transfer failure
- [ ] Check for players in range of cross-server portals
- [ ] Initiate transfer when player activates portal

## World Systems

### Persistence (`pkg/world/persistence.go`)
- [ ] Trigger auto-save at regular intervals

## Engine Systems

### Book System (`pkg/engine/book_spawning.go`)
- [ ] Create dedicated BookshelfDialogProvider in dialog_system.go

### Interaction System (`pkg/engine/interaction_system.go`)
- [ ] Start lock-picking mini-game via MiniGameSystem
- [ ] Check player inventory for key when bookshelf.RequiresKey is true

### Mini-Game System (`pkg/engine/minigame_system.go`)
- [ ] Generate reward items based on game type
- [ ] Integrate reward items with InventorySystem.AddItem() (Phase 27.3 feature)

### Discovery System (`pkg/engine/discovery_system.go`)
- [ ] Integrate with quest generation system for dynamic quest triggers

### Vehicle Combat System (`pkg/engine/vehicle_combat_system.go`)
- [ ] Integrate with projectile system when available for vehicle weapons

## Implementation Priority

### High Priority (v5.0)
- Trade system two-phase commit (critical for multiplayer integrity)
- Chat system end-to-end encryption (security)
- Mini-game reward integration (gameplay completeness)

### Medium Priority (v5.1)
- Federation protocol implementation (server-to-server communication)
- Lock-picking mini-game (gameplay depth)
- Vehicle projectile system integration (combat richness)

### Low Priority (Future)
- Trust score system (social features)
- Auto-save triggers (quality of life)
- Bookshelf dialog provider (minor polish)

## Notes

All TODO comments have been removed from source code and tracked here.
New TODO items should be added to this document or created as GitHub issues.
When implementing features, move them from this document to ROADMAP and update status.
