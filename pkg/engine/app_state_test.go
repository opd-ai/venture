package engine

import (
	"errors"
	"testing"
)

func TestAppState_String(t *testing.T) {
	tests := []struct {
		name  string
		state AppState
		want  string
	}{
		{"main menu", AppStateMainMenu, "MainMenu"},
		{"single player menu", AppStateSinglePlayerMenu, "SinglePlayerMenu"},
		{"multiplayer menu", AppStateMultiPlayerMenu, "MultiPlayerMenu"},
		{"character creation", AppStateCharacterCreation, "CharacterCreation"},
		{"gameplay", AppStateGameplay, "Gameplay"},
		{"settings", AppStateSettings, "Settings"},
		{"unknown", AppState(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("AppState.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAppStateManager(t *testing.T) {
	asm := NewAppStateManager()

	if asm.CurrentState() != AppStateMainMenu {
		t.Errorf("NewAppStateManager() initial state = %v, want %v", asm.CurrentState(), AppStateMainMenu)
	}

	if asm.PreviousState() != AppStateMainMenu {
		t.Errorf("NewAppStateManager() previous state = %v, want %v", asm.PreviousState(), AppStateMainMenu)
	}
}

func TestAppStateManager_TransitionTo(t *testing.T) {
	tests := []struct {
		name        string
		fromState   AppState
		toState     AppState
		shouldError bool
	}{
		// Valid transitions from main menu
		{"main menu to single player", AppStateMainMenu, AppStateSinglePlayerMenu, false},
		{"main menu to multiplayer", AppStateMainMenu, AppStateMultiPlayerMenu, false},
		{"main menu to settings", AppStateMainMenu, AppStateSettings, false},
		{"main menu to character creation", AppStateMainMenu, AppStateCharacterCreation, false},
		{"main menu to gameplay", AppStateMainMenu, AppStateGameplay, false},

		// Valid transitions from single player menu
		{"single player to gameplay", AppStateSinglePlayerMenu, AppStateGameplay, false},
		{"single player to character creation", AppStateSinglePlayerMenu, AppStateCharacterCreation, false},
		{"single player to main menu", AppStateSinglePlayerMenu, AppStateMainMenu, false},

		// Valid transitions from multiplayer menu
		{"multiplayer to gameplay", AppStateMultiPlayerMenu, AppStateGameplay, false},
		{"multiplayer to main menu", AppStateMultiPlayerMenu, AppStateMainMenu, false},

		// Valid transitions from character creation
		{"character creation to gameplay", AppStateCharacterCreation, AppStateGameplay, false},
		{"character creation to single player", AppStateCharacterCreation, AppStateSinglePlayerMenu, false},

		// Valid transitions from gameplay
		{"gameplay to main menu", AppStateGameplay, AppStateMainMenu, false},

		// Valid transitions from settings
		{"settings to main menu", AppStateSettings, AppStateMainMenu, false},

		// Invalid transitions
		{"gameplay to single player", AppStateGameplay, AppStateSinglePlayerMenu, true},
		{"gameplay to settings", AppStateGameplay, AppStateSettings, true},
		{"multiplayer to single player", AppStateMultiPlayerMenu, AppStateSinglePlayerMenu, true},
		{"settings to gameplay", AppStateSettings, AppStateGameplay, true},

		// Same state transition (should be no-op, no error)
		{"same state main menu", AppStateMainMenu, AppStateMainMenu, false},
		{"same state gameplay", AppStateGameplay, AppStateGameplay, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asm := NewAppStateManager()
			asm.currentState = tt.fromState

			err := asm.TransitionTo(tt.toState)

			if tt.shouldError && err == nil {
				t.Errorf("TransitionTo() expected error but got nil")
			}

			if !tt.shouldError && err != nil {
				t.Errorf("TransitionTo() unexpected error: %v", err)
			}

			if !tt.shouldError && asm.CurrentState() != tt.toState {
				t.Errorf("TransitionTo() current state = %v, want %v", asm.CurrentState(), tt.toState)
			}

			// Verify previous state is preserved (except for no-op transitions)
			if !tt.shouldError && tt.fromState != tt.toState && asm.PreviousState() != tt.fromState {
				t.Errorf("TransitionTo() previous state = %v, want %v", asm.PreviousState(), tt.fromState)
			}
		})
	}
}

func TestAppStateManager_Back(t *testing.T) {
	tests := []struct {
		name          string
		fromState     AppState
		expectedState AppState
		shouldError   bool
	}{
		{"back from single player menu", AppStateSinglePlayerMenu, AppStateMainMenu, false},
		{"back from multiplayer menu", AppStateMultiPlayerMenu, AppStateMainMenu, false},
		{"back from settings", AppStateSettings, AppStateMainMenu, false},
		{"back from character creation", AppStateCharacterCreation, AppStateMainMenu, false},
		{"cannot back from main menu", AppStateMainMenu, AppStateMainMenu, true},
		{"cannot back from gameplay", AppStateGameplay, AppStateGameplay, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asm := NewAppStateManager()
			asm.currentState = tt.fromState

			err := asm.Back()

			if tt.shouldError && err == nil {
				t.Errorf("Back() expected error but got nil")
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Back() unexpected error: %v", err)
			}

			if !tt.shouldError && asm.CurrentState() != tt.expectedState {
				t.Errorf("Back() current state = %v, want %v", asm.CurrentState(), tt.expectedState)
			}
		})
	}
}

func TestAppStateManager_SetStateChangeCallback(t *testing.T) {
	tests := []struct {
		name            string
		callbackReturns error
		shouldTransit   bool
	}{
		{"callback allows transition", nil, true},
		{"callback prevents transition", errors.New("transition denied"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asm := NewAppStateManager()
			callbackCalled := false

			asm.SetStateChangeCallback(func(from, to AppState) error {
				callbackCalled = true
				if from != AppStateMainMenu {
					t.Errorf("callback from = %v, want %v", from, AppStateMainMenu)
				}
				if to != AppStateSinglePlayerMenu {
					t.Errorf("callback to = %v, want %v", to, AppStateSinglePlayerMenu)
				}
				return tt.callbackReturns
			})

			err := asm.TransitionTo(AppStateSinglePlayerMenu)

			if !callbackCalled {
				t.Error("state change callback was not called")
			}

			if tt.shouldTransit {
				if err != nil {
					t.Errorf("TransitionTo() unexpected error: %v", err)
				}
				if asm.CurrentState() != AppStateSinglePlayerMenu {
					t.Errorf("TransitionTo() current state = %v, want %v", asm.CurrentState(), AppStateSinglePlayerMenu)
				}
			} else {
				if err == nil {
					t.Error("TransitionTo() expected error but got nil")
				}
				if asm.CurrentState() != AppStateMainMenu {
					t.Errorf("TransitionTo() current state = %v, want %v (should not change)", asm.CurrentState(), AppStateMainMenu)
				}
			}
		})
	}
}

func TestAppStateManager_IsInMenu(t *testing.T) {
	tests := []struct {
		name  string
		state AppState
		want  bool
	}{
		{"main menu is menu", AppStateMainMenu, true},
		{"single player menu is menu", AppStateSinglePlayerMenu, true},
		{"multiplayer menu is menu", AppStateMultiPlayerMenu, true},
		{"character creation is menu", AppStateCharacterCreation, true},
		{"settings is menu", AppStateSettings, true},
		{"gameplay is not menu", AppStateGameplay, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asm := NewAppStateManager()
			asm.currentState = tt.state

			if got := asm.IsInMenu(); got != tt.want {
				t.Errorf("IsInMenu() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppStateManager_IsInGameplay(t *testing.T) {
	tests := []struct {
		name  string
		state AppState
		want  bool
	}{
		{"gameplay is gameplay", AppStateGameplay, true},
		{"main menu is not gameplay", AppStateMainMenu, false},
		{"single player menu is not gameplay", AppStateSinglePlayerMenu, false},
		{"multiplayer menu is not gameplay", AppStateMultiPlayerMenu, false},
		{"character creation is not gameplay", AppStateCharacterCreation, false},
		{"settings is not gameplay", AppStateSettings, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asm := NewAppStateManager()
			asm.currentState = tt.state

			if got := asm.IsInGameplay(); got != tt.want {
				t.Errorf("IsInGameplay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppStateManager_TransitionSequence(t *testing.T) {
	// Test a complete user flow: Main Menu -> Single Player -> Gameplay -> Main Menu
	asm := NewAppStateManager()

	// Start in main menu
	if asm.CurrentState() != AppStateMainMenu {
		t.Fatalf("initial state = %v, want %v", asm.CurrentState(), AppStateMainMenu)
	}

	// Go to single player menu
	if err := asm.TransitionTo(AppStateSinglePlayerMenu); err != nil {
		t.Fatalf("transition to single player failed: %v", err)
	}
	if asm.CurrentState() != AppStateSinglePlayerMenu {
		t.Errorf("state = %v, want %v", asm.CurrentState(), AppStateSinglePlayerMenu)
	}

	// Start new game (go to gameplay)
	if err := asm.TransitionTo(AppStateGameplay); err != nil {
		t.Fatalf("transition to gameplay failed: %v", err)
	}
	if asm.CurrentState() != AppStateGameplay {
		t.Errorf("state = %v, want %v", asm.CurrentState(), AppStateGameplay)
	}

	// Quit to main menu
	if err := asm.TransitionTo(AppStateMainMenu); err != nil {
		t.Fatalf("transition back to main menu failed: %v", err)
	}
	if asm.CurrentState() != AppStateMainMenu {
		t.Errorf("state = %v, want %v", asm.CurrentState(), AppStateMainMenu)
	}
}

func TestAppStateManager_BackNavigation(t *testing.T) {
	// Test back navigation through menus
	asm := NewAppStateManager()

	// Main Menu -> Single Player Menu
	if err := asm.TransitionTo(AppStateSinglePlayerMenu); err != nil {
		t.Fatalf("transition to single player failed: %v", err)
	}

	// Back to main menu
	if err := asm.Back(); err != nil {
		t.Fatalf("back navigation failed: %v", err)
	}
	if asm.CurrentState() != AppStateMainMenu {
		t.Errorf("state after back = %v, want %v", asm.CurrentState(), AppStateMainMenu)
	}

	// Main Menu -> Multiplayer Menu -> Back
	if err := asm.TransitionTo(AppStateMultiPlayerMenu); err != nil {
		t.Fatalf("transition to multiplayer failed: %v", err)
	}
	if err := asm.Back(); err != nil {
		t.Fatalf("back navigation failed: %v", err)
	}
	if asm.CurrentState() != AppStateMainMenu {
		t.Errorf("state after back = %v, want %v", asm.CurrentState(), AppStateMainMenu)
	}
}
