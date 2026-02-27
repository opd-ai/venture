// Package engine provides the main game loop and Ebiten integration.
// This file implements EbitenGame which ties together the ECS world, rendering
// systems, and the Ebiten game engine. EbitenGame implements both ebiten.Game
// and GameRunner interfaces.
package engine

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/world/housing"
	"github.com/sirupsen/logrus"
)

// EbitenGame represents the main game instance with the ECS world and game loop.
// Implements both ebiten.Game and GameRunner interfaces.
type EbitenGame struct {
	World          *World
	lastUpdateTime time.Time
	ScreenWidth    int
	ScreenHeight   int
	Paused         bool

	// Application state management
	StateManager              *AppStateManager
	MainMenuUI                *MainMenuUI
	SinglePlayerMenu          *SinglePlayerMenu   // Submenu for single-player options
	GenreSelectionMenu        *GenreSelectionMenu // Genre selection for single-player
	MultiplayerMenu           *MultiplayerMenu    // Submenu for multiplayer options
	ServerAddressInput        *ServerAddressInput // Text input for server address
	SettingsUI                *SettingsUI
	SettingsManager           *SettingsManager
	CharacterCreation         *EbitenCharacterCreation
	CharacterCreationTutorial *CharacterCreationTutorial // Tutorial overlay for character creation
	OnboardingManager         *OnboardingManager         // Coordinates all tutorial layers (Phase 3.2)
	LoadingUI                 *LoadingUI                 // World generation loading screen (Phase Performance Audit)
	pendingCharData           *CharacterData
	isMultiplayerMode         bool   // Track if character creation is for multiplayer
	selectedGenreID           string // Selected genre for world generation
	pendingServerAddress      string // Server address for Join option
	worldSeed                 int64  // World generation seed for deterministic content

	// Rendering systems
	CameraSystem        *CameraSystem
	RenderSystem        *EbitenRenderSystem
	TerrainRenderSystem *TerrainRenderSystem
	LightingSystem      *LightingSystem       // Dynamic lighting system (Phase 5.1)
	PostProcessor       *PostProcessorAdapter // Post-processing effects (Phase 5.3)
	sceneBuffer         *ebiten.Image         // Reusable buffer for lighting/post-processing
	litBuffer           *ebiten.Image         // Cached buffer for lighting effects (Performance Audit fix)
	HUDSystem           *EbitenHUDSystem
	TutorialSystem      *EbitenTutorialSystem
	HelpSystem          *EbitenHelpSystem
	MenuSystem          *EbitenMenuSystem

	// Context-sensitive tutorial manager (Phase 3.3 - ShowTutorials wiring)
	ContextualTutorial ContextualTutorialProvider

	// UI systems
	InventoryUI *EbitenInventoryUI
	QuestUI     *EbitenQuestUI
	CharacterUI *EbitenCharacterUI
	SkillsUI    *EbitenSkillsUI
	MapUI       *EbitenMapUI
	ShopUI      *ShopUI     // Commerce and merchant interaction UI
	CraftingUI  *CraftingUI // Crafting and recipe UI
	MailboxUI   *MailboxUI  // Mail system UI (Phase 40.3)
	TradeUI     *TradeUI    // Player-to-player trading UI (Phase 3.3)

	// INTEGRATION FIX [Category B]: V8.0 UI systems
	// Gap: V8 systems (housing, gallery) fully implemented but no UI fields
	// Fix: Added UI fields for housing and gallery management
	// Roadmap: ROADMAP_V8.md Phase 49.1, 49.4
	HousingUI *housing.HousingUI // Player housing management UI (Phase 49.1, 51.2, 51.3)
	GalleryUI *GalleryUI         // Image gallery viewer UI (Phase 49.4)

	// Phase 3.2 (PLAN.md): Guild Federation
	GuildUI *GuildUI // Guild management UI

	// Phase 4.2 (PLAN.md): Advanced Classes
	AdvancedClassUI *AdvancedClassUI // Advanced class management UI (multi-classing, prestige, talents)

	// Phase 4.3 (PLAN.md): Territory Control
	TerritoryUI *TerritoryUI // Territory management UI (guild warfare, territory capture)

	// Phase 6.1 (PLAN.md): Branching Narratives
	StoryChoiceUI *StoryChoiceUI // Branching narrative choice UI (story decisions, consequences)

	// Phase 6.2 (PLAN.md): Dialog System
	DialogUI *DialogUI // NPC dialog UI (conversations, dialog choices)

	// Phase 4.2 (PLAN.md): Prestige System
	PrestigeUI PrestigeUIProvider // Prestige and paragon point allocation UI - stored as interface to avoid import cycle

	// Audio system (for settings integration)
	AudioManager *AudioManager

	// Player entity reference (for UI systems)
	PlayerEntity *Entity

	// Callbacks for state transitions
	onNewGame            func() error
	onMultiplayerConnect func(serverAddr string) error
	onQuitToMenu         func() error

	// Async world generation (Performance Audit - async terrain loading)
	terrainLoader       interface{}         // *terrain.AsyncLoader - stored as interface{} to avoid circular import
	terrainLoadComplete func(*Entity) error // Callback when terrain loading completes

	// Logger for game operations
	logger *logrus.Entry

	// Performance monitoring
	frameTimeTracker *FrameTimeTracker
	frameCount       uint64
	profilingEnabled bool

	// Render interpolation: alpha = fraction of tick elapsed since last Update(),
	// used to interpolate entity positions between PrevX/PrevY and X/Y in Draw().
	// This eliminates visual jitter on high-refresh-rate monitors where Draw() is
	// called more frequently than Update().
	renderAlpha float64

	// Cached reference to InputSystem to avoid linear scan of all systems every Draw frame
	cachedInputSystem *InputSystem

	// Cached entity list for Draw() to avoid redundant GetEntities() calls per frame
	cachedDrawEntities []*Entity
}

// NewEbitenGame creates a new game instance with Ebiten integration.
func NewEbitenGame(screenWidth, screenHeight int) *EbitenGame {
	return NewEbitenGameWithLogger(screenWidth, screenHeight, nil)
}

// NewEbitenGameWithLogger creates a new game instance with a logger.
func NewEbitenGameWithLogger(screenWidth, screenHeight int, logger *logrus.Logger) *EbitenGame {
	logEntry := initializeLogger(logger)
	world := NewWorldWithLogger(logger)

	coreComponents := initializeCoreComponents(world, screenWidth, screenHeight, logger, logEntry)
	uiComponents := initializeUIComponents(world, screenWidth, screenHeight, coreComponents.settingsManager)

	game := buildGameInstance(screenWidth, screenHeight, world, logEntry, coreComponents, uiComponents)
	setupGameCallbacks(game, logEntry)

	if logEntry != nil {
		logEntry.WithFields(logrus.Fields{
			"screenWidth":  screenWidth,
			"screenHeight": screenHeight,
		}).Info("game initialized")
	}

	return game
}

// initializeLogger creates a structured logger entry for the game system.
func initializeLogger(logger *logrus.Logger) *logrus.Entry {
	if logger != nil {
		return logger.WithFields(logrus.Fields{"system": "game"})
	}
	return nil
}

// coreComponents holds core game components needed during initialization.
type coreComponents struct {
	cameraSystem     *CameraSystem
	renderSystem     *EbitenRenderSystem
	hudSystem        *EbitenHUDSystem
	menuSystem       *EbitenMenuSystem
	settingsManager  *SettingsManager
	frameTimeTracker *FrameTimeTracker
	lightingSystem   *LightingSystem
	sceneBuffer      *ebiten.Image
}

// initializeCoreComponents creates all core game systems.
func initializeCoreComponents(world *World, screenWidth, screenHeight int, logger *logrus.Logger, logEntry *logrus.Entry) *coreComponents {
	cameraSystem := NewCameraSystem(screenWidth, screenHeight)
	renderSystem := NewRenderSystem(cameraSystem)
	hudSystem := NewEbitenHUDSystem(screenWidth, screenHeight)

	menuSystem, err := NewEbitenMenuSystem(world, screenWidth, screenHeight, "./saves")
	if err != nil {
		logInitializationError(logEntry, err, "menu system")
	}

	settingsManager := initializeSettingsManager(logEntry)
	frameTimeTracker := NewFrameTimeTracker(1000)

	lightingConfig := NewLightingConfig()
	// Lighting enabled by default for immersive experience (V3.0 feature)
	// LightingConfig already defaults to Enabled=true in NewLightingConfig()
	lightingSystem := NewLightingSystemWithLogger(world, lightingConfig, logger)

	return &coreComponents{
		cameraSystem:     cameraSystem,
		renderSystem:     renderSystem,
		hudSystem:        hudSystem,
		menuSystem:       menuSystem,
		settingsManager:  settingsManager,
		frameTimeTracker: frameTimeTracker,
		lightingSystem:   lightingSystem,
		sceneBuffer:      nil,
	}
}

// initializeSettingsManager creates and loads the settings manager.
func initializeSettingsManager(logEntry *logrus.Entry) *SettingsManager {
	settingsManager, err := NewSettingsManager()
	if err != nil {
		logInitializationError(logEntry, err, "settings manager, using defaults")
		return &SettingsManager{settings: DefaultSettings()}
	}

	if err := settingsManager.LoadSettings(); err != nil {
		logInitializationError(logEntry, err, "loading settings, using defaults")
	}
	return settingsManager
}

// logInitializationError logs initialization errors with consistent formatting.
func logInitializationError(logEntry *logrus.Entry, err error, context string) {
	if logEntry != nil {
		logEntry.WithError(err).Warnf("failed to initialize %s", context)
	} else {
		fmt.Fprintf(os.Stderr, "WARNING: failed to initialize %s: %v\n", context, err)
	}
}

// uiComponents holds UI-related game components.
type uiComponents struct {
	inventoryUI        *EbitenInventoryUI
	questUI            *EbitenQuestUI
	characterUI        *EbitenCharacterUI
	skillsUI           *EbitenSkillsUI
	mapUI              *EbitenMapUI
	settingsUI         *SettingsUI
	mainMenuUI         *MainMenuUI
	singlePlayerMenu   *SinglePlayerMenu
	genreSelectionMenu *GenreSelectionMenu
	multiplayerMenu    *MultiplayerMenu
	serverAddressInput *ServerAddressInput
	characterCreation  *EbitenCharacterCreation
	galleryUI          *GalleryUI         // V8.0 Gallery UI (Phase 49.4)
	housingUI          *housing.HousingUI // V8.0 Housing UI (Phase 49.1) - INTEGRATION FIX [Category B]
	guildUI            *GuildUI           // Phase 3.2 Guild UI (PLAN.md)
	prestigeUI         PrestigeUIProvider // Phase 4.2 Prestige UI (PLAN.md) - stored as interface to avoid import cycle
}

// initializeUIComponents creates all UI systems in parallel for faster startup.
// Components are grouped by dependencies and initialized concurrently.
// Expected improvement: ~20-40ms startup reduction from sequential initialization.
func initializeUIComponents(world *World, screenWidth, screenHeight int, settingsManager *SettingsManager) *uiComponents {
	ui := &uiComponents{}
	var wg sync.WaitGroup

	// Group 1: World-dependent UI components (5 components)
	wg.Add(5)
	go func() {
		defer wg.Done()
		ui.inventoryUI = NewEbitenInventoryUI(world, screenWidth, screenHeight)
	}()
	go func() {
		defer wg.Done()
		ui.questUI = NewEbitenQuestUI(world, screenWidth, screenHeight)
	}()
	go func() {
		defer wg.Done()
		ui.characterUI = NewEbitenCharacterUI(world, screenWidth, screenHeight)
	}()
	go func() {
		defer wg.Done()
		ui.skillsUI = NewEbitenSkillsUI(world, screenWidth, screenHeight)
	}()
	go func() {
		defer wg.Done()
		ui.mapUI = NewEbitenMapUI(world, screenWidth, screenHeight)
	}()

	// Group 2: Independent UI components (9 components)
	wg.Add(9)
	go func() {
		defer wg.Done()
		ui.settingsUI = NewSettingsUI(screenWidth, screenHeight, settingsManager)
	}()
	go func() {
		defer wg.Done()
		ui.mainMenuUI = NewMainMenuUI(screenWidth, screenHeight)
	}()
	go func() {
		defer wg.Done()
		ui.singlePlayerMenu = NewSinglePlayerMenu(screenWidth, screenHeight)
	}()
	go func() {
		defer wg.Done()
		ui.genreSelectionMenu = NewGenreSelectionMenu(screenWidth, screenHeight)
	}()
	go func() {
		defer wg.Done()
		ui.multiplayerMenu = NewMultiplayerMenu(screenWidth, screenHeight)
	}()
	go func() {
		defer wg.Done()
		ui.serverAddressInput = NewServerAddressInput(screenWidth, screenHeight)
	}()
	go func() {
		defer wg.Done()
		ui.characterCreation = NewCharacterCreation(screenWidth, screenHeight)
	}()
	go func() {
		defer wg.Done()
		ui.galleryUI = NewGalleryUI(screenWidth, screenHeight)
	}()
	go func() {
		defer wg.Done()
		ui.housingUI = housing.NewHousingUI(screenWidth, screenHeight)
	}()

	// Group 3: World-dependent with additional dependencies (2 components)
	wg.Add(2)
	go func() {
		defer wg.Done()
		ui.guildUI = NewGuildUI(world, nil, screenWidth, screenHeight)
	}()

	// Note: PrestigeUI initialization happens in handlers.go to avoid circular imports
	// with pkg/engine/prestige. The PrestigeUI field will be set later during wiring.

	wg.Wait()
	return ui
}

// buildGameInstance constructs the EbitenGame struct with all components.
func buildGameInstance(screenWidth, screenHeight int, world *World, logEntry *logrus.Entry, core *coreComponents, ui *uiComponents) *EbitenGame {
	return &EbitenGame{
		World:                     world,
		lastUpdateTime:            time.Now(),
		ScreenWidth:               screenWidth,
		ScreenHeight:              screenHeight,
		StateManager:              NewAppStateManager(),
		MainMenuUI:                ui.mainMenuUI,
		SinglePlayerMenu:          ui.singlePlayerMenu,
		GenreSelectionMenu:        ui.genreSelectionMenu,
		MultiplayerMenu:           ui.multiplayerMenu,
		ServerAddressInput:        ui.serverAddressInput,
		SettingsUI:                ui.settingsUI,
		SettingsManager:           core.settingsManager,
		CharacterCreation:         ui.characterCreation,
		CharacterCreationTutorial: NewCharacterCreationTutorial(),
		OnboardingManager:         NewOnboardingManager(logEntry),
		LoadingUI:                 NewLoadingUI(screenWidth, screenHeight),
		CameraSystem:              core.cameraSystem,
		RenderSystem:              core.renderSystem,
		LightingSystem:            core.lightingSystem,
		sceneBuffer:               core.sceneBuffer,
		HUDSystem:                 core.hudSystem,
		MenuSystem:                core.menuSystem,
		InventoryUI:               ui.inventoryUI,
		QuestUI:                   ui.questUI,
		CharacterUI:               ui.characterUI,
		SkillsUI:                  ui.skillsUI,
		MapUI:                     ui.mapUI,
		GalleryUI:                 ui.galleryUI,
		HousingUI:                 ui.housingUI,  // V8.0 Housing UI (Phase 49.1) - INTEGRATION FIX [Category B]
		GuildUI:                   ui.guildUI,    // Phase 3.2 Guild UI (PLAN.md)
		PrestigeUI:                ui.prestigeUI, // Phase 4.2 Prestige UI (PLAN.md)
		logger:                    logEntry,
		frameTimeTracker:          core.frameTimeTracker,
		frameCount:                0,
		profilingEnabled:          false,
	}
}

// setupGameCallbacks configures all UI callback handlers.
func setupGameCallbacks(game *EbitenGame, logEntry *logrus.Entry) {
	game.MainMenuUI.SetSelectCallback(game.handleMainMenuSelection)

	game.SinglePlayerMenu.SetNewGameCallback(game.handleSinglePlayerMenuNewGame)
	game.SinglePlayerMenu.SetLoadGameCallback(game.handleSinglePlayerMenuLoadGame)
	game.SinglePlayerMenu.SetBackCallback(game.handleSinglePlayerMenuBack)

	game.GenreSelectionMenu.SetGenreSelectCallback(game.handleGenreSelection)
	game.GenreSelectionMenu.SetBackCallback(game.handleGenreSelectionBack)

	game.MultiplayerMenu.SetJoinCallback(game.handleMultiplayerMenuJoin)
	game.MultiplayerMenu.SetHostCallback(game.handleMultiplayerMenuHost)
	game.MultiplayerMenu.SetBackCallback(game.handleMultiplayerMenuBack)

	game.ServerAddressInput.SetConnectCallback(game.handleServerAddressConnect)
	game.ServerAddressInput.SetCancelCallback(game.handleServerAddressCancel)

	game.SettingsUI.SetBackCallback(func() {
		if err := game.StateManager.TransitionTo(AppStateMainMenu); err != nil {
			if logEntry != nil {
				logEntry.WithError(err).Error("failed to return to main menu from settings")
			}
		}
	})
	game.SettingsUI.SetApplyCallback(func() {
		if err := game.ApplySettings(); err != nil {
			if logEntry != nil {
				logEntry.WithError(err).Error("failed to apply settings")
			}
		}
	})

	// Wire onboarding manager to tutorial systems (Phase 3.2)
	if game.OnboardingManager != nil {
		game.OnboardingManager.SetCreationTutorial(game.CharacterCreationTutorial)
		// TutorialSystem is set later after initializeTutorialAndHelp in cmd/client
	}
}

// SetNewGameCallback sets the callback function called when New Game is selected.
func (g *EbitenGame) SetNewGameCallback(callback func() error) {
	g.onNewGame = callback
}

// SetMultiplayerConnectCallback sets the callback function called when connecting to multiplayer.
func (g *EbitenGame) SetMultiplayerConnectCallback(callback func(serverAddr string) error) {
	g.onMultiplayerConnect = callback
}

// SetQuitToMenuCallback sets the callback function called when quitting to main menu.
func (g *EbitenGame) SetQuitToMenuCallback(callback func() error) {
	g.onQuitToMenu = callback
}

// SetTerrainLoader sets the async terrain loader for tracking world generation progress.
// The loader should be started before transitioning to AppStateLoading.
func (g *EbitenGame) SetTerrainLoader(loader interface{}) {
	g.terrainLoader = loader
}

// GetTerrainLoader returns the current terrain loader reference.
func (g *EbitenGame) GetTerrainLoader() interface{} {
	return g.terrainLoader
}

// SetTerrainLoadCompleteCallback sets the callback function called when terrain loading completes.
// The callback receives the player entity and should finalize world initialization.
func (g *EbitenGame) SetTerrainLoadCompleteCallback(callback func(*Entity) error) {
	g.terrainLoadComplete = callback
}

// SetWorldSeed sets the world generation seed and applies it to character creation defaults.
// This ensures deterministic default character naming based on the world seed.
func (g *EbitenGame) SetWorldSeed(seed int64) {
	g.worldSeed = seed
	if g.CharacterCreation != nil {
		g.CharacterCreation.SetDefaultNameFromSeed(seed)
	}
}

// GetWorldSeed returns the current world generation seed.
func (g *EbitenGame) GetWorldSeed() int64 {
	return g.worldSeed
}

// SetTutorialSystemForOnboarding wires the tutorial system to the onboarding manager.
// Phase 3.2: Call this after initializeTutorialAndHelp in cmd/client to complete the wiring.
func (g *EbitenGame) SetTutorialSystemForOnboarding(tutorialSystem *EbitenTutorialSystem) {
	g.TutorialSystem = tutorialSystem
	if g.OnboardingManager != nil {
		g.OnboardingManager.SetTutorialSystem(tutorialSystem)
	}
}

// SetContextualTutorial sets the context-sensitive tutorial manager.
// Phase 3.3: Wires ShowTutorials setting to TutorialManager from pkg/rendering/ui.
func (g *EbitenGame) SetContextualTutorial(provider ContextualTutorialProvider) {
	g.ContextualTutorial = provider

	// Apply current ShowTutorials setting immediately
	if g.SettingsManager != nil {
		settings := g.SettingsManager.GetSettings()
		if provider != nil {
			if settings.ShowTutorials {
				provider.Enable()
			} else {
				provider.Disable()
			}
		}
	}
}

// handleMainMenuSelection processes main menu option selections and triggers state transitions.
func (g *EbitenGame) handleMainMenuSelection(option MainMenuOption) {
	switch option {
	case MainMenuOptionSinglePlayer:
		g.handleSinglePlayerOption()
	case MainMenuOptionMultiPlayer:
		g.handleMultiPlayerOption()
	case MainMenuOptionSettings:
		g.handleSettingsOption()
	case MainMenuOptionQuit:
		g.handleQuitOption()
	}
}

// handleSinglePlayerOption transitions to single-player submenu.
func (g *EbitenGame) handleSinglePlayerOption() {
	if err := g.StateManager.TransitionTo(AppStateSinglePlayerMenu); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to single-player menu")
		}
		return
	}

	if g.SinglePlayerMenu != nil {
		g.SinglePlayerMenu.Show()
	}

	if g.logger != nil {
		g.logger.Info("entering single-player menu")
	}
}

// handleMultiPlayerOption transitions to multiplayer submenu.
func (g *EbitenGame) handleMultiPlayerOption() {
	if err := g.StateManager.TransitionTo(AppStateMultiPlayerMenu); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to multiplayer menu")
		}
		return
	}

	if g.MultiplayerMenu != nil {
		g.MultiplayerMenu.Show()
	}

	if g.logger != nil {
		g.logger.Info("entering multiplayer menu")
	}
}

// handleSettingsOption transitions to settings menu.
func (g *EbitenGame) handleSettingsOption() {
	if err := g.StateManager.TransitionTo(AppStateSettings); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to settings")
		}
		return
	}

	g.SettingsUI.Show()

	if g.logger != nil {
		g.logger.Info("entering settings menu")
	}
}

// handleQuitOption logs the quit action.
func (g *EbitenGame) handleQuitOption() {
	if g.logger != nil {
		g.logger.Info("quit selected")
	}
}

// handleSinglePlayerMenuNewGame handles the New Game selection from single-player menu.
func (g *EbitenGame) handleSinglePlayerMenuNewGame() {
	// Transition to genre selection
	if err := g.StateManager.TransitionTo(AppStateGenreSelection); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to genre selection")
		}
		return
	}

	// Show genre selection menu
	if g.GenreSelectionMenu != nil {
		g.GenreSelectionMenu.Show()
	}

	// Hide single-player menu
	if g.SinglePlayerMenu != nil {
		g.SinglePlayerMenu.Hide()
	}

	if g.logger != nil {
		g.logger.Info("entering genre selection for new single-player game")
	}
}

// handleGenreSelection handles genre selection from the genre selection menu.
func (g *EbitenGame) handleGenreSelection(genreID string) {
	// Store selected genre
	g.selectedGenreID = genreID

	// Transition to character creation
	if err := g.StateManager.TransitionTo(AppStateCharacterCreation); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to character creation")
		}
		return
	}

	// Reset character creation UI for new game
	g.CharacterCreation.Reset()

	// Reset character creation tutorial for new game.
	// In the new-game flow, no player entity exists yet (it is created
	// after terrain-load completion), so we always reset the tutorial here.
	// Returning players who load a saved game will have tutorial state
	// restored via ImportState on their persisted player entity.
	if g.CharacterCreationTutorial != nil {
		g.CharacterCreationTutorial.Reset()
	}

	// Set default name from world seed so user has a starting name
	g.CharacterCreation.SetDefaultNameFromSeed(g.worldSeed)

	g.isMultiplayerMode = false // Single-player mode

	// Hide genre selection menu
	if g.GenreSelectionMenu != nil {
		g.GenreSelectionMenu.Hide()
	}

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"genre": genreID,
		}).Info("genre selected, entering character creation")
	}
}

// handleGenreSelectionBack handles the Back selection from genre selection menu.
func (g *EbitenGame) handleGenreSelectionBack() {
	// Transition back to single-player menu
	if err := g.StateManager.TransitionTo(AppStateSinglePlayerMenu); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition back to single-player menu")
		}
		return
	}

	// Hide genre selection menu
	if g.GenreSelectionMenu != nil {
		g.GenreSelectionMenu.Hide()
	}

	// Show single-player menu
	if g.SinglePlayerMenu != nil {
		g.SinglePlayerMenu.Show()
	}

	if g.logger != nil {
		g.logger.Info("returning to single-player menu from genre selection")
	}
}

// handleSinglePlayerMenuLoadGame handles the Load Game selection.
// Loads the most recent saved game using the save/load system.
func (g *EbitenGame) handleSinglePlayerMenuLoadGame() {
	if g.logger != nil {
		g.logger.Info("load game selected - attempting to load recent save")
	}

	// Note: Quick load (F9) is the primary load mechanism during gameplay.
	// Menu-driven load is available but deferred for future enhancement (load save browser UI).
	// For now, users should use F9 (quick load) during gameplay.
	if g.logger != nil {
		g.logger.Info("menu-driven load game UI deferred - use F9 quick load during gameplay")
	}
}

// handleSinglePlayerMenuBack handles the Back selection from single-player menu.
func (g *EbitenGame) handleSinglePlayerMenuBack() {
	// Transition back to main menu
	if err := g.StateManager.TransitionTo(AppStateMainMenu); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition back to main menu")
		}
		return
	}

	// Hide single-player menu
	if g.SinglePlayerMenu != nil {
		g.SinglePlayerMenu.Hide()
	}

	if g.logger != nil {
		g.logger.Info("returning to main menu from single-player menu")
	}
}

// handleMultiplayerMenuJoin handles the Join Server selection from multiplayer menu.
func (g *EbitenGame) handleMultiplayerMenuJoin() {
	// Transition to server address input
	if err := g.StateManager.TransitionTo(AppStateServerAddressInput); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to server address input")
		}
		return
	}

	// Show server address input
	if g.ServerAddressInput != nil {
		g.ServerAddressInput.Show()
	}

	// Hide multiplayer menu
	if g.MultiplayerMenu != nil {
		g.MultiplayerMenu.Hide()
	}

	if g.logger != nil {
		g.logger.Info("showing server address input for join")
	}
}

// handleMultiplayerMenuHost handles the Host Game selection from multiplayer menu.
func (g *EbitenGame) handleMultiplayerMenuHost() {
	g.logHostSelected()
	g.pendingServerAddress = "localhost:8080"

	if err := g.transitionToGameplayFromHost(); err != nil {
		return
	}

	g.hideMultiplayerMenu()

	if err := g.connectToHostedServer(); err != nil {
		g.revertToMultiplayerMenu()
		return
	}

	g.logHostConnection()
}

// logHostSelected logs host game selection.
func (g *EbitenGame) logHostSelected() {
	if g.logger != nil {
		g.logger.Info("host game selected - menu-driven host deferred")
	}
}

// transitionToGameplayFromHost transitions from menu to gameplay state.
func (g *EbitenGame) transitionToGameplayFromHost() error {
	if err := g.StateManager.TransitionTo(AppStateGameplay); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to gameplay")
		}
		return err
	}
	return nil
}

// hideMultiplayerMenu hides the multiplayer menu UI.
func (g *EbitenGame) hideMultiplayerMenu() {
	if g.MultiplayerMenu != nil {
		g.MultiplayerMenu.Hide()
	}
}

// connectToHostedServer attempts connection to the hosted server.
func (g *EbitenGame) connectToHostedServer() error {
	if g.onMultiplayerConnect == nil {
		return nil
	}
	if err := g.onMultiplayerConnect(g.pendingServerAddress); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to connect to hosted server")
		}
		return err
	}
	return nil
}

// revertToMultiplayerMenu transitions back to multiplayer menu on connection failure.
func (g *EbitenGame) revertToMultiplayerMenu() {
	_ = g.StateManager.TransitionTo(AppStateMultiPlayerMenu)
	if g.MultiplayerMenu != nil {
		g.MultiplayerMenu.Show()
	}
}

// logHostConnection logs successful connection to hosted server.
func (g *EbitenGame) logHostConnection() {
	if g.logger != nil {
		g.logger.WithField("address", g.pendingServerAddress).Info("connected to hosted server")
	}
}

// handleMultiplayerMenuBack handles the Back selection from multiplayer menu.
func (g *EbitenGame) handleMultiplayerMenuBack() {
	// Transition back to main menu
	if err := g.StateManager.TransitionTo(AppStateMainMenu); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition back to main menu")
		}
		return
	}

	// Hide multiplayer menu
	if g.MultiplayerMenu != nil {
		g.MultiplayerMenu.Hide()
	}

	if g.logger != nil {
		g.logger.Info("returning to main menu from multiplayer menu")
	}
}

// handleServerAddressConnect handles connecting to the entered server address.
func (g *EbitenGame) handleServerAddressConnect(address string) {
	g.logServerConnection(address)
	g.pendingServerAddress = address

	if err := g.transitionToGameplayFromServer(); err != nil {
		return
	}

	g.hideServerAddressInput()

	if err := g.connectToServer(address); err != nil {
		g.revertToServerAddressInput()
		return
	}

	g.logSuccessfulConnection(address)
}

// logServerConnection logs the connection attempt to server.
func (g *EbitenGame) logServerConnection(address string) {
	if g.logger != nil {
		g.logger.WithField("address", address).Info("connecting to server")
	}
}

// transitionToGameplayFromServer transitions from server input to gameplay state.
func (g *EbitenGame) transitionToGameplayFromServer() error {
	if err := g.StateManager.TransitionTo(AppStateGameplay); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to gameplay")
		}
		return err
	}
	return nil
}

// hideServerAddressInput hides the server address input UI.
func (g *EbitenGame) hideServerAddressInput() {
	if g.ServerAddressInput != nil {
		g.ServerAddressInput.Hide()
	}
}

// connectToServer attempts connection to the specified server address.
func (g *EbitenGame) connectToServer(address string) error {
	if g.onMultiplayerConnect == nil {
		return nil
	}
	if err := g.onMultiplayerConnect(address); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to connect to server")
		}
		return err
	}
	return nil
}

// revertToServerAddressInput transitions back to server input on connection failure.
func (g *EbitenGame) revertToServerAddressInput() {
	_ = g.StateManager.TransitionTo(AppStateServerAddressInput)
	if g.ServerAddressInput != nil {
		g.ServerAddressInput.Show()
	}
}

// logSuccessfulConnection logs successful connection to server.
func (g *EbitenGame) logSuccessfulConnection(address string) {
	if g.logger != nil {
		g.logger.WithField("address", address).Info("connected to server")
	}
}

// handleServerAddressCancel handles canceling server address input.
func (g *EbitenGame) handleServerAddressCancel() {
	// Transition back to multiplayer menu
	if err := g.StateManager.TransitionTo(AppStateMultiPlayerMenu); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition back to multiplayer menu")
		}
		return
	}

	// Hide server address input
	if g.ServerAddressInput != nil {
		g.ServerAddressInput.Hide()
	}

	// Show multiplayer menu
	if g.MultiplayerMenu != nil {
		g.MultiplayerMenu.Show()
	}

	if g.logger != nil {
		g.logger.Info("cancelled server address input, returning to multiplayer menu")
	}
}

// IsInMainMenu returns true if currently displaying the main menu.
func (g *EbitenGame) IsInMainMenu() bool {
	return g.StateManager.IsInMenu()
}

// trackFramePerformance records frame timing metrics and logs performance statistics.
func (g *EbitenGame) trackFramePerformance(frameStart time.Time) {
	if !g.profilingEnabled || g.frameTimeTracker == nil {
		return
	}

	g.frameTimeTracker.RecordFrame(time.Since(frameStart))
	g.frameCount++

	if g.frameCount%300 == 0 && g.logger != nil {
		g.logFrameStats()
	}
}

// logFrameStats logs current frame time statistics with stuttering detection.
func (g *EbitenGame) logFrameStats() {
	stats := g.frameTimeTracker.GetStats()
	fields := logrus.Fields{
		"avg_ms":      stats.Average.Milliseconds(),
		"min_ms":      stats.Min.Milliseconds(),
		"max_ms":      stats.Max.Milliseconds(),
		"1pct_low_ms": stats.Percentile1.Milliseconds(),
		"avg_fps":     stats.GetFPS(),
		"worst_fps":   stats.GetWorstFPS(),
		"samples":     stats.SampleCount,
	}

	if stats.IsStuttering() {
		fields["stuttering"] = true
		g.logger.WithFields(fields).Warn("frame time stuttering detected")
	} else {
		g.logger.WithFields(fields).Info("frame time stats")
	}
}

// calculateDeltaTime returns the fixed timestep matching Ebiten's TPS rate.
// Using a constant fixed timestep instead of wall-clock time.Now() deltas eliminates
// jitter caused by OS scheduling variance, GC pauses, and Ebiten catch-up ticks
// (where Update() is called multiple times in succession with near-zero wall-clock deltas).
func (g *EbitenGame) calculateDeltaTime() float64 {
	tps := ebiten.TPS()
	if tps <= 0 {
		return 1.0 / 60.0 // Fallback to 60 TPS default
	}
	return 1.0 / float64(tps)
}

// updateMenuState handles updates for all menu-related application states.
func (g *EbitenGame) updateMenuState() (handled bool) {
	switch g.StateManager.CurrentState() {
	case AppStateMainMenu:
		g.MainMenuUI.Update()
		return true
	case AppStateSinglePlayerMenu:
		if g.SinglePlayerMenu != nil {
			g.SinglePlayerMenu.Update()
		}
		return true
	case AppStateGenreSelection:
		if g.GenreSelectionMenu != nil {
			g.GenreSelectionMenu.Update()
		}
		return true
	case AppStateMultiPlayerMenu:
		if g.MultiplayerMenu != nil {
			g.MultiplayerMenu.Update()
		}
		return true
	case AppStateServerAddressInput:
		if g.ServerAddressInput != nil {
			g.ServerAddressInput.Update()
		}
		return true
	case AppStateSettings:
		g.SettingsUI.Update()
		return true
	}

	if g.StateManager.IsInMenu() {
		g.MainMenuUI.Update()
		return true
	}
	return false
}

// handleCharacterCreation processes character creation updates and transitions to gameplay.
func (g *EbitenGame) handleCharacterCreation() error {
	// Use the same fixed timestep as the rest of the game to keep tutorial timing consistent.
	deltaTime := g.calculateDeltaTime()

	// Update the character creation tutorial overlay alongside character creation.
	// The tutorial is updated FIRST so it can consume input (e.g., ENTER to dismiss
	// the welcome overlay) before the underlying character creation processes it.
	if g.CharacterCreationTutorial != nil && g.CharacterCreationTutorial.IsActive() {
		g.CharacterCreationTutorial.Update(int(g.CharacterCreation.currentStep), deltaTime)
	}

	// If the tutorial consumed input this frame (e.g., welcome overlay dismissal),
	// skip the character creation update to prevent double-advancement.
	if g.CharacterCreationTutorial != nil && g.CharacterCreationTutorial.InputConsumed {
		return nil
	}

	completed := g.CharacterCreation.Update()
	if !completed {
		return nil
	}

	if g.logger != nil {
		g.logger.Info("character creation completed, transitioning to gameplay")
	}

	// Mark the character creation tutorial as complete
	if g.CharacterCreationTutorial != nil && !g.CharacterCreationTutorial.IsComplete() {
		g.CharacterCreationTutorial.CompleteTutorial()
	}

	charData := g.CharacterCreation.GetCharacterData()

	// Transition onboarding to in-game tutorial phase (Phase 3.2)
	// Store player class for class-aware tutorial content
	if g.OnboardingManager != nil {
		g.OnboardingManager.SetPlayerClass(charData.Class)
		g.OnboardingManager.TransitionToInGameTutorial()
	}

	g.pendingCharData = &charData
	g.CharacterCreation.Cleanup()

	// If terrain was already loaded (default startup path: Loading → CharacterCreation),
	// finishLoadingToGameplay transitions to AppStateGameplay and fires the
	// terrainLoadComplete callback that creates the player entity.
	if g.terrainLoader != nil || g.terrainLoadComplete != nil {
		return g.finishLoadingToGameplay()
	}

	// Menu-driven path: terrain hasn't loaded yet, transition directly.
	if err := g.StateManager.TransitionTo(AppStateGameplay); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to gameplay after character creation")
		}
		return err
	}

	return g.triggerGameStartCallback(charData)
}

// triggerGameStartCallback invokes the appropriate callback for multiplayer or single-player mode.
func (g *EbitenGame) triggerGameStartCallback(charData CharacterData) error {
	if g.isMultiplayerMode {
		return g.startMultiplayerGame(charData)
	}
	return g.startSinglePlayerGame(charData)
}

// handleLoadingState updates the loading screen during async terrain generation.
// Polls terrain loader progress and transitions to gameplay when complete.
func (g *EbitenGame) handleLoadingState() error {
	loader, err := g.validateTerrainLoader()
	if err != nil {
		return err
	}
	if loader == nil {
		return nil
	}

	if err := g.updateLoadingProgress(loader); err != nil {
		return err
	}

	if loader.IsDone() {
		return g.handleLoadingComplete()
	}

	return nil
}

// asyncLoader defines the interface for async terrain loading to avoid circular imports.
type asyncLoader interface {
	GetProgress() (float64, error)
	IsDone() bool
}

// validateTerrainLoader checks if the terrain loader is properly configured.
func (g *EbitenGame) validateTerrainLoader() (asyncLoader, error) {
	if g.terrainLoader == nil {
		if g.logger != nil {
			g.logger.Error("in loading state but no terrain loader set")
		}
		_ = g.StateManager.TransitionTo(AppStateMainMenu)
		return nil, nil
	}

	loader, ok := g.terrainLoader.(asyncLoader)
	if !ok {
		if g.logger != nil {
			g.logger.Error("terrain loader does not implement required methods")
		}
		_ = g.StateManager.TransitionTo(AppStateMainMenu)
		return nil, nil
	}

	return loader, nil
}

// updateLoadingProgress retrieves and displays current loading progress.
func (g *EbitenGame) updateLoadingProgress(loader asyncLoader) error {
	progress, err := loader.GetProgress()
	if err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("terrain generation failed")
		}
		_ = g.StateManager.TransitionTo(AppStateMainMenu)
		return err
	}

	if g.LoadingUI != nil {
		g.LoadingUI.SetProgress(progress)
	}

	return nil
}

// handleLoadingComplete finalizes terrain loading.
// If the player hasn't been through character creation yet, transitions to
// AppStateCharacterCreation so they can pick name/class/equipment.
// Otherwise goes directly to gameplay (e.g. when loading a saved game).
func (g *EbitenGame) handleLoadingComplete() error {
	// If character creation was already completed (e.g. via the menu new-game
	// flow, or when loading a save), skip straight to gameplay.
	if g.pendingCharData != nil || g.PlayerEntity != nil {
		return g.finishLoadingToGameplay()
	}

	// Terrain is ready — now let the player create their character.
	if g.logger != nil {
		g.logger.Info("terrain loading complete, entering character creation")
	}

	if err := g.StateManager.TransitionTo(AppStateCharacterCreation); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to character creation after terrain load")
		}
		return err
	}

	// Reset character creation UI for a fresh start.
	g.CharacterCreation.Reset()
	if g.CharacterCreationTutorial != nil {
		g.CharacterCreationTutorial.Reset()
	}
	g.CharacterCreation.SetDefaultNameFromSeed(g.worldSeed)
	g.isMultiplayerMode = false

	return nil
}

// finishLoadingToGameplay transitions from loading (or character creation) to
// active gameplay and fires the terrainLoadComplete callback that creates the
// player entity and sets up the world.
func (g *EbitenGame) finishLoadingToGameplay() error {
	if g.logger != nil {
		g.logger.Info("transitioning to gameplay")
	}

	if err := g.StateManager.TransitionTo(AppStateGameplay); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("failed to transition to gameplay")
		}
		return err
	}

	if g.terrainLoadComplete != nil {
		if err := g.terrainLoadComplete(g.PlayerEntity); err != nil {
			if g.logger != nil {
				g.logger.WithError(err).Error("terrain load complete callback failed")
			}
			return err
		}
	}

	// Clear terrain loader after callback completes (callback may need to access it)
	g.terrainLoader = nil

	return nil
}

// startMultiplayerGame connects to multiplayer server with character data.
func (g *EbitenGame) startMultiplayerGame(charData CharacterData) error {
	if g.onMultiplayerConnect != nil {
		if err := g.onMultiplayerConnect(""); err != nil {
			if g.logger != nil {
				g.logger.WithError(err).Error("multiplayer connect callback failed")
			}
			_ = g.StateManager.TransitionTo(AppStateMainMenu)
			g.pendingCharData = nil
			return err
		}
	}

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"name":  charData.Name,
			"class": charData.Class.String(),
			"mode":  "multiplayer",
		}).Info("character created, connecting to server")
	}
	return nil
}

// startSinglePlayerGame starts a new single-player game session.
func (g *EbitenGame) startSinglePlayerGame(charData CharacterData) error {
	if g.onNewGame != nil {
		if err := g.onNewGame(); err != nil {
			if g.logger != nil {
				g.logger.WithError(err).Error("new game callback failed")
			}
			_ = g.StateManager.TransitionTo(AppStateMainMenu)
			g.pendingCharData = nil
			return err
		}
	}

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"name":  charData.Name,
			"class": charData.Class.String(),
			"mode":  "single-player",
		}).Info("character created, starting game")
	}
	return nil
}

// updateGameplayUI updates all UI systems during active gameplay.
func (g *EbitenGame) updateGameplayUI(deltaTime float64) {
	g.updateCoreUIScreens(deltaTime)
	g.updateCommerceUIScreens(deltaTime)
	g.updatePhaseUIScreens(deltaTime)
	g.updateVirtualControlsVisibility()
}

// updateCoreUIScreens updates the primary UI screens.
func (g *EbitenGame) updateCoreUIScreens(deltaTime float64) {
	g.InventoryUI.Update(nil, deltaTime)
	g.QuestUI.Update(nil, deltaTime)
	g.CharacterUI.Update(nil, deltaTime)
	g.SkillsUI.Update(nil, deltaTime)
	g.MapUI.Update(nil, deltaTime)

	// Note: TutorialSystem.Update() is called via World.Update() since it is
	// registered with World.AddSystem(). Calling it here as well would cause
	// double-evaluation of step conditions per frame, leading to steps being
	// skipped or appearing misordered.
}

// updateCommerceUIScreens updates shop, crafting, and trade UIs.
func (g *EbitenGame) updateCommerceUIScreens(deltaTime float64) {
	if g.ShopUI != nil {
		g.ShopUI.Update(g.World.GetEntities(), deltaTime)
	}

	if g.CraftingUI != nil {
		g.CraftingUI.Update(nil, deltaTime)
	}

	if g.TradeUI != nil {
		g.TradeUI.Update(deltaTime)
	}
}

// updatePhaseUIScreens updates all phase-specific UI screens.
func (g *EbitenGame) updatePhaseUIScreens(deltaTime float64) {
	if g.HousingUI != nil {
		g.HousingUI.Update()
	}

	if g.GalleryUI != nil {
		g.GalleryUI.Update()
	}

	if g.GuildUI != nil {
		g.GuildUI.Update()
	}

	if g.AdvancedClassUI != nil {
		g.AdvancedClassUI.Update()
	}

	if g.TerritoryUI != nil {
		g.TerritoryUI.Update()
	}

	if g.StoryChoiceUI != nil {
		g.StoryChoiceUI.Update(deltaTime)
	}

	if g.DialogUI != nil {
		g.DialogUI.Update()
	}
}

// updateVirtualControlsVisibility manages virtual control visibility based on UI state.
// Hides virtual controls (D-pad, action buttons) when any menu is open to prevent overlay.
// Shows virtual controls during active gameplay when no blocking UI is visible.
func (g *EbitenGame) updateVirtualControlsVisibility() {
	// Check if we're in menu states (main menu, settings, character creation, etc.)
	inMenuState := g.StateManager.CurrentState() != AppStateGameplay

	// Check if any gameplay UI is open
	// BUG FIX: Phase 3.7 - MailboxUI missing from virtual controls visibility check
	// Resolution: Added MailboxUI.IsOpen() check to hide virtual controls when mailbox is open
	// Phase 4.3: Added TerritoryUI.IsVisible() check for territory control UI
	anyUIOpen := g.InventoryUI.IsVisible() ||
		g.QuestUI.IsVisible() ||
		g.CharacterUI.IsVisible() ||
		g.SkillsUI.IsVisible() ||
		g.MapUI.IsFullScreen() ||
		(g.ShopUI != nil && g.ShopUI.IsVisible()) ||
		(g.CraftingUI != nil && g.CraftingUI.IsVisible()) ||
		(g.TradeUI != nil && g.TradeUI.IsVisible()) ||
		(g.MailboxUI != nil && g.MailboxUI.IsOpen()) ||
		(g.GuildUI != nil && g.GuildUI.IsVisible()) ||
		(g.AdvancedClassUI != nil && g.AdvancedClassUI.IsVisible()) ||
		(g.TerritoryUI != nil && g.TerritoryUI.IsVisible()) ||
		(g.StoryChoiceUI != nil && g.StoryChoiceUI.IsVisible()) ||
		(g.DialogUI != nil && g.DialogUI.IsVisible()) ||
		(g.MenuSystem != nil && g.MenuSystem.IsActive())

	// Virtual controls should be hidden if:
	// 1. In a menu state (not gameplay)
	// 2. Any gameplay UI is open
	shouldHide := inMenuState || anyUIOpen

	// Update virtual controls visibility for all input systems
	for _, system := range g.World.GetSystems() {
		if inputSys, ok := system.(*InputSystem); ok {
			inputSys.SetVirtualControlsVisible(!shouldHide)
			break
		}
	}
}

// shouldUpdateWorld checks if world updates should proceed based on UI visibility.
// BUG FIX: Phase 3.7 - MailboxUI missing from world update check
// Resolution: Added MailboxUI.IsOpen() check to pause world when mailbox is open
// Phase 4.3: Added TerritoryUI.IsVisible() check to pause world during territory management
// Phase 6.2: Added DialogUI.IsVisible() check to pause world during NPC conversations
func (g *EbitenGame) shouldUpdateWorld() bool {
	return !g.InventoryUI.IsVisible() &&
		!g.QuestUI.IsVisible() &&
		!g.CharacterUI.IsVisible() &&
		!g.SkillsUI.IsVisible() &&
		!g.MapUI.IsFullScreen() &&
		(g.ShopUI == nil || !g.ShopUI.IsVisible()) &&
		(g.CraftingUI == nil || !g.CraftingUI.IsVisible()) &&
		(g.TradeUI == nil || !g.TradeUI.IsVisible()) &&
		(g.MailboxUI == nil || !g.MailboxUI.IsOpen()) &&
		(g.GuildUI == nil || !g.GuildUI.IsVisible()) &&
		(g.AdvancedClassUI == nil || !g.AdvancedClassUI.IsVisible()) &&
		(g.TerritoryUI == nil || !g.TerritoryUI.IsVisible()) &&
		(g.StoryChoiceUI == nil || !g.StoryChoiceUI.IsVisible()) &&
		(g.DialogUI == nil || !g.DialogUI.IsVisible())
}

func (g *EbitenGame) Update() error {
	frameStart := time.Now()
	defer func() { g.trackFramePerformance(frameStart) }()

	deltaTime := g.calculateDeltaTime()

	// Handle loading state BEFORE other state checks
	if g.StateManager.CurrentState() == AppStateLoading {
		return g.handleLoadingState()
	}

	// Handle character creation BEFORE updateMenuState to prevent main menu from consuming input
	if g.StateManager.CurrentState() == AppStateCharacterCreation {
		return g.handleCharacterCreation()
	}

	if g.updateMenuState() {
		return nil
	}

	if g.MenuSystem != nil && g.MenuSystem.IsActive() {
		g.Paused = true
		g.MenuSystem.Update(g.World.GetEntities(), deltaTime)
		return nil
	}

	if g.Paused {
		return nil
	}

	g.updateGameplayUI(deltaTime)

	// Update camera BEFORE world systems so that camera and entities are in sync
	// for the same tick. This eliminates the one-frame lag where entities move
	// but the camera hasn't caught up yet.
	g.CameraSystem.Update(g.World.GetEntities(), deltaTime)

	if g.shouldUpdateWorld() {
		g.World.Update(deltaTime)
	}

	// Cache entity list once per Update for Draw() to reuse, avoiding redundant calls
	g.cachedDrawEntities = g.World.GetEntities()

	// Record wall-clock timestamp for render interpolation alpha calculation in Draw()
	g.lastUpdateTime = time.Now()
	return nil
}

// Draw implements ebiten.Game interface. Called every frame.
// Calculates render interpolation alpha to smoothly interpolate entity positions
// between the previous and current simulation tick, eliminating visual jitter
// on monitors with refresh rates higher than the TPS.
func (g *EbitenGame) Draw(screen *ebiten.Image) {
	if g.drawMenuState(screen) {
		return
	}

	// Calculate render interpolation alpha: fraction of time elapsed since last Update().
	// On a 60 TPS / 144 Hz setup, Draw() is called ~2.4x per Update(). Alpha smoothly
	// advances from 0.0 to 1.0 between ticks so entity positions are interpolated.
	tps := ebiten.TPS()
	if tps > 0 {
		elapsed := time.Since(g.lastUpdateTime).Seconds()
		tickDuration := 1.0 / float64(tps)
		g.renderAlpha = elapsed / tickDuration
		if g.renderAlpha > 1.0 {
			g.renderAlpha = 1.0
		}
		if g.renderAlpha < 0.0 {
			g.renderAlpha = 0.0
		}
	} else {
		g.renderAlpha = 1.0
	}

	// Pass interpolation alpha to render system
	g.RenderSystem.SetRenderAlpha(g.renderAlpha)

	g.drawGameplayScene(screen)
	g.drawOverlays(screen)
	g.drawVirtualControls(screen)
}

// drawMenuState renders the appropriate menu based on current application state.
// Returns true if a menu was drawn, false if gameplay should be rendered.
func (g *EbitenGame) drawMenuState(screen *ebiten.Image) bool {
	switch g.StateManager.CurrentState() {
	case AppStateMainMenu:
		g.MainMenuUI.Draw(screen)
		return true
	case AppStateSinglePlayerMenu:
		if g.SinglePlayerMenu != nil {
			g.SinglePlayerMenu.Draw(screen)
		}
		return true
	case AppStateGenreSelection:
		if g.GenreSelectionMenu != nil {
			g.GenreSelectionMenu.Draw(screen)
		}
		return true
	case AppStateMultiPlayerMenu:
		if g.MultiplayerMenu != nil {
			g.MultiplayerMenu.Draw(screen)
		}
		return true
	case AppStateLoading:
		if g.LoadingUI != nil {
			g.LoadingUI.Draw(screen)
		}
		return true
	case AppStateServerAddressInput:
		if g.ServerAddressInput != nil {
			g.ServerAddressInput.Draw(screen)
		}
		return true
	case AppStateSettings:
		g.SettingsUI.Draw(screen)
		return true
	case AppStateCharacterCreation:
		g.CharacterCreation.Draw(screen)
		// Draw tutorial overlay on top of character creation UI
		if g.CharacterCreationTutorial != nil {
			g.CharacterCreationTutorial.Draw(screen)
		}
		return true
	}

	if g.StateManager.IsInMenu() {
		g.MainMenuUI.Draw(screen)
		return true
	}

	return false
}

// drawGameplayScene renders the main game scene with terrain, entities, and visual effects.
func (g *EbitenGame) drawGameplayScene(screen *ebiten.Image) {
	if g.LightingSystem != nil && g.LightingSystem.IsEnabled() {
		g.drawLitScene(screen)
	} else {
		g.drawStandardScene(screen)
	}
}

// drawLitScene renders the game scene with dynamic lighting and post-processing pipeline.
func (g *EbitenGame) drawLitScene(screen *ebiten.Image) {
	if g.sceneBuffer == nil {
		g.sceneBuffer = ebiten.NewImage(g.ScreenWidth, g.ScreenHeight)
	}

	g.sceneBuffer.Clear()

	if g.TerrainRenderSystem != nil {
		g.TerrainRenderSystem.Draw(g.sceneBuffer, g.CameraSystem)
	}

	// Use cached entity list from Update() to avoid redundant GetEntities() calls
	entities := g.cachedDrawEntities
	if entities == nil {
		entities = g.World.GetEntities()
	}

	g.RenderSystem.Draw(g.sceneBuffer, entities)

	if g.CameraSystem != nil {
		camX, camY := g.CameraSystem.GetPosition()
		g.LightingSystem.SetViewport(camX, camY, g.ScreenWidth, g.ScreenHeight)
	}

	// Apply lighting to cached buffer (Performance Audit fix: reuse litBuffer)
	if g.litBuffer == nil || g.litBuffer.Bounds().Dx() != g.ScreenWidth || g.litBuffer.Bounds().Dy() != g.ScreenHeight {
		if g.litBuffer != nil {
			g.litBuffer.Dispose()
		}
		g.litBuffer = ebiten.NewImage(g.ScreenWidth, g.ScreenHeight)
	}
	g.litBuffer.Clear()
	g.LightingSystem.ApplyLighting(g.litBuffer, g.sceneBuffer, entities)

	// Apply post-processing if enabled (Phase 5.3)
	finalImage := g.litBuffer
	if g.PostProcessor != nil && g.PostProcessor.IsEnabled() {
		finalImage = g.PostProcessor.Apply(g.litBuffer)
	}

	// Draw final image to screen
	screen.DrawImage(finalImage, nil)
}

// drawStandardScene renders the game scene without lighting effects, with optional post-processing.
func (g *EbitenGame) drawStandardScene(screen *ebiten.Image) {
	// Use cached entity list from Update() to avoid redundant GetEntities() calls
	entities := g.cachedDrawEntities
	if entities == nil {
		entities = g.World.GetEntities()
	}

	// Render to buffer if post-processing is enabled
	if g.PostProcessor != nil && g.PostProcessor.IsEnabled() {
		if g.sceneBuffer == nil {
			g.sceneBuffer = ebiten.NewImage(g.ScreenWidth, g.ScreenHeight)
		}

		g.sceneBuffer.Clear()

		if g.TerrainRenderSystem != nil {
			g.TerrainRenderSystem.Draw(g.sceneBuffer, g.CameraSystem)
		}

		g.RenderSystem.Draw(g.sceneBuffer, entities)

		// Apply post-processing
		finalImage := g.PostProcessor.Apply(g.sceneBuffer)
		screen.DrawImage(finalImage, nil)
	} else {
		// Direct rendering without post-processing
		if g.TerrainRenderSystem != nil {
			g.TerrainRenderSystem.Draw(screen, g.CameraSystem)
		}

		g.RenderSystem.Draw(screen, entities)
	}
}

// drawOverlays renders all UI overlays including HUD, menus, inventory, and other interfaces.
func (g *EbitenGame) drawOverlays(screen *ebiten.Image) {
	g.drawSystemOverlays(screen)
	g.drawCoreUIOverlays(screen)
	g.drawCommerceUIOverlays(screen)
	g.drawPhaseUIOverlays(screen)
}

// drawSystemOverlays draws HUD, tutorial, help, and menu systems.
func (g *EbitenGame) drawSystemOverlays(screen *ebiten.Image) {
	g.HUDSystem.Draw(screen)

	if g.TutorialSystem != nil && g.TutorialSystem.Enabled {
		g.TutorialSystem.Draw(screen)
	}

	if g.HelpSystem != nil && g.HelpSystem.Visible {
		g.HelpSystem.Draw(screen)
	}

	if g.MenuSystem != nil && g.MenuSystem.IsActive() {
		g.MenuSystem.Draw(screen)
	}
}

// drawCoreUIOverlays draws primary UI screens.
func (g *EbitenGame) drawCoreUIOverlays(screen *ebiten.Image) {
	g.InventoryUI.Draw(screen)
	g.QuestUI.Draw(screen)
	g.CharacterUI.Draw(screen)
	g.SkillsUI.Draw(screen)
	g.MapUI.Draw(screen)
}

// drawCommerceUIOverlays draws shop and crafting UIs.
func (g *EbitenGame) drawCommerceUIOverlays(screen *ebiten.Image) {
	if g.ShopUI != nil {
		g.ShopUI.Draw(screen)
	}

	if g.CraftingUI != nil {
		g.CraftingUI.Draw(screen)
	}
}

// drawPhaseUIOverlays draws all phase-specific UI screens.
func (g *EbitenGame) drawPhaseUIOverlays(screen *ebiten.Image) {
	if g.TradeUI != nil {
		g.TradeUI.Draw(screen)
	}

	if g.AdvancedClassUI != nil {
		g.AdvancedClassUI.Draw(screen)
	}

	if g.TerritoryUI != nil {
		g.TerritoryUI.Draw(screen)
	}

	if g.StoryChoiceUI != nil {
		g.StoryChoiceUI.Draw(screen)
	}

	if g.DialogUI != nil {
		g.DialogUI.Draw(screen)
	}

	g.drawMailboxUI(screen)

	if g.HousingUI != nil {
		g.HousingUI.Draw(screen)
	}

	if g.GalleryUI != nil {
		g.GalleryUI.Draw(screen)
	}

	if g.GuildUI != nil {
		g.GuildUI.Draw(screen)
	}
}

// drawMailboxUI renders the mailbox interface if open and loads mail data from player entity.
// Performance Audit V6 fix: now draws directly to screen without CPU→GPU image conversion.
func (g *EbitenGame) drawMailboxUI(screen *ebiten.Image) {
	if g.MailboxUI == nil || !g.MailboxUI.IsOpen() {
		return
	}

	if g.PlayerEntity != nil {
		if mailComp, ok := g.PlayerEntity.GetComponent("mail"); ok {
			g.MailboxUI.LoadFromMailComponent(mailComp.(*MailComponent))
		}
	}

	// Draw directly to screen - eliminates ebiten.NewImageFromImage conversion spike
	g.MailboxUI.Draw(screen)
}

// drawVirtualControls renders mobile touch controls on top of all other elements.
// Uses a cached InputSystem reference to avoid linear scanning all 44+ systems every frame.
func (g *EbitenGame) drawVirtualControls(screen *ebiten.Image) {
	// Lazily cache the InputSystem reference on first call
	if g.cachedInputSystem == nil {
		for _, system := range g.World.GetSystems() {
			if inputSys, ok := system.(*InputSystem); ok {
				g.cachedInputSystem = inputSys
				break
			}
		}
	}
	if g.cachedInputSystem != nil {
		g.cachedInputSystem.DrawVirtualControls(screen)
	}
}

// screenDimensionConsumer is implemented by systems that maintain cached copies
// of the screen dimensions and need to be kept in sync when the window is resized.
type screenDimensionConsumer interface {
	OnScreenResize(width, height int)
}

// Layout implements ebiten.Game interface. Returns the game's screen size.
func (g *EbitenGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth > 0 && outsideHeight > 0 {
		// Only propagate when the dimensions actually change to avoid redundant work.
		if outsideWidth != g.ScreenWidth || outsideHeight != g.ScreenHeight {
			g.ScreenWidth = outsideWidth
			g.ScreenHeight = outsideHeight

			// Keep sceneBuffer in sync with the current screen size. LightingSystem.ApplyLighting
			// uses the rendered scene size (sceneBuffer.Size()) to build its lighting buffer,
			// so a stale buffer size after a resize would cause clipping/misalignment.
			if g.sceneBuffer != nil {
				g.sceneBuffer.Dispose()
				g.sceneBuffer = ebiten.NewImage(g.ScreenWidth, g.ScreenHeight)
			}

			g.propagateScreenResize()
		}
	}
	return g.ScreenWidth, g.ScreenHeight
}

// propagateScreenResize updates any systems that cache the screen dimensions so
// they remain in sync with EbitenGame.ScreenWidth/ScreenHeight.
func (g *EbitenGame) propagateScreenResize() {
	if g.CameraSystem != nil {
		g.CameraSystem.ScreenWidth = g.ScreenWidth
		g.CameraSystem.ScreenHeight = g.ScreenHeight

		// Recalculate camera bounds for the new screen dimensions so the
		// viewport stays clamped to terrain edges after orientation changes.
		g.CameraSystem.RecalculateBounds()
	}

	if g.World == nil {
		return
	}

	systems := g.World.GetSystems()
	for _, system := range systems {
		if consumer, ok := system.(screenDimensionConsumer); ok {
			consumer.OnScreenResize(g.ScreenWidth, g.ScreenHeight)
		}
	}
}

// SetPlayerEntity sets the player entity for the game and UI systems.
// This should be called after creating the player entity.
func (g *EbitenGame) SetPlayerEntity(entity *Entity) {
	g.PlayerEntity = entity
	g.InventoryUI.SetPlayerEntity(entity)
	g.QuestUI.SetPlayerEntity(entity)
	g.CharacterUI.SetPlayerEntity(entity)
	g.SkillsUI.SetPlayerEntity(entity)
	g.MapUI.SetPlayerEntity(entity)

	// Set player for shop UI (if initialized)
	if g.ShopUI != nil {
		g.ShopUI.SetPlayerEntity(entity)
	}

	// Set player for crafting UI (if initialized)
	if g.CraftingUI != nil {
		g.CraftingUI.SetPlayerEntity(entity)
	}

	// Phase 3.3 (PLAN.md): Set player for trade UI (if initialized)
	if g.TradeUI != nil {
		g.TradeUI.SetPlayerEntity(entity)
	}

	// Phase 3.2: Set player for onboarding manager to enable save/load persistence
	if g.OnboardingManager != nil {
		g.OnboardingManager.SetPlayerEntity(entity)
	}

	// Set player for render system aim indicator (drawn below player sprite)
	if g.RenderSystem != nil {
		g.RenderSystem.SetAimPlayerEntity(entity)
	}
}

// SetAudioManager sets the audio manager for the game.
// This should be called before ApplySettings() to enable volume control.
func (g *EbitenGame) SetAudioManager(audioManager *AudioManager) {
	g.AudioManager = audioManager

	// Apply current settings to audio if settings manager exists
	if g.SettingsManager != nil {
		_ = g.ApplySettings() // Ignore error, just apply what we can
	}
}

// SetInventorySystem connects the inventory system to the inventory UI for item actions.
func (g *EbitenGame) SetInventorySystem(system *InventorySystem) {
	g.InventoryUI.SetInventorySystem(system)
}

// GetPendingCharacterData returns the character data from character creation (if any).
// Returns nil if no character data is pending. Clears the pending data after retrieval.
func (g *EbitenGame) GetPendingCharacterData() *CharacterData {
	data := g.pendingCharData
	g.pendingCharData = nil // Clear after retrieval
	return data
}

// GetSelectedGenreID returns the selected genre ID and clears it.
func (g *EbitenGame) GetSelectedGenreID() string {
	genreID := g.selectedGenreID
	g.selectedGenreID = "" // Clear after retrieval
	return genreID
}

// SetupInputCallbacks connects the input system callbacks to the UI systems.
// This should be called after the InputSystem is added to the world.
// GAP-014 REPAIR: Accept objective tracker for quest progress tracking
// H-008 FIX: Returns error if any callback registration fails
func (g *EbitenGame) SetupInputCallbacks(inputSystem *InputSystem, objectiveTracker *ObjectiveTrackerSystem) error {
	if err := g.setupUICallbacks(inputSystem, objectiveTracker); err != nil {
		return err
	}
	if err := g.setupOptionalUICallbacks(inputSystem, objectiveTracker); err != nil {
		return err
	}
	return g.setupMenuCallback(inputSystem)
}

// setupUICallbacks connects core UI toggle callbacks (inventory, quests, character, skills, map).
func (g *EbitenGame) setupUICallbacks(inputSystem *InputSystem, objectiveTracker *ObjectiveTrackerSystem) error {
	callbacks := []struct {
		name    string
		setter  func(func()) error
		toggler func()
		uiName  string
	}{
		{"inventory", inputSystem.SetInventoryCallback, g.InventoryUI.Toggle, "inventory"},
		{"quests", inputSystem.SetQuestsCallback, g.QuestUI.Toggle, "quest_log"},
		{"character", inputSystem.SetCharacterCallback, g.CharacterUI.Toggle, "character"},
		{"skills", inputSystem.SetSkillsCallback, g.SkillsUI.Toggle, "skills"},
		{"map", inputSystem.SetMapCallback, g.MapUI.ToggleFullScreen, "map"},
	}

	for _, cb := range callbacks {
		if err := cb.setter(g.makeUIToggleCallback(cb.toggler, cb.uiName, objectiveTracker)); err != nil {
			return fmt.Errorf("failed to set %s callback: %w", cb.name, err)
		}
	}
	return nil
}

// setupOptionalUICallbacks connects optional UI callbacks (crafting, mailbox, housing, gallery).
func (g *EbitenGame) setupOptionalUICallbacks(inputSystem *InputSystem, objectiveTracker *ObjectiveTrackerSystem) error {
	optionalCallbacks := []struct {
		name   string
		setter func(func()) error
		ui     interface{ Toggle() }
		uiName string
	}{
		{"crafting", inputSystem.SetCraftingCallback, g.CraftingUI, "crafting"},
		{"mailbox", inputSystem.SetMailboxCallback, g.MailboxUI, "mailbox"},
		{"trade", inputSystem.SetTradeCallback, g.TradeUI, "trade"}, // Phase 3.3 (PLAN.md)
		{"guild", inputSystem.SetGuildCallback, g.GuildUI, "guild"}, // Phase 3.2 (PLAN.md)
		{"housing", inputSystem.SetHousingCallback, g.HousingUI, "housing"},
		{"gallery", inputSystem.SetGalleryCallback, g.GalleryUI, "gallery"},
	}

	for _, cb := range optionalCallbacks {
		if err := cb.setter(g.makeOptionalUIToggleCallback(cb.ui, cb.uiName, objectiveTracker)); err != nil {
			return fmt.Errorf("failed to set %s callback: %w", cb.name, err)
		}
	}
	return nil
}

// setupMenuCallback connects the pause menu toggle callback.
func (g *EbitenGame) setupMenuCallback(inputSystem *InputSystem) error {
	if g.MenuSystem != nil {
		if err := inputSystem.SetMenuToggleCallback(func() {
			g.MenuSystem.Toggle()
		}); err != nil {
			return fmt.Errorf("failed to set menu toggle callback: %w", err)
		}
	}
	return nil
}

// makeUIToggleCallback creates a callback that toggles UI and tracks objective completion.
func (g *EbitenGame) makeUIToggleCallback(toggler func(), uiName string, objectiveTracker *ObjectiveTrackerSystem) func() {
	return func() {
		toggler()
		if objectiveTracker != nil && g.PlayerEntity != nil {
			objectiveTracker.OnUIOpened(g.PlayerEntity, uiName)
		}
	}
}

// makeOptionalUIToggleCallback creates a callback for optional UI elements.
func (g *EbitenGame) makeOptionalUIToggleCallback(ui interface{ Toggle() }, uiName string, objectiveTracker *ObjectiveTrackerSystem) func() {
	return func() {
		if ui != nil {
			ui.Toggle()
			if objectiveTracker != nil && g.PlayerEntity != nil {
				objectiveTracker.OnUIOpened(g.PlayerEntity, uiName)
			}
		}
	}
}

// GetWorld returns the ECS world instance (implements GameRunner interface).
func (g *EbitenGame) GetWorld() *World {
	return g.World
}

// GetScreenSize returns the current screen dimensions (implements GameRunner interface).
func (g *EbitenGame) GetScreenSize() (width, height int) {
	return g.ScreenWidth, g.ScreenHeight
}

// IsPaused returns whether the game is currently paused (implements GameRunner interface).
func (g *EbitenGame) IsPaused() bool {
	return g.Paused
}

// SetPaused sets the game pause state (implements GameRunner interface).
func (g *EbitenGame) SetPaused(paused bool) {
	g.Paused = paused
}

// GetPlayerEntity returns the current player entity (implements GameRunner interface).
func (g *EbitenGame) GetPlayerEntity() *Entity {
	return g.PlayerEntity
}

// ApplySettings applies settings from SettingsManager to game systems.
// This should be called when settings are loaded or changed.
// Returns error if application fails for critical settings.
func (g *EbitenGame) ApplySettings() error {
	if g.SettingsManager == nil {
		return nil // No settings to apply
	}

	settings := g.SettingsManager.GetSettings()

	// Apply audio volumes
	if g.AudioManager != nil {
		// Master volume affects both music and SFX as a multiplier
		g.AudioManager.SetMusicVolume(settings.MasterVolume * settings.MusicVolume)
		g.AudioManager.SetSFXVolume(settings.MasterVolume * settings.SFXVolume)

		if g.logger != nil {
			g.logger.WithFields(logrus.Fields{
				"masterVolume": settings.MasterVolume,
				"musicVolume":  settings.MusicVolume,
				"sfxVolume":    settings.SFXVolume,
			}).Debug("applied audio settings")
		}
	}

	// Apply VSync
	ebiten.SetVsyncEnabled(settings.VSync)

	// Apply fullscreen (Note: this triggers a window mode change)
	if settings.Fullscreen != ebiten.IsFullscreen() {
		ebiten.SetFullscreen(settings.Fullscreen)
	}

	// Apply ShowTutorials setting to active tutorial systems (Task 3.3 from PLAN.md)
	// When OnboardingManager exists, delegate to it so child systems are only
	// enabled for the correct onboarding phase. This prevents premature
	// activation of the in-game tutorial during character creation (e.g. when
	// SetAudioManager triggers ApplySettings before character creation finishes).
	if g.OnboardingManager != nil {
		g.OnboardingManager.SetEnabled(settings.ShowTutorials)
	} else {
		// No onboarding manager: directly control individual systems
		if g.TutorialSystem != nil {
			g.TutorialSystem.Enabled = settings.ShowTutorials
			g.TutorialSystem.ShowUI = settings.ShowTutorials
		}
		if g.CharacterCreationTutorial != nil {
			g.CharacterCreationTutorial.SetEnabled(settings.ShowTutorials)
		}
	}
	// ContextualTutorial is independent of the onboarding state machine
	if g.ContextualTutorial != nil {
		if settings.ShowTutorials {
			g.ContextualTutorial.Enable()
		} else {
			g.ContextualTutorial.Disable()
		}
	}

	// Graphics quality and ShowFPS are informational for now
	// Future: could affect particle counts, sprite quality, etc.

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"vsync":         settings.VSync,
			"fullscreen":    settings.Fullscreen,
			"quality":       settings.GraphicsQuality,
			"showFPS":       settings.ShowFPS,
			"showTutorials": settings.ShowTutorials,
		}).Debug("applied display settings")
	}

	return nil
}

// EnableFrameTimeProfiling enables performance profiling with frame time tracking.
// This should be called before starting the game loop.
func (g *EbitenGame) EnableFrameTimeProfiling() {
	g.profilingEnabled = true
	if g.logger != nil {
		g.logger.Info("frame time profiling enabled")
	}
}

// DisableFrameTimeProfiling disables performance profiling.
func (g *EbitenGame) DisableFrameTimeProfiling() {
	g.profilingEnabled = false
	if g.logger != nil {
		g.logger.Info("frame time profiling disabled")
	}
}

// GetFrameTimeStats returns the current frame time statistics.
// Returns empty stats if profiling is disabled or no frames recorded.
func (g *EbitenGame) GetFrameTimeStats() FrameTimeStats {
	if g.frameTimeTracker == nil {
		return FrameTimeStats{}
	}
	return g.frameTimeTracker.GetStats()
}

// CurrentFPS implements stability.FPSProvider interface.
// Returns the current average frames per second from frame time tracking.
func (g *EbitenGame) CurrentFPS() float64 {
	if g.frameTimeTracker == nil {
		return 60.0 // Default to target FPS when tracking disabled
	}
	stats := g.frameTimeTracker.GetStats()
	return stats.GetFPS()
}

// Run starts the game loop.
func (g *EbitenGame) Run(title string) error {
	ebiten.SetWindowSize(g.ScreenWidth, g.ScreenHeight)
	ebiten.SetWindowTitle(title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil {
		return fmt.Errorf("failed to run game: %w", err)
	}

	return nil
}

// SetFullscreen sets fullscreen mode.
func (g *EbitenGame) SetFullscreen(fullscreen bool) {
	ebiten.SetFullscreen(fullscreen)
}

// Compile-time interface checks
var (
	_ GameRunner  = (*EbitenGame)(nil)
	_ ebiten.Game = (*EbitenGame)(nil)
)

// EnableLighting enables or disables the dynamic lighting system.
// When enabled, uses post-processing rendering pipeline with light sources.
func (g *EbitenGame) EnableLighting(enabled bool) {
	if g.LightingSystem != nil {
		g.LightingSystem.SetEnabled(enabled)

		if g.logger != nil {
			g.logger.WithField("enabled", enabled).Info("lighting system toggled")
		}
	}
}

// SetLightingGenrePreset configures lighting for the specified genre.
// This should be called when the genre is selected or changed.
func (g *EbitenGame) SetLightingGenrePreset(genreID string) {
	if g.LightingSystem != nil {
		config := g.LightingSystem.GetConfig()
		if config != nil {
			config.SetGenrePreset(genreID)
			g.LightingSystem.SetConfig(config)

			if g.logger != nil {
				g.logger.WithField("genre", genreID).Info("lighting genre preset applied")
			}
		}
	}
}
