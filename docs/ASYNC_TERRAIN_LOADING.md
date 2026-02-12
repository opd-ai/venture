# Async Terrain Loading

## Overview

The async terrain loading system prevents UI freezes during world generation by showing a loading screen with progress bar while terrain generates in the background. This is particularly important for large composite terrains that can take 2-8 seconds to generate.

## Architecture

### State Flow

```
AppStateCharacterCreation
    ↓
AppStateLoading (NEW)
    ↓
AppStateGameplay
```

### Components

1. **LoadingUI** (`pkg/engine/loading_ui.go`)
   - Simple progress bar with percentage display
   - Updates every frame based on terrain generation progress
   - Renders on dark background with centered message

2. **AppStateLoading** (`pkg/engine/app_state.go`)
   - New application state between character creation and gameplay
   - Transition rules:
     - FROM: `AppStateCharacterCreation` only
     - TO: `AppStateGameplay` or `AppStateMainMenu` (on error)

3. **AsyncLoader** (`pkg/procgen/terrain/async_loader.go`)
   - Existing component now properly utilized
   - Runs terrain generation in background goroutine
   - Provides thread-safe progress tracking (0.0 to 1.0)

### Initialization Sequence

#### Before (Blocking)
```go
1. setupAllGameSystems()
2. setupWorldTerrain()          ← BLOCKS 2-8s for large terrains
3. spawnWorldEntities()
4. setupPlayerEntity()
5. setupGameUI()
6. runGameLoop()                ← Window appears here
```

#### After (Async)
```go
1. setupAllGameSystems()
2. startAsyncTerrainGeneration() ← Returns immediately
3. setupAsyncTerrainCallbacks()  ← Registers completion handler
4. TransitionTo(AppStateLoading)
5. runGameLoop()                 ← Window appears here with loading screen

--- In Update() loop ---
6. handleLoadingState()          ← Polls progress, updates UI
7. When done:
   - completeWorldInitialization()
   - spawnWorldEntities()
   - setupPlayerEntity()
   - setupGameUI()
   - TransitionTo(AppStateGameplay)
```

## Performance Impact

### Before
- **BSP terrain (25ms)**: Small freeze, barely noticeable
- **Composite terrain (2-8s)**: Significant freeze, appears unresponsive
- **User experience**: No feedback, uncertain if game crashed

### After
- **BSP terrain (25ms)**: Loading screen flashes briefly (1-2 frames)
- **Composite terrain (2-8s)**: Smooth progress bar from 0% to 100%
- **User experience**: Clear feedback, 60 FPS loading screen
- **Time to window**: Immediate (from game start)
- **Time to gameplay**: Same total time, but perceived as faster

## Implementation Details

### EbitenGame Fields
```go
type EbitenGame struct {
    // ...
    LoadingUI            *LoadingUI
    terrainLoader        interface{} // *terrain.AsyncLoader
    terrainLoadComplete  func(*Entity) error
    // ...
}
```

### Loading State Handler
```go
func (g *EbitenGame) handleLoadingState() error {
    loader := g.terrainLoader.(asyncLoader)
    
    // Update progress display
    progress, err := loader.GetProgress()
    g.LoadingUI.SetProgress(progress)
    
    // Check if complete
    if loader.IsDone() {
        g.terrainLoader = nil
        g.StateManager.TransitionTo(AppStateGameplay)
        
        // Run completion callback
        if g.terrainLoadComplete != nil {
            g.terrainLoadComplete(g.PlayerEntity)
        }
    }
    
    return nil
}
```

### Terrain Completion Callback
```go
setupAsyncTerrainCallbacks(game, sys, networkClient, logger, clientLogger)
// Sets callback that:
// 1. Retrieves terrain from loader.Wait()
// 2. Initializes terrain rendering, lighting, collision
// 3. Generates factions
// 4. Spawns entities and effects
// 5. Creates player entity
// 6. Sets up UI systems
// 7. Rebuilds spatial partition
```

## Testing

### Unit Tests
- `loading_state_test.go` - State transition validation
- `loading_ui_test.go` - UI component behavior

### Manual Testing
To test with different terrain sizes:
```bash
# Small terrain (BSP, ~25ms) - loading screen flashes
./venture --seed 12345

# Large terrain (Composite) - loading screen visible for ~2-8s
# (Requires custom terrain generator configuration)
```

### Expected Behavior
1. Window appears immediately after character creation
2. Loading screen shows "Generating world..." message
3. Progress bar fills from 0% to 100%
4. Percentage text updates in real-time
5. Smooth 60 FPS during loading
6. Seamless transition to gameplay when complete

## Error Handling

### Terrain Generation Failure
- Loading screen detects error from `loader.GetProgress()`
- Transitions back to `AppStateMainMenu`
- Logs error with context

### Invalid Loader State
- Validates loader interface on state entry
- Falls back to main menu if loader is nil or invalid type
- Prevents crash from missing terrain data

## Future Improvements

1. **Progressive Loading**
   - Show partially generated terrain chunks as they complete
   - Allow player to start in first room while rest generates

2. **Enhanced Progress Display**
   - Show current generation phase (rooms, corridors, decorations)
   - Display estimated time remaining

3. **Asset Preloading**
   - Use loading time to pre-generate common sprites
   - Warm sprite cache during terrain generation

4. **Cancellation Support**
   - Add "Cancel" button to return to menu
   - Gracefully stop terrain generation goroutine

## References

- Performance Audit: `AUDIT.md` Priority 2, Item 4
- Async Loader: `pkg/procgen/terrain/async_loader.go`
- App States: `pkg/engine/app_state.go`
- Loading UI: `pkg/engine/loading_ui.go`
