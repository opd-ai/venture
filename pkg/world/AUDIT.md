# Audit: pkg/world/ (Core World Package)

**Date:** 2026-02-16
**Coverage:** 88.8%
**Files Audited:** 9 source files, 9 test files

## Issues Found & Fixed

### HIGH-1: Multi-goal event progress tracking logic bug (metagame.go)
- **Severity:** High
- **Status:** ✅ Fixed
- **Description:** `checkDefaultGoals()` and `allGoalsMet()` compared a single participant progress int against all goal values. For multi-goal events (e.g., seasonal challenges with `quests_completed: 50` and `bosses_defeated: 5`), a participant with progress 50 would pass both checks even if only one goal was actually met.
- **Fix:** Changed `UpdateProgress()` to use compound keys (`serverID:goalKey`) for per-participant per-goal tracking. Rewrote `checkDefaultGoals()` to check each goal independently per participant. Removed the flawed `allGoalsMet()` function.

### HIGH-2: Potential panic in ListBackups sort (persistence.go)
- **Severity:** High
- **Status:** ✅ Fixed
- **Description:** `ListBackups()` sort function called `os.Stat()` and ignored errors. If a file was deleted between the existence check and sort, the nil `FileInfo` would panic on `ModTime()`.
- **Fix:** Added error handling in sort comparator; files that can't be stat'd sort last.

### MED-1: Missing input validation for negative counts (territory.go)
- **Severity:** Medium
- **Status:** ✅ Fixed
- **Description:** `UpdateControlPoint()` accepted negative `attackers` and `defenders` values, which would produce nonsensical capture/decay calculations.
- **Fix:** Added validation returning error for negative values.

### MED-2: Empty row access in compression (chunk_compression.go)
- **Severity:** Medium
- **Status:** ✅ Fixed
- **Description:** `CompressChunk()` accessed `chunk.Terrain[0]` to get width without checking if the first row was empty, risking index-out-of-range panic.
- **Fix:** Added empty column check before accessing `Terrain[0]`.

## Remaining Issues (Low Priority)

### LOW-1: Silent out-of-bounds writes (state.go)
- `SetTile()` silently ignores out-of-bounds writes with no error return. This is consistent with `GetTile()` returning nil, but callers have no way to know a write failed.

### LOW-2: Chunk loader calls LoadWorld per chunk (chunk_loader.go)
- `loadChunk()` calls `c.persistence.LoadWorld()` for every chunk load attempt. Should cache the loaded state to avoid repeated disk reads.

### LOW-3: Expired events never removed from map (metagame.go)
- `EventManager.events` map grows indefinitely. Expired events are marked inactive but never deleted. Could cause memory growth over long server sessions.

### LOW-4: No thread safety for EventManager (metagame.go)
- `EventManager` has no mutex protection. If used from multiple goroutines (e.g., game server), concurrent map access could panic.

### LOW-5: No thread safety for TerritoryManager (territory.go)
- `TerritoryManager.zones` map has no synchronization. Concurrent combat updates could cause data races.
