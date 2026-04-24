// Package companion is the top-level namespace for companion-system packages
// that live outside the main engine package.
//
// # Namespace Map
//
// The companion feature surface is spread across three locations:
//
//   - pkg/engine/companion_*.go — core ECS systems and components wired into the
//     game world (CompanionProgressionSystem, CompanionHousingSystem, etc.).
//     These live in pkg/engine because they operate on engine.Entity and engine.World
//     and must avoid import cycles with pkg/engine.
//
//   - pkg/procgen/companion/ — procedural generation of companion personalities,
//     skill trees, dialogue, and equipment.  Depends only on pkg/procgen primitives.
//
//   - pkg/companion/learning/ (this package) — adaptive-AI learning subsystem that
//     adjusts companion behaviour based on player interaction history.  Kept here to
//     isolate the ML-style logic from the main engine loop.
//
// # Usage
//
// New companion-related packages that do not require direct access to engine.World
// should be placed under pkg/companion/. Packages that must register ECS systems or
// components belong in pkg/engine.
//
// See also: pkg/integration/companion_housing/ for cross-system integration between
// companion AI and the player-housing subsystem.
package companion
