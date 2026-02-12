package engine

import (
	"sync"
	"testing"

	"github.com/opd-ai/venture/pkg/world/housing"
)

// BenchmarkInitializeUIComponents_Parallel benchmarks the parallel UI initialization.
func BenchmarkInitializeUIComponents_Parallel(b *testing.B) {
	world := NewWorld()
	settingsManager, err := NewSettingsManager()
	if err != nil {
		b.Fatalf("Failed to create settings manager: %v", err)
	}
	screenWidth, screenHeight := 800, 600

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = initializeUIComponents(world, screenWidth, screenHeight, settingsManager)
	}
}

// BenchmarkInitializeUIComponents_Sequential benchmarks sequential UI initialization
// for comparison with parallel implementation.
func BenchmarkInitializeUIComponents_Sequential(b *testing.B) {
	world := NewWorld()
	settingsManager, err := NewSettingsManager()
	if err != nil {
		b.Fatalf("Failed to create settings manager: %v", err)
	}
	screenWidth, screenHeight := 800, 600

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Sequential initialization (old pattern)
		ui := &uiComponents{}
		ui.inventoryUI = NewEbitenInventoryUI(world, screenWidth, screenHeight)
		ui.questUI = NewEbitenQuestUI(world, screenWidth, screenHeight)
		ui.characterUI = NewEbitenCharacterUI(world, screenWidth, screenHeight)
		ui.skillsUI = NewEbitenSkillsUI(world, screenWidth, screenHeight)
		ui.mapUI = NewEbitenMapUI(world, screenWidth, screenHeight)
		ui.settingsUI = NewSettingsUI(screenWidth, screenHeight, settingsManager)
		ui.mainMenuUI = NewMainMenuUI(screenWidth, screenHeight)
		ui.singlePlayerMenu = NewSinglePlayerMenu(screenWidth, screenHeight)
		ui.genreSelectionMenu = NewGenreSelectionMenu(screenWidth, screenHeight)
		ui.multiplayerMenu = NewMultiplayerMenu(screenWidth, screenHeight)
		ui.serverAddressInput = NewServerAddressInput(screenWidth, screenHeight)
		ui.characterCreation = NewCharacterCreation(screenWidth, screenHeight)
		ui.galleryUI = NewGalleryUI(screenWidth, screenHeight)
		ui.housingUI = housing.NewHousingUI(screenWidth, screenHeight)
		ui.guildUI = NewGuildUI(world, nil, screenWidth, screenHeight)
		_ = ui
	}
}

// BenchmarkInitializeUIComponents_ParallelMultipleWorlds benchmarks parallel initialization
// with different world instances (simulates multiple game instances).
func BenchmarkInitializeUIComponents_ParallelMultipleWorlds(b *testing.B) {
	screenWidth, screenHeight := 800, 600

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world := NewWorld()
		settingsManager, err := NewSettingsManager()
		if err != nil {
			b.Fatalf("Failed to create settings manager: %v", err)
		}
		_ = initializeUIComponents(world, screenWidth, screenHeight, settingsManager)
	}
}

// BenchmarkInitializeUIComponents_ConcurrentInitializations benchmarks concurrent
// initialization of multiple UI sets (stress test for goroutine overhead).
func BenchmarkInitializeUIComponents_ConcurrentInitializations(b *testing.B) {
	world := NewWorld()
	settingsManager, err := NewSettingsManager()
	if err != nil {
		b.Fatalf("Failed to create settings manager: %v", err)
	}
	screenWidth, screenHeight := 800, 600

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(4)
		for j := 0; j < 4; j++ {
			go func() {
				defer wg.Done()
				_ = initializeUIComponents(world, screenWidth, screenHeight, settingsManager)
			}()
		}
		wg.Wait()
	}
}

// BenchmarkUIComponent_InventoryInit benchmarks individual inventory UI initialization
// to measure per-component overhead.
func BenchmarkUIComponent_InventoryInit(b *testing.B) {
	world := NewWorld()
	screenWidth, screenHeight := 800, 600

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewEbitenInventoryUI(world, screenWidth, screenHeight)
	}
}

// BenchmarkUIComponent_MainMenuInit benchmarks individual main menu UI initialization.
func BenchmarkUIComponent_MainMenuInit(b *testing.B) {
	screenWidth, screenHeight := 800, 600

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewMainMenuUI(screenWidth, screenHeight)
	}
}

// BenchmarkUIComponent_SettingsInit benchmarks individual settings UI initialization.
func BenchmarkUIComponent_SettingsInit(b *testing.B) {
	settingsManager, err := NewSettingsManager()
	if err != nil {
		b.Fatalf("Failed to create settings manager: %v", err)
	}
	screenWidth, screenHeight := 800, 600

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewSettingsUI(screenWidth, screenHeight, settingsManager)
	}
}
