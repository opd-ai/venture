package engine

import "testing"

func TestAppState_Loading(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"loading state string", "Loading"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AppStateLoading.String(); got != tt.want {
				t.Errorf("AppStateLoading.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppStateManager_LoadingTransitions(t *testing.T) {
	tests := []struct {
		name      string
		fromState AppState
		toState   AppState
		wantErr   bool
	}{
		{"character creation to loading", AppStateCharacterCreation, AppStateLoading, false},
		{"loading to gameplay", AppStateLoading, AppStateGameplay, false},
		{"loading to main menu", AppStateLoading, AppStateMainMenu, false},
		{"main menu to loading", AppStateMainMenu, AppStateLoading, false}, // Valid: direct loading from main menu
		{"gameplay to loading", AppStateGameplay, AppStateLoading, true},  // Invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asm := NewAppStateManager()
			// Set up initial state
			if tt.fromState != AppStateMainMenu {
				_ = asm.TransitionTo(tt.fromState)
			}

			err := asm.TransitionTo(tt.toState)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransitionTo() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && asm.CurrentState() != tt.toState {
				t.Errorf("TransitionTo() state = %v, want %v", asm.CurrentState(), tt.toState)
			}
		})
	}
}

func TestAppStateManager_IsInMenu_WithLoading(t *testing.T) {
	asm := NewAppStateManager()

	// Loading state should not be considered a menu
	_ = asm.TransitionTo(AppStateCharacterCreation)
	_ = asm.TransitionTo(AppStateLoading)

	if asm.IsInMenu() {
		t.Error("IsInMenu() should return false for Loading state")
	}

	if asm.IsInGameplay() {
		t.Error("IsInGameplay() should return false for Loading state")
	}
}
