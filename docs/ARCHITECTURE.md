# Architecture Decision Records

## ADR-001: Entity-Component-System (ECS) Architecture

**Status:** Accepted

**Context:**
The game requires a flexible architecture to handle diverse procedurally-generated content including entities, items, abilities, and behaviors. Traditional object-oriented hierarchies would become unwieldy with the variety of possible combinations.

**Decision:**
Implement an Entity-Component-System (ECS) architecture where:
- **Entities** are unique identifiers with component collections
- **Components** are pure data structures (position, health, sprite, etc.)
- **Systems** contain behavior logic and operate on entities with specific components

**Consequences:**
- **Positive:** Excellent composition flexibility, cache-friendly data access, easy to add new content types
- **Positive:** Parallel system execution potential
- **Negative:** More verbose than traditional OOP for simple cases
- **Negative:** Requires discipline to avoid putting logic in components

## ADR-002: Pure Go with No External Assets

**Status:** Accepted

**Context:**
Requirements specify 100% procedural generation for graphics and audio, with no external asset files.

**Decision:**
All visual and audio content will be generated at runtime using:
- Procedural graphics via Ebiten's image manipulation
- Waveform synthesis for audio
- Algorithmic generation seeded for determinism

**Consequences:**
- **Positive:** Single binary distribution, no asset pipeline
- **Positive:** Infinite content variety within generation rules
- **Negative:** Higher CPU/memory usage for generation
- **Negative:** Initial generation time on startup
- **Mitigation:** Cache generated assets, lazy generation, progressive loading

## ADR-003: Client-Server Network Architecture

**Status:** Accepted

**Context:**
Multiplayer support required with high-latency tolerance (200-5000ms) for co-op gameplay, including support for slow connections like onion services (Tor).

**Decision:**
Implement authoritative server model with client-side prediction:
- Server maintains canonical game state
- Clients predict local actions for responsiveness
- Server reconciliation corrects prediction errors
- Entity interpolation smooths network jitter

**Consequences:**
- **Positive:** Prevents cheating, consistent game state
- **Positive:** Works well with high latency through prediction
- **Negative:** More complex than peer-to-peer
- **Negative:** Requires dedicated server for multiplayer

## ADR-004: Package-Based Module Organization

**Status:** Accepted

**Context:**
Large codebase needs clear organization and separation of concerns.

**Decision:**
Use `pkg/` directory with domain-focused packages:
- `engine/` - Core ECS and game loop
- `procgen/` - All generation systems
- `rendering/` - Visual output
- `audio/` - Sound synthesis
- `network/` - Multiplayer
- `combat/` - Combat mechanics
- `world/` - World state

**Consequences:**
- **Positive:** Clear responsibility boundaries
- **Positive:** Easier testing and reusability
- **Positive:** Supports parallel development
- **Negative:** Requires careful interface design to avoid circular dependencies

## ADR-005: Deterministic Generation with Seeds

**Status:** Accepted

**Context:**
Need reproducible content for multiplayer sync and testing.

**Decision:**
All procedural generation uses deterministic algorithms with seed values:
- Base world seed derives all other seeds
- Each content type gets independent but deterministic sub-seeds
- Same seed always produces same content

**Consequences:**
- **Positive:** Multiplayer clients can generate same content independently
- **Positive:** Easy to reproduce bugs and test scenarios
- **Positive:** Share interesting worlds via seed sharing
- **Negative:** Must avoid using system time or other non-deterministic sources

## ADR-006: Genre System for Content Variation

**Status:** Accepted

**Context:**
Game should support multiple adventure genres (fantasy, sci-fi, horror, etc.) with appropriate theming.

**Decision:**
Implement genre as a modifier system that affects:
- Visual palette and style
- Audio themes and instruments
- Entity naming and types
- Item and ability flavoring
- Environment themes

**Consequences:**
- **Positive:** Huge variety from same generation systems
- **Positive:** Player can choose preferred setting
- **Negative:** Requires careful abstraction of generation rules
- **Negative:** Need to validate all genre combinations work

## ADR-007: Performance Targets

**Status:** Accepted

**Context:**
Game should run on modest hardware (Intel i5/Ryzen 5, 8GB RAM, integrated graphics) to be accessible to a wide audience.

**Decision:**
Target specifications:
- 60 FPS minimum framerate
- <500MB client memory
- <1GB server memory (4 players)
- <2s world generation time
- <100KB/s per player network usage

**Consequences:**
- **Positive:** Accessible to wider audience
- **Positive:** Forces efficient algorithms
- **Negative:** May limit graphical complexity
- **Negative:** Requires careful optimization

**Achieved (V3.0):** 106 FPS with 2000 entities, 73MB memory usage (70% above FPS target, 86% below memory budget)

## ADR-008: Enhanced Visual Quality (V3.0)

**Status:** Accepted

**Context:**
V2.0 established functional procedural generation for all visual content, but visual quality improvements were needed to rival hand-crafted games while maintaining zero external assets and deterministic generation.

**Decision:**
Implement comprehensive visual enhancement system across six phases (15-20):
- **Phase 15:** Enhanced sprite generation with anatomical accuracy, anti-aliasing, and genre variations
- **Phase 16:** Advanced tile rendering with texture patterns, smooth transitions, and depth effects
- **Phase 17:** Sophisticated lighting with soft shadows, colored lights, and bloom effects
- **Phase 18:** Rich particle and weather systems with fluid simulation
- **Phase 19:** UI polish with dynamic color palettes and visual hierarchy
- **Phase 20:** Environmental detail with parallax backgrounds, time-of-day, and post-processing

**Implementation:**
- **Sprite Detail:** 40% increase through pixel-perfect anatomical templates (head 4×4, torso 4×6, legs 4×8)
- **Anti-Aliasing:** Sub-pixel rendering with 4 quality levels (Off, Low, Medium, High) using 2x2 to 8x8 super-sampling
- **Lighting:** Soft shadows with gradient edges, colored lighting, bloom effects, advanced ambient occlusion
- **Weather:** Comprehensive weather systems (rain, snow, fog, dust, ash) with genre-specific variations
- **Textures:** 50+ procedural patterns per genre (stone, wood, metal, organic)
- **Performance:** Maintained 106 FPS with all enhancements active (<5% lighting overhead)

**Consequences:**
- **Positive:** Visual quality rivals hand-crafted games while remaining 100% procedural
- **Positive:** 40% sprite detail increase, silhouette scores improved from 0.65 to 0.75 average
- **Positive:** All enhancements maintain deterministic seed-based generation
- **Positive:** Performance targets maintained (106 FPS, 73MB memory)
- **Positive:** Full backward compatibility with V2.0 saves
- **Negative:** Sprite generation time increased from 2-3ms to 3-5ms (still within <5ms target)
- **Mitigation:** Sprite caching maintained at 95.9% hit rate, offsetting generation cost

---

## ADR-007: V4.0 Gameplay Expansion Systems

**Status:** Accepted

**Context:**
V3.0 delivered polished visual and audio systems. V4.0 expands gameplay depth with ten major systems (Phases 21-30) that add procedural vehicles, companions, books, advanced magic, character classes, social features, mini-games, reputation mechanics, adaptive music, and environmental storytelling.

**Decision:**
Implement comprehensive gameplay expansion across ten phases (21-30):
- **Phase 21:** Vehicle System - Procedural vehicles with physics, fuel, durability, and combat
- **Phase 22:** Companion System - AI companions with loyalty, inventory, progression, and skill inheritance
- **Phase 23:** Book System - Procedural books with genre-specific content and reading mechanics
- **Phase 24:** Expanded Magic System - Spell effects and combination system for diverse spellcasting
- **Phase 25:** Class Progression - Character classes with specializations and unique abilities
- **Phase 26:** Expression System - Player emotes, gestures, and achievement tracking
- **Phase 27:** Mini-Game System - Embedded games (puzzles, dice, card games) within the world
- **Phase 28:** Reputation System - Faction reputation with moral choices and alignment tracking
- **Phase 29:** Adaptive Music System - Context-aware music that responds to gameplay events
- **Phase 30:** Environmental Storytelling - Discoverable story fragments, branching narratives, and procedural archaeology

**Implementation:**
- **System Count:** 65 active systems in client (63 via registerAllSystems + 2 in main.go)
- **Performance:** All V4 systems combined add <24% FPS impact (106 → ~80 FPS, well above 60 FPS minimum)
- **Memory:** +48MB for V4 features (73MB → 121MB, <500MB budget with 379MB margin)
- **Network:** +63KB/s per player (<100KB/s target)
- **Coverage:** All V4 packages maintain ≥65% test coverage
- **Determinism:** Vehicle/companion/book generation fully deterministic (seed-based), with controlled non-determinism for NPC dialog in V5.0

**Key Systems:**
1. **VehicleMovementSystem:** Physics-based vehicle movement with fuel consumption and terrain interaction
2. **VehicleDurabilitySystem:** Damage tracking and repair mechanics
3. **MountingSystem:** Player/NPC mounting and dismounting logic
4. **VehicleCombatSystem:** Vehicle-mounted weapons and combat
5. **CompanionAISystem:** Companion pathfinding, commands, and behavior
6. **CompanionProgressionSystem:** Level-up and stat advancement for companions
7. **CompanionLoyaltySystem:** Loyalty mechanics affecting companion behavior
8. **CompanionInventorySystem:** Companion-specific inventory management
9. **SkillInheritanceSystem:** Skill transfer from companion to player
10. **BookReadingSystem:** Interactive book reading with progress tracking
11. **SpellEffectSystem:** Visual and gameplay effects for spells
12. **SpellCombinationSystem:** Combining spells for enhanced effects
13. **ClassProgressionSystem:** Class-based abilities and progression paths
14. **ExpressionSystem:** Player emote execution and display
15. **MiniGameSystem:** Embedded game state management
16. **ReputationSystem:** Faction standing and reputation tracking
17. **MusicTriggerSystem:** Context-aware music transitions
18. **DiscoverySystem:** Story fragment discovery and journal tracking

**Multiplayer Integration:**
- **V4 Components Synchronized:** VehicleComponent, CompanionComponent, MountComponent, AchievementComponent, ExpressionComponent, ClassProgressionComponent
- **Serialization:** All synced components implement Serialize/Deserialize methods (17-29 bytes per component)
- **Server Entity Spawning:** Server independently spawns vehicles, companions, and bookshelves for authoritative state
- **Snapshot Extensions:** EntitySnapshot struct extended to include V4 component data
- **Network Overhead:** Measured at 63KB/s per player (37KB/s margin under 100KB/s budget)

**Performance Benchmarks:**
- **VehicleCombatSystem:** 64.7µs per update (100 vehicles) = 15,445 ops/sec
- **CompanionAISystem:** 1.88µs per update (50 companions) = 532,340 ops/sec  
- **CompanionProgressionSystem:** 3.55µs per update (100 companions) = 281,768 ops/sec
- **MiniGameSystem:** 0.22µs per update (20 games) = 4,595,018 ops/sec
- **ExpressionSystem:** 0.89µs per update (100 expressions) = 1,121,076 ops/sec
- **V4 Integrated Scenario:** 3.35µs per frame (30 vehicles + 10 companions + 5 games + 5 expressions) = 4,970x faster than 16.67ms budget

**Consequences:**
- **Positive:** Transforms Venture from action-RPG into deep gameplay experience with diverse systems
- **Positive:** All V4 features maintain deterministic seed-based generation (except dialog in V5.0)
- **Positive:** Performance targets exceeded (80+ FPS vs 60 FPS minimum, 4,970x headroom)
- **Positive:** Multiplayer-compatible design with full network synchronization
- **Positive:** Backward compatible with V3.0 saves (default values for new components)
- **Positive:** 100% procedural content across all new features (zero external assets)
- **Negative:** System complexity increased (41 → 65 systems in client)
- **Mitigation:** Comprehensive testing (≥65% coverage), clear system boundaries, extensive documentation
- **Negative:** Memory footprint increased (+48MB for V4 features)
- **Mitigation:** Object pooling, sprite caching, lazy initialization where appropriate

---

## Related Documentation

For implementation details, development workflows, testing strategies, and code quality standards, see:
- **[Development Guide](DEVELOPMENT.md)** - Complete development workflow and best practices
- **[Contributing Guide](CONTRIBUTING.md)** - Contribution guidelines and code standards
- **[Technical Specification](TECHNICAL_SPEC.md)** - Detailed technical architecture

---

## ADR-008: V8.0 Player Housing System

**Status:** Accepted (Phase 49.1 Complete)

**Context:**
V7.0 completed advanced visual improvements (display scaling, enhanced sprites, smoothed walls, pixel-perfect collision). V8.0 begins with player housing system as foundation for persistent structures, guilds, and territory control.

**Decision:**
Implement player housing core infrastructure (Phase 49.1) with:
- **Package:** `pkg/world/housing/` for plot management, collision detection, persistence
- **Building Sizes:** Four tiers (Small 8×8, Medium 16×16, Large 24×24, Estate 32×32 tiles)
- **Spatial Indexing:** Uniform grid with 64-unit cells for O(1) plot queries
- **Permissions:** Four levels (None, Visit, Friend, CoOwner) with per-player overrides
- **Persistence:** JSON + gzip compression (8x compression ratio achieved)
- **ECS Integration:** HousingComponent for entity-based housing representation

**Implementation:**
- **Manager:** Central plot placement and validation system with spatial grid
- **Plot:** Core type with ownership, position, size, permissions, metadata
- **SpatialGrid:** Efficient collision detection using uniform grid partitioning
- **PermissionSet:** Access control with default and per-player permissions
- **Persistence:** Save/load with incremental updates and player-specific export
- **CLI Tool:** `cmd/housingtest/` for demonstration and testing

**Performance Metrics (Phase 49.1):**
- Plot allocation: <1ms per placement (50x better than <50ms target)
- Collision detection: <0.1ms for 1000 plots (100x better than <10ms target)
- Save: ~25ms for 50 plots (<100ms target)
- Load: ~15ms for 50 plots (<200ms target)
- Memory: ~25KB per 100 plots (200x better than <5MB target)
- Compression: 8x ratio (1.23KB for 50 plots)
- Test Coverage: 95.1% (exceeds 65% requirement)

**Consequences:**
- **Positive:** Foundation for guilds, territory control, cross-server structures (Phases 49.2-51)
- **Positive:** Spatial grid enables efficient collision detection at scale (1000+ plots)
- **Positive:** Compression keeps save files small (<2KB per 100 plots vs 50MB budget)
- **Positive:** Permission system supports co-ownership and visitor access
- **Positive:** Deterministic plot IDs for multiplayer synchronization
- **Negative:** Plot placement limited to non-overlapping areas (by design for gameplay)
- **Mitigation:** 1-tile minimum spacing enforced, clear error messages for invalid placement
- **Future:** Building generation (Phase 51.1), furniture (51.3), guild halls (51.2)

**Package Structure:**
```
pkg/world/housing/
├── doc.go          # Package documentation
├── types.go        # BuildingSize, PermissionLevel, Plot, Vector2
├── component.go    # HousingComponent for ECS
├── manager.go      # Manager with plot placement and queries
├── spatial.go      # SpatialGrid for collision detection
├── persistence.go  # Save/load with JSON + gzip
└── *_test.go       # Comprehensive tests (95.1% coverage)
```

**CLI Integration:**
- Added `-enable-housing` flag to client (default: true)
- Created `cmd/housingtest/` demo tool with actions: demo, save, load, benchmark

**Next Steps (Remaining V8.0 Phases):**
- Phase 49.2: Persistent trust & reputation system
- Phase 49.3: Chat history with delta compression
- Phase 49.4: Persistent image storage & gallery
- Phase 50: Guilds, territory control, enhanced physics
- Phase 51: Building generation, guild halls, furniture
- Phase 52: WebRTC federation, mobile federation
- Phase 53: Deep companion AI, branching narratives, advanced classes
- Phase 54: Server modding, blueprint sharing, content tools

