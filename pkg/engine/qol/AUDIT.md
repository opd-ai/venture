# Audit: github.com/opd-ai/venture/pkg/engine/qol
**Date**: 2026-02-13
**Status**: Needs Work

## Summary
The QoL package implements quality-of-life features (auto-loot, craft queue, guild invitations, mount whistle, storage sorting, recipe tracking) with solid architecture and excellent test coverage (94.6%). The package is well-structured with clean separation of concerns and proper thread safety. Critical issues include missing godoc comments on 31 exported methods, absence of structured logging, and lack of serialization support for the persistent QoLComponent.

## Issues Found
- [ ] **med** doc-coverage — 31 exported methods missing godoc comments across all manager types (`autoloot.go:24,31,42,54`, `craftqueue.go:27,56,74,90`, `guildinvitation.go:29,49,70,93`, `mountwhistle.go:27,43,50,59`, `recipetracker.go:25,59,77,87`, `storagesorter.go:52,59,76`, `system.go:47,54,61,68,75,82,89,96`)
- [ ] **med** error-handling — No structured logging with logrus.WithFields on error paths; package lacks any logging infrastructure (all files)
- [ ] **med** integration — QoLComponent missing Serialize/Deserialize methods for persistence support; other persistent components (AchievementComponent, etc.) implement serialization (`types.go:111`)
- [ ] **low** deterministic-procgen — Uses time.Now() for timestamps, but this is acceptable as QoL features are time-based mechanics, not procedural generation (`craftqueue.go:48`, `guildinvitation.go:34,37,88`, `mountwhistle.go:36`, `types.go:142`)

## Test Coverage
94.6% (target: 65%) ✅

## Integration Status
**Well-integrated into engine architecture:**
- QoLSystemWrapper in `pkg/engine/qol_system_wrapper.go` implements System interface with periodic cleanup
- Manager initialized in client handlers (`cmd/client/handlers.go:1116`)
- Integration tests verify wrapper behavior (`pkg/engine/qol_system_wrapper_test.go`)

**Missing integrations:**
- No Serialize/Deserialize support means QoL preferences not persisted across sessions
- No structured logging means difficult to debug QoL issues in production
- No registration evidence in `system_init.go` or server code (client-only feature currently)

**Cross-system connections:**
- V4 Companions: Auto-loot queries manager for collection decisions
- V8 Crafting: Craft queue integrates with crafting system
- V8 Guilds: Offline invitation system with 7-day expiry
- V4 Vehicles: Mount whistle summons vehicles via pathfinding

## Recommendations
1. **Add godoc comments** to all 31 exported methods following Go conventions (e.g., "SetConfig sets auto-loot configuration for a companion")
2. **Implement structured logging** with logrus.WithFields for error paths (especially in AddRecipe errors, AcceptInvitation failures, CleanupExpired results)
3. **Add Serialize/Deserialize** to QoLComponent to persist player QoL preferences (auto-loot radius, sort presets, tracked recipes) across sessions
4. **Add integration logging** in QoLSystemWrapper.Update for cleanup operations and potential issues
5. Consider adding benchmark tests for SortItems to validate <10ms claim for 100 items
