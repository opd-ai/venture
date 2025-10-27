# Genre Selection Menu Documentation

## Overview

The Genre Selection Menu is a new navigation screen added in Phase 1.1 that allows players to choose their preferred game genre before starting a new single-player game. This selection influences the procedural generation of all game content, including terrain, entities, items, magic abilities, and skill trees.

## Navigation Flow

```
Main Menu → Single-Player → New Game → Genre Selection → Character Creation → Gameplay
```

The genre selection screen appears between the "New Game" selection and character creation, ensuring players choose their genre before customizing their character.

## Available Genres

The system provides five distinct genres, each with unique thematic elements:

1. **Fantasy** (`"fantasy"`)
   - Medieval fantasy theme with magic, dungeons, and mythical creatures
   - Color palette: Earth tones, magical glows
   - Entity names: Ancient, Dark, Elder, Fire, Shadow, Drake, Lord, Wyrm, Knight
   - Equipment: Swords, bows, staffs, plate armor, potions

2. **Sci-Fi** (`"scifi"`)
   - Futuristic technology and space themes
   - Color palette: Neon blues, metallics, energy glows
   - Entity names: Combat, Security, Battle, Titan, Omega, Android, Cyborg, Mech, Unit, Destroyer
   - Equipment: Laser rifles, plasma guns, powered armor, nanobots, energy shields

3. **Horror** (`"horror"`)
   - Dark, scary atmosphere with emphasis on fear
   - Color palette: Dark tones, blood reds, sickly greens
   - Visual effects: Limited visibility, fog
   - Thematic focus: Dread, survival, tension

4. **Cyberpunk** (`"cyberpunk"`)
   - Urban future with hacking and neon aesthetics
   - Color palette: Neon pinks, purples, blues against dark backgrounds
   - Thematic focus: Technology, corporate themes, urban decay

5. **Post-Apocalyptic** (`"postapoc"`)
   - Survival in wasteland environments
   - Color palette: Browns, grays, rust tones
   - Entity names: Mutated, decayed themes
   - Thematic focus: Scarcity, makeshift equipment, survival

## Input Controls

### Keyboard Navigation
- **Arrow Keys** or **WASD**: Navigate up/down through genre list
- **Number Keys (1-5)**: Quick select corresponding genre
- **Enter/Space**: Select highlighted genre
- **Escape**: Return to Single-Player menu

### Mouse Navigation
- **Hover**: Highlight genre under cursor
- **Left Click**: Select genre
- **Click outside**: No action (menu remains visible)

## Implementation Details

### State Machine
- **State**: `AppStateGenreSelection` (value 2)
- **Valid Transitions**:
  - From: `AppStateSinglePlayerMenu`
  - To: `AppStateCharacterCreation`, `AppStateSinglePlayerMenu`, `AppStateMainMenu`

### Components

#### GenreSelectionMenu (`pkg/engine/genre_selection_menu.go`)
- **Purpose**: UI component for genre selection
- **Genre Source**: `pkg/procgen/genre.DefaultRegistry()`
- **Fields**:
  - `genres []genre.Definition`: List of available genres
  - `selectedIndex int`: Current selection (keyboard navigation)
  - `isVisible bool`: Visibility state
  - `onGenreSelect func(string)`: Callback for genre selection
  - `onBack func()`: Callback for back navigation

#### Key Methods
- `NewGenreSelectionMenu(screenWidth, screenHeight int)`: Initialize with 5 default genres
- `Update()`: Handle input (keyboard and mouse)
- `Draw(screen *ebiten.Image)`: Render menu with genre list
- `Show() / Hide()`: Control visibility
- `getGenreAtPosition(x, y int) int`: Mouse hit detection

### Game Integration (`pkg/engine/game.go`)

#### State Management
```go
// Added field
GenreSelectionMenu *GenreSelectionMenu

// Added field for genre storage
selectedGenreID string

// Handler when "New Game" selected from Single-Player menu
func (g *EbitenGame) handleSinglePlayerMenuNewGame() {
    g.StateManager.TransitionTo(AppStateGenreSelection)
    g.GenreSelectionMenu.Show()
    g.SinglePlayerMenu.Hide()
}

// Handler when genre selected
func (g *EbitenGame) handleGenreSelection(genreID string) {
    g.selectedGenreID = genreID
    g.StateManager.TransitionTo(AppStateCharacterCreation)
    g.GenreSelectionMenu.Hide()
    // Character creation will use selectedGenreID
}

// Getter for world generation
func (g *EbitenGame) GetSelectedGenreID() string {
    genreID := g.selectedGenreID
    g.selectedGenreID = "" // Clear after retrieval
    return genreID
}
```

#### Update Loop
```go
case AppStateGenreSelection:
    if g.GenreSelectionMenu != nil {
        g.GenreSelectionMenu.Update()
    }
```

#### Draw Loop
```go
case AppStateGenreSelection:
    if g.GenreSelectionMenu != nil {
        g.GenreSelectionMenu.Draw(screen)
    }
```

### Layout Calculations

```go
// Screen dimensions
screenWidth := 1280
screenHeight := 720

// Menu positioning
startY := screenHeight/2 - 100  // 260 for 720p
genreHeight := 50

// Genre Y positions (for 720p):
// Fantasy: 260
// Sci-Fi: 310
// Horror: 360
// Cyberpunk: 410
// Post-Apocalyptic: 460
```

## Testing

### Test Coverage
Comprehensive test suite in `pkg/engine/genre_selection_menu_test.go`:

- **Initialization**: `TestNewGenreSelectionMenu`
- **Visibility**: `TestGenreSelectionMenu_ShowHide`
- **Callbacks**: `TestGenreSelectionMenu_SetCallbacks`
- **Selection**: `TestGenreSelectionMenu_SelectCurrentGenre`
- **Navigation**: `TestGenreSelectionMenu_Navigation`
- **Mouse Detection**: `TestGenreSelectionMenu_GetGenreAtPosition`
- **Genre Validation**: `TestGenreSelectionMenu_AllGenresPresent`
- **Integration**: `TestGenreSelectionMenu_IntegrationWithGame`
- **State Retrieval**: `TestGetSelectedGenreID`

### Test Execution
```bash
# Run genre selection tests
go test -v ./pkg/engine -run TestGenreSelection

# Run full engine test suite
go test ./pkg/engine -v

# Check test coverage
go test -cover ./pkg/engine
```

### Example Test: Integration Flow
```go
func TestGenreSelectionMenu_IntegrationWithGame(t *testing.T) {
    game := NewEbitenGame(1280, 720)

    // Verify initialization
    if game.GenreSelectionMenu == nil {
        t.Fatal("expected GenreSelectionMenu to be initialized")
    }

    // Test state flow: Main → Single-Player → Genre Selection
    game.handleMainMenuSelection(MainMenuOptionSinglePlayer)
    game.handleSinglePlayerMenuNewGame()
    
    if game.StateManager.CurrentState() != AppStateGenreSelection {
        t.Error("expected GenreSelection state")
    }

    // Test genre selection
    game.handleGenreSelection("fantasy")
    
    if game.StateManager.CurrentState() != AppStateCharacterCreation {
        t.Error("expected CharacterCreation state")
    }

    // Verify genre stored
    genreID := game.GetSelectedGenreID()
    if genreID != "fantasy" {
        t.Errorf("expected 'fantasy', got '%s'", genreID)
    }
}
```

## Usage for World Generation

The selected genre ID is used when initializing the game world:

```go
// In world generation code
genreID := game.GetSelectedGenreID()
if genreID == "" {
    genreID = "fantasy" // Default fallback
}

// Create generation parameters with genre
params := procgen.GenerationParams{
    Difficulty: 0.5,
    Depth:      1,
    GenreID:    genreID,
    Custom:     make(map[string]interface{}),
}

// Generate world content (terrain, entities, items, etc.)
terrainGen := terrain.NewBSPGenerator()
terrain, err := terrainGen.Generate(worldSeed, params)
```

## Visual Design

### Menu Appearance
- **Title**: "Select Genre" (centered, large font)
- **Genre List**: Vertical list with name and description
- **Number Shortcuts**: Displayed before each genre name (1-5)
- **Highlight**: Selected genre has different color/background
- **Controls Hint**: Bottom of screen shows navigation instructions

### Example Layout
```
                        Select Genre

        1. Fantasy - Medieval fantasy with magic and dungeons
      → 2. Sci-Fi - Futuristic technology and space themes
        3. Horror - Dark and scary atmosphere
        4. Cyberpunk - Urban future with neon and hacking
        5. Post-Apocalyptic - Survival in wasteland

        [1-5] Quick Select  [↑↓/WASD] Navigate  [Enter] Select  [Esc] Back
```

## Future Enhancements

Potential improvements for future phases:

1. **Genre Descriptions**: Expanded descriptions with visual previews
2. **Genre Blending**: Allow mixing two genres (already supported in procgen system)
3. **Custom Genres**: Player-defined genre combinations
4. **Genre Icons**: Visual representations of each genre
5. **Preview Mode**: Show sample terrain/entities before selection
6. **Save Preferences**: Remember last selected genre

## Related Documentation

- [Architecture](ARCHITECTURE.md) - Overall system design
- [Technical Spec](TECHNICAL_SPEC.md) - Implementation details
- [Development Guide](DEVELOPMENT.md) - Contributing guidelines
- [Phase 1 Complete](../PHASE1_COMPLETE.md) - Phase 1 completion status
- [Procedural Genre System](../pkg/procgen/genre/doc.go) - Genre definition system

## Code References

### Key Files
- `pkg/engine/genre_selection_menu.go` - Genre selection menu component (252 lines)
- `pkg/engine/genre_selection_menu_test.go` - Comprehensive test suite (484 lines)
- `pkg/engine/game.go` - Game integration with state management
- `pkg/engine/app_state.go` - State machine with AppStateGenreSelection
- `pkg/procgen/genre/registry.go` - Genre definitions and registry

### Genre System
- `pkg/procgen/genre/` - Core genre system package
- `pkg/procgen/genre/registry.go` - DefaultRegistry() with 5 genres
- All generators accept `GenreID` in `GenerationParams`

## Troubleshooting

### Menu Not Appearing
- Verify state transition: `game.StateManager.CurrentState() == AppStateGenreSelection`
- Check menu visibility: `game.GenreSelectionMenu.IsVisible()`
- Ensure `Show()` called in `handleSinglePlayerMenuNewGame()`

### Genre Not Applied
- Verify `GetSelectedGenreID()` called during world generation
- Check `selectedGenreID` field populated in `handleGenreSelection()`
- Ensure `GenreID` passed to `GenerationParams` for all generators

### State Transition Errors
- Check `isValidAppTransition()` allows:
  - `SinglePlayerMenu → GenreSelection`
  - `GenreSelection → CharacterCreation`
  - `GenreSelection → SinglePlayerMenu` (back)

## Performance Considerations

- **Initialization**: O(1) - 5 genres loaded from registry
- **Update**: O(1) - Simple input handling
- **Draw**: O(n) - Linear in number of genres (n=5, negligible)
- **Memory**: ~2KB for menu state + genre definitions

No performance concerns with current implementation.
