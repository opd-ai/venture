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
// ECS System Pattern (narrative_world, political_warfare, world_events):
// The system wraps a Manager and is registered with the engine World for
// per-frame Update() calls.
//
//	sys := political_warfare.NewSystem(world, guildMgr, seed)
//	world.AddSystem(sys)
//
// Manager-Only Pattern (guild_housing, trade_routes, guild_vehicle):
// A Manager is created and stored in the client's systemsContainer. It is
// called directly from handlers rather than through the ECS update loop.
//
//	mgr := guild_housing.NewManager()
//	container.guildHousingManager = mgr
//
// Pure Integration Pattern (choice_consequences, companion_housing, housing_crafting):
// No registration needed. Functions are called directly from other systems
// that already participate in the ECS loop.
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
