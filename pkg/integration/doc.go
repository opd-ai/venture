// Package integration provides cross-system production integrations and
// multiplayer integration tests for Venture.
//
// # Sub-Packages (Production Systems)
//
// Nine sub-packages implement cross-domain features that bridge engine,
// world, network, procgen, and narrative systems:
//
//   - choice_consequences — Tracks narrative choices and applies consequences
//   - companion_housing   — Companion/pet home systems with bedding and training
//   - guild_housing       — Guild housing with permissions and federation sync
//   - guild_vehicle       — Guild fleet management with vehicle deployment
//   - housing_crafting    — Housing + crafting station integration
//   - narrative_world     — Narrative events with companion memory tracking
//   - political_warfare   — Guild wars, treaties, embargoes, and diplomacy
//   - trade_routes        — Trade route management with economy integration
//   - world_events        — World event scheduling and processing
//
// # System Registration Patterns
//
// Integration systems follow three registration patterns. Choose based on
// whether the system needs per-frame updates:
//
// # ECS System Pattern
//
// The system wraps a Manager and is registered with the engine World for
// per-frame Update() calls. Used by narrative_world, political_warfare, and
// world_events packages.
//
// Example:
//
//	// Initialize guild manager (required dependency)
//	guildMgr := engine.NewGuildSystem(world)
//
//	// Create political warfare system with seed for deterministic events
//	warSys := political_warfare.NewSystem(world, guildMgr, 12345)
//
//	// Register with engine for per-frame updates
//	world.AddSystem(warSys)
//
//	// System will now receive Update(entities, deltaTime) calls each frame
//	// and process war mechanics, treaty timers, siege progress, etc.
//
// # Manager-Only Pattern
//
// A Manager is created and stored in the client's systemsContainer. It is
// called directly from handlers rather than through the ECS update loop.
// Used by guild_housing, trade_routes, and guild_vehicle packages.
//
// Example:
//
//	// Create housing manager with world reference
//	housingMgr := guild_housing.NewManager()
//
//	// Store in systems container for handler access
//	container.guildHousingManager = housingMgr
//
//	// Call directly from handlers (not through ECS loop)
//	err := housingMgr.PurchaseRoom(guildID, "barracks")
//	if err != nil {
//	    logrus.WithError(err).Error("failed to purchase room")
//	}
//
//	// For networked updates
//	housingMgr.ProcessRoomUpdate(guildID, roomData)
//
// # Pure Integration Pattern
//
// No registration needed. Functions are called directly from other systems
// that already participate in the ECS loop. Used by choice_consequences,
// companion_housing, and housing_crafting packages.
//
// Example:
//
//	// In your quest system's Update() method:
//	func (s *QuestSystem) Update(entities []*engine.Entity, deltaTime float64) {
//	    for _, entity := range entities {
//	        quest, ok := entity.GetComponent("quest").(*QuestComponent)
//	        if !ok || !quest.JustCompleted {
//	            continue
//	        }
//
//	        // Apply narrative consequences directly
//	        choice_consequences.RecordChoice(entity.ID, quest.ChoiceID)
//	        choice_consequences.ApplyConsequences(entity.ID, world)
//
//	        // Update companion memory if player has companion
//	        if entity.HasComponent("companion") {
//	            narrative_world.UpdateCompanionMemory(entity.ID, quest.Outcome)
//	        }
//	    }
//	}
//
// # Determinism
//
// All sub-packages use injectable TimeProvider for deterministic timestamps.
// Set a fake provider in tests to get reproducible results:
//
//	pkg.SetTimeProvider(func() time.Time { return fixedTime })
//	defer pkg.ResetTimeProvider()
//
// # Testing
//
// Several sub-packages transitively import Ebiten via pkg/world/housing or
// pkg/engine, causing GLFW init panics in headless environments. Use the
// test-integration script or set DISPLAY=:99:
//
//	make test-integration            # recommended
//	DISPLAY=:99 go test ./pkg/integration/...  # manual
//
// The root package also contains multiplayer_test.go with integration tests
// for deterministic content generation, cross-genre consistency, save/load
// compatibility, and network latency simulation.
package integration
