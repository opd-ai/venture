package engine

import (
	"testing"
)

// TestGameState_String tests state name formatting
func TestGameState_String(t *testing.T) {
	tests := []struct {
		state GameState
		want  string
	}{
		{StateExploring, "Exploring"},
		{StateCombat, "Combat"},
		{StateMenu, "Menu"},
		{StateDialogue, "Dialogue"},
		{StateInventory, "Inventory"},
		{StateCharacterScreen, "Character"},
		{StateSkillTree, "Skills"},
		{StateQuestLog, "Quests"},
		{StateMap, "Map"},
		{GameState(9999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("GameState.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestGameState_AllowsMovement tests movement permission logic
func TestGameState_AllowsMovement(t *testing.T) {
	tests := []struct {
		state GameState
		want  bool
	}{
		{StateExploring, true},
		{StateCombat, true},
		{StateMenu, false},
		{StateDialogue, false},
		{StateInventory, false},
		{StateCharacterScreen, false},
		{StateSkillTree, false},
		{StateQuestLog, false},
		{StateMap, false},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			got := tt.state.AllowsMovement()
			if got != tt.want {
				t.Errorf("%s.AllowsMovement() = %v, want %v", tt.state.String(), got, tt.want)
			}
		})
	}
}

// TestGameState_AllowsCombat tests combat permission logic
func TestGameState_AllowsCombat(t *testing.T) {
	tests := []struct {
		state GameState
		want  bool
	}{
		{StateExploring, true},
		{StateCombat, true},
		{StateMenu, false},
		{StateDialogue, false},
		{StateInventory, false},
		{StateCharacterScreen, false},
		{StateSkillTree, false},
		{StateQuestLog, false},
		{StateMap, false},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			got := tt.state.AllowsCombat()
			if got != tt.want {
				t.Errorf("%s.AllowsCombat() = %v, want %v", tt.state.String(), got, tt.want)
			}
		})
	}
}

// TestGameState_AllowsUIToggle tests UI toggle permission logic
func TestGameState_AllowsUIToggle(t *testing.T) {
	tests := []struct {
		state GameState
		want  bool
	}{
		{StateExploring, true},
		{StateCombat, true},
		{StateMenu, true},
		{StateDialogue, false}, // Must finish dialogue first
		{StateInventory, true},
		{StateCharacterScreen, true},
		{StateSkillTree, true},
		{StateQuestLog, true},
		{StateMap, true},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			got := tt.state.AllowsUIToggle()
			if got != tt.want {
				t.Errorf("%s.AllowsUIToggle() = %v, want %v", tt.state.String(), got, tt.want)
			}
		})
	}
}

// TestGameState_IsUIState tests UI state detection
func TestGameState_IsUIState(t *testing.T) {
	tests := []struct {
		state GameState
		want  bool
	}{
		{StateExploring, false},
		{StateCombat, false},
		{StateMenu, true},
		{StateDialogue, false}, // Dialogue is special, not pure UI
		{StateInventory, true},
		{StateCharacterScreen, true},
		{StateSkillTree, true},
		{StateQuestLog, true},
		{StateMap, true},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			got := tt.state.IsUIState()
			if got != tt.want {
				t.Errorf("%s.IsUIState() = %v, want %v", tt.state.String(), got, tt.want)
			}
		})
	}
}

// TestInputSystem_GameStateManagement tests state get/set methods
func TestInputSystem_GameStateManagement(t *testing.T) {
	system := NewInputSystem()

	// Verify default state
	if system.GetGameState() != StateExploring {
		t.Errorf("NewInputSystem() default state = %v, want %v", system.GetGameState(), StateExploring)
	}

	// Change to menu state
	system.SetGameState(StateMenu)
	if system.GetGameState() != StateMenu {
		t.Errorf("After SetGameState(StateMenu), GetGameState() = %v, want %v", system.GetGameState(), StateMenu)
	}

	// Change to inventory state
	system.SetGameState(StateInventory)
	if system.GetGameState() != StateInventory {
		t.Errorf("After SetGameState(StateInventory), GetGameState() = %v, want %v", system.GetGameState(), StateInventory)
	}

	// Back to exploring
	system.SetGameState(StateExploring)
	if system.GetGameState() != StateExploring {
		t.Errorf("After SetGameState(StateExploring), GetGameState() = %v, want %v", system.GetGameState(), StateExploring)
	}
}

// TestInputSystem_KeyBindingsAccess tests key bindings registry access
func TestInputSystem_KeyBindingsAccess(t *testing.T) {
	system := NewInputSystem()

	bindings := system.GetKeyBindings()
	if bindings == nil {
		t.Fatal("GetKeyBindings() returned nil")
	}

	// Verify bindings registry is functional
	label := bindings.GetActionLabel(ActionMoveUp)
	expectedLabel := "Move Up [W]"
	if label != expectedLabel {
		t.Errorf("GetKeyBindings().GetActionLabel(ActionMoveUp) = %q, want %q", label, expectedLabel)
	}
}

// TestInputSystem_StateTransitions tests typical state transition flows
func TestInputSystem_StateTransitions(t *testing.T) {
	system := NewInputSystem()

	// Start in exploring
	if system.GetGameState() != StateExploring {
		t.Fatal("Expected initial state to be Exploring")
	}

	// Open inventory
	system.SetGameState(StateInventory)
	if system.GetGameState().AllowsMovement() {
		t.Error("Movement should be blocked in inventory")
	}
	if system.GetGameState().AllowsCombat() {
		t.Error("Combat should be blocked in inventory")
	}
	if !system.GetGameState().IsUIState() {
		t.Error("Inventory should be a UI state")
	}

	// Close inventory (back to exploring)
	system.SetGameState(StateExploring)
	if !system.GetGameState().AllowsMovement() {
		t.Error("Movement should be allowed in exploring")
	}
	if !system.GetGameState().AllowsCombat() {
		t.Error("Combat should be allowed in exploring")
	}

	// Enter combat
	system.SetGameState(StateCombat)
	if !system.GetGameState().AllowsMovement() {
		t.Error("Movement should be allowed in combat")
	}
	if !system.GetGameState().AllowsCombat() {
		t.Error("Combat should be allowed in combat state")
	}

	// Open map during combat (valid)
	system.SetGameState(StateMap)
	if system.GetGameState().AllowsMovement() {
		t.Error("Movement should be blocked in map")
	}
	if !system.GetGameState().IsUIState() {
		t.Error("Map should be a UI state")
	}

	// Close map (back to combat)
	system.SetGameState(StateCombat)
	if !system.GetGameState().AllowsCombat() {
		t.Error("Combat should be allowed after closing map")
	}
}

// TestInputSystem_DialogueStateRestrictions tests dialogue state limitations
func TestInputSystem_DialogueStateRestrictions(t *testing.T) {
	system := NewInputSystem()

	// Enter dialogue
	system.SetGameState(StateDialogue)

	// Dialogue blocks movement
	if system.GetGameState().AllowsMovement() {
		t.Error("Movement should be blocked during dialogue")
	}

	// Dialogue blocks combat
	if system.GetGameState().AllowsCombat() {
		t.Error("Combat should be blocked during dialogue")
	}

	// Dialogue blocks UI toggles (must finish dialogue first)
	if system.GetGameState().AllowsUIToggle() {
		t.Error("UI toggle should be blocked during dialogue")
	}

	// Dialogue is not a pure UI state (it's gameplay)
	if system.GetGameState().IsUIState() {
		t.Error("Dialogue should not be classified as UI state")
	}
}

// TestInputSystem_MenuStateBlocksGameplay tests menu state input blocking
func TestInputSystem_MenuStateBlocksGameplay(t *testing.T) {
	system := NewInputSystem()

	// Enter menu
	system.SetGameState(StateMenu)

	// Menu blocks movement
	if system.GetGameState().AllowsMovement() {
		t.Error("Movement should be blocked in menu")
	}

	// Menu blocks combat
	if system.GetGameState().AllowsCombat() {
		t.Error("Combat should be blocked in menu")
	}

	// Menu allows UI toggle (to close menu)
	if !system.GetGameState().AllowsUIToggle() {
		t.Error("UI toggle should be allowed in menu (to close it)")
	}

	// Menu is a UI state
	if !system.GetGameState().IsUIState() {
		t.Error("Menu should be classified as UI state")
	}
}

// TestInputSystem_AllUIStatesBlockGameplay tests all UI states block movement/combat
func TestInputSystem_AllUIStatesBlockGameplay(t *testing.T) {
	system := NewInputSystem()

	uiStates := []GameState{
		StateMenu,
		StateInventory,
		StateCharacterScreen,
		StateSkillTree,
		StateQuestLog,
		StateMap,
	}

	for _, state := range uiStates {
		t.Run(state.String(), func(t *testing.T) {
			system.SetGameState(state)

			if system.GetGameState().AllowsMovement() {
				t.Errorf("%s should block movement", state.String())
			}

			if system.GetGameState().AllowsCombat() {
				t.Errorf("%s should block combat", state.String())
			}

			if !system.GetGameState().IsUIState() {
				t.Errorf("%s should be classified as UI state", state.String())
			}
		})
	}
}

// TestInputSystem_ExploringAndCombatAllowGameplay tests gameplay states allow actions
func TestInputSystem_ExploringAndCombatAllowGameplay(t *testing.T) {
	system := NewInputSystem()

	gameplayStates := []GameState{
		StateExploring,
		StateCombat,
	}

	for _, state := range gameplayStates {
		t.Run(state.String(), func(t *testing.T) {
			system.SetGameState(state)

			if !system.GetGameState().AllowsMovement() {
				t.Errorf("%s should allow movement", state.String())
			}

			if !system.GetGameState().AllowsCombat() {
				t.Errorf("%s should allow combat", state.String())
			}

			if system.GetGameState().IsUIState() {
				t.Errorf("%s should not be classified as UI state", state.String())
			}
		})
	}
}
