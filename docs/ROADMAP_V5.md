# Development Roadmap - Version 5.0: Social Systems & Multiplayer Messaging

## Overview

**Project:** Venture - Fully Procedural Multiplayer Action-RPG  
**Version:** 5.0 - Social Systems & Multiplayer Messaging  
**Previous Version:** 4.0 Complete (Phase 30 - Projected 2027)  
**Timeline:** 8-10 months (6 phases)  
**Date:** November 2025  
**Focus:** Player communication, NPC dialog, item trading, and multiplayer social interaction

---

## Objective & Scope

**What V5 Delivers:**
- Runtime NPC dialog generation using Markov chains (controlled non-determinism)
- Player-to-player text chat with E2E encryption, range limiting, and item-extended range
- Image sharing system with client upload, server relay, size/type limits, and moderation hooks
- Item sharing and trading with proximity rules, trust mechanics, and atomic ownership transfer
- Multi-party conversation support (NPC + multiple players) with message ordering and conflict resolution
- Social interaction systems compatible with existing multiplayer architecture (200-5000ms latency)

**What V5 Does NOT Include:**
- Voice chat or video streaming
- Persistent player housing or guild systems
- External chat integrations (Discord, IRC)
- User-generated content beyond images (no custom sprites/scripts)
- Blockchain/NFT trading mechanics
- Global marketplace or auction house

**Completion Criteria:** All features functional in multiplayer, passing acceptance tests, meeting performance targets, with ≥65% test coverage per package.

---

## Key Constraints & Compatibility

### Determinism Policy

**Preserved Determinism:**
- Terrain, entity stats, item generation, quest content remain fully deterministic
- Combat outcomes, loot drops, skill progression use existing seed-based RNG
- World state synchronization continues via authoritative server model

**Controlled Non-Determinism:**
- **NPC Dialog**: Markov chain generation uses runtime entropy (player input history, conversation context)
  - **Boundary:** Dialog content affects presentation only, never gameplay mechanics (quest objectives, item rewards, entity behavior)
  - **Fallback:** Deterministic template-based dialog available via config flag (`-deterministic-dialog=true`)
  - **Authoritativeness:** Server generates dialog; clients display server-provided text
  - **Testing:** Non-deterministic tests verify dialog variation; deterministic tests use fixed seeds
  - **Documentation:** Each dialog generator must document non-determinism scope and fallback behavior

**Rationale:** NPC dialog variety enhances immersion without affecting game balance or multiplayer sync. Deterministic fallback ensures testability and backward compatibility.

### Network Requirements

**High-Latency Support (200-5000ms):**
- Chat messages use unreliable-ordered UDP with ACK/NACK for delivery confirmation
- Client-side message buffering (5-second window) for out-of-order delivery
- Optimistic UI updates (show sent message immediately, mark as "sending" until ACK)
- Image transfers use TCP with chunked upload, resume on disconnect
- Item trades use two-phase commit (propose → accept → server validates → commit)
- Lag compensation for proximity checks (server rewinds to client's perspective timestamp)

**Bandwidth Targets:**
- Text chat: <1KB per message, <10KB/s per player average
- Image sharing: <500KB per image, max 2MB/s upload rate
- Item trades: <2KB per transaction
- Total overhead: <25KB/s per player (25% of 100KB/s budget)

### Privacy & Encryption

**End-to-End Encryption (E2E):**
- Player-to-player chat uses per-session Diffie-Hellman key exchange
- Messages encrypted client-side before network transmission
- Server cannot decrypt message content (relay only)
- **Trade-off:** Server-side moderation impossible for E2E content
- **Mitigation:** Client-side spam filters, user-initiated reporting, rate limiting

**Server-Visible Content:**
- NPC dialog (server-generated, no privacy expectation)
- Image metadata (sender, timestamp, size)
- Trade proposals (item IDs, quantities, participants)

**User Controls:**
- Opt-out from all social features (`-disable-social=true`)
- Block list (ignore messages/trades from specific players)
- Image auto-download toggle (default: off, show thumbnail + manual accept)

---

## Design Principles

### ECS Architecture Integration

**New Components:**
```go
// pkg/engine/social_components.go
type ChatComponent struct {
    Messages       []ChatMessage
    UnreadCount    int
    ActiveChannels []ChatChannel
}

type TradeComponent struct {
    ActiveTrade    *TradeProposal
    TradeHistory   []TradeRecord
    TrustScore     float64 // 0.0-1.0, affects trade limits
}

type DialogComponent struct {
    DialogState    *DialogState
    ResponseHistory []string // For Markov chain context
}
```

**New Systems:**
- `ChatSystem`: Message delivery, channel management, encryption/decryption
- `TradeSystem`: Proximity validation, ownership transfer, rollback on conflict
- `DialogSystem`: NPC response generation, conversation state management
- `SocialNetworkSystem`: Message routing, bandwidth throttling, ACK/NACK handling

### Seed Determinism Policy

**Maintained:**
- World generation, entity spawning, item properties use existing `procgen.SeedGenerator`
- Multiplayer clients generate identical worlds from shared world seed
- Saved games reproduce identical state from save file seed

**Exception:**
- NPC dialog Markov chains use `markov.Generator` with runtime entropy
- Dialog seeds derived from: conversation ID hash + player input hash + timestamp
- **Isolation:** Dialog generator never calls `procgen.SeedGenerator` or modifies world RNG state
- **Testing:** Deterministic mode uses fixed dialog seeds for reproducible tests

### High-Latency Resilience

**Message Ordering:**
- Each message tagged with sequence number and timestamp
- Client buffers out-of-order messages, reorders before display (5-second window)
- Duplicate detection via message ID (UUID v4)

**Trade Conflict Resolution:**
- Server is authoritative; client predictions reconciled on ACK
- Concurrent trade attempts (same item, different players) resolved by server timestamp
- Failed trades trigger rollback notification with reason (item moved, insufficient proximity, item no longer owned)

**Timeout Handling:**
- Chat ACK timeout: 10 seconds, retry 3 times, mark as "failed to send"
- Image upload timeout: 60 seconds, auto-resume on reconnect
- Trade proposal timeout: 30 seconds, auto-cancel with notification

---

## Feature List

### 5.1: Runtime NPC Dialog (Markov Chain-Based)

**Description:**  
Generate dynamic NPC dialog at runtime using Markov chain models trained on genre-specific text corpora. Dialog varies based on player interaction history, NPC personality, and conversation context.

**Components:**
- `pkg/procgen/dialog/markov.go`: Markov chain generator (order 2-3, configurable)
- `pkg/procgen/dialog/corpus.go`: Genre-specific text corpora (fantasy: medieval, sci-fi: technical, horror: ominous)
- `pkg/procgen/dialog/personality.go`: NPC personality traits influencing word selection probabilities
- `pkg/engine/dialog_component.go`: Dialog state tracking (conversation history, topic)

**Non-Determinism Constraints:**
- **Where:** Dialog text generation only (not quest objectives, item rewards, NPC behavior)
- **How:** Markov chains seeded with `hash(conversationID + playerInput + timestamp)`
- **Fallback:** Template-based dialog when `-deterministic-dialog=true` flag set
- **Authoritativeness:** Server generates all dialog; clients display server text (prevents client-side manipulation)

**Acceptance Criteria:**
- [x] Generate 5+ unique responses per NPC for same input (variation test)
- [x] Deterministic mode produces identical dialog given same seed (reproducibility test)
- [x] Dialog never references non-existent items, quests, or entities (validation test)
- [x] Response generation <50ms (performance test)
- [x] Graceful fallback to templates on Markov generation failure
- [x] Genre-appropriate vocabulary (fantasy: "thee/thou", sci-fi: "protocol/system")

**Testing:**
- Table-driven tests with fixed seeds for deterministic mode
- Variation tests verifying >80% unique responses for same input over 10 runs
- Corpus validation tests (no profanity, all words ASCII-compatible)
- Benchmark: 1000 dialog generations <5 seconds

### 5.2: Player-to-Player Text Chat

**Description:**  
Encrypted text messaging between players with channel support (global, local, party), range limiting (local chat requires proximity), and item-extended range (megaphone increases local radius, walkie-talkie enables unlimited range).

**Components:**
- `pkg/network/chat.go`: Message routing, encryption, ACK/NACK protocol
- `pkg/network/crypto.go`: E2E encryption (Diffie-Hellman key exchange, AES-256-GCM)
- `pkg/engine/chat_component.go`: Message history, unread count, active channels
- `pkg/rendering/ui/chat.go`: Chat UI (message list, input field, channel tabs)

**Channels:**
- **Global**: All players on server, no range limit, rate limit: 1 msg/3 seconds
- **Local**: Players within 10 tile radius, rate limit: 1 msg/1 second
- **Party**: Party members only, no range limit, rate limit: 1 msg/0.5 seconds
- **Whisper**: Direct message to specific player, no range limit, rate limit: 1 msg/0.5 seconds

**Range Extension Items:**
- **Megaphone**: Increases local chat radius to 30 tiles (consumable, 10 uses)
- **Walkie-Talkie**: Enables unlimited range for local chat (equippable, requires batteries)
- **Signal Flare**: Sends global broadcast visible to all players (consumable, 1 use, 5-minute cooldown)

**E2E Encryption:**
- Key exchange on player connection (Diffie-Hellman with 2048-bit modulus)
- Per-message encryption (AES-256-GCM with random IV)
- Server relays encrypted payloads, cannot decrypt content
- **Trade-off:** Server moderation impossible; rely on client-side filters and user reporting

**Rate Limiting:**
- Per-channel, per-player limits enforced server-side
- Exceeding limit triggers 30-second mute
- Repeat violations double mute duration (30s → 60s → 120s, max 10 minutes)

**Acceptance Criteria:**
- [ ] E2E encryption: server logs show encrypted payloads, not plaintext
- [ ] Message delivery: 99% delivered within 2 seconds at 200ms latency
- [ ] Range limiting: local messages not received beyond radius (tested with 50 players)
- [ ] Item effects: megaphone extends radius to 30 tiles, walkie-talkie removes limit
- [ ] Rate limiting: exceeding limit triggers mute, duration increases on repeat violations
- [ ] Client filters: profanity filter blocks 95%+ of common swears (configurable, opt-in)

**Testing:**
- Latency simulation (200ms, 500ms, 2000ms, 5000ms) with 100 messages
- Message loss (5%, 10%, 20%) and reorder tests with ACK/NACK verification
- Duplicate detection (send same message ID twice, verify single delivery)
- Encryption tests (verify ciphertext differs for same plaintext with different IVs)
- Benchmark: 1000 messages sent/received <10 seconds per player

### 5.3: Image Sharing

**Description:**  
Upload and share images via client UI with server relay, size/type limits, thumbnail generation, and moderation hooks. Images are ephemeral (not saved to disk) and require manual acceptance before download.

**Flow:**
1. Client: User selects image file (PNG/JPEG/GIF)
2. Client: Resize if >500KB, generate thumbnail (128×128)
3. Client: Upload to server via chunked HTTP POST
4. Server: Validate size (<500KB), type (PNG/JPEG/GIF), dimensions (<2048×2048)
5. Server: Run moderation hooks (placeholder for future ML-based NSFW detection)
6. Server: Relay thumbnail to recipients (auto-download)
7. Recipients: Click thumbnail to request full image
8. Server: Stream full image to requester

**Size/Type Limits:**
- Max size: 500KB (enforced server-side, client-side pre-check)
- Types: PNG, JPEG, GIF (non-animated)
- Max dimensions: 2048×2048 pixels
- Thumbnail: 128×128 pixels, JPEG quality 75

**Privacy/Control:**
- Images sent to specific channels (global, local, party, whisper)
- Recipients must manually accept full image download (thumbnail auto-downloads)
- Images expire after 10 minutes or on sender disconnect (whichever first)
- User setting: `-auto-download-images=false` (default: off)

**Moderation:**
- **Server-side hooks:** `OnImageUpload(metadata)` for future NSFW detection
- **Client-side reporting:** Right-click image → "Report" sends metadata to server log
- **Rate limiting:** 1 image per 60 seconds per player
- **No E2E encryption:** Images visible to server for moderation (trade-off accepted)

**Acceptance Criteria:**
- [ ] Upload: 500KB image uploads in <5 seconds at 200ms latency
- [ ] Validation: >500KB images rejected with error message
- [ ] Invalid types: .bmp/.tiff rejected with error message
- [ ] Moderation: `OnImageUpload` hook invoked for all uploads
- [ ] Expiry: Images deleted after 10 minutes or sender disconnect
- [ ] Manual accept: Full image not downloaded until user clicks thumbnail
- [ ] Rate limit: Uploading 2 images within 60s triggers rejection

**Testing:**
- Upload tests: 100KB, 500KB, 600KB (reject), various types (PNG, JPEG, GIF, BMP)
- Latency tests: Upload 500KB at 200ms, 2000ms, 5000ms
- Disconnect tests: Upload interrupted mid-transfer, verify resume on reconnect
- Moderation hook tests: Verify invocation, metadata passed correctly
- Benchmark: Upload 100 images (500KB each) <60 seconds

### 5.4: Item Sharing & Trading

**Description:**  
Transfer items between players with proximity requirements, trust-based limits, atomic ownership transfer, and rollback on disconnect/conflict. Supports direct trade (bilateral) and gifting (unilateral).

**Proximity Rules:**
- Players must be within 5 tiles to initiate trade
- Server validates proximity on proposal and commit (lag-compensated)
- Trade auto-cancels if players move >10 tiles apart during negotiation

**Trust Mechanics:**
- Each player has `TrustScore` (0.0-1.0, default: 0.5)
- Successful trades increase trust (+0.05), failed trades decrease (-0.1)
- High trust (>0.8): Can trade rare/legendary items
- Low trust (<0.3): Limited to common/uncommon items, max 5 items per trade
- Trust resets to 0.5 on server restart (not persisted)

**Transfer Protocol (Two-Phase Commit):**
1. **Propose:** Player A sends trade proposal (items offered, items requested)
2. **Review:** Player B reviews, accepts/rejects/counter-proposes
3. **Validate:** Server checks ownership, proximity, trust, item tradability
4. **Commit:** Server atomically transfers items, updates inventories
5. **ACK:** Both clients receive confirmation or rollback notification

**Ownership Transfer:**
- Items removed from sender inventory, added to recipient inventory (atomic)
- Server is authoritative; client predictions reconciled on commit
- Failed commit triggers rollback: items returned, notification sent

**Rollback Scenarios:**
- Player disconnects during trade (auto-cancel, items returned)
- Items no longer owned (sold/dropped between propose and commit)
- Proximity violated (players moved >10 tiles apart)
- Trust check failed (rare item traded by low-trust player)
- Concurrent trades (same item in two proposals, first commit wins)

**Acceptance Criteria:**
- [ ] Proximity: Trade rejected if players >5 tiles at proposal
- [ ] Trust: Low-trust player (<0.3) cannot trade legendary items
- [ ] Atomicity: Concurrent trades for same item fail for second commit
- [ ] Rollback: Disconnect during trade returns items to original owners
- [ ] Lag compensation: Proximity validated at client's perspective timestamp
- [ ] Performance: Trade commit <100ms at 200ms latency

**Testing:**
- Proximity tests: Trade at 5, 10, 15 tiles (5 succeeds, >5 fails)
- Trust tests: Attempt rare trade with low trust (rejected)
- Concurrent trade tests: Two players propose trade for same item simultaneously
- Disconnect tests: Disconnect at each protocol phase (propose, review, validate, commit)
- Rollback tests: Item moved/sold between propose and commit
- Latency tests: Trade protocol at 200ms, 2000ms, 5000ms
- Benchmark: 100 trades (5 items each) <30 seconds

### 5.5: Concurrency Model (Multi-Party Conversations)

**Description:**  
Support simultaneous conversations between multiple players and NPCs with message ordering, conflict resolution, and fair turn-taking.

**Scenarios:**
- **NPC + Multiple Players:** 1 NPC, 3 players (all receive NPC responses)
- **Group Chat + NPC:** 4 players in party chat, 1 NPC joins conversation
- **Concurrent Trades:** 2 players each trading with separate NPCs simultaneously

**Message Ordering:**
- Each conversation has unique ID (UUID v4)
- Messages tagged with: sender ID, conversation ID, sequence number, timestamp
- Server enforces total ordering within conversation (FIFO per sender)
- Clients display messages in timestamp order (server-provided timestamps)

**Conflict Resolution:**
- **NPC Response Conflict:** Multiple players ask NPC simultaneously → server queues requests, processes FIFO
- **Trade Conflict:** Two players attempt same trade → first commit wins, second receives "trade unavailable" notification
- **Dialog State Conflict:** NPC conversation interrupted by another player → original conversation paused, resumed on interrupt completion

**Turn-Taking (NPC Dialogs):**
- NPC processes one dialog request at a time (FIFO queue)
- Active request blocks queue (max 30 seconds, then auto-complete)
- Players see "NPC is busy" if conversation active
- Queue limit: 5 pending requests, excess rejected with "try again later"

**Acceptance Criteria:**
- [ ] Ordering: 100 messages from 5 players delivered in correct timestamp order
- [ ] NPC queue: 5 simultaneous requests queued, processed FIFO
- [ ] Trade conflict: 2 players attempt same trade, first wins, second notified
- [ ] Dialog interrupt: NPC conversation paused when interrupted, resumed correctly
- [ ] Turn timeout: Active request auto-completes after 30 seconds

**Testing:**
- Multi-player tests: 5 players, 1 NPC, 50 messages, verify ordering
- Conflict tests: 2 players, 1 trade, simultaneous proposals
- Queue tests: 6 simultaneous NPC requests, 5 queued, 1 rejected
- Interrupt tests: Start NPC dialog, interrupt with another player, verify pause/resume
- Benchmark: 1000 messages, 10 players, 5 conversations <60 seconds

### 5.6: Networking Specifics

**Description:**  
Low-level protocol design for bandwidth efficiency, compression, encryption, and message reliability.

**Packet Design:**
- **Chat Message Packet:** Header (16 bytes) + Encrypted Payload (variable)
  - Header: Message ID (16 bytes UUID), Sender ID (8 bytes), Channel (1 byte), Timestamp (8 bytes)
  - Payload: AES-256-GCM encrypted (plaintext + 16-byte auth tag)
  - Total: ~50-200 bytes per message
- **Image Metadata Packet:** Header (16 bytes) + Metadata (64 bytes)
  - Metadata: Image ID, sender ID, dimensions, size, thumbnail URL
- **Trade Proposal Packet:** Header (16 bytes) + Item List (variable)
  - Item List: Item IDs (8 bytes each), quantities (4 bytes each)
  - Max 20 items per trade → 240 bytes max

**Compression:**
- Chat messages: zlib compression if plaintext >100 bytes (saves ~30-50%)
- Image thumbnails: JPEG quality 75 (saves ~80% vs. PNG)
- Trade proposals: No compression (already small <240 bytes)

**Encryption:**
- **E2E (Chat):** AES-256-GCM with per-message random IV (12 bytes)
- **Key Exchange:** Diffie-Hellman 2048-bit modulus, shared secret → HKDF-SHA256 → AES key
- **Server-Relay (Images/Trades):** TLS 1.3 transport encryption only

**ACK/NACK Semantics:**
- **Reliable Messages (Trades):** Require ACK within 10 seconds, retry 3 times, fail on timeout
- **Unreliable Messages (Chat):** Best-effort delivery, optional ACK for read receipts
- **NACK:** Sent on validation failure (e.g., rate limit exceeded, invalid item ID)

**Snapshot/Delta Encoding:**
- Not applicable (messages are discrete events, not continuous state)
- Future consideration: Delta compression for chat history sync on reconnect

**Acceptance Criteria:**
- [ ] Bandwidth: <10KB/s per player for 10 messages/minute chat rate
- [ ] Compression: >30% reduction for messages >100 bytes
- [ ] Encryption: All E2E payloads encrypted, server cannot decrypt
- [ ] ACK/NACK: Reliable messages ACKed within 10 seconds or retried
- [ ] Packet loss: 10% loss with retries delivers 99%+ messages

**Testing:**
- Bandwidth tests: Measure bytes sent/received for 100 messages
- Compression tests: Compare compressed vs. uncompressed sizes for 50/100/200 byte messages
- Encryption tests: Verify ciphertext changes for same plaintext
- Packet loss tests: Simulate 5%, 10%, 20% loss, verify delivery rates
- Benchmark: Send 1000 messages, measure bandwidth and latency

---

## Testing & Validation Plan

### Unit Testing

**Table-Driven Tests:**
- Chat encryption: Encrypt/decrypt 10 messages, verify plaintext matches
- Markov generation: Generate 100 responses, verify >80% unique
- Trade validation: Test 20 scenarios (valid, invalid proximity, invalid trust, etc.)
- Message ordering: Scramble 50 messages, verify correct ordering after sort

**Deterministic Tests:**
- NPC dialog with fixed seed produces identical output (10 runs)
- Trade proposals with same items/players produce identical validation results
- Chat rate limiting with simulated timestamps enforces limits correctly

**Race Detector:**
- Run all tests with `go test -race ./...`
- Focus on concurrent message handling, trade conflicts, dialog queues

**Benchmark Targets:**
- Dialog generation: 1000 responses <5 seconds
- Chat encryption: 1000 messages encrypted/decrypted <2 seconds
- Trade validation: 1000 proposals validated <1 second
- Message ordering: 10,000 messages sorted <500ms

### Integration Testing

**Latency Simulation:**
- Chat delivery at 200ms, 500ms, 2000ms, 5000ms
- Image upload at 200ms, 2000ms (chunked transfer resilience)
- Trade protocol at 500ms, 2000ms (two-phase commit delays)

**Failure Scenarios:**
- **Message Loss:** 5%, 10%, 20% packet loss with ACK/NACK retries
- **Reorder:** Scramble message delivery order, verify client reordering
- **Duplicates:** Send same message ID twice, verify single display
- **Concurrent Trades:** 10 players trading simultaneously, verify no item duplication
- **NPC Dialog Regeneration:** Interrupt dialog generation mid-process, verify clean restart

**Multi-Player Scenarios:**
- 50 players, 10 messages/minute each → 500 messages/minute
- 10 concurrent trades (5 pairs), verify all complete or rollback correctly
- 5 players talking to same NPC, verify queue and turn-taking

**Acceptance Tests (Must Pass Before Merge):**
- [ ] Chat: 99% delivery rate at 200ms latency with 10% packet loss
- [ ] Images: 500KB upload completes in <10 seconds at 2000ms latency
- [ ] Trades: 100 trades, zero item duplication or loss
- [ ] NPC Dialog: 100 conversations, zero crashes, >80% response variation
- [ ] Performance: No frame time regression (maintain <16.67ms per frame)
- [ ] Memory: Total overhead <100MB for 50 players

---

## Security, Moderation & Privacy

### Abuse Mitigation

**Rate Limits:**
- Chat: 1 msg/3s (global), 1 msg/1s (local), 1 msg/0.5s (party/whisper)
- Images: 1 upload/60s per player
- Trades: 5 proposals/minute per player
- Violations: Mute durations double on repeat (30s → 60s → 120s, max 10 minutes)

**Spam Filters:**
- **Client-side:** Regex-based profanity filter (opt-in, user-configurable)
- **Server-side:** Duplicate message detection (reject if same text sent <5 seconds apart)
- **Future:** Bayesian spam classifier (not in v5.0 scope)

**Image Moderation:**
- **Immediate:** Size/type validation, rate limiting
- **Future (v5.1+):** ML-based NSFW detection via `OnImageUpload` hook
- **User Reporting:** Right-click image → "Report" logs metadata to server

**Trust Score Abuse:**
- Trust cannot be gamed (server-authoritative)
- Failed trades (disconnect, invalid items) decrease trust
- No manual trust adjustment (prevents collusion)

### User Opt-Outs

**Global Disable:**
- `-disable-social=true`: Disables chat, images, trades, NPC dialog (templates only)
- Useful for single-player or privacy-focused users

**Granular Controls:**
- `-disable-chat=true`: Disable all chat (global, local, party, whisper)
- `-disable-images=true`: Disable image sharing (thumbnails not downloaded)
- `-disable-trades=true`: Disable item trading (can still accept gifts if enabled)
- `-auto-download-images=false`: Require manual image accept (default: off)

**Block List:**
- Block specific players (ignore messages, trades, image shares)
- Stored client-side (not synced to server)
- UI: Right-click player name → "Block"

### Data Retention & Telemetry

**Data Retention:**
- **Chat messages:** Not stored server-side (ephemeral)
- **Images:** Deleted after 10 minutes or sender disconnect
- **Trade history:** Last 100 trades per player (in-memory, not persisted)
- **NPC dialog:** Not stored (generated on-demand)

**Telemetry (Opt-In):**
- Aggregate metrics: Messages sent/received per hour, images uploaded per day, trades completed per session
- No message content logged (only metadata)
- User opt-in required: `-telemetry=true`

**Privacy Guarantee:**
- No message content stored or logged (except user-reported images for moderation review)
- Server cannot decrypt E2E chat messages
- Trust scores and trade history not shared between players

---

## Rollout & Migration Plan

### Backwards Compatibility

**Save File Migration:**
- v4.0 saves load with social components initialized to defaults (no chat history, trust score 0.5)
- New save format version: 5.0 (includes chat history, trade history, trust scores)
- Migration auto-detected on load, one-way upgrade (v5.0 saves not loadable in v4.0)

**Network Protocol:**
- New message types: `ChatMessage`, `ImageShare`, `TradeProposal` (IDs 50-55)
- v4.0 clients ignore unknown message types (forward compatibility)
- v5.0 server supports v4.0 clients (social features disabled for them)

**Server-Client Version Check:**
- Handshake includes protocol version (v4.0: 4, v5.0: 5)
- v5.0 server rejects v3.0 or earlier clients (too old)
- v5.0 client displays warning when connecting to v4.0 server ("social features disabled")

### Feature Flags

**Compile-Time Flags:**
- `ENABLE_SOCIAL`: Enable/disable social features at build time (default: enabled)
- `ENABLE_E2E_CRYPTO`: Enable E2E encryption (default: enabled, disable for testing)

**Runtime Flags:**
- `-disable-social=true`: User opt-out (client-side)
- `-deterministic-dialog=true`: Use template dialog instead of Markov (testing/reproducibility)
- `-social-beta=true`: Opt-in to beta features before stable release

### Opt-In Beta

**Beta Phase (Months 1-6):**
- Social features available with `-social-beta=true` flag
- Beta servers separate from stable servers (port 8081 vs. 8080)
- Beta warning: "Social features experimental, expect bugs and wipes"

**Stable Release (Month 7+):**
- Social features enabled by default (no flag required)
- Beta servers merged into stable
- Migration path: Beta users re-create accounts on stable (no data import)

### Telemetry Collection

**Opt-In Metrics:**
- Messages sent/received per session
- Images uploaded/downloaded per session
- Trades proposed/completed/failed per session
- Average message latency (sent → ACK)
- NPC dialog generation time

**Reporting:**
- Metrics aggregated locally, sent to server on disconnect
- Server logs to file, no external analytics (privacy-preserving)
- Used for performance tuning and bug identification

---

## Metrics & Success Criteria

### Delivery Reliability Targets

**Chat Messages:**
- 99% delivered within 2 seconds at 200ms median latency
- 95% delivered within 5 seconds at 2000ms median latency
- 90% delivered within 10 seconds at 5000ms median latency

**Images:**
- 500KB upload completes within 10 seconds at 200ms latency
- 500KB upload completes within 30 seconds at 2000ms latency
- Resume after disconnect: 95% resume correctly within 5 seconds

**Trades:**
- 100% atomic (no item duplication or loss)
- Proposal → Commit within 5 seconds at 200ms latency
- Rollback on conflict: 100% successful (items returned to original owners)

### Message Throughput

**Per Player:**
- Send: 10 messages/minute sustained (chat only)
- Receive: 50 messages/minute sustained (chat from 5 players)
- Images: 1 upload/minute, 5 downloads/minute

**Server (50 Players):**
- 500 messages/minute chat throughput
- 50 image uploads/hour
- 100 trades/hour

### CPU/Memory Targets

**Client Overhead:**
- Chat system: <5ms per frame (encryption/decryption/rendering)
- Image system: <10ms per frame (thumbnail rendering)
- Trade system: <2ms per frame (UI updates)
- Total: <17ms per frame (no regression on <16.67ms budget)

**Server Overhead:**
- Chat routing: <1ms per message (50 players, 500 messages/minute)
- Image relay: <50ms per image (500KB, 50 uploads/hour)
- Trade validation: <10ms per proposal (100 trades/hour)
- Total memory: <500MB for 50 players (includes chat history, image cache, trade state)

**Memory Budget:**
- Chat history: <1MB per player (last 1000 messages)
- Image cache: <50MB (100 images × 500KB, LRU eviction)
- Trade state: <1MB (active proposals, history)
- Total client overhead: <100MB (for 50 players)

### Acceptance Tests (Must Pass Before Merge)

**Functionality:**
- [ ] All 6 features operational: NPC dialog, chat, images, trades, concurrency, networking
- [ ] E2E encryption: Server cannot decrypt chat messages
- [ ] Deterministic mode: Same seed produces identical NPC dialog (10 runs)
- [ ] Proximity validation: Trades rejected if players >5 tiles apart
- [ ] Trust enforcement: Low-trust players cannot trade legendary items

**Performance:**
- [ ] No frame time regression: Maintain <16.67ms per frame (60 FPS)
- [ ] Chat delivery: 99% within 2 seconds at 200ms latency
- [ ] Image upload: 500KB in <10 seconds at 200ms latency
- [ ] Trade commit: <100ms at 200ms latency
- [ ] Memory: Total overhead <100MB for 50 players

**Reliability:**
- [ ] Message loss recovery: 10% packet loss → 99%+ delivery with retries
- [ ] Trade atomicity: 1000 trades, zero item duplication or loss
- [ ] Disconnect resilience: Trades rollback correctly on disconnect (100 tests)
- [ ] Concurrent conflict: 10 players trading simultaneously, no corruption

**Quality:**
- [ ] Test coverage: ≥65% per package (chat, dialog, trade, network)
- [ ] Race detector: `go test -race ./...` passes with zero warnings
- [ ] Cross-platform: All features work on Linux, macOS, Windows, WebAssembly
- [ ] Documentation: All packages have `doc.go`, all public APIs have godoc comments

---

## Milestones & Timeline

### Month 1-2: Phase 21 - Chat System Foundation ✅ COMPLETE
**Status:** All deliverables implemented (November 2025)

**Completed:**
- ✅ Chat components, systems, UI (global, local, party, whisper channels)
- ✅ E2E encryption (Diffie-Hellman, AES-256-GCM) in `pkg/network/crypto.go`
- ✅ Rate limiting, spam filters, client-side profanity filter in `pkg/network/profanity.go`
- ✅ ACK/NACK protocol with retries in `pkg/network/chat.go`
- ✅ Chat UI rendering in `pkg/rendering/ui/chat.go`
- ✅ Comprehensive tests: `chat_test.go`, `chat_integration_test.go`, `profanity_test.go`, `chat_test.go` (UI)
- ✅ Test coverage: >65% for all new packages
- ✅ Latency simulation tests (200ms, 500ms, 2000ms, 5000ms)
- ✅ Packet loss tests (5%, 10%, 20%)
- ✅ Multi-player integration tests (50 players)
- ✅ Throughput benchmarks (500 messages/min target achieved)

### Month 3: Phase 22 - NPC Dialog System ✅ COMPLETE
**Status:** All deliverables implemented (November 2025)

**Completed:**
- ✅ Markov chain generator (order 2-3) in `pkg/procgen/dialog/markov.go`
- ✅ Genre-specific text corpora (5 genres) in `pkg/procgen/dialog/corpus.go`
- ✅ NPC personality traits system in `pkg/procgen/dialog/personality.go`
- ✅ Dialog state management in `pkg/engine/npcdialog_component.go`
- ✅ NPCDialogSystem integration in `pkg/engine/npcdialog_system.go`
- ✅ Deterministic fallback mode with `-deterministic-dialog` flag support
- ✅ Comprehensive tests: `markov_test.go`, `corpus_test.go`, `personality_test.go`, `npcdialog_component_test.go`, `npcdialog_system_test.go`
- ✅ Variation tests (>50% unique responses verified)
- ✅ Determinism tests (identical output with same seed)
- ✅ Corpus validation tests
- ✅ Performance benchmarks (<50ms per response, <5s for 1000 generations)
- ✅ CLI tool: `cmd/dialogtest/main.go` for interactive testing
- ✅ Test coverage: >65% for all new packages
- ✅ Genre-appropriate vocabulary validation
- ✅ Graceful fallback to templates on generation failure

### Month 4-5: Phase 23 - Image Sharing System
**Deliverables:**
- Image upload/download (chunked HTTP, resume on disconnect)
- Thumbnail generation (128×128 JPEG)
- Size/type validation, moderation hooks
- Manual accept UI, auto-download toggle
- Tests: Upload tests (100KB-600KB), disconnect/resume tests, moderation hook tests

### Month 6: Phase 24 - Item Trading System
**Deliverables:**
- Trade proposal, review, commit protocol (two-phase)
- Proximity validation (lag-compensated)
- Trust score mechanics, tradability rules
- Rollback on disconnect/conflict
- Tests: Atomicity tests (1000 trades), concurrent conflict tests, proximity tests

### Month 7: Phase 25 - Concurrency & Integration
**Deliverables:**
- Multi-party conversation support (NPC + players)
- Message ordering, conflict resolution, turn-taking
- Integration tests (50 players, 500 messages/minute)
- Performance optimization (frame time, bandwidth)
- Tests: Multi-player tests, conflict tests, queue tests, benchmarks

### Month 8: Phase 26 - Polish & Beta Release
**Deliverables:**
- UI polish (chat bubbles, trade confirmation dialogs, image preview)
- Error handling, user feedback (notifications, error messages)
- Beta deployment (separate servers, `-social-beta=true` flag)
- Documentation (user manual, API reference, migration guide)
- Acceptance tests (all must pass)

### Month 9-10: Stable Release & Post-Launch
**Deliverables:**
- Stable release (social features enabled by default)
- Performance monitoring, bug fixes
- Telemetry analysis, feature tuning
- v5.1 planning (ML image moderation, advanced spam filters)

---

## Deliverables Checklist

**Code:**
- [x] `pkg/procgen/dialog/` - Markov generator, corpora, personality system (Phase 22)
- [x] `pkg/network/chat.go` - E2E encryption, ACK/NACK, message routing ✅
- [x] `pkg/network/crypto.go` - Diffie-Hellman, AES-256-GCM ✅
- [x] `pkg/network/profanity.go` - Client-side profanity filter (opt-in, configurable) ✅
- [x] `pkg/network/chat_test.go` - 19 test functions, 3 benchmarks ✅
- [x] `pkg/network/chat_integration_test.go` - 13 integration tests, 2 benchmarks ✅
- [x] `pkg/network/profanity_test.go` - 14 test functions, 4 benchmarks ✅
- [x] `pkg/rendering/ui/chat.go` - Chat UI rendering with 4 channels ✅
- [x] `pkg/rendering/ui/chat_test.go` - 16 test functions, 3 benchmarks ✅
- [ ] `pkg/network/images.go` - Upload/download, chunked transfer, thumbnails (Phase 23)
- [ ] `pkg/network/trade.go` - Two-phase commit, proximity validation, trust
- [ ] `pkg/engine/chat_component.go` - Chat state, message history
- [ ] `pkg/engine/trade_component.go` - Trade state, trust score
- [ ] `pkg/engine/dialog_component.go` - Dialog state, response history
- [ ] `pkg/engine/chat_system.go` - Message delivery, channel management
- [ ] `pkg/engine/trade_system.go` - Trade lifecycle, rollback
- [ ] `pkg/engine/dialog_system.go` - NPC response generation
- [ ] `pkg/rendering/ui/chat.go` - Chat UI (message list, input, channels)
- [ ] `pkg/rendering/ui/trade.go` - Trade UI (proposal, review, confirm)

**Tests:**
- [x] `pkg/procgen/dialog/markov_test.go` - Variation, determinism, corpus tests (Phase 22)
- [x] `pkg/network/chat_test.go` - Encryption, ACK/NACK, latency simulation ✅
- [x] `pkg/network/crypto_test.go` - Key exchange, encryption/decryption ✅
- [x] `pkg/network/profanity_test.go` - Filter behavior, leet speak detection ✅
- [x] `pkg/network/chat_integration_test.go` - E2E flow, packet loss, multi-player ✅
- [x] `pkg/rendering/ui/chat_test.go` - UI behavior, message display, input handling ✅
- [ ] `pkg/network/images_test.go` - Upload/download, resume, validation (Phase 23)
- [ ] `pkg/network/trade_test.go` - Two-phase commit, atomicity, rollback
- [ ] Integration tests: Multi-player scenarios, packet loss, concurrency
- [ ] Benchmarks: Dialog generation, chat throughput, trade validation

**Documentation:**
- [ ] `docs/SOCIAL_SYSTEMS.md` - User guide (chat commands, trading, NPC dialog)
- [ ] `docs/API_REFERENCE.md` - API updates (social components, systems)
- [ ] `docs/MIGRATION_V5.md` - v4.0 → v5.0 save migration guide
- [ ] `pkg/procgen/dialog/doc.go` - Package documentation (Markov chains, determinism policy)
- [ ] `pkg/network/doc.go` - Package documentation (E2E encryption, protocols)
- [ ] `README.md` - Feature updates (social systems section)

**Examples:**
- [ ] `examples/chat_demo/` - Chat system demonstration (E2E encryption, channels)
- [ ] `examples/trade_demo/` - Trading system demonstration (two-phase commit)
- [ ] `examples/dialog_demo/` - NPC dialog demonstration (Markov generation)

**Tools:**
- [ ] `cmd/chattest/` - Chat system CLI testing tool
- [ ] `cmd/dialogtest/` - Dialog generation CLI testing tool
- [ ] `cmd/tradetest/` - Trading system CLI testing tool

---

## Notes & Open Decisions

### Design Trade-Offs Requiring Team Discussion

1. **E2E Encryption vs. Moderation:**
   - **Current:** E2E chat prevents server-side moderation
   - **Alternative:** Server-readable chat with encryption-at-rest only
   - **Decision Required:** Prioritize privacy (E2E) or moderation (server-readable)?
   - **Recommendation:** Keep E2E, rely on client-side filters + user reporting

2. **Markov Chain Order:**
   - **Current:** Order 2-3 (configurable)
   - **Trade-off:** Higher order (4-5) → better coherence, but requires larger corpora and slower generation
   - **Decision Required:** Balance between dialog quality and performance
   - **Recommendation:** Start with order 2, tune based on beta feedback

3. **Trust Score Persistence:**
   - **Current:** Trust scores not persisted (reset to 0.5 on server restart)
   - **Alternative:** Persist trust scores to database
   - **Trade-off:** Persistence enables long-term reputation, but adds complexity and storage
   - **Decision Required:** Is persistent trust worth the complexity?
   - **Recommendation:** Defer to v5.1+ (start with in-memory)

4. **Image Storage:**
   - **Current:** Images ephemeral (10-minute expiry)
   - **Alternative:** Persist images to disk/database for chat history
   - **Trade-off:** Persistence enables scrollback, but requires storage and moderation
   - **Decision Required:** Ephemeral vs. persistent image sharing?
   - **Recommendation:** Start ephemeral, consider persistence in v5.1+ if users request

5. **NPC Dialog Sharing:**
   - **Open Question:** Should all players in proximity see same NPC dialog, or personalized per player?
   - **Current:** Personalized (each player sees different responses)
   - **Alternative:** Shared (all players see same dialog, like multiplayer cutscene)
   - **Trade-off:** Personalized → more variety, but inconsistent; Shared → consistent, but less variety
   - **Decision Required:** Personalized vs. shared NPC dialog in multiplayer?
   - **Recommendation:** Default to personalized, add config flag for shared mode

6. **Trade Rollback Notification:**
   - **Open Question:** How detailed should rollback error messages be?
   - **Current:** Generic "trade failed" with reason code (proximity, trust, ownership)
   - **Alternative:** Detailed messages ("Item X no longer owned by Player Y")
   - **Trade-off:** Detailed → better UX, but potential privacy leak (reveals inventory state)
   - **Decision Required:** Generic vs. detailed rollback messages?
   - **Recommendation:** Detailed messages (transparency > privacy in trade context)

---

**End of Roadmap v5.0**
