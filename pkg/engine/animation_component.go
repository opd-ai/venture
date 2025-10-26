// Package engine provides animation component for multi-frame sprite animations.
package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// AnimationState represents the current animation state of an entity.
type AnimationState string

const (
	// Animation states for entities
	AnimationStateIdle   AnimationState = "idle"
	AnimationStateWalk   AnimationState = "walk"
	AnimationStateRun    AnimationState = "run"
	AnimationStateAttack AnimationState = "attack"
	AnimationStateCast   AnimationState = "cast"
	AnimationStateHit    AnimationState = "hit"
	AnimationStateDeath  AnimationState = "death"
	AnimationStateJump   AnimationState = "jump"
	AnimationStateCrouch AnimationState = "crouch"
	AnimationStateUse    AnimationState = "use"
)

// String returns the string representation of the animation state.
func (a AnimationState) String() string {
	return string(a)
}

// AnimationComponent holds animation state and frame data for an entity.
// Integrates with ECS to provide multi-frame sprite animations.
type AnimationComponent struct {
	// Current animation state
	CurrentState AnimationState

	// Previous state (for transition detection)
	PreviousState AnimationState

	// Frames for current animation (cached)
	Frames []*ebiten.Image

	// Current frame index
	FrameIndex int

	// Time per frame (in seconds)
	FrameTime float64

	// Time accumulator for frame transitions
	TimeAccumulator float64

	// Whether animation loops
	Loop bool

	// Callback when animation completes (for one-shot animations)
	OnComplete func()

	// Whether animation is playing
	Playing bool

	// Base seed for procedural frame generation
	Seed int64

	// Frame count for current state
	FrameCount int

	// Dirty flag - regenerate frames if true
	Dirty bool

	// Last facing direction (for maintaining direction during idle)
	LastFacing string
}

// Type returns the component type identifier.
func (a *AnimationComponent) Type() string {
	return "animation"
}

// Play starts the animation from the beginning.
func (a *AnimationComponent) Play() {
	a.Playing = true
	a.FrameIndex = 0
	a.TimeAccumulator = 0.0
}

// Pause pauses the animation at the current frame.
func (a *AnimationComponent) Pause() {
	a.Playing = false
}

// Resume resumes the animation from the current frame.
func (a *AnimationComponent) Resume() {
	a.Playing = true
}

// Stop stops the animation and resets to first frame.
func (a *AnimationComponent) Stop() {
	a.Playing = false
	a.FrameIndex = 0
	a.TimeAccumulator = 0.0
}

// SetState changes the animation state and marks frames as dirty.
// Frames will be regenerated on next update.
func (a *AnimationComponent) SetState(state AnimationState) {
	if a.CurrentState != state {
		a.PreviousState = a.CurrentState
		a.CurrentState = state
		a.Dirty = true
		a.FrameIndex = 0
		a.TimeAccumulator = 0.0
		a.Playing = true // CRITICAL: Always start playing when state changes

		// CRITICAL FIX: Set loop based on animation type
		// Action animations (attack, hit, death, cast, use) should play once
		// Movement animations (idle, walk, run, jump, crouch) should loop
		switch state {
		case AnimationStateAttack, AnimationStateHit, AnimationStateDeath,
			AnimationStateCast, AnimationStateUse:
			a.Loop = false // Play once, then call OnComplete
		case AnimationStateIdle, AnimationStateWalk, AnimationStateRun,
			AnimationStateJump, AnimationStateCrouch:
			a.Loop = true // Loop continuously
		default:
			a.Loop = true // Default to looping
		}
	} else {
		// CRITICAL FIX: If setting to same state (e.g., attack → attack),
		// restart the animation from the beginning. This allows re-triggering
		// attack animations or other actions without changing state.
		a.FrameIndex = 0
		a.TimeAccumulator = 0.0
		a.Playing = true
		a.Dirty = true // Force regeneration to ensure frames are fresh
	}
}

// CurrentFrame returns the current frame image.
// Returns nil if no frames are available.
func (a *AnimationComponent) CurrentFrame() *ebiten.Image {
	if len(a.Frames) == 0 || a.FrameIndex >= len(a.Frames) {
		return nil
	}
	return a.Frames[a.FrameIndex]
}

// IsComplete returns true if a non-looping animation has completed.
func (a *AnimationComponent) IsComplete() bool {
	if a.Loop {
		return false
	}
	return a.FrameIndex >= len(a.Frames)-1 && !a.Playing
}

// Reset resets the animation to initial state.
func (a *AnimationComponent) Reset() {
	a.FrameIndex = 0
	a.TimeAccumulator = 0.0
	a.Playing = false
	a.CurrentState = AnimationStateIdle
	a.PreviousState = AnimationStateIdle
	a.Dirty = true
}

// NewAnimationComponent creates a new animation component with default values.
func NewAnimationComponent(seed int64) *AnimationComponent {
	return &AnimationComponent{
		CurrentState:    AnimationStateIdle,
		PreviousState:   AnimationStateIdle,
		Frames:          nil,
		FrameIndex:      0,
		FrameTime:       0.1, // 10 FPS default
		TimeAccumulator: 0.0,
		Loop:            true,
		Playing:         true,
		Seed:            seed,
		FrameCount:      4, // Default 4 frames per animation
		Dirty:           true,
	}
}
