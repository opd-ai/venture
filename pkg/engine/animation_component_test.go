package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestAnimationComponent_NewAnimationComponent tests component initialization.
func TestAnimationComponent_NewAnimationComponent(t *testing.T) {
	seed := int64(12345)
	anim := NewAnimationComponent(seed)

	if anim == nil {
		t.Fatal("NewAnimationComponent returned nil")
	}

	if anim.Seed != seed {
		t.Errorf("Expected seed %d, got %d", seed, anim.Seed)
	}

	if anim.CurrentState != AnimationStateIdle {
		t.Errorf("Expected initial state %v, got %v", AnimationStateIdle, anim.CurrentState)
	}

	if anim.FrameTime != 0.1 {
		t.Errorf("Expected frame time 0.1, got %f", anim.FrameTime)
	}

	if !anim.Playing {
		t.Error("Expected animation to be playing by default")
	}

	if !anim.Loop {
		t.Error("Expected animation to loop by default")
	}

	if !anim.Dirty {
		t.Error("Expected animation to be dirty initially")
	}
}

// TestAnimationComponent_Type tests component type identifier.
func TestAnimationComponent_Type(t *testing.T) {
	anim := NewAnimationComponent(12345)
	if anim.Type() != "animation" {
		t.Errorf("Expected type 'animation', got '%s'", anim.Type())
	}
}

// TestAnimationComponent_PlayPauseResume tests playback control.
func TestAnimationComponent_PlayPauseResume(t *testing.T) {
	anim := NewAnimationComponent(12345)

	// Test pause
	anim.Pause()
	if anim.Playing {
		t.Error("Expected animation to be paused")
	}

	// Test resume
	anim.Resume()
	if !anim.Playing {
		t.Error("Expected animation to be playing after resume")
	}

	// Test play resets frame
	anim.FrameIndex = 5
	anim.TimeAccumulator = 0.5
	anim.Play()

	if anim.FrameIndex != 0 {
		t.Errorf("Expected frame index 0 after play, got %d", anim.FrameIndex)
	}

	if anim.TimeAccumulator != 0.0 {
		t.Errorf("Expected time accumulator 0.0 after play, got %f", anim.TimeAccumulator)
	}
}

// TestAnimationComponent_Stop tests stopping animation.
func TestAnimationComponent_Stop(t *testing.T) {
	anim := NewAnimationComponent(12345)
	anim.FrameIndex = 3
	anim.TimeAccumulator = 0.25
	anim.Stop()

	if anim.Playing {
		t.Error("Expected animation to be stopped")
	}

	if anim.FrameIndex != 0 {
		t.Errorf("Expected frame index 0 after stop, got %d", anim.FrameIndex)
	}

	if anim.TimeAccumulator != 0.0 {
		t.Errorf("Expected time accumulator 0.0 after stop, got %f", anim.TimeAccumulator)
	}
}

// TestAnimationComponent_SetState tests state transitions.
func TestAnimationComponent_SetState(t *testing.T) {
	anim := NewAnimationComponent(12345)
	anim.Dirty = false
	anim.FrameIndex = 5

	// Set new state
	anim.SetState(AnimationStateWalk)

	if anim.CurrentState != AnimationStateWalk {
		t.Errorf("Expected state %v, got %v", AnimationStateWalk, anim.CurrentState)
	}

	if anim.PreviousState != AnimationStateIdle {
		t.Errorf("Expected previous state %v, got %v", AnimationStateIdle, anim.PreviousState)
	}

	if !anim.Dirty {
		t.Error("Expected dirty flag to be set after state change")
	}

	if anim.FrameIndex != 0 {
		t.Errorf("Expected frame index to reset to 0, got %d", anim.FrameIndex)
	}

	// Setting same state should restart animation (allows re-triggering attacks, etc.)
	anim.Dirty = false
	anim.FrameIndex = 3
	anim.TimeAccumulator = 0.25
	anim.SetState(AnimationStateWalk)

	if !anim.Dirty {
		t.Error("Expected dirty flag to be set when restarting animation with same state")
	}

	if anim.FrameIndex != 0 {
		t.Errorf("Expected frame index to reset to 0 when restarting, got %d", anim.FrameIndex)
	}

	if anim.TimeAccumulator != 0.0 {
		t.Errorf("Expected time accumulator to reset to 0.0 when restarting, got %f", anim.TimeAccumulator)
	}

	if !anim.Playing {
		t.Error("Expected animation to be playing after restart")
	}
}

// TestAnimationComponent_CurrentFrame tests frame retrieval.
func TestAnimationComponent_CurrentFrame(t *testing.T) {
	anim := NewAnimationComponent(12345)

	// No frames
	if frame := anim.CurrentFrame(); frame != nil {
		t.Error("Expected nil frame when no frames loaded")
	}

	// Add frames
	anim.Frames = []*ebiten.Image{
		ebiten.NewImage(10, 10),
		ebiten.NewImage(10, 10),
		ebiten.NewImage(10, 10),
	}

	// Get first frame
	frame := anim.CurrentFrame()
	if frame == nil {
		t.Error("Expected non-nil frame")
	}

	// Get middle frame
	anim.FrameIndex = 1
	frame = anim.CurrentFrame()
	if frame != anim.Frames[1] {
		t.Error("Expected frame at index 1")
	}

	// Out of bounds
	anim.FrameIndex = 10
	frame = anim.CurrentFrame()
	if frame != nil {
		t.Error("Expected nil for out of bounds index")
	}
}

// TestAnimationComponent_IsComplete tests completion detection.
func TestAnimationComponent_IsComplete(t *testing.T) {
	anim := NewAnimationComponent(12345)
	anim.Frames = []*ebiten.Image{
		ebiten.NewImage(10, 10),
		ebiten.NewImage(10, 10),
	}

	// Looping animation never completes
	anim.Loop = true
	anim.FrameIndex = 1
	anim.Playing = false

	if anim.IsComplete() {
		t.Error("Expected looping animation to not be complete")
	}

	// Non-looping animation completes at last frame
	anim.Loop = false
	if !anim.IsComplete() {
		t.Error("Expected non-looping animation at last frame to be complete")
	}

	// Not complete if still playing
	anim.Playing = true
	if anim.IsComplete() {
		t.Error("Expected playing animation to not be complete")
	}

	// Not complete if not at last frame
	anim.FrameIndex = 0
	anim.Playing = false
	if anim.IsComplete() {
		t.Error("Expected animation not at last frame to not be complete")
	}
}

// TestAnimationComponent_Reset tests reset functionality.
func TestAnimationComponent_Reset(t *testing.T) {
	anim := NewAnimationComponent(12345)
	anim.CurrentState = AnimationStateAttack
	anim.PreviousState = AnimationStateWalk
	anim.FrameIndex = 5
	anim.TimeAccumulator = 0.5
	anim.Playing = true
	anim.Dirty = false

	anim.Reset()

	if anim.FrameIndex != 0 {
		t.Errorf("Expected frame index 0, got %d", anim.FrameIndex)
	}

	if anim.TimeAccumulator != 0.0 {
		t.Errorf("Expected time accumulator 0.0, got %f", anim.TimeAccumulator)
	}

	if anim.Playing {
		t.Error("Expected animation to be stopped")
	}

	if anim.CurrentState != AnimationStateIdle {
		t.Errorf("Expected state %v, got %v", AnimationStateIdle, anim.CurrentState)
	}

	if anim.PreviousState != AnimationStateIdle {
		t.Errorf("Expected previous state %v, got %v", AnimationStateIdle, anim.PreviousState)
	}

	if !anim.Dirty {
		t.Error("Expected dirty flag to be set")
	}
}

// TestAnimationState_String tests state string representation.
func TestAnimationState_String(t *testing.T) {
	tests := []struct {
		state    AnimationState
		expected string
	}{
		{AnimationStateIdle, "idle"},
		{AnimationStateWalk, "walk"},
		{AnimationStateRun, "run"},
		{AnimationStateAttack, "attack"},
		{AnimationStateCast, "cast"},
		{AnimationStateHit, "hit"},
		{AnimationStateDeath, "death"},
		{AnimationStateJump, "jump"},
		{AnimationStateCrouch, "crouch"},
		{AnimationStateUse, "use"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.state.String() != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, tt.state.String())
			}
		})
	}
}

// TestAnimationComponent_OnComplete tests completion callback.
func TestAnimationComponent_OnComplete(t *testing.T) {
	anim := NewAnimationComponent(12345)
	anim.Frames = []*ebiten.Image{
		ebiten.NewImage(10, 10),
		ebiten.NewImage(10, 10),
	}
	anim.Loop = false

	// Set callback
	callbackCalled := false
	anim.OnComplete = func() {
		callbackCalled = true
	}

	// Manually trigger completion (normally done by AnimationSystem)
	anim.FrameIndex = 1
	anim.Playing = false
	if anim.OnComplete != nil {
		anim.OnComplete()
	}

	if !callbackCalled {
		t.Error("Expected OnComplete callback to be called")
	}
}

// BenchmarkAnimationComponent_SetState benchmarks state transitions.
func BenchmarkAnimationComponent_SetState(b *testing.B) {
	anim := NewAnimationComponent(12345)

	states := []AnimationState{
		AnimationStateIdle,
		AnimationStateWalk,
		AnimationStateAttack,
		AnimationStateCast,
		AnimationStateHit,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		anim.SetState(states[i%len(states)])
	}
}

// BenchmarkAnimationComponent_CurrentFrame benchmarks frame retrieval.
func BenchmarkAnimationComponent_CurrentFrame(b *testing.B) {
	anim := NewAnimationComponent(12345)
	anim.Frames = make([]*ebiten.Image, 8)
	for i := range anim.Frames {
		anim.Frames[i] = ebiten.NewImage(28, 28)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		anim.FrameIndex = i % 8
		_ = anim.CurrentFrame()
	}
}
