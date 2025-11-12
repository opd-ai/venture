package engine

import (
	"testing"
)

// TestInputSystem_ExpressionIntegration tests expression system integration with input.
// Phase 26.1: Expression hotkey handling (Shift+1 through Shift+=).
func TestInputSystem_ExpressionIntegration(t *testing.T) {
	// Create world and systems
	world := NewWorld()
	inputSys := NewInputSystem()
	audioMgr := NewAudioManager(44100, 12345)
	exprSys := NewExpressionSystem(world, audioMgr)

	// Connect expression system to input system
	inputSys.SetExpressionSystem(exprSys)

	// Verify setter worked
	if inputSys.expressionSystem == nil {
		t.Fatal("Expected expression system to be set")
	}
	if inputSys.expressionSystem != exprSys {
		t.Error("Expected expression system reference to match")
	}

	// Create player entity with input component
	player := world.CreateEntity()
	player.AddComponent(&EbitenInput{})
	world.Update(0) // Process entity addition

	// Note: We can't actually test key press simulation without ebiten runtime,
	// but we've verified the system is connected properly

	// Verify player can receive expressions
	success := exprSys.TriggerExpression(player.ID, ExpressionWave)
	if !success {
		t.Error("Expected expression trigger to succeed")
	}

	// Verify expression component was added
	expComp, ok := player.GetComponent("expression")
	if !ok {
		t.Error("Expected expression component to be added")
	}
	if expComp == nil {
		t.Error("Expected non-nil expression component")
	}
}

// TestInputSystem_SetExpressionSystem tests the setter method.
func TestInputSystem_SetExpressionSystem(t *testing.T) {
	inputSys := NewInputSystem()
	world := NewWorld()
	exprSys := NewExpressionSystem(world, nil)

	// Initially nil
	if inputSys.expressionSystem != nil {
		t.Error("Expected expression system to be nil initially")
	}

	// Set it
	inputSys.SetExpressionSystem(exprSys)

	// Verify it's set
	if inputSys.expressionSystem == nil {
		t.Fatal("Expected expression system to be set")
	}
	if inputSys.expressionSystem != exprSys {
		t.Error("Expected expression system reference to match")
	}

	// Can set to nil
	inputSys.SetExpressionSystem(nil)
	if inputSys.expressionSystem != nil {
		t.Error("Expected expression system to be nil after setting to nil")
	}
}
