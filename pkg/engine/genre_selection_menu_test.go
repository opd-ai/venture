package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/procgen/genre"
)

// TestNewGenreSelectionMenu tests menu initialization.
func TestNewGenreSelectionMenu(t *testing.T) {
	width, height := 1280, 720
	menu := NewGenreSelectionMenu(width, height)

	if menu == nil {
		t.Fatal("expected non-nil menu")
	}

	if menu.screenWidth != width {
		t.Errorf("expected width %d, got %d", width, menu.screenWidth)
	}

	if menu.screenHeight != height {
		t.Errorf("expected height %d, got %d", height, menu.screenHeight)
	}

	if menu.selectedIdx != 0 {
		t.Errorf("expected initial selection 0, got %d", menu.selectedIdx)
	}

	if menu.visible {
		t.Error("expected menu to be initially invisible")
	}

	// Should have 5 predefined genres
	expectedGenres := 5
	if len(menu.genres) != expectedGenres {
		t.Errorf("expected %d genres, got %d", expectedGenres, len(menu.genres))
	}

	// Verify genres are loaded from registry
	registry := genre.DefaultRegistry()
	registryGenres := registry.All()
	if len(menu.genres) != len(registryGenres) {
		t.Errorf("expected %d genres from registry, got %d in menu", len(registryGenres), len(menu.genres))
	}
}

// TestGenreSelectionMenu_ShowHide tests visibility toggling.
func TestGenreSelectionMenu_ShowHide(t *testing.T) {
	menu := NewGenreSelectionMenu(1280, 720)

	// Initially invisible
	if menu.IsVisible() {
		t.Error("expected menu to be initially invisible")
	}

	// Show menu
	menu.Show()
	if !menu.IsVisible() {
		t.Error("expected menu to be visible after Show()")
	}

	// Show resets selection to first genre
	menu.selectedIdx = 2
	menu.Show()
	if menu.selectedIdx != 0 {
		t.Errorf("expected selection reset to 0 after Show(), got %d", menu.selectedIdx)
	}

	// Hide menu
	menu.Hide()
	if menu.IsVisible() {
		t.Error("expected menu to be invisible after Hide()")
	}
}

// TestGenreSelectionMenu_SetCallbacks tests callback registration.
func TestGenreSelectionMenu_SetCallbacks(t *testing.T) {
	menu := NewGenreSelectionMenu(1280, 720)

	genreSelectCalled := false
	var selectedGenreID string
	backCalled := false

	menu.SetGenreSelectCallback(func(genreID string) {
		genreSelectCalled = true
		selectedGenreID = genreID
	})

	menu.SetBackCallback(func() {
		backCalled = true
	})

	// Verify callbacks are set
	if menu.onGenreSelect == nil {
		t.Error("expected onGenreSelect callback to be set")
	}
	if menu.onBack == nil {
		t.Error("expected onBack callback to be set")
	}

	// Test genre select callback
	testGenreID := "fantasy"
	menu.onGenreSelect(testGenreID)
	if !genreSelectCalled {
		t.Error("expected onGenreSelect to be called")
	}
	if selectedGenreID != testGenreID {
		t.Errorf("expected genreID %q, got %q", testGenreID, selectedGenreID)
	}

	// Test back callback
	menu.onBack()
	if !backCalled {
		t.Error("expected onBack to be called")
	}
}

// TestGenreSelectionMenu_Update_NotVisible tests that update does nothing when invisible.
func TestGenreSelectionMenu_Update_NotVisible(t *testing.T) {
	menu := NewGenreSelectionMenu(1280, 720)
	menu.Hide() // Ensure invisible

	callbackCalled := false
	menu.SetGenreSelectCallback(func(genreID string) {
		callbackCalled = true
	})

	// Update should return false and not process input
	selected := menu.Update()
	if selected {
		t.Error("expected Update to return false when menu is invisible")
	}

	if callbackCalled {
		t.Error("expected no callback to be called when menu is invisible")
	}
}

// TestGenreSelectionMenu_SelectCurrentGenre tests genre selection.
func TestGenreSelectionMenu_SelectCurrentGenre(t *testing.T) {
	menu := NewGenreSelectionMenu(1280, 720)
	menu.Show()

	var selectedGenreID string
	menu.SetGenreSelectCallback(func(genreID string) {
		selectedGenreID = genreID
	})

	// Select first genre (any valid genre)
	menu.selectedIdx = 0
	selected := menu.selectCurrentGenre()
	if !selected {
		t.Error("expected selectCurrentGenre to return true")
	}

	firstGenreID := selectedGenreID
	registry := genre.DefaultRegistry()
	if g, err := registry.Get(firstGenreID); err != nil || g == nil {
		t.Errorf("first genre ID %q is not in registry: %v", firstGenreID, err)
	}

	// Select second genre (should be different from first if registry has at least 2 genres)
	menu.selectedIdx = 1
	selected = menu.selectCurrentGenre()
	if !selected {
		t.Error("expected selectCurrentGenre to return true")
	}

	secondGenreID := selectedGenreID
	if g, err := registry.Get(secondGenreID); err != nil || g == nil {
		t.Errorf("second genre ID %q is not in registry: %v", secondGenreID, err)
	}

	// Verify different genres if registry has at least 2
	if menu.GetGenreCount() >= 2 && firstGenreID == secondGenreID {
		t.Errorf("expected different genres at indices 0 and 1")
	}
}

// TestGenreSelectionMenu_Navigation tests keyboard navigation (simulated).
func TestGenreSelectionMenu_Navigation(t *testing.T) {
	menu := NewGenreSelectionMenu(1280, 720)
	menu.Show()

	genreCount := len(menu.genres)

	// Test navigation wrapping
	tests := []struct {
		name        string
		startIdx    int
		navigation  func()
		expectedIdx int
	}{
		{"down from first", 0, func() {
			menu.selectedIdx++
			if menu.selectedIdx >= genreCount {
				menu.selectedIdx = 0
			}
		}, 1},
		{"down from last wrap", genreCount - 1, func() {
			menu.selectedIdx++
			if menu.selectedIdx >= genreCount {
				menu.selectedIdx = 0
			}
		}, 0},
		{"up from first wrap", 0, func() {
			menu.selectedIdx--
			if menu.selectedIdx < 0 {
				menu.selectedIdx = genreCount - 1
			}
		}, genreCount - 1},
		{"up from second", 1, func() {
			menu.selectedIdx--
			if menu.selectedIdx < 0 {
				menu.selectedIdx = genreCount - 1
			}
		}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu.selectedIdx = tt.startIdx
			tt.navigation()
			if menu.selectedIdx != tt.expectedIdx {
				t.Errorf("expected index %d, got %d", tt.expectedIdx, menu.selectedIdx)
			}
		})
	}
}

// TestGenreSelectionMenu_GetGenreAtPosition tests mouse position detection.
func TestGenreSelectionMenu_GetGenreAtPosition(t *testing.T) {
	menu := NewGenreSelectionMenu(1280, 720)

	tests := []struct {
		name     string
		x        int
		y        int
		expected int
	}{
		{"outside menu top", 10, 10, -1},
		{"far right", 1200, 310, -1},
		{"far left", 50, 310, -1},
		{"near first genre", 640, 260, 0},  // startY=260 (720/2-100), height=50
		{"near second genre", 640, 310, 1}, // 260 + 50 = 310
		{"near third genre", 640, 360, 2},  // 260 + 100 = 360
		{"far below", 640, 700, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := menu.getGenreAtPosition(tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestGenreSelectionMenu_GetSelectedGenre tests getting the currently selected genre.
func TestGenreSelectionMenu_GetSelectedGenre(t *testing.T) {
	menu := NewGenreSelectionMenu(1280, 720)

	// First genre selected by default (any genre is valid, just verify it's not nil)
	selectedGenre := menu.GetSelectedGenre()
	if selectedGenre == nil {
		t.Fatal("expected non-nil genre")
	}
	firstGenreID := selectedGenre.ID

	// Verify it's a valid genre ID from the registry
	registry := genre.DefaultRegistry()
	if g, err := registry.Get(firstGenreID); err != nil || g == nil {
		t.Errorf("first genre ID %q is not in registry: %v", firstGenreID, err)
	}

	// Change selection to second index
	menu.selectedIdx = 1
	genre2 := menu.GetSelectedGenre()
	if genre2 == nil {
		t.Fatal("expected non-nil genre for second index")
	}
	secondGenreID := genre2.ID

	// Verify it's different from first (assuming registry has at least 2 genres)
	if menu.GetGenreCount() >= 2 && firstGenreID == secondGenreID {
		t.Errorf("expected different genres at indices 0 and 1")
	}

	// Verify second genre is also valid
	if g, err := registry.Get(secondGenreID); err != nil || g == nil {
		t.Errorf("second genre ID %q is not in registry: %v", secondGenreID, err)
	}

	// Out of bounds should return nil
	menu.selectedIdx = 999
	selectedGenre = menu.GetSelectedGenre()
	if selectedGenre != nil {
		t.Error("expected nil genre for out-of-bounds index")
	}
}

// TestGenreSelectionMenu_GetGenreCount tests genre count.
func TestGenreSelectionMenu_GetGenreCount(t *testing.T) {
	menu := NewGenreSelectionMenu(1280, 720)

	count := menu.GetGenreCount()
	expected := 5 // 5 predefined genres
	if count != expected {
		t.Errorf("expected %d genres, got %d", expected, count)
	}
}

// TestGenreSelectionMenu_Draw tests drawing (no-op when invisible).
func TestGenreSelectionMenu_Draw(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Ebiten-dependent test in short mode")
	}

	menu := NewGenreSelectionMenu(1280, 720)

	// Create a minimal test image
	img := ebiten.NewImage(1280, 720)

	// Draw when invisible (should be no-op)
	menu.Hide()
	menu.Draw(img) // Should not panic

	// Draw when visible (should render)
	menu.Show()
	menu.Draw(img) // Should not panic

	// No assertions - just verify it doesn't crash
}

// TestGenreSelectionMenu_IntegrationWithGame tests menu integration with game state.
func TestGenreSelectionMenu_IntegrationWithGame(t *testing.T) {
	game := NewEbitenGame(1280, 720)

	// Verify menu is initialized
	if game.GenreSelectionMenu == nil {
		t.Fatal("expected GenreSelectionMenu to be initialized")
	}

	// Verify menu is initially invisible
	if game.GenreSelectionMenu.IsVisible() {
		t.Error("expected menu to be initially invisible")
	}

	// Verify callbacks are wired
	if game.GenreSelectionMenu.onGenreSelect == nil {
		t.Error("expected onGenreSelect callback to be wired")
	}
	if game.GenreSelectionMenu.onBack == nil {
		t.Error("expected onBack callback to be wired")
	}

	// Test state transitions
	initialState := game.StateManager.CurrentState()
	if initialState != AppStateMainMenu {
		t.Errorf("expected initial state MainMenu, got %s", initialState.String())
	}

	// Simulate: Main Menu → Single-Player → New Game → Genre Selection
	game.handleMainMenuSelection(MainMenuOptionSinglePlayer)
	if game.StateManager.CurrentState() != AppStateSinglePlayerMenu {
		t.Errorf("expected SinglePlayerMenu state, got %s", game.StateManager.CurrentState().String())
	}

	game.handleSinglePlayerMenuNewGame()
	if game.StateManager.CurrentState() != AppStateGenreSelection {
		t.Errorf("expected GenreSelection state, got %s", game.StateManager.CurrentState().String())
	}

	// Menu should be visible
	if !game.GenreSelectionMenu.IsVisible() {
		t.Error("expected menu to be visible after transition")
	}

	// Test genre selection
	testGenreID := "fantasy"
	game.handleGenreSelection(testGenreID)

	// Should transition to character creation
	if game.StateManager.CurrentState() != AppStateCharacterCreation {
		t.Errorf("expected CharacterCreation state, got %s", game.StateManager.CurrentState().String())
	}

	// Menu should be hidden
	if game.GenreSelectionMenu.IsVisible() {
		t.Error("expected menu to be hidden after genre selection")
	}

	// Genre should be stored
	if game.selectedGenreID != testGenreID {
		t.Errorf("expected selectedGenreID %q, got %q", testGenreID, game.selectedGenreID)
	}

	// Test back navigation
	game.StateManager.TransitionTo(AppStateGenreSelection)
	game.GenreSelectionMenu.Show()
	game.handleGenreSelectionBack()

	// Should return to single-player menu
	if game.StateManager.CurrentState() != AppStateSinglePlayerMenu {
		t.Errorf("expected SinglePlayerMenu state after back, got %s", game.StateManager.CurrentState().String())
	}

	// Genre selection menu should be hidden
	if game.GenreSelectionMenu.IsVisible() {
		t.Error("expected menu to be hidden after back")
	}

	// Single-player menu should be visible
	if !game.SinglePlayerMenu.IsVisible() {
		t.Error("expected single-player menu to be visible after back")
	}
}

// TestGenreSelectionMenu_AllGenresPresent tests that all expected genres are available.
func TestGenreSelectionMenu_AllGenresPresent(t *testing.T) {
	menu := NewGenreSelectionMenu(1280, 720)

	expectedGenres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	genreMap := make(map[string]bool)
	for _, g := range menu.genres {
		genreMap[g.ID] = true
	}

	for _, expected := range expectedGenres {
		if !genreMap[expected] {
			t.Errorf("expected genre %q not found in menu", expected)
		}
	}
}

// TestGetSelectedGenreID tests genre ID retrieval and clearing.
func TestGetSelectedGenreID(t *testing.T) {
	game := NewEbitenGame(1280, 720)

	// Initially empty
	genreID := game.GetSelectedGenreID()
	if genreID != "" {
		t.Errorf("expected empty genre ID, got %q", genreID)
	}

	// Set genre
	testGenreID := "cyberpunk"
	game.selectedGenreID = testGenreID

	// Get should return and clear
	genreID = game.GetSelectedGenreID()
	if genreID != testGenreID {
		t.Errorf("expected genre ID %q, got %q", testGenreID, genreID)
	}

	// Should be cleared after retrieval
	genreID = game.GetSelectedGenreID()
	if genreID != "" {
		t.Errorf("expected empty genre ID after second retrieval, got %q", genreID)
	}
}

// Benchmark genre selection operations
func BenchmarkGenreSelectionMenu_SelectCurrentGenre(b *testing.B) {
	menu := NewGenreSelectionMenu(1280, 720)
	menu.Show()
	menu.SetGenreSelectCallback(func(genreID string) {})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		menu.selectCurrentGenre()
	}
}

func BenchmarkGenreSelectionMenu_GetGenreAtPosition(b *testing.B) {
	menu := NewGenreSelectionMenu(1280, 720)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		menu.getGenreAtPosition(640, 310)
	}
}
