// Package engine provides application-level state management.
// This file implements the state machine for transitioning between
// main menu, submenus, and gameplay phases. This is different from
// GameState in input_system.go which manages UI input filtering.
package engine

import "fmt"

// AppState represents the current phase of the game application.
// Controls what scene is displayed and whether world simulation is active.
type AppState int

const (
	// AppStateMainMenu is the initial state showing main menu options.
	AppStateMainMenu AppState = iota
	// AppStateSinglePlayerMenu shows New Game / Load Game / Back options.
	AppStateSinglePlayerMenu
	// AppStateGenreSelection shows genre selection for single-player.
	AppStateGenreSelection
	// AppStateMultiPlayerMenu shows Join/Host/Back options for multiplayer.
	AppStateMultiPlayerMenu
	// AppStateServerAddressInput shows text input for server address.
	AppStateServerAddressInput
	// AppStateCharacterCreation shows character creation UI (future).
	AppStateCharacterCreation
	// AppStateGameplay is the active game state with world simulation.
	AppStateGameplay
	// AppStateSettings shows game settings (future).
	AppStateSettings
)

// String returns a human-readable name for the app state.
func (s AppState) String() string {
	switch s {
	case AppStateMainMenu:
		return "MainMenu"
	case AppStateSinglePlayerMenu:
		return "SinglePlayerMenu"
	case AppStateGenreSelection:
		return "GenreSelection"
	case AppStateMultiPlayerMenu:
		return "MultiPlayerMenu"
	case AppStateServerAddressInput:
		return "ServerAddressInput"
	case AppStateCharacterCreation:
		return "CharacterCreation"
	case AppStateGameplay:
		return "Gameplay"
	case AppStateSettings:
		return "Settings"
	default:
		return "Unknown"
	}
}

// AppStateManager manages application state transitions and validation.
// Ensures only valid state transitions occur and provides hooks for
// state change callbacks.
type AppStateManager struct {
	currentState  AppState
	previousState AppState
	onStateChange func(from, to AppState) error
}

// NewAppStateManager creates a new state manager starting in main menu.
func NewAppStateManager() *AppStateManager {
	return &AppStateManager{
		currentState:  AppStateMainMenu,
		previousState: AppStateMainMenu,
	}
}

// CurrentState returns the current application state.
func (asm *AppStateManager) CurrentState() AppState {
	return asm.currentState
}

// PreviousState returns the previous application state (useful for back navigation).
func (asm *AppStateManager) PreviousState() AppState {
	return asm.previousState
}

// SetStateChangeCallback sets a callback function that is called on state transitions.
// The callback can return an error to prevent the state change.
func (asm *AppStateManager) SetStateChangeCallback(callback func(from, to AppState) error) {
	asm.onStateChange = callback
}

// TransitionTo changes the application state to the specified state.
// Returns an error if the transition is invalid or if the state change callback fails.
func (asm *AppStateManager) TransitionTo(newState AppState) error {
	if asm.currentState == newState {
		return nil // No-op if already in target state
	}

	// Validate transition
	if !isValidAppTransition(asm.currentState, newState) {
		return fmt.Errorf("invalid state transition from %s to %s", asm.currentState, newState)
	}

	// Call state change callback if set
	if asm.onStateChange != nil {
		if err := asm.onStateChange(asm.currentState, newState); err != nil {
			return fmt.Errorf("state change callback failed: %w", err)
		}
	}

	asm.previousState = asm.currentState
	asm.currentState = newState
	return nil
}

// Back transitions to the previous state (useful for back buttons).
// Only works for menu states, not from gameplay back to menu.
func (asm *AppStateManager) Back() error {
	// Define back navigation logic
	var targetState AppState
	switch asm.currentState {
	case AppStateSinglePlayerMenu, AppStateMultiPlayerMenu, AppStateSettings:
		targetState = AppStateMainMenu
	case AppStateCharacterCreation:
		// For MVP, go back to main menu (skip single player submenu)
		targetState = AppStateMainMenu
	default:
		return fmt.Errorf("cannot navigate back from %s", asm.currentState)
	}

	return asm.TransitionTo(targetState)
}

// isValidAppTransition checks if a state transition is allowed.
// This enforces the state machine rules and prevents invalid transitions.
func isValidAppTransition(from, to AppState) bool {
	// Allow transitions from any state to main menu (for quitting gameplay)
	if to == AppStateMainMenu {
		return true
	}

	switch from {
	case AppStateMainMenu:
		// Can go to any submenu, settings, or directly to gameplay (MVP) from main menu
		return to == AppStateSinglePlayerMenu || to == AppStateMultiPlayerMenu || to == AppStateSettings || to == AppStateCharacterCreation || to == AppStateGameplay
	case AppStateSinglePlayerMenu:
		// Can start new game (via genre selection), load game, or go back to main menu
		return to == AppStateGenreSelection || to == AppStateCharacterCreation || to == AppStateGameplay || to == AppStateMainMenu
	case AppStateMultiPlayerMenu:
		// Can show server address input (Join), start host game, or go back
		return to == AppStateServerAddressInput || to == AppStateGameplay || to == AppStateMainMenu
	case AppStateServerAddressInput:
		// Can connect to server (gameplay) or go back to multiplayer menu
		return to == AppStateGameplay || to == AppStateMultiPlayerMenu || to == AppStateMainMenu
	case AppStateGenreSelection:
		// Can proceed to character creation or go back to single-player menu
		return to == AppStateCharacterCreation || to == AppStateSinglePlayerMenu || to == AppStateMainMenu
	case AppStateCharacterCreation:
		// Can complete creation and start game or go back to main menu (MVP flow)
		return to == AppStateGameplay || to == AppStateSinglePlayerMenu || to == AppStateMainMenu
	case AppStateSettings:
		// Can only go back to main menu
		return to == AppStateMainMenu
	case AppStateGameplay:
		// Can only quit to main menu from gameplay
		return to == AppStateMainMenu
	default:
		return false
	}
}

// IsInMenu returns true if currently in any menu state (not gameplay).
func (asm *AppStateManager) IsInMenu() bool {
	return asm.currentState != AppStateGameplay
}

// IsInGameplay returns true if currently in gameplay state.
func (asm *AppStateManager) IsInGameplay() bool {
	return asm.currentState == AppStateGameplay
}
