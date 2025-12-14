package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// BenchmarkInputSystem_DetectAnyKeyPress_Baseline measures the ORIGINAL implementation
// that allocates a new slice every call. This is for comparison purposes.
func BenchmarkInputSystem_DetectAnyKeyPress_Baseline(b *testing.B) {
	input := &EbitenInput{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input.AnyKeyPressed = false
		// Original implementation: allocates slice every call
		if !input.AnyKeyPressed {
			pressedKeys := inpututil.AppendPressedKeys(nil)
			if len(pressedKeys) > 0 {
				input.AnyKeyPressed = true
			}
		}
	}
}

// BenchmarkInputSystem_DetectAnyKeyPress measures the performance of detectAnyKeyPress
// with the buffer reuse optimization. This function is called every frame for each
// player entity with an input component.
func BenchmarkInputSystem_DetectAnyKeyPress(b *testing.B) {
	inputSys := NewInputSystem()
	input := &EbitenInput{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset the flag as would happen in resetInputFlags each frame
		input.AnyKeyPressed = false
		inputSys.detectAnyKeyPress(input)
	}
}

// BenchmarkInputSystem_IsAnyKeyPressed measures the performance of IsAnyKeyPressed
// with the buffer reuse optimization.
func BenchmarkInputSystem_IsAnyKeyPressed(b *testing.B) {
	inputSys := NewInputSystem()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = inputSys.IsAnyKeyPressed()
	}
}

// BenchmarkInputSystem_GetPressedKeys measures the performance of GetPressedKeys.
// Note: This method still allocates on non-empty results for API safety.
func BenchmarkInputSystem_GetPressedKeys(b *testing.B) {
	inputSys := NewInputSystem()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = inputSys.GetPressedKeys()
	}
}

// BenchmarkInputSystem_GetAnyPressedKey measures the performance of GetAnyPressedKey
// with the buffer reuse optimization.
func BenchmarkInputSystem_GetAnyPressedKey(b *testing.B) {
	inputSys := NewInputSystem()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = inputSys.GetAnyPressedKey()
	}
}

// BenchmarkInputSystem_ProcessEntityInputs simulates the per-frame input processing
// for multiple player entities.
func BenchmarkInputSystem_ProcessEntityInputs(b *testing.B) {
	world := NewWorld()
	inputSys := NewInputSystem()

	// Create 4 player entities (typical multiplayer scenario)
	entities := make([]*Entity, 4)
	for i := 0; i < 4; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&EbitenInput{})
		entity.AddComponent(&VelocityComponent{})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inputSys.processEntityInputs(entities, 0.016)
	}
}
