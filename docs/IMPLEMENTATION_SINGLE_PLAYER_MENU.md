# Single-Player Submenu Implementation

**Status**: ✅ Complete (October 27, 2025)  
**Phase**: 1.1 (Menu System Enhancements)  
**Coverage**: selectCurrentOption 90.0%, Draw 88.9%, getOptionAtPosition 100%, all helper methods 100%

## Overview

The Single-Player Submenu provides a dedicated interface for single-player game mode selection, replacing the direct transition from Main Menu → Character Creation. This implementation follows Phase 1.1 enhancement requirements, adding a structured submenu for New Game and future Load Game functionality.

### Key Features

✅ **Three-Option Menu**: New Game, Load Game (placeholder), Back  
✅ **Keyboard Navigation**: Arrow keys, WASD, number keys (1-3)  
✅ **Mouse Support**: Click-to-select with hover detection  
✅ **Dual-Exit Pattern**: ESC key for back navigation  
✅ **Disabled State**: Load Game grayed out with "Coming Soon" message  
✅ **State Management**: Uses `AppStateSinglePlayerMenu` from existing state machine  
✅ **Performance**: 3.19 ns/op for selection, 2.67 ns/op for position detection, zero allocations  

## Architecture

### Component Structure

```
MainMenuUI (Main Menu)
    ↓ (Select "Single-Player")
    ↓
SinglePlayerMenu (Submenu)
    ↓ (Select "New Game")
    ↓
CharacterCreation
    ↓
Gameplay
```

### State Transitions

```
AppStateMainMenu
    ↓ (MainMenuOptionSinglePlayer)
AppStateSinglePlayerMenu
    ↓ (SinglePlayerMenuOptionNewGame)
AppStateCharacterCreation
    ↓ (Character creation complete)
AppStateGameplay

OR

AppStateSinglePlayerMenu
    ↓ (SinglePlayerMenuOptionBack or ESC)
AppStateMainMenu
```

## Implementation

### File: `pkg/engine/single_player_menu.go` (276 lines)

#### Key Types

**SinglePlayerMenuOption** (enum):
- `SinglePlayerMenuOptionNewGame` (0): Starts new game after character creation
- `SinglePlayerMenuOptionLoadGame` (1): Loads saved game (Phase 8.3, currently disabled)
- `SinglePlayerMenuOptionBack` (2): Returns to main menu

**SinglePlayerMenu** (struct):
```go
type SinglePlayerMenu struct {
    screenWidth  int
    screenHeight int
    selectedIdx  int
    options      []SinglePlayerMenuOption
    visible      bool

    // Callbacks
    onNewGame  func()
    onLoadGame func()
    onBack     func()
}
```

#### Key Methods

**NewSinglePlayerMenu(screenWidth, screenHeight int)**
- Creates and initializes the submenu
- Sets up three options in order: New Game, Load Game, Back
- Starts invisible with first option selected

**Show() / Hide() / IsVisible()**
- `Show()`: Makes menu visible, resets selection to first option
- `Hide()`: Makes menu invisible
- `IsVisible()`: Returns current visibility state

**Update() bool**
- Processes keyboard and mouse input when visible
- Returns `false` if invisible (no processing)
- Returns `true` if an option was selected
- Handles:
  * ESC key: triggers back callback (dual-exit pattern)
  * Up/Down arrows, W/S: navigation with wrapping
  * Number keys 1-3: direct selection shortcuts
  * Enter/Space: select current option
  * Mouse click: select option at cursor
  * Mouse hover: update selection

**selectCurrentOption() bool**
- Triggers callback for currently selected option
- Returns `false` for disabled options (Load Game)
- Returns `true` for enabled options (New Game, Back)

**Draw(screen *ebiten.Image)**
- No-op if invisible
- Renders:
  * Semi-transparent dark background (RGBA 0,0,0,200)
  * "Single-Player" title
  * "Select a game mode" subtitle
  * Three menu options with selection indicator (">")
  * "(Coming Soon)" suffix for disabled Load Game
  * Hint text for disabled option when selected
  * Controls reminder at bottom
- Disabled options show "(Coming Soon)" suffix but no gray coloring (DebugPrintAt limitation)

### Integration Points

#### pkg/engine/game.go

**Added Field:**
```go
SinglePlayerMenu *SinglePlayerMenu // Submenu for single-player options
```

**Initialization** (NewEbitenGameWithLogger):
```go
SinglePlayerMenu: NewSinglePlayerMenu(screenWidth, screenHeight),
```

**Callback Wiring:**
```go
game.SinglePlayerMenu.SetNewGameCallback(game.handleSinglePlayerMenuNewGame)
game.SinglePlayerMenu.SetLoadGameCallback(game.handleSinglePlayerMenuLoadGame)
game.SinglePlayerMenu.SetBackCallback(game.handleSinglePlayerMenuBack)
```

**Handler Methods:**

`handleSinglePlayerMenuNewGame()`:
- Transitions to `AppStateCharacterCreation`
- Resets character creation UI
- Sets `isMultiplayerMode = false`
- Hides single-player menu

`handleSinglePlayerMenuLoadGame()`:
- Placeholder for Phase 8.3
- Logs "not yet implemented"

`handleSinglePlayerMenuBack()`:
- Transitions back to `AppStateMainMenu`
- Hides single-player menu

**Update Loop:**
```go
if g.StateManager.CurrentState() == AppStateSinglePlayerMenu {
    if g.SinglePlayerMenu != nil {
        g.SinglePlayerMenu.Update()
    }
    return nil
}
```

**Draw Loop:**
```go
if g.StateManager.CurrentState() == AppStateSinglePlayerMenu {
    if g.SinglePlayerMenu != nil {
        g.SinglePlayerMenu.Draw(screen)
    }
    return
}
```

**Modified handleMainMenuSelection:**
```go
case MainMenuOptionSinglePlayer:
    // Transition to single-player submenu
    if err := g.StateManager.TransitionTo(AppStateSinglePlayerMenu); err != nil {
        // error handling
        return
    }

    // Show single-player menu
    if g.SinglePlayerMenu != nil {
        g.SinglePlayerMenu.Show()
    }
```

### State Machine (Already Existed)

The `AppStateSinglePlayerMenu` state was already defined in `pkg/engine/app_state.go` as part of the original Phase 1 design. This implementation fulfills the intended use of that state.

## User Experience

### Navigation Flow

1. **Game Start** → Main Menu (AppStateMainMenu)
2. **Select "Single-Player"** → Single-Player Submenu (AppStateSinglePlayerMenu)
3. **Select "New Game"** → Character Creation (AppStateCharacterCreation)
4. **Complete Character** → Gameplay (AppStateGameplay)

### Alternative Flows

**Back Navigation:**
- Single-Player Submenu → **Press ESC or select "Back"** → Main Menu

**Disabled Option:**
- Single-Player Submenu → **Select "Load Game"** → Shows hint "Save/Load system coming in Phase 8.3", no transition

### Keyboard Controls

- **Arrow Up/Down** or **W/S**: Navigate options
- **1**: Quick-select New Game
- **2**: Quick-select Load Game (disabled, shows hint)
- **3**: Quick-select Back
- **Enter** or **Space**: Select current option
- **ESC**: Back to main menu (dual-exit)

### Mouse Controls

- **Click**: Select option at cursor position
- **Hover**: Highlight option under cursor
- Clickable area: Approximate text bounds ± 20 pixels

## Testing

### Test File: `pkg/engine/single_player_menu_test.go` (393 lines)

#### Test Functions (12 total)

1. **TestSinglePlayerMenuOption_String**: String representations (4 sub-tests)
2. **TestNewSinglePlayerMenu**: Initialization and defaults
3. **TestSinglePlayerMenu_ShowHide**: Visibility toggling and selection reset
4. **TestSinglePlayerMenu_SetCallbacks**: Callback registration and invocation
5. **TestSinglePlayerMenu_Update_NotVisible**: No-op when invisible
6. **TestSinglePlayerMenu_SelectCurrentOption_NewGame**: New Game callback
7. **TestSinglePlayerMenu_SelectCurrentOption_LoadGame**: Disabled option (no callback)
8. **TestSinglePlayerMenu_SelectCurrentOption_Back**: Back callback
9. **TestSinglePlayerMenu_Navigation**: Arrow key wrapping logic (6 sub-tests)
10. **TestSinglePlayerMenu_GetOptionAtPosition**: Mouse position detection (7 sub-tests)
11. **TestSinglePlayerMenu_Draw**: Rendering (Ebiten-dependent, skipped in CI)
12. **TestSinglePlayerMenu_IntegrationWithGame**: Full state transition flow

#### Benchmarks (2 total)

1. **BenchmarkSinglePlayerMenu_SelectCurrentOption**: 3.191 ns/op, 0 allocs
2. **BenchmarkSinglePlayerMenu_GetOptionAtPosition**: 2.671 ns/op, 0 allocs

#### Coverage Results

```
Function                    Coverage
------------------------------------
String()                    100.0%
NewSinglePlayerMenu()       100.0%
SetNewGameCallback()        100.0%
SetLoadGameCallback()       100.0%
SetBackCallback()           100.0%
Show()                      100.0%
Hide()                      100.0%
IsVisible()                 100.0%
Update()                      5.9%  (Ebiten input handling)
selectCurrentOption()        90.0%
getOptionAtPosition()       100.0%
Draw()                       88.9%
```

**Overall**: Core logic 90-100%, Update() low due to untestable Ebiten input handling (standard limitation).

### Test Scenarios Covered

✅ Option string representations (including unknown)  
✅ Menu initialization with correct defaults  
✅ Show/hide functionality and visibility state  
✅ Selection reset on Show()  
✅ Callback registration and invocation  
✅ Update returns false when invisible  
✅ New Game selection triggers callback  
✅ Load Game selection does NOT trigger callback (disabled)  
✅ Back selection triggers callback  
✅ Navigation wrapping (up/down)  
✅ Mouse position detection (hit testing)  
✅ Draw operations (smoke test)  
✅ Full integration with game state machine  

## Performance

### Benchmark Results

```
BenchmarkSinglePlayerMenu_SelectCurrentOption-16    376817089    3.191 ns/op    0 B/op    0 allocs/op
BenchmarkSinglePlayerMenu_GetOptionAtPosition-16    443520403    2.671 ns/op    0 B/op    0 allocs/op
```

**Analysis:**
- **SelectCurrentOption**: 3.19 nanoseconds per operation (extremely fast)
- **GetOptionAtPosition**: 2.67 nanoseconds per operation (mouse detection)
- **Zero allocations**: All operations are stack-only, no heap pressure
- **Called frequency**: Once per frame when visible (~60 FPS = 16.6ms budget)
- **Performance impact**: Negligible (<0.001% of frame budget)

## Future Enhancements (Phase 8.3)

When the Save/Load system is implemented in Phase 8.3:

### 1. Enable Load Game Option

**File**: `pkg/engine/single_player_menu.go`

```go
// Remove disabled check in selectCurrentOption
case SinglePlayerMenuOptionLoadGame:
    if m.onLoadGame != nil {
        m.onLoadGame()
    }
    return true  // Now returns true instead of false
```

```go
// Remove "(Coming Soon)" suffix in Draw
if isDisabled {
    optionText = optionText + " (Coming Soon)"  // DELETE THIS LINE
}
```

### 2. Implement Load Game Handler

**File**: `pkg/engine/game.go`

```go
func (g *EbitenGame) handleSinglePlayerMenuLoadGame() {
    // Transition to save file selection UI
    if err := g.StateManager.TransitionTo(AppStateSaveLoadMenu); err != nil {
        // error handling
        return
    }

    // Show save file browser
    if g.SaveLoadUI != nil {
        g.SaveLoadUI.Show()
    }

    // Hide single-player menu
    if g.SinglePlayerMenu != nil {
        g.SinglePlayerMenu.Hide()
    }
}
```

### 3. Add Save File Selection State

**File**: `pkg/engine/app_state.go`

```go
const (
    // ... existing states ...
    AppStateSaveLoadMenu  // New state for browsing save files
)
```

### 4. Create SaveLoadUI Component

Similar to `SinglePlayerMenu`, implement:
- `pkg/engine/save_load_ui.go`: File browser for save files
- List saved games with metadata (character name, level, play time, screenshot)
- Select file → Load game → Transition to AppStateGameplay
- Integration with `pkg/saveload` package

## Known Limitations

1. **Load Game Disabled**: Placeholder until Phase 8.3 (Save/Load System)
2. **No Color Support**: DebugPrintAt used for rendering (monochrome)
3. **Ebiten Input Dependency**: Update() method cannot be fully tested without Ebiten runtime
4. **Fixed Layout**: Screen coordinates hardcoded (not resolution-independent)

## Integration Checklist

For developers modifying the single-player menu:

- [ ] Update `SinglePlayerMenuOption` enum if adding options
- [ ] Add corresponding case in `selectCurrentOption()`
- [ ] Add corresponding case in `Draw()` for rendering
- [ ] Add callback setter method (e.g., `SetNewOptionCallback`)
- [ ] Wire callback in `NewEbitenGameWithLogger`
- [ ] Add handler method in `game.go` (e.g., `handleSinglePlayerMenuNewOption`)
- [ ] Add test cases in `single_player_menu_test.go`
- [ ] Update this documentation

## Comparison with Main Menu

### Similarities (Design Consistency)

- Enum-based option system
- Callback architecture for actions
- Show/Hide visibility pattern
- Keyboard navigation with wrapping
- Mouse hover and click support
- Number key shortcuts (1-3)
- DebugPrintAt rendering (monochrome)

### Differences (Submenu-Specific)

- **ESC Key Behavior**: SinglePlayerMenu uses ESC for back navigation (dual-exit), MainMenu doesn't handle ESC
- **Disabled Options**: SinglePlayerMenu has disabled Load Game with visual feedback, MainMenu has no disabled options
- **Visibility Control**: SinglePlayerMenu has Show()/Hide() methods, MainMenu is always "visible" (controlled by AppState)
- **Title/Subtitle**: SinglePlayerMenu has "Single-Player" title + "Select a game mode" subtitle, MainMenu shows venture logo title

## Conclusion

The Single-Player Submenu successfully implements Phase 1.1 enhancement requirements:

✅ **Three-option menu** (New Game, Load Game placeholder, Back)  
✅ **Dual-exit pattern** (ESC key for back)  
✅ **State machine integration** (AppStateSinglePlayerMenu)  
✅ **Comprehensive testing** (12 test functions, 90%+ coverage on testable code)  
✅ **Performance validated** (<4ns per operation, zero allocations)  
✅ **Production-ready** (zero build errors, all tests passing)  

**Status**: Ready for production use. Players can now access New Game through a dedicated submenu. Load Game placeholder prepared for Phase 8.3 implementation.
